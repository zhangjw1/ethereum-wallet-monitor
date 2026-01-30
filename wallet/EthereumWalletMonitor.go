package wallet

import (
	"context"
	"ethereum-monitor/config"
	"ethereum-monitor/database"
	"ethereum-monitor/logger"
	"ethereum-monitor/model"
	"ethereum-monitor/utils"
	"fmt"
	"math/big"
	"os"
	"strings"

	"go.uber.org/zap"

	"github.com/ethereum/go-ethereum/common"

	ethereum "github.com/HydroProtocol/ethereum-watcher"
	"github.com/HydroProtocol/ethereum-watcher/structs"
)

type EtherenumThransactionPlugin struct {
	targetAddress string
	threshold     *big.Int
}

func (p *EtherenumThransactionPlugin) AcceptTx(tx structs.RemovableTx) {
	logger.Debug("收到交易",
		zap.Uint64("block", tx.GetBlockNumber()),
		zap.String("hash", tx.GetHash()))

	// 使用不区分大小写的比较
	from := strings.ToLower(tx.GetFrom())
	to := strings.ToLower(tx.GetTo())
	target := strings.ToLower(p.targetAddress)

	if from != target && to != target {
		return
	}

	value := tx.GetValue()
	gasPrice := tx.GetGasPrice()
	logger.Info("匹配到目标地址的交易",
		zap.String("hash", tx.GetHash()),
		zap.String("amount", weiToEth(&value)+" ETH"),
		zap.String("from", tx.GetFrom()),
		zap.String("to", tx.GetTo()),
		zap.Uint64("block", tx.GetBlockNumber()),
		zap.String("gasPrice", gasPrice.String()))

	if value.Cmp(p.threshold) > 0 {
		p.processTransaction(tx)
	}
}

// 监控交易的信息
func (p *EtherenumThransactionPlugin) Accept(tx *structs.RemovableTxAndReceipt) {
	logger.Debug("匹配到目标地址的区块", zap.Any("logs", tx.Receipt.GetLogs()))
}

type USDTTransferPlugin struct {
	targetAddress string
	threshold     *big.Int
	mevDetector   *utils.MevDetector
	pushPlus      *utils.PushPlusNotifier
	wechatRepo    *database.WechatAlterRepository
}

func (p *USDTTransferPlugin) Accept(log *structs.RemovableReceiptLog) {
	logger.Debug("收到 USDT Transfer 事件",
		zap.String("blockHash", log.GetBlockHash()),
		zap.Int("blockNum", log.GetBlockNum()),
		zap.String("txHash", log.GetTransactionHash()))

	if log.IsRemoved {
		logger.Warn("日志被删除", zap.String("blockHash", log.GetBlockHash()))
		return
	}

	topics := log.GetTopics()
	if len(topics) < 3 {
		logger.Warn("USDT Transfer 事件 topics 数量不足", zap.Int("count", len(topics)))
		return
	}

	// topics[0] 是事件签名 Transfer(address,address,uint256)
	// topics[1] 是 from 地址
	// topics[2] 是 to 地址
	from := strings.ToLower(extractAddress(topics[1]))
	to := strings.ToLower(extractAddress(topics[2]))
	target := strings.ToLower(p.targetAddress)

	logger.Debug("USDT Transfer 地址信息",
		zap.String("from", from),
		zap.String("to", to),
		zap.String("target", target))

	/*	if from != target && to != target {
		return
	}*/

	// data 字段包含转账金额
	value := new(big.Int).SetBytes(common.FromHex(log.GetData()))

	// 将 USDT 金额转换为可读格式（6位小数）
	usdtAmount := new(big.Float).SetInt(value)
	divisor := new(big.Float).SetFloat64(1e6)
	result := new(big.Float).Quo(usdtAmount, divisor)

	logger.Debug("检测到 USDT 转账",
		zap.String("from", from),
		zap.String("to", to),
		zap.String("amount", result.String()+" USDT"),
		zap.String("txHash", log.GetTransactionHash()))

	if value.Cmp(p.threshold) > 0 {
		// 使用 MEV 检测器检查交易
		txHash := log.GetTransactionHash()
		mevResult, err := p.mevDetector.DetectMev(txHash)
		if err != nil {
			logger.Error("MEV 检测失败", zap.String("txHash", txHash), zap.Error(err))
			// 检测失败时仍然发出告警
		} else if mevResult.IsMev {
			// 如果是 MEV 攻击，记录但不告警
			logger.Info("检测到 MEV Bot 转账，跳过告警",
				zap.String("mevType", string(mevResult.MevType)),
				zap.Float64("confidence", mevResult.Confidence),
				zap.String("from", from),
				zap.String("to", to),
				zap.String("amount", result.String()+" USDT"),
				zap.String("txHash", txHash),
				zap.Strings("evidence", mevResult.Evidence))
			// 可选：发送 MEV 检测通知（不是告警）
			return
		}

		// 非 MEV 攻击的大额转账，发出告警
		direction := "转入"
		if from == target {
			direction = "转出"
		}

		logger.Warn("🚨 USDT 大额转账告警",
			zap.String("direction", direction),
			zap.String("from", from),
			zap.String("to", to),
			zap.String("amount", result.String()+" USDT"),
			zap.String("txHash", log.GetTransactionHash()),
			zap.Int("blockNum", log.GetBlockNum()))

		// 发送微信通知
		notifStatus := "success"
		var errorMsg string

		if p.pushPlus != nil {
			err := p.pushPlus.SendUSDTAlert(
				direction,
				from,
				to,
				result.String(),
				txHash,
				log.GetBlockNum(),
			)
			if err != nil {
				logger.Error("发送微信通知失败", zap.Error(err))
				notifStatus = "failed"
				errorMsg = err.Error()
			}
		}

		// 记录到数据库
		if p.wechatRepo != nil {
			notifLog := &model.WechatAlter{
				Type:         "USDT_ALERT",
				Direction:    direction,
				FromAddress:  from,
				ToAddress:    to,
				Amount:       result.String(),
				Currency:     "USDT",
				TxHash:       txHash,
				BlockNum:     log.GetBlockNum(),
				Content:      fmt.Sprintf("🚨 USDT 大额%s告警: %s USDT", direction, result.String()),
				Status:       notifStatus,
				ErrorMsg:     errorMsg,
				PublishType:  "pushplus",
				PublishToken: os.Getenv("PUSHPLUS_TOKEN"),
			}

			if err := p.wechatRepo.Create(notifLog); err != nil {
				logger.Error("保存通知记录失败", zap.Error(err))
			}
		}
	}
}

func (p *USDTTransferPlugin) FromContract() string {
	return config.UsdtContractAddress
}

func (p *USDTTransferPlugin) InterestedTopics() []string {
	return []string{config.UsdtTransferTopic}
}

func (p *USDTTransferPlugin) NeedReceiptLog(receiptLog *structs.RemovableReceiptLog) bool {
	return true
}

// 辅助函数：从 Topic 中提取地址
func extractAddress(topic string) string {
	// Topic 是 32 字节，地址是后 20 字节
	if len(topic) >= 66 { // "0x" + 64 个字符
		return "0x" + topic[26:] // 跳过前 26 个字符（0x + 24个0）
	}
	return topic
}

func (p *EtherenumThransactionPlugin) processTransaction(tx structs.RemovableTx) {
	if tx.IsRemoved {
		logger.Warn("交易被删除", zap.String("hash", tx.GetHash()))
		return
	}
	direction := "转入"
	if tx.GetFrom() == p.targetAddress {
		direction = "转出"
	}
	value := tx.GetValue()

	// 检测 MEV 攻击
	mevDetector, err := utils.NewMevDetector(config.GetEthereumRpcUrl())
	if err == nil {
		defer mevDetector.Close()
		mevResult, err := mevDetector.DetectMev(tx.GetHash())
		if err == nil && mevResult.IsMev {
			logger.Warn("⚠️  检测到 MEV 攻击",
				zap.String("type", string(mevResult.MevType)),
				zap.Float64("confidence", mevResult.Confidence),
				zap.Strings("evidence", mevResult.Evidence))
		}
	}

	logger.Warn("🚨 大额交易告警",
		zap.String("direction", direction),
		zap.String("hash", tx.GetHash()),
		zap.String("amount", weiToEth(&value)+" ETH"),
		zap.String("from", tx.GetFrom()),
		zap.String("to", tx.GetTo()),
		zap.Uint64("block", tx.GetBlockNumber()))
}

// 创建 ETH 的阈值（基于配置）
func createThreshold() *big.Int {
	threshold := big.NewInt(0)
	// 将 ETH 阈值转换为 Wei 单位 (1 ETH = 10^18 Wei)
	// config.ETH_THRESHOLD 是以 ETH 为单位的阈值，这里是 10 ETH
	ethValue := new(big.Int).Mul(big.NewInt(int64(config.EthThreshold)), big.NewInt(1000000000000000000))
	threshold.Set(ethValue)
	return threshold
}

func createUSDTThreshold(amount int64) *big.Int {
	// USDT 是 6 位小数
	threshold := big.NewInt(amount)
	threshold.Mul(threshold, big.NewInt(1000000)) // 乘以 10^6
	return threshold
}

func weiToEth(wei *big.Int) string {
	ethWei := new(big.Float).SetInt(wei)
	divisor := new(big.Float).SetFloat64(1e18)
	result := new(big.Float).Quo(ethWei, divisor)
	return result.String()
}

func AddressAddMonitor() {
	logger.Info("🚀 以太坊钱包监控程序启动")

	// 必须在创建 watcher 之前设置代理
	utils.SetGlobalProxy("http://127.0.0.1:7890")

	// 创建 MEV 检测器
	mevDetector, err := utils.NewMevDetector(config.GetEthereumRpcUrl())
	if err != nil {
		logger.Fatal("创建 MEV 检测器失败", zap.Error(err))
	}
	defer mevDetector.Close()

	// 创建 PushPlus 通知器
	var pushPlus *utils.PushPlusNotifier
	pushPlusToken := os.Getenv("PUSHPLUS_TOKEN")
	if pushPlusToken != "" {
		pushPlus = utils.NewPushPlusNotifier(pushPlusToken)
		logger.Info("PushPlus 微信通知已启用")
	} else {
		logger.Warn("未配置 PushPlus，将只记录日志")
	}

	// 创建通知记录 Repository
	wechatRepo := database.NewWechatAlterRepository()

	ethereumPlugin := &EtherenumThransactionPlugin{
		targetAddress: config.OkxWalletAddress,
		threshold:     createThreshold(),
	}

	usdtTransferPlugin := &USDTTransferPlugin{
		targetAddress: config.OkxWalletAddress,
		threshold:     createUSDTThreshold(config.UsdtThreshold),
		mevDetector:   mevDetector,
		pushPlus:      pushPlus,
		wechatRepo:    wechatRepo,
	}

	logger.Info("正在创建 Watcher...")
	watcher := ethereum.NewHttpBasedEthWatcher(context.Background(), config.GetEthereumRpcUrl())

	// 设置轮询间隔（秒）
	watcher.SetSleepSecondsForNewBlock(config.SleepSecondsForNewBlock)
	logger.Info("配置完成", zap.String("address", config.OkxWalletAddress), zap.Int("threshold", config.EthThreshold))

	watcher.RegisterTxPlugin(ethereumPlugin)
	logger.Info("ETH 交易插件已注册")

	watcher.RegisterReceiptLogPlugin(usdtTransferPlugin)
	logger.Info("USDT Transfer 插件已注册",
		zap.String("contract", config.UsdtContractAddress),
		zap.String("topic", config.UsdtTransferTopic),
		zap.Int64("threshold", config.UsdtThreshold))

	logger.Info("⏳ 等待新区块...")

	err = watcher.RunTillExit()
	if err != nil {
		logger.Error("运行错误", zap.Error(err))
		return
	}
}
