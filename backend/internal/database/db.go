package database

import (
	"fmt"
	"nft_backend/internal/blockchain/block/log"
	"nft_backend/internal/blockchain/block/status"
	model2 "nft_backend/internal/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB(dsn string) (*gorm.DB, error) {

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		// 开启详细日志，打印所有 SQL
		//Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}

	fmt.Println("数据库迁移...")
	if err := db.AutoMigrate(&model2.User{},
		&model2.NFT{},                // NFT记录
		&model2.NFTOperateRecord{},   // nft操作记录
		&model2.NFTTransferRecord{},  // NFT转账记录
		&log.BlockProcessLog{},       // 区块处理日志(主要)
		&status.BlockProcessStatus{}, // 区块处理状态(单条记录)
		&model2.Event{},              // NFT事件记录
	); err != nil {
		return nil, fmt.Errorf("数据库迁移失败: %v", err)
	}
	// 检查是否存在默认用户
	if err = EnsureDefaultAdmin(db); err != nil {
		panic(err)
	}

	return db, nil
}
