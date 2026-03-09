package timer

import (
	"log"
	"os"
	"time"

	"github.com/robfig/cron/v3"
)

// 全局调度器实例（单例）
var Cron *cron.Cron

// 任务ID映射：key=业务唯一标识，value=定时器任务ID
var TaskIDMap = make(map[string]cron.EntryID)

// 初始化定时器（内部方法，外部无需调用）
func initCron() {
	// 加载东八区时区
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		log.Printf("【Cron】加载时区失败，使用本地时区：%v", err)
		loc = time.Local
	}

	// 常规生产配置：秒级+东八区+自定义日志+最大并发10
	Cron = cron.New(
		cron.WithSeconds(),
		cron.WithLocation(loc),
		cron.WithLogger(NewLogger()),
		//cron.WithMaxConcurrentJobs(10),
	)

	Cron.Start()
	log.Println("【Cron】调度器初始化并启动成功")
}

// 自定义日志器
func NewLogger() cron.Logger {
	return cron.VerbosePrintfLogger(
		log.New(os.Stdout, "[Cron-Scheduler] ", log.LstdFlags|log.Lshortfile),
	)
}

// 注册定时任务（对外暴露，封装错误捕获）
func RegisterTask(bizKey string, cronExpr string, taskFunc func()) error {
	wrapperFunc := func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[Cron-Task] 业务[%s]执行panic：%v，时间：%s",
					bizKey, r, time.Now().Format("2006-01-02 15:04:05.000"))
			}
		}()
		taskFunc()
	}

	// 移除旧任务，避免重复
	if oldID, ok := TaskIDMap[bizKey]; ok {
		Cron.Remove(oldID)
		log.Printf("[Cron-Task] 业务[%s]旧任务已移除（ID：%d）", bizKey, oldID)
	}

	// 添加新任务
	taskID, err := Cron.AddFunc(cronExpr, wrapperFunc)
	if err != nil {
		log.Printf("[Cron-Task] 业务[%s]注册失败：%v", bizKey, err)
		return err
	}

	TaskIDMap[bizKey] = taskID
	log.Printf("[Cron-Task] 业务[%s]注册成功，Cron：%s，ID：%d", bizKey, cronExpr, taskID)
	return nil
}

// 移除定时任务
func RemoveTask(bizKey string) {
	if taskID, ok := TaskIDMap[bizKey]; ok {
		Cron.Remove(taskID)
		delete(TaskIDMap, bizKey)
		log.Printf("[Cron-Task] 业务[%s]任务已删除（ID：%d）", bizKey, taskID)
	} else {
		log.Printf("[Cron-Task] 业务[%s]任务不存在，无需删除", bizKey)
	}
}

// 优雅停止调度器（内部方法）
func stopCron() {
	if Cron != nil {
		ctx := Cron.Stop()
		<-ctx.Done()
		log.Println("【Cron】调度器已优雅停止，所有任务执行完毕")
	}
}
