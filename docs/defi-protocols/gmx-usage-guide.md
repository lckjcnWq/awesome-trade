# GMX 协议 详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [环境准备](#环境准备)
3. [GLP池机制](#glp池机制)
4. [永续合约交易](#永续合约交易)
5. [流动性提供](#流动性提供)
6. [费用和奖励](#费用和奖励)
7. [风险管理](#风险管理)
8. [套利机会](#套利机会)
9. [实际应用](#实际应用)

## 基础概念

### 1.1 GMX 简介

GMX 是去中心化永续合约交易所，采用独特的GLP多资产池模型，提供零滑点交易、高达50倍杠杆和实时价格发现机制。

```bash
# 安装GMX相关依赖
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

// GMX 核心合约地址 (Arbitrum)
var (
    // GMX Token
    GMXTokenAddress = common.HexToAddress("0xfc5A1A6EB076a2C7aD06eD22C90d7E710E35ad0a")
    
    // GLP Token
    GLPTokenAddress = common.HexToAddress("0x4277f8F2c384827B5273592FF7CeBd9f2C1ac258")
    
    // Vault (核心合约)
    VaultAddress = common.HexToAddress("0x489ee077994B6658eAfA855C308275EAd8097C4A")
    
    // Router
    RouterAddress = common.HexToAddress("0xaBBc5F99639c9B6bCb58544ddf04EFA6802F4064")
    
    // Position Router
    PositionRouterAddress = common.HexToAddress("0xb87a436B93fFE9D75c5cFA7bAcFff96430b09868")
    
    // GLP Manager
    GLPManagerAddress = common.HexToAddress("0x3963FfC9dff443c2A94f21b129D429891E32ec18")
    
    // Reward Router
    RewardRouterAddress = common.HexToAddress("0xA906F338CB21815cBc4Bc87ace9e68c87eF8d8F1")
    
    // Fee GLP Tracker
    FeeGLPTrackerAddress = common.HexToAddress("0x4e971a87900b931fF39d1Aad67697F49835400b6")
    
    // Staked GLP Tracker
    StakedGLPTrackerAddress = common.HexToAddress("0x1aDDD80E6039594eE970E5872D247bf0414C8903")
    
    // USDG (GMX稳定币)
    USDGAddress = common.HexToAddress("0x45096e7aA921f27590f8F19e457794EB09678141")
)

// 支持的代币地址 (Arbitrum)
var (
    WETHAddress  = common.HexToAddress("0x82aF49447D8a07e3bd95BD0d56f35241523fBab1")
    WBTCAddress  = common.HexToAddress("0x2f2a2543B76A4166549F7aaB2e75Bef0aefC5B0f")
    USDCAddress  = common.HexToAddress("0xFF970A61A04b1cA14834A43f5dE4533eBDDB5CC8")
    USDTAddress  = common.HexToAddress("0xFd086bC7CD5C481DCC9C85ebE478A1C0b69FCbb9")
    DAIAddress   = common.HexToAddress("0xDA10009cBd5D07dd0CeCc66161FC93D7c9000da1")
    FRAXAddress  = common.HexToAddress("0x17FC002b466eEc40DaE837Fc4bE5c67993ddBd6F")
    LINKAddress  = common.HexToAddress("0xf97f4df75117a78c1A5a0DBb814Af92458539FB4")
    UNIAddress   = common.HexToAddress("0xFa7F8980b0f1E64A2062791cc3b0871572f1F7f0")
)

// 持仓信息
type Position struct {
    Account         common.Address
    CollateralToken common.Address
    IndexToken      common.Address
    IsLong          bool
    Size            *big.Int        // 持仓大小 (USD, 30位精度)
    Collateral      *big.Int        // 抵押品 (USD, 30位精度)
    AveragePrice    *big.Int        // 平均价格 (30位精度)
    EntryFundingRate *big.Int       // 入场资金费率
    ReserveAmount   *big.Int        // 保留数量
    RealisedPnl     *big.Int        // 已实现盈亏
    LastIncreasedTime *big.Int      // 最后增加时间
}

// GLP信息
type GLPInfo struct {
    TotalSupply     *big.Int                    // 总供应量
    AUM             *big.Int                    // 管理资产总值
    Price           *big.Int                    // GLP价格
    Composition     map[common.Address]*big.Int // 资产组成
    TargetWeights   map[common.Address]*big.Int // 目标权重
    CurrentWeights  map[common.Address]*big.Int // 当前权重
    UtilizationRate decimal.Decimal             // 利用率
}

// 交易费用
type TradingFees struct {
    PositionFee     decimal.Decimal // 开仓/平仓费用
    SwapFee         decimal.Decimal // 兑换费用
    FundingRate     decimal.Decimal // 资金费率
    BorrowingRate   decimal.Decimal // 借贷费率
    LiquidationFee  decimal.Decimal // 清算费用
}

// 市场信息
type MarketInfo struct {
    Token           common.Address
    Symbol          string
    MaxLeverage     *big.Int
    MaxGlobalLong   *big.Int
    MaxGlobalShort  *big.Int
    GuaranteedUsd   *big.Int
    PoolAmount      *big.Int
    ReservedAmount  *big.Int
    UtilizationRate decimal.Decimal
    FundingRate     decimal.Decimal
    MaxPrice        *big.Int
    MinPrice        *big.Int
}

// 奖励信息
type RewardInfo struct {
    GMXRewards      *big.Int        // GMX奖励
    ETHRewards      *big.Int        // ETH奖励
    EsGMXRewards    *big.Int        // esGMX奖励
    MultiplierPoints *big.Int       // 乘数点数
    StakedGMX       *big.Int        // 质押的GMX
    StakedGLP       *big.Int        // 质押的GLP
    TotalRewards    *big.Int        // 总奖励
}
```

## 环境准备

### 2.1 合约ABI定义

```go
// contracts/gmx_abi.go
package contracts

// Vault ABI (简化版)
const VaultABI = `[
    {
        "inputs": [
            {"name": "_account", "type": "address"},
            {"name": "_collateralToken", "type": "address"},
            {"name": "_indexToken", "type": "address"},
            {"name": "_isLong", "type": "bool"}
        ],
        "name": "getPosition",
        "outputs": [
            {"name": "", "type": "uint256"},
            {"name": "", "type": "uint256"},
            {"name": "", "type": "uint256"},
            {"name": "", "type": "uint256"},
            {"name": "", "type": "uint256"},
            {"name": "", "type": "uint256"},
            {"name": "", "type": "bool"},
            {"name": "", "type": "uint256"}
        ],
        "type": "function"
    },
    {
        "inputs": [{"name": "_token", "type": "address"}],
        "name": "getMaxPrice",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    },
    {
        "inputs": [{"name": "_token", "type": "address"}],
        "name": "getMinPrice",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    },
    {
        "inputs": [{"name": "_token", "type": "address"}],
        "name": "poolAmounts",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    },
    {
        "inputs": [{"name": "_token", "type": "address"}],
        "name": "reservedAmounts",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    },
    {
        "inputs": [{"name": "_token", "type": "address"}],
        "name": "guaranteedUsd",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    }
]`

// Position Router ABI (简化版)
const PositionRouterABI = `[
    {
        "inputs": [
            {"name": "_path", "type": "address[]"},
            {"name": "_indexToken", "type": "address"},
            {"name": "_amountIn", "type": "uint256"},
            {"name": "_minOut", "type": "uint256"},
            {"name": "_sizeDelta", "type": "uint256"},
            {"name": "_isLong", "type": "bool"},
            {"name": "_acceptablePrice", "type": "uint256"},
            {"name": "_executionFee", "type": "uint256"},
            {"name": "_referralCode", "type": "bytes32"},
            {"name": "_callbackTarget", "type": "address"}
        ],
        "name": "createIncreasePosition",
        "outputs": [{"name": "", "type": "bytes32"}],
        "payable": true,
        "type": "function"
    },
    {
        "inputs": [
            {"name": "_path", "type": "address[]"},
            {"name": "_indexToken", "type": "address"},
            {"name": "_collateralDelta", "type": "uint256"},
            {"name": "_sizeDelta", "type": "uint256"},
            {"name": "_isLong", "type": "bool"},
            {"name": "_receiver", "type": "address"},
            {"name": "_acceptablePrice", "type": "uint256"},
            {"name": "_minOut", "type": "uint256"},
            {"name": "_executionFee", "type": "uint256"},
            {"name": "_withdrawETH", "type": "bool"},
            {"name": "_callbackTarget", "type": "address"}
        ],
        "name": "createDecreasePosition",
        "outputs": [{"name": "", "type": "bytes32"}],
        "payable": true,
        "type": "function"
    }
]`

// GLP Manager ABI (简化版)
const GLPManagerABI = `[
    {
        "inputs": [],
        "name": "getAum",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    },
    {
        "inputs": [{"name": "_token", "type": "address"}],
        "name": "getAumInUsdg",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    },
    {
        "inputs": [
            {"name": "_token", "type": "address"},
            {"name": "_amount", "type": "uint256"},
            {"name": "_minUsdg", "type": "uint256"},
            {"name": "_minGlp", "type": "uint256"}
        ],
        "name": "addLiquidity",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    },
    {
        "inputs": [
            {"name": "_tokenOut", "type": "address"},
            {"name": "_glpAmount", "type": "uint256"},
            {"name": "_minOut", "type": "uint256"},
            {"name": "_receiver", "type": "address"}
        ],
        "name": "removeLiquidity",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    }
]`
```

### 2.2 GMX客户端设置

```go
// client/gmx_client.go
package client

import (
    "context"
    "math/big"
    
    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
)

type GMXClient struct {
    ethClient         *ethclient.Client
    vaultABI          abi.ABI
    positionRouterABI abi.ABI
    glpManagerABI     abi.ABI
    rewardRouterABI   abi.ABI
}

func NewGMXClient(rpcURL string) (*GMXClient, error) {
    ethClient, err := ethclient.Dial(rpcURL)
    if err != nil {
        return nil, err
    }
    
    vaultABI, err := abi.JSON(strings.NewReader(VaultABI))
    if err != nil {
        return nil, err
    }
    
    positionRouterABI, err := abi.JSON(strings.NewReader(PositionRouterABI))
    if err != nil {
        return nil, err
    }
    
    glpManagerABI, err := abi.JSON(strings.NewReader(GLPManagerABI))
    if err != nil {
        return nil, err
    }
    
    return &GMXClient{
        ethClient:         ethClient,
        vaultABI:          vaultABI,
        positionRouterABI: positionRouterABI,
        glpManagerABI:     glpManagerABI,
    }, nil
}

// 获取持仓信息
func (c *GMXClient) GetPosition(
    account common.Address,
    collateralToken common.Address,
    indexToken common.Address,
    isLong bool,
) (*Position, error) {
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    var result []interface{}
    err := c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &VaultAddress,
        Data: c.vaultABI.Methods["getPosition"].ID,
    }, nil)
    if err != nil {
        return nil, err
    }
    
    position := &Position{
        Account:         account,
        CollateralToken: collateralToken,
        IndexToken:      indexToken,
        IsLong:          isLong,
        Size:            result[0].(*big.Int),
        Collateral:      result[1].(*big.Int),
        AveragePrice:    result[2].(*big.Int),
        EntryFundingRate: result[3].(*big.Int),
        ReserveAmount:   result[4].(*big.Int),
        RealisedPnl:     result[5].(*big.Int),
        LastIncreasedTime: result[7].(*big.Int),
    }
    
    return position, nil
}

// 获取代币价格
func (c *GMXClient) GetTokenPrice(token common.Address, isMax bool) (*big.Int, error) {
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    var price *big.Int
    var err error
    
    if isMax {
        err = c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
            To: &VaultAddress,
            Data: c.vaultABI.Methods["getMaxPrice"].ID,
        }, nil)
    } else {
        err = c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
            To: &VaultAddress,
            Data: c.vaultABI.Methods["getMinPrice"].ID,
        }, nil)
    }
    
    if err != nil {
        return nil, err
    }
    
    return price, nil
}

// 获取GLP信息
func (c *GMXClient) GetGLPInfo() (*GLPInfo, error) {
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    // 获取AUM
    var aum *big.Int
    err := c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &GLPManagerAddress,
        Data: c.glpManagerABI.Methods["getAum"].ID,
    }, nil)
    if err != nil {
        return nil, err
    }
    
    // 获取GLP总供应量
    var totalSupply *big.Int
    err = c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &GLPTokenAddress,
        Data: []byte{}, // ERC20 totalSupply方法
    }, nil)
    if err != nil {
        return nil, err
    }
    
    // 计算GLP价格
    price := new(big.Int).Div(
        new(big.Int).Mul(aum, big.NewInt(1e18)),
        totalSupply,
    )
    
    glpInfo := &GLPInfo{
        TotalSupply: totalSupply,
        AUM:         aum,
        Price:       price,
        Composition: make(map[common.Address]*big.Int),
    }
    
    return glpInfo, nil
}

// 获取市场信息
func (c *GMXClient) GetMarketInfo(token common.Address) (*MarketInfo, error) {
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    // 获取池子数量
    var poolAmount *big.Int
    err := c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &VaultAddress,
        Data: c.vaultABI.Methods["poolAmounts"].ID,
    }, nil)
    if err != nil {
        return nil, err
    }
    
    // 获取保留数量
    var reservedAmount *big.Int
    err = c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &VaultAddress,
        Data: c.vaultABI.Methods["reservedAmounts"].ID,
    }, nil)
    if err != nil {
        return nil, err
    }
    
    // 获取保证金USD
    var guaranteedUsd *big.Int
    err = c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &VaultAddress,
        Data: c.vaultABI.Methods["guaranteedUsd"].ID,
    }, nil)
    if err != nil {
        return nil, err
    }
    
    // 获取价格
    maxPrice, _ := c.GetTokenPrice(token, true)
    minPrice, _ := c.GetTokenPrice(token, false)
    
    // 计算利用率
    utilizationRate := decimal.Zero
    if poolAmount.Cmp(big.NewInt(0)) > 0 {
        utilizationRate = decimal.NewFromBigInt(reservedAmount, 0).Div(
            decimal.NewFromBigInt(poolAmount, 0),
        )
    }
    
    marketInfo := &MarketInfo{
        Token:           token,
        PoolAmount:      poolAmount,
        ReservedAmount:  reservedAmount,
        GuaranteedUsd:   guaranteedUsd,
        MaxPrice:        maxPrice,
        MinPrice:        minPrice,
        UtilizationRate: utilizationRate,
    }
    
    return marketInfo, nil
}
```

## 永续合约交易

### 3.1 交易服务

```go
// services/trading_service.go
package services

import (
    "context"
    "crypto/ecdsa"
    "math/big"
    
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/shopspring/decimal"
)

type GMXTradingService struct {
    client     *GMXClient
    privateKey *ecdsa.PrivateKey
}

func NewGMXTradingService(client *GMXClient, privateKey *ecdsa.PrivateKey) *GMXTradingService {
    return &GMXTradingService{
        client:     client,
        privateKey: privateKey,
    }
}

// 开多仓
func (s *GMXTradingService) OpenLongPosition(
    collateralToken common.Address,
    indexToken common.Address,
    collateralAmount *big.Int,
    sizeDelta *big.Int,
    acceptablePrice *big.Int,
) (*types.Transaction, error) {
    return s.createPosition(
        []common.Address{collateralToken},
        indexToken,
        collateralAmount,
        sizeDelta,
        true,
        acceptablePrice,
    )
}

// 开空仓
func (s *GMXTradingService) OpenShortPosition(
    collateralToken common.Address,
    indexToken common.Address,
    collateralAmount *big.Int,
    sizeDelta *big.Int,
    acceptablePrice *big.Int,
) (*types.Transaction, error) {
    return s.createPosition(
        []common.Address{collateralToken},
        indexToken,
        collateralAmount,
        sizeDelta,
        false,
        acceptablePrice,
    )
}

// 增加持仓
func (s *GMXTradingService) IncreasePosition(
    path []common.Address,
    indexToken common.Address,
    amountIn *big.Int,
    sizeDelta *big.Int,
    isLong bool,
    acceptablePrice *big.Int,
) (*types.Transaction, error) {
    return s.createPosition(path, indexToken, amountIn, sizeDelta, isLong, acceptablePrice)
}

// 减少持仓
func (s *GMXTradingService) DecreasePosition(
    path []common.Address,
    indexToken common.Address,
    collateralDelta *big.Int,
    sizeDelta *big.Int,
    isLong bool,
    acceptablePrice *big.Int,
) (*types.Transaction, error) {
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 获取执行费用
    executionFee := big.NewInt(200000000000000) // 0.0002 ETH
    
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
    
    // 构建减少持仓交易数据
    data, err := s.client.positionRouterABI.Pack(
        "createDecreasePosition",
        path,
        indexToken,
        collateralDelta,
        sizeDelta,
        isLong,
        fromAddress,
        acceptablePrice,
        big.NewInt(0), // minOut
        executionFee,
        false, // withdrawETH
        common.Address{}, // callbackTarget
    )
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(
        nonce,
        PositionRouterAddress,
        executionFee,
        500000,
        gasPrice,
        data,
    )
    
    return s.signAndSendTransaction(tx)
}

// 平仓
func (s *GMXTradingService) ClosePosition(
    collateralToken common.Address,
    indexToken common.Address,
    isLong bool,
) (*types.Transaction, error) {
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 获取当前持仓
    position, err := s.client.GetPosition(fromAddress, collateralToken, indexToken, isLong)
    if err != nil {
        return nil, err
    }
    
    if position.Size.Cmp(big.NewInt(0)) == 0 {
        return nil, fmt.Errorf("没有持仓需要平仓")
    }
    
    // 获取当前价格作为可接受价格
    acceptablePrice, err := s.client.GetTokenPrice(indexToken, !isLong)
    if err != nil {
        return nil, err
    }
    
    // 平仓 = 减少全部持仓
    return s.DecreasePosition(
        []common.Address{collateralToken},
        indexToken,
        position.Collateral, // 全部抵押品
        position.Size,       // 全部持仓大小
        isLong,
        acceptablePrice,
    )
}

// 计算持仓盈亏
func (s *GMXTradingService) CalculatePositionPnL(
    position *Position,
    currentPrice *big.Int,
) (*big.Int, error) {
    if position.Size.Cmp(big.NewInt(0)) == 0 {
        return big.NewInt(0), nil
    }
    
    // 计算价格差异
    priceDelta := new(big.Int)
    if position.IsLong {
        // 多仓: 当前价格 - 平均价格
        priceDelta.Sub(currentPrice, position.AveragePrice)
    } else {
        // 空仓: 平均价格 - 当前价格
        priceDelta.Sub(position.AveragePrice, currentPrice)
    }
    
    // 计算盈亏 = (价格差异 / 平均价格) * 持仓大小
    if position.AveragePrice.Cmp(big.NewInt(0)) == 0 {
        return big.NewInt(0), nil
    }
    
    pnl := new(big.Int).Mul(priceDelta, position.Size)
    pnl.Div(pnl, position.AveragePrice)
    
    return pnl, nil
}

// 计算清算价格
func (s *GMXTradingService) CalculateLiquidationPrice(position *Position) (*big.Int, error) {
    if position.Size.Cmp(big.NewInt(0)) == 0 {
        return big.NewInt(0), nil
    }
    
    // 简化的清算价格计算
    // 实际计算需要考虑资金费率、借贷费率等
    
    // 清算阈值 (90%的抵押品)
    liquidationThreshold := new(big.Int).Mul(position.Collateral, big.NewInt(90))
    liquidationThreshold.Div(liquidationThreshold, big.NewInt(100))
    
    var liquidationPrice *big.Int
    
    if position.IsLong {
        // 多仓清算价格 = 平均价格 - (清算阈值 * 平均价格 / 持仓大小)
        priceDrop := new(big.Int).Mul(liquidationThreshold, position.AveragePrice)
        priceDrop.Div(priceDrop, position.Size)
        liquidationPrice = new(big.Int).Sub(position.AveragePrice, priceDrop)
    } else {
        // 空仓清算价格 = 平均价格 + (清算阈值 * 平均价格 / 持仓大小)
        priceRise := new(big.Int).Mul(liquidationThreshold, position.AveragePrice)
        priceRise.Div(priceRise, position.Size)
        liquidationPrice = new(big.Int).Add(position.AveragePrice, priceRise)
    }
    
    return liquidationPrice, nil
}

// 获取最优杠杆
func (s *GMXTradingService) GetOptimalLeverage(
    collateralAmount *big.Int,
    indexToken common.Address,
    riskTolerance decimal.Decimal,
) (*big.Int, error) {
    // 获取市场信息
    marketInfo, err := s.client.GetMarketInfo(indexToken)
    if err != nil {
        return nil, err
    }
    
    // 基于风险承受能力计算最优杠杆
    // 风险承受能力: 0.1 (保守) 到 1.0 (激进)
    
    maxLeverage := decimal.NewFromInt(50) // GMX最大50倍杠杆
    utilizationPenalty := marketInfo.UtilizationRate.Mul(decimal.NewFromFloat(0.5))
    
    optimalLeverage := maxLeverage.Mul(riskTolerance).Sub(utilizationPenalty)
    
    // 确保杠杆在合理范围内
    if optimalLeverage.LessThan(decimal.NewFromInt(1)) {
        optimalLeverage = decimal.NewFromInt(1)
    }
    if optimalLeverage.GreaterThan(maxLeverage) {
        optimalLeverage = maxLeverage
    }
    
    return optimalLeverage.BigInt(), nil
}

// 辅助函数
func (s *GMXTradingService) createPosition(
    path []common.Address,
    indexToken common.Address,
    amountIn *big.Int,
    sizeDelta *big.Int,
    isLong bool,
    acceptablePrice *big.Int,
) (*types.Transaction, error) {
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 获取执行费用
    executionFee := big.NewInt(200000000000000) // 0.0002 ETH
    
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
    
    // 构建增加持仓交易数据
    data, err := s.client.positionRouterABI.Pack(
        "createIncreasePosition",
        path,
        indexToken,
        amountIn,
        big.NewInt(0), // minOut
        sizeDelta,
        isLong,
        acceptablePrice,
        executionFee,
        [32]byte{}, // referralCode
        common.Address{}, // callbackTarget
    )
    if err != nil {
        return nil, err
    }
    
    // 计算总价值 (抵押品 + 执行费用)
    totalValue := new(big.Int).Add(amountIn, executionFee)
    
    tx := types.NewTransaction(
        nonce,
        PositionRouterAddress,
        totalValue,
        500000,
        gasPrice,
        data,
    )
    
    return s.signAndSendTransaction(tx)
}

func (s *GMXTradingService) signAndSendTransaction(tx *types.Transaction) (*types.Transaction, error) {
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

## GLP池机制

### 4.1 流动性服务

```go
// services/liquidity_service.go
package services

import (
    "context"
    "crypto/ecdsa"
    "math/big"
    
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/shopspring/decimal"
)

type GLPLiquidityService struct {
    client     *GMXClient
    privateKey *ecdsa.PrivateKey
}

func NewGLPLiquidityService(client *GMXClient, privateKey *ecdsa.PrivateKey) *GLPLiquidityService {
    return &GLPLiquidityService{
        client:     client,
        privateKey: privateKey,
    }
}

// 添加流动性 (铸造GLP)
func (s *GLPLiquidityService) AddLiquidity(
    token common.Address,
    amount *big.Int,
    minUsdg *big.Int,
    minGlp *big.Int,
) (*types.Transaction, error) {
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 授权GLP Manager使用代币
    if err := s.approveToken(token, GLPManagerAddress, amount); err != nil {
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
    
    // 构建添加流动性交易数据
    data, err := s.client.glpManagerABI.Pack(
        "addLiquidity",
        token,
        amount,
        minUsdg,
        minGlp,
    )
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, GLPManagerAddress, big.NewInt(0), 300000, gasPrice, data)
    
    return s.signAndSendTransaction(tx)
}

// 移除流动性 (销毁GLP)
func (s *GLPLiquidityService) RemoveLiquidity(
    tokenOut common.Address,
    glpAmount *big.Int,
    minOut *big.Int,
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
    
    // 构建移除流动性交易数据
    data, err := s.client.glpManagerABI.Pack(
        "removeLiquidity",
        tokenOut,
        glpAmount,
        minOut,
        fromAddress,
    )
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, GLPManagerAddress, big.NewInt(0), 300000, gasPrice, data)
    
    return s.signAndSendTransaction(tx)
}

// 计算GLP铸造数量
func (s *GLPLiquidityService) CalculateGLPMintAmount(
    token common.Address,
    amount *big.Int,
) (*GLPMintInfo, error) {
    // 获取代币价格
    tokenPrice, err := s.client.GetTokenPrice(token, false) // 使用最小价格
    if err != nil {
        return nil, err
    }
    
    // 获取GLP信息
    glpInfo, err := s.client.GetGLPInfo()
    if err != nil {
        return nil, err
    }
    
    // 计算USD价值
    usdValue := new(big.Int).Mul(amount, tokenPrice)
    usdValue.Div(usdValue, big.NewInt(1e30)) // 调整精度
    
    // 计算GLP数量 = USD价值 / GLP价格
    glpAmount := new(big.Int).Div(
        new(big.Int).Mul(usdValue, big.NewInt(1e18)),
        glpInfo.Price,
    )
    
    // 计算费用 (基于池子权重)
    fee := s.calculateMintFee(token, amount)
    
    // 扣除费用后的GLP数量
    netGLPAmount := new(big.Int).Sub(glpAmount, fee)
    
    return &GLPMintInfo{
        TokenAmount:    amount,
        USDValue:       usdValue,
        GLPAmount:      netGLPAmount,
        Fee:            fee,
        GLPPrice:       glpInfo.Price,
        PriceImpact:    s.calculatePriceImpact(token, amount),
    }, nil
}

// 计算GLP赎回数量
func (s *GLPLiquidityService) CalculateGLPRedeemAmount(
    tokenOut common.Address,
    glpAmount *big.Int,
) (*GLPRedeemInfo, error) {
    // 获取GLP信息
    glpInfo, err := s.client.GetGLPInfo()
    if err != nil {
        return nil, err
    }
    
    // 获取代币价格
    tokenPrice, err := s.client.GetTokenPrice(tokenOut, true) // 使用最大价格
    if err != nil {
        return nil, err
    }
    
    // 计算USD价值
    usdValue := new(big.Int).Mul(glpAmount, glpInfo.Price)
    usdValue.Div(usdValue, big.NewInt(1e18))
    
    // 计算代币数量 = USD价值 / 代币价格
    tokenAmount := new(big.Int).Div(
        new(big.Int).Mul(usdValue, big.NewInt(1e30)),
        tokenPrice,
    )
    
    // 计算费用
    fee := s.calculateRedeemFee(tokenOut, tokenAmount)
    
    // 扣除费用后的代币数量
    netTokenAmount := new(big.Int).Sub(tokenAmount, fee)
    
    return &GLPRedeemInfo{
        GLPAmount:      glpAmount,
        USDValue:       usdValue,
        TokenAmount:    netTokenAmount,
        Fee:            fee,
        GLPPrice:       glpInfo.Price,
        PriceImpact:    s.calculatePriceImpact(tokenOut, tokenAmount),
    }, nil
}

// 获取最佳铸造代币
func (s *GLPLiquidityService) GetBestMintToken(amount *big.Int) (*TokenRecommendation, error) {
    supportedTokens := []common.Address{
        WETHAddress, WBTCAddress, USDCAddress, USDTAddress, DAIAddress,
    }
    
    var bestToken common.Address
    var lowestFee decimal.Decimal = decimal.NewFromInt(100) // 100%
    
    for _, token := range supportedTokens {
        mintInfo, err := s.CalculateGLPMintAmount(token, amount)
        if err != nil {
            continue
        }
        
        feeRate := decimal.NewFromBigInt(mintInfo.Fee, -18).Div(
            decimal.NewFromBigInt(mintInfo.GLPAmount, -18),
        )
        
        if feeRate.LessThan(lowestFee) {
            lowestFee = feeRate
            bestToken = token
        }
    }
    
    return &TokenRecommendation{
        Token:       bestToken,
        FeeRate:     lowestFee,
        Reason:      "最低铸造费用",
    }, nil
}

// 获取最佳赎回代币
func (s *GLPLiquidityService) GetBestRedeemToken(glpAmount *big.Int) (*TokenRecommendation, error) {
    supportedTokens := []common.Address{
        WETHAddress, WBTCAddress, USDCAddress, USDTAddress, DAIAddress,
    }
    
    var bestToken common.Address
    var lowestFee decimal.Decimal = decimal.NewFromInt(100) // 100%
    
    for _, token := range supportedTokens {
        redeemInfo, err := s.CalculateGLPRedeemAmount(token, glpAmount)
        if err != nil {
            continue
        }
        
        feeRate := decimal.NewFromBigInt(redeemInfo.Fee, -18).Div(
            decimal.NewFromBigInt(redeemInfo.TokenAmount, -18),
        )
        
        if feeRate.LessThan(lowestFee) {
            lowestFee = feeRate
            bestToken = token
        }
    }
    
    return &TokenRecommendation{
        Token:       bestToken,
        FeeRate:     lowestFee,
        Reason:      "最低赎回费用",
    }, nil
}

// 辅助函数
func (s *GLPLiquidityService) calculateMintFee(token common.Address, amount *big.Int) *big.Int {
    // 简化的费用计算
    // 实际费用基于池子权重和目标权重的差异
    baseFee := decimal.NewFromFloat(0.003) // 0.3%基础费用
    feeAmount := decimal.NewFromBigInt(amount, -18).Mul(baseFee)
    return feeAmount.Mul(decimal.NewFromInt(1e18)).BigInt()
}

func (s *GLPLiquidityService) calculateRedeemFee(token common.Address, amount *big.Int) *big.Int {
    // 简化的费用计算
    baseFee := decimal.NewFromFloat(0.003) // 0.3%基础费用
    feeAmount := decimal.NewFromBigInt(amount, -18).Mul(baseFee)
    return feeAmount.Mul(decimal.NewFromInt(1e18)).BigInt()
}

func (s *GLPLiquidityService) calculatePriceImpact(token common.Address, amount *big.Int) decimal.Decimal {
    // 简化的价格影响计算
    // 实际计算需要考虑池子深度和权重
    return decimal.NewFromFloat(0.001) // 0.1%
}

func (s *GLPLiquidityService) approveToken(token, spender common.Address, amount *big.Int) error {
    // 实现ERC20授权逻辑
    return nil
}

func (s *GLPLiquidityService) signAndSendTransaction(tx *types.Transaction) (*types.Transaction, error) {
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

type GLPMintInfo struct {
    TokenAmount *big.Int
    USDValue    *big.Int
    GLPAmount   *big.Int
    Fee         *big.Int
    GLPPrice    *big.Int
    PriceImpact decimal.Decimal
}

type GLPRedeemInfo struct {
    GLPAmount   *big.Int
    USDValue    *big.Int
    TokenAmount *big.Int
    Fee         *big.Int
    GLPPrice    *big.Int
    PriceImpact decimal.Decimal
}

type TokenRecommendation struct {
    Token   common.Address
    FeeRate decimal.Decimal
    Reason  string
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
    "github.com/shopspring/decimal"
    
    "your-project/client"
    "your-project/services"
)

func main() {
    // 创建GMX客户端 (Arbitrum)
    gmxClient, err := client.NewGMXClient("https://arb1.arbitrum.io/rpc")
    if err != nil {
        log.Fatal("创建GMX客户端失败:", err)
    }
    
    // 加载私钥
    privateKey, err := crypto.HexToECDSA("your_private_key_here")
    if err != nil {
        log.Fatal("加载私钥失败:", err)
    }
    
    // 获取用户地址
    userAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
    
    // 创建服务
    tradingService := services.NewGMXTradingService(gmxClient, privateKey)
    liquidityService := services.NewGLPLiquidityService(gmxClient, privateKey)
    
    // 1. 获取GLP信息
    fmt.Printf("=== GLP池信息 ===\n")
    
    glpInfo, err := gmxClient.GetGLPInfo()
    if err != nil {
        log.Fatal("获取GLP信息失败:", err)
    }
    
    fmt.Printf("GLP总供应量: %s\n", glpInfo.TotalSupply.String())
    fmt.Printf("管理资产总值: %s USD\n", 
        decimal.NewFromBigInt(glpInfo.AUM, -30).String())
    fmt.Printf("GLP价格: %s USD\n", 
        decimal.NewFromBigInt(glpInfo.Price, -18).String())
    fmt.Printf("利用率: %s%%\n", 
        glpInfo.UtilizationRate.Mul(decimal.NewFromInt(100)).String())
    
    // 2. 查询市场信息
    fmt.Printf("\n=== 市场信息 ===\n")
    
    markets := []struct {
        Token  common.Address
        Symbol string
    }{
        {client.WETHAddress, "ETH"},
        {client.WBTCAddress, "BTC"},
        {client.USDCAddress, "USDC"},
    }
    
    for _, market := range markets {
        marketInfo, err := gmxClient.GetMarketInfo(market.Token)
        if err != nil {
            log.Printf("获取%s市场信息失败: %v", market.Symbol, err)
            continue
        }
        
        fmt.Printf("%s市场:\n", market.Symbol)
        fmt.Printf("  池子数量: %s\n", marketInfo.PoolAmount.String())
        fmt.Printf("  保留数量: %s\n", marketInfo.ReservedAmount.String())
        fmt.Printf("  利用率: %s%%\n", 
            marketInfo.UtilizationRate.Mul(decimal.NewFromInt(100)).String())
        fmt.Printf("  最大价格: %s USD\n", 
            decimal.NewFromBigInt(marketInfo.MaxPrice, -30).String())
        fmt.Printf("  最小价格: %s USD\n", 
            decimal.NewFromBigInt(marketInfo.MinPrice, -30).String())
    }
    
    // 3. 查询用户持仓
    fmt.Printf("\n=== 用户持仓 ===\n")
    
    positions := []struct {
        CollateralToken common.Address
        IndexToken      common.Address
        IsLong          bool
        Name            string
    }{
        {client.USDCAddress, client.WETHAddress, true, "ETH多仓"},
        {client.USDCAddress, client.WETHAddress, false, "ETH空仓"},
        {client.USDCAddress, client.WBTCAddress, true, "BTC多仓"},
        {client.USDCAddress, client.WBTCAddress, false, "BTC空仓"},
    }
    
    for _, pos := range positions {
        position, err := gmxClient.GetPosition(
            userAddress,
            pos.CollateralToken,
            pos.IndexToken,
            pos.IsLong,
        )
        if err != nil {
            log.Printf("获取%s持仓失败: %v", pos.Name, err)
            continue
        }
        
        if position.Size.Cmp(big.NewInt(0)) > 0 {
            fmt.Printf("%s:\n", pos.Name)
            fmt.Printf("  持仓大小: %s USD\n", 
                decimal.NewFromBigInt(position.Size, -30).String())
            fmt.Printf("  抵押品: %s USD\n", 
                decimal.NewFromBigInt(position.Collateral, -30).String())
            fmt.Printf("  平均价格: %s USD\n", 
                decimal.NewFromBigInt(position.AveragePrice, -30).String())
            
            // 计算当前盈亏
            currentPrice, err := gmxClient.GetTokenPrice(pos.IndexToken, pos.IsLong)
            if err == nil {
                pnl, err := tradingService.CalculatePositionPnL(position, currentPrice)
                if err == nil {
                    fmt.Printf("  未实现盈亏: %s USD\n", 
                        decimal.NewFromBigInt(pnl, -30).String())
                }
                
                // 计算清算价格
                liquidationPrice, err := tradingService.CalculateLiquidationPrice(position)
                if err == nil {
                    fmt.Printf("  清算价格: %s USD\n", 
                        decimal.NewFromBigInt(liquidationPrice, -30).String())
                }
            }
        } else {
            fmt.Printf("%s: 无持仓\n", pos.Name)
        }
    }
    
    // 4. GLP流动性操作示例
    fmt.Printf("\n=== GLP流动性操作示例 ===\n")
    
    // 计算添加流动性
    addAmount := big.NewInt(1000e6) // 1000 USDC
    
    mintInfo, err := liquidityService.CalculateGLPMintAmount(client.USDCAddress, addAmount)
    if err != nil {
        log.Printf("计算GLP铸造失败: %v", err)
    } else {
        fmt.Printf("添加 %s USDC 流动性:\n", 
            decimal.NewFromBigInt(addAmount, -6).String())
        fmt.Printf("  获得GLP: %s\n", 
            decimal.NewFromBigInt(mintInfo.GLPAmount, -18).String())
        fmt.Printf("  费用: %s USD\n", 
            decimal.NewFromBigInt(mintInfo.Fee, -18).String())
        fmt.Printf("  价格影响: %s%%\n", 
            mintInfo.PriceImpact.Mul(decimal.NewFromInt(100)).String())
    }
    
    // 获取最佳铸造代币建议
    bestMintToken, err := liquidityService.GetBestMintToken(big.NewInt(1000e18))
    if err != nil {
        log.Printf("获取最佳铸造代币失败: %v", err)
    } else {
        fmt.Printf("最佳铸造代币: %s\n", bestMintToken.Token.Hex())
        fmt.Printf("费用率: %s%%\n", 
            bestMintToken.FeeRate.Mul(decimal.NewFromInt(100)).String())
        fmt.Printf("原因: %s\n", bestMintToken.Reason)
    }
    
    // 5. 永续合约交易示例
    fmt.Printf("\n=== 永续合约交易示例 ===\n")
    
    // 计算最优杠杆
    collateralAmount := big.NewInt(500e6) // 500 USDC作为抵押品
    riskTolerance := decimal.NewFromFloat(0.3) // 30%风险承受能力
    
    optimalLeverage, err := tradingService.GetOptimalLeverage(
        collateralAmount,
        client.WETHAddress,
        riskTolerance,
    )
    if err != nil {
        log.Printf("计算最优杠杆失败: %v", err)
    } else {
        fmt.Printf("基于风险承受能力的最优杠杆: %sx\n", optimalLeverage.String())
    }
    
    // 开多仓示例
    ethPrice, err := gmxClient.GetTokenPrice(client.WETHAddress, true)
    if err != nil {
        log.Printf("获取ETH价格失败: %v", err)
    } else {
        fmt.Printf("当前ETH价格: %s USD\n", 
            decimal.NewFromBigInt(ethPrice, -30).String())
        
        // 计算持仓大小 = 抵押品 * 杠杆
        positionSize := new(big.Int).Mul(collateralAmount, optimalLeverage)
        
        fmt.Printf("准备开ETH多仓:\n")
        fmt.Printf("  抵押品: %s USDC\n", 
            decimal.NewFromBigInt(collateralAmount, -6).String())
        fmt.Printf("  持仓大小: %s USD\n", 
            decimal.NewFromBigInt(positionSize, -6).String())
        fmt.Printf("  杠杆: %sx\n", optimalLeverage.String())
        
        // 设置可接受价格 (当前价格 + 0.5%滑点)
        acceptablePrice := new(big.Int).Mul(ethPrice, big.NewInt(1005))
        acceptablePrice.Div(acceptablePrice, big.NewInt(1000))
        
        // 执行开仓 (注释掉实际执行)
        // tx, err := tradingService.OpenLongPosition(
        //     client.USDCAddress,
        //     client.WETHAddress,
        //     collateralAmount,
        //     positionSize,
        //     acceptablePrice,
        // )
        // if err != nil {
        //     log.Printf("开多仓失败: %v", err)
        // } else {
        //     fmt.Printf("开仓交易已提交: %s\n", tx.Hash().Hex())
        // }
    }
    
    // 6. 费用分析
    fmt.Printf("\n=== 费用分析 ===\n")
    
    fmt.Printf("GMX费用结构:\n")
    fmt.Printf("  开仓/平仓费用: 0.1%%\n")
    fmt.Printf("  兑换费用: 0.2%% - 0.8%%\n")
    fmt.Printf("  借贷费用: 0.01%% / 小时\n")
    fmt.Printf("  资金费用: 动态调整\n")
    fmt.Printf("  GLP铸造/赎回: 0.3%% - 0.8%%\n")
    
    // 7. 风险提示
    fmt.Printf("\n=== 风险提示 ===\n")
    
    fmt.Printf("使用GMX时请注意:\n")
    fmt.Printf("  1. 高杠杆交易风险极高\n")
    fmt.Printf("  2. 关注清算价格避免强制平仓\n")
    fmt.Printf("  3. 资金费率可能影响持仓成本\n")
    fmt.Printf("  4. GLP持有者承担池子的无常损失\n")
    fmt.Printf("  5. 市场波动可能导致快速清算\n")
    fmt.Printf("  6. 确保有足够的抵押品维持持仓\n")
    
    // 8. 策略建议
    fmt.Printf("\n=== 策略建议 ===\n")
    
    fmt.Printf("GMX交易策略:\n")
    fmt.Printf("  保守策略: 2-5x杠杆，充足抵押品\n")
    fmt.Printf("  平衡策略: 5-10x杠杆，设置止损\n")
    fmt.Printf("  激进策略: 10-20x杠杆，短期交易\n")
    
    fmt.Printf("\nGLP投资策略:\n")
    fmt.Printf("  长期持有: 享受交易费用分成\n")
    fmt.Printf("  套利机会: 利用权重偏差\n")
    fmt.Printf("  风险对冲: 与交易者对手盘\n")
}
```

这个GMX使用指南提供了完整的去中心化永续合约交易所集成方案，涵盖了永续合约交易、GLP流动性提供、风险管理、套利策略等核心功能，是构建高级DeFi交易应用的重要参考文档。
