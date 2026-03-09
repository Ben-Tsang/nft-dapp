package handler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"nft_backend/internal/app/service"
	"nft_backend/internal/di"
	"nft_backend/internal/model"
)

// CommonSaveEvent 所有事件共用的「查-比-增/更」逻辑
// 参数：newEvent=构建好的事件对象，eventType=事件名称（用于日志/错误提示）
func CommonSaveEvent(newEvent *model.Event, eventType string) error {
	// 从DI容器获取eventService（全局单例，不用每个Handler传）
	eventService, _ := di.Resolve[*service.EventService]()
	ctx := context.Background()
	// 1. 查询本地记录
	existEvent, err := eventService.GetByTxHashAndLogIndex(ctx, newEvent.TxHash, newEvent.LogIndex)
	if err != nil {
		// 只有「不是记录不存在」的错误才返回
		if !errors.Is(err, service.ErrRecordNotFound) {
			return fmt.Errorf("查询[%s]事件失败: %w", eventType, err)
		}
		// 无记录 → 新增
		log.Printf("[%s]事件本地无记录，执行新增：TxHash=%s, LogIndex=%d", eventType, newEvent.TxHash, newEvent.LogIndex)
		return eventService.Create(ctx, newEvent)
	}

	// 2. 有记录 → 比对核心字段，不同才更新
	if isEventChanged(existEvent, newEvent) {
		newEvent.ID = existEvent.ID // 继承主键实现覆盖
		log.Printf("[%s]事件记录有变更，执行覆盖更新：TxHash=%s, LogIndex=%d", eventType, newEvent.TxHash, newEvent.LogIndex)
		return eventService.Update(ctx, newEvent)
	}

	// 3. 无变更 → 跳过
	log.Printf("[%s]事件记录无变更，跳过更新：TxHash=%s, LogIndex=%d", eventType, newEvent.TxHash, newEvent.LogIndex)
	return nil
}

// isEventChanged 通用字段比对方法（只比链上字段，忽略本地字段）
func isEventChanged(old, new *model.Event) bool {
	if old.EventType != new.EventType ||
		old.TokenID != new.TokenID ||
		old.FromAddress != new.FromAddress ||
		old.ToAddress != new.ToAddress ||
		old.ExtraData != new.ExtraData ||
		old.BlockNumber != new.BlockNumber ||
		old.ContractAddress != new.ContractAddress ||
		old.Amount != new.Amount {
		return true
	}
	return false
}
