package di

import (
	"log"
	repo2 "nft_backend/internal/app/repository"
	service2 "nft_backend/internal/app/service"

	"github.com/ethereum/go-ethereum/ethclient"
	"gorm.io/gorm"
)

// Bootstrap 统一初始化所有业务实例（对外暴露的唯一入口）
// main函数只需调用这个方法，即可完成DI容器+所有实例的初始化
func Bootstrap(db *gorm.DB, client *ethclient.Client) {
	// 1. 先初始化容器核心
	Init()
	// 2. 注册所有业务服务（集中管理）
	registerServices(db, client)
}

// registerServices 注册所有业务服务（包内私有，细分逻辑）
func registerServices(db *gorm.DB, client *ethclient.Client) {
	// 注册NFT服务
	nftRepo := repo2.NewNftRepo(db)
	nftSvc := service2.NewNftService(nftRepo)
	if err := Register[*service2.NftService](nftSvc); err != nil {
		log.Fatalf("注册NFT服务失败：%v", err)
	}

	// 注册事件服务
	eventRepo := repo2.NewEventRepo(db)
	eventSvc := service2.NewEventService(eventRepo)
	if err := Register[*service2.EventService](eventSvc); err != nil {
		log.Fatalf("注册事件服务失败：%v", err)
	}

	// 注册区块链客户端
	if err := Register[*ethclient.Client](client); err != nil {
		log.Fatalf("注册区块链client失败：%v", err)
	}

	// 新增服务时，只需在这里添加即可
	// tokenSvc := &status.TokenServiceImpl{}
	// if err := Register[status.TokenService](tokenSvc); err != nil {
	// 	log.Fatalf("注册Token服务失败：%v", err)
	// }
}
