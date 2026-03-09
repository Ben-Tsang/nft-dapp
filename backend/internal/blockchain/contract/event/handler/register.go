package handler

import (
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/types"
)

var (
	handlersMu sync.RWMutex
	handlers   = make(map[string]func() ContractEventHandler)
)

type ContractEventHandler interface {
	SupportedEvents() []string
	HandleEvent(abi abi.ABI, eventName string, vLog types.Log, blockTime string) error
}

type EventHandler interface {
	Handle(abi abi.ABI, vLog types.Log, blockTime string) error
}

// Register 注册处理器工厂函数
func Register(name string, factory func() ContractEventHandler) {
	handlersMu.Lock()
	defer handlersMu.Unlock()

	if _, exists := handlers[name]; exists {
		panic("handler already registered: " + name)
	}
	handlers[name] = factory
}

// GetHandler 获取处理器实例, 在获取时才创建实例
func GetHandler(name string) ContractEventHandler {
	handlersMu.RLock()
	factory, ok := handlers[name]
	handlersMu.RUnlock()

	if !ok {
		return nil
	}
	return factory()
}

// GetHandlerNames 获取所有已注册的处理器名称
func GetHandlerNames() []string {
	handlersMu.RLock()
	defer handlersMu.RUnlock()

	names := make([]string, 0, len(handlers))
	for name := range handlers {
		names = append(names, name)
	}
	return names
}
