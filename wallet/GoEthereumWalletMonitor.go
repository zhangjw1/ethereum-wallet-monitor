package wallet

import (
	"context"
	"etherum-monitor/config"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// GoEthereumWalletMonitor 使用官方go-ethereum库的以太坊钱包监控器
type GoEthereumWalletMonitor struct {
	client *ethclient.Client
}

// NewGoEthereumWalletMonitor 创建新的go-ethereum钱包监控器实例
func NewGoEthereumWalletMonitor(rpcUrl string) (*GoEthereumWalletMonitor, error) {
	client, err := ethclient.Dial(rpcUrl)
	if err != nil {
		return nil, fmt.Errorf("连接以太坊节点失败: %v", err)
	}

	return &GoEthereumWalletMonitor{
		client: client,
	}, nil
}

// GetBalance 获取指定地址的余额
func (g *GoEthereumWalletMonitor) GetBalance(address string) (*big.Float, error) {
	addr := common.HexToAddress(address)
	balance, err := g.client.BalanceAt(context.Background(), addr, nil)
	if err != nil {
		return nil, fmt.Errorf("获取余额失败: %v", err)
	}

	// 将wei转换为ether
	balanceInEth := new(big.Float).SetInt(balance)
	balanceInEth.Quo(balanceInEth, big.NewFloat(1e18))

	return balanceInEth, nil
}

// GetTransactionByHash 根据交易哈希获取交易信息
func (g *GoEthereumWalletMonitor) GetTransactionByHash(hash common.Hash) (*types.Transaction, bool, error) {
	tx, pending, err := g.client.TransactionByHash(context.Background(), hash)
	if err != nil {
		return nil, false, fmt.Errorf("获取交易失败: %v", err)
	}

	return tx, pending, nil
}

// FilterLogs 过滤特定地址的事件日志
func (g *GoEthereumWalletMonitor) FilterLogs(address string) error {
	addr := common.HexToAddress(address)
	query := ethereum.FilterQuery{
		Addresses: []common.Address{addr},
	}

	logs, err := g.client.FilterLogs(context.Background(), query)
	if err != nil {
		return fmt.Errorf("过滤日志失败: %v", err)
	}

	fmt.Printf("找到 %d 条日志记录\n", len(logs))
	for _, vLog := range logs {
		fmt.Printf("区块号: %d, 交易索引: %d\n", vLog.BlockNumber, vLog.TxIndex)
	}

	return nil
}

// SubscribeNewHead 订阅新区块头
func (g *GoEthereumWalletMonitor) SubscribeNewHead() error {
	headers := make(chan *types.Header)
	sub, err := g.client.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		return fmt.Errorf("订阅新区块失败: %v", err)
	}
	defer sub.Unsubscribe()

	fmt.Println("开始监听新区块...")
	for {
		select {
		case err := <-sub.Err():
			return fmt.Errorf("订阅错误: %v", err)
		case header := <-headers:
			fmt.Printf("新块到达: #%d - %s\n", header.Number, header.Hash().Hex())

			// 在这里可以添加处理新块的逻辑
			g.handleNewBlock(header)
		}
	}
}

// handleNewBlock 处理新块到达时的逻辑
func (g *GoEthereumWalletMonitor) handleNewBlock(header *types.Header) {
	fmt.Printf("处理新块 #%d\n", header.Number)

	// 获取块中的交易数量
	block, err := g.client.BlockByHash(context.Background(), header.Hash())
	if err != nil {
		log.Printf("获取块失败: %v", err)
		return
	}

	fmt.Printf("块中有 %d 笔交易\n", len(block.Transactions()))

	// 检查块中的交易是否涉及目标地址
	g.checkTransactionsForTargetAddress(block)
}

// checkTransactionsForTargetAddress 检查块中的交易是否涉及目标地址
func (g *GoEthereumWalletMonitor) checkTransactionsForTargetAddress(block *types.Block) {
	targetAddress := config.OKX_WALLET_ADDRESS
	addr := common.HexToAddress(targetAddress)

	for i, tx := range block.Transactions() {
		// 检查to地址
		if tx.To() != nil && *tx.To() == addr {
			fmt.Printf("发现目标地址接收交易 - 块号: %d, 交易索引: %d, 交易哈希: %s\n",
				block.NumberU64(), i, tx.Hash().Hex())
		}

		// 如果交易的from地址是我们监控的地址，也需要关注
		// 注意：需要从签名恢复发送方地址
		signer := types.LatestSignerForChainID(tx.ChainId())
		from, err := types.Sender(signer, tx)
		if err == nil && from == addr {
			fmt.Printf("发现目标地址发送交易 - 块号: %d, 交易索引: %d, 交易哈希: %s\n",
				block.NumberU64(), i, tx.Hash().Hex())
		}
	}
}

// Close 关闭客户端连接
func (g *GoEthereumWalletMonitor) Close() {
	g.client.Close()
}

// GoEthereumAddressAddMonitor 使用官方go-ethereum库的新监控函数
func GoEthereumAddressAddMonitor() {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("🚀 使用官方go-ethereum库的以太坊钱包监控程序启动")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	monitor, err := NewGoEthereumWalletMonitor(config.GetEthereumRpcUrl())
	if err != nil {
		fmt.Printf("❌ 创建监控器失败: %v\n", err)
		return
	}
	defer monitor.Close()

	fmt.Printf("✅ 成功连接到以太坊节点: %s\n", config.GetEthereumRpcUrl())

	// 获取目标地址余额
	balance, err := monitor.GetBalance(config.OKX_WALLET_ADDRESS)
	if err != nil {
		fmt.Printf("⚠️  获取余额失败: %v\n", err)
	} else {
		fmt.Printf("💰 目标地址余额: %s ETH\n", balance.Text('f', 6))
	}

	// 开始监听新区块
	if err := monitor.SubscribeNewHead(); err != nil {
		fmt.Printf("❌ 监听新区块失败: %v\n", err)
		return
	}
}
