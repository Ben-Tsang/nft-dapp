package router

import (
	"nft_backend/internal/app/handler"

	"github.com/gin-gonic/gin"

	"nft_backend/internal/middleware"
)

func PersonRegisterRoutes(r *gin.Engine, handler *handler.PersonHandler) *gin.Engine {
	g := r.Group("/person")
	{
		// 注册person中间件
		g.Use(middleware.Auth())
		// 注册person路由
		// 组装数据
		/*repository := person.NewRepository(database)
		service := person.NewService(repository)
		handler := handler2.NewNftHandler(service)*/
		g.POST("", handler.Create)
		g.DELETE("/:id", handler.Delete)
		g.PUT("/:id", handler.Update)
		g.GET("/:id", handler.Get)
		g.GET("/page", handler.PageList)
	}
	return r
}
