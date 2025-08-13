# SushiSwap 协议 详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [环境准备](#环境准备)
3. [交易功能](#交易功能)
4. [流动性挖矿](#流动性挖矿)
5. [收益农场](#收益农场)
6. [治理代币](#治理代币)
7. [跨链部署](#跨链部署)
8. [高级功能](#高级功能)
9. [实际应用](#实际应用)

## 基础概念

### 1.1 SushiSwap 简介

SushiSwap 是社区驱动的去中心化交易所，从 Uniswap 分叉而来，通过创新的代币经济学和治理机制，成为多链 DeFi 生态的重要组成部分。

```bash
# 安装SushiSwap相关依赖
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

// SushiSwap 核心合约地址 (Mainnet)
var (
    // SushiSwap Router V2
    SushiRouterAddress     = common.HexToAddress("0xd9e1cE17f2641f24aE83637ab66a2cca9C378B9F")
    
    // SushiSwap Factory
    SushiFactoryAddress    = common.HexToAddress("0xC0AEe478e3658e2610c5F7A4A2E1777cE9e4f2Ac")
    
    // SUSHI Token
    SUSHITokenAddress      = common.HexToAddress("0x6B3595068778DD592e39A122f4f5a5cF09C90fE2")
    
    // MasterChef (流动性挖矿)
    MasterChefAddress      = common.HexToAddress("0xc2EdaD668740f1aA35E4D8f227fB8E17dcA888Cd")
    
    // MasterChef V2
    MasterChefV2Address    = common.HexToAddress("0xEF0881eC094552b2e128Cf945EF17a6752B4Ec5d")
    
    // SushiBar (xSUSHI质押)
    SushiBarAddress        = common.HexToAddress("0x8798249c2E607446EfB7Ad49eC89dD1865Ff4272")
    
    // Kashi (借贷平台)
    KashiAddress           = common.HexToAddress("0x2cBA6Ab6574646Badc84F0544d05059e57a5dc42")
    
    // Miso (代币发行平台)
    MisoAddress            = common.HexToAddress("0x4c4564a1FE775D97297F9e3Dc2e762e0Ed5Dda0e")
)

// 交易对信息
type SushiPair struct {
    Address     common.Address
    Token0      common.Address
    Token1      common.Address
    Reserve0    *big.Int
    Reserve1    *big.Int
    TotalSupply *big.Int
    Fee         *big.Int // 0.3% = 3000
}

// 农场信息
type FarmInfo struct {
    PoolID          *big.Int
    LPToken         common.Address
    AllocPoint      *big.Int
    LastRewardBlock *big.Int
    AccSushiPerShare *big.Int
    TotalStaked     *big.Int
    APR             decimal.Decimal
}

// 用户农场信息
type UserFarmInfo struct {
    Amount      *big.Int
    RewardDebt  *big.Int
    PendingReward *big.Int
}
```

## 环境准备

### 2.1 合约ABI定义

```go
// contracts/sushi_abi.go
package contracts

// SushiSwap Router ABI (简化版)
const SushiRouterABI = `[
    {
        "inputs": [
            {"name": "amountIn", "type": "uint256"},
            {"name": "amountOutMin", "type": "uint256"},
            {"name": "path", "type": "address[]"},
            {"name": "to", "type": "address"},
            {"name": "deadline", "type": "uint256"}
        ],
        "name": "swapExactTokensForTokens",
        "outputs": [{"name": "amounts", "type": "uint256[]"}],
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
        "outputs": [{"name": "amounts", "type": "uint256[]"}],
        "type": "function"
    },
    {
        "inputs": [
            {"name": "tokenA", "type": "address"},
            {"name": "tokenB", "type": "address"},
            {"name": "amountADesired", "type": "uint256"},
            {"name": "amountBDesired", "type": "uint256"},
            {"name": "amountAMin", "type": "uint256"},
            {"name": "amountBMin", "type": "uint256"},
            {"name": "to", "type": "address"},
            {"name": "deadline", "type": "uint256"}
        ],
        "name": "addLiquidity",
        "outputs": [
            {"name": "amountA", "type": "uint256"},
            {"name": "amountB", "type": "uint256"},
            {"name": "liquidity", "type": "uint256"}
        ],
        "type": "function"
    },
    {
        "inputs": [
            {"name": "amountIn", "type": "uint256"},
            {"name": "path", "type": "address[]"}
        ],
        "name": "getAmountsOut",
        "outputs": [{"name": "amounts", "type": "uint256[]"}],
        "type": "function"
    }
]`

// MasterChef ABI (简化版)
const MasterChefABI = `[
    {
        "inputs": [
            {"name": "_pid", "type": "uint256"},
            {"name": "_amount", "type": "uint256"}
        ],
        "name": "deposit",
        "outputs": [],
        "type": "function"
    },
    {
        "inputs": [
            {"name": "_pid", "type": "uint256"},
            {"name": "_amount", "type": "uint256"}
        ],
        "name": "withdraw",
        "outputs": [],
        "type": "function"
    },
    {
        "inputs": [{"name": "_pid", "type": "uint256"}],
        "name": "emergencyWithdraw",
        "outputs": [],
        "type": "function"
    },
    {
        "inputs": [
            {"name": "_pid", "type": "uint256"},
            {"name": "_user", "type": "address"}
        ],
        "name": "pendingSushi",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    },
    {
        "inputs": [{"name": "_pid", "type": "uint256"}],
        "name": "poolInfo",
        "outputs": [
            {"name": "lpToken", "type": "address"},
            {"name": "allocPoint", "type": "uint256"},
            {"name": "lastRewardBlock", "type": "uint256"},
            {"name": "accSushiPerShare", "type": "uint256"}
        ],
        "type": "function"
    },
    {
        "inputs": [
            {"name": "_pid", "type": "uint256"},
            {"name": "_user", "type": "address"}
        ],
        "name": "userInfo",
        "outputs": [
            {"name": "amount", "type": "uint256"},
            {"name": "rewardDebt", "type": "uint256"}
        ],
        "type": "function"
    }
]`

// SushiBar ABI (简化版)
const SushiBarABI = `[
    {
        "inputs": [{"name": "_amount", "type": "uint256"}],
        "name": "enter",
        "outputs": [],
        "type": "function"
    },
    {
        "inputs": [{"name": "_share", "type": "uint256"}],
        "name": "leave",
        "outputs": [],
        "type": "function"
    },
    {
        "inputs": [{"name": "account", "type": "address"}],
        "name": "balanceOf",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    }
]`
```

### 2.2 客户端设置

```go
// client/sushi_client.go
package client

import (
    "context"
    "math/big"
    
    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
)

type SushiClient struct {
    ethClient       *ethclient.Client
    routerABI       abi.ABI
    masterChefABI   abi.ABI
    sushiBarABI     abi.ABI
    routerAddress   common.Address
    masterChefAddr  common.Address
    sushiBarAddr    common.Address
}

func NewSushiClient(rpcURL string) (*SushiClient, error) {
    ethClient, err := ethclient.Dial(rpcURL)
    if err != nil {
        return nil, err
    }
    
    routerABI, err := abi.JSON(strings.NewReader(SushiRouterABI))
    if err != nil {
        return nil, err
    }
    
    masterChefABI, err := abi.JSON(strings.NewReader(MasterChefABI))
    if err != nil {
        return nil, err
    }
    
    sushiBarABI, err := abi.JSON(strings.NewReader(SushiBarABI))
    if err != nil {
        return nil, err
    }
    
    return &SushiClient{
        ethClient:      ethClient,
        routerABI:      routerABI,
        masterChefABI:  masterChefABI,
        sushiBarABI:    sushiBarABI,
        routerAddress:  SushiRouterAddress,
        masterChefAddr: MasterChefAddress,
        sushiBarAddr:   SushiBarAddress,
    }, nil
}

// 获取交易输出金额
func (c *SushiClient) GetAmountsOut(amountIn *big.Int, path []common.Address) ([]*big.Int, error) {
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    var result []*big.Int
    err := c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &c.routerAddress,
        Data: c.routerABI.Methods["getAmountsOut"].ID,
    }, nil)
    
    if err != nil {
        return nil, err
    }
    
    return result, nil
}

// 获取农场信息
func (c *SushiClient) GetPoolInfo(poolID *big.Int) (*FarmInfo, error) {
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    var result []interface{}
    err := c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &c.masterChefAddr,
        Data: c.masterChefABI.Methods["poolInfo"].ID,
    }, nil)
    
    if err != nil {
        return nil, err
    }
    
    return &FarmInfo{
        PoolID:          poolID,
        LPToken:         result[0].(common.Address),
        AllocPoint:      result[1].(*big.Int),
        LastRewardBlock: result[2].(*big.Int),
        AccSushiPerShare: result[3].(*big.Int),
    }, nil
}
```

## 交易功能

### 3.1 交易服务

```go
// services/sushi_swap_service.go
package services

import (
    "context"
    "crypto/ecdsa"
    "math/big"
    "time"
    
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/crypto"
)

type SushiSwapService struct {
    client     *SushiClient
    privateKey *ecdsa.PrivateKey
}

func NewSushiSwapService(client *SushiClient, privateKey *ecdsa.PrivateKey) *SushiSwapService {
    return &SushiSwapService{
        client:     client,
        privateKey: privateKey,
    }
}

// 代币交换
func (s *SushiSwapService) SwapExactTokensForTokens(
    amountIn *big.Int,
    amountOutMin *big.Int,
    path []common.Address,
    to common.Address,
    deadline *big.Int,
) (*types.Transaction, error) {
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 授权Router使用输入代币
    if err := s.approveToken(path[0], s.client.routerAddress, amountIn); err != nil {
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
    
    // 构建交换交易数据
    data, err := s.client.routerABI.Pack(
        "swapExactTokensForTokens",
        amountIn,
        amountOutMin,
        path,
        to,
        deadline,
    )
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, s.client.routerAddress, big.NewInt(0), 200000, gasPrice, data)
    
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

// ETH换代币
func (s *SushiSwapService) SwapExactETHForTokens(
    amountOutMin *big.Int,
    path []common.Address,
    to common.Address,
    deadline *big.Int,
    ethAmount *big.Int,
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
    
    // 构建交换交易数据
    data, err := s.client.routerABI.Pack(
        "swapExactETHForTokens",
        amountOutMin,
        path,
        to,
        deadline,
    )
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, s.client.routerAddress, ethAmount, 200000, gasPrice, data)
    
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
func (s *SushiSwapService) AddLiquidity(
    tokenA, tokenB common.Address,
    amountADesired, amountBDesired *big.Int,
    amountAMin, amountBMin *big.Int,
    to common.Address,
    deadline *big.Int,
) (*types.Transaction, error) {
    // 授权Router使用两个代币
    if err := s.approveToken(tokenA, s.client.routerAddress, amountADesired); err != nil {
        return nil, err
    }
    if err := s.approveToken(tokenB, s.client.routerAddress, amountBDesired); err != nil {
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
    
    // 构建添加流动性交易数据
    data, err := s.client.routerABI.Pack(
        "addLiquidity",
        tokenA,
        tokenB,
        amountADesired,
        amountBDesired,
        amountAMin,
        amountBMin,
        to,
        deadline,
    )
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, s.client.routerAddress, big.NewInt(0), 300000, gasPrice, data)
    
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

// 计算交换输出
func (s *SushiSwapService) CalculateSwapOutput(
    amountIn *big.Int,
    path []common.Address,
) (*SwapQuote, error) {
    amounts, err := s.client.GetAmountsOut(amountIn, path)
    if err != nil {
        return nil, err
    }
    
    if len(amounts) < 2 {
        return nil, fmt.Errorf("无效的交换路径")
    }
    
    outputAmount := amounts[len(amounts)-1]
    
    // 计算价格影响
    priceImpact := s.calculatePriceImpact(amountIn, outputAmount, path)
    
    return &SwapQuote{
        InputAmount:  amountIn,
        OutputAmount: outputAmount,
        Path:         path,
        PriceImpact:  priceImpact,
        Fee:          s.calculateFee(amountIn),
    }, nil
}

type SwapQuote struct {
    InputAmount  *big.Int
    OutputAmount *big.Int
    Path         []common.Address
    PriceImpact  decimal.Decimal
    Fee          *big.Int
}

// 计算手续费 (0.3%)
func (s *SushiSwapService) calculateFee(amountIn *big.Int) *big.Int {
    fee := new(big.Int).Mul(amountIn, big.NewInt(3))
    fee.Div(fee, big.NewInt(1000))
    return fee
}

// 生成截止时间 (当前时间 + 20分钟)
func (s *SushiSwapService) GenerateDeadline() *big.Int {
    return big.NewInt(time.Now().Unix() + 1200)
}
```

## 流动性挖矿

### 4.1 MasterChef服务

```go
// services/masterchef_service.go
package services

import (
    "context"
    "crypto/ecdsa"
    "math/big"
    
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/shopspring/decimal"
)

type MasterChefService struct {
    client     *SushiClient
    privateKey *ecdsa.PrivateKey
}

func NewMasterChefService(client *SushiClient, privateKey *ecdsa.PrivateKey) *MasterChefService {
    return &MasterChefService{
        client:     client,
        privateKey: privateKey,
    }
}

// 存入LP代币到农场
func (s *MasterChefService) Deposit(poolID *big.Int, amount *big.Int) (*types.Transaction, error) {
    // 获取农场信息
    farmInfo, err := s.client.GetPoolInfo(poolID)
    if err != nil {
        return nil, err
    }
    
    // 授权MasterChef使用LP代币
    if err := s.approveToken(farmInfo.LPToken, s.client.masterChefAddr, amount); err != nil {
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
    data, err := s.client.masterChefABI.Pack("deposit", poolID, amount)
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, s.client.masterChefAddr, big.NewInt(0), 200000, gasPrice, data)
    
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

// 从农场提取LP代币
func (s *MasterChefService) Withdraw(poolID *big.Int, amount *big.Int) (*types.Transaction, error) {
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
    data, err := s.client.masterChefABI.Pack("withdraw", poolID, amount)
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, s.client.masterChefAddr, big.NewInt(0), 150000, gasPrice, data)
    
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

// 紧急提取 (不领取奖励)
func (s *MasterChefService) EmergencyWithdraw(poolID *big.Int) (*types.Transaction, error) {
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
    
    // 构建emergencyWithdraw交易数据
    data, err := s.client.masterChefABI.Pack("emergencyWithdraw", poolID)
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, s.client.masterChefAddr, big.NewInt(0), 100000, gasPrice, data)
    
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

// 获取待领取的SUSHI奖励
func (s *MasterChefService) GetPendingSushi(poolID *big.Int, user common.Address) (*big.Int, error) {
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    var pendingReward *big.Int
    err := s.client.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &s.client.masterChefAddr,
        Data: s.client.masterChefABI.Methods["pendingSushi"].ID,
    }, nil)
    
    if err != nil {
        return nil, err
    }
    
    return pendingReward, nil
}

// 获取用户农场信息
func (s *MasterChefService) GetUserInfo(poolID *big.Int, user common.Address) (*UserFarmInfo, error) {
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    var result []interface{}
    err := s.client.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &s.client.masterChefAddr,
        Data: s.client.masterChefABI.Methods["userInfo"].ID,
    }, nil)
    
    if err != nil {
        return nil, err
    }
    
    // 获取待领取奖励
    pendingReward, err := s.GetPendingSushi(poolID, user)
    if err != nil {
        pendingReward = big.NewInt(0)
    }
    
    return &UserFarmInfo{
        Amount:        result[0].(*big.Int),
        RewardDebt:    result[1].(*big.Int),
        PendingReward: pendingReward,
    }, nil
}

// 计算农场APR
func (s *MasterChefService) CalculateFarmAPR(poolID *big.Int) (decimal.Decimal, error) {
    farmInfo, err := s.client.GetPoolInfo(poolID)
    if err != nil {
        return decimal.Zero, err
    }
    
    // 获取SUSHI价格 (需要实现价格预言机)
    sushiPrice, err := s.getSushiPrice()
    if err != nil {
        return decimal.Zero, err
    }
    
    // 获取LP代币价格
    lpPrice, err := s.getLPTokenPrice(farmInfo.LPToken)
    if err != nil {
        return decimal.Zero, err
    }
    
    // 计算年化奖励
    // 简化计算: (每块SUSHI奖励 * 分配点数 * 块/年 * SUSHI价格) / (总质押 * LP价格)
    blocksPerYear := decimal.NewFromInt(2102400) // 假设15秒一个块
    sushiPerBlock := decimal.NewFromFloat(100)   // 假设每块100 SUSHI
    
    allocPoint := decimal.NewFromBigInt(farmInfo.AllocPoint, 0)
    totalAllocPoint := decimal.NewFromInt(1000) // 需要从合约获取
    
    yearlyReward := sushiPerBlock.Mul(blocksPerYear).Mul(allocPoint).Div(totalAllocPoint).Mul(sushiPrice)
    totalStakedValue := decimal.NewFromBigInt(farmInfo.TotalStaked, -18).Mul(lpPrice)
    
    if totalStakedValue.IsZero() {
        return decimal.Zero, nil
    }
    
    apr := yearlyReward.Div(totalStakedValue).Mul(decimal.NewFromInt(100))
    
    return apr, nil
}

// 获取SUSHI价格 (简化实现)
func (s *MasterChefService) getSushiPrice() (decimal.Decimal, error) {
    // 这里应该从价格预言机或DEX获取实际价格
    return decimal.NewFromFloat(1.5), nil // 假设价格
}

// 获取LP代币价格 (简化实现)
func (s *MasterChefService) getLPTokenPrice(lpToken common.Address) (decimal.Decimal, error) {
    // 这里应该计算LP代币的实际价格
    return decimal.NewFromFloat(100), nil // 假设价格
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
    // 创建SushiSwap客户端
    sushiClient, err := client.NewSushiClient("https://mainnet.infura.io/v3/YOUR_PROJECT_ID")
    if err != nil {
        log.Fatal("创建SushiSwap客户端失败:", err)
    }
    
    // 加载私钥
    privateKey, err := crypto.HexToECDSA("your_private_key_here")
    if err != nil {
        log.Fatal("加载私钥失败:", err)
    }
    
    // 获取用户地址
    userAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
    
    // 创建服务
    swapService := services.NewSushiSwapService(sushiClient, privateKey)
    masterChefService := services.NewMasterChefService(sushiClient, privateKey)
    
    // 代币地址
    wethAddress := common.HexToAddress("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")
    usdcAddress := common.HexToAddress("0xA0b86a33E6417c8f4c8c8c8c8c8c8c8c8c8c8c8c")
    
    // 1. 查询WETH到USDC的交换率
    swapAmount := big.NewInt(1e18) // 1 WETH
    path := []common.Address{wethAddress, usdcAddress}
    
    quote, err := swapService.CalculateSwapOutput(swapAmount, path)
    if err != nil {
        log.Fatal("计算交换失败:", err)
    }
    
    fmt.Printf("交换 1 WETH 可获得 %s USDC\n", quote.OutputAmount.String())
    fmt.Printf("价格影响: %s%%\n", quote.PriceImpact.String())
    fmt.Printf("手续费: %s\n", quote.Fee.String())
    
    // 2. 执行WETH到USDC的交换
    minOutput := new(big.Int).Mul(quote.OutputAmount, big.NewInt(99))
    minOutput.Div(minOutput, big.NewInt(100)) // 1% 滑点保护
    
    deadline := swapService.GenerateDeadline()
    
    tx, err := swapService.SwapExactTokensForTokens(
        swapAmount,
        minOutput,
        path,
        userAddress,
        deadline,
    )
    if err != nil {
        log.Fatal("交换失败:", err)
    }
    
    fmt.Printf("交换交易已提交: %s\n", tx.Hash().Hex())
    
    // 3. 添加WETH-USDC流动性
    wethAmount := big.NewInt(5e17)  // 0.5 WETH
    usdcAmount := big.NewInt(1000e6) // 1000 USDC
    
    // 设置最小金额 (5% 滑点)
    minWETH := new(big.Int).Mul(wethAmount, big.NewInt(95))
    minWETH.Div(minWETH, big.NewInt(100))
    
    minUSDC := new(big.Int).Mul(usdcAmount, big.NewInt(95))
    minUSDC.Div(minUSDC, big.NewInt(100))
    
    tx, err = swapService.AddLiquidity(
        wethAddress,
        usdcAddress,
        wethAmount,
        usdcAmount,
        minWETH,
        minUSDC,
        userAddress,
        deadline,
    )
    if err != nil {
        log.Fatal("添加流动性失败:", err)
    }
    
    fmt.Printf("添加流动性交易已提交: %s\n", tx.Hash().Hex())
    
    // 4. 将LP代币存入农场挖矿
    poolID := big.NewInt(1) // WETH-USDC池ID
    lpAmount := big.NewInt(1e18) // 假设获得了1个LP代币
    
    tx, err = masterChefService.Deposit(poolID, lpAmount)
    if err != nil {
        log.Fatal("存入农场失败:", err)
    }
    
    fmt.Printf("农场存入交易已提交: %s\n", tx.Hash().Hex())
    
    // 5. 查询农场信息
    farmInfo, err := sushiClient.GetPoolInfo(poolID)
    if err != nil {
        log.Fatal("查询农场信息失败:", err)
    }
    
    fmt.Printf("农场信息:\n")
    fmt.Printf("  LP代币: %s\n", farmInfo.LPToken.Hex())
    fmt.Printf("  分配点数: %s\n", farmInfo.AllocPoint.String())
    
    // 6. 查询用户农场信息
    userInfo, err := masterChefService.GetUserInfo(poolID, userAddress)
    if err != nil {
        log.Fatal("查询用户信息失败:", err)
    }
    
    fmt.Printf("用户农场信息:\n")
    fmt.Printf("  质押数量: %s\n", userInfo.Amount.String())
    fmt.Printf("  待领取奖励: %s SUSHI\n", userInfo.PendingReward.String())
    
    // 7. 计算农场APR
    apr, err := masterChefService.CalculateFarmAPR(poolID)
    if err != nil {
        log.Fatal("计算APR失败:", err)
    }
    
    fmt.Printf("农场年化收益率: %s%%\n", apr.String())
}
```

这个SushiSwap使用指南提供了完整的社区驱动DEX集成方案，涵盖了交易、流动性提供、收益农场等核心功能，是DeFi AMM协议的重要参考文档。
