package analyzer

import (
	"encoding/json"
	"ethereum-monitor/config"
	"ethereum-monitor/model"
	"fmt"
)

// TokenRiskScorer 代币风险评分器
type TokenRiskScorer struct{}

// NewTokenRiskScorer 创建风险评分器
func NewTokenRiskScorer() *TokenRiskScorer {
	return &TokenRiskScorer{}
}

// CalculateRiskScore 计算风险评分
func (s *TokenRiskScorer) CalculateRiskScore(analysis *model.TokenAnalysis) (float64, string, []string) {
	score := 0.0
	riskFlags := []string{}

	// 1. 未验证合约 +30
	if !analysis.IsVerified {
		score += config.RiskScoreUnverified
		riskFlags = append(riskFlags, "合约未验证")
	}

	// 2. 蜜罐 +50
	if analysis.IsHoneypot {
		score += config.RiskScoreHoneypot
		riskFlags = append(riskFlags, "⚠️ 检测到蜜罐: "+analysis.HoneypotReason)
	}

	// 3. 高税率 (>10%) +20
	if analysis.BuyTax > config.HighTaxThreshold {
		score += config.RiskScoreHighTax
		riskFlags = append(riskFlags, "买入税过高: "+formatPercent(analysis.BuyTax))
	}
	if analysis.SellTax > config.HighTaxThreshold {
		score += config.RiskScoreHighTax
		riskFlags = append(riskFlags, "卖出税过高: "+formatPercent(analysis.SellTax))
	}

	// 4. 持有者过度集中 (>50%) +25
	if analysis.Top10HoldingPct > config.Top10HoldingThreshold {
		score += config.RiskScoreConcentratedHolding
		riskFlags = append(riskFlags, "持有者过度集中: 前10持有"+formatPercent(analysis.Top10HoldingPct))
	}

	// 5. 无流动性 +40
	if !analysis.HasLiquidity {
		score += config.RiskScoreNoLiquidity
		riskFlags = append(riskFlags, "无流动性")
	} else if analysis.LiquidityUSD < config.MemeMinLiquidityUSD {
		score += config.RiskScoreNoLiquidity / 2 // 流动性不足 +20
		riskFlags = append(riskFlags, "流动性不足: $"+formatFloat(analysis.LiquidityUSD))
	}

	// 6. 未放弃所有权 +15
	if !analysis.IsOwnershipRenounced && analysis.OwnerAddress != "" {
		score += config.RiskScoreNotRenounced
		riskFlags = append(riskFlags, "未放弃所有权")
	}

	// 限制最大值为 100
	if score > 100 {
		score = 100
	}

	// 确定风险等级
	riskLevel := s.determineRiskLevel(score)

	return score, riskLevel, riskFlags
}

// determineRiskLevel 确定风险等级
func (s *TokenRiskScorer) determineRiskLevel(score float64) string {
	if score < 20 {
		return "low"
	} else if score < 40 {
		return "medium"
	} else if score < 70 {
		return "high"
	}
	return "critical"
}

// IsLowRisk 判断是否是低风险代币
func (s *TokenRiskScorer) IsLowRisk(analysis *model.TokenAnalysis) bool {
	return analysis.RiskScore < config.MemeRiskScoreThresholdLow
}

// IsPotentialGem 判断是否是潜力币
func (s *TokenRiskScorer) IsPotentialGem(analysis *model.TokenAnalysis) bool {
	// 低风险 + 初始市值合理 + 有流动性
	return s.IsLowRisk(analysis) &&
		analysis.InitialMarketCap > 0 &&
		analysis.InitialMarketCap < config.MemeMarketCapThreshold &&
		analysis.HasLiquidity &&
		analysis.LiquidityUSD >= config.MemeMinLiquidityUSD
}

// GenerateRiskReport 生成风险报告
func (s *TokenRiskScorer) GenerateRiskReport(analysis *model.TokenAnalysis) string {
	report := "🔍 代币风险分析报告\n\n"
	report += "📊 基本信息:\n"
	report += "名称: " + analysis.Name + "\n"
	report += "符号: " + analysis.Symbol + "\n"
	report += "地址: " + analysis.TokenAddress + "\n\n"

	report += "⚠️ 风险评分: " + formatFloat(analysis.RiskScore) + "/100\n"
	report += "风险等级: " + getRiskLevelEmoji(analysis.RiskLevel) + " " + analysis.RiskLevel + "\n\n"

	// 解析风险标志
	var riskFlags []string
	if analysis.RiskFlags != "" {
		json.Unmarshal([]byte(analysis.RiskFlags), &riskFlags)
	}

	if len(riskFlags) > 0 {
		report += "🚩 风险标志:\n"
		for _, flag := range riskFlags {
			report += "  • " + flag + "\n"
		}
		report += "\n"
	}

	report += "💰 流动性: $" + formatFloat(analysis.LiquidityUSD) + "\n"
	report += "📈 初始市值: $" + formatFloat(analysis.InitialMarketCap) + "\n"
	report += "👥 持有者数量: " + formatInt(analysis.HolderCount) + "\n"
	report += "💸 买入税: " + formatPercent(analysis.BuyTax) + "\n"
	report += "💸 卖出税: " + formatPercent(analysis.SellTax) + "\n"

	return report
}

// 辅助函数
func formatPercent(value float64) string {
	return fmt.Sprintf("%.2f%%", value)
}

func formatFloat(value float64) string {
	return fmt.Sprintf("%.2f", value)
}

func formatInt(value int) string {
	return fmt.Sprintf("%d", value)
}

func getRiskLevelEmoji(level string) string {
	switch level {
	case "low":
		return "✅"
	case "medium":
		return "⚠️"
	case "high":
		return "🔴"
	case "critical":
		return "💀"
	default:
		return "❓"
	}
}
