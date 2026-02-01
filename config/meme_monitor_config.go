package config

// Meme 币监控配置

// Uniswap V2 配置
const (
	// Uniswap V2 Factory 地址
	UniswapV2FactoryAddress = "0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f"

	// PairCreated 事件签名
	// event PairCreated(address indexed token0, address indexed token1, address pair, uint)
	UniswapV2PairCreatedTopic = "0x0d3648bd0f6ba80134a33ba9275ac585d9d315f0ad8355cddefde31afa28d0e9"

	// WETH 地址（用于识别 ETH 交易对）
	WETHAddress = "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"
)

// Meme 币风险评分阈值
const (
	// 低风险阈值（<30 分）
	MemeRiskScoreThresholdLow = 30.0

	// 初始市值阈值（USD）
	MemeMarketCapThreshold = 100000.0

	// 最小流动性阈值（USD）
	MemeMinLiquidityUSD = 5000.0
)

// 蜜罐检测 API
const (
	// Honeypot.is API
	HoneypotAPIURL = "https://api.honeypot.is/v2/IsHoneypot"

	// GoPlus Security API
	GoPlusAPIURL = "https://api.gopluslabs.io/api/v1/token_security/1"
)

// 风险评分权重
const (
	RiskScoreUnverified          = 30.0 // 未验证合约
	RiskScoreHoneypot            = 50.0 // 蜜罐
	RiskScoreHighTax             = 20.0 // 高税率 (>10%)
	RiskScoreConcentratedHolding = 25.0 // 持有者集中 (>50%)
	RiskScoreNoLiquidity         = 40.0 // 无流动性
	RiskScoreNotRenounced        = 15.0 // 未放弃所有权
)

// 持有者集中度阈值
const (
	Top10HoldingThreshold = 50.0 // 前10持有者占比阈值
)

// 税率阈值
const (
	HighTaxThreshold = 10.0 // 高税率阈值（百分比）
)
