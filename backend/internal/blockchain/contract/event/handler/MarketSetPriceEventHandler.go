package handler

import (
	"context"
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

type MarketSetPriceEventHandler struct {
	eventName      string
	eventService   *service2.EventService
	nftService     *service2.NftService
	operateService *service2.OperateService // 新增：操作记录服务
}

func NewMarketSetPriceEventHandler(operateService *service2.OperateService) *MarketSetPriceEventHandler {
	// 从容器获取依赖
	eventService, _ := di.Resolve[*service2.EventService]()
	nftService, _ := di.Resolve[*service2.NftService]()

	return &MarketSetPriceEventHandler{
		eventName:      _type.SetPrice.String(),
		eventService:   eventService,
		nftService:     nftService,
		operateService: operateService, // 注入操作记录服务
	}
}

// Handle 处理设置价格事件，优化版
// 核心逻辑：新事件新增+插操作记录，已有事件按需更新，保证NFT价格最终一致性
func (h *MarketSetPriceEventHandler) Handle(abi abi.ABI, vLog types.Log, blockTime string) error {
	// 上下文传递，便于链路追踪和超时控制
	ctx := context.Background()
	log.Printf("[SetPriceEvent][Tx:%s][Log:%d] 开始处理设置价格事件，区块时间：%s",
		vLog.TxHash.Hex(), vLog.Index, blockTime)

	// ========== 步骤1：解析Topic（索引字段）- 增加错误日志上下文 ==========
	ownerAddress, tokenId, err := h.parseTopics(vLog.Topics)
	if err != nil {
		log.Printf("[SetPriceEvent][Tx:%s][Log:%d] 解析Topic失败: %v", vLog.TxHash.Hex(), vLog.Index, err)
		return fmt.Errorf("解析Topic失败 [Tx:%s Log:%d]: %w", vLog.TxHash.Hex(), vLog.Index, err)
	}
	tokenIDStr := tokenId.String()
	contractAddr := vLog.Address.Hex()

	// ========== 步骤2：解析Data（非索引字段）- 增加错误日志上下文 ==========
	extraData, err := h.parseData(abi, vLog.Data)
	if err != nil {
		log.Printf("[SetPriceEvent][Tx:%s][Log:%d][Token:%s] 解析Data失败: %v",
			vLog.TxHash.Hex(), vLog.Index, tokenIDStr, err)
		return fmt.Errorf("解析Data失败 [Tx:%s Log:%d Token:%s]: %w",
			vLog.TxHash.Hex(), vLog.Index, tokenIDStr, err)
	}

	// ========== 步骤3：解析ExtraData为结构化数据 - 增加校验 ==========
	var setPriceEventData SetPriceEventData
	if err := json.Unmarshal([]byte(extraData), &setPriceEventData); err != nil {
		log.Printf("[SetPriceEvent][Tx:%s][Log:%d][Token:%s] 解析ExtraData失败: %v",
			vLog.TxHash.Hex(), vLog.Index, tokenIDStr, err)
		return fmt.Errorf("解析ExtraData失败 [Tx:%s Log:%d Token:%s]: %w",
			vLog.TxHash.Hex(), vLog.Index, tokenIDStr, err)
	}
	// 校验核心字段非空
	if setPriceEventData.Price == "" {
		log.Printf("[SetPriceEvent][Tx:%s][Log:%d][Token:%s] 价格为空，无效事件",
			vLog.TxHash.Hex(), vLog.Index, tokenIDStr)
		return errors.New(fmt.Sprintf("设置价格事件价格为空 [Tx:%s Log:%d Token:%s]",
			vLog.TxHash.Hex(), vLog.Index, tokenIDStr))
	}

	// ========== 步骤4：构建Event对象 - 简化赋值，增加注释 ==========
	newEvent := &model.Event{
		EventType:       h.eventName,
		TxHash:          vLog.TxHash.Hex(),
		BlockNumber:     vLog.BlockNumber,
		LogIndex:        vLog.Index,
		ContractAddress: contractAddr,
		TokenID:         tokenIDStr,
		FromAddress:     ownerAddress, // 设置价格的所有者地址
		ToAddress:       ownerAddress, // 无接收方，同所有者
		Amount:          "",
		ExtraData:       extraData,
		// CreatedAt由gorm autoCreateTime自动生成
		// UpdatedAt由gorm autoUpdateTime自动生成
	}

	log.Printf("[SetPriceEvent][Tx:%s][Log:%d][Token:%s] 解析完成的设置价格事件：%+v",
		vLog.TxHash.Hex(), vLog.Index, tokenIDStr, newEvent)

	// ========== 步骤5：查询-比对-新增/更新 - 优化逻辑结构，增加重试意识 ==========
	existEvent, err := h.eventService.GetByTxHashAndLogIndex(ctx, newEvent.TxHash, newEvent.LogIndex)
	if err != nil {
		// 仅当不是"记录不存在"错误时，才返回错误
		if !errors.Is(err, service2.ErrRecordNotFound) {
			log.Printf("[SetPriceEvent][Tx:%s][Log:%d][Token:%s] 查询事件记录失败: %v",
				vLog.TxHash.Hex(), vLog.Index, tokenIDStr, err)
			return fmt.Errorf("查询本地设置价格事件记录失败 [Tx:%s Log:%d Token:%s]: %w",
				vLog.TxHash.Hex(), vLog.Index, tokenIDStr, err)
		}

		// 情况1：无记录 → 新增Event + 插入操作记录
		log.Printf("[SetPriceEvent][Tx:%s][Log:%d][Token:%s] 本地无事件记录，执行新增",
			vLog.TxHash.Hex(), vLog.Index, tokenIDStr)

		// 新增Event记录（带上下文）
		if err := h.eventService.Create(ctx, newEvent); err != nil {
			log.Printf("[SetPriceEvent][Tx:%s][Log:%d][Token:%s] 新增事件记录失败: %v",
				vLog.TxHash.Hex(), vLog.Index, tokenIDStr, err)
			return fmt.Errorf("存储设置价格事件失败 [Tx:%s Log:%d Token:%s]: %w",
				vLog.TxHash.Hex(), vLog.Index, tokenIDStr, err)
		}

		// 构建操作记录（提取为变量，提升可读性）
		operateRecord := &model.NFTOperateRecord{
			ContractAddress: contractAddr,
			TokenID:         tokenIDStr,
			UserAddress:     ownerAddress,
			OwnerAddress:    ownerAddress,
			OperateType:     string(constant.OperateTypeSetPrice), // 使用常量，避免硬编码
			Amount:          setPriceEventData.Price,
			TxHash:          vLog.TxHash.Hex(),
			Status:          string(constant.OperateStatusSuccess), // 使用常量
			BlockNumber:     fmt.Sprintf("%d", vLog.BlockNumber),
			OperateAt:       time.Now(),
			Remark:          fmt.Sprintf("设置NFT价格为%s ETH", setPriceEventData.Price),
		}

		// 创建设置价格操作记录（容错+详细日志）
		if err := h.operateService.CreateOperateRecord(ctx, operateRecord); err != nil {
			log.Printf("[SetPriceEvent][Tx:%s][Log:%d][Token:%s] 创建设置价格操作记录失败: %v",
				vLog.TxHash.Hex(), vLog.Index, tokenIDStr, err)
			// 业务决策：操作记录创建失败不阻断主流程，但记录错误便于排查
			// 如需严格一致性，可取消注释返回错误
			// return fmt.Errorf("创建设置价格操作记录失败 [Tx:%s Log:%d Token:%s]: %w",
			// 	vLog.TxHash.Hex(), vLog.Index, tokenIDStr, err)
		} else {
			log.Printf("[SetPriceEvent][Tx:%s][Log:%d][Token:%s] 操作记录创建成功",
				vLog.TxHash.Hex(), vLog.Index, tokenIDStr)
		}
	} else {
		// 情况2：有记录 → 比对后更新
		if h.isEventChanged(existEvent, newEvent) {
			newEvent.ID = existEvent.ID // 继承主键实现覆盖更新
			log.Printf("[SetPriceEvent][Tx:%s][Log:%d][Token:%s] 事件记录有变更，执行更新",
				vLog.TxHash.Hex(), vLog.Index, tokenIDStr)

			if err := h.eventService.Update(ctx, newEvent); err != nil {
				log.Printf("[SetPriceEvent][Tx:%s][Log:%d][Token:%s] 更新事件记录失败: %v",
					vLog.TxHash.Hex(), vLog.Index, tokenIDStr, err)
				return fmt.Errorf("更新设置价格事件记录失败 [Tx:%s Log:%d Token:%s]: %w",
					vLog.TxHash.Hex(), vLog.Index, tokenIDStr, err)
			}
		} else {
			log.Printf("[SetPriceEvent][Tx:%s][Log:%d][Token:%s] 事件记录无变更，跳过更新",
				vLog.TxHash.Hex(), vLog.Index, tokenIDStr)
		}
	}

	// ========== 步骤6：更新NFT价格 - 增加日志和上下文 ==========
	log.Printf("[SetPriceEvent][Tx:%s][Log:%d][Token:%s] 开始更新NFT价格为%s ETH",
		vLog.TxHash.Hex(), vLog.Index, tokenIDStr, setPriceEventData.Price)

	if err := h.nftService.UpdatePrice(ctx, tokenIDStr, setPriceEventData.Price, setPriceEventData.SetAt); err != nil {
		log.Printf("[SetPriceEvent][Tx:%s][Log:%d][Token:%s] 更新NFT价格失败: %v",
			vLog.TxHash.Hex(), vLog.Index, tokenIDStr, err)
		return fmt.Errorf("更新NFT价格失败 [Tx:%s Log:%d Token:%s]: %w",
			vLog.TxHash.Hex(), vLog.Index, tokenIDStr, err)
	}

	log.Printf("[SetPriceEvent][Tx:%s][Log:%d][Token:%s] 设置价格事件处理完成，NFT价格已更新为%s ETH",
		vLog.TxHash.Hex(), vLog.Index, tokenIDStr, setPriceEventData.Price)
	return nil
}

// 字段比对方法（和其他处理器保持一致）
func (h *MarketSetPriceEventHandler) isEventChanged(old, new *model.Event) bool {
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

// 解析设置价格事件Topic（优化注释和错误提示）
func (h *MarketSetPriceEventHandler) parseTopics(topics []common.Hash) (string, *big.Int, error) {
	// 校验Topics长度
	if len(topics) < 3 {
		return "", nil, fmt.Errorf("Topics长度不足，期望至少3个，实际：%d", len(topics))
	}

	// 所有者地址
	ownerAddrBytes := topics[1].Bytes()[12:]
	ownerAddress := common.BytesToAddress(ownerAddrBytes).Hex()

	// tokenId
	tokenId := new(big.Int).SetBytes(topics[2].Bytes())
	if tokenId.Cmp(big.NewInt(0)) < 0 {
		return "", nil, errors.New("tokenId不能为负数")
	}

	// 修正6：错误提示从「解析的地址不合法」改为「解析的所有者地址不合法」（精准）
	if !common.IsHexAddress(ownerAddress) {
		return "", nil, fmt.Errorf("解析的所有者地址不合法：%s", ownerAddress)
	}

	return ownerAddress, tokenId, nil
}

// 解析设置价格事件Data（优化日志和错误提示）
func (h *MarketSetPriceEventHandler) parseData(abi abi.ABI, data []byte) (string, error) {
	eventData := make(map[string]interface{})

	if len(data) > 0 {
		if err := abi.UnpackIntoMap(eventData, h.eventName, data); err != nil {
			return "", fmt.Errorf("UnpackIntoMap失败: %w", err)
		}
	}

	// 修正7：日志带业务语义（设置价格事件）
	log.Printf("设置价格事件解析出的eventData: %+v", eventData)

	// 提取字段（优化错误提示，带业务语义）
	price, ok := eventData["price"].(*big.Int)
	if !ok {
		return "", errors.New("Price字段解析失败（设置价格事件）")
	}
	setAt, ok := eventData["setAt"].(*big.Int)
	if !ok {
		return "", errors.New("setAt字段解析失败（设置价格事件）")
	}

	// 序列化为ExtraData
	extraData, err := json.Marshal(map[string]string{
		"price": price.String(),
		"setAt": setAt.String(),
	})
	if err != nil {
		return "", fmt.Errorf("序列化设置价格事件ExtraData失败: %w", err)
	}

	return string(extraData), nil
}
