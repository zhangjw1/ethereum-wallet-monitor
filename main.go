package main

import (
	"etherum-monitor/logger"
	"etherum-monitor/wallet"
	"fmt"
	"log"

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

	fmt.Println("选择要运行的监控模式:")
	fmt.Println("1. Hydro Protocol 监控")
	fmt.Println("2. 官方 Go-Ethereum 监控")
	fmt.Print("请输入选择 (1 或 2): ")

	var choice int
	fmt.Scanf("%d", &choice)

	switch choice {
	case 1:
		fmt.Println("\n启动 Hydro Protocol 监控...")
		wallet.AddressAddMonitor()
	case 2:
		fmt.Println("\n启动官方 Go-Ethereum 监控...")
		wallet.GoEthereumAddressAddMonitor()
	default:
		fmt.Println("无效选择，默认启动 Hydro Protocol 监控")
		wallet.AddressAddMonitor()
	}
}
