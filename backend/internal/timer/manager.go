package timer

import (
	"fmt"
	"nft_backend/internal/blockchain/contract/block"
	"nft_backend/internal/logger"

	"go.uber.org/zap"
)

// -------------------------- 新增2个对外方法（供main调用，核心） --------------------------
// StartCron 初始化+启动定时器+注册任务，非阻塞执行（替代原StartAndManageCron）
// 原阻塞的信号监听逻辑删除，由main统一管理优雅关闭
func StartCron() {
	initCron()         // 复用原有初始化逻辑
	registerAllTasks() // 复用原有任务注册逻辑
}

// StopCron 优雅关闭定时器（供main收到退出信号后统一调用，复用原有stopCron）
func StopCron() {
	stopCron()
}

// -------------------------- 原有代码保留（仅删除信号监听阻塞逻辑） --------------------------
// registerAllTasks 注册所有定时任务（集中管理，新增/修改任务只改这里，原封不动）
func registerAllTasks() {
	//_ = RegisterTask("block_correct", "*/10 * * * * *", nftCorrectService.Correct) // 每10秒同步NFT
	//_ = RegisterTask("nft_status", "*/30 * * * * *", NFTStatusCheckTask) // 每30秒检查NFT状态
	// 新增任务直接在这里加一行即可，示例：
	// _ = RegisterTask("order_check", "0 */5 * * * *", task.OrderCheckTask) // 每5分钟检查订单
}

// 新增：按链注册定时任务（任务名拼接链标识，避免冲突）
func RegisterCronTaskByChain(chainKey string, chainCron string, service *block.NFTCorrectService) {
	// 任务名拼接链标识：如 eth_sepolia_nft_correct
	taskName := fmt.Sprintf("%s_nft_correct", chainKey)
	// 原有注册逻辑，仅修改任务名
	if err := RegisterTask(taskName, chainCron, service.Correct); err != nil {
		logger.L.Error("多链定时任务注册失败", zap.String("chainKey", chainKey), zap.String("taskName", taskName), zap.Error(err))
	}
}

// 【删除原StartAndManageCron方法】彻底移除信号监听和<-quit阻塞逻辑
