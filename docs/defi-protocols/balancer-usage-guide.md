# Balancer 协议 详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [环境准备](#环境准备)
3. [权重池机制](#权重池机制)
4. [流动性管理](#流动性管理)
5. [交易和套利](#交易和套利)
6. [治理和激励](#治理和激励)
7. [智能池策略](#智能池策略)
8. [收益优化](#收益优化)
9. [实际应用](#实际应用)

## 基础概念

### 1.1 Balancer 简介

Balancer 是自动化做市商(AMM)和去中心化交易所，支持多资产权重池，提供灵活的流动性管理和交易功能，是DeFi生态的重要基础设施。

```bash
# 安装Balancer相关依赖
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

// Balancer V2 核心合约地址 (Mainnet)
var (
    // Vault (核心合约)
    VaultAddress = common.HexToAddress("0xBA12222222228d8Ba445958a75a0704d566BF2C8")
    
    // Weighted Pool Factory
    WeightedPoolFactoryAddress = common.HexToAddress("0x8E9aa87E45f92bad84D5F8DD1bff34Fb92637dE9")
    
    // Stable Pool Factory
    StablePoolFactoryAddress = common.HexToAddress("0xc66Ba2B6595D3613CCab350C886aCE23866EDe24")
    
    // Liquidity Bootstrapping Pool Factory
    LBPFactoryAddress = common.HexToAddress("0x751A0bC0e3f75b38e01Cf25bFCE7fF36DE1C87DE")
    
    // BAL Token
    BALTokenAddress = common.HexToAddress("0xba100000625a3754423978a60c9317c58a424e3D")
    
    // Balancer Minter
    BalancerMinterAddress = common.HexToAddress("0x239e55F427D44C3cc793f49bFB507ebe76638a2b")
    
    // Gauge Controller
    GaugeControllerAddress = common.HexToAddress("0xC128468b7Ce63eA702C1f104D55A2566b13D3ABD")
    
    // Authorizer
    AuthorizerAddress = common.HexToAddress("0xA331D84eC860Bf466b4CdCcFb4aC09a1B43F3aE6")
)

// 池类型定义
const (
    WeightedPoolType = "WeightedPool"
    StablePoolType   = "StablePool"
    MetaStablePoolType = "MetaStablePool"
    LiquidityBootstrappingPoolType = "LiquidityBootstrappingPool"
    InvestmentPoolType = "InvestmentPool"
)

// 池信息
type Pool struct {
    ID              [32]byte
    Address         common.Address
    PoolType        string
    Tokens          []common.Address
    Balances        []*big.Int
    Weights         []*big.Int
    SwapFeePercentage *big.Int
    TotalSupply     *big.Int
    Owner           common.Address
    SwapEnabled     bool
    PauseWindowEndTime *big.Int
    BufferPeriodEndTime *big.Int
}

// 权重池配置
type WeightedPoolConfig struct {
    Name            string
    Symbol          string
    Tokens          []common.Address
    Weights         []*big.Int
    SwapFeePercentage *big.Int
    Owner           common.Address
}

// 稳定池配置
type StablePoolConfig struct {
    Name            string
    Symbol          string
    Tokens          []common.Address
    AmplificationParameter *big.Int
    SwapFeePercentage *big.Int
    Owner           common.Address
}

// 交易请求
type SwapRequest struct {
    Kind            uint8           // 0: GIVEN_IN, 1: GIVEN_OUT
    TokenIn         common.Address
    TokenOut        common.Address
    Amount          *big.Int
    PoolId          [32]byte
    LastChangeBlock *big.Int
    From            common.Address
    To              common.Address
    UserData        []byte
}

// 批量交易
type BatchSwapStep struct {
    PoolId          [32]byte
    AssetInIndex    *big.Int
    AssetOutIndex   *big.Int
    Amount          *big.Int
    UserData        []byte
}

// 流动性操作
type JoinPoolRequest struct {
    Assets          []common.Address
    MaxAmountsIn    []*big.Int
    UserData        []byte
    FromInternalBalance bool
}

type ExitPoolRequest struct {
    Assets          []common.Address
    MinAmountsOut   []*big.Int
    UserData        []byte
    ToInternalBalance bool
}

// 治理信息
type GovernanceInfo struct {
    BALBalance      *big.Int
    VeBALBalance    *big.Int
    VotingPower     *big.Int
    LockedUntil     *big.Int
    GaugeWeights    map[common.Address]*big.Int
}

// 收益信息
type YieldInfo struct {
    SwapFees        *big.Int
    BALRewards      *big.Int
    ExternalRewards map[common.Address]*big.Int
    APR             decimal.Decimal
    TotalYield      *big.Int
}

// 套利机会
type ArbitrageOpportunity struct {
    PoolA           common.Address
    PoolB           common.Address
    TokenIn         common.Address
    TokenOut        common.Address
    AmountIn        *big.Int
    ExpectedProfit  *big.Int
    ProfitPercentage decimal.Decimal
    GasCost         *big.Int
    NetProfit       *big.Int
}
```

## 环境准备

### 2.1 合约ABI定义

```go
// contracts/balancer_abi.go
package contracts

// Vault ABI (简化版)
const VaultABI = `[
    {
        "inputs": [
            {"name": "poolId", "type": "bytes32"}
        ],
        "name": "getPool",
        "outputs": [
            {"name": "", "type": "address"},
            {"name": "", "type": "uint8"}
        ],
        "type": "function"
    },
    {
        "inputs": [
            {"name": "poolId", "type": "bytes32"}
        ],
        "name": "getPoolTokens",
        "outputs": [
            {"name": "tokens", "type": "address[]"},
            {"name": "balances", "type": "uint256[]"},
            {"name": "lastChangeBlock", "type": "uint256"}
        ],
        "type": "function"
    },
    {
        "inputs": [
            {"name": "singleSwap", "type": "tuple"},
            {"name": "funds", "type": "tuple"},
            {"name": "limit", "type": "uint256"},
            {"name": "deadline", "type": "uint256"}
        ],
        "name": "swap",
        "outputs": [{"name": "", "type": "uint256"}],
        "payable": true,
        "type": "function"
    },
    {
        "inputs": [
            {"name": "kind", "type": "uint8"},
            {"name": "swaps", "type": "tuple[]"},
            {"name": "assets", "type": "address[]"},
            {"name": "funds", "type": "tuple"},
            {"name": "limits", "type": "int256[]"},
            {"name": "deadline", "type": "uint256"}
        ],
        "name": "batchSwap",
        "outputs": [{"name": "", "type": "int256[]"}],
        "payable": true,
        "type": "function"
    },
    {
        "inputs": [
            {"name": "poolId", "type": "bytes32"},
            {"name": "sender", "type": "address"},
            {"name": "recipient", "type": "address"},
            {"name": "request", "type": "tuple"}
        ],
        "name": "joinPool",
        "outputs": [],
        "payable": true,
        "type": "function"
    },
    {
        "inputs": [
            {"name": "poolId", "type": "bytes32"},
            {"name": "sender", "type": "address"},
            {"name": "recipient", "type": "address"},
            {"name": "request", "type": "tuple"}
        ],
        "name": "exitPool",
        "outputs": [],
        "type": "function"
    }
]`

// Weighted Pool Factory ABI (简化版)
const WeightedPoolFactoryABI = `[
    {
        "inputs": [
            {"name": "name", "type": "string"},
            {"name": "symbol", "type": "string"},
            {"name": "tokens", "type": "address[]"},
            {"name": "weights", "type": "uint256[]"},
            {"name": "swapFeePercentage", "type": "uint256"},
            {"name": "owner", "type": "address"}
        ],
        "name": "create",
        "outputs": [{"name": "", "type": "address"}],
        "type": "function"
    }
]`

// BAL Token ABI (简化版)
const BALTokenABI = `[
    {
        "inputs": [{"name": "account", "type": "address"}],
        "name": "balanceOf",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    },
    {
        "inputs": [],
        "name": "totalSupply",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    }
]`
```

### 2.2 Balancer客户端设置

```go
// client/balancer_client.go
package client

import (
    "context"
    "math/big"
    
    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
)

type BalancerClient struct {
    ethClient           *ethclient.Client
    vaultABI            abi.ABI
    weightedPoolFactoryABI abi.ABI
    balTokenABI         abi.ABI
}

func NewBalancerClient(rpcURL string) (*BalancerClient, error) {
    ethClient, err := ethclient.Dial(rpcURL)
    if err != nil {
        return nil, err
    }
    
    vaultABI, err := abi.JSON(strings.NewReader(VaultABI))
    if err != nil {
        return nil, err
    }
    
    weightedPoolFactoryABI, err := abi.JSON(strings.NewReader(WeightedPoolFactoryABI))
    if err != nil {
        return nil, err
    }
    
    balTokenABI, err := abi.JSON(strings.NewReader(BALTokenABI))
    if err != nil {
        return nil, err
    }
    
    return &BalancerClient{
        ethClient:              ethClient,
        vaultABI:               vaultABI,
        weightedPoolFactoryABI: weightedPoolFactoryABI,
        balTokenABI:            balTokenABI,
    }, nil
}

// 获取池信息
func (c *BalancerClient) GetPool(poolId [32]byte) (*Pool, error) {
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    // 获取池地址和类型
    var poolResult []interface{}
    err := c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &VaultAddress,
        Data: c.vaultABI.Methods["getPool"].ID,
    }, nil)
    if err != nil {
        return nil, err
    }
    
    poolAddress := poolResult[0].(common.Address)
    poolType := poolResult[1].(uint8)
    
    // 获取池代币和余额
    var tokensResult []interface{}
    err = c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &VaultAddress,
        Data: c.vaultABI.Methods["getPoolTokens"].ID,
    }, nil)
    if err != nil {
        return nil, err
    }
    
    tokens := tokensResult[0].([]common.Address)
    balances := tokensResult[1].([]*big.Int)
    
    pool := &Pool{
        ID:       poolId,
        Address:  poolAddress,
        Tokens:   tokens,
        Balances: balances,
    }
    
    // 根据池类型设置相应信息
    switch poolType {
    case 0:
        pool.PoolType = WeightedPoolType
    case 1:
        pool.PoolType = StablePoolType
    default:
        pool.PoolType = "Unknown"
    }
    
    return pool, nil
}

// 获取池代币信息
func (c *BalancerClient) GetPoolTokens(poolId [32]byte) ([]common.Address, []*big.Int, error) {
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    var result []interface{}
    err := c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &VaultAddress,
        Data: c.vaultABI.Methods["getPoolTokens"].ID,
    }, nil)
    if err != nil {
        return nil, nil, err
    }
    
    tokens := result[0].([]common.Address)
    balances := result[1].([]*big.Int)
    
    return tokens, balances, nil
}

// 计算交易输出
func (c *BalancerClient) QuerySwap(
    poolId [32]byte,
    tokenIn common.Address,
    tokenOut common.Address,
    amountIn *big.Int,
) (*big.Int, error) {
    // 这里需要调用池合约的查询函数
    // 简化实现，实际需要根据池类型调用不同的计算函数
    
    // 获取池信息
    pool, err := c.GetPool(poolId)
    if err != nil {
        return nil, err
    }
    
    // 简化的价格计算 (实际需要根据池类型使用不同公式)
    var amountOut *big.Int
    
    if pool.PoolType == WeightedPoolType {
        amountOut = c.calculateWeightedPoolSwap(pool, tokenIn, tokenOut, amountIn)
    } else if pool.PoolType == StablePoolType {
        amountOut = c.calculateStablePoolSwap(pool, tokenIn, tokenOut, amountIn)
    }
    
    return amountOut, nil
}

// 获取BAL代币余额
func (c *BalancerClient) GetBALBalance(account common.Address) (*big.Int, error) {
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    var balance *big.Int
    err := c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &BALTokenAddress,
        Data: c.balTokenABI.Methods["balanceOf"].ID,
    }, nil)
    if err != nil {
        return nil, err
    }
    
    return balance, nil
}

// 辅助函数 - 权重池交易计算
func (c *BalancerClient) calculateWeightedPoolSwap(
    pool *Pool,
    tokenIn common.Address,
    tokenOut common.Address,
    amountIn *big.Int,
) *big.Int {
    // 简化的权重池计算
    // 实际公式: amountOut = balanceOut * (1 - (balanceIn / (balanceIn + amountIn))^(weightIn/weightOut))
    
    // 找到代币索引
    var inIndex, outIndex int = -1, -1
    for i, token := range pool.Tokens {
        if token == tokenIn {
            inIndex = i
        }
        if token == tokenOut {
            outIndex = i
        }
    }
    
    if inIndex == -1 || outIndex == -1 {
        return big.NewInt(0)
    }
    
    // 简化计算 (实际需要使用精确的数学公式)
    balanceIn := pool.Balances[inIndex]
    balanceOut := pool.Balances[outIndex]
    
    // 简单的比例计算 (不考虑权重和费用)
    amountOut := new(big.Int).Mul(amountIn, balanceOut)
    amountOut.Div(amountOut, new(big.Int).Add(balanceIn, amountIn))
    
    return amountOut
}

// 辅助函数 - 稳定池交易计算
func (c *BalancerClient) calculateStablePoolSwap(
    pool *Pool,
    tokenIn common.Address,
    tokenOut common.Address,
    amountIn *big.Int,
) *big.Int {
    // 简化的稳定池计算
    // 实际使用StableSwap不变量公式
    
    // 找到代币索引
    var inIndex, outIndex int = -1, -1
    for i, token := range pool.Tokens {
        if token == tokenIn {
            inIndex = i
        }
        if token == tokenOut {
            outIndex = i
        }
    }
    
    if inIndex == -1 || outIndex == -1 {
        return big.NewInt(0)
    }
    
    // 简化计算 (稳定币池接近1:1兑换)
    swapFee := big.NewInt(3) // 0.3%
    feeAmount := new(big.Int).Mul(amountIn, swapFee)
    feeAmount.Div(feeAmount, big.NewInt(1000))
    
    amountOut := new(big.Int).Sub(amountIn, feeAmount)
    
    return amountOut
}
```

## 权重池机制

### 3.1 池管理服务

```go
// services/pool_service.go
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

type BalancerPoolService struct {
    client     *BalancerClient
    privateKey *ecdsa.PrivateKey
}

func NewBalancerPoolService(client *BalancerClient, privateKey *ecdsa.PrivateKey) *BalancerPoolService {
    return &BalancerPoolService{
        client:     client,
        privateKey: privateKey,
    }
}

// 创建权重池
func (s *BalancerPoolService) CreateWeightedPool(config *WeightedPoolConfig) (*types.Transaction, error) {
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 验证权重总和为100%
    totalWeight := big.NewInt(0)
    for _, weight := range config.Weights {
        totalWeight.Add(totalWeight, weight)
    }
    
    expectedTotal := big.NewInt(1e18) // 100% = 1e18
    if totalWeight.Cmp(expectedTotal) != 0 {
        return nil, fmt.Errorf("权重总和必须为100%%, 当前为: %s", totalWeight.String())
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
    
    // 构建创建池交易数据
    data, err := s.client.weightedPoolFactoryABI.Pack(
        "create",
        config.Name,
        config.Symbol,
        config.Tokens,
        config.Weights,
        config.SwapFeePercentage,
        config.Owner,
    )
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(
        nonce,
        WeightedPoolFactoryAddress,
        big.NewInt(0),
        500000,
        gasPrice,
        data,
    )
    
    return s.signAndSendTransaction(tx)
}

// 添加流动性
func (s *BalancerPoolService) JoinPool(
    poolId [32]byte,
    assets []common.Address,
    maxAmountsIn []*big.Int,
    userData []byte,
) (*types.Transaction, error) {
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 构建加入池请求
    joinRequest := JoinPoolRequest{
        Assets:              assets,
        MaxAmountsIn:        maxAmountsIn,
        UserData:            userData,
        FromInternalBalance: false,
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
    
    // 构建加入池交易数据
    data, err := s.client.vaultABI.Pack(
        "joinPool",
        poolId,
        fromAddress,
        fromAddress,
        joinRequest,
    )
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(
        nonce,
        VaultAddress,
        big.NewInt(0),
        400000,
        gasPrice,
        data,
    )
    
    return s.signAndSendTransaction(tx)
}

// 移除流动性
func (s *BalancerPoolService) ExitPool(
    poolId [32]byte,
    assets []common.Address,
    minAmountsOut []*big.Int,
    userData []byte,
) (*types.Transaction, error) {
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 构建退出池请求
    exitRequest := ExitPoolRequest{
        Assets:            assets,
        MinAmountsOut:     minAmountsOut,
        UserData:          userData,
        ToInternalBalance: false,
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
    
    // 构建退出池交易数据
    data, err := s.client.vaultABI.Pack(
        "exitPool",
        poolId,
        fromAddress,
        fromAddress,
        exitRequest,
    )
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(
        nonce,
        VaultAddress,
        big.NewInt(0),
        400000,
        gasPrice,
        data,
    )
    
    return s.signAndSendTransaction(tx)
}

// 计算最优权重配置
func (s *BalancerPoolService) CalculateOptimalWeights(
    tokens []common.Address,
    targetAllocations []decimal.Decimal,
) ([]*big.Int, error) {
    if len(tokens) != len(targetAllocations) {
        return nil, fmt.Errorf("代币数量与目标配置数量不匹配")
    }
    
    // 验证配置总和为100%
    totalAllocation := decimal.Zero
    for _, allocation := range targetAllocations {
        totalAllocation = totalAllocation.Add(allocation)
    }
    
    if !totalAllocation.Equal(decimal.NewFromInt(1)) {
        return nil, fmt.Errorf("目标配置总和必须为100%%")
    }
    
    // 转换为权重 (1e18精度)
    weights := make([]*big.Int, len(targetAllocations))
    for i, allocation := range targetAllocations {
        weight := allocation.Mul(decimal.NewFromInt(1e18))
        weights[i] = weight.BigInt()
    }
    
    return weights, nil
}

// 计算无常损失
func (s *BalancerPoolService) CalculateImpermanentLoss(
    poolId [32]byte,
    initialPrices []decimal.Decimal,
    currentPrices []decimal.Decimal,
    weights []decimal.Decimal,
) (*ImpermanentLossInfo, error) {
    if len(initialPrices) != len(currentPrices) || len(initialPrices) != len(weights) {
        return nil, fmt.Errorf("价格和权重数组长度不匹配")
    }
    
    // 计算价格变化比率
    priceRatios := make([]decimal.Decimal, len(initialPrices))
    for i := range initialPrices {
        if initialPrices[i].IsZero() {
            return nil, fmt.Errorf("初始价格不能为零")
        }
        priceRatios[i] = currentPrices[i].Div(initialPrices[i])
    }
    
    // 计算池价值变化
    poolValueRatio := decimal.NewFromInt(1)
    for i, ratio := range priceRatios {
        poolValueRatio = poolValueRatio.Mul(ratio.Pow(weights[i]))
    }
    
    // 计算持有价值变化
    holdValueRatio := decimal.Zero
    for i, ratio := range priceRatios {
        holdValueRatio = holdValueRatio.Add(weights[i].Mul(ratio))
    }
    
    // 计算无常损失
    impermanentLoss := poolValueRatio.Div(holdValueRatio).Sub(decimal.NewFromInt(1))
    
    return &ImpermanentLossInfo{
        PoolValueRatio:    poolValueRatio,
        HoldValueRatio:    holdValueRatio,
        ImpermanentLoss:   impermanentLoss,
        LossPercentage:    impermanentLoss.Mul(decimal.NewFromInt(100)),
    }, nil
}

// 获取池统计信息
func (s *BalancerPoolService) GetPoolStats(poolId [32]byte) (*PoolStats, error) {
    // 获取池信息
    pool, err := s.client.GetPool(poolId)
    if err != nil {
        return nil, err
    }
    
    // 计算总价值锁定 (TVL)
    tvl := decimal.Zero
    for i, balance := range pool.Balances {
        // 这里需要获取代币价格，简化实现
        tokenPrice := decimal.NewFromInt(1) // 假设价格
        tokenValue := decimal.NewFromBigInt(balance, -18).Mul(tokenPrice)
        tvl = tvl.Add(tokenValue)
    }
    
    // 计算24小时交易量 (需要从事件日志获取)
    volume24h := decimal.NewFromInt(1000000) // 简化实现
    
    // 计算费用收入
    swapFeeRate := decimal.NewFromBigInt(pool.SwapFeePercentage, -18)
    feeIncome := volume24h.Mul(swapFeeRate)
    
    // 计算APR
    apr := feeIncome.Mul(decimal.NewFromInt(365)).Div(tvl).Mul(decimal.NewFromInt(100))
    
    return &PoolStats{
        TVL:           tvl,
        Volume24h:     volume24h,
        FeeIncome24h:  feeIncome,
        APR:           apr,
        SwapFeeRate:   swapFeeRate,
        TokenCount:    len(pool.Tokens),
    }, nil
}

// 辅助函数
func (s *BalancerPoolService) signAndSendTransaction(tx *types.Transaction) (*types.Transaction, error) {
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

type ImpermanentLossInfo struct {
    PoolValueRatio  decimal.Decimal
    HoldValueRatio  decimal.Decimal
    ImpermanentLoss decimal.Decimal
    LossPercentage  decimal.Decimal
}

type PoolStats struct {
    TVL          decimal.Decimal
    Volume24h    decimal.Decimal
    FeeIncome24h decimal.Decimal
    APR          decimal.Decimal
    SwapFeeRate  decimal.Decimal
    TokenCount   int
}
```

## 交易和套利

### 4.1 交易服务

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

type BalancerTradingService struct {
    client     *BalancerClient
    privateKey *ecdsa.PrivateKey
}

func NewBalancerTradingService(client *BalancerClient, privateKey *ecdsa.PrivateKey) *BalancerTradingService {
    return &BalancerTradingService{
        client:     client,
        privateKey: privateKey,
    }
}

// 单次交易
func (s *BalancerTradingService) Swap(
    poolId [32]byte,
    tokenIn common.Address,
    tokenOut common.Address,
    amountIn *big.Int,
    minAmountOut *big.Int,
    deadline *big.Int,
) (*types.Transaction, error) {
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 构建单次交易请求
    singleSwap := struct {
        PoolId   [32]byte
        Kind     uint8
        AssetIn  common.Address
        AssetOut common.Address
        Amount   *big.Int
        UserData []byte
    }{
        PoolId:   poolId,
        Kind:     0, // GIVEN_IN
        AssetIn:  tokenIn,
        AssetOut: tokenOut,
        Amount:   amountIn,
        UserData: []byte{},
    }
    
    // 构建资金管理
    funds := struct {
        Sender              common.Address
        FromInternalBalance bool
        Recipient           common.Address
        ToInternalBalance   bool
    }{
        Sender:              fromAddress,
        FromInternalBalance: false,
        Recipient:           fromAddress,
        ToInternalBalance:   false,
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
    
    // 构建交易数据
    data, err := s.client.vaultABI.Pack(
        "swap",
        singleSwap,
        funds,
        minAmountOut,
        deadline,
    )
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(
        nonce,
        VaultAddress,
        big.NewInt(0),
        300000,
        gasPrice,
        data,
    )
    
    return s.signAndSendTransaction(tx)
}

// 批量交易
func (s *BalancerTradingService) BatchSwap(
    swaps []BatchSwapStep,
    assets []common.Address,
    limits []*big.Int,
    deadline *big.Int,
) (*types.Transaction, error) {
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 构建资金管理
    funds := struct {
        Sender              common.Address
        FromInternalBalance bool
        Recipient           common.Address
        ToInternalBalance   bool
    }{
        Sender:              fromAddress,
        FromInternalBalance: false,
        Recipient:           fromAddress,
        ToInternalBalance:   false,
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
    
    // 构建批量交易数据
    data, err := s.client.vaultABI.Pack(
        "batchSwap",
        uint8(0), // GIVEN_IN
        swaps,
        assets,
        funds,
        limits,
        deadline,
    )
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(
        nonce,
        VaultAddress,
        big.NewInt(0),
        500000,
        gasPrice,
        data,
    )
    
    return s.signAndSendTransaction(tx)
}

// 寻找套利机会
func (s *BalancerTradingService) FindArbitrageOpportunities(
    tokenA common.Address,
    tokenB common.Address,
    pools [][32]byte,
) ([]*ArbitrageOpportunity, error) {
    var opportunities []*ArbitrageOpportunity
    
    // 检查所有池对之间的价格差异
    for i := 0; i < len(pools); i++ {
        for j := i + 1; j < len(pools); j++ {
            poolA := pools[i]
            poolB := pools[j]
            
            // 计算池A中的价格
            amountTest := big.NewInt(1e18) // 1个代币
            amountOutA, err := s.client.QuerySwap(poolA, tokenA, tokenB, amountTest)
            if err != nil {
                continue
            }
            
            // 计算池B中的价格
            amountOutB, err := s.client.QuerySwap(poolB, tokenA, tokenB, amountTest)
            if err != nil {
                continue
            }
            
            // 检查是否存在套利机会
            if amountOutA.Cmp(amountOutB) > 0 {
                // 池A价格更高，在池B买入，池A卖出
                opportunity := s.calculateArbitrageProfit(poolB, poolA, tokenA, tokenB, amountTest)
                if opportunity.NetProfit.Cmp(big.NewInt(0)) > 0 {
                    opportunities = append(opportunities, opportunity)
                }
            } else if amountOutB.Cmp(amountOutA) > 0 {
                // 池B价格更高，在池A买入，池B卖出
                opportunity := s.calculateArbitrageProfit(poolA, poolB, tokenA, tokenB, amountTest)
                if opportunity.NetProfit.Cmp(big.NewInt(0)) > 0 {
                    opportunities = append(opportunities, opportunity)
                }
            }
        }
    }
    
    return opportunities, nil
}

// 执行套利交易
func (s *BalancerTradingService) ExecuteArbitrage(opportunity *ArbitrageOpportunity) (*types.Transaction, error) {
    // 构建批量交易步骤
    swaps := []BatchSwapStep{
        {
            PoolId:        [32]byte{}, // 需要设置为opportunity.PoolA的ID
            AssetInIndex:  big.NewInt(0),
            AssetOutIndex: big.NewInt(1),
            Amount:        opportunity.AmountIn,
            UserData:      []byte{},
        },
        {
            PoolId:        [32]byte{}, // 需要设置为opportunity.PoolB的ID
            AssetInIndex:  big.NewInt(1),
            AssetOutIndex: big.NewInt(0),
            Amount:        big.NewInt(0), // 使用上一步的输出
            UserData:      []byte{},
        },
    }
    
    assets := []common.Address{opportunity.TokenIn, opportunity.TokenOut}
    limits := []*big.Int{opportunity.AmountIn, big.NewInt(0)} // 最大输入，最小输出
    deadline := big.NewInt(time.Now().Unix() + 1200) // 20分钟后过期
    
    return s.BatchSwap(swaps, assets, limits, deadline)
}

// 获取最佳交易路径
func (s *BalancerTradingService) GetBestTradingPath(
    tokenIn common.Address,
    tokenOut common.Address,
    amountIn *big.Int,
    pools [][32]byte,
) (*TradingPath, error) {
    var bestPath *TradingPath
    var bestAmountOut *big.Int = big.NewInt(0)
    
    // 直接交易路径
    for _, poolId := range pools {
        amountOut, err := s.client.QuerySwap(poolId, tokenIn, tokenOut, amountIn)
        if err != nil {
            continue
        }
        
        if amountOut.Cmp(bestAmountOut) > 0 {
            bestAmountOut = amountOut
            bestPath = &TradingPath{
                Type:      "DIRECT",
                Pools:     [][32]byte{poolId},
                Tokens:    []common.Address{tokenIn, tokenOut},
                AmountIn:  amountIn,
                AmountOut: amountOut,
                Hops:      1,
            }
        }
    }
    
    // 多跳交易路径 (简化实现，实际需要更复杂的路径搜索算法)
    // 这里可以实现Dijkstra算法或其他图搜索算法
    
    return bestPath, nil
}

// 辅助函数
func (s *BalancerTradingService) calculateArbitrageProfit(
    buyPool [32]byte,
    sellPool [32]byte,
    tokenIn common.Address,
    tokenOut common.Address,
    amountIn *big.Int,
) *ArbitrageOpportunity {
    // 计算在buyPool买入的成本
    amountOut1, _ := s.client.QuerySwap(buyPool, tokenIn, tokenOut, amountIn)
    
    // 计算在sellPool卖出的收益
    amountOut2, _ := s.client.QuerySwap(sellPool, tokenOut, tokenIn, amountOut1)
    
    // 计算利润
    profit := new(big.Int).Sub(amountOut2, amountIn)
    
    // 估算gas成本
    gasCost := big.NewInt(500000 * 20e9) // 500k gas * 20 gwei
    
    // 计算净利润
    netProfit := new(big.Int).Sub(profit, gasCost)
    
    // 计算利润百分比
    profitPercentage := decimal.Zero
    if amountIn.Cmp(big.NewInt(0)) > 0 {
        profitPercentage = decimal.NewFromBigInt(profit, -18).Div(
            decimal.NewFromBigInt(amountIn, -18),
        ).Mul(decimal.NewFromInt(100))
    }
    
    return &ArbitrageOpportunity{
        PoolA:            common.Address{}, // 需要从poolId转换
        PoolB:            common.Address{}, // 需要从poolId转换
        TokenIn:          tokenIn,
        TokenOut:         tokenOut,
        AmountIn:         amountIn,
        ExpectedProfit:   profit,
        ProfitPercentage: profitPercentage,
        GasCost:          gasCost,
        NetProfit:        netProfit,
    }
}

func (s *BalancerTradingService) signAndSendTransaction(tx *types.Transaction) (*types.Transaction, error) {
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

type TradingPath struct {
    Type      string
    Pools     [][32]byte
    Tokens    []common.Address
    AmountIn  *big.Int
    AmountOut *big.Int
    Hops      int
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
    // 创建Balancer客户端
    balancerClient, err := client.NewBalancerClient("https://mainnet.infura.io/v3/YOUR_PROJECT_ID")
    if err != nil {
        log.Fatal("创建Balancer客户端失败:", err)
    }
    
    // 加载私钥
    privateKey, err := crypto.HexToECDSA("your_private_key_here")
    if err != nil {
        log.Fatal("加载私钥失败:", err)
    }
    
    // 获取用户地址
    userAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
    
    // 创建服务
    poolService := services.NewBalancerPoolService(balancerClient, privateKey)
    tradingService := services.NewBalancerTradingService(balancerClient, privateKey)
    
    // 1. 查询BAL代币余额
    fmt.Printf("=== BAL代币信息 ===\n")
    
    balBalance, err := balancerClient.GetBALBalance(userAddress)
    if err != nil {
        log.Fatal("获取BAL余额失败:", err)
    }
    
    fmt.Printf("用户地址: %s\n", userAddress.Hex())
    fmt.Printf("BAL余额: %s (%.6f BAL)\n", 
        balBalance.String(), 
        decimal.NewFromBigInt(balBalance, -18).InexactFloat64())
    
    // 2. 查询热门池信息
    fmt.Printf("\n=== 热门池信息 ===\n")
    
    // 示例池ID (需要替换为实际的池ID)
    popularPools := []struct {
        ID   [32]byte
        Name string
    }{
        {[32]byte{0x5c, 0x6e, 0xe3, 0x04}, "80BAL-20WETH"},
        {[32]byte{0x96, 0x64, 0x6a, 0x2c}, "50WBTC-50WETH"},
        {[32]byte{0x06, 0xdf, 0x3b, 0x2b}, "33WETH-33USDC-33USDT"},
    }
    
    for _, poolInfo := range popularPools {
        pool, err := balancerClient.GetPool(poolInfo.ID)
        if err != nil {
            log.Printf("获取池 %s 信息失败: %v", poolInfo.Name, err)
            continue
        }
        
        fmt.Printf("%s池:\n", poolInfo.Name)
        fmt.Printf("  地址: %s\n", pool.Address.Hex())
        fmt.Printf("  类型: %s\n", pool.PoolType)
        fmt.Printf("  代币数量: %d\n", len(pool.Tokens))
        
        // 获取池统计
        stats, err := poolService.GetPoolStats(poolInfo.ID)
        if err != nil {
            log.Printf("获取池统计失败: %v", err)
        } else {
            fmt.Printf("  TVL: $%s\n", stats.TVL.StringFixed(2))
            fmt.Printf("  24h交易量: $%s\n", stats.Volume24h.StringFixed(2))
            fmt.Printf("  APR: %s%%\n", stats.APR.StringFixed(2))
            fmt.Printf("  交易费率: %s%%\n", stats.SwapFeeRate.Mul(decimal.NewFromInt(100)).StringFixed(3))
        }
        
        fmt.Println()
    }
    
    // 3. 创建权重池示例
    fmt.Printf("=== 创建权重池示例 ===\n")
    
    // 示例代币地址
    wethAddress := common.HexToAddress("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")
    usdcAddress := common.HexToAddress("0xA0b86a33E6441b8dB4B2b8b8b8b8b8b8b8b8b8b8")
    
    poolConfig := &services.WeightedPoolConfig{
        Name:              "60WETH-40USDC",
        Symbol:            "60WETH-40USDC-BPT",
        Tokens:            []common.Address{wethAddress, usdcAddress},
        Weights:           []*big.Int{big.NewInt(6e17), big.NewInt(4e17)}, // 60%, 40%
        SwapFeePercentage: big.NewInt(3e15), // 0.3%
        Owner:             userAddress,
    }
    
    fmt.Printf("准备创建权重池:\n")
    fmt.Printf("  名称: %s\n", poolConfig.Name)
    fmt.Printf("  符号: %s\n", poolConfig.Symbol)
    fmt.Printf("  权重: 60%% WETH, 40%% USDC\n")
    fmt.Printf("  交易费: 0.3%%\n")
    
    // 计算最优权重
    targetAllocations := []decimal.Decimal{
        decimal.NewFromFloat(0.6), // 60%
        decimal.NewFromFloat(0.4), // 40%
    }
    
    optimalWeights, err := poolService.CalculateOptimalWeights(poolConfig.Tokens, targetAllocations)
    if err != nil {
        log.Printf("计算最优权重失败: %v", err)
    } else {
        fmt.Printf("最优权重配置验证通过\n")
        
        // 创建池 (注释掉实际执行)
        // tx, err := poolService.CreateWeightedPool(poolConfig)
        // if err != nil {
        //     log.Printf("创建权重池失败: %v", err)
        // } else {
        //     fmt.Printf("创建池交易已提交: %s\n", tx.Hash().Hex())
        // }
    }
    
    // 4. 交易路径优化示例
    fmt.Printf("\n=== 交易路径优化示例 ===\n")
    
    tokenIn := wethAddress
    tokenOut := usdcAddress
    amountIn := big.NewInt(1e18) // 1 WETH
    
    // 示例池ID列表
    availablePools := [][32]byte{
        popularPools[0].ID,
        popularPools[1].ID,
        popularPools[2].ID,
    }
    
    fmt.Printf("寻找最佳交易路径:\n")
    fmt.Printf("  输入: %s WETH\n", decimal.NewFromBigInt(amountIn, -18).String())
    fmt.Printf("  输出代币: USDC\n")
    
    bestPath, err := tradingService.GetBestTradingPath(tokenIn, tokenOut, amountIn, availablePools)
    if err != nil {
        log.Printf("获取最佳交易路径失败: %v", err)
    } else if bestPath != nil {
        fmt.Printf("最佳路径:\n")
        fmt.Printf("  类型: %s\n", bestPath.Type)
        fmt.Printf("  跳数: %d\n", bestPath.Hops)
        fmt.Printf("  预期输出: %s USDC\n", 
            decimal.NewFromBigInt(bestPath.AmountOut, -6).String())
        
        // 执行交易 (注释掉实际执行)
        // deadline := big.NewInt(time.Now().Unix() + 1200)
        // tx, err := tradingService.Swap(
        //     bestPath.Pools[0],
        //     tokenIn,
        //     tokenOut,
        //     amountIn,
        //     bestPath.AmountOut,
        //     deadline,
        // )
        // if err != nil {
        //     log.Printf("执行交易失败: %v", err)
        // } else {
        //     fmt.Printf("交易已提交: %s\n", tx.Hash().Hex())
        // }
    }
    
    // 5. 套利机会分析
    fmt.Printf("\n=== 套利机会分析 ===\n")
    
    arbitrageOpportunities, err := tradingService.FindArbitrageOpportunities(
        tokenIn,
        tokenOut,
        availablePools,
    )
    if err != nil {
        log.Printf("寻找套利机会失败: %v", err)
    } else {
        fmt.Printf("发现 %d 个套利机会:\n", len(arbitrageOpportunities))
        
        for i, opportunity := range arbitrageOpportunities {
            fmt.Printf("  机会 %d:\n", i+1)
            fmt.Printf("    输入金额: %s\n", 
                decimal.NewFromBigInt(opportunity.AmountIn, -18).String())
            fmt.Printf("    预期利润: %s\n", 
                decimal.NewFromBigInt(opportunity.ExpectedProfit, -18).String())
            fmt.Printf("    利润率: %s%%\n", 
                opportunity.ProfitPercentage.StringFixed(2))
            fmt.Printf("    Gas成本: %s ETH\n", 
                decimal.NewFromBigInt(opportunity.GasCost, -18).String())
            fmt.Printf("    净利润: %s ETH\n", 
                decimal.NewFromBigInt(opportunity.NetProfit, -18).String())
            
            if opportunity.NetProfit.Cmp(big.NewInt(0)) > 0 {
                fmt.Printf("    状态: ✅ 有利可图\n")
            } else {
                fmt.Printf("    状态: ❌ 无利可图\n")
            }
        }
    }
    
    // 6. 无常损失计算示例
    fmt.Printf("\n=== 无常损失计算示例 ===\n")
    
    // 假设价格变化
    initialPrices := []decimal.Decimal{
        decimal.NewFromInt(2000), // ETH初始价格 $2000
        decimal.NewFromInt(1),    // USDC初始价格 $1
    }
    
    currentPrices := []decimal.Decimal{
        decimal.NewFromInt(3000), // ETH当前价格 $3000 (+50%)
        decimal.NewFromInt(1),    // USDC当前价格 $1
    }
    
    weights := []decimal.Decimal{
        decimal.NewFromFloat(0.6), // 60% ETH
        decimal.NewFromFloat(0.4), // 40% USDC
    }
    
    ilInfo, err := poolService.CalculateImpermanentLoss(
        popularPools[0].ID,
        initialPrices,
        currentPrices,
        weights,
    )
    if err != nil {
        log.Printf("计算无常损失失败: %v", err)
    } else {
        fmt.Printf("无常损失分析:\n")
        fmt.Printf("  价格变化: ETH +50%%, USDC 0%%\n")
        fmt.Printf("  池价值比率: %s\n", ilInfo.PoolValueRatio.StringFixed(4))
        fmt.Printf("  持有价值比率: %s\n", ilInfo.HoldValueRatio.StringFixed(4))
        fmt.Printf("  无常损失: %s%%\n", ilInfo.LossPercentage.StringFixed(2))
        
        if ilInfo.ImpermanentLoss.IsNegative() {
            fmt.Printf("  结果: ❌ 存在无常损失\n")
        } else {
            fmt.Printf("  结果: ✅ 无无常损失\n")
        }
    }
    
    // 7. Balancer特性总结
    fmt.Printf("\n=== Balancer特性总结 ===\n")
    
    fmt.Printf("Balancer优势:\n")
    fmt.Printf("  - 灵活的权重配置 (不限于50/50)\n")
    fmt.Printf("  - 多资产池 (最多8个代币)\n")
    fmt.Printf("  - 智能池和可编程流动性\n")
    fmt.Printf("  - 流动性引导池 (LBP)\n")
    fmt.Printf("  - 强大的治理机制\n")
    
    fmt.Printf("\n池类型:\n")
    fmt.Printf("  - 权重池: 自定义权重配置\n")
    fmt.Printf("  - 稳定池: 相似价值资产\n")
    fmt.Printf("  - 元稳定池: 包装代币池\n")
    fmt.Printf("  - 流动性引导池: 价格发现\n")
    fmt.Printf("  - 投资池: 主动管理\n")
    
    // 8. 最佳实践建议
    fmt.Printf("\n=== 最佳实践建议 ===\n")
    
    fmt.Printf("使用Balancer时请注意:\n")
    fmt.Printf("  1. 理解不同池类型的特点和风险\n")
    fmt.Printf("  2. 合理设置权重以降低无常损失\n")
    fmt.Printf("  3. 关注池的流动性和交易量\n")
    fmt.Printf("  4. 考虑BAL代币激励和治理参与\n")
    fmt.Printf("  5. 监控套利机会和MEV风险\n")
    fmt.Printf("  6. 定期重新平衡投资组合\n")
    fmt.Printf("  7. 了解智能池的管理费用\n")
    fmt.Printf("  8. 评估gas成本对小额交易的影响\n")
}
```

这个Balancer使用指南提供了完整的权重池DEX集成方案，涵盖了多资产池管理、灵活权重配置、智能交易路径、套利策略、无常损失计算等核心功能，是构建高级DeFi投资组合管理应用的重要参考文档。
