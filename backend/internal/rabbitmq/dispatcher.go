package rabbitmq

import (
	"nft_backend/internal/logger"
	"reflect"

	"gorm.io/gorm"
)

// 定义一个消息分发器, 属性是消息处理函数的映射表, key是队列名称, value是对应的消息处理函数
type Dispatcher struct {
	handlers map[string]Handler
}

// 用于消费者分发消息到对应的业务处理函数
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		handlers: make(map[string]Handler),
	}
}

// 用于消费者注册消息处理函数
func (d *Dispatcher) RegisterHandler(queue string, handler Handler) {
	d.handlers[queue] = handler
}

// 自动注册消息处理函数, 用于消费者注册时自动注册
// 找出所有实现了Handler接口的对象, 并注册到消息处理函数映射表中
func (d *Dispatcher) AutoRegister(db *gorm.DB) {
	handlers := getHandlers(db)
	for _, handler := range handlers {
		// 获取处理器的类型
		handlerType := reflect.TypeOf(handler).Elem()
		// 获取 QueueName 标签值（队列名称）
		queueName := handlerType.Field(0).Tag.Get("json")
		d.RegisterHandler("logger_queue", handler)
		logger.L.Info("注册事件处理器, 队列名称: " + queueName)
	}

}

// 用于消费者从队列中获取消息并分发到对应的业务处理函数
func (d *Dispatcher) Dispatch(queue string, msg []byte) error {
	if handler, ok := d.handlers[queue]; ok {
		return handler.handle(msg)
	}
	return nil
}

// 每次新增一个消息处理函数时, 都需要在这里注册一下
// handler的队列名称通过结构体的第一个字段的json标签获取
func getHandlers(db *gorm.DB) []Handler {
	return []Handler{
		//NewLoggerHandler(database),
	}
}
