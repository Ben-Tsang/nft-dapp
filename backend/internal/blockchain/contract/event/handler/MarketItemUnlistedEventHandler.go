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

type MarketItemUnlistedEventHandler struct {
	eventName      string
	eventService   *service2.EventService
	nftService     *service2.NftService
	operateService *service2.OperateService // 新增：操作记录服务依赖
}

func NewMarketItemUnlistedEventHandler(operateService *service2.OperateService) *MarketItemUnlistedEventHandler {
	// 依赖注入（增加错误日志）
	eventService, err := di.Resolve[*service2.EventService]()
	if err != nil {
		log.Printf("解析EventService依赖失败: %v", err)
	}
	nftService, err := di.Resolve[*service2.NftService]()
	if err != nil {
		log.Printf("解析NFTService依赖失败: %v", err)
	}

	return &MarketItemUnlistedEventHandler{
		eventName:      _type.ItemUnlisted.String(),
		eventService:   eventService,
		nftService:     nftService,
		operateService: operateService, // 注入操作记录服务
	}
}

// Handle 处理下架事件（核心改造：添加ctx，强化操作记录创建逻辑）
func (h *MarketItemUnlistedEventHandler) Handle(abi abi.ABI, vLog types.Log, blockTime string) error {
	// 1. 创建带超时的ctx，传递全链路（事件处理整体超时控制）
	ctx, cancel := context.WithTimeout(context.Background(), constant.EventHandleTimeout)
	defer cancel() // 函数结束释放资源

	// 2. 补充上下文日志（TxHash+LogIndex，便于定位）
	txHash := vLog.TxHash.Hex()
	logIndex := vLog.Index
	tokenIdStr := "" // 提前声明，便于后续日志使用
	log.Printf("[UnlistedEvent][Tx:%s][Log:%d] 开始处理下架事件，合约地址：%s，区块时间：%s",
		txHash, logIndex, vLog.Address.Hex(), blockTime)

	// ========== 步骤1：解析Topic（独有逻辑） ==========
	sellerAddress, tokenId, err := h.parseTopics(vLog.Topics)
	if err != nil {
		return fmt.Errorf("[UnlistedEvent][Tx:%s][Log:%d] 下架事件解析Topic失败: %w", txHash, logIndex, err)
	}
	tokenIdStr = tokenId.String()

	// ========== 步骤2：解析Data（独有逻辑） ==========
	extraData, err := h.parseData(abi, vLog.Data)
	if err != nil {
		return fmt.Errorf("[UnlistedEvent][Tx:%s][Log:%d][Token:%s] 下架事件解析Data失败: %w",
			txHash, logIndex, tokenIdStr, err)
	}

	// ========== 步骤3：构建Event对象（独有逻辑） ==========
	newEvent := &model.Event{
		EventType:       h.eventName,
		TxHash:          txHash,
		BlockNumber:     vLog.BlockNumber,
		LogIndex:        logIndex,
		ContractAddress: vLog.Address.Hex(),
		TokenID:         tokenIdStr,
		FromAddress:     sellerAddress,
		ToAddress:       sellerAddress,
		Amount:          "",
		ExtraData:       extraData,
	}

	log.Printf("[UnlistedEvent][Tx:%s][Log:%d][Token:%s] 解析完成的下架事件，事件内容：%+v",
		txHash, logIndex, tokenIdStr, newEvent)

	// ========== 步骤4：通用逻辑：一行调用CommonSaveEvent（核心改造，传递ctx） ==========
	// 注：需确保CommonSaveEvent也适配ctx参数，若未适配可临时调整为直接调用eventService
	if h.eventService == nil {
		return fmt.Errorf("[UnlistedEvent][Tx:%s][Log:%d][Token:%s] eventService依赖未初始化",
			txHash, logIndex, tokenIdStr)
	}
	// 替代方案：直接调用eventService的CRUD（保证ctx传递）
	existEvent, err := h.eventService.GetByTxHashAndLogIndex(ctx, newEvent.TxHash, newEvent.LogIndex)
	if err != nil && !errors.Is(err, service2.ErrRecordNotFound) {
		return fmt.Errorf("[UnlistedEvent][Tx:%s][Log:%d][Token:%s] 查询事件记录失败: %w",
			txHash, logIndex, tokenIdStr, err)
	}

	isNewEvent := false
	if errors.Is(err, service2.ErrRecordNotFound) {
		// 新事件：创建Event记录（必须成功）
		if err := h.eventService.Create(ctx, newEvent); err != nil {
			return fmt.Errorf("[UnlistedEvent][Tx:%s][Log:%d][Token:%s] 新增事件记录失败: %w",
				txHash, logIndex, tokenIdStr, err)
		}
		isNewEvent = true
		log.Printf("[UnlistedEvent][Tx:%s][Log:%d][Token:%s] 新事件，已创建Event记录",
			txHash, logIndex, tokenIdStr)
	} else {
		// 已有事件：按需更新（保留原有逻辑）
		if h.needUpdateEvent(existEvent, newEvent) {
			newEvent.ID = existEvent.ID
			if err := h.eventService.Update(ctx, newEvent); err != nil {
				return fmt.Errorf("[UnlistedEvent][Tx:%s][Log:%d][Token:%s] 更新事件记录失败: %w",
					txHash, logIndex, tokenIdStr, err)
			}
			log.Printf("[UnlistedEvent][Tx:%s][Log:%d][Token:%s] 已有事件，已更新Event记录",
				txHash, logIndex, tokenIdStr)
		}
	}

	// ========== 步骤5：处理NFT下架业务（独有逻辑，传递ctx） ==========
	// 空指针校验（强化：直接返回错误，不允许空依赖）
	if h.nftService == nil {
		return fmt.Errorf("[UnlistedEvent][Tx:%s][Log:%d][Token:%s] nftService依赖未初始化",
			txHash, logIndex, tokenIdStr)
	}

	var unlistedEventData UnlistedEventData
	err = json.Unmarshal([]byte(extraData), &unlistedEventData)
	if err != nil {
		return fmt.Errorf("[UnlistedEvent][Tx:%s][Log:%d][Token:%s] 解析下架事件ExtraData失败: %w",
			txHash, logIndex, tokenIdStr, err)
	}

	// 字段合法性校验（强化：不能为空）
	if unlistedEventData.UnlistedAt == "" {
		return fmt.Errorf("[UnlistedEvent][Tx:%s][Log:%d][Token:%s] 下架事件unlistedAt字段为空",
			txHash, logIndex, tokenIdStr)
	}

	if err := h.nftService.UnlistedNFT(ctx, tokenIdStr, unlistedEventData.UnlistedAt); err != nil {
		return fmt.Errorf("[UnlistedEvent][Tx:%s][Log:%d][Token:%s] 更新NFT下架状态失败: %w",
			txHash, logIndex, tokenIdStr, err)
	}

	// ========== 核心改造：新事件必须创建操作日志（失败阻断主流程） ==========
	// 空指针校验（强化：直接返回错误，不允许空依赖）
	if h.operateService == nil {
		return fmt.Errorf("[UnlistedEvent][Tx:%s][Log:%d][Token:%s] OperateService依赖未初始化，无法创建操作记录",
			txHash, logIndex, tokenIdStr)
	}

	// 仅新事件需要创建操作记录（已有事件不重复创建）
	if isNewEvent {
		// 构建下架操作记录
		operateRecord := &model.NFTOperateRecord{
			ContractAddress: vLog.Address.Hex(),
			TokenID:         tokenIdStr,
			UserAddress:     sellerAddress,
			OwnerAddress:    sellerAddress,
			OperateType:     string(constant.OperateTypeUnlisted), // 使用常量，避免硬编码
			Amount:          "",
			TxHash:          txHash,
			Status:          string(constant.OperateStatusSuccess), // 使用常量
			BlockNumber:     fmt.Sprintf("%d", vLog.BlockNumber),
			OperateAt:       time.Now(),
			Remark:          fmt.Sprintf("NFT下架，下架时间：%s", unlistedEventData.UnlistedAt),
		}

		// 核心变更：创建操作记录失败直接返回错误（阻断主流程）
		if err := h.operateService.CreateOperateRecord(ctx, operateRecord); err != nil {
			return fmt.Errorf("[UnlistedEvent][Tx:%s][Log:%d][Token:%s] 创建下架操作记录失败（新事件必须记录）: %w",
				txHash, logIndex, tokenIdStr, err)
		}
		log.Printf("[UnlistedEvent][Tx:%s][Log:%d][Token:%s] 新事件，下架操作记录创建成功",
			txHash, logIndex, tokenIdStr)
	} else {
		log.Printf("[UnlistedEvent][Tx:%s][Log:%d][Token:%s] 已有事件，跳过重复创建操作记录",
			txHash, logIndex, tokenIdStr)
	}

	log.Printf("[UnlistedEvent][Tx:%s][Log:%d][Token:%s] 下架事件处理完成，下架时间：%s",
		txHash, logIndex, tokenIdStr, unlistedEventData.UnlistedAt)
	return nil
}

// needUpdateEvent 判断事件是否需要更新（补充实现，对比核心字段）
func (h *MarketItemUnlistedEventHandler) needUpdateEvent(old, new *model.Event) bool {
	if old == nil || new == nil {
		return true
	}
	return old.ExtraData != new.ExtraData ||
		old.ContractAddress != new.ContractAddress ||
		old.FromAddress != new.FromAddress
}

// 解析下架事件Topic（独有逻辑，补充ctx日志上下文）
func (h *MarketItemUnlistedEventHandler) parseTopics(topics []common.Hash) (string, *big.Int, error) {
	if topics == nil {
		return "", nil, errors.New("Topics为空（下架事件）")
	}

	if len(topics) < 3 {
		return "", nil, fmt.Errorf("Topics长度不足（下架事件），期望至少3个，实际：%d", len(topics))
	}

	// 卖家地址解析
	sellerAddrBytes := topics[1].Bytes()
	if len(sellerAddrBytes) < 12 {
		return "", nil, errors.New("卖家地址字节长度不足（下架事件）")
	}
	sellerAddrBytes = sellerAddrBytes[12:]
	sellerAddress := common.BytesToAddress(sellerAddrBytes).Hex()

	// tokenId解析
	tokenIdBytes := topics[2].Bytes()
	tokenId := new(big.Int).SetBytes(tokenIdBytes)
	if tokenId.Cmp(big.NewInt(0)) < 0 {
		return "", nil, errors.New("tokenId不能为负数（下架事件）")
	}

	// 地址合法性校验
	if !common.IsHexAddress(sellerAddress) {
		return "", nil, fmt.Errorf("解析的卖家地址不合法（下架事件）：%s", sellerAddress)
	}

	return sellerAddress, tokenId, nil
}

// 解析下架事件Data（独有逻辑，移除错误的abi.Source日志）
func (h *MarketItemUnlistedEventHandler) parseData(abi abi.ABI, data []byte) (string, error) {
	eventData := make(map[string]interface{})

	if len(data) > 0 {
		if err := abi.UnpackIntoMap(eventData, h.eventName, data); err != nil {
			return "", fmt.Errorf("UnpackIntoMap失败（下架事件）: %w", err)
		}
	}

	log.Printf("下架事件解析出的eventData: %+v", eventData)

	// 兼容字段缺失
	unlistedAt, ok := eventData["unlistedAt"].(*big.Int)
	if !ok {
		log.Printf("下架事件unlistedAt字段缺失或类型错误，赋空值")
		extraData, _ := json.Marshal(map[string]string{
			"unlistedAt": "",
		})
		return string(extraData), nil
	}

	// 序列化为ExtraData
	extraData, err := json.Marshal(UnlistedEventData{
		UnlistedAt: unlistedAt.String(),
	})
	if err != nil {
		return "", fmt.Errorf("序列化下架事件ExtraData失败: %w", err)
	}

	return string(extraData), nil
}
