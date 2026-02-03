package main

import (
	"context"
	"ethereum-monitor/database"
	"ethereum-monitor/logger"
	"ethereum-monitor/monitor"
	"ethereum-monitor/scheduler"
	"ethereum-monitor/utils"
	"ethereum-monitor/wallet"
	"log"
	"os"

	"go.uber.org/zap"

	"github.com/joho/godotenv"
)

func main() {
	// 加载 .env 文件（如果存在）
	if err := godotenv.Load(); err != nil {
		log.Println("未找到 .env 文件，将使用系统环境变量")
	}

	// 初始化日志系统
	logger.Init()
	defer logger.Log.Sync()

	// 设置日志记录器
	utils.SetLogger(logger.Log)

	// 初始化数据库
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./ethereum_monitor.db"
	}
	if err := database.NewSqlite(dbPath); err != nil {
		logger.Log.Fatal("数据库初始化失败: " + err.Error())
	}
	defer database.Close()
	logger.Log.Info("数据库初始化成功", zap.Any("fields", map[string]interface{}{"path": dbPath}))

	// 配置代理（如果设置了 HTTP_PROXY）
	if proxyURL := os.Getenv("HTTP_PROXY"); proxyURL != "" {
		if err := utils.SetGlobalProxy(proxyURL); err != nil {
			logger.Log.Error("设置代理失败: " + err.Error())
		}
	}

	// ==================== 选择启动模式 ====================

	// 初始化调度器 (必须在 StartMemeMonitor 之前)
	scheduler.Init()
	scheduler.Start()
	defer scheduler.Stop()

	// 方式 1: 启动地址监控（币安 + OKX）
	// go startAddressMonitor() // Run in background

	// 方式 2: 启动 Meme 币监控（推荐）
	// 监听新合约部署和 Uniswap 交易对创建
	if err := monitor.StartMemeMonitor(); err != nil {
		logger.Log.Error("启动 Meme 监控失败", zap.Error(err))
	}

	err := wallet.StartWatcherMonitor(context.Background())
	if err != nil {
		logger.Log.Error("启动 Watcher 监控失败", zap.Error(err))
	}

	// 阻塞主程，防止退出
	// select {}
	// StartMemeMonitor 本身是阻塞的，所以不需要 select {}，除非 StartMemeMonitor 出错返回

}
