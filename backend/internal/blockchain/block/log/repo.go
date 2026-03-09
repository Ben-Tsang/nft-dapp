package log

import (
	"gorm.io/gorm"
)

// BlockProcessLogRepo 日志仓储层（需实现数据库操作）
type BlockProcessLogRepo struct {
	db *gorm.DB
}

func NewBlockProcessLogRepo(db *gorm.DB) *BlockProcessLogRepo {
	return &BlockProcessLogRepo{db: db}
}

// CreateLog 插入区块处理日志
func (r *BlockProcessLogRepo) CreateLog(log *BlockProcessLog) error {
	return r.db.Create(log).Error
}
