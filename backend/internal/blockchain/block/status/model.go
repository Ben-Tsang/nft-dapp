package status

import (
	"time"
)

// BlockProcessStatus 存储已成功处理的最新区块号（仅1条记录）
type BlockProcessStatus struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键自增" json:"id"`
	ChainID     int64     `gorm:"column:chain_id;not null;default:0;comment:链id" json:"chain_id"`
	LatestBlock int64     `gorm:"column:latest_block;not null;default:0;comment:已成功处理的最新区块号" json:"latest_block"`
	UpdateTime  time.Time `gorm:"column:update_time;not null;default:CURRENT_TIMESTAMP;autoUpdateTime;comment:最后更新时间" json:"update_time"`
	Version     int       `gorm:"column:version;not null;default:0;comment:乐观锁版本号（防并发更新）" json:"version"`
}
