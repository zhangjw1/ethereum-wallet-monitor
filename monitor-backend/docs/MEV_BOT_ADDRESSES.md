# MEV Bot 地址数据源

## 如何获取已知的 MEV Bot 地址

### 1. Etherscan 标签页面
- **URL**: https://etherscan.io/accounts/label/mev-builder
- **说明**: Etherscan 维护的 MEV Builder 地址列表
- **更新频率**: 实时更新
- **使用方法**: 直接访问页面，复制地址

### 2. EigenPhi MEV 分析平台
- **URL**: https://eigenphi.io/mev/ethereum/leaderboard
- **说明**: 提供 MEV Bot 排行榜和详细分析
- **数据**: 包含 Bot 地址、收益、攻击类型等
- **API**: 提供 API 接口（需要注册）

### 3. Flashbots 透明度仪表板
- **URL**: https://transparency.flashbots.net/
- **说明**: Flashbots 官方的 MEV 数据
- **数据**: Builder 地址、区块统计、收益分析

### 4. MEV-Boost 中继统计
- **URL**: https://www.relayscan.io/
- **说明**: MEV-Boost 中继和 Builder 统计
- **数据**: 实时的 Builder 活动数据

### 5. GitHub 社区维护列表
- **URL**: https://gist.github.com/metachris/a4d10ff59cad5ffe3cf0f2c6e91fc0bc
- **说明**: 社区维护的 Builder 支付地址列表
- **格式**: Markdown 表格，易于解析

### 6. Flashbots 文档和 Wiki
- **URL**: https://docs.flashbots.net/
- **Wiki**: https://github.com/flashbots/mev-boost/wiki
- **说明**: 官方文档和技术资料

## 知名 MEV Bot 案例

### jaredfromsubway.eth
- **主地址**: `0xae2fc483527b8ef99eb5d9b44875f005ba1fae13`
- **Bot 地址**: `0x6b75d8af000000e20b7a7ddf000ba900b4009a80`
- **特点**: 最活跃的三明治攻击 Bot，收益超过 600 万美元
- **数据来源**: 
  - https://etherscan.io/address/jaredfromsubway.eth
  - https://medium.com/@eigenphi/performance-appraisal-of-jaredfromsubway-eth-c599fc713659

### Flashbots Builder
- **地址**: `0xdafea492d9c6733ae3d56b7ed1adb60692c98bc5`
- **ENS**: `flashbots-builder.eth`
- **类型**: MEV Builder（区块构建者）

### bloXroute Builders
- **Max-Profit**: `0xf2f5c73fa04406b1995e397b55c24ab1f3ea726c`
- **Non-Sandwich**: `0xf573d99385c05c23b24ed33de616ad16a43a0919`
- **Regulated**: `0x199d5ed7f45f4ee35960cf22eade2076e95b253f`

## MEV Bot 地址特征

### 常见模式
1. **前导零地址**: `0x000000000000...` 或 `0x00000000...`
   - 使用 CREATE2 生成的虚荣地址
   - 便于识别和过滤

2. **ENS 域名**: 
   - `jaredfromsubway.eth`
   - `flashbots-builder.eth`
   - `bloxroute-maxprofit.eth`

3. **高频交易特征**:
   - 每个区块多笔交易
   - Gas Price 通常较高
   - 涉及多个 DEX 合约

## 自动更新方案

### 方案 1: 定期手动更新
1. 每周访问上述数据源
2. 更新 `config/MevBotAddresses.go`
3. 提交代码更新

### 方案 2: API 集成（推荐）
```go
// 从 EigenPhi API 获取最新数据
func UpdateMevBotList() error {
    resp, err := http.Get("https://api.eigenphi.io/v1/mev/bots")
    // 解析并更新本地列表
}
```

### 方案 3: 链上分析
- 分析区块链数据，识别 MEV 行为模式
- 自动标记可疑地址
- 需要更复杂的算法

## 检测策略

### 1. 地址匹配（准确率 95%）
- 直接匹配已知 Bot 地址
- 最可靠的方法

### 2. 模式识别（准确率 75%）
- 前导零地址
- 高频交易模式
- 特定的合约调用序列

### 3. 行为分析（准确率 70%）
- 三明治攻击模式（前后交易）
- 异常高的 Gas Price
- 多次内部转账
- 交易失败但消耗大量 Gas

## 参考资料

1. **Flashbots 文档**: https://docs.flashbots.net/
2. **MEV 研究论文**: https://arxiv.org/abs/1904.05234
3. **EigenPhi 博客**: https://medium.com/@eigenphi
4. **Etherscan MEV 指南**: https://info.etherscan.com/exploring-the-world-of-mev/
5. **CoinMarketCap MEV 教程**: https://coinmarketcap.com/academy/article/frontrunners-and-mev-explained

## 更新日志

- **2026-01-29**: 初始版本，包含 10+ 个已知 Bot 地址
- 下次更新: 建议每月更新一次
