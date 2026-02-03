package scheduler

import (
	"ethereum-monitor/analyzer"
	"ethereum-monitor/config"
	"ethereum-monitor/database"
	"ethereum-monitor/logger"
	"ethereum-monitor/model"
	"time"

	"go.uber.org/zap"
)

type LiquidityScanner struct {
	repo              *database.TokenAnalysisRepository
	liquidityAnalyzer *analyzer.LiquidityAnalyzer
	tokenReader       *analyzer.TokenInfoReader
}

func NewLiquidityScanner(rpcURL string) (*LiquidityScanner, error) {
	la, err := analyzer.NewLiquidityAnalyzer(rpcURL)
	if err != nil {
		return nil, err
	}

	tr, err := analyzer.NewTokenInfoReader(rpcURL)
	if err != nil {
		return nil, err
	}

	return &LiquidityScanner{
		repo:              database.NewTokenAnalysisRepository(),
		liquidityAnalyzer: la,
		tokenReader:       tr,
	}, nil
}

// Run 执行一次扫描
func (s *LiquidityScanner) Run() {
	// 获取待处理的代币
	// 每次处理 20 个，避免 RPC 压力过大
	tokens, err := s.repo.GetPendingLiquidityTokens(20)
	if err != nil {
		logger.Log.Error("获取待扫描代币失败", zap.Error(err))
		return
	}

	if len(tokens) == 0 {
		return
	}

	logger.Log.Info("开始扫描流动性", zap.Int("pending_count", len(tokens)))

	for _, token := range tokens {
		s.processToken(&token)
	}
}

func (s *LiquidityScanner) processToken(t *model.TokenAnalysis) {
	// 1. 检查流动性
	liqUSD, _, err := s.liquidityAnalyzer.GetLiquidityInfo(t.PairAddress, t.TokenAddress)
	if err != nil {
		logger.Log.Warn("获取流动性失败", zap.String("token", t.TokenAddress), zap.Error(err))
		// 暂时不处理错误，等待下一次重试
		return
	}

	// 更新最后检查时间
	t.LastCheckAt = time.Now()
	t.LiquidityUSD = liqUSD

	// 2. 判断流动性是否达标
	// 阈值：例如 $5000 (config.MemeMinLiquidityUSD)
	if liqUSD < config.MemeMinLiquidityUSD {
		// 流动性不足

		// 检查是否超时（例如 2 小时未加池）
		if time.Since(t.PairCreatedAt) > 2*time.Hour {
			t.Status = "REJECTED"
			t.RiskFlags = `["timeout_no_liquidity"]`
			logger.Log.Info("🗑️ 代币超时未加池，已丢弃", zap.String("symbol", t.Symbol), zap.String("addr", t.TokenAddress))
			s.repo.Update(t)
		} else {
			// 还没超时，只更新 LastCheckAt，保持 PENDING 状态
			if liqUSD > 100 {
				logger.Log.Debug("流动性不足但非零", zap.String("addr", t.TokenAddress), zap.Float64("usd", liqUSD))
			}
			s.repo.Update(t)
		}
		return
	}

	// 3. 流动性达标！开始处理
	t.HasLiquidity = true
	t.LiquidityAddedAt = time.Now()
	t.InitialMarketCap = liqUSD // 粗略估算，假设全流通

	// 4. 补充基本信息 (Name, Symbol)
	info, err := s.tokenReader.ReadTokenInfo(t.TokenAddress)
	if err == nil && info.IsValid {
		t.Name = info.Name
		t.Symbol = info.Symbol
		t.Decimals = info.Decimals
		t.TotalSupply = info.TotalSupply.String()
	} else {
		logger.Log.Warn("读取代币信息失败", zap.String("token", t.TokenAddress))
		// 即使读取失败，也继续推进，可能网络抖动
	}

	// 5. 状态流转 -> ANALYZING
	t.Status = "ANALYZING"

	logger.Log.Info("💧 发现流动性达标代币",
		zap.String("symbol", t.Symbol),
		zap.Float64("liquidity", liqUSD),
		zap.String("eth", "ETH")) // ethAmount undefined in logging context? No, valid var.

	if err := s.repo.Update(t); err != nil {
		logger.Log.Error("更新代币状态失败", zap.Error(err))
	}
}

// Close 关闭资源
func (s *LiquidityScanner) Close() {
	s.liquidityAnalyzer.Close()
	s.tokenReader.Close()
}
