package config

import (
	"fmt"
	"os"
)

const (
	OkxWalletAddress     = "0x6ea08ca8f313d860808ef7431fc72c6fbcf4a72d" // OKX 钱包地址
	BinanceWalletAddress = "0xf91773ceef22691a825b47a3f14fd68c1d876adf" // 币安买币地址

	UsdtContractAddress = "0xdac17f958d2ee523a2206206994597c13d831ec7"

	// 以太坊平均出块时间约 12 秒，设置为 20 秒可以减少更多请求
	SleepSecondsForNewBlock = 20

	EthThreshold = 10

	UsdtThreshold = 500000
	// USDT 转账事件
	UsdtTransferTopic = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
)

// GetEthereumRpcUrl 从环境变量获取 Infura Key 并构建 RPC URL
func GetEthereumRpcUrl() string {
	infuraKey := os.Getenv("INFURA_KEY")

	if infuraKey == "" {
		// 如果没有设置环境变量，使用公共节点
		// 注意：公共节点可能不稳定，建议使用 Infura 或 Alchemy
		fmt.Println("⚠️  未设置 INFURA_KEY 环境变量，使用公共 RPC 节点")
		fmt.Println("⚠️  公共节点可能不稳定，建议设置 INFURA_KEY 或 ALCHEMY_KEY")

		// 尝试使用 Alchemy 公共端点
		if alchemyKey := os.Getenv("ALCHEMY_KEY"); alchemyKey != "" {
			fmt.Println("✓ 使用 Alchemy RPC 节点")
			return fmt.Sprintf("https://eth-mainnet.g.alchemy.com/v2/%s", alchemyKey)
		}

		// 备用公共节点列表（按优先级）
		publicNodes := []string{
			"https://eth.llamarpc.com",
			"https://rpc.ankr.com/eth",
			"https://ethereum.publicnode.com",
			"https://1rpc.io/eth",
		}

		// 使用第一个公共节点
		fmt.Printf("使用公共节点: %s\n", publicNodes[0])
		return publicNodes[0]
	}
	fmt.Println("✓ 使用 Infura RPC 节点")
	return fmt.Sprintf("https://mainnet.infura.io/v3/%s", infuraKey)
}

// GetEthereumWsUrl 获取 WebSocket URL
func GetEthereumWsUrl() string {
	infuraKey := os.Getenv("INFURA_KEY")

	if infuraKey == "" {
		return "" // 没有 WebSocket，使用轮询模式
	}
	return fmt.Sprintf("wss://mainnet.infura.io/ws/v3/%s", infuraKey)
}
