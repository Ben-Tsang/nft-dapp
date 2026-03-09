package model

import (
	"time"

	"gorm.io/gorm"
)

// NFT
type NFT struct {
	ID              string     `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	TokenID         string     `json:"token_id" gorm:"not null;index:idx_contract_token,unique"` // ✅ 标准名
	OwnerID         string     `json:"owner_id" gorm:"not null"`
	NftName         string     `json:"nft_name"`
	NftDescription  string     `json:"nft_description"`
	NftURI          string     `json:"nft_uri"`
	ContractAddress string     `json:"contract_address" gorm:"not null;index:idx_contract_token,unique"`
	BlockNumber     string     `json:"block_number"`
	Price           string     `json:"price"`
	IsListed        bool       `json:"is_listed"`
	ListedAt        *time.Time `json:"listed_at"`
	UnListedAt      *time.Time `json:"unlisted_at"`
	LastCorrectTime *time.Time `json:"last_correct_time"`
	BuyAt           *time.Time `json:"buy_at"`
	CreatedAt       time.Time  `json:"created_at" gorm:"default:current_timestamp"`
	UpdatedAt       time.Time  `json:"updated_at" gorm:"default:current_timestamp;autoUpdateTime"`
}

// NFTOperateRecord 专业版
type NFTOperateRecord struct {
	ID              string    `gorm:"column:id;primaryKey;type:varchar(64);default:uuid_generate_v4()" json:"id"`
	ContractAddress string    `gorm:"column:contract_address;not null;type:varchar(42);index" json:"contract_address"`
	TokenID         string    `gorm:"column:token_id;not null;type:varchar(64);index" json:"token_id"` // ✅ 标准
	UserAddress     string    `gorm:"column:user_address;not null;type:varchar(42);index" json:"user_address"`
	OwnerAddress    string    `gorm:"column:owner_address;not null;type:varchar(42)" json:"owner_address"`
	OperateType     string    `gorm:"column:operate_type;not null;type:varchar(32);index" json:"operate_type"`
	Amount          string    `gorm:"column:amount;type:varchar(64)" json:"amount"`
	TxHash          string    `gorm:"column:tx_hash;type:varchar(66);index" json:"tx_hash"`
	Status          string    `gorm:"column:status;not null;type:varchar(16);default:pending;index" json:"status"`
	BlockNumber     string    `gorm:"column:block_number;type:varchar(32)" json:"block_number"`
	OperateAt       time.Time `gorm:"column:operate_at;not null;default:current_timestamp;index" json:"operate_at"`
	CreatedAt       time.Time `json:"created_at" gorm:"default:current_timestamp"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"default:current_timestamp;autoUpdateTime"`
	Remark          string    `gorm:"column:remark;type:varchar(255)" json:"remark"`
}

// NFTTransferRecord 对应数据库表 nft_transfer_record（NFT转账记录）
type NFTTransferRecord struct {
	ID          uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                  // 主键ID
	TokenID     string    `gorm:"column:token_id;not null" json:"token_id"`                      // NFT编号（注意：表中写的是token_idtoken_int，实际应为token_id）
	FromAddress string    `gorm:"column:from_address;type:varchar(255)" json:"from_address"`     // 转出地址
	ToAddress   string    `gorm:"column:to_address;type:varchar(255)" json:"to_address"`         // 转入地址
	BlockNumber uint64    `gorm:"column:block_number;not null" json:"block_number"`              // 区块号
	TxIndex     int       `gorm:"column:tx_index;not null" json:"tx_index"`                      // 交易索引
	LogIndex    int       `gorm:"column:log_index;not null" json:"log_index"`                    // 日志索引
	TxHash      string    `gorm:"column:tx_hash;type:varchar(255)" json:"tx_hash"`               // 交易哈希
	IsRemoved   int       `gorm:"column:is_removed;default:0" json:"is_removed"`                 // 事件是否回滚（0=正常 1=回滚）
	CreatedAt   time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"` // 记录时间
}

type QuestStatusEvent struct {
	ID         int            `gorm:"primarykey;autoIncrement" json:"id"`                 // 主键ID
	QuestID    string         `gorm:"type:varchar(64);not null;index" json:"quest_id"`    // 任务ID
	OldStatus  string         `gorm:"type:varchar(32);not null" json:"old_status"`        // 旧状态
	NewStatus  string         `gorm:"type:varchar(32);not null" json:"new_status"`        // 新状态
	OperatorID string         `gorm:"type:varchar(64);not null;index" json:"operator_id"` // 操作者ID
	OccurredAt time.Time      `gorm:"type:datetime;not null;index" json:"occurred_at"`    // 发生时间
	CreatedAt  time.Time      `json:"created_at"`                                         // 创建时间
	UpdatedAt  time.Time      `json:"updated_at"`                                         // 更新时间
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deleted_at"`                            // 软删除时间
}

type QuestStatus struct {
	NextStatus string
}
