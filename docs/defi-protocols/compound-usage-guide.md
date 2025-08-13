# Compound 协议 详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [环境准备](#环境准备)
3. [合约交互](#合约交互)
4. [资产供应](#资产供应)
5. [资产借贷](#资产借贷)
6. [利率计算](#利率计算)
7. [清算机制](#清算机制)
8. [治理代币](#治理代币)
9. [实际应用](#实际应用)

## 基础概念

### 1.1 Compound 协议简介

Compound 是以太坊上的去中心化借贷协议，用户可以供应资产赚取利息，或抵押资产借出其他代币。

```bash
# 安装Compound相关依赖
go get github.com/ethereum/go-ethereum
go get github.com/ethereum/go-ethereum/accounts/abi
go get github.com/ethereum/go-ethereum/accounts/abi/bind
go get github.com/shopspring/decimal
```

### 1.2 核心概念

```go
// 主要包导入
import (
    "context"
    "math/big"
    "strings"
    
    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/shopspring/decimal"
)

// Compound 合约地址 (Mainnet)
var (
    ComptrollerAddress = common.HexToAddress("0x3d9819210A31b4961b30EF54bE2aeD79B9c9Cd3B")
    CETHAddress        = common.HexToAddress("0x4Ddc2D193948926D02f9B1fE9e1daa0718270ED5")
    CDAIAddress        = common.HexToAddress("0x5d3a536E4D6DbD6114cc1Ead35777bAB948E3643")
    CUSDCAddress       = common.HexToAddress("0x39AA39c021dfbaE8faC545936693aC917d5E7563")
    COMPAddress        = common.HexToAddress("0xc00e94Cb662C3520282E6f5717214004A7f26888")
)

// cToken 接口
type CToken struct {
    Address     common.Address
    Symbol      string
    Decimals    uint8
    Underlying  common.Address
}

// 市场信息
type Market struct {
    CToken              CToken
    CollateralFactor    *big.Int
    LiquidationThreshold *big.Int
    SupplyAPY           decimal.Decimal
    BorrowAPY           decimal.Decimal
    TotalSupply         *big.Int
    TotalBorrow         *big.Int
}
```

## 环境准备

### 2.1 合约ABI定义

```go
// contracts/compound_abi.go
package contracts

// Comptroller ABI (简化版)
const ComptrollerABI = `[
    {
        "constant": true,
        "inputs": [{"name": "account", "type": "address"}],
        "name": "getAccountLiquidity",
        "outputs": [
            {"name": "", "type": "uint256"},
            {"name": "", "type": "uint256"},
            {"name": "", "type": "uint256"}
        ],
        "type": "function"
    },
    {
        "constant": false,
        "inputs": [{"name": "cTokens", "type": "address[]"}],
        "name": "enterMarkets",
        "outputs": [{"name": "", "type": "uint256[]"}],
        "type": "function"
    },
    {
        "constant": true,
        "inputs": [{"name": "cToken", "type": "address"}],
        "name": "markets",
        "outputs": [
            {"name": "isListed", "type": "bool"},
            {"name": "collateralFactorMantissa", "type": "uint256"}
        ],
        "type": "function"
    }
]`

// CToken ABI (简化版)
const CTokenABI = `[
    {
        "constant": false,
        "inputs": [],
        "name": "mint",
        "outputs": [{"name": "", "type": "uint256"}],
        "payable": true,
        "type": "function"
    },
    {
        "constant": false,
        "inputs": [{"name": "mintAmount", "type": "uint256"}],
        "name": "mint",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    },
    {
        "constant": false,
        "inputs": [{"name": "redeemTokens", "type": "uint256"}],
        "name": "redeem",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    },
    {
        "constant": false,
        "inputs": [{"name": "borrowAmount", "type": "uint256"}],
        "name": "borrow",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    },
    {
        "constant": false,
        "inputs": [{"name": "repayAmount", "type": "uint256"}],
        "name": "repayBorrow",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    },
    {
        "constant": true,
        "inputs": [{"name": "account", "type": "address"}],
        "name": "balanceOf",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    },
    {
        "constant": true,
        "inputs": [{"name": "account", "type": "address"}],
        "name": "borrowBalanceStored",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    },
    {
        "constant": true,
        "inputs": [],
        "name": "exchangeRateStored",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    },
    {
        "constant": true,
        "inputs": [],
        "name": "supplyRatePerBlock",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    },
    {
        "constant": true,
        "inputs": [],
        "name": "borrowRatePerBlock",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    }
]`
```

### 2.2 客户端设置

```go
// client/compound_client.go
package client

import (
    "context"
    "math/big"
    
    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
)

type CompoundClient struct {
    ethClient        *ethclient.Client
    comptrollerABI   abi.ABI
    cTokenABI        abi.ABI
    comptrollerAddr  common.Address
}

func NewCompoundClient(rpcURL string) (*CompoundClient, error) {
    ethClient, err := ethclient.Dial(rpcURL)
    if err != nil {
        return nil, err
    }
    
    comptrollerABI, err := abi.JSON(strings.NewReader(ComptrollerABI))
    if err != nil {
        return nil, err
    }
    
    cTokenABI, err := abi.JSON(strings.NewReader(CTokenABI))
    if err != nil {
        return nil, err
    }
    
    return &CompoundClient{
        ethClient:       ethClient,
        comptrollerABI:  comptrollerABI,
        cTokenABI:       cTokenABI,
        comptrollerAddr: ComptrollerAddress,
    }, nil
}

// 获取账户流动性
func (c *CompoundClient) GetAccountLiquidity(account common.Address) (*big.Int, *big.Int, error) {
    callOpts := &bind.CallOpts{
        Context: context.Background(),
    }
    
    var result []interface{}
    err := c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To:   &c.comptrollerAddr,
        Data: c.comptrollerABI.Methods["getAccountLiquidity"].ID,
    }, nil)
    
    if err != nil {
        return nil, nil, err
    }
    
    err = c.comptrollerABI.UnpackIntoInterface(&result, "getAccountLiquidity", result)
    if err != nil {
        return nil, nil, err
    }
    
    liquidity := result[1].(*big.Int)
    shortfall := result[2].(*big.Int)
    
    return liquidity, shortfall, nil
}
```

## 合约交互

### 3.1 市场信息查询

```go
// services/market_service.go
package services

import (
    "context"
    "math/big"
    
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/shopspring/decimal"
)

type MarketService struct {
    client *CompoundClient
}

func NewMarketService(client *CompoundClient) *MarketService {
    return &MarketService{
        client: client,
    }
}

// 获取市场信息
func (s *MarketService) GetMarketInfo(cTokenAddr common.Address) (*Market, error) {
    // 获取抵押因子
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    var marketInfo struct {
        IsListed              bool
        CollateralFactorMantissa *big.Int
    }
    
    err := s.client.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &s.client.comptrollerAddr,
        Data: s.client.comptrollerABI.Methods["markets"].ID,
    }, nil)
    
    if err != nil {
        return nil, err
    }
    
    // 获取供应利率
    supplyRate, err := s.getSupplyRate(cTokenAddr)
    if err != nil {
        return nil, err
    }
    
    // 获取借贷利率
    borrowRate, err := s.getBorrowRate(cTokenAddr)
    if err != nil {
        return nil, err
    }
    
    // 计算年化利率
    blocksPerYear := big.NewInt(2102400) // 以太坊每年大约区块数
    supplyAPY := s.calculateAPY(supplyRate, blocksPerYear)
    borrowAPY := s.calculateAPY(borrowRate, blocksPerYear)
    
    return &Market{
        CToken: CToken{
            Address: cTokenAddr,
        },
        CollateralFactor: marketInfo.CollateralFactorMantissa,
        SupplyAPY:       supplyAPY,
        BorrowAPY:       borrowAPY,
    }, nil
}

// 获取供应利率
func (s *MarketService) getSupplyRate(cTokenAddr common.Address) (*big.Int, error) {
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    var rate *big.Int
    err := s.client.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &cTokenAddr,
        Data: s.client.cTokenABI.Methods["supplyRatePerBlock"].ID,
    }, nil)
    
    if err != nil {
        return nil, err
    }
    
    return rate, nil
}

// 获取借贷利率
func (s *MarketService) getBorrowRate(cTokenAddr common.Address) (*big.Int, error) {
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    var rate *big.Int
    err := s.client.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &cTokenAddr,
        Data: s.client.cTokenABI.Methods["borrowRatePerBlock"].ID,
    }, nil)
    
    if err != nil {
        return nil, err
    }
    
    return rate, nil
}

// 计算年化利率
func (s *MarketService) calculateAPY(ratePerBlock *big.Int, blocksPerYear *big.Int) decimal.Decimal {
    rate := decimal.NewFromBigInt(ratePerBlock, 0)
    blocks := decimal.NewFromBigInt(blocksPerYear, 0)
    mantissa := decimal.NewFromInt(1e18)
    
    // APY = (1 + rate/1e18)^blocksPerYear - 1
    rateDecimal := rate.Div(mantissa)
    annualRate := rateDecimal.Mul(blocks)
    
    return annualRate.Mul(decimal.NewFromInt(100)) // 转换为百分比
}

// 获取所有市场信息
func (s *MarketService) GetAllMarkets() ([]*Market, error) {
    cTokenAddresses := []common.Address{
        CETHAddress,
        CDAIAddress,
        CUSDCAddress,
        // 添加更多cToken地址
    }
    
    var markets []*Market
    for _, addr := range cTokenAddresses {
        market, err := s.GetMarketInfo(addr)
        if err != nil {
            continue // 跳过错误的市场
        }
        markets = append(markets, market)
    }
    
    return markets, nil
}
```

## 资产供应

### 4.1 供应资产

```go
// services/supply_service.go
package services

import (
    "context"
    "crypto/ecdsa"
    "math/big"
    
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/crypto"
)

type SupplyService struct {
    client     *CompoundClient
    privateKey *ecdsa.PrivateKey
}

func NewSupplyService(client *CompoundClient, privateKey *ecdsa.PrivateKey) *SupplyService {
    return &SupplyService{
        client:     client,
        privateKey: privateKey,
    }
}

// 供应ETH
func (s *SupplyService) SupplyETH(amount *big.Int) (*types.Transaction, error) {
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 获取nonce
    nonce, err := s.client.ethClient.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        return nil, err
    }
    
    // 获取gas价格
    gasPrice, err := s.client.ethClient.SuggestGasPrice(context.Background())
    if err != nil {
        return nil, err
    }
    
    // 构建交易选项
    auth, err := bind.NewKeyedTransactorWithChainID(s.privateKey, big.NewInt(1)) // Mainnet
    if err != nil {
        return nil, err
    }
    
    auth.Nonce = big.NewInt(int64(nonce))
    auth.Value = amount
    auth.GasLimit = uint64(250000)
    auth.GasPrice = gasPrice
    
    // 调用cETH的mint函数
    data, err := s.client.cTokenABI.Pack("mint")
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, CETHAddress, amount, 250000, gasPrice, data)
    
    // 签名交易
    chainID, err := s.client.ethClient.NetworkID(context.Background())
    if err != nil {
        return nil, err
    }
    
    signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), s.privateKey)
    if err != nil {
        return nil, err
    }
    
    // 发送交易
    err = s.client.ethClient.SendTransaction(context.Background(), signedTx)
    if err != nil {
        return nil, err
    }
    
    return signedTx, nil
}

// 供应ERC20代币
func (s *SupplyService) SupplyERC20(cTokenAddr, tokenAddr common.Address, amount *big.Int) (*types.Transaction, error) {
    // 首先需要授权cToken合约使用代币
    err := s.approveToken(tokenAddr, cTokenAddr, amount)
    if err != nil {
        return nil, err
    }
    
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 获取nonce
    nonce, err := s.client.ethClient.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        return nil, err
    }
    
    // 获取gas价格
    gasPrice, err := s.client.ethClient.SuggestGasPrice(context.Background())
    if err != nil {
        return nil, err
    }
    
    // 构建mint交易数据
    data, err := s.client.cTokenABI.Pack("mint", amount)
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, cTokenAddr, big.NewInt(0), 300000, gasPrice, data)
    
    // 签名交易
    chainID, err := s.client.ethClient.NetworkID(context.Background())
    if err != nil {
        return nil, err
    }
    
    signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), s.privateKey)
    if err != nil {
        return nil, err
    }
    
    // 发送交易
    err = s.client.ethClient.SendTransaction(context.Background(), signedTx)
    if err != nil {
        return nil, err
    }
    
    return signedTx, nil
}

// 授权代币
func (s *SupplyService) approveToken(tokenAddr, spender common.Address, amount *big.Int) error {
    // ERC20 approve函数的ABI
    erc20ABI := `[{"constant":false,"inputs":[{"name":"spender","type":"address"},{"name":"value","type":"uint256"}],"name":"approve","outputs":[{"name":"","type":"bool"}],"type":"function"}]`
    
    parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
    if err != nil {
        return err
    }
    
    // 构建approve交易数据
    data, err := parsedABI.Pack("approve", spender, amount)
    if err != nil {
        return err
    }
    
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 获取nonce
    nonce, err := s.client.ethClient.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        return err
    }
    
    // 获取gas价格
    gasPrice, err := s.client.ethClient.SuggestGasPrice(context.Background())
    if err != nil {
        return err
    }
    
    tx := types.NewTransaction(nonce, tokenAddr, big.NewInt(0), 100000, gasPrice, data)
    
    // 签名并发送交易
    chainID, err := s.client.ethClient.NetworkID(context.Background())
    if err != nil {
        return err
    }
    
    signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), s.privateKey)
    if err != nil {
        return err
    }
    
    return s.client.ethClient.SendTransaction(context.Background(), signedTx)
}

// 赎回资产
func (s *SupplyService) Redeem(cTokenAddr common.Address, cTokenAmount *big.Int) (*types.Transaction, error) {
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 获取nonce
    nonce, err := s.client.ethClient.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        return nil, err
    }
    
    // 获取gas价格
    gasPrice, err := s.client.ethClient.SuggestGasPrice(context.Background())
    if err != nil {
        return nil, err
    }
    
    // 构建redeem交易数据
    data, err := s.client.cTokenABI.Pack("redeem", cTokenAmount)
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, cTokenAddr, big.NewInt(0), 300000, gasPrice, data)
    
    // 签名交易
    chainID, err := s.client.ethClient.NetworkID(context.Background())
    if err != nil {
        return nil, err
    }
    
    signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), s.privateKey)
    if err != nil {
        return nil, err
    }
    
    // 发送交易
    err = s.client.ethClient.SendTransaction(context.Background(), signedTx)
    if err != nil {
        return nil, err
    }
    
    return signedTx, nil
}
```

## 资产借贷

### 5.1 借贷操作

```go
// services/borrow_service.go
package services

import (
    "context"
    "crypto/ecdsa"
    "math/big"
    
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/crypto"
)

type BorrowService struct {
    client     *CompoundClient
    privateKey *ecdsa.PrivateKey
}

func NewBorrowService(client *CompoundClient, privateKey *ecdsa.PrivateKey) *BorrowService {
    return &BorrowService{
        client:     client,
        privateKey: privateKey,
    }
}

// 进入市场（启用抵押）
func (s *BorrowService) EnterMarkets(cTokenAddrs []common.Address) (*types.Transaction, error) {
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 获取nonce
    nonce, err := s.client.ethClient.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        return nil, err
    }
    
    // 获取gas价格
    gasPrice, err := s.client.ethClient.SuggestGasPrice(context.Background())
    if err != nil {
        return nil, err
    }
    
    // 构建enterMarkets交易数据
    data, err := s.client.comptrollerABI.Pack("enterMarkets", cTokenAddrs)
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, s.client.comptrollerAddr, big.NewInt(0), 200000, gasPrice, data)
    
    // 签名交易
    chainID, err := s.client.ethClient.NetworkID(context.Background())
    if err != nil {
        return nil, err
    }
    
    signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), s.privateKey)
    if err != nil {
        return nil, err
    }
    
    // 发送交易
    err = s.client.ethClient.SendTransaction(context.Background(), signedTx)
    if err != nil {
        return nil, err
    }
    
    return signedTx, nil
}

// 借贷资产
func (s *BorrowService) Borrow(cTokenAddr common.Address, borrowAmount *big.Int) (*types.Transaction, error) {
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 获取nonce
    nonce, err := s.client.ethClient.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        return nil, err
    }
    
    // 获取gas价格
    gasPrice, err := s.client.ethClient.SuggestGasPrice(context.Background())
    if err != nil {
        return nil, err
    }
    
    // 构建borrow交易数据
    data, err := s.client.cTokenABI.Pack("borrow", borrowAmount)
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, cTokenAddr, big.NewInt(0), 300000, gasPrice, data)
    
    // 签名交易
    chainID, err := s.client.ethClient.NetworkID(context.Background())
    if err != nil {
        return nil, err
    }
    
    signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), s.privateKey)
    if err != nil {
        return nil, err
    }
    
    // 发送交易
    err = s.client.ethClient.SendTransaction(context.Background(), signedTx)
    if err != nil {
        return nil, err
    }
    
    return signedTx, nil
}

// 还款
func (s *BorrowService) RepayBorrow(cTokenAddr common.Address, repayAmount *big.Int) (*types.Transaction, error) {
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 获取nonce
    nonce, err := s.client.ethClient.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        return nil, err
    }
    
    // 获取gas价格
    gasPrice, err := s.client.ethClient.SuggestGasPrice(context.Background())
    if err != nil {
        return nil, err
    }
    
    var data []byte
    var err error
    var value *big.Int = big.NewInt(0)
    
    // 如果是ETH，需要发送ETH
    if cTokenAddr == CETHAddress {
        data, err = s.client.cTokenABI.Pack("repayBorrow")
        value = repayAmount
    } else {
        // ERC20代币需要先授权
        data, err = s.client.cTokenABI.Pack("repayBorrow", repayAmount)
    }
    
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, cTokenAddr, value, 300000, gasPrice, data)
    
    // 签名交易
    chainID, err := s.client.ethClient.NetworkID(context.Background())
    if err != nil {
        return nil, err
    }
    
    signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), s.privateKey)
    if err != nil {
        return nil, err
    }
    
    // 发送交易
    err = s.client.ethClient.SendTransaction(context.Background(), signedTx)
    if err != nil {
        return nil, err
    }
    
    return signedTx, nil
}
```

## 利率计算

### 6.1 利率模型

```go
// utils/interest_rate.go
package utils

import (
    "math/big"
    
    "github.com/shopspring/decimal"
)

// 利率计算器
type InterestRateCalculator struct {
    BlocksPerYear *big.Int
}

func NewInterestRateCalculator() *InterestRateCalculator {
    return &InterestRateCalculator{
        BlocksPerYear: big.NewInt(2102400), // 以太坊每年约2,102,400个区块
    }
}

// 计算复合年化利率
func (calc *InterestRateCalculator) CalculateCompoundAPY(ratePerBlock *big.Int) decimal.Decimal {
    rate := decimal.NewFromBigInt(ratePerBlock, -18) // 转换为小数
    blocksPerYear := decimal.NewFromBigInt(calc.BlocksPerYear, 0)
    
    // 复合利率公式: (1 + rate)^blocks - 1
    onePlusRate := decimal.NewFromInt(1).Add(rate)
    
    // 简化计算，实际应该使用幂运算
    // 这里使用近似计算: rate * blocksPerYear
    apy := rate.Mul(blocksPerYear)
    
    return apy.Mul(decimal.NewFromInt(100)) // 转换为百分比
}

// 计算借贷能力
func (calc *InterestRateCalculator) CalculateBorrowCapacity(
    collateralValue *big.Int,
    collateralFactor *big.Int,
) *big.Int {
    // 借贷能力 = 抵押品价值 * 抵押因子
    capacity := new(big.Int).Mul(collateralValue, collateralFactor)
    capacity.Div(capacity, big.NewInt(1e18)) // 除以1e18（抵押因子的精度）
    
    return capacity
}

// 计算健康因子
func (calc *InterestRateCalculator) CalculateHealthFactor(
    collateralValue *big.Int,
    borrowValue *big.Int,
    liquidationThreshold *big.Int,
) decimal.Decimal {
    if borrowValue.Cmp(big.NewInt(0)) == 0 {
        return decimal.NewFromInt(999999) // 无借贷时健康因子为无穷大
    }
    
    collateral := decimal.NewFromBigInt(collateralValue, 0)
    borrow := decimal.NewFromBigInt(borrowValue, 0)
    threshold := decimal.NewFromBigInt(liquidationThreshold, -18)
    
    // 健康因子 = (抵押品价值 * 清算阈值) / 借贷价值
    healthFactor := collateral.Mul(threshold).Div(borrow)
    
    return healthFactor
}
```

## 清算机制

### 7.1 清算服务

```go
// services/liquidation_service.go
package services

import (
    "context"
    "crypto/ecdsa"
    "math/big"
    
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
)

type LiquidationService struct {
    client     *CompoundClient
    privateKey *ecdsa.PrivateKey
}

func NewLiquidationService(client *CompoundClient, privateKey *ecdsa.PrivateKey) *LiquidationService {
    return &LiquidationService{
        client:     client,
        privateKey: privateKey,
    }
}

// 检查可清算账户
func (s *LiquidationService) CheckLiquidatableAccounts(accounts []common.Address) ([]common.Address, error) {
    var liquidatable []common.Address
    
    for _, account := range accounts {
        liquidity, shortfall, err := s.client.GetAccountLiquidity(account)
        if err != nil {
            continue
        }
        
        // 如果shortfall > 0，说明账户可以被清算
        if shortfall.Cmp(big.NewInt(0)) > 0 {
            liquidatable = append(liquidatable, account)
        }
    }
    
    return liquidatable, nil
}

// 执行清算
func (s *LiquidationService) Liquidate(
    borrower common.Address,
    cTokenBorrowed common.Address,
    cTokenCollateral common.Address,
    repayAmount *big.Int,
) (*types.Transaction, error) {
    // 构建清算交易数据
    // 这需要调用cToken的liquidateBorrow函数
    
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 获取nonce
    nonce, err := s.client.ethClient.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        return nil, err
    }
    
    // 获取gas价格
    gasPrice, err := s.client.ethClient.SuggestGasPrice(context.Background())
    if err != nil {
        return nil, err
    }
    
    // 构建liquidateBorrow交易数据
    data, err := s.client.cTokenABI.Pack(
        "liquidateBorrow",
        borrower,
        repayAmount,
        cTokenCollateral,
    )
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, cTokenBorrowed, big.NewInt(0), 500000, gasPrice, data)
    
    // 签名交易
    chainID, err := s.client.ethClient.NetworkID(context.Background())
    if err != nil {
        return nil, err
    }
    
    signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), s.privateKey)
    if err != nil {
        return nil, err
    }
    
    // 发送交易
    err = s.client.ethClient.SendTransaction(context.Background(), signedTx)
    if err != nil {
        return nil, err
    }
    
    return signedTx, nil
}

// 计算清算奖励
func (s *LiquidationService) CalculateLiquidationReward(
    repayAmount *big.Int,
    collateralPrice *big.Int,
    borrowPrice *big.Int,
) *big.Int {
    // 清算奖励 = 还款金额 * (1 + 清算激励) * 借贷价格 / 抵押品价格
    // 清算激励通常为5%
    
    incentive := big.NewInt(105) // 105% = 100% + 5%激励
    base := big.NewInt(100)
    
    reward := new(big.Int).Mul(repayAmount, incentive)
    reward.Div(reward, base)
    reward.Mul(reward, borrowPrice)
    reward.Div(reward, collateralPrice)
    
    return reward
}
```

## 治理代币

### 8.1 COMP代币操作

```go
// services/governance_service.go
package services

import (
    "context"
    "math/big"
    
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
)

type GovernanceService struct {
    client *CompoundClient
}

func NewGovernanceService(client *CompoundClient) *GovernanceService {
    return &GovernanceService{
        client: client,
    }
}

// 查询COMP余额
func (s *GovernanceService) GetCOMPBalance(account common.Address) (*big.Int, error) {
    // ERC20 balanceOf函数
    erc20ABI := `[{"constant":true,"inputs":[{"name":"account","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"type":"function"}]`
    
    parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
    if err != nil {
        return nil, err
    }
    
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    var balance *big.Int
    err = s.client.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &COMPAddress,
        Data: parsedABI.Methods["balanceOf"].ID,
    }, nil)
    
    if err != nil {
        return nil, err
    }
    
    return balance, nil
}

// 查询可领取的COMP奖励
func (s *GovernanceService) GetAccruedCOMP(account common.Address) (*big.Int, error) {
    // 调用Comptroller的compAccrued函数
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    var accrued *big.Int
    err := s.client.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &s.client.comptrollerAddr,
        Data: s.client.comptrollerABI.Methods["compAccrued"].ID,
    }, nil)
    
    if err != nil {
        return nil, err
    }
    
    return accrued, nil
}

// 领取COMP奖励
func (s *GovernanceService) ClaimCOMP(account common.Address) (*types.Transaction, error) {
    // 调用Comptroller的claimComp函数
    data, err := s.client.comptrollerABI.Pack("claimComp", account)
    if err != nil {
        return nil, err
    }
    
    // 构建和发送交易的逻辑...
    
    return nil, nil
}
```

## 实际应用

### 9.1 完整示例

```go
// main.go
package main

import (
    "fmt"
    "log"
    "math/big"
    
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/crypto"
    
    "your-project/client"
    "your-project/services"
)

func main() {
    // 创建Compound客户端
    compoundClient, err := client.NewCompoundClient("https://mainnet.infura.io/v3/YOUR_PROJECT_ID")
    if err != nil {
        log.Fatal("创建Compound客户端失败:", err)
    }
    
    // 加载私钥
    privateKey, err := crypto.HexToECDSA("your_private_key_here")
    if err != nil {
        log.Fatal("加载私钥失败:", err)
    }
    
    // 创建服务
    marketService := services.NewMarketService(compoundClient)
    supplyService := services.NewSupplyService(compoundClient, privateKey)
    borrowService := services.NewBorrowService(compoundClient, privateKey)
    
    // 查询市场信息
    markets, err := marketService.GetAllMarkets()
    if err != nil {
        log.Fatal("获取市场信息失败:", err)
    }
    
    fmt.Println("Compound市场信息:")
    for _, market := range markets {
        fmt.Printf("市场: %s, 供应APY: %s%%, 借贷APY: %s%%\n",
            market.CToken.Address.Hex(),
            market.SupplyAPY.String(),
            market.BorrowAPY.String(),
        )
    }
    
    // 供应1 ETH
    supplyAmount := big.NewInt(1e18) // 1 ETH
    tx, err := supplyService.SupplyETH(supplyAmount)
    if err != nil {
        log.Fatal("供应ETH失败:", err)
    }
    
    fmt.Printf("供应交易已提交: %s\n", tx.Hash().Hex())
    
    // 进入市场以启用抵押
    cTokens := []common.Address{CETHAddress}
    tx, err = borrowService.EnterMarkets(cTokens)
    if err != nil {
        log.Fatal("进入市场失败:", err)
    }
    
    fmt.Printf("进入市场交易已提交: %s\n", tx.Hash().Hex())
    
    // 借贷100 USDC
    borrowAmount := big.NewInt(100 * 1e6) // 100 USDC
    tx, err = borrowService.Borrow(CUSDCAddress, borrowAmount)
    if err != nil {
        log.Fatal("借贷失败:", err)
    }
    
    fmt.Printf("借贷交易已提交: %s\n", tx.Hash().Hex())
}
```

这个Compound使用指南提供了完整的DeFi借贷协议集成方案，涵盖了供应、借贷、清算等核心功能。
