# Curve Finance 详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [环境准备](#环境准备)
3. [池子类型](#池子类型)
4. [交易操作](#交易操作)
5. [流动性提供](#流动性提供)
6. [收益挖矿](#收益挖矿)
7. [治理代币](#治理代币)
8. [跨链部署](#跨链部署)
9. [实际应用](#实际应用)

## 基础概念

### 1.1 Curve Finance 简介

Curve 是专门为稳定币和相似资产设计的去中心化交易所，通过优化的AMM算法实现极低滑点交易，是DeFi生态中最重要的稳定币交易基础设施。

```bash
# 安装Curve相关依赖
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

// Curve 主要合约地址 (Mainnet)
var (
    // 3Pool (USDC/USDT/DAI)
    ThreePoolAddress     = common.HexToAddress("0xbEbc44782C7dB0a1A60Cb6fe97d0b483032FF1C7")
    ThreePoolLPToken     = common.HexToAddress("0x6c3F90f043a72FA612cbac8115EE7e52BDe6E490")
    
    // stETH Pool
    StETHPoolAddress     = common.HexToAddress("0xDC24316b9AE028F1497c275EB9192a3Ea0f67022")
    
    // CRV Token
    CRVTokenAddress      = common.HexToAddress("0xD533a949740bb3306d119CC777fa900bA034cd52")
    
    // Gauge Controller
    GaugeControllerAddress = common.HexToAddress("0x2F50D538606Fa9EDD2B11E2446BEb18C9D5846bB")
    
    // Minter
    MinterAddress        = common.HexToAddress("0xd061D61a4d941c39E5453435B6345Dc261C2fcE0")
)

// 池子类型
type PoolType int

const (
    PlainPool PoolType = iota  // 普通池 (如3Pool)
    LendingPool               // 借贷池 (如Compound池)
    MetaPool                  // 元池 (与基础池配对)
    CryptoPool               // 加密货币池 (如ETH/BTC)
    FactoryPool              // 工厂池
)

// 池子信息
type CurvePool struct {
    Address     common.Address
    Name        string
    Type        PoolType
    Coins       []common.Address
    LPToken     common.Address
    Gauge       common.Address
    A           *big.Int  // 放大系数
    Fee         *big.Int  // 交易费用
    AdminFee    *big.Int  // 管理费用
}
```

## 环境准备

### 2.1 合约ABI定义

```go
// contracts/curve_abi.go
package contracts

// Curve Pool ABI (简化版)
const CurvePoolABI = `[
    {
        "name": "exchange",
        "inputs": [
            {"name": "i", "type": "int128"},
            {"name": "j", "type": "int128"},
            {"name": "dx", "type": "uint256"},
            {"name": "min_dy", "type": "uint256"}
        ],
        "outputs": [],
        "type": "function"
    },
    {
        "name": "get_dy",
        "inputs": [
            {"name": "i", "type": "int128"},
            {"name": "j", "type": "int128"},
            {"name": "dx", "type": "uint256"}
        ],
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    },
    {
        "name": "add_liquidity",
        "inputs": [
            {"name": "amounts", "type": "uint256[3]"},
            {"name": "min_mint_amount", "type": "uint256"}
        ],
        "outputs": [],
        "type": "function"
    },
    {
        "name": "remove_liquidity",
        "inputs": [
            {"name": "_amount", "type": "uint256"},
            {"name": "min_amounts", "type": "uint256[3]"}
        ],
        "outputs": [],
        "type": "function"
    },
    {
        "name": "remove_liquidity_one_coin",
        "inputs": [
            {"name": "_token_amount", "type": "uint256"},
            {"name": "i", "type": "int128"},
            {"name": "min_amount", "type": "uint256"}
        ],
        "outputs": [],
        "type": "function"
    },
    {
        "name": "balances",
        "inputs": [{"name": "arg0", "type": "uint256"}],
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    },
    {
        "name": "coins",
        "inputs": [{"name": "arg0", "type": "uint256"}],
        "outputs": [{"name": "", "type": "address"}],
        "type": "function"
    },
    {
        "name": "A",
        "inputs": [],
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    },
    {
        "name": "fee",
        "inputs": [],
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    }
]`

// Gauge ABI (简化版)
const GaugeABI = `[
    {
        "name": "deposit",
        "inputs": [{"name": "_value", "type": "uint256"}],
        "outputs": [],
        "type": "function"
    },
    {
        "name": "withdraw",
        "inputs": [{"name": "_value", "type": "uint256"}],
        "outputs": [],
        "type": "function"
    },
    {
        "name": "claim_rewards",
        "inputs": [],
        "outputs": [],
        "type": "function"
    },
    {
        "name": "balanceOf",
        "inputs": [{"name": "arg0", "type": "address"}],
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    },
    {
        "name": "claimable_tokens",
        "inputs": [{"name": "addr", "type": "address"}],
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    }
]`
```

### 2.2 客户端设置

```go
// client/curve_client.go
package client

import (
    "context"
    "math/big"
    
    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
)

type CurveClient struct {
    ethClient *ethclient.Client
    poolABI   abi.ABI
    gaugeABI  abi.ABI
}

func NewCurveClient(rpcURL string) (*CurveClient, error) {
    ethClient, err := ethclient.Dial(rpcURL)
    if err != nil {
        return nil, err
    }
    
    poolABI, err := abi.JSON(strings.NewReader(CurvePoolABI))
    if err != nil {
        return nil, err
    }
    
    gaugeABI, err := abi.JSON(strings.NewReader(GaugeABI))
    if err != nil {
        return nil, err
    }
    
    return &CurveClient{
        ethClient: ethClient,
        poolABI:   poolABI,
        gaugeABI:  gaugeABI,
    }, nil
}

// 获取交易输出金额
func (c *CurveClient) GetDy(poolAddress common.Address, i, j int, dx *big.Int) (*big.Int, error) {
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    var result *big.Int
    err := c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &poolAddress,
        Data: c.poolABI.Methods["get_dy"].ID,
    }, nil)
    
    if err != nil {
        return nil, err
    }
    
    return result, nil
}

// 获取池子余额
func (c *CurveClient) GetBalance(poolAddress common.Address, coinIndex int) (*big.Int, error) {
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    var result *big.Int
    err := c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &poolAddress,
        Data: c.poolABI.Methods["balances"].ID,
    }, nil)
    
    if err != nil {
        return nil, err
    }
    
    return result, nil
}
```

## 池子类型

### 3.1 稳定币池

```go
// services/stableswap_service.go
package services

import (
    "math/big"
    
    "github.com/ethereum/go-ethereum/common"
    "github.com/shopspring/decimal"
)

type StableSwapService struct {
    client *CurveClient
}

func NewStableSwapService(client *CurveClient) *StableSwapService {
    return &StableSwapService{
        client: client,
    }
}

// 3Pool 稳定币索引
const (
    DAI_INDEX  = 0
    USDC_INDEX = 1
    USDT_INDEX = 2
)

// 计算稳定币交换
func (s *StableSwapService) CalculateStableSwap(
    poolAddress common.Address,
    fromCoin, toCoin int,
    amount *big.Int,
) (*SwapCalculation, error) {
    // 获取预期输出
    expectedOutput, err := s.client.GetDy(poolAddress, fromCoin, toCoin, amount)
    if err != nil {
        return nil, err
    }
    
    // 获取池子信息
    poolInfo, err := s.getPoolInfo(poolAddress)
    if err != nil {
        return nil, err
    }
    
    // 计算价格影响
    priceImpact := s.calculatePriceImpact(poolInfo, fromCoin, toCoin, amount, expectedOutput)
    
    // 计算交易费用
    fee := s.calculateTradingFee(amount, poolInfo.Fee)
    
    return &SwapCalculation{
        InputAmount:    amount,
        OutputAmount:   expectedOutput,
        PriceImpact:    priceImpact,
        TradingFee:     fee,
        ExchangeRate:   s.calculateExchangeRate(amount, expectedOutput),
    }, nil
}

// 获取最优稳定币路径
func (s *StableSwapService) GetOptimalStablePath(
    fromToken, toToken common.Address,
    amount *big.Int,
) (*OptimalPath, error) {
    // 定义主要稳定币池
    pools := []PoolRoute{
        {
            Pool:     ThreePoolAddress,
            FromCoin: s.getTokenIndex(fromToken, ThreePoolAddress),
            ToCoin:   s.getTokenIndex(toToken, ThreePoolAddress),
        },
        // 可以添加更多池子路径
    }
    
    var bestPath *OptimalPath
    var bestOutput *big.Int = big.NewInt(0)
    
    for _, route := range pools {
        if route.FromCoin == -1 || route.ToCoin == -1 {
            continue // 池子不支持这些代币
        }
        
        output, err := s.client.GetDy(route.Pool, route.FromCoin, route.ToCoin, amount)
        if err != nil {
            continue
        }
        
        if output.Cmp(bestOutput) > 0 {
            bestOutput = output
            bestPath = &OptimalPath{
                Route:        route,
                OutputAmount: output,
            }
        }
    }
    
    return bestPath, nil
}

type SwapCalculation struct {
    InputAmount    *big.Int
    OutputAmount   *big.Int
    PriceImpact    decimal.Decimal
    TradingFee     *big.Int
    ExchangeRate   decimal.Decimal
}

type PoolRoute struct {
    Pool     common.Address
    FromCoin int
    ToCoin   int
}

type OptimalPath struct {
    Route        PoolRoute
    OutputAmount *big.Int
}
```

### 3.2 加密货币池

```go
// services/cryptoswap_service.go
package services

import (
    "math/big"
    
    "github.com/ethereum/go-ethereum/common"
    "github.com/shopspring/decimal"
)

type CryptoSwapService struct {
    client *CurveClient
}

func NewCryptoSwapService(client *CurveClient) *CryptoSwapService {
    return &CryptoSwapService{
        client: client,
    }
}

// 加密货币池特点：
// 1. 支持波动性资产 (ETH, BTC, etc.)
// 2. 动态费用
// 3. 内部预言机

// tricrypto池 (USDT/WBTC/WETH)
var TricryptoPoolAddress = common.HexToAddress("0xD51a44d3FaE010294C616388b506AcdA1bfAAE46")

const (
    TRICRYPTO_USDT_INDEX = 0
    TRICRYPTO_WBTC_INDEX = 1
    TRICRYPTO_WETH_INDEX = 2
)

// 计算加密货币交换
func (s *CryptoSwapService) CalculateCryptoSwap(
    poolAddress common.Address,
    fromCoin, toCoin int,
    amount *big.Int,
) (*CryptoSwapCalculation, error) {
    // 获取预期输出
    expectedOutput, err := s.client.GetDy(poolAddress, fromCoin, toCoin, amount)
    if err != nil {
        return nil, err
    }
    
    // 获取动态费用
    dynamicFee, err := s.getDynamicFee(poolAddress)
    if err != nil {
        return nil, err
    }
    
    // 计算价格影响 (加密货币池通常有更高的价格影响)
    priceImpact := s.calculateCryptoPriceImpact(poolAddress, fromCoin, toCoin, amount)
    
    return &CryptoSwapCalculation{
        InputAmount:    amount,
        OutputAmount:   expectedOutput,
        PriceImpact:    priceImpact,
        DynamicFee:     dynamicFee,
        Slippage:       s.calculateSlippage(amount, expectedOutput),
    }, nil
}

// 获取内部预言机价格
func (s *CryptoSwapService) GetInternalOraclePrice(
    poolAddress common.Address,
    coinIndex int,
) (*big.Int, error) {
    // Curve加密货币池有内部预言机
    // 这里需要调用特定的预言机函数
    
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    var result *big.Int
    err := s.client.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &poolAddress,
        Data: s.client.poolABI.Methods["price_oracle"].ID,
    }, nil)
    
    if err != nil {
        return nil, err
    }
    
    return result, nil
}

type CryptoSwapCalculation struct {
    InputAmount    *big.Int
    OutputAmount   *big.Int
    PriceImpact    decimal.Decimal
    DynamicFee     *big.Int
    Slippage       decimal.Decimal
}
```

## 交易操作

### 4.1 交换服务

```go
// services/exchange_service.go
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

type ExchangeService struct {
    client     *CurveClient
    privateKey *ecdsa.PrivateKey
}

func NewExchangeService(client *CurveClient, privateKey *ecdsa.PrivateKey) *ExchangeService {
    return &ExchangeService{
        client:     client,
        privateKey: privateKey,
    }
}

// 执行代币交换
func (s *ExchangeService) Exchange(
    poolAddress common.Address,
    fromCoin, toCoin int,
    amount *big.Int,
    minOutput *big.Int,
) (*types.Transaction, error) {
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 获取输入代币地址
    inputTokenAddress, err := s.getCoinAddress(poolAddress, fromCoin)
    if err != nil {
        return nil, err
    }
    
    // 授权池子使用代币
    if err := s.approveToken(inputTokenAddress, poolAddress, amount); err != nil {
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
    
    // 构建exchange交易数据
    data, err := s.client.poolABI.Pack(
        "exchange",
        big.NewInt(int64(fromCoin)),
        big.NewInt(int64(toCoin)),
        amount,
        minOutput,
    )
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, poolAddress, big.NewInt(0), 200000, gasPrice, data)
    
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

// 批量交换 (通过多个池子)
func (s *ExchangeService) MultiPoolExchange(
    routes []ExchangeRoute,
    amount *big.Int,
    minFinalOutput *big.Int,
) ([]*types.Transaction, error) {
    var transactions []*types.Transaction
    currentAmount := amount
    
    for i, route := range routes {
        // 计算最小输出 (除了最后一步)
        var minOutput *big.Int
        if i == len(routes)-1 {
            minOutput = minFinalOutput
        } else {
            // 为中间步骤设置较低的滑点保护
            expectedOutput, err := s.client.GetDy(route.Pool, route.FromCoin, route.ToCoin, currentAmount)
            if err != nil {
                return nil, err
            }
            minOutput = s.calculateMinOutput(expectedOutput, 0.5) // 0.5% 滑点
        }
        
        tx, err := s.Exchange(route.Pool, route.FromCoin, route.ToCoin, currentAmount, minOutput)
        if err != nil {
            return nil, err
        }
        
        transactions = append(transactions, tx)
        
        // 更新下一步的输入金额
        if i < len(routes)-1 {
            currentAmount, err = s.client.GetDy(route.Pool, route.FromCoin, route.ToCoin, currentAmount)
            if err != nil {
                return nil, err
            }
        }
    }
    
    return transactions, nil
}

type ExchangeRoute struct {
    Pool     common.Address
    FromCoin int
    ToCoin   int
}

// 计算最小输出 (滑点保护)
func (s *ExchangeService) calculateMinOutput(expectedOutput *big.Int, slippagePercent float64) *big.Int {
    slippage := decimal.NewFromFloat(slippagePercent / 100)
    expected := decimal.NewFromBigInt(expectedOutput, 0)
    minOutput := expected.Mul(decimal.NewFromFloat(1).Sub(slippage))
    
    result, _ := minOutput.BigInt()
    return result
}
```

## 流动性提供

### 5.1 流动性服务

```go
// services/liquidity_service.go
package services

import (
    "context"
    "crypto/ecdsa"
    "math/big"
    
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/crypto"
)

type LiquidityService struct {
    client     *CurveClient
    privateKey *ecdsa.PrivateKey
}

func NewLiquidityService(client *CurveClient, privateKey *ecdsa.PrivateKey) *LiquidityService {
    return &LiquidityService{
        client:     client,
        privateKey: privateKey,
    }
}

// 添加流动性 (3Pool示例)
func (s *LiquidityService) AddLiquidity3Pool(
    daiAmount, usdcAmount, usdtAmount *big.Int,
    minMintAmount *big.Int,
) (*types.Transaction, error) {
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 代币地址
    daiAddress := common.HexToAddress("0x6B175474E89094C44Da98b954EedeAC495271d0F")
    usdcAddress := common.HexToAddress("0xA0b86a33E6417c8f4c8c8c8c8c8c8c8c8c8c8c8c")
    usdtAddress := common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7")
    
    // 授权代币
    if daiAmount.Cmp(big.NewInt(0)) > 0 {
        if err := s.approveToken(daiAddress, ThreePoolAddress, daiAmount); err != nil {
            return nil, err
        }
    }
    if usdcAmount.Cmp(big.NewInt(0)) > 0 {
        if err := s.approveToken(usdcAddress, ThreePoolAddress, usdcAmount); err != nil {
            return nil, err
        }
    }
    if usdtAmount.Cmp(big.NewInt(0)) > 0 {
        if err := s.approveToken(usdtAddress, ThreePoolAddress, usdtAmount); err != nil {
            return nil, err
        }
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
    
    // 构建add_liquidity交易数据
    amounts := [3]*big.Int{daiAmount, usdcAmount, usdtAmount}
    data, err := s.client.poolABI.Pack("add_liquidity", amounts, minMintAmount)
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, ThreePoolAddress, big.NewInt(0), 300000, gasPrice, data)
    
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

// 移除流动性
func (s *LiquidityService) RemoveLiquidity3Pool(
    lpTokenAmount *big.Int,
    minAmounts [3]*big.Int,
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
    
    // 构建remove_liquidity交易数据
    data, err := s.client.poolABI.Pack("remove_liquidity", lpTokenAmount, minAmounts)
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, ThreePoolAddress, big.NewInt(0), 250000, gasPrice, data)
    
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

// 单币移除流动性
func (s *LiquidityService) RemoveLiquidityOneCoin(
    poolAddress common.Address,
    lpTokenAmount *big.Int,
    coinIndex int,
    minAmount *big.Int,
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
    
    // 构建remove_liquidity_one_coin交易数据
    data, err := s.client.poolABI.Pack(
        "remove_liquidity_one_coin",
        lpTokenAmount,
        big.NewInt(int64(coinIndex)),
        minAmount,
    )
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, poolAddress, big.NewInt(0), 200000, gasPrice, data)
    
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

// 计算添加流动性获得的LP代币数量
func (s *LiquidityService) CalculateAddLiquidity(
    poolAddress common.Address,
    amounts []*big.Int,
) (*big.Int, error) {
    // 这需要调用池子的calc_token_amount函数
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    var result *big.Int
    err := s.client.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &poolAddress,
        Data: s.client.poolABI.Methods["calc_token_amount"].ID,
    }, nil)
    
    if err != nil {
        return nil, err
    }
    
    return result, nil
}
```

## 收益挖矿

### 6.1 Gauge质押服务

```go
// services/gauge_service.go
package services

import (
    "context"
    "crypto/ecdsa"
    "math/big"
    
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/crypto"
)

type GaugeService struct {
    client     *CurveClient
    privateKey *ecdsa.PrivateKey
}

func NewGaugeService(client *CurveClient, privateKey *ecdsa.PrivateKey) *GaugeService {
    return &GaugeService{
        client:     client,
        privateKey: privateKey,
    }
}

// 质押LP代币到Gauge
func (s *GaugeService) DepositToGauge(
    gaugeAddress common.Address,
    lpTokenAddress common.Address,
    amount *big.Int,
) (*types.Transaction, error) {
    // 授权Gauge使用LP代币
    if err := s.approveToken(lpTokenAddress, gaugeAddress, amount); err != nil {
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
    
    // 构建deposit交易数据
    data, err := s.client.gaugeABI.Pack("deposit", amount)
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, gaugeAddress, big.NewInt(0), 150000, gasPrice, data)
    
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

// 从Gauge提取LP代币
func (s *GaugeService) WithdrawFromGauge(
    gaugeAddress common.Address,
    amount *big.Int,
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
    data, err := s.client.gaugeABI.Pack("withdraw", amount)
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, gaugeAddress, big.NewInt(0), 120000, gasPrice, data)
    
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

// 领取CRV奖励
func (s *GaugeService) ClaimRewards(gaugeAddress common.Address) (*types.Transaction, error) {
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
    
    // 构建claim_rewards交易数据
    data, err := s.client.gaugeABI.Pack("claim_rewards")
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, gaugeAddress, big.NewInt(0), 100000, gasPrice, data)
    
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

// 获取可领取的CRV数量
func (s *GaugeService) GetClaimableTokens(gaugeAddress, userAddress common.Address) (*big.Int, error) {
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    var result *big.Int
    err := s.client.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &gaugeAddress,
        Data: s.client.gaugeABI.Methods["claimable_tokens"].ID,
    }, nil)
    
    if err != nil {
        return nil, err
    }
    
    return result, nil
}

// 获取用户在Gauge中的余额
func (s *GaugeService) GetGaugeBalance(gaugeAddress, userAddress common.Address) (*big.Int, error) {
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    var result *big.Int
    err := s.client.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &gaugeAddress,
        Data: s.client.gaugeABI.Methods["balanceOf"].ID,
    }, nil)
    
    if err != nil {
        return nil, err
    }
    
    return result, nil
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
    // 创建Curve客户端
    curveClient, err := client.NewCurveClient("https://mainnet.infura.io/v3/YOUR_PROJECT_ID")
    if err != nil {
        log.Fatal("创建Curve客户端失败:", err)
    }
    
    // 加载私钥
    privateKey, err := crypto.HexToECDSA("your_private_key_here")
    if err != nil {
        log.Fatal("加载私钥失败:", err)
    }
    
    // 创建服务
    exchangeService := services.NewExchangeService(curveClient, privateKey)
    liquidityService := services.NewLiquidityService(curveClient, privateKey)
    gaugeService := services.NewGaugeService(curveClient, privateKey)
    stableSwapService := services.NewStableSwapService(curveClient)
    
    // 代币地址
    usdcAddress := common.HexToAddress("0xA0b86a33E6417c8f4c8c8c8c8c8c8c8c8c8c8c8c")
    daiAddress := common.HexToAddress("0x6B175474E89094C44Da98b954EedeAC495271d0F")
    
    // 1. 查询USDC到DAI的交换率
    swapAmount := big.NewInt(1000 * 1e6) // 1000 USDC
    calculation, err := stableSwapService.CalculateStableSwap(
        ThreePoolAddress,
        USDC_INDEX,
        DAI_INDEX,
        swapAmount,
    )
    if err != nil {
        log.Fatal("计算交换失败:", err)
    }
    
    fmt.Printf("交换 1000 USDC 可获得 %s DAI\n", calculation.OutputAmount.String())
    fmt.Printf("价格影响: %s%%\n", calculation.PriceImpact.String())
    fmt.Printf("交易费用: %s\n", calculation.TradingFee.String())
    
    // 2. 执行USDC到DAI的交换
    minOutput := new(big.Int).Mul(calculation.OutputAmount, big.NewInt(99))
    minOutput.Div(minOutput, big.NewInt(100)) // 1% 滑点保护
    
    tx, err := exchangeService.Exchange(
        ThreePoolAddress,
        USDC_INDEX,
        DAI_INDEX,
        swapAmount,
        minOutput,
    )
    if err != nil {
        log.Fatal("交换失败:", err)
    }
    
    fmt.Printf("交换交易已提交: %s\n", tx.Hash().Hex())
    
    // 3. 添加流动性到3Pool
    daiAmount := big.NewInt(1000 * 1e18) // 1000 DAI
    usdcAmount := big.NewInt(1000 * 1e6) // 1000 USDC
    usdtAmount := big.NewInt(1000 * 1e6) // 1000 USDT
    
    // 计算最小LP代币数量
    minLPTokens, err := liquidityService.CalculateAddLiquidity(
        ThreePoolAddress,
        []*big.Int{daiAmount, usdcAmount, usdtAmount},
    )
    if err != nil {
        log.Fatal("计算LP代币失败:", err)
    }
    
    // 应用滑点保护
    minLPTokens = new(big.Int).Mul(minLPTokens, big.NewInt(99))
    minLPTokens.Div(minLPTokens, big.NewInt(100))
    
    tx, err = liquidityService.AddLiquidity3Pool(daiAmount, usdcAmount, usdtAmount, minLPTokens)
    if err != nil {
        log.Fatal("添加流动性失败:", err)
    }
    
    fmt.Printf("添加流动性交易已提交: %s\n", tx.Hash().Hex())
    
    // 4. 质押LP代币到Gauge获取CRV奖励
    threePoolGauge := common.HexToAddress("0xbFcF63294aD7105dEa65aA58F8AE5BE2D9d0952A")
    lpTokenAmount := big.NewInt(1000 * 1e18) // 假设获得了1000个LP代币
    
    tx, err = gaugeService.DepositToGauge(threePoolGauge, ThreePoolLPToken, lpTokenAmount)
    if err != nil {
        log.Fatal("质押到Gauge失败:", err)
    }
    
    fmt.Printf("Gauge质押交易已提交: %s\n", tx.Hash().Hex())
    
    // 5. 查询可领取的CRV奖励
    userAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
    claimableRewards, err := gaugeService.GetClaimableTokens(threePoolGauge, userAddress)
    if err != nil {
        log.Fatal("查询奖励失败:", err)
    }
    
    fmt.Printf("可领取的CRV奖励: %s\n", claimableRewards.String())
    
    // 6. 领取CRV奖励
    if claimableRewards.Cmp(big.NewInt(0)) > 0 {
        tx, err = gaugeService.ClaimRewards(threePoolGauge)
        if err != nil {
            log.Fatal("领取奖励失败:", err)
        }
        
        fmt.Printf("领取奖励交易已提交: %s\n", tx.Hash().Hex())
    }
}
```

这个Curve使用指南提供了完整的稳定币DEX集成方案，涵盖了交换、流动性提供、收益挖矿等核心功能，是DeFi稳定币交易的重要参考文档。
