package rabbitmq

import (
	"fmt"
	"nft_backend/internal/config"

	"github.com/streadway/amqp"
)

// RabbitMQConnection 用于存储 RabbitMQ 的连接信息
type RabbitMQConnection struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
}

// Connect 连接到 RabbitMQ 服务
func Connect(mqCfg config.RabbitMQSection) (*RabbitMQConnection, error) {
	// 构造 RabbitMQ 连接 URL
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/%s", mqCfg.User, mqCfg.Password, mqCfg.Host, mqCfg.Port, mqCfg.Vhost)
	fmt.Println("rabbitmq url: " + url)
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	// 返回连接和信道信息, 这里通道是共享的
	return &RabbitMQConnection{
		Connection: conn,
		Channel:    channel,
	}, nil
}

// Close 关闭 RabbitMQ 连接
func (r *RabbitMQConnection) Close() {
	if r.Channel != nil {
		r.Channel.Close()
	}
	if r.Connection != nil {
		r.Connection.Close()
	}
}
