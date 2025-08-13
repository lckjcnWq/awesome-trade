# Polygon SDK 详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [环境准备](#环境准备)
3. [网络连接](#网络连接)
4. [MATIC操作](#MATIC操作)
5. [ERC20代币](#ERC20代币)
6. [Layer2桥接](#Layer2桥接)
7. [DeFi集成](#DeFi集成)
8. [NFT操作](#NFT操作)
9. [实际应用](#实际应用)

## 基础概念

### 1.1 Polygon简介

Polygon是以太坊的Layer2扩容解决方案，提供更快的交易速度和更低的费用，同时保持与以太坊的兼容性。

```bash
# 安装Polygon相关依赖
go get github.com/ethereum/go-ethereum
go get github.com/maticnetwork/polygon-sdk
go get github.com/shopspring/decimal
```

### 1.2 网络配置

```go
// config/polygon.go
package config

import (
    "math/big"
)

type PolygonConfig struct {
    MainnetRPC     string
    TestnetRPC     string
    ChainID        *big.Int
    TestChainID    *big.Int
    EthereumRPC    string
    GasLimit       uint64
    GasPrice       *big.Int
    ExplorerURL    string
}

func DefaultPolygonConfig() *PolygonConfig {
    return &PolygonConfig{
        MainnetRPC:  "https://polygon-rpc.com/",
        TestnetRPC:  "https://rpc-mumbai.maticvigil.com/",
        ChainID:     big.NewInt(137),  // Polygon主网
        TestChainID: big.NewInt(80001), // Mumbai测试网
        EthereumRPC: "https://mainnet.infura.io/v3/YOUR_PROJECT_ID",
        GasLimit:    21000,
        GasPrice:    big.NewInt(30000000000), // 30 Gwei
        ExplorerURL: "https://polygonscan.com",
    }
}

// 常用合约地址
var (
    // 主网代币合约
    USDC_POLYGON = "0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174"
    USDT_POLYGON = "0xc2132D05D31c914a87C6611C10748AEb04B58e8F"
    WMATIC      = "0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270"
    
    // DeFi协议
    QUICKSWAP_ROUTER = "0xa5E0829CaCEd8fFDD4De3c43696c57F7D7A678ff"
    AAVE_POOL       = "0x8dFf5E27EA6b7AC08EbFdf9eB090F32ee9a30fcf"
    
    // 桥接合约
    POLYGON_BRIDGE = "0xA0c68C638235ee32657e8f720a23ceC1bFc77C77"
    ROOT_CHAIN_MANAGER = "0xA0c68C638235ee32657e8f720a23ceC1bFc77C77"
)
```

## 环境准备

### 2.1 多链客户端

```go
// client/polygon_client.go
package client

import (
    "context"
    "fmt"
    "math/big"

    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
)

type PolygonClient struct {
    polygonClient  *ethclient.Client
    ethereumClient *ethclient.Client
    chainID        *big.Int
    ethChainID     *big.Int
}

func NewPolygonClient(polygonRPC, ethereumRPC string, polygonChainID, ethChainID *big.Int) (*PolygonClient, error) {
    // 连接Polygon网络
    polygonClient, err := ethclient.Dial(polygonRPC)
    if err != nil {
        return nil, fmt.Errorf("连接Polygon网络失败: %v", err)
    }

    // 连接以太坊网络
    ethereumClient, err := ethclient.Dial(ethereumRPC)
    if err != nil {
        return nil, fmt.Errorf("连接以太坊网络失败: %v", err)
    }

    return &PolygonClient{
        polygonClient:  polygonClient,
        ethereumClient: ethereumClient,
        chainID:        polygonChainID,
        ethChainID:     ethChainID,
    }, nil
}

// 获取Polygon客户端
func (p *PolygonClient) GetPolygonClient() *ethclient.Client {
    return p.polygonClient
}

// 获取以太坊客户端
func (p *PolygonClient) GetEthereumClient() *ethclient.Client {
    return p.ethereumClient
}

// 获取MATIC余额
func (p *PolygonClient) GetMATICBalance(address common.Address) (*big.Int, error) {
    balance, err := p.polygonClient.BalanceAt(context.Background(), address, nil)
    if err != nil {
        return nil, err
    }
    return balance, nil
}

// 获取ETH余额
func (p *PolygonClient) GetETHBalance(address common.Address) (*big.Int, error) {
    balance, err := p.ethereumClient.BalanceAt(context.Background(), address, nil)
    if err != nil {
        return nil, err
    }
    return balance, nil
}

// 获取Polygon网络状态
func (p *PolygonClient) GetPolygonNetworkStatus() (*NetworkStatus, error) {
    header, err := p.polygonClient.HeaderByNumber(context.Background(), nil)
    if err != nil {
        return nil, err
    }

    gasPrice, err := p.polygonClient.SuggestGasPrice(context.Background())
    if err != nil {
        return nil, err
    }

    return &NetworkStatus{
        BlockNumber: header.Number.Uint64(),
        GasPrice:    gasPrice,
        Timestamp:   header.Time,
    }, nil
}

// 发送Polygon交易
func (p *PolygonClient) SendPolygonTransaction(signedTx *types.Transaction) error {
    return p.polygonClient.SendTransaction(context.Background(), signedTx)
}

// 发送以太坊交易
func (p *PolygonClient) SendEthereumTransaction(signedTx *types.Transaction) error {
    return p.ethereumClient.SendTransaction(context.Background(), signedTx)
}

type NetworkStatus struct {
    BlockNumber uint64
    GasPrice    *big.Int
    Timestamp   uint64
}
```

## 网络连接

### 3.1 网络管理器

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
    mainnetClient *client.PolygonClient
    testnetClient *client.PolygonClient
    config        *config.PolygonConfig
}

func NewNetworkManager(cfg *config.PolygonConfig) (*NetworkManager, error) {
    // 创建主网客户端
    mainnetClient, err := client.NewPolygonClient(
        cfg.MainnetRPC,
        cfg.EthereumRPC,
        cfg.ChainID,
        big.NewInt(1), // 以太坊主网
    )
    if err != nil {
        return nil, fmt.Errorf("创建主网客户端失败: %v", err)
    }

    // 创建测试网客户端
    testnetClient, err := client.NewPolygonClient(
        cfg.TestnetRPC,
        "https://goerli.infura.io/v3/YOUR_PROJECT_ID", // Goerli测试网
        cfg.TestChainID,
        big.NewInt(5), // Goerli链ID
    )
    if err != nil {
        return nil, fmt.Errorf("创建测试网客户端失败: %v", err)
    }

    return &NetworkManager{
        mainnetClient: mainnetClient,
        testnetClient: testnetClient,
        config:        cfg,
    }, nil
}

// 获取主网客户端
func (n *NetworkManager) GetMainnetClient() *client.PolygonClient {
    return n.mainnetClient
}

// 获取测试网客户端
func (n *NetworkManager) GetTestnetClient() *client.PolygonClient {
    return n.testnetClient
}

// 检查网络连通性
func (n *NetworkManager) CheckConnectivity() error {
    // 检查Polygon主网
    _, err := n.mainnetClient.GetPolygonNetworkStatus()
    if err != nil {
        return fmt.Errorf("Polygon主网连接失败: %v", err)
    }

    // 检查测试网
    _, err = n.testnetClient.GetPolygonNetworkStatus()
    if err != nil {
        return fmt.Errorf("Polygon测试网连接失败: %v", err)
    }

    return nil
}

// 比较网络费用
func (n *NetworkManager) CompareNetworkFees() (*FeeComparison, error) {
    // 获取Polygon费用
    polygonStatus, err := n.mainnetClient.GetPolygonNetworkStatus()
    if err != nil {
        return nil, err
    }

    // 获取以太坊费用
    ethHeader, err := n.mainnetClient.GetEthereumClient().HeaderByNumber(context.Background(), nil)
    if err != nil {
        return nil, err
    }

    ethGasPrice, err := n.mainnetClient.GetEthereumClient().SuggestGasPrice(context.Background())
    if err != nil {
        return nil, err
    }

    return &FeeComparison{
        PolygonGasPrice: polygonStatus.GasPrice,
        EthereumGasPrice: ethGasPrice,
        PolygonBlock:    polygonStatus.BlockNumber,
        EthereumBlock:   ethHeader.Number.Uint64(),
    }, nil
}

type FeeComparison struct {
    PolygonGasPrice  *big.Int
    EthereumGasPrice *big.Int
    PolygonBlock     uint64
    EthereumBlock    uint64
}
```

## MATIC操作

### 4.1 MATIC转账

```go
// matic/transfer.go
package matic

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

type MATICTransfer struct {
    client *client.PolygonClient
    wallet *wallet.WalletManager
}

func NewMATICTransfer(client *client.PolygonClient, wallet *wallet.WalletManager) *MATICTransfer {
    return &MATICTransfer{
        client: client,
        wallet: wallet,
    }
}

// MATIC转账
func (m *MATICTransfer) Transfer(to common.Address, amount decimal.Decimal) (*types.Transaction, error) {
    // 转换金额为Wei
    amountWei := amount.Mul(decimal.NewFromFloat(1e18)).BigInt()

    // 获取nonce
    nonce, err := m.client.GetPolygonClient().PendingNonceAt(
        context.Background(), 
        m.wallet.GetAddress(),
    )
    if err != nil {
        return nil, fmt.Errorf("获取nonce失败: %v", err)
    }

    // 获取gas价格
    gasPrice, err := m.client.GetPolygonClient().SuggestGasPrice(context.Background())
    if err != nil {
        return nil, fmt.Errorf("获取gas价格失败: %v", err)
    }

    // 创建交易
    tx := types.NewTransaction(
        nonce,
        to,
        amountWei,
        21000, // MATIC转账的gas限制
        gasPrice,
        nil,
    )

    // 签名交易
    signedTx, err := m.wallet.SignTransaction(tx, m.client.chainID)
    if err != nil {
        return nil, fmt.Errorf("签名交易失败: %v", err)
    }

    // 发送交易
    err = m.client.SendPolygonTransaction(signedTx)
    if err != nil {
        return nil, fmt.Errorf("发送交易失败: %v", err)
    }

    return signedTx, nil
}

// 快速转账（使用更高的gas价格）
func (m *MATICTransfer) FastTransfer(to common.Address, amount decimal.Decimal) (*types.Transaction, error) {
    amountWei := amount.Mul(decimal.NewFromFloat(1e18)).BigInt()

    nonce, err := m.client.GetPolygonClient().PendingNonceAt(
        context.Background(), 
        m.wallet.GetAddress(),
    )
    if err != nil {
        return nil, err
    }

    // 使用更高的gas价格以加快确认
    gasPrice, err := m.client.GetPolygonClient().SuggestGasPrice(context.Background())
    if err != nil {
        return nil, err
    }
    
    // 增加50%的gas价格
    fastGasPrice := new(big.Int).Mul(gasPrice, big.NewInt(150))
    fastGasPrice = fastGasPrice.Div(fastGasPrice, big.NewInt(100))

    tx := types.NewTransaction(
        nonce,
        to,
        amountWei,
        21000,
        fastGasPrice,
        nil,
    )

    signedTx, err := m.wallet.SignTransaction(tx, m.client.chainID)
    if err != nil {
        return nil, err
    }

    err = m.client.SendPolygonTransaction(signedTx)
    if err != nil {
        return nil, err
    }

    return signedTx, nil
}

// 估算转账费用
func (m *MATICTransfer) EstimateTransferCost(amount decimal.Decimal) (*TransferCost, error) {
    gasPrice, err := m.client.GetPolygonClient().SuggestGasPrice(context.Background())
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

// 批量转账
func (m *MATICTransfer) BatchTransfer(transfers []TransferData) ([]*types.Transaction, error) {
    var transactions []*types.Transaction
    
    // 获取起始nonce
    nonce, err := m.client.GetPolygonClient().PendingNonceAt(
        context.Background(), 
        m.wallet.GetAddress(),
    )
    if err != nil {
        return nil, err
    }

    gasPrice, err := m.client.GetPolygonClient().SuggestGasPrice(context.Background())
    if err != nil {
        return nil, err
    }

    for i, transfer := range transfers {
        amountWei := transfer.Amount.Mul(decimal.NewFromFloat(1e18)).BigInt()

        tx := types.NewTransaction(
            nonce+uint64(i),
            transfer.To,
            amountWei,
            21000,
            gasPrice,
            nil,
        )

        signedTx, err := m.wallet.SignTransaction(tx, m.client.chainID)
        if err != nil {
            return nil, err
        }

        err = m.client.SendPolygonTransaction(signedTx)
        if err != nil {
            return nil, err
        }

        transactions = append(transactions, signedTx)
    }

    return transactions, nil
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

## ERC20代币

### 5.1 Polygon代币操作

```go
// token/polygon_erc20.go
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

type PolygonERC20 struct {
    client       *client.PolygonClient
    wallet       *wallet.WalletManager
    contract     *bind.BoundContract
    contractAddr common.Address
    abi          abi.ABI
    decimals     uint8
    symbol       string
    name         string
}

func NewPolygonERC20(client *client.PolygonClient, wallet *wallet.WalletManager, contractAddr common.Address) (*PolygonERC20, error) {
    parsedABI, err := abi.JSON(strings.NewReader(ERC20ABI))
    if err != nil {
        return nil, fmt.Errorf("解析ABI失败: %v", err)
    }

    contract := bind.NewBoundContract(
        contractAddr, 
        parsedABI, 
        client.GetPolygonClient(), 
        client.GetPolygonClient(), 
        client.GetPolygonClient(),
    )

    token := &PolygonERC20{
        client:       client,
        wallet:       wallet,
        contract:     contract,
        contractAddr: contractAddr,
        abi:          parsedABI,
    }

    if err := token.loadTokenInfo(); err != nil {
        return nil, fmt.Errorf("加载代币信息失败: %v", err)
    }

    return token, nil
}

// 加载代币信息
func (t *PolygonERC20) loadTokenInfo() error {
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
func (t *PolygonERC20) GetBalance(address common.Address) (decimal.Decimal, error) {
    var result []interface{}
    err := t.contract.Call(nil, &result, "balanceOf", address)
    if err != nil {
        return decimal.Zero, err
    }

    balance := result[0].(*big.Int)
    return decimal.NewFromBigInt(balance, -int32(t.decimals)), nil
}

// 转账
func (t *PolygonERC20) Transfer(to common.Address, amount decimal.Decimal) (*types.Transaction, error) {
    amountBig := amount.Shift(int32(t.decimals)).BigInt()

    nonce, err := t.client.GetPolygonClient().PendingNonceAt(
        context.Background(), 
        t.wallet.GetAddress(),
    )
    if err != nil {
        return nil, err
    }

    gasPrice, err := t.client.GetPolygonClient().SuggestGasPrice(context.Background())
    if err != nil {
        return nil, err
    }

    auth := &bind.TransactOpts{
        From:     t.wallet.GetAddress(),
        Nonce:    big.NewInt(int64(nonce)),
        GasLimit: 65000, // ERC20转账的gas限制
        GasPrice: gasPrice,
        Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
            return t.wallet.SignTransaction(tx, t.client.chainID)
        },
    }

    tx, err := t.contract.Transact(auth, "transfer", to, amountBig)
    if err != nil {
        return nil, err
    }

    return tx, nil
}

// 快速转账（更高gas费）
func (t *PolygonERC20) FastTransfer(to common.Address, amount decimal.Decimal) (*types.Transaction, error) {
    amountBig := amount.Shift(int32(t.decimals)).BigInt()

    nonce, err := t.client.GetPolygonClient().PendingNonceAt(
        context.Background(), 
        t.wallet.GetAddress(),
    )
    if err != nil {
        return nil, err
    }

    gasPrice, err := t.client.GetPolygonClient().SuggestGasPrice(context.Background())
    if err != nil {
        return nil, err
    }

    // 增加gas价格
    fastGasPrice := new(big.Int).Mul(gasPrice, big.NewInt(150))
    fastGasPrice = fastGasPrice.Div(fastGasPrice, big.NewInt(100))

    auth := &bind.TransactOpts{
        From:     t.wallet.GetAddress(),
        Nonce:    big.NewInt(int64(nonce)),
        GasLimit: 65000,
        GasPrice: fastGasPrice,
        Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
            return t.wallet.SignTransaction(tx, t.client.chainID)
        },
    }

    tx, err := t.contract.Transact(auth, "transfer", to, amountBig)
    if err != nil {
        return nil, err
    }

    return tx, nil
}

// 授权
func (t *PolygonERC20) Approve(spender common.Address, amount decimal.Decimal) (*types.Transaction, error) {
    amountBig := amount.Shift(int32(t.decimals)).BigInt()

    nonce, err := t.client.GetPolygonClient().PendingNonceAt(
        context.Background(), 
        t.wallet.GetAddress(),
    )
    if err != nil {
        return nil, err
    }

    gasPrice, err := t.client.GetPolygonClient().SuggestGasPrice(context.Background())
    if err != nil {
        return nil, err
    }

    auth := &bind.TransactOpts{
        From:     t.wallet.GetAddress(),
        Nonce:    big.NewInt(int64(nonce)),
        GasLimit: 50000,
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

// 获取代币价格（通过DEX）
func (t *PolygonERC20) GetTokenPrice() (decimal.Decimal, error) {
    // 这里可以集成QuickSwap或其他DEX来获取价格
    // 简化实现，返回模拟价格
    return decimal.NewFromFloat(1.0), nil
}

// 获取代币信息
func (t *PolygonERC20) GetTokenInfo() TokenInfo {
    return TokenInfo{
        Name:     t.name,
        Symbol:   t.symbol,
        Decimals: t.decimals,
        Address:  t.contractAddr.Hex(),
        Network:  "Polygon",
    }
}

type TokenInfo struct {
    Name     string
    Symbol   string
    Decimals uint8
    Address  string
    Network  string
}
```

## Layer2桥接

### 6.1 以太坊-Polygon桥接

```go
// bridge/polygon_bridge.go
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
    "github.com/shopspring/decimal"

    "your-project/client"
    "your-project/wallet"
)

// 桥接合约ABI（简化版）
const BridgeABI = `[
    {
        "inputs": [
            {"name": "token", "type": "address"},
            {"name": "amount", "type": "uint256"}
        ],
        "name": "depositFor",
        "outputs": [],
        "stateMutability": "nonpayable",
        "type": "function"
    },
    {
        "inputs": [
            {"name": "token", "type": "address"},
            {"name": "amount", "type": "uint256"}
        ],
        "name": "withdraw",
        "outputs": [],
        "stateMutability": "nonpayable",
        "type": "function"
    }
]`

type PolygonBridge struct {
    client         *client.PolygonClient
    wallet         *wallet.WalletManager
    bridgeContract *bind.BoundContract
    bridgeAddr     common.Address
}

func NewPolygonBridge(client *client.PolygonClient, wallet *wallet.WalletManager) (*PolygonBridge, error) {
    bridgeAddr := common.HexToAddress("0xA0c68C638235ee32657e8f720a23ceC1bFc77C77")
    
    parsedABI, err := abi.JSON(strings.NewReader(BridgeABI))
    if err != nil {
        return nil, fmt.Errorf("解析桥接ABI失败: %v", err)
    }

    bridgeContract := bind.NewBoundContract(
        bridgeAddr, 
        parsedABI, 
        client.GetEthereumClient(), 
        client.GetEthereumClient(), 
        client.GetEthereumClient(),
    )

    return &PolygonBridge{
        client:         client,
        wallet:         wallet,
        bridgeContract: bridgeContract,
        bridgeAddr:     bridgeAddr,
    }, nil
}

// 从以太坊存款到Polygon
func (b *PolygonBridge) DepositToPolygon(tokenAddr common.Address, amount decimal.Decimal) (*types.Transaction, error) {
    amountBig := amount.Shift(18).BigInt() // 假设18位小数

    // 获取以太坊网络的nonce
    nonce, err := b.client.GetEthereumClient().PendingNonceAt(
        context.Background(), 
        b.wallet.GetAddress(),
    )
    if err != nil {
        return nil, fmt.Errorf("获取nonce失败: %v", err)
    }

    // 获取以太坊gas价格
    gasPrice, err := b.client.GetEthereumClient().SuggestGasPrice(context.Background())
    if err != nil {
        return nil, fmt.Errorf("获取gas价格失败: %v", err)
    }

    auth := &bind.TransactOpts{
        From:     b.wallet.GetAddress(),
        Nonce:    big.NewInt(int64(nonce)),
        GasLimit: 200000, // 桥接操作需要更多gas
        GasPrice: gasPrice,
        Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
            return b.wallet.SignTransaction(tx, b.client.ethChainID)
        },
    }

    // 调用存款方法
    tx, err := b.bridgeContract.Transact(auth, "depositFor", tokenAddr, amountBig)
    if err != nil {
        return nil, fmt.Errorf("存款交易失败: %v", err)
    }

    return tx, nil
}

// 从Polygon提取到以太坊
func (b *PolygonBridge) WithdrawToEthereum(tokenAddr common.Address, amount decimal.Decimal) (*types.Transaction, error) {
    amountBig := amount.Shift(18).BigInt()

    // 获取Polygon网络的nonce
    nonce, err := b.client.GetPolygonClient().PendingNonceAt(
        context.Background(), 
        b.wallet.GetAddress(),
    )
    if err != nil {
        return nil, fmt.Errorf("获取nonce失败: %v", err)
    }

    gasPrice, err := b.client.GetPolygonClient().SuggestGasPrice(context.Background())
    if err != nil {
        return nil, fmt.Errorf("获取gas价格失败: %v", err)
    }

    // 创建Polygon网络的合约实例
    polygonBridgeContract := bind.NewBoundContract(
        b.bridgeAddr, 
        b.bridgeContract.Abi, 
        b.client.GetPolygonClient(), 
        b.client.GetPolygonClient(), 
        b.client.GetPolygonClient(),
    )

    auth := &bind.TransactOpts{
        From:     b.wallet.GetAddress(),
        Nonce:    big.NewInt(int64(nonce)),
        GasLimit: 150000,
        GasPrice: gasPrice,
        Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
            return b.wallet.SignTransaction(tx, b.client.chainID)
        },
    }

    tx, err := polygonBridgeContract.Transact(auth, "withdraw", tokenAddr, amountBig)
    if err != nil {
        return nil, fmt.Errorf("提取交易失败: %v", err)
    }

    return tx, nil
}

// 估算桥接费用
func (b *PolygonBridge) EstimateBridgeCost(direction BridgeDirection, amount decimal.Decimal) (*BridgeCost, error) {
    var gasPrice *big.Int
    var err error
    var gasLimit uint64

    switch direction {
    case EthereumToPolygon:
        gasPrice, err = b.client.GetEthereumClient().SuggestGasPrice(context.Background())
        gasLimit = 200000
    case PolygonToEthereum:
        gasPrice, err = b.client.GetPolygonClient().SuggestGasPrice(context.Background())
        gasLimit = 150000
    default:
        return nil, fmt.Errorf("无效的桥接方向")
    }

    if err != nil {
        return nil, err
    }

    gasCost := new(big.Int).Mul(gasPrice, big.NewInt(int64(gasLimit)))

    return &BridgeCost{
        Direction: direction,
        Amount:    amount,
        GasCost:   decimal.NewFromBigInt(gasCost, -18),
        GasPrice:  decimal.NewFromBigInt(gasPrice, -9),
        GasLimit:  gasLimit,
    }, nil
}

// 检查桥接状态
func (b *PolygonBridge) CheckBridgeStatus(txHash common.Hash, direction BridgeDirection) (*BridgeStatus, error) {
    var receipt *types.Receipt
    var err error

    switch direction {
    case EthereumToPolygon:
        receipt, err = b.client.GetEthereumClient().TransactionReceipt(context.Background(), txHash)
    case PolygonToEthereum:
        receipt, err = b.client.GetPolygonClient().TransactionReceipt(context.Background(), txHash)
    default:
        return nil, fmt.Errorf("无效的桥接方向")
    }

    if err != nil {
        return &BridgeStatus{
            TxHash:    txHash,
            Status:    "pending",
            Confirmed: false,
        }, nil
    }

    status := "failed"
    if receipt.Status == 1 {
        status = "success"
    }

    return &BridgeStatus{
        TxHash:      txHash,
        Status:      status,
        Confirmed:   true,
        BlockNumber: receipt.BlockNumber.Uint64(),
        GasUsed:     receipt.GasUsed,
    }, nil
}

type BridgeDirection int

const (
    EthereumToPolygon BridgeDirection = iota
    PolygonToEthereum
)

type BridgeCost struct {
    Direction BridgeDirection
    Amount    decimal.Decimal
    GasCost   decimal.Decimal
    GasPrice  decimal.Decimal
    GasLimit  uint64
}

type BridgeStatus struct {
    TxHash      common.Hash
    Status      string
    Confirmed   bool
    BlockNumber uint64
    GasUsed     uint64
}
```

## DeFi集成

### 7.1 QuickSwap集成

```go
// defi/quickswap.go
package defi

import (
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

// QuickSwap路由器ABI（简化版）
const QuickSwapRouterABI = `[
    {
        "inputs": [
            {"name": "amountIn", "type": "uint256"},
            {"name": "path", "type": "address[]"}
        ],
        "name": "getAmountsOut",
        "outputs": [
            {"name": "amounts", "type": "uint256[]"}
        ],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [
            {"name": "amountIn", "type": "uint256"},
            {"name": "amountOutMin", "type": "uint256"},
            {"name": "path", "type": "address[]"},
            {"name": "to", "type": "address"},
            {"name": "deadline", "type": "uint256"}
        ],
        "name": "swapExactTokensForTokens",
        "outputs": [
            {"name": "amounts", "type": "uint256[]"}
        ],
        "stateMutability": "nonpayable",
        "type": "function"
    }
]`

type QuickSwap struct {
    client     *client.PolygonClient
    wallet     *wallet.WalletManager
    router     *bind.BoundContract
    routerAddr common.Address
}

func NewQuickSwap(client *client.PolygonClient, wallet *wallet.WalletManager) (*QuickSwap, error) {
    routerAddr := common.HexToAddress("0xa5E0829CaCEd8fFDD4De3c43696c57F7D7A678ff")
    
    parsedABI, err := abi.JSON(strings.NewReader(QuickSwapRouterABI))
    if err != nil {
        return nil, fmt.Errorf("解析QuickSwap ABI失败: %v", err)
    }

    router := bind.NewBoundContract(
        routerAddr, 
        parsedABI, 
        client.GetPolygonClient(), 
        client.GetPolygonClient(), 
        client.GetPolygonClient(),
    )

    return &QuickSwap{
        client:     client,
        wallet:     wallet,
        router:     router,
        routerAddr: routerAddr,
    }, nil
}

// 获取交换价格
func (q *QuickSwap) GetAmountsOut(amountIn decimal.Decimal, path []common.Address) ([]decimal.Decimal, error) {
    amountInBig := amountIn.Shift(18).BigInt()

    var result []interface{}
    err := q.router.Call(nil, &result, "getAmountsOut", amountInBig, path)
    if err != nil {
        return nil, err
    }

    amounts := result[0].([]*big.Int)
    var decimalAmounts []decimal.Decimal
    
    for _, amount := range amounts {
        decimalAmounts = append(decimalAmounts, decimal.NewFromBigInt(amount, -18))
    }

    return decimalAmounts, nil
}

// 代币交换
func (q *QuickSwap) SwapExactTokensForTokens(
    amountIn decimal.Decimal,
    amountOutMin decimal.Decimal,
    path []common.Address,
    deadline *big.Int,
) (*types.Transaction, error) {
    
    amountInBig := amountIn.Shift(18).BigInt()
    amountOutMinBig := amountOutMin.Shift(18).BigInt()

    nonce, err := q.client.GetPolygonClient().PendingNonceAt(
        context.Background(), 
        q.wallet.GetAddress(),
    )
    if err != nil {
        return nil, err
    }

    gasPrice, err := q.client.GetPolygonClient().SuggestGasPrice(context.Background())
    if err != nil {
        return nil, err
    }

    auth := &bind.TransactOpts{
        From:     q.wallet.GetAddress(),
        Nonce:    big.NewInt(int64(nonce)),
        GasLimit: 300000,
        GasPrice: gasPrice,
        Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
            return q.wallet.SignTransaction(tx, q.client.chainID)
        },
    }

    return q.router.Transact(
        auth,
        "swapExactTokensForTokens",
        amountInBig,
        amountOutMinBig,
        path,
        q.wallet.GetAddress(),
        deadline,
    )
}

// 计算最优交换路径
func (q *QuickSwap) FindBestPath(tokenA, tokenB common.Address, amountIn decimal.Decimal) ([]common.Address, decimal.Decimal, error) {
    // WMATIC地址
    WMATIC := common.HexToAddress("0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270")
    
    // 直接路径
    directPath := []common.Address{tokenA, tokenB}
    directAmounts, err := q.GetAmountsOut(amountIn, directPath)
    if err != nil {
        return nil, decimal.Zero, err
    }

    // 通过WMATIC的路径
    wmaticPath := []common.Address{tokenA, WMATIC, tokenB}
    wmaticAmounts, err := q.GetAmountsOut(amountIn, wmaticPath)
    if err != nil {
        return directPath, directAmounts[len(directAmounts)-1], nil
    }

    // 比较输出金额
    if wmaticAmounts[len(wmaticAmounts)-1].GreaterThan(directAmounts[len(directAmounts)-1]) {
        return wmaticPath, wmaticAmounts[len(wmaticAmounts)-1], nil
    }

    return directPath, directAmounts[len(directAmounts)-1], nil
}

// 获取流动性池信息
func (q *QuickSwap) GetPoolInfo(tokenA, tokenB common.Address) (*PoolInfo, error) {
    // 这里可以调用QuickSwap工厂合约获取池信息
    // 简化实现
    return &PoolInfo{
        TokenA:    tokenA,
        TokenB:    tokenB,
        ReserveA:  decimal.NewFromFloat(1000000),
        ReserveB:  decimal.NewFromFloat(2000000),
        Fee:       decimal.NewFromFloat(0.003), // 0.3%
    }, nil
}

type PoolInfo struct {
    TokenA   common.Address
    TokenB   common.Address
    ReserveA decimal.Decimal
    ReserveB decimal.Decimal
    Fee      decimal.Decimal
}
```

## NFT操作

### 8.1 Polygon NFT

```go
// nft/polygon_nft.go
package nft

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

// ERC721 ABI（简化版）
const ERC721ABI = `[
    {
        "inputs": [{"name": "owner", "type": "address"}],
        "name": "balanceOf",
        "outputs": [{"name": "", "type": "uint256"}],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [{"name": "tokenId", "type": "uint256"}],
        "name": "ownerOf",
        "outputs": [{"name": "", "type": "address"}],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [
            {"name": "from", "type": "address"},
            {"name": "to", "type": "address"},
            {"name": "tokenId", "type": "uint256"}
        ],
        "name": "transferFrom",
        "outputs": [],
        "stateMutability": "nonpayable",
        "type": "function"
    },
    {
        "inputs": [{"name": "tokenId", "type": "uint256"}],
        "name": "tokenURI",
        "outputs": [{"name": "", "type": "string"}],
        "stateMutability": "view",
        "type": "function"
    }
]`

type PolygonNFT struct {
    client       *client.PolygonClient
    wallet       *wallet.WalletManager
    contract     *bind.BoundContract
    contractAddr common.Address
}

func NewPolygonNFT(client *client.PolygonClient, wallet *wallet.WalletManager, contractAddr common.Address) (*PolygonNFT, error) {
    parsedABI, err := abi.JSON(strings.NewReader(ERC721ABI))
    if err != nil {
        return nil, fmt.Errorf("解析ERC721 ABI失败: %v", err)
    }

    contract := bind.NewBoundContract(
        contractAddr, 
        parsedABI, 
        client.GetPolygonClient(), 
        client.GetPolygonClient(), 
        client.GetPolygonClient(),
    )

    return &PolygonNFT{
        client:       client,
        wallet:       wallet,
        contract:     contract,
        contractAddr: contractAddr,
    }, nil
}

// 获取NFT余额
func (n *PolygonNFT) GetBalance(owner common.Address) (*big.Int, error) {
    var result []interface{}
    err := n.contract.Call(nil, &result, "balanceOf", owner)
    if err != nil {
        return nil, err
    }

    return result[0].(*big.Int), nil
}

// 获取NFT所有者
func (n *PolygonNFT) GetOwner(tokenId *big.Int) (common.Address, error) {
    var result []interface{}
    err := n.contract.Call(nil, &result, "ownerOf", tokenId)
    if err != nil {
        return common.Address{}, err
    }

    return result[0].(common.Address), nil
}

// 转移NFT
func (n *PolygonNFT) TransferFrom(from, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
    nonce, err := n.client.GetPolygonClient().PendingNonceAt(
        context.Background(), 
        n.wallet.GetAddress(),
    )
    if err != nil {
        return nil, err
    }

    gasPrice, err := n.client.GetPolygonClient().SuggestGasPrice(context.Background())
    if err != nil {
        return nil, err
    }

    auth := &bind.TransactOpts{
        From:     n.wallet.GetAddress(),
        Nonce:    big.NewInt(int64(nonce)),
        GasLimit: 100000,
        GasPrice: gasPrice,
        Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
            return n.wallet.SignTransaction(tx, n.client.chainID)
        },
    }

    return n.contract.Transact(auth, "transferFrom", from, to, tokenId)
}

// 获取NFT元数据URI
func (n *PolygonNFT) GetTokenURI(tokenId *big.Int) (string, error) {
    var result []interface{}
    err := n.contract.Call(nil, &result, "tokenURI", tokenId)
    if err != nil {
        return "", err
    }

    return result[0].(string), nil
}

// 获取NFT详细信息
func (n *PolygonNFT) GetNFTInfo(tokenId *big.Int) (*NFTInfo, error) {
    owner, err := n.GetOwner(tokenId)
    if err != nil {
        return nil, err
    }

    tokenURI, err := n.GetTokenURI(tokenId)
    if err != nil {
        return nil, err
    }

    return &NFTInfo{
        TokenId:     tokenId,
        Owner:       owner,
        TokenURI:    tokenURI,
        Contract:    n.contractAddr,
        Network:     "Polygon",
    }, nil
}

type NFTInfo struct {
    TokenId  *big.Int
    Owner    common.Address
    TokenURI string
    Contract common.Address
    Network  string
}
```

## 实际应用

### 9.1 完整Polygon应用

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
    "your-project/network"
    "your-project/wallet"
    "your-project/matic"
    "your-project/token"
    "your-project/defi"
    "your-project/bridge"
)

func main() {
    // 创建Polygon配置
    cfg := config.DefaultPolygonConfig()

    // 创建网络管理器
    networkManager, err := network.NewNetworkManager(cfg)
    if err != nil {
        log.Fatal("创建网络管理器失败:", err)
    }

    // 检查网络连通性
    err = networkManager.CheckConnectivity()
    if err != nil {
        log.Printf("网络连接检查失败: %v", err)
    } else {
        fmt.Println("网络连接正常")
    }

    // 创建钱包
    walletManager, err := wallet.NewWalletManager("your_private_key_here")
    if err != nil {
        log.Fatal("创建钱包失败:", err)
    }

    fmt.Printf("钱包地址: %s\n", walletManager.GetAddress().Hex())

    // 使用测试网客户端
    testnetClient := networkManager.GetTestnetClient()

    // 查询MATIC余额
    maticBalance, err := testnetClient.GetMATICBalance(walletManager.GetAddress())
    if err != nil {
        log.Printf("查询MATIC余额失败: %v", err)
    } else {
        maticDecimal := decimal.NewFromBigInt(maticBalance, -18)
        fmt.Printf("MATIC余额: %s MATIC\n", maticDecimal.String())
    }

    // MATIC转账示例
    maticTransfer := matic.NewMATICTransfer(testnetClient, walletManager)
    
    transferAmount := decimal.NewFromFloat(0.01)
    cost, err := maticTransfer.EstimateTransferCost(transferAmount)
    if err != nil {
        log.Printf("估算转账费用失败: %v", err)
    } else {
        fmt.Printf("MATIC转账费用估算:\n")
        fmt.Printf("  转账金额: %s MATIC\n", cost.Amount.String())
        fmt.Printf("  Gas费用: %s MATIC\n", cost.GasCost.String())
        fmt.Printf("  总费用: %s MATIC\n", cost.Total.String())
    }

    // ERC20代币操作示例（USDC）
    usdcAddr := common.HexToAddress("0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174")
    usdcToken, err := token.NewPolygonERC20(testnetClient, walletManager, usdcAddr)
    if err != nil {
        log.Printf("创建USDC代币实例失败: %v", err)
    } else {
        usdcBalance, err := usdcToken.GetBalance(walletManager.GetAddress())
        if err != nil {
            log.Printf("查询USDC余额失败: %v", err)
        } else {
            fmt.Printf("USDC余额: %s USDC\n", usdcBalance.String())
        }

        tokenInfo := usdcToken.GetTokenInfo()
        fmt.Printf("代币信息: %s (%s) - %d位小数 - %s网络\n", 
            tokenInfo.Name, tokenInfo.Symbol, tokenInfo.Decimals, tokenInfo.Network)
    }

    // QuickSwap交换示例
    quickSwap, err := defi.NewQuickSwap(testnetClient, walletManager)
    if err != nil {
        log.Printf("创建QuickSwap实例失败: %v", err)
    } else {
        wmaticAddr := common.HexToAddress("0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270")
        path := []common.Address{usdcAddr, wmaticAddr}
        
        swapAmount := decimal.NewFromFloat(100)
        amounts, err := quickSwap.GetAmountsOut(swapAmount, path)
        if err != nil {
            log.Printf("查询交换价格失败: %v", err)
        } else {
            fmt.Printf("QuickSwap价格: %s USDC = %s WMATIC\n", 
                amounts[0].String(), amounts[1].String())
        }

        // 寻找最优路径
        bestPath, bestAmount, err := quickSwap.FindBestPath(usdcAddr, wmaticAddr, swapAmount)
        if err != nil {
            log.Printf("寻找最优路径失败: %v", err)
        } else {
            fmt.Printf("最优路径输出: %s WMATIC\n", bestAmount.String())
            fmt.Printf("路径长度: %d\n", len(bestPath))
        }

        // 获取流动性池信息
        poolInfo, err := quickSwap.GetPoolInfo(usdcAddr, wmaticAddr)
        if err != nil {
            log.Printf("获取池信息失败: %v", err)
        } else {
            fmt.Printf("流动性池信息:\n")
            fmt.Printf("  储备A: %s\n", poolInfo.ReserveA.String())
            fmt.Printf("  储备B: %s\n", poolInfo.ReserveB.String())
            fmt.Printf("  手续费: %s%%\n", poolInfo.Fee.Mul(decimal.NewFromInt(100)).String())
        }
    }

    // 桥接操作示例
    polygonBridge, err := bridge.NewPolygonBridge(testnetClient, walletManager)
    if err != nil {
        log.Printf("创建桥接实例失败: %v", err)
    } else {
        // 估算桥接费用
        bridgeAmount := decimal.NewFromFloat(10)
        
        ethToPolygonCost, err := polygonBridge.EstimateBridgeCost(
            bridge.EthereumToPolygon, 
            bridgeAmount,
        )
        if err != nil {
            log.Printf("估算以太坊到Polygon桥接费用失败: %v", err)
        } else {
            fmt.Printf("以太坊到Polygon桥接费用:\n")
            fmt.Printf("  金额: %s\n", ethToPolygonCost.Amount.String())
            fmt.Printf("  Gas费用: %s ETH\n", ethToPolygonCost.GasCost.String())
        }

        polygonToEthCost, err := polygonBridge.EstimateBridgeCost(
            bridge.PolygonToEthereum, 
            bridgeAmount,
        )
        if err != nil {
            log.Printf("估算Polygon到以太坊桥接费用失败: %v", err)
        } else {
            fmt.Printf("Polygon到以太坊桥接费用:\n")
            fmt.Printf("  金额: %s\n", polygonToEthCost.Amount.String())
            fmt.Printf("  Gas费用: %s MATIC\n", polygonToEthCost.GasCost.String())
        }
    }

    // 比较网络费用
    feeComparison, err := networkManager.CompareNetworkFees()
    if err != nil {
        log.Printf("比较网络费用失败: %v", err)
    } else {
        fmt.Printf("网络费用比较:\n")
        fmt.Printf("  Polygon Gas价格: %s Gwei\n", 
            decimal.NewFromBigInt(feeComparison.PolygonGasPrice, -9).String())
        fmt.Printf("  以太坊Gas价格: %s Gwei\n", 
            decimal.NewFromBigInt(feeComparison.EthereumGasPrice, -9).String())
        
        ratio := decimal.NewFromBigInt(feeComparison.EthereumGasPrice, 0).
                Div(decimal.NewFromBigInt(feeComparison.PolygonGasPrice, 0))
        fmt.Printf("  费用比率 (ETH/Polygon): %s倍\n", ratio.String())
    }

    fmt.Println("Polygon操作演示完成!")
}
```
