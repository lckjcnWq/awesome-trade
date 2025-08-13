# 🛠️ Web3 基础设施集成指南

## 📋 概述

本目录包含 Web3 开发必备的基础设施和工具集成指南，涵盖区块链客户端、钱包连接、域名服务、分布式存储、预言机网络等核心组件，为构建完整的 Web3 应用提供技术支持。

## 📚 文档列表

### 核心基础设施

| 文档 | 服务 | 类型 | 主要功能 | 重要性 |
|------|------|------|----------|--------|
| [web3js-go-usage-guide.md](./web3js-go-usage-guide.md) | Web3.js Go | 区块链客户端 | 链上交互基础 | ⭐⭐⭐⭐⭐ |
| [metamask-connector-usage-guide.md](./metamask-connector-usage-guide.md) | MetaMask | 钱包连接 | 用户身份认证 | ⭐⭐⭐⭐⭐ |

### 域名和身份服务

| 文档 | 服务 | 类型 | 主要功能 | 采用率 |
|------|------|------|----------|--------|
| [ens-usage-guide.md](./ens-usage-guide.md) | ENS | 域名服务 | 地址解析 | ⭐⭐⭐⭐ |

### 分布式存储

| 文档 | 服务 | 类型 | 主要功能 | 去中心化程度 |
|------|------|------|----------|-------------|
| [ipfs-usage-guide.md](./ipfs-usage-guide.md) | IPFS | 分布式存储 | 内容寻址存储 | ⭐⭐⭐⭐⭐ |

### 数据和预言机

| 文档 | 服务 | 类型 | 主要功能 | 可靠性 |
|------|------|------|----------|--------|
| [chainlink-usage-guide.md](./chainlink-usage-guide.md) | Chainlink | 预言机网络 | 外部数据接入 | ⭐⭐⭐⭐⭐ |

### NFT 和数字资产

| 文档 | 服务 | 类型 | 主要功能 | 市场份额 |
|------|------|------|----------|----------|
| [opensea-usage-guide.md](./opensea-usage-guide.md) | OpenSea | NFT市场 | 数字资产交易 | ⭐⭐⭐⭐⭐ |

## 🚀 快速开始

### 1. 基础设施选择

**DApp 前端开发**：
- 必需：MetaMask 连接器
- 推荐：ENS 域名解析
- 可选：IPFS 内容存储

**数据驱动应用**：
- 核心：Chainlink 预言机
- 补充：The Graph 索引
- 备选：自建数据源

**NFT 应用开发**：
- 市场：OpenSea API
- 存储：IPFS + Pinata
- 元数据：标准化格式

### 2. 环境准备

```bash
# Web3基础设施依赖
go get github.com/ethereum/go-ethereum
go get github.com/gorilla/websocket
go get github.com/gin-gonic/gin

# IPFS相关
go get github.com/ipfs/go-ipfs-api
go get github.com/ipfs/go-ipfs-files

# ENS相关
go get github.com/wealdtech/go-ens/v3
```

### 3. 通用集成模式

```go
// 1. Web3基础设施管理器
type Web3Infrastructure struct {
    ethClient    *ethclient.Client
    ipfsClient   *ipfs.Shell
    ensResolver  *ens.Resolver
    walletConn   *WalletConnector
}

// 2. 统一的服务接口
type Web3Service interface {
    Initialize() error
    IsHealthy() bool
    GetStatus() ServiceStatus
}

// 3. 服务发现和管理
type ServiceRegistry struct {
    services map[string]Web3Service
    health   map[string]bool
}
```

## 🔧 基础设施特性对比

### 存储解决方案

| 服务 | 类型 | 去中心化 | 持久性 | 访问速度 | 成本 |
|------|------|----------|--------|----------|------|
| IPFS | 内容寻址 | 高 | 需要固定 | 中等 | 低 |
| Arweave | 永久存储 | 高 | 永久 | 慢 | 一次性 |
| Filecoin | 激励存储 | 高 | 合约保证 | 慢 | 中等 |
| AWS S3 | 中心化 | 无 | 高 | 快 | 中等 |

### 预言机服务

| 服务 | 数据类型 | 去中心化 | 更新频率 | 准确性 | 成本 |
|------|----------|----------|----------|--------|------|
| Chainlink | 多样化 | 高 | 实时 | 极高 | 中等 |
| Band Protocol | 金融数据 | 高 | 实时 | 高 | 低 |
| Tellor | 开放数据 | 中等 | 按需 | 中等 | 低 |
| 自建预言机 | 定制 | 低 | 可控 | 可变 | 高 |

### 钱包连接方案

| 钱包 | 用户基数 | 支持链 | 开发友好度 | 安全性 |
|------|----------|--------|------------|--------|
| MetaMask | 极大 | 多链 | 极高 | 高 |
| WalletConnect | 大 | 多链 | 高 | 高 |
| Coinbase Wallet | 大 | 有限 | 中等 | 高 |
| Trust Wallet | 中等 | 多链 | 中等 | 中等 |

## 💡 最佳实践

### 1. 基础设施架构设计

```go
// Web3基础设施管理器
type InfrastructureManager struct {
    services    map[string]Web3Service
    healthCheck *HealthChecker
    failover    *FailoverManager
    metrics     *MetricsCollector
}

func NewInfrastructureManager() *InfrastructureManager {
    return &InfrastructureManager{
        services:    make(map[string]Web3Service),
        healthCheck: NewHealthChecker(),
        failover:    NewFailoverManager(),
        metrics:     NewMetricsCollector(),
    }
}

func (im *InfrastructureManager) RegisterService(name string, service Web3Service) {
    im.services[name] = service
    im.healthCheck.AddService(name, service)
}

func (im *InfrastructureManager) GetService(name string) (Web3Service, error) {
    service, exists := im.services[name]
    if !exists {
        return nil, fmt.Errorf("服务 %s 不存在", name)
    }
    
    if !im.healthCheck.IsHealthy(name) {
        // 尝试故障转移
        backup, err := im.failover.GetBackupService(name)
        if err != nil {
            return nil, fmt.Errorf("服务 %s 不可用且无备用服务", name)
        }
        return backup, nil
    }
    
    return service, nil
}
```

### 2. 钱包连接管理

```go
// 钱包连接管理器
type WalletManager struct {
    connectors map[string]WalletConnector
    current    string
    events     chan WalletEvent
}

type WalletConnector interface {
    Connect() error
    Disconnect() error
    GetAccounts() ([]string, error)
    SignMessage(message []byte) ([]byte, error)
    SendTransaction(tx *types.Transaction) (string, error)
    IsConnected() bool
}

func (wm *WalletManager) ConnectWallet(walletType string) error {
    connector, exists := wm.connectors[walletType]
    if !exists {
        return fmt.Errorf("不支持的钱包类型: %s", walletType)
    }
    
    err := connector.Connect()
    if err != nil {
        return err
    }
    
    wm.current = walletType
    wm.events <- WalletEvent{
        Type:   "connected",
        Wallet: walletType,
    }
    
    return nil
}

type WalletEvent struct {
    Type   string
    Wallet string
    Data   interface{}
}
```

### 3. 分布式存储管理

```go
// 存储管理器
type StorageManager struct {
    ipfs     *ipfs.Shell
    pinata   *PinataClient
    arweave  *ArweaveClient
    strategy StorageStrategy
}

type StorageStrategy int

const (
    IPFSOnly StorageStrategy = iota
    IPFSWithPinning
    MultiProvider
    PermanentStorage
)

func (sm *StorageManager) Store(data []byte, strategy StorageStrategy) (*StorageResult, error) {
    switch strategy {
    case IPFSOnly:
        return sm.storeIPFS(data)
    case IPFSWithPinning:
        return sm.storeWithPinning(data)
    case MultiProvider:
        return sm.storeMultiProvider(data)
    case PermanentStorage:
        return sm.storePermanent(data)
    default:
        return nil, fmt.Errorf("不支持的存储策略")
    }
}

func (sm *StorageManager) storeIPFS(data []byte) (*StorageResult, error) {
    hash, err := sm.ipfs.Add(bytes.NewReader(data))
    if err != nil {
        return nil, err
    }
    
    return &StorageResult{
        Hash:     hash,
        Provider: "ipfs",
        URL:      fmt.Sprintf("ipfs://%s", hash),
    }, nil
}

type StorageResult struct {
    Hash     string
    Provider string
    URL      string
    Metadata map[string]interface{}
}
```

### 4. 预言机数据管理

```go
// 预言机管理器
type OracleManager struct {
    chainlink *ChainlinkClient
    band      *BandClient
    custom    map[string]CustomOracle
    cache     *DataCache
}

func (om *OracleManager) GetPrice(asset string, sources []string) (*PriceData, error) {
    var prices []*PriceData
    
    for _, source := range sources {
        price, err := om.getPriceFromSource(asset, source)
        if err != nil {
            continue
        }
        prices = append(prices, price)
    }
    
    if len(prices) == 0 {
        return nil, fmt.Errorf("无法获取 %s 的价格数据", asset)
    }
    
    // 计算加权平均价格
    return om.calculateWeightedPrice(prices), nil
}

func (om *OracleManager) getPriceFromSource(asset, source string) (*PriceData, error) {
    // 先检查缓存
    if cached := om.cache.Get(asset, source); cached != nil {
        return cached, nil
    }
    
    var price *PriceData
    var err error
    
    switch source {
    case "chainlink":
        price, err = om.chainlink.GetPrice(asset)
    case "band":
        price, err = om.band.GetPrice(asset)
    default:
        if oracle, exists := om.custom[source]; exists {
            price, err = oracle.GetPrice(asset)
        } else {
            err = fmt.Errorf("未知的预言机源: %s", source)
        }
    }
    
    if err == nil {
        om.cache.Set(asset, source, price, 5*time.Minute)
    }
    
    return price, err
}

type PriceData struct {
    Asset     string
    Price     decimal.Decimal
    Timestamp time.Time
    Source    string
    Confidence decimal.Decimal
}
```

## 🔍 监控和健康检查

### 服务健康监控

```go
// 健康检查器
type HealthChecker struct {
    services map[string]Web3Service
    status   map[string]ServiceHealth
    alerts   chan HealthAlert
}

type ServiceHealth struct {
    IsHealthy    bool
    LastCheck    time.Time
    ResponseTime time.Duration
    ErrorCount   int
    Uptime       time.Duration
}

func (hc *HealthChecker) StartMonitoring() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        hc.checkAllServices()
    }
}

func (hc *HealthChecker) checkAllServices() {
    for name, service := range hc.services {
        start := time.Now()
        healthy := service.IsHealthy()
        responseTime := time.Since(start)
        
        health := hc.status[name]
        health.LastCheck = time.Now()
        health.ResponseTime = responseTime
        
        if healthy != health.IsHealthy {
            // 状态变化，发送警报
            hc.alerts <- HealthAlert{
                Service:   name,
                OldStatus: health.IsHealthy,
                NewStatus: healthy,
                Timestamp: time.Now(),
            }
        }
        
        health.IsHealthy = healthy
        if !healthy {
            health.ErrorCount++
        }
        
        hc.status[name] = health
    }
}

type HealthAlert struct {
    Service   string
    OldStatus bool
    NewStatus bool
    Timestamp time.Time
}
```

### 性能指标收集

```go
// 指标收集器
type MetricsCollector struct {
    metrics map[string]*ServiceMetrics
    mu      sync.RWMutex
}

type ServiceMetrics struct {
    RequestCount    int64
    ErrorCount      int64
    TotalLatency    time.Duration
    AverageLatency  time.Duration
    LastRequestTime time.Time
}

func (mc *MetricsCollector) RecordRequest(service string, latency time.Duration, success bool) {
    mc.mu.Lock()
    defer mc.mu.Unlock()
    
    metrics, exists := mc.metrics[service]
    if !exists {
        metrics = &ServiceMetrics{}
        mc.metrics[service] = metrics
    }
    
    metrics.RequestCount++
    metrics.TotalLatency += latency
    metrics.AverageLatency = time.Duration(int64(metrics.TotalLatency) / metrics.RequestCount)
    metrics.LastRequestTime = time.Now()
    
    if !success {
        metrics.ErrorCount++
    }
}

func (mc *MetricsCollector) GetMetrics(service string) *ServiceMetrics {
    mc.mu.RLock()
    defer mc.mu.RUnlock()
    
    if metrics, exists := mc.metrics[service]; exists {
        // 返回副本以避免并发问题
        return &ServiceMetrics{
            RequestCount:    metrics.RequestCount,
            ErrorCount:      metrics.ErrorCount,
            TotalLatency:    metrics.TotalLatency,
            AverageLatency:  metrics.AverageLatency,
            LastRequestTime: metrics.LastRequestTime,
        }
    }
    
    return nil
}
```

## 🔒 安全考虑

### 1. 钱包安全

- 永远不要在服务端存储私钥
- 使用安全的签名流程
- 实现会话管理和超时
- 验证所有用户输入

### 2. 数据完整性

- 验证 IPFS 内容哈希
- 使用多个预言机源交叉验证
- 实现数据签名验证
- 监控异常数据模式

### 3. 服务可用性

- 实现多重备份机制
- 使用负载均衡
- 设置合理的超时时间
- 准备降级方案

## 🤝 贡献指南

### 添加新基础设施

1. 创建服务特定的使用指南
2. 实现标准的 Web3Service 接口
3. 添加健康检查和监控
4. 编写集成测试用例
5. 更新本 README 文档

### 文档改进

1. 补充实际集成案例
2. 更新服务 API 变化
3. 添加故障排除指南
4. 完善安全最佳实践

---

**最后更新**: 2025-01-13  
**维护团队**: Awesome Trade 开发团队
