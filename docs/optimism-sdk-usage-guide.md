# Optimism SDK Go 详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [环境准备](#环境准备)
3. [L1到L2存款](#L1到L2存款)
4. [L2到L1提取](#L2到L1提取)
5. [代币桥接](#代币桥接)
6. [消息传递](#消息传递)
7. [状态根验证](#状态根验证)
8. [批量操作](#批量操作)
9. [实际应用](#实际应用)

## 基础概念

### 1.1 Optimism简介

Optimism是以太坊Layer 2扩容解决方案，使用Optimistic Rollup技术提供快速、低成本的交易处理。

```bash
# 安装Optimism相关依赖
go get github.com/ethereum/go-ethereum
go get github.com/ethereum-optimism/optimism/l2geth
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
// - L2: Optimism网络
// - Portal: 跨层门户合约
// - Messenger: 消息传递器
// - StateRoot: 状态根
// - Withdrawal: 提取过程
// - Finalization: 最终确认
```

## 环境准备

### 2.1 Optimism配置

```go
// config/optimism.go
package config

import (
    "math/big"
    "time"
    
    "github.com/ethereum/go-ethereum/common"
)

type OptimismConfig struct {
    // 网络配置
    L1NetworkName    string
    L2NetworkName    string
    L1ChainID        *big.Int
    L2ChainID        *big.Int
    L1RPC            string
    L2RPC            string
    
    // 核心合约地址
    L1CrossDomainMessenger common.Address
    L2CrossDomainMessenger common.Address
    L1StandardBridge       common.Address
    L2StandardBridge       common.Address
    OptimismPortal         common.Address
    L2OutputOracle         common.Address
    
    // 交易配置
    L1GasLimit       uint64
    L2GasLimit       uint64
    L1GasPrice       *big.Int
    L2GasPrice       *big.Int
    
    // 时间配置
    Timeout              time.Duration
    FinalizationPeriod   time.Duration
    ProofWindow          time.Duration
    
    // 重试配置
    RetryDelay       time.Duration
    MaxRetries       int
}

func DefaultOptimismConfig() *OptimismConfig {
    return &OptimismConfig{
        L1NetworkName: "ethereum",
        L2NetworkName: "optimism",
        L1ChainID:     big.NewInt(1),
        L2ChainID:     big.NewInt(10),
        L1RPC:         "https://mainnet.infura.io/v3/YOUR_PROJECT_ID",
        L2RPC:         "https://mainnet.optimism.io",
        
        // Optimism主网合约地址
        L1CrossDomainMessenger: common.HexToAddress("0x25ace71c97B33Cc4729CF772ae268934F7ab5fA1"),
        L2CrossDomainMessenger: common.HexToAddress("0x4200000000000000000000000000000000000007"),
        L1StandardBridge:       common.HexToAddress("0x99C9fc46f92E8a1c0deC1b1747d010903E884bE1"),
        L2StandardBridge:       common.HexToAddress("0x4200000000000000000000000000000000000010"),
        OptimismPortal:         common.HexToAddress("0xbEb5Fc579115071764c7423A4f12eDde41f106Ed"),
        L2OutputOracle:         common.HexToAddress("0xdfe97868233d1aa22e815a266982f2cf17685a27"),
        
        L1GasLimit:       500000,
        L2GasLimit:       2000000,
        L1GasPrice:       big.NewInt(20000000000), // 20 Gwei
        L2GasPrice:       big.NewInt(1000000),     // 0.001 Gwei
        
        Timeout:            60 * time.Second,
        FinalizationPeriod: 7 * 24 * time.Hour, // 7天
        ProofWindow:        7 * 24 * time.Hour,
        RetryDelay:         5 * time.Second,
        MaxRetries:         3,
    }
}

// Optimism Goerli测试网配置
func OptimismGoerliConfig() *OptimismConfig {
    cfg := DefaultOptimismConfig()
    cfg.L1NetworkName = "goerli"
    cfg.L2NetworkName = "optimism-goerli"
    cfg.L1ChainID = big.NewInt(5)
    cfg.L2ChainID = big.NewInt(420)
    cfg.L1RPC = "https://goerli.infura.io/v3/YOUR_PROJECT_ID"
    cfg.L2RPC = "https://goerli.optimism.io"
    
    // 测试网合约地址
    cfg.L1CrossDomainMessenger = common.HexToAddress("0x5086d1eEF304eb5284A0f6720f79403b4e9bE294")
    cfg.L2CrossDomainMessenger = common.HexToAddress("0x4200000000000000000000000000000000000007")
    cfg.L1StandardBridge = common.HexToAddress("0x636Af16bf2f682dD3109e60102b8E1A089FedAa8")
    cfg.L2StandardBridge = common.HexToAddress("0x4200000000000000000000000000000000000010")
    cfg.OptimismPortal = common.HexToAddress("0x5b47E1A08Ea6d985D6649300584e6722Ec4B1383")
    cfg.L2OutputOracle = common.HexToAddress("0xE6Dfba0953616Bacab0c9A8ecb3a9BBa77FC15c0")
    
    return cfg
}
```

## L1到L2存款

### 3.1 存款管理器

```go
// deposit/manager.go
package deposit

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

// OptimismPortal ABI（简化版）
const OptimismPortalABI = `[
    {
        "inputs": [
            {"name": "_to", "type": "address"},
            {"name": "_value", "type": "uint256"},
            {"name": "_gasLimit", "type": "uint64"},
            {"name": "_isCreation", "type": "bool"},
            {"name": "_data", "type": "bytes"}
        ],
        "name": "depositTransaction",
        "outputs": [],
        "stateMutability": "payable",
        "type": "function"
    }
]`

// L1StandardBridge ABI（简化版）
const L1StandardBridgeABI = `[
    {
        "inputs": [
            {"name": "_l2Token", "type": "address"},
            {"name": "_amount", "type": "uint256"},
            {"name": "_l2Gas", "type": "uint32"},
            {"name": "_data", "type": "bytes"}
        ],
        "name": "depositERC20",
        "outputs": [],
        "stateMutability": "nonpayable",
        "type": "function"
    },
    {
        "inputs": [
            {"name": "_l2Gas", "type": "uint32"},
            {"name": "_data", "type": "bytes"}
        ],
        "name": "depositETH",
        "outputs": [],
        "stateMutability": "payable",
        "type": "function"
    }
]`

type DepositManager struct {
    l1Client       *ethclient.Client
    l2Client       *ethclient.Client
    config         *config.OptimismConfig
    wallet         *wallet.WalletManager
    portal         *bind.BoundContract
    l1Bridge       *bind.BoundContract
}

func NewDepositManager(cfg *config.OptimismConfig, wallet *wallet.WalletManager) (*DepositManager, error) {
    // 连接客户端
    l1Client, err := ethclient.Dial(cfg.L1RPC)
    if err != nil {
        return nil, fmt.Errorf("连接L1节点失败: %v", err)
    }

    l2Client, err := ethclient.Dial(cfg.L2RPC)
    if err != nil {
        return nil, fmt.Errorf("连接L2节点失败: %v", err)
    }

    // 解析Portal ABI
    portalABI, err := abi.JSON(strings.NewReader(OptimismPortalABI))
    if err != nil {
        return nil, fmt.Errorf("解析Portal ABI失败: %v", err)
    }

    // 解析Bridge ABI
    bridgeABI, err := abi.JSON(strings.NewReader(L1StandardBridgeABI))
    if err != nil {
        return nil, fmt.Errorf("解析Bridge ABI失败: %v", err)
    }

    // 创建合约实例
    portal := bind.NewBoundContract(cfg.OptimismPortal, portalABI, l1Client, l1Client, l1Client)
    l1Bridge := bind.NewBoundContract(cfg.L1StandardBridge, bridgeABI, l1Client, l1Client, l1Client)

    return &DepositManager{
        l1Client: l1Client,
        l2Client: l2Client,
        config:   cfg,
        wallet:   wallet,
        portal:   portal,
        l1Bridge: l1Bridge,
    }, nil
}

// 存入ETH
func (dm *DepositManager) DepositETH(amount *big.Int, l2Gas uint32) (*DepositResult, error) {
    // 创建交易选项
    auth, err := dm.wallet.CreateTransactOpts(dm.config.L1ChainID)
    if err != nil {
        return nil, fmt.Errorf("创建交易选项失败: %v", err)
    }

    auth.Value = amount
    auth.GasLimit = dm.config.L1GasLimit
    auth.GasPrice = dm.config.L1GasPrice

    // 调用depositETH
    tx, err := dm.l1Bridge.Transact(auth, "depositETH", l2Gas, []byte{})
    if err != nil {
        return nil, fmt.Errorf("存入ETH失败: %v", err)
    }

    return &DepositResult{
        L1TxHash: tx.Hash(),
        Amount:   amount,
        To:       dm.wallet.GetAddress(),
        Type:     "ETH",
        L2Gas:    l2Gas,
    }, nil
}

// 存入ERC20代币
func (dm *DepositManager) DepositERC20(l1Token, l2Token common.Address, amount *big.Int, l2Gas uint32) (*DepositResult, error) {
    // 首先批准代币
    err := dm.approveToken(l1Token, amount)
    if err != nil {
        return nil, fmt.Errorf("批准代币失败: %v", err)
    }

    // 创建交易选项
    auth, err := dm.wallet.CreateTransactOpts(dm.config.L1ChainID)
    if err != nil {
        return nil, fmt.Errorf("创建交易选项失败: %v", err)
    }

    auth.GasLimit = dm.config.L1GasLimit
    auth.GasPrice = dm.config.L1GasPrice

    // 调用depositERC20
    tx, err := dm.l1Bridge.Transact(auth, "depositERC20", l2Token, amount, l2Gas, []byte{})
    if err != nil {
        return nil, fmt.Errorf("存入ERC20失败: %v", err)
    }

    return &DepositResult{
        L1TxHash: tx.Hash(),
        Amount:   amount,
        To:       dm.wallet.GetAddress(),
        Type:     "ERC20",
        L1Token:  l1Token,
        L2Token:  l2Token,
        L2Gas:    l2Gas,
    }, nil
}

// 通过Portal存入交易
func (dm *DepositManager) DepositTransaction(to common.Address, value *big.Int, gasLimit uint64, data []byte) (*DepositResult, error) {
    // 创建交易选项
    auth, err := dm.wallet.CreateTransactOpts(dm.config.L1ChainID)
    if err != nil {
        return nil, fmt.Errorf("创建交易选项失败: %v", err)
    }

    // 计算总费用（包括L2执行费用）
    l2ExecutionFee := new(big.Int).Mul(big.NewInt(int64(gasLimit)), dm.config.L2GasPrice)
    totalValue := new(big.Int).Add(value, l2ExecutionFee)

    auth.Value = totalValue
    auth.GasLimit = dm.config.L1GasLimit
    auth.GasPrice = dm.config.L1GasPrice

    // 调用depositTransaction
    tx, err := dm.portal.Transact(auth, "depositTransaction", to, value, gasLimit, false, data)
    if err != nil {
        return nil, fmt.Errorf("存入交易失败: %v", err)
    }

    return &DepositResult{
        L1TxHash: tx.Hash(),
        Amount:   value,
        To:       to,
        Type:     "TRANSACTION",
        L2Gas:    uint32(gasLimit),
        Data:     data,
    }, nil
}

// 批准代币
func (dm *DepositManager) approveToken(tokenAddr common.Address, amount *big.Int) error {
    // ERC20 approve逻辑
    // 简化实现
    return nil
}

// 等待L2交易确认
func (dm *DepositManager) WaitForL2Transaction(l1TxHash common.Hash) (*L2TransactionResult, error) {
    // 等待L1交易确认
    l1Receipt, err := dm.waitForL1Confirmation(l1TxHash)
    if err != nil {
        return nil, err
    }

    // 从L1交易日志中提取L2交易信息
    l2TxHash, err := dm.extractL2TxHash(l1Receipt)
    if err != nil {
        return nil, err
    }

    // 等待L2交易确认
    l2Receipt, err := dm.waitForL2Confirmation(l2TxHash)
    if err != nil {
        return nil, err
    }

    return &L2TransactionResult{
        L1TxHash:  l1TxHash,
        L2TxHash:  l2TxHash,
        L1Receipt: l1Receipt,
        L2Receipt: l2Receipt,
        Success:   l2Receipt.Status == 1,
    }, nil
}

// 等待L1确认
func (dm *DepositManager) waitForL1Confirmation(txHash common.Hash) (*types.Receipt, error) {
    for i := 0; i < dm.config.MaxRetries; i++ {
        receipt, err := dm.l1Client.TransactionReceipt(context.Background(), txHash)
        if err == nil {
            return receipt, nil
        }
        time.Sleep(dm.config.RetryDelay)
    }
    return nil, fmt.Errorf("等待L1交易确认超时")
}

// 等待L2确认
func (dm *DepositManager) waitForL2Confirmation(txHash common.Hash) (*types.Receipt, error) {
    for i := 0; i < dm.config.MaxRetries; i++ {
        receipt, err := dm.l2Client.TransactionReceipt(context.Background(), txHash)
        if err == nil {
            return receipt, nil
        }
        time.Sleep(dm.config.RetryDelay)
    }
    return nil, fmt.Errorf("等待L2交易确认超时")
}

// 从L1交易日志提取L2交易哈希
func (dm *DepositManager) extractL2TxHash(receipt *types.Receipt) (common.Hash, error) {
    // 解析TransactionDeposited事件
    // 简化实现，返回模拟哈希
    return common.HexToHash("0xabcdef1234567890"), nil
}

type DepositResult struct {
    L1TxHash common.Hash
    Amount   *big.Int
    To       common.Address
    Type     string
    L1Token  common.Address
    L2Token  common.Address
    L2Gas    uint32
    Data     []byte
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

// L2StandardBridge ABI（简化版）
const L2StandardBridgeABI = `[
    {
        "inputs": [
            {"name": "_l1Token", "type": "address"},
            {"name": "_amount", "type": "uint256"},
            {"name": "_l1Gas", "type": "uint32"},
            {"name": "_data", "type": "bytes"}
        ],
        "name": "withdraw",
        "outputs": [],
        "stateMutability": "nonpayable",
        "type": "function"
    },
    {
        "inputs": [
            {"name": "_amount", "type": "uint256"},
            {"name": "_l1Gas", "type": "uint32"},
            {"name": "_data", "type": "bytes"}
        ],
        "name": "withdrawETH",
        "outputs": [],
        "stateMutability": "nonpayable",
        "type": "function"
    }
]`

// OptimismPortal ABI（简化版）
const OptimismPortalWithdrawABI = `[
    {
        "inputs": [
            {"name": "_tx", "type": "tuple", "components": [
                {"name": "nonce", "type": "uint256"},
                {"name": "sender", "type": "address"},
                {"name": "target", "type": "address"},
                {"name": "value", "type": "uint256"},
                {"name": "gasLimit", "type": "uint256"},
                {"name": "data", "type": "bytes"}
            ]},
            {"name": "_l2OutputIndex", "type": "uint256"},
            {"name": "_outputRootProof", "type": "tuple", "components": [
                {"name": "version", "type": "bytes32"},
                {"name": "stateRoot", "type": "bytes32"},
                {"name": "messagePasserStorageRoot", "type": "bytes32"},
                {"name": "latestBlockhash", "type": "bytes32"}
            ]},
            {"name": "_withdrawalProof", "type": "bytes[]"}
        ],
        "name": "proveWithdrawalTransaction",
        "outputs": [],
        "stateMutability": "nonpayable",
        "type": "function"
    },
    {
        "inputs": [
            {"name": "_tx", "type": "tuple", "components": [
                {"name": "nonce", "type": "uint256"},
                {"name": "sender", "type": "address"},
                {"name": "target", "type": "address"},
                {"name": "value", "type": "uint256"},
                {"name": "gasLimit", "type": "uint256"},
                {"name": "data", "type": "bytes"}
            ]}
        ],
        "name": "finalizeWithdrawalTransaction",
        "outputs": [],
        "stateMutability": "nonpayable",
        "type": "function"
    }
]`

type WithdrawManager struct {
    l1Client *ethclient.Client
    l2Client *ethclient.Client
    config   *config.OptimismConfig
    wallet   *wallet.WalletManager
    l2Bridge *bind.BoundContract
    portal   *bind.BoundContract
}

func NewWithdrawManager(cfg *config.OptimismConfig, wallet *wallet.WalletManager) (*WithdrawManager, error) {
    // 连接客户端
    l1Client, err := ethclient.Dial(cfg.L1RPC)
    if err != nil {
        return nil, fmt.Errorf("连接L1节点失败: %v", err)
    }

    l2Client, err := ethclient.Dial(cfg.L2RPC)
    if err != nil {
        return nil, fmt.Errorf("连接L2节点失败: %v", err)
    }

    // 解析L2Bridge ABI
    l2BridgeABI, err := abi.JSON(strings.NewReader(L2StandardBridgeABI))
    if err != nil {
        return nil, fmt.Errorf("解析L2Bridge ABI失败: %v", err)
    }

    // 解析Portal ABI
    portalABI, err := abi.JSON(strings.NewReader(OptimismPortalWithdrawABI))
    if err != nil {
        return nil, fmt.Errorf("解析Portal ABI失败: %v", err)
    }

    // 创建合约实例
    l2Bridge := bind.NewBoundContract(cfg.L2StandardBridge, l2BridgeABI, l2Client, l2Client, l2Client)
    portal := bind.NewBoundContract(cfg.OptimismPortal, portalABI, l1Client, l1Client, l1Client)

    return &WithdrawManager{
        l1Client: l1Client,
        l2Client: l2Client,
        config:   cfg,
        wallet:   wallet,
        l2Bridge: l2Bridge,
        portal:   portal,
    }, nil
}

// 发起ETH提取
func (wm *WithdrawManager) WithdrawETH(amount *big.Int, l1Gas uint32) (*WithdrawInitiation, error) {
    // 创建L2交易选项
    auth, err := wm.wallet.CreateTransactOpts(wm.config.L2ChainID)
    if err != nil {
        return nil, fmt.Errorf("创建交易选项失败: %v", err)
    }

    auth.GasLimit = wm.config.L2GasLimit
    auth.GasPrice = wm.config.L2GasPrice

    // 调用withdrawETH
    tx, err := wm.l2Bridge.Transact(auth, "withdrawETH", amount, l1Gas, []byte{})
    if err != nil {
        return nil, fmt.Errorf("发起ETH提取失败: %v", err)
    }

    return &WithdrawInitiation{
        L2TxHash: tx.Hash(),
        Amount:   amount,
        Type:     "ETH",
        L1Gas:    l1Gas,
    }, nil
}

// 发起ERC20提取
func (wm *WithdrawManager) WithdrawERC20(l1Token common.Address, amount *big.Int, l1Gas uint32) (*WithdrawInitiation, error) {
    // 创建L2交易选项
    auth, err := wm.wallet.CreateTransactOpts(wm.config.L2ChainID)
    if err != nil {
        return nil, fmt.Errorf("创建交易选项失败: %v", err)
    }

    auth.GasLimit = wm.config.L2GasLimit
    auth.GasPrice = wm.config.L2GasPrice

    // 调用withdraw
    tx, err := wm.l2Bridge.Transact(auth, "withdraw", l1Token, amount, l1Gas, []byte{})
    if err != nil {
        return nil, fmt.Errorf("发起ERC20提取失败: %v", err)
    }

    return &WithdrawInitiation{
        L2TxHash: tx.Hash(),
        Amount:   amount,
        Type:     "ERC20",
        L1Token:  l1Token,
        L1Gas:    l1Gas,
    }, nil
}

// 证明提取交易
func (wm *WithdrawManager) ProveWithdrawal(proof WithdrawProof) (*WithdrawProofResult, error) {
    // 创建L1交易选项
    auth, err := wm.wallet.CreateTransactOpts(wm.config.L1ChainID)
    if err != nil {
        return nil, fmt.Errorf("创建交易选项失败: %v", err)
    }

    auth.GasLimit = wm.config.L1GasLimit
    auth.GasPrice = wm.config.L1GasPrice

    // 构建交易结构
    withdrawalTx := struct {
        Nonce    *big.Int
        Sender   common.Address
        Target   common.Address
        Value    *big.Int
        GasLimit *big.Int
        Data     []byte
    }{
        Nonce:    proof.Nonce,
        Sender:   proof.Sender,
        Target:   proof.Target,
        Value:    proof.Value,
        GasLimit: proof.GasLimit,
        Data:     proof.Data,
    }

    // 构建输出根证明
    outputRootProof := struct {
        Version                   [32]byte
        StateRoot                 [32]byte
        MessagePasserStorageRoot  [32]byte
        LatestBlockhash          [32]byte
    }{
        Version:                  proof.OutputRootProof.Version,
        StateRoot:                proof.OutputRootProof.StateRoot,
        MessagePasserStorageRoot: proof.OutputRootProof.MessagePasserStorageRoot,
        LatestBlockhash:         proof.OutputRootProof.LatestBlockhash,
    }

    // 调用proveWithdrawalTransaction
    tx, err := wm.portal.Transact(
        auth,
        "proveWithdrawalTransaction",
        withdrawalTx,
        proof.L2OutputIndex,
        outputRootProof,
        proof.WithdrawalProof,
    )
    if err != nil {
        return nil, fmt.Errorf("证明提取交易失败: %v", err)
    }

    return &WithdrawProofResult{
        L1TxHash: tx.Hash(),
        Proof:    proof,
    }, nil
}

// 最终确认提取
func (wm *WithdrawManager) FinalizeWithdrawal(withdrawalTx WithdrawalTransaction) (*WithdrawFinalization, error) {
    // 创建L1交易选项
    auth, err := wm.wallet.CreateTransactOpts(wm.config.L1ChainID)
    if err != nil {
        return nil, fmt.Errorf("创建交易选项失败: %v", err)
    }

    auth.GasLimit = wm.config.L1GasLimit
    auth.GasPrice = wm.config.L1GasPrice

    // 构建交易结构
    tx := struct {
        Nonce    *big.Int
        Sender   common.Address
        Target   common.Address
        Value    *big.Int
        GasLimit *big.Int
        Data     []byte
    }{
        Nonce:    withdrawalTx.Nonce,
        Sender:   withdrawalTx.Sender,
        Target:   withdrawalTx.Target,
        Value:    withdrawalTx.Value,
        GasLimit: withdrawalTx.GasLimit,
        Data:     withdrawalTx.Data,
    }

    // 调用finalizeWithdrawalTransaction
    finalTx, err := wm.portal.Transact(auth, "finalizeWithdrawalTransaction", tx)
    if err != nil {
        return nil, fmt.Errorf("最终确认提取失败: %v", err)
    }

    return &WithdrawFinalization{
        L1TxHash:      finalTx.Hash(),
        WithdrawalTx:  withdrawalTx,
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
        L2TxHash:   l2TxHash,
        L2Receipt:  l2Receipt,
        Status:     "initiated",
        L2BlockNum: l2Receipt.BlockNumber.Uint64(),
    }

    // 检查是否可以证明
    canProve, err := wm.canProveWithdrawal(l2Receipt)
    if err != nil {
        return status, err
    }

    if canProve {
        status.Status = "ready_to_prove"
        
        // 检查是否已证明
        isProven, err := wm.isWithdrawalProven(l2TxHash)
        if err != nil {
            return status, err
        }

        if isProven {
            status.Status = "proven"
            
            // 检查是否可以最终确认
            canFinalize, err := wm.canFinalizeWithdrawal(l2Receipt)
            if err != nil {
                return status, err
            }

            if canFinalize {
                status.Status = "ready_to_finalize"
            } else {
                status.Status = "waiting_for_finalization_period"
            }
        }
    } else {
        status.Status = "waiting_for_state_root"
    }

    return status, nil
}

// 检查是否可以证明提取
func (wm *WithdrawManager) canProveWithdrawal(l2Receipt *types.Receipt) (bool, error) {
    // 检查L2输出是否已提交到L1
    // 简化实现
    return true, nil
}

// 检查提取是否已证明
func (wm *WithdrawManager) isWithdrawalProven(l2TxHash common.Hash) (bool, error) {
    // 检查Portal合约中的证明状态
    // 简化实现
    return false, nil
}

// 检查是否可以最终确认
func (wm *WithdrawManager) canFinalizeWithdrawal(l2Receipt *types.Receipt) (bool, error) {
    // 检查最终确认期是否已过
    // 简化实现
    return false, nil
}

type WithdrawInitiation struct {
    L2TxHash common.Hash
    Amount   *big.Int
    Type     string
    L1Token  common.Address
    L1Gas    uint32
}

type WithdrawProof struct {
    Nonce            *big.Int
    Sender           common.Address
    Target           common.Address
    Value            *big.Int
    GasLimit         *big.Int
    Data             []byte
    L2OutputIndex    *big.Int
    OutputRootProof  OutputRootProof
    WithdrawalProof  [][]byte
}

type OutputRootProof struct {
    Version                  [32]byte
    StateRoot                [32]byte
    MessagePasserStorageRoot [32]byte
    LatestBlockhash         [32]byte
}

type WithdrawProofResult struct {
    L1TxHash common.Hash
    Proof    WithdrawProof
}

type WithdrawalTransaction struct {
    Nonce    *big.Int
    Sender   common.Address
    Target   common.Address
    Value    *big.Int
    GasLimit *big.Int
    Data     []byte
}

type WithdrawFinalization struct {
    L1TxHash     common.Hash
    WithdrawalTx WithdrawalTransaction
}

type WithdrawStatus struct {
    L2TxHash   common.Hash
    L2Receipt  *types.Receipt
    Status     string
    L2BlockNum uint64
}
```

## 实际应用

### 9.1 完整Optimism应用

```go
// main.go
package main

import (
    "fmt"
    "log"
    "math/big"

    "github.com/ethereum/go-ethereum/common"

    "your-project/config"
    "your-project/deposit"
    "your-project/withdraw"
    "your-project/wallet"
)

func main() {
    // 创建Optimism配置
    cfg := config.OptimismGoerliConfig() // 使用测试网

    // 创建钱包
    walletManager, err := wallet.NewWalletManager("your_private_key_here")
    if err != nil {
        log.Fatal("创建钱包失败:", err)
    }

    fmt.Printf("钱包地址: %s\n", walletManager.GetAddress().Hex())

    // 存款管理示例
    fmt.Println("=== 存款管理示例 ===")
    
    depositManager, err := deposit.NewDepositManager(cfg, walletManager)
    if err != nil {
        log.Fatal("创建存款管理器失败:", err)
    }

    // 存入ETH到L2
    ethAmount := big.NewInt(100000000000000000) // 0.1 ETH
    ethDeposit, err := depositManager.DepositETH(ethAmount, 200000)
    if err != nil {
        log.Printf("存入ETH失败: %v", err)
    } else {
        fmt.Printf("ETH存款成功:\n")
        fmt.Printf("  L1交易哈希: %s\n", ethDeposit.L1TxHash.Hex())
        fmt.Printf("  金额: %s Wei\n", ethDeposit.Amount.String())
        fmt.Printf("  类型: %s\n", ethDeposit.Type)
        fmt.Printf("  L2 Gas: %d\n", ethDeposit.L2Gas)

        // 等待L2交易确认
        l2Result, err := depositManager.WaitForL2Transaction(ethDeposit.L1TxHash)
        if err != nil {
            log.Printf("等待L2交易失败: %v", err)
        } else {
            fmt.Printf("L2交易确认:\n")
            fmt.Printf("  L2交易哈希: %s\n", l2Result.L2TxHash.Hex())
            fmt.Printf("  成功: %t\n", l2Result.Success)
        }
    }

    // ERC20存款示例
    testL1Token := common.HexToAddress("0x1234567890123456789012345678901234567890")
    testL2Token := common.HexToAddress("0x0987654321098765432109876543210987654321")
    tokenAmount := big.NewInt(1000000000000000000) // 1 token

    tokenDeposit, err := depositManager.DepositERC20(testL1Token, testL2Token, tokenAmount, 200000)
    if err != nil {
        log.Printf("存入ERC20失败: %v", err)
    } else {
        fmt.Printf("ERC20存款成功:\n")
        fmt.Printf("  L1交易哈希: %s\n", tokenDeposit.L1TxHash.Hex())
        fmt.Printf("  L1代币: %s\n", tokenDeposit.L1Token.Hex())
        fmt.Printf("  L2代币: %s\n", tokenDeposit.L2Token.Hex())
        fmt.Printf("  金额: %s\n", tokenDeposit.Amount.String())
    }

    // 提取管理示例
    fmt.Println("\n=== 提取管理示例 ===")
    
    withdrawManager, err := withdraw.NewWithdrawManager(cfg, walletManager)
    if err != nil {
        log.Fatal("创建提取管理器失败:", err)
    }

    // 发起ETH提取
    withdrawAmount := big.NewInt(50000000000000000) // 0.05 ETH
    ethWithdraw, err := withdrawManager.WithdrawETH(withdrawAmount, 100000)
    if err != nil {
        log.Printf("发起ETH提取失败: %v", err)
    } else {
        fmt.Printf("ETH提取发起成功:\n")
        fmt.Printf("  L2交易哈希: %s\n", ethWithdraw.L2TxHash.Hex())
        fmt.Printf("  金额: %s Wei\n", ethWithdraw.Amount.String())
        fmt.Printf("  类型: %s\n", ethWithdraw.Type)

        // 检查提取状态
        withdrawStatus, err := withdrawManager.CheckWithdrawStatus(ethWithdraw.L2TxHash)
        if err != nil {
            log.Printf("检查提取状态失败: %v", err)
        } else {
            fmt.Printf("提取状态: %s\n", withdrawStatus.Status)
            if withdrawStatus.L2Receipt != nil {
                fmt.Printf("L2区块号: %d\n", withdrawStatus.L2BlockNum)
            }
        }
    }

    // ERC20提取示例
    tokenWithdraw, err := withdrawManager.WithdrawERC20(testL1Token, tokenAmount, 100000)
    if err != nil {
        log.Printf("发起ERC20提取失败: %v", err)
    } else {
        fmt.Printf("ERC20提取发起成功:\n")
        fmt.Printf("  L2交易哈希: %s\n", tokenWithdraw.L2TxHash.Hex())
        fmt.Printf("  L1代币: %s\n", tokenWithdraw.L1Token.Hex())
        fmt.Printf("  金额: %s\n", tokenWithdraw.Amount.String())
    }

    // 提取证明示例（模拟数据）
    fmt.Println("\n=== 提取证明示例 ===")
    
    // 构建模拟证明数据
    mockProof := withdraw.WithdrawProof{
        Nonce:         big.NewInt(1),
        Sender:        walletManager.GetAddress(),
        Target:        walletManager.GetAddress(),
        Value:         withdrawAmount,
        GasLimit:      big.NewInt(100000),
        Data:          []byte{},
        L2OutputIndex: big.NewInt(100),
        OutputRootProof: withdraw.OutputRootProof{
            Version:                  [32]byte{},
            StateRoot:                [32]byte{},
            MessagePasserStorageRoot: [32]byte{},
            LatestBlockhash:         [32]byte{},
        },
        WithdrawalProof: [][]byte{},
    }

    proofResult, err := withdrawManager.ProveWithdrawal(mockProof)
    if err != nil {
        log.Printf("证明提取失败: %v", err)
    } else {
        fmt.Printf("提取证明成功:\n")
        fmt.Printf("  L1交易哈希: %s\n", proofResult.L1TxHash.Hex())
    }

    // 最终确认示例
    mockWithdrawalTx := withdraw.WithdrawalTransaction{
        Nonce:    big.NewInt(1),
        Sender:   walletManager.GetAddress(),
        Target:   walletManager.GetAddress(),
        Value:    withdrawAmount,
        GasLimit: big.NewInt(100000),
        Data:     []byte{},
    }

    finalization, err := withdrawManager.FinalizeWithdrawal(mockWithdrawalTx)
    if err != nil {
        log.Printf("最终确认提取失败: %v", err)
    } else {
        fmt.Printf("提取最终确认成功:\n")
        fmt.Printf("  L1交易哈希: %s\n", finalization.L1TxHash.Hex())
    }

    fmt.Println("Optimism操作演示完成!")
}
```
