package auth

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

// nonce
func GenerateNonce(rdb *redis.Client) (string, error) {
	nonceBytes := make([]byte, 16)  // 16 字节的随机数切片, 都是初始值 0
	_, err := rand.Read(nonceBytes) // 填充随机数
	if err != nil {
		return "", nil
	} // 填充随机数

	nonce := fmt.Sprintf("%x", nonceBytes)
	// 保存到redis, 10秒过期
	rdb.Set(context.Background(), nonce, "1", 180*time.Second)
	log.Println("nonce:", nonce)
	return nonce, nil // 转换为十六进制字符串
}
