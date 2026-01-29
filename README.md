# Ethereum Wallet Monitor

以太坊钱包监控工具 - 实时监控以太坊钱包地址的交易活动，支持大额交易告警

## 功能特性

- 🔍 实时监控指定以太坊钱包地址的交易
- 💰 支持大额交易告警（默认阈值：10 ETH）
- 📊 显示交易详情（金额、发送方、接收方、区块号、Gas 价格）
- 🔄 自动轮询新区块（默认间隔：10 秒）
- 🌐 支持代理配置
- ⚡ 支持多种监控后端（Hydro Protocol, Go-Ethereum官方库）

## 技术栈

- Go 1.21
- ethereum-watcher - 以太坊区块链监控库
- go-ethereum - 官方Go以太坊实现
- Infura - 以太坊节点服务

## 快速开始

### 安装依赖

```bash
go mod download
```

### 配置

编辑 `config/MonitorConfig.go` 文件：

```go
const (
    ETHEREUM_RPC_URL   = "https://mainnet.infura.io/v3/YOUR_INFURA_KEY"
    OKX_WALLET_ADDRESS = "YOUR_WALLET_ADDRESS"
    SLEEP_SECONDS_FOR_NEW_BLOCK = 10
    ETH_THRESHOLD = 10
)
```

### 运行

```bash
go run main.go
```

程序会提示选择监控模式：
- 1: 使用 Hydro Protocol 监控（轮询模式）
- 2: 使用官方 Go-Ethereum 监控（事件订阅模式）

## 项目结构

```
.
├── main.go                          # 程序入口
├── config/
│   └── MonitorConfig.go            # 配置文件
├── utils/
│   └── HttpClientUtils.go          # HTTP 客户端工具
└── wallet/
    ├── EthereumWalletMonitor.go    # 原有监控实现（Hydro Protocol）
    └── GoEthereumWalletMonitor.go  # 新增监控实现（官方Go-Ethereum）
```

## 配置说明

- `INFURA_KEY`: Infura 项目的 API 密钥（注意：不是完整URL，而是密钥部分）
- `OKX_WALLET_ADDRESS`: 要监控的钱包地址
- `SLEEP_SECONDS_FOR_NEW_BLOCK`: 轮询新区块的间隔（秒）
- `ETH_THRESHOLD`: 大额交易告警阈值（ETH）

## 两种监控方式对比

| 特性 | Hydro Protocol | Go-Ethereum |
|------|----------------|-------------|
| 实现方式 | HTTP轮询 | WebSocket订阅 |
| 实时性 | 较低（取决于轮询间隔） | 更高（实时推送） |
| 资源消耗 | 轮询开销 | 订阅开销 |
| 官方支持 | 第三方 | 以太坊基金会 |
| 功能覆盖 | 基础功能 | 完整功能 |

## 注意事项

- 需要有效的 Infura API Key 或其他以太坊节点服务
- 如果在国内使用，可能需要配置代理
- 代理配置在 `utils/HttpClientUtils.go` 中
- 新增了官方go-ethereum库支持，提供更多功能选项

## License

MIT
