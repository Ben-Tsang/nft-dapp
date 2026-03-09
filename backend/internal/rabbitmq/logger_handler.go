package rabbitmq

import (
	"gorm.io/gorm"
)

// 这里加入一个属性,用于指定自定义队列名称
type LoggerHandler struct {
	db        *gorm.DB
	QueueName `json:"logger_queue"`
}

func NewLoggerHandler(db *gorm.DB) *LoggerHandler {
	return &LoggerHandler{
		QueueName: QueueName{Name: "logger_queue"},
		db:        db,
	}
}

/*func (l LoggerHandler) handle(msg []byte) error {
	var event model.QuestStatusEvent
	err := json.Unmarshal(msg, &event)
	if err != nil {
		return err
	}
	logger.L.Info("mq消费者: logger日志处理, 记录日志")
	// 记录日志
	var log model.QuestLog
	log.QuestID = event.QuestID
	log.OldStatus = event.OldStatus
	log.NewStatus = event.NewStatus
	log.OperatorID = event.OperatorID
	log.CreatedAt = time.Now()
	if l.database == nil {
		logger.L.Error("mq消费者: logger日志处理失败, db为空")
	} else {
		if err := l.database.Create(&log).Error; err != nil {
			return err
		}
	}

	return nil
}*/
