package wallet

import (
	"context"
	"ethereum-monitor/database"
	"ethereum-monitor/logger"
	"ethereum-monitor/model"
	"ethereum-monitor/utils"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

// TransactionEvent 交易事件
type TransactionEvent struct {
	Direction string
	From      string
	To        string
	Amount    string
	Currency  string
	TxHash    string
	BlockNum  int
}

// AddressMonitor 地址监控器（基于 Go-Ethereum）
type AddressMonitor struct {
	client        *ethclient.Client
	wsClient      *ethclient.Client
	addressLabels map[common.Address]string
	addressSet    map[common.Address]struct{}
	mevDetector   *utils.MevDetector
	pushPlus      *utils.PushPlusNotifier
	wechatRepo    *database.WechatAlterRepository
	usdcContract  common.Address
	transferTopic common.Hash

	// 异步处理通道
	blockChan chan *types.Block
	txChan    chan *TransactionEvent
	workerNum int
}

// NewAddressMonitor 创建地址监控器
func NewAddressMonitor(rpcURL, wsURL string, addresses map[string]string) (*AddressMonitor, error) {
	// HTTP 客户端（用于查询）
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("连接 RPC 失败: %w", err)
	}

	// WebSocket 客户端（用于订阅）
	var wsClient *ethclient.Client
	if wsURL != "" {
		wsClient, err = ethclient.Dial(wsURL)
		if err != nil {
			logger.Warn("WebSocket 连接失败，将使用轮询模式", zap.Error(err))
		}
	}

	// MEV 检测器
	mevDetector, err := utils.NewMevDetector(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("创建 MEV 检测器失败: %w", err)
	}

	// PushPlus 通知器
	var pushPlus *utils.PushPlusNotifier
	if token := os.Getenv("PUSHPLUS_TOKEN"); token != "" {
		pushPlus = utils.NewPushPlusNotifier(token)
	}

	// USDC 合约地址
	usdcContract := common.HexToAddress("0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48")

	// Transfer 事件签名
	transferTopic := common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")

	addressLabels := make(map[common.Address]string, len(addresses))
	addressSet := make(map[common.Address]struct{}, len(addresses))
	for addr, label := range addresses {
		parsed := common.HexToAddress(addr)
		addressLabels[parsed] = label
		addressSet[parsed] = struct{}{}
	}

	return &AddressMonitor{
		client:        client,
		wsClient:      wsClient,
		addressLabels: addressLabels,
		addressSet:    addressSet,
		mevDetector:   mevDetector,
		pushPlus:      pushPlus,
		wechatRepo:    database.NewWechatAlterRepository(),
		usdcContract:  usdcContract,
		transferTopic: transferTopic,
	}, nil
}

// Start 启动监控
func (m *AddressMonitor) Start(ctx context.Context) error {
	logger.Info("🚀 启动地址监控",
		zap.Int("address_count", len(m.addressSet)),
		zap.Strings("addresses", m.addressLabelList()),
		zap.Bool("websocket", m.wsClient != nil))

	if m.wsClient != nil {
		// 使用 WebSocket 实时订阅
		return m.startWebSocketMonitor(ctx)
	} else {
		// 使用轮询模式
		return m.startPollingMonitor(ctx)
	}
}

// startWebSocketMonitor WebSocket 实时监控
func (m *AddressMonitor) startWebSocketMonitor(ctx context.Context) error {
	// 订阅新区块头
	headers := make(chan *types.Header)
	sub, err := m.wsClient.SubscribeNewHead(ctx, headers)
	if err != nil {
		return fmt.Errorf("订阅区块失败: %w", err)
	}
	defer sub.Unsubscribe()

	logger.Info("✅ WebSocket 订阅成功，开始实时监控...")

	for {
		select {
		case err := <-sub.Err():
			logger.Error("订阅错误", zap.Error(err))
			return err
		case header := <-headers:
			// 获取完整区块
			block, err := m.client.BlockByHash(ctx, header.Hash())
			if err != nil {
				logger.Error("获取区块失败", zap.Error(err))
				continue
			}

			// 检查区块中的交易
			m.checkBlockTransactions(ctx, block)

		case <-ctx.Done():
			logger.Info("监控已停止")
			return nil
		}
	}
}

// startPollingMonitor 轮询模式监控
func (m *AddressMonitor) startPollingMonitor(ctx context.Context) error {
	logger.Info("使用轮询模式监控...")

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	var lastBlock uint64

	for {
		select {
		case <-ticker.C:
			// 获取最新区块号
			header, err := m.client.HeaderByNumber(ctx, nil)
			if err != nil {
				logger.Error("获取区块头失败", zap.Error(err))
				continue
			}

			currentBlock := header.Number.Uint64()
			if lastBlock == 0 {
				lastBlock = currentBlock
				continue
			}

			// 检查新区块
			for blockNum := lastBlock + 1; blockNum <= currentBlock; blockNum++ {
				block, err := m.client.BlockByNumber(ctx, big.NewInt(int64(blockNum)))
				if err != nil {
					logger.Error("获取区块失败", zap.Uint64("block", blockNum), zap.Error(err))
					continue
				}

				m.checkBlockTransactions(ctx, block)
			}

			lastBlock = currentBlock

		case <-ctx.Done():
			logger.Info("监控已停止")
			return nil
		}
	}
}

// checkBlockTransactions 检查区块中的交易
func (m *AddressMonitor) checkBlockTransactions(ctx context.Context, block *types.Block) {
	for _, tx := range block.Transactions() {
		// 检查是否与目标地址相关
		if m.isRelatedTransaction(tx) {
			m.handleTransaction(ctx, tx, block.Number().Uint64())
		}
	}

	// 检查 ERC20 Transfer 事件
	m.checkERC20Transfers(ctx, block)
}

// isRelatedTransaction 判断交易是否与目标地址相关
func (m *AddressMonitor) isRelatedTransaction(tx *types.Transaction) bool {
	// 检查接收方
	if tx.To() != nil && m.isMonitoredAddress(*tx.To()) {
		return true
	}

	// 检查发送方
	msg, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
	if err == nil && m.isMonitoredAddress(msg) {
		return true
	}

	return false
}

// handleTransaction 处理交易
func (m *AddressMonitor) handleTransaction(ctx context.Context, tx *types.Transaction, blockNum uint64) {
	txHash := tx.Hash().Hex()

	// 检查是否已处理
	if m.wechatRepo.ExistsByTxHash(txHash) {
		return
	}

	// 判断方向与归属地址
	from, _ := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
	var to common.Address
	var toHex string
	if tx.To() != nil {
		to = *tx.To()
		toHex = to.Hex()
	}

	fromMonitored := m.isMonitoredAddress(from)
	toMonitored := tx.To() != nil && m.isMonitoredAddress(to)

	direction := "转入"
	targetLabel := ""
	if fromMonitored {
		direction = "转出"
		targetLabel = m.getAddressLabel(from)
	} else if toMonitored {
		targetLabel = m.getAddressLabel(to)
	}

	// 计算金额
	ethAmount := new(big.Float).Quo(new(big.Float).SetInt(tx.Value()), big.NewFloat(1e18))
	amountStr := ethAmount.Text('f', 6)

	logger.Info("🔔 检测到 ETH 交易",
		zap.String("direction", direction),
		zap.String("from", from.Hex()),
		zap.String("to", toHex),
		zap.String("amount", amountStr+" ETH"),
		zap.String("tx", txHash),
		zap.String("label", targetLabel))

	// MEV 检测
	if m.mevDetector != nil {
		mevResult, err := m.mevDetector.DetectMev(txHash)
		if err == nil && mevResult.IsMev {
			logger.Info("检测到 MEV 交易，跳过通知",
				zap.String("type", string(mevResult.MevType)),
				zap.Float64("confidence", mevResult.Confidence))
			return
		}
	}

	// 发送通知
	m.sendNotification(direction, targetLabel, from.Hex(), toHex, amountStr, "ETH", txHash, int(blockNum))
}

// checkERC20Transfers 检查 ERC20 Transfer 事件
func (m *AddressMonitor) checkERC20Transfers(ctx context.Context, block *types.Block) {
	// 构建过滤器查询
	query := ethereum.FilterQuery{
		FromBlock: block.Number(),
		ToBlock:   block.Number(),
		Addresses: []common.Address{m.usdcContract},
		Topics: [][]common.Hash{
			{m.transferTopic},
		},
	}

	logs, err := m.client.FilterLogs(ctx, query)
	if err != nil {
		logger.Error("查询日志失败", zap.Error(err))
		return
	}

	for _, vLog := range logs {
		m.handleERC20Transfer(vLog, int(block.Number().Uint64()))
	}
}

// handleERC20Transfer 处理 ERC20 Transfer 事件
func (m *AddressMonitor) handleERC20Transfer(vLog types.Log, blockNum int) {
	if len(vLog.Topics) < 3 {
		return
	}

	// 解析 from 和 to
	from := common.HexToAddress(vLog.Topics[1].Hex())
	to := common.HexToAddress(vLog.Topics[2].Hex())

	// 检查是否与目标地址相关
	if !m.isMonitoredAddress(from) && !m.isMonitoredAddress(to) {
		return
	}

	txHash := vLog.TxHash.Hex()

	// 检查是否已处理
	if m.wechatRepo.ExistsByTxHash(txHash) {
		return
	}

	// 解析金额（USDC 是 6 位小数）
	amount := new(big.Int).SetBytes(vLog.Data)
	usdcAmount := new(big.Float).Quo(new(big.Float).SetInt(amount), big.NewFloat(1e6))
	amountStr := usdcAmount.Text('f', 2)

	direction := "转入"
	targetLabel := ""
	if m.isMonitoredAddress(from) {
		direction = "转出"
		targetLabel = m.getAddressLabel(from)
	} else {
		targetLabel = m.getAddressLabel(to)
	}

	logger.Info("🔔 检测到 USDC 交易",
		zap.String("direction", direction),
		zap.String("from", from.Hex()),
		zap.String("to", to.Hex()),
		zap.String("amount", amountStr+" USDC"),
		zap.String("tx", txHash),
		zap.String("label", targetLabel))

	// 发送通知
	m.sendNotification(direction, targetLabel, from.Hex(), to.Hex(), amountStr, "USDC", txHash, blockNum)
}

// sendNotification 发送通知
func (m *AddressMonitor) sendNotification(direction, label, from, to, amount, currency, txHash string, blockNum int) {
	notifStatus := "success"
	var errorMsg string

	// 发送 PushPlus 通知
	if m.pushPlus != nil {
		emoji := "📥"
		if direction == "转出" {
			emoji = "📤"
		}

		title := fmt.Sprintf("%s %s %s", emoji, currency, direction)
		content := fmt.Sprintf(`## 交易详情

**监控地址**: %s  
**币种**: %s  
**金额**: %s %s  
**方向**: %s  
**发送方**: %s  
**接收方**: %s  
**区块**: %d  
**交易**: [查看详情](https://etherscan.io/tx/%s)  
**时间**: %s`,
			label,
			currency,
			amount,
			currency,
			direction,
			from,
			to,
			blockNum,
			txHash,
			time.Now().Format("2006-01-02 15:04:05"))

		err := m.pushPlus.Send(title, content)
		if err != nil {
			logger.Error("发送通知失败", zap.Error(err))
			notifStatus = "failed"
			errorMsg = err.Error()
		}
	}

	// 记录到数据库
	if m.wechatRepo != nil {
		notifLog := &model.WechatAlter{
			Type:         fmt.Sprintf("%s_TRANSFER", currency),
			Direction:    direction,
			FromAddress:  strings.ToLower(from),
			ToAddress:    strings.ToLower(to),
			Amount:       amount,
			Currency:     currency,
			TxHash:       strings.ToLower(txHash),
			BlockNum:     blockNum,
			Content:      fmt.Sprintf("%s %s %s: %s %s (%s)", emoji(direction), currency, direction, amount, currency, label),
			Status:       notifStatus,
			ErrorMsg:     errorMsg,
			PublishType:  "pushplus",
			PublishToken: os.Getenv("PUSHPLUS_TOKEN"),
		}

		if err := m.wechatRepo.Create(notifLog); err != nil {
			logger.Error("保存通知记录失败", zap.Error(err))
		}
	}
}

func emoji(direction string) string {
	if direction == "转出" {
		return "📤"
	}
	return "📥"
}

func (m *AddressMonitor) isMonitoredAddress(address common.Address) bool {
	_, ok := m.addressSet[address]
	return ok
}

func (m *AddressMonitor) getAddressLabel(address common.Address) string {
	if label, ok := m.addressLabels[address]; ok && label != "" {
		return label
	}
	return address.Hex()
}

func (m *AddressMonitor) addressLabelList() []string {
	labels := make([]string, 0, len(m.addressLabels))
	for address, label := range m.addressLabels {
		if label == "" {
			labels = append(labels, address.Hex())
		} else {
			labels = append(labels, fmt.Sprintf("%s(%s)", label, address.Hex()))
		}
	}
	return labels
}

// Close 关闭监控器
func (m *AddressMonitor) Close() {
	if m.client != nil {
		m.client.Close()
	}
	if m.wsClient != nil {
		m.wsClient.Close()
	}
	if m.mevDetector != nil {
		m.mevDetector.Close()
	}
}
