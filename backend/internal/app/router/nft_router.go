package router

import (
	nftHandler "nft_backend/internal/app/handler"
	"nft_backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func NftRegisterRoutes(r *gin.Engine, handler *nftHandler.NftHandler) *gin.Engine {
	g := r.Group("/nft")
	{
		// 注册中间件
		g.Use(middleware.Auth())
		g.GET("/myList", handler.MyNFTPageList)
		g.GET("/discoveryList", handler.DiscoveryList)
	}
	g1 := r.Group("/nft/mint")
	{
		g1.Use(middleware.Auth())
		// 组装数据
		g1.POST("/checkFileDuplicate", handler.CheckFileDuplicate)
	}
	g2 := r.Group("/test")
	{
		g2.GET("/blockQuery", handler.BlockQuery)
	}
	return r
}
