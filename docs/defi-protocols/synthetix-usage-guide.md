# Synthetix 协议 详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [环境准备](#环境准备)
3. [合成资产系统](#合成资产系统)
4. [质押和铸造](#质押和铸造)
5. [交易和兑换](#交易和兑换)
6. [债务池机制](#债务池机制)
7. [奖励系统](#奖励系统)
8. [清算机制](#清算机制)
9. [实际应用](#实际应用)

## 基础概念

### 1.1 Synthetix 简介

Synthetix 是去中心化合成资产协议，允许用户通过质押SNX代币铸造合成资产(Synths)，提供对传统金融资产的链上敞口。

```bash
# 安装Synthetix相关依赖
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

// Synthetix 核心合约地址 (Mainnet)
var (
    // SNX Token
    SNXTokenAddress        = common.HexToAddress("0xC011a73ee8576Fb46F5E1c5751cA3B9Fe0af2a6F")
    
    // Synthetix Core
    SynthetixAddress       = common.HexToAddress("0xC011a72400E58ecD99Ee497CF89E3775d4bd732F")
    
    // sUSD Token
    SUSDTokenAddress       = common.HexToAddress("0x57Ab1ec28D129707052df4dF418D58a2D46d5f51")
    
    // Exchanger
    ExchangerAddress       = common.HexToAddress("0xAc86855865CbF31c8f9FBB68C749AD5Bd72802e3")
    
    // Fee Pool
    FeePoolAddress         = common.HexToAddress("0xb440DD674e1243644791a4AdfE3A2AbB0A92d309")
    
    // Rewards Escrow
    RewardsEscrowAddress   = common.HexToAddress("0xb671F2210B1F6621A2607EA63E6B2DC3e2464d1F")
    
    // Liquidations
    LiquidationsAddress    = common.HexToAddress("0xA8E31E3C38aDD6052A9407298FAEB8fD393A6cF9")
    
    // System Settings
    SystemSettingsAddress  = common.HexToAddress("0x545973f28950f50fc6c7F52AAb4Ad214A27C0564")
    
    // Depot (sUSD/ETH)
    DepotAddress          = common.HexToAddress("0xE1f64079aDa6Ef07b03982Ca34f1dD7152AA3b86")
)

// 合成资产类型
var (
    // 货币类合成资产
    SUSD = [4]byte{0x73, 0x55, 0x53, 0x44} // sUSD
    SEUR = [4]byte{0x73, 0x45, 0x55, 0x52} // sEUR
    SJPY = [4]byte{0x73, 0x4A, 0x50, 0x59} // sJPY
    SGBP = [4]byte{0x73, 0x47, 0x42, 0x50} // sGBP
    SCHF = [4]byte{0x73, 0x43, 0x48, 0x46} // sCHF
    
    // 加密货币合成资产
    SBTC = [4]byte{0x73, 0x42, 0x54, 0x43} // sBTC
    SETH = [4]byte{0x73, 0x45, 0x54, 0x48} // sETH
    SLINK = [4]byte{0x73, 0x4C, 0x49, 0x4E} // sLINK
    
    // 商品合成资产
    SGOLD = [4]byte{0x73, 0x47, 0x4F, 0x4C} // sGOLD
    SSILVER = [4]byte{0x73, 0x53, 0x49, 0x4C} // sSILVER
    SOIL = [4]byte{0x73, 0x4F, 0x49, 0x4C} // sOIL
    
    // 股票指数合成资产
    SFTSE = [4]byte{0x73, 0x46, 0x54, 0x53} // sFTSE
    SNIKKEI = [4]byte{0x73, 0x4E, 0x49, 0x4B} // sNIKKEI
)

// 质押信息
type StakingInfo struct {
    CollateralAmount    *big.Int        // 质押的SNX数量
    DebtBalance        *big.Int        // 债务余额
    CollateralisationRatio decimal.Decimal // 抵押率
    MaxIssuableSynths  *big.Int        // 最大可铸造Synths
    TransferableSNX    *big.Int        // 可转移的SNX
    EscrowedSNX        *big.Int        // 托管的SNX
    RewardEscrowBalance *big.Int       // 奖励托管余额
}

// 合成资产信息
type SynthInfo struct {
    CurrencyKey    [4]byte
    Name          string
    Symbol        string
    TotalSupply   *big.Int
    Address       common.Address
    Rate          *big.Int  // 汇率 (18位精度)
    InversePricing bool     // 是否为反向定价
    UpperLimit    *big.Int  // 上限价格
    LowerLimit    *big.Int  // 下限价格
}

// 交易信息
type ExchangeEntry struct {
    Src            [4]byte
    Amount         *big.Int
    Dest           [4]byte
    AmountReceived *big.Int
    ExchangeFee    *big.Int
    Timestamp      *big.Int
    RoundIdForSrc  *big.Int
    RoundIdForDest *big.Int
}

// 费用信息
type FeeInfo struct {
    ExchangeFeeRate    decimal.Decimal // 交易费率
    TargetThreshold    decimal.Decimal // 目标阈值
    SpeedThreshold     decimal.Decimal // 速度阈值
    LastExchangeRate   decimal.Decimal // 上次交易费率
    LastVolumeIssuance *big.Int        // 上次发行量
}

// 清算信息
type LiquidationInfo struct {
    Account           common.Address
    Deadline          *big.Int
    Caller            common.Address
    CollateralAmount  *big.Int
    DebtAmount        *big.Int
    EscrowAmount      *big.Int
}
```

## 环境准备

### 2.1 合约ABI定义

```go
// contracts/synthetix_abi.go
package contracts

// Synthetix Core ABI (简化版)
const SynthetixABI = `[
    {
        "inputs": [{"name": "amount", "type": "uint256"}],
        "name": "issueSynths",
        "outputs": [],
        "type": "function"
    },
    {
        "inputs": [{"name": "amount", "type": "uint256"}],
        "name": "burnSynths",
        "outputs": [],
        "type": "function"
    },
    {
        "inputs": [],
        "name": "burnSynthsToTarget",
        "outputs": [],
        "type": "function"
    },
    {
        "inputs": [{"name": "account", "type": "address"}],
        "name": "collateralisationRatio",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    },
    {
        "inputs": [{"name": "account", "type": "address"}],
        "name": "debtBalanceOf",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    },
    {
        "inputs": [{"name": "account", "type": "address"}],
        "name": "maxIssuableSynths",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    },
    {
        "inputs": [{"name": "account", "type": "address"}],
        "name": "remainingIssuableSynths",
        "outputs": [
            {"name": "maxIssuable", "type": "uint256"},
            {"name": "alreadyIssued", "type": "uint256"},
            {"name": "totalSystemDebt", "type": "uint256"}
        ],
        "type": "function"
    },
    {
        "inputs": [{"name": "account", "type": "address"}],
        "name": "transferableSynthetix",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    }
]`

// Exchanger ABI (简化版)
const ExchangerABI = `[
    {
        "inputs": [
            {"name": "sourceCurrencyKey", "type": "bytes32"},
            {"name": "sourceAmount", "type": "uint256"},
            {"name": "destinationCurrencyKey", "type": "bytes32"},
            {"name": "trackingCode", "type": "bytes32"}
        ],
        "name": "exchange",
        "outputs": [{"name": "amountReceived", "type": "uint256"}],
        "type": "function"
    },
    {
        "inputs": [
            {"name": "account", "type": "address"},
            {"name": "currencyKey", "type": "bytes32"}
        ],
        "name": "settlementOwing",
        "outputs": [
            {"name": "reclaimAmount", "type": "uint256"},
            {"name": "rebateAmount", "type": "uint256"},
            {"name": "numEntries", "type": "uint256"}
        ],
        "type": "function"
    },
    {
        "inputs": [
            {"name": "from", "type": "address"},
            {"name": "currencyKey", "type": "bytes32"}
        ],
        "name": "settle",
        "outputs": [
            {"name": "reclaimed", "type": "uint256"},
            {"name": "refunded", "type": "uint256"},
            {"name": "numEntries", "type": "uint256"}
        ],
        "type": "function"
    },
    {
        "inputs": [
            {"name": "currencyKey", "type": "bytes32"}
        ],
        "name": "feeRateForExchange",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    }
]`

// Fee Pool ABI (简化版)
const FeePoolABI = `[
    {
        "inputs": [],
        "name": "claimFees",
        "outputs": [{"name": "", "type": "bool"}],
        "type": "function"
    },
    {
        "inputs": [],
        "name": "claimOnBehalf",
        "outputs": [{"name": "", "type": "bool"}],
        "type": "function"
    },
    {
        "inputs": [{"name": "account", "type": "address"}],
        "name": "feesAvailable",
        "outputs": [
            {"name": "", "type": "uint256"},
            {"name": "", "type": "uint256"}
        ],
        "type": "function"
    },
    {
        "inputs": [{"name": "account", "type": "address"}],
        "name": "isFeesClaimable",
        "outputs": [{"name": "", "type": "bool"}],
        "type": "function"
    }
]`
```

### 2.2 客户端设置

```go
// client/synthetix_client.go
package client

import (
    "context"
    "math/big"
    
    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
)

type SynthetixClient struct {
    ethClient       *ethclient.Client
    synthetixABI    abi.ABI
    exchangerABI    abi.ABI
    feePoolABI      abi.ABI
    synthetixAddr   common.Address
    exchangerAddr   common.Address
    feePoolAddr     common.Address
}

func NewSynthetixClient(rpcURL string) (*SynthetixClient, error) {
    ethClient, err := ethclient.Dial(rpcURL)
    if err != nil {
        return nil, err
    }
    
    synthetixABI, err := abi.JSON(strings.NewReader(SynthetixABI))
    if err != nil {
        return nil, err
    }
    
    exchangerABI, err := abi.JSON(strings.NewReader(ExchangerABI))
    if err != nil {
        return nil, err
    }
    
    feePoolABI, err := abi.JSON(strings.NewReader(FeePoolABI))
    if err != nil {
        return nil, err
    }
    
    return &SynthetixClient{
        ethClient:     ethClient,
        synthetixABI:  synthetixABI,
        exchangerABI:  exchangerABI,
        feePoolABI:    feePoolABI,
        synthetixAddr: SynthetixAddress,
        exchangerAddr: ExchangerAddress,
        feePoolAddr:   FeePoolAddress,
    }, nil
}

// 获取质押信息
func (c *SynthetixClient) GetStakingInfo(account common.Address) (*StakingInfo, error) {
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    // 获取抵押率
    var collateralisationRatio *big.Int
    err := c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &c.synthetixAddr,
        Data: c.synthetixABI.Methods["collateralisationRatio"].ID,
    }, nil)
    if err != nil {
        return nil, err
    }
    
    // 获取债务余额
    var debtBalance *big.Int
    err = c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &c.synthetixAddr,
        Data: c.synthetixABI.Methods["debtBalanceOf"].ID,
    }, nil)
    if err != nil {
        return nil, err
    }
    
    // 获取最大可铸造Synths
    var maxIssuableSynths *big.Int
    err = c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &c.synthetixAddr,
        Data: c.synthetixABI.Methods["maxIssuableSynths"].ID,
    }, nil)
    if err != nil {
        return nil, err
    }
    
    // 获取可转移的SNX
    var transferableSNX *big.Int
    err = c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &c.synthetixAddr,
        Data: c.synthetixABI.Methods["transferableSynthetix"].ID,
    }, nil)
    if err != nil {
        return nil, err
    }
    
    // 计算抵押率百分比
    collateralisationRatioDecimal := decimal.NewFromBigInt(collateralisationRatio, -18)
    
    return &StakingInfo{
        DebtBalance:            debtBalance,
        CollateralisationRatio: collateralisationRatioDecimal,
        MaxIssuableSynths:      maxIssuableSynths,
        TransferableSNX:        transferableSNX,
    }, nil
}

// 获取交易费率
func (c *SynthetixClient) GetExchangeFeeRate(currencyKey [4]byte) (decimal.Decimal, error) {
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    var feeRate *big.Int
    err := c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &c.exchangerAddr,
        Data: c.exchangerABI.Methods["feeRateForExchange"].ID,
    }, nil)
    if err != nil {
        return decimal.Zero, err
    }
    
    return decimal.NewFromBigInt(feeRate, -18), nil
}

// 获取可领取费用
func (c *SynthetixClient) GetFeesAvailable(account common.Address) (*big.Int, *big.Int, error) {
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    var result []interface{}
    err := c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &c.feePoolAddr,
        Data: c.feePoolABI.Methods["feesAvailable"].ID,
    }, nil)
    if err != nil {
        return nil, nil, err
    }
    
    susdFees := result[0].(*big.Int)
    snxRewards := result[1].(*big.Int)
    
    return susdFees, snxRewards, nil
}
```

## 质押和铸造

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

type StakingService struct {
    client     *SynthetixClient
    privateKey *ecdsa.PrivateKey
}

func NewStakingService(client *SynthetixClient, privateKey *ecdsa.PrivateKey) *StakingService {
    return &StakingService{
        client:     client,
        privateKey: privateKey,
    }
}

// 铸造sUSD
func (s *StakingService) IssueSynths(amount *big.Int) (*types.Transaction, error) {
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 检查质押状态
    stakingInfo, err := s.client.GetStakingInfo(fromAddress)
    if err != nil {
        return nil, err
    }
    
    // 检查是否可以铸造指定数量
    if amount.Cmp(stakingInfo.MaxIssuableSynths) > 0 {
        return nil, fmt.Errorf("铸造数量超过最大可铸造量")
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
    data, err := s.client.synthetixABI.Pack("issueSynths", amount)
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, s.client.synthetixAddr, big.NewInt(0), 300000, gasPrice, data)
    
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

// 销毁sUSD
func (s *StakingService) BurnSynths(amount *big.Int) (*types.Transaction, error) {
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
    
    // 构建交易数据
    data, err := s.client.synthetixABI.Pack("burnSynths", amount)
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, s.client.synthetixAddr, big.NewInt(0), 250000, gasPrice, data)
    
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

// 销毁到目标抵押率
func (s *StakingService) BurnSynthsToTarget() (*types.Transaction, error) {
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
    
    // 构建交易数据
    data, err := s.client.synthetixABI.Pack("burnSynthsToTarget")
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, s.client.synthetixAddr, big.NewInt(0), 300000, gasPrice, data)
    
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

// 计算最优铸造数量
func (s *StakingService) CalculateOptimalMintAmount(
    snxBalance *big.Int,
    snxPrice decimal.Decimal,
    targetCRatio decimal.Decimal,
) *big.Int {
    // 计算SNX总价值
    snxValue := decimal.NewFromBigInt(snxBalance, -18).Mul(snxPrice)
    
    // 计算最大可铸造sUSD = SNX价值 / 目标抵押率
    maxMintable := snxValue.Div(targetCRatio)
    
    // 转换为wei
    maxMintableWei := maxMintable.Mul(decimal.NewFromInt(1e18))
    
    return maxMintableWei.BigInt()
}

// 检查健康度
func (s *StakingService) CheckHealthFactor(account common.Address) (*HealthFactor, error) {
    stakingInfo, err := s.client.GetStakingInfo(account)
    if err != nil {
        return nil, err
    }
    
    // 计算健康度
    healthFactor := &HealthFactor{
        CurrentCRatio:    stakingInfo.CollateralisationRatio,
        TargetCRatio:     decimal.NewFromFloat(4.0), // 400%
        LiquidationRatio: decimal.NewFromFloat(1.5), // 150%
        DebtBalance:      stakingInfo.DebtBalance,
        MaxIssuable:      stakingInfo.MaxIssuableSynths,
    }
    
    // 判断风险等级
    if healthFactor.CurrentCRatio.LessThan(healthFactor.LiquidationRatio) {
        healthFactor.RiskLevel = "CRITICAL"
    } else if healthFactor.CurrentCRatio.LessThan(decimal.NewFromFloat(2.0)) {
        healthFactor.RiskLevel = "HIGH"
    } else if healthFactor.CurrentCRatio.LessThan(healthFactor.TargetCRatio) {
        healthFactor.RiskLevel = "MEDIUM"
    } else {
        healthFactor.RiskLevel = "LOW"
    }
    
    return healthFactor, nil
}

type HealthFactor struct {
    CurrentCRatio    decimal.Decimal
    TargetCRatio     decimal.Decimal
    LiquidationRatio decimal.Decimal
    DebtBalance      *big.Int
    MaxIssuable      *big.Int
    RiskLevel        string
}
```

## 交易和兑换

### 4.1 交易服务

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
    "github.com/shopspring/decimal"
)

type ExchangeService struct {
    client     *SynthetixClient
    privateKey *ecdsa.PrivateKey
}

func NewExchangeService(client *SynthetixClient, privateKey *ecdsa.PrivateKey) *ExchangeService {
    return &ExchangeService{
        client:     client,
        privateKey: privateKey,
    }
}

// 兑换合成资产
func (s *ExchangeService) Exchange(
    sourceCurrencyKey [4]byte,
    sourceAmount *big.Int,
    destinationCurrencyKey [4]byte,
    trackingCode [32]byte,
) (*types.Transaction, error) {
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 获取交易费率
    feeRate, err := s.client.GetExchangeFeeRate(sourceCurrencyKey)
    if err != nil {
        return nil, err
    }
    
    // 计算预期接收数量 (简化计算)
    expectedAmount := s.calculateExpectedAmount(sourceAmount, sourceCurrencyKey, destinationCurrencyKey, feeRate)
    
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
    data, err := s.client.exchangerABI.Pack(
        "exchange",
        sourceCurrencyKey,
        sourceAmount,
        destinationCurrencyKey,
        trackingCode,
    )
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, s.client.exchangerAddr, big.NewInt(0), 400000, gasPrice, data)
    
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

// 结算交易
func (s *ExchangeService) Settle(currencyKey [4]byte) (*types.Transaction, error) {
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
    
    // 构建交易数据
    data, err := s.client.exchangerABI.Pack("settle", fromAddress, currencyKey)
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, s.client.exchangerAddr, big.NewInt(0), 200000, gasPrice, data)
    
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

// 获取结算信息
func (s *ExchangeService) GetSettlementOwing(
    account common.Address,
    currencyKey [4]byte,
) (*SettlementInfo, error) {
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    var result []interface{}
    err := s.client.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &s.client.exchangerAddr,
        Data: s.client.exchangerABI.Methods["settlementOwing"].ID,
    }, nil)
    if err != nil {
        return nil, err
    }
    
    return &SettlementInfo{
        ReclaimAmount: result[0].(*big.Int),
        RebateAmount:  result[1].(*big.Int),
        NumEntries:    result[2].(*big.Int),
    }, nil
}

// 计算预期兑换数量
func (s *ExchangeService) calculateExpectedAmount(
    sourceAmount *big.Int,
    sourceCurrency [4]byte,
    destCurrency [4]byte,
    feeRate decimal.Decimal,
) *big.Int {
    // 这里需要从价格预言机获取实际汇率
    // 简化实现，假设1:1兑换减去手续费
    
    sourceDecimal := decimal.NewFromBigInt(sourceAmount, -18)
    feeAmount := sourceDecimal.Mul(feeRate)
    netAmount := sourceDecimal.Sub(feeAmount)
    
    return netAmount.Mul(decimal.NewFromInt(1e18)).BigInt()
}

// 批量兑换
func (s *ExchangeService) BatchExchange(exchanges []ExchangeParams) ([]*types.Transaction, error) {
    var transactions []*types.Transaction
    
    for _, exchange := range exchanges {
        tx, err := s.Exchange(
            exchange.SourceCurrency,
            exchange.SourceAmount,
            exchange.DestCurrency,
            exchange.TrackingCode,
        )
        if err != nil {
            return nil, err
        }
        
        transactions = append(transactions, tx)
        
        // 等待一段时间避免nonce冲突
        time.Sleep(1 * time.Second)
    }
    
    return transactions, nil
}

// 获取最佳兑换路径
func (s *ExchangeService) GetBestExchangePath(
    sourceCurrency [4]byte,
    destCurrency [4]byte,
    amount *big.Int,
) (*ExchangePath, error) {
    // 直接兑换
    directFeeRate, err := s.client.GetExchangeFeeRate(sourceCurrency)
    if err != nil {
        return nil, err
    }
    
    directAmount := s.calculateExpectedAmount(amount, sourceCurrency, destCurrency, directFeeRate)
    
    // 通过sUSD中转
    susdFeeRate, err := s.client.GetExchangeFeeRate(SUSD)
    if err != nil {
        return nil, err
    }
    
    // 第一步：源货币 -> sUSD
    susdAmount := s.calculateExpectedAmount(amount, sourceCurrency, SUSD, directFeeRate)
    
    // 第二步：sUSD -> 目标货币
    indirectAmount := s.calculateExpectedAmount(susdAmount, SUSD, destCurrency, susdFeeRate)
    
    // 选择最优路径
    if directAmount.Cmp(indirectAmount) > 0 {
        return &ExchangePath{
            Type:           "DIRECT",
            Steps:          []ExchangeStep{{From: sourceCurrency, To: destCurrency, Amount: amount}},
            ExpectedAmount: directAmount,
            TotalFee:       directFeeRate,
        }, nil
    } else {
        return &ExchangePath{
            Type: "INDIRECT",
            Steps: []ExchangeStep{
                {From: sourceCurrency, To: SUSD, Amount: amount},
                {From: SUSD, To: destCurrency, Amount: susdAmount},
            },
            ExpectedAmount: indirectAmount,
            TotalFee:       directFeeRate.Add(susdFeeRate),
        }, nil
    }
}

type ExchangeParams struct {
    SourceCurrency [4]byte
    SourceAmount   *big.Int
    DestCurrency   [4]byte
    TrackingCode   [32]byte
}

type SettlementInfo struct {
    ReclaimAmount *big.Int
    RebateAmount  *big.Int
    NumEntries    *big.Int
}

type ExchangePath struct {
    Type           string
    Steps          []ExchangeStep
    ExpectedAmount *big.Int
    TotalFee       decimal.Decimal
}

type ExchangeStep struct {
    From   [4]byte
    To     [4]byte
    Amount *big.Int
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
    // 创建Synthetix客户端
    synthetixClient, err := client.NewSynthetixClient("https://mainnet.infura.io/v3/YOUR_PROJECT_ID")
    if err != nil {
        log.Fatal("创建Synthetix客户端失败:", err)
    }
    
    // 加载私钥
    privateKey, err := crypto.HexToECDSA("your_private_key_here")
    if err != nil {
        log.Fatal("加载私钥失败:", err)
    }
    
    // 获取用户地址
    userAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
    
    // 创建服务
    stakingService := services.NewStakingService(synthetixClient, privateKey)
    exchangeService := services.NewExchangeService(synthetixClient, privateKey)
    
    // 1. 查询质押信息
    fmt.Printf("=== 质押信息查询 ===\n")
    
    stakingInfo, err := synthetixClient.GetStakingInfo(userAddress)
    if err != nil {
        log.Fatal("获取质押信息失败:", err)
    }
    
    fmt.Printf("债务余额: %s sUSD\n", stakingInfo.DebtBalance.String())
    fmt.Printf("抵押率: %s%%\n", stakingInfo.CollateralisationRatio.Mul(decimal.NewFromInt(100)).String())
    fmt.Printf("最大可铸造: %s sUSD\n", stakingInfo.MaxIssuableSynths.String())
    fmt.Printf("可转移SNX: %s\n", stakingInfo.TransferableSNX.String())
    
    // 2. 检查健康度
    fmt.Printf("\n=== 健康度检查 ===\n")
    
    healthFactor, err := stakingService.CheckHealthFactor(userAddress)
    if err != nil {
        log.Fatal("检查健康度失败:", err)
    }
    
    fmt.Printf("当前抵押率: %s%%\n", healthFactor.CurrentCRatio.Mul(decimal.NewFromInt(100)).String())
    fmt.Printf("目标抵押率: %s%%\n", healthFactor.TargetCRatio.Mul(decimal.NewFromInt(100)).String())
    fmt.Printf("清算抵押率: %s%%\n", healthFactor.LiquidationRatio.Mul(decimal.NewFromInt(100)).String())
    fmt.Printf("风险等级: %s\n", healthFactor.RiskLevel)
    
    // 3. 铸造sUSD示例
    fmt.Printf("\n=== 铸造sUSD示例 ===\n")
    
    if stakingInfo.MaxIssuableSynths.Cmp(big.NewInt(0)) > 0 {
        // 铸造一半的最大可铸造量
        mintAmount := new(big.Int).Div(stakingInfo.MaxIssuableSynths, big.NewInt(2))
        
        fmt.Printf("准备铸造 %s sUSD\n", mintAmount.String())
        
        tx, err := stakingService.IssueSynths(mintAmount)
        if err != nil {
            log.Printf("铸造sUSD失败: %v", err)
        } else {
            fmt.Printf("铸造交易已提交: %s\n", tx.Hash().Hex())
        }
    } else {
        fmt.Printf("当前无法铸造sUSD，请检查SNX质押状态\n")
    }
    
    // 4. 合成资产兑换示例
    fmt.Printf("\n=== 合成资产兑换示例 ===\n")
    
    // 查询sUSD到sBTC的兑换费率
    feeRate, err := synthetixClient.GetExchangeFeeRate(SUSD)
    if err != nil {
        log.Printf("获取兑换费率失败: %v", err)
    } else {
        fmt.Printf("sUSD兑换费率: %s%%\n", feeRate.Mul(decimal.NewFromInt(100)).String())
    }
    
    // 获取最佳兑换路径
    exchangeAmount := big.NewInt(100e18) // 100 sUSD
    
    path, err := exchangeService.GetBestExchangePath(SUSD, SBTC, exchangeAmount)
    if err != nil {
        log.Printf("获取兑换路径失败: %v", err)
    } else {
        fmt.Printf("最佳兑换路径: %s\n", path.Type)
        fmt.Printf("预期接收: %s sBTC\n", path.ExpectedAmount.String())
        fmt.Printf("总手续费: %s%%\n", path.TotalFee.Mul(decimal.NewFromInt(100)).String())
        
        // 执行兑换
        if len(path.Steps) == 1 {
            // 直接兑换
            tx, err := exchangeService.Exchange(
                path.Steps[0].From,
                path.Steps[0].Amount,
                path.Steps[0].To,
                [32]byte{}, // 空的tracking code
            )
            if err != nil {
                log.Printf("兑换失败: %v", err)
            } else {
                fmt.Printf("兑换交易已提交: %s\n", tx.Hash().Hex())
            }
        }
    }
    
    // 5. 查询可领取费用
    fmt.Printf("\n=== 费用查询 ===\n")
    
    susdFees, snxRewards, err := synthetixClient.GetFeesAvailable(userAddress)
    if err != nil {
        log.Printf("查询费用失败: %v", err)
    } else {
        fmt.Printf("可领取sUSD费用: %s\n", susdFees.String())
        fmt.Printf("可领取SNX奖励: %s\n", snxRewards.String())
        
        // 如果有可领取的费用，可以调用claimFees
        if susdFees.Cmp(big.NewInt(0)) > 0 || snxRewards.Cmp(big.NewInt(0)) > 0 {
            fmt.Printf("检测到可领取费用，可调用claimFees()领取\n")
        }
    }
    
    // 6. 结算检查
    fmt.Printf("\n=== 结算检查 ===\n")
    
    settlementInfo, err := exchangeService.GetSettlementOwing(userAddress, SBTC)
    if err != nil {
        log.Printf("查询结算信息失败: %v", err)
    } else {
        if settlementInfo.NumEntries.Cmp(big.NewInt(0)) > 0 {
            fmt.Printf("待结算条目数: %s\n", settlementInfo.NumEntries.String())
            fmt.Printf("回收金额: %s\n", settlementInfo.ReclaimAmount.String())
            fmt.Printf("退还金额: %s\n", settlementInfo.RebateAmount.String())
            
            // 执行结算
            tx, err := exchangeService.Settle(SBTC)
            if err != nil {
                log.Printf("结算失败: %v", err)
            } else {
                fmt.Printf("结算交易已提交: %s\n", tx.Hash().Hex())
            }
        } else {
            fmt.Printf("sBTC无待结算条目\n")
        }
    }
    
    // 7. 风险管理建议
    fmt.Printf("\n=== 风险管理建议 ===\n")
    
    switch healthFactor.RiskLevel {
    case "CRITICAL":
        fmt.Printf("⚠️ 紧急警告：抵押率过低，面临清算风险！\n")
        fmt.Printf("建议立即：\n")
        fmt.Printf("1. 增加SNX质押\n")
        fmt.Printf("2. 销毁部分sUSD\n")
        fmt.Printf("3. 调用burnSynthsToTarget()自动调整到目标抵押率\n")
        
    case "HIGH":
        fmt.Printf("⚠️ 高风险：抵押率偏低\n")
        fmt.Printf("建议：销毁部分sUSD或增加SNX质押\n")
        
    case "MEDIUM":
        fmt.Printf("⚡ 中等风险：抵押率低于目标值\n")
        fmt.Printf("建议：考虑调整抵押率到400%以上\n")
        
    case "LOW":
        fmt.Printf("✅ 低风险：抵押率健康\n")
        fmt.Printf("可以考虑：\n")
        fmt.Printf("1. 铸造更多sUSD增加收益\n")
        fmt.Printf("2. 参与合成资产交易\n")
    }
    
    // 8. 计算最优操作
    fmt.Printf("\n=== 最优操作建议 ===\n")
    
    if healthFactor.RiskLevel == "LOW" {
        // 计算最优铸造数量
        snxBalance := big.NewInt(10000e18) // 假设有10000 SNX
        snxPrice := decimal.NewFromFloat(3.5) // 假设SNX价格为$3.5
        targetCRatio := decimal.NewFromFloat(4.0) // 400%目标抵押率
        
        optimalMint := stakingService.CalculateOptimalMintAmount(snxBalance, snxPrice, targetCRatio)
        
        fmt.Printf("基于当前SNX持仓，最优铸造数量: %s sUSD\n", optimalMint.String())
        fmt.Printf("这将使抵押率保持在目标400%%\n")
    }
    
    // 9. 合成资产组合建议
    fmt.Printf("\n=== 合成资产组合建议 ===\n")
    
    fmt.Printf("推荐的合成资产配置：\n")
    fmt.Printf("1. 40%% sUSD - 稳定币基础\n")
    fmt.Printf("2. 30%% sBTC - 加密货币敞口\n")
    fmt.Printf("3. 20%% sETH - 以太坊敞口\n")
    fmt.Printf("4. 10%% 其他 - sGOLD, sEUR等多样化\n")
    
    fmt.Printf("\n合成资产交易优势：\n")
    fmt.Printf("- 无滑点交易\n")
    fmt.Printf("- 24/7交易传统资产\n")
    fmt.Printf("- 无需持有底层资产\n")
    fmt.Printf("- 全球资产敞口\n")
}
```

这个Synthetix使用指南提供了完整的合成资产协议集成方案，涵盖了质押铸造、合成资产交易、债务池管理、风险控制等核心功能，是DeFi合成资产交易的重要参考文档。
