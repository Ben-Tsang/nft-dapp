package repository

import (
	"context" // 新增：引入context包
	"fmt"
	"nft_backend/internal/app/dto" // 导入公共DTO包（不再依赖Service）
	"nft_backend/internal/model"

	"gorm.io/gorm"
)

// OperateRepo 操作记录仓库（彻底解耦Service）
type OperateRepo struct {
	db *gorm.DB
}

// NewOperateRepo 创建仓库实例
func NewOperateRepo(db *gorm.DB) *OperateRepo {
	return &OperateRepo{
		db: db,
	}
}

// ========== Repo方法1：根据ID查询 ==========
func (r *OperateRepo) GetByID(ctx context.Context, id string) (*model.NFTOperateRecord, error) {
	var record model.NFTOperateRecord
	// 传递ctx到GORM，支持超时/链路追踪
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// ========== Repo方法2：根据交易哈希查询 ==========
func (r *OperateRepo) GetByTxHash(ctx context.Context, txHash string) (*model.NFTOperateRecord, error) {
	var record model.NFTOperateRecord
	// 传递ctx到GORM
	err := r.db.WithContext(ctx).Where("tx_hash = ?", txHash).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// ========== Repo方法3：分页查询（新增关键字模糊搜索） ==========
func (r *OperateRepo) Query(ctx context.Context, filter dto.OperateQueryFilter, page, pageSize int, sortBy string, sortDesc bool) ([]model.NFTOperateRecord, int64, error) {
	// 1. 构建查询条件
	query := r.db.WithContext(ctx).Model(&model.NFTOperateRecord{})

	// 原有条件过滤（完全保留）
	if filter.OwnerAddress != "" {
		query = query.Where("owner_address = ?", filter.OwnerAddress)
	}
	if filter.ContractAddress != "" {
		query = query.Where("contract_address = ?", filter.ContractAddress)
	}
	if filter.TokenID != "" {
		query = query.Where("token_id = ?", filter.TokenID)
	}
	if filter.OperateType != "" {
		query = query.Where("operate_type = ?", filter.OperateType)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	// 新增：关键字模糊搜索（多字段OR匹配）
	if filter.SearchTerm != "" {
		// 拼接模糊查询条件：token_id / tx_hash / contract_address 三个字段模糊匹配
		searchStr := "%" + filter.SearchTerm + "%" // GORM的LIKE匹配符
		query = query.Where(
			"token_id LIKE ? OR tx_hash LIKE ? OR contract_address LIKE ?",
			searchStr, searchStr, searchStr,
		)
	}

	// 2. 获取总条数（保留原有逻辑）
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count records failed: %w", err)
	}

	// 3. 排序（保留原有逻辑）
	var sortStr string
	if sortDesc {
		sortStr = fmt.Sprintf("%s DESC", sortBy)
	} else {
		sortStr = fmt.Sprintf("%s ASC", sortBy)
	}
	query = query.Order(sortStr)

	// 4. 分页（保留原有逻辑）
	offset := (page - 1) * pageSize
	query = query.Offset(offset).Limit(pageSize)

	// 5. 查询数据（保留原有逻辑）
	var records []model.NFTOperateRecord
	if err := query.Find(&records).Error; err != nil {
		return nil, 0, fmt.Errorf("find records failed: %w", err)
	}

	return records, total, nil
}

// ========== Repo方法4：更新状态 ==========
func (r *OperateRepo) UpdateStatusByTxHash(ctx context.Context, txHash, status string) error {
	// 传递ctx到GORM
	return r.db.WithContext(ctx).Model(&model.NFTOperateRecord{}).
		Where("tx_hash = ?", txHash).
		Update("status", status).Error
}

// ========== Repo方法5：创建记录 ==========
func (r *OperateRepo) Create(ctx context.Context, record *model.NFTOperateRecord) error {
	// 传递ctx到GORM
	return r.db.WithContext(ctx).Create(record).Error
}
