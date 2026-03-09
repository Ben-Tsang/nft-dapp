package handler

import (
	"context"
	"fmt"
	"nft_backend/internal/app/dto"
	service "nft_backend/internal/app/service"
	"nft_backend/internal/app/web/response" // 引入你已有的响应包

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
)

// OperateHandler 操作记录处理器（适配现有响应结构）
type OperateHandler struct {
	operateService *service.OperateService
}

// NewOperateHandler 创建操作记录处理器实例
func NewOperateHandler(operateService *service.OperateService) *OperateHandler {
	return &OperateHandler{
		operateService: operateService,
	}
}

// ========== 核心接口1：分页查询操作记录（适配前端TanStack Table） ==========
func (h *OperateHandler) ListOperateRecords(c *gin.Context) {

	// 获取用户id
	ownerID := c.GetString("userId")
	// 校验地址是否合法
	if !common.IsHexAddress(ownerID) {
		response.Fail(c, -1, "地址不合法")
		return
	}
	ownerID = common.HexToAddress(c.GetString("userId")).Hex() // 转成标准EIP-55地址格式

	// 提取分页参数
	var pagination dto.Pagination
	if err := c.ShouldBindQuery(&pagination); err != nil {
		response.Fail(c, -1, err.Error())
		return // 补充return，避免后续逻辑执行
	}

	// 获取分页参数
	pageNo := pagination.PageNo
	pageSize := pagination.PageSize
	sortBy := pagination.SortBy

	// 构造过滤条件（新增SearchTerm参数接收）
	filter := dto.OperateQueryFilter{
		OwnerAddress:    ownerID,
		ContractAddress: c.Query("contract_address"),
		TokenID:         c.Query("token_id"),
		OperateType:     c.Query("operateType"),
		Status:          c.Query("status"),
		SearchTerm:      c.Query("searchTerm"), // 新增：接收前端的关键字搜索参数
	}

	ctx := context.Background()
	// 调用Service查询
	result, err := h.operateService.QueryOperateRecords(ctx, filter, pageNo, pageSize, sortBy, true)
	if err != nil {
		response.Fail(c, -1, err.Error())
		return
	}

	response.OK(c, "获取用户操作记录成功", result)
}

// ========== 核心接口2：根据ID查询单条操作记录 ==========
// @Summary 根据ID查询NFT操作记录
// @Description 获取单条操作记录详情
// @Tags 操作记录
// @Accept json
// @Produce json
// @Param id path string true "操作记录ID"
// @Success 200 {object} response.Response{data=model.NFTOperateRecord}
// @Router /api/v1/operate/records/{id} [get]
func (h *OperateHandler) GetOperateRecordByID(c *gin.Context) {
	// 1. 解析路径参数
	id := c.Param("id")
	if id == "" {
		response.Fail(c, 400, "操作记录ID不能为空")
		return
	}
	ctx := context.Background()
	// 2. 调用Service查询
	record, err := h.operateService.GetOperateRecordByID(ctx, id)
	if err != nil {
		if err == service.ErrRecordNotFound {
			response.Fail(c, 404, "操作记录不存在")
			return
		}
		response.Fail(c, 500, fmt.Sprintf("查询操作记录失败: %v", err))
		return
	}

	// 3. 返回结果（适配你的Response结构）
	response.OK(c, "查询成功", record)
}

// ========== 核心接口3：根据交易哈希查询操作记录 ==========
// @Summary 根据交易哈希查询NFT操作记录
// @Description 按链上TxHash溯源操作记录
// @Tags 操作记录
// @Accept json
// @Produce json
// @Param tx_hash query string true "交易哈希"
// @Success 200 {object} response.Response{data=model.NFTOperateRecord}
// @Router /api/v1/operate/records/tx [get]
func (h *OperateHandler) GetOperateRecordByTxHash(c *gin.Context) {
	// 1. 解析查询参数
	txHash := c.Query("tx_hash")
	if txHash == "" {
		response.Fail(c, 400, "交易哈希不能为空")
		return
	}
	ctx := context.Background()
	// 2. 调用Service查询
	record, err := h.operateService.GetOperateRecordByTxHash(ctx, txHash)
	if err != nil {
		if err == service.ErrRecordNotFound {
			response.Fail(c, 404, "未找到该交易哈希对应的操作记录")
			return
		}
		response.Fail(c, 500, fmt.Sprintf("查询操作记录失败: %v", err))
		return
	}

	// 3. 返回结果（适配你的Response结构）
	response.OK(c, "查询成功", record)
}

// ========== 核心接口4：更新操作记录状态 ==========
// @Summary 更新NFT操作记录状态
// @Description 手动更新操作记录状态（如pending→success/failed）
// @Tags 操作记录
// @Accept json
// @Produce json
// @Param tx_hash query string true "交易哈希"
// @Param status query string true "状态（pending/success/failed）"
// @Success 200 {object} response.Response{data=bool}
// @Router /api/v1/operate/records/status [put]
func (h *OperateHandler) UpdateOperateRecordStatus(c *gin.Context) {
	// 1. 解析参数
	txHash := c.Query("tx_hash")
	status := c.Query("status")
	if txHash == "" {
		response.Fail(c, 400, "交易哈希不能为空")
		return
	}
	if status == "" {
		response.Fail(c, 400, "状态不能为空")
		return
	}

	// 2. 校验状态合法性
	validStatus := map[string]bool{
		"pending": true,
		"success": true,
		"failed":  true,
	}
	if !validStatus[status] {
		response.Fail(c, 400, "状态只能是 pending/success/failed")
		return
	}
	ctx := context.Background()
	// 3. 调用Service更新
	err := h.operateService.UpdateOperateRecordStatus(ctx, txHash, status)
	if err != nil {
		response.Fail(c, 500, fmt.Sprintf("更新操作记录状态失败: %v", err))
		return
	}

	// 4. 返回结果（适配你的Response结构）
	response.OK(c, "更新成功", true)
}

// 搜索栏枚举列表
func (h *OperateHandler) ListSelects(c *gin.Context) {

	result := h.operateService.GetSelects()

	response.OK(c, "获取枚举列表成功", result)
}

// ========== 补充：所需的Service层结构体/方法定义（参考） ==========
/*
// 1. 在 internal/app/service/operate_service.go 中定义筛选结构体
type OperateQueryFilter struct {
	OwnerAddress     string
	ContractAddress string
	TokenID         string
	OperateType     string
	Status          string
	StartTime       time.Time
	EndTime         time.Time
}

// 2. Service层核心方法（需实现）
func (s *OperateService) QueryOperateRecords(filter OperateQueryFilter, page, pageSize int, sortBy string, sortDesc bool) ([]model.NFTOperateRecord, int64, error) {
	// 实现分页查询逻辑
}

func (s *OperateService) GetOperateRecordByID(id string) (*model.NFTOperateRecord, error) {
	// 实现根据ID查询逻辑
}

func (s *OperateService) GetOperateRecordByTxHash(txHash string) (*model.NFTOperateRecord, error) {
	// 实现根据交易哈希查询逻辑
}

func (s *OperateService) UpdateOperateRecordStatus(txHash, status string) error {
	// 实现更新状态逻辑
}

// 3. 错误常量定义
var ErrRecordNotFound = errors.New("record not found")
*/
