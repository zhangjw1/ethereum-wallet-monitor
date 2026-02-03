package analyzer

import (
	"context"
	"ethereum-monitor/config"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// LiquidityAnalyzer 流动性分析器
type LiquidityAnalyzer struct {
	client *ethclient.Client
}

// NewLiquidityAnalyzer 创建流动性分析器
func NewLiquidityAnalyzer(rpcURL string) (*LiquidityAnalyzer, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to dial rpc: %w", err)
	}
	return &LiquidityAnalyzer{client: client}, nil
}

// PairReserves 交易对储备量
type PairReserves struct {
	Reserve0       *big.Int
	Reserve1       *big.Int
	BlockTimestamp uint32
}

// GetReserves 读取 Uniswap Pair 的储备量
func (la *LiquidityAnalyzer) GetReserves(pairAddress string) (*PairReserves, error) {
	addr := common.HexToAddress(pairAddress)
	// getReserves() signature: 0902f1ac
	// returns (uint112 reserve0, uint112 reserve1, uint32 blockTimestampLast)
	data := common.Hex2Bytes("0902f1ac")

	msg := ethereum.CallMsg{
		To:   &addr,
		Data: data,
	}

	result, err := la.client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return nil, err
	}

	if len(result) < 64 { // 至少要有两个 32 字节的数据
		return nil, fmt.Errorf("invalid getReserves response")
	}

	// 解析返回数据
	// Solidity 返回的是 3 个 32字节的 word (尽管 uint112 只有 14 字节，但也是按 32 字节对齐返回的)
	reserve0 := new(big.Int).SetBytes(result[0:32])
	reserve1 := new(big.Int).SetBytes(result[32:64])

	// timestamp 可能是第三个 word，也可能不是，这里暂时不需要 timestamp

	return &PairReserves{
		Reserve0: reserve0,
		Reserve1: reserve1,
	}, nil
}

// GetLiquidityInfo 获取流动性详情（包含 USD 估值）
func (la *LiquidityAnalyzer) GetLiquidityInfo(pairAddress, tokenAddress string) (liquidityUSD float64, ethAmount float64, err error) {
	reserves, err := la.GetReserves(pairAddress)
	if err != nil {
		return 0, 0, err
	}

	// 我们需要知道 tokenAddress 是 token0 还是 token1
	// 为此我们需要查询 Pair 的 token0() 方法，但这会增加 RPC 调用
	// 更好的方法是：我们在 PairCreatedPlugin 里就知道谁是 token0/token1，并存在 DB 里。
	// 但现在 DB 只存了 tokenAddress。
	// 这里我们做一个假设：Pair 里必然有一个是 WETH。
	// 我们可以调用 pair.token0() 来确认。

	token0Addr, err := la.getToken0(pairAddress)
	if err != nil {
		return 0, 0, err
	}

	weth := strings.ToLower(config.WETHAddress)
	var wethReserve *big.Int

	// 判断哪个 reserve 是 WETH 的
	if strings.ToLower(token0Addr) == weth {
		wethReserve = reserves.Reserve0
	} else {
		// 如果 token0 不是 WETH，我们就假设 token1 是 WETH
		// (因为我们在 PairCreated 处理时只过滤了含 WETH 的池子)
		wethReserve = reserves.Reserve1
	}

	// WETH 是 18 位精度
	ethVal := new(big.Float).SetInt(wethReserve)
	ethDiv := new(big.Float).SetFloat64(1e18)
	ethCount, _ := new(big.Float).Quo(ethVal, ethDiv).Float64()

	// 假定 ETH 价格 $2500 (后续可优化)
	// 池子总价值 = WETH 价值 * 2 (因为恒定乘积做市，两边价值理应相等)
	// 但通常我们在说 Liquidity 时候，指的是 USDT 计价的总池子深度
	// 这里 liquidityUSD = ethCount * 2500 * 2

	// 注意：有些工具显示 Liquidity 仅指 ETH 这一侧的价值，有些指双侧。
	// 这里我们按 **双侧总价值** 计算。
	ethPrice := 2500.0 // Mock Price
	liquidityUSD = ethCount * ethPrice * 2

	return liquidityUSD, ethCount, nil
}

// getToken0 读取 token0 地址
func (la *LiquidityAnalyzer) getToken0(pairAddress string) (string, error) {
	// token0() signature: 0dfe1681
	data := common.Hex2Bytes("0dfe1681")
	msg := ethereum.CallMsg{
		To:   &common.Address{}, // 稍后设置
		Data: data,
	}
	// Copy address to avoid pointer issues if reused? No need.
	addr := common.HexToAddress(pairAddress)
	msg.To = &addr

	result, err := la.client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return "", err
	}
	if len(result) < 32 {
		return "", fmt.Errorf("invalid token0 response")
	}

	// Address is last 20 bytes of 32-byte word
	return common.BytesToAddress(result).Hex(), nil
}

// Close 关闭
func (la *LiquidityAnalyzer) Close() {
	la.client.Close()
}
