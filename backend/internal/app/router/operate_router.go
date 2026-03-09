package router

import (
	"nft_backend/internal/app/handler"
	"nft_backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterOperateRoutes 注册操作记录相关路由
func OperateRegisterRoutes(r *gin.Engine, operateHandler *handler.OperateHandler) {

	// 操作记录接口分组
	operateGroup := r.Group("/nft/operate")
	operateGroup.Use(middleware.Auth())
	{
		operateGroup.GET("/records", operateHandler.ListOperateRecords)               // 分页查询
		operateGroup.GET("/records/:id", operateHandler.GetOperateRecordByID)         // 按ID查询
		operateGroup.GET("/records/tx", operateHandler.GetOperateRecordByTxHash)      // 按交易哈希查询
		operateGroup.PUT("/records/status", operateHandler.UpdateOperateRecordStatus) // 更新状态
		operateGroup.GET("/selects", operateHandler.ListSelects)                      // 搜索栏枚举列表
	}
}
