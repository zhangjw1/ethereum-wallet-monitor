package scheduler

import (
	"ethereum-monitor/analyzer"
	"ethereum-monitor/database"
	"ethereum-monitor/logger"
	"ethereum-monitor/model"
	"ethereum-monitor/utils" // 确保有这个 utils
	"fmt"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
)

type SafetyScanner struct {
	repo         *database.TokenAnalysisRepository
	memeAnalyzer *analyzer.MemeTokenAnalyzer
	notifier     *utils.PushPlusNotifier
}

func NewSafetyScanner(rpcURL, goPlusKey string) (*SafetyScanner, error) {
	ma, err := analyzer.NewMemeTokenAnalyzer(rpcURL, goPlusKey)
	if err != nil {
		return nil, err
	}

	var notifier *utils.PushPlusNotifier
	if token := os.Getenv("PUSHPLUS_TOKEN"); token != "" {
		notifier = utils.NewPushPlusNotifier(token)
	}

	return &SafetyScanner{
		repo:         database.NewTokenAnalysisRepository(),
		memeAnalyzer: ma,
		notifier:     notifier,
	}, nil
}

func (s *SafetyScanner) Run() {
	// 获取待分析的代币 (ANALYZING 或 需要重试的 MONITORING)
	tokens, err := s.repo.GetTokensForSafetyCheck(10)
	if err != nil {
		logger.Log.Error("获取待安全分析代币失败", zap.Error(err))
		return
	}

	if len(tokens) == 0 {
		return
	}

	logger.Log.Info("开始安全分析", zap.Int("count", len(tokens)))

	for _, token := range tokens {
		s.processToken(&token)
	}
}

func (s *SafetyScanner) processToken(t *model.TokenAnalysis) {
	oldStatus := t.Status
	oldSafetyStatus := t.SafetyStatus

	// 执行安全检测
	if err := s.memeAnalyzer.AnalyzeSafetyOnly(t); err != nil {
		// 如果 API 失败（网络错误），暂不改变状态，等待重试
		return
	}

	// 检查是否是 API 数据未找到 (404)
	isDataNotFound := false
	if t.HoneypotReason != "" && (strings.Contains(t.HoneypotReason, "not found") || strings.Contains(t.HoneypotReason, "too new")) {
		isDataNotFound = true
	}

	shouldNotify := false

	if isDataNotFound {
		// 数据未找到，Token 太新
		// 策略：放入 MONITORING 列表，但标记需要重试
		t.Status = "MONITORING"
		t.SafetyStatus = "RETRY_NEEDED"
		t.RiskLevel = "unknown"

		// 只有从 ANALYZING 变为 MONITORING 时才通知
		if oldStatus == "ANALYZING" {
			shouldNotify = true
		}
		logger.Log.Info("⚠️ 代币安全数据未找到，暂时放行并标记重试", zap.String("symbol", t.Symbol))

	} else if t.IsHoneypot || t.RiskLevel == "critical" {
		t.Status = "REJECTED"
		t.SafetyStatus = "COMPLETED"
		logger.Log.Info("⛔ 拒绝高风险/蜜罐代币",
			zap.String("symbol", t.Symbol),
			zap.String("reason", t.HoneypotReason))
	} else {
		// 通过！
		t.Status = "MONITORING"
		t.SafetyStatus = "COMPLETED"

		if oldStatus == "ANALYZING" {
			shouldNotify = true
		} else if oldStatus == "MONITORING" && oldSafetyStatus == "RETRY_NEEDED" {
			// 重试成功！
			logger.Log.Info("✅ 代币重试检测通过", zap.String("symbol", t.Symbol))
			// 可选：发送更新通知，或者单纯记录
			// shouldNotify = true
		}

		logger.Log.Info("✅ 代币通过安全检测",
			zap.String("symbol", t.Symbol),
			zap.Float64("score", t.RiskScore))
	}

	t.AnalyzedAt = time.Now()
	if err := s.repo.Update(t); err != nil {
		logger.Log.Error("更新代币分析结果失败", zap.Error(err))
	}

	if shouldNotify {
		s.sendNewTokenAlert(t)
	}
}

func (s *SafetyScanner) sendNewTokenAlert(t *model.TokenAnalysis) {
	if s.notifier == nil {
		return
	}

	title := "👀 新币上线: " + t.Symbol
	content := "### 发现新 Token 上线 (已过初筛)\n\n"
	content += "**名称**: " + t.Name + "\n"
	content += "**合约**: `" + t.TokenAddress + "`\n"
	content += fmt.Sprintf("**流动性**: $%.0f\n", t.LiquidityUSD)

	if t.SafetyStatus == "RETRY_NEEDED" {
		content += "\n⚠️ **风险未知** (API未收录)\n"
		content += "系统将持续扫描，请谨慎操作。\n"
	} else {
		content += fmt.Sprintf("**风险分**: %.1f (%s)\n", t.RiskScore, t.RiskLevel)
		if t.RiskLevel == "low" {
			content += "\n✅ **低风险** - 值得关注!\n"
		}
	}

	content += "\n[Etherscan](https://etherscan.io/address/" + t.TokenAddress + ") | "
	content += "[Uniswap](https://app.uniswap.org/#/swap?outputCurrency=" + t.TokenAddress + ")"

	go s.notifier.SendCustomAlert(title, content)
}

// Close 关闭资源
func (s *SafetyScanner) Close() {
	s.memeAnalyzer.Close()
}
