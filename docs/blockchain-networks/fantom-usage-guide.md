# Fantom 区块链 详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [环境准备](#环境准备)
3. [Opera主网](#opera主网)
4. [智能合约部署](#智能合约部署)
5. [DeFi生态集成](#defi生态集成)
6. [跨链桥接](#跨链桥接)
7. [性能优化](#性能优化)
8. [监控和分析](#监控和分析)
9. [实际应用](#实际应用)

## 基础概念

### 1.1 Fantom 简介

Fantom 是高性能、可扩展的区块链平台，采用Lachesis共识机制实现快速确认和低费用，支持EVM兼容的智能合约和DeFi应用。

```bash
# 安装Fantom相关依赖
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

// Fantom 网络配置
var (
    // Mainnet (Opera)
    MainnetRPC = "https://rpc.ftm.tools"
    MainnetChainID = big.NewInt(250)
    
    // Testnet
    TestnetRPC = "https://rpc.testnet.fantom.network"
    TestnetChainID = big.NewInt(4002)
    
    // 备用RPC端点
    AlternativeRPCs = []string{
        "https://rpc.fantom.network",
        "https://rpc2.fantom.network",
        "https://rpc3.fantom.network",
        "https://fantom-mainnet.gateway.pokt.network/v1/lb/62759259ea1b320039c9e7ac",
    }
)

// Fantom 网络信息
type NetworkInfo struct {
    NetworkID    *big.Int
    NetworkName  string
    ChainID      *big.Int
    RPCEndpoint  string
    WSEndpoint   string
    ExplorerURL  string
    Currency     string
    Decimals     int
}

var (
    MainnetInfo = NetworkInfo{
        NetworkID:   big.NewInt(250),
        NetworkName: "Fantom Opera",
        ChainID:     MainnetChainID,
        RPCEndpoint: MainnetRPC,
        WSEndpoint:  "wss://wsapi.fantom.network",
        ExplorerURL: "https://ftmscan.com",
        Currency:    "FTM",
        Decimals:    18,
    }
    
    TestnetInfo = NetworkInfo{
        NetworkID:   big.NewInt(4002),
        NetworkName: "Fantom Testnet",
        ChainID:     TestnetChainID,
        RPCEndpoint: TestnetRPC,
        WSEndpoint:  "wss://wsapi.testnet.fantom.network",
        ExplorerURL: "https://testnet.ftmscan.com",
        Currency:    "FTM",
        Decimals:    18,
    }
)

// Fantom 核心合约地址
var (
    // Wrapped FTM
    WFTMAddress = common.HexToAddress("0x21be370D5312f44cB42ce377BC9b8a0cEF1A4C83")
    
    // Fantom Foundation
    FantomFoundationAddress = common.HexToAddress("0xFC00FACE00000000000000000000000000000000")
    
    // SFC (Special Fee Contract) - 质押合约
    SFCAddress = common.HexToAddress("0xFC00FACE00000000000000000000000000000000")
    
    // Multicall
    MulticallAddress = common.HexToAddress("0xb828C456600857abd4ed6C32FAcc607bD0464F4F")
)

// DeFi协议地址
var (
    // SpookySwap
    SpookySwapRouterAddress = common.HexToAddress("0xF491e7B69E4244ad4002BC14e878a34207E38c29")
    SpookySwapFactoryAddress = common.HexToAddress("0x152eE697f2E276fA89E96742e9bB9aB1F2E61bE3")
    
    // SpiritSwap
    SpiritSwapRouterAddress = common.HexToAddress("0x16327E3FbDaCA3bcF7E38F5Af2599D2DDc33aE52")
    SpiritSwapFactoryAddress = common.HexToAddress("0xEF45d134b73241eDa7703fa787148D9C9F4950b0")
    
    // Beethoven X (Balancer)
    BeethovenXVaultAddress = common.HexToAddress("0x20dd72Ed959b6147912C2e529F0a0C651c33c9ce")
    
    // Geist Finance (Aave Fork)
    GeistLendingPoolAddress = common.HexToAddress("0x9FAD24f572045c7869117160A571B2e50b10d068")
    
    // Tarot Finance
    TarotFactoryAddress = common.HexToAddress("0x35C052bBf8338b06351782A565aa9AaD173432eA")
    
    // Tomb Finance
    TombFinanceAddress = common.HexToAddress("0x6c021Ae822BEa943b2E66552bDe1D2696a53fbB7")
)

// 验证者信息
type Validator struct {
    ID               *big.Int
    Address          common.Address
    CreatedEpoch     *big.Int
    CreatedTime      *big.Int
    DeactivatedEpoch *big.Int
    DeactivatedTime  *big.Int
    Status           *big.Int
    ReceivedStake    *big.Int
    StakeShare       decimal.Decimal
    DelegatedMe      *big.Int
    DagAddress       common.Address
}

// 委托信息
type Delegation struct {
    ValidatorID      *big.Int
    Delegator        common.Address
    Stake            *big.Int
    LockedStake      *big.Int
    LockupFromEpoch  *big.Int
    LockupEndTime    *big.Int
    LockupDuration   *big.Int
    EarlyUnlockFee   *big.Int
    Rewards          *big.Int
}

// 网络统计
type NetworkStats struct {
    TotalSupply      *big.Int
    CirculatingSupply *big.Int
    TotalStaked      *big.Int
    TotalDelegated   *big.Int
    StakingAPR       decimal.Decimal
    ValidatorCount   *big.Int
    DelegatorCount   *big.Int
    AverageBlockTime decimal.Decimal
    TPS              decimal.Decimal
}
```

## 环境准备

### 2.1 Fantom客户端设置

```go
// client/fantom_client.go
package client

import (
    "context"
    "math/big"
    "time"
    
    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/shopspring/decimal"
)

type FantomClient struct {
    ethClient    *ethclient.Client
    networkInfo  *NetworkInfo
    sfcABI       abi.ABI
    multicallABI abi.ABI
}

func NewFantomClient(networkInfo *NetworkInfo) (*FantomClient, error) {
    // 尝试连接到主RPC端点
    ethClient, err := ethclient.Dial(networkInfo.RPCEndpoint)
    if err != nil {
        // 如果主端点失败，尝试备用端点
        if networkInfo.NetworkName == "Fantom Opera" {
            for _, rpc := range AlternativeRPCs {
                ethClient, err = ethclient.Dial(rpc)
                if err == nil {
                    break
                }
            }
        }
        if err != nil {
            return nil, err
        }
    }
    
    // 加载SFC ABI
    sfcABI, err := abi.JSON(strings.NewReader(SFCABI))
    if err != nil {
        return nil, err
    }
    
    // 加载Multicall ABI
    multicallABI, err := abi.JSON(strings.NewReader(MulticallABI))
    if err != nil {
        return nil, err
    }
    
    return &FantomClient{
        ethClient:    ethClient,
        networkInfo:  networkInfo,
        sfcABI:       sfcABI,
        multicallABI: multicallABI,
    }, nil
}

// 获取网络信息
func (c *FantomClient) GetNetworkInfo() *NetworkInfo {
    return c.networkInfo
}

// 获取以太坊客户端
func (c *FantomClient) EthClient() *ethclient.Client {
    return c.ethClient
}

// 获取FTM余额
func (c *FantomClient) GetBalance(address common.Address) (*big.Int, error) {
    ctx := context.Background()
    return c.ethClient.BalanceAt(ctx, address, nil)
}

// 获取网络统计
func (c *FantomClient) GetNetworkStats() (*NetworkStats, error) {
    ctx := context.Background()
    
    // 获取最新区块
    latestBlock, err := c.ethClient.BlockByNumber(ctx, nil)
    if err != nil {
        return nil, err
    }
    
    // 获取前一个区块计算平均出块时间
    prevBlock, err := c.ethClient.BlockByNumber(ctx, big.NewInt(latestBlock.Number().Int64()-100))
    if err != nil {
        return nil, err
    }
    
    // 计算平均出块时间
    timeDiff := latestBlock.Time() - prevBlock.Time()
    blockDiff := latestBlock.Number().Int64() - prevBlock.Number().Int64()
    avgBlockTime := decimal.NewFromInt(int64(timeDiff)).Div(decimal.NewFromInt(blockDiff))
    
    // 计算TPS (简化计算)
    tps := decimal.NewFromInt(int64(len(latestBlock.Transactions()))).Div(avgBlockTime)
    
    return &NetworkStats{
        AverageBlockTime: avgBlockTime,
        TPS:              tps,
        ValidatorCount:   big.NewInt(50), // 需要从SFC合约获取实际数据
    }, nil
}

// 获取验证者信息
func (c *FantomClient) GetValidator(validatorID *big.Int) (*Validator, error) {
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    // 这里需要调用SFC合约的getValidator方法
    // 简化实现
    return &Validator{
        ID:      validatorID,
        Status:  big.NewInt(1), // Active
    }, nil
}

// 获取所有验证者
func (c *FantomClient) GetValidators() ([]*Validator, error) {
    // 这里需要调用SFC合约获取验证者列表
    // 简化实现
    var validators []*Validator
    
    for i := 1; i <= 50; i++ {
        validator := &Validator{
            ID:     big.NewInt(int64(i)),
            Status: big.NewInt(1),
        }
        validators = append(validators, validator)
    }
    
    return validators, nil
}

// 获取委托信息
func (c *FantomClient) GetDelegation(delegator common.Address, validatorID *big.Int) (*Delegation, error) {
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    // 这里需要调用SFC合约的getDelegation方法
    // 简化实现
    return &Delegation{
        ValidatorID: validatorID,
        Delegator:   delegator,
        Stake:       big.NewInt(0),
    }, nil
}

// 批量调用
func (c *FantomClient) Multicall(calls []MulticallData) ([][]byte, error) {
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    // 构建multicall调用
    var multicallCalls []struct {
        Target   common.Address
        CallData []byte
    }
    
    for _, call := range calls {
        multicallCalls = append(multicallCalls, struct {
            Target   common.Address
            CallData []byte
        }{
            Target:   call.Target,
            CallData: call.CallData,
        })
    }
    
    // 调用multicall合约
    var results [][]byte
    err := c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &MulticallAddress,
        Data: c.multicallABI.Methods["aggregate"].ID,
    }, nil)
    if err != nil {
        return nil, err
    }
    
    return results, nil
}

type MulticallData struct {
    Target   common.Address
    CallData []byte
}

// SFC合约ABI (简化版)
const SFCABI = `[
    {
        "inputs": [{"name": "validatorID", "type": "uint256"}],
        "name": "getValidator",
        "outputs": [
            {"name": "status", "type": "uint256"},
            {"name": "receivedStake", "type": "uint256"},
            {"name": "createdEpoch", "type": "uint256"},
            {"name": "createdTime", "type": "uint256"},
            {"name": "deactivatedEpoch", "type": "uint256"},
            {"name": "deactivatedTime", "type": "uint256"}
        ],
        "type": "function"
    },
    {
        "inputs": [
            {"name": "delegator", "type": "address"},
            {"name": "validatorID", "type": "uint256"}
        ],
        "name": "getDelegation",
        "outputs": [
            {"name": "stake", "type": "uint256"},
            {"name": "lockedStake", "type": "uint256"},
            {"name": "lockupFromEpoch", "type": "uint256"},
            {"name": "lockupEndTime", "type": "uint256"},
            {"name": "lockupDuration", "type": "uint256"},
            {"name": "earlyUnlockFee", "type": "uint256"}
        ],
        "type": "function"
    },
    {
        "inputs": [{"name": "validatorID", "type": "uint256"}],
        "name": "delegate",
        "outputs": [],
        "payable": true,
        "type": "function"
    },
    {
        "inputs": [
            {"name": "validatorID", "type": "uint256"},
            {"name": "amount", "type": "uint256"}
        ],
        "name": "undelegate",
        "outputs": [],
        "type": "function"
    },
    {
        "inputs": [{"name": "validatorID", "type": "uint256"}],
        "name": "claimRewards",
        "outputs": [],
        "type": "function"
    }
]`

// Multicall合约ABI (简化版)
const MulticallABI = `[
    {
        "inputs": [
            {
                "components": [
                    {"name": "target", "type": "address"},
                    {"name": "callData", "type": "bytes"}
                ],
                "name": "calls",
                "type": "tuple[]"
            }
        ],
        "name": "aggregate",
        "outputs": [
            {"name": "blockNumber", "type": "uint256"},
            {"name": "returnData", "type": "bytes[]"}
        ],
        "type": "function"
    }
]`
```

## Opera主网

### 3.1 质押服务

```go
// services/staking_service.go
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

type FantomStakingService struct {
    client     *FantomClient
    privateKey *ecdsa.PrivateKey
}

func NewFantomStakingService(client *FantomClient, privateKey *ecdsa.PrivateKey) *FantomStakingService {
    return &FantomStakingService{
        client:     client,
        privateKey: privateKey,
    }
}

// 委托质押
func (s *FantomStakingService) Delegate(validatorID *big.Int, amount *big.Int) (*types.Transaction, error) {
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 检查余额
    balance, err := s.client.GetBalance(fromAddress)
    if err != nil {
        return nil, err
    }
    
    if balance.Cmp(amount) < 0 {
        return nil, fmt.Errorf("余额不足")
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
    
    // 构建委托交易数据
    data, err := s.client.sfcABI.Pack("delegate", validatorID)
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, SFCAddress, amount, 200000, gasPrice, data)
    
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

// 取消委托
func (s *FantomStakingService) Undelegate(validatorID *big.Int, amount *big.Int) (*types.Transaction, error) {
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 检查委托余额
    delegation, err := s.client.GetDelegation(fromAddress, validatorID)
    if err != nil {
        return nil, err
    }
    
    if delegation.Stake.Cmp(amount) < 0 {
        return nil, fmt.Errorf("取消委托数量超过已委托数量")
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
    
    // 构建取消委托交易数据
    data, err := s.client.sfcABI.Pack("undelegate", validatorID, amount)
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, SFCAddress, big.NewInt(0), 150000, gasPrice, data)
    
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

// 领取奖励
func (s *FantomStakingService) ClaimRewards(validatorID *big.Int) (*types.Transaction, error) {
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
    
    // 构建领取奖励交易数据
    data, err := s.client.sfcABI.Pack("claimRewards", validatorID)
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, SFCAddress, big.NewInt(0), 100000, gasPrice, data)
    
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

// 计算质押奖励
func (s *FantomStakingService) CalculateStakingRewards(
    stakedAmount *big.Int,
    validatorID *big.Int,
    stakingDuration time.Duration,
) (*StakingRewards, error) {
    // 获取验证者信息
    validator, err := s.client.GetValidator(validatorID)
    if err != nil {
        return nil, err
    }
    
    // 基础年化收益率 (简化计算)
    baseAPR := decimal.NewFromFloat(0.04) // 4%
    
    // 根据验证者表现调整
    performanceMultiplier := decimal.NewFromFloat(1.0)
    if validator.Status.Cmp(big.NewInt(1)) == 0 { // Active
        performanceMultiplier = decimal.NewFromFloat(1.1)
    }
    
    // 计算年化收益
    stakedDecimal := decimal.NewFromBigInt(stakedAmount, -18)
    annualReward := stakedDecimal.Mul(baseAPR).Mul(performanceMultiplier)
    
    // 计算期间收益
    durationYears := decimal.NewFromFloat(stakingDuration.Hours() / 8760) // 小时转年
    periodReward := annualReward.Mul(durationYears)
    
    return &StakingRewards{
        StakedAmount:    stakedAmount,
        ValidatorID:     validatorID,
        APR:             baseAPR.Mul(performanceMultiplier),
        AnnualReward:    annualReward.Mul(decimal.NewFromInt(1e18)).BigInt(),
        PeriodReward:    periodReward.Mul(decimal.NewFromInt(1e18)).BigInt(),
        Duration:        stakingDuration,
    }, nil
}

// 获取最佳验证者
func (s *FantomStakingService) GetBestValidators(limit int) ([]*ValidatorRanking, error) {
    validators, err := s.client.GetValidators()
    if err != nil {
        return nil, err
    }
    
    var rankings []*ValidatorRanking
    
    for _, validator := range validators {
        if validator.Status.Cmp(big.NewInt(1)) != 0 { // 只考虑活跃验证者
            continue
        }
        
        // 计算评分 (简化算法)
        score := s.calculateValidatorScore(validator)
        
        ranking := &ValidatorRanking{
            Validator: validator,
            Score:     score,
            APR:       decimal.NewFromFloat(4.0), // 简化APR
            Uptime:    decimal.NewFromFloat(99.5), // 简化在线时间
        }
        
        rankings = append(rankings, ranking)
        
        if len(rankings) >= limit {
            break
        }
    }
    
    // 按评分排序
    sort.Slice(rankings, func(i, j int) bool {
        return rankings[i].Score.GreaterThan(rankings[j].Score)
    })
    
    return rankings, nil
}

// 计算验证者评分
func (s *FantomStakingService) calculateValidatorScore(validator *Validator) decimal.Decimal {
    // 简化的评分算法
    baseScore := decimal.NewFromFloat(100)
    
    // 根据质押量调整 (更多质押 = 更稳定)
    stakeScore := decimal.NewFromBigInt(validator.ReceivedStake, -18).Div(decimal.NewFromInt(1000000))
    
    // 根据运行时间调整
    timeScore := decimal.NewFromBigInt(validator.CreatedEpoch, 0).Div(decimal.NewFromInt(1000))
    
    totalScore := baseScore.Add(stakeScore).Add(timeScore)
    
    return totalScore
}

type StakingRewards struct {
    StakedAmount *big.Int
    ValidatorID  *big.Int
    APR          decimal.Decimal
    AnnualReward *big.Int
    PeriodReward *big.Int
    Duration     time.Duration
}

type ValidatorRanking struct {
    Validator *Validator
    Score     decimal.Decimal
    APR       decimal.Decimal
    Uptime    decimal.Decimal
}
```

## DeFi生态集成

### 4.1 SpookySwap集成

```go
// services/spookyswap_service.go
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

type SpookySwapService struct {
    client     *FantomClient
    privateKey *ecdsa.PrivateKey
    routerABI  abi.ABI
}

func NewSpookySwapService(client *FantomClient, privateKey *ecdsa.PrivateKey) *SpookySwapService {
    routerABI, _ := abi.JSON(strings.NewReader(UniswapV2RouterABI))
    
    return &SpookySwapService{
        client:     client,
        privateKey: privateKey,
        routerABI:  routerABI,
    }
}

// FTM换代币
func (s *SpookySwapService) SwapFTMForTokens(
    tokenOut common.Address,
    amountIn *big.Int,
    amountOutMin *big.Int,
    deadline *big.Int,
) (*types.Transaction, error) {
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 构建交换路径
    path := []common.Address{WFTMAddress, tokenOut}
    
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
    data, err := s.routerABI.Pack(
        "swapExactETHForTokens",
        amountOutMin,
        path,
        fromAddress,
        deadline,
    )
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, SpookySwapRouterAddress, amountIn, 300000, gasPrice, data)
    
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

// 代币换FTM
func (s *SpookySwapService) SwapTokensForFTM(
    tokenIn common.Address,
    amountIn *big.Int,
    amountOutMin *big.Int,
    deadline *big.Int,
) (*types.Transaction, error) {
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 授权Router使用代币
    if err := s.approveToken(tokenIn, SpookySwapRouterAddress, amountIn); err != nil {
        return nil, err
    }
    
    // 构建交换路径
    path := []common.Address{tokenIn, WFTMAddress}
    
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
    data, err := s.routerABI.Pack(
        "swapExactTokensForETH",
        amountIn,
        amountOutMin,
        path,
        fromAddress,
        deadline,
    )
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, SpookySwapRouterAddress, big.NewInt(0), 300000, gasPrice, data)
    
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

// 添加流动性
func (s *SpookySwapService) AddLiquidityFTM(
    token common.Address,
    amountTokenDesired *big.Int,
    amountTokenMin *big.Int,
    amountFTMMin *big.Int,
    deadline *big.Int,
    ftmAmount *big.Int,
) (*types.Transaction, error) {
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 授权Router使用代币
    if err := s.approveToken(token, SpookySwapRouterAddress, amountTokenDesired); err != nil {
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
    
    // 构建交易数据
    data, err := s.routerABI.Pack(
        "addLiquidityETH",
        token,
        amountTokenDesired,
        amountTokenMin,
        amountFTMMin,
        fromAddress,
        deadline,
    )
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, SpookySwapRouterAddress, ftmAmount, 400000, gasPrice, data)
    
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
func (s *SpookySwapService) approveToken(token, spender common.Address, amount *big.Int) error {
    // 这里需要实现ERC20代币授权逻辑
    // 简化实现
    return nil
}

// Uniswap V2 Router ABI (SpookySwap兼容)
const UniswapV2RouterABI = `[
    {
        "inputs": [
            {"name": "amountOutMin", "type": "uint256"},
            {"name": "path", "type": "address[]"},
            {"name": "to", "type": "address"},
            {"name": "deadline", "type": "uint256"}
        ],
        "name": "swapExactETHForTokens",
        "outputs": [{"name": "amounts", "type": "uint256[]"}],
        "payable": true,
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
        "outputs": [{"name": "amounts", "type": "uint256[]"}],
        "type": "function"
    },
    {
        "inputs": [
            {"name": "token", "type": "address"},
            {"name": "amountTokenDesired", "type": "uint256"},
            {"name": "amountTokenMin", "type": "uint256"},
            {"name": "amountETHMin", "type": "uint256"},
            {"name": "to", "type": "address"},
            {"name": "deadline", "type": "uint256"}
        ],
        "name": "addLiquidityETH",
        "outputs": [
            {"name": "amountToken", "type": "uint256"},
            {"name": "amountETH", "type": "uint256"},
            {"name": "liquidity", "type": "uint256"}
        ],
        "payable": true,
        "type": "function"
    }
]`
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
    "time"
    
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/crypto"
    
    "your-project/client"
    "your-project/services"
)

func main() {
    // 创建Fantom客户端 (使用主网)
    fantomClient, err := client.NewFantomClient(&client.MainnetInfo)
    if err != nil {
        log.Fatal("创建Fantom客户端失败:", err)
    }
    
    // 加载私钥
    privateKey, err := crypto.HexToECDSA("your_private_key_here")
    if err != nil {
        log.Fatal("加载私钥失败:", err)
    }
    
    // 获取用户地址
    userAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
    
    // 创建服务
    stakingService := services.NewFantomStakingService(fantomClient, privateKey)
    spookySwapService := services.NewSpookySwapService(fantomClient, privateKey)
    
    // 1. 获取网络信息
    networkInfo := fantomClient.GetNetworkInfo()
    fmt.Printf("=== Fantom网络信息 ===\n")
    fmt.Printf("网络名称: %s\n", networkInfo.NetworkName)
    fmt.Printf("链ID: %s\n", networkInfo.ChainID.String())
    fmt.Printf("RPC端点: %s\n", networkInfo.RPCEndpoint)
    fmt.Printf("浏览器: %s\n", networkInfo.ExplorerURL)
    fmt.Printf("原生代币: %s\n", networkInfo.Currency)
    
    // 2. 获取FTM余额
    fmt.Printf("\n=== 账户信息 ===\n")
    
    balance, err := fantomClient.GetBalance(userAddress)
    if err != nil {
        log.Fatal("获取FTM余额失败:", err)
    }
    
    fmt.Printf("地址: %s\n", userAddress.Hex())
    fmt.Printf("FTM余额: %s (%.6f FTM)\n", 
        balance.String(), 
        float64(balance.Int64())/1e18)
    
    // 3. 获取网络统计
    fmt.Printf("\n=== 网络统计 ===\n")
    
    networkStats, err := fantomClient.GetNetworkStats()
    if err != nil {
        log.Printf("获取网络统计失败: %v", err)
    } else {
        fmt.Printf("平均出块时间: %s秒\n", networkStats.AverageBlockTime.String())
        fmt.Printf("当前TPS: %s\n", networkStats.TPS.String())
        fmt.Printf("验证者数量: %s\n", networkStats.ValidatorCount.String())
    }
    
    // 4. 查询验证者信息
    fmt.Printf("\n=== 验证者信息 ===\n")
    
    bestValidators, err := stakingService.GetBestValidators(5)
    if err != nil {
        log.Printf("获取最佳验证者失败: %v", err)
    } else {
        fmt.Printf("推荐的前5个验证者:\n")
        for i, ranking := range bestValidators {
            fmt.Printf("  %d. 验证者ID: %s\n", i+1, ranking.Validator.ID.String())
            fmt.Printf("     评分: %s\n", ranking.Score.String())
            fmt.Printf("     APR: %s%%\n", ranking.APR.String())
            fmt.Printf("     在线时间: %s%%\n", ranking.Uptime.String())
        }
    }
    
    // 5. 质押示例
    fmt.Printf("\n=== 质押示例 ===\n")
    
    if balance.Cmp(big.NewInt(10e18)) > 0 { // 如果余额大于10 FTM
        stakeAmount := big.NewInt(5e18) // 质押5 FTM
        validatorID := big.NewInt(1)    // 选择验证者1
        
        fmt.Printf("准备质押 %s FTM 到验证者 %s\n", 
            stakeAmount.String(), 
            validatorID.String())
        
        // 计算预期奖励
        rewards, err := stakingService.CalculateStakingRewards(
            stakeAmount,
            validatorID,
            365*24*time.Hour, // 1年
        )
        if err != nil {
            log.Printf("计算质押奖励失败: %v", err)
        } else {
            fmt.Printf("预期年化收益率: %s%%\n", rewards.APR.Mul(decimal.NewFromInt(100)).String())
            fmt.Printf("预期年收益: %s FTM\n", 
                decimal.NewFromBigInt(rewards.AnnualReward, -18).String())
        }
        
        // 执行质押
        tx, err := stakingService.Delegate(validatorID, stakeAmount)
        if err != nil {
            log.Printf("质押失败: %v", err)
        } else {
            fmt.Printf("质押交易已提交: %s\n", tx.Hash().Hex())
        }
    } else {
        fmt.Printf("余额不足，无法进行质押示例\n")
    }
    
    // 6. SpookySwap交易示例
    fmt.Printf("\n=== SpookySwap交易示例 ===\n")
    
    // 假设的代币地址 (USDC)
    usdcAddress := common.HexToAddress("0x04068DA6C83AFCFA0e13ba15A6696662335D5B75")
    
    if balance.Cmp(big.NewInt(2e18)) > 0 { // 如果余额大于2 FTM
        swapAmount := big.NewInt(1e18) // 用1 FTM换USDC
        minOutput := big.NewInt(0)     // 最小输出 (实际应该计算)
        deadline := big.NewInt(time.Now().Unix() + 1200) // 20分钟后过期
        
        fmt.Printf("准备用 %s FTM 换取 USDC\n", swapAmount.String())
        
        tx, err := spookySwapService.SwapFTMForTokens(
            usdcAddress,
            swapAmount,
            minOutput,
            deadline,
        )
        if err != nil {
            log.Printf("SpookySwap交换失败: %v", err)
        } else {
            fmt.Printf("SpookySwap交换交易已提交: %s\n", tx.Hash().Hex())
        }
    } else {
        fmt.Printf("余额不足，无法进行SpookySwap交易示例\n")
    }
    
    // 7. 查询委托信息
    fmt.Printf("\n=== 委托信息查询 ===\n")
    
    validatorID := big.NewInt(1)
    delegation, err := fantomClient.GetDelegation(userAddress, validatorID)
    if err != nil {
        log.Printf("查询委托信息失败: %v", err)
    } else {
        if delegation.Stake.Cmp(big.NewInt(0)) > 0 {
            fmt.Printf("验证者 %s 的委托信息:\n", validatorID.String())
            fmt.Printf("  质押数量: %s FTM\n", 
                decimal.NewFromBigInt(delegation.Stake, -18).String())
            fmt.Printf("  锁定数量: %s FTM\n", 
                decimal.NewFromBigInt(delegation.LockedStake, -18).String())
            
            if delegation.LockupEndTime.Cmp(big.NewInt(0)) > 0 {
                lockupEnd := time.Unix(delegation.LockupEndTime.Int64(), 0)
                fmt.Printf("  锁定到期: %s\n", lockupEnd.Format("2006-01-02 15:04:05"))
            }
        } else {
            fmt.Printf("未在验证者 %s 进行委托\n", validatorID.String())
        }
    }
    
    // 8. 性能对比
    fmt.Printf("\n=== 性能对比 ===\n")
    
    fmt.Printf("Fantom vs 其他区块链:\n")
    fmt.Printf("  出块时间: ~1秒 (vs 以太坊 ~13秒)\n")
    fmt.Printf("  交易费用: ~$0.01 (vs 以太坊 $5-50)\n")
    fmt.Printf("  TPS: ~4000 (vs 以太坊 ~15)\n")
    fmt.Printf("  最终确认: 1-2秒 (vs 以太坊 6分钟)\n")
    
    // 9. DeFi生态概览
    fmt.Printf("\n=== Fantom DeFi生态 ===\n")
    
    fmt.Printf("主要DeFi协议:\n")
    fmt.Printf("  SpookySwap: %s\n", SpookySwapRouterAddress.Hex())
    fmt.Printf("  SpiritSwap: %s\n", SpiritSwapRouterAddress.Hex())
    fmt.Printf("  Beethoven X: %s\n", BeethovenXVaultAddress.Hex())
    fmt.Printf("  Geist Finance: %s\n", GeistLendingPoolAddress.Hex())
    fmt.Printf("  Tarot Finance: %s\n", TarotFactoryAddress.Hex())
    
    fmt.Printf("\n生态特点:\n")
    fmt.Printf("  - 低费用高速交易\n")
    fmt.Printf("  - 丰富的DeFi协议\n")
    fmt.Printf("  - 活跃的开发社区\n")
    fmt.Printf("  - 与以太坊完全兼容\n")
    
    // 10. 风险提示
    fmt.Printf("\n=== 风险提示 ===\n")
    
    fmt.Printf("使用Fantom网络时请注意:\n")
    fmt.Printf("  1. 确保使用正确的网络配置\n")
    fmt.Printf("  2. 小额测试后再进行大额操作\n")
    fmt.Printf("  3. 关注网络升级和维护公告\n")
    fmt.Printf("  4. 验证合约地址的真实性\n")
    fmt.Printf("  5. 保管好私钥和助记词\n")
}
```

这个Fantom使用指南提供了完整的高性能区块链集成方案，涵盖了质押委托、DeFi协议集成、跨链操作、性能优化等核心功能，是构建快速低费用DeFi应用的重要参考文档。
