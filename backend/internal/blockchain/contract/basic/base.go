package basic

import (
	"fmt"
	"nft_backend/internal/logger"

	"github.com/ethereum/go-ethereum/ethclient"
)

// 全局常量：以太坊零地址
const ZeroAddress = "0x0000000000000000000000000000000000000000"

// 区块链连接相关
func Connect(url string) (*ethclient.Client, error) {
	// 连接到以太坊节点（在这里使用本地节点：http://127.0.0.1:8545）
	client, err := ethclient.Dial(url) // 这里改成你的以太坊节点地址，或者使用Infura等服务
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the Ethereum client: %v", err)
	}

	// 成功连接，返回 client
	logger.L.Info("成功连接到以太坊节点...")
	return client, nil
}
