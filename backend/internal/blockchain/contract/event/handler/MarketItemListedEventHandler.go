package handler

import (
	"context" // 新增ctx依赖
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	service2 "nft_backend/internal/app/service"
	"nft_backend/internal/blockchain/contract/event/constant"
	"nft_backend/internal/blockchain/contract/event/type"
	"nft_backend/internal/di"
	"nft_backend/internal/model"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type MarketItemListedEventHandler struct {
	eventName      string
	eventService   *service2.EventService
	nftService     *service2.NftService
	operateService *service2.OperateService // 新增：操作记录服务依赖
}

func NewMarketItemListedEventHandler(operateService *service2.OperateService) *MarketItemListedEventHandler {
	// 依赖注入（补充错误日志，便于排查依赖问题）
	eventService, err := di.Resolve[*service2.EventService]()
	if err != nil {
		log.Printf("解析EventService依赖失败: %v", err)
	}
	nftService, err := di.Resolve[*service2.NftService]()
	if err != nil {
		log.Printf("解析NFTService依赖失败: %v", err)
	}

	return &MarketItemListedEventHandler{
		eventName:      _type.ItemListed.String(),
		eventService:   eventService,
		nftService:     nftService,
		operateService: operateService, // 注入操作记录服务
	}
}

// Handle 处理上架事件（核心改造：添加ctx，强化操作记录必录逻辑）
func (h *MarketItemListedEventHandler) Handle(abi abi.ABI, vLog types.Log, blockTime string) error {
	// 1. 创建带超时的ctx，传递全链路（控制事件处理整体超时）
	ctx, cancel := context.WithTimeout(context.Background(), constant.EventHandleTimeout)
	defer cancel() // 函数结束释放资源

	// 2. 补充上下文日志（TxHash+LogIndex，便于精准定位）
	txHash := vLog.TxHash.Hex()
	logIndex := vLog.Index
	tokenIdStr := "" // 提前声明，便于后续日志使用
	log.Printf("[ListedEvent][Tx:%s][Log:%d] 开始处理上架事件，区块时间：%s", txHash, logIndex, blockTime)

	// ========== 1. 独有逻辑：解析Topic（上架事件专属） ==========
	sellerAddress, contractAddress, tokenId, err := h.parseTopics(vLog.Topics)
	if err != nil {
		return fmt.Errorf("[ListedEvent][Tx:%s][Log:%d] 解析Topic失败: %w", txHash, logIndex, err)
	}
	tokenIdStr = tokenId.String()

	// ========== 2. 独有逻辑：解析Data（上架事件专属） ==========
	extraData, err := h.parseData(abi, vLog.Data)
	if err != nil {
		return fmt.Errorf("[ListedEvent][Tx:%s][Log:%d][Token:%s] 解析Data失败: %w",
			txHash, logIndex, tokenIdStr, err)
	}

	// ========== 3. 独有逻辑：构建Event对象 ==========
	newEvent := &model.Event{
		EventType:       h.eventName,
		TxHash:          txHash,
		BlockNumber:     vLog.BlockNumber,
		LogIndex:        logIndex,
		ContractAddress: contractAddress, // 保留修正后的合约地址解析逻辑
		TokenID:         tokenIdStr,
		FromAddress:     sellerAddress,
		ToAddress:       sellerAddress,
		Amount:          "",
		ExtraData:       extraData,
	}
	log.Printf("[ListedEvent][Tx:%s][Log:%d][Token:%s] 解析完成的上架事件：%+v",
		txHash, logIndex, tokenIdStr, newEvent)

	// ========== 4. 通用逻辑：替换CommonSaveEvent为显式CRUD（保证ctx传递） ==========
	// 空指针校验：eventService不能为空
	if h.eventService == nil {
		return fmt.Errorf("[ListedEvent][Tx:%s][Log:%d][Token:%s] eventService依赖未初始化",
			txHash, logIndex, tokenIdStr)
	}

	// 查询事件是否已存在
	existEvent, err := h.eventService.GetByTxHashAndLogIndex(ctx, newEvent.TxHash, newEvent.LogIndex)
	if err != nil && !errors.Is(err, service2.ErrRecordNotFound) {
		return fmt.Errorf("[ListedEvent][Tx:%s][Log:%d][Token:%s] 查询事件记录失败: %w",
			txHash, logIndex, tokenIdStr, err)
	}

	isNewEvent := false
	if errors.Is(err, service2.ErrRecordNotFound) {
		// 新事件：创建Event记录（必须成功）
		if err := h.eventService.Create(ctx, newEvent); err != nil {
			return fmt.Errorf("[ListedEvent][Tx:%s][Log:%d][Token:%s] 新增事件记录失败: %w",
				txHash, logIndex, tokenIdStr, err)
		}
		isNewEvent = true
		log.Printf("[ListedEvent][Tx:%s][Log:%d][Token:%s] 新事件，已创建Event记录",
			txHash, logIndex, tokenIdStr)
	} else {
		// 已有事件：按需更新（避免重复创建）
		if h.needUpdateEvent(existEvent, newEvent) {
			newEvent.ID = existEvent.ID
			if err := h.eventService.Update(ctx, newEvent); err != nil {
				return fmt.Errorf("[ListedEvent][Tx:%s][Log:%d][Token:%s] 更新事件记录失败: %w",
					txHash, logIndex, tokenIdStr, err)
			}
			log.Printf("[ListedEvent][Tx:%s][Log:%d][Token:%s] 已有事件，已更新Event记录",
				txHash, logIndex, tokenIdStr)
		}
	}

	// ========== 5. 独有逻辑：处理NFT上架业务（上架事件专属，传递ctx） ==========
	// 空指针校验：nftService不能为空
	if h.nftService == nil {
		return fmt.Errorf("[ListedEvent][Tx:%s][Log:%d][Token:%s] nftService依赖未初始化",
			txHash, logIndex, tokenIdStr)
	}

	var listedEventData ListedEventData
	err = json.Unmarshal([]byte(extraData), &listedEventData)
	if err != nil {
		return fmt.Errorf("[ListedEvent][Tx:%s][Log:%d][Token:%s] 解析ExtraData失败: %w",
			txHash, logIndex, tokenIdStr, err)
	}

	// 调用NFT服务更新上架状态（传递ctx）
	if err := h.nftService.ListNFT(ctx, tokenIdStr, listedEventData.Price, listedEventData.ListedAt); err != nil {
		return fmt.Errorf("[ListedEvent][Tx:%s][Log:%d][Token:%s] 更新NFT上架状态失败: %w",
			txHash, logIndex, tokenIdStr, err)
	}

	// ========== 核心改造：新事件必须创建操作日志（失败阻断主流程） ==========
	// 空指针校验：operateService不能为空（新事件必须记录，不允许跳过）
	if h.operateService == nil {
		return fmt.Errorf("[ListedEvent][Tx:%s][Log:%d][Token:%s] OperateService依赖未初始化，无法创建操作记录",
			txHash, logIndex, tokenIdStr)
	}

	// 仅新事件创建操作记录（避免重复）
	if isNewEvent {
		// 构建上架操作记录
		operateRecord := &model.NFTOperateRecord{
			ContractAddress: contractAddress,
			TokenID:         tokenIdStr,
			UserAddress:     sellerAddress,
			OwnerAddress:    sellerAddress,
			OperateType:     string(constant.OperateTypeListed), // 使用常量，避免硬编码
			Amount:          listedEventData.Price,
			TxHash:          txHash,
			Status:          string(constant.OperateStatusSuccess), // 使用常量
			BlockNumber:     fmt.Sprintf("%d", vLog.BlockNumber),
			OperateAt:       time.Now(),
			Remark:          fmt.Sprintf("NFT上架，价格：%s ETH，上架时间：%s", listedEventData.Price, listedEventData.ListedAt),
		}

		// 核心变更：创建失败直接返回错误（阻断主流程，保证操作日志必录）
		if err := h.operateService.CreateOperateRecord(ctx, operateRecord); err != nil {
			return fmt.Errorf("[ListedEvent][Tx:%s][Log:%d][Token:%s] 创建上架操作记录失败（新事件必须记录）: %w",
				txHash, logIndex, tokenIdStr, err)
		}
		log.Printf("[ListedEvent][Tx:%s][Log:%d][Token:%s] 新事件，上架操作记录创建成功",
			txHash, logIndex, tokenIdStr)
	} else {
		log.Printf("[ListedEvent][Tx:%s][Log:%d][Token:%s] 已有事件，跳过重复创建操作记录",
			txHash, logIndex, tokenIdStr)
	}

	log.Printf("[ListedEvent][Tx:%s][Log:%d][Token:%s] 上架事件处理完成，TokenID：%s",
		txHash, logIndex, tokenIdStr, tokenIdStr)
	return nil
}

// needUpdateEvent 判断事件是否需要更新（补充实现，对比核心字段）
func (h *MarketItemListedEventHandler) needUpdateEvent(old, new *model.Event) bool {
	if old == nil || new == nil {
		return true
	}
	return old.ExtraData != new.ExtraData ||
		old.ContractAddress != new.ContractAddress ||
		old.Amount != new.Amount
}

// 上架事件专属：解析Topic（保留原有修正的合约地址解析逻辑）
func (h *MarketItemListedEventHandler) parseTopics(topics []common.Hash) (string, string, *big.Int, error) {
	if len(topics) < 4 {
		return "", "", nil, fmt.Errorf("Topics长度不足，期望至少4个，实际：%d", len(topics))
	}

	// 卖家地址 → topics[1]
	sellerAddrBytes := topics[1].Bytes()[12:]
	sellerAddress := common.BytesToAddress(sellerAddrBytes).Hex()

	// 合约地址 → topics[2]（保留修正后的逻辑）
	contractAddrBytes := topics[2].Bytes()[12:]
	contractAddress := common.BytesToAddress(contractAddrBytes).Hex()

	// tokenId → topics[3]
	tokenId := new(big.Int).SetBytes(topics[3].Bytes())
	if tokenId.Cmp(big.NewInt(0)) < 0 {
		return "", "", nil, errors.New("tokenId不能为负数")
	}

	// 地址合法性校验
	if !common.IsHexAddress(sellerAddress) {
		return "", "", nil, fmt.Errorf("解析的卖家地址不合法：%s", sellerAddress)
	}
	if !common.IsHexAddress(contractAddress) {
		return "", "", nil, fmt.Errorf("解析的合约地址不合法：%s", contractAddress)
	}

	return sellerAddress, contractAddress, tokenId, nil
}

// 上架事件专属：解析Data（保留原有逻辑）
func (h *MarketItemListedEventHandler) parseData(abi abi.ABI, data []byte) (string, error) {
	eventData := make(map[string]interface{})
	if len(data) > 0 {
		if err := abi.UnpackIntoMap(eventData, h.eventName, data); err != nil {
			return "", fmt.Errorf("UnpackIntoMap失败: %w", err)
		}
	}

	// 提取price和listedAt（上架事件专属字段）
	price, ok := eventData["price"].(*big.Int)
	if !ok {
		return "", errors.New("Price字段解析失败")
	}
	listedAt, ok := eventData["listedAt"].(*big.Int)
	if !ok {
		return "", errors.New("ListedAt字段解析失败")
	}

	// 序列化为JSON
	extraData, err := json.Marshal(ListedEventData{
		Price:    price.String(),
		ListedAt: listedAt.String(),
	})
	if err != nil {
		return "", fmt.Errorf("序列化ExtraData失败: %w", err)
	}

	return string(extraData), nil
}
