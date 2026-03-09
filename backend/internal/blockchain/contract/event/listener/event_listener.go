package listener

import (
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
)

// EventListener 事件监听器，管理多个合约监听
type EventListener struct {
	client            *ethclient.Client
	stopChan          chan struct{}
	ContractListeners []*ContractListener
}

// NewEventListener 创建事件监听器，管理多个合约
func NewEventListener(client *ethclient.Client, contracts []Contract) (*EventListener, error) {
	var contractListeners []*ContractListener

	// 为每个合约创建一个 ContractListener
	for _, contract := range contracts {
		contractListener, err := NewContractListener(client, contract)
		if err != nil {
			return nil, err
		}
		contractListeners = append(contractListeners, contractListener)
	}

	return &EventListener{
		client:            client,
		ContractListeners: contractListeners,
		stopChan:          make(chan struct{}),
	}, nil
}

// Start 启动所有合约的监听
func (el *EventListener) Start() error {
	for _, listener := range el.ContractListeners {
		if err := listener.Start(); err != nil {
			return fmt.Errorf("启动合约监听失败: %v", err)
		}
	}
	return nil
}

// Stop 停止所有合约的监听
func (el *EventListener) Stop() {
	for _, listener := range el.ContractListeners {
		listener.Stop()
	}
	close(el.stopChan)
}

// 新增：Close方法，优雅停止所有监听协程
func (el *EventListener) Close() {
	close(el.stopChan) // 发送关闭信号
	//el.wg.Wait()       // 等待所有监听协程结束
}
