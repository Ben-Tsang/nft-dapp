package service

import (
	"context" // 新增：引入context包
	"errors"
	"fmt"
	"nft_backend/internal/app/dto" // 导入公共DTO
	"nft_backend/internal/app/repository"
	"nft_backend/internal/blockchain/contract/event/constant"
	"nft_backend/internal/common"
	"nft_backend/internal/model"
	"time"

	"gorm.io/gorm"
)

// OperateService 操作记录服务
type OperateService struct {
	operateRepo *repository.OperateRepo // 仅Service依赖Repo，反向无依赖
}

// NewOperateService 创建服务实例
func NewOperateService(operateRepo *repository.OperateRepo) *OperateService {
	return &OperateService{
		operateRepo: operateRepo,
	}
}

// ========== 核心方法1：根据ID查询 ==========
func (s *OperateService) GetOperateRecordByID(ctx context.Context, id string) (*model.NFTOperateRecord, error) {
	if id == "" {
		return nil, errors.New("id不能为空")
	}

	record, err := s.operateRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, fmt.Errorf("查询操作记录失败: %w", err)
	}

	return record, nil
}

// ========== 核心方法2：根据交易哈希查询 ==========
func (s *OperateService) GetOperateRecordByTxHash(ctx context.Context, txHash string) (*model.NFTOperateRecord, error) {
	if txHash == "" {
		return nil, errors.New("交易哈希不能为空")
	}

	record, err := s.operateRepo.GetByTxHash(ctx, txHash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, fmt.Errorf("查询操作记录失败: %w", err)
	}

	return record, nil
}

type NFTOperateRecordDTO struct {
	model.NFTOperateRecord        // 嵌套原有结构体，继承所有字段
	OperateTypeLabel       string `json:"operate_type_label"`
	StatusLabel            string `json:"status_label"`
}

// ========== 核心方法3：分页查询（使用公共DTO） ==========
func (s *OperateService) QueryOperateRecords(ctx context.Context, filter dto.OperateQueryFilter, pageNo int, pageSize int, sortBy string, sortDesc bool) (common.PageResult[NFTOperateRecordDTO], error) {
	var result common.PageResult[NFTOperateRecordDTO]

	// 参数校验
	if pageNo < 1 {
		pageNo = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 排序字段白名单（防SQL注入）
	validSortFields := map[string]bool{
		"id":           true,
		"operate_at":   true,
		"block_number": true,
	}
	if !validSortFields[sortBy] {
		sortBy = "operate_at"
	}

	// 调用Repo层（传递ctx）
	records, total, err := s.operateRepo.Query(ctx, filter, pageNo, pageSize, sortBy, sortDesc)
	if err != nil {
		return result, fmt.Errorf("分页查询失败: %w", err)
	}

	// 替换操作类型显示
	dtos := make([]NFTOperateRecordDTO, 0)
	for _, record := range records {
		dto := NFTOperateRecordDTO{
			NFTOperateRecord: record,
			OperateTypeLabel: constant.OperateTypeLabelMap[constant.OperateType(record.OperateType)],
			StatusLabel:      "",
		}
		dtos = append(dtos, dto)
	}
	// 组装分页结果
	result.Records = dtos
	result.Current = pageNo
	result.Size = pageSize
	result.Total = int(total)
	result.Pages = (result.Total + pageSize - 1) / pageSize

	return result, nil
}

// ========== 核心方法4：更新状态 ==========
func (s *OperateService) UpdateOperateRecordStatus(ctx context.Context, txHash, status string) error {
	if txHash == "" {
		return errors.New("交易哈希不能为空")
	}

	// 状态合法性校验
	validStatus := map[string]bool{"pending": true, "success": true, "failed": true}
	if !validStatus[status] {
		return nil
	}

	// 调用Repo更新（传递ctx）
	if err := s.operateRepo.UpdateStatusByTxHash(ctx, txHash, status); err != nil {
		return fmt.Errorf("更新状态失败: %w", err)
	}

	return nil
}

// ========== 核心方法5：创建操作记录 ==========
func (s *OperateService) CreateOperateRecord(ctx context.Context, record *model.NFTOperateRecord) error {
	if record == nil {
		return errors.New("操作记录不能为空")
	}

	// 补充默认值
	if record.Status == "" {
		record.Status = "success"
	}
	if record.OperateAt.IsZero() {
		record.OperateAt = time.Now()
	}

	// 调用Repo创建（传递ctx）
	if err := s.operateRepo.Create(ctx, record); err != nil {
		return fmt.Errorf("创建操作记录失败: %w", err)
	}

	return nil
}

// 获取枚举列表

type selects struct {
	Value string
	Label string
}

func (s *OperateService) GetSelects() map[string][]selects {
	// 1. 初始化返回结果（避免 nil map 赋值报错）
	result := make(map[string][]selects)

	// 2. 从 constant 包的映射表构建操作类型选项列表
	var operateTypeList []selects
	for typ, label := range constant.OperateTypeLabelMap {
		operateTypeList = append(operateTypeList, selects{
			Value: string(typ), // 转成字符串（如 "mint"）
			Label: label,       // 中文标签（如 "NFT铸造"）
		})
	}
	var operateStatusList []selects
	for status, label := range constant.OperateStatusLabelMap {
		operateStatusList = append(operateStatusList, selects{
			Value: string(status),
			Label: label,
		})
	}
	// 3. 赋值到返回结果中，key 为 "operateType"
	result["operateType"] = operateTypeList
	result["operateStatus"] = operateStatusList
	return result
}
