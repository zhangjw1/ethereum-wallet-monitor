package main

import (
	"ethereum-monitor/database"
	"ethereum-monitor/logger"
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

	// MEV 检测示例
	utils.ExampleUsage()

	// TODO: 启动实际的监控服务
	// wallet.AddressAddMonitor()
}
