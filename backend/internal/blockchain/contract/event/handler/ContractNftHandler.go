package handler

import (
	"errors"
	"log"
	"nft_backend/internal/app/service"
	"nft_backend/internal/blockchain/contract/event/type"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/types"
)

// ========== 2. NFT事件处理器核心定义 ==========
type ContractNftHandler struct {
	mintedHandler *NftMintedEventHandler
}

// 改造NewNftContractHandler：带参接收operateService（传给子处理器）
func NewNftContractHandler(operateService *service.OperateService) ContractEventHandler {
	// 子处理器构造函数改为带参，传入operateService
	mintedHandler := NewNftMintedEventHandler(operateService)
	return &ContractNftHandler{
		mintedHandler: mintedHandler,
	}
}

func (n *ContractNftHandler) SupportedEvents() []string {
	return []string{
		_type.NFTMinted.String(),
	}
}

// HandleEvent 处理Minted事件（原有逻辑保留，仅优化错误返回）
func (n *ContractNftHandler) HandleEvent(abi abi.ABI, eventName string, vLog types.Log, blockTime string) error {
	// 前置校验
	if n == nil {
		return errors.New("NftHandler 实例为空")
	}
	if len(vLog.Topics) == 0 {
		return errors.New("日志Topics为空，无法解析事件")
	}

	switch eventName {
	case _type.NFTMinted.String():
		// 注意：原代码是handle（小写），需确认子处理器方法名是否为Handle（大写），避免大小写问题
		err := n.mintedHandler.Handle(abi, vLog, blockTime)
		if err != nil {
			return errors.Join(errors.New("处理NFT Minted事件失败"), err)
		}

	default:
		log.Printf("忽略未知事件类型：%s", eventName)
	}

	return nil
}

// ========== 3. 核心：带参的初始化注册函数（替代原有的init()） ==========
// initContractNftHandler：手动传参注册NFT处理器，适配注册函数签名
// 入参：需要传给子处理器的operateService（无需DI，直接传）
func InitContractNftHandler(operateService *service.OperateService) {
	log.Println("注册 nft 事件处理器（手动传参，不依赖DI）")

	// 关键：闭包捕获operateService，生成注册函数要求的「无参工厂函数」
	factory := func() ContractEventHandler {
		// 闭包内调用带参的NewNftContractHandler，传入operateService
		return NewNftContractHandler(operateService)
	}

	// 注册：factory是无参函数，匹配Register的签名要求
	Register("nft", factory)
}
