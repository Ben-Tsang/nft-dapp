package repository

import (
	"context" // 新增ctx依赖
	"nft_backend/internal/model"

	"gorm.io/gorm"
)

// 原有Repo保留（兼容历史代码），新增NFT专属Repo
/*type Repo struct {
	Db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{Db: db}
}*/

// ========== 新增：NFT专属仓储方法（抽离Service中的数据库操作） ==========
type NftRepo struct {
	db *gorm.DB
}

func NewNftRepo(db *gorm.DB) *NftRepo {
	return &NftRepo{db: db}
}

// GetByContractAndTokenID 根据合约地址+TokenID查询NFT
func (r *NftRepo) GetByContractAndTokenID(ctx context.Context, contractAddress, tokenId string) (*model.NFT, error) {
	var nft model.NFT
	err := r.db.WithContext(ctx).
		Where("contract_address = ? AND token_id = ?", contractAddress, tokenId).
		First(&nft).Error
	if err != nil {
		return nil, err
	}
	return &nft, nil
}

// GetByTokenID 根据TokenID查询NFT
func (r *NftRepo) GetByTokenID(ctx context.Context, tokenId string) (*model.NFT, error) {
	var nft model.NFT
	err := r.db.WithContext(ctx).
		Where("token_id = ?", tokenId).
		First(&nft).Error
	if err != nil {
		return nil, err
	}
	return &nft, nil
}

// FirstOrCreate 原子创建NFT（防重复）
func (r *NftRepo) FirstOrCreate(ctx context.Context, nft *model.NFT) error {
	return r.db.WithContext(ctx).
		Where("contract_address = ? AND token_id = ?", nft.ContractAddress, nft.TokenID).
		FirstOrCreate(nft).Error
}

// UpdateByContractAndTokenID 根据合约地址+TokenID更新NFT字段
func (r *NftRepo) UpdateByContractAndTokenID(ctx context.Context, contractAddress, tokenId string, updateFields map[string]interface{}) (int64, error) {
	result := r.db.WithContext(ctx).
		Model(&model.NFT{}).
		Where("contract_address = ? AND token_id = ?", contractAddress, tokenId).
		Updates(updateFields)
	return result.RowsAffected, result.Error
}

// CountByOwnerID 统计指定所有者的NFT数量
func (r *NftRepo) CountByOwnerID(ctx context.Context, ownerId string) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&model.NFT{}).
		Where("owner_id = ?", ownerId).
		Count(&total).Error
	return total, err
}

// ListByOwnerID 分页查询指定所有者的NFT
func (r *NftRepo) ListByOwnerID(ctx context.Context, ownerId string, offset, limit int) ([]model.NFT, error) {
	var nfts []model.NFT
	err := r.db.WithContext(ctx).
		Where("owner_id = ?", ownerId).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&nfts).Error
	return nfts, err
}

// CountMarketNFTs 统计非指定所有者的已上架NFT数量
func (r *NftRepo) CountMarketNFTs(ctx context.Context, ownerId string) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&model.NFT{}).
		Where("owner_id != ? AND is_listed = ?", ownerId, true).
		Count(&total).Error
	return total, err
}

// ListMarketNFTs 分页查询非指定所有者的已上架NFT
func (r *NftRepo) ListMarketNFTs(ctx context.Context, ownerId string, offset, limit int) ([]model.NFT, error) {
	var nfts []model.NFT
	err := r.db.WithContext(ctx).
		Where("owner_id != ? AND is_listed = ?", ownerId, true).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&nfts).Error
	return nfts, err
}

// UpdateByTokenID 根据TokenID更新NFT（全字段更新）
func (r *NftRepo) UpdateByTokenID(ctx context.Context, nft *model.NFT) error {
	return r.db.WithContext(ctx).Save(nft).Error
}

// CountByNftURIHash 统计包含指定哈希的NFT数量（查重用）
func (r *NftRepo) CountByNftURIHash(ctx context.Context, hash string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.NFT{}).
		Where("nft_uri LIKE ?", "%"+hash+"%").
		Count(&count).Error
	return count, err
}

// 保留原有Create方法（兼容历史调用）
/*func (r *Repo) Create(ctx context.Context, nft *model.NFT) error {
	return r.Db.WithContext(ctx).Create(nft).Error
}*/
