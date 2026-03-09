package block

import (
	"context"
	"fmt"
	"nft_backend/internal/app/service"
	blocklog "nft_backend/internal/blockchain/block/log"
	"nft_backend/internal/blockchain/block/status"
	"nft_backend/internal/blockchain/contract/basic"
	"nft_backend/internal/blockchain/contract/event/listener"
	"nft_backend/internal/logger"
	"nft_backend/internal/model"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"gorm.io/gorm"
)

// NFTCorrectService 校正服务
type NFTCorrectService struct {
	db                        *gorm.DB
	client                    *ethclient.Client
	parser                    *BlockParser
	nftService                service.NFTServicer
	ContractListeners         []*listener.ContractListener
	blockProcessStatusService *status.BlockProcessStatusService
	blockProcessLogService    *blocklog.BlockProcessLogService
}

// NewNFTCorrectService 创建校正服务
func NewNFTCorrectService(
	db *gorm.DB,
	client *ethclient.Client,
	parser *BlockParser,
	nftService service.NFTServicer,
	contractListeners []*listener.ContractListener,
	blockProcessStatusService *status.BlockProcessStatusService,
	blockProcessLogService *blocklog.BlockProcessLogService,
) *NFTCorrectService {
	return &NFTCorrectService{
		db:                        db,
		client:                    client,
		parser:                    parser,
		nftService:                nftService,
		ContractListeners:         contractListeners,
		blockProcessStatusService: blockProcessStatusService,
		blockProcessLogService:    blockProcessLogService,
	}
}

const TIME_OUT = 300
const MAX_BLOCK_NUM = 50

// Correct 执行单次校正（核心入口，重构后）
func (s *NFTCorrectService) Correct() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(TIME_OUT)*time.Second)
	defer cancel()

	logger.Sugar.Infof("===== 开始执行NFT数据校正 =====")
	//startTime := time.Now()
	//logger.Sugar.Infof("待校正合约个数: %d", len(s.parser.Contracts))

	// 每条连先遍历区块，再处理每个区块内的所有合约

	chainID, err := s.client.ChainID(ctx)
	if err != nil {
		return
	}

	if err := s.correctByBlock(ctx, chainID.Int64()); err != nil {
		logger.Sugar.Errorf("区块校正流程执行失败: %v", err)
	}

	//logger.Sugar.Infof("===== 校正完成 | 总耗时: %v =====", time.Since(startTime))
}

// 按区块维度执行校正（核心重构方法）
func (s *NFTCorrectService) correctByBlock(ctx context.Context, chainID int64) error {
	// 1. 获取全局区块区间
	lastBlock, err := s.GetLastBlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("获取上次校正高度失败: %w", err)
	}
	latestBlock, err := s.parser.GetLatestBlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("获取最新区块高度失败: %w", err)
	}

	// 无新区块需处理
	if lastBlock >= latestBlock {
		logger.Sugar.Infof("无新区块需校正（上次高度: %d, 最新高度: %d）", lastBlock, latestBlock)
		return nil
	}

	// 起始高度：首次校正用所有合约中最小的部署高度，否则用上次高度+1
	startBlock := lastBlock + 1

	// 限制单次最大遍历区块数
	endBlock := startBlock + uint64(MAX_BLOCK_NUM) - 1
	if endBlock > latestBlock {
		endBlock = latestBlock
	}
	logger.Sugar.Infof("待校正区块区间: %d - %d", startBlock, endBlock)

	// 2. 遍历区块（核心：一个区块只查询一次）
	var totalCorrected int
	for blockNum := startBlock; blockNum <= endBlock; blockNum++ {
		// 检查上下文超时
		select {
		case <-ctx.Done():
			return fmt.Errorf("处理区块%d超时: %w", blockNum, ctx.Err())
		default:
		}

		//logger.Sugar.Infof("开始处理区块%d", blockNum)

		// 修改后的代码
		contractVlogsMap, err := s.parser.ParseBlockWithVlogs(ctx, chainID, blockNum)
		if err != nil {
			logger.Sugar.Errorf("解析区块%d失败，跳过: %v", blockNum, err)
			continue
		}

		// 检查map
		if contractVlogsMap == nil || len(contractVlogsMap) == 0 {
			cid, err := s.client.ChainID(ctx)
			if err != nil {
				return err
			}
			logger.Sugar.Infof("%s区块%d没有合约vlogs，跳过", cid, blockNum)
			err0 := s.UpdateLastBlockNumber(ctx, blockNum)
			if err0 != nil {
				return err0
			}
			continue
		}

		// 遍历所有合约
		blockCorrected := 0
		for _, contract := range s.parser.Contracts {
			contractAddress := contract.Contract.Address
			var contractListener *listener.ContractListener
			for _, c := range s.ContractListeners {
				if strings.EqualFold(contractAddress, contract.Contract.Address) {
					contractListener = c
					break
				}
			}

			// 检查map中是否有该地址
			vlogs, exists := contractVlogsMap[contractAddress]
			if !exists {
				// 尝试大小写不敏感的匹配
				for mapAddr := range contractVlogsMap {
					if strings.EqualFold(mapAddr, contractAddress) {
						vlogs = contractVlogsMap[mapAddr]
						exists = true
						break
					}
				}

				if !exists {
					logger.Sugar.Infof("合约%s在区块%d没有vlogs，跳过", contractAddress, blockNum)
					err := s.UpdateLastBlockNumber(ctx, blockNum)
					if err != nil {
						return err
					}
					continue
				}
			}

			// 检查vlogs是否为空
			if vlogs == nil || len(vlogs) == 0 {
				logger.Sugar.Infof("合约%s在区块%d有零条vlogs", contractAddress, blockNum)
				// 根据业务决定是否继续处理空数组
			}

			// 处理数据

			err := s.correctContractVlogs(ctx, contract, contractListener, blockNum, vlogs)
			if err != nil {
				logger.Sugar.Errorf("合约%s在区块%d校正失败: %v", contractAddress, blockNum, err)
				continue
			}
		}

		// 即使无数据，也更新全局区块状态（确保断点续跑）
		if err := s.UpdateLastBlockNumber(ctx, blockNum); err != nil {
			logger.Sugar.Errorf("更新区块%d状态失败: %v", blockNum, err)
		} else {
			// logger.Sugar.Infof("区块%d处理完成 | 解析%d条原始数据 | 校正%d条 | 状态已更新", blockNum, len(contractVlogsMap), blockCorrected)
		}
		totalCorrected += blockCorrected
	}

	// logger.Sugar.Infof("区块校正完成 | 总处理区块数: %d | 总校正条数: %d", endBlock-startBlock+1, totalCorrected)
	return nil
}

// correctContractVlogs 处理单个合约在指定区块的NFT数据
func (s *NFTCorrectService) correctContractVlogs(ctx context.Context, contractConfig basic.ContractConfig, contractListener *listener.ContractListener, blockNum uint64, vlogs []*types.Log) error {
	logger.Sugar.Infof("开始矫正")
	if len(vlogs) == 0 {
		return fmt.Errorf("合约%s区块%d无vlog数据", contractConfig.Contract.Address, blockNum)
	}
	// 执行校正
	for _, vlog := range vlogs {
		contractListener.DispatchEvent(*vlog)
	}
	logger.Sugar.Infof("合约%s在区块%d校正完成 | vlog数据%d条 ",
		contractConfig.Contract.Address, blockNum, len(vlogs))
	return nil
}

// 以下方法保持不变，仅适配重构后的逻辑
// filterValidNFTData 过滤有效NFT数据（基于nft.NFT）
func (s *NFTCorrectService) filterValidNFTData(datas []model.NFT) []model.NFT {
	var valid []model.NFT

	for _, d := range datas {
		// 1. 校验TokenID有效性
		if d.TokenID == "" || d.TokenID == "0" {
			continue
		}
		// 2. 校验Owner地址有效性
		if d.OwnerID == "" || d.OwnerID == basic.ZeroAddress {
			continue
		}
		// 3. 校验合约地址
		if d.ContractAddress == "" || d.ContractAddress == basic.ZeroAddress {
			continue
		}
		valid = append(valid, d)
	}
	return valid
}

// GetLastBlockNumber 获取上一个查询的区块编号
func (s *NFTCorrectService) GetLastBlockNumber(ctx context.Context) (uint64, error) {
	chainID, err := s.client.ChainID(ctx)
	if err != nil {
		return 0, err
	}
	lastBlockNumber, err := s.blockProcessStatusService.GetLatestBlockNumber(chainID.Int64())
	if err != nil {
		return 0, err
	}
	logger.Sugar.Infof("上次区块处理编号: %d", lastBlockNumber)
	return uint64(lastBlockNumber), nil
}

// UpdateLastBlockNumber 更新最新区块编号
func (s *NFTCorrectService) UpdateLastBlockNumber(ctx context.Context, block uint64) error {
	chainID, err0 := s.client.ChainID(ctx)
	if err0 != nil {
		return err0
	}
	err := s.blockProcessStatusService.UpdateLatestBlock(chainID.Int64(), int64(block))
	if err != nil {
		return fmt.Errorf("更新区块%d状态失败: %w", block, err)
	}
	logger.Sugar.Infof("区块%d状态已更新到最新", block)
	return nil
}
