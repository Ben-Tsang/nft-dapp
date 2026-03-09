package block

import (
	"context"
	"fmt"
	"log"
	"math/big"
	blocklog "nft_backend/internal/blockchain/block/log"
	"nft_backend/internal/blockchain/contract/basic"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"gorm.io/gorm"
)

// ========== 核心修改2：BlockParser维护多合约ABI映射 ==========
type BlockParser struct {
	db         *gorm.DB
	client     *ethclient.Client
	Contracts  []basic.ContractConfig // 多合约配置（含各自ABI）
	logService *blocklog.BlockProcessLogService
}

// ========== 核心修改3：NewBlockParser支持多合约ABI加载 ==========
func NewBlockParser(
	db *gorm.DB,
	client *ethclient.Client,
	contracts []basic.ContractConfig,
	logService *blocklog.BlockProcessLogService,
) (*BlockParser, error) {
	if len(contracts) == 0 {
		return nil, fmt.Errorf("未配置任何合约")
	}

	// 1. 遍历加载每个合约的ABI并初始化事件映射
	var loadedContracts []basic.ContractConfig
	for i, cfg := range contracts {
		// 校验必填字段
		if cfg.Contract.Address == "" {
			return nil, fmt.Errorf("第%d个合约地址为空", i+1)
		}
		if cfg.Contract.AbiPath == "" {
			return nil, fmt.Errorf("合约%s的ABI文件路径为空", cfg.Contract.Address)
		}

		// 2. 加载ABI（复用listener包的LoadABI）
		contractABI, err := basic.LoadABI(cfg.Contract.AbiPath)
		if err != nil {
			return nil, fmt.Errorf("加载合约%s的ABI失败: %w", cfg.Contract.Address, err)
		}

		// 3. 初始化事件哈希映射
		eventTopicMap := basic.InitEventTopicMap(contractABI)

		// 4. 填充ABI和事件映射
		cfg.EventTopicMap = eventTopicMap
		cfg.ABI = contractABI

		loadedContracts = append(loadedContracts, cfg)

	}

	return &BlockParser{
		db:         db,
		client:     client,
		Contracts:  loadedContracts,
		logService: logService,
	}, nil
}

// ========== 核心修改4：IsMyContract保持不变（兼容多合约） ==========
func (p *BlockParser) IsMyContract(addr common.Address) bool {

	for _, contractConfig := range p.Contracts {
		if common.HexToAddress(contractConfig.Contract.Address) == addr {
			return true
		}
	}
	return false
}

//	从每个区块中获取合约相关的vlog
//
// 返回值是一个map，key是合约地址，value是该合约在该区块下的vlog数组
func (p *BlockParser) ParseBlockWithVlogs(ctx context.Context, chainID int64, blockNum uint64) (map[string][]*types.Log, error) {
	logEntity := &blocklog.BlockProcessLog{
		ChainID:     chainID,
		BlockNumber: int64(blockNum),
		ProcessTime: time.Now(),
	}

	// 1. 获取区块数据
	block, err := p.client.BlockByNumber(ctx, big.NewInt(int64(blockNum)))
	if err != nil {
		logEntity.Status = blocklog.ProcessStatusFailed
		logEntity.ErrorMsg = fmt.Sprintf("获取区块数据失败: %v", err)
		_ = p.logService.CreateLog(logEntity)
		return nil, fmt.Errorf("获取区块%d失败: %w", blockNum, err)
	}

	if block == nil || len(block.Transactions()) == 0 {
		logEntity.Status = blocklog.ProcessStatusNoNeed
		logEntity.ErrorMsg = "区块无交易数据"
		_ = p.logService.CreateLog(logEntity)
		return nil, nil
	}

	// 初始化结果map
	contractVlogs := make(map[string][]*types.Log)

	// 提前准备合约地址集合，用于快速查找
	contractSet := make(map[string]bool)
	for _, cfg := range p.Contracts {
		contractSet[strings.ToLower(cfg.Contract.Address)] = true
	}

	// 2. 遍历区块内交易
	for _, tx := range block.Transactions() {
		if tx.To() == nil {
			continue
		}

		// 检查是否是目标合约
		contractAddr := strings.ToLower(tx.To().Hex())
		if !contractSet[contractAddr] {
			continue
		}

		// 获取交易收据
		receipt, err := p.client.TransactionReceipt(ctx, tx.Hash())
		if err != nil {
			log.Printf("获取交易%s收据失败: %v", tx.Hash().Hex(), err)
			continue
		}
		if receipt.Status != types.ReceiptStatusSuccessful {
			continue
		}

		// 3. 处理日志
		for _, vlog := range receipt.Logs {
			// 校验日志的合约地址
			vlogAddr := strings.ToLower(vlog.Address.Hex())
			if !contractSet[vlogAddr] {
				continue
			}

			// 直接添加到对应合约的日志列表
			// 注意：这里使用懒初始化，不需要预先检查 nil
			contractVlogs[vlogAddr] = append(contractVlogs[vlogAddr], vlog)
		}
	}

	// 4. 记录处理结果
	totalLogs := 0
	for _, logs := range contractVlogs {
		totalLogs += len(logs)
	}

	if len(contractVlogs) == 0 {
		logEntity.Status = blocklog.ProcessStatusNoNeed
		logEntity.ErrorMsg = "区块内无目标合约交易数据"
	} else {
		logEntity.Status = blocklog.ProcessStatusSuccess
		logEntity.ErrorMsg = fmt.Sprintf("解析出%d个合约共%d条日志", len(contractVlogs), totalLogs)
	}
	_ = p.logService.CreateLog(logEntity)

	return contractVlogs, nil
}

// GetLatestBlockNumber 保持不变
func (p *BlockParser) GetLatestBlockNumber(ctx context.Context) (uint64, error) {
	num, err := p.client.BlockNumber(ctx)
	if err != nil {
		return 0, fmt.Errorf("获取最新区块高度失败: %w", err)
	}
	return num, nil
}
