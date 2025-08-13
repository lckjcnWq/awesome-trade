# 🗄️ Web3 数据库集成指南

## 📋 概述

本目录包含 Web3 应用与数据库交互的完整指南，专注于使用 Go 语言构建高性能、可扩展的区块链应用数据层架构。涵盖数据库设计、链上数据同步、缓存策略、事务处理等核心技术。

## 📚 文档列表

### 核心集成指南

| 文档 | 数据库 | 类型 | 主要功能 | 难度 |
|------|--------|------|----------|------|
| [web3-mysql-usage-guide.md](./web3-mysql-usage-guide.md) | MySQL | 关系型数据库 | Web3数据存储与查询 | ⭐⭐⭐⭐ |
| [blockchain-data-sync-guide.md](./blockchain-data-sync-guide.md) | Multi-DB | 数据同步 | 链上数据实时同步 | ⭐⭐⭐⭐⭐ |
| [web3-caching-strategies-guide.md](./web3-caching-strategies-guide.md) | Redis/Memory | 缓存层 | 高性能数据缓存 | ⭐⭐⭐ |

### 专项技术指南

| 文档 | 技术栈 | 应用场景 | 复杂度 |
|------|--------|----------|--------|
| [defi-data-modeling-guide.md](./defi-data-modeling-guide.md) | MySQL + Go | DeFi协议数据建模 | ⭐⭐⭐⭐ |
| [nft-metadata-storage-guide.md](./nft-metadata-storage-guide.md) | IPFS + MySQL | NFT元数据管理 | ⭐⭐⭐ |
| [transaction-indexing-guide.md](./transaction-indexing-guide.md) | Elasticsearch + MySQL | 交易数据索引 | ⭐⭐⭐⭐⭐ |

## 🎯 核心特性

### 🏗️ 架构设计原则

- **分层架构**: 数据访问层、业务逻辑层、缓存层分离
- **读写分离**: 主从数据库架构，优化读写性能  
- **数据一致性**: 强一致性与最终一致性的平衡
- **扩展性**: 支持水平和垂直扩展
- **容错性**: 故障恢复和数据备份机制

### 📊 数据类型覆盖

- **链上数据**: 区块、交易、事件日志、状态变更
- **用户数据**: 账户信息、钱包地址、权限管理
- **业务数据**: DeFi操作、NFT交易、代币转账
- **元数据**: 合约ABI、代币信息、价格数据
- **审计数据**: 操作日志、合规记录、监控指标

### ⚡ 性能优化

- **索引策略**: 基于查询模式的智能索引设计
- **分区表**: 基于时间和哈希的数据分区
- **连接池**: 数据库连接池管理和优化
- **查询优化**: SQL查询性能调优
- **缓存策略**: 多级缓存架构设计

## 🚀 快速开始

### 1. 选择您的应用场景

**DeFi 协议开发**：
```bash
# 推荐学习路径
docs/database/web3-mysql-usage-guide.md          # 基础数据库设计
docs/database/defi-data-modeling-guide.md        # DeFi数据建模
docs/database/blockchain-data-sync-guide.md      # 实时数据同步
docs/database/web3-caching-strategies-guide.md   # 缓存优化
```

**NFT 市场开发**：
```bash
# 推荐学习路径  
docs/database/web3-mysql-usage-guide.md          # 基础数据库设计
docs/database/nft-metadata-storage-guide.md      # NFT元数据管理
docs/database/transaction-indexing-guide.md      # 交易索引系统
```

**区块链浏览器**：
```bash
# 推荐学习路径
docs/database/blockchain-data-sync-guide.md      # 全节点数据同步
docs/database/transaction-indexing-guide.md      # 海量交易索引
docs/database/web3-caching-strategies-guide.md   # 查询性能优化
```

### 2. 环境准备

所有文档都基于以下基础环境：

```bash
# Go 环境要求
Go 1.21+

# 数据库依赖
go get gorm.io/gorm
go get gorm.io/driver/mysql
go get github.com/ethereum/go-ethereum

# 缓存和队列
go get github.com/go-redis/redis/v8
go get github.com/Shopify/sarama

# 监控和日志
go get go.uber.org/zap
go get github.com/prometheus/client_golang
```

### 3. 项目结构模板

```
awesome-trade/
├── internal/
│   ├── database/              # 数据库相关
│   │   ├── models/           # 数据模型定义
│   │   ├── migrations/       # 数据库迁移
│   │   ├── repositories/     # 数据访问接口
│   │   └── seeders/          # 测试数据
│   ├── blockchain/           # 区块链交互
│   │   ├── clients/          # 区块链客户端
│   │   ├── listeners/        # 事件监听器
│   │   └── synchronizers/    # 数据同步器
│   └── cache/                # 缓存管理
│       ├── strategies/       # 缓存策略
│       └── invalidators/     # 缓存失效器
├── pkg/
│   ├── database/             # 数据库工具包
│   │   ├── connection/       # 连接管理
│   │   ├── transaction/      # 事务管理
│   │   └── monitoring/       # 性能监控
│   └── blockchain/           # 区块链工具包
│       ├── parser/           # 数据解析器
│       └── validator/        # 数据验证器
└── deployments/
    ├── mysql/               # MySQL配置
    ├── redis/              # Redis配置
    └── monitoring/         # 监控配置
```

## 🔧 数据库架构设计

### 核心数据表分类

```sql
-- 1. 区块链基础数据表
CREATE TABLE blocks (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    number BIGINT UNIQUE NOT NULL,
    hash VARCHAR(66) UNIQUE NOT NULL,
    parent_hash VARCHAR(66) NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    miner VARCHAR(42) NOT NULL,
    gas_used BIGINT NOT NULL,
    gas_limit BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_number (number),
    INDEX idx_timestamp (timestamp),
    INDEX idx_miner (miner)
);

CREATE TABLE transactions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    hash VARCHAR(66) UNIQUE NOT NULL,
    block_id BIGINT NOT NULL,
    block_number BIGINT NOT NULL,
    transaction_index INT NOT NULL,
    from_address VARCHAR(42) NOT NULL,
    to_address VARCHAR(42),
    value DECIMAL(78,0) NOT NULL,
    gas_price BIGINT NOT NULL,
    gas_used BIGINT,
    status TINYINT NOT NULL,
    input TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (block_id) REFERENCES blocks(id),
    INDEX idx_block_number (block_number),
    INDEX idx_from_address (from_address),
    INDEX idx_to_address (to_address),
    INDEX idx_hash (hash)
);

-- 2. DeFi 业务数据表
CREATE TABLE defi_protocols (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    contract_address VARCHAR(42) UNIQUE NOT NULL,
    protocol_type ENUM('DEX', 'LENDING', 'YIELD', 'INSURANCE') NOT NULL,
    chain_id INT NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE token_transfers (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    transaction_id BIGINT NOT NULL,
    token_address VARCHAR(42) NOT NULL,
    from_address VARCHAR(42) NOT NULL,
    to_address VARCHAR(42) NOT NULL,
    amount DECIMAL(78,0) NOT NULL,
    log_index INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (transaction_id) REFERENCES transactions(id),
    INDEX idx_token_address (token_address),
    INDEX idx_from_address (from_address),
    INDEX idx_to_address (to_address)
);

-- 3. 用户和钱包数据表
CREATE TABLE wallets (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    address VARCHAR(42) UNIQUE NOT NULL,
    first_seen_at TIMESTAMP NOT NULL,
    last_activity_at TIMESTAMP NOT NULL,
    transaction_count BIGINT DEFAULT 0,
    balance_eth DECIMAL(78,18) DEFAULT 0,
    is_contract BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_address (address),
    INDEX idx_last_activity (last_activity_at)
);
```

### 分库分表策略

```yaml
分库策略:
  主库 (awesome_trade_main):
    - 用户账户数据
    - 系统配置数据
    - 业务核心数据
    
  区块链数据库 (awesome_trade_blockchain):
    - 区块和交易数据
    - 事件日志数据
    - 智能合约数据
    
  业务数据库 (awesome_trade_business):
    - DeFi操作记录
    - NFT交易数据
    - 代币转账记录

分表策略:
  按时间分区 (transactions):
    - 按月分区: transactions_202501, transactions_202502
    - 自动分区管理: 保留最近12个月数据
    
  按哈希分片 (token_transfers):
    - 16个分片: token_transfers_00 到 token_transfers_15
    - 基于 token_address 哈希分片
```

## 💡 最佳实践

### 1. 数据一致性保证

```go
// 事务管理器
type TransactionManager struct {
    db *gorm.DB
    tx map[string]*gorm.DB
    mu sync.RWMutex
}

// 分布式事务处理
func (tm *TransactionManager) ExecuteDistributedTransaction(
    ctx context.Context, 
    operations []Operation,
) error {
    // 1. 预提交阶段
    for _, op := range operations {
        if err := op.Prepare(ctx); err != nil {
            tm.rollbackAll(operations)
            return err
        }
    }
    
    // 2. 提交阶段
    for _, op := range operations {
        if err := op.Commit(ctx); err != nil {
            tm.rollbackAll(operations)
            return err
        }
    }
    
    return nil
}
```

### 2. 读写分离配置

```go
// 数据库配置
type DatabaseConfig struct {
    Master DatabaseInstance `yaml:"master"`
    Slaves []DatabaseInstance `yaml:"slaves"`
    ReadWriteRatio float64 `yaml:"read_write_ratio"`
}

// 智能路由器
type DatabaseRouter struct {
    master *gorm.DB
    slaves []*gorm.DB
    balancer LoadBalancer
}

func (dr *DatabaseRouter) GetReadDB() *gorm.DB {
    return dr.balancer.SelectSlave(dr.slaves)
}

func (dr *DatabaseRouter) GetWriteDB() *gorm.DB {
    return dr.master
}
```

### 3. 性能监控

```go
// 数据库性能监控
type DatabaseMonitor struct {
    queryDuration prometheus.HistogramVec
    connectionPool prometheus.GaugeVec
    slowQueries   prometheus.CounterVec
}

func (dm *DatabaseMonitor) RecordQuery(
    operation string, 
    table string, 
    duration time.Duration,
) {
    dm.queryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
    
    // 记录慢查询
    if duration > time.Second {
        dm.slowQueries.WithLabelValues(operation, table).Inc()
    }
}
```

## 📈 扩展方案

### 水平扩展

- **读取扩展**: 多个只读副本 + 负载均衡
- **写入扩展**: 分片写入 + 数据路由
- **存储扩展**: 分区表 + 历史数据归档

### 垂直扩展  

- **计算资源**: CPU、内存动态扩展
- **存储资源**: SSD存储 + 热冷数据分离
- **网络资源**: 专用网络 + 连接池优化

### 缓存扩展

- **本地缓存**: 应用内存缓存
- **分布式缓存**: Redis Cluster
- **CDN缓存**: 静态数据全球分发

## 🛡️ 安全和合规

### 数据安全

- **访问控制**: 基于角色的权限管理
- **数据加密**: 传输加密 + 存储加密
- **审计日志**: 完整的操作审计链

### 合规要求

- **数据保留**: 符合监管要求的数据保留策略
- **隐私保护**: GDPR/CCPA 兼容的数据处理
- **监管报告**: 自动生成合规报告

## 🤝 贡献指南

### 添加新的数据库集成

1. 创建数据库特定的使用指南
2. 提供完整的迁移脚本
3. 添加性能基准测试
4. 编写单元和集成测试
5. 更新本 README 文档

### 文档改进

1. 补充实际项目案例
2. 更新技术栈版本信息  
3. 添加故障排除指南
4. 完善性能优化建议

---

**最后更新**: 2025-01-13  
**维护团队**: Awesome Trade 开发团队
