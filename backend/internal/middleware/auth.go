package middleware

import (
	"fmt"
	"nft_backend/internal/app/web/response"
	"nft_backend/internal/auth"
	"strings"

	"github.com/gin-gonic/gin"
)

const ()

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 要求使用Bearer Token进行认证
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Fail(c, 401, "未登录")
			c.Abort()
			return
		}

		// 获取token
		parts := strings.Fields(authHeader)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Fail(c, 401, "认证失败")
			c.Abort()
			return
		}
		tokenStr := parts[1]

		// 验证token
		token, err := auth.ParseToken(tokenStr)
		if err != nil {
			response.Fail(c, 401, "认证失败")
			c.Abort()
			return
		}
		fmt.Println(">>> token from middleware:", token) // 打印整个 claims
		c.Set("userId", token.UserID)
		c.Set("role", token.Role)
		c.Next()
	}
}
