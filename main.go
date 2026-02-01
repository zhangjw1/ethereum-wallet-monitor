package main

import (
	"ethereum-monitor/database"
	"ethereum-monitor/logger"
	"ethereum-monitor/monitor"
	"ethereum-monitor/utils"
	"go.uber.org/zap"
	"log"
	"os"

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

	// 方式 1: 启动 Meme 币监控（推荐）
	// 监听新合约部署和 Uniswap 交易对创建
	monitor.StartMemeMonitor()

	// 方式 2: 使用示例函数启动
	// monitor.ExampleMemeMonitor()

	// 方式 3: 自定义配置启动
	// monitor.StartMemeMonitorWithCustomConfig(
	//     config.GetEthereumRpcUrl(),  // RPC URL
	//     10,                           // 轮询间隔（秒）
	//     true,                         // 启用 PairCreated 监听
	// )

	// 方式 4: 测试分析指定代币
	// monitor.TestAnalyzeToken("0xdAC17F958D2ee523a2206206994597C13D831ec7") // USDT

	// 方式 5: 同时启动钱包监控和 Meme 币监控
	// wallet.AddressAddMonitor() // 在 goroutine 中启动钱包监控
	// monitor.StartMemeMonitor()      // 启动 Meme 币监控

	// ==================== 其他功能 ====================

	// 定时任务调度器
	// scheduler.Init()
	// scheduler.Start()
	// defer scheduler.Stop()

	// MEV 检测示例
	// utils.ExampleUsage()

	// Covalent API 示例
	// utils.ExampleCovalentUsage()
}
