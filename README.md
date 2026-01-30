# Ethereum MEV Monitor

以太坊 MEV 监控工具 - 实时监控以太坊交易中的 MEV（最大可提取价值）攻击行为

## 功能特性

- 🔍 MEV 攻击检测（三明治攻击、抢跑、套利等）
- 🤖 已知 MEV Bot 地址识别
- 📊 交易详情分析（Gas 价格、事件日志等）
- 💾 SQLite 数据库存储
- 🌐 支持代理配置

## 技术栈

- Go 1.24+
- go-ethereum - 官方 Go 以太坊实现
- GORM - ORM 框架
- Zap - 日志库

## 快速开始

### 安装依赖

```bash
go mod download
```

### 配置

创建 `.env` 文件：

```env
# 以太坊 RPC 配置
INFURA_KEY=your_infura_key_here
LOG_LEVEL=info

# 数据库配置
DB_PATH=./ethereum_monitor.db
```

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
├── utils/
│   ├── HttpClientUtils.go     # HTTP 客户端
│   ├── MevDetector.go         # MEV 检测器
│   └── MevDetector_example.go # 使用示例
└── wallet/
    ├── EthereumWalletMonitor.go    # 钱包监控
    └── GoEthereumWalletMonitor.go  # Go-Ethereum 监控
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

## 注意事项

- 需要有效的 Infura API Key 或其他以太坊 RPC 节点
- 代理可能影响 RPC 请求，建议关闭或配置白名单
- Go 版本需要 1.24.0+

## License

MIT
