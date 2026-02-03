# Ethereum MEV Monitor

以太坊 MEV 监控工具 - 实时监控以太坊交易中的 MEV（最大可提取价值）攻击行为

## 功能特性

- 🔍 MEV 攻击检测（三明治攻击、抢跑、套利等）
- 🤖 已知 MEV Bot 地址识别
- 📊 交易详情分析（Gas 价格、事件日志等）
- 💰 多链钱包余额查询（Covalent API）
- 🎯 **Meme 币自动发现和风险分析**（新功能）
- 🛡️ **蜜罐检测和安全评分**（新功能）
- ⏰ 定时任务调度（robfig/cron）
- 💾 SQLite 数据库存储
- 🔔 微信通知（PushPlus）
- 🌐 支持代理配置

## 技术栈

- Go 1.24+
- go-ethereum - 官方 Go 以太坊实现
- ethereum-watcher - 交易监控
- Covalent (GoldRush) API - 多链余额查询
- robfig/cron - 定时任务
- GORM - ORM 框架
- Zap - 日志库

## 快速开始

### 安装依赖

```bash
go mod download
```

### 配置

复制 `.env.example` 并重命名为 `.env`：

```bash
cp .env.example .env
```

编辑 `.env` 文件，填入你的 API Keys：

```env
# 以太坊 RPC 配置
INFURA_KEY=your_infura_key_here
LOG_LEVEL=info

# 数据库配置
DB_PATH=./ethereum_monitor.db

# Covalent API 配置（可选，用于多链余额查询）
COVALENT_API_KEY=your_covalent_api_key_here

# PushPlus 微信通知（可选）
PUSHPLUS_TOKEN=your_pushplus_token_here

# Etherscan API（用于合约验证和持有者查询）
ETHERSCAN_API_KEY=your_etherscan_api_key_here

# GoPlus Security API（用于蜜罐检测，可选）
GOPLUS_API_KEY=your_goplus_api_key_here
```

**获取 API Keys：**
- Infura: https://infura.io/
- Covalent: https://goldrush.dev/platform
- PushPlus: https://www.pushplus.plus/
- Etherscan: https://etherscan.io/apis
- GoPlus: https://gopluslabs.io/

### 运行

```bash
go run main.go
```

## 项目结构

```
.
├── main.go                     # 程序入口
├── config/
│   ├── MonitorConfig.go       # 监控配置
│   └── MevBotAddresses.go     # MEV Bot 地址库
├── database/
│   ├── sqlite.go              # 数据库连接
│   └── mev_builder_repository.go  # MEV Builder 数据访问
├── logger/
│   └── logger.go              # 日志配置
├── model/
│   └── MevBuilder.go          # 数据模型
├── scheduler/                  # 定时任务模块
│   ├── scheduler.go           # 定时任务调度器
│   ├── example.go             # 使用示例
│   └── README.md              # 定时任务文档
├── utils/
│   ├── HttpClientUtils.go     # HTTP 客户端
│   ├── MevDetector.go         # MEV 检测器
│   ├── CovalentClient.go      # Covalent API 客户端
│   ├── WechatNotifier.go      # 微信通知
│   └── *_example.go           # 使用示例
├── wallet/
│   ├── EthereumWalletMonitor.go    # 钱包监控
│   └── GoEthereumWalletMonitor.go  # Go-Ethereum 监控
└── docs/
    ├── COVALENT_USAGE.md      # Covalent 使用指南
    ├── MEV_BOT_ADDRESSES.md   # MEV Bot 地址说明
    └── USAGE.md               # 使用说明
```

## MEV 检测原理

程序通过以下特征识别 MEV 攻击：

1. **三明治攻击**：检测前后夹击的交易模式
2. **已知 Bot 地址**：匹配 MEV Bot 地址库
3. **Gas 价格异常**：检测异常高的 Gas 价格
4. **事件模式**：分析交易日志中的 Transfer 事件

## 已知 MEV Bot 地址

项目内置了主流 MEV Builder 和 Bot 地址：
- Flashbots Builder
- bloXroute (Max-Profit, Non-Sandwich, Regulated)
- Eden Network
- beaverbuild.org
- rsync-builder.xyz
- Titan Builder
- jaredfromsubway.eth (著名三明治攻击 Bot)

## 使用示例

### Meme 币监控（新功能）

自动发现和分析新部署的代币，识别潜力币和蜜罐：

```go
// 启动 Meme 币监控
monitor.ExampleMemeMonitor()

// 或测试分析指定代币
monitor.TestAnalyzeToken("0x代币合约地址")
```

**功能特点**：
- ✅ 自动检测新币部署
- ✅ 蜜罐检测（Honeypot.is + GoPlus）
- ✅ 风险评分（0-100分）
- ✅ 税率检测
- ✅ 持有者分析
- ✅ 流动性检查
- ✅ 低风险新币自动告警

详细文档：[docs/MEME_MONITOR_USAGE.md](docs/MEME_MONITOR_USAGE.md)

### 查询多链钱包余额

```go
// 创建 Covalent 客户端
client := utils.NewCovalentClient(os.Getenv("COVALENT_API_KEY"))

// 查询以太坊主网余额
balances, err := client.GetTokenBalances("eth-mainnet", "0x...")
if err != nil {
    log.Fatal(err)
}

// 遍历所有代币
for _, token := range balances.Data.Items {
    fmt.Printf("%s: %s (价值: $%.2f)\n",
        token.ContractTickerSymbol,
        token.Balance,
        token.Quote)
}
```

详细文档：[docs/COVALENT_USAGE.md](docs/COVALENT_USAGE.md)

### 定时任务

项目已集成定时任务功能，默认每天 0:00 执行任务。

详细文档：[scheduler/README.md](scheduler/README.md)

## 注意事项

- 需要有效的 Infura API Key 或其他以太坊 RPC 节点
- Covalent API Key 用于多链余额查询（可选）
- 代理可能影响 RPC 请求，建议关闭或配置白名单
- Go 版本需要 1.24.0+

## 相关文档

- [Meme 币监控使用指南](docs/MEME_MONITOR_USAGE.md)
- [Covalent API 使用指南](docs/COVALENT_USAGE.md)
- [定时任务使用说明](scheduler/README.md)
- [MEV Bot 地址说明](docs/MEV_BOT_ADDRESSES.md)

## License

MIT
