package monitor

import (
	"context"
	"ethereum-monitor/config"
	"ethereum-monitor/logger"
	"ethereum-monitor/utils"
	"os"

	ethereum "github.com/HydroProtocol/ethereum-watcher"
	"go.uber.org/zap"
)

// StartMemeMonitor 启动 Meme 币监控
func StartMemeMonitor() error {
	logger.Log.Info("🚀 Meme 币监控启动")

	// 必须在创建 watcher 之前设置代理
	// 从环境变量读取代理配置
	if proxyURL := os.Getenv("HTTP_PROXY"); proxyURL != "" {
		logger.Log.Info("设置代理", zap.String("proxy", proxyURL))
		if err := utils.SetGlobalProxy(proxyURL); err != nil {
			logger.Log.Error("设置代理失败", zap.Error(err))
		}
	} else {
		// 如果没有配置环境变量，使用默认代理
		logger.Log.Info("使用默认代理: http://127.0.0.1:7890")
		utils.SetGlobalProxy("http://127.0.0.1:7890")
	}

	// 创建 PairCreated 事件监听插件
	pairCreatedPlugin, err := NewPairCreatedPlugin(config.GetEthereumRpcUrl())
	if err != nil {
		logger.Log.Fatal("创建 PairCreated 插件失败", zap.Error(err))
		return err
	}
	defer pairCreatedPlugin.Close()

	// 创建 Watcher
	logger.Log.Info("正在创建 Watcher...")
	watcher := ethereum.NewHttpBasedEthWatcher(context.Background(), config.GetEthereumRpcUrl())

	// 设置轮询间隔（秒）
	watcher.SetSleepSecondsForNewBlock(config.SleepSecondsForNewBlock)
	logger.Log.Info("配置完成",
		zap.Int("pollInterval", config.SleepSecondsForNewBlock))

	// 注册 PairCreated 事件监听插件
	watcher.RegisterReceiptLogPlugin(pairCreatedPlugin)
	logger.Log.Info("✅ Uniswap PairCreated 事件监听插件已注册",
		zap.String("factory", config.UniswapV2FactoryAddress),
		zap.String("topic", config.UniswapV2PairCreatedTopic))

	logger.Log.Info("⏳ 开始监听新区块...")
	logger.Log.Info("💡 提示：")
	logger.Log.Info("   - 监听 Uniswap 新交易对创建事件")
	logger.Log.Info("   - 检测到新 ETH 交易对时会自动记录")
	logger.Log.Info("   - 新代币信息会保存到数据库")

	// 运行监听器
	err = watcher.RunTillExit()
	if err != nil {
		logger.Log.Error("监听器运行错误", zap.Error(err))
		return err
	}

	return nil
}
