package rabbitmq

// 定义一个消息处理函数接口, 所有需要处理的消息都需要实现该接口
type Handler interface {
	handle(msg []byte) error
}

// QueueName 标签用来指定自定义队列名称
type QueueName struct {
	Name string `json:"queue_name"` // 自定义队列名称
}
