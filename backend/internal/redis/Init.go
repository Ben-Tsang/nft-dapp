package redis

import (
	"fmt"
	"log"
	"nft_backend/internal/config"
	"time"

	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
)

// Redis 客户端对象
var Rdb *redis.Client
var Ctx = context.Background()

// 初始化 Redis 客户端
func InitRedis() *redis.Client {
	appConfig, _ := config.Get()
	// 创建 Redis 客户端
	Rdb = redis.NewClient(&redis.Options{
		Addr:     appConfig.Redis.Host + ":" + fmt.Sprintf("%d", appConfig.Redis.Port), // Redis 地址
		Password: appConfig.Redis.Password,                                             // 默认无密码
		DB:       0,                                                                    // 使用默认数据库
	})

	// 测试 Redis 连接
	_, err := Rdb.Ping(Ctx).Result()
	if err != nil {
		log.Fatalf("redis连接失败: %v", err)
	}

	Rdb.Set(context.Background(), "name", "admin", 10*time.Second)

	fmt.Println("连接redis成功")

	return Rdb
}
