package basic

import (
	"log"
	"nft_backend/internal/config"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// ========== 核心修改1：扩展ContractConfig，增加ABI相关字段 ==========
type ContractConfig struct {
	Contract      config.ContractCfg
	EventTopicMap EventTopicMap
	ABI           abi.ABI
}

// EventTopicMap 预构建事件哈希->事件名映射
type EventTopicMap map[common.Hash]string

// GetEventType 通过topic0获取事件名（统一逻辑）
func (m EventTopicMap) GetEventType(topic common.Hash) string {
	if name, ok := m[topic]; ok {
		return name
	}
	log.Printf("未匹配到事件名的topic0哈希：%s", topic.Hex())
	return "Unknown"
}

// InitEventTopicMap 初始化事件哈希映射
func InitEventTopicMap(contractABI abi.ABI) EventTopicMap {
	topicMap := make(EventTopicMap)
	for name, event := range contractABI.Events {
		if !event.Anonymous {
			topicMap[event.ID] = name
		}
	}
	return topicMap
}
