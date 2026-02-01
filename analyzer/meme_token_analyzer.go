package analyzer

import (
	"encoding/json"
	"ethereum-monitor/database"
	"ethereum-monitor/logger"
	"ethereum-monitor/model"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// MemeTokenAnalyzer Meme 币分析器
type MemeTokenAnalyzer struct {
	tokenReader      *TokenInfoReader
	honeypotDetector *HoneypotDetector
	riskScorer       *TokenRiskScorer
	tokenRepo        *database.TokenAnalysisRepository
}

// NewMemeTokenAnalyzer 创建 Meme 币分析器
func NewMemeTokenAnalyzer(rpcURL, goPlusAPIKey string) (*MemeTokenAnalyzer, error) {
	tokenReader, err := NewTokenInfoReader(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create token reader: %w", err)
	}

	return &MemeTokenAnalyzer{
		tokenReader:      tokenReader,
		honeypotDetector: NewHoneypotDetector(goPlusAPIKey),
		riskScorer:       NewTokenRiskScorer(),
		tokenRepo:        database.NewTokenAnalysisRepository(),
	}, nil
}

// AnalyzeToken 分析代币
func (a *MemeTokenAnalyzer) AnalyzeToken(tokenAddress string) (*model.TokenAnalysis, error) {
	logger.Log.Info("开始分析代币", zap.String("address", tokenAddress))

	analysis := &model.TokenAnalysis{
		TokenAddress: tokenAddress,
		AnalyzedAt:   time.Now(),
	}

	// 1. 读取代币基本信息
	tokenInfo, err := a.tokenReader.ReadTokenInfo(tokenAddress)
	if err != nil || !tokenInfo.IsValid {
		logger.Log.Warn("无法读取代币信息，可能不是有效的 ERC20 代币",
			zap.String("address", tokenAddress),
			zap.Error(err))
		return nil, fmt.Errorf("invalid ERC20 token")
	}

	analysis.Name = tokenInfo.Name
	analysis.Symbol = tokenInfo.Symbol
	analysis.Decimals = tokenInfo.Decimals
	analysis.TotalSupply = tokenInfo.TotalSupply.String()

	logger.Log.Info("代币基本信息",
		zap.String("name", analysis.Name),
		zap.String("symbol", analysis.Symbol),
		zap.Uint8("decimals", analysis.Decimals))

	// 2. 检查合约验证状态（需要 Etherscan API，这里简化处理）
	analysis.IsVerified = false // TODO: 集成 Etherscan API

	// 3. 蜜罐检测
	honeypotResult, err := a.honeypotDetector.CheckHoneypot(tokenAddress)
	if err != nil {
		logger.Log.Warn("蜜罐检测失败", zap.Error(err))
		// 继续分析，不中断
	} else {
		analysis.IsHoneypot = honeypotResult.IsHoneypot
		analysis.HoneypotReason = honeypotResult.Reason
		analysis.BuyTax = honeypotResult.BuyTax
		analysis.SellTax = honeypotResult.SellTax

		logger.Log.Info("蜜罐检测结果",
			zap.Bool("isHoneypot", analysis.IsHoneypot),
			zap.Float64("buyTax", analysis.BuyTax),
			zap.Float64("sellTax", analysis.SellTax))
	}

	// 4. 检查流动性（简化版，实际需要查询 Uniswap）
	// TODO: 实现流动性检查
	analysis.HasLiquidity = false
	analysis.LiquidityUSD = 0
	analysis.InitialMarketCap = 0

	// 5. 持有者分析（需要 Etherscan API 或链上查询）
	// TODO: 实现持有者分析
	analysis.HolderCount = 0
	analysis.Top10HoldingPct = 0

	// 6. 所有权检查
	// TODO: 实现所有权检查
	analysis.IsOwnershipRenounced = false

	// 7. 计算风险评分
	score, level, flags := a.riskScorer.CalculateRiskScore(analysis)
	analysis.RiskScore = score
	analysis.RiskLevel = level

	// 将风险标志转换为 JSON
	flagsJSON, _ := json.Marshal(flags)
	analysis.RiskFlags = string(flagsJSON)

	logger.Log.Info("风险评分完成",
		zap.Float64("score", score),
		zap.String("level", level),
		zap.Int("flagCount", len(flags)))

	// 8. 保存到数据库
	if err := a.tokenRepo.Create(analysis); err != nil {
		logger.Log.Error("保存代币分析失败", zap.Error(err))
		return analysis, err
	}

	logger.Log.Info("代币分析完成", zap.String("symbol", analysis.Symbol))

	return analysis, nil
}

// IsLowRiskToken 判断是否是低风险代币
func (a *MemeTokenAnalyzer) IsLowRiskToken(analysis *model.TokenAnalysis) bool {
	return a.riskScorer.IsLowRisk(analysis)
}

// IsPotentialGem 判断是否是潜力币
func (a *MemeTokenAnalyzer) IsPotentialGem(analysis *model.TokenAnalysis) bool {
	return a.riskScorer.IsPotentialGem(analysis)
}

// GenerateReport 生成分析报告
func (a *MemeTokenAnalyzer) GenerateReport(analysis *model.TokenAnalysis) string {
	return a.riskScorer.GenerateRiskReport(analysis)
}

// Close 关闭资源
func (a *MemeTokenAnalyzer) Close() {
	if a.tokenReader != nil {
		a.tokenReader.Close()
	}
}
