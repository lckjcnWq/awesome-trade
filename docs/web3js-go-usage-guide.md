# Web3.js Go绑定 详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [环境准备](#环境准备)
3. [以太坊连接](#以太坊连接)
4. [账户管理](#账户管理)
5. [智能合约交互](#智能合约交互)
6. [交易处理](#交易处理)
7. [事件监听](#事件监听)
8. [工具函数](#工具函数)
9. [实际应用](#实际应用)

## 基础概念

### 1.1 Web3.js Go绑定简介

Web3.js Go绑定提供了在Go语言中使用类似Web3.js API的功能，用于与以太坊区块链进行交互。

```bash
# 安装Web3相关依赖
go get github.com/ethereum/go-ethereum
go get github.com/ethereum/go-ethereum/ethclient
go get github.com/ethereum/go-ethereum/accounts/abi
go get github.com/shopspring/decimal
```

### 1.2 核心组件

```go
// 主要包导入
import (
    "context"
    "math/big"
    "crypto/ecdsa"
    
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/crypto"
)

// 核心概念：
// - Provider: 区块链节点连接提供者
// - Signer: 交易签名器
// - Contract: 智能合约实例
// - Transaction: 交易对象
// - Event: 事件监听器
```

## 环境准备

### 2.1 Web3配置

```go
// config/web3.go
package config

import (
    "math/big"
    "time"
)

type Web3Config struct {
    // 网络配置
    NetworkName    string
    ChainID        *big.Int
    NetworkID      *big.Int
    
    // RPC配置
    RPCURL         string
    WSUrl          string
    
    // 交易配置
    GasLimit       uint64
    GasPrice       *big.Int
    MaxFeePerGas   *big.Int
    MaxPriorityFee *big.Int
    
    // 超时配置
    Timeout        time.Duration
    BlockTimeout   time.Duration
    
    // 重试配置
    RetryCount     int
    RetryDelay     time.Duration
}

func DefaultWeb3Config() *Web3Config {
    return &Web3Config{
        NetworkName:    "mainnet",
        ChainID:        big.NewInt(1),
        NetworkID:      big.NewInt(1),
        RPCURL:         "https://mainnet.infura.io/v3/YOUR_PROJECT_ID",
        WSUrl:          "wss://mainnet.infura.io/ws/v3/YOUR_PROJECT_ID",
        GasLimit:       21000,
        GasPrice:       big.NewInt(20000000000), // 20 Gwei
        MaxFeePerGas:   big.NewInt(30000000000), // 30 Gwei
        MaxPriorityFee: big.NewInt(2000000000),  // 2 Gwei
        Timeout:        30 * time.Second,
        BlockTimeout:   60 * time.Second,
        RetryCount:     3,
        RetryDelay:     time.Second,
    }
}

// 测试网配置
func GoerliConfig() *Web3Config {
    cfg := DefaultWeb3Config()
    cfg.NetworkName = "goerli"
    cfg.ChainID = big.NewInt(5)
    cfg.NetworkID = big.NewInt(5)
    cfg.RPCURL = "https://goerli.infura.io/v3/YOUR_PROJECT_ID"
    cfg.WSUrl = "wss://goerli.infura.io/ws/v3/YOUR_PROJECT_ID"
    return cfg
}

// Polygon配置
func PolygonConfig() *Web3Config {
    cfg := DefaultWeb3Config()
    cfg.NetworkName = "polygon"
    cfg.ChainID = big.NewInt(137)
    cfg.NetworkID = big.NewInt(137)
    cfg.RPCURL = "https://polygon-rpc.com/"
    cfg.WSUrl = "wss://polygon-rpc.com/"
    cfg.GasPrice = big.NewInt(30000000000) // 30 Gwei
    return cfg
}
```

## 以太坊连接

### 3.1 Provider管理

```go
// provider/manager.go
package provider

import (
    "context"
    "fmt"
    "math/big"
    "time"

    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/ethereum/go-ethereum/rpc"
    
    "your-project/config"
)

type ProviderManager struct {
    client    *ethclient.Client
    rpcClient *rpc.Client
    config    *config.Web3Config
    ctx       context.Context
    cancel    context.CancelFunc
}

func NewProviderManager(cfg *config.Web3Config) (*ProviderManager, error) {
    // 创建RPC客户端
    rpcClient, err := rpc.DialContext(context.Background(), cfg.RPCURL)
    if err != nil {
        return nil, fmt.Errorf("连接RPC失败: %v", err)
    }

    // 创建以太坊客户端
    client := ethclient.NewClient(rpcClient)

    // 验证连接
    ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
    defer cancel()

    _, err = client.NetworkID(ctx)
    if err != nil {
        return nil, fmt.Errorf("验证网络连接失败: %v", err)
    }

    providerCtx, providerCancel := context.WithCancel(context.Background())

    return &ProviderManager{
        client:    client,
        rpcClient: rpcClient,
        config:    cfg,
        ctx:       providerCtx,
        cancel:    providerCancel,
    }, nil
}

// 获取以太坊客户端
func (pm *ProviderManager) GetClient() *ethclient.Client {
    return pm.client
}

// 获取RPC客户端
func (pm *ProviderManager) GetRPCClient() *rpc.Client {
    return pm.rpcClient
}

// 获取网络信息
func (pm *ProviderManager) GetNetworkInfo() (*NetworkInfo, error) {
    ctx, cancel := context.WithTimeout(pm.ctx, pm.config.Timeout)
    defer cancel()

    // 获取链ID
    chainID, err := pm.client.ChainID(ctx)
    if err != nil {
        return nil, fmt.Errorf("获取链ID失败: %v", err)
    }

    // 获取网络ID
    networkID, err := pm.client.NetworkID(ctx)
    if err != nil {
        return nil, fmt.Errorf("获取网络ID失败: %v", err)
    }

    // 获取最新区块号
    blockNumber, err := pm.client.BlockNumber(ctx)
    if err != nil {
        return nil, fmt.Errorf("获取区块号失败: %v", err)
    }

    // 获取Gas价格
    gasPrice, err := pm.client.SuggestGasPrice(ctx)
    if err != nil {
        return nil, fmt.Errorf("获取Gas价格失败: %v", err)
    }

    return &NetworkInfo{
        ChainID:     chainID,
        NetworkID:   networkID,
        BlockNumber: blockNumber,
        GasPrice:    gasPrice,
        Name:        pm.config.NetworkName,
    }, nil
}

// 检查连接状态
func (pm *ProviderManager) IsConnected() bool {
    ctx, cancel := context.WithTimeout(pm.ctx, 5*time.Second)
    defer cancel()

    _, err := pm.client.NetworkID(ctx)
    return err == nil
}

// 获取区块信息
func (pm *ProviderManager) GetBlock(blockNumber *big.Int) (*BlockInfo, error) {
    ctx, cancel := context.WithTimeout(pm.ctx, pm.config.BlockTimeout)
    defer cancel()

    block, err := pm.client.BlockByNumber(ctx, blockNumber)
    if err != nil {
        return nil, fmt.Errorf("获取区块失败: %v", err)
    }

    return &BlockInfo{
        Number:       block.Number(),
        Hash:         block.Hash(),
        ParentHash:   block.ParentHash(),
        Timestamp:    block.Time(),
        GasLimit:     block.GasLimit(),
        GasUsed:      block.GasUsed(),
        Transactions: len(block.Transactions()),
    }, nil
}

// 获取最新区块
func (pm *ProviderManager) GetLatestBlock() (*BlockInfo, error) {
    return pm.GetBlock(nil)
}

// 关闭连接
func (pm *ProviderManager) Close() {
    pm.cancel()
    pm.client.Close()
    pm.rpcClient.Close()
}

type NetworkInfo struct {
    ChainID     *big.Int
    NetworkID   *big.Int
    BlockNumber uint64
    GasPrice    *big.Int
    Name        string
}

type BlockInfo struct {
    Number       *big.Int
    Hash         common.Hash
    ParentHash   common.Hash
    Timestamp    uint64
    GasLimit     uint64
    GasUsed      uint64
    Transactions int
}
```

## 账户管理

### 4.1 钱包管理

```go
// wallet/manager.go
package wallet

import (
    "crypto/ecdsa"
    "fmt"
    "math/big"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/core/types"
    
    "your-project/provider"
)

type WalletManager struct {
    provider   *provider.ProviderManager
    privateKey *ecdsa.PrivateKey
    publicKey  *ecdsa.PublicKey
    address    common.Address
}

func NewWalletManager(provider *provider.ProviderManager, privateKeyHex string) (*WalletManager, error) {
    // 解析私钥
    privateKey, err := crypto.HexToECDSA(privateKeyHex)
    if err != nil {
        return nil, fmt.Errorf("解析私钥失败: %v", err)
    }

    // 获取公钥
    publicKey := privateKey.Public()
    publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
    if !ok {
        return nil, fmt.Errorf("获取公钥失败")
    }

    // 生成地址
    address := crypto.PubkeyToAddress(*publicKeyECDSA)

    return &WalletManager{
        provider:   provider,
        privateKey: privateKey,
        publicKey:  publicKeyECDSA,
        address:    address,
    }, nil
}

// 生成新钱包
func GenerateWallet(provider *provider.ProviderManager) (*WalletManager, error) {
    privateKey, err := crypto.GenerateKey()
    if err != nil {
        return nil, fmt.Errorf("生成私钥失败: %v", err)
    }

    privateKeyHex := fmt.Sprintf("%x", crypto.FromECDSA(privateKey))
    return NewWalletManager(provider, privateKeyHex)
}

// 获取地址
func (wm *WalletManager) GetAddress() common.Address {
    return wm.address
}

// 获取私钥十六进制
func (wm *WalletManager) GetPrivateKeyHex() string {
    return fmt.Sprintf("%x", crypto.FromECDSA(wm.privateKey))
}

// 获取余额
func (wm *WalletManager) GetBalance() (*big.Int, error) {
    client := wm.provider.GetClient()
    balance, err := client.BalanceAt(wm.provider.ctx, wm.address, nil)
    if err != nil {
        return nil, fmt.Errorf("获取余额失败: %v", err)
    }
    return balance, nil
}

// 获取Nonce
func (wm *WalletManager) GetNonce() (uint64, error) {
    client := wm.provider.GetClient()
    nonce, err := client.PendingNonceAt(wm.provider.ctx, wm.address)
    if err != nil {
        return 0, fmt.Errorf("获取nonce失败: %v", err)
    }
    return nonce, nil
}

// 签名交易
func (wm *WalletManager) SignTransaction(tx *types.Transaction, chainID *big.Int) (*types.Transaction, error) {
    signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), wm.privateKey)
    if err != nil {
        return nil, fmt.Errorf("签名交易失败: %v", err)
    }
    return signedTx, nil
}

// 签名消息
func (wm *WalletManager) SignMessage(message []byte) ([]byte, error) {
    hash := crypto.Keccak256Hash(message)
    signature, err := crypto.Sign(hash.Bytes(), wm.privateKey)
    if err != nil {
        return nil, fmt.Errorf("签名消息失败: %v", err)
    }
    return signature, nil
}

// 签名个人消息（带前缀）
func (wm *WalletManager) SignPersonalMessage(message string) ([]byte, error) {
    prefixedMessage := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
    return wm.SignMessage([]byte(prefixedMessage))
}

// 创建交易选项
func (wm *WalletManager) CreateTransactOpts(chainID *big.Int) (*bind.TransactOpts, error) {
    auth, err := bind.NewKeyedTransactorWithChainID(wm.privateKey, chainID)
    if err != nil {
        return nil, fmt.Errorf("创建交易选项失败: %v", err)
    }
    return auth, nil
}

// 发送ETH
func (wm *WalletManager) SendETH(to common.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int) (*types.Transaction, error) {
    client := wm.provider.GetClient()
    
    // 获取nonce
    nonce, err := wm.GetNonce()
    if err != nil {
        return nil, err
    }

    // 创建交易
    tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, nil)

    // 获取链ID
    networkInfo, err := wm.provider.GetNetworkInfo()
    if err != nil {
        return nil, err
    }

    // 签名交易
    signedTx, err := wm.SignTransaction(tx, networkInfo.ChainID)
    if err != nil {
        return nil, err
    }

    // 发送交易
    err = client.SendTransaction(wm.provider.ctx, signedTx)
    if err != nil {
        return nil, fmt.Errorf("发送交易失败: %v", err)
    }

    return signedTx, nil
}
```

## 智能合约交互

### 5.1 合约管理

```go
// contract/manager.go
package contract

import (
    "context"
    "fmt"
    "math/big"
    "strings"

    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    
    "your-project/provider"
    "your-project/wallet"
)

type ContractManager struct {
    provider *provider.ProviderManager
    wallet   *wallet.WalletManager
    contract *bind.BoundContract
    abi      abi.ABI
    address  common.Address
}

func NewContractManager(
    provider *provider.ProviderManager,
    wallet *wallet.WalletManager,
    contractAddress common.Address,
    abiJSON string,
) (*ContractManager, error) {
    
    // 解析ABI
    parsedABI, err := abi.JSON(strings.NewReader(abiJSON))
    if err != nil {
        return nil, fmt.Errorf("解析ABI失败: %v", err)
    }

    // 创建合约绑定
    client := provider.GetClient()
    contract := bind.NewBoundContract(contractAddress, parsedABI, client, client, client)

    return &ContractManager{
        provider: provider,
        wallet:   wallet,
        contract: contract,
        abi:      parsedABI,
        address:  contractAddress,
    }, nil
}

// 调用只读方法
func (cm *ContractManager) Call(methodName string, result interface{}, params ...interface{}) error {
    err := cm.contract.Call(nil, result, methodName, params...)
    if err != nil {
        return fmt.Errorf("调用合约方法失败: %v", err)
    }
    return nil
}

// 调用写入方法
func (cm *ContractManager) Transact(methodName string, gasLimit uint64, gasPrice *big.Int, value *big.Int, params ...interface{}) (*types.Transaction, error) {
    // 获取网络信息
    networkInfo, err := cm.provider.GetNetworkInfo()
    if err != nil {
        return nil, err
    }

    // 创建交易选项
    auth, err := cm.wallet.CreateTransactOpts(networkInfo.ChainID)
    if err != nil {
        return nil, err
    }

    // 设置交易参数
    auth.GasLimit = gasLimit
    auth.GasPrice = gasPrice
    auth.Value = value

    // 执行交易
    tx, err := cm.contract.Transact(auth, methodName, params...)
    if err != nil {
        return nil, fmt.Errorf("执行合约交易失败: %v", err)
    }

    return tx, nil
}

// 估算Gas费用
func (cm *ContractManager) EstimateGas(methodName string, params ...interface{}) (uint64, error) {
    // 打包方法调用数据
    data, err := cm.abi.Pack(methodName, params...)
    if err != nil {
        return 0, fmt.Errorf("打包方法数据失败: %v", err)
    }

    // 创建调用消息
    msg := ethereum.CallMsg{
        From: cm.wallet.GetAddress(),
        To:   &cm.address,
        Data: data,
    }

    // 估算Gas
    client := cm.provider.GetClient()
    gasLimit, err := client.EstimateGas(cm.provider.ctx, msg)
    if err != nil {
        return 0, fmt.Errorf("估算Gas失败: %v", err)
    }

    return gasLimit, nil
}

// 监听事件
func (cm *ContractManager) WatchEvents(eventName string, handler func(types.Log)) error {
    // 获取事件定义
    event, exists := cm.abi.Events[eventName]
    if !exists {
        return fmt.Errorf("事件不存在: %s", eventName)
    }

    // 创建过滤器查询
    query := ethereum.FilterQuery{
        Addresses: []common.Address{cm.address},
        Topics:    [][]common.Hash{{event.ID}},
    }

    // 订阅日志
    client := cm.provider.GetClient()
    logs := make(chan types.Log)
    sub, err := client.SubscribeFilterLogs(cm.provider.ctx, query, logs)
    if err != nil {
        return fmt.Errorf("订阅事件失败: %v", err)
    }

    // 处理事件
    go func() {
        defer sub.Unsubscribe()
        for {
            select {
            case err := <-sub.Err():
                fmt.Printf("事件订阅错误: %v\n", err)
                return
            case vLog := <-logs:
                handler(vLog)
            }
        }
    }()

    return nil
}

// 解析事件日志
func (cm *ContractManager) ParseEventLog(eventName string, log types.Log) (map[string]interface{}, error) {
    // 获取事件定义
    event, exists := cm.abi.Events[eventName]
    if !exists {
        return nil, fmt.Errorf("事件不存在: %s", eventName)
    }

    // 解析事件数据
    values := make(map[string]interface{})
    err := cm.abi.UnpackIntoMap(values, eventName, log.Data)
    if err != nil {
        return nil, fmt.Errorf("解析事件数据失败: %v", err)
    }

    // 解析索引参数
    var indexed abi.Arguments
    for _, arg := range event.Inputs {
        if arg.Indexed {
            indexed = append(indexed, arg)
        }
    }

    if len(log.Topics) > 1 {
        err = abi.ParseTopics(values, indexed, log.Topics[1:])
        if err != nil {
            return nil, fmt.Errorf("解析事件主题失败: %v", err)
        }
    }

    return values, nil
}

// 获取合约代码
func (cm *ContractManager) GetCode() ([]byte, error) {
    client := cm.provider.GetClient()
    code, err := client.CodeAt(cm.provider.ctx, cm.address, nil)
    if err != nil {
        return nil, fmt.Errorf("获取合约代码失败: %v", err)
    }
    return code, nil
}

// 检查合约是否存在
func (cm *ContractManager) ContractExists() (bool, error) {
    code, err := cm.GetCode()
    if err != nil {
        return false, err
    }
    return len(code) > 0, nil
}
```

## 交易处理

### 6.1 交易管理

```go
// transaction/manager.go
package transaction

import (
    "context"
    "fmt"
    "math/big"
    "time"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    
    "your-project/provider"
    "your-project/wallet"
)

type TransactionManager struct {
    provider *provider.ProviderManager
    wallet   *wallet.WalletManager
}

func NewTransactionManager(provider *provider.ProviderManager, wallet *wallet.WalletManager) *TransactionManager {
    return &TransactionManager{
        provider: provider,
        wallet:   wallet,
    }
}

// 创建交易
func (tm *TransactionManager) CreateTransaction(params TransactionParams) (*types.Transaction, error) {
    // 获取nonce
    nonce, err := tm.wallet.GetNonce()
    if err != nil {
        return nil, err
    }

    // 设置默认值
    if params.GasLimit == 0 {
        params.GasLimit = 21000
    }
    if params.GasPrice == nil {
        params.GasPrice, err = tm.provider.GetClient().SuggestGasPrice(tm.provider.ctx)
        if err != nil {
            return nil, fmt.Errorf("获取Gas价格失败: %v", err)
        }
    }
    if params.Value == nil {
        params.Value = big.NewInt(0)
    }

    // 创建交易
    var tx *types.Transaction
    if params.Data != nil {
        tx = types.NewTransaction(nonce, params.To, params.Value, params.GasLimit, params.GasPrice, params.Data)
    } else {
        tx = types.NewTransaction(nonce, params.To, params.Value, params.GasLimit, params.GasPrice, nil)
    }

    return tx, nil
}

// 发送交易
func (tm *TransactionManager) SendTransaction(params TransactionParams) (*TransactionResult, error) {
    // 创建交易
    tx, err := tm.CreateTransaction(params)
    if err != nil {
        return nil, err
    }

    // 获取网络信息
    networkInfo, err := tm.provider.GetNetworkInfo()
    if err != nil {
        return nil, err
    }

    // 签名交易
    signedTx, err := tm.wallet.SignTransaction(tx, networkInfo.ChainID)
    if err != nil {
        return nil, err
    }

    // 发送交易
    client := tm.provider.GetClient()
    err = client.SendTransaction(tm.provider.ctx, signedTx)
    if err != nil {
        return nil, fmt.Errorf("发送交易失败: %v", err)
    }

    return &TransactionResult{
        Hash:      signedTx.Hash(),
        Nonce:     signedTx.Nonce(),
        GasPrice:  signedTx.GasPrice(),
        GasLimit:  signedTx.Gas(),
        Value:     signedTx.Value(),
        Data:      signedTx.Data(),
        Timestamp: time.Now(),
    }, nil
}

// 等待交易确认
func (tm *TransactionManager) WaitForConfirmation(txHash common.Hash, confirmations uint64) (*types.Receipt, error) {
    client := tm.provider.GetClient()
    
    for {
        // 获取交易收据
        receipt, err := client.TransactionReceipt(tm.provider.ctx, txHash)
        if err != nil {
            time.Sleep(time.Second)
            continue
        }

        // 获取当前区块号
        currentBlock, err := client.BlockNumber(tm.provider.ctx)
        if err != nil {
            return nil, fmt.Errorf("获取当前区块号失败: %v", err)
        }

        // 检查确认数
        if currentBlock-receipt.BlockNumber.Uint64() >= confirmations {
            return receipt, nil
        }

        time.Sleep(time.Second)
    }
}

// 获取交易状态
func (tm *TransactionManager) GetTransactionStatus(txHash common.Hash) (*TransactionStatus, error) {
    client := tm.provider.GetClient()

    // 尝试获取交易收据
    receipt, err := client.TransactionReceipt(tm.provider.ctx, txHash)
    if err != nil {
        // 检查交易是否在内存池中
        _, isPending, err := client.TransactionByHash(tm.provider.ctx, txHash)
        if err != nil {
            return &TransactionStatus{
                Hash:   txHash,
                Status: "not_found",
            }, nil
        }

        if isPending {
            return &TransactionStatus{
                Hash:   txHash,
                Status: "pending",
            }, nil
        }

        return &TransactionStatus{
            Hash:   txHash,
            Status: "unknown",
        }, nil
    }

    // 确定交易状态
    status := "failed"
    if receipt.Status == 1 {
        status = "success"
    }

    // 获取当前区块号计算确认数
    currentBlock, err := client.BlockNumber(tm.provider.ctx)
    if err != nil {
        currentBlock = receipt.BlockNumber.Uint64()
    }

    confirmations := currentBlock - receipt.BlockNumber.Uint64()

    return &TransactionStatus{
        Hash:          txHash,
        Status:        status,
        BlockNumber:   receipt.BlockNumber.Uint64(),
        BlockHash:     receipt.BlockHash,
        GasUsed:       receipt.GasUsed,
        Confirmations: confirmations,
        Receipt:       receipt,
    }, nil
}

// 取消交易（通过发送更高Gas价格的空交易）
func (tm *TransactionManager) CancelTransaction(originalTx *types.Transaction, newGasPrice *big.Int) (*TransactionResult, error) {
    // 创建取消交易参数
    params := TransactionParams{
        To:       tm.wallet.GetAddress(), // 发送给自己
        Value:    big.NewInt(0),
        GasLimit: 21000,
        GasPrice: newGasPrice,
    }

    // 使用相同的nonce发送新交易
    nonce := originalTx.Nonce()
    tx := types.NewTransaction(nonce, params.To, params.Value, params.GasLimit, params.GasPrice, nil)

    // 获取网络信息
    networkInfo, err := tm.provider.GetNetworkInfo()
    if err != nil {
        return nil, err
    }

    // 签名并发送
    signedTx, err := tm.wallet.SignTransaction(tx, networkInfo.ChainID)
    if err != nil {
        return nil, err
    }

    client := tm.provider.GetClient()
    err = client.SendTransaction(tm.provider.ctx, signedTx)
    if err != nil {
        return nil, fmt.Errorf("发送取消交易失败: %v", err)
    }

    return &TransactionResult{
        Hash:      signedTx.Hash(),
        Nonce:     signedTx.Nonce(),
        GasPrice:  signedTx.GasPrice(),
        GasLimit:  signedTx.Gas(),
        Value:     signedTx.Value(),
        Timestamp: time.Now(),
    }, nil
}

type TransactionParams struct {
    To       common.Address
    Value    *big.Int
    GasLimit uint64
    GasPrice *big.Int
    Data     []byte
}

type TransactionResult struct {
    Hash      common.Hash
    Nonce     uint64
    GasPrice  *big.Int
    GasLimit  uint64
    Value     *big.Int
    Data      []byte
    Timestamp time.Time
}

type TransactionStatus struct {
    Hash          common.Hash
    Status        string
    BlockNumber   uint64
    BlockHash     common.Hash
    GasUsed       uint64
    Confirmations uint64
    Receipt       *types.Receipt
}
```

## 事件监听

### 7.1 事件管理器

```go
// events/manager.go
package events

import (
    "context"
    "fmt"
    "math/big"

    "github.com/ethereum/go-ethereum"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    
    "your-project/provider"
)

type EventManager struct {
    provider *provider.ProviderManager
}

func NewEventManager(provider *provider.ProviderManager) *EventManager {
    return &EventManager{
        provider: provider,
    }
}

// 监听新区块
func (em *EventManager) WatchNewBlocks(handler func(*types.Header)) error {
    client := em.provider.GetClient()
    
    headers := make(chan *types.Header)
    sub, err := client.SubscribeNewHead(em.provider.ctx, headers)
    if err != nil {
        return fmt.Errorf("订阅新区块失败: %v", err)
    }

    go func() {
        defer sub.Unsubscribe()
        for {
            select {
            case err := <-sub.Err():
                fmt.Printf("新区块订阅错误: %v\n", err)
                return
            case header := <-headers:
                handler(header)
            }
        }
    }()

    return nil
}

// 监听待处理交易
func (em *EventManager) WatchPendingTransactions(handler func(common.Hash)) error {
    client := em.provider.GetClient()
    
    txHashes := make(chan common.Hash)
    sub, err := client.SubscribePendingTransactions(em.provider.ctx, txHashes)
    if err != nil {
        return fmt.Errorf("订阅待处理交易失败: %v", err)
    }

    go func() {
        defer sub.Unsubscribe()
        for {
            select {
            case err := <-sub.Err():
                fmt.Printf("待处理交易订阅错误: %v\n", err)
                return
            case txHash := <-txHashes:
                handler(txHash)
            }
        }
    }()

    return nil
}

// 监听合约事件
func (em *EventManager) WatchContractEvents(
    contractAddress common.Address,
    eventSignature common.Hash,
    handler func(types.Log),
) error {
    client := em.provider.GetClient()
    
    query := ethereum.FilterQuery{
        Addresses: []common.Address{contractAddress},
        Topics:    [][]common.Hash{{eventSignature}},
    }

    logs := make(chan types.Log)
    sub, err := client.SubscribeFilterLogs(em.provider.ctx, query, logs)
    if err != nil {
        return fmt.Errorf("订阅合约事件失败: %v", err)
    }

    go func() {
        defer sub.Unsubscribe()
        for {
            select {
            case err := <-sub.Err():
                fmt.Printf("合约事件订阅错误: %v\n", err)
                return
            case vLog := <-logs:
                handler(vLog)
            }
        }
    }()

    return nil
}

// 获取历史日志
func (em *EventManager) GetLogs(query ethereum.FilterQuery) ([]types.Log, error) {
    client := em.provider.GetClient()
    
    logs, err := client.FilterLogs(em.provider.ctx, query)
    if err != nil {
        return nil, fmt.Errorf("获取历史日志失败: %v", err)
    }

    return logs, nil
}

// 获取合约历史事件
func (em *EventManager) GetContractEvents(
    contractAddress common.Address,
    eventSignature common.Hash,
    fromBlock, toBlock *big.Int,
) ([]types.Log, error) {
    
    query := ethereum.FilterQuery{
        FromBlock: fromBlock,
        ToBlock:   toBlock,
        Addresses: []common.Address{contractAddress},
        Topics:    [][]common.Hash{{eventSignature}},
    }

    return em.GetLogs(query)
}

// 监听地址转账
func (em *EventManager) WatchAddressTransfers(
    address common.Address,
    handler func(TransferEvent),
) error {
    
    // ERC20 Transfer事件签名
    transferSignature := crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))
    
    query := ethereum.FilterQuery{
        Topics: [][]common.Hash{
            {transferSignature},
            {common.BytesToHash(address.Bytes())}, // from
            {common.BytesToHash(address.Bytes())}, // to
        },
    }

    client := em.provider.GetClient()
    logs := make(chan types.Log)
    sub, err := client.SubscribeFilterLogs(em.provider.ctx, query, logs)
    if err != nil {
        return fmt.Errorf("订阅转账事件失败: %v", err)
    }

    go func() {
        defer sub.Unsubscribe()
        for {
            select {
            case err := <-sub.Err():
                fmt.Printf("转账事件订阅错误: %v\n", err)
                return
            case vLog := <-logs:
                // 解析转账事件
                if len(vLog.Topics) >= 3 && len(vLog.Data) >= 32 {
                    from := common.BytesToAddress(vLog.Topics[1].Bytes())
                    to := common.BytesToAddress(vLog.Topics[2].Bytes())
                    value := new(big.Int).SetBytes(vLog.Data[:32])
                    
                    transferEvent := TransferEvent{
                        From:            from,
                        To:              to,
                        Value:           value,
                        ContractAddress: vLog.Address,
                        BlockNumber:     vLog.BlockNumber,
                        TxHash:          vLog.TxHash,
                    }
                    
                    handler(transferEvent)
                }
            }
        }
    }()

    return nil
}

type TransferEvent struct {
    From            common.Address
    To              common.Address
    Value           *big.Int
    ContractAddress common.Address
    BlockNumber     uint64
    TxHash          common.Hash
}
```

## 工具函数

### 8.1 实用工具

```go
// utils/helpers.go
package utils

import (
    "fmt"
    "math/big"
    "strconv"
    "strings"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/shopspring/decimal"
)

// 单位转换
type UnitConverter struct{}

func NewUnitConverter() *UnitConverter {
    return &UnitConverter{}
}

// Wei转Ether
func (uc *UnitConverter) WeiToEther(wei *big.Int) decimal.Decimal {
    return decimal.NewFromBigInt(wei, -18)
}

// Ether转Wei
func (uc *UnitConverter) EtherToWei(ether decimal.Decimal) *big.Int {
    return ether.Shift(18).BigInt()
}

// Gwei转Wei
func (uc *UnitConverter) GweiToWei(gwei decimal.Decimal) *big.Int {
    return gwei.Shift(9).BigInt()
}

// Wei转Gwei
func (uc *UnitConverter) WeiToGwei(wei *big.Int) decimal.Decimal {
    return decimal.NewFromBigInt(wei, -9)
}

// 格式化地址
func FormatAddress(address common.Address) string {
    addr := address.Hex()
    return fmt.Sprintf("%s...%s", addr[:6], addr[len(addr)-4:])
}

// 验证地址
func IsValidAddress(address string) bool {
    return common.IsHexAddress(address)
}

// 计算合约地址
func CalculateContractAddress(deployer common.Address, nonce uint64) common.Address {
    return crypto.CreateAddress(deployer, nonce)
}

// 计算CREATE2合约地址
func CalculateCreate2Address(deployer common.Address, salt [32]byte, bytecodeHash [32]byte) common.Address {
    return crypto.CreateAddress2(deployer, salt, bytecodeHash)
}

// 哈希工具
type HashUtils struct{}

func NewHashUtils() *HashUtils {
    return &HashUtils{}
}

// Keccak256哈希
func (hu *HashUtils) Keccak256(data []byte) common.Hash {
    return crypto.Keccak256Hash(data)
}

// 字符串哈希
func (hu *HashUtils) Keccak256String(data string) common.Hash {
    return crypto.Keccak256Hash([]byte(data))
}

// 签名工具
type SignatureUtils struct{}

func NewSignatureUtils() *SignatureUtils {
    return &SignatureUtils{}
}

// 恢复签名者地址
func (su *SignatureUtils) RecoverSigner(message []byte, signature []byte) (common.Address, error) {
    hash := crypto.Keccak256Hash(message)
    
    // 调整v值
    if signature[64] >= 27 {
        signature[64] -= 27
    }

    pubkey, err := crypto.SigToPub(hash.Bytes(), signature)
    if err != nil {
        return common.Address{}, fmt.Errorf("恢复公钥失败: %v", err)
    }

    return crypto.PubkeyToAddress(*pubkey), nil
}

// 验证签名
func (su *SignatureUtils) VerifySignature(message []byte, signature []byte, signer common.Address) bool {
    recoveredSigner, err := su.RecoverSigner(message, signature)
    if err != nil {
        return false
    }
    return recoveredSigner == signer
}

// 编码工具
type EncodingUtils struct{}

func NewEncodingUtils() *EncodingUtils {
    return &EncodingUtils{}
}

// 十六进制编码
func (eu *EncodingUtils) ToHex(data []byte) string {
    return common.Bytes2Hex(data)
}

// 十六进制解码
func (eu *EncodingUtils) FromHex(hexStr string) []byte {
    return common.FromHex(hexStr)
}

// 大整数转十六进制
func (eu *EncodingUtils) BigIntToHex(num *big.Int) string {
    return "0x" + num.Text(16)
}

// 十六进制转大整数
func (eu *EncodingUtils) HexToBigInt(hexStr string) (*big.Int, error) {
    if strings.HasPrefix(hexStr, "0x") {
        hexStr = hexStr[2:]
    }
    
    num := new(big.Int)
    num, ok := num.SetString(hexStr, 16)
    if !ok {
        return nil, fmt.Errorf("无效的十六进制字符串: %s", hexStr)
    }
    
    return num, nil
}

// 字符串转uint64
func (eu *EncodingUtils) StringToUint64(str string) (uint64, error) {
    return strconv.ParseUint(str, 10, 64)
}

// uint64转字符串
func (eu *EncodingUtils) Uint64ToString(num uint64) string {
    return strconv.FormatUint(num, 10)
}
```

## 实际应用

### 9.1 完整Web3应用

```go
// main.go
package main

import (
    "fmt"
    "log"
    "math/big"
    "time"

    "github.com/ethereum/go-ethereum/common"
    "github.com/shopspring/decimal"

    "your-project/config"
    "your-project/provider"
    "your-project/wallet"
    "your-project/transaction"
    "your-project/events"
    "your-project/utils"
)

func main() {
    // 创建Web3配置
    cfg := config.GoerliConfig() // 使用测试网

    // 创建Provider
    providerManager, err := provider.NewProviderManager(cfg)
    if err != nil {
        log.Fatal("创建Provider失败:", err)
    }
    defer providerManager.Close()

    // 检查连接
    if !providerManager.IsConnected() {
        log.Fatal("无法连接到以太坊网络")
    }

    // 获取网络信息
    networkInfo, err := providerManager.GetNetworkInfo()
    if err != nil {
        log.Printf("获取网络信息失败: %v", err)
    } else {
        fmt.Printf("网络信息:\n")
        fmt.Printf("  链ID: %s\n", networkInfo.ChainID.String())
        fmt.Printf("  网络ID: %s\n", networkInfo.NetworkID.String())
        fmt.Printf("  当前区块: %d\n", networkInfo.BlockNumber)
        fmt.Printf("  Gas价格: %s Gwei\n", 
            utils.NewUnitConverter().WeiToGwei(networkInfo.GasPrice).String())
    }

    // 创建钱包
    walletManager, err := wallet.NewWalletManager(providerManager, "your_private_key_here")
    if err != nil {
        log.Fatal("创建钱包失败:", err)
    }

    fmt.Printf("钱包地址: %s\n", walletManager.GetAddress().Hex())

    // 获取余额
    balance, err := walletManager.GetBalance()
    if err != nil {
        log.Printf("获取余额失败: %v", err)
    } else {
        etherBalance := utils.NewUnitConverter().WeiToEther(balance)
        fmt.Printf("ETH余额: %s ETH\n", etherBalance.String())
    }

    // 交易管理示例
    txManager := transaction.NewTransactionManager(providerManager, walletManager)

    // 发送ETH交易
    recipient := common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96c4b4d8b6")
    amount := utils.NewUnitConverter().EtherToWei(decimal.NewFromFloat(0.001)) // 0.001 ETH

    txParams := transaction.TransactionParams{
        To:       recipient,
        Value:    amount,
        GasLimit: 21000,
        GasPrice: networkInfo.GasPrice,
    }

    fmt.Println("发送ETH交易...")
    txResult, err := txManager.SendTransaction(txParams)
    if err != nil {
        log.Printf("发送交易失败: %v", err)
    } else {
        fmt.Printf("交易已发送:\n")
        fmt.Printf("  哈希: %s\n", txResult.Hash.Hex())
        fmt.Printf("  Nonce: %d\n", txResult.Nonce)
        fmt.Printf("  Gas价格: %s Gwei\n", 
            utils.NewUnitConverter().WeiToGwei(txResult.GasPrice).String())

        // 等待交易确认
        fmt.Println("等待交易确认...")
        receipt, err := txManager.WaitForConfirmation(txResult.Hash, 1)
        if err != nil {
            log.Printf("等待确认失败: %v", err)
        } else {
            fmt.Printf("交易已确认:\n")
            fmt.Printf("  区块号: %d\n", receipt.BlockNumber.Uint64())
            fmt.Printf("  Gas使用: %d\n", receipt.GasUsed)
            fmt.Printf("  状态: %d\n", receipt.Status)
        }
    }

    // 事件监听示例
    eventManager := events.NewEventManager(providerManager)

    // 监听新区块
    fmt.Println("开始监听新区块...")
    err = eventManager.WatchNewBlocks(func(header *types.Header) {
        fmt.Printf("新区块: #%d, 哈希: %s, 交易数: %d\n", 
            header.Number.Uint64(), 
            header.Hash().Hex()[:10]+"...", 
            len(header.TxHash))
    })
    if err != nil {
        log.Printf("监听新区块失败: %v", err)
    }

    // 监听待处理交易
    fmt.Println("开始监听待处理交易...")
    err = eventManager.WatchPendingTransactions(func(txHash common.Hash) {
        fmt.Printf("新的待处理交易: %s\n", txHash.Hex()[:10]+"...")
    })
    if err != nil {
        log.Printf("监听待处理交易失败: %v", err)
    }

    // 工具函数示例
    unitConverter := utils.NewUnitConverter()
    hashUtils := utils.NewHashUtils()
    sigUtils := utils.NewSignatureUtils()
    encUtils := utils.NewEncodingUtils()

    // 单位转换示例
    weiAmount := big.NewInt(1000000000000000000) // 1 ETH in Wei
    etherAmount := unitConverter.WeiToEther(weiAmount)
    fmt.Printf("单位转换: %s Wei = %s ETH\n", weiAmount.String(), etherAmount.String())

    // 哈希示例
    message := "Hello, Web3!"
    messageHash := hashUtils.Keccak256String(message)
    fmt.Printf("消息哈希: %s -> %s\n", message, messageHash.Hex())

    // 签名示例
    signature, err := walletManager.SignPersonalMessage(message)
    if err != nil {
        log.Printf("签名失败: %v", err)
    } else {
        fmt.Printf("消息签名: %s\n", encUtils.ToHex(signature))

        // 验证签名
        prefixedMessage := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
        isValid := sigUtils.VerifySignature([]byte(prefixedMessage), signature, walletManager.GetAddress())
        fmt.Printf("签名验证: %t\n", isValid)
    }

    // 地址格式化示例
    formattedAddr := utils.FormatAddress(walletManager.GetAddress())
    fmt.Printf("格式化地址: %s\n", formattedAddr)

    // 获取最新区块信息
    latestBlock, err := providerManager.GetLatestBlock()
    if err != nil {
        log.Printf("获取最新区块失败: %v", err)
    } else {
        fmt.Printf("最新区块信息:\n")
        fmt.Printf("  区块号: %d\n", latestBlock.Number.Uint64())
        fmt.Printf("  时间戳: %d\n", latestBlock.Timestamp)
        fmt.Printf("  Gas限制: %d\n", latestBlock.GasLimit)
        fmt.Printf("  Gas使用: %d\n", latestBlock.GasUsed)
        fmt.Printf("  交易数: %d\n", latestBlock.Transactions)
    }

    // 运行一段时间以观察事件
    fmt.Println("监听事件中，按Ctrl+C退出...")
    time.Sleep(30 * time.Second)

    fmt.Println("Web3.js Go绑定演示完成!")
}
```
