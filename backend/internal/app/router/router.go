package router

import (
	handler2 "nft_backend/internal/app/handler"
	"nft_backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(authHandler *handler2.AuthHandler, personHandler *handler2.PersonHandler, nftHandler *handler2.NftHandler, operateHandler *handler2.OperateHandler) *gin.Engine {
	r := gin.New()
	// 挂在全局中间件
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())
	// 注册路由
	AuthRegisterRoutes(r, authHandler)
	PersonRegisterRoutes(r, personHandler)
	NftRegisterRoutes(r, nftHandler)
	OperateRegisterRoutes(r, operateHandler)
	return r
}
