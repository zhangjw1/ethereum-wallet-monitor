package wallet

import (
	"ethereum-monitor/database"
	"ethereum-monitor/logger"
	"ethereum-monitor/model"
	"ethereum-monitor/utils"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

// TokenConfig ERC20 代币配置
type TokenConfig struct {
	Address  common.Address
	Symbol   string
	Decimals int
}

// MonitorConfig 监控配置
type MonitorConfig struct {
	Addresses      map[string]string // 地址 -> 标签
	Tokens         []TokenConfig     // 要监控的 ERC20 代币
	ETHThreshold   *big.Int          // ETH 阈值（Wei）
	TokenThreshold *big.Int          // 代币阈值（最小单位）
}

// AddressManager 地址管理器
type AddressManager struct {
	addressLabels map[common.Address]string
	addressSet    map[common.Address]struct{}
}

// NewAddressManager 创建地址管理器
func NewAddressManager(addresses map[string]string) *AddressManager {
	addressLabels := make(map[common.Address]string, len(addresses))
	addressSet := make(map[common.Address]struct{}, len(addresses))

	for addr, label := range addresses {
		parsed := common.HexToAddress(addr)
		addressLabels[parsed] = label
		addressSet[parsed] = struct{}{}
	}

	return &AddressManager{
		addressLabels: addressLabels,
		addressSet:    addressSet,
	}
}

// IsMonitored 检查地址是否被监控
func (am *AddressManager) IsMonitored(address common.Address) bool {
	_, ok := am.addressSet[address]
	return ok
}

// GetLabel 获取地址标签
func (am *AddressManager) GetLabel(address common.Address) string {
	if label, ok := am.addressLabels[address]; ok && label != "" {
		return label
	}
	return address.Hex()
}

// GetLabelList 获取所有地址标签列表
func (am *AddressManager) GetLabelList() []string {
	labels := make([]string, 0, len(am.addressLabels))
	for address, label := range am.addressLabels {
		if label == "" {
			labels = append(labels, address.Hex())
		} else {
			labels = append(labels, fmt.Sprintf("%s(%s)", label, address.Hex()))
		}
	}
	return labels
}

// NotificationService 通知服务
type NotificationService struct {
	pushPlus   *utils.PushPlusNotifier
	wechatRepo *database.WechatAlterRepository
}

// NewNotificationService 创建通知服务
func NewNotificationService() *NotificationService {
	var pushPlus *utils.PushPlusNotifier
	if token := os.Getenv("PUSHPLUS_TOKEN"); token != "" {
		pushPlus = utils.NewPushPlusNotifier(token)
	}

	return &NotificationService{
		pushPlus:   pushPlus,
		wechatRepo: database.NewWechatAlterRepository(),
	}
}

// TransferNotification 转账通知信息
type TransferNotification struct {
	Direction   string
	Label       string
	From        string
	To          string
	Amount      string
	Currency    string
	TxHash      string
	BlockNum    int
	ShouldAlert bool // 是否需要告警（大额交易）
}

// SendTransferNotification 发送转账通知
func (ns *NotificationService) SendTransferNotification(notif *TransferNotification) error {
	notifStatus := "success"
	var errorMsg string

	// 发送 PushPlus 通知
	if ns.pushPlus != nil && notif.ShouldAlert {
		emoji := "📥"
		if notif.Direction == "转出" {
			emoji = "📤"
		}

		title := fmt.Sprintf("%s %s %s", emoji, notif.Currency, notif.Direction)
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
			notif.Label,
			notif.Currency,
			notif.Amount,
			notif.Currency,
			notif.Direction,
			notif.From,
			notif.To,
			notif.BlockNum,
			notif.TxHash,
			time.Now().Format("2006-01-02 15:04:05"))

		err := ns.pushPlus.Send(title, content)
		if err != nil {
			logger.Error("发送通知失败", zap.Error(err))
			notifStatus = "failed"
			errorMsg = err.Error()
		}
	}

	// 记录到数据库
	if ns.wechatRepo != nil {
		emoji := "📥"
		if notif.Direction == "转出" {
			emoji = "📤"
		}

		notifLog := &model.WechatAlter{
			Type:         fmt.Sprintf("%s_TRANSFER", notif.Currency),
			Direction:    notif.Direction,
			FromAddress:  strings.ToLower(notif.From),
			ToAddress:    strings.ToLower(notif.To),
			Amount:       notif.Amount,
			Currency:     notif.Currency,
			TxHash:       strings.ToLower(notif.TxHash),
			BlockNum:     notif.BlockNum,
			Content:      fmt.Sprintf("%s %s %s: %s %s (%s)", emoji, notif.Currency, notif.Direction, notif.Amount, notif.Currency, notif.Label),
			Status:       notifStatus,
			ErrorMsg:     errorMsg,
			PublishType:  "pushplus",
			PublishToken: os.Getenv("PUSHPLUS_TOKEN"),
		}

		if err := ns.wechatRepo.Create(notifLog); err != nil {
			logger.Error("保存通知记录失败", zap.Error(err))
			return err
		}
	}

	return nil
}

// IsProcessed 检查交易是否已处理
func (ns *NotificationService) IsProcessed(txHash string) bool {
	if ns.wechatRepo == nil {
		return false
	}
	return ns.wechatRepo.ExistsByTxHash(txHash)
}

// MevFilter MEV 过滤器
type MevFilter struct {
	detector *utils.MevDetector
}

// NewMevFilter 创建 MEV 过滤器
func NewMevFilter(rpcURL string) (*MevFilter, error) {
	detector, err := utils.NewMevDetector(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("创建 MEV 检测器失败: %w", err)
	}

	return &MevFilter{
		detector: detector,
	}, nil
}

// IsMevTransaction 检查是否是 MEV 交易
func (mf *MevFilter) IsMevTransaction(txHash string) bool {
	if mf.detector == nil {
		return false
	}

	result, err := mf.detector.DetectMev(txHash)
	if err != nil {
		logger.Debug("MEV 检测失败", zap.String("txHash", txHash), zap.Error(err))
		return false
	}

	if result.IsMev {
		logger.Info("检测到 MEV 交易，跳过通知",
			zap.String("type", string(result.MevType)),
			zap.Float64("confidence", result.Confidence),
			zap.String("txHash", txHash))
		return true
	}

	return false
}

// Close 关闭 MEV 过滤器
func (mf *MevFilter) Close() {
	if mf.detector != nil {
		mf.detector.Close()
	}
}

// TokenHandler ERC20 代币处理器
type TokenHandler struct {
	tokens        map[common.Address]*TokenConfig
	transferTopic common.Hash
}

// NewTokenHandler 创建代币处理器
func NewTokenHandler(tokens []TokenConfig) *TokenHandler {
	tokenMap := make(map[common.Address]*TokenConfig)
	for i := range tokens {
		tokenMap[tokens[i].Address] = &tokens[i]
	}

	// Transfer(address indexed from, address indexed to, uint256 value)
	transferTopic := common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")

	return &TokenHandler{
		tokens:        tokenMap,
		transferTopic: transferTopic,
	}
}

// GetTokenConfig 获取代币配置
func (th *TokenHandler) GetTokenConfig(address common.Address) (*TokenConfig, bool) {
	config, ok := th.tokens[address]
	return config, ok
}

// GetTransferTopic 获取 Transfer 事件主题
func (th *TokenHandler) GetTransferTopic() common.Hash {
	return th.transferTopic
}

// GetMonitoredTokens 获取所有监控的代币地址
func (th *TokenHandler) GetMonitoredTokens() []common.Address {
	addresses := make([]common.Address, 0, len(th.tokens))
	for addr := range th.tokens {
		addresses = append(addresses, addr)
	}
	return addresses
}

// ParseTransferAmount 解析转账金额
func (th *TokenHandler) ParseTransferAmount(tokenAddress common.Address, rawAmount *big.Int) string {
	config, ok := th.GetTokenConfig(tokenAddress)
	if !ok {
		return rawAmount.String()
	}

	// 转换为可读格式
	divisor := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(config.Decimals)), nil))
	amount := new(big.Float).SetInt(rawAmount)
	result := new(big.Float).Quo(amount, divisor)

	// 根据小数位数格式化
	precision := 2
	if config.Decimals > 6 {
		precision = 6
	}

	return result.Text('f', precision)
}

// ExtractAddressFromTopic 从 Topic 中提取地址
func ExtractAddressFromTopic(topic string) string {
	// Topic 是 32 字节，地址是后 20 字节
	if len(topic) >= 66 { // "0x" + 64 个字符
		return "0x" + topic[26:] // 跳过前 26 个字符（0x + 24个0）
	}
	return topic
}

// WeiToEth 将 Wei 转换为 ETH
func WeiToEth(wei *big.Int) string {
	ethAmount := new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(1e18))
	return ethAmount.Text('f', 6)
}

// CreateETHThreshold 创建 ETH 阈值
func CreateETHThreshold(ethAmount int64) *big.Int {
	threshold := new(big.Int).Mul(big.NewInt(ethAmount), big.NewInt(1e18))
	return threshold
}

// CreateTokenThreshold 创建代币阈值
func CreateTokenThreshold(amount int64, decimals int) *big.Int {
	threshold := new(big.Int).Mul(
		big.NewInt(amount),
		new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil),
	)
	return threshold
}
