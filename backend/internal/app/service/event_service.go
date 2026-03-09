package service

import (
	"context" // 新增：引入context包
	"errors"
	"nft_backend/internal/app/repository"
	"nft_backend/internal/model"

	"gorm.io/gorm"
)

// 定义全局错误变量，供上层判断
var ErrRecordNotFound = gorm.ErrRecordNotFound

type EventService struct {
	repo *repository.EventRepo
}

func NewEventService(repo *repository.EventRepo) *EventService {
	return &EventService{
		repo: repo,
	}
}

// 创建事件记录（补充ctx参数，传递到repo层）
// ctx: 上下文，用于超时控制、链路追踪等
// event: 要创建的事件记录
func (s *EventService) Create(ctx context.Context, event *model.Event) error {
	// 根据事件类型处理业务逻辑（你原有注释的逻辑可后续补充）
	switch event.EventType {
	case "Mint":
		// 后续补充：铸造事件业务逻辑（如新增NFT持有记录）
	case "Transfer":
		// 后续补充：转账事件业务逻辑（如更新NFT持有记录）
	case "Burn":
		// 后续补充：销毁事件业务逻辑（如删除NFT持有记录）
	}

	// 执行数据库创建，传递ctx到repository层
	return s.repo.Create(ctx, event)
}

// 根据交易hash和日志索引获取事件记录（新增ctx参数，适配业务逻辑）
// ctx: 上下文
// hash: 交易哈希
// index: 日志索引
// return: 事件记录/错误
func (s *EventService) GetByTxHashAndLogIndex(ctx context.Context, hash string, index uint) (*model.Event, error) {
	// 传递ctx到repository层
	return s.repo.GetByTxHashAndLogIndex(ctx, hash, index)
}

// 更新事件记录（用于覆盖更新，新增ctx参数）
// ctx: 上下文
// event: 要更新的事件记录（必须包含主键ID）
// return: 错误信息
func (s *EventService) Update(ctx context.Context, event *model.Event) error {
	// 前置校验：主键ID不能为空（更新必须基于主键）
	if event.ID == 0 {
		return errors.New("更新事件记录失败：主键ID不能为空")
	}
	// 执行数据库更新，传递ctx到repository层
	return s.repo.Update(ctx, event)
}
