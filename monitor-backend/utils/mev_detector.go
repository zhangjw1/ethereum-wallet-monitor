package utils

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// MevDetector MEV 攻击检测器
type MevDetector struct {
	client *ethclient.Client
}

// MevType MEV 攻击类型
type MevType string

const (
	MevTypeSandwich    MevType = "三明治攻击"
	MevTypeFrontRun    MevType = "抢跑交易"
	MevTypeBackRun     MevType = "尾随交易"
	MevTypeLiquidation MevType = "清算攻击"
	MevTypeHighGas     MevType = "异常高Gas"
	MevTypeNone        MevType = "正常交易"
)

// MevDetectionResult MEV 检测结果
type MevDetectionResult struct {
	IsMev       bool     // 是否是 MEV 攻击
	MevType     MevType  // MEV 类型
	Confidence  float64  // 置信度 (0-1)
	Description string   // 详细描述
	Evidence    []string // 证据列表
}

// NewMevDetector 创建 MEV 检测器
func NewMevDetector(rpcUrl string) (*MevDetector, error) {
	client, err := ethclient.Dial(rpcUrl)
	if err != nil {
		return nil, err
	}
	return &MevDetector{client: client}, nil
}

// DetectMev 检测交易是否为 MEV 攻击
func (m *MevDetector) DetectMev(txHash string) (*MevDetectionResult, error) {
	hash := common.HexToHash(txHash)
	tx, pending, err := m.client.TransactionByHash(context.Background(), hash)
	if err != nil {
		return nil, err
	}

	if pending {
		return &MevDetectionResult{
			IsMev:       false,
			MevType:     MevTypeNone,
			Confidence:  0,
			Description: "交易待确认，无法判断",
		}, nil
	}

	receipt, err := m.getReceiptWithRetry(hash)
	if err != nil {
		return nil, err
	}

	result := &MevDetectionResult{
		IsMev:    false,
		MevType:  MevTypeNone,
		Evidence: []string{},
	}

	// 检测各种 MEV 特征
	m.checkKnownMevBots(tx, result) // 优先检查已知 Bot
	m.checkHighGasPrice(tx, result)
	m.checkSandwichAttack(tx, receipt, result)
	m.checkFrontRunning(tx, receipt, result)
	m.checkInternalTransfers(receipt, result) // 检查内部转账
	m.checkFailedButExecuted(receipt, result) // 检查失败但执行的交易

	return result, nil
}

func (m *MevDetector) getReceiptWithRetry(hash common.Hash) (*types.Receipt, error) {
	const maxAttempts = 3
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		receipt, err := m.client.TransactionReceipt(context.Background(), hash)
		if err == nil {
			return receipt, nil
		}

		// Some RPC providers intermittently return truncated JSON; retry briefly.
		if !strings.Contains(err.Error(), "unexpected end of JSON input") {
			return nil, err
		}

		if attempt < maxAttempts {
			time.Sleep(time.Duration(attempt) * 300 * time.Millisecond)
			continue
		}
		return nil, err
	}

	return nil, fmt.Errorf("GetTransactionReceipt failed after retries")
}

// checkHighGasPrice 检测异常高的 Gas Price
func (m *MevDetector) checkHighGasPrice(tx *types.Transaction, result *MevDetectionResult) {
	gasPrice := tx.GasPrice()
	if gasPrice == nil {
		return
	}

	// 如果 Gas Price 超过 500 Gwei，可能是 MEV
	highGasThreshold := big.NewInt(500000000000) // 500 Gwei
	if gasPrice.Cmp(highGasThreshold) > 0 {
		result.IsMev = true
		result.MevType = MevTypeHighGas
		result.Confidence = 0.6
		result.Evidence = append(result.Evidence,
			"Gas Price 异常高: "+weiToGwei(gasPrice)+" Gwei")
	}
}

// checkSandwichAttack 检测三明治攻击
func (m *MevDetector) checkSandwichAttack(tx *types.Transaction, receipt *types.Receipt, result *MevDetectionResult) {
	// 三明治攻击特征：
	// 1. 在同一个区块内
	// 2. 有相同的交易对
	// 3. 前后各有一笔交易来自同一地址

	blockNumber := receipt.BlockNumber
	block, err := m.client.BlockByNumber(context.Background(), blockNumber)
	if err != nil {
		return
	}

	txIndex := receipt.TransactionIndex
	transactions := block.Transactions()

	if txIndex == 0 || int(txIndex) >= len(transactions)-1 {
		return
	}

	// 检查前一笔和后一笔交易
	prevTx := transactions[txIndex-1]
	nextTx := transactions[txIndex+1]

	prevFrom, _ := types.Sender(types.LatestSignerForChainID(prevTx.ChainId()), prevTx)
	nextFrom, _ := types.Sender(types.LatestSignerForChainID(nextTx.ChainId()), nextTx)

	// 如果前后交易来自同一地址，且都与目标交易交互相同合约
	if prevFrom == nextFrom && prevTx.To() != nil && nextTx.To() != nil && tx.To() != nil {
		if *prevTx.To() == *tx.To() && *nextTx.To() == *tx.To() {
			result.IsMev = true
			result.MevType = MevTypeSandwich
			result.Confidence = 0.8
			result.Evidence = append(result.Evidence,
				"检测到三明治攻击模式: 前后交易来自同一地址 "+prevFrom.Hex())
		}
	}
}

// checkFrontRunning 检测抢跑交易
func (m *MevDetector) checkFrontRunning(tx *types.Transaction, receipt *types.Receipt, result *MevDetectionResult) {
	// 抢跑特征：
	// 1. Gas Price 明显高于平均值
	// 2. 在目标交易之前执行
	// 3. 与目标交易交互相同合约

	blockNumber := receipt.BlockNumber
	block, err := m.client.BlockByNumber(context.Background(), blockNumber)
	if err != nil {
		return
	}

	txIndex := receipt.TransactionIndex
	if txIndex == 0 {
		return
	}

	// 计算区块内平均 Gas Price
	var totalGasPrice big.Int
	txCount := 0
	for _, blockTx := range block.Transactions() {
		if blockTx.GasPrice() != nil {
			totalGasPrice.Add(&totalGasPrice, blockTx.GasPrice())
			txCount++
		}
	}

	if txCount == 0 {
		return
	}

	avgGasPrice := new(big.Int).Div(&totalGasPrice, big.NewInt(int64(txCount)))
	txGasPrice := tx.GasPrice()

	if txGasPrice == nil {
		return
	}

	// 如果 Gas Price 是平均值的 2 倍以上
	threshold := new(big.Int).Mul(avgGasPrice, big.NewInt(2))
	if txGasPrice.Cmp(threshold) > 0 {
		result.IsMev = true
		result.MevType = MevTypeFrontRun
		result.Confidence = 0.7
		result.Evidence = append(result.Evidence,
			"Gas Price 是区块平均值的 2 倍以上")
	}
}

// checkKnownMevBots 检测已知的 MEV Bot 地址
func (m *MevDetector) checkKnownMevBots(tx *types.Transaction, result *MevDetectionResult) {
	// 已知的 MEV Bot 地址列表（更新和扩展）
	knownMevBots := map[string]string{
		"0xa69babef1ca67a37ffaf7a485dfff3382056e78c": "Flashbots",
		"0x00000000000007736e2f9af06b8f5f3b6d0e8f13": "MEV Bot",
		"0x000000000000084e91743124a982076c59f10084": "Sandwich Bot",
		"0xd2269f890854a8c5f03e8ea091e3d5a2e0e0f890": "MEV Bot",
		"0x6b75d8af000000e20b7a7ddf000ba900b4009a80": "MEV Bot",
		"0x51c72848c68a965f66fa7a88855f9f7784502a7f": "jaredfromsubway.eth",
		"0x00000000003b3cc22af3ae1eac0440bcee416b40": "MEV Bot",
	}

	from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
	if err != nil {
		return
	}

	fromAddr := strings.ToLower(from.Hex())

	// 检查完整地址匹配
	for botAddr, botName := range knownMevBots {
		if strings.ToLower(botAddr) == fromAddr {
			result.IsMev = true
			result.Confidence = 0.95
			result.Evidence = append(result.Evidence,
				"交易来自已知 MEV Bot: "+botName)
			return
		}
	}

	// 检查地址前缀模式（MEV Bot 常用模式）
	if strings.HasPrefix(fromAddr, "0x000000000000") ||
		strings.HasPrefix(fromAddr, "0x00000000") {
		result.IsMev = true
		result.Confidence = 0.75
		result.Evidence = append(result.Evidence,
			"地址符合 MEV Bot 特征模式（前缀为多个0）")
	}

	// 检查 to 地址
	if tx.To() != nil {
		toAddr := strings.ToLower(tx.To().Hex())
		for botAddr, botName := range knownMevBots {
			if strings.ToLower(botAddr) == toAddr {
				result.IsMev = true
				result.Confidence = 0.9
				result.Evidence = append(result.Evidence,
					"交易发送到已知 MEV Bot: "+botName)
				return
			}
		}
	}
}

// weiToGwei 将 Wei 转换为 Gwei
func weiToGwei(wei *big.Int) string {
	gwei := new(big.Float).SetInt(wei)
	divisor := new(big.Float).SetFloat64(1e9)
	result := new(big.Float).Quo(gwei, divisor)
	return result.Text('f', 2)
}

// Close 关闭客户端连接
func (m *MevDetector) Close() {
	m.client.Close()
}

// checkInternalTransfers 检测异常的内部转账模式
func (m *MevDetector) checkInternalTransfers(receipt *types.Receipt, result *MevDetectionResult) {
	// MEV 攻击通常涉及多次内部转账
	// 检查日志中的 Transfer 事件数量
	transferCount := 0
	for _, log := range receipt.Logs {
		// Transfer 事件的 topic[0]
		if len(log.Topics) > 0 &&
			log.Topics[0].Hex() == "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef" {
			transferCount++
		}
	}

	// 如果有 3 次以上的 Transfer 事件，可能是三明治攻击
	if transferCount >= 3 {
		result.IsMev = true
		if result.MevType == MevTypeNone {
			result.MevType = MevTypeSandwich
		}
		result.Confidence = 0.7
		result.Evidence = append(result.Evidence,
			fmt.Sprintf("检测到 %d 次 Transfer 事件，符合 MEV 攻击模式", transferCount))
	}
}

// checkFailedButExecuted 检测失败但仍执行的交易（MEV 常见特征）
func (m *MevDetector) checkFailedButExecuted(receipt *types.Receipt, result *MevDetectionResult) {
	// 如果交易失败（status = 0）但仍然消耗了大量 Gas
	if receipt.Status == 0 {
		gasUsed := receipt.GasUsed
		// 如果消耗了超过 100,000 Gas 但失败了，可能是 MEV 尝试
		if gasUsed > 100000 {
			result.IsMev = true
			result.Confidence = 0.65
			result.Evidence = append(result.Evidence,
				fmt.Sprintf("交易失败但消耗了 %d Gas，可能是 MEV 攻击尝试", gasUsed))
		}
	}
}
