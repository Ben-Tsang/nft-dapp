package rabbitmq

import (
	"log"
	"nft_backend/internal/logger"

	"github.com/streadway/amqp"
)

// Consumer 用于从队列中消费消息
type Consumer struct {
	Channel    *amqp.Channel
	dispatcher *Dispatcher
	queueName  string
}

// NewConsumer 创建一个新的 Consumer
func NewConsumer(channel *amqp.Channel, dispatcher *Dispatcher, queueName string) *Consumer {
	// 声明队列一次，并在初始化时配置
	_, err := channel.QueueDeclare(
		queueName, // 队列名称
		true,      // 是否持久化
		false,     // 是否自动删除
		false,     // 是否排他
		false,     // 是否阻塞
		nil,       // 额外的属性
	)
	if err != nil {
		log.Fatalf("Queue declaration failed: %v", err)
	}

	return &Consumer{
		Channel:    channel,
		dispatcher: dispatcher,
		queueName:  queueName,
	}
}

// Consume 从指定队列中消费消息
func (c *Consumer) Consume() (<-chan amqp.Delivery, error) {
	// 从队列中消费消息
	logger.L.Warn("rabbitmq comsume方法处理队列: " + c.queueName)
	msgs, err := c.Channel.Consume(
		c.queueName, // 队列名称
		"",          // 消费者名称
		true,        // 自动确认
		false,       // 排他队列
		false,       // 是否阻塞
		false,       // 是否在消费队列中等待
		nil,         // 额外的属性
	)
	if err != nil {
		return nil, err
	}

	return msgs, nil
}

// Process 消费到的消息
func (c *Consumer) Process(msgs <-chan amqp.Delivery) {
	logger.L.Warn("rabbitmq 收到消费消息")
	for msg := range msgs {
		log.Printf("rabbitmq 消费者收到消息: %s", msg.Body)
		// 在这里你可以处理消息，例如写入数据库、触发其他操作等
		log.Printf("队列名: %s", msg.RoutingKey)
		//c.dispatcher.Dispatch(msg.RoutingKey, msg.Body)
		// 消息处理完毕后确认
		msg.Ack(false)
	}
}
