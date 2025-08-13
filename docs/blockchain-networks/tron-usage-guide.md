# Tron 区块链 详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [环境准备](#环境准备)
3. [账户系统](#账户系统)
4. [智能合约](#智能合约)
5. [DeFi生态](#defi生态)
6. [资源管理](#资源管理)
7. [治理机制](#治理机制)
8. [性能优化](#性能优化)
9. [实际应用](#实际应用)

## 基础概念

### 1.1 Tron 简介

Tron 是高性能区块链平台，采用DPoS共识机制，支持智能合约和DApp开发，以高吞吐量、低费用和强大的DeFi生态著称。

```bash
# 安装Tron相关依赖
go get github.com/ethereum/go-ethereum
go get github.com/fbsobreira/gotron-sdk
go get github.com/shopspring/decimal
```

### 1.2 核心概念

```go
// 主要包导入
import (
    "context"
    "encoding/hex"
    "math/big"
    "strings"
    
    "github.com/fbsobreira/gotron-sdk/pkg/client"
    "github.com/fbsobreira/gotron-sdk/pkg/proto/api"
    "github.com/fbsobreira/gotron-sdk/pkg/proto/core"
    "github.com/shopspring/decimal"
)

// Tron 网络配置
var (
    // Mainnet
    MainnetNodes = []string{
        "grpc.trongrid.io:50051",
        "grpc.shasta.trongrid.io:50051",
    }
    
    // Testnet (Shasta)
    TestnetNodes = []string{
        "grpc.shasta.trongrid.io:50051",
    }
    
    // Nile Testnet
    NileTestnetNodes = []string{
        "grpc.nile.trongrid.io:50051",
    }
)

// Tron 网络信息
type NetworkInfo struct {
    NetworkID   string
    NetworkName string
    ChainID     string
    Nodes       []string
    ExplorerURL string
    Currency    string
    Decimals    int
}

var (
    MainnetInfo = NetworkInfo{
        NetworkID:   "1",
        NetworkName: "Tron Mainnet",
        ChainID:     "0x2b6653dc",
        Nodes:       MainnetNodes,
        ExplorerURL: "https://tronscan.org",
        Currency:    "TRX",
        Decimals:    6,
    }
    
    ShastaInfo = NetworkInfo{
        NetworkID:   "2",
        NetworkName: "Shasta Testnet",
        ChainID:     "0x94a9059e",
        Nodes:       TestnetNodes,
        ExplorerURL: "https://shasta.tronscan.org",
        Currency:    "TRX",
        Decimals:    6,
    }
)

// Tron 核心合约地址
var (
    // USDT TRC20
    USDTAddress = "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"
    
    // USDC TRC20
    USDCAddress = "TEkxiTehnzSmSe2XqrBj4w32RUN966rdz8"
    
    // WTRX
    WTRXAddress = "TNUC9Qb1rRpS5CbWLmNMxXBjyFoydXjWFR"
    
    // JustSwap Router
    JustSwapRouterAddress = "TKzxdSv2FZKQrEqkKVgp5DcwEXBEKMg2Ax"
    
    // SUN Token
    SUNAddress = "TSSMHYeV2uE9qYH95DqyoCuNCzEL1NvU3S"
    
    // BTT Token
    BTTAddress = "TAFjULxiVgT4qWk6UZwjqwZXTSaGaqnVp4"
    
    // WIN Token
    WINAddress = "TLa2f6VPqDgRE67v1736s7bJ8Ray5wYjU7"
)

// 账户信息
type Account struct {
    Address         string
    Balance         *big.Int
    TRC20Balances   map[string]*big.Int
    Bandwidth       *big.Int
    Energy          *big.Int
    FrozenBalance   *big.Int
    VotingPower     *big.Int
    AccountResource *AccountResource
}

// 账户资源
type AccountResource struct {
    FreeNetUsed     *big.Int
    FreeNetLimit    *big.Int
    NetUsed         *big.Int
    NetLimit        *big.Int
    EnergyUsed      *big.Int
    EnergyLimit     *big.Int
    TotalNetLimit   *big.Int
    TotalNetWeight  *big.Int
    TotalEnergyLimit *big.Int
    TotalEnergyWeight *big.Int
}

// 交易信息
type Transaction struct {
    TxID            string
    BlockNumber     *big.Int
    BlockTimeStamp  *big.Int
    ContractResult  []string
    Receipt         *TransactionReceipt
    Log             []*TransactionLog
}

type TransactionReceipt struct {
    EnergyUsage     *big.Int
    EnergyFee       *big.Int
    OriginEnergyUsage *big.Int
    EnergyUsageTotal *big.Int
    NetUsage        *big.Int
    NetFee          *big.Int
    Result          string
}

type TransactionLog struct {
    Address string
    Topics  []string
    Data    string
}

// 智能合约信息
type Contract struct {
    Address         string
    ByteCode        string
    Name            string
    ABI             string
    SourceCode      string
    CompilerVersion string
    OptimizationUsed bool
    Runs            int
    ConstructorArguments string
}

// 超级代表信息
type SuperRepresentative struct {
    Address     string
    URL         string
    VoteCount   *big.Int
    TotalProduced *big.Int
    TotalMissed *big.Int
    LatestBlockNum *big.Int
    LatestSlotNum *big.Int
    IsJobs      bool
}

// 提案信息
type Proposal struct {
    ProposalID      *big.Int
    ProposerAddress string
    Parameters      map[string]*big.Int
    ExpirationTime  *big.Int
    CreateTime      *big.Int
    Approvals       []string
    State           string
}

// 冻结信息
type FrozenInfo struct {
    FrozenBalance   *big.Int
    ExpireTime      *big.Int
    Resource        string // "BANDWIDTH" or "ENERGY"
}

// DeFi协议信息
type DeFiProtocol struct {
    Name            string
    Address         string
    TVL             decimal.Decimal
    APY             decimal.Decimal
    TokenPairs      []TokenPair
    Volume24h       decimal.Decimal
}

type TokenPair struct {
    Token0          string
    Token1          string
    Reserve0        *big.Int
    Reserve1        *big.Int
    LPTokenAddress  string
    TotalSupply     *big.Int
}
```

## 环境准备

### 2.1 Tron客户端设置

```go
// client/tron_client.go
package client

import (
    "context"
    "encoding/hex"
    "math/big"
    
    "github.com/fbsobreira/gotron-sdk/pkg/client"
    "github.com/fbsobreira/gotron-sdk/pkg/proto/api"
    "github.com/fbsobreira/gotron-sdk/pkg/proto/core"
)

type TronClient struct {
    grpcClient  *client.GrpcClient
    networkInfo *NetworkInfo
}

func NewTronClient(networkInfo *NetworkInfo) (*TronClient, error) {
    // 连接到Tron节点
    grpcClient := client.NewGrpcClient(networkInfo.Nodes[0])
    err := grpcClient.Start()
    if err != nil {
        return nil, err
    }
    
    return &TronClient{
        grpcClient:  grpcClient,
        networkInfo: networkInfo,
    }, nil
}

// 获取网络信息
func (c *TronClient) GetNetworkInfo() *NetworkInfo {
    return c.networkInfo
}

// 获取账户信息
func (c *TronClient) GetAccount(address string) (*Account, error) {
    // 转换地址格式
    addr, err := client.Base58ToAddress(address)
    if err != nil {
        return nil, err
    }
    
    // 获取账户信息
    account, err := c.grpcClient.GetAccount(addr.String())
    if err != nil {
        return nil, err
    }
    
    // 获取账户资源
    resource, err := c.grpcClient.GetAccountResource(addr.String())
    if err != nil {
        return nil, err
    }
    
    // 构建账户信息
    accountInfo := &Account{
        Address:         address,
        Balance:         big.NewInt(account.Balance),
        TRC20Balances:   make(map[string]*big.Int),
        Bandwidth:       big.NewInt(0),
        Energy:          big.NewInt(0),
        FrozenBalance:   big.NewInt(0),
        VotingPower:     big.NewInt(0),
        AccountResource: &AccountResource{
            FreeNetUsed:      big.NewInt(resource.FreeNetUsed),
            FreeNetLimit:     big.NewInt(resource.FreeNetLimit),
            NetUsed:          big.NewInt(resource.NetUsed),
            NetLimit:         big.NewInt(resource.NetLimit),
            EnergyUsed:       big.NewInt(resource.EnergyUsed),
            EnergyLimit:      big.NewInt(resource.EnergyLimit),
            TotalNetLimit:    big.NewInt(resource.TotalNetLimit),
            TotalNetWeight:   big.NewInt(resource.TotalNetWeight),
            TotalEnergyLimit: big.NewInt(resource.TotalEnergyLimit),
            TotalEnergyWeight: big.NewInt(resource.TotalEnergyWeight),
        },
    }
    
    // 计算冻结余额
    for _, frozen := range account.Frozen {
        accountInfo.FrozenBalance.Add(accountInfo.FrozenBalance, big.NewInt(frozen.FrozenBalance))
    }
    
    // 计算投票权力
    for _, vote := range account.Votes {
        accountInfo.VotingPower.Add(accountInfo.VotingPower, big.NewInt(vote.VoteCount))
    }
    
    return accountInfo, nil
}

// 获取TRC20代币余额
func (c *TronClient) GetTRC20Balance(contractAddress, ownerAddress string) (*big.Int, error) {
    // 构建balanceOf调用
    contract, err := client.Base58ToAddress(contractAddress)
    if err != nil {
        return nil, err
    }
    
    owner, err := client.Base58ToAddress(ownerAddress)
    if err != nil {
        return nil, err
    }
    
    // 调用智能合约
    result, err := c.grpcClient.TriggerConstantContract(
        owner.String(),
        contract.String(),
        "balanceOf(address)",
        hex.EncodeToString(owner.Bytes()),
    )
    if err != nil {
        return nil, err
    }
    
    if len(result.ConstantResult) == 0 {
        return big.NewInt(0), nil
    }
    
    // 解析结果
    balance := new(big.Int)
    balance.SetBytes(result.ConstantResult[0])
    
    return balance, nil
}

// 发送TRX
func (c *TronClient) SendTRX(from, to string, amount *big.Int, privateKey string) (*Transaction, error) {
    // 转换地址
    fromAddr, err := client.Base58ToAddress(from)
    if err != nil {
        return nil, err
    }
    
    toAddr, err := client.Base58ToAddress(to)
    if err != nil {
        return nil, err
    }
    
    // 创建转账交易
    tx, err := c.grpcClient.Transfer(fromAddr.String(), toAddr.String(), amount.Int64())
    if err != nil {
        return nil, err
    }
    
    // 签名交易
    signedTx, err := c.grpcClient.SignTransaction(tx, privateKey)
    if err != nil {
        return nil, err
    }
    
    // 广播交易
    result, err := c.grpcClient.Broadcast(signedTx)
    if err != nil {
        return nil, err
    }
    
    return &Transaction{
        TxID: hex.EncodeToString(result.Txid),
    }, nil
}

// 发送TRC20代币
func (c *TronClient) SendTRC20(
    contractAddress, from, to string,
    amount *big.Int,
    privateKey string,
) (*Transaction, error) {
    // 转换地址
    contract, err := client.Base58ToAddress(contractAddress)
    if err != nil {
        return nil, err
    }
    
    fromAddr, err := client.Base58ToAddress(from)
    if err != nil {
        return nil, err
    }
    
    toAddr, err := client.Base58ToAddress(to)
    if err != nil {
        return nil, err
    }
    
    // 构建transfer调用参数
    params := hex.EncodeToString(toAddr.Bytes()) + 
             hex.EncodeToString(common.LeftPadBytes(amount.Bytes(), 32))
    
    // 创建智能合约调用交易
    tx, err := c.grpcClient.TriggerContract(
        fromAddr.String(),
        contract.String(),
        "transfer(address,uint256)",
        params,
        0, // feeLimit
        0, // callValue
    )
    if err != nil {
        return nil, err
    }
    
    // 签名交易
    signedTx, err := c.grpcClient.SignTransaction(tx.Transaction, privateKey)
    if err != nil {
        return nil, err
    }
    
    // 广播交易
    result, err := c.grpcClient.Broadcast(signedTx)
    if err != nil {
        return nil, err
    }
    
    return &Transaction{
        TxID: hex.EncodeToString(result.Txid),
    }, nil
}

// 获取交易信息
func (c *TronClient) GetTransaction(txID string) (*Transaction, error) {
    // 转换交易ID
    txIDBytes, err := hex.DecodeString(txID)
    if err != nil {
        return nil, err
    }
    
    // 获取交易信息
    tx, err := c.grpcClient.GetTransactionByID(hex.EncodeToString(txIDBytes))
    if err != nil {
        return nil, err
    }
    
    // 获取交易收据
    receipt, err := c.grpcClient.GetTransactionInfoByID(hex.EncodeToString(txIDBytes))
    if err != nil {
        return nil, err
    }
    
    transaction := &Transaction{
        TxID:           txID,
        BlockNumber:    big.NewInt(receipt.BlockNumber),
        BlockTimeStamp: big.NewInt(receipt.BlockTimeStamp),
        ContractResult: receipt.ContractResult,
        Receipt: &TransactionReceipt{
            EnergyUsage:       big.NewInt(receipt.Receipt.EnergyUsage),
            EnergyFee:         big.NewInt(receipt.Receipt.EnergyFee),
            OriginEnergyUsage: big.NewInt(receipt.Receipt.OriginEnergyUsage),
            EnergyUsageTotal:  big.NewInt(receipt.Receipt.EnergyUsageTotal),
            NetUsage:          big.NewInt(receipt.Receipt.NetUsage),
            NetFee:            big.NewInt(receipt.Receipt.NetFee),
            Result:            receipt.Receipt.Result.String(),
        },
    }
    
    // 解析日志
    for _, log := range receipt.Log {
        transactionLog := &TransactionLog{
            Address: client.AddressToBase58(log.Address),
            Topics:  make([]string, len(log.Topics)),
            Data:    hex.EncodeToString(log.Data),
        }
        
        for i, topic := range log.Topics {
            transactionLog.Topics[i] = hex.EncodeToString(topic)
        }
        
        transaction.Log = append(transaction.Log, transactionLog)
    }
    
    return transaction, nil
}

// 获取超级代表列表
func (c *TronClient) GetSuperRepresentatives() ([]*SuperRepresentative, error) {
    witnesses, err := c.grpcClient.ListWitnesses()
    if err != nil {
        return nil, err
    }
    
    var srs []*SuperRepresentative
    for _, witness := range witnesses.Witnesses {
        sr := &SuperRepresentative{
            Address:        client.AddressToBase58(witness.Address),
            URL:            string(witness.Url),
            VoteCount:      big.NewInt(witness.VoteCount),
            TotalProduced:  big.NewInt(witness.TotalProduced),
            TotalMissed:    big.NewInt(witness.TotalMissed),
            LatestBlockNum: big.NewInt(witness.LatestBlockNum),
            LatestSlotNum:  big.NewInt(witness.LatestSlotNum),
            IsJobs:         witness.IsJobs,
        }
        srs = append(srs, sr)
    }
    
    return srs, nil
}

// 冻结TRX获取资源
func (c *TronClient) FreezeBalance(
    ownerAddress string,
    frozenBalance *big.Int,
    frozenDuration int64,
    resource string,
    privateKey string,
) (*Transaction, error) {
    // 转换地址
    owner, err := client.Base58ToAddress(ownerAddress)
    if err != nil {
        return nil, err
    }
    
    // 确定资源类型
    var resourceType core.ResourceCode
    switch resource {
    case "BANDWIDTH":
        resourceType = core.ResourceCode_BANDWIDTH
    case "ENERGY":
        resourceType = core.ResourceCode_ENERGY
    default:
        return nil, fmt.Errorf("无效的资源类型: %s", resource)
    }
    
    // 创建冻结交易
    tx, err := c.grpcClient.FreezeBalance(
        owner.String(),
        frozenBalance.Int64(),
        frozenDuration,
        resourceType,
    )
    if err != nil {
        return nil, err
    }
    
    // 签名交易
    signedTx, err := c.grpcClient.SignTransaction(tx, privateKey)
    if err != nil {
        return nil, err
    }
    
    // 广播交易
    result, err := c.grpcClient.Broadcast(signedTx)
    if err != nil {
        return nil, err
    }
    
    return &Transaction{
        TxID: hex.EncodeToString(result.Txid),
    }, nil
}

// 解冻TRX
func (c *TronClient) UnfreezeBalance(
    ownerAddress string,
    resource string,
    privateKey string,
) (*Transaction, error) {
    // 转换地址
    owner, err := client.Base58ToAddress(ownerAddress)
    if err != nil {
        return nil, err
    }
    
    // 确定资源类型
    var resourceType core.ResourceCode
    switch resource {
    case "BANDWIDTH":
        resourceType = core.ResourceCode_BANDWIDTH
    case "ENERGY":
        resourceType = core.ResourceCode_ENERGY
    default:
        return nil, fmt.Errorf("无效的资源类型: %s", resource)
    }
    
    // 创建解冻交易
    tx, err := c.grpcClient.UnfreezeBalance(owner.String(), resourceType)
    if err != nil {
        return nil, err
    }
    
    // 签名交易
    signedTx, err := c.grpcClient.SignTransaction(tx, privateKey)
    if err != nil {
        return nil, err
    }
    
    // 广播交易
    result, err := c.grpcClient.Broadcast(signedTx)
    if err != nil {
        return nil, err
    }
    
    return &Transaction{
        TxID: hex.EncodeToString(result.Txid),
    }, nil
}
```

## 智能合约

### 3.1 合约服务

```go
// services/contract_service.go
package services

import (
    "context"
    "encoding/hex"
    "math/big"
    
    "github.com/fbsobreira/gotron-sdk/pkg/client"
)

type TronContractService struct {
    client     *TronClient
    privateKey string
}

func NewTronContractService(client *TronClient, privateKey string) *TronContractService {
    return &TronContractService{
        client:     client,
        privateKey: privateKey,
    }
}

// 部署智能合约
func (s *TronContractService) DeployContract(
    ownerAddress string,
    contractName string,
    abi string,
    bytecode string,
    constructorParams string,
    feeLimit int64,
) (*Transaction, error) {
    // 转换地址
    owner, err := client.Base58ToAddress(ownerAddress)
    if err != nil {
        return nil, err
    }
    
    // 创建部署交易
    tx, err := s.client.grpcClient.DeployContract(
        owner.String(),
        contractName,
        abi,
        bytecode,
        constructorParams,
        feeLimit,
        0, // callValue
    )
    if err != nil {
        return nil, err
    }
    
    // 签名交易
    signedTx, err := s.client.grpcClient.SignTransaction(tx.Transaction, s.privateKey)
    if err != nil {
        return nil, err
    }
    
    // 广播交易
    result, err := s.client.grpcClient.Broadcast(signedTx)
    if err != nil {
        return nil, err
    }
    
    return &Transaction{
        TxID: hex.EncodeToString(result.Txid),
    }, nil
}

// 调用智能合约
func (s *TronContractService) CallContract(
    ownerAddress string,
    contractAddress string,
    functionSelector string,
    parameter string,
    feeLimit int64,
    callValue int64,
) (*Transaction, error) {
    // 转换地址
    owner, err := client.Base58ToAddress(ownerAddress)
    if err != nil {
        return nil, err
    }
    
    contract, err := client.Base58ToAddress(contractAddress)
    if err != nil {
        return nil, err
    }
    
    // 创建合约调用交易
    tx, err := s.client.grpcClient.TriggerContract(
        owner.String(),
        contract.String(),
        functionSelector,
        parameter,
        feeLimit,
        callValue,
    )
    if err != nil {
        return nil, err
    }
    
    // 签名交易
    signedTx, err := s.client.grpcClient.SignTransaction(tx.Transaction, s.privateKey)
    if err != nil {
        return nil, err
    }
    
    // 广播交易
    result, err := s.client.grpcClient.Broadcast(signedTx)
    if err != nil {
        return nil, err
    }
    
    return &Transaction{
        TxID: hex.EncodeToString(result.Txid),
    }, nil
}

// 查询智能合约
func (s *TronContractService) QueryContract(
    ownerAddress string,
    contractAddress string,
    functionSelector string,
    parameter string,
) ([]byte, error) {
    // 转换地址
    owner, err := client.Base58ToAddress(ownerAddress)
    if err != nil {
        return nil, err
    }
    
    contract, err := client.Base58ToAddress(contractAddress)
    if err != nil {
        return nil, err
    }
    
    // 调用常量合约方法
    result, err := s.client.grpcClient.TriggerConstantContract(
        owner.String(),
        contract.String(),
        functionSelector,
        parameter,
    )
    if err != nil {
        return nil, err
    }
    
    if len(result.ConstantResult) == 0 {
        return []byte{}, nil
    }
    
    return result.ConstantResult[0], nil
}

// 获取合约信息
func (s *TronContractService) GetContract(contractAddress string) (*Contract, error) {
    // 转换地址
    contract, err := client.Base58ToAddress(contractAddress)
    if err != nil {
        return nil, err
    }
    
    // 获取合约信息
    contractInfo, err := s.client.grpcClient.GetContract(contract.String())
    if err != nil {
        return nil, err
    }
    
    return &Contract{
        Address:  contractAddress,
        ByteCode: hex.EncodeToString(contractInfo.Bytecode),
        Name:     contractInfo.Name,
        ABI:      contractInfo.Abi.String(),
    }, nil
}

// 估算合约调用费用
func (s *TronContractService) EstimateEnergy(
    ownerAddress string,
    contractAddress string,
    functionSelector string,
    parameter string,
    callValue int64,
) (*big.Int, error) {
    // 转换地址
    owner, err := client.Base58ToAddress(ownerAddress)
    if err != nil {
        return nil, err
    }
    
    contract, err := client.Base58ToAddress(contractAddress)
    if err != nil {
        return nil, err
    }
    
    // 估算能量消耗
    result, err := s.client.grpcClient.EstimateEnergy(
        owner.String(),
        contract.String(),
        functionSelector,
        parameter,
        callValue,
    )
    if err != nil {
        return nil, err
    }
    
    return big.NewInt(result.EnergyRequired), nil
}
```

## DeFi生态

### 4.1 JustSwap集成

```go
// services/justswap_service.go
package services

import (
    "encoding/hex"
    "math/big"
    
    "github.com/fbsobreira/gotron-sdk/pkg/client"
)

type JustSwapService struct {
    client     *TronClient
    privateKey string
}

func NewJustSwapService(client *TronClient, privateKey string) *JustSwapService {
    return &JustSwapService{
        client:     client,
        privateKey: privateKey,
    }
}

// TRX换代币
func (s *JustSwapService) SwapTRXForTokens(
    ownerAddress string,
    tokenAddress string,
    trxAmount *big.Int,
    minTokenAmount *big.Int,
    deadline *big.Int,
) (*Transaction, error) {
    // 构建交换参数
    params := s.buildSwapParams(minTokenAmount, []string{WTRXAddress, tokenAddress}, ownerAddress, deadline)
    
    // 调用JustSwap路由合约
    return s.client.grpcClient.TriggerContract(
        ownerAddress,
        JustSwapRouterAddress,
        "swapExactETHForTokens(uint256,address[],address,uint256)",
        params,
        1000000, // feeLimit
        trxAmount.Int64(), // callValue
    )
}

// 代币换TRX
func (s *JustSwapService) SwapTokensForTRX(
    ownerAddress string,
    tokenAddress string,
    tokenAmount *big.Int,
    minTRXAmount *big.Int,
    deadline *big.Int,
) (*Transaction, error) {
    // 先授权代币
    err := s.approveToken(ownerAddress, tokenAddress, JustSwapRouterAddress, tokenAmount)
    if err != nil {
        return nil, err
    }
    
    // 构建交换参数
    params := s.buildSwapParams(tokenAmount, []string{tokenAddress, WTRXAddress}, ownerAddress, deadline)
    params = hex.EncodeToString(common.LeftPadBytes(minTRXAmount.Bytes(), 32)) + params
    
    // 调用JustSwap路由合约
    return s.client.grpcClient.TriggerContract(
        ownerAddress,
        JustSwapRouterAddress,
        "swapExactTokensForETH(uint256,uint256,address[],address,uint256)",
        params,
        1000000, // feeLimit
        0, // callValue
    )
}

// 添加流动性
func (s *JustSwapService) AddLiquidity(
    ownerAddress string,
    tokenA string,
    tokenB string,
    amountADesired *big.Int,
    amountBDesired *big.Int,
    amountAMin *big.Int,
    amountBMin *big.Int,
    deadline *big.Int,
) (*Transaction, error) {
    // 授权两个代币
    err := s.approveToken(ownerAddress, tokenA, JustSwapRouterAddress, amountADesired)
    if err != nil {
        return nil, err
    }
    
    err = s.approveToken(ownerAddress, tokenB, JustSwapRouterAddress, amountBDesired)
    if err != nil {
        return nil, err
    }
    
    // 构建添加流动性参数
    params := s.buildAddLiquidityParams(
        tokenA, tokenB,
        amountADesired, amountBDesired,
        amountAMin, amountBMin,
        ownerAddress, deadline,
    )
    
    // 调用JustSwap路由合约
    return s.client.grpcClient.TriggerContract(
        ownerAddress,
        JustSwapRouterAddress,
        "addLiquidity(address,address,uint256,uint256,uint256,uint256,address,uint256)",
        params,
        1000000, // feeLimit
        0, // callValue
    )
}

// 移除流动性
func (s *JustSwapService) RemoveLiquidity(
    ownerAddress string,
    tokenA string,
    tokenB string,
    liquidity *big.Int,
    amountAMin *big.Int,
    amountBMin *big.Int,
    deadline *big.Int,
) (*Transaction, error) {
    // 获取LP代币地址
    pairAddress, err := s.getPairAddress(tokenA, tokenB)
    if err != nil {
        return nil, err
    }
    
    // 授权LP代币
    err = s.approveToken(ownerAddress, pairAddress, JustSwapRouterAddress, liquidity)
    if err != nil {
        return nil, err
    }
    
    // 构建移除流动性参数
    params := s.buildRemoveLiquidityParams(
        tokenA, tokenB,
        liquidity,
        amountAMin, amountBMin,
        ownerAddress, deadline,
    )
    
    // 调用JustSwap路由合约
    return s.client.grpcClient.TriggerContract(
        ownerAddress,
        JustSwapRouterAddress,
        "removeLiquidity(address,address,uint256,uint256,uint256,address,uint256)",
        params,
        1000000, // feeLimit
        0, // callValue
    )
}

// 辅助函数
func (s *JustSwapService) approveToken(
    ownerAddress string,
    tokenAddress string,
    spenderAddress string,
    amount *big.Int,
) error {
    // 构建approve参数
    spender, _ := client.Base58ToAddress(spenderAddress)
    params := hex.EncodeToString(spender.Bytes()) + 
             hex.EncodeToString(common.LeftPadBytes(amount.Bytes(), 32))
    
    // 调用代币合约的approve方法
    _, err := s.client.grpcClient.TriggerContract(
        ownerAddress,
        tokenAddress,
        "approve(address,uint256)",
        params,
        1000000, // feeLimit
        0, // callValue
    )
    
    return err
}

func (s *JustSwapService) buildSwapParams(
    amount *big.Int,
    path []string,
    to string,
    deadline *big.Int,
) string {
    // 构建交换参数 (简化实现)
    params := hex.EncodeToString(common.LeftPadBytes(amount.Bytes(), 32))
    
    // 添加路径
    for _, addr := range path {
        address, _ := client.Base58ToAddress(addr)
        params += hex.EncodeToString(address.Bytes())
    }
    
    // 添加接收者和截止时间
    toAddr, _ := client.Base58ToAddress(to)
    params += hex.EncodeToString(toAddr.Bytes())
    params += hex.EncodeToString(common.LeftPadBytes(deadline.Bytes(), 32))
    
    return params
}

func (s *JustSwapService) buildAddLiquidityParams(
    tokenA, tokenB string,
    amountADesired, amountBDesired *big.Int,
    amountAMin, amountBMin *big.Int,
    to string,
    deadline *big.Int,
) string {
    // 构建添加流动性参数 (简化实现)
    params := ""
    
    // 添加代币地址
    addrA, _ := client.Base58ToAddress(tokenA)
    addrB, _ := client.Base58ToAddress(tokenB)
    params += hex.EncodeToString(addrA.Bytes())
    params += hex.EncodeToString(addrB.Bytes())
    
    // 添加数量参数
    params += hex.EncodeToString(common.LeftPadBytes(amountADesired.Bytes(), 32))
    params += hex.EncodeToString(common.LeftPadBytes(amountBDesired.Bytes(), 32))
    params += hex.EncodeToString(common.LeftPadBytes(amountAMin.Bytes(), 32))
    params += hex.EncodeToString(common.LeftPadBytes(amountBMin.Bytes(), 32))
    
    // 添加接收者和截止时间
    toAddr, _ := client.Base58ToAddress(to)
    params += hex.EncodeToString(toAddr.Bytes())
    params += hex.EncodeToString(common.LeftPadBytes(deadline.Bytes(), 32))
    
    return params
}

func (s *JustSwapService) buildRemoveLiquidityParams(
    tokenA, tokenB string,
    liquidity *big.Int,
    amountAMin, amountBMin *big.Int,
    to string,
    deadline *big.Int,
) string {
    // 构建移除流动性参数 (简化实现)
    params := ""
    
    // 添加代币地址
    addrA, _ := client.Base58ToAddress(tokenA)
    addrB, _ := client.Base58ToAddress(tokenB)
    params += hex.EncodeToString(addrA.Bytes())
    params += hex.EncodeToString(addrB.Bytes())
    
    // 添加流动性和最小数量
    params += hex.EncodeToString(common.LeftPadBytes(liquidity.Bytes(), 32))
    params += hex.EncodeToString(common.LeftPadBytes(amountAMin.Bytes(), 32))
    params += hex.EncodeToString(common.LeftPadBytes(amountBMin.Bytes(), 32))
    
    // 添加接收者和截止时间
    toAddr, _ := client.Base58ToAddress(to)
    params += hex.EncodeToString(toAddr.Bytes())
    params += hex.EncodeToString(common.LeftPadBytes(deadline.Bytes(), 32))
    
    return params
}

func (s *JustSwapService) getPairAddress(tokenA, tokenB string) (string, error) {
    // 查询交易对地址 (简化实现)
    // 实际需要调用Factory合约的getPair方法
    return "TYour_LP_Token_Address_Here", nil
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
    
    "github.com/shopspring/decimal"
    
    "your-project/client"
    "your-project/services"
)

func main() {
    // 创建Tron客户端 (使用主网)
    tronClient, err := client.NewTronClient(&client.MainnetInfo)
    if err != nil {
        log.Fatal("创建Tron客户端失败:", err)
    }
    
    // 用户私钥和地址
    privateKey := "your_private_key_here"
    userAddress := "TYour_Address_Here"
    
    // 创建服务
    contractService := services.NewTronContractService(tronClient, privateKey)
    justSwapService := services.NewJustSwapService(tronClient, privateKey)
    
    // 1. 获取网络信息
    networkInfo := tronClient.GetNetworkInfo()
    fmt.Printf("=== Tron网络信息 ===\n")
    fmt.Printf("网络名称: %s\n", networkInfo.NetworkName)
    fmt.Printf("链ID: %s\n", networkInfo.ChainID)
    fmt.Printf("浏览器: %s\n", networkInfo.ExplorerURL)
    fmt.Printf("原生代币: %s\n", networkInfo.Currency)
    fmt.Printf("精度: %d\n", networkInfo.Decimals)
    
    // 2. 获取账户信息
    fmt.Printf("\n=== 账户信息 ===\n")
    
    account, err := tronClient.GetAccount(userAddress)
    if err != nil {
        log.Fatal("获取账户信息失败:", err)
    }
    
    fmt.Printf("地址: %s\n", account.Address)
    fmt.Printf("TRX余额: %s (%.6f TRX)\n", 
        account.Balance.String(), 
        decimal.NewFromBigInt(account.Balance, -6).InexactFloat64())
    fmt.Printf("冻结余额: %s TRX\n", 
        decimal.NewFromBigInt(account.FrozenBalance, -6).String())
    fmt.Printf("投票权力: %s\n", account.VotingPower.String())
    
    // 3. 获取TRC20代币余额
    fmt.Printf("\n=== TRC20代币余额 ===\n")
    
    tokens := []struct {
        Address string
        Symbol  string
        Decimals int
    }{
        {client.USDTAddress, "USDT", 6},
        {client.USDCAddress, "USDC", 6},
        {client.SUNAddress, "SUN", 18},
        {client.BTTAddress, "BTT", 18},
        {client.WINAddress, "WIN", 6},
    }
    
    for _, token := range tokens {
        balance, err := tronClient.GetTRC20Balance(token.Address, userAddress)
        if err != nil {
            log.Printf("获取%s余额失败: %v", token.Symbol, err)
            continue
        }
        
        fmt.Printf("%s余额: %s (%.6f %s)\n", 
            token.Symbol,
            balance.String(),
            decimal.NewFromBigInt(balance, -token.Decimals).InexactFloat64(),
            token.Symbol)
    }
    
    // 4. 获取账户资源信息
    fmt.Printf("\n=== 账户资源信息 ===\n")
    
    resource := account.AccountResource
    fmt.Printf("带宽信息:\n")
    fmt.Printf("  免费带宽已用: %s\n", resource.FreeNetUsed.String())
    fmt.Printf("  免费带宽限制: %s\n", resource.FreeNetLimit.String())
    fmt.Printf("  质押带宽已用: %s\n", resource.NetUsed.String())
    fmt.Printf("  质押带宽限制: %s\n", resource.NetLimit.String())
    
    fmt.Printf("能量信息:\n")
    fmt.Printf("  能量已用: %s\n", resource.EnergyUsed.String())
    fmt.Printf("  能量限制: %s\n", resource.EnergyLimit.String())
    
    // 5. 获取超级代表信息
    fmt.Printf("\n=== 超级代表信息 ===\n")
    
    srs, err := tronClient.GetSuperRepresentatives()
    if err != nil {
        log.Printf("获取超级代表失败: %v", err)
    } else {
        fmt.Printf("前5名超级代表:\n")
        for i, sr := range srs[:5] {
            fmt.Printf("  %d. %s\n", i+1, sr.Address)
            fmt.Printf("     URL: %s\n", sr.URL)
            fmt.Printf("     得票数: %s\n", sr.VoteCount.String())
            fmt.Printf("     生产区块: %s\n", sr.TotalProduced.String())
            fmt.Printf("     错过区块: %s\n", sr.TotalMissed.String())
        }
    }
    
    // 6. TRX转账示例
    fmt.Printf("\n=== TRX转账示例 ===\n")
    
    if account.Balance.Cmp(big.NewInt(10e6)) > 0 { // 如果余额大于10 TRX
        transferAmount := big.NewInt(1e6) // 转账1 TRX
        recipientAddress := "TRecipient_Address_Here"
        
        fmt.Printf("准备转账 %s TRX 到 %s\n", 
            decimal.NewFromBigInt(transferAmount, -6).String(),
            recipientAddress)
        
        // 执行转账
        tx, err := tronClient.SendTRX(userAddress, recipientAddress, transferAmount, privateKey)
        if err != nil {
            log.Printf("TRX转账失败: %v", err)
        } else {
            fmt.Printf("转账交易已提交: %s\n", tx.TxID)
        }
    } else {
        fmt.Printf("余额不足，无法进行转账示例\n")
    }
    
    // 7. TRC20代币转账示例
    fmt.Printf("\n=== TRC20代币转账示例 ===\n")
    
    usdtBalance, err := tronClient.GetTRC20Balance(client.USDTAddress, userAddress)
    if err != nil {
        log.Printf("获取USDT余额失败: %v", err)
    } else if usdtBalance.Cmp(big.NewInt(10e6)) > 0 { // 如果USDT余额大于10
        transferAmount := big.NewInt(1e6) // 转账1 USDT
        recipientAddress := "TRecipient_Address_Here"
        
        fmt.Printf("准备转账 %s USDT 到 %s\n", 
            decimal.NewFromBigInt(transferAmount, -6).String(),
            recipientAddress)
        
        // 执行USDT转账
        tx, err := tronClient.SendTRC20(
            client.USDTAddress,
            userAddress,
            recipientAddress,
            transferAmount,
            privateKey,
        )
        if err != nil {
            log.Printf("USDT转账失败: %v", err)
        } else {
            fmt.Printf("USDT转账交易已提交: %s\n", tx.TxID)
        }
    } else {
        fmt.Printf("USDT余额不足，无法进行转账示例\n")
    }
    
    // 8. JustSwap交易示例
    fmt.Printf("\n=== JustSwap交易示例 ===\n")
    
    if account.Balance.Cmp(big.NewInt(100e6)) > 0 { // 如果TRX余额大于100
        swapAmount := big.NewInt(10e6) // 用10 TRX换USDT
        minOutput := big.NewInt(0)     // 最小输出 (实际应该计算)
        deadline := big.NewInt(time.Now().Unix() + 1200) // 20分钟后过期
        
        fmt.Printf("准备用 %s TRX 换取 USDT\n", 
            decimal.NewFromBigInt(swapAmount, -6).String())
        
        // 执行TRX换USDT
        tx, err := justSwapService.SwapTRXForTokens(
            userAddress,
            client.USDTAddress,
            swapAmount,
            minOutput,
            deadline,
        )
        if err != nil {
            log.Printf("JustSwap交换失败: %v", err)
        } else {
            fmt.Printf("JustSwap交换交易已提交: %s\n", tx.TxID)
        }
    } else {
        fmt.Printf("TRX余额不足，无法进行JustSwap交易示例\n")
    }
    
    // 9. 资源管理示例
    fmt.Printf("\n=== 资源管理示例 ===\n")
    
    if account.Balance.Cmp(big.NewInt(1000e6)) > 0 { // 如果余额大于1000 TRX
        freezeAmount := big.NewInt(100e6) // 冻结100 TRX
        freezeDuration := int64(3)        // 冻结3天
        
        fmt.Printf("准备冻结 %s TRX 获取能量\n", 
            decimal.NewFromBigInt(freezeAmount, -6).String())
        
        // 冻结TRX获取能量
        tx, err := tronClient.FreezeBalance(
            userAddress,
            freezeAmount,
            freezeDuration,
            "ENERGY",
            privateKey,
        )
        if err != nil {
            log.Printf("冻结TRX失败: %v", err)
        } else {
            fmt.Printf("冻结交易已提交: %s\n", tx.TxID)
        }
    } else {
        fmt.Printf("余额不足，无法进行冻结示例\n")
    }
    
    // 10. Tron特性总结
    fmt.Printf("\n=== Tron特性总结 ===\n")
    
    fmt.Printf("Tron优势:\n")
    fmt.Printf("  - 高吞吐量 (2000+ TPS)\n")
    fmt.Printf("  - 低交易费用\n")
    fmt.Printf("  - 丰富的DeFi生态\n")
    fmt.Printf("  - 强大的DApp支持\n")
    fmt.Printf("  - 活跃的社区治理\n")
    
    fmt.Printf("\n主要DeFi协议:\n")
    fmt.Printf("  - JustSwap: DEX交易\n")
    fmt.Printf("  - JustLend: 借贷协议\n")
    fmt.Printf("  - SUN.io: 挖矿平台\n")
    fmt.Printf("  - WINk: 游戏平台\n")
    
    fmt.Printf("\n资源系统:\n")
    fmt.Printf("  - 带宽: 转账和合约调用\n")
    fmt.Printf("  - 能量: 智能合约执行\n")
    fmt.Printf("  - 冻结机制: 获取资源\n")
    fmt.Printf("  - 投票权: 参与治理\n")
    
    // 11. 最佳实践建议
    fmt.Printf("\n=== 最佳实践建议 ===\n")
    
    fmt.Printf("使用Tron时请注意:\n")
    fmt.Printf("  1. 合理管理带宽和能量资源\n")
    fmt.Printf("  2. 冻结TRX获取免费资源\n")
    fmt.Printf("  3. 参与超级代表投票获得奖励\n")
    fmt.Printf("  4. 关注智能合约的能量消耗\n")
    fmt.Printf("  5. 使用TRC20代币时预留能量\n")
    fmt.Printf("  6. 定期检查账户资源使用情况\n")
    fmt.Printf("  7. 了解不同DeFi协议的风险\n")
    fmt.Printf("  8. 保管好私钥和助记词\n")
}
```

这个Tron使用指南提供了完整的高性能区块链集成方案，涵盖了账户管理、资源系统、智能合约、DeFi生态、治理机制等核心功能，是构建Tron DApp和DeFi应用的重要参考文档。
