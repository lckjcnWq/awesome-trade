# Aave 协议 详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [环境准备](#环境准备)
3. [协议架构](#协议架构)
4. [资产供应](#资产供应)
5. [资产借贷](#资产借贷)
6. [闪电贷](#闪电贷)
7. [利率模式](#利率模式)
8. [清算机制](#清算机制)
9. [实际应用](#实际应用)

## 基础概念

### 1.1 Aave 协议简介

Aave 是领先的去中心化借贷协议，支持超额抵押借贷、闪电贷、利率切换等创新功能，是DeFi借贷领域的重要基础设施。

```bash
# 安装Aave相关依赖
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

// Aave V3 合约地址 (Mainnet)
var (
    PoolAddress                = common.HexToAddress("0x87870Bca3F3fD6335C3F4ce8392D69350B4fA4E2")
    PoolDataProviderAddress    = common.HexToAddress("0x7B4EB56E7CD4b454BA8ff71E4518426369a138a3")
    PriceOracleAddress         = common.HexToAddress("0x54586bE62E3c3580375aE3723C145253060Ca0C2")
    AaveTokenAddress           = common.HexToAddress("0x7Fc66500c84A76Ad7e9c93437bFc5Ac33E2DDaE9")
    StkAaveAddress             = common.HexToAddress("0x4da27a545c0c5B758a6BA100e3a049001de870f5")
)

// aToken 和 债务代币接口
type AToken struct {
    Address     common.Address
    Symbol      string
    Decimals    uint8
    Underlying  common.Address
}

// 储备资产信息
type ReserveData struct {
    Configuration       ReserveConfiguration
    LiquidityIndex      *big.Int
    VariableBorrowIndex *big.Int
    CurrentLiquidityRate *big.Int
    CurrentVariableBorrowRate *big.Int
    CurrentStableBorrowRate   *big.Int
    LastUpdateTimestamp uint40
    ATokenAddress       common.Address
    StableDebtTokenAddress   common.Address
    VariableDebtTokenAddress common.Address
    InterestRateStrategyAddress common.Address
    AccruedToTreasury   *big.Int
    Unbacked           *big.Int
    IsolationModeTotalDebt *big.Int
}
```

## 环境准备

### 2.1 合约ABI定义

```go
// contracts/aave_abi.go
package contracts

// Pool ABI (简化版)
const PoolABI = `[
    {
        "inputs": [
            {"name": "asset", "type": "address"},
            {"name": "amount", "type": "uint256"},
            {"name": "onBehalfOf", "type": "address"},
            {"name": "referralCode", "type": "uint16"}
        ],
        "name": "supply",
        "outputs": [],
        "type": "function"
    },
    {
        "inputs": [
            {"name": "asset", "type": "address"},
            {"name": "amount", "type": "uint256"},
            {"name": "to", "type": "address"}
        ],
        "name": "withdraw",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    },
    {
        "inputs": [
            {"name": "asset", "type": "address"},
            {"name": "amount", "type": "uint256"},
            {"name": "interestRateMode", "type": "uint256"},
            {"name": "referralCode", "type": "uint16"},
            {"name": "onBehalfOf", "type": "address"}
        ],
        "name": "borrow",
        "outputs": [],
        "type": "function"
    },
    {
        "inputs": [
            {"name": "asset", "type": "address"},
            {"name": "amount", "type": "uint256"},
            {"name": "rateMode", "type": "uint256"},
            {"name": "onBehalfOf", "type": "address"}
        ],
        "name": "repay",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    },
    {
        "inputs": [
            {"name": "assets", "type": "address[]"},
            {"name": "amounts", "type": "uint256[]"},
            {"name": "modes", "type": "uint256[]"},
            {"name": "onBehalfOf", "type": "address"},
            {"name": "params", "type": "bytes"},
            {"name": "referralCode", "type": "uint16"}
        ],
        "name": "flashLoan",
        "outputs": [],
        "type": "function"
    },
    {
        "inputs": [{"name": "user", "type": "address"}],
        "name": "getUserAccountData",
        "outputs": [
            {"name": "totalCollateralETH", "type": "uint256"},
            {"name": "totalDebtETH", "type": "uint256"},
            {"name": "availableBorrowsETH", "type": "uint256"},
            {"name": "currentLiquidationThreshold", "type": "uint256"},
            {"name": "ltv", "type": "uint256"},
            {"name": "healthFactor", "type": "uint256"}
        ],
        "type": "function"
    }
]`

// PoolDataProvider ABI (简化版)
const PoolDataProviderABI = `[
    {
        "inputs": [{"name": "asset", "type": "address"}],
        "name": "getReserveData",
        "outputs": [
            {"name": "unbacked", "type": "uint256"},
            {"name": "accruedToTreasuryScaled", "type": "uint256"},
            {"name": "totalAToken", "type": "uint256"},
            {"name": "totalStableDebt", "type": "uint256"},
            {"name": "totalVariableDebt", "type": "uint256"},
            {"name": "liquidityRate", "type": "uint256"},
            {"name": "variableBorrowRate", "type": "uint256"},
            {"name": "stableBorrowRate", "type": "uint256"},
            {"name": "averageStableBorrowRate", "type": "uint256"},
            {"name": "liquidityIndex", "type": "uint256"},
            {"name": "variableBorrowIndex", "type": "uint256"},
            {"name": "lastUpdateTimestamp", "type": "uint40"}
        ],
        "type": "function"
    },
    {
        "inputs": [
            {"name": "user", "type": "address"},
            {"name": "asset", "type": "address"}
        ],
        "name": "getUserReserveData",
        "outputs": [
            {"name": "currentATokenBalance", "type": "uint256"},
            {"name": "currentStableDebt", "type": "uint256"},
            {"name": "currentVariableDebt", "type": "uint256"},
            {"name": "principalStableDebt", "type": "uint256"},
            {"name": "scaledVariableDebt", "type": "uint256"},
            {"name": "stableBorrowRate", "type": "uint256"},
            {"name": "liquidityRate", "type": "uint256"},
            {"name": "stableRateLastUpdated", "type": "uint40"},
            {"name": "usageAsCollateralEnabled", "type": "bool"}
        ],
        "type": "function"
    }
]`
```

### 2.2 客户端设置

```go
// client/aave_client.go
package client

import (
    "context"
    "math/big"
    
    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
)

type AaveClient struct {
    ethClient           *ethclient.Client
    poolABI            abi.ABI
    dataProviderABI    abi.ABI
    poolAddress        common.Address
    dataProviderAddress common.Address
}

func NewAaveClient(rpcURL string) (*AaveClient, error) {
    ethClient, err := ethclient.Dial(rpcURL)
    if err != nil {
        return nil, err
    }
    
    poolABI, err := abi.JSON(strings.NewReader(PoolABI))
    if err != nil {
        return nil, err
    }
    
    dataProviderABI, err := abi.JSON(strings.NewReader(PoolDataProviderABI))
    if err != nil {
        return nil, err
    }
    
    return &AaveClient{
        ethClient:           ethClient,
        poolABI:            poolABI,
        dataProviderABI:    dataProviderABI,
        poolAddress:        PoolAddress,
        dataProviderAddress: PoolDataProviderAddress,
    }, nil
}

// 获取用户账户数据
func (c *AaveClient) GetUserAccountData(user common.Address) (*UserAccountData, error) {
    callOpts := &bind.CallOpts{
        Context: context.Background(),
    }
    
    var result []interface{}
    err := c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To:   &c.poolAddress,
        Data: c.poolABI.Methods["getUserAccountData"].ID,
    }, nil)
    
    if err != nil {
        return nil, err
    }
    
    err = c.poolABI.UnpackIntoInterface(&result, "getUserAccountData", result)
    if err != nil {
        return nil, err
    }
    
    return &UserAccountData{
        TotalCollateralETH:          result[0].(*big.Int),
        TotalDebtETH:               result[1].(*big.Int),
        AvailableBorrowsETH:        result[2].(*big.Int),
        CurrentLiquidationThreshold: result[3].(*big.Int),
        LTV:                        result[4].(*big.Int),
        HealthFactor:               result[5].(*big.Int),
    }, nil
}

type UserAccountData struct {
    TotalCollateralETH          *big.Int
    TotalDebtETH               *big.Int
    AvailableBorrowsETH        *big.Int
    CurrentLiquidationThreshold *big.Int
    LTV                        *big.Int
    HealthFactor               *big.Int
}
```

## 协议架构

### 3.1 Aave V3 架构

```go
// services/aave_service.go
package services

import (
    "context"
    "fmt"
    "math/big"
    
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/shopspring/decimal"
)

type AaveService struct {
    client *AaveClient
}

func NewAaveService(client *AaveClient) *AaveService {
    return &AaveService{
        client: client,
    }
}

// 获取储备资产数据
func (s *AaveService) GetReserveData(asset common.Address) (*ReserveInfo, error) {
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    var result []interface{}
    err := s.client.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &s.client.dataProviderAddress,
        Data: s.client.dataProviderABI.Methods["getReserveData"].ID,
    }, nil)
    
    if err != nil {
        return nil, err
    }
    
    return &ReserveInfo{
        Asset:               asset,
        TotalAToken:         result[2].(*big.Int),
        TotalStableDebt:     result[3].(*big.Int),
        TotalVariableDebt:   result[4].(*big.Int),
        LiquidityRate:       result[5].(*big.Int),
        VariableBorrowRate:  result[6].(*big.Int),
        StableBorrowRate:    result[7].(*big.Int),
        LiquidityIndex:      result[9].(*big.Int),
        VariableBorrowIndex: result[10].(*big.Int),
        LastUpdateTimestamp: result[11].(uint40),
    }, nil
}

// 计算年化利率
func (s *AaveService) CalculateAPY(ratePerSecond *big.Int) decimal.Decimal {
    // Aave V3 使用每秒利率
    secondsPerYear := decimal.NewFromInt(31536000) // 365 * 24 * 3600
    rate := decimal.NewFromBigInt(ratePerSecond, -27) // Ray精度 (1e27)
    
    // APY = (1 + rate)^secondsPerYear - 1
    // 简化计算: rate * secondsPerYear (对于小利率近似)
    apy := rate.Mul(secondsPerYear)
    
    return apy.Mul(decimal.NewFromInt(100)) // 转换为百分比
}

type ReserveInfo struct {
    Asset               common.Address
    TotalAToken         *big.Int
    TotalStableDebt     *big.Int
    TotalVariableDebt   *big.Int
    LiquidityRate       *big.Int
    VariableBorrowRate  *big.Int
    StableBorrowRate    *big.Int
    LiquidityIndex      *big.Int
    VariableBorrowIndex *big.Int
    LastUpdateTimestamp uint40
}
```

## 资产供应

### 4.1 供应操作

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
    client     *AaveClient
    privateKey *ecdsa.PrivateKey
}

func NewSupplyService(client *AaveClient, privateKey *ecdsa.PrivateKey) *SupplyService {
    return &SupplyService{
        client:     client,
        privateKey: privateKey,
    }
}

// 供应资产到Aave
func (s *SupplyService) Supply(
    asset common.Address,
    amount *big.Int,
    onBehalfOf common.Address,
    referralCode uint16,
) (*types.Transaction, error) {
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 首先需要授权Pool合约使用代币
    if err := s.approveToken(asset, s.client.poolAddress, amount); err != nil {
        return nil, err
    }
    
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
    
    // 构建supply交易数据
    data, err := s.client.poolABI.Pack("supply", asset, amount, onBehalfOf, referralCode)
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, s.client.poolAddress, big.NewInt(0), 300000, gasPrice, data)
    
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

// 提取资产
func (s *SupplyService) Withdraw(
    asset common.Address,
    amount *big.Int, // 使用 type(uint256).max 提取全部
    to common.Address,
) (*types.Transaction, error) {
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
    
    // 构建withdraw交易数据
    data, err := s.client.poolABI.Pack("withdraw", asset, amount, to)
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, s.client.poolAddress, big.NewInt(0), 300000, gasPrice, data)
    
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
    
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/crypto"
)

type BorrowService struct {
    client     *AaveClient
    privateKey *ecdsa.PrivateKey
}

func NewBorrowService(client *AaveClient, privateKey *ecdsa.PrivateKey) *BorrowService {
    return &BorrowService{
        client:     client,
        privateKey: privateKey,
    }
}

// 利率模式
const (
    StableRate   = 1
    VariableRate = 2
)

// 借贷资产
func (s *BorrowService) Borrow(
    asset common.Address,
    amount *big.Int,
    interestRateMode uint64, // 1=稳定利率, 2=浮动利率
    referralCode uint16,
    onBehalfOf common.Address,
) (*types.Transaction, error) {
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
    data, err := s.client.poolABI.Pack("borrow", asset, amount, big.NewInt(int64(interestRateMode)), referralCode, onBehalfOf)
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, s.client.poolAddress, big.NewInt(0), 400000, gasPrice, data)
    
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
func (s *BorrowService) Repay(
    asset common.Address,
    amount *big.Int, // 使用 type(uint256).max 还清全部债务
    rateMode uint64, // 1=稳定利率, 2=浮动利率
    onBehalfOf common.Address,
) (*types.Transaction, error) {
    // 首先需要授权Pool合约使用代币
    if err := s.approveToken(asset, s.client.poolAddress, amount); err != nil {
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
    
    // 构建repay交易数据
    data, err := s.client.poolABI.Pack("repay", asset, amount, big.NewInt(int64(rateMode)), onBehalfOf)
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, s.client.poolAddress, big.NewInt(0), 300000, gasPrice, data)
    
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

// 切换利率模式
func (s *BorrowService) SwapBorrowRateMode(
    asset common.Address,
    rateMode uint64,
) (*types.Transaction, error) {
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
    
    // 构建swapBorrowRateMode交易数据
    data, err := s.client.poolABI.Pack("swapBorrowRateMode", asset, big.NewInt(int64(rateMode)))
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, s.client.poolAddress, big.NewInt(0), 200000, gasPrice, data)
    
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

## 闪电贷

### 6.1 闪电贷实现

```go
// services/flashloan_service.go
package services

import (
    "context"
    "crypto/ecdsa"
    "math/big"
    
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/crypto"
)

type FlashLoanService struct {
    client     *AaveClient
    privateKey *ecdsa.PrivateKey
}

func NewFlashLoanService(client *AaveClient, privateKey *ecdsa.PrivateKey) *FlashLoanService {
    return &FlashLoanService{
        client:     client,
        privateKey: privateKey,
    }
}

// 闪电贷模式
const (
    NoDebt    = 0 // 无债务，必须在同一交易中还款
    StableDebt = 1 // 转为稳定利率债务
    VariableDebt = 2 // 转为浮动利率债务
)

// 执行闪电贷
func (s *FlashLoanService) FlashLoan(
    assets []common.Address,
    amounts []*big.Int,
    modes []*big.Int, // 0=无债务, 1=稳定债务, 2=浮动债务
    onBehalfOf common.Address,
    params []byte, // 传递给executeOperation的参数
    referralCode uint16,
) (*types.Transaction, error) {
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
    
    // 构建flashLoan交易数据
    data, err := s.client.poolABI.Pack("flashLoan", assets, amounts, modes, onBehalfOf, params, referralCode)
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, s.client.poolAddress, big.NewInt(0), 1000000, gasPrice, data)
    
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

// 计算闪电贷手续费
func (s *FlashLoanService) CalculateFlashLoanFee(amount *big.Int) *big.Int {
    // Aave V3 闪电贷手续费为 0.05%
    fee := new(big.Int).Mul(amount, big.NewInt(5))
    fee.Div(fee, big.NewInt(10000))
    return fee
}

// 闪电贷接收器合约示例
const FlashLoanReceiverABI = `[
    {
        "inputs": [
            {"name": "assets", "type": "address[]"},
            {"name": "amounts", "type": "uint256[]"},
            {"name": "premiums", "type": "uint256[]"},
            {"name": "initiator", "type": "address"},
            {"name": "params", "type": "bytes"}
        ],
        "name": "executeOperation",
        "outputs": [{"name": "", "type": "bool"}],
        "type": "function"
    }
]`

// 闪电贷套利示例
type FlashLoanArbitrage struct {
    flashLoanService *FlashLoanService
    dexA            DEXInterface
    dexB            DEXInterface
}

func (fla *FlashLoanArbitrage) ExecuteArbitrage(
    asset common.Address,
    amount *big.Int,
) (*types.Transaction, error) {
    // 检查套利机会
    priceA, err := fla.dexA.GetPrice(asset)
    if err != nil {
        return nil, err
    }
    
    priceB, err := fla.dexB.GetPrice(asset)
    if err != nil {
        return nil, err
    }
    
    // 计算预期利润
    expectedProfit := fla.calculateProfit(amount, priceA, priceB)
    flashLoanFee := fla.flashLoanService.CalculateFlashLoanFee(amount)
    
    if expectedProfit.Cmp(flashLoanFee) <= 0 {
        return nil, fmt.Errorf("套利不盈利")
    }
    
    // 构建闪电贷参数
    params := fla.encodeArbitrageParams(asset, amount, priceA, priceB)
    
    // 执行闪电贷
    return fla.flashLoanService.FlashLoan(
        []common.Address{asset},
        []*big.Int{amount},
        []*big.Int{big.NewInt(NoDebt)},
        fla.flashLoanService.client.poolAddress,
        params,
        0,
    )
}

type DEXInterface interface {
    GetPrice(asset common.Address) (*big.Int, error)
    Swap(assetIn, assetOut common.Address, amountIn *big.Int) (*big.Int, error)
}
```

## 利率模式

### 7.1 利率计算

```go
// utils/interest_rate.go
package utils

import (
    "math/big"
    
    "github.com/shopspring/decimal"
)

// Aave利率计算器
type AaveInterestCalculator struct {
    RAY *big.Int // 1e27
}

func NewAaveInterestCalculator() *AaveInterestCalculator {
    ray := new(big.Int)
    ray.Exp(big.NewInt(10), big.NewInt(27), nil)
    
    return &AaveInterestCalculator{
        RAY: ray,
    }
}

// 计算复合年化利率
func (calc *AaveInterestCalculator) CalculateCompoundedAPY(ratePerSecond *big.Int) decimal.Decimal {
    // 转换为小数
    rate := decimal.NewFromBigInt(ratePerSecond, -27)
    
    // 一年的秒数
    secondsPerYear := decimal.NewFromInt(31536000)
    
    // 复合利率公式: (1 + rate)^secondsPerYear - 1
    // 对于小利率，使用近似公式: rate * secondsPerYear
    apy := rate.Mul(secondsPerYear)
    
    return apy.Mul(decimal.NewFromInt(100))
}

// 计算健康因子
func (calc *AaveInterestCalculator) CalculateHealthFactor(
    totalCollateralETH *big.Int,
    totalDebtETH *big.Int,
    liquidationThreshold *big.Int,
) decimal.Decimal {
    if totalDebtETH.Cmp(big.NewInt(0)) == 0 {
        return decimal.NewFromInt(999999) // 无债务时健康因子为无穷大
    }
    
    collateral := decimal.NewFromBigInt(totalCollateralETH, -18)
    debt := decimal.NewFromBigInt(totalDebtETH, -18)
    threshold := decimal.NewFromBigInt(liquidationThreshold, -4) // 基点转换
    
    // 健康因子 = (抵押品价值 * 清算阈值) / 债务价值
    healthFactor := collateral.Mul(threshold).Div(debt)
    
    return healthFactor
}

// 计算最大借贷金额
func (calc *AaveInterestCalculator) CalculateMaxBorrowAmount(
    totalCollateralETH *big.Int,
    totalDebtETH *big.Int,
    ltv *big.Int,
) *big.Int {
    collateral := decimal.NewFromBigInt(totalCollateralETH, -18)
    debt := decimal.NewFromBigInt(totalDebtETH, -18)
    ltvRatio := decimal.NewFromBigInt(ltv, -4) // 基点转换
    
    // 最大借贷 = 抵押品价值 * LTV - 当前债务
    maxBorrow := collateral.Mul(ltvRatio).Sub(debt)
    
    if maxBorrow.LessThanOrEqual(decimal.Zero) {
        return big.NewInt(0)
    }
    
    result, _ := maxBorrow.Mul(decimal.NewFromInt(1e18)).BigInt()
    return result
}

// 计算清算价格
func (calc *AaveInterestCalculator) CalculateLiquidationPrice(
    collateralAmount *big.Int,
    debtAmount *big.Int,
    liquidationThreshold *big.Int,
) decimal.Decimal {
    debt := decimal.NewFromBigInt(debtAmount, -18)
    collateral := decimal.NewFromBigInt(collateralAmount, -18)
    threshold := decimal.NewFromBigInt(liquidationThreshold, -4)
    
    // 清算价格 = 债务价值 / (抵押品数量 * 清算阈值)
    liquidationPrice := debt.Div(collateral.Mul(threshold))
    
    return liquidationPrice
}
```

## 清算机制

### 8.1 清算服务

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
    client     *AaveClient
    privateKey *ecdsa.PrivateKey
    calculator *AaveInterestCalculator
}

func NewLiquidationService(client *AaveClient, privateKey *ecdsa.PrivateKey) *LiquidationService {
    return &LiquidationService{
        client:     client,
        privateKey: privateKey,
        calculator: NewAaveInterestCalculator(),
    }
}

// 检查可清算用户
func (s *LiquidationService) CheckLiquidatableUsers(users []common.Address) ([]LiquidationOpportunity, error) {
    var opportunities []LiquidationOpportunity
    
    for _, user := range users {
        accountData, err := s.client.GetUserAccountData(user)
        if err != nil {
            continue
        }
        
        // 计算健康因子
        healthFactor := s.calculator.CalculateHealthFactor(
            accountData.TotalCollateralETH,
            accountData.TotalDebtETH,
            accountData.CurrentLiquidationThreshold,
        )
        
        // 健康因子 < 1 时可以清算
        if healthFactor.LessThan(decimal.NewFromInt(1)) {
            opportunity := LiquidationOpportunity{
                User:         user,
                HealthFactor: healthFactor,
                CollateralETH: accountData.TotalCollateralETH,
                DebtETH:      accountData.TotalDebtETH,
            }
            
            opportunities = append(opportunities, opportunity)
        }
    }
    
    return opportunities, nil
}

// 执行清算
func (s *LiquidationService) LiquidateCall(
    collateralAsset common.Address,
    debtAsset common.Address,
    user common.Address,
    debtToCover *big.Int,
    receiveAToken bool,
) (*types.Transaction, error) {
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 首先需要授权Pool合约使用债务代币
    if err := s.approveToken(debtAsset, s.client.poolAddress, debtToCover); err != nil {
        return nil, err
    }
    
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
    
    // 构建liquidationCall交易数据
    data, err := s.client.poolABI.Pack(
        "liquidationCall",
        collateralAsset,
        debtAsset,
        user,
        debtToCover,
        receiveAToken,
    )
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, s.client.poolAddress, big.NewInt(0), 500000, gasPrice, data)
    
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
func (s *LiquidationService) CalculateLiquidationBonus(
    collateralAsset common.Address,
    debtToCover *big.Int,
) (*big.Int, error) {
    // Aave V3 清算奖励通常为 5-10%
    // 这里需要从协议配置中获取具体的清算奖励比例
    
    // 简化计算，假设 5% 奖励
    bonus := new(big.Int).Mul(debtToCover, big.NewInt(5))
    bonus.Div(bonus, big.NewInt(100))
    
    return bonus, nil
}

type LiquidationOpportunity struct {
    User          common.Address
    HealthFactor  decimal.Decimal
    CollateralETH *big.Int
    DebtETH       *big.Int
    MaxLiquidation *big.Int
    ExpectedProfit *big.Int
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
    // 创建Aave客户端
    aaveClient, err := client.NewAaveClient("https://mainnet.infura.io/v3/YOUR_PROJECT_ID")
    if err != nil {
        log.Fatal("创建Aave客户端失败:", err)
    }
    
    // 加载私钥
    privateKey, err := crypto.HexToECDSA("your_private_key_here")
    if err != nil {
        log.Fatal("加载私钥失败:", err)
    }
    
    // 获取用户地址
    publicKey := privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    userAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 创建服务
    aaveService := services.NewAaveService(aaveClient)
    supplyService := services.NewSupplyService(aaveClient, privateKey)
    borrowService := services.NewBorrowService(aaveClient, privateKey)
    
    // USDC地址
    usdcAddress := common.HexToAddress("0xA0b86a33E6417c8f4c8c8c8c8c8c8c8c8c8c8c8c")
    
    // 查询用户账户数据
    accountData, err := aaveClient.GetUserAccountData(userAddress)
    if err != nil {
        log.Fatal("获取账户数据失败:", err)
    }
    
    fmt.Printf("总抵押品价值: %s ETH\n", accountData.TotalCollateralETH.String())
    fmt.Printf("总债务价值: %s ETH\n", accountData.TotalDebtETH.String())
    fmt.Printf("可借贷金额: %s ETH\n", accountData.AvailableBorrowsETH.String())
    fmt.Printf("健康因子: %s\n", accountData.HealthFactor.String())
    
    // 查询USDC储备数据
    reserveData, err := aaveService.GetReserveData(usdcAddress)
    if err != nil {
        log.Fatal("获取储备数据失败:", err)
    }
    
    supplyAPY := aaveService.CalculateAPY(reserveData.LiquidityRate)
    borrowAPY := aaveService.CalculateAPY(reserveData.VariableBorrowRate)
    
    fmt.Printf("USDC 供应APY: %s%%\n", supplyAPY.String())
    fmt.Printf("USDC 借贷APY: %s%%\n", borrowAPY.String())
    
    // 供应 1000 USDC
    supplyAmount := big.NewInt(1000 * 1e6) // 1000 USDC
    tx, err := supplyService.Supply(usdcAddress, supplyAmount, userAddress, 0)
    if err != nil {
        log.Fatal("供应失败:", err)
    }
    
    fmt.Printf("供应交易已提交: %s\n", tx.Hash().Hex())
    
    // 借贷 500 USDC (浮动利率)
    borrowAmount := big.NewInt(500 * 1e6) // 500 USDC
    tx, err = borrowService.Borrow(usdcAddress, borrowAmount, VariableRate, 0, userAddress)
    if err != nil {
        log.Fatal("借贷失败:", err)
    }
    
    fmt.Printf("借贷交易已提交: %s\n", tx.Hash().Hex())
}
```

这个Aave使用指南提供了完整的DeFi借贷协议集成方案，涵盖了供应、借贷、闪电贷、清算等核心功能，是DeFi开发的重要参考文档。
