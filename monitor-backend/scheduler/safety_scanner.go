package scheduler

import (
	"ethereum-monitor/analyzer"
	"ethereum-monitor/database"
	"ethereum-monitor/logger"
	"ethereum-monitor/model"
	"ethereum-monitor/utils" // 确保有这个 utils
	"fmt"
	"os"
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
	// 获取待分析的代币 (ANALYZING)
	tokens, err := s.repo.GetByStatus("ANALYZING", 10)
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
	// 执行安全检测
	if err := s.memeAnalyzer.AnalyzeSafetyOnly(t); err != nil {
		// 如果 API 失败，暂不改变状态，等待重试
		return
	}

	// 判定结果
	if t.IsHoneypot || t.RiskLevel == "critical" {
		t.Status = "REJECTED"
		logger.Log.Info("⛔ 拒绝高风险/蜜罐代币",
			zap.String("symbol", t.Symbol),
			zap.String("reason", t.HoneypotReason))
	} else {
		// 通过！进入观察期
		t.Status = "MONITORING"
		logger.Log.Info("✅ 代币通过安全检测，进入观察列表",
			zap.String("symbol", t.Symbol),
			zap.Float64("score", t.RiskScore))

		// 发送初次上线通知
		s.sendNewTokenAlert(t)
	}

	t.AnalyzedAt = time.Now()
	if err := s.repo.Update(t); err != nil {
		logger.Log.Error("更新代币分析结果失败", zap.Error(err))
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
	content += fmt.Sprintf("**风险分**: %.1f (%s)\n", t.RiskScore, t.RiskLevel)

	if t.RiskLevel == "low" {
		content += "\n✅ **低风险** - 值得关注!\n"
	}

	content += "\n[Etherscan](https://etherscan.io/address/" + t.TokenAddress + ") | "
	content += "[Uniswap](https://app.uniswap.org/#/swap?outputCurrency=" + t.TokenAddress + ")"

	go s.notifier.SendCustomAlert(title, content)
}

// Close 关闭资源
func (s *SafetyScanner) Close() {
	s.memeAnalyzer.Close()
}
