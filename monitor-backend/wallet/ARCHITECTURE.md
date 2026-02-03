# Wallet 包架构设计

## 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                        应用层                                 │
│  ┌──────────────────┐         ┌──────────────────┐          │
│  │  StartGoEthMonitor│         │StartWatcherMonitor│         │
│  │   (example.go)   │         │   (example.go)   │          │
│  └────────┬─────────┘         └────────┬─────────┘          │
└───────────┼──────────────────────────────┼──────────────────┘
            │                              │
┌───────────┼──────────────────────────────┼──────────────────┐
│           │        监控实现层             │                   │
│  ┌────────▼─────────┐         ┌─────────▼────────┐          │
│  │  GoEthMonitor    │         │ WatcherMonitor   │          │
│  │ (goeth_monitor)  │         │(watcher_monitor) │          │
│  │                  │         │                  │          │
│  │ • WebSocket订阅  │         │ • HTTP轮询       │          │
│  │ • HTTP轮询       │         │ • 插件架构       │          │
│  │ • 自动降级       │         │ • 易扩展         │          │
│  └────────┬─────────┘         └─────────┬────────┘          │
└───────────┼──────────────────────────────┼──────────────────┘
            │                              │
            └──────────────┬───────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│                      公共组件层 (common.go)                   │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │AddressManager│  │NotificationSvc│  │  MevFilter   │      │
│  │              │  │              │  │              │      │
│  │• 地址管理    │  │• 通知发送    │  │• MEV检测     │      │
│  │• 标签映射    │  │• 数据库记录  │  │• 交易过滤    │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│  ┌──────────────┐  ┌──────────────────────────────────┐    │
│  │TokenHandler  │  │      工具函数                     │    │
│  │              │  │ • WeiToEth                       │    │
│  │• 代币配置    │  │ • CreateETHThreshold             │    │
│  │• 金额解析    │  │ • ExtractAddressFromTopic        │    │
│  └──────────────┘  └──────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│                      基础设施层                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ go-ethereum  │  │ethereum-watcher│ │   Database   │      │
│  │              │  │              │  │              │      │
│  │• ethclient   │  │• Watcher     │  │• SQLite      │      │
│  │• types       │  │• Plugin      │  │• Repository  │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│  ┌──────────────┐  ┌──────────────┐                        │
│  │   Logger     │  │  Notifier    │                        │
│  │              │  │              │                        │
│  │• zap         │  │• PushPlus    │                        │
│  └──────────────┘  └──────────────┘                        │
└─────────────────────────────────────────────────────────────┘
```

## 数据流图

### GoEthMonitor 数据流

```
┌─────────────┐
│ Ethereum    │
│ Blockchain  │
└──────┬──────┘
       │
       │ WebSocket/HTTP
       ▼
┌─────────────────┐
│  ethclient      │
│  (go-ethereum)  │
└────────┬────────┘
         │
         │ Block/Transaction
         ▼
┌─────────────────────────────────────────┐
│         GoEthMonitor                    │
│  ┌─────────────────────────────────┐   │
│  │  checkBlockTransactions()       │   │
│  │    │                             │   │
│  │    ├─► handleETHTransaction()   │   │
│  │    │                             │   │
│  │    └─► checkERC20Transfers()    │   │
│  └─────────────────────────────────┘   │
└────────┬────────────────────────────────┘
         │
         │ TransferNotification
         ▼
┌─────────────────────────────────────────┐
│      公共组件处理流程                     │
│                                         │
│  1. AddressManager.IsMonitored()       │
│         │                               │
│         ▼                               │
│  2. MevFilter.IsMevTransaction()       │
│         │                               │
│         ▼                               │
│  3. NotificationService.Send()         │
│         │                               │
│         ├─► PushPlus 通知              │
│         └─► Database 记录              │
└─────────────────────────────────────────┘
```

### WatcherMonitor 数据流

```
┌─────────────┐
│ Ethereum    │
│ Blockchain  │
└──────┬──────┘
       │
       │ HTTP Polling
       ▼
┌─────────────────┐
│ ethereum-watcher│
│   Framework     │
└────────┬────────┘
         │
         │ Plugin Events
         ▼
┌─────────────────────────────────────────┐
│         WatcherMonitor                  │
│  ┌─────────────────────────────────┐   │
│  │  ethTransactionPlugin           │   │
│  │    • AcceptTx()                 │   │
│  └─────────────────────────────────┘   │
│  ┌─────────────────────────────────┐   │
│  │  erc20TransferPlugin            │   │
│  │    • Accept()                   │   │
│  └─────────────────────────────────┘   │
└────────┬────────────────────────────────┘
         │
         │ TransferNotification
         ▼
┌─────────────────────────────────────────┐
│      公共组件处理流程                     │
│  (同 GoEthMonitor)                      │
└─────────────────────────────────────────┘
```

## 类图

### 核心类关系

```
┌─────────────────────┐
│   MonitorConfig     │
│─────────────────────│
│ + Addresses         │
│ + Tokens            │
│ + ETHThreshold      │
│ + TokenThreshold    │
└─────────────────────┘
          △
          │ uses
          │
┌─────────┴───────────────────────────────┐
│                                         │
┌─────────────────────┐   ┌─────────────────────┐
│   GoEthMonitor      │   │  WatcherMonitor     │
│─────────────────────│   │─────────────────────│
│ - client            │   │ - watcher           │
│ - wsClient          │   │ - addressMgr        │
│ - addressMgr        │   │ - notifSvc          │
│ - notifSvc          │   │ - mevFilter         │
│ - mevFilter         │   │ - tokenHandler      │
│ - tokenHandler      │   │─────────────────────│
│─────────────────────│   │ + Start()           │
│ + Start()           │   │ + Close()           │
│ + Close()           │   └─────────────────────┘
│ - startWebSocket()  │
│ - startPolling()    │
└─────────────────────┘
          │
          │ uses
          ▼
┌─────────────────────────────────────────┐
│         公共组件                         │
│                                         │
│  ┌─────────────────┐                   │
│  │AddressManager   │                   │
│  │─────────────────│                   │
│  │+ IsMonitored()  │                   │
│  │+ GetLabel()     │                   │
│  └─────────────────┘                   │
│                                         │
│  ┌─────────────────┐                   │
│  │NotificationSvc  │                   │
│  │─────────────────│                   │
│  │+ Send()         │                   │
│  │+ IsProcessed()  │                   │
│  └─────────────────┘                   │
│                                         │
│  ┌─────────────────┐                   │
│  │  MevFilter      │                   │
│  │─────────────────│                   │
│  │+ IsMevTx()      │                   │
│  └─────────────────┘                   │
│                                         │
│  ┌─────────────────┐                   │
│  │ TokenHandler    │                   │
│  │─────────────────│                   │
│  │+ ParseAmount()  │                   │
│  │+ GetConfig()    │                   │
│  └─────────────────┘                   │
└─────────────────────────────────────────┘
```

## 时序图

### ETH 转账监控流程

```
用户钱包    Blockchain    Monitor      AddressMgr   MevFilter   NotifSvc    PushPlus   Database
   │            │            │              │            │           │           │          │
   │  发起转账   │            │              │            │           │           │          │
   ├───────────►│            │              │            │           │           │          │
   │            │            │              │            │           │           │          │
   │            │  新区块     │              │            │           │           │          │
   │            ├───────────►│              │            │           │           │          │
   │            │            │              │            │           │           │          │
   │            │            │ IsMonitored? │            │           │           │          │
   │            │            ├─────────────►│            │           │           │          │
   │            │            │     Yes      │            │           │           │          │
   │            │            │◄─────────────┤            │           │           │          │
   │            │            │              │            │           │           │          │
   │            │            │  IsMevTx?    │            │           │           │          │
   │            │            ├──────────────┼───────────►│           │           │          │
   │            │            │              │    No      │           │           │          │
   │            │            │◄─────────────┼────────────┤           │           │          │
   │            │            │              │            │           │           │          │
   │            │            │  SendNotification         │           │           │          │
   │            │            ├───────────────────────────┼──────────►│           │          │
   │            │            │              │            │           │           │          │
   │            │            │              │            │           │  Send     │          │
   │            │            │              │            │           ├──────────►│          │
   │            │            │              │            │           │   OK      │          │
   │            │            │              │            │           │◄──────────┤          │
   │            │            │              │            │           │           │          │
   │            │            │              │            │           │  Save     │          │
   │            │            │              │            │           ├──────────────────────►│
   │            │            │              │            │           │           │    OK    │
   │            │            │              │            │           │◄──────────────────────┤
   │            │            │              │            │           │           │          │
   │            │            │◄──────────────────────────┼───────────┤           │          │
   │            │            │              │            │           │           │          │
```

## 模块职责

### 1. 监控实现层

#### GoEthMonitor
**职责：**
- 管理 go-ethereum 客户端连接
- 实现 WebSocket 实时订阅
- 实现 HTTP 轮询备用方案
- 自动在两种模式间切换
- 解析区块和交易数据

**不负责：**
- 地址匹配逻辑（委托给 AddressManager）
- MEV 检测（委托给 MevFilter）
- 通知发送（委托给 NotificationService）

#### WatcherMonitor
**职责：**
- 管理 ethereum-watcher 框架
- 实现插件式架构
- 注册和管理插件
- 处理插件事件

**不负责：**
- 具体的业务逻辑（在插件中实现）
- 公共功能（委托给公共组件）

### 2. 公共组件层

#### AddressManager
**单一职责：** 地址管理
- 维护监控地址列表
- 提供地址查询接口
- 管理地址标签

#### NotificationService
**单一职责：** 通知管理
- 发送各种通知
- 记录通知历史
- 防止重复通知

#### MevFilter
**单一职责：** MEV 检测
- 识别 MEV Bot 交易
- 提供过滤决策

#### TokenHandler
**单一职责：** 代币处理
- 管理代币配置
- 解析代币金额
- 提供代币信息

## 设计模式

### 1. 策略模式 (Strategy Pattern)
两种监控实现（GoEthMonitor 和 WatcherMonitor）是不同的策略，可以根据需求选择。

```go
// 策略接口（隐式）
type Monitor interface {
    Start(ctx context.Context) error
    Close()
}

// 具体策略
type GoEthMonitor struct { ... }
type WatcherMonitor struct { ... }
```

### 2. 组合模式 (Composition Pattern)
监控器通过组合公共组件来实现功能，而不是继承。

```go
type GoEthMonitor struct {
    addressMgr   *AddressManager      // 组合
    notifSvc     *NotificationService // 组合
    mevFilter    *MevFilter           // 组合
    tokenHandler *TokenHandler        // 组合
}
```

### 3. 插件模式 (Plugin Pattern)
WatcherMonitor 使用插件模式，易于扩展。

```go
type ethTransactionPlugin struct {
    monitor *WatcherMonitor
}

func (p *ethTransactionPlugin) AcceptTx(tx structs.RemovableTx) {
    // 处理交易
}
```

### 4. 单一职责原则 (Single Responsibility)
每个组件只负责一个功能：
- AddressManager → 地址管理
- NotificationService → 通知
- MevFilter → MEV 检测
- TokenHandler → 代币处理

### 5. 依赖注入 (Dependency Injection)
通过构造函数注入依赖，便于测试。

```go
func NewGoEthMonitor(rpcURL, wsURL string, config *MonitorConfig) (*GoEthMonitor, error) {
    addressMgr := NewAddressManager(config.Addresses)
    notifSvc := NewNotificationService()
    // ...
}
```

## 扩展点

### 1. 添加新的监控模式
实现相同的接口，使用公共组件：

```go
type GraphQLMonitor struct {
    addressMgr   *AddressManager
    notifSvc     *NotificationService
    // ...
}

func (m *GraphQLMonitor) Start(ctx context.Context) error {
    // 使用 GraphQL 订阅
}
```

### 2. 添加新的通知渠道
扩展 NotificationService：

```go
type NotificationService struct {
    pushPlus   *utils.PushPlusNotifier
    telegram   *utils.TelegramNotifier  // 新增
    discord    *utils.DiscordNotifier   // 新增
}
```

### 3. 添加新的过滤器
类似 MevFilter 的设计：

```go
type SpamFilter struct {
    // 过滤垃圾交易
}

func (sf *SpamFilter) IsSpamTransaction(txHash string) bool {
    // 检测逻辑
}
```

### 4. 添加新的代币类型
扩展 TokenHandler：

```go
type NFTHandler struct {
    // 处理 ERC721/ERC1155
}
```

## 性能考虑

### 1. 并发处理
```go
// 使用 goroutine 并发处理区块
for _, tx := range block.Transactions() {
    go m.handleTransaction(ctx, tx, blockNum)
}
```

### 2. 连接池
```go
// 复用 HTTP 连接
client := &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
    },
}
```

### 3. 缓存
```go
// 缓存已处理的交易
type NotificationService struct {
    processedCache *lru.Cache
}
```

## 安全考虑

### 1. 输入验证
```go
func NewAddressManager(addresses map[string]string) *AddressManager {
    // 验证地址格式
    for addr := range addresses {
        if !common.IsHexAddress(addr) {
            panic("invalid address")
        }
    }
}
```

### 2. 错误处理
```go
// 优雅处理错误，不中断监控
if err := m.handleTransaction(ctx, tx, blockNum); err != nil {
    logger.Error("处理交易失败", zap.Error(err))
    // 继续处理下一个交易
}
```

### 3. 资源清理
```go
func (m *GoEthMonitor) Close() {
    if m.client != nil {
        m.client.Close()
    }
    if m.wsClient != nil {
        m.wsClient.Close()
    }
    // ...
}
```

## 总结

这个架构设计具有以下优点：

1. **清晰的分层**：应用层、监控层、组件层、基础设施层
2. **高内聚低耦合**：每个模块职责单一，依赖关系清晰
3. **易于扩展**：新功能可以通过组合现有组件实现
4. **易于测试**：依赖注入，可以 mock 各个组件
5. **高可维护性**：代码结构清晰，易于理解和修改
