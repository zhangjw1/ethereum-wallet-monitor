package wallet

import (
	"context"
	"ethereum-monitor/logger"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

// GoEthMonitor 基于 go-ethereum 的监控器（支持 WebSocket + HTTP 轮询）
type GoEthMonitor struct {
	client   *ethclient.Client // HTTP RPC 客户端，用于查询区块和交易数据
	wsClient *ethclient.Client // WebSocket 客户端，用于实时订阅新区块（可选，如果为 nil 则使用轮询模式）

	addressMgr   *AddressManager      // 地址管理器，管理监控的钱包地址列表和标签
	notifSvc     *NotificationService // 通知服务，负责发送通知和记录到数据库
	mevFilter    *MevFilter           // MEV 过滤器，用于检测和过滤 MEV Bot 交易
	tokenHandler *TokenHandler        // 代币处理器，管理 ERC20 代币配置和金额解析

	ethThreshold   *big.Int // ETH 转账阈值（Wei 单位），只有超过此金额的交易才会触发通知
	tokenThreshold *big.Int // ERC20 代币转账阈值（最小单位），只有超过此金额的交易才会触发通知
}

// NewGoEthMonitor 创建 go-ethereum 监控器
func NewGoEthMonitor(rpcURL, wsURL string, config *MonitorConfig) (*GoEthMonitor, error) {
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

	// 创建地址管理器
	addressMgr := NewAddressManager(config.Addresses)

	// 创建通知服务
	notifSvc := NewNotificationService()

	// 创建 MEV 过滤器
	mevFilter, err := NewMevFilter(rpcURL)
	if err != nil {
		logger.Warn("创建 MEV 过滤器失败", zap.Error(err))
		// 不返回错误，继续创建监控器
	}

	// 创建代币处理器
	tokenHandler := NewTokenHandler(config.Tokens)

	return &GoEthMonitor{
		client:         client,
		wsClient:       wsClient,
		addressMgr:     addressMgr,
		notifSvc:       notifSvc,
		mevFilter:      mevFilter,
		tokenHandler:   tokenHandler,
		ethThreshold:   config.ETHThreshold,
		tokenThreshold: config.TokenThreshold,
	}, nil
}

// Start 启动监控
func (m *GoEthMonitor) Start(ctx context.Context) error {
	logger.Info("🚀 启动 go-ethereum 地址监控",
		zap.Int("address_count", len(m.addressMgr.addressSet)),
		zap.Strings("addresses", m.addressMgr.GetLabelList()),
		zap.Bool("websocket", m.wsClient != nil))

	if m.wsClient != nil {
		// 使用 WebSocket 实时订阅
		return m.startWebSocketMonitor(ctx)
	}
	// 使用轮询模式
	return m.startPollingMonitor(ctx)
}

// startWebSocketMonitor WebSocket 实时监控
func (m *GoEthMonitor) startWebSocketMonitor(ctx context.Context) error {
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
func (m *GoEthMonitor) startPollingMonitor(ctx context.Context) error {
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
func (m *GoEthMonitor) checkBlockTransactions(ctx context.Context, block *types.Block) {
	// 检查 ETH 交易
	for _, tx := range block.Transactions() {
		if m.isRelatedTransaction(tx) {
			m.handleETHTransaction(ctx, tx, block.Number().Uint64())
		}
	}

	// 检查 ERC20 Transfer 事件
	m.checkERC20Transfers(ctx, block)
}

// isRelatedTransaction 判断交易是否与目标地址相关
func (m *GoEthMonitor) isRelatedTransaction(tx *types.Transaction) bool {
	// 检查接收方
	if tx.To() != nil && m.addressMgr.IsMonitored(*tx.To()) {
		return true
	}

	// 检查发送方
	msg, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
	if err == nil && m.addressMgr.IsMonitored(msg) {
		return true
	}

	return false
}

// handleETHTransaction 处理 ETH 交易
func (m *GoEthMonitor) handleETHTransaction(ctx context.Context, tx *types.Transaction, blockNum uint64) {
	txHash := tx.Hash().Hex()

	// 检查是否已处理
	if m.notifSvc.IsProcessed(txHash) {
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

	fromMonitored := m.addressMgr.IsMonitored(from)
	toMonitored := tx.To() != nil && m.addressMgr.IsMonitored(to)

	direction := "转入"
	targetLabel := ""
	if fromMonitored {
		direction = "转出"
		targetLabel = m.addressMgr.GetLabel(from)
	} else if toMonitored {
		targetLabel = m.addressMgr.GetLabel(to)
	}

	// 计算金额
	amountStr := WeiToEth(tx.Value())

	logger.Info("🔔 检测到 ETH 交易",
		zap.String("direction", direction),
		zap.String("from", from.Hex()),
		zap.String("to", toHex),
		zap.String("amount", amountStr+" ETH"),
		zap.String("tx", txHash),
		zap.String("label", targetLabel))

	// MEV 检测
	if m.mevFilter != nil && m.mevFilter.IsMevTransaction(txHash) {
		return
	}

	// 检查是否超过阈值
	shouldAlert := m.ethThreshold != nil && tx.Value().Cmp(m.ethThreshold) > 0

	// 发送通知
	notif := &TransferNotification{
		Direction:   direction,
		Label:       targetLabel,
		From:        from.Hex(),
		To:          toHex,
		Amount:      amountStr,
		Currency:    "ETH",
		TxHash:      txHash,
		BlockNum:    int(blockNum),
		ShouldAlert: shouldAlert,
	}

	if err := m.notifSvc.SendTransferNotification(notif); err != nil {
		logger.Error("发送通知失败", zap.Error(err))
	}
}

// checkERC20Transfers 检查 ERC20 Transfer 事件
func (m *GoEthMonitor) checkERC20Transfers(ctx context.Context, block *types.Block) {
	monitoredTokens := m.tokenHandler.GetMonitoredTokens()
	if len(monitoredTokens) == 0 {
		return
	}

	// 构建过滤器查询
	query := ethereum.FilterQuery{
		FromBlock: block.Number(),
		ToBlock:   block.Number(),
		Addresses: monitoredTokens,
		Topics: [][]common.Hash{
			{m.tokenHandler.GetTransferTopic()},
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
func (m *GoEthMonitor) handleERC20Transfer(vLog types.Log, blockNum int) {
	if len(vLog.Topics) < 3 {
		return
	}

	// 解析 from 和 to
	from := common.HexToAddress(vLog.Topics[1].Hex())
	to := common.HexToAddress(vLog.Topics[2].Hex())

	// 检查是否与目标地址相关
	if !m.addressMgr.IsMonitored(from) && !m.addressMgr.IsMonitored(to) {
		return
	}

	txHash := vLog.TxHash.Hex()

	// 检查是否已处理
	if m.notifSvc.IsProcessed(txHash) {
		return
	}

	// 获取代币配置
	tokenConfig, ok := m.tokenHandler.GetTokenConfig(vLog.Address)
	if !ok {
		return
	}

	// 解析金额
	amount := new(big.Int).SetBytes(vLog.Data)
	amountStr := m.tokenHandler.ParseTransferAmount(vLog.Address, amount)

	direction := "转入"
	targetLabel := ""
	if m.addressMgr.IsMonitored(from) {
		direction = "转出"
		targetLabel = m.addressMgr.GetLabel(from)
	} else {
		targetLabel = m.addressMgr.GetLabel(to)
	}

	logger.Info("🔔 检测到代币交易",
		zap.String("token", tokenConfig.Symbol),
		zap.String("direction", direction),
		zap.String("from", from.Hex()),
		zap.String("to", to.Hex()),
		zap.String("amount", amountStr+" "+tokenConfig.Symbol),
		zap.String("tx", txHash),
		zap.String("label", targetLabel))

	// 检查是否超过阈值
	shouldAlert := m.tokenThreshold != nil && amount.Cmp(m.tokenThreshold) > 0

	// 发送通知
	notif := &TransferNotification{
		Direction:   direction,
		Label:       targetLabel,
		From:        from.Hex(),
		To:          to.Hex(),
		Amount:      amountStr,
		Currency:    tokenConfig.Symbol,
		TxHash:      txHash,
		BlockNum:    blockNum,
		ShouldAlert: shouldAlert,
	}

	if err := m.notifSvc.SendTransferNotification(notif); err != nil {
		logger.Error("发送通知失败", zap.Error(err))
	}
}

// Close 关闭监控器
func (m *GoEthMonitor) Close() {
	if m.client != nil {
		m.client.Close()
	}
	if m.wsClient != nil {
		m.wsClient.Close()
	}
	if m.mevFilter != nil {
		m.mevFilter.Close()
	}
}
