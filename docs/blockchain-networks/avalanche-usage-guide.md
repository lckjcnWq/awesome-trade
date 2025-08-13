# Avalanche 区块链 详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [环境准备](#环境准备)
3. [三链架构](#三链架构)
4. [智能合约部署](#智能合约部署)
5. [跨链操作](#跨链操作)
6. [子网开发](#子网开发)
7. [DeFi集成](#defi集成)
8. [性能优化](#性能优化)
9. [实际应用](#实际应用)

## 基础概念

### 1.1 Avalanche 简介

Avalanche 是高性能、可扩展的区块链平台，采用独特的三链架构和雪崩共识机制，支持智能合约、DeFi应用和自定义子网。

```bash
# 安装Avalanche相关依赖
go get github.com/ava-labs/avalanchego
go get github.com/ava-labs/coreth
go get github.com/ethereum/go-ethereum
go get github.com/shopspring/decimal
```

### 1.2 核心概念

```go
// 主要包导入
import (
    "context"
    "math/big"
    "strings"
    
    "github.com/ava-labs/avalanchego/ids"
    "github.com/ava-labs/avalanchego/utils/formatting"
    "github.com/ava-labs/avalanchego/vms/avm"
    "github.com/ava-labs/avalanchego/vms/platformvm"
    "github.com/ava-labs/coreth/ethclient"
    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/common"
    "github.com/shopspring/decimal"
)

// Avalanche 网络配置
var (
    // Mainnet
    MainnetRPC = "https://api.avax.network/ext/bc/C/rpc"
    MainnetChainID = big.NewInt(43114)
    
    // Testnet (Fuji)
    TestnetRPC = "https://api.avax-test.network/ext/bc/C/rpc"
    TestnetChainID = big.NewInt(43113)
    
    // 本地网络
    LocalRPC = "http://localhost:9650/ext/bc/C/rpc"
    LocalChainID = big.NewInt(43112)
)

// 链ID常量
const (
    XChainID = "2oYMqpc5uqyyQBNPxuYbp8YxMcUdoVGPJzqeqHmyc6e4TtHF4K" // X-Chain
    PChainID = "11111111111111111111111111111111LpoYY"           // P-Chain
    CChainID = "2q9e4r6Mu3U68nU1fYjgbR6JvwrRx36CohpAX5UQxse55x1Q5"  // C-Chain
)

// 资产信息
type Asset struct {
    ID       ids.ID
    Symbol   string
    Name     string
    Decimals uint8
    Balance  *big.Int
}

// AVAX资产
var AVAXAsset = Asset{
    Symbol:   "AVAX",
    Name:     "Avalanche",
    Decimals: 18,
}

// 网络信息
type NetworkInfo struct {
    NetworkID   uint32
    NetworkName string
    ChainID     *big.Int
    RPCEndpoint string
    WSEndpoint  string
    Explorer    string
}

// 主网配置
var MainnetInfo = NetworkInfo{
    NetworkID:   1,
    NetworkName: "Avalanche Mainnet",
    ChainID:     MainnetChainID,
    RPCEndpoint: MainnetRPC,
    WSEndpoint:  "wss://api.avax.network/ext/bc/C/ws",
    Explorer:    "https://snowtrace.io",
}

// 测试网配置
var TestnetInfo = NetworkInfo{
    NetworkID:   5,
    NetworkName: "Avalanche Fuji Testnet",
    ChainID:     TestnetChainID,
    RPCEndpoint: TestnetRPC,
    WSEndpoint:  "wss://api.avax-test.network/ext/bc/C/ws",
    Explorer:    "https://testnet.snowtrace.io",
}

// 验证者信息
type Validator struct {
    NodeID        ids.NodeID
    StartTime     uint64
    EndTime       uint64
    StakeAmount   *big.Int
    RewardOwner   *platformvm.Owner
    PotentialReward *big.Int
    DelegationFee uint32
    Uptime        float64
    Connected     bool
}

// 委托信息
type Delegation struct {
    TxID        ids.ID
    NodeID      ids.NodeID
    StartTime   uint64
    EndTime     uint64
    StakeAmount *big.Int
    RewardOwner *platformvm.Owner
    PotentialReward *big.Int
}
```

## 环境准备

### 2.1 客户端设置

```go
// client/avalanche_client.go
package client

import (
    "context"
    "math/big"
    
    "github.com/ava-labs/avalanchego/api"
    "github.com/ava-labs/avalanchego/api/info"
    "github.com/ava-labs/avalanchego/vms/avm"
    "github.com/ava-labs/avalanchego/vms/platformvm"
    "github.com/ava-labs/coreth/ethclient"
    "github.com/ethereum/go-ethereum/common"
)

type AvalancheClient struct {
    // C-Chain (EVM兼容)
    cChainClient *ethclient.Client
    
    // X-Chain (资产交换)
    xChainClient avm.Client
    
    // P-Chain (平台链)
    pChainClient platformvm.Client
    
    // 信息API
    infoClient info.Client
    
    networkInfo *NetworkInfo
}

func NewAvalancheClient(networkInfo *NetworkInfo) (*AvalancheClient, error) {
    // 创建C-Chain客户端 (EVM兼容)
    cChainClient, err := ethclient.Dial(networkInfo.RPCEndpoint)
    if err != nil {
        return nil, err
    }
    
    // 创建API客户端
    apiClient := api.NewClient(networkInfo.RPCEndpoint, "")
    
    // 创建X-Chain客户端
    xChainClient := avm.NewClient(apiClient, "X")
    
    // 创建P-Chain客户端
    pChainClient := platformvm.NewClient(apiClient, "P")
    
    // 创建信息客户端
    infoClient := info.NewClient(apiClient)
    
    return &AvalancheClient{
        cChainClient: cChainClient,
        xChainClient: xChainClient,
        pChainClient: pChainClient,
        infoClient:   infoClient,
        networkInfo:  networkInfo,
    }, nil
}

// 获取网络信息
func (c *AvalancheClient) GetNetworkInfo() (*NetworkInfo, error) {
    return c.networkInfo, nil
}

// 获取C-Chain客户端
func (c *AvalancheClient) CChain() *ethclient.Client {
    return c.cChainClient
}

// 获取X-Chain客户端
func (c *AvalancheClient) XChain() avm.Client {
    return c.xChainClient
}

// 获取P-Chain客户端
func (c *AvalancheClient) PChain() platformvm.Client {
    return c.pChainClient
}

// 获取节点信息
func (c *AvalancheClient) GetNodeInfo() (*info.GetNodeInfoReply, error) {
    ctx := context.Background()
    return c.infoClient.GetNodeInfo(ctx)
}

// 获取网络名称
func (c *AvalancheClient) GetNetworkName() (string, error) {
    ctx := context.Background()
    return c.infoClient.GetNetworkName(ctx)
}

// 获取区块链信息
func (c *AvalancheClient) GetBlockchains() ([]info.Blockchain, error) {
    ctx := context.Background()
    return c.infoClient.GetBlockchains(ctx)
}
```

## 三链架构

### 3.1 X-Chain 操作

```go
// services/xchain_service.go
package services

import (
    "context"
    "math/big"
    
    "github.com/ava-labs/avalanchego/ids"
    "github.com/ava-labs/avalanchego/vms/avm"
    "github.com/ava-labs/avalanchego/vms/secp256k1fx"
    "github.com/shopspring/decimal"
)

type XChainService struct {
    client *AvalancheClient
}

func NewXChainService(client *AvalancheClient) *XChainService {
    return &XChainService{
        client: client,
    }
}

// 获取余额
func (s *XChainService) GetBalance(address string, assetID ids.ID) (*big.Int, error) {
    ctx := context.Background()
    
    balance, err := s.client.XChain().GetBalance(ctx, address, assetID.String())
    if err != nil {
        return nil, err
    }
    
    return big.NewInt(int64(balance)), nil
}

// 获取所有余额
func (s *XChainService) GetAllBalances(address string) (map[ids.ID]*big.Int, error) {
    ctx := context.Background()
    
    balances, err := s.client.XChain().GetAllBalances(ctx, address)
    if err != nil {
        return nil, err
    }
    
    result := make(map[ids.ID]*big.Int)
    for assetID, balance := range balances {
        result[assetID] = big.NewInt(int64(balance))
    }
    
    return result, nil
}

// 发送资产
func (s *XChainService) Send(
    from string,
    to string,
    amount *big.Int,
    assetID ids.ID,
    memo string,
) (ids.ID, error) {
    ctx := context.Background()
    
    // 构建输出
    outputs := []*avm.TransferableOutput{
        {
            Asset: avm.Asset{ID: assetID},
            Out: &secp256k1fx.TransferOutput{
                Amt: amount.Uint64(),
                OutputOwners: secp256k1fx.OutputOwners{
                    Threshold: 1,
                    Addrs:     []ids.ShortID{}, // 需要解析to地址
                },
            },
        },
    }
    
    // 发送交易
    txID, err := s.client.XChain().Send(
        ctx,
        from,
        outputs,
        memo,
    )
    if err != nil {
        return ids.Empty, err
    }
    
    return txID, nil
}

// 创建资产
func (s *XChainService) CreateAsset(
    name string,
    symbol string,
    denomination uint8,
    initialHolders map[string]*big.Int,
    minterSets [][]string,
) (ids.ID, error) {
    ctx := context.Background()
    
    // 构建初始状态
    var initialStates []*avm.InitialState
    
    // 添加初始持有者
    for address, amount := range initialHolders {
        initialState := &avm.InitialState{
            FxIndex: 0, // secp256k1fx
            Outs: []avm.TransferableOut{
                &avm.TransferableOutput{
                    Asset: avm.Asset{}, // 将在创建时设置
                    Out: &secp256k1fx.TransferOutput{
                        Amt: amount.Uint64(),
                        OutputOwners: secp256k1fx.OutputOwners{
                            Threshold: 1,
                            Addrs:     []ids.ShortID{}, // 需要解析地址
                        },
                    },
                },
            },
        }
        initialStates = append(initialStates, initialState)
    }
    
    // 创建资产
    assetID, err := s.client.XChain().CreateAsset(
        ctx,
        name,
        symbol,
        denomination,
        initialStates,
    )
    if err != nil {
        return ids.Empty, err
    }
    
    return assetID, nil
}

// 铸造资产
func (s *XChainService) MintAsset(
    assetID ids.ID,
    amount *big.Int,
    to string,
) (ids.ID, error) {
    ctx := context.Background()
    
    // 构建铸造输出
    outputs := []*avm.TransferableOutput{
        {
            Asset: avm.Asset{ID: assetID},
            Out: &secp256k1fx.TransferOutput{
                Amt: amount.Uint64(),
                OutputOwners: secp256k1fx.OutputOwners{
                    Threshold: 1,
                    Addrs:     []ids.ShortID{}, // 需要解析to地址
                },
            },
        },
    }
    
    // 执行铸造
    txID, err := s.client.XChain().Mint(
        ctx,
        outputs,
        assetID,
    )
    if err != nil {
        return ids.Empty, err
    }
    
    return txID, nil
}

// 获取交易状态
func (s *XChainService) GetTxStatus(txID ids.ID) (avm.Status, error) {
    ctx := context.Background()
    return s.client.XChain().GetTxStatus(ctx, txID)
}

// 获取UTXO
func (s *XChainService) GetUTXOs(addresses []string, limit uint32) ([]*avm.UTXO, error) {
    ctx := context.Background()
    
    utxos, _, err := s.client.XChain().GetUTXOs(
        ctx,
        addresses,
        limit,
        ids.Empty, // startIndex
        ids.Empty, // sourceChain
    )
    if err != nil {
        return nil, err
    }
    
    return utxos, nil
}
```

### 3.2 P-Chain 操作

```go
// services/pchain_service.go
package services

import (
    "context"
    "math/big"
    "time"
    
    "github.com/ava-labs/avalanchego/ids"
    "github.com/ava-labs/avalanchego/vms/platformvm"
)

type PChainService struct {
    client *AvalancheClient
}

func NewPChainService(client *AvalancheClient) *PChainService {
    return &PChainService{
        client: client,
    }
}

// 获取余额
func (s *PChainService) GetBalance(address string) (*big.Int, error) {
    ctx := context.Background()
    
    balance, err := s.client.PChain().GetBalance(ctx, address)
    if err != nil {
        return nil, err
    }
    
    return big.NewInt(int64(balance)), nil
}

// 添加验证者
func (s *PChainService) AddValidator(
    nodeID ids.NodeID,
    startTime time.Time,
    endTime time.Time,
    stakeAmount *big.Int,
    rewardAddress string,
    delegationFeeRate uint32,
) (ids.ID, error) {
    ctx := context.Background()
    
    // 构建验证者
    validator := &platformvm.Validator{
        NodeID: nodeID,
        Start:  uint64(startTime.Unix()),
        End:    uint64(endTime.Unix()),
        Wght:   stakeAmount.Uint64(),
    }
    
    // 添加验证者
    txID, err := s.client.PChain().AddValidator(
        ctx,
        validator,
        rewardAddress,
        delegationFeeRate,
    )
    if err != nil {
        return ids.Empty, err
    }
    
    return txID, nil
}

// 添加委托者
func (s *PChainService) AddDelegator(
    nodeID ids.NodeID,
    startTime time.Time,
    endTime time.Time,
    stakeAmount *big.Int,
    rewardAddress string,
) (ids.ID, error) {
    ctx := context.Background()
    
    // 构建委托者
    delegator := &platformvm.Validator{
        NodeID: nodeID,
        Start:  uint64(startTime.Unix()),
        End:    uint64(endTime.Unix()),
        Wght:   stakeAmount.Uint64(),
    }
    
    // 添加委托者
    txID, err := s.client.PChain().AddDelegator(
        ctx,
        delegator,
        rewardAddress,
    )
    if err != nil {
        return ids.Empty, err
    }
    
    return txID, nil
}

// 获取验证者列表
func (s *PChainService) GetCurrentValidators() ([]*Validator, error) {
    ctx := context.Background()
    
    validators, err := s.client.PChain().GetCurrentValidators(ctx, ids.Empty)
    if err != nil {
        return nil, err
    }
    
    var result []*Validator
    for _, v := range validators {
        validator := &Validator{
            NodeID:      v.NodeID,
            StartTime:   v.StartTime,
            EndTime:     v.EndTime,
            StakeAmount: big.NewInt(int64(v.Weight)),
            DelegationFee: v.DelegationFee,
            Uptime:      v.Uptime,
            Connected:   v.Connected,
        }
        result = append(result, validator)
    }
    
    return result, nil
}

// 获取待处理验证者
func (s *PChainService) GetPendingValidators() ([]*Validator, error) {
    ctx := context.Background()
    
    validators, err := s.client.PChain().GetPendingValidators(ctx, ids.Empty)
    if err != nil {
        return nil, err
    }
    
    var result []*Validator
    for _, v := range validators {
        validator := &Validator{
            NodeID:      v.NodeID,
            StartTime:   v.StartTime,
            EndTime:     v.EndTime,
            StakeAmount: big.NewInt(int64(v.Weight)),
            DelegationFee: v.DelegationFee,
        }
        result = append(result, validator)
    }
    
    return result, nil
}

// 获取委托信息
func (s *PChainService) GetStake(address string) ([]*Delegation, error) {
    ctx := context.Background()
    
    stakeInfo, err := s.client.PChain().GetStake(ctx, address)
    if err != nil {
        return nil, err
    }
    
    var delegations []*Delegation
    for _, stake := range stakeInfo.Stakes {
        delegation := &Delegation{
            TxID:        stake.TxID,
            StartTime:   stake.StartTime,
            EndTime:     stake.EndTime,
            StakeAmount: big.NewInt(int64(stake.Amount)),
        }
        delegations = append(delegations, delegation)
    }
    
    return delegations, nil
}

// 创建子网
func (s *PChainService) CreateSubnet(controlKeys []string, threshold uint32) (ids.ID, error) {
    ctx := context.Background()
    
    // 转换控制密钥
    var owners []ids.ShortID
    for _, key := range controlKeys {
        // 需要将字符串转换为ShortID
        // 这里简化处理
    }
    
    // 创建子网
    subnetID, err := s.client.PChain().CreateSubnet(
        ctx,
        owners,
        threshold,
    )
    if err != nil {
        return ids.Empty, err
    }
    
    return subnetID, nil
}

// 获取子网列表
func (s *PChainService) GetSubnets() ([]platformvm.APISubnet, error) {
    ctx := context.Background()
    return s.client.PChain().GetSubnets(ctx, nil)
}

// 导入/导出操作
func (s *PChainService) ImportAVAX(
    sourceChain string,
    to string,
    amount *big.Int,
) (ids.ID, error) {
    ctx := context.Background()
    
    // 从源链导入AVAX到P-Chain
    txID, err := s.client.PChain().ImportAVAX(
        ctx,
        to,
        sourceChain,
    )
    if err != nil {
        return ids.Empty, err
    }
    
    return txID, nil
}

func (s *PChainService) ExportAVAX(
    to string,
    destinationChain string,
    amount *big.Int,
) (ids.ID, error) {
    ctx := context.Background()
    
    // 从P-Chain导出AVAX到目标链
    txID, err := s.client.PChain().ExportAVAX(
        ctx,
        amount.Uint64(),
        destinationChain,
        to,
    )
    if err != nil {
        return ids.Empty, err
    }
    
    return txID, nil
}
```

### 3.3 C-Chain 操作

```go
// services/cchain_service.go
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

type CChainService struct {
    client     *AvalancheClient
    privateKey *ecdsa.PrivateKey
}

func NewCChainService(client *AvalancheClient, privateKey *ecdsa.PrivateKey) *CChainService {
    return &CChainService{
        client:     client,
        privateKey: privateKey,
    }
}

// 获取余额
func (s *CChainService) GetBalance(address common.Address) (*big.Int, error) {
    ctx := context.Background()
    return s.client.CChain().BalanceAt(ctx, address, nil)
}

// 发送AVAX
func (s *CChainService) SendAVAX(to common.Address, amount *big.Int) (*types.Transaction, error) {
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 获取nonce
    nonce, err := s.client.CChain().PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        return nil, err
    }
    
    // 获取gas价格
    gasPrice, err := s.client.CChain().SuggestGasPrice(context.Background())
    if err != nil {
        return nil, err
    }
    
    // 创建交易
    tx := types.NewTransaction(nonce, to, amount, 21000, gasPrice, nil)
    
    // 签名交易
    chainID, err := s.client.CChain().NetworkID(context.Background())
    if err != nil {
        return nil, err
    }
    
    signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), s.privateKey)
    if err != nil {
        return nil, err
    }
    
    // 发送交易
    err = s.client.CChain().SendTransaction(context.Background(), signedTx)
    if err != nil {
        return nil, err
    }
    
    return signedTx, nil
}

// 部署合约
func (s *CChainService) DeployContract(
    contractABI string,
    contractBytecode string,
    constructorParams ...interface{},
) (common.Address, *types.Transaction, error) {
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 解析ABI
    parsedABI, err := abi.JSON(strings.NewReader(contractABI))
    if err != nil {
        return common.Address{}, nil, err
    }
    
    // 获取认证器
    auth, err := bind.NewKeyedTransactorWithChainID(s.privateKey, s.client.networkInfo.ChainID)
    if err != nil {
        return common.Address{}, nil, err
    }
    
    // 部署合约
    address, tx, _, err := bind.DeployContract(
        auth,
        parsedABI,
        common.FromHex(contractBytecode),
        s.client.CChain(),
        constructorParams...,
    )
    if err != nil {
        return common.Address{}, nil, err
    }
    
    return address, tx, nil
}

// 调用合约方法
func (s *CChainService) CallContract(
    contractAddress common.Address,
    contractABI string,
    methodName string,
    params ...interface{},
) (*types.Transaction, error) {
    // 解析ABI
    parsedABI, err := abi.JSON(strings.NewReader(contractABI))
    if err != nil {
        return nil, err
    }
    
    // 获取认证器
    auth, err := bind.NewKeyedTransactorWithChainID(s.privateKey, s.client.networkInfo.ChainID)
    if err != nil {
        return nil, err
    }
    
    // 创建合约实例
    contract := bind.NewBoundContract(contractAddress, parsedABI, s.client.CChain(), s.client.CChain(), s.client.CChain())
    
    // 调用方法
    tx, err := contract.Transact(auth, methodName, params...)
    if err != nil {
        return nil, err
    }
    
    return tx, nil
}

// 查询合约状态
func (s *CChainService) QueryContract(
    contractAddress common.Address,
    contractABI string,
    methodName string,
    result interface{},
    params ...interface{},
) error {
    // 解析ABI
    parsedABI, err := abi.JSON(strings.NewReader(contractABI))
    if err != nil {
        return err
    }
    
    // 创建调用选项
    callOpts := &bind.CallOpts{
        Context: context.Background(),
    }
    
    // 创建合约实例
    contract := bind.NewBoundContract(contractAddress, parsedABI, s.client.CChain(), s.client.CChain(), s.client.CChain())
    
    // 查询方法
    err = contract.Call(callOpts, result, methodName, params...)
    if err != nil {
        return err
    }
    
    return nil
}

// 获取交易收据
func (s *CChainService) GetTransactionReceipt(txHash common.Hash) (*types.Receipt, error) {
    ctx := context.Background()
    return s.client.CChain().TransactionReceipt(ctx, txHash)
}

// 等待交易确认
func (s *CChainService) WaitForTransaction(txHash common.Hash) (*types.Receipt, error) {
    ctx := context.Background()
    
    for {
        receipt, err := s.client.CChain().TransactionReceipt(ctx, txHash)
        if err == nil {
            return receipt, nil
        }
        
        // 等待一段时间后重试
        time.Sleep(2 * time.Second)
    }
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
    "time"
    
    "github.com/ava-labs/avalanchego/ids"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/crypto"
    
    "your-project/client"
    "your-project/services"
)

func main() {
    // 创建Avalanche客户端 (使用测试网)
    avalancheClient, err := client.NewAvalancheClient(&client.TestnetInfo)
    if err != nil {
        log.Fatal("创建Avalanche客户端失败:", err)
    }
    
    // 加载私钥
    privateKey, err := crypto.HexToECDSA("your_private_key_here")
    if err != nil {
        log.Fatal("加载私钥失败:", err)
    }
    
    // 获取地址
    publicKey := privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    address := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 创建服务
    xChainService := services.NewXChainService(avalancheClient)
    pChainService := services.NewPChainService(avalancheClient)
    cChainService := services.NewCChainService(avalancheClient, privateKey)
    
    // 1. 获取网络信息
    networkInfo, err := avalancheClient.GetNetworkInfo()
    if err != nil {
        log.Fatal("获取网络信息失败:", err)
    }
    
    fmt.Printf("网络信息:\n")
    fmt.Printf("  网络名称: %s\n", networkInfo.NetworkName)
    fmt.Printf("  链ID: %s\n", networkInfo.ChainID.String())
    fmt.Printf("  RPC端点: %s\n", networkInfo.RPCEndpoint)
    
    // 2. 获取节点信息
    nodeInfo, err := avalancheClient.GetNodeInfo()
    if err != nil {
        log.Fatal("获取节点信息失败:", err)
    }
    
    fmt.Printf("节点信息:\n")
    fmt.Printf("  节点ID: %s\n", nodeInfo.ID)
    fmt.Printf("  版本: %s\n", nodeInfo.Version)
    
    // 3. C-Chain操作示例
    fmt.Printf("\n=== C-Chain操作 ===\n")
    
    // 获取AVAX余额
    balance, err := cChainService.GetBalance(address)
    if err != nil {
        log.Fatal("获取C-Chain余额失败:", err)
    }
    
    fmt.Printf("C-Chain AVAX余额: %s\n", balance.String())
    
    // 发送AVAX
    if balance.Cmp(big.NewInt(1e18)) > 0 { // 如果余额大于1 AVAX
        recipientAddress := common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96c4b4d8b6")
        sendAmount := big.NewInt(1e17) // 0.1 AVAX
        
        tx, err := cChainService.SendAVAX(recipientAddress, sendAmount)
        if err != nil {
            log.Printf("发送AVAX失败: %v", err)
        } else {
            fmt.Printf("AVAX转账交易已提交: %s\n", tx.Hash().Hex())
            
            // 等待交易确认
            receipt, err := cChainService.WaitForTransaction(tx.Hash())
            if err != nil {
                log.Printf("等待交易确认失败: %v", err)
            } else {
                fmt.Printf("交易已确认，区块号: %d\n", receipt.BlockNumber.Uint64())
            }
        }
    }
    
    // 4. P-Chain操作示例
    fmt.Printf("\n=== P-Chain操作 ===\n")
    
    // 获取P-Chain余额
    pBalance, err := pChainService.GetBalance(address.Hex())
    if err != nil {
        log.Printf("获取P-Chain余额失败: %v", err)
    } else {
        fmt.Printf("P-Chain AVAX余额: %s\n", pBalance.String())
    }
    
    // 获取当前验证者
    validators, err := pChainService.GetCurrentValidators()
    if err != nil {
        log.Printf("获取验证者失败: %v", err)
    } else {
        fmt.Printf("当前验证者数量: %d\n", len(validators))
        
        // 显示前5个验证者
        for i, validator := range validators {
            if i >= 5 {
                break
            }
            fmt.Printf("  验证者 %d:\n", i+1)
            fmt.Printf("    节点ID: %s\n", validator.NodeID.String())
            fmt.Printf("    质押金额: %s AVAX\n", validator.StakeAmount.String())
            fmt.Printf("    委托费率: %d%%\n", validator.DelegationFee/10000)
            fmt.Printf("    在线时间: %.2f%%\n", validator.Uptime*100)
        }
    }
    
    // 获取子网信息
    subnets, err := pChainService.GetSubnets()
    if err != nil {
        log.Printf("获取子网失败: %v", err)
    } else {
        fmt.Printf("子网数量: %d\n", len(subnets))
    }
    
    // 5. X-Chain操作示例
    fmt.Printf("\n=== X-Chain操作 ===\n")
    
    // 获取X-Chain余额
    avaxAssetID, _ := ids.FromString("FvwEAhmxKfeiG8SnEvq42hc6whRyY3EFYAvebMqDNDGCgxN5Z")
    xBalance, err := xChainService.GetBalance(address.Hex(), avaxAssetID)
    if err != nil {
        log.Printf("获取X-Chain余额失败: %v", err)
    } else {
        fmt.Printf("X-Chain AVAX余额: %s\n", xBalance.String())
    }
    
    // 获取所有资产余额
    allBalances, err := xChainService.GetAllBalances(address.Hex())
    if err != nil {
        log.Printf("获取所有余额失败: %v", err)
    } else {
        fmt.Printf("X-Chain资产数量: %d\n", len(allBalances))
        for assetID, balance := range allBalances {
            if balance.Cmp(big.NewInt(0)) > 0 {
                fmt.Printf("  资产 %s: %s\n", assetID.String(), balance.String())
            }
        }
    }
    
    // 6. 跨链操作示例
    fmt.Printf("\n=== 跨链操作示例 ===\n")
    
    // 从C-Chain导出到P-Chain
    if balance.Cmp(big.NewInt(2e18)) > 0 { // 如果C-Chain余额大于2 AVAX
        exportAmount := big.NewInt(1e18) // 1 AVAX
        
        // 注意：实际的跨链操作需要更复杂的实现
        fmt.Printf("准备从C-Chain导出 %s AVAX到P-Chain\n", exportAmount.String())
        
        // 这里应该实现实际的导出逻辑
        // txID, err := cChainService.ExportAVAX(address.Hex(), "P", exportAmount)
        // if err != nil {
        //     log.Printf("导出失败: %v", err)
        // } else {
        //     fmt.Printf("导出交易已提交: %s\n", txID.String())
        // }
    }
    
    // 7. 智能合约部署示例
    fmt.Printf("\n=== 智能合约部署示例 ===\n")
    
    // 简单的ERC20合约ABI和字节码
    erc20ABI := `[{"constant":true,"inputs":[],"name":"totalSupply","outputs":[{"name":"","type":"uint256"}],"type":"function"}]`
    erc20Bytecode := "608060405234801561001057600080fd5b50..." // 简化的字节码
    
    // 部署合约
    contractAddress, deployTx, err := cChainService.DeployContract(
        erc20ABI,
        erc20Bytecode,
        "TestToken",     // 代币名称
        "TEST",          // 代币符号
        big.NewInt(18),  // 小数位数
        big.NewInt(1000000e18), // 总供应量
    )
    if err != nil {
        log.Printf("部署合约失败: %v", err)
    } else {
        fmt.Printf("合约部署交易: %s\n", deployTx.Hash().Hex())
        fmt.Printf("合约地址: %s\n", contractAddress.Hex())
    }
    
    // 8. 性能测试
    fmt.Printf("\n=== 性能测试 ===\n")
    
    // 测试交易吞吐量
    startTime := time.Now()
    var successCount int
    
    for i := 0; i < 10; i++ {
        // 发送小额测试交易
        testAmount := big.NewInt(1e15) // 0.001 AVAX
        tx, err := cChainService.SendAVAX(address, testAmount) // 发送给自己
        if err == nil {
            successCount++
            fmt.Printf("  测试交易 %d: %s\n", i+1, tx.Hash().Hex())
        }
        
        time.Sleep(100 * time.Millisecond) // 避免nonce冲突
    }
    
    duration := time.Since(startTime)
    fmt.Printf("性能测试结果:\n")
    fmt.Printf("  成功交易: %d/10\n", successCount)
    fmt.Printf("  总耗时: %v\n", duration)
    fmt.Printf("  平均TPS: %.2f\n", float64(successCount)/duration.Seconds())
}
```

这个Avalanche使用指南提供了完整的高性能区块链集成方案，涵盖了三链架构、智能合约部署、跨链操作、子网开发等核心功能，是多链DeFi应用开发的重要参考文档。
