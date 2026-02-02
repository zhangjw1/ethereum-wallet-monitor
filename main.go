package main

import (
	"context"
	"ethereum-monitor/config"
	"ethereum-monitor/database"
	"ethereum-monitor/logger"
	"ethereum-monitor/utils"
	"ethereum-monitor/wallet"
	"log"
	"os"
	"os/signal"
	"syscall"

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

	// 方式 1: 启动地址监控（币安 + OKX）
	startAddressMonitor()

	// 方式 2: 启动 Meme 币监控（推荐）
	// 监听新合约部署和 Uniswap 交易对创建
	// monitor.StartMemeMonitor()

	// 方式 3: 使用示例函数启动
	// monitor.ExampleMemeMonitor()

	// 方式 4: 自定义配置启动
	// monitor.StartMemeMonitorWithCustomConfig(
	//     config.GetEthereumRpcUrl(),  // RPC URL
	//     10,                           // 轮询间隔（秒）
	//     true,                         // 启用 PairCreated 监听
	// )

	// 方式 5: 测试分析指定代币
	// monitor.TestAnalyzeToken("0xdAC17F958D2ee523a2206206994597C13D831ec7") // USDT

	// 方式 6: 同时启动钱包监控和 Meme 币监控
	// go startAddressMonitor()    // 在 goroutine 中启动地址监控
	// monitor.StartMemeMonitor()  // 启动 Meme 币监控

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

// startAddressMonitor 启动地址监控（支持多地址）
func startAddressMonitor() {
	// 配置要监控的地址
	addresses := map[string]string{
		config.BinanceWalletAddress: "币安",
		config.OkxWalletAddress:     "OKX",
	}

	// 创建监控器
	addressMonitor, err := wallet.NewAddressMonitor(
		config.GetEthereumRpcUrl(),
		config.GetEthereumWsUrl(),
		addresses,
	)
	if err != nil {
		logger.Log.Fatal("创建地址监控器失败: " + err.Error())
	}
	defer addressMonitor.Close()

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 监听系统信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动监控
	go func() {
		if err := addressMonitor.Start(ctx); err != nil {
			logger.Log.Error("地址监控错误: " + err.Error())
		}
	}()

	// 等待退出信号
	<-sigChan
	logger.Log.Info("收到退出信号，正在停止监控...")
	cancel()
}
