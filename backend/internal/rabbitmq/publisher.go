package rabbitmq

import (
	"log"

	"github.com/streadway/amqp"
)

// Publisher 用于发布消息到 RabbitMQ
type Publisher struct {
	Channel *amqp.Channel
}

// NewPublisher 创建一个新的 Publisher
func NewPublisher(channel *amqp.Channel) *Publisher {
	return &Publisher{Channel: channel}
}

// Publish 向指定队列发送消息
func (p *Publisher) Publish(queueName string, message []byte) error {
	// 声明队列
	q, err := p.Channel.QueueDeclare(
		queueName, // 队列名称
		true,      // 是否持久化
		false,     // 是否自动删除
		false,     // 是否排他
		false,     // 是否阻塞
		nil,       // 额外的属性
	)
	if err != nil {
		return err
	}

	// 发布消息
	err = p.Channel.Publish(
		"",     // 默认交换机
		q.Name, // 队列名称
		true,   // 是否确认
		true,   // 是否立即
		amqp.Publishing{
			ContentType: "text/plain",

			Body: []byte(message),
		},
	)
	if err != nil {
		return err
	}

	log.Printf("发送到消息队列: %s", message)
	return nil
}
