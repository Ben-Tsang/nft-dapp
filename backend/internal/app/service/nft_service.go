package service

import (
	"context" // 新增ctx依赖
	"fmt"
	"nft_backend/internal/app/repository"
	"nft_backend/internal/common"
	"nft_backend/internal/logger"
	"nft_backend/internal/model"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 接口校验：确保Service完全实现NFTServicer，编译期校验
var _ NFTServicer = (*NftService)(nil)

// NftService NFT服务实现层，依赖NFT专属Repo（职责更清晰）
type NftService struct {
	nftRepo *repository.NftRepo // 替换为专属NftRepo，不再直接依赖通用Repo
}

// NewNftService 实例化NFT服务，做依赖非空校验
func NewNftService(nftRepo *repository.NftRepo) *NftService {
	if nftRepo == nil {
		panic("nft service init failed: nft repository is nil")
	}
	return &NftService{nftRepo: nftRepo}
}

// 分页查询操作记录（添加ctx，兼容原有参数）
func (s *NftService) PageListOperationRecord(ctx context.Context, c *gin.Context, id string, no int, size int) (*common.PageResult[model.NFT], error) {
	// 注：如需实现该方法，可基于nftRepo扩展，此处保留原有返回
	return nil, nil
}

// -------------------------- 铸币/校正核心方法：接口实现 --------------------------
// GetNFT 根据合约地址+TokenID查询NFT，查不到返回gorm.ErrRecordNotFound
func (s *NftService) GetNFT(ctx context.Context, contractAddress, tokenId string) (*model.NFT, error) {
	nft, err := s.nftRepo.GetByContractAndTokenID(ctx, contractAddress, tokenId)
	if err != nil {
		logger.Sugar.Infof("未找到记录, contextAddress: %s, tokenId: %s", contractAddress, tokenId)
		return nil, err
	}
	return nft, nil
}

// CreateNft 创建NFT记录，原子防重复
func (s *NftService) CreateNft(ctx context.Context, tokenId, ownerId, name, description, tokenURI, contractAddress string, blockNum uint64) error {
	// 类型转换：适配模型的BlockNumber为string类型
	blockNumStr := fmt.Sprintf("%d", blockNum)
	now := time.Now()
	// 构建NFT模型
	nft := &model.NFT{
		TokenID:         tokenId,
		OwnerID:         ownerId,
		NftName:         name,
		NftDescription:  description,
		NftURI:          tokenURI,
		ContractAddress: contractAddress,
		BlockNumber:     blockNumStr,
		LastCorrectTime: &now,
		IsListed:        false,
	}

	// 调用Repo原子创建
	if err := s.nftRepo.FirstOrCreate(ctx, nft); err != nil {
		return fmt.Errorf("nft FirstOrCreate failed: %w", err)
	}
	return nil
}

// UpdateNft 更新NFT记录，根据合约地址+TokenID唯一定位
func (s *NftService) UpdateNft(ctx context.Context, contractAddress, tokenId, ownerId, name, description, tokenURI string, blockNum uint64) error {
	blockNumStr := fmt.Sprintf("%d", blockNum)
	now := time.Now()
	// 仅更新业务相关字段
	updateFields := map[string]interface{}{
		"owner_id":          ownerId,
		"nft_name":          name,
		"nft_description":   description,
		"nft_uri":           tokenURI,
		"block_number":      blockNumStr,
		"last_correct_time": &now,
	}

	// 调用Repo更新
	rowsAffected, err := s.nftRepo.UpdateByContractAndTokenID(ctx, contractAddress, tokenId, updateFields)
	if err != nil {
		return fmt.Errorf("nft update failed: %w", err)
	}
	// 无匹配记录，返回GORM原生未找到错误
	if rowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// -------------------------- 分页查询方法：接口实现 --------------------------
// PageListMyNFT 分页查询当前用户的NFT列表
func (s *NftService) PageListMyNFT(ctx context.Context, ownerId string, pageNo, pageSize int) (common.PageResult[model.NFT], error) {
	var result common.PageResult[model.NFT]
	result.Records = make([]model.NFT, 0)

	// 分页参数校验
	if pageNo <= 0 {
		pageNo = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	// 查询总数（调用Repo）
	total, countErr := s.nftRepo.CountByOwnerID(ctx, ownerId)
	if countErr != nil {
		return result, fmt.Errorf("nft my list count failed: %w", countErr)
	}

	// 无数据直接返回
	if total == 0 {
		result.Current = pageNo
		result.Size = pageSize
		result.Total = int(total)
		result.Pages = 0
		return result, nil
	}

	// 分页查询（调用Repo）
	offset := (pageNo - 1) * pageSize
	nfts, queryErr := s.nftRepo.ListByOwnerID(ctx, ownerId, offset, pageSize)
	if queryErr != nil {
		return result, fmt.Errorf("nft my list query failed: %w", queryErr)
	}

	// 组装分页结果
	result.Records = nfts
	result.Current = pageNo
	result.Size = pageSize
	result.Total = int(total)
	result.Pages = (result.Total + pageSize - 1) / pageSize

	return result, nil
}

// PageListExcludedOwnerNFTs 分页查询非自身的已上架NFT
func (s *NftService) PageListExcludedOwnerNFTs(ctx context.Context, ownerId string, pageNo, pageSize int) (common.PageResult[model.NFT], error) {
	var result common.PageResult[model.NFT]
	result.Records = make([]model.NFT, 0)

	// 分页参数校验
	if pageNo <= 0 {
		pageNo = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	// 查询总数（调用Repo）
	total, countErr := s.nftRepo.CountMarketNFTs(ctx, ownerId)
	if countErr != nil {
		return result, fmt.Errorf("nft market list count failed: %w", countErr)
	}

	if total == 0 {
		result.Current = pageNo
		result.Size = pageSize
		result.Total = int(total)
		result.Pages = 0
		return result, nil
	}

	// 分页查询（调用Repo）
	offset := (pageNo - 1) * pageSize
	nfts, queryErr := s.nftRepo.ListMarketNFTs(ctx, ownerId, offset, pageSize)
	if queryErr != nil {
		return result, fmt.Errorf("nft market list query failed: %w", queryErr)
	}

	// 组装分页结果
	result.Records = nfts
	result.Current = pageNo
	result.Size = pageSize
	result.Total = int(total)
	result.Pages = (result.Total + pageSize - 1) / pageSize

	return result, nil
}

// -------------------------- NFT业务操作：上架/下架/改价/转赠 --------------------------
// ListNFT NFT上架
func (s *NftService) ListNFT(ctx context.Context, tokenId, price, listedAt string) error {
	logger.L.Info("nft list", zap.String("tokenId", tokenId), zap.String("price", price), zap.String("listedAt", listedAt))
	// 解析秒级时间戳
	timestamp, err := strconv.ParseInt(listedAt, 10, 64)
	if err != nil {
		return fmt.Errorf("parse listedAt timestamp failed: %w", err)
	}
	parsedTime := time.Unix(timestamp, 0)

	// 查询NFT（调用Repo）
	nft, err := s.nftRepo.GetByTokenID(ctx, tokenId)
	if err != nil {
		return fmt.Errorf("get nft by tokenId failed: %w", err)
	}

	// 更新上架相关字段
	nft.Price = price
	nft.IsListed = true
	nft.ListedAt = &parsedTime
	nft.UnListedAt = nil

	// 执行更新（调用Repo）
	if err := s.nftRepo.UpdateByTokenID(ctx, nft); err != nil {
		return fmt.Errorf("update nft list status failed: %w", err)
	}
	return nil
}

// UnlistedNFT NFT下架
func (s *NftService) UnlistedNFT(ctx context.Context, tokenId, unlistedAt string) error {
	logger.L.Info("nft unlisted", zap.String("tokenId", tokenId), zap.String("unlistedAt", unlistedAt))
	// 解析时间戳
	timestamp, err := strconv.ParseInt(unlistedAt, 10, 64)
	if err != nil {
		return fmt.Errorf("parse unlistedAt timestamp failed: %w", err)
	}
	parsedTime := time.Unix(timestamp, 0)

	// 查询NFT（调用Repo）
	nft, err := s.nftRepo.GetByTokenID(ctx, tokenId)
	if err != nil {
		return fmt.Errorf("get nft by tokenId failed: %w", err)
	}

	// 更新下架相关字段
	nft.Price = "0"
	nft.IsListed = false
	nft.UnListedAt = &parsedTime

	// 执行更新（调用Repo）
	if err := s.nftRepo.UpdateByTokenID(ctx, nft); err != nil {
		return fmt.Errorf("update nft unlisted status failed: %w", err)
	}
	return nil
}

// UpdatePrice 更新NFT上架价格
func (s *NftService) UpdatePrice(ctx context.Context, tokenId, price, time string) error {
	logger.L.Info("nft update price", zap.String("tokenId", tokenId), zap.String("price", price))
	// 查询NFT（调用Repo）
	nft, err := s.nftRepo.GetByTokenID(ctx, tokenId)
	if err != nil {
		return fmt.Errorf("get nft by tokenId failed: %w", err)
	}

	// 仅更新价格
	nft.Price = price
	if err := s.nftRepo.UpdateByTokenID(ctx, nft); err != nil {
		return fmt.Errorf("update nft price failed: %w", err)
	}
	return nil
}

// ChangeOwner 变更NFT拥有者
func (s *NftService) ChangeOwner(ctx context.Context, tokenId, ownerId, buyAt string) error {
	logger.L.Info("nft change owner", zap.String("tokenId", tokenId), zap.String("newOwnerId", ownerId), zap.String("buyAt", buyAt))
	// 解析购买时间戳
	timestamp, err := strconv.ParseInt(buyAt, 10, 64)
	if err != nil {
		return fmt.Errorf("parse buyAt timestamp failed: %w", err)
	}
	buyTime := time.Unix(timestamp, 0)

	// 查询NFT（调用Repo）
	nft, err := s.nftRepo.GetByTokenID(ctx, tokenId)
	if err != nil {
		return fmt.Errorf("get nft by tokenId failed: %w", err)
	}

	// 更新拥有者及相关状态
	nft.OwnerID = ownerId
	nft.Price = "0"
	nft.IsListed = false
	nft.ListedAt = nil
	nft.UnListedAt = nil
	nft.BuyAt = &buyTime
	nft.LastCorrectTime = &buyTime

	// 执行更新（调用Repo）
	if err := s.nftRepo.UpdateByTokenID(ctx, nft); err != nil {
		return fmt.Errorf("change nft owner failed: %w", err)
	}
	return nil
}

// -------------------------- 辅助方法：接口实现 --------------------------
// CheckFileDuplicate 检查NFT资源哈希是否重复
func (s *NftService) CheckFileDuplicate(ctx context.Context, hash string) bool {
	// 调用Repo查重
	count, _ := s.nftRepo.CountByNftURIHash(ctx, hash)
	return count > 0
}
