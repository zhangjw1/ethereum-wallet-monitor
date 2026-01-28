# Ethereum Wallet Monitor

以太坊钱包监控工具 - 实时监控以太坊钱包地址的交易活动，支持大额交易告警

## 功能特性

- 🔍 实时监控指定以太坊钱包地址的交易
- 💰 支持大额交易告警（默认阈值：10 ETH）
- 📊 显示交易详情（金额、发送方、接收方、区块号、Gas 价格）
- 🔄 自动轮询新区块（默认间隔：10 秒）
- 🌐 支持代理配置

## 技术栈

- Go 1.21
- ethereum-watcher - 以太坊区块链监控库
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

## 项目结构

```
.
├── main.go                          # 程序入口
├── config/
│   └── MonitorConfig.go            # 配置文件
├── utils/
│   └── HttpClientUtils.go          # HTTP 客户端工具
└── wallet/
    └── EthereumWalletMonitor.go    # 钱包监控核心逻辑
```

## 配置说明

- `ETHEREUM_RPC_URL`: Infura 或其他以太坊节点的 RPC 地址
- `OKX_WALLET_ADDRESS`: 要监控的钱包地址
- `SLEEP_SECONDS_FOR_NEW_BLOCK`: 轮询新区块的间隔（秒）
- `ETH_THRESHOLD`: 大额交易告警阈值（ETH）

## 注意事项

- 需要有效的 Infura API Key
- 如果在国内使用，可能需要配置代理
- 代理配置在 `utils/HttpClientUtils.go` 中

## License

MIT
