package model

import "time"

// nft事件模型
type Event struct {
	// 核心修复：PostgreSQL 自增标签（gorm 标签适配 PG）
	ID              int64     `gorm:"primaryKey;type:bigserial;autoIncrement" database:"id" json:"id" comment:"主键ID"`
	EventType       string    `database:"event_type" json:"eventType" comment:"事件类型：Mint/Transfer/Burn"`
	TxHash          string    `database:"tx_hash" json:"txHash" comment:"交易哈希"`
	BlockNumber     uint64    `database:"block_number" json:"blockNumber" comment:"区块高度"`
	LogIndex        uint      `database:"log_index" json:"logIndex" comment:"日志索引"`
	ContractAddress string    `database:"contract_address" json:"contractAddress" comment:"合约地址"`
	TokenID         string    `database:"token_id" json:"tokenId" comment:"Token ID"`
	FromAddress     string    `database:"from_address" json:"fromAddress" comment:"发送方地址"`
	ToAddress       string    `database:"to_address" json:"toAddress" comment:"接收方地址"`
	Amount          string    `database:"amount" json:"amount" comment:"数量"`
	ExtraData       string    `database:"extra_data" json:"extraData" comment:"额外数据"`
	CreatedAt       time.Time `gorm:"autoCreateTime" database:"created_at" json:"createdAt" comment:"创建时间"`
}
