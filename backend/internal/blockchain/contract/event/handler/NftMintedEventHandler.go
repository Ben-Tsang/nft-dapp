package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"nft_backend/internal/app/service"
	"nft_backend/internal/blockchain/contract/basic"
	"nft_backend/internal/blockchain/contract/event/constant"
	_type "nft_backend/internal/blockchain/contract/event/type"
	"nft_backend/internal/di"
	"nft_backend/internal/model"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"gorm.io/gorm"
)

type NftMintedEventHandler struct {
	eventName      string
	eventService   *service.EventService
	nftService     *service.NftService
	operateService *service.OperateService
}

func NewNftMintedEventHandler(operateService *service.OperateService) *NftMintedEventHandler {
	// 依赖注入（补充错误日志）
	eventService, err := di.Resolve[*service.EventService]()
	if err != nil {
		log.Printf("解析EventService依赖失败: %v", err)
	}
	nftService, err := di.Resolve[*service.NftService]()
	if err != nil {
		log.Printf("解析NFTService依赖失败: %v", err)
	}

	return &NftMintedEventHandler{
		eventName:      _type.NFTMinted.String(),
		eventService:   eventService,
		nftService:     nftService,
		operateService: operateService,
	}
}

func (h *NftMintedEventHandler) Handle(abi abi.ABI, vLog types.Log, blockTime string) error {
	// 1. 创建带超时的ctx（使用常量定义的超时时间）
	ctx, cancel := context.WithTimeout(context.Background(), constant.EventHandleTimeout)
	defer cancel() // 函数结束释放资源

	// 2. 补充上下文日志（TxHash+LogIndex，便于定位）
	txHash := vLog.TxHash.Hex()
	logIndex := vLog.Index
	tokenIdStr := "" // 提前声明，便于后续日志使用
	log.Printf("[MintEvent][Tx:%s][Log:%d] 开始处理铸造事件，区块时间：%s", txHash, logIndex, blockTime)

	// ========== 步骤1：解析Topic（索引字段） ==========
	toAddress, tokenId, err := h.parseMintedTopics(vLog.Topics)
	if err != nil {
		return fmt.Errorf("[MintEvent][Tx:%s][Log:%d] 解析Topic失败: %w", txHash, logIndex, err)
	}
	tokenIdStr = tokenId.String()

	// ========== 步骤2：解析Data（非索引字段） ==========
	extraData, err := h.parseMintedData(abi, vLog.Data)
	if err != nil {
		return fmt.Errorf("[MintEvent][Tx:%s][Log:%d][Token:%s] 解析Data失败: %w",
			txHash, logIndex, tokenIdStr, err)
	}

	// ========== 步骤3：构建NFTEvent对象 ==========
	newEvent := &model.Event{
		EventType:       h.eventName,
		TxHash:          txHash,
		BlockNumber:     vLog.BlockNumber,
		LogIndex:        logIndex,
		ContractAddress: vLog.Address.Hex(),
		TokenID:         tokenIdStr,
		FromAddress:     basic.ZeroAddress,
		ToAddress:       toAddress,
		Amount:          "",
		ExtraData:       extraData,
		CreatedAt:       time.Now(),
	}

	log.Printf("[MintEvent][Tx:%s][Log:%d][Token:%s] 解析完成的铸造事件：%+v",
		txHash, logIndex, tokenIdStr, newEvent)

	// ========== 步骤4：存储事件记录（传递ctx，强化校验） ==========
	// 空指针校验：eventService不能为空
	if h.eventService == nil {
		return fmt.Errorf("[MintEvent][Tx:%s][Log:%d][Token:%s] eventService依赖未初始化",
			txHash, logIndex, tokenIdStr)
	}

	// 先查询事件是否已存在（避免重复创建）
	_, err = h.eventService.GetByTxHashAndLogIndex(ctx, newEvent.TxHash, newEvent.LogIndex)
	isNewEvent := false
	if err != nil {
		if errors.Is(err, service.ErrRecordNotFound) {
			// 新事件：执行创建
			if err := h.eventService.Create(ctx, newEvent); err != nil {
				return fmt.Errorf("[MintEvent][Tx:%s][Log:%d][Token:%s] 存储事件失败: %w",
					txHash, logIndex, tokenIdStr, err)
			}
			isNewEvent = true
			log.Printf("[MintEvent][Tx:%s][Log:%d][Token:%s] 新铸造事件，已创建Event记录",
				txHash, logIndex, tokenIdStr)
		} else {
			// 其他查询错误
			return fmt.Errorf("[MintEvent][Tx:%s][Log:%d][Token:%s] 查询事件记录失败: %w",
				txHash, logIndex, tokenIdStr, err)
		}
	} else {
		log.Printf("[MintEvent][Tx:%s][Log:%d][Token:%s] 已有铸造事件记录，跳过重复创建",
			txHash, logIndex, tokenIdStr)
	}

	// ========== 步骤5：解析extraData为MintEventData ==========
	var mintedEventData MintEventData
	err = json.Unmarshal([]byte(extraData), &mintedEventData)
	if err != nil {
		log.Printf("[MintEvent][Tx:%s][Log:%d][Token:%s] 解析extraData失败: %v",
			txHash, logIndex, tokenIdStr, err)
		return fmt.Errorf("[MintEvent][Tx:%s][Log:%d][Token:%s] 解析extraData失败: %w",
			txHash, logIndex, tokenIdStr, err)
	}

	// ========== 核心：NFT记录 查→创/更 ==========
	contractAddress := vLog.Address.Hex()
	blockNum := vLog.BlockNumber

	// 空指针校验：nftService不能为空
	if h.nftService == nil {
		return fmt.Errorf("[MintEvent][Tx:%s][Log:%d][Token:%s] nftService依赖未初始化",
			txHash, logIndex, tokenIdStr)
	}

	// 1. 查询NFT
	_, err = h.nftService.GetNFT(ctx, contractAddress, tokenIdStr)
	if err != nil {
		// 查不到，执行创建
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err = h.nftService.CreateNft(
				ctx,
				tokenIdStr,
				toAddress,
				mintedEventData.Name,
				mintedEventData.Description,
				mintedEventData.TokenURI,
				contractAddress,
				blockNum,
			); err != nil {
				return fmt.Errorf("[MintEvent][Tx:%s][Log:%d][Token:%s] 创建NFT记录失败: %w",
					txHash, logIndex, tokenIdStr, err)
			}
			log.Printf("[MintEvent][Tx:%s][Log:%d][Token:%s] NFT记录创建成功",
				txHash, logIndex, tokenIdStr)

			// ========== 核心改造：新事件必须创建铸造操作记录（失败阻断） ==========
			if isNewEvent {
				// 空指针校验：operateService不能为空
				if h.operateService == nil {
					return fmt.Errorf("[MintEvent][Tx:%s][Log:%d][Token:%s] OperateService依赖未初始化，无法创建操作记录",
						txHash, logIndex, tokenIdStr)
				}

				// 构建铸造操作记录（使用常量）
				operateRecord := &model.NFTOperateRecord{
					ContractAddress: contractAddress,
					TokenID:         tokenIdStr,
					UserAddress:     toAddress,
					OwnerAddress:    toAddress,
					OperateType:     string(constant.OperateTypeMint),      // 铸造专属类型
					Amount:          "",                                    // 铸造无金额
					TxHash:          txHash,                                // 交易哈希
					Status:          string(constant.OperateStatusSuccess), // 使用常量
					BlockNumber:     fmt.Sprintf("%d", vLog.BlockNumber),
					OperateAt:       time.Now(),
					Remark:          "NFT铸造事件自动生成",
				}

				// 核心变更：创建失败直接返回错误（阻断主流程）
				if err := h.operateService.CreateOperateRecord(ctx, operateRecord); err != nil {
					return fmt.Errorf("[MintEvent][Tx:%s][Log:%d][Token:%s] 创建铸造操作记录失败（新事件必须记录）: %w",
						txHash, logIndex, tokenIdStr, err)
				}
				log.Printf("[MintEvent][Tx:%s][Log:%d][Token:%s] 新事件，铸造操作记录创建成功",
					txHash, logIndex, tokenIdStr)
			} else {
				log.Printf("[MintEvent][Tx:%s][Log:%d][Token:%s] 已有事件，跳过重复创建操作记录",
					txHash, logIndex, tokenIdStr)
			}
			return nil
		}
		// 其他查询异常
		return fmt.Errorf("[MintEvent][Tx:%s][Log:%d][Token:%s] 查询NFT失败: %w",
			txHash, logIndex, tokenIdStr, err)
	}

	// 2. 查到，执行更新
	if err = h.nftService.UpdateNft(
		ctx,
		contractAddress,
		tokenIdStr,
		toAddress,
		mintedEventData.Name,
		mintedEventData.Description,
		mintedEventData.TokenURI,
		blockNum,
	); err != nil {
		return fmt.Errorf("[MintEvent][Tx:%s][Log:%d][Token:%s] 更新NFT记录失败: %w",
			txHash, logIndex, tokenIdStr, err)
	}
	log.Printf("[MintEvent][Tx:%s][Log:%d][Token:%s] NFT记录更新成功",
		txHash, logIndex, tokenIdStr)
	log.Printf("[MintEvent][Tx:%s][Log:%d][Token:%s] 铸造事件处理完成",
		txHash, logIndex, tokenIdStr)
	return nil
}

// parseMintedTopics 解析Minted事件的Topic（保留原有逻辑，补充日志上下文）
func (h *NftMintedEventHandler) parseMintedTopics(topics []common.Hash) (string, *big.Int, error) {
	if len(topics) < 3 {
		return "", nil, fmt.Errorf("Topics长度不足，期望至少3个，实际：%d", len(topics))
	}

	// 解析toAddress
	toAddrBytes := topics[1].Bytes()[12:]
	toAddress := common.BytesToAddress(toAddrBytes).Hex()
	// 解析tokenId
	tokenId := new(big.Int).SetBytes(topics[2].Bytes())

	// 合法性校验
	if tokenId.Cmp(big.NewInt(0)) < 0 {
		return "", nil, errors.New("tokenId不能为负数")
	}
	if !common.IsHexAddress(toAddress) {
		return "", nil, fmt.Errorf("解析的地址不合法：%s", toAddress)
	}

	return toAddress, tokenId, nil
}

// parseMintedData 解析Minted事件的Data（保留原有逻辑）
func (h *NftMintedEventHandler) parseMintedData(abi abi.ABI, data []byte) (string, error) {
	eventData := make(map[string]interface{})
	if len(data) > 0 {
		if err := abi.UnpackIntoMap(eventData, h.eventName, data); err != nil {
			return "", fmt.Errorf("UnpackIntoMap失败: %w", err)
		}
	}

	// 兼容参数缺失
	name, _ := eventData["name"].(string)
	description, _ := eventData["description"].(string)
	uri, _ := eventData["tokenURI"].(string)

	// 序列化为JSON字符串
	extraData, err := json.Marshal(MintEventData{
		Name:        name,
		Description: description,
		TokenURI:    uri,
	})
	if err != nil {
		return "", fmt.Errorf("序列化ExtraData失败: %w", err)
	}

	return string(extraData), nil
}
