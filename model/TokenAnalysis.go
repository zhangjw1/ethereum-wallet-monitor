package model

import "time"

// TokenAnalysis 代币分析结果
type TokenAnalysis struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	TokenAddress string `gorm:"type:varchar(42);uniqueIndex;not null" json:"token_address"`

	// 基本信息
	Name        string `gorm:"type:varchar(255)" json:"name"`
	Symbol      string `gorm:"type:varchar(50);index" json:"symbol"`
	Decimals    uint8  `json:"decimals"`
	TotalSupply string `gorm:"type:varchar(100)" json:"total_supply"`

	// 流动性信息
	HasLiquidity     bool    `gorm:"default:false" json:"has_liquidity"`
	LiquidityUSD     float64 `json:"liquidity_usd"`
	InitialMarketCap float64 `json:"initial_market_cap"`
	PairAddress      string  `gorm:"type:varchar(42)" json:"pair_address"` // Uniswap Pair 地址

	// 安全检查
	IsVerified     bool   `gorm:"default:false" json:"is_verified"`
	IsHoneypot     bool   `gorm:"default:false" json:"is_honeypot"`
	HoneypotReason string `gorm:"type:text" json:"honeypot_reason"`

	// 税率
	BuyTax  float64 `json:"buy_tax"`
	SellTax float64 `json:"sell_tax"`

	// 持有者分析
	HolderCount     int     `json:"holder_count"`
	Top10HoldingPct float64 `json:"top10_holding_pct"` // 前10持有者占比

	// 所有权
	OwnerAddress         string `gorm:"type:varchar(42)" json:"owner_address"`
	IsOwnershipRenounced bool   `gorm:"default:false" json:"is_ownership_renounced"`

	// 风险评分
	RiskScore float64 `gorm:"index" json:"risk_score"`                  // 0-100，越低越安全
	RiskLevel string  `gorm:"type:varchar(20);index" json:"risk_level"` // "low", "medium", "high", "critical"
	RiskFlags string  `gorm:"type:text" json:"risk_flags"`              // JSON 数组，危险信号列表

	// 社交信息（可选）
	Website  string `gorm:"type:varchar(500)" json:"website"`
	Twitter  string `gorm:"type:varchar(500)" json:"twitter"`
	Telegram string `gorm:"type:varchar(500)" json:"telegram"`

	AnalyzedAt time.Time `gorm:"index" json:"analyzed_at"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (TokenAnalysis) TableName() string {
	return "token_analyses"
}
