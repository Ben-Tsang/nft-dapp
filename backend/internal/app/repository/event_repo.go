package repository

import (
	"context" // 新增：引入context包
	"errors"
	"nft_backend/internal/model"

	"gorm.io/gorm"
)

type EventRepo struct {
	db *gorm.DB
}

func NewEventRepo(db *gorm.DB) *EventRepo {
	return &EventRepo{
		db: db,
	}
}

// 创建事件记录（新增ctx参数，传递到数据库操作）
// ctx: 上下文，用于超时控制、链路追踪
// event: 要创建的事件记录
func (r *EventRepo) Create(ctx context.Context, event *model.Event) error {
	// 使用WithContext传递ctx，支持超时/取消/链路追踪
	return r.db.WithContext(ctx).Create(event).Error
}

// 根据交易哈希和日志索引查询事件记录（新增ctx参数）
// ctx: 上下文
// hash: 交易哈希
// index: 日志索引
// return: 事件记录/错误
func (r *EventRepo) GetByTxHashAndLogIndex(ctx context.Context, hash string, index uint) (*model.Event, error) {
	var event model.Event
	// 组合唯一条件：TxHash + LogIndex（链上事件的唯一标识）
	// 传递ctx到数据库查询，支持超时控制
	result := r.db.WithContext(ctx).Where("tx_hash = ? AND log_index = ?", hash, index).First(&event)
	if result.Error != nil {
		// 区分“记录不存在”和其他数据库错误
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, result.Error
	}
	return &event, nil
}

// 更新事件记录（根据主键ID覆盖，新增ctx参数）
// ctx: 上下文
// event: 要更新的事件记录（必须包含主键ID）
// return: 错误信息
func (r *EventRepo) Update(ctx context.Context, event *model.Event) error {
	// 传递ctx到数据库操作，保留原有字段筛选逻辑（避免覆盖CreatedAt等自动字段）
	return r.db.WithContext(ctx).Model(event).Select(
		"event_type", "block_number", "contract_address",
		"token_id", "from_address", "to_address",
		"amount", "extra_data",
	).Updates(event).Error
}
