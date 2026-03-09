package listener

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"nft_backend/internal/blockchain/contract/basic"
	"nft_backend/internal/blockchain/contract/event/handler"
	"nft_backend/internal/logger"

	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Contract 合约配置结构体
type Contract struct {
	Address     common.Address
	AbiPath     string
	HandlerName string
	Name        string `yaml:"name"`         // 合约名称
	DeployBlock uint64 `yaml:"deploy_block"` // 部署区块高度
	RPCUrl      string `yaml:"rpc_url"`      // RPC地址
}

// ContractListener 合约监听器
type ContractListener struct {
	client          *ethclient.Client
	contractABI     abi.ABI
	ContractAddress common.Address
	contractHandler handler.ContractEventHandler
	stopChan        chan struct{}
	lastBlock       uint64 // 最后处理的区块号
	pollInterval    time.Duration
	errorCount      int
	maxErrors       int
	eventTopicMap   basic.EventTopicMap // 公共事件哈希映射
}

// NewContractListener 创建新的合约监听器
func NewContractListener(client *ethclient.Client, contract Contract) (*ContractListener, error) {
	contractABI, err := basic.LoadABI(contract.AbiPath)
	if err != nil {
		return nil, err
	}

	// 获取对应的事件处理器
	eventHandler := handler.GetHandler(contract.HandlerName)
	if eventHandler == nil {
		logger.Sugar.Panic("合约" + contract.Name + "找不到处理器")
	}
	// 初始化事件哈希映射
	eventTopicMap := basic.InitEventTopicMap(contractABI)

	contractListener := &ContractListener{
		client:          client,
		contractABI:     contractABI,
		ContractAddress: contract.Address,
		contractHandler: eventHandler,
		stopChan:        make(chan struct{}),
		pollInterval:    2 * time.Second, // 轮询间隔
		lastBlock:       0,
		errorCount:      0,
		maxErrors:       10, // 最大错误次数
		eventTopicMap:   eventTopicMap,
	}
	return contractListener, nil
}

// Start 启动合约监听
func (cl *ContractListener) Start() error {
	currentBlock, err := cl.getCurrentBlock()
	if err != nil {
		return fmt.Errorf("获取起始区块失败: %v", err)
	}

	cl.lastBlock = currentBlock

	// 启动轮询
	go cl.pollingLoop()

	return nil
}

// Stop 停止监听
func (cl *ContractListener) Stop() {
	close(cl.stopChan)
}

// DispatchEvent 分发事件（核心处理逻辑）
func (cl *ContractListener) DispatchEvent(vLog types.Log) {
	// 获取区块时间
	blockTime := cl.getBlockTime(vLog.BlockNumber)

	// 获取事件类型（统一逻辑）
	eventType := cl.eventTopicMap.GetEventType(vLog.Topics[0])
	if eventType == "Unknown" {
		fmt.Printf("未知事件类型: %s (区块: #%d)\n", vLog.Topics[0], vLog.BlockNumber)
		return
	}
	log.Printf("处理事件: %s, 区块: #%d\n", eventType, vLog.BlockNumber)

	// 检查是否支持该事件
	if cl.contractHandler != nil {
		supported := false
		supportedEvents := cl.contractHandler.SupportedEvents()
		for _, se := range supportedEvents {
			if se == eventType {
				supported = true
				break
			}
		}
		if !supported {
			fmt.Printf("不支持的事件类型: %s (区块: #%d)\n", eventType, vLog.BlockNumber)
			return
		}

		// 处理事件（可传递解析后的NFT数据）
		if err := cl.contractHandler.HandleEvent(cl.contractABI, eventType, vLog, blockTime); err != nil {
			fmt.Printf("事件处理失败: %v\n", err)
		}
	} else {
		log.Println("无法处理事件, 事件处理器为nil")
	}
}

// pollingLoop 轮询事件
func (cl *ContractListener) pollingLoop() {
	ticker := time.NewTicker(cl.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-cl.stopChan:
			return
		case <-ticker.C:
			if err := cl.pollEvents(); err != nil {
				cl.errorCount++
				if cl.errorCount >= cl.maxErrors {
					fmt.Println("超出最大错误次数，停止监听")
					return
				}
				time.Sleep(time.Duration(cl.errorCount) * time.Second) // 错误时增加间隔
			} else {
				cl.errorCount = 0 // 成功时重置错误计数
			}
		}
	}
}

// pollEvents 查询事件
func (cl *ContractListener) pollEvents() error {
	currentBlock, err := cl.getCurrentBlock()
	if err != nil {
		return fmt.Errorf("获取当前区块失败: %v", err)
	}

	if currentBlock <= cl.lastBlock {
		return nil
	}

	fromBlock := cl.lastBlock + 1
	toBlock := currentBlock

	// 查询事件
	_, err = cl.QueryBlockRange(fromBlock, toBlock)
	if err != nil {
		return fmt.Errorf("查询区块范围失败: %v", err)
	}

	cl.lastBlock = toBlock
	return nil
}

// QueryBlockRange 查询区块范围
func (cl *ContractListener) QueryBlockRange(fromBlock, toBlock uint64) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(fromBlock)),
		ToBlock:   big.NewInt(int64(toBlock)),
		Addresses: []common.Address{cl.ContractAddress},
	}

	logs, err := cl.client.FilterLogs(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("过滤日志失败: %v", err)
	}

	if len(logs) == 0 {
		return 0, nil
	}

	// 处理事件
	for _, vLog := range logs {
		cl.DispatchEvent(vLog)
	}

	return len(logs), nil
}

// getCurrentBlock 获取当前区块号
func (cl *ContractListener) getCurrentBlock() (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	blockNumber, err := cl.client.BlockNumber(ctx)
	if err != nil {
		return 0, fmt.Errorf("获取区块号失败: %v", err)
	}

	return blockNumber, nil
}

// getBlockTime 获取区块时间
func (cl *ContractListener) getBlockTime(blockNumber uint64) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	header, err := cl.client.HeaderByNumber(ctx, big.NewInt(int64(blockNumber)))
	if err != nil {
		return "时间获取失败"
	}

	return time.Unix(int64(header.Time), 0).Format("2006-01-02 15:04:05")
}

// handleUnknownEvent 处理未知事件
func (cl *ContractListener) handleUnknownEvent(vLog types.Log, eventType, blockTime string) {
	fmt.Printf("未知事件类型: %s (区块: #%d)\n", eventType, vLog.BlockNumber)
}
