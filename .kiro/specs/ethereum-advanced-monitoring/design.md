# 以太坊高级监控功能设计文档

## 系统架构

### 整体架构图

```
┌─────────────────────────────────────────────────────────────┐
│                     Ethereum RPC Node                        │
│                    (Infura/Alchemy)                          │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                   Event Listeners Layer                      │
├─────────────────────────────────────────────────────────────┤
│  LiquidationListener │ ContractDeployListener │              │
│  MevTransactionListener │ WalletActivityListener            │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                   Analysis & Detection Layer                 │
├─────────────────────────────────────────────────────────────┤
│  LiquidationAnalyzer │ MemeTokenAnalyzer │                   │
│  MevProfitCalculator │ WalletBehaviorAnalyzer               │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                      Storage Layer                           │
├─────────────────────────────────────────────────────────────┤
│  SQLite Database (GORM)                                      │
│  - liquidations, mev_transactions, wallets, tokens           │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                   Notification Layer                         │
├─────────────────────────────────────────────────────────────┤
│  PushPlus (WeChat) │ Logger (File/Console)                  │
└─────────────────────────────────────────────────────────────┘
```

## 模块设计

### 1. 清算事件监控模块

#### 数据模型

```go
// model/Liquidation.go
type Liquidation struct {
    ID                  uint      `gorm:"primaryKey"`
    TxHash              string    `gorm:"uniqueIndex;not null"`
    BlockNumber         uint64    `gorm:"index"`
    Timestamp           time.Time `gorm:"index"`
    Protocol            string    `gorm:"index"` // "Aave", "Compound"
    
    // 清算详情
    User                string    `gorm:"index"` // 被清算用户
    Liquidator          string    `gorm:"index"` // 清算人
    CollateralAsset     string    // 抵押品代币地址
    CollateralSymbol    string    // 抵押品符号
    CollateralAmount    string    // 抵押品数量
    DebtAsset           string    // 债务代币地址
    DebtSymbol          string    // 债务符号
    DebtAmount          string    // 债务数量
    
    // 价值计算（USD）
    CollateralValueUSD  float64   // 抵押品价值
    DebtValueUSD        float64   // 债务价值
    ProfitUSD           float64   // 清算利润
    DiscountPercent     float64   // 折扣百分比
    
    // Gas 信息
    GasUsed             uint64
    GasPrice            string
    GasCostUSD          float64
    
    CreatedAt           time.Time
}
```


#### 核心组件

**LiquidationListener**
- 监听 Aave V3 LiquidationCall 事件
- 监听 Compound V3 AbsorbDebt 事件
- 解析事件参数，提取清算数据

**LiquidationAnalyzer**
- 查询代币价格（Chainlink/Uniswap）
- 计算清算利润和折扣
- 评估清算风险等级

**LiquidationRepository**
- 数据持久化
- 查询历史清算记录
- 生成统计报告

#### 关键算法

```go
// 计算清算利润
func CalculateLiquidationProfit(
    collateralAmount *big.Int,
    collateralPrice float64,
    debtAmount *big.Int,
    debtPrice float64,
    gasCost *big.Int,
) float64 {
    collateralValue := toFloat(collateralAmount) * collateralPrice
    debtValue := toFloat(debtAmount) * debtPrice
    gasCostUSD := toFloat(gasCost) * ethPrice
    
    profit := collateralValue - debtValue - gasCostUSD
    return profit
}

// 计算折扣百分比
func CalculateDiscount(collateralValue, debtValue float64) float64 {
    return (collateralValue - debtValue) / collateralValue * 100
}
```

### 2. MEV 深度监控升级模块

#### 数据模型

```go
// model/MevTransaction.go
type MevTransaction struct {
    ID              uint      `gorm:"primaryKey"`
    TxHash          string    `gorm:"uniqueIndex;not null"`
    BlockNumber     uint64    `gorm:"index"`
    Timestamp       time.Time `gorm:"index"`
    MevType         string    `gorm:"index"` // "sandwich", "arbitrage", "liquidation"
    
    // 参与方
    BotAddress      string    `gorm:"index"` // MEV Bot 地址
    VictimAddress   string    `gorm:"index"` // 受害者地址（如果有）
    
    // 交易详情
    TargetToken     string    // 目标代币
    TargetDex       string    // 目标 DEX
    
    // 利润分析
    Revenue         string    // 收入（Wei）
    Cost            string    // 成本（Wei）
    NetProfit       string    // 净利润（Wei）
    NetProfitUSD    float64   // 净利润（USD）
    
    // Gas 信息
    GasUsed         uint64
    GasPrice        string
    
    // 三明治攻击特有字段
    FrontRunTxHash  string    // 前置交易
    BackRunTxHash   string    // 后置交易
    VictimLoss      string    // 受害者损失
    
    CreatedAt       time.Time
}

// model/MevBot.go
type MevBot struct {
    ID              uint      `gorm:"primaryKey"`
    Address         string    `gorm:"uniqueIndex;not null"`
    Label           string    // Bot 标签/名称
    
    // 统计数据
    TotalTransactions int
    SuccessRate       float64
    TotalProfitUSD    float64
    AvgProfitUSD      float64
    
    // 偏好分析
    PreferredMevType  string   // 最常用的 MEV 类型
    PreferredDex      string   // 最常用的 DEX
    
    FirstSeenAt       time.Time
    LastSeenAt        time.Time
    UpdatedAt         time.Time
}
```


#### 核心组件

**MevTransactionAnalyzer**
- 扩展现有 MevDetector
- 计算 MEV 利润（分析交易前后余额变化）
- 识别三明治攻击的前后交易
- 提取受害者信息

**MevBotTracker**
- 自动识别和追踪 MEV Bot
- 更新 Bot 统计数据
- 分析 Bot 行为模式

**MevStatisticsGenerator**
- 生成每日/每周 MEV 报告
- 排行榜生成
- 趋势分析

#### 关键算法

```go
// 计算三明治攻击利润
func CalculateSandwichProfit(
    frontRunTx, victimTx, backRunTx *types.Transaction,
    frontReceipt, backReceipt *types.Receipt,
) (*big.Int, error) {
    // 1. 获取 Bot 在前置交易前的代币余额
    // 2. 获取 Bot 在后置交易后的代币余额
    // 3. 计算差额 = 利润
    // 4. 减去 Gas 成本
    
    balanceBefore := getTokenBalance(botAddress, targetToken, frontRunTx.BlockNumber-1)
    balanceAfter := getTokenBalance(botAddress, targetToken, backRunTx.BlockNumber)
    
    profit := new(big.Int).Sub(balanceAfter, balanceBefore)
    gasCost := calculateGasCost(frontReceipt, backReceipt)
    
    netProfit := new(big.Int).Sub(profit, gasCost)
    return netProfit, nil
}

// 识别三明治攻击的三笔交易
func IdentifySandwichTransactions(targetTxHash string) (front, victim, back string, err error) {
    // 1. 获取目标交易所在区块
    // 2. 找到目标交易的索引
    // 3. 检查前一笔和后一笔交易
    // 4. 验证是否来自同一地址且操作同一池子
}
```

### 3. 钱包行为分析模块

#### 数据模型

```go
// model/MonitoredWallet.go
type MonitoredWallet struct {
    ID              uint      `gorm:"primaryKey"`
    Address         string    `gorm:"uniqueIndex;not null"`
    Label           string    // 地址标签
    WalletType      string    `gorm:"index"` // "smart_money", "whale", "dormant", "normal"
    Source          string    // "auto_discovered", "manual", "etherscan"
    
    // 资产信息
    EthBalance      string
    TotalValueUSD   float64   `gorm:"index"`
    
    // 交易统计
    TotalTrades     int
    WinRate         float64   // 胜率
    TotalProfitUSD  float64
    AvgHoldDays     float64   // 平均持仓天数
    
    // 偏好分析
    PreferredTokens string    // JSON 数组
    PreferredDex    string
    
    // 活跃度
    LastActiveAt    time.Time `gorm:"index"`
    IsDormant       bool      // 是否休眠（>365天未活动）
    
    CreatedAt       time.Time
    UpdatedAt       time.Time
}

// model/WalletTransaction.go
type WalletTransaction struct {
    ID              uint      `gorm:"primaryKey"`
    WalletAddress   string    `gorm:"index"`
    TxHash          string    `gorm:"uniqueIndex"`
    BlockNumber     uint64
    Timestamp       time.Time `gorm:"index"`
    
    // 交易类型
    TxType          string    // "buy", "sell", "transfer", "swap"
    TokenAddress    string
    TokenSymbol     string
    Amount          string
    ValueUSD        float64
    
    // 盈亏分析（仅针对卖出）
    BuyPrice        float64
    SellPrice       float64
    ProfitUSD       float64
    ProfitPercent   float64
    HoldDays        int
    
    CreatedAt       time.Time
}
```


#### 核心组件

**WalletDiscoveryService**
- 自动发现大额交易的发送/接收地址
- 计算地址资产规模
- 自动分类（聪明钱/巨鲸/普通）

**WalletBehaviorAnalyzer**
- 分析交易历史
- 计算胜率和收益率
- 识别交易模式

**AddressLabelService**
- 集成 Etherscan API 获取地址标签
- 集成 DeBank API 获取资产信息
- 缓存标签数据

**DormantWalletDetector**
- 检测休眠地址唤醒
- 分析唤醒后的交易行为

#### 关键算法

```go
// 自动发现聪明钱
func DiscoverSmartMoney(address string) (bool, error) {
    // 1. 获取最近 100 笔交易
    // 2. 分析买入/卖出配对
    // 3. 计算胜率和平均收益
    // 4. 胜率 > 60% 且平均收益 > 20% -> 聪明钱
    
    trades := getRecentTrades(address, 100)
    winRate, avgProfit := analyzeTrades(trades)
    
    if winRate > 0.6 && avgProfit > 0.2 {
        return true, nil
    }
    return false, nil
}

// 检测休眠地址唤醒
func DetectDormantWakeup(address string, tx *types.Transaction) bool {
    lastTx := getLastTransaction(address)
    if lastTx == nil {
        return false
    }
    
    daysSinceLastTx := time.Since(lastTx.Timestamp).Hours() / 24
    return daysSinceLastTx > 365
}
```

### 4. 智能合约部署监控 - Meme 币分析模块

#### 数据模型

```go
// model/ContractDeployment.go
type ContractDeployment struct {
    ID              uint      `gorm:"primaryKey"`
    ContractAddress string    `gorm:"uniqueIndex;not null"`
    DeployerAddress string    `gorm:"index"`
    TxHash          string    `gorm:"uniqueIndex"`
    BlockNumber     uint64    `gorm:"index"`
    Timestamp       time.Time `gorm:"index"`
    
    // 合约信息
    IsToken         bool
    IsVerified      bool
    ContractType    string    // "ERC20", "ERC721", "Other"
    
    CreatedAt       time.Time
}

// model/TokenAnalysis.go
type TokenAnalysis struct {
    ID              uint      `gorm:"primaryKey"`
    TokenAddress    string    `gorm:"uniqueIndex;not null"`
    
    // 基本信息
    Name            string
    Symbol          string
    Decimals        uint8
    TotalSupply     string
    
    // 流动性信息
    HasLiquidity    bool
    LiquidityUSD    float64
    InitialMarketCap float64
    PairAddress     string    // Uniswap Pair 地址
    
    // 安全检查
    IsVerified      bool
    IsHoneypot      bool
    HoneypotReason  string
    
    // 税率
    BuyTax          float64
    SellTax         float64
    
    // 持有者分析
    HolderCount     int
    Top10HoldingPct float64   // 前10持有者占比
    
    // 所有权
    OwnerAddress    string
    IsOwnershipRenounced bool
    
    // 风险评分
    RiskScore       float64   `gorm:"index"` // 0-100，越低越安全
    RiskLevel       string    // "low", "medium", "high", "critical"
    RiskFlags       string    // JSON 数组，危险信号列表
    
    // 社交信息（可选）
    Website         string
    Twitter         string
    Telegram        string
    
    AnalyzedAt      time.Time
    CreatedAt       time.Time
}
```


#### 核心组件

**ContractDeploymentListener**
- 监听区块中 to 地址为 null 的交易
- 监听 Uniswap V2/V3 Factory 的 PairCreated 事件
- 过滤出 ERC20 代币合约

**MemeTokenAnalyzer**
- 读取代币基本信息（name, symbol, totalSupply）
- 检查合约验证状态
- 分析持有者分布
- 检测流动性和初始市值

**HoneypotDetector**
- 集成 Honeypot.is API
- 集成 GoPlus Security API
- 模拟买卖交易检测蜜罐

**TokenRiskScorer**
- 综合评估风险因素
- 计算风险评分（0-100）
- 生成风险报告

#### 关键算法

```go
// 风险评分算法
func CalculateRiskScore(analysis *TokenAnalysis) float64 {
    score := 0.0
    
    // 未验证合约 +30
    if !analysis.IsVerified {
        score += 30
    }
    
    // 蜜罐 +50
    if analysis.IsHoneypot {
        score += 50
    }
    
    // 高税率 (>10%) +20
    if analysis.BuyTax > 10 || analysis.SellTax > 10 {
        score += 20
    }
    
    // 持有者过度集中 (>50%) +25
    if analysis.Top10HoldingPct > 50 {
        score += 25
    }
    
    // 无流动性 +40
    if !analysis.HasLiquidity {
        score += 40
    }
    
    // 未放弃所有权 +15
    if !analysis.IsOwnershipRenounced {
        score += 15
    }
    
    // 限制最大值为 100
    if score > 100 {
        score = 100
    }
    
    return score
}

// 检测是否是 ERC20 代币
func IsERC20Token(contractAddress string) bool {
    // 尝试调用 ERC20 标准方法
    hasName := callMethod(contractAddress, "name()")
    hasSymbol := callMethod(contractAddress, "symbol()")
    hasDecimals := callMethod(contractAddress, "decimals()")
    hasTotalSupply := callMethod(contractAddress, "totalSupply()")
    
    return hasName && hasSymbol && hasDecimals && hasTotalSupply
}

// 获取持有者分布
func GetHolderDistribution(tokenAddress string) (int, float64, error) {
    // 使用 Etherscan API 或链上查询
    // 返回：总持有者数，前10持有者占比
    holders := getTopHolders(tokenAddress, 10)
    totalSupply := getTotalSupply(tokenAddress)
    
    top10Amount := big.NewInt(0)
    for _, holder := range holders {
        top10Amount.Add(top10Amount, holder.Balance)
    }
    
    top10Pct := float64(top10Amount.Int64()) / float64(totalSupply.Int64()) * 100
    
    return len(getAllHolders(tokenAddress)), top10Pct, nil
}
```

## 外部 API 集成

### 1. 价格数据 API

**Chainlink Price Feeds**
```go
// 获取 ETH/USD 价格
func GetEthPrice() (float64, error) {
    // Chainlink ETH/USD Price Feed
    // 0x5f4eC3Df9cbd43714FE2740f5E3616155c5b8419
}
```

**Uniswap V2/V3 TWAP**
```go
// 从 Uniswap 获取代币价格
func GetTokenPriceFromUniswap(tokenAddress string) (float64, error) {
    // 查询 WETH/Token Pair
    // 计算价格
}
```

### 2. 蜜罐检测 API

**Honeypot.is**
```go
func CheckHoneypotAPI(tokenAddress string) (bool, string, error) {
    url := fmt.Sprintf("https://api.honeypot.is/v2/IsHoneypot?address=%s", tokenAddress)
    // 返回：是否蜜罐，原因
}
```

**GoPlus Security**
```go
func CheckTokenSecurity(tokenAddress string) (*SecurityReport, error) {
    url := fmt.Sprintf("https://api.gopluslabs.io/api/v1/token_security/1?contract_addresses=%s", tokenAddress)
    // 返回：详细安全报告
}
```

### 3. 地址标签 API

**Etherscan**
```go
func GetAddressLabel(address string) (string, error) {
    // 需要 Etherscan API Key
    url := fmt.Sprintf("https://api.etherscan.io/api?module=account&action=addresslabel&address=%s&apikey=%s", 
        address, apiKey)
}
```

## 配置管理

### 环境变量扩展

```env
# 现有配置
INFURA_KEY=xxx
COVALENT_API_KEY=xxx
PUSHPLUS_TOKEN=xxx

# 新增配置
ETHERSCAN_API_KEY=xxx
ALCHEMY_API_KEY=xxx
GOPLUS_API_KEY=xxx

# 监控配置
LIQUIDATION_ALERT_THRESHOLD_USD=100000
MEME_RISK_SCORE_THRESHOLD=30
MEME_MARKET_CAP_THRESHOLD=100000
SMART_MONEY_MIN_WIN_RATE=0.6
WHALE_MIN_BALANCE_ETH=1000

# 功能开关
ENABLE_LIQUIDATION_MONITOR=true
ENABLE_MEV_DEEP_ANALYTICS=true
ENABLE_WALLET_BEHAVIOR_ANALYTICS=true
ENABLE_MEME_COIN_ANALYZER=true
```

## 性能优化

### 1. 批量处理
- 批量查询价格数据
- 批量插入数据库

### 2. 缓存策略
- 代币价格缓存（5分钟）
- 地址标签缓存（24小时）
- 合约验证状态缓存（永久）

### 3. 异步处理
- 事件监听和数据分析解耦
- 使用 channel 和 goroutine 并发处理

### 4. 数据库优化
- 添加必要索引
- 定期清理历史数据
- 考虑分表策略

## 错误处理

### RPC 节点故障
```go
type RpcManager struct {
    primaryRpc   string
    fallbackRpcs []string
    currentIndex int
}

func (m *RpcManager) GetClient() (*ethclient.Client, error) {
    // 尝试主节点
    // 失败则切换到备用节点
}
```

### API 限流处理
```go
type RateLimiter struct {
    requestsPerSecond int
    lastRequest       time.Time
}

func (r *RateLimiter) Wait() {
    // 限流等待
}
```

## 测试策略

### 单元测试
- 每个分析器独立测试
- Mock 外部 API 调用
- 测试边界条件

### 集成测试
- 使用测试网（Goerli/Sepolia）
- 测试完整监控流程
- 验证数据持久化

### 性能测试
- 压力测试：并发监控 100+ 地址
- 延迟测试：事件响应时间
- 资源测试：内存和 CPU 使用

## 部署方案

### 开发环境
- 使用 Infura 免费层
- SQLite 数据库
- 本地运行

### 生产环境
- 使用 Alchemy Growth 计划
- PostgreSQL 数据库
- Docker 容器化部署
- 配置监控和告警（Prometheus + Grafana）
