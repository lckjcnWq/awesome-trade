# Uniswap Go SDK 详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [环境准备](#环境准备)
3. [V2协议交互](#V2协议交互)
4. [V3协议交互](#V3协议交互)
5. [流动性管理](#流动性管理)
6. [价格查询](#价格查询)
7. [交易执行](#交易执行)
8. [高级功能](#高级功能)
9. [实际应用](#实际应用)

## 基础概念

### 1.1 Uniswap简介

Uniswap是去中心化交易所(DEX)协议，支持自动化做市商(AMM)模式，提供代币交换和流动性挖矿功能。

```bash
# 安装Uniswap相关依赖
go get github.com/ethereum/go-ethereum
go get github.com/uniswap/uniswap-sdk-core
go get github.com/shopspring/decimal
```

### 1.2 协议版本

```go
// config/uniswap.go
package config

import (
    "math/big"
    
    "github.com/ethereum/go-ethereum/common"
)

type UniswapConfig struct {
    // V2合约地址
    V2RouterAddr   common.Address
    V2FactoryAddr  common.Address
    
    // V3合约地址
    V3RouterAddr   common.Address
    V3FactoryAddr  common.Address
    V3PoolAddr     common.Address
    
    // 网络配置
    ChainID        *big.Int
    WETH           common.Address
    
    // 交易配置
    SlippageTolerance decimal.Decimal
    DeadlineMinutes   int64
}

func DefaultUniswapConfig() *UniswapConfig {
    return &UniswapConfig{
        // 以太坊主网地址
        V2RouterAddr:  common.HexToAddress("0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D"),
        V2FactoryAddr: common.HexToAddress("0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f"),
        V3RouterAddr:  common.HexToAddress("0xE592427A0AEce92De3Edee1F18E0157C05861564"),
        V3FactoryAddr: common.HexToAddress("0x1F98431c8aD98523631AE4a59f267346ea31F984"),
        
        ChainID: big.NewInt(1), // 以太坊主网
        WETH:    common.HexToAddress("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"),
        
        SlippageTolerance: decimal.NewFromFloat(0.005), // 0.5%
        DeadlineMinutes:   20,
    }
}

// 常用代币地址
var (
    USDC_ADDR = common.HexToAddress("0xA0b86a33E6441b8435b662303c0f218C8c7c8e8e")
    USDT_ADDR = common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7")
    DAI_ADDR  = common.HexToAddress("0x6B175474E89094C44Da98b954EedeAC495271d0F")
    WBTC_ADDR = common.HexToAddress("0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599")
)
```

## 环境准备

### 2.1 客户端设置

```go
// client/uniswap_client.go
package client

import (
    "context"
    "fmt"
    "math/big"

    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
)

type UniswapClient struct {
    ethClient *ethclient.Client
    config    *UniswapConfig
    chainID   *big.Int
}

func NewUniswapClient(rpcURL string, config *UniswapConfig) (*UniswapClient, error) {
    client, err := ethclient.Dial(rpcURL)
    if err != nil {
        return nil, fmt.Errorf("连接以太坊节点失败: %v", err)
    }

    // 验证链ID
    chainID, err := client.NetworkID(context.Background())
    if err != nil {
        return nil, fmt.Errorf("获取网络ID失败: %v", err)
    }

    if chainID.Cmp(config.ChainID) != 0 {
        return nil, fmt.Errorf("链ID不匹配: 期望 %s, 实际 %s", 
            config.ChainID.String(), chainID.String())
    }

    return &UniswapClient{
        ethClient: client,
        config:    config,
        chainID:   chainID,
    }, nil
}

// 获取以太坊客户端
func (u *UniswapClient) GetEthClient() *ethclient.Client {
    return u.ethClient
}

// 获取配置
func (u *UniswapClient) GetConfig() *UniswapConfig {
    return u.config
}

// 获取当前区块号
func (u *UniswapClient) GetBlockNumber() (uint64, error) {
    header, err := u.ethClient.HeaderByNumber(context.Background(), nil)
    if err != nil {
        return 0, err
    }
    return header.Number.Uint64(), nil
}

// 获取Gas价格
func (u *UniswapClient) GetGasPrice() (*big.Int, error) {
    return u.ethClient.SuggestGasPrice(context.Background())
}

// 发送交易
func (u *UniswapClient) SendTransaction(signedTx *types.Transaction) error {
    return u.ethClient.SendTransaction(context.Background(), signedTx)
}

// 等待交易确认
func (u *UniswapClient) WaitForReceipt(txHash common.Hash) (*types.Receipt, error) {
    for {
        receipt, err := u.ethClient.TransactionReceipt(context.Background(), txHash)
        if err != nil {
            continue
        }
        return receipt, nil
    }
}
```

## V2协议交互

### 3.1 V2路由器

```go
// v2/router.go
package v2

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
    "github.com/shopspring/decimal"

    "your-project/client"
    "your-project/wallet"
)

// Uniswap V2 Router ABI（简化版）
const V2RouterABI = `[
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
    },
    {
        "inputs": [
            {"name": "amountOutMin", "type": "uint256"},
            {"name": "path", "type": "address[]"},
            {"name": "to", "type": "address"},
            {"name": "deadline", "type": "uint256"}
        ],
        "name": "swapExactETHForTokens",
        "outputs": [
            {"name": "amounts", "type": "uint256[]"}
        ],
        "stateMutability": "payable",
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
        "name": "swapExactTokensForETH",
        "outputs": [
            {"name": "amounts", "type": "uint256[]"}
        ],
        "stateMutability": "nonpayable",
        "type": "function"
    }
]`

type V2Router struct {
    client     *client.UniswapClient
    wallet     *wallet.WalletManager
    contract   *bind.BoundContract
    routerAddr common.Address
}

func NewV2Router(client *client.UniswapClient, wallet *wallet.WalletManager) (*V2Router, error) {
    routerAddr := client.GetConfig().V2RouterAddr
    
    parsedABI, err := abi.JSON(strings.NewReader(V2RouterABI))
    if err != nil {
        return nil, fmt.Errorf("解析V2 Router ABI失败: %v", err)
    }

    contract := bind.NewBoundContract(
        routerAddr, 
        parsedABI, 
        client.GetEthClient(), 
        client.GetEthClient(), 
        client.GetEthClient(),
    )

    return &V2Router{
        client:     client,
        wallet:     wallet,
        contract:   contract,
        routerAddr: routerAddr,
    }, nil
}

// 获取交换输出金额
func (r *V2Router) GetAmountsOut(amountIn decimal.Decimal, path []common.Address) ([]decimal.Decimal, error) {
    amountInBig := amountIn.Shift(18).BigInt() // 假设18位小数

    var result []interface{}
    err := r.contract.Call(nil, &result, "getAmountsOut", amountInBig, path)
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

// 代币到代币交换
func (r *V2Router) SwapExactTokensForTokens(
    amountIn decimal.Decimal,
    amountOutMin decimal.Decimal,
    path []common.Address,
) (*types.Transaction, error) {
    
    amountInBig := amountIn.Shift(18).BigInt()
    amountOutMinBig := amountOutMin.Shift(18).BigInt()
    
    // 计算截止时间
    deadline := big.NewInt(time.Now().Unix() + r.client.GetConfig().DeadlineMinutes*60)

    nonce, err := r.client.GetEthClient().PendingNonceAt(
        context.Background(), 
        r.wallet.GetAddress(),
    )
    if err != nil {
        return nil, err
    }

    gasPrice, err := r.client.GetGasPrice()
    if err != nil {
        return nil, err
    }

    auth := &bind.TransactOpts{
        From:     r.wallet.GetAddress(),
        Nonce:    big.NewInt(int64(nonce)),
        GasLimit: 300000,
        GasPrice: gasPrice,
        Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
            return r.wallet.SignTransaction(tx, r.client.chainID)
        },
    }

    return r.contract.Transact(
        auth,
        "swapExactTokensForTokens",
        amountInBig,
        amountOutMinBig,
        path,
        r.wallet.GetAddress(),
        deadline,
    )
}

// ETH到代币交换
func (r *V2Router) SwapExactETHForTokens(
    amountIn decimal.Decimal,
    amountOutMin decimal.Decimal,
    path []common.Address,
) (*types.Transaction, error) {
    
    amountInBig := amountIn.Shift(18).BigInt()
    amountOutMinBig := amountOutMin.Shift(18).BigInt()
    deadline := big.NewInt(time.Now().Unix() + r.client.GetConfig().DeadlineMinutes*60)

    nonce, err := r.client.GetEthClient().PendingNonceAt(
        context.Background(), 
        r.wallet.GetAddress(),
    )
    if err != nil {
        return nil, err
    }

    gasPrice, err := r.client.GetGasPrice()
    if err != nil {
        return nil, err
    }

    auth := &bind.TransactOpts{
        From:     r.wallet.GetAddress(),
        Nonce:    big.NewInt(int64(nonce)),
        GasLimit: 300000,
        GasPrice: gasPrice,
        Value:    amountInBig, // ETH金额
        Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
            return r.wallet.SignTransaction(tx, r.client.chainID)
        },
    }

    return r.contract.Transact(
        auth,
        "swapExactETHForTokens",
        amountOutMinBig,
        path,
        r.wallet.GetAddress(),
        deadline,
    )
}

// 代币到ETH交换
func (r *V2Router) SwapExactTokensForETH(
    amountIn decimal.Decimal,
    amountOutMin decimal.Decimal,
    path []common.Address,
) (*types.Transaction, error) {
    
    amountInBig := amountIn.Shift(18).BigInt()
    amountOutMinBig := amountOutMin.Shift(18).BigInt()
    deadline := big.NewInt(time.Now().Unix() + r.client.GetConfig().DeadlineMinutes*60)

    nonce, err := r.client.GetEthClient().PendingNonceAt(
        context.Background(), 
        r.wallet.GetAddress(),
    )
    if err != nil {
        return nil, err
    }

    gasPrice, err := r.client.GetGasPrice()
    if err != nil {
        return nil, err
    }

    auth := &bind.TransactOpts{
        From:     r.wallet.GetAddress(),
        Nonce:    big.NewInt(int64(nonce)),
        GasLimit: 300000,
        GasPrice: gasPrice,
        Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
            return r.wallet.SignTransaction(tx, r.client.chainID)
        },
    }

    return r.contract.Transact(
        auth,
        "swapExactTokensForETH",
        amountInBig,
        amountOutMinBig,
        path,
        r.wallet.GetAddress(),
        deadline,
    )
}

// 计算最优路径
func (r *V2Router) FindBestPath(tokenA, tokenB common.Address, amountIn decimal.Decimal) ([]common.Address, decimal.Decimal, error) {
    weth := r.client.GetConfig().WETH
    
    // 直接路径
    directPath := []common.Address{tokenA, tokenB}
    directAmounts, err := r.GetAmountsOut(amountIn, directPath)
    if err != nil {
        return nil, decimal.Zero, err
    }

    // 通过WETH的路径
    wethPath := []common.Address{tokenA, weth, tokenB}
    wethAmounts, err := r.GetAmountsOut(amountIn, wethPath)
    if err != nil {
        return directPath, directAmounts[len(directAmounts)-1], nil
    }

    // 比较输出金额
    if wethAmounts[len(wethAmounts)-1].GreaterThan(directAmounts[len(directAmounts)-1]) {
        return wethPath, wethAmounts[len(wethAmounts)-1], nil
    }

    return directPath, directAmounts[len(directAmounts)-1], nil
}

// 计算滑点保护的最小输出
func (r *V2Router) CalculateMinAmountOut(expectedAmount decimal.Decimal) decimal.Decimal {
    slippage := r.client.GetConfig().SlippageTolerance
    return expectedAmount.Mul(decimal.NewFromInt(1).Sub(slippage))
}
```

## V3协议交互

### 4.1 V3路由器

```go
// v3/router.go
package v3

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
    "github.com/shopspring/decimal"

    "your-project/client"
    "your-project/wallet"
)

// Uniswap V3 Router ABI（简化版）
const V3RouterABI = `[
    {
        "inputs": [
            {
                "components": [
                    {"name": "tokenIn", "type": "address"},
                    {"name": "tokenOut", "type": "address"},
                    {"name": "fee", "type": "uint24"},
                    {"name": "recipient", "type": "address"},
                    {"name": "deadline", "type": "uint256"},
                    {"name": "amountIn", "type": "uint256"},
                    {"name": "amountOutMinimum", "type": "uint256"},
                    {"name": "sqrtPriceLimitX96", "type": "uint160"}
                ],
                "name": "params",
                "type": "tuple"
            }
        ],
        "name": "exactInputSingle",
        "outputs": [
            {"name": "amountOut", "type": "uint256"}
        ],
        "stateMutability": "payable",
        "type": "function"
    }
]`

type V3Router struct {
    client     *client.UniswapClient
    wallet     *wallet.WalletManager
    contract   *bind.BoundContract
    routerAddr common.Address
}

func NewV3Router(client *client.UniswapClient, wallet *wallet.WalletManager) (*V3Router, error) {
    routerAddr := client.GetConfig().V3RouterAddr
    
    parsedABI, err := abi.JSON(strings.NewReader(V3RouterABI))
    if err != nil {
        return nil, fmt.Errorf("解析V3 Router ABI失败: %v", err)
    }

    contract := bind.NewBoundContract(
        routerAddr, 
        parsedABI, 
        client.GetEthClient(), 
        client.GetEthClient(), 
        client.GetEthClient(),
    )

    return &V3Router{
        client:     client,
        wallet:     wallet,
        contract:   contract,
        routerAddr: routerAddr,
    }, nil
}

// V3单跳交换
func (r *V3Router) ExactInputSingle(params ExactInputSingleParams) (*types.Transaction, error) {
    deadline := big.NewInt(time.Now().Unix() + r.client.GetConfig().DeadlineMinutes*60)
    
    // 构建参数结构
    swapParams := struct {
        TokenIn           common.Address
        TokenOut          common.Address
        Fee               *big.Int
        Recipient         common.Address
        Deadline          *big.Int
        AmountIn          *big.Int
        AmountOutMinimum  *big.Int
        SqrtPriceLimitX96 *big.Int
    }{
        TokenIn:           params.TokenIn,
        TokenOut:          params.TokenOut,
        Fee:               big.NewInt(int64(params.Fee)),
        Recipient:         r.wallet.GetAddress(),
        Deadline:          deadline,
        AmountIn:          params.AmountIn.Shift(18).BigInt(),
        AmountOutMinimum:  params.AmountOutMinimum.Shift(18).BigInt(),
        SqrtPriceLimitX96: big.NewInt(0), // 0表示无价格限制
    }

    nonce, err := r.client.GetEthClient().PendingNonceAt(
        context.Background(), 
        r.wallet.GetAddress(),
    )
    if err != nil {
        return nil, err
    }

    gasPrice, err := r.client.GetGasPrice()
    if err != nil {
        return nil, err
    }

    auth := &bind.TransactOpts{
        From:     r.wallet.GetAddress(),
        Nonce:    big.NewInt(int64(nonce)),
        GasLimit: 400000, // V3需要更多gas
        GasPrice: gasPrice,
        Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
            return r.wallet.SignTransaction(tx, r.client.chainID)
        },
    }

    return r.contract.Transact(auth, "exactInputSingle", swapParams)
}

// 获取池子费率
func (r *V3Router) GetPoolFees() []uint32 {
    return []uint32{
        500,   // 0.05%
        3000,  // 0.3%
        10000, // 1%
    }
}

// 寻找最佳费率池
func (r *V3Router) FindBestFeePool(tokenA, tokenB common.Address, amountIn decimal.Decimal) (uint32, decimal.Decimal, error) {
    fees := r.GetPoolFees()
    bestFee := fees[0]
    bestOutput := decimal.Zero

    for _, fee := range fees {
        // 这里需要调用quoter合约来获取报价
        // 简化实现，返回模拟数据
        output := amountIn.Mul(decimal.NewFromFloat(0.99)) // 模拟1%滑点
        
        if output.GreaterThan(bestOutput) {
            bestOutput = output
            bestFee = fee
        }
    }

    return bestFee, bestOutput, nil
}

type ExactInputSingleParams struct {
    TokenIn          common.Address
    TokenOut         common.Address
    Fee              uint32
    AmountIn         decimal.Decimal
    AmountOutMinimum decimal.Decimal
}
```

## 流动性管理

### 5.1 流动性提供

```go
// liquidity/provider.go
package liquidity

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

// V2 Pair ABI（简化版）
const V2PairABI = `[
    {
        "inputs": [],
        "name": "getReserves",
        "outputs": [
            {"name": "_reserve0", "type": "uint112"},
            {"name": "_reserve1", "type": "uint112"},
            {"name": "_blockTimestampLast", "type": "uint32"}
        ],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [],
        "name": "totalSupply",
        "outputs": [
            {"name": "", "type": "uint256"}
        ],
        "stateMutability": "view",
        "type": "function"
    }
]`

type LiquidityProvider struct {
    client *client.UniswapClient
    wallet *wallet.WalletManager
}

func NewLiquidityProvider(client *client.UniswapClient, wallet *wallet.WalletManager) *LiquidityProvider {
    return &LiquidityProvider{
        client: client,
        wallet: wallet,
    }
}

// 获取V2池子储备
func (lp *LiquidityProvider) GetV2PoolReserves(pairAddr common.Address) (*PoolReserves, error) {
    parsedABI, err := abi.JSON(strings.NewReader(V2PairABI))
    if err != nil {
        return nil, err
    }

    contract := bind.NewBoundContract(
        pairAddr, 
        parsedABI, 
        lp.client.GetEthClient(), 
        lp.client.GetEthClient(), 
        lp.client.GetEthClient(),
    )

    var result []interface{}
    err = contract.Call(nil, &result, "getReserves")
    if err != nil {
        return nil, err
    }

    reserve0 := result[0].(*big.Int)
    reserve1 := result[1].(*big.Int)
    timestamp := result[2].(uint32)

    // 获取总供应量
    var totalSupplyResult []interface{}
    err = contract.Call(nil, &totalSupplyResult, "totalSupply")
    if err != nil {
        return nil, err
    }

    totalSupply := totalSupplyResult[0].(*big.Int)

    return &PoolReserves{
        Reserve0:    decimal.NewFromBigInt(reserve0, -18),
        Reserve1:    decimal.NewFromBigInt(reserve1, -18),
        TotalSupply: decimal.NewFromBigInt(totalSupply, -18),
        Timestamp:   timestamp,
    }, nil
}

// 计算添加流动性所需的代币数量
func (lp *LiquidityProvider) CalculateAddLiquidity(
    pairAddr common.Address,
    tokenAAmount decimal.Decimal,
    tokenBAmount decimal.Decimal,
) (*LiquidityCalculation, error) {
    
    reserves, err := lp.GetV2PoolReserves(pairAddr)
    if err != nil {
        return nil, err
    }

    // 计算最优比例
    if reserves.Reserve0.IsZero() || reserves.Reserve1.IsZero() {
        // 新池子，使用提供的比例
        return &LiquidityCalculation{
            TokenAAmount: tokenAAmount,
            TokenBAmount: tokenBAmount,
            LPTokens:     tokenAAmount.Add(tokenBAmount), // 简化计算
        }, nil
    }

    // 现有池子，计算最优比例
    ratio := reserves.Reserve1.Div(reserves.Reserve0)
    
    // 基于tokenA计算tokenB
    calculatedTokenB := tokenAAmount.Mul(ratio)
    
    // 基于tokenB计算tokenA
    calculatedTokenA := tokenBAmount.Div(ratio)

    var finalTokenA, finalTokenB decimal.Decimal
    
    if calculatedTokenB.LessThanOrEqual(tokenBAmount) {
        finalTokenA = tokenAAmount
        finalTokenB = calculatedTokenB
    } else {
        finalTokenA = calculatedTokenA
        finalTokenB = tokenBAmount
    }

    // 计算LP代币数量
    lpTokens := finalTokenA.Mul(reserves.TotalSupply).Div(reserves.Reserve0)

    return &LiquidityCalculation{
        TokenAAmount: finalTokenA,
        TokenBAmount: finalTokenB,
        LPTokens:     lpTokens,
    }, nil
}

// 计算移除流动性获得的代币数量
func (lp *LiquidityProvider) CalculateRemoveLiquidity(
    pairAddr common.Address,
    lpTokenAmount decimal.Decimal,
) (*RemoveLiquidityCalculation, error) {
    
    reserves, err := lp.GetV2PoolReserves(pairAddr)
    if err != nil {
        return nil, err
    }

    // 计算份额
    share := lpTokenAmount.Div(reserves.TotalSupply)
    
    tokenAAmount := reserves.Reserve0.Mul(share)
    tokenBAmount := reserves.Reserve1.Mul(share)

    return &RemoveLiquidityCalculation{
        TokenAAmount: tokenAAmount,
        TokenBAmount: tokenBAmount,
        Share:        share,
    }, nil
}

// 计算无常损失
func (lp *LiquidityProvider) CalculateImpermanentLoss(
    initialPriceRatio decimal.Decimal,
    currentPriceRatio decimal.Decimal,
) decimal.Decimal {
    
    // 无常损失公式: 2*sqrt(ratio) / (1 + ratio) - 1
    ratio := currentPriceRatio.Div(initialPriceRatio)
    
    sqrt := ratio.Pow(decimal.NewFromFloat(0.5))
    numerator := decimal.NewFromInt(2).Mul(sqrt)
    denominator := decimal.NewFromInt(1).Add(ratio)
    
    result := numerator.Div(denominator).Sub(decimal.NewFromInt(1))
    
    return result
}

type PoolReserves struct {
    Reserve0    decimal.Decimal
    Reserve1    decimal.Decimal
    TotalSupply decimal.Decimal
    Timestamp   uint32
}

type LiquidityCalculation struct {
    TokenAAmount decimal.Decimal
    TokenBAmount decimal.Decimal
    LPTokens     decimal.Decimal
}

type RemoveLiquidityCalculation struct {
    TokenAAmount decimal.Decimal
    TokenBAmount decimal.Decimal
    Share        decimal.Decimal
}
```

## 价格查询

### 6.1 价格预言机

```go
// oracle/price.go
package oracle

import (
    "fmt"
    "math/big"

    "github.com/ethereum/go-ethereum/common"
    "github.com/shopspring/decimal"

    "your-project/client"
    "your-project/liquidity"
)

type PriceOracle struct {
    client            *client.UniswapClient
    liquidityProvider *liquidity.LiquidityProvider
}

func NewPriceOracle(client *client.UniswapClient) *PriceOracle {
    return &PriceOracle{
        client:            client,
        liquidityProvider: liquidity.NewLiquidityProvider(client, nil),
    }
}

// 获取代币价格（相对于另一个代币）
func (po *PriceOracle) GetTokenPrice(
    tokenA, tokenB common.Address,
    pairAddr common.Address,
) (decimal.Decimal, error) {
    
    reserves, err := po.liquidityProvider.GetV2PoolReserves(pairAddr)
    if err != nil {
        return decimal.Zero, err
    }

    if reserves.Reserve0.IsZero() {
        return decimal.Zero, fmt.Errorf("池子储备为零")
    }

    // 价格 = reserve1 / reserve0
    price := reserves.Reserve1.Div(reserves.Reserve0)
    return price, nil
}

// 获取代币美元价格（通过USDC池）
func (po *PriceOracle) GetTokenUSDPrice(
    token common.Address,
    usdcPairAddr common.Address,
) (decimal.Decimal, error) {
    
    // 假设USDC是reserve1
    price, err := po.GetTokenPrice(token, common.Address{}, usdcPairAddr)
    if err != nil {
        return decimal.Zero, err
    }

    return price, nil
}

// 计算时间加权平均价格(TWAP)
func (po *PriceOracle) CalculateTWAP(
    pairAddr common.Address,
    timeWindow int64,
) (decimal.Decimal, error) {
    
    // 这里需要实现TWAP计算逻辑
    // 需要获取历史价格数据
    // 简化实现，返回当前价格
    reserves, err := po.liquidityProvider.GetV2PoolReserves(pairAddr)
    if err != nil {
        return decimal.Zero, err
    }

    if reserves.Reserve0.IsZero() {
        return decimal.Zero, fmt.Errorf("池子储备为零")
    }

    return reserves.Reserve1.Div(reserves.Reserve0), nil
}

// 获取多个代币的价格
func (po *PriceOracle) GetMultipleTokenPrices(
    tokens []common.Address,
    basePairs []common.Address,
) (map[common.Address]decimal.Decimal, error) {
    
    prices := make(map[common.Address]decimal.Decimal)
    
    for i, token := range tokens {
        if i < len(basePairs) {
            price, err := po.GetTokenPrice(token, common.Address{}, basePairs[i])
            if err != nil {
                continue
            }
            prices[token] = price
        }
    }

    return prices, nil
}

// 价格影响计算
func (po *PriceOracle) CalculatePriceImpact(
    pairAddr common.Address,
    amountIn decimal.Decimal,
    isTokenAIn bool,
) (decimal.Decimal, error) {
    
    reserves, err := po.liquidityProvider.GetV2PoolReserves(pairAddr)
    if err != nil {
        return decimal.Zero, err
    }

    var reserveIn, reserveOut decimal.Decimal
    if isTokenAIn {
        reserveIn = reserves.Reserve0
        reserveOut = reserves.Reserve1
    } else {
        reserveIn = reserves.Reserve1
        reserveOut = reserves.Reserve0
    }

    // 计算交换后的输出
    amountInWithFee := amountIn.Mul(decimal.NewFromFloat(0.997)) // 0.3%手续费
    numerator := amountInWithFee.Mul(reserveOut)
    denominator := reserveIn.Add(amountInWithFee)
    amountOut := numerator.Div(denominator)

    // 计算价格影响
    originalPrice := reserveOut.Div(reserveIn)
    newReserveIn := reserveIn.Add(amountIn)
    newReserveOut := reserveOut.Sub(amountOut)
    newPrice := newReserveOut.Div(newReserveIn)

    priceImpact := originalPrice.Sub(newPrice).Div(originalPrice).Abs()
    
    return priceImpact, nil
}
```

## 交易执行

### 7.1 交易管理器

```go
// trading/manager.go
package trading

import (
    "fmt"
    "math/big"
    "time"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/shopspring/decimal"

    "your-project/client"
    "your-project/wallet"
    "your-project/v2"
    "your-project/v3"
    "your-project/oracle"
)

type TradingManager struct {
    client      *client.UniswapClient
    wallet      *wallet.WalletManager
    v2Router    *v2.V2Router
    v3Router    *v3.V3Router
    priceOracle *oracle.PriceOracle
}

func NewTradingManager(client *client.UniswapClient, wallet *wallet.WalletManager) (*TradingManager, error) {
    v2Router, err := v2.NewV2Router(client, wallet)
    if err != nil {
        return nil, err
    }

    v3Router, err := v3.NewV3Router(client, wallet)
    if err != nil {
        return nil, err
    }

    priceOracle := oracle.NewPriceOracle(client)

    return &TradingManager{
        client:      client,
        wallet:      wallet,
        v2Router:    v2Router,
        v3Router:    v3Router,
        priceOracle: priceOracle,
    }, nil
}

// 执行最优交换
func (tm *TradingManager) ExecuteBestSwap(params SwapParams) (*SwapResult, error) {
    // 比较V2和V3的价格
    v2Output, err := tm.getV2Output(params)
    if err != nil {
        return nil, fmt.Errorf("获取V2输出失败: %v", err)
    }

    v3Output, err := tm.getV3Output(params)
    if err != nil {
        return nil, fmt.Errorf("获取V3输出失败: %v", err)
    }

    // 选择最优路径
    if v3Output.GreaterThan(v2Output) {
        return tm.executeV3Swap(params)
    } else {
        return tm.executeV2Swap(params)
    }
}

// 获取V2输出
func (tm *TradingManager) getV2Output(params SwapParams) (decimal.Decimal, error) {
    path, output, err := tm.v2Router.FindBestPath(
        params.TokenIn,
        params.TokenOut,
        params.AmountIn,
    )
    if err != nil {
        return decimal.Zero, err
    }

    _ = path // 使用最优路径
    return output, nil
}

// 获取V3输出
func (tm *TradingManager) getV3Output(params SwapParams) (decimal.Decimal, error) {
    fee, output, err := tm.v3Router.FindBestFeePool(
        params.TokenIn,
        params.TokenOut,
        params.AmountIn,
    )
    if err != nil {
        return decimal.Zero, err
    }

    _ = fee // 使用最优费率
    return output, nil
}

// 执行V2交换
func (tm *TradingManager) executeV2Swap(params SwapParams) (*SwapResult, error) {
    path, expectedOutput, err := tm.v2Router.FindBestPath(
        params.TokenIn,
        params.TokenOut,
        params.AmountIn,
    )
    if err != nil {
        return nil, err
    }

    minOutput := tm.v2Router.CalculateMinAmountOut(expectedOutput)

    var tx *types.Transaction
    
    if params.TokenIn == tm.client.GetConfig().WETH {
        // ETH到代币
        tx, err = tm.v2Router.SwapExactETHForTokens(
            params.AmountIn,
            minOutput,
            path,
        )
    } else if params.TokenOut == tm.client.GetConfig().WETH {
        // 代币到ETH
        tx, err = tm.v2Router.SwapExactTokensForETH(
            params.AmountIn,
            minOutput,
            path,
        )
    } else {
        // 代币到代币
        tx, err = tm.v2Router.SwapExactTokensForTokens(
            params.AmountIn,
            minOutput,
            path,
        )
    }

    if err != nil {
        return nil, err
    }

    return &SwapResult{
        TxHash:         tx.Hash(),
        Protocol:       "V2",
        ExpectedOutput: expectedOutput,
        MinOutput:      minOutput,
        Path:           path,
        Timestamp:      time.Now(),
    }, nil
}

// 执行V3交换
func (tm *TradingManager) executeV3Swap(params SwapParams) (*SwapResult, error) {
    fee, expectedOutput, err := tm.v3Router.FindBestFeePool(
        params.TokenIn,
        params.TokenOut,
        params.AmountIn,
    )
    if err != nil {
        return nil, err
    }

    minOutput := expectedOutput.Mul(
        decimal.NewFromInt(1).Sub(tm.client.GetConfig().SlippageTolerance),
    )

    v3Params := v3.ExactInputSingleParams{
        TokenIn:          params.TokenIn,
        TokenOut:         params.TokenOut,
        Fee:              fee,
        AmountIn:         params.AmountIn,
        AmountOutMinimum: minOutput,
    }

    tx, err := tm.v3Router.ExactInputSingle(v3Params)
    if err != nil {
        return nil, err
    }

    return &SwapResult{
        TxHash:         tx.Hash(),
        Protocol:       "V3",
        ExpectedOutput: expectedOutput,
        MinOutput:      minOutput,
        Path:           []common.Address{params.TokenIn, params.TokenOut},
        Fee:            fee,
        Timestamp:      time.Now(),
    }, nil
}

// 监控交易状态
func (tm *TradingManager) MonitorTransaction(txHash common.Hash) (*TransactionStatus, error) {
    receipt, err := tm.client.WaitForReceipt(txHash)
    if err != nil {
        return &TransactionStatus{
            TxHash: txHash,
            Status: "pending",
        }, nil
    }

    status := "failed"
    if receipt.Status == 1 {
        status = "success"
    }

    return &TransactionStatus{
        TxHash:      txHash,
        Status:      status,
        BlockNumber: receipt.BlockNumber.Uint64(),
        GasUsed:     receipt.GasUsed,
        Confirmed:   true,
    }, nil
}

type SwapParams struct {
    TokenIn   common.Address
    TokenOut  common.Address
    AmountIn  decimal.Decimal
    Recipient common.Address
}

type SwapResult struct {
    TxHash         common.Hash
    Protocol       string
    ExpectedOutput decimal.Decimal
    MinOutput      decimal.Decimal
    Path           []common.Address
    Fee            uint32
    Timestamp      time.Time
}

type TransactionStatus struct {
    TxHash      common.Hash
    Status      string
    BlockNumber uint64
    GasUsed     uint64
    Confirmed   bool
}
```

## 高级功能

### 8.1 套利机器人

```go
// arbitrage/bot.go
package arbitrage

import (
    "fmt"
    "log"
    "time"

    "github.com/ethereum/go-ethereum/common"
    "github.com/shopspring/decimal"

    "your-project/trading"
    "your-project/oracle"
)

type ArbitrageBot struct {
    tradingManager *trading.TradingManager
    priceOracle    *oracle.PriceOracle
    minProfitRate  decimal.Decimal
    maxGasPrice    decimal.Decimal
}

func NewArbitrageBot(
    tradingManager *trading.TradingManager,
    priceOracle *oracle.PriceOracle,
    minProfitRate decimal.Decimal,
) *ArbitrageBot {
    return &ArbitrageBot{
        tradingManager: tradingManager,
        priceOracle:    priceOracle,
        minProfitRate:  minProfitRate,
        maxGasPrice:    decimal.NewFromFloat(100), // 100 Gwei
    }
}

// 寻找套利机会
func (ab *ArbitrageBot) FindArbitrageOpportunities(
    tokenPairs []TokenPair,
) ([]ArbitrageOpportunity, error) {
    
    var opportunities []ArbitrageOpportunity

    for _, pair := range tokenPairs {
        // 获取V2价格
        v2Price, err := ab.priceOracle.GetTokenPrice(
            pair.TokenA,
            pair.TokenB,
            pair.V2PairAddr,
        )
        if err != nil {
            continue
        }

        // 获取V3价格（简化实现）
        v3Price := v2Price.Mul(decimal.NewFromFloat(1.01)) // 模拟1%差价

        // 计算价格差异
        priceDiff := v3Price.Sub(v2Price).Div(v2Price).Abs()

        if priceDiff.GreaterThan(ab.minProfitRate) {
            opportunity := ArbitrageOpportunity{
                TokenA:      pair.TokenA,
                TokenB:      pair.TokenB,
                V2Price:     v2Price,
                V3Price:     v3Price,
                PriceDiff:   priceDiff,
                Timestamp:   time.Now(),
            }

            // 计算最优交易金额
            optimalAmount, profit := ab.calculateOptimalAmount(opportunity)
            opportunity.OptimalAmount = optimalAmount
            opportunity.ExpectedProfit = profit

            opportunities = append(opportunities, opportunity)
        }
    }

    return opportunities, nil
}

// 计算最优交易金额
func (ab *ArbitrageBot) calculateOptimalAmount(
    opportunity ArbitrageOpportunity,
) (decimal.Decimal, decimal.Decimal) {
    
    // 简化计算，实际需要考虑滑点、gas费等
    baseAmount := decimal.NewFromFloat(1000) // 1000 USD等值
    profit := baseAmount.Mul(opportunity.PriceDiff)
    
    return baseAmount, profit
}

// 执行套利交易
func (ab *ArbitrageBot) ExecuteArbitrage(
    opportunity ArbitrageOpportunity,
) (*ArbitrageResult, error) {
    
    // 检查gas价格
    gasPrice, err := ab.tradingManager.client.GetGasPrice()
    if err != nil {
        return nil, err
    }

    gasPriceGwei := decimal.NewFromBigInt(gasPrice, -9)
    if gasPriceGwei.GreaterThan(ab.maxGasPrice) {
        return nil, fmt.Errorf("gas价格过高: %s Gwei", gasPriceGwei.String())
    }

    // 执行第一笔交易（买入）
    buyParams := trading.SwapParams{
        TokenIn:  opportunity.TokenA,
        TokenOut: opportunity.TokenB,
        AmountIn: opportunity.OptimalAmount,
    }

    buyResult, err := ab.tradingManager.ExecuteBestSwap(buyParams)
    if err != nil {
        return nil, fmt.Errorf("买入交易失败: %v", err)
    }

    // 等待确认
    buyStatus, err := ab.tradingManager.MonitorTransaction(buyResult.TxHash)
    if err != nil || buyStatus.Status != "success" {
        return nil, fmt.Errorf("买入交易未确认")
    }

    // 执行第二笔交易（卖出）
    sellParams := trading.SwapParams{
        TokenIn:  opportunity.TokenB,
        TokenOut: opportunity.TokenA,
        AmountIn: buyResult.ExpectedOutput,
    }

    sellResult, err := ab.tradingManager.ExecuteBestSwap(sellParams)
    if err != nil {
        return nil, fmt.Errorf("卖出交易失败: %v", err)
    }

    // 等待确认
    sellStatus, err := ab.tradingManager.MonitorTransaction(sellResult.TxHash)
    if err != nil || sellStatus.Status != "success" {
        return nil, fmt.Errorf("卖出交易未确认")
    }

    // 计算实际利润
    actualProfit := sellResult.ExpectedOutput.Sub(opportunity.OptimalAmount)

    return &ArbitrageResult{
        Opportunity:   opportunity,
        BuyTxHash:     buyResult.TxHash,
        SellTxHash:    sellResult.TxHash,
        ActualProfit:  actualProfit,
        ExecutionTime: time.Now(),
        Success:       true,
    }, nil
}

// 启动套利机器人
func (ab *ArbitrageBot) Start(tokenPairs []TokenPair, interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            opportunities, err := ab.FindArbitrageOpportunities(tokenPairs)
            if err != nil {
                log.Printf("寻找套利机会失败: %v", err)
                continue
            }

            for _, opportunity := range opportunities {
                log.Printf("发现套利机会: %s/%s, 价格差异: %s%%",
                    opportunity.TokenA.Hex()[:8],
                    opportunity.TokenB.Hex()[:8],
                    opportunity.PriceDiff.Mul(decimal.NewFromInt(100)).String(),
                )

                // 执行套利
                result, err := ab.ExecuteArbitrage(opportunity)
                if err != nil {
                    log.Printf("套利执行失败: %v", err)
                    continue
                }

                log.Printf("套利成功: 利润 %s", result.ActualProfit.String())
            }
        }
    }
}

type TokenPair struct {
    TokenA     common.Address
    TokenB     common.Address
    V2PairAddr common.Address
    V3PoolAddr common.Address
}

type ArbitrageOpportunity struct {
    TokenA         common.Address
    TokenB         common.Address
    V2Price        decimal.Decimal
    V3Price        decimal.Decimal
    PriceDiff      decimal.Decimal
    OptimalAmount  decimal.Decimal
    ExpectedProfit decimal.Decimal
    Timestamp      time.Time
}

type ArbitrageResult struct {
    Opportunity   ArbitrageOpportunity
    BuyTxHash     common.Hash
    SellTxHash    common.Hash
    ActualProfit  decimal.Decimal
    ExecutionTime time.Time
    Success       bool
}
```

## 实际应用

### 9.1 完整Uniswap应用

```go
// main.go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/ethereum/go-ethereum/common"
    "github.com/shopspring/decimal"

    "your-project/config"
    "your-project/client"
    "your-project/wallet"
    "your-project/trading"
    "your-project/oracle"
    "your-project/arbitrage"
)

func main() {
    // 创建Uniswap配置
    cfg := config.DefaultUniswapConfig()

    // 创建客户端
    uniswapClient, err := client.NewUniswapClient(
        "https://mainnet.infura.io/v3/YOUR_PROJECT_ID",
        cfg,
    )
    if err != nil {
        log.Fatal("创建Uniswap客户端失败:", err)
    }

    // 创建钱包
    walletManager, err := wallet.NewWalletManager("your_private_key_here")
    if err != nil {
        log.Fatal("创建钱包失败:", err)
    }

    fmt.Printf("钱包地址: %s\n", walletManager.GetAddress().Hex())

    // 创建交易管理器
    tradingManager, err := trading.NewTradingManager(uniswapClient, walletManager)
    if err != nil {
        log.Fatal("创建交易管理器失败:", err)
    }

    // 创建价格预言机
    priceOracle := oracle.NewPriceOracle(uniswapClient)

    // 代币交换示例
    swapParams := trading.SwapParams{
        TokenIn:  config.USDC_ADDR,
        TokenOut: cfg.WETH,
        AmountIn: decimal.NewFromFloat(100), // 100 USDC
    }

    fmt.Println("执行代币交换...")
    swapResult, err := tradingManager.ExecuteBestSwap(swapParams)
    if err != nil {
        log.Printf("交换失败: %v", err)
    } else {
        fmt.Printf("交换结果:\n")
        fmt.Printf("  协议: %s\n", swapResult.Protocol)
        fmt.Printf("  交易哈希: %s\n", swapResult.TxHash.Hex())
        fmt.Printf("  预期输出: %s ETH\n", swapResult.ExpectedOutput.String())
        fmt.Printf("  最小输出: %s ETH\n", swapResult.MinOutput.String())

        // 监控交易状态
        fmt.Println("监控交易状态...")
        status, err := tradingManager.MonitorTransaction(swapResult.TxHash)
        if err != nil {
            log.Printf("监控交易失败: %v", err)
        } else {
            fmt.Printf("交易状态: %s\n", status.Status)
            if status.Confirmed {
                fmt.Printf("区块号: %d\n", status.BlockNumber)
                fmt.Printf("Gas使用: %d\n", status.GasUsed)
            }
        }
    }

    // 价格查询示例
    fmt.Println("\n价格查询示例:")
    
    // 假设的USDC/ETH池地址
    usdcEthPair := common.HexToAddress("0xB4e16d0168e52d35CaCD2c6185b44281Ec28C9Dc")
    
    price, err := priceOracle.GetTokenPrice(config.USDC_ADDR, cfg.WETH, usdcEthPair)
    if err != nil {
        log.Printf("获取价格失败: %v", err)
    } else {
        fmt.Printf("USDC/ETH价格: %s\n", price.String())
    }

    // 计算价格影响
    priceImpact, err := priceOracle.CalculatePriceImpact(
        usdcEthPair,
        decimal.NewFromFloat(1000), // 1000 USDC
        true, // USDC是tokenA
    )
    if err != nil {
        log.Printf("计算价格影响失败: %v", err)
    } else {
        fmt.Printf("价格影响: %s%%\n", 
            priceImpact.Mul(decimal.NewFromInt(100)).String())
    }

    // 套利机器人示例
    fmt.Println("\n启动套利机器人...")
    
    arbitrageBot := arbitrage.NewArbitrageBot(
        tradingManager,
        priceOracle,
        decimal.NewFromFloat(0.01), // 1%最小利润率
    )

    // 定义监控的代币对
    tokenPairs := []arbitrage.TokenPair{
        {
            TokenA:     config.USDC_ADDR,
            TokenB:     cfg.WETH,
            V2PairAddr: usdcEthPair,
            V3PoolAddr: common.HexToAddress("0x88e6A0c2dDD26FEEb64F039a2c41296FcB3f5640"),
        },
    }

    // 寻找套利机会
    opportunities, err := arbitrageBot.FindArbitrageOpportunities(tokenPairs)
    if err != nil {
        log.Printf("寻找套利机会失败: %v", err)
    } else {
        fmt.Printf("发现 %d 个套利机会\n", len(opportunities))
        
        for i, opportunity := range opportunities {
            fmt.Printf("机会 %d:\n", i+1)
            fmt.Printf("  代币对: %s/%s\n", 
                opportunity.TokenA.Hex()[:8], 
                opportunity.TokenB.Hex()[:8])
            fmt.Printf("  V2价格: %s\n", opportunity.V2Price.String())
            fmt.Printf("  V3价格: %s\n", opportunity.V3Price.String())
            fmt.Printf("  价格差异: %s%%\n", 
                opportunity.PriceDiff.Mul(decimal.NewFromInt(100)).String())
            fmt.Printf("  预期利润: %s\n", opportunity.ExpectedProfit.String())
        }
    }

    // 获取多个代币价格
    fmt.Println("\n多代币价格查询:")
    tokens := []common.Address{config.USDC_ADDR, config.USDT_ADDR, config.DAI_ADDR}
    basePairs := []common.Address{usdcEthPair, usdcEthPair, usdcEthPair} // 简化示例

    prices, err := priceOracle.GetMultipleTokenPrices(tokens, basePairs)
    if err != nil {
        log.Printf("获取多代币价格失败: %v", err)
    } else {
        for token, price := range prices {
            fmt.Printf("代币 %s 价格: %s\n", token.Hex()[:8], price.String())
        }
    }

    fmt.Println("Uniswap操作演示完成!")
}
```
