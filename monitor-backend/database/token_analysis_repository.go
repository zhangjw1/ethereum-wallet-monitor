package database

import (
	"ethereum-monitor/model"
	"time"
)

// TokenAnalysisRepository 代币分析数据访问层
type TokenAnalysisRepository struct{}

// NewTokenAnalysisRepository 创建 Repository
func NewTokenAnalysisRepository() *TokenAnalysisRepository {
	return &TokenAnalysisRepository{}
}

// Create 创建代币分析记录
func (r *TokenAnalysisRepository) Create(analysis *model.TokenAnalysis) error {
	return DB.Create(analysis).Error
}

// Update 更新代币分析记录
func (r *TokenAnalysisRepository) Update(analysis *model.TokenAnalysis) error {
	return DB.Save(analysis).Error
}

// GetByAddress 根据代币地址查询
func (r *TokenAnalysisRepository) GetByAddress(address string) (*model.TokenAnalysis, error) {
	var analysis model.TokenAnalysis
	err := DB.Where("token_address = ?", address).First(&analysis).Error
	return &analysis, err
}

// GetLowRiskTokens 获取低风险代币
func (r *TokenAnalysisRepository) GetLowRiskTokens(maxRiskScore float64, limit int) ([]model.TokenAnalysis, error) {
	var tokens []model.TokenAnalysis
	err := DB.Where("risk_score <= ?", maxRiskScore).
		Order("analyzed_at DESC").
		Limit(limit).
		Find(&tokens).Error
	return tokens, err
}

// GetRecentAnalyses 获取最近分析的代币
func (r *TokenAnalysisRepository) GetRecentAnalyses(limit int) ([]model.TokenAnalysis, error) {
	var tokens []model.TokenAnalysis
	err := DB.Order("analyzed_at DESC").
		Limit(limit).
		Find(&tokens).Error
	return tokens, err
}

// GetByStatus 根据状态查询
func (r *TokenAnalysisRepository) GetByStatus(status string, limit int) ([]model.TokenAnalysis, error) {
	var tokens []model.TokenAnalysis
	err := DB.Where("status = ?", status).
		Order("pair_created_at DESC").
		Limit(limit).
		Find(&tokens).Error
	return tokens, err
}

// GetPendingLiquidityTokens 获取待扫描流动性的代币
func (r *TokenAnalysisRepository) GetPendingLiquidityTokens(limit int) ([]model.TokenAnalysis, error) {
	// 查找状态为 PENDING_LIQUIDITY 且创建时间在 2 小时以内的
	// 如果超过 2 小时还没加池，可能是死币，暂不优先扫描（后续由过期任务清理）
	twoHoursAgo := time.Now().Add(-2 * time.Hour)

	var tokens []model.TokenAnalysis
	err := DB.Where("status = ? AND pair_created_at > ?", "PENDING_LIQUIDITY", twoHoursAgo).
		Order("pair_created_at ASC"). // 按时间正序，优先处理最早的
		Limit(limit).
		Find(&tokens).Error
	return tokens, err
}
func (r *TokenAnalysisRepository) GetByRiskLevel(riskLevel string, limit int) ([]model.TokenAnalysis, error) {
	var tokens []model.TokenAnalysis
	err := DB.Where("risk_level = ?", riskLevel).
		Order("analyzed_at DESC").
		Limit(limit).
		Find(&tokens).Error
	return tokens, err
}

// GetDailyStats 获取每日统计
func (r *TokenAnalysisRepository) GetDailyStats(date time.Time) (map[string]interface{}, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	stats := make(map[string]interface{})

	// 总数
	var total int64
	DB.Model(&model.TokenAnalysis{}).
		Where("analyzed_at BETWEEN ? AND ?", startOfDay, endOfDay).
		Count(&total)
	stats["total"] = total

	// 按风险等级统计
	var riskStats []struct {
		RiskLevel string
		Count     int64
	}
	DB.Model(&model.TokenAnalysis{}).
		Select("risk_level, COUNT(*) as count").
		Where("analyzed_at BETWEEN ? AND ?", startOfDay, endOfDay).
		Group("risk_level").
		Scan(&riskStats)
	stats["risk_distribution"] = riskStats

	// 蜜罐数量
	var honeypotCount int64
	DB.Model(&model.TokenAnalysis{}).
		Where("analyzed_at BETWEEN ? AND ? AND is_honeypot = ?", startOfDay, endOfDay, true).
		Count(&honeypotCount)
	stats["honeypot_count"] = honeypotCount

	return stats, nil
}
