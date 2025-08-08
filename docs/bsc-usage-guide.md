# Binance Smart Chain (BSC) 详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [环境准备](#环境准备)
3. [网络连接](#网络连接)
4. [账户管理](#账户管理)
5. [BNB操作](#BNB操作)
6. [BEP20代币](#BEP20代币)
7. [智能合约交互](#智能合约交互)
8. [DeFi协议集成](#DeFi协议集成)
9. [实际应用](#实际应用)

## 基础概念

### 1.1 BSC简介

Binance Smart Chain是币安推出的区块链网络，兼容以太坊虚拟机(EVM)，支持智能合约和DeFi应用。

```bash
# 安装BSC相关依赖
go get github.com/ethereum/go-ethereum
go get github.com/binance-chain/bsc
go get github.com/shopspring/decimal
```

### 1.2 网络信息

```go
// config/bsc.go
package config

import (
    "math/big"
)

// BSC网络配置
type BSCConfig struct {
    MainnetRPC    string
    TestnetRPC    string
    ChainID       *big.Int
    TestChainID   *big.Int
    GasLimit      uint64
    GasPrice      *big.Int
    ExplorerURL   string
}

func DefaultBSCConfig() *BSCConfig {
    return &BSCConfig{
        MainnetRPC:  "https://bsc-dataseed1.binance.org/",
        TestnetRPC:  "https://data-seed-prebsc-1-s1.binance.org:8545/",
        ChainID:     big.NewInt(56),  // BSC主网
        TestChainID: big.NewInt(97),  // BSC测试网
        GasLimit:    21000,
        GasPrice:    big.NewInt(5000000000), // 5 Gwei
        ExplorerURL: "https://bscscan.com",
    }
}

// 常用合约地址
var (
    // BEP20代币合约
    BUSD_CONTRACT = "0xe9e7CEA3DedcA5984780Bafc599bD69ADd087D56"
    USDT_CONTRACT = "0x55d398326f99059fF775485246999027B3197955"
    CAKE_CONTRACT = "0x0E09FaBB73Bd3Ade0a17ECC321fD13a19e81cE82"
    
    // DeFi协议合约
    PANCAKESWAP_ROUTER = "0x10ED43C718714eb63d5aA57B78B54704E256024E"
    PANCAKESWAP_FACTORY = "0xcA143Ce32Fe78f1f7019d7d551a6402fC5350c73"
)
```

## 环境准备

### 2.1 客户端连接

```go
// client/bsc_client.go
package client

import (
    "context"
    "fmt"
    "math/big"

    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
)

type BSCClient struct {
    client  *ethclient.Client
    chainID *big.Int
    config  *BSCConfig
}

func NewBSCClient(rpcURL string, chainID *big.Int) (*BSCClient, error) {
    client, err := ethclient.Dial(rpcURL)
    if err != nil {
        return nil, fmt.Errorf("连接BSC节点失败: %v", err)
    }

    return &BSCClient{
        client:  client,
        chainID: chainID,
    }, nil
}

// 获取最新区块号
func (b *BSCClient) GetLatestBlockNumber() (uint64, error) {
    header, err := b.client.HeaderByNumber(context.Background(), nil)
    if err != nil {
        return 0, err
    }
    return header.Number.Uint64(), nil
}

// 获取BNB余额
func (b *BSCClient) GetBNBBalance(address common.Address) (*big.Int, error) {
    balance, err := b.client.BalanceAt(context.Background(), address, nil)
    if err != nil {
        return nil, err
    }
    return balance, nil
}

// 获取Gas价格
func (b *BSCClient) GetGasPrice() (*big.Int, error) {
    gasPrice, err := b.client.SuggestGasPrice(context.Background())
    if err != nil {
        return nil, err
    }
    return gasPrice, nil
}

// 获取Nonce
func (b *BSCClient) GetNonce(address common.Address) (uint64, error) {
    nonce, err := b.client.PendingNonceAt(context.Background(), address)
    if err != nil {
        return 0, err
    }
    return nonce, nil
}

// 发送交易
func (b *BSCClient) SendTransaction(signedTx *types.Transaction) error {
    return b.client.SendTransaction(context.Background(), signedTx)
}

// 等待交易确认
func (b *BSCClient) WaitForReceipt(txHash common.Hash) (*types.Receipt, error) {
    for {
        receipt, err := b.client.TransactionReceipt(context.Background(), txHash)
        if err != nil {
            continue
        }
        return receipt, nil
    }
}
```

## 网络连接

### 3.1 多网络管理

```go
// network/manager.go
package network

import (
    "fmt"
    "math/big"

    "your-project/client"
    "your-project/config"
)

type NetworkManager struct {
    mainnetClient *client.BSCClient
    testnetClient *client.BSCClient
    config        *config.BSCConfig
}

func NewNetworkManager(cfg *config.BSCConfig) (*NetworkManager, error) {
    // 连接主网
    mainnetClient, err := client.NewBSCClient(cfg.MainnetRPC, cfg.ChainID)
    if err != nil {
        return nil, fmt.Errorf("连接BSC主网失败: %v", err)
    }

    // 连接测试网
    testnetClient, err := client.NewBSCClient(cfg.TestnetRPC, cfg.TestChainID)
    if err != nil {
        return nil, fmt.Errorf("连接BSC测试网失败: %v", err)
    }

    return &NetworkManager{
        mainnetClient: mainnetClient,
        testnetClient: testnetClient,
        config:        cfg,
    }, nil
}

// 获取主网客户端
func (n *NetworkManager) GetMainnetClient() *client.BSCClient {
    return n.mainnetClient
}

// 获取测试网客户端
func (n *NetworkManager) GetTestnetClient() *client.BSCClient {
    return n.testnetClient
}

// 根据链ID获取客户端
func (n *NetworkManager) GetClientByChainID(chainID *big.Int) *client.BSCClient {
    if chainID.Cmp(n.config.ChainID) == 0 {
        return n.mainnetClient
    }
    return n.testnetClient
}

// 检查网络连接状态
func (n *NetworkManager) CheckNetworkStatus() error {
    // 检查主网
    _, err := n.mainnetClient.GetLatestBlockNumber()
    if err != nil {
        return fmt.Errorf("主网连接异常: %v", err)
    }

    // 检查测试网
    _, err = n.testnetClient.GetLatestBlockNumber()
    if err != nil {
        return fmt.Errorf("测试网连接异常: %v", err)
    }

    return nil
}
```

## 账户管理

### 4.1 钱包操作

```go
// wallet/manager.go
package wallet

import (
    "crypto/ecdsa"
    "fmt"
    "math/big"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/crypto"
)

type WalletManager struct {
    privateKey *ecdsa.PrivateKey
    publicKey  *ecdsa.PublicKey
    address    common.Address
}

func NewWalletManager(privateKeyHex string) (*WalletManager, error) {
    privateKey, err := crypto.HexToECDSA(privateKeyHex)
    if err != nil {
        return nil, fmt.Errorf("解析私钥失败: %v", err)
    }

    publicKey := privateKey.Public()
    publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
    if !ok {
        return nil, fmt.Errorf("获取公钥失败")
    }

    address := crypto.PubkeyToAddress(*publicKeyECDSA)

    return &WalletManager{
        privateKey: privateKey,
        publicKey:  publicKeyECDSA,
        address:    address,
    }, nil
}

// 生成新钱包
func GenerateWallet() (*WalletManager, error) {
    privateKey, err := crypto.GenerateKey()
    if err != nil {
        return nil, fmt.Errorf("生成私钥失败: %v", err)
    }

    publicKey := privateKey.Public()
    publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
    if !ok {
        return nil, fmt.Errorf("获取公钥失败")
    }

    address := crypto.PubkeyToAddress(*publicKeyECDSA)

    return &WalletManager{
        privateKey: privateKey,
        publicKey:  publicKeyECDSA,
        address:    address,
    }, nil
}

// 获取地址
func (w *WalletManager) GetAddress() common.Address {
    return w.address
}

// 获取私钥十六进制
func (w *WalletManager) GetPrivateKeyHex() string {
    return fmt.Sprintf("%x", crypto.FromECDSA(w.privateKey))
}

// 签名交易
func (w *WalletManager) SignTransaction(tx *types.Transaction, chainID *big.Int) (*types.Transaction, error) {
    signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), w.privateKey)
    if err != nil {
        return nil, fmt.Errorf("签名交易失败: %v", err)
    }
    return signedTx, nil
}

// 签名消息
func (w *WalletManager) SignMessage(message []byte) ([]byte, error) {
    hash := crypto.Keccak256Hash(message)
    signature, err := crypto.Sign(hash.Bytes(), w.privateKey)
    if err != nil {
        return nil, fmt.Errorf("签名消息失败: %v", err)
    }
    return signature, nil
}
```

## BNB操作

### 5.1 BNB转账

```go
// bnb/transfer.go
package bnb

import (
    "context"
    "fmt"
    "math/big"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/shopspring/decimal"

    "your-project/client"
    "your-project/wallet"
)

type BNBTransfer struct {
    client *client.BSCClient
    wallet *wallet.WalletManager
}

func NewBNBTransfer(client *client.BSCClient, wallet *wallet.WalletManager) *BNBTransfer {
    return &BNBTransfer{
        client: client,
        wallet: wallet,
    }
}

// BNB转账
func (b *BNBTransfer) Transfer(to common.Address, amount decimal.Decimal) (*types.Transaction, error) {
    // 转换金额为Wei
    amountWei := amount.Mul(decimal.NewFromFloat(1e18)).BigInt()

    // 获取nonce
    nonce, err := b.client.GetNonce(b.wallet.GetAddress())
    if err != nil {
        return nil, fmt.Errorf("获取nonce失败: %v", err)
    }

    // 获取gas价格
    gasPrice, err := b.client.GetGasPrice()
    if err != nil {
        return nil, fmt.Errorf("获取gas价格失败: %v", err)
    }

    // 创建交易
    tx := types.NewTransaction(
        nonce,
        to,
        amountWei,
        21000, // BNB转账的gas限制
        gasPrice,
        nil,
    )

    // 签名交易
    signedTx, err := b.wallet.SignTransaction(tx, b.client.chainID)
    if err != nil {
        return nil, fmt.Errorf("签名交易失败: %v", err)
    }

    // 发送交易
    err = b.client.SendTransaction(signedTx)
    if err != nil {
        return nil, fmt.Errorf("发送交易失败: %v", err)
    }

    return signedTx, nil
}

// 批量转账
func (b *BNBTransfer) BatchTransfer(transfers []TransferData) ([]*types.Transaction, error) {
    var transactions []*types.Transaction
    
    // 获取起始nonce
    nonce, err := b.client.GetNonce(b.wallet.GetAddress())
    if err != nil {
        return nil, fmt.Errorf("获取nonce失败: %v", err)
    }

    // 获取gas价格
    gasPrice, err := b.client.GetGasPrice()
    if err != nil {
        return nil, fmt.Errorf("获取gas价格失败: %v", err)
    }

    for i, transfer := range transfers {
        // 转换金额为Wei
        amountWei := transfer.Amount.Mul(decimal.NewFromFloat(1e18)).BigInt()

        // 创建交易
        tx := types.NewTransaction(
            nonce+uint64(i),
            transfer.To,
            amountWei,
            21000,
            gasPrice,
            nil,
        )

        // 签名交易
        signedTx, err := b.wallet.SignTransaction(tx, b.client.chainID)
        if err != nil {
            return nil, fmt.Errorf("签名交易失败: %v", err)
        }

        // 发送交易
        err = b.client.SendTransaction(signedTx)
        if err != nil {
            return nil, fmt.Errorf("发送交易失败: %v", err)
        }

        transactions = append(transactions, signedTx)
    }

    return transactions, nil
}

// 估算转账费用
func (b *BNBTransfer) EstimateTransferCost(amount decimal.Decimal) (*TransferCost, error) {
    gasPrice, err := b.client.GetGasPrice()
    if err != nil {
        return nil, err
    }

    gasLimit := big.NewInt(21000)
    gasCost := new(big.Int).Mul(gasPrice, gasLimit)
    
    amountWei := amount.Mul(decimal.NewFromFloat(1e18)).BigInt()
    total := new(big.Int).Add(amountWei, gasCost)

    return &TransferCost{
        Amount:   amount,
        GasCost:  decimal.NewFromBigInt(gasCost, -18),
        Total:    decimal.NewFromBigInt(total, -18),
        GasPrice: decimal.NewFromBigInt(gasPrice, -9), // Gwei
        GasLimit: gasLimit.Uint64(),
    }, nil
}

type TransferData struct {
    To     common.Address
    Amount decimal.Decimal
}

type TransferCost struct {
    Amount   decimal.Decimal
    GasCost  decimal.Decimal
    Total    decimal.Decimal
    GasPrice decimal.Decimal
    GasLimit uint64
}
```

## BEP20代币

### 6.1 代币操作

```go
// token/bep20.go
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
    "github.com/shopspring/decimal"

    "your-project/client"
    "your-project/wallet"
)

// BEP20标准ABI
const BEP20ABI = `[
    {
        "constant": true,
        "inputs": [],
        "name": "name",
        "outputs": [{"name": "", "type": "string"}],
        "type": "function"
    },
    {
        "constant": true,
        "inputs": [],
        "name": "symbol",
        "outputs": [{"name": "", "type": "string"}],
        "type": "function"
    },
    {
        "constant": true,
        "inputs": [],
        "name": "decimals",
        "outputs": [{"name": "", "type": "uint8"}],
        "type": "function"
    },
    {
        "constant": true,
        "inputs": [],
        "name": "totalSupply",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    },
    {
        "constant": true,
        "inputs": [{"name": "_owner", "type": "address"}],
        "name": "balanceOf",
        "outputs": [{"name": "balance", "type": "uint256"}],
        "type": "function"
    },
    {
        "constant": false,
        "inputs": [
            {"name": "_to", "type": "address"},
            {"name": "_value", "type": "uint256"}
        ],
        "name": "transfer",
        "outputs": [{"name": "", "type": "bool"}],
        "type": "function"
    },
    {
        "constant": true,
        "inputs": [
            {"name": "_owner", "type": "address"},
            {"name": "_spender", "type": "address"}
        ],
        "name": "allowance",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    },
    {
        "constant": false,
        "inputs": [
            {"name": "_spender", "type": "address"},
            {"name": "_value", "type": "uint256"}
        ],
        "name": "approve",
        "outputs": [{"name": "", "type": "bool"}],
        "type": "function"
    }
]`

type BEP20Token struct {
    client       *client.BSCClient
    wallet       *wallet.WalletManager
    contract     *bind.BoundContract
    contractAddr common.Address
    abi          abi.ABI
    decimals     uint8
    symbol       string
    name         string
}

func NewBEP20Token(client *client.BSCClient, wallet *wallet.WalletManager, contractAddr common.Address) (*BEP20Token, error) {
    parsedABI, err := abi.JSON(strings.NewReader(BEP20ABI))
    if err != nil {
        return nil, fmt.Errorf("解析ABI失败: %v", err)
    }

    contract := bind.NewBoundContract(contractAddr, parsedABI, client.client, client.client, client.client)

    token := &BEP20Token{
        client:       client,
        wallet:       wallet,
        contract:     contract,
        contractAddr: contractAddr,
        abi:          parsedABI,
    }

    // 获取代币信息
    if err := token.loadTokenInfo(); err != nil {
        return nil, fmt.Errorf("加载代币信息失败: %v", err)
    }

    return token, nil
}

// 加载代币信息
func (t *BEP20Token) loadTokenInfo() error {
    // 获取decimals
    var decimalsResult []interface{}
    err := t.contract.Call(nil, &decimalsResult, "decimals")
    if err != nil {
        return err
    }
    t.decimals = decimalsResult[0].(uint8)

    // 获取symbol
    var symbolResult []interface{}
    err = t.contract.Call(nil, &symbolResult, "symbol")
    if err != nil {
        return err
    }
    t.symbol = symbolResult[0].(string)

    // 获取name
    var nameResult []interface{}
    err = t.contract.Call(nil, &nameResult, "name")
    if err != nil {
        return err
    }
    t.name = nameResult[0].(string)

    return nil
}

// 获取余额
func (t *BEP20Token) GetBalance(address common.Address) (decimal.Decimal, error) {
    var result []interface{}
    err := t.contract.Call(nil, &result, "balanceOf", address)
    if err != nil {
        return decimal.Zero, err
    }

    balance := result[0].(*big.Int)
    return decimal.NewFromBigInt(balance, -int32(t.decimals)), nil
}

// 转账
func (t *BEP20Token) Transfer(to common.Address, amount decimal.Decimal) (*types.Transaction, error) {
    // 转换金额
    amountBig := amount.Shift(int32(t.decimals)).BigInt()

    // 获取nonce
    nonce, err := t.client.GetNonce(t.wallet.GetAddress())
    if err != nil {
        return nil, err
    }

    // 获取gas价格
    gasPrice, err := t.client.GetGasPrice()
    if err != nil {
        return nil, err
    }

    // 创建交易选项
    auth := &bind.TransactOpts{
        From:     t.wallet.GetAddress(),
        Nonce:    big.NewInt(int64(nonce)),
        GasLimit: 100000, // BEP20转账的gas限制
        GasPrice: gasPrice,
        Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
            return t.wallet.SignTransaction(tx, t.client.chainID)
        },
    }

    // 调用transfer方法
    tx, err := t.contract.Transact(auth, "transfer", to, amountBig)
    if err != nil {
        return nil, err
    }

    return tx, nil
}

// 授权
func (t *BEP20Token) Approve(spender common.Address, amount decimal.Decimal) (*types.Transaction, error) {
    amountBig := amount.Shift(int32(t.decimals)).BigInt()

    nonce, err := t.client.GetNonce(t.wallet.GetAddress())
    if err != nil {
        return nil, err
    }

    gasPrice, err := t.client.GetGasPrice()
    if err != nil {
        return nil, err
    }

    auth := &bind.TransactOpts{
        From:     t.wallet.GetAddress(),
        Nonce:    big.NewInt(int64(nonce)),
        GasLimit: 60000,
        GasPrice: gasPrice,
        Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
            return t.wallet.SignTransaction(tx, t.client.chainID)
        },
    }

    tx, err := t.contract.Transact(auth, "approve", spender, amountBig)
    if err != nil {
        return nil, err
    }

    return tx, nil
}

// 获取授权额度
func (t *BEP20Token) GetAllowance(owner, spender common.Address) (decimal.Decimal, error) {
    var result []interface{}
    err := t.contract.Call(nil, &result, "allowance", owner, spender)
    if err != nil {
        return decimal.Zero, err
    }

    allowance := result[0].(*big.Int)
    return decimal.NewFromBigInt(allowance, -int32(t.decimals)), nil
}

// 获取代币信息
func (t *BEP20Token) GetTokenInfo() TokenInfo {
    return TokenInfo{
        Name:     t.name,
        Symbol:   t.symbol,
        Decimals: t.decimals,
        Address:  t.contractAddr.Hex(),
    }
}

type TokenInfo struct {
    Name     string
    Symbol   string
    Decimals uint8
    Address  string
}
```

## 智能合约交互

### 7.1 合约调用

```go
// contract/interaction.go
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

    "your-project/client"
    "your-project/wallet"
)

type ContractInteraction struct {
    client   *client.BSCClient
    wallet   *wallet.WalletManager
    contract *bind.BoundContract
    abi      abi.ABI
    address  common.Address
}

func NewContractInteraction(client *client.BSCClient, wallet *wallet.WalletManager, contractAddr common.Address, abiJSON string) (*ContractInteraction, error) {
    parsedABI, err := abi.JSON(strings.NewReader(abiJSON))
    if err != nil {
        return nil, fmt.Errorf("解析ABI失败: %v", err)
    }

    contract := bind.NewBoundContract(contractAddr, parsedABI, client.client, client.client, client.client)

    return &ContractInteraction{
        client:   client,
        wallet:   wallet,
        contract: contract,
        abi:      parsedABI,
        address:  contractAddr,
    }, nil
}

// 调用只读方法
func (c *ContractInteraction) CallMethod(methodName string, result interface{}, params ...interface{}) error {
    return c.contract.Call(nil, result, methodName, params...)
}

// 调用写入方法
func (c *ContractInteraction) TransactMethod(methodName string, gasLimit uint64, params ...interface{}) (*types.Transaction, error) {
    nonce, err := c.client.GetNonce(c.wallet.GetAddress())
    if err != nil {
        return nil, err
    }

    gasPrice, err := c.client.GetGasPrice()
    if err != nil {
        return nil, err
    }

    auth := &bind.TransactOpts{
        From:     c.wallet.GetAddress(),
        Nonce:    big.NewInt(int64(nonce)),
        GasLimit: gasLimit,
        GasPrice: gasPrice,
        Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
            return c.wallet.SignTransaction(tx, c.client.chainID)
        },
    }

    return c.contract.Transact(auth, methodName, params...)
}

// 估算Gas费用
func (c *ContractInteraction) EstimateGas(methodName string, params ...interface{}) (uint64, error) {
    // 打包方法调用数据
    data, err := c.abi.Pack(methodName, params...)
    if err != nil {
        return 0, err
    }

    // 估算gas
    msg := ethereum.CallMsg{
        From: c.wallet.GetAddress(),
        To:   &c.address,
        Data: data,
    }

    gasLimit, err := c.client.client.EstimateGas(context.Background(), msg)
    if err != nil {
        return 0, err
    }

    return gasLimit, nil
}

// 监听合约事件
func (c *ContractInteraction) WatchEvents(eventName string, handler func(types.Log)) error {
    query := ethereum.FilterQuery{
        Addresses: []common.Address{c.address},
        Topics:    [][]common.Hash{},
    }

    // 获取事件签名
    event := c.abi.Events[eventName]
    if event.ID != (common.Hash{}) {
        query.Topics = append(query.Topics, []common.Hash{event.ID})
    }

    logs := make(chan types.Log)
    sub, err := c.client.client.SubscribeFilterLogs(context.Background(), query, logs)
    if err != nil {
        return err
    }

    go func() {
        defer sub.Unsubscribe()
        for {
            select {
            case err := <-sub.Err():
                fmt.Printf("事件监听错误: %v\n", err)
                return
            case vLog := <-logs:
                handler(vLog)
            }
        }
    }()

    return nil
}
```

## DeFi协议集成

### 8.1 PancakeSwap集成

```go
// defi/pancakeswap.go
package defi

import (
    "fmt"
    "math/big"

    "github.com/ethereum/go-ethereum/common"
    "github.com/shopspring/decimal"

    "your-project/client"
    "your-project/wallet"
    "your-project/contract"
)

// PancakeSwap路由器ABI（简化版）
const PancakeRouterABI = `[
    {
        "inputs": [
            {"internalType": "uint256", "name": "amountIn", "type": "uint256"},
            {"internalType": "address[]", "name": "path", "type": "address[]"}
        ],
        "name": "getAmountsOut",
        "outputs": [
            {"internalType": "uint256[]", "name": "amounts", "type": "uint256[]"}
        ],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [
            {"internalType": "uint256", "name": "amountIn", "type": "uint256"},
            {"internalType": "uint256", "name": "amountOutMin", "type": "uint256"},
            {"internalType": "address[]", "name": "path", "type": "address[]"},
            {"internalType": "address", "name": "to", "type": "address"},
            {"internalType": "uint256", "name": "deadline", "type": "uint256"}
        ],
        "name": "swapExactTokensForTokens",
        "outputs": [
            {"internalType": "uint256[]", "name": "amounts", "type": "uint256[]"}
        ],
        "stateMutability": "nonpayable",
        "type": "function"
    }
]`

type PancakeSwap struct {
    client     *client.BSCClient
    wallet     *wallet.WalletManager
    router     *contract.ContractInteraction
    routerAddr common.Address
}

func NewPancakeSwap(client *client.BSCClient, wallet *wallet.WalletManager) (*PancakeSwap, error) {
    routerAddr := common.HexToAddress("0x10ED43C718714eb63d5aA57B78B54704E256024E")
    
    router, err := contract.NewContractInteraction(client, wallet, routerAddr, PancakeRouterABI)
    if err != nil {
        return nil, fmt.Errorf("创建路由器合约失败: %v", err)
    }

    return &PancakeSwap{
        client:     client,
        wallet:     wallet,
        router:     router,
        routerAddr: routerAddr,
    }, nil
}

// 获取交换价格
func (p *PancakeSwap) GetAmountsOut(amountIn decimal.Decimal, path []common.Address) ([]decimal.Decimal, error) {
    amountInBig := amountIn.Shift(18).BigInt() // 假设18位小数

    var result []interface{}
    err := p.router.CallMethod("getAmountsOut", &result, amountInBig, path)
    if err != nil {
        return nil, err
    }

    amounts := result[0].([]*big.Int)
    var decimalsAmounts []decimal.Decimal
    
    for _, amount := range amounts {
        decimalsAmounts = append(decimalsAmounts, decimal.NewFromBigInt(amount, -18))
    }

    return decimalsAmounts, nil
}

// 代币交换
func (p *PancakeSwap) SwapExactTokensForTokens(
    amountIn decimal.Decimal,
    amountOutMin decimal.Decimal,
    path []common.Address,
    deadline *big.Int,
) (*types.Transaction, error) {
    
    amountInBig := amountIn.Shift(18).BigInt()
    amountOutMinBig := amountOutMin.Shift(18).BigInt()

    return p.router.TransactMethod(
        "swapExactTokensForTokens",
        300000, // gas limit
        amountInBig,
        amountOutMinBig,
        path,
        p.wallet.GetAddress(),
        deadline,
    )
}

// 计算最优交换路径
func (p *PancakeSwap) FindBestPath(tokenA, tokenB common.Address, amountIn decimal.Decimal) ([]common.Address, decimal.Decimal, error) {
    // WBNB地址
    WBNB := common.HexToAddress("0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c")
    
    // 直接路径
    directPath := []common.Address{tokenA, tokenB}
    directAmounts, err := p.GetAmountsOut(amountIn, directPath)
    if err != nil {
        return nil, decimal.Zero, err
    }

    // 通过WBNB的路径
    wbnbPath := []common.Address{tokenA, WBNB, tokenB}
    wbnbAmounts, err := p.GetAmountsOut(amountIn, wbnbPath)
    if err != nil {
        return directPath, directAmounts[len(directAmounts)-1], nil
    }

    // 比较输出金额
    if wbnbAmounts[len(wbnbAmounts)-1].GreaterThan(directAmounts[len(directAmounts)-1]) {
        return wbnbPath, wbnbAmounts[len(wbnbAmounts)-1], nil
    }

    return directPath, directAmounts[len(directAmounts)-1], nil
}
```

## 实际应用

### 9.1 完整交易应用

```go
// main.go
package main

import (
    "fmt"
    "log"
    "math/big"

    "github.com/ethereum/go-ethereum/common"
    "github.com/shopspring/decimal"

    "your-project/config"
    "your-project/client"
    "your-project/wallet"
    "your-project/bnb"
    "your-project/token"
    "your-project/defi"
)

func main() {
    // 创建BSC配置
    cfg := config.DefaultBSCConfig()

    // 连接BSC测试网
    bscClient, err := client.NewBSCClient(cfg.TestnetRPC, cfg.TestChainID)
    if err != nil {
        log.Fatal("连接BSC失败:", err)
    }

    // 创建钱包（请替换为您的私钥）
    walletManager, err := wallet.NewWalletManager("your_private_key_here")
    if err != nil {
        log.Fatal("创建钱包失败:", err)
    }

    fmt.Printf("钱包地址: %s\n", walletManager.GetAddress().Hex())

    // 查询BNB余额
    bnbBalance, err := bscClient.GetBNBBalance(walletManager.GetAddress())
    if err != nil {
        log.Printf("查询BNB余额失败: %v", err)
    } else {
        bnbDecimal := decimal.NewFromBigInt(bnbBalance, -18)
        fmt.Printf("BNB余额: %s BNB\n", bnbDecimal.String())
    }

    // BNB转账示例
    bnbTransfer := bnb.NewBNBTransfer(bscClient, walletManager)
    
    // 估算转账费用
    transferAmount := decimal.NewFromFloat(0.01) // 0.01 BNB
    cost, err := bnbTransfer.EstimateTransferCost(transferAmount)
    if err != nil {
        log.Printf("估算转账费用失败: %v", err)
    } else {
        fmt.Printf("转账费用估算:\n")
        fmt.Printf("  转账金额: %s BNB\n", cost.Amount.String())
        fmt.Printf("  Gas费用: %s BNB\n", cost.GasCost.String())
        fmt.Printf("  总费用: %s BNB\n", cost.Total.String())
    }

    // BEP20代币操作示例（BUSD）
    busdAddr := common.HexToAddress("0xe9e7CEA3DedcA5984780Bafc599bD69ADd087D56")
    busdToken, err := token.NewBEP20Token(bscClient, walletManager, busdAddr)
    if err != nil {
        log.Printf("创建BUSD代币实例失败: %v", err)
    } else {
        // 查询BUSD余额
        busdBalance, err := busdToken.GetBalance(walletManager.GetAddress())
        if err != nil {
            log.Printf("查询BUSD余额失败: %v", err)
        } else {
            fmt.Printf("BUSD余额: %s BUSD\n", busdBalance.String())
        }

        // 获取代币信息
        tokenInfo := busdToken.GetTokenInfo()
        fmt.Printf("代币信息: %s (%s) - %d位小数\n", 
            tokenInfo.Name, tokenInfo.Symbol, tokenInfo.Decimals)
    }

    // PancakeSwap交换示例
    pancakeSwap, err := defi.NewPancakeSwap(bscClient, walletManager)
    if err != nil {
        log.Printf("创建PancakeSwap实例失败: %v", err)
    } else {
        // 查询BUSD到BNB的交换价格
        wbnbAddr := common.HexToAddress("0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c")
        path := []common.Address{busdAddr, wbnbAddr}
        
        swapAmount := decimal.NewFromFloat(100) // 100 BUSD
        amounts, err := pancakeSwap.GetAmountsOut(swapAmount, path)
        if err != nil {
            log.Printf("查询交换价格失败: %v", err)
        } else {
            fmt.Printf("交换价格: %s BUSD = %s BNB\n", 
                amounts[0].String(), amounts[1].String())
        }

        // 寻找最优交换路径
        bestPath, bestAmount, err := pancakeSwap.FindBestPath(busdAddr, wbnbAddr, swapAmount)
        if err != nil {
            log.Printf("寻找最优路径失败: %v", err)
        } else {
            fmt.Printf("最优路径输出: %s BNB\n", bestAmount.String())
            fmt.Printf("路径长度: %d\n", len(bestPath))
        }
    }

    fmt.Println("BSC操作演示完成!")
}
```
