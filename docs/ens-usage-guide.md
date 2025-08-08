# ENS Go 详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [环境准备](#环境准备)
3. [域名解析](#域名解析)
4. [反向解析](#反向解析)
5. [域名注册](#域名注册)
6. [记录管理](#记录管理)
7. [子域名操作](#子域名操作)
8. [高级功能](#高级功能)
9. [实际应用](#实际应用)

## 基础概念

### 1.1 ENS简介

ENS (Ethereum Name Service) 是以太坊域名服务，提供人类可读的域名到以太坊地址的映射，类似于DNS系统。

```bash
# 安装ENS相关依赖
go get github.com/ethereum/go-ethereum
go get github.com/wealdtech/go-ens/v3
go get github.com/shopspring/decimal
```

### 1.2 核心概念

```go
// 主要包导入
import (
    "context"
    "math/big"
    "strings"
    
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/ethereum/go-ethereum/common"
    "github.com/wealdtech/go-ens/v3"
)

// 核心概念：
// - Domain: 域名（如 alice.eth）
// - Resolver: 解析器合约
// - Registry: ENS注册表
// - Registrar: 域名注册器
// - Node: 域名的哈希表示
// - Record: 域名记录（地址、文本等）
```

## environment准备

### 2.1 ENS客户端

```go
// config/ens.go
package config

import (
    "time"
)

type ENSConfig struct {
    // 网络配置
    EthereumRPC    string
    ChainID        int64
    
    // ENS合约地址
    RegistryAddr   string
    RegistrarAddr  string
    ResolverAddr   string
    
    // 超时配置
    Timeout        time.Duration
    
    // 缓存配置
    CacheEnabled   bool
    CacheTTL       time.Duration
}

func DefaultENSConfig() *ENSConfig {
    return &ENSConfig{
        EthereumRPC:   "https://mainnet.infura.io/v3/YOUR_PROJECT_ID",
        ChainID:       1, // 以太坊主网
        RegistryAddr:  "0x00000000000C2E074eC69A0dFb2997BA6C7d2e1e", // ENS Registry
        RegistrarAddr: "0x57f1887a8BF19b14fC0dF6Fd9B2acc9Af147eA85", // ETH Registrar
        ResolverAddr:  "0x4976fb03C32e5B8cfe2b6cCB31c09Ba78EBaBa41", // Public Resolver
        Timeout:       30 * time.Second,
        CacheEnabled:  true,
        CacheTTL:      5 * time.Minute,
    }
}

// 测试网配置
func GoerliENSConfig() *ENSConfig {
    cfg := DefaultENSConfig()
    cfg.EthereumRPC = "https://goerli.infura.io/v3/YOUR_PROJECT_ID"
    cfg.ChainID = 5
    cfg.RegistryAddr = "0x00000000000C2E074eC69A0dFb2997BA6C7d2e1e"
    return cfg
}
```

## 域名解析

### 3.1 解析器

```go
// resolver/ens_resolver.go
package resolver

import (
    "context"
    "fmt"
    "strings"
    "time"

    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/ethereum/go-ethereum/common"
    "github.com/wealdtech/go-ens/v3"
    
    "your-project/config"
)

type ENSResolver struct {
    client *ethclient.Client
    config *config.ENSConfig
    cache  map[string]CacheEntry
}

type CacheEntry struct {
    Value     interface{}
    ExpiresAt time.Time
}

func NewENSResolver(cfg *config.ENSConfig) (*ENSResolver, error) {
    client, err := ethclient.Dial(cfg.EthereumRPC)
    if err != nil {
        return nil, fmt.Errorf("连接以太坊节点失败: %v", err)
    }

    return &ENSResolver{
        client: client,
        config: cfg,
        cache:  make(map[string]CacheEntry),
    }, nil
}

// 解析域名到地址
func (er *ENSResolver) Resolve(domain string) (common.Address, error) {
    // 检查缓存
    if er.config.CacheEnabled {
        if entry, exists := er.cache[domain]; exists && time.Now().Before(entry.ExpiresAt) {
            if addr, ok := entry.Value.(common.Address); ok {
                return addr, nil
            }
        }
    }

    // 标准化域名
    domain = strings.ToLower(domain)
    if !strings.HasSuffix(domain, ".eth") {
        domain += ".eth"
    }

    ctx, cancel := context.WithTimeout(context.Background(), er.config.Timeout)
    defer cancel()

    // 解析地址
    address, err := ens.Resolve(er.client, domain)
    if err != nil {
        return common.Address{}, fmt.Errorf("解析域名失败: %v", err)
    }

    // 缓存结果
    if er.config.CacheEnabled {
        er.cache[domain] = CacheEntry{
            Value:     address,
            ExpiresAt: time.Now().Add(er.config.CacheTTL),
        }
    }

    return address, nil
}

// 批量解析域名
func (er *ENSResolver) ResolveBatch(domains []string) (map[string]common.Address, error) {
    results := make(map[string]common.Address)
    
    for _, domain := range domains {
        address, err := er.Resolve(domain)
        if err != nil {
            // 记录错误但继续处理其他域名
            fmt.Printf("解析域名 %s 失败: %v\n", domain, err)
            continue
        }
        results[domain] = address
    }

    return results, nil
}

// 解析文本记录
func (er *ENSResolver) ResolveText(domain, key string) (string, error) {
    // 检查缓存
    cacheKey := fmt.Sprintf("%s:text:%s", domain, key)
    if er.config.CacheEnabled {
        if entry, exists := er.cache[cacheKey]; exists && time.Now().Before(entry.ExpiresAt) {
            if text, ok := entry.Value.(string); ok {
                return text, nil
            }
        }
    }

    domain = strings.ToLower(domain)
    if !strings.HasSuffix(domain, ".eth") {
        domain += ".eth"
    }

    ctx, cancel := context.WithTimeout(context.Background(), er.config.Timeout)
    defer cancel()

    text, err := ens.Text(er.client, domain, key)
    if err != nil {
        return "", fmt.Errorf("解析文本记录失败: %v", err)
    }

    // 缓存结果
    if er.config.CacheEnabled {
        er.cache[cacheKey] = CacheEntry{
            Value:     text,
            ExpiresAt: time.Now().Add(er.config.CacheTTL),
        }
    }

    return text, nil
}

// 解析内容哈希
func (er *ENSResolver) ResolveContentHash(domain string) ([]byte, error) {
    domain = strings.ToLower(domain)
    if !strings.HasSuffix(domain, ".eth") {
        domain += ".eth"
    }

    ctx, cancel := context.WithTimeout(context.Background(), er.config.Timeout)
    defer cancel()

    contentHash, err := ens.Contenthash(er.client, domain)
    if err != nil {
        return nil, fmt.Errorf("解析内容哈希失败: %v", err)
    }

    return contentHash, nil
}

// 获取域名的所有记录
func (er *ENSResolver) GetAllRecords(domain string) (*DomainRecords, error) {
    domain = strings.ToLower(domain)
    if !strings.HasSuffix(domain, ".eth") {
        domain += ".eth"
    }

    records := &DomainRecords{
        Domain: domain,
    }

    // 解析地址
    if address, err := er.Resolve(domain); err == nil {
        records.Address = address
    }

    // 解析常用文本记录
    textKeys := []string{"email", "url", "avatar", "description", "notice", "keywords"}
    records.TextRecords = make(map[string]string)
    
    for _, key := range textKeys {
        if text, err := er.ResolveText(domain, key); err == nil && text != "" {
            records.TextRecords[key] = text
        }
    }

    // 解析内容哈希
    if contentHash, err := er.ResolveContentHash(domain); err == nil {
        records.ContentHash = contentHash
    }

    return records, nil
}

// 检查域名是否存在
func (er *ENSResolver) DomainExists(domain string) (bool, error) {
    domain = strings.ToLower(domain)
    if !strings.HasSuffix(domain, ".eth") {
        domain += ".eth"
    }

    ctx, cancel := context.WithTimeout(context.Background(), er.config.Timeout)
    defer cancel()

    exists, err := ens.DomainExists(er.client, domain)
    if err != nil {
        return false, fmt.Errorf("检查域名存在性失败: %v", err)
    }

    return exists, nil
}

type DomainRecords struct {
    Domain      string
    Address     common.Address
    TextRecords map[string]string
    ContentHash []byte
}
```

## 反向解析

### 4.1 反向解析器

```go
// resolver/reverse_resolver.go
package resolver

import (
    "context"
    "fmt"
    "strings"

    "github.com/ethereum/go-ethereum/common"
    "github.com/wealdtech/go-ens/v3"
)

// 反向解析地址到域名
func (er *ENSResolver) ReverseResolve(address common.Address) (string, error) {
    // 检查缓存
    cacheKey := fmt.Sprintf("reverse:%s", address.Hex())
    if er.config.CacheEnabled {
        if entry, exists := er.cache[cacheKey]; exists && time.Now().Before(entry.ExpiresAt) {
            if domain, ok := entry.Value.(string); ok {
                return domain, nil
            }
        }
    }

    ctx, cancel := context.WithTimeout(context.Background(), er.config.Timeout)
    defer cancel()

    domain, err := ens.ReverseResolve(er.client, address)
    if err != nil {
        return "", fmt.Errorf("反向解析失败: %v", err)
    }

    // 验证反向解析结果
    if domain != "" {
        resolvedAddress, err := er.Resolve(domain)
        if err != nil || resolvedAddress != address {
            return "", fmt.Errorf("反向解析验证失败")
        }
    }

    // 缓存结果
    if er.config.CacheEnabled {
        er.cache[cacheKey] = CacheEntry{
            Value:     domain,
            ExpiresAt: time.Now().Add(er.config.CacheTTL),
        }
    }

    return domain, nil
}

// 批量反向解析
func (er *ENSResolver) ReverseResolveBatch(addresses []common.Address) (map[common.Address]string, error) {
    results := make(map[common.Address]string)
    
    for _, address := range addresses {
        domain, err := er.ReverseResolve(address)
        if err != nil {
            fmt.Printf("反向解析地址 %s 失败: %v\n", address.Hex(), err)
            continue
        }
        if domain != "" {
            results[address] = domain
        }
    }

    return results, nil
}

// 获取地址的主域名
func (er *ENSResolver) GetPrimaryDomain(address common.Address) (string, error) {
    return er.ReverseResolve(address)
}
```

## 域名注册

### 5.1 注册管理器

```go
// registrar/manager.go
package registrar

import (
    "context"
    "fmt"
    "math/big"
    "strings"
    "time"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/wealdtech/go-ens/v3"
    
    "your-project/config"
    "your-project/wallet"
)

type RegistrarManager struct {
    client *ethclient.Client
    config *config.ENSConfig
    wallet *wallet.WalletManager
}

func NewRegistrarManager(cfg *config.ENSConfig, wallet *wallet.WalletManager) (*RegistrarManager, error) {
    client, err := ethclient.Dial(cfg.EthereumRPC)
    if err != nil {
        return nil, fmt.Errorf("连接以太坊节点失败: %v", err)
    }

    return &RegistrarManager{
        client: client,
        config: cfg,
        wallet: wallet,
    }, nil
}

// 检查域名可用性
func (rm *RegistrarManager) IsAvailable(name string) (bool, error) {
    name = strings.ToLower(name)
    
    ctx, cancel := context.WithTimeout(context.Background(), rm.config.Timeout)
    defer cancel()

    available, err := ens.NameAvailable(rm.client, name)
    if err != nil {
        return false, fmt.Errorf("检查域名可用性失败: %v", err)
    }

    return available, nil
}

// 获取域名注册费用
func (rm *RegistrarManager) GetRegistrationCost(name string, duration *big.Int) (*big.Int, error) {
    name = strings.ToLower(name)
    
    ctx, cancel := context.WithTimeout(context.Background(), rm.config.Timeout)
    defer cancel()

    cost, err := ens.RentCost(rm.client, name, duration)
    if err != nil {
        return nil, fmt.Errorf("获取注册费用失败: %v", err)
    }

    return cost, nil
}

// 提交域名注册承诺
func (rm *RegistrarManager) MakeCommitment(name string, owner common.Address, secret [32]byte) (*types.Transaction, error) {
    name = strings.ToLower(name)
    
    // 创建承诺
    commitment, err := ens.MakeCommitment(name, owner, secret)
    if err != nil {
        return nil, fmt.Errorf("创建承诺失败: %v", err)
    }

    // 获取注册器合约
    registrar, err := ens.NewRegistrar(rm.client, common.HexToAddress(rm.config.RegistrarAddr))
    if err != nil {
        return nil, fmt.Errorf("获取注册器合约失败: %v", err)
    }

    // 创建交易选项
    auth, err := rm.wallet.CreateTransactOpts(big.NewInt(rm.config.ChainID))
    if err != nil {
        return nil, err
    }

    // 提交承诺
    tx, err := registrar.Commit(auth, commitment)
    if err != nil {
        return nil, fmt.Errorf("提交承诺失败: %v", err)
    }

    return tx, nil
}

// 注册域名
func (rm *RegistrarManager) Register(name string, owner common.Address, duration *big.Int, secret [32]byte, resolver common.Address) (*types.Transaction, error) {
    name = strings.ToLower(name)
    
    // 检查域名可用性
    available, err := rm.IsAvailable(name)
    if err != nil {
        return nil, err
    }
    if !available {
        return nil, fmt.Errorf("域名不可用: %s", name)
    }

    // 获取注册费用
    cost, err := rm.GetRegistrationCost(name, duration)
    if err != nil {
        return nil, err
    }

    // 获取注册器合约
    registrar, err := ens.NewRegistrar(rm.client, common.HexToAddress(rm.config.RegistrarAddr))
    if err != nil {
        return nil, fmt.Errorf("获取注册器合约失败: %v", err)
    }

    // 创建交易选项
    auth, err := rm.wallet.CreateTransactOpts(big.NewInt(rm.config.ChainID))
    if err != nil {
        return nil, err
    }

    // 设置注册费用
    auth.Value = cost

    // 注册域名
    tx, err := registrar.RegisterWithConfig(auth, name, owner, duration, secret, resolver, owner)
    if err != nil {
        return nil, fmt.Errorf("注册域名失败: %v", err)
    }

    return tx, nil
}

// 续费域名
func (rm *RegistrarManager) Renew(name string, duration *big.Int) (*types.Transaction, error) {
    name = strings.ToLower(name)
    
    // 获取续费费用
    cost, err := rm.GetRegistrationCost(name, duration)
    if err != nil {
        return nil, err
    }

    // 获取注册器合约
    registrar, err := ens.NewRegistrar(rm.client, common.HexToAddress(rm.config.RegistrarAddr))
    if err != nil {
        return nil, fmt.Errorf("获取注册器合约失败: %v", err)
    }

    // 创建交易选项
    auth, err := rm.wallet.CreateTransactOpts(big.NewInt(rm.config.ChainID))
    if err != nil {
        return nil, err
    }

    // 设置续费费用
    auth.Value = cost

    // 续费域名
    tx, err := registrar.Renew(auth, name, duration)
    if err != nil {
        return nil, fmt.Errorf("续费域名失败: %v", err)
    }

    return tx, nil
}

// 获取域名到期时间
func (rm *RegistrarManager) GetExpiry(name string) (time.Time, error) {
    name = strings.ToLower(name)
    
    ctx, cancel := context.WithTimeout(context.Background(), rm.config.Timeout)
    defer cancel()

    expiry, err := ens.NameExpires(rm.client, name)
    if err != nil {
        return time.Time{}, fmt.Errorf("获取域名到期时间失败: %v", err)
    }

    return time.Unix(expiry.Int64(), 0), nil
}

// 生成随机密钥
func (rm *RegistrarManager) GenerateSecret() [32]byte {
    var secret [32]byte
    copy(secret[:], crypto.Keccak256([]byte(fmt.Sprintf("%d", time.Now().UnixNano()))))
    return secret
}

type RegistrationInfo struct {
    Name       string
    Owner      common.Address
    Expiry     time.Time
    Available  bool
    Cost       *big.Int
}

// 获取域名注册信息
func (rm *RegistrarManager) GetRegistrationInfo(name string, duration *big.Int) (*RegistrationInfo, error) {
    name = strings.ToLower(name)
    
    info := &RegistrationInfo{
        Name: name,
    }

    // 检查可用性
    available, err := rm.IsAvailable(name)
    if err != nil {
        return nil, err
    }
    info.Available = available

    // 如果不可用，获取所有者和到期时间
    if !available {
        owner, err := ens.Owner(rm.client, name+".eth")
        if err == nil {
            info.Owner = owner
        }

        expiry, err := rm.GetExpiry(name)
        if err == nil {
            info.Expiry = expiry
        }
    }

    // 获取注册费用
    cost, err := rm.GetRegistrationCost(name, duration)
    if err == nil {
        info.Cost = cost
    }

    return info, nil
}
```

## 记录管理

### 6.1 记录管理器

```go
// records/manager.go
package records

import (
    "context"
    "fmt"
    "math/big"
    "strings"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/wealdtech/go-ens/v3"
    
    "your-project/config"
    "your-project/wallet"
)

type RecordsManager struct {
    client *ethclient.Client
    config *config.ENSConfig
    wallet *wallet.WalletManager
}

func NewRecordsManager(cfg *config.ENSConfig, wallet *wallet.WalletManager) (*RecordsManager, error) {
    client, err := ethclient.Dial(cfg.EthereumRPC)
    if err != nil {
        return nil, fmt.Errorf("连接以太坊节点失败: %v", err)
    }

    return &RecordsManager{
        client: client,
        config: cfg,
        wallet: wallet,
    }, nil
}

// 设置地址记录
func (rm *RecordsManager) SetAddress(domain string, address common.Address) (*types.Transaction, error) {
    domain = strings.ToLower(domain)
    if !strings.HasSuffix(domain, ".eth") {
        domain += ".eth"
    }

    // 获取解析器地址
    resolver, err := ens.Resolver(rm.client, domain)
    if err != nil {
        return nil, fmt.Errorf("获取解析器失败: %v", err)
    }

    // 创建解析器合约实例
    resolverContract, err := ens.NewResolver(rm.client, resolver)
    if err != nil {
        return nil, fmt.Errorf("创建解析器合约失败: %v", err)
    }

    // 创建交易选项
    auth, err := rm.wallet.CreateTransactOpts(big.NewInt(rm.config.ChainID))
    if err != nil {
        return nil, err
    }

    // 计算域名节点
    node := ens.NameHash(domain)

    // 设置地址记录
    tx, err := resolverContract.SetAddr(auth, node, address)
    if err != nil {
        return nil, fmt.Errorf("设置地址记录失败: %v", err)
    }

    return tx, nil
}

// 设置文本记录
func (rm *RecordsManager) SetText(domain, key, value string) (*types.Transaction, error) {
    domain = strings.ToLower(domain)
    if !strings.HasSuffix(domain, ".eth") {
        domain += ".eth"
    }

    // 获取解析器地址
    resolver, err := ens.Resolver(rm.client, domain)
    if err != nil {
        return nil, fmt.Errorf("获取解析器失败: %v", err)
    }

    // 创建解析器合约实例
    resolverContract, err := ens.NewResolver(rm.client, resolver)
    if err != nil {
        return nil, fmt.Errorf("创建解析器合约失败: %v", err)
    }

    // 创建交易选项
    auth, err := rm.wallet.CreateTransactOpts(big.NewInt(rm.config.ChainID))
    if err != nil {
        return nil, err
    }

    // 计算域名节点
    node := ens.NameHash(domain)

    // 设置文本记录
    tx, err := resolverContract.SetText(auth, node, key, value)
    if err != nil {
        return nil, fmt.Errorf("设置文本记录失败: %v", err)
    }

    return tx, nil
}

// 设置内容哈希
func (rm *RecordsManager) SetContentHash(domain string, contentHash []byte) (*types.Transaction, error) {
    domain = strings.ToLower(domain)
    if !strings.HasSuffix(domain, ".eth") {
        domain += ".eth"
    }

    // 获取解析器地址
    resolver, err := ens.Resolver(rm.client, domain)
    if err != nil {
        return nil, fmt.Errorf("获取解析器失败: %v", err)
    }

    // 创建解析器合约实例
    resolverContract, err := ens.NewResolver(rm.client, resolver)
    if err != nil {
        return nil, fmt.Errorf("创建解析器合约失败: %v", err)
    }

    // 创建交易选项
    auth, err := rm.wallet.CreateTransactOpts(big.NewInt(rm.config.ChainID))
    if err != nil {
        return nil, err
    }

    // 计算域名节点
    node := ens.NameHash(domain)

    // 设置内容哈希
    tx, err := resolverContract.SetContenthash(auth, node, contentHash)
    if err != nil {
        return nil, fmt.Errorf("设置内容哈希失败: %v", err)
    }

    return tx, nil
}

// 设置反向记录
func (rm *RecordsManager) SetReverse(domain string) (*types.Transaction, error) {
    domain = strings.ToLower(domain)
    if !strings.HasSuffix(domain, ".eth") {
        domain += ".eth"
    }

    // 获取反向注册器
    reverseRegistrar, err := ens.ReverseRegistrar(rm.client)
    if err != nil {
        return nil, fmt.Errorf("获取反向注册器失败: %v", err)
    }

    // 创建交易选项
    auth, err := rm.wallet.CreateTransactOpts(big.NewInt(rm.config.ChainID))
    if err != nil {
        return nil, err
    }

    // 设置反向记录
    tx, err := reverseRegistrar.SetName(auth, domain)
    if err != nil {
        return nil, fmt.Errorf("设置反向记录失败: %v", err)
    }

    return tx, nil
}

// 批量设置记录
func (rm *RecordsManager) SetMultipleRecords(domain string, records map[string]interface{}) ([]*types.Transaction, error) {
    var transactions []*types.Transaction

    for key, value := range records {
        switch key {
        case "address":
            if addr, ok := value.(common.Address); ok {
                tx, err := rm.SetAddress(domain, addr)
                if err != nil {
                    return transactions, fmt.Errorf("设置地址记录失败: %v", err)
                }
                transactions = append(transactions, tx)
            }
        case "contenthash":
            if hash, ok := value.([]byte); ok {
                tx, err := rm.SetContentHash(domain, hash)
                if err != nil {
                    return transactions, fmt.Errorf("设置内容哈希失败: %v", err)
                }
                transactions = append(transactions, tx)
            }
        default:
            // 文本记录
            if text, ok := value.(string); ok {
                tx, err := rm.SetText(domain, key, text)
                if err != nil {
                    return transactions, fmt.Errorf("设置文本记录 %s 失败: %v", key, err)
                }
                transactions = append(transactions, tx)
            }
        }
    }

    return transactions, nil
}
```

## 子域名操作

### 7.1 子域名管理器

```go
// subdomain/manager.go
package subdomain

import (
    "context"
    "fmt"
    "math/big"
    "strings"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/wealdtech/go-ens/v3"
    
    "your-project/config"
    "your-project/wallet"
)

type SubdomainManager struct {
    client *ethclient.Client
    config *config.ENSConfig
    wallet *wallet.WalletManager
}

func NewSubdomainManager(cfg *config.ENSConfig, wallet *wallet.WalletManager) (*SubdomainManager, error) {
    client, err := ethclient.Dial(cfg.EthereumRPC)
    if err != nil {
        return nil, fmt.Errorf("连接以太坊节点失败: %v", err)
    }

    return &SubdomainManager{
        client: client,
        config: cfg,
        wallet: wallet,
    }, nil
}

// 创建子域名
func (sm *SubdomainManager) CreateSubdomain(parentDomain, subdomain string, owner common.Address, resolver common.Address) (*types.Transaction, error) {
    parentDomain = strings.ToLower(parentDomain)
    subdomain = strings.ToLower(subdomain)
    
    if !strings.HasSuffix(parentDomain, ".eth") {
        parentDomain += ".eth"
    }

    fullDomain := subdomain + "." + parentDomain

    // 获取ENS注册表
    registry, err := ens.NewRegistry(sm.client, common.HexToAddress(sm.config.RegistryAddr))
    if err != nil {
        return nil, fmt.Errorf("获取ENS注册表失败: %v", err)
    }

    // 创建交易选项
    auth, err := sm.wallet.CreateTransactOpts(big.NewInt(sm.config.ChainID))
    if err != nil {
        return nil, err
    }

    // 计算父域名和子域名节点
    parentNode := ens.NameHash(parentDomain)
    subdomainLabel := crypto.Keccak256Hash([]byte(subdomain))

    // 设置子域名记录
    tx, err := registry.SetSubnodeRecord(auth, parentNode, subdomainLabel, owner, resolver, big.NewInt(0))
    if err != nil {
        return nil, fmt.Errorf("创建子域名失败: %v", err)
    }

    return tx, nil
}

// 设置子域名所有者
func (sm *SubdomainManager) SetSubdomainOwner(parentDomain, subdomain string, owner common.Address) (*types.Transaction, error) {
    parentDomain = strings.ToLower(parentDomain)
    subdomain = strings.ToLower(subdomain)
    
    if !strings.HasSuffix(parentDomain, ".eth") {
        parentDomain += ".eth"
    }

    // 获取ENS注册表
    registry, err := ens.NewRegistry(sm.client, common.HexToAddress(sm.config.RegistryAddr))
    if err != nil {
        return nil, fmt.Errorf("获取ENS注册表失败: %v", err)
    }

    // 创建交易选项
    auth, err := sm.wallet.CreateTransactOpts(big.NewInt(sm.config.ChainID))
    if err != nil {
        return nil, err
    }

    // 计算父域名和子域名节点
    parentNode := ens.NameHash(parentDomain)
    subdomainLabel := crypto.Keccak256Hash([]byte(subdomain))

    // 设置子域名所有者
    tx, err := registry.SetSubnodeOwner(auth, parentNode, subdomainLabel, owner)
    if err != nil {
        return nil, fmt.Errorf("设置子域名所有者失败: %v", err)
    }

    return tx, nil
}

// 获取子域名所有者
func (sm *SubdomainManager) GetSubdomainOwner(parentDomain, subdomain string) (common.Address, error) {
    parentDomain = strings.ToLower(parentDomain)
    subdomain = strings.ToLower(subdomain)
    
    if !strings.HasSuffix(parentDomain, ".eth") {
        parentDomain += ".eth"
    }

    fullDomain := subdomain + "." + parentDomain

    ctx, cancel := context.WithTimeout(context.Background(), sm.config.Timeout)
    defer cancel()

    owner, err := ens.Owner(sm.client, fullDomain)
    if err != nil {
        return common.Address{}, fmt.Errorf("获取子域名所有者失败: %v", err)
    }

    return owner, nil
}

// 列出域名的子域名
func (sm *SubdomainManager) ListSubdomains(parentDomain string) ([]SubdomainInfo, error) {
    // 注意：这个功能需要监听ENS事件或使用第三方索引服务
    // 这里提供一个基础实现框架
    
    var subdomains []SubdomainInfo
    
    // 实际实现需要：
    // 1. 监听NewOwner事件
    // 2. 过滤出指定父域名的子域名
    // 3. 获取每个子域名的详细信息
    
    return subdomains, nil
}

type SubdomainInfo struct {
    Name     string
    Owner    common.Address
    Resolver common.Address
    TTL      *big.Int
}
```

## 高级功能

### 8.1 ENS工具集

```go
// utils/ens_utils.go
package utils

import (
    "fmt"
    "regexp"
    "strings"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/wealdtech/go-ens/v3"
)

type ENSUtils struct{}

func NewENSUtils() *ENSUtils {
    return &ENSUtils{}
}

// 验证域名格式
func (eu *ENSUtils) ValidateDomainName(domain string) error {
    domain = strings.ToLower(domain)
    
    // 检查长度
    if len(domain) < 3 {
        return fmt.Errorf("域名长度不能少于3个字符")
    }
    
    if len(domain) > 253 {
        return fmt.Errorf("域名长度不能超过253个字符")
    }

    // 检查字符
    validPattern := regexp.MustCompile(`^[a-z0-9-]+(\.[a-z0-9-]+)*$`)
    if !validPattern.MatchString(domain) {
        return fmt.Errorf("域名包含无效字符")
    }

    // 检查是否以连字符开头或结尾
    if strings.HasPrefix(domain, "-") || strings.HasSuffix(domain, "-") {
        return fmt.Errorf("域名不能以连字符开头或结尾")
    }

    return nil
}

// 标准化域名
func (eu *ENSUtils) NormalizeDomain(domain string) string {
    domain = strings.ToLower(domain)
    domain = strings.TrimSpace(domain)
    
    if !strings.HasSuffix(domain, ".eth") {
        domain += ".eth"
    }
    
    return domain
}

// 计算域名哈希
func (eu *ENSUtils) NameHash(domain string) [32]byte {
    return ens.NameHash(domain)
}

// 计算标签哈希
func (eu *ENSUtils) LabelHash(label string) [32]byte {
    return crypto.Keccak256Hash([]byte(label))
}

// 检查是否为有效的ENS域名
func (eu *ENSUtils) IsENSDomain(domain string) bool {
    return strings.HasSuffix(strings.ToLower(domain), ".eth")
}

// 提取域名标签
func (eu *ENSUtils) ExtractLabels(domain string) []string {
    domain = strings.ToLower(domain)
    if strings.HasSuffix(domain, ".eth") {
        domain = strings.TrimSuffix(domain, ".eth")
    }
    
    return strings.Split(domain, ".")
}

// 构建域名
func (eu *ENSUtils) BuildDomain(labels []string) string {
    domain := strings.Join(labels, ".")
    if !strings.HasSuffix(domain, ".eth") {
        domain += ".eth"
    }
    return domain
}

// 获取父域名
func (eu *ENSUtils) GetParentDomain(domain string) string {
    labels := eu.ExtractLabels(domain)
    if len(labels) <= 1 {
        return ""
    }
    
    return eu.BuildDomain(labels[1:])
}

// 获取顶级标签
func (eu *ENSUtils) GetTopLabel(domain string) string {
    labels := eu.ExtractLabels(domain)
    if len(labels) == 0 {
        return ""
    }
    
    return labels[0]
}

// 格式化地址显示
func (eu *ENSUtils) FormatAddress(address common.Address) string {
    addr := address.Hex()
    return fmt.Sprintf("%s...%s", addr[:6], addr[len(addr)-4:])
}

// 生成域名建议
func (eu *ENSUtils) GenerateSuggestions(baseName string, count int) []string {
    var suggestions []string
    
    // 添加数字后缀
    for i := 1; i <= count && len(suggestions) < count; i++ {
        suggestion := fmt.Sprintf("%s%d", baseName, i)
        suggestions = append(suggestions, suggestion)
    }
    
    // 添加常用后缀
    suffixes := []string{"app", "dao", "defi", "nft", "web3", "crypto"}
    for _, suffix := range suffixes {
        if len(suggestions) >= count {
            break
        }
        suggestion := fmt.Sprintf("%s%s", baseName, suffix)
        suggestions = append(suggestions, suggestion)
    }
    
    return suggestions
}
```

## 实际应用

### 9.1 完整ENS应用

```go
// main.go
package main

import (
    "fmt"
    "log"
    "math/big"

    "github.com/ethereum/go-ethereum/common"

    "your-project/config"
    "your-project/resolver"
    "your-project/registrar"
    "your-project/records"
    "your-project/subdomain"
    "your-project/utils"
    "your-project/wallet"
)

func main() {
    // 创建ENS配置
    cfg := config.GoerliENSConfig() // 使用测试网

    // 创建钱包
    walletManager, err := wallet.NewWalletManager("your_private_key_here")
    if err != nil {
        log.Fatal("创建钱包失败:", err)
    }

    fmt.Printf("钱包地址: %s\n", walletManager.GetAddress().Hex())

    // 创建ENS解析器
    ensResolver, err := resolver.NewENSResolver(cfg)
    if err != nil {
        log.Fatal("创建ENS解析器失败:", err)
    }

    // 创建管理器
    registrarManager, err := registrar.NewRegistrarManager(cfg, walletManager)
    if err != nil {
        log.Fatal("创建注册管理器失败:", err)
    }

    recordsManager, err := records.NewRecordsManager(cfg, walletManager)
    if err != nil {
        log.Fatal("创建记录管理器失败:", err)
    }

    subdomainManager, err := subdomain.NewSubdomainManager(cfg, walletManager)
    if err != nil {
        log.Fatal("创建子域名管理器失败:", err)
    }

    ensUtils := utils.NewENSUtils()

    // 域名解析示例
    fmt.Println("=== 域名解析示例 ===")
    
    testDomains := []string{"vitalik.eth", "ethereum.eth", "ens.eth"}
    
    for _, domain := range testDomains {
        // 检查域名是否存在
        exists, err := ensResolver.DomainExists(domain)
        if err != nil {
            log.Printf("检查域名 %s 存在性失败: %v", domain, err)
            continue
        }
        
        if !exists {
            fmt.Printf("域名 %s 不存在\n", domain)
            continue
        }

        // 解析地址
        address, err := ensResolver.Resolve(domain)
        if err != nil {
            log.Printf("解析域名 %s 失败: %v", domain, err)
            continue
        }
        
        fmt.Printf("域名: %s -> 地址: %s\n", domain, address.Hex())

        // 反向解析
        reverseDomain, err := ensResolver.ReverseResolve(address)
        if err != nil {
            log.Printf("反向解析地址 %s 失败: %v", address.Hex(), err)
        } else if reverseDomain != "" {
            fmt.Printf("地址: %s -> 域名: %s\n", ensUtils.FormatAddress(address), reverseDomain)
        }

        // 获取所有记录
        records, err := ensResolver.GetAllRecords(domain)
        if err != nil {
            log.Printf("获取域名 %s 记录失败: %v", domain, err)
        } else {
            fmt.Printf("域名 %s 的记录:\n", domain)
            fmt.Printf("  地址: %s\n", records.Address.Hex())
            for key, value := range records.TextRecords {
                fmt.Printf("  %s: %s\n", key, value)
            }
            if len(records.ContentHash) > 0 {
                fmt.Printf("  内容哈希: %x\n", records.ContentHash)
            }
        }
    }

    // 域名注册示例
    fmt.Println("\n=== 域名注册示例 ===")
    
    testName := "mytest123"
    
    // 验证域名格式
    err = ensUtils.ValidateDomainName(testName)
    if err != nil {
        log.Printf("域名格式无效: %v", err)
    } else {
        fmt.Printf("域名 %s 格式有效\n", testName)
    }

    // 检查域名可用性
    available, err := registrarManager.IsAvailable(testName)
    if err != nil {
        log.Printf("检查域名可用性失败: %v", err)
    } else {
        fmt.Printf("域名 %s 可用性: %t\n", testName, available)
    }

    // 获取注册信息
    duration := big.NewInt(365 * 24 * 3600) // 1年
    regInfo, err := registrarManager.GetRegistrationInfo(testName, duration)
    if err != nil {
        log.Printf("获取注册信息失败: %v", err)
    } else {
        fmt.Printf("域名注册信息:\n")
        fmt.Printf("  名称: %s\n", regInfo.Name)
        fmt.Printf("  可用: %t\n", regInfo.Available)
        if regInfo.Cost != nil {
            fmt.Printf("  费用: %s Wei\n", regInfo.Cost.String())
        }
        if !regInfo.Available {
            fmt.Printf("  所有者: %s\n", regInfo.Owner.Hex())
            fmt.Printf("  到期时间: %s\n", regInfo.Expiry.Format("2006-01-02 15:04:05"))
        }
    }

    // 域名工具示例
    fmt.Println("\n=== 域名工具示例 ===")
    
    testDomain := "alice.eth"
    
    // 计算域名哈希
    nameHash := ensUtils.NameHash(testDomain)
    fmt.Printf("域名 %s 的哈希: %x\n", testDomain, nameHash)

    // 提取标签
    labels := ensUtils.ExtractLabels(testDomain)
    fmt.Printf("域名 %s 的标签: %+v\n", testDomain, labels)

    // 获取父域名
    parentDomain := ensUtils.GetParentDomain("sub.alice.eth")
    fmt.Printf("sub.alice.eth 的父域名: %s\n", parentDomain)

    // 生成域名建议
    suggestions := ensUtils.GenerateSuggestions("alice", 5)
    fmt.Printf("基于 'alice' 的域名建议: %+v\n", suggestions)

    // 批量解析示例
    fmt.Println("\n=== 批量解析示例 ===")
    
    batchDomains := []string{"vitalik.eth", "ethereum.eth", "ens.eth"}
    batchResults, err := ensResolver.ResolveBatch(batchDomains)
    if err != nil {
        log.Printf("批量解析失败: %v", err)
    } else {
        fmt.Println("批量解析结果:")
        for domain, address := range batchResults {
            fmt.Printf("  %s -> %s\n", domain, address.Hex())
        }
    }

    // 批量反向解析示例
    var addresses []common.Address
    for _, address := range batchResults {
        addresses = append(addresses, address)
    }
    
    reverseBatchResults, err := ensResolver.ReverseResolveBatch(addresses)
    if err != nil {
        log.Printf("批量反向解析失败: %v", err)
    } else {
        fmt.Println("批量反向解析结果:")
        for address, domain := range reverseBatchResults {
            fmt.Printf("  %s -> %s\n", ensUtils.FormatAddress(address), domain)
        }
    }

    fmt.Println("ENS操作演示完成!")
}
```
