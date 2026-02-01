package monitor

import (
	"ethereum-monitor/analyzer"
	"ethereum-monitor/config"
	"ethereum-monitor/database"
	"ethereum-monitor/logger"
	"ethereum-monitor/model"
	"ethereum-monitor/utils"
	"os"
	"strings"
	"time"

	"github.com/HydroProtocol/ethereum-watcher/structs"
	"go.uber.org/zap"
)

// ContractDeploymentPlugin 合约部署监听插件
type ContractDeploymentPlugin struct {
	deploymentRepo *database.ContractDeploymentRepository
	analyzer       *analyzer.MemeTokenAnalyzer
	pushPlus       *utils.PushPlusNotifier
	tokenReader    *analyzer.TokenInfoReader
}

// NewContractDeploymentPlugin 创建合约部署监听插件
func NewContractDeploymentPlugin(rpcURL string) (*ContractDeploymentPlugin, error) {
	// 创建代币信息读取器
	tokenReader, err := analyzer.NewTokenInfoReader(rpcURL)
	if err != nil {
		return nil, err
	}

	// 创建通知器
	var pushPlus *utils.PushPlusNotifier
	if token := os.Getenv("PUSHPLUS_TOKEN"); token != "" {
		pushPlus = utils.NewPushPlusNotifier(token)
	}

	// 暂时不创建 Meme 币分析器（需要 GoPlus API Key）
	// goPlusAPIKey := os.Getenv("GOPLUS_API_KEY")
	// memeAnalyzer, err := analyzer.NewMemeTokenAnalyzer(rpcURL, goPlusAPIKey)
	// if err != nil {
	// 	return nil, err
	// }

	return &ContractDeploymentPlugin{
		deploymentRepo: database.NewContractDeploymentRepository(),
		analyzer:       nil, // 暂时设为 nil
		pushPlus:       pushPlus,
		tokenReader:    tokenReader,
	}, nil
}

// AcceptTx 处理交易（检测合约部署）
func (p *ContractDeploymentPlugin) AcceptTx(tx structs.RemovableTx) {
	// 检查是否是合约部署交易（to 地址为空）
	if tx.GetTo() != "" {
		return // 不是合约部署
	}

	logger.Log.Debug("检测到合约部署交易",
		zap.String("hash", tx.GetHash()),
		zap.String("from", tx.GetFrom()),
		zap.Uint64("block", tx.GetBlockNumber()))

	// 注意：在这个阶段我们还不知道合约地址
	// 需要在 Accept 方法中通过 receipt 获取
}

// Accept 处理交易和回执（获取合约地址并分析）
func (p *ContractDeploymentPlugin) Accept(txAndReceipt *structs.RemovableTxAndReceipt) {
	tx := txAndReceipt.Tx
	receipt := txAndReceipt.Receipt

	// 再次检查是否是合约部署
	if tx.GetTo() != "" {
		return
	}

	// 检查交易是否成功
	if !receipt.GetResult() {
		logger.Log.Debug("合约部署交易失败，跳过",
			zap.String("txHash", tx.GetHash()))
		return // 交易失败
	}

	// 对于合约部署交易，从日志中获取合约地址
	// 大多数合约部署都会产生日志，日志的 address 字段就是合约地址
	logs := receipt.GetLogs()
	var contractAddress string

	if len(logs) > 0 {
		// 从第一个日志获取合约地址
		contractAddress = logs[0].GetAddress()
		logger.Log.Debug("从日志中获取合约地址",
			zap.String("address", contractAddress),
			zap.String("txHash", tx.GetHash()))
	} else {
		// 如果没有日志，说明这个合约部署可能没有触发任何事件
		// 这种情况下我们无法获取合约地址，跳过
		logger.Log.Debug("合约部署无日志，无法获取合约地址",
			zap.String("txHash", tx.GetHash()),
			zap.String("from", tx.GetFrom()))
		return
	}

	if contractAddress == "" {
		logger.Log.Debug("无法获取合约地址，跳过",
			zap.String("txHash", tx.GetHash()))
		return
	}

	logger.Log.Info("✅ 检测到合约部署",
		zap.String("address", contractAddress),
		zap.String("txHash", tx.GetHash()),
		zap.String("deployer", tx.GetFrom()),
		zap.Uint64("block", tx.GetBlockNumber()))

	// 暂时跳过 ERC20 检测（会产生大量 RPC 请求）
	// TODO: 优化 ERC20 检测逻辑，添加批量查询或缓存
	isToken := false
	// isToken := p.tokenReader.IsERC20Token(contractAddress)

	// 保存部署记录
	deployment := &model.ContractDeployment{
		ContractAddress: contractAddress,
		DeployerAddress: tx.GetFrom(),
		TxHash:          tx.GetHash(),
		BlockNumber:     tx.GetBlockNumber(),
		Timestamp:       time.Unix(int64(txAndReceipt.TimeStamp), 0),
		IsToken:         isToken,
		ContractType:    "Unknown",
	}

	// 暂时注释掉代币分析
	// if isToken {
	// 	deployment.ContractType = "ERC20"
	// 	logger.Log.Info("🎯 检测到 ERC20 代币部署",
	// 		zap.String("address", contractAddress),
	// 		zap.String("txHash", tx.GetHash()))
	//
	// 	// 异步分析代币
	// 	go p.analyzeNewToken(contractAddress)
	// }

	if err := p.deploymentRepo.Create(deployment); err != nil {
		logger.Log.Error("保存部署记录失败", zap.Error(err))
	}
}

// analyzeNewToken 分析新代币
func (p *ContractDeploymentPlugin) analyzeNewToken(tokenAddress string) {
	// 等待一段时间，让合约初始化完成
	time.Sleep(30 * time.Second)

	logger.Log.Info("开始分析新代币", zap.String("address", tokenAddress))

	// 暂时跳过完整分析，只记录日志
	// TODO: 实现完整的代币分析
	logger.Log.Info("代币分析功能开发中",
		zap.String("address", tokenAddress))

	/* 完整分析代码（需要 GoPlus API Key）
	analysis, err := p.analyzer.AnalyzeToken(tokenAddress)
	if err != nil {
		logger.Log.Error("代币分析失败",
			zap.String("address", tokenAddress),
			zap.Error(err))
		return
	}

	// 检查是否是潜力币
	if p.analyzer.IsPotentialGem(analysis) {
		p.sendPotentialGemAlert(analysis)
	} else if p.analyzer.IsLowRiskToken(analysis) {
		p.sendLowRiskTokenAlert(analysis)
	}
	*/
}

// sendPotentialGemAlert 发送潜力币告警
func (p *ContractDeploymentPlugin) sendPotentialGemAlert(analysis *model.TokenAnalysis) {
	logger.Log.Info("🎯 发现潜力 Meme 币！",
		zap.String("symbol", analysis.Symbol),
		zap.Float64("riskScore", analysis.RiskScore))

	if p.pushPlus == nil {
		return
	}

	title := "🎯 发现潜力 Meme 币: " + analysis.Symbol
	content := p.analyzer.GenerateReport(analysis)
	content += "\n\n💎 这是一个低风险且有潜力的新币！"
	content += "\n\n**合约地址**: `" + analysis.TokenAddress + "`"
	content += "\n**Etherscan**: https://etherscan.io/address/" + analysis.TokenAddress

	if err := p.pushPlus.SendCustomAlert(title, content); err != nil {
		logger.Log.Error("发送告警失败", zap.Error(err))
	}
}

// sendLowRiskTokenAlert 发送低风险代币告警
func (p *ContractDeploymentPlugin) sendLowRiskTokenAlert(analysis *model.TokenAnalysis) {
	logger.Log.Info("✅ 发现低风险新币",
		zap.String("symbol", analysis.Symbol),
		zap.Float64("riskScore", analysis.RiskScore))

	// 低风险但不是潜力币，只记录日志，不发送告警
	// 如果想要告警，可以取消下面的注释
	/*
		if p.pushPlus != nil {
			title := "✅ 发现低风险新币: " + analysis.Symbol
			content := p.analyzer.GenerateReport(analysis)
			p.pushPlus.SendCustomAlert(title, content)
		}
	*/
}

// Close 关闭资源
func (p *ContractDeploymentPlugin) Close() {
	// if p.analyzer != nil {
	// 	p.analyzer.Close()
	// }
	if p.tokenReader != nil {
		p.tokenReader.Close()
	}
}

// PairCreatedPlugin Uniswap PairCreated 事件监听插件
type PairCreatedPlugin struct {
	deploymentRepo *database.ContractDeploymentRepository
	analyzer       *analyzer.MemeTokenAnalyzer
	pushPlus       *utils.PushPlusNotifier
}

// NewPairCreatedPlugin 创建 PairCreated 事件监听插件
func NewPairCreatedPlugin(rpcURL string) (*PairCreatedPlugin, error) {
	// 创建通知器
	var pushPlus *utils.PushPlusNotifier
	if token := os.Getenv("PUSHPLUS_TOKEN"); token != "" {
		pushPlus = utils.NewPushPlusNotifier(token)
	}

	// 创建 Meme 币分析器
	// 注意：即使没有 GoPlus API Key，分析器也可以工作，只是蜜罐检测会失败
	goPlusAPIKey := os.Getenv("GOPLUS_API_KEY")
	memeAnalyzer, err := analyzer.NewMemeTokenAnalyzer(rpcURL, goPlusAPIKey)
	if err != nil {
		logger.Log.Warn("创建 Meme 币分析器失败，将跳过代币分析", zap.Error(err))
		// 不返回错误，继续创建插件
		return &PairCreatedPlugin{
			deploymentRepo: database.NewContractDeploymentRepository(),
			analyzer:       nil,
			pushPlus:       pushPlus,
		}, nil
	}

	return &PairCreatedPlugin{
		deploymentRepo: database.NewContractDeploymentRepository(),
		analyzer:       memeAnalyzer,
		pushPlus:       pushPlus,
	}, nil
}

// Accept 处理 PairCreated 事件
func (p *PairCreatedPlugin) Accept(log *structs.RemovableReceiptLog) {
	logger.Log.Info("收到 PairCreated 事件")
	if log.IsRemoved {
		return
	}

	topics := log.GetTopics()
	if len(topics) < 3 {
		return
	}

	// PairCreated(address indexed token0, address indexed token1, address pair, uint)
	// topics[0] = 事件签名
	// topics[1] = token0
	// topics[2] = token1
	// data = pair address + pair index

	token0 := extractAddress(topics[1])
	token1 := extractAddress(topics[2])

	// 判断哪个是 WETH，哪个是新代币
	wethAddress := strings.ToLower(config.WETHAddress)
	var newTokenAddress string

	if strings.ToLower(token0) == wethAddress {
		newTokenAddress = token1
	} else if strings.ToLower(token1) == wethAddress {
		newTokenAddress = token0
	} else {
		// 不是 ETH 交易对，跳过
		return
	}

	logger.Log.Info("🔥 检测到新的 ETH 交易对创建",
		zap.String("token", newTokenAddress),
		zap.String("txHash", log.GetTransactionHash()),
		zap.Int("block", log.GetBlockNum()))

	// 异步分析代币
	go p.analyzeNewPairToken(newTokenAddress)
}

// analyzeNewPairToken 分析新交易对的代币
func (p *PairCreatedPlugin) analyzeNewPairToken(tokenAddress string) {
	// 等待流动性添加完成
	time.Sleep(60 * time.Second)

	logger.Log.Info("开始分析新交易对代币", zap.String("address", tokenAddress))

	// 检查分析器是否初始化
	if p.analyzer == nil {
		logger.Log.Warn("分析器未初始化，跳过分析",
			zap.String("address", tokenAddress),
			zap.String("提示", "请配置 GOPLUS_API_KEY 以启用完整分析"))
		return
	}

	// 执行代币分析
	analysis, err := p.analyzer.AnalyzeToken(tokenAddress)
	if err != nil {
		logger.Log.Error("代币分析失败",
			zap.String("address", tokenAddress),
			zap.Error(err))
		return
	}

	logger.Log.Info("✅ 代币分析完成",
		zap.String("symbol", analysis.Symbol),
		zap.String("name", analysis.Name),
		zap.Float64("riskScore", analysis.RiskScore),
		zap.String("riskLevel", analysis.RiskLevel))

	// 检查是否是潜力币
	if p.analyzer.IsPotentialGem(analysis) {
		p.sendPotentialGemAlert(analysis)
	} else if p.analyzer.IsLowRiskToken(analysis) {
		logger.Log.Info("✅ 发现低风险新交易对",
			zap.String("symbol", analysis.Symbol),
			zap.Float64("riskScore", analysis.RiskScore))
	}
}

// sendPotentialGemAlert 发送潜力币告警
func (p *PairCreatedPlugin) sendPotentialGemAlert(analysis *model.TokenAnalysis) {
	logger.Log.Info("🎯 发现潜力 Meme 币（新交易对）！",
		zap.String("symbol", analysis.Symbol),
		zap.Float64("riskScore", analysis.RiskScore))

	if p.pushPlus == nil {
		return
	}

	title := "🔥 新交易对创建: " + analysis.Symbol
	content := p.analyzer.GenerateReport(analysis)
	content += "\n\n💎 这是一个低风险且有潜力的新币！"
	content += "\n\n**合约地址**: `" + analysis.TokenAddress + "`"
	content += "\n**Etherscan**: https://etherscan.io/address/" + analysis.TokenAddress
	content += "\n**Uniswap**: https://app.uniswap.org/#/swap?outputCurrency=" + analysis.TokenAddress

	if err := p.pushPlus.SendCustomAlert(title, content); err != nil {
		logger.Log.Error("发送告警失败", zap.Error(err))
	}
}

// FromContract 返回监听的合约地址
func (p *PairCreatedPlugin) FromContract() string {
	return config.UniswapV2FactoryAddress
}

// InterestedTopics 返回感兴趣的事件主题
func (p *PairCreatedPlugin) InterestedTopics() []string {
	return []string{config.UniswapV2PairCreatedTopic}
}

// NeedReceiptLog 是否需要处理该日志
func (p *PairCreatedPlugin) NeedReceiptLog(receiptLog *structs.RemovableReceiptLog) bool {
	return true
}

// Close 关闭资源
func (p *PairCreatedPlugin) Close() {
	if p.analyzer != nil {
		p.analyzer.Close()
	}
}

// extractAddress 从 Topic 中提取地址
func extractAddress(topic string) string {
	// Topic 是 32 字节，地址是后 20 字节
	if len(topic) >= 66 { // "0x" + 64 个字符
		return "0x" + topic[26:] // 跳过前 26 个字符（0x + 24个0）
	}
	return topic
}
