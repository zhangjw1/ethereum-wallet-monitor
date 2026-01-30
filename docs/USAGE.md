# 项目说明 - 引入go-ethereum库

## 概述

本项目已成功引入官方的go-ethereum库 (`github.com/ethereum/go-ethereum`)，提供了更全面的以太坊功能支持。

## 主要变更

1. 在go.mod中添加了 `github.com/ethereum/go-ethereum v1.13.13` 依赖
2. 新增 `GoEthereumWalletMonitor.go` 文件，包含使用官方库的监控实现
3. 更新了 `main.go` 以支持两种监控模式的选择

## 功能特性

### 新增功能（基于go-ethereum）

- 连接到以太坊节点
- 查询地址余额
- 获取交易详情
- 过滤特定地址的日志
- 实时监听新区块
- 检测目标地址的收发交易

### 已有功能（基于hydro-protocol）

- 基于HTTP轮询的交易监控
- 大额交易告警
- 配置化的阈值检测

## 使用方法

### 1. 运行程序

```bash
go run main.go
```

程序会提示选择监控模式：
- 1: 使用原有的Hydro Protocol监控
- 2: 使用新的官方go-ethereum监控

### 2. 代码示例

#### 使用官方go-ethereum库

```go
// 创建监控器
monitor, err := wallet.NewGoEthereumWalletMonitor("YOUR_RPC_URL")
if err != nil {
    log.Fatal(err)
}
defer monitor.Close()

// 获取余额
balance, err := monitor.GetBalance("0x...")
if err != nil {
    log.Printf("获取余额失败: %v", err)
} else {
    fmt.Printf("余额: %s ETH\n", balance.Text('f', 6))
}

// 监听新区块
err = monitor.SubscribeNewHead()
```

#### 使用原有库

```go
// 继续使用原有的监控方式
wallet.AddressAddMonitor()
```

## 依赖说明

- `github.com/ethereum/go-ethereum`: 官方Go语言以太坊实现
- `github.com/HydroProtocol/ethereum-watcher`: 原有的以太坊监控库

## 优势对比

| 特性 | Hydro Protocol | go-ethereum |
|------|----------------|-------------|
| 官方支持 | 第三方 | 官方维护 |
| 功能覆盖 | 有限 | 全面 |
| 实时性 | 轮询 | 事件订阅 |
| 社区活跃度 | 一般 | 高 |

## 代码结构

```
ethereum-monitor/
├── wallet/
│   ├── EthereumWalletMonitor.go      # 原有监控实现
│   └── GoEthereumWalletMonitor.go    # 新增官方库实现
├── config/
│   └── MonitorConfig.go
├── utils/
│   └── HttpClientUtils.go
├── main.go                          # 更新后的主程序
└── go.mod                           # 添加了go-ethereum依赖
```

## 注意事项

1. 两个监控系统可以独立运行
2. go-ethereum库提供了更好的实时性和更低的延迟
3. 保留原有实现以供对比和回退
4. 确保RPC端点可用且未超出请求限制