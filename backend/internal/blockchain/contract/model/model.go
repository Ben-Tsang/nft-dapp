package model

import (
	"math/big"
	"time"
)

// NFTModel NFT数据库模型（映射nft.NFT，保持字段兼容）
type NFTModel struct {
	ID              string    `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	ContractAddr    string    `gorm:"index;not null"`
	TokenID         string    `gorm:"index;not null"`
	OwnerID         string    `gorm:"not null"`
	IsListed        bool      `gorm:"default:false"`
	Price           *big.Int  `gorm:"type:bigint"`
	LastCorrectTime time.Time `gorm:"autoUpdateTime"`
}

// TaskConfig 任务配置
type TaskConfig struct {
	CronSpec      string `yaml:"cron_spec"`
	BatchSize     int    `yaml:"batch_size"`
	MaxBlockRange int    `yaml:"max_block_range"`
	Timeout       int    `yaml:"timeout"`
	RetryTimes    int    `yaml:"retry_times"`
}
