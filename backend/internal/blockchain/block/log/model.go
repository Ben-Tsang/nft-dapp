package log

import "time"

// -------------------------- 辅助日志表：block_process_log --------------------------
// 定义区块处理状态常量（与日志表status字段对应）
const (
	ProcessStatusFailed  = 0 // 处理失败
	ProcessStatusSuccess = 1 // 处理成功（有合约交易且处理完成）
	ProcessStatusNoNeed  = 2 // 无需处理（无合约交易）
)

// BlockProcessLog 记录每个区块的处理结果（成功/失败/无需处理）
type BlockProcessLog struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement;comment:自增主键" json:"id"`
	ChainID     int64     `gorm:"column:chain_id;comment:链id" json:"chain_id"`
	BlockNumber int64     `gorm:"column:block_number;not null;comment:区块号" json:"block_number"`
	Status      int       `gorm:"column:status;not null;comment:处理状态：0=失败 1=成功 2=无需处理" json:"status"`
	ProcessTime time.Time `gorm:"column:process_time;not null;default:CURRENT_TIMESTAMP;comment:处理时间" json:"process_time"`
	ErrorMsg    string    `gorm:"column:error_msg;type:text;null;comment:失败时的错误信息" json:"error_msg"`
}
