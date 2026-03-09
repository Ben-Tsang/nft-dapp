package status

import (
	"time"

	"gorm.io/gorm"
)

// BlockProcessStatusRepo 状态仓储层
type BlockProcessStatusRepo struct {
	db *gorm.DB
}

func NewBlockProcessStatusRepo(db *gorm.DB) *BlockProcessStatusRepo {
	return &BlockProcessStatusRepo{db: db}
}

// InitBlockStatus 初始化状态表（首次执行）
func (r *BlockProcessStatusRepo) InitBlockStatus(baseNumber int64) error {
	var count int64
	if err := r.db.Model(&BlockProcessStatus{}).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return r.db.Create(&BlockProcessStatus{
			ID:          1,
			LatestBlock: baseNumber,
			Version:     0,
		}).Error
	}
	return nil
}

// GetLatestBlock 获取最新区块处理状态，无记录则自动创建
func (r *BlockProcessStatusRepo) GetLatestBlock(chainID int64) (*BlockProcessStatus, error) {
	var status BlockProcessStatus
	err := r.db.Where("chain_id = ?", chainID).First(&status).Error

	// 处理三种情况：无错误、记录未找到、其他错误
	switch err {
	case nil:
		// 找到记录，直接返回
		return &status, nil
	case gorm.ErrRecordNotFound:
		// 未找到记录，创建默认记录
		defaultStatus := &BlockProcessStatus{
			ChainID:     chainID, // 链id
			LatestBlock: 0,       // 默认区块号为0
			UpdateTime:  time.Now(),
			Version:     0, // 默认版本号为0
		}
		// 创建记录
		if createErr := r.db.Create(defaultStatus).Error; createErr != nil {
			// 创建失败时返回错误
			return nil, createErr
		}
		return defaultStatus, nil
	default:
		// 其他数据库错误（如连接失败等），返回错误
		return nil, err
	}
}

// Update 更新区块状态
func (r *BlockProcessStatusRepo) Update(chainID int64, status *BlockProcessStatus) error {
	// 乐观锁更新（防止并发问题）
	return r.db.Model(status).Where("chain_id = ? AND version = ?", chainID, status.Version).
		Updates(map[string]interface{}{
			"latest_block": status.LatestBlock,
			"update_time":  gorm.Expr("NOW()"),
			"version":      status.Version + 1,
		}).Error
}
