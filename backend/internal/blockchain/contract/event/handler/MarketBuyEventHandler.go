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

type MarketBuyEventHandler struct {
	eventName      string
	eventService   *service2.EventService
	nftService     *service2.NftService
	operateService *service2.OperateService // 新增：操作记录服务依赖
}

func NewMarketBuyEventHandler(operateService *service2.OperateService) *MarketBuyEventHandler {
	// 依赖注入（补充错误日志，便于排查依赖问题）
	eventService, err := di.Resolve[*service2.EventService]()
	if err != nil {
		log.Printf("解析EventService依赖失败: %v", err)
	}
	nftService, err := di.Resolve[*service2.NftService]()
	if err != nil {
		log.Printf("解析NFTService依赖失败: %v", err)
	}

	return &MarketBuyEventHandler{
		eventName:      _type.Buy.String(),
		eventService:   eventService,
		nftService:     nftService,
		operateService: operateService, // 注入操作记录服务
	}
}

// Handle 处理购买事件（核心改造：添加ctx，强化操作记录必录逻辑）
func (h *MarketBuyEventHandler) Handle(abi abi.ABI, vLog types.Log, blockTime string) error {
	// 1. 创建带超时的ctx，传递全链路（控制事件处理整体超时）
	ctx, cancel := context.WithTimeout(context.Background(), constant.EventHandleTimeout)
	defer cancel() // 函数结束释放资源

	// 2. 补充上下文日志（TxHash+LogIndex，便于精准定位）
	txHash := vLog.TxHash.Hex()
	logIndex := vLog.Index
	tokenIdStr := "" // 提前声明，便于后续日志使用
	log.Printf("[BuyEvent][Tx:%s][Log:%d] 开始处理购买事件，区块时间：%s", txHash, logIndex, blockTime)

	// ========== 步骤1：解析Topic（索引字段） ==========
	buyAddress, tokenId, err := h.parseTopics(vLog.Topics)
	if err != nil {
		return fmt.Errorf("[BuyEvent][Tx:%s][Log:%d] 解析Topic失败: %w", txHash, logIndex, err)
	}
	tokenIdStr = tokenId.String()

	// ========== 步骤2：解析Data（非索引字段） ==========
	extraData, err := h.parseData(abi, vLog.Data)
	if err != nil {
		return fmt.Errorf("[BuyEvent][Tx:%s][Log:%d][Token:%s] 解析Data失败: %w",
			txHash, logIndex, tokenIdStr, err)
	}

	// ========== 步骤3：构建Event对象（适配结构体） ==========
	newEvent := &model.Event{
		EventType:       h.eventName,
		TxHash:          txHash,
		BlockNumber:     vLog.BlockNumber,
		LogIndex:        logIndex,
		ContractAddress: vLog.Address.Hex(),
		TokenID:         tokenIdStr,
		FromAddress:     buyAddress,
		ToAddress:       buyAddress,
		Amount:          "",
		ExtraData:       extraData,
	}

	log.Printf("[BuyEvent][Tx:%s][Log:%d][Token:%s] 解析完成的购买事件：%+v",
		txHash, logIndex, tokenIdStr, newEvent)

	// ========== 步骤4：核心逻辑：查询-比对-新增/更新（传递ctx） ==========
	// 空指针校验：eventService不能为空
	if h.eventService == nil {
		return fmt.Errorf("[BuyEvent][Tx:%s][Log:%d][Token:%s] eventService依赖未初始化",
			txHash, logIndex, tokenIdStr)
	}

	existEvent, err := h.eventService.GetByTxHashAndLogIndex(ctx, newEvent.TxHash, newEvent.LogIndex)
	if err != nil {
		if !errors.Is(err, service2.ErrRecordNotFound) {
			return fmt.Errorf("[BuyEvent][Tx:%s][Log:%d][Token:%s] 查询本地购买事件记录失败: %w",
				txHash, logIndex, tokenIdStr, err)
		}

		// 情况1：无记录 → 新增（传递ctx）
		log.Printf("[BuyEvent][Tx:%s][Log:%d][Token:%s] 本地无该购买事件记录，执行新增",
			txHash, logIndex, tokenIdStr)
		if err := h.eventService.Create(ctx, newEvent); err != nil {
			return fmt.Errorf("[BuyEvent][Tx:%s][Log:%d][Token:%s] 存储购买事件失败: %w",
				txHash, logIndex, tokenIdStr, err)
		}
	} else {
		// 情况2：有记录 → 比对后更新（传递ctx）
		if h.isEventChanged(existEvent, newEvent) {
			newEvent.ID = existEvent.ID
			log.Printf("[BuyEvent][Tx:%s][Log:%d][Token:%s] 购买事件记录有变更，执行覆盖更新",
				txHash, logIndex, tokenIdStr)
			if err := h.eventService.Update(ctx, newEvent); err != nil {
				return fmt.Errorf("[BuyEvent][Tx:%s][Log:%d][Token:%s] 更新购买事件记录失败: %w",
					txHash, logIndex, tokenIdStr, err)
			}
		} else {
			log.Printf("[BuyEvent][Tx:%s][Log:%d][Token:%s] 购买事件记录无变更，跳过更新",
				txHash, logIndex, tokenIdStr)
		}
	}

	// 标记是否为新事件（用于后续操作记录判断）
	isNewEvent := errors.Is(err, service2.ErrRecordNotFound)

	// ========== 步骤5：同步NFT所有者信息（传递ctx） ==========
	// 空指针校验：nftService不能为空
	if h.nftService == nil {
		return fmt.Errorf("[BuyEvent][Tx:%s][Log:%d][Token:%s] nftService依赖未初始化",
			txHash, logIndex, tokenIdStr)
	}

	var buyEventData BuyEventData
	err = json.Unmarshal([]byte(extraData), &buyEventData)
	if err != nil {
		return fmt.Errorf("[BuyEvent][Tx:%s][Log:%d][Token:%s] 解析购买事件ExtraData失败: %w",
			txHash, logIndex, tokenIdStr, err)
	}
	if err := h.nftService.ChangeOwner(ctx, tokenIdStr, buyAddress, buyEventData.BuyAt); err != nil {
		return fmt.Errorf("[BuyEvent][Tx:%s][Log:%d][Token:%s] 修改NFT所有者记录失败: %w",
			txHash, logIndex, tokenIdStr, err)
	}

	// ========== 核心改造：新事件必须创建操作日志（失败阻断主流程） ==========
	// 空指针校验：operateService不能为空（新事件必须记录，不允许跳过）
	if h.operateService == nil {
		return fmt.Errorf("[BuyEvent][Tx:%s][Log:%d][Token:%s] OperateService依赖未初始化，无法创建操作记录",
			txHash, logIndex, tokenIdStr)
	}

	// 仅新事件创建操作记录（避免重复）
	if isNewEvent {
		// 构建购买操作记录（购买事件核心字段）
		operateRecord := &model.NFTOperateRecord{
			ContractAddress: vLog.Address.Hex(),
			TokenID:         tokenIdStr,
			UserAddress:     buyAddress,
			OwnerAddress:    buyAddress,
			OperateType:     string(constant.OperateTypeBuy), // 使用常量，避免硬编码
			Amount:          buyEventData.Price,
			TxHash:          txHash,
			Status:          string(constant.OperateStatusSuccess), // 使用常量
			BlockNumber:     fmt.Sprintf("%d", vLog.BlockNumber),
			OperateAt:       time.Now(),
			Remark:          fmt.Sprintf("NFT购买，价格：%s ETH，购买时间：%s", buyEventData.Price, buyEventData.BuyAt),
		}

		// 核心变更：创建失败直接返回错误（阻断主流程，保证操作日志必录）
		if err := h.operateService.CreateOperateRecord(ctx, operateRecord); err != nil {
			return fmt.Errorf("[BuyEvent][Tx:%s][Log:%d][Token:%s] 创建购买操作记录失败（新事件必须记录）: %w",
				txHash, logIndex, tokenIdStr, err)
		}
		log.Printf("[BuyEvent][Tx:%s][Log:%d][Token:%s] 新事件，购买操作记录创建成功",
			txHash, logIndex, tokenIdStr)
	} else {
		log.Printf("[BuyEvent][Tx:%s][Log:%d][Token:%s] 已有事件，跳过重复创建操作记录",
			txHash, logIndex, tokenIdStr)
	}

	log.Printf("[BuyEvent][Tx:%s][Log:%d][Token:%s] 购买事件处理完成，TokenID：%s",
		txHash, logIndex, tokenIdStr, tokenIdStr)
	return nil
}

// 字段比对方法（适配Event结构体，保留原有逻辑）
func (h *MarketBuyEventHandler) isEventChanged(old, new *model.Event) bool {
	if old.EventType != new.EventType ||
		old.TokenID != new.TokenID ||
		old.FromAddress != new.FromAddress ||
		old.ToAddress != new.ToAddress ||
		old.ExtraData != new.ExtraData ||
		old.BlockNumber != new.BlockNumber ||
		old.ContractAddress != new.ContractAddress ||
		old.Amount != new.Amount {
		return true
	}
	return false
}

// 解析购买事件Topic（保留原有bug修正逻辑）
func (h *MarketBuyEventHandler) parseTopics(topics []common.Hash) (string, *big.Int, error) {
	// 修正1：Topic长度提示错误（期望3个，原代码写的4个）
	if len(topics) < 3 {
		return "", nil, fmt.Errorf("Topics长度不足，期望至少3个，实际：%d", len(topics))
	}

	// 买家地址（topics[1]）
	buyAddrBytes := topics[1].Bytes()[12:]
	buyAddress := common.BytesToAddress(buyAddrBytes).Hex()

	// tokenId（topics[2]）
	tokenId := new(big.Int).SetBytes(topics[2].Bytes())
	if tokenId.Cmp(big.NewInt(0)) < 0 {
		return "", nil, errors.New("tokenId不能为负数")
	}

	// 校验地址合法性
	if !common.IsHexAddress(buyAddress) {
		return "", nil, fmt.Errorf("解析的买家地址不合法：%s", buyAddress)
	}

	return buyAddress, tokenId, nil
}

// 解析购买事件Data（保留原有bug修正逻辑）
func (h *MarketBuyEventHandler) parseData(abi abi.ABI, data []byte) (string, error) {
	eventData := make(map[string]interface{})

	if len(data) > 0 {
		if err := abi.UnpackIntoMap(eventData, h.eventName, data); err != nil {
			return "", fmt.Errorf("UnpackIntoMap失败: %w", err)
		}
	}

	log.Printf("[BuyEvent] 解析出的eventData: %+v", eventData)

	// 提取price字段
	price, ok := eventData["price"].(*big.Int)
	if !ok {
		return "", errors.New("Price字段解析失败（购买事件）")
	}

	// 修正2：字段名从listedAt改为buyAt（和JSON字段对应）
	buyAt, ok := eventData["buyAt"].(*big.Int)
	if !ok {
		return "", errors.New("BuyAt字段解析失败（购买事件）")
	}

	// 序列化ExtraData（使用结构化结构体，更规范）
	extraData, err := json.Marshal(BuyEventData{
		Price: price.String(),
		BuyAt: buyAt.String(),
	})
	if err != nil {
		return "", fmt.Errorf("序列化购买事件ExtraData失败: %w", err)
	}

	return string(extraData), nil
}
