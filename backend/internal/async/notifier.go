package async

/*
import (
	"context"
	"gorm-lab/internal/event"
	"gorm-lab/internal/logger"
	"time"

	"go.uber.org/zap"
)

type Notifier struct {
	ch chan event.QuestStatusEvent
}

func NewNotifier(bufferSize int) *Notifier {
	if bufferSize <= 0 {
		bufferSize = 100
	}
	return &Notifier{
		// make 是初始化一个channel，bufferSize是channel的容量
		ch: make(chan event.QuestStatusEvent, bufferSize),
	}
}

// 发送事件
func (n *Notifier) Publish(evt event.QuestStatusEvent) {

	select {
	case n.ch <- evt:
	default:
		logger.L.Warn("async event channel full, drop event",
			zap.Int("quest_id", evt.QuestID),
			zap.String("old_status", evt.OldStatus),
			zap.String("new_status", evt.NewStatus),
			zap.Int("operator_id", evt.OperatorID),
		)
	}
}

// 启动后台gorountine消费事件
func (n *Notifier) Start(ctx context.Context) {
	go func() {
		logger.L.Info("async event notifier started")
		defer logger.L.Info("async event notifier stopped")
		for {
			select {
			case <-ctx.Done():
				return
			case evt := <-n.ch:
				// 这里就是异步处理逻辑
				handleQuestStatusEvent(evt)
			}
		}
	}()
}

func handleQuestStatusEvent(evt event.QuestStatusEvent) {
	logger.L.Info("quest status changed (async)",
		zap.Int("quest_id", evt.QuestID),
		zap.String("old_status", evt.OldStatus),
		zap.String("new_status", evt.NewStatus),
		zap.Int("operator_id", evt.OperatorID),
		zap.String("occurred_at", evt.OccurredAt.Format(time.RFC3339)),
	)
}
*/
