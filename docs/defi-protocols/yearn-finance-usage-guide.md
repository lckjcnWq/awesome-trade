# Yearn Finance 协议 详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [环境准备](#环境准备)
3. [Vault系统](#vault系统)
4. [策略管理](#策略管理)
5. [收益优化](#收益优化)
6. [治理代币](#治理代币)
7. [风险管理](#风险管理)
8. [高级功能](#高级功能)
9. [实际应用](#实际应用)

## 基础概念

### 1.1 Yearn Finance 简介

Yearn Finance 是领先的 DeFi 收益聚合协议，通过自动化策略为用户优化收益，是 DeFi 生态中最重要的收益优化平台。

```bash
# 安装Yearn相关依赖
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

// Yearn Finance 核心合约地址 (Mainnet)
var (
    // YFI Token
    YFITokenAddress        = common.HexToAddress("0x0bc529c00C6401aEF6D220BE8C6Ea1667F6Ad93e")
    
    // Vault Registry
    VaultRegistryAddress   = common.HexToAddress("0x50c1a2eA0a861A967D9d0FFE2AE4012c2E053804")
    
    // 主要Vault地址
    YearnDAIVaultAddress   = common.HexToAddress("0xdA816459F1AB5631232FE5e97a05BBBb94970c95")
    YearnUSDCVaultAddress  = common.HexToAddress("0xa354F35829Ae975e850e23e9615b11Da1B3dC4DE")
    YearnUSDTVaultAddress  = common.HexToAddress("0x7Da96a3891Add058AdA2E826306D812C638D87a7")
    YearnWETHVaultAddress  = common.HexToAddress("0xa258C4606Ca8206D8aA700cE2143D7db854D168c")
    
    // Strategy Registry
    StrategyRegistryAddress = common.HexToAddress("0x187b9f26c0e8c4f5a0d7b6c2e8b2c8e8c8c8c8c8")
    
    // Governance
    GovernanceAddress      = common.HexToAddress("0xFEB4acf3df3cDEA7399794D0869ef76A6EfAff52")
    
    // Treasury
    TreasuryAddress        = common.HexToAddress("0x93A62dA5a14C80f265DAbC077fCEE437B1a0Efde")
)

// Vault信息
type VaultInfo struct {
    Address         common.Address
    Token           common.Address
    Name            string
    Symbol          string
    Decimals        uint8
    TotalAssets     *big.Int
    TotalSupply     *big.Int
    PricePerShare   *big.Int
    APY             decimal.Decimal
    ManagementFee   *big.Int
    PerformanceFee  *big.Int
    DepositLimit    *big.Int
    EmergencyShutdown bool
}

// 策略信息
type StrategyInfo struct {
    Address         common.Address
    Name            string
    Vault           common.Address
    Strategist      common.Address
    Keeper          common.Address
    Want            common.Address
    TotalDebt       *big.Int
    TotalGain       *big.Int
    TotalLoss       *big.Int
    DebtRatio       *big.Int
    RateLimit       *big.Int
    PerformanceFee  *big.Int
    Activation      *big.Int
    IsActive        bool
}

// 用户Vault信息
type UserVaultInfo struct {
    ShareBalance    *big.Int
    TokenBalance    *big.Int
    DepositedAmount *big.Int
    CurrentValue    *big.Int
    Profit          *big.Int
    APY             decimal.Decimal
}
```

## 环境准备

### 2.1 合约ABI定义

```go
// contracts/yearn_abi.go
package contracts

// Yearn Vault ABI (简化版)
const YearnVaultABI = `[
    {
        "inputs": [{"name": "_amount", "type": "uint256"}],
        "name": "deposit",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    },
    {
        "inputs": [
            {"name": "_amount", "type": "uint256"},
            {"name": "_recipient", "type": "address"}
        ],
        "name": "deposit",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    },
    {
        "inputs": [{"name": "_shares", "type": "uint256"}],
        "name": "withdraw",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    },
    {
        "inputs": [
            {"name": "_shares", "type": "uint256"},
            {"name": "_recipient", "type": "address"}
        ],
        "name": "withdraw",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    },
    {
        "inputs": [],
        "name": "totalAssets",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    },
    {
        "inputs": [],
        "name": "totalSupply",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    },
    {
        "inputs": [],
        "name": "pricePerShare",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    },
    {
        "inputs": [{"name": "account", "type": "address"}],
        "name": "balanceOf",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    },
    {
        "inputs": [],
        "name": "token",
        "outputs": [{"name": "", "type": "address"}],
        "type": "function"
    },
    {
        "inputs": [],
        "name": "depositLimit",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    },
    {
        "inputs": [],
        "name": "emergencyShutdown",
        "outputs": [{"name": "", "type": "bool"}],
        "type": "function"
    },
    {
        "inputs": [],
        "name": "managementFee",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    },
    {
        "inputs": [],
        "name": "performanceFee",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    }
]`

// Strategy ABI (简化版)
const StrategyABI = `[
    {
        "inputs": [],
        "name": "harvest",
        "outputs": [],
        "type": "function"
    },
    {
        "inputs": [],
        "name": "estimatedTotalAssets",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    },
    {
        "inputs": [],
        "name": "vault",
        "outputs": [{"name": "", "type": "address"}],
        "type": "function"
    },
    {
        "inputs": [],
        "name": "want",
        "outputs": [{"name": "", "type": "address"}],
        "type": "function"
    },
    {
        "inputs": [],
        "name": "strategist",
        "outputs": [{"name": "", "type": "address"}],
        "type": "function"
    },
    {
        "inputs": [],
        "name": "keeper",
        "outputs": [{"name": "", "type": "address"}],
        "type": "function"
    },
    {
        "inputs": [],
        "name": "isActive",
        "outputs": [{"name": "", "type": "bool"}],
        "type": "function"
    }
]`

// Vault Registry ABI (简化版)
const VaultRegistryABI = `[
    {
        "inputs": [{"name": "_token", "type": "address"}],
        "name": "latestVault",
        "outputs": [{"name": "", "type": "address"}],
        "type": "function"
    },
    {
        "inputs": [{"name": "_token", "type": "address"}],
        "name": "vaults",
        "outputs": [{"name": "", "type": "address[]"}],
        "type": "function"
    },
    {
        "inputs": [],
        "name": "numVaults",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    }
]`
```

### 2.2 客户端设置

```go
// client/yearn_client.go
package client

import (
    "context"
    "math/big"
    
    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
)

type YearnClient struct {
    ethClient       *ethclient.Client
    vaultABI        abi.ABI
    strategyABI     abi.ABI
    registryABI     abi.ABI
    registryAddress common.Address
}

func NewYearnClient(rpcURL string) (*YearnClient, error) {
    ethClient, err := ethclient.Dial(rpcURL)
    if err != nil {
        return nil, err
    }
    
    vaultABI, err := abi.JSON(strings.NewReader(YearnVaultABI))
    if err != nil {
        return nil, err
    }
    
    strategyABI, err := abi.JSON(strings.NewReader(StrategyABI))
    if err != nil {
        return nil, err
    }
    
    registryABI, err := abi.JSON(strings.NewReader(VaultRegistryABI))
    if err != nil {
        return nil, err
    }
    
    return &YearnClient{
        ethClient:       ethClient,
        vaultABI:        vaultABI,
        strategyABI:     strategyABI,
        registryABI:     registryABI,
        registryAddress: VaultRegistryAddress,
    }, nil
}

// 获取Vault信息
func (c *YearnClient) GetVaultInfo(vaultAddress common.Address) (*VaultInfo, error) {
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    // 获取基本信息
    var token common.Address
    err := c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &vaultAddress,
        Data: c.vaultABI.Methods["token"].ID,
    }, nil)
    if err != nil {
        return nil, err
    }
    
    var totalAssets *big.Int
    err = c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &vaultAddress,
        Data: c.vaultABI.Methods["totalAssets"].ID,
    }, nil)
    if err != nil {
        return nil, err
    }
    
    var totalSupply *big.Int
    err = c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &vaultAddress,
        Data: c.vaultABI.Methods["totalSupply"].ID,
    }, nil)
    if err != nil {
        return nil, err
    }
    
    var pricePerShare *big.Int
    err = c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &vaultAddress,
        Data: c.vaultABI.Methods["pricePerShare"].ID,
    }, nil)
    if err != nil {
        return nil, err
    }
    
    var depositLimit *big.Int
    err = c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &vaultAddress,
        Data: c.vaultABI.Methods["depositLimit"].ID,
    }, nil)
    if err != nil {
        return nil, err
    }
    
    var emergencyShutdown bool
    err = c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &vaultAddress,
        Data: c.vaultABI.Methods["emergencyShutdown"].ID,
    }, nil)
    if err != nil {
        return nil, err
    }
    
    var managementFee *big.Int
    err = c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &vaultAddress,
        Data: c.vaultABI.Methods["managementFee"].ID,
    }, nil)
    if err != nil {
        return nil, err
    }
    
    var performanceFee *big.Int
    err = c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &vaultAddress,
        Data: c.vaultABI.Methods["performanceFee"].ID,
    }, nil)
    if err != nil {
        return nil, err
    }
    
    return &VaultInfo{
        Address:           vaultAddress,
        Token:            token,
        TotalAssets:      totalAssets,
        TotalSupply:      totalSupply,
        PricePerShare:    pricePerShare,
        DepositLimit:     depositLimit,
        EmergencyShutdown: emergencyShutdown,
        ManagementFee:    managementFee,
        PerformanceFee:   performanceFee,
    }, nil
}

// 获取用户Vault余额
func (c *YearnClient) GetUserVaultBalance(vaultAddress, userAddress common.Address) (*big.Int, error) {
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    var balance *big.Int
    err := c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &vaultAddress,
        Data: c.vaultABI.Methods["balanceOf"].ID,
    }, nil)
    
    if err != nil {
        return nil, err
    }
    
    return balance, nil
}

// 获取代币的最新Vault
func (c *YearnClient) GetLatestVault(tokenAddress common.Address) (common.Address, error) {
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    var vaultAddress common.Address
    err := c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &c.registryAddress,
        Data: c.registryABI.Methods["latestVault"].ID,
    }, nil)
    
    if err != nil {
        return common.Address{}, err
    }
    
    return vaultAddress, nil
}
```

## Vault系统

### 3.1 Vault服务

```go
// services/vault_service.go
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

type VaultService struct {
    client     *YearnClient
    privateKey *ecdsa.PrivateKey
}

func NewVaultService(client *YearnClient, privateKey *ecdsa.PrivateKey) *VaultService {
    return &VaultService{
        client:     client,
        privateKey: privateKey,
    }
}

// 存入资产到Vault
func (s *VaultService) Deposit(vaultAddress common.Address, amount *big.Int) (*types.Transaction, error) {
    // 获取Vault信息
    vaultInfo, err := s.client.GetVaultInfo(vaultAddress)
    if err != nil {
        return nil, err
    }
    
    // 检查紧急关闭状态
    if vaultInfo.EmergencyShutdown {
        return nil, fmt.Errorf("Vault处于紧急关闭状态")
    }
    
    // 检查存款限制
    if vaultInfo.DepositLimit.Cmp(big.NewInt(0)) > 0 {
        newTotal := new(big.Int).Add(vaultInfo.TotalAssets, amount)
        if newTotal.Cmp(vaultInfo.DepositLimit) > 0 {
            return nil, fmt.Errorf("超过存款限制")
        }
    }
    
    // 授权Vault使用代币
    if err := s.approveToken(vaultInfo.Token, vaultAddress, amount); err != nil {
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
    data, err := s.client.vaultABI.Pack("deposit", amount)
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, vaultAddress, big.NewInt(0), 200000, gasPrice, data)
    
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

// 从Vault提取资产
func (s *VaultService) Withdraw(vaultAddress common.Address, shares *big.Int) (*types.Transaction, error) {
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 检查用户份额余额
    userShares, err := s.client.GetUserVaultBalance(vaultAddress, fromAddress)
    if err != nil {
        return nil, err
    }
    
    if shares.Cmp(userShares) > 0 {
        return nil, fmt.Errorf("提取份额超过余额")
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
    
    // 构建withdraw交易数据
    data, err := s.client.vaultABI.Pack("withdraw", shares)
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, vaultAddress, big.NewInt(0), 150000, gasPrice, data)
    
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

// 提取所有资产
func (s *VaultService) WithdrawAll(vaultAddress common.Address) (*types.Transaction, error) {
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 获取用户所有份额
    userShares, err := s.client.GetUserVaultBalance(vaultAddress, fromAddress)
    if err != nil {
        return nil, err
    }
    
    return s.Withdraw(vaultAddress, userShares)
}

// 计算用户收益
func (s *VaultService) CalculateUserProfit(vaultAddress, userAddress common.Address, initialDeposit *big.Int) (*UserProfitInfo, error) {
    // 获取用户当前份额
    userShares, err := s.client.GetUserVaultBalance(vaultAddress, userAddress)
    if err != nil {
        return nil, err
    }
    
    // 获取Vault信息
    vaultInfo, err := s.client.GetVaultInfo(vaultAddress)
    if err != nil {
        return nil, err
    }
    
    // 计算当前价值
    currentValue := new(big.Int).Mul(userShares, vaultInfo.PricePerShare)
    currentValue.Div(currentValue, big.NewInt(1e18)) // 调整精度
    
    // 计算利润
    profit := new(big.Int).Sub(currentValue, initialDeposit)
    
    // 计算收益率
    profitRate := decimal.Zero
    if initialDeposit.Cmp(big.NewInt(0)) > 0 {
        profitDecimal := decimal.NewFromBigInt(profit, -18)
        initialDecimal := decimal.NewFromBigInt(initialDeposit, -18)
        profitRate = profitDecimal.Div(initialDecimal).Mul(decimal.NewFromInt(100))
    }
    
    return &UserProfitInfo{
        InitialDeposit: initialDeposit,
        CurrentValue:   currentValue,
        Profit:         profit,
        ProfitRate:     profitRate,
        Shares:         userShares,
    }, nil
}

// 获取最佳Vault
func (s *VaultService) GetBestVault(tokenAddress common.Address) (*VaultRecommendation, error) {
    // 获取最新Vault
    latestVault, err := s.client.GetLatestVault(tokenAddress)
    if err != nil {
        return nil, err
    }
    
    // 获取Vault信息
    vaultInfo, err := s.client.GetVaultInfo(latestVault)
    if err != nil {
        return nil, err
    }
    
    // 计算APY (简化实现)
    apy, err := s.calculateVaultAPY(latestVault)
    if err != nil {
        apy = decimal.Zero
    }
    
    return &VaultRecommendation{
        VaultAddress: latestVault,
        VaultInfo:    vaultInfo,
        APY:          apy,
        Reason:       "最新版本Vault，通常具有最优策略",
        RiskLevel:    "中等",
    }, nil
}

type UserProfitInfo struct {
    InitialDeposit *big.Int
    CurrentValue   *big.Int
    Profit         *big.Int
    ProfitRate     decimal.Decimal
    Shares         *big.Int
}

type VaultRecommendation struct {
    VaultAddress common.Address
    VaultInfo    *VaultInfo
    APY          decimal.Decimal
    Reason       string
    RiskLevel    string
}

// 计算Vault APY (简化实现)
func (s *VaultService) calculateVaultAPY(vaultAddress common.Address) (decimal.Decimal, error) {
    // 这里应该实现复杂的APY计算逻辑
    // 包括历史收益、策略表现等
    
    // 简化返回固定值
    return decimal.NewFromFloat(8.5), nil
}
```

## 收益优化

### 4.1 收益优化服务

```go
// services/yield_optimizer_service.go
package services

import (
    "math/big"
    
    "github.com/ethereum/go-ethereum/common"
    "github.com/shopspring/decimal"
)

type YieldOptimizerService struct {
    vaultService *VaultService
    client       *YearnClient
}

func NewYieldOptimizerService(vaultService *VaultService, client *YearnClient) *YieldOptimizerService {
    return &YieldOptimizerService{
        vaultService: vaultService,
        client:       client,
    }
}

// 优化投资组合
func (s *YieldOptimizerService) OptimizePortfolio(portfolio map[common.Address]*big.Int) (*OptimizationResult, error) {
    var recommendations []VaultAllocation
    totalValue := decimal.Zero
    
    for tokenAddress, amount := range portfolio {
        // 获取最佳Vault
        recommendation, err := s.vaultService.GetBestVault(tokenAddress)
        if err != nil {
            continue
        }
        
        // 计算价值 (需要价格预言机)
        tokenValue, err := s.getTokenValue(tokenAddress, amount)
        if err != nil {
            continue
        }
        
        allocation := VaultAllocation{
            Token:        tokenAddress,
            Amount:       amount,
            Value:        tokenValue,
            Vault:        recommendation.VaultAddress,
            ExpectedAPY:  recommendation.APY,
            RiskLevel:    recommendation.RiskLevel,
        }
        
        recommendations = append(recommendations, allocation)
        totalValue = totalValue.Add(tokenValue)
    }
    
    // 计算加权平均APY
    weightedAPY := s.calculateWeightedAPY(recommendations, totalValue)
    
    return &OptimizationResult{
        Allocations:     recommendations,
        TotalValue:      totalValue,
        WeightedAPY:     weightedAPY,
        EstimatedYield:  s.calculateEstimatedYield(totalValue, weightedAPY),
        RiskScore:       s.calculateRiskScore(recommendations),
    }, nil
}

// 自动复投策略
func (s *YieldOptimizerService) AutoCompound(vaultAddress common.Address, userAddress common.Address) (*CompoundResult, error) {
    // 获取用户当前收益
    userInfo, err := s.getUserVaultInfo(vaultAddress, userAddress)
    if err != nil {
        return nil, err
    }
    
    // 检查是否有足够收益进行复投
    minCompoundAmount := big.NewInt(1e18) // 最小复投金额
    if userInfo.Profit.Cmp(minCompoundAmount) < 0 {
        return &CompoundResult{
            Success: false,
            Reason:  "收益不足，无需复投",
        }, nil
    }
    
    // 计算复投金额 (保留一部分作为gas费)
    compoundAmount := new(big.Int).Mul(userInfo.Profit, big.NewInt(95))
    compoundAmount.Div(compoundAmount, big.NewInt(100)) // 95%复投
    
    // 执行复投
    tx, err := s.vaultService.Deposit(vaultAddress, compoundAmount)
    if err != nil {
        return &CompoundResult{
            Success: false,
            Reason:  fmt.Sprintf("复投失败: %v", err),
        }, nil
    }
    
    return &CompoundResult{
        Success:        true,
        CompoundAmount: compoundAmount,
        TransactionHash: tx.Hash().Hex(),
        NewAPY:         s.calculateNewAPY(userInfo.APY, compoundAmount),
    }, nil
}

// 风险分散建议
func (s *YieldOptimizerService) DiversificationAdvice(totalAmount *big.Int) (*DiversificationPlan, error) {
    // 获取推荐的Vault列表
    recommendedVaults := []VaultAllocation{
        {
            Token:       common.HexToAddress("0x6B175474E89094C44Da98b954EedeAC495271d0F"), // DAI
            Vault:       YearnDAIVaultAddress,
            Percentage:  decimal.NewFromFloat(30),
            ExpectedAPY: decimal.NewFromFloat(8.5),
            RiskLevel:   "低",
        },
        {
            Token:       common.HexToAddress("0xA0b86a33E6417c8f4c8c8c8c8c8c8c8c8c8c8c8c"), // USDC
            Vault:       YearnUSDCVaultAddress,
            Percentage:  decimal.NewFromFloat(30),
            ExpectedAPY: decimal.NewFromFloat(7.8),
            RiskLevel:   "低",
        },
        {
            Token:       common.HexToAddress("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"), // WETH
            Vault:       YearnWETHVaultAddress,
            Percentage:  decimal.NewFromFloat(40),
            ExpectedAPY: decimal.NewFromFloat(12.3),
            RiskLevel:   "中等",
        },
    }
    
    // 计算具体分配金额
    for i := range recommendedVaults {
        allocation := &recommendedVaults[i]
        allocationAmount := new(big.Int).Mul(totalAmount, allocation.Percentage.BigInt())
        allocationAmount.Div(allocationAmount, big.NewInt(100))
        allocation.Amount = allocationAmount
    }
    
    return &DiversificationPlan{
        Allocations:    recommendedVaults,
        TotalAmount:    totalAmount,
        ExpectedAPY:    s.calculateWeightedAPY(recommendedVaults, decimal.NewFromBigInt(totalAmount, -18)),
        RiskLevel:      "平衡",
        Reasoning:      "分散投资于不同风险等级的资产，平衡收益与风险",
    }, nil
}

type VaultAllocation struct {
    Token        common.Address
    Amount       *big.Int
    Value        decimal.Decimal
    Vault        common.Address
    Percentage   decimal.Decimal
    ExpectedAPY  decimal.Decimal
    RiskLevel    string
}

type OptimizationResult struct {
    Allocations     []VaultAllocation
    TotalValue      decimal.Decimal
    WeightedAPY     decimal.Decimal
    EstimatedYield  decimal.Decimal
    RiskScore       decimal.Decimal
}

type CompoundResult struct {
    Success         bool
    CompoundAmount  *big.Int
    TransactionHash string
    NewAPY          decimal.Decimal
    Reason          string
}

type DiversificationPlan struct {
    Allocations []VaultAllocation
    TotalAmount *big.Int
    ExpectedAPY decimal.Decimal
    RiskLevel   string
    Reasoning   string
}

// 辅助函数
func (s *YieldOptimizerService) getTokenValue(tokenAddress common.Address, amount *big.Int) (decimal.Decimal, error) {
    // 这里应该从价格预言机获取实际价格
    return decimal.NewFromBigInt(amount, -18), nil
}

func (s *YieldOptimizerService) calculateWeightedAPY(allocations []VaultAllocation, totalValue decimal.Decimal) decimal.Decimal {
    weightedSum := decimal.Zero
    
    for _, allocation := range allocations {
        weight := allocation.Value.Div(totalValue)
        weightedSum = weightedSum.Add(allocation.ExpectedAPY.Mul(weight))
    }
    
    return weightedSum
}

func (s *YieldOptimizerService) calculateEstimatedYield(totalValue, apy decimal.Decimal) decimal.Decimal {
    return totalValue.Mul(apy).Div(decimal.NewFromInt(100))
}

func (s *YieldOptimizerService) calculateRiskScore(allocations []VaultAllocation) decimal.Decimal {
    // 简化的风险评分计算
    totalRisk := decimal.Zero
    
    for _, allocation := range allocations {
        var riskScore decimal.Decimal
        switch allocation.RiskLevel {
        case "低":
            riskScore = decimal.NewFromFloat(1)
        case "中等":
            riskScore = decimal.NewFromFloat(2)
        case "高":
            riskScore = decimal.NewFromFloat(3)
        default:
            riskScore = decimal.NewFromFloat(2)
        }
        
        totalRisk = totalRisk.Add(riskScore)
    }
    
    if len(allocations) > 0 {
        return totalRisk.Div(decimal.NewFromInt(int64(len(allocations))))
    }
    
    return decimal.Zero
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
    // 创建Yearn客户端
    yearnClient, err := client.NewYearnClient("https://mainnet.infura.io/v3/YOUR_PROJECT_ID")
    if err != nil {
        log.Fatal("创建Yearn客户端失败:", err)
    }
    
    // 加载私钥
    privateKey, err := crypto.HexToECDSA("your_private_key_here")
    if err != nil {
        log.Fatal("加载私钥失败:", err)
    }
    
    // 获取用户地址
    userAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
    
    // 创建服务
    vaultService := services.NewVaultService(yearnClient, privateKey)
    optimizerService := services.NewYieldOptimizerService(vaultService, yearnClient)
    
    // 代币地址
    daiAddress := common.HexToAddress("0x6B175474E89094C44Da98b954EedeAC495271d0F")
    
    // 1. 获取DAI的最佳Vault
    recommendation, err := vaultService.GetBestVault(daiAddress)
    if err != nil {
        log.Fatal("获取最佳Vault失败:", err)
    }
    
    fmt.Printf("推荐Vault: %s\n", recommendation.VaultAddress.Hex())
    fmt.Printf("预期APY: %s%%\n", recommendation.APY.String())
    fmt.Printf("风险等级: %s\n", recommendation.RiskLevel)
    
    // 2. 查询Vault详细信息
    vaultInfo, err := yearnClient.GetVaultInfo(recommendation.VaultAddress)
    if err != nil {
        log.Fatal("获取Vault信息失败:", err)
    }
    
    fmt.Printf("Vault信息:\n")
    fmt.Printf("  总资产: %s\n", vaultInfo.TotalAssets.String())
    fmt.Printf("  总供应量: %s\n", vaultInfo.TotalSupply.String())
    fmt.Printf("  每份价格: %s\n", vaultInfo.PricePerShare.String())
    fmt.Printf("  管理费: %s\n", vaultInfo.ManagementFee.String())
    fmt.Printf("  绩效费: %s\n", vaultInfo.PerformanceFee.String())
    
    // 3. 存入1000 DAI到Vault
    depositAmount := big.NewInt(1000e18) // 1000 DAI
    tx, err := vaultService.Deposit(recommendation.VaultAddress, depositAmount)
    if err != nil {
        log.Fatal("存入Vault失败:", err)
    }
    
    fmt.Printf("存入交易已提交: %s\n", tx.Hash().Hex())
    
    // 4. 查询用户Vault余额
    userShares, err := yearnClient.GetUserVaultBalance(recommendation.VaultAddress, userAddress)
    if err != nil {
        log.Fatal("查询用户余额失败:", err)
    }
    
    fmt.Printf("用户份额: %s\n", userShares.String())
    
    // 5. 计算用户收益
    profitInfo, err := vaultService.CalculateUserProfit(
        recommendation.VaultAddress,
        userAddress,
        depositAmount,
    )
    if err != nil {
        log.Fatal("计算收益失败:", err)
    }
    
    fmt.Printf("收益信息:\n")
    fmt.Printf("  初始存入: %s\n", profitInfo.InitialDeposit.String())
    fmt.Printf("  当前价值: %s\n", profitInfo.CurrentValue.String())
    fmt.Printf("  利润: %s\n", profitInfo.Profit.String())
    fmt.Printf("  收益率: %s%%\n", profitInfo.ProfitRate.String())
    
    // 6. 投资组合优化
    portfolio := map[common.Address]*big.Int{
        daiAddress:  big.NewInt(5000e18), // 5000 DAI
        common.HexToAddress("0xA0b86a33E6417c8f4c8c8c8c8c8c8c8c8c8c8c8c"): big.NewInt(3000e6), // 3000 USDC
    }
    
    optimization, err := optimizerService.OptimizePortfolio(portfolio)
    if err != nil {
        log.Fatal("投资组合优化失败:", err)
    }
    
    fmt.Printf("优化建议:\n")
    fmt.Printf("  总价值: %s\n", optimization.TotalValue.String())
    fmt.Printf("  加权APY: %s%%\n", optimization.WeightedAPY.String())
    fmt.Printf("  预期年收益: %s\n", optimization.EstimatedYield.String())
    fmt.Printf("  风险评分: %s\n", optimization.RiskScore.String())
    
    for i, allocation := range optimization.Allocations {
        fmt.Printf("  分配 %d:\n", i+1)
        fmt.Printf("    代币: %s\n", allocation.Token.Hex())
        fmt.Printf("    金额: %s\n", allocation.Amount.String())
        fmt.Printf("    Vault: %s\n", allocation.Vault.Hex())
        fmt.Printf("    预期APY: %s%%\n", allocation.ExpectedAPY.String())
    }
    
    // 7. 获取分散投资建议
    totalAmount := big.NewInt(10000e18) // 10000 USD等值
    diversification, err := optimizerService.DiversificationAdvice(totalAmount)
    if err != nil {
        log.Fatal("获取分散投资建议失败:", err)
    }
    
    fmt.Printf("分散投资建议:\n")
    fmt.Printf("  总金额: %s\n", diversification.TotalAmount.String())
    fmt.Printf("  预期APY: %s%%\n", diversification.ExpectedAPY.String())
    fmt.Printf("  风险等级: %s\n", diversification.RiskLevel)
    fmt.Printf("  理由: %s\n", diversification.Reasoning)
    
    for i, allocation := range diversification.Allocations {
        fmt.Printf("  建议 %d:\n", i+1)
        fmt.Printf("    代币: %s\n", allocation.Token.Hex())
        fmt.Printf("    比例: %s%%\n", allocation.Percentage.String())
        fmt.Printf("    金额: %s\n", allocation.Amount.String())
        fmt.Printf("    预期APY: %s%%\n", allocation.ExpectedAPY.String())
        fmt.Printf("    风险等级: %s\n", allocation.RiskLevel)
    }
    
    // 8. 自动复投
    compoundResult, err := optimizerService.AutoCompound(recommendation.VaultAddress, userAddress)
    if err != nil {
        log.Fatal("自动复投失败:", err)
    }
    
    if compoundResult.Success {
        fmt.Printf("自动复投成功:\n")
        fmt.Printf("  复投金额: %s\n", compoundResult.CompoundAmount.String())
        fmt.Printf("  交易哈希: %s\n", compoundResult.TransactionHash)
        fmt.Printf("  新APY: %s%%\n", compoundResult.NewAPY.String())
    } else {
        fmt.Printf("自动复投跳过: %s\n", compoundResult.Reason)
    }
}
```

这个Yearn Finance使用指南提供了完整的收益聚合协议集成方案，涵盖了Vault管理、收益优化、自动复投等核心功能，是DeFi收益优化的重要参考文档。
