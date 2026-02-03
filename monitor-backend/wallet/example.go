package wallet

import (
	"context"
	"ethereum-monitor/config"

	"github.com/ethereum/go-ethereum/common"
)

// StartGoEthMonitor 启动 go-ethereum 监控器（WebSocket + 轮询双模式）
func StartGoEthMonitor(ctx context.Context) error {
	// 配置监控参数
	monitorConfig := &MonitorConfig{
		Addresses: map[string]string{
			config.OkxWalletAddress: "OKX钱包",
			// 可以添加更多地址
		},
		Tokens: []TokenConfig{
			{
				Address:  common.HexToAddress("0xdac17f958d2ee523a2206206994597c13d831ec7"), // USDT
				Symbol:   "USDT",
				Decimals: 6,
			},
			{
				Address:  common.HexToAddress("0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48"), // USDC
				Symbol:   "USDC",
				Decimals: 6,
			},
		},
		ETHThreshold:   CreateETHThreshold(int64(config.EthThreshold)),
		TokenThreshold: CreateTokenThreshold(int64(config.UsdtThreshold), 6), // USDT/USDC 都是 6 位小数
	}

	// 创建监控器
	monitor, err := NewGoEthMonitor(
		config.GetEthereumRpcUrl(),
		config.GetEthereumWsUrl(),
		monitorConfig,
	)
	if err != nil {
		return err
	}
	defer monitor.Close()

	// 启动监控
	return monitor.Start(ctx)
}

// StartWatcherMonitor 启动 ethereum-watcher 监控器（HTTP 轮询）
func StartWatcherMonitor(ctx context.Context) error {
	// 配置监控参数
	monitorConfig := &MonitorConfig{
		Addresses: map[string]string{
			config.OkxWalletAddress: "OKX钱包",
		},
		Tokens: []TokenConfig{
			{
				Address:  common.HexToAddress("0xdac17f958d2ee523a2206206994597c13d831ec7"), // USDT
				Symbol:   "USDT",
				Decimals: 6,
			},
		},
		ETHThreshold:   CreateETHThreshold(int64(config.EthThreshold)),
		TokenThreshold: CreateTokenThreshold(int64(config.UsdtThreshold), 6),
	}

	// 创建监控器
	monitor, err := NewWatcherMonitor(
		config.GetEthereumRpcUrl(),
		monitorConfig,
	)
	if err != nil {
		return err
	}
	defer monitor.Close()

	// 启动监控
	return monitor.Start(ctx, config.GetEthereumRpcUrl(), config.SleepSecondsForNewBlock)
}
