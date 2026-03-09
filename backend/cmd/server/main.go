package main

import (
	"context"
	"fmt"
	"net/http"
	handler2 "nft_backend/internal/app/handler"
	"nft_backend/internal/app/repository"
	"nft_backend/internal/app/router"
	"nft_backend/internal/app/service"
	"nft_backend/internal/app/web/validate"
	blockLog "nft_backend/internal/blockchain/block/log"
	"nft_backend/internal/blockchain/block/status"
	"nft_backend/internal/blockchain/contract/basic"
	"nft_backend/internal/blockchain/contract/block"
	"nft_backend/internal/blockchain/contract/event/handler"
	listener2 "nft_backend/internal/blockchain/contract/event/listener"
	"nft_backend/internal/config"
	"nft_backend/internal/database"
	"nft_backend/internal/di"
	"nft_backend/internal/logger"
	"nft_backend/internal/person"
	"nft_backend/internal/rabbitmq"
	"nft_backend/internal/redis"
	"nft_backend/internal/timer"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

func main() {
	// ********** 步骤1：初始化配置 - 严格错误处理 **********
	if err := config.Init(); err != nil {
		fmt.Printf("配置初始化失败：%v\n", err)
		return
	}
	cfg, err := config.Get()
	if err != nil {
		logger.L.Error("获取配置实例失败", zap.Error(err))
		return
	}
	if cfg.App.HTTPPort <= 0 {
		logger.L.Error("端口配置非法", zap.Int("port", cfg.App.HTTPPort))
		return
	}

	// ********** 步骤2：初始化核心基础资源 - 完全复用，无改动 **********
	if err := logger.InitLogger(cfg.App.Env); err != nil {
		fmt.Printf("初始化logger失败: %v\n", err)
		return
	}
	defer func() {
		if err := logger.L.Sync(); err != nil {
			fmt.Printf("日志刷盘失败: %v\n", err)
		}
	}()
	logger.L.Info("logger初始化成功")

	validate.InitValidator()
	logger.L.Info("参数校验器初始化成功")

	gormDB, err := database.InitDB(cfg.DB.DSN())
	if err != nil {
		logger.L.Error("初始化数据库连接失败", zap.Error(err))
		return
	}
	logger.L.Info("数据库连接成功")
	defer func() {
		sqlDB, _ := gormDB.DB()
		if err := sqlDB.Close(); err != nil {
			logger.L.Error("数据库关闭失败", zap.Error(err))
		}
	}()

	rdb := redis.InitRedis()
	logger.L.Info("Redis连接成功")
	defer func() {
		if err := rdb.Close(); err != nil {
			logger.L.Error("Redis关闭失败", zap.Error(err))
		}
	}()

	// ********** 步骤3：初始化DI容器 - 完全复用 **********
	// 多链下ETH客户端改为后续按链初始化，此处DI可按需传入基础资源
	di.Bootstrap(gormDB, nil)
	logger.L.Info("DI容器初始化成功")

	// ********** 步骤4：初始化RabbitMQ - 完全复用 **********
	mqconn, err := rabbitmq.Connect(cfg.RabbitMQ)
	if err != nil {
		logger.L.Fatal("初始化RabbitMQ连接失败", zap.Error(err))
	}
	logger.L.Info("RabbitMQ连接成功")
	defer func() {
		if err := mqconn.Channel.Close(); err != nil {
			logger.L.Error("RabbitMQ通道关闭失败", zap.Error(err))
		}
	}()

	mqconn.Channel.Qos(1, 0, false)
	dispatcher := rabbitmq.NewDispatcher()
	dispatcher.AutoRegister(gormDB)
	consumer := rabbitmq.NewConsumer(mqconn.Channel, dispatcher, "logger_queue")
	go func() {
		logger.L.Info("启动RabbitMQ消费者...")
		msgs, err := consumer.Consume()
		if err != nil {
			logger.L.Fatal("RabbitMQ消费失败", zap.Error(err))
		}
		consumer.Process(msgs)
	}()

	// ********** 步骤5：初始化基础业务层 - Repo/基础Service 完全复用 **********
	// Repo层（所有链共用，无链隔离）
	userRepo := repository.NewUserRepo(gormDB)
	personRepo := person.NewRepo(gormDB)
	nftRepo := repository.NewNftRepo(gormDB)
	blockProcessStatusRepo := status.NewBlockProcessStatusRepo(gormDB)
	blockProcessLogRepo := blockLog.NewBlockProcessLogRepo(gormDB)
	operateRepo := repository.NewOperateRepo(gormDB)
	// 基础Service层（所有链共用）
	userService := service.NewUserService(userRepo)
	personService := person.NewService(personRepo)
	nftService := service.NewNftService(nftRepo)
	operateService := service.NewOperateService(operateRepo)
	// 多链改造：区块状态/日志Service按链隔离，后续初始化
	var chainStatusServices = make(map[string]*status.BlockProcessStatusService)
	var chainLogServices = make(map[string]*blockLog.BlockProcessLogService)
	// 多链核心存储：链标识 -> 链相关核心实例（客户端/监听器/解析器/校正服务）
	var chainCores = make(handler2.ChainCores, len(cfg.Chains))

	// ********** 步骤6：多链核心初始化 - 遍历所有链，按链创建独立实例（核心改造点） **********
	logger.L.Info("开始初始化多链核心实例", zap.Int("chainCount", len(cfg.Chains)))
	var enableChainCount = 0
	// ********** 定时器注册 - 多链校正服务批量注册 **********
	timer.StartCron()

	handler.InitContractMarketHandler(operateService)
	handler.InitContractNftHandler(operateService)
	for chainKey, chainCfg := range cfg.Chains {
		if !chainCfg.Enable {
			logger.Sugar.Infof("%s 链不启用, 不初始化", chainKey)
			break
		}
		enableChainCount++
		logger.L.Info("初始化链核心实例", zap.String("chainKey", chainKey), zap.Int64("chainID", chainCfg.Basic.ChainID))
		// 6.1 按链创建独立ETH客户端（多链隔离，避免客户端冲突）
		ethClient, err := basic.Connect(chainCfg.Basic.RPCUrl)
		if err != nil {
			logger.L.Fatal("连接区块链节点失败", zap.String("chainKey", chainKey), zap.Error(err))
		}
		logger.L.Info("链ETH客户端连接成功", zap.String("chainKey", chainKey))

		// 6.2 按链初始化区块状态/日志Service（链隔离，避免多链区块状态混乱）
		chainStatusServices[chainKey] = status.NewBlockProcessStatusService(blockProcessStatusRepo)
		chainLogServices[chainKey] = blockLog.NewBlockProcessLogService(blockProcessLogRepo)
		// 初始化该链区块起始高度
		if err := chainStatusServices[chainKey].Init(int64(chainCfg.Basic.StartBlock)); err != nil {
			logger.L.Fatal("初始化链区块状态失败", zap.String("chainKey", chainKey), zap.Error(err))
		}
		logger.L.Info("链区块状态初始化成功", zap.String("chainKey", chainKey), zap.Int64("startBlock", int64(chainCfg.Basic.StartBlock)))

		// 6.3 按链创建独立事件监听器
		// 转换合约配置为监听器所需格式
		var listenerContracts []listener2.Contract
		for _, c := range chainCfg.Contracts {
			listenerContracts = append(listenerContracts, listener2.Contract{
				Name:        c.Name,
				Address:     common.HexToAddress(c.Address),
				AbiPath:     c.AbiPath,
				HandlerName: c.Handler,
				RPCUrl:      chainCfg.Basic.RPCUrl, // 优先用链全局RPC
			})
		}
		eventListener, err := listener2.NewEventListener(ethClient, listenerContracts)
		if err != nil {
			logger.L.Fatal("创建链事件监听器失败", zap.String("chainKey", chainKey), zap.Error(err))
		}
		// 后台启动该链事件监听（多链独立协程）
		go func(ck string, el *listener2.EventListener) {
			if err := el.Start(); err != nil {
				logger.L.Fatal("链事件监听启动失败", zap.String("chainKey", ck), zap.Error(err))
			}
		}(chainKey, eventListener)
		logger.L.Info("链事件监听器已启动", zap.String("chainKey", chainKey))

		// 6.4 按链创建独立BlockParser
		// 转换合约配置为解析器所需格式
		var contractConfigs []basic.ContractConfig
		for _, c := range chainCfg.Contracts {
			contractConfigs = append(contractConfigs, basic.ContractConfig{
				Contract: c,
			})
		}
		blockParser, err := block.NewBlockParser(gormDB, ethClient, contractConfigs, chainLogServices[chainKey])
		if err != nil {
			logger.L.Fatal("初始化链区块解析器失败", zap.String("chainKey", chainKey), zap.Error(err))
		}

		// 6.5 按链创建独立NFTCorrectService（链核心业务服务）
		nftCorrectService := block.NewNFTCorrectService(
			gormDB, ethClient, blockParser, nftService,
			eventListener.ContractListeners, chainStatusServices[chainKey], chainLogServices[chainKey],
		)

		// 6.6 保存该链所有核心实例到全局映射
		chainCores[chainKey] = &handler2.ChainCore{
			EthClient:         ethClient,
			EventListener:     eventListener,
			BlockParser:       blockParser,
			NFTCorrectService: nftCorrectService,
		}
		// 从配置中读取当前链的cron表达式：chainCfg.Basic.CorrectCron
		chainCronSpec := chainCfg.Basic.CorrectCron
		// 校验cron表达式非空（避免配置遗漏导致任务注册失败）
		if chainCronSpec == "" {
			logger.L.Warn("链未配置校正cron表达式，跳过定时任务注册", zap.String("chainKey", chainKey))
			continue
		}
		// 调用改造后的方法，传递【链标识+配置的cron+校正服务】
		timer.RegisterCronTaskByChain(chainKey, chainCronSpec, nftCorrectService)
		logger.L.Info("链定时任务注册成功",
			zap.String("chainKey", chainKey),
			zap.String("taskName", fmt.Sprintf("%s_nft_correct", chainKey)),
			zap.String("cronSpec", chainCronSpec))
	}

	logger.Sugar.Infof("多链核心实例初始化完成, 启用链数: %d", enableChainCount)

	// 遍历所有链，注册该链的定时任务

	//logger.L.Info("多链定时任务注册成功", zap.Int("taskCount", len(cfg.Chains)))

	// ********** 步骤8：Web层初始化 - 适配多链，传入多链校正服务 **********
	// 改造NftHandler，支持多链NFT校正（按链标识分发请求）
	authHandler := handler2.NewAuthHandler(userService, rdb)
	personHandler := handler2.NewPersonHandler(personService)
	nftHandler := handler2.NewNftHandler(nftService, chainCores) // 关键：传入多链核心实例，替代原单链校正服务
	operateHandler := handler2.NewOperateHandler(operateService)
	logger.L.Info("多链业务Handler初始化成功")

	r := router.SetupRouter(authHandler, personHandler, nftHandler, operateHandler)
	logger.L.Info("Gin路由注册成功")

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.App.HTTPPort),
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.L.Fatal("Web服务启动失败", zap.Error(err), zap.String("addr", srv.Addr))
		}
	}()
	logger.L.Info("Web服务启动成功", zap.String("addr", srv.Addr))

	// ********** 步骤10：优雅关闭 - 多链资源统一关闭 **********
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit
	logger.L.Info("收到系统退出信号，开始优雅关闭服务...")

	// 关闭定时器
	timer.StopCron()

	// 多链资源优雅关闭：遍历所有链，关闭ETH客户端/事件监听器
	for chainKey, core := range chainCores {
		// 关闭ETH客户端：Close无返回值，直接调用（修复核心报错）
		core.EthClient.Close()
		// 关闭事件监听器（之前改造的Close方法，无返回值）
		core.EventListener.Close()
		// 打印成功日志
		logger.L.Info("链资源优雅关闭成功", zap.String("chainKey", chainKey))
	}

	// 关闭Web服务
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.L.Fatal("Web服务强制关闭", zap.Error(err))
	}
	logger.L.Info("Web服务优雅关闭成功")

	logger.L.Info("多链服务已全部优雅关闭，程序退出")
}
