package config

import (
	"fmt"
	"os"
)

const (
	OKX_WALLET_ADDRESS = "0x6ea08ca8f313d860808ef7431fc72c6fbcf4a72d"

	USDT_CONTRACT_ADDRESS = "0xdac17f958d2ee523a2206206994597c13d831ec7"

	SLEEP_SECONDS_FOR_NEW_BLOCK = 10

	ETH_THRESHOLD = 10

	USDT_THRESHOLD = 10000
	// USDT 转账事件
	USDT_TRANSFER_TOPIC = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
)

// GetEthereumRpcUrl 从环境变量获取 Infura Key 并构建 RPC URL
func GetEthereumRpcUrl() string {
	infuraKey := os.Getenv("INFURA_KEY")

	if infuraKey == "" {
		// 如果没有设置环境变量，使用公共节点
		fmt.Println("⚠️  未设置 INFURA_KEY 环境变量，使用公共 RPC 节点")
		return "https://eth.llamarpc.com"
	}
	fmt.Println("✓ 使用 Infura RPC 节点")
	return fmt.Sprintf("https://mainnet.infura.io/v3/%s", infuraKey)
}
