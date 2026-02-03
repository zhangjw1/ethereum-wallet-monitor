package scheduler

import (
	"ethereum-monitor/logger"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

var cronScheduler *cron.Cron

// Init 初始化定时任务调度器
func Init() {
	// 创建一个支持秒级精度的 cron 调度器
	cronScheduler = cron.New(cron.WithSeconds())

	logger.Log.Info("定时任务调度器初始化成功")
}

// Start 启动所有定时任务
func Start() {
	if cronScheduler == nil {
		logger.Log.Error("定时任务调度器未初始化")
		return
	}

	// 注册每天 0:00 执行的任务
	_, err := cronScheduler.AddFunc("0 0 0 * * *", DailyMidnightTask)
	if err != nil {
		logger.Log.Error("注册每日 0:00 任务失败", zap.Error(err))
		return
	}
	logger.Log.Info("✅ 已注册定时任务: 每天 0:00 执行")

	// 启动调度器
	cronScheduler.Start()
	logger.Log.Info("🚀 定时任务调度器已启动")
}

// Stop 停止定时任务调度器
func Stop() {
	if cronScheduler != nil {
		cronScheduler.Stop()
		logger.Log.Info("定时任务调度器已停止")
	}
}

// DailyMidnightTask 每天 0:00 执行的任务
func DailyMidnightTask() {
	logger.Log.Info("⏰ 执行每日 0:00 定时任务")

	// TODO: 在这里添加你的业务逻辑
	// 例如：
	// 1. 清理过期数据
	// 2. 生成每日统计报告
	// 3. 发送每日汇总通知
	// 4. 数据库备份
	//monitor, err := wallet.NewGoEthereumWalletMonitor(config.GetEthereumRpcUrl())
	//if err != nil {
	//	return
	//}
	//balance, err := monitor.GetBalance(config.OkxWalletAddress)
	//if err != nil {
	//	return
	//}

	logger.Log.Info("✅ 每日 0:00 定时任务执行完成")
}
