package dto

// OperateQueryFilter 操作记录查询筛选条件（公共结构体，解决循环依赖）
type OperateQueryFilter struct {
	OwnerAddress    string // 原有：用户地址
	ContractAddress string // 原有：合约地址
	TokenID         string // 原有：TokenID
	OperateType     string // 原有：操作类型
	Status          string // 原有：状态
	SearchTerm      string // 新增：关键字搜索（NFT ID/交易哈希/合约地址）
}

type Pagination struct {
	PageNo   int    `form:"pageNo" binding:"omitempty,min=1"`
	PageSize int    `form:"pageSize" binding:"omitempty,min=10,max=100"`
	SortBy   string `form:"sortBy"`
	SortDesc bool   `form:"sortDesc"`
}
