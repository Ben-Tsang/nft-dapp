package handler

import (
	"context"
	"fmt"
	"log"
	"nft_backend/internal/app/dto"
	"nft_backend/internal/app/service"
	"nft_backend/internal/app/web/response"
	log2 "nft_backend/internal/blockchain/block/log"
	"nft_backend/internal/blockchain/block/status"
	"nft_backend/internal/blockchain/contract/block"
	"nft_backend/internal/blockchain/contract/event/constant"
	listener2 "nft_backend/internal/blockchain/contract/event/listener"
	"nft_backend/internal/logger"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
)

// 补充HTTP接口专属常量（若constant包未定义，需添加到 internal/constant/constant.go）
// 建议在constant/constant.go中添加：
// const (
//     // HTTPRequestTimeout HTTP接口默认超时时间（查询类接口）
//     HTTPRequestTimeout = 3 * time.Second
//     // EventHandleTimeout 区块链事件处理超时（原有）
//     EventHandleTimeout = 10 * time.Second
// )

// 全局ChainCore：存储单链所有核心实例（main/handler共用）
// 所有链强绑定的实例都放这，按链标识映射管理
type ChainCore struct {
	EthClient          *ethclient.Client                 // 链独立ETH客户端
	EventListener      *listener2.EventListener          // 链独立事件监听器
	BlockParser        *block.BlockParser                // 链独立区块解析器
	NFTCorrectService  *block.NFTCorrectService          // 链独立NFT校正服务
	BlockStatusService *status.BlockProcessStatusService // 链区块状态服务
	BlockLogService    *log2.BlockProcessLogService      // 链区块日志服务
}

// 可选：定义ChainCores为多链核心实例映射（简化main/handler中的类型声明）
type ChainCores map[string]*ChainCore

type NftHandler struct {
	service    service.NFTServicer
	chainCores ChainCores
}

func NewNftHandler(service service.NFTServicer, chainCores ChainCores) *NftHandler {
	return &NftHandler{
		service:    service,
		chainCores: chainCores,
	}
}

// 我的NFT分页列表
func (h *NftHandler) MyNFTPageList(c *gin.Context) {
	// ========== 1. 正确获取ctx：复用Gin请求ctx + HTTP专属超时 ==========
	ctx, cancel := context.WithTimeout(c.Request.Context(), constant.HTTPRequestTimeout)
	defer cancel() // 释放资源，避免内存泄漏

	// ========== 2. 参数校验（补充return，避免后续执行） ==========
	// 获取并校验用户ID
	ownerID := c.GetString("userId")
	if ownerID == "" {
		response.Fail(c, -1, "用户ID为空")
		return
	}
	// 校验地址合法性
	if !common.IsHexAddress(ownerID) {
		response.Fail(c, -1, "地址不合法")
		return
	}
	// 标准化地址格式（EIP-55）
	ownerID = common.HexToAddress(ownerID).Hex()

	// ========== 3. 绑定并校验分页参数 ==========
	var pagination dto.Pagination
	if err := c.ShouldBindQuery(&pagination); err != nil {
		response.Fail(c, -1, fmt.Sprintf("分页参数解析失败：%s", err.Error()))
		return
	}
	// 分页参数合法性校验（设置默认值+范围限制）
	pageNo := pagination.PageNo
	pageSize := pagination.PageSize
	if pageNo < 1 {
		pageNo = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// ========== 4. 调用服务（解耦：仅传ctx，不传Gin的c） ==========
	result, err := h.service.PageListMyNFT(ctx, ownerID, pageNo, pageSize)
	if err != nil {
		response.Fail(c, -1, fmt.Sprintf("获取我的NFT列表失败：%s", err.Error()))
		return
	}

	// ========== 5. 成功返回 ==========
	response.OK(c, "获取数据成功", result)
}

// 检查文件是否重复
type CheckFileDuplicateRequest struct {
	Hash string `json:"hash" binding:"required"` // 增加binding校验，简化手动判断
}

type CheckFileDuplicateResponse struct {
	IsDuplication bool `json:"isDuplication"`
}

func (h *NftHandler) CheckFileDuplicate(c *gin.Context) {
	// ========== 1. 正确获取ctx：复用Gin请求ctx + HTTP专属超时 ==========
	ctx, cancel := context.WithTimeout(c.Request.Context(), constant.HTTPRequestTimeout)
	defer cancel()

	// ========== 2. 绑定并校验请求参数 ==========
	var req CheckFileDuplicateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, -1, fmt.Sprintf("参数解析失败：%s", err.Error()))
		return
	}
	// 双重校验（防止binding失效）
	hash := req.Hash
	if hash == "" {
		response.Fail(c, -1, "hash不能为空")
		return
	}

	log.Println("检查文件重复，hash:", hash)

	// ========== 3. 调用服务（解耦：仅传ctx，不传Gin的c） ==========
	isDuplication := h.service.CheckFileDuplicate(ctx, hash)

	// ========== 4. 构建响应并返回 ==========
	result := CheckFileDuplicateResponse{
		IsDuplication: isDuplication,
	}
	response.OK(c, "检查文件是否重复成功", result)
}

// 发现页(过滤自己上架的nft, 只保留自己能购买的nft)
func (h *NftHandler) DiscoveryList(c *gin.Context) {
	// ========== 1. 正确获取ctx：复用Gin请求ctx + HTTP专属超时 ==========
	ctx, cancel := context.WithTimeout(c.Request.Context(), constant.HTTPRequestTimeout)
	defer cancel()

	// ========== 2. 参数校验（补充return，避免后续执行） ==========
	// 获取并校验用户ID
	ownerID := c.GetString("userId")
	if ownerID == "" {
		response.Fail(c, -1, "用户ID为空")
		return
	}
	// 校验地址合法性
	if !common.IsHexAddress(ownerID) {
		response.Fail(c, -1, "地址不合法")
		return
	}
	// 标准化地址格式（EIP-55）
	ownerID = common.HexToAddress(ownerID).Hex()

	// ========== 3. 绑定并校验分页参数 ==========
	var pagination dto.Pagination
	if err := c.ShouldBindQuery(&pagination); err != nil {
		response.Fail(c, -1, fmt.Sprintf("分页参数解析失败：%s", err.Error()))
		return
	}
	// 分页参数合法性校验
	pageNo := pagination.PageNo
	pageSize := pagination.PageSize
	if pageNo < 1 {
		pageNo = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// ========== 4. 调用服务（解耦：仅传ctx，不传Gin的c） ==========
	result, err := h.service.PageListExcludedOwnerNFTs(ctx, ownerID, pageNo, pageSize)
	if err != nil {
		response.Fail(c, -1, fmt.Sprintf("获取发现页NFT列表失败：%s", err.Error()))
		return
	}

	// ========== 5. 成功返回 ==========
	response.OK(c, "获取数据成功", result)
}

// 区块查询测试（修正方法参数名：ctx → c，符合Gin规范）
func (h *NftHandler) BlockQuery(c *gin.Context) {
	// ========== 1. 正确获取ctx：复用Gin请求ctx + 长耗时超时 ==========
	_, cancel := context.WithTimeout(c.Request.Context(), constant.HTTPLongRequestTimeout)
	defer cancel()

	logger.L.Sugar().Info("测试区块读取...")
	chainName := "hardhat"

	// ========== 2. 校验链实例是否存在 ==========
	chainCore, exist := h.chainCores[chainName]
	if !exist {
		response.Fail(c, -1, fmt.Sprintf("未找到%s链的核心实例", chainName))
		return
	}
	if chainCore.NFTCorrectService == nil {
		response.Fail(c, -1, fmt.Sprintf("%s链的NFT校正服务未初始化", chainName))
		return
	}

	// ========== 3. 执行校正操作（传递ctx，支持超时/取消） ==========
	// 注意：若Correct方法未接收ctx，建议改造为 chainCore.NFTCorrectService.Correct(ctx)
	chainCore.NFTCorrectService.Correct()

	response.OK(c, "区块查询测试成功", nil)
}
