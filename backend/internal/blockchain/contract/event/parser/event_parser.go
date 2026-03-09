package parser

import (
	"errors"
	"fmt"
	"math/big"
	"nft_backend/internal/blockchain/contract/basic"
	"nft_backend/internal/blockchain/contract/event/type"
	"nft_backend/internal/model"
	"strconv"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// TransferEvent Transfer事件结构体
type TransferEvent struct {
	From    common.Address
	To      common.Address
	TokenID *big.Int
}

// MintedEvent Minted事件结构体
type MintedEvent struct {
	To      common.Address `json:"to"`
	TokenId *big.Int       `json:"tokenId"`
	Uri     string         `json:"uri"`
}

// 统一解析types.Log为NFT结构体（核心统一逻辑）
func ParseLogToNFT(vLog *types.Log, contractABI abi.ABI, eventType string) (*model.NFT, error) {
	if vLog == nil {
		return nil, errors.New("日志为空")
	}

	var nftData *model.NFT
	switch eventType {
	case "Transfer":
		transferEvent, err := parseTransferEvent(vLog, contractABI)
		if err != nil {
			return nil, fmt.Errorf("解析Transfer事件失败: %w", err)
		}
		// 直接构建nft.NFT，无需中间NFTData
		nftData = &model.NFT{
			NftID:           transferEvent.TokenID.String(),
			OwnerID:         transferEvent.To.Hex(),
			ContractAddress: vLog.Address.Hex(),
			BlockNumber:     strconv.Itoa(int(vLog.BlockNumber)),
			Price:           "0",
			IsListed:        false,
		}
	case "NFTMinted":
		mintedEvent, err := parseMintedEvent(vLog, contractABI)
		if err != nil {
			return nil, fmt.Errorf("解析NFTMinted事件失败: %w", err)
		}
		nftData = &model.NFT{
			NftID:           mintedEvent.TokenId.String(),
			OwnerID:         mintedEvent.To.Hex(),
			NftURI:          mintedEvent.Uri,
			ContractAddress: vLog.Address.Hex(),
			BlockNumber:     strconv.Itoa(int(vLog.BlockNumber)),
			Price:           "0",
			IsListed:        false,
		}
	default:
		return nil, fmt.Errorf("不支持的事件类型: %s", eventType)
	}

	// 基础校验
	if nftData.NftID == "" || nftData.NftID == "0" {
		return nil, errors.New("无效的TokenID")
	}
	if nftData.OwnerID == "" || nftData.OwnerID == basic.ZeroAddress {
		return nil, errors.New("无效的所有者地址")
	}

	return nftData, nil
}

// parseTransferEvent 解析Transfer事件（统一逻辑）
func parseTransferEvent(vLog *types.Log, contractABI abi.ABI) (*TransferEvent, error) {
	var transferEvent TransferEvent
	// 兼容ABI解析和手动解析
	if len(vLog.Topics) >= 4 {
		// 手动解析（原correct包逻辑）
		transferEvent.From = common.HexToAddress(vLog.Topics[1].Hex())
		transferEvent.To = common.HexToAddress(vLog.Topics[2].Hex())
		transferEvent.TokenID = new(big.Int).SetBytes(vLog.Topics[3].Bytes())
	} else {
		// ABI解析（原listener包逻辑）
		if err := contractABI.UnpackIntoInterface(&transferEvent, _type.Transfer.String(), vLog.Data); err != nil {
			return nil, err
		}
	}
	return &transferEvent, nil
}

// parseMintedEvent 解析NFTMinted事件（统一逻辑）
func parseMintedEvent(vLog *types.Log, contractABI abi.ABI) (*MintedEvent, error) {
	var mintedEvent MintedEvent
	if err := contractABI.UnpackIntoInterface(&mintedEvent, _type.NFTMinted.String(), vLog.Data); err != nil {
		return nil, err
	}
	return &mintedEvent, nil
}
