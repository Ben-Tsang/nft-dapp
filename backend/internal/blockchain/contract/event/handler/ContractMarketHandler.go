package handler

import (
	"log"
	"nft_backend/internal/app/service"
	"nft_backend/internal/blockchain/contract/event/type"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/types"
)

// ========== 1. 先明确注册函数的签名（你原有定义） ==========
// 注册函数要求传入：无参、返回ContractEventHandler的函数
type HandlerFactory func() ContractEventHandler

// 全局注册表（你的原有逻辑）
var handlerRegistry = make(map[string]HandlerFactory)

type ContractMarketHandler struct {
	ItemListedHandler        *MarketItemListedEventHandler
	ItemUnlistedEventHandler *MarketItemUnlistedEventHandler
	SetPriceEventHandler     *MarketSetPriceEventHandler
	BuyEventHandler          *MarketBuyEventHandler
}

// NewMarketHandler：带参构造函数（接收你要传入的operateService）
func NewMarketHandler(operateService *service.OperateService) ContractEventHandler {
	// 给子处理器传入operateService（子处理器构造函数同步改造成带参）
	itemListedHandler := NewMarketItemListedEventHandler(operateService)
	itemUnlistedHandler := NewMarketItemUnlistedEventHandler(operateService)
	setPriceHandler := NewMarketSetPriceEventHandler(operateService)
	buyHandler := NewMarketBuyEventHandler(operateService)

	return &ContractMarketHandler{
		ItemListedHandler:        itemListedHandler,
		ItemUnlistedEventHandler: itemUnlistedHandler,
		SetPriceEventHandler:     setPriceHandler,
		BuyEventHandler:          buyHandler,
	}
}

// ========== 3. 核心：带参的初始化函数 + 闭包适配签名 ==========
// initContractMarketHandler：你要直接传参的初始化函数
// 入参：你想传入的operateService（无需DI，直接传）
func InitContractMarketHandler(operateService *service.OperateService) {
	log.Printf("注册 Market 事件处理器（直接传参，不依赖DI）")

	// 关键：用闭包捕获operateService，返回注册函数要求的「无参HandlerFactory」
	factory := func() ContractEventHandler {
		// 闭包内可以访问外部的operateService，调用带参的NewMarketHandler
		return NewMarketHandler(operateService)
	}

	// 注册：传入的factory是无参函数，匹配Register的签名要求
	Register("market", factory)
}

// ========== 4. 原有方法逻辑（仅补充Buy事件） ==========
func (m ContractMarketHandler) SupportedEvents() []string {
	return []string{
		_type.ItemListed.String(),
		_type.ItemUnlisted.String(),
		_type.SetPrice.String(),
		_type.Buy.String(), // 补充Buy事件，避免漏处理
	}
}

func (m ContractMarketHandler) HandleEvent(abi abi.ABI, eventName string, vLog types.Log, blockTime string) error {
	switch eventName {
	case _type.ItemListed.String():
		return m.ItemListedHandler.Handle(abi, vLog, blockTime)
	case _type.ItemUnlisted.String():
		return m.ItemUnlistedEventHandler.Handle(abi, vLog, blockTime)
	case _type.SetPrice.String():
		return m.SetPriceEventHandler.Handle(abi, vLog, blockTime)
	case _type.Buy.String():
		return m.BuyEventHandler.Handle(abi, vLog, blockTime)
	default:
		log.Printf("未知事件：%s", eventName)
		return nil
	}
}

// 其他子处理器（Unlisted/SetPrice/Buy）按相同逻辑改造：
// func NewMarketItemUnlistedEventHandler(operateService *service.OperateService) *MarketItemUnlistedEventHandler
// func NewMarketSetPriceEventHandler(operateService *service.OperateService) *MarketSetPriceEventHandler
// func NewMarketBuyEventHandler(operateService *service.OperateService) *MarketBuyEventHandler

// ========== 辅助：如果其他依赖也不想用DI，可手动初始化 ==========
func initEventService() *service.EventService {
	// 手动创建EventService实例（比如直接new，或传入数据库连接等）
	return &service.EventService{}
}

func initNftService() *service.NftService {
	// 手动创建NftService实例
	return &service.NftService{}
}
