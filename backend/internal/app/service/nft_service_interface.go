package service

import (
	"context" // 新增ctx依赖
	"nft_backend/internal/common"
	"nft_backend/internal/model"
)

// NFTServicer NFT服务抽象接口，供外部包（如事件处理、控制器、校正流程）依赖
// 方法名、入参、返回值与Service实现严格一致，保证接口校验通过
type NFTServicer interface {
	// 铸币/校正用：根据合约地址+TokenID查询NFT，查不到返回gorm.ErrRecordNotFound
	// ctx: 上下文（超时/取消/链路追踪）
	// contractAddress: 合约地址
	// tokenId: TokenID
	// return: NFT记录/错误
	GetNFT(ctx context.Context, contractAddress, tokenId string) (*model.NFT, error)

	// 铸币/校正用：创建NFT记录，带FirstOrCreate原子防重复，入参匹配铸币事件数据
	// ctx: 上下文
	// tokenId: TokenID
	// ownerId: 拥有者地址
	// name: NFT名称
	// description: NFT描述
	// tokenURI: NFT资源URI
	// contractAddress: 合约地址
	// blockNum: 区块号
	// return: 错误信息
	CreateNft(ctx context.Context, tokenId, ownerId, name, description, tokenURI, contractAddress string, blockNum uint64) error

	// 铸币/校正用：更新NFT记录，根据合约地址+TokenID唯一定位，无匹配返回gorm.ErrRecordNotFound
	// ctx: 上下文
	// contractAddress: 合约地址
	// tokenId: TokenID
	// ownerId: 新拥有者地址
	// name: 新名称
	// description: 新描述
	// tokenURI: 新资源URI
	// blockNum: 区块号
	// return: 错误信息
	UpdateNft(ctx context.Context, contractAddress, tokenId, ownerId, name, description, tokenURI string, blockNum uint64) error

	// 业务用：分页查询当前用户的NFT列表
	// ctx: 上下文
	// c: Gin上下文（供日志/参数获取等）
	// ownerId: 用户地址
	// pageNo: 页码
	// pageSize: 每页条数
	// return: 分页结果/错误
	PageListMyNFT(ctx context.Context, ownerId string, pageNo, pageSize int) (common.PageResult[model.NFT], error)

	// 业务用：分页查询非自身的已上架NFT（市场列表）
	// ctx: 上下文
	// c: Gin上下文
	// ownerId: 当前用户地址（排除自身）
	// pageNo: 页码
	// pageSize: 每页条数
	// return: 分页结果/错误
	PageListExcludedOwnerNFTs(ctx context.Context, ownerId string, pageNo, pageSize int) (common.PageResult[model.NFT], error)

	// 业务用：NFT上架，传入tokenId/价格/上架时间戳（秒级）
	// ctx: 上下文
	// tokenId: TokenID
	// price: 上架价格
	// listedAt: 上架时间戳（秒级）
	// return: 错误信息
	ListNFT(ctx context.Context, tokenId, price, listedAt string) error

	// 业务用：NFT下架，传入tokenId/下架时间戳（秒级）
	// ctx: 上下文
	// tokenId: TokenID
	// unlistedAt: 下架时间戳（秒级）
	// return: 错误信息
	UnlistedNFT(ctx context.Context, tokenId, unlistedAt string) error

	// 业务用：更新NFT上架价格，传入tokenId/新价格/操作时间戳（秒级，预留）
	// ctx: 上下文
	// tokenId: TokenID
	// price: 新价格
	// time: 操作时间戳（预留）
	// return: 错误信息
	UpdatePrice(ctx context.Context, tokenId, price, time string) error

	// 业务用：变更NFT拥有者（购买/转让），传入tokenId/新拥有者/购买时间戳（秒级）
	// ctx: 上下文
	// tokenId: TokenID
	// ownerId: 新拥有者地址
	// buyAt: 购买时间戳（秒级）
	// return: 错误信息
	ChangeOwner(ctx context.Context, tokenId, ownerId, buyAt string) error

	// 业务用：检查NFT资源哈希是否重复（防重传），传入哈希值
	// ctx: 上下文
	// c: Gin上下文
	// hash: 资源哈希值
	// return: 是否重复（true=重复，false=不重复）
	CheckFileDuplicate(ctx context.Context, hash string) bool
}
