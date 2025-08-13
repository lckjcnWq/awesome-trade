# Arbitrum SDK Go 详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [环境准备](#环境准备)
3. [L1到L2桥接](#L1到L2桥接)
4. [L2到L1提取](#L2到L1提取)
5. [代币桥接](#代币桥接)
6. [交易监控](#交易监控)
7. [Gas费估算](#Gas费估算)
8. [批量操作](#批量操作)
9. [实际应用](#实际应用)

## 基础概念

### 1.1 Arbitrum简介

Arbitrum是以太坊Layer 2扩容解决方案，使用Optimistic Rollup技术提供更快、更便宜的交易处理。

```bash
# 安装Arbitrum相关依赖
go get github.com/ethereum/go-ethereum
go get github.com/offchainlabs/arbitrum/packages/arb-util
go get github.com/shopspring/decimal
```

### 1.2 核心概念

```go
// 主要包导入
import (
    "context"
    "math/big"
    "time"
    
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
)

// 核心概念：
// - L1: 以太坊主网
// - L2: Arbitrum网络
// - Bridge: 跨层桥接合约
// - Retryable: 可重试交易
// - Outbox: L2到L1消息输出
// - Sequencer: 排序器节点
```

## 环境准备

### 2.1 Arbitrum配置

```go
// config/arbitrum.go
package config

import (
    "math/big"
    "time"
    
    "github.com/ethereum/go-ethereum/common"
)

type ArbitrumConfig struct {
    // 网络配置
    L1NetworkName    string
    L2NetworkName    string
    L1ChainID        *big.Int
    L2ChainID        *big.Int
    L1RPC            string
    L2RPC            string
    
    // 桥接合约地址
    L1GatewayRouter  common.Address
    L2GatewayRouter  common.Address
    L1EthGateway     common.Address
    L2EthGateway     common.Address
    Inbox            common.Address
    Outbox           common.Address
    
    // 交易配置
    L1GasLimit       uint64
    L2GasLimit       uint64
    L1GasPrice       *big.Int
    L2GasPrice       *big.Int
    
    // 超时配置
    Timeout          time.Duration
    ConfirmationBlocks uint64
    
    // 重试配置
    RetryDelay       time.Duration
    MaxRetries       int
}

func DefaultArbitrumConfig() *ArbitrumConfig {
    return &ArbitrumConfig{
        L1NetworkName: "ethereum",
        L2NetworkName: "arbitrum-one",
        L1ChainID:     big.NewInt(1),
        L2ChainID:     big.NewInt(42161),
        L1RPC:         "https://mainnet.infura.io/v3/YOUR_PROJECT_ID",
        L2RPC:         "https://arb1.arbitrum.io/rpc",
        
        // Arbitrum One合约地址
        L1GatewayRouter: common.HexToAddress("0x72Ce9c846789fdB6fC1f34aC4AD25Dd9ef7031ef"),
        L2GatewayRouter: common.HexToAddress("0x5288c571Fd7aD117beA99bF60FE0846C4E84F933"),
        L1EthGateway:    common.HexToAddress("0xd92023E9d9911199a6711321D1277285e6d4e2db"),
        L2EthGateway:    common.HexToAddress("0x09e9222E96E7B4AE2a407B98d48e330053351EEe"),
        Inbox:           common.HexToAddress("0x4Dbd4fc535Ac27206064B68FfCf827b0A60BAB3f"),
        Outbox:          common.HexToAddress("0x0B9857ae2D4A3DBe74ffE1d7DF045bb7F96E4840"),
        
        L1GasLimit:       500000,
        L2GasLimit:       2000000,
        L1GasPrice:       big.NewInt(20000000000), // 20 Gwei
        L2GasPrice:       big.NewInt(100000000),   // 0.1 Gwei
        
        Timeout:          60 * time.Second,
        ConfirmationBlocks: 12,
        RetryDelay:       5 * time.Second,
        MaxRetries:       3,
    }
}

// Arbitrum Goerli测试网配置
func ArbitrumGoerliConfig() *ArbitrumConfig {
    cfg := DefaultArbitrumConfig()
    cfg.L1NetworkName = "goerli"
    cfg.L2NetworkName = "arbitrum-goerli"
    cfg.L1ChainID = big.NewInt(5)
    cfg.L2ChainID = big.NewInt(421613)
    cfg.L1RPC = "https://goerli.infura.io/v3/YOUR_PROJECT_ID"
    cfg.L2RPC = "https://goerli-rollup.arbitrum.io/rpc"
    
    // 测试网合约地址
    cfg.L1GatewayRouter = common.HexToAddress("0x4c7708168395aEa569453Fc36862D2ffcDaC588c")
    cfg.L2GatewayRouter = common.HexToAddress("0xE5B9d8d42d656d1DcB8065A6c012FE3780246041")
    cfg.Inbox = common.HexToAddress("0x6BEbC4925716945D46F0Ec336D5C2564F419682C")
    
    return cfg
}
```

## L1到L2桥接

### 3.1 ETH桥接

```go
// bridge/eth_bridge.go
package bridge

import (
    "context"
    "fmt"
    "math/big"
    "strings"

    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/ethclient"
    
    "your-project/config"
    "your-project/wallet"
)

// Inbox ABI（简化版）
const InboxABI = `[
    {
        "inputs": [
            {"name": "to", "type": "address"},
            {"name": "l2CallValue", "type": "uint256"},
            {"name": "maxSubmissionCost", "type": "uint256"},
            {"name": "excessFeeRefundAddress", "type": "address"},
            {"name": "callValueRefundAddress", "type": "address"},
            {"name": "gasLimit", "type": "uint256"},
            {"name": "maxFeePerGas", "type": "uint256"},
            {"name": "data", "type": "bytes"}
        ],
        "name": "createRetryableTicket",
        "outputs": [{"name": "", "type": "uint256"}],
        "stateMutability": "payable",
        "type": "function"
    },
    {
        "inputs": [
            {"name": "refundTo", "type": "address"}
        ],
        "name": "depositEth",
        "outputs": [{"name": "", "type": "uint256"}],
        "stateMutability": "payable",
        "type": "function"
    }
]`

type ETHBridge struct {
    l1Client *ethclient.Client
    l2Client *ethclient.Client
    config   *config.ArbitrumConfig
    wallet   *wallet.WalletManager
    inbox    *bind.BoundContract
}

func NewETHBridge(cfg *config.ArbitrumConfig, wallet *wallet.WalletManager) (*ETHBridge, error) {
    // 连接L1客户端
    l1Client, err := ethclient.Dial(cfg.L1RPC)
    if err != nil {
        return nil, fmt.Errorf("连接L1节点失败: %v", err)
    }

    // 连接L2客户端
    l2Client, err := ethclient.Dial(cfg.L2RPC)
    if err != nil {
        return nil, fmt.Errorf("连接L2节点失败: %v", err)
    }

    // 解析Inbox ABI
    parsedABI, err := abi.JSON(strings.NewReader(InboxABI))
    if err != nil {
        return nil, fmt.Errorf("解析Inbox ABI失败: %v", err)
    }

    // 创建Inbox合约实例
    inbox := bind.NewBoundContract(cfg.Inbox, parsedABI, l1Client, l1Client, l1Client)

    return &ETHBridge{
        l1Client: l1Client,
        l2Client: l2Client,
        config:   cfg,
        wallet:   wallet,
        inbox:    inbox,
    }, nil
}

// 存入ETH到L2
func (eb *ETHBridge) DepositETH(amount *big.Int) (*DepositResult, error) {
    // 创建交易选项
    auth, err := eb.wallet.CreateTransactOpts(eb.config.L1ChainID)
    if err != nil {
        return nil, fmt.Errorf("创建交易选项失败: %v", err)
    }

    auth.Value = amount
    auth.GasLimit = eb.config.L1GasLimit
    auth.GasPrice = eb.config.L1GasPrice

    // 调用depositEth
    tx, err := eb.inbox.Transact(auth, "depositEth", eb.wallet.GetAddress())
    if err != nil {
        return nil, fmt.Errorf("存入ETH失败: %v", err)
    }

    return &DepositResult{
        L1TxHash: tx.Hash(),
        Amount:   amount,
        To:       eb.wallet.GetAddress(),
        Type:     "ETH_DEPOSIT",
    }, nil
}

// 创建可重试票据
func (eb *ETHBridge) CreateRetryableTicket(params RetryableParams) (*RetryableResult, error) {
    // 创建交易选项
    auth, err := eb.wallet.CreateTransactOpts(eb.config.L1ChainID)
    if err != nil {
        return nil, fmt.Errorf("创建交易选项失败: %v", err)
    }

    // 计算总费用
    totalValue := new(big.Int).Add(params.L2CallValue, params.MaxSubmissionCost)
    totalValue.Add(totalValue, new(big.Int).Mul(params.GasLimit, params.MaxFeePerGas))

    auth.Value = totalValue
    auth.GasLimit = eb.config.L1GasLimit
    auth.GasPrice = eb.config.L1GasPrice

    // 创建可重试票据
    tx, err := eb.inbox.Transact(
        auth,
        "createRetryableTicket",
        params.To,
        params.L2CallValue,
        params.MaxSubmissionCost,
        params.ExcessFeeRefundAddress,
        params.CallValueRefundAddress,
        params.GasLimit,
        params.MaxFeePerGas,
        params.Data,
    )
    if err != nil {
        return nil, fmt.Errorf("创建可重试票据失败: %v", err)
    }

    return &RetryableResult{
        L1TxHash: tx.Hash(),
        Params:   params,
    }, nil
}

// 估算L2 Gas费用
func (eb *ETHBridge) EstimateL2Gas(to common.Address, data []byte) (*GasEstimate, error) {
    // 获取L2 Gas价格
    gasPrice, err := eb.l2Client.SuggestGasPrice(context.Background())
    if err != nil {
        return nil, fmt.Errorf("获取L2 Gas价格失败: %v", err)
    }

    // 估算Gas限制
    msg := ethereum.CallMsg{
        From: eb.wallet.GetAddress(),
        To:   &to,
        Data: data,
    }

    gasLimit, err := eb.l2Client.EstimateGas(context.Background(), msg)
    if err != nil {
        return nil, fmt.Errorf("估算L2 Gas失败: %v", err)
    }

    // 计算最大提交成本
    maxSubmissionCost := eb.calculateMaxSubmissionCost(len(data))

    return &GasEstimate{
        GasLimit:          big.NewInt(int64(gasLimit)),
        GasPrice:          gasPrice,
        MaxSubmissionCost: maxSubmissionCost,
        TotalCost:         new(big.Int).Add(new(big.Int).Mul(big.NewInt(int64(gasLimit)), gasPrice), maxSubmissionCost),
    }, nil
}

// 计算最大提交成本
func (eb *ETHBridge) calculateMaxSubmissionCost(dataLength int) *big.Int {
    // 基础成本 + 数据成本
    baseCost := big.NewInt(40000)
    dataCost := big.NewInt(int64(dataLength * 16))
    return new(big.Int).Add(baseCost, dataCost)
}

// 监控存款状态
func (eb *ETHBridge) WaitForL2Transaction(l1TxHash common.Hash) (*L2TransactionResult, error) {
    // 等待L1交易确认
    receipt, err := eb.waitForL1Confirmation(l1TxHash)
    if err != nil {
        return nil, err
    }

    // 从L1交易日志中提取L2交易信息
    l2TxHash, err := eb.extractL2TxHash(receipt)
    if err != nil {
        return nil, err
    }

    // 等待L2交易确认
    l2Receipt, err := eb.waitForL2Confirmation(l2TxHash)
    if err != nil {
        return nil, err
    }

    return &L2TransactionResult{
        L1TxHash:    l1TxHash,
        L2TxHash:    l2TxHash,
        L1Receipt:   receipt,
        L2Receipt:   l2Receipt,
        Success:     l2Receipt.Status == 1,
    }, nil
}

// 等待L1确认
func (eb *ETHBridge) waitForL1Confirmation(txHash common.Hash) (*types.Receipt, error) {
    for i := 0; i < eb.config.MaxRetries; i++ {
        receipt, err := eb.l1Client.TransactionReceipt(context.Background(), txHash)
        if err == nil {
            return receipt, nil
        }
        time.Sleep(eb.config.RetryDelay)
    }
    return nil, fmt.Errorf("等待L1交易确认超时")
}

// 等待L2确认
func (eb *ETHBridge) waitForL2Confirmation(txHash common.Hash) (*types.Receipt, error) {
    for i := 0; i < eb.config.MaxRetries; i++ {
        receipt, err := eb.l2Client.TransactionReceipt(context.Background(), txHash)
        if err == nil {
            return receipt, nil
        }
        time.Sleep(eb.config.RetryDelay)
    }
    return nil, fmt.Errorf("等待L2交易确认超时")
}

// 从L1交易日志提取L2交易哈希
func (eb *ETHBridge) extractL2TxHash(receipt *types.Receipt) (common.Hash, error) {
    // 这里需要解析InboxMessageDelivered事件
    // 简化实现，返回模拟哈希
    return common.HexToHash("0x1234567890abcdef"), nil
}

type RetryableParams struct {
    To                      common.Address
    L2CallValue             *big.Int
    MaxSubmissionCost       *big.Int
    ExcessFeeRefundAddress  common.Address
    CallValueRefundAddress  common.Address
    GasLimit                *big.Int
    MaxFeePerGas            *big.Int
    Data                    []byte
}

type DepositResult struct {
    L1TxHash common.Hash
    Amount   *big.Int
    To       common.Address
    Type     string
}

type RetryableResult struct {
    L1TxHash common.Hash
    Params   RetryableParams
}

type GasEstimate struct {
    GasLimit          *big.Int
    GasPrice          *big.Int
    MaxSubmissionCost *big.Int
    TotalCost         *big.Int
}

type L2TransactionResult struct {
    L1TxHash  common.Hash
    L2TxHash  common.Hash
    L1Receipt *types.Receipt
    L2Receipt *types.Receipt
    Success   bool
}
```

## L2到L1提取

### 4.1 提取管理器

```go
// withdraw/manager.go
package withdraw

import (
    "context"
    "fmt"
    "math/big"
    "strings"
    "time"

    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/ethclient"
    
    "your-project/config"
    "your-project/wallet"
)

// ArbSys ABI（简化版）
const ArbSysABI = `[
    {
        "inputs": [
            {"name": "destination", "type": "address"},
            {"name": "data", "type": "bytes"}
        ],
        "name": "sendTxToL1",
        "outputs": [{"name": "", "type": "uint256"}],
        "stateMutability": "payable",
        "type": "function"
    }
]`

// Outbox ABI（简化版）
const OutboxABI = `[
    {
        "inputs": [
            {"name": "proof", "type": "bytes32[]"},
            {"name": "index", "type": "uint256"},
            {"name": "l2Sender", "type": "address"},
            {"name": "to", "type": "address"},
            {"name": "l2Block", "type": "uint256"},
            {"name": "l1Block", "type": "uint256"},
            {"name": "l2Timestamp", "type": "uint256"},
            {"name": "value", "type": "uint256"},
            {"name": "data", "type": "bytes"}
        ],
        "name": "executeTransaction",
        "outputs": [],
        "stateMutability": "nonpayable",
        "type": "function"
    }
]`

type WithdrawManager struct {
    l1Client *ethclient.Client
    l2Client *ethclient.Client
    config   *config.ArbitrumConfig
    wallet   *wallet.WalletManager
    arbSys   *bind.BoundContract
    outbox   *bind.BoundContract
}

func NewWithdrawManager(cfg *config.ArbitrumConfig, wallet *wallet.WalletManager) (*WithdrawManager, error) {
    // 连接客户端
    l1Client, err := ethclient.Dial(cfg.L1RPC)
    if err != nil {
        return nil, fmt.Errorf("连接L1节点失败: %v", err)
    }

    l2Client, err := ethclient.Dial(cfg.L2RPC)
    if err != nil {
        return nil, fmt.Errorf("连接L2节点失败: %v", err)
    }

    // 解析ArbSys ABI
    arbSysABI, err := abi.JSON(strings.NewReader(ArbSysABI))
    if err != nil {
        return nil, fmt.Errorf("解析ArbSys ABI失败: %v", err)
    }

    // 解析Outbox ABI
    outboxABI, err := abi.JSON(strings.NewReader(OutboxABI))
    if err != nil {
        return nil, fmt.Errorf("解析Outbox ABI失败: %v", err)
    }

    // ArbSys预编译合约地址
    arbSysAddr := common.HexToAddress("0x0000000000000000000000000000000000000064")
    arbSys := bind.NewBoundContract(arbSysAddr, arbSysABI, l2Client, l2Client, l2Client)

    // Outbox合约
    outbox := bind.NewBoundContract(cfg.Outbox, outboxABI, l1Client, l1Client, l1Client)

    return &WithdrawManager{
        l1Client: l1Client,
        l2Client: l2Client,
        config:   cfg,
        wallet:   wallet,
        arbSys:   arbSys,
        outbox:   outbox,
    }, nil
}

// 发起提取
func (wm *WithdrawManager) InitiateWithdraw(to common.Address, amount *big.Int, data []byte) (*WithdrawInitiation, error) {
    // 创建L2交易选项
    auth, err := wm.wallet.CreateTransactOpts(wm.config.L2ChainID)
    if err != nil {
        return nil, fmt.Errorf("创建交易选项失败: %v", err)
    }

    auth.Value = amount
    auth.GasLimit = wm.config.L2GasLimit
    auth.GasPrice = wm.config.L2GasPrice

    // 调用ArbSys.sendTxToL1
    tx, err := wm.arbSys.Transact(auth, "sendTxToL1", to, data)
    if err != nil {
        return nil, fmt.Errorf("发起提取失败: %v", err)
    }

    return &WithdrawInitiation{
        L2TxHash: tx.Hash(),
        To:       to,
        Amount:   amount,
        Data:     data,
    }, nil
}

// 执行提取
func (wm *WithdrawManager) ExecuteWithdraw(proof WithdrawProof) (*WithdrawExecution, error) {
    // 创建L1交易选项
    auth, err := wm.wallet.CreateTransactOpts(wm.config.L1ChainID)
    if err != nil {
        return nil, fmt.Errorf("创建交易选项失败: %v", err)
    }

    auth.GasLimit = wm.config.L1GasLimit
    auth.GasPrice = wm.config.L1GasPrice

    // 执行提取交易
    tx, err := wm.outbox.Transact(
        auth,
        "executeTransaction",
        proof.MerkleProof,
        proof.Index,
        proof.L2Sender,
        proof.To,
        proof.L2Block,
        proof.L1Block,
        proof.L2Timestamp,
        proof.Value,
        proof.Data,
    )
    if err != nil {
        return nil, fmt.Errorf("执行提取失败: %v", err)
    }

    return &WithdrawExecution{
        L1TxHash: tx.Hash(),
        Proof:    proof,
    }, nil
}

// 获取提取证明
func (wm *WithdrawManager) GetWithdrawProof(l2TxHash common.Hash) (*WithdrawProof, error) {
    // 这里需要调用Arbitrum节点API获取Merkle证明
    // 简化实现，返回模拟数据
    
    return &WithdrawProof{
        MerkleProof: [][32]byte{},
        Index:       big.NewInt(0),
        L2Sender:    wm.wallet.GetAddress(),
        To:          wm.wallet.GetAddress(),
        L2Block:     big.NewInt(0),
        L1Block:     big.NewInt(0),
        L2Timestamp: big.NewInt(time.Now().Unix()),
        Value:       big.NewInt(0),
        Data:        []byte{},
    }, nil
}

// 检查提取状态
func (wm *WithdrawManager) CheckWithdrawStatus(l2TxHash common.Hash) (*WithdrawStatus, error) {
    // 获取L2交易收据
    l2Receipt, err := wm.l2Client.TransactionReceipt(context.Background(), l2TxHash)
    if err != nil {
        return &WithdrawStatus{
            L2TxHash: l2TxHash,
            Status:   "pending",
        }, nil
    }

    status := &WithdrawStatus{
        L2TxHash:    l2TxHash,
        L2Receipt:   l2Receipt,
        Status:      "confirmed_on_l2",
        L2BlockNum:  l2Receipt.BlockNumber.Uint64(),
    }

    // 检查是否可以执行提取
    canExecute, err := wm.canExecuteWithdraw(l2Receipt)
    if err != nil {
        return status, err
    }

    if canExecute {
        status.Status = "ready_for_execution"
    } else {
        status.Status = "waiting_for_challenge_period"
    }

    return status, nil
}

// 检查是否可以执行提取
func (wm *WithdrawManager) canExecuteWithdraw(l2Receipt *types.Receipt) (bool, error) {
    // 检查挑战期是否已过
    // Arbitrum的挑战期通常是7天
    challengePeriod := 7 * 24 * time.Hour
    
    // 获取L2区块时间
    l2Block, err := wm.l2Client.BlockByNumber(context.Background(), l2Receipt.BlockNumber)
    if err != nil {
        return false, err
    }

    blockTime := time.Unix(int64(l2Block.Time()), 0)
    return time.Since(blockTime) > challengePeriod, nil
}

type WithdrawInitiation struct {
    L2TxHash common.Hash
    To       common.Address
    Amount   *big.Int
    Data     []byte
}

type WithdrawProof struct {
    MerkleProof [][32]byte
    Index       *big.Int
    L2Sender    common.Address
    To          common.Address
    L2Block     *big.Int
    L1Block     *big.Int
    L2Timestamp *big.Int
    Value       *big.Int
    Data        []byte
}

type WithdrawExecution struct {
    L1TxHash common.Hash
    Proof    WithdrawProof
}

type WithdrawStatus struct {
    L2TxHash   common.Hash
    L2Receipt  *types.Receipt
    Status     string
    L2BlockNum uint64
}
```

## 代币桥接

### 5.1 ERC20桥接

```go
// token/bridge.go
package token

import (
    "context"
    "fmt"
    "math/big"
    "strings"

    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/ethclient"
    
    "your-project/config"
    "your-project/wallet"
)

// L1 Gateway Router ABI（简化版）
const L1GatewayRouterABI = `[
    {
        "inputs": [
            {"name": "_token", "type": "address"},
            {"name": "_to", "type": "address"},
            {"name": "_amount", "type": "uint256"},
            {"name": "_maxGas", "type": "uint256"},
            {"name": "_gasPriceBid", "type": "uint256"},
            {"name": "_data", "type": "bytes"}
        ],
        "name": "outboundTransfer",
        "outputs": [{"name": "", "type": "bytes"}],
        "stateMutability": "payable",
        "type": "function"
    },
    {
        "inputs": [
            {"name": "l1Token", "type": "address"}
        ],
        "name": "getGateway",
        "outputs": [{"name": "gateway", "type": "address"}],
        "stateMutability": "view",
        "type": "function"
    }
]`

// ERC20 ABI（简化版）
const ERC20ABI = `[
    {
        "inputs": [
            {"name": "spender", "type": "address"},
            {"name": "amount", "type": "uint256"}
        ],
        "name": "approve",
        "outputs": [{"name": "", "type": "bool"}],
        "stateMutability": "nonpayable",
        "type": "function"
    },
    {
        "inputs": [
            {"name": "account", "type": "address"}
        ],
        "name": "balanceOf",
        "outputs": [{"name": "", "type": "uint256"}],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [
            {"name": "owner", "type": "address"},
            {"name": "spender", "type": "address"}
        ],
        "name": "allowance",
        "outputs": [{"name": "", "type": "uint256"}],
        "stateMutability": "view",
        "type": "function"
    }
]`

type TokenBridge struct {
    l1Client        *ethclient.Client
    l2Client        *ethclient.Client
    config          *config.ArbitrumConfig
    wallet          *wallet.WalletManager
    l1GatewayRouter *bind.BoundContract
    l2GatewayRouter *bind.BoundContract
}

func NewTokenBridge(cfg *config.ArbitrumConfig, wallet *wallet.WalletManager) (*TokenBridge, error) {
    // 连接客户端
    l1Client, err := ethclient.Dial(cfg.L1RPC)
    if err != nil {
        return nil, fmt.Errorf("连接L1节点失败: %v", err)
    }

    l2Client, err := ethclient.Dial(cfg.L2RPC)
    if err != nil {
        return nil, fmt.Errorf("连接L2节点失败: %v", err)
    }

    // 解析Gateway Router ABI
    gatewayABI, err := abi.JSON(strings.NewReader(L1GatewayRouterABI))
    if err != nil {
        return nil, fmt.Errorf("解析Gateway ABI失败: %v", err)
    }

    // 创建Gateway Router合约实例
    l1GatewayRouter := bind.NewBoundContract(cfg.L1GatewayRouter, gatewayABI, l1Client, l1Client, l1Client)
    l2GatewayRouter := bind.NewBoundContract(cfg.L2GatewayRouter, gatewayABI, l2Client, l2Client, l2Client)

    return &TokenBridge{
        l1Client:        l1Client,
        l2Client:        l2Client,
        config:          cfg,
        wallet:          wallet,
        l1GatewayRouter: l1GatewayRouter,
        l2GatewayRouter: l2GatewayRouter,
    }, nil
}

// 存入ERC20代币到L2
func (tb *TokenBridge) DepositToken(params TokenDepositParams) (*TokenDepositResult, error) {
    // 检查并批准代币
    err := tb.ensureTokenApproval(params.L1Token, params.Amount)
    if err != nil {
        return nil, fmt.Errorf("代币批准失败: %v", err)
    }

    // 估算L2 Gas费用
    gasEstimate, err := tb.estimateL2TokenGas(params)
    if err != nil {
        return nil, fmt.Errorf("估算Gas费用失败: %v", err)
    }

    // 创建交易选项
    auth, err := tb.wallet.CreateTransactOpts(tb.config.L1ChainID)
    if err != nil {
        return nil, fmt.Errorf("创建交易选项失败: %v", err)
    }

    auth.Value = gasEstimate.TotalCost
    auth.GasLimit = tb.config.L1GasLimit
    auth.GasPrice = tb.config.L1GasPrice

    // 调用outboundTransfer
    tx, err := tb.l1GatewayRouter.Transact(
        auth,
        "outboundTransfer",
        params.L1Token,
        params.To,
        params.Amount,
        gasEstimate.GasLimit,
        gasEstimate.GasPrice,
        params.Data,
    )
    if err != nil {
        return nil, fmt.Errorf("存入代币失败: %v", err)
    }

    return &TokenDepositResult{
        L1TxHash:    tx.Hash(),
        L1Token:     params.L1Token,
        Amount:      params.Amount,
        To:          params.To,
        GasEstimate: gasEstimate,
    }, nil
}

// 确保代币批准
func (tb *TokenBridge) ensureTokenApproval(tokenAddr common.Address, amount *big.Int) error {
    // 解析ERC20 ABI
    erc20ABI, err := abi.JSON(strings.NewReader(ERC20ABI))
    if err != nil {
        return fmt.Errorf("解析ERC20 ABI失败: %v", err)
    }

    // 创建ERC20合约实例
    token := bind.NewBoundContract(tokenAddr, erc20ABI, tb.l1Client, tb.l1Client, tb.l1Client)

    // 检查当前授权额度
    var allowanceResult []interface{}
    err = token.Call(nil, &allowanceResult, "allowance", tb.wallet.GetAddress(), tb.config.L1GatewayRouter)
    if err != nil {
        return fmt.Errorf("检查授权额度失败: %v", err)
    }

    currentAllowance := allowanceResult[0].(*big.Int)

    // 如果授权额度不足，进行批准
    if currentAllowance.Cmp(amount) < 0 {
        auth, err := tb.wallet.CreateTransactOpts(tb.config.L1ChainID)
        if err != nil {
            return fmt.Errorf("创建交易选项失败: %v", err)
        }

        auth.GasLimit = 100000
        auth.GasPrice = tb.config.L1GasPrice

        // 批准代币
        _, err = token.Transact(auth, "approve", tb.config.L1GatewayRouter, amount)
        if err != nil {
            return fmt.Errorf("批准代币失败: %v", err)
        }
    }

    return nil
}

// 估算L2代币Gas费用
func (tb *TokenBridge) estimateL2TokenGas(params TokenDepositParams) (*GasEstimate, error) {
    // 获取L2 Gas价格
    gasPrice, err := tb.l2Client.SuggestGasPrice(context.Background())
    if err != nil {
        return nil, fmt.Errorf("获取L2 Gas价格失败: %v", err)
    }

    // 代币转账的Gas限制
    gasLimit := big.NewInt(200000)

    // 计算最大提交成本
    maxSubmissionCost := big.NewInt(50000)

    return &GasEstimate{
        GasLimit:          gasLimit,
        GasPrice:          gasPrice,
        MaxSubmissionCost: maxSubmissionCost,
        TotalCost:         new(big.Int).Add(new(big.Int).Mul(gasLimit, gasPrice), maxSubmissionCost),
    }, nil
}

// 获取代币余额
func (tb *TokenBridge) GetTokenBalance(tokenAddr common.Address, account common.Address, isL2 bool) (*big.Int, error) {
    var client *ethclient.Client
    if isL2 {
        client = tb.l2Client
    } else {
        client = tb.l1Client
    }

    // 解析ERC20 ABI
    erc20ABI, err := abi.JSON(strings.NewReader(ERC20ABI))
    if err != nil {
        return nil, fmt.Errorf("解析ERC20 ABI失败: %v", err)
    }

    // 创建ERC20合约实例
    token := bind.NewBoundContract(tokenAddr, erc20ABI, client, client, client)

    // 获取余额
    var balanceResult []interface{}
    err = token.Call(nil, &balanceResult, "balanceOf", account)
    if err != nil {
        return nil, fmt.Errorf("获取代币余额失败: %v", err)
    }

    return balanceResult[0].(*big.Int), nil
}

// 获取代币网关
func (tb *TokenBridge) GetTokenGateway(l1Token common.Address) (common.Address, error) {
    var gatewayResult []interface{}
    err := tb.l1GatewayRouter.Call(nil, &gatewayResult, "getGateway", l1Token)
    if err != nil {
        return common.Address{}, fmt.Errorf("获取代币网关失败: %v", err)
    }

    return gatewayResult[0].(common.Address), nil
}

type TokenDepositParams struct {
    L1Token common.Address
    To      common.Address
    Amount  *big.Int
    Data    []byte
}

type TokenDepositResult struct {
    L1TxHash    common.Hash
    L1Token     common.Address
    Amount      *big.Int
    To          common.Address
    GasEstimate *GasEstimate
}
```

## 实际应用

### 9.1 完整Arbitrum应用

```go
// main.go
package main

import (
    "fmt"
    "log"
    "math/big"

    "github.com/ethereum/go-ethereum/common"

    "your-project/config"
    "your-project/bridge"
    "your-project/withdraw"
    "your-project/token"
    "your-project/wallet"
)

func main() {
    // 创建Arbitrum配置
    cfg := config.ArbitrumGoerliConfig() // 使用测试网

    // 创建钱包
    walletManager, err := wallet.NewWalletManager("your_private_key_here")
    if err != nil {
        log.Fatal("创建钱包失败:", err)
    }

    fmt.Printf("钱包地址: %s\n", walletManager.GetAddress().Hex())

    // ETH桥接示例
    fmt.Println("=== ETH桥接示例 ===")
    
    ethBridge, err := bridge.NewETHBridge(cfg, walletManager)
    if err != nil {
        log.Fatal("创建ETH桥接失败:", err)
    }

    // 存入ETH到L2
    depositAmount := big.NewInt(100000000000000000) // 0.1 ETH
    depositResult, err := ethBridge.DepositETH(depositAmount)
    if err != nil {
        log.Printf("存入ETH失败: %v", err)
    } else {
        fmt.Printf("ETH存入成功:\n")
        fmt.Printf("  L1交易哈希: %s\n", depositResult.L1TxHash.Hex())
        fmt.Printf("  金额: %s Wei\n", depositResult.Amount.String())
        fmt.Printf("  接收地址: %s\n", depositResult.To.Hex())

        // 等待L2交易确认
        l2Result, err := ethBridge.WaitForL2Transaction(depositResult.L1TxHash)
        if err != nil {
            log.Printf("等待L2交易失败: %v", err)
        } else {
            fmt.Printf("L2交易确认:\n")
            fmt.Printf("  L2交易哈希: %s\n", l2Result.L2TxHash.Hex())
            fmt.Printf("  成功: %t\n", l2Result.Success)
        }
    }

    // 代币桥接示例
    fmt.Println("\n=== 代币桥接示例 ===")
    
    tokenBridge, err := token.NewTokenBridge(cfg, walletManager)
    if err != nil {
        log.Fatal("创建代币桥接失败:", err)
    }

    // 假设的测试代币地址
    testTokenAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")
    
    // 获取L1代币余额
    l1Balance, err := tokenBridge.GetTokenBalance(testTokenAddr, walletManager.GetAddress(), false)
    if err != nil {
        log.Printf("获取L1代币余额失败: %v", err)
    } else {
        fmt.Printf("L1代币余额: %s\n", l1Balance.String())
    }

    // 存入代币到L2
    tokenDepositParams := token.TokenDepositParams{
        L1Token: testTokenAddr,
        To:      walletManager.GetAddress(),
        Amount:  big.NewInt(1000000000000000000), // 1 token
        Data:    []byte{},
    }

    tokenDepositResult, err := tokenBridge.DepositToken(tokenDepositParams)
    if err != nil {
        log.Printf("存入代币失败: %v", err)
    } else {
        fmt.Printf("代币存入成功:\n")
        fmt.Printf("  L1交易哈希: %s\n", tokenDepositResult.L1TxHash.Hex())
        fmt.Printf("  代币地址: %s\n", tokenDepositResult.L1Token.Hex())
        fmt.Printf("  金额: %s\n", tokenDepositResult.Amount.String())
        fmt.Printf("  Gas费用: %s Wei\n", tokenDepositResult.GasEstimate.TotalCost.String())
    }

    // 提取管理示例
    fmt.Println("\n=== 提取管理示例 ===")
    
    withdrawManager, err := withdraw.NewWithdrawManager(cfg, walletManager)
    if err != nil {
        log.Fatal("创建提取管理器失败:", err)
    }

    // 发起ETH提取
    withdrawAmount := big.NewInt(50000000000000000) // 0.05 ETH
    withdrawInit, err := withdrawManager.InitiateWithdraw(
        walletManager.GetAddress(),
        withdrawAmount,
        []byte{},
    )
    if err != nil {
        log.Printf("发起提取失败: %v", err)
    } else {
        fmt.Printf("提取发起成功:\n")
        fmt.Printf("  L2交易哈希: %s\n", withdrawInit.L2TxHash.Hex())
        fmt.Printf("  提取金额: %s Wei\n", withdrawInit.Amount.String())
        fmt.Printf("  目标地址: %s\n", withdrawInit.To.Hex())

        // 检查提取状态
        withdrawStatus, err := withdrawManager.CheckWithdrawStatus(withdrawInit.L2TxHash)
        if err != nil {
            log.Printf("检查提取状态失败: %v", err)
        } else {
            fmt.Printf("提取状态: %s\n", withdrawStatus.Status)
            if withdrawStatus.L2Receipt != nil {
                fmt.Printf("L2区块号: %d\n", withdrawStatus.L2BlockNum)
            }
        }
    }

    // Gas费估算示例
    fmt.Println("\n=== Gas费估算示例 ===")
    
    // 估算L2 Gas费用
    gasEstimate, err := ethBridge.EstimateL2Gas(walletManager.GetAddress(), []byte{})
    if err != nil {
        log.Printf("估算Gas费用失败: %v", err)
    } else {
        fmt.Printf("L2 Gas估算:\n")
        fmt.Printf("  Gas限制: %s\n", gasEstimate.GasLimit.String())
        fmt.Printf("  Gas价格: %s Wei\n", gasEstimate.GasPrice.String())
        fmt.Printf("  最大提交成本: %s Wei\n", gasEstimate.MaxSubmissionCost.String())
        fmt.Printf("  总费用: %s Wei\n", gasEstimate.TotalCost.String())
    }

    // 网关查询示例
    fmt.Println("\n=== 网关查询示例 ===")
    
    gateway, err := tokenBridge.GetTokenGateway(testTokenAddr)
    if err != nil {
        log.Printf("获取代币网关失败: %v", err)
    } else {
        fmt.Printf("代币 %s 的网关地址: %s\n", testTokenAddr.Hex(), gateway.Hex())
    }

    fmt.Println("Arbitrum操作演示完成!")
}
```
