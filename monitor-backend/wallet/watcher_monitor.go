package wallet

import (
	"context"
	"ethereum-monitor/logger"
	"math/big"
	"strings"

	ethereum "github.com/HydroProtocol/ethereum-watcher"
	"github.com/HydroProtocol/ethereum-watcher/structs"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

// WatcherMonitor 基于 ethereum-watcher 的监控器（HTTP 轮询）
type WatcherMonitor struct {
	watcher *ethereum.AbstractWatcher // ethereum-watcher 框架的监听器实例，负责轮询区块和分发事件

	addressMgr   *AddressManager      // 地址管理器，管理监控的钱包地址列表和标签
	notifSvc     *NotificationService // 通知服务，负责发送通知和记录到数据库
	mevFilter    *MevFilter           // MEV 过滤器，用于检测和过滤 MEV Bot 交易
	tokenHandler *TokenHandler        // 代币处理器，管理 ERC20 代币配置和金额解析

	ethThreshold   *big.Int // ETH 转账阈值（Wei 单位），只有超过此金额的交易才会触发通知
	tokenThreshold *big.Int // ERC20 代币转账阈值（最小单位），只有超过此金额的交易才会触发通知
}

// NewWatcherMonitor 创建 ethereum-watcher 监控器
func NewWatcherMonitor(rpcURL string, config *MonitorConfig) (*WatcherMonitor, error) {
	// 创建地址管理器
	addressMgr := NewAddressManager(config.Addresses)

	// 创建通知服务
	notifSvc := NewNotificationService()

	// 创建 MEV 过滤器
	mevFilter, err := NewMevFilter(rpcURL)
	if err != nil {
		logger.Warn("创建 MEV 过滤器失败", zap.Error(err))
	}

	// 创建代币处理器
	tokenHandler := NewTokenHandler(config.Tokens)

	return &WatcherMonitor{
		addressMgr:     addressMgr,
		notifSvc:       notifSvc,
		mevFilter:      mevFilter,
		tokenHandler:   tokenHandler,
		ethThreshold:   config.ETHThreshold,
		tokenThreshold: config.TokenThreshold,
	}, nil
}

// Start 启动监控
func (m *WatcherMonitor) Start(ctx context.Context, rpcURL string, pollInterval int) error {
	logger.Info("🚀 启动 ethereum-watcher 地址监控",
		zap.Int("address_count", len(m.addressMgr.addressSet)),
		zap.Strings("addresses", m.addressMgr.GetLabelList()),
		zap.Int("pollInterval", pollInterval))

	// 创建 Watcher
	m.watcher = ethereum.NewHttpBasedEthWatcher(ctx, rpcURL)
	m.watcher.SetSleepSecondsForNewBlock(pollInterval)

	// 注册 ETH 交易插件
	ethPlugin := &ethTransactionPlugin{
		monitor: m,
	}
	m.watcher.RegisterTxPlugin(ethPlugin)
	logger.Info("✅ ETH 交易插件已注册")

	// 注册 ERC20 Transfer 插件
	for _, tokenAddr := range m.tokenHandler.GetMonitoredTokens() {
		transferPlugin := &erc20TransferPlugin{
			monitor:      m,
			tokenAddress: tokenAddr,
		}
		m.watcher.RegisterReceiptLogPlugin(transferPlugin)

		if config, ok := m.tokenHandler.GetTokenConfig(tokenAddr); ok {
			logger.Info("✅ ERC20 Transfer 插件已注册",
				zap.String("token", config.Symbol),
				zap.String("address", tokenAddr.Hex()))
		}
	}

	logger.Info("⏳ 开始监听新区块...")

	// 运行监听器
	err := m.watcher.RunTillExit()
	if err != nil {
		logger.Error("监听器运行错误", zap.Error(err))
		return err
	}

	return nil
}

// Close 关闭监控器
func (m *WatcherMonitor) Close() {
	if m.mevFilter != nil {
		m.mevFilter.Close()
	}
}

// ethTransactionPlugin ETH 交易插件
// 实现 ITxPlugin 接口，用于监听和处理 ETH 原生代币的转账交易
type ethTransactionPlugin struct {
	monitor *WatcherMonitor // 监控器实例，用于访问地址管理、通知服务、MEV 过滤等公共组件
}

func (p *ethTransactionPlugin) AcceptTx(tx structs.RemovableTx) {
	if tx.IsRemoved {
		return
	}

	// 使用不区分大小写的比较
	from := strings.ToLower(tx.GetFrom())
	to := strings.ToLower(tx.GetTo())

	// 检查是否与监控地址相关
	fromAddr := common.HexToAddress(from)
	toAddr := common.HexToAddress(to)

	if !p.monitor.addressMgr.IsMonitored(fromAddr) && !p.monitor.addressMgr.IsMonitored(toAddr) {
		return
	}

	txHash := tx.GetHash()

	// 检查是否已处理
	if p.monitor.notifSvc.IsProcessed(txHash) {
		return
	}

	value := tx.GetValue()

	// 检查是否超过阈值
	if p.monitor.ethThreshold != nil && value.Cmp(p.monitor.ethThreshold) <= 0 {
		return
	}

	// 判断方向
	direction := "转入"
	targetLabel := ""
	if p.monitor.addressMgr.IsMonitored(fromAddr) {
		direction = "转出"
		targetLabel = p.monitor.addressMgr.GetLabel(fromAddr)
	} else {
		targetLabel = p.monitor.addressMgr.GetLabel(toAddr)
	}

	amountStr := WeiToEth(&value)

	logger.Info("🔔 检测到 ETH 交易",
		zap.String("direction", direction),
		zap.String("from", tx.GetFrom()),
		zap.String("to", tx.GetTo()),
		zap.String("amount", amountStr+" ETH"),
		zap.String("tx", txHash),
		zap.String("label", targetLabel))

	// MEV 检测
	if p.monitor.mevFilter != nil && p.monitor.mevFilter.IsMevTransaction(txHash) {
		return
	}

	// 发送通知
	notif := &TransferNotification{
		Direction:   direction,
		Label:       targetLabel,
		From:        tx.GetFrom(),
		To:          tx.GetTo(),
		Amount:      amountStr,
		Currency:    "ETH",
		TxHash:      txHash,
		BlockNum:    int(tx.GetBlockNumber()),
		ShouldAlert: true, // 已经过阈值检查
	}

	if err := p.monitor.notifSvc.SendTransferNotification(notif); err != nil {
		logger.Error("发送通知失败", zap.Error(err))
	}
}

// erc20TransferPlugin ERC20 Transfer 插件
// 实现 IReceiptLogPlugin 接口，用于监听特定 ERC20 代币的 Transfer 事件
type erc20TransferPlugin struct {
	monitor      *WatcherMonitor // 监控器实例，用于访问地址管理、通知服务等公共组件
	tokenAddress common.Address  // 要监听的 ERC20 代币合约地址（如 USDT、USDC 等）
}

func (p *erc20TransferPlugin) Accept(log *structs.RemovableReceiptLog) {
	if log.IsRemoved {
		return
	}

	topics := log.GetTopics()
	if len(topics) < 3 {
		return
	}

	// 解析 from 和 to
	from := strings.ToLower(ExtractAddressFromTopic(topics[1]))
	to := strings.ToLower(ExtractAddressFromTopic(topics[2]))

	fromAddr := common.HexToAddress(from)
	toAddr := common.HexToAddress(to)

	// 检查是否与监控地址相关
	if !p.monitor.addressMgr.IsMonitored(fromAddr) && !p.monitor.addressMgr.IsMonitored(toAddr) {
		logger.Info("不相关地址的转账，不做处理~~~")
		return
	}

	txHash := log.GetTransactionHash()

	// 检查是否已处理
	if p.monitor.notifSvc.IsProcessed(txHash) {
		return
	}

	// 获取代币配置
	tokenConfig, ok := p.monitor.tokenHandler.GetTokenConfig(p.tokenAddress)
	if !ok {
		return
	}

	// 解析金额
	amount := new(big.Int).SetBytes(common.FromHex(log.GetData()))

	// 检查是否超过阈值
	if p.monitor.tokenThreshold != nil && amount.Cmp(p.monitor.tokenThreshold) <= 0 {
		return
	}

	amountStr := p.monitor.tokenHandler.ParseTransferAmount(p.tokenAddress, amount)

	// 判断方向
	direction := "转入"
	targetLabel := ""
	if p.monitor.addressMgr.IsMonitored(fromAddr) {
		direction = "转出"
		targetLabel = p.monitor.addressMgr.GetLabel(fromAddr)
	} else {
		targetLabel = p.monitor.addressMgr.GetLabel(toAddr)
	}

	logger.Info("🔔 检测到代币交易",
		zap.String("token", tokenConfig.Symbol),
		zap.String("direction", direction),
		zap.String("from", from),
		zap.String("to", to),
		zap.String("amount", amountStr+" "+tokenConfig.Symbol),
		zap.String("tx", txHash),
		zap.String("label", targetLabel))

	// 发送通知
	notif := &TransferNotification{
		Direction:   direction,
		Label:       targetLabel,
		From:        from,
		To:          to,
		Amount:      amountStr,
		Currency:    tokenConfig.Symbol,
		TxHash:      txHash,
		BlockNum:    log.GetBlockNum(),
		ShouldAlert: true, // 已经过阈值检查
	}

	if err := p.monitor.notifSvc.SendTransferNotification(notif); err != nil {
		logger.Error("发送通知失败", zap.Error(err))
	}
}

func (p *erc20TransferPlugin) FromContract() string {
	return p.tokenAddress.Hex()
}

func (p *erc20TransferPlugin) InterestedTopics() []string {
	return []string{p.monitor.tokenHandler.GetTransferTopic().Hex()}
}

func (p *erc20TransferPlugin) NeedReceiptLog(receiptLog *structs.RemovableReceiptLog) bool {
	return true
}
