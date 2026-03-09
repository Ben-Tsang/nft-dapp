package router

import (
	"nft_backend/internal/app/handler"

	"github.com/gin-gonic/gin"
)

func AuthRegisterRoutes(r *gin.Engine, authHandler *handler.AuthHandler) {
	g := r.Group("/auth")
	{
		// 创建 UserService

		// 注册路由
		//authHandler := handler.NewAuthHandler(userService, c, rdb)
		g.POST("/login", authHandler.Login)
		g.GET("/nonce", authHandler.Nonce)
	}

}
