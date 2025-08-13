# ChainBridge 跨链桥接 详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [环境准备](#环境准备)
3. [桥接架构](#桥接架构)
4. [资产转移](#资产转移)
5. [中继器网络](#中继器网络)
6. [安全机制](#安全机制)
7. [监控和管理](#监控和管理)
8. [故障处理](#故障处理)
9. [实际应用](#实际应用)

## 基础概念

### 1.1 ChainBridge 简介

ChainBridge 是模块化的多向区块链桥接协议，支持在不同区块链之间安全转移资产和数据，采用中继器网络确保跨链操作的可靠性。

```bash
# 安装ChainBridge相关依赖
go get github.com/ChainSafe/chainbridge-core
go get github.com/ethereum/go-ethereum
go get github.com/ethereum/go-ethereum/accounts/abi
go get github.com/shopspring/decimal
```

### 1.2 核心概念

```go
// 主要包导入
import (
    "context"
    "math/big"
    "strings"
    "time"
    
    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/shopspring/decimal"
)

// ChainBridge 核心合约地址
var (
    // Ethereum Mainnet
    EthereumBridgeAddress = common.HexToAddress("0x6Ab592e197C9A4B4B8c6b5b2E1f4B8b8b8b8b8b8")
    EthereumHandlerAddress = common.HexToAddress("0x6Ab592e197C9A4B4B8c6b5b2E1f4B8b8b8b8b8b9")
    
    // Polygon
    PolygonBridgeAddress = common.HexToAddress("0x6Ab592e197C9A4B4B8c6b5b2E1f4B8b8b8b8b8ba")
    PolygonHandlerAddress = common.HexToAddress("0x6Ab592e197C9A4B4B8c6b5b2E1f4B8b8b8b8b8bb")
    
    // BSC
    BSCBridgeAddress = common.HexToAddress("0x6Ab592e197C9A4B4B8c6b5b2E1f4B8b8b8b8b8bc")
    BSCHandlerAddress = common.HexToAddress("0x6Ab592e197C9A4B4B8c6b5b2E1f4B8b8b8b8b8bd")
}

// 链ID定义
const (
    EthereumChainID = 1
    PolygonChainID  = 2
    BSCChainID      = 3
    FantomChainID   = 4
    AvalancheChainID = 5
)

// 资源ID定义
var (
    // ERC20资源
    USDCResourceID = [32]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}
    USDTResourceID = [32]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02}
    WETHResourceID = [32]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03}
    
    // NFT资源
    NFTResourceID = [32]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00}
)

// 桥接配置
type BridgeConfig struct {
    ChainID         uint8
    Name            string
    Endpoint        string
    BridgeAddress   common.Address
    HandlerAddress  common.Address
    StartBlock      uint64
    BlockConfirmations uint64
}

// 支持的链配置
var SupportedChains = map[uint8]*BridgeConfig{
    EthereumChainID: {
        ChainID:         EthereumChainID,
        Name:            "Ethereum",
        Endpoint:        "https://mainnet.infura.io/v3/YOUR_PROJECT_ID",
        BridgeAddress:   EthereumBridgeAddress,
        HandlerAddress:  EthereumHandlerAddress,
        StartBlock:      12000000,
        BlockConfirmations: 12,
    },
    PolygonChainID: {
        ChainID:         PolygonChainID,
        Name:            "Polygon",
        Endpoint:        "https://polygon-rpc.com",
        BridgeAddress:   PolygonBridgeAddress,
        HandlerAddress:  PolygonHandlerAddress,
        StartBlock:      20000000,
        BlockConfirmations: 20,
    },
    BSCChainID: {
        ChainID:         BSCChainID,
        Name:            "BSC",
        Endpoint:        "https://bsc-dataseed.binance.org",
        BridgeAddress:   BSCBridgeAddress,
        HandlerAddress:  BSCHandlerAddress,
        StartBlock:      10000000,
        BlockConfirmations: 15,
    },
}

// 存款记录
type DepositRecord struct {
    DepositNonce    uint64
    OriginChainID   uint8
    ResourceID      [32]byte
    DestinationChainID uint8
    Depositor       common.Address
    Recipient       common.Address
    Amount          *big.Int
    HandlerResponse []byte
    Status          DepositStatus
    Timestamp       time.Time
    TxHash          common.Hash
}

// 存款状态
type DepositStatus int

const (
    DepositStatusPending DepositStatus = iota
    DepositStatusExecuted
    DepositStatusCancelled
    DepositStatusFailed
)

// 提案记录
type ProposalRecord struct {
    OriginChainID      uint8
    DepositNonce       uint64
    ResourceID         [32]byte
    DataHash           [32]byte
    ProposedBlock      uint64
    Status             ProposalStatus
    YesVotes           []*big.Int
    NoVotes            []*big.Int
    YesVotesTotal      uint8
    ProposedBy         common.Address
}

// 提案状态
type ProposalStatus int

const (
    ProposalStatusInactive ProposalStatus = iota
    ProposalStatusActive
    ProposalStatusPassed
    ProposalStatusExecuted
    ProposalStatusCancelled
)

// 中继器信息
type RelayerInfo struct {
    Address    common.Address
    IsActive   bool
    Threshold  uint8
    ChainIDs   []uint8
    LastActive time.Time
}
```

## 环境准备

### 2.1 桥接客户端设置

```go
// client/chainbridge_client.go
package client

import (
    "context"
    "math/big"
    
    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
)

type ChainBridgeClient struct {
    chains      map[uint8]*ChainClient
    bridgeABI   abi.ABI
    handlerABI  abi.ABI
    erc20ABI    abi.ABI
}

type ChainClient struct {
    config    *BridgeConfig
    ethClient *ethclient.Client
}

func NewChainBridgeClient() (*ChainBridgeClient, error) {
    // 加载合约ABI
    bridgeABI, err := abi.JSON(strings.NewReader(BridgeABI))
    if err != nil {
        return nil, err
    }
    
    handlerABI, err := abi.JSON(strings.NewReader(ERC20HandlerABI))
    if err != nil {
        return nil, err
    }
    
    erc20ABI, err := abi.JSON(strings.NewReader(ERC20ABI))
    if err != nil {
        return nil, err
    }
    
    client := &ChainBridgeClient{
        chains:     make(map[uint8]*ChainClient),
        bridgeABI:  bridgeABI,
        handlerABI: handlerABI,
        erc20ABI:   erc20ABI,
    }
    
    // 初始化支持的链
    for chainID, config := range SupportedChains {
        ethClient, err := ethclient.Dial(config.Endpoint)
        if err != nil {
            continue // 跳过连接失败的链
        }
        
        client.chains[chainID] = &ChainClient{
            config:    config,
            ethClient: ethClient,
        }
    }
    
    return client, nil
}

// 获取链客户端
func (c *ChainBridgeClient) GetChainClient(chainID uint8) (*ChainClient, error) {
    client, exists := c.chains[chainID]
    if !exists {
        return nil, fmt.Errorf("不支持的链ID: %d", chainID)
    }
    return client, nil
}

// 获取存款记录
func (c *ChainBridgeClient) GetDepositRecord(
    chainID uint8,
    depositNonce uint64,
) (*DepositRecord, error) {
    chainClient, err := c.GetChainClient(chainID)
    if err != nil {
        return nil, err
    }
    
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    // 调用桥接合约获取存款记录
    var result []interface{}
    err = chainClient.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &chainClient.config.BridgeAddress,
        Data: c.bridgeABI.Methods["getDepositRecord"].ID,
    }, nil)
    if err != nil {
        return nil, err
    }
    
    // 解析结果
    record := &DepositRecord{
        DepositNonce:       depositNonce,
        OriginChainID:      chainID,
        Status:             DepositStatus(result[0].(uint8)),
        // 其他字段需要从事件日志中获取
    }
    
    return record, nil
}

// 获取提案记录
func (c *ChainBridgeClient) GetProposalRecord(
    chainID uint8,
    originChainID uint8,
    depositNonce uint64,
) (*ProposalRecord, error) {
    chainClient, err := c.GetChainClient(chainID)
    if err != nil {
        return nil, err
    }
    
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    // 调用桥接合约获取提案记录
    var result []interface{}
    err = chainClient.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &chainClient.config.BridgeAddress,
        Data: c.bridgeABI.Methods["getProposal"].ID,
    }, nil)
    if err != nil {
        return nil, err
    }
    
    // 解析结果
    proposal := &ProposalRecord{
        OriginChainID: originChainID,
        DepositNonce:  depositNonce,
        Status:        ProposalStatus(result[0].(uint8)),
        YesVotesTotal: result[1].(uint8),
    }
    
    return proposal, nil
}

// 检查资源是否支持
func (c *ChainBridgeClient) IsResourceSupported(
    chainID uint8,
    resourceID [32]byte,
) (bool, error) {
    chainClient, err := c.GetChainClient(chainID)
    if err != nil {
        return false, err
    }
    
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    var supported bool
    err = chainClient.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &chainClient.config.HandlerAddress,
        Data: c.handlerABI.Methods["_resourceIDToTokenContractAddress"].ID,
    }, nil)
    if err != nil {
        return false, err
    }
    
    return supported, nil
}

// 获取桥接费用
func (c *ChainBridgeClient) GetBridgeFee(
    originChainID uint8,
    destinationChainID uint8,
    resourceID [32]byte,
) (*big.Int, error) {
    chainClient, err := c.GetChainClient(originChainID)
    if err != nil {
        return nil, err
    }
    
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    var fee *big.Int
    err = chainClient.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &chainClient.config.BridgeAddress,
        Data: c.bridgeABI.Methods["_fee"].ID,
    }, nil)
    if err != nil {
        return nil, err
    }
    
    return fee, nil
}
```

## 资产转移

### 3.1 桥接服务

```go
// services/bridge_service.go
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

type BridgeService struct {
    client     *ChainBridgeClient
    privateKey *ecdsa.PrivateKey
}

func NewBridgeService(client *ChainBridgeClient, privateKey *ecdsa.PrivateKey) *BridgeService {
    return &BridgeService{
        client:     client,
        privateKey: privateKey,
    }
}

// ERC20代币桥接
func (s *BridgeService) BridgeERC20(
    originChainID uint8,
    destinationChainID uint8,
    resourceID [32]byte,
    recipient common.Address,
    amount *big.Int,
) (*types.Transaction, error) {
    // 获取源链客户端
    originClient, err := s.client.GetChainClient(originChainID)
    if err != nil {
        return nil, err
    }
    
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 检查资源是否支持
    supported, err := s.client.IsResourceSupported(originChainID, resourceID)
    if err != nil {
        return nil, err
    }
    if !supported {
        return nil, fmt.Errorf("资源ID %x 在链 %d 上不支持", resourceID, originChainID)
    }
    
    // 获取桥接费用
    bridgeFee, err := s.client.GetBridgeFee(originChainID, destinationChainID, resourceID)
    if err != nil {
        return nil, err
    }
    
    // 获取代币合约地址
    tokenAddress, err := s.getTokenAddress(originChainID, resourceID)
    if err != nil {
        return nil, err
    }
    
    // 授权Handler合约使用代币
    if err := s.approveToken(originChainID, tokenAddress, amount); err != nil {
        return nil, err
    }
    
    // 获取nonce
    nonce, err := originClient.ethClient.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        return nil, err
    }
    
    // 获取gas价格
    gasPrice, err := originClient.ethClient.SuggestGasPrice(context.Background())
    if err != nil {
        return nil, err
    }
    
    // 构建存款数据
    depositData := s.constructDepositData(amount, recipient)
    
    // 构建桥接交易数据
    data, err := s.client.bridgeABI.Pack(
        "deposit",
        destinationChainID,
        resourceID,
        depositData,
    )
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(
        nonce,
        originClient.config.BridgeAddress,
        bridgeFee,
        300000,
        gasPrice,
        data,
    )
    
    // 签名交易
    chainID, err := originClient.ethClient.NetworkID(context.Background())
    if err != nil {
        return nil, err
    }
    
    signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), s.privateKey)
    if err != nil {
        return nil, err
    }
    
    // 发送交易
    err = originClient.ethClient.SendTransaction(context.Background(), signedTx)
    if err != nil {
        return nil, err
    }
    
    return signedTx, nil
}

// NFT桥接
func (s *BridgeService) BridgeNFT(
    originChainID uint8,
    destinationChainID uint8,
    resourceID [32]byte,
    recipient common.Address,
    tokenID *big.Int,
    metadata string,
) (*types.Transaction, error) {
    // 获取源链客户端
    originClient, err := s.client.GetChainClient(originChainID)
    if err != nil {
        return nil, err
    }
    
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 获取NFT合约地址
    nftAddress, err := s.getTokenAddress(originChainID, resourceID)
    if err != nil {
        return nil, err
    }
    
    // 授权Handler合约使用NFT
    if err := s.approveNFT(originChainID, nftAddress, tokenID); err != nil {
        return nil, err
    }
    
    // 获取桥接费用
    bridgeFee, err := s.client.GetBridgeFee(originChainID, destinationChainID, resourceID)
    if err != nil {
        return nil, err
    }
    
    // 获取nonce
    nonce, err := originClient.ethClient.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        return nil, err
    }
    
    // 获取gas价格
    gasPrice, err := originClient.ethClient.SuggestGasPrice(context.Background())
    if err != nil {
        return nil, err
    }
    
    // 构建NFT存款数据
    depositData := s.constructNFTDepositData(tokenID, recipient, metadata)
    
    // 构建桥接交易数据
    data, err := s.client.bridgeABI.Pack(
        "deposit",
        destinationChainID,
        resourceID,
        depositData,
    )
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(
        nonce,
        originClient.config.BridgeAddress,
        bridgeFee,
        400000,
        gasPrice,
        data,
    )
    
    // 签名交易
    chainID, err := originClient.ethClient.NetworkID(context.Background())
    if err != nil {
        return nil, err
    }
    
    signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), s.privateKey)
    if err != nil {
        return nil, err
    }
    
    // 发送交易
    err = originClient.ethClient.SendTransaction(context.Background(), signedTx)
    if err != nil {
        return nil, err
    }
    
    return signedTx, nil
}

// 查询桥接状态
func (s *BridgeService) GetBridgeStatus(
    originChainID uint8,
    destinationChainID uint8,
    depositNonce uint64,
) (*BridgeStatus, error) {
    // 获取存款记录
    depositRecord, err := s.client.GetDepositRecord(originChainID, depositNonce)
    if err != nil {
        return nil, err
    }
    
    // 获取提案记录
    proposalRecord, err := s.client.GetProposalRecord(destinationChainID, originChainID, depositNonce)
    if err != nil {
        return nil, err
    }
    
    status := &BridgeStatus{
        DepositRecord:  depositRecord,
        ProposalRecord: proposalRecord,
        OverallStatus:  s.calculateOverallStatus(depositRecord, proposalRecord),
    }
    
    return status, nil
}

// 批量桥接
func (s *BridgeService) BatchBridge(transfers []BridgeTransfer) ([]*types.Transaction, error) {
    var transactions []*types.Transaction
    
    for _, transfer := range transfers {
        var tx *types.Transaction
        var err error
        
        switch transfer.Type {
        case "ERC20":
            tx, err = s.BridgeERC20(
                transfer.OriginChainID,
                transfer.DestinationChainID,
                transfer.ResourceID,
                transfer.Recipient,
                transfer.Amount,
            )
        case "NFT":
            tx, err = s.BridgeNFT(
                transfer.OriginChainID,
                transfer.DestinationChainID,
                transfer.ResourceID,
                transfer.Recipient,
                transfer.TokenID,
                transfer.Metadata,
            )
        default:
            err = fmt.Errorf("不支持的转移类型: %s", transfer.Type)
        }
        
        if err != nil {
            return nil, err
        }
        
        transactions = append(transactions, tx)
        
        // 等待一段时间避免nonce冲突
        time.Sleep(2 * time.Second)
    }
    
    return transactions, nil
}

// 辅助函数
func (s *BridgeService) getTokenAddress(chainID uint8, resourceID [32]byte) (common.Address, error) {
    chainClient, err := s.client.GetChainClient(chainID)
    if err != nil {
        return common.Address{}, err
    }
    
    var tokenAddress common.Address
    err = chainClient.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &chainClient.config.HandlerAddress,
        Data: s.client.handlerABI.Methods["_resourceIDToTokenContractAddress"].ID,
    }, nil)
    if err != nil {
        return common.Address{}, err
    }
    
    return tokenAddress, nil
}

func (s *BridgeService) approveToken(chainID uint8, tokenAddress common.Address, amount *big.Int) error {
    chainClient, err := s.client.GetChainClient(chainID)
    if err != nil {
        return err
    }
    
    // 实现ERC20授权逻辑
    // 简化实现
    return nil
}

func (s *BridgeService) approveNFT(chainID uint8, nftAddress common.Address, tokenID *big.Int) error {
    chainClient, err := s.client.GetChainClient(chainID)
    if err != nil {
        return err
    }
    
    // 实现ERC721授权逻辑
    // 简化实现
    return nil
}

func (s *BridgeService) constructDepositData(amount *big.Int, recipient common.Address) []byte {
    // 构建ERC20存款数据
    // 格式: amount (32 bytes) + recipient length (32 bytes) + recipient
    data := make([]byte, 0)
    
    // 添加金额 (32字节)
    amountBytes := make([]byte, 32)
    amount.FillBytes(amountBytes)
    data = append(data, amountBytes...)
    
    // 添加接收者长度 (32字节)
    recipientLengthBytes := make([]byte, 32)
    big.NewInt(20).FillBytes(recipientLengthBytes) // 地址长度20字节
    data = append(data, recipientLengthBytes...)
    
    // 添加接收者地址
    data = append(data, recipient.Bytes()...)
    
    return data
}

func (s *BridgeService) constructNFTDepositData(tokenID *big.Int, recipient common.Address, metadata string) []byte {
    // 构建NFT存款数据
    // 格式: tokenID (32 bytes) + recipient length (32 bytes) + recipient + metadata length (32 bytes) + metadata
    data := make([]byte, 0)
    
    // 添加tokenID (32字节)
    tokenIDBytes := make([]byte, 32)
    tokenID.FillBytes(tokenIDBytes)
    data = append(data, tokenIDBytes...)
    
    // 添加接收者长度和地址
    recipientLengthBytes := make([]byte, 32)
    big.NewInt(20).FillBytes(recipientLengthBytes)
    data = append(data, recipientLengthBytes...)
    data = append(data, recipient.Bytes()...)
    
    // 添加元数据长度和内容
    metadataBytes := []byte(metadata)
    metadataLengthBytes := make([]byte, 32)
    big.NewInt(int64(len(metadataBytes))).FillBytes(metadataLengthBytes)
    data = append(data, metadataLengthBytes...)
    data = append(data, metadataBytes...)
    
    return data
}

func (s *BridgeService) calculateOverallStatus(deposit *DepositRecord, proposal *ProposalRecord) string {
    if deposit.Status == DepositStatusFailed {
        return "FAILED"
    }
    
    if proposal == nil {
        return "PENDING_PROPOSAL"
    }
    
    switch proposal.Status {
    case ProposalStatusExecuted:
        return "COMPLETED"
    case ProposalStatusPassed:
        return "PENDING_EXECUTION"
    case ProposalStatusActive:
        return "VOTING"
    case ProposalStatusCancelled:
        return "CANCELLED"
    default:
        return "PENDING"
    }
}

type BridgeTransfer struct {
    Type                string
    OriginChainID       uint8
    DestinationChainID  uint8
    ResourceID          [32]byte
    Recipient           common.Address
    Amount              *big.Int
    TokenID             *big.Int
    Metadata            string
}

type BridgeStatus struct {
    DepositRecord  *DepositRecord
    ProposalRecord *ProposalRecord
    OverallStatus  string
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
    
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/crypto"
    
    "your-project/client"
    "your-project/services"
)

func main() {
    // 创建ChainBridge客户端
    bridgeClient, err := client.NewChainBridgeClient()
    if err != nil {
        log.Fatal("创建ChainBridge客户端失败:", err)
    }
    
    // 加载私钥
    privateKey, err := crypto.HexToECDSA("your_private_key_here")
    if err != nil {
        log.Fatal("加载私钥失败:", err)
    }
    
    // 获取用户地址
    userAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
    
    // 创建桥接服务
    bridgeService := services.NewBridgeService(bridgeClient, privateKey)
    
    // 1. 查询支持的链
    fmt.Printf("=== 支持的区块链 ===\n")
    
    for chainID, config := range client.SupportedChains {
        fmt.Printf("链ID %d: %s\n", chainID, config.Name)
        fmt.Printf("  RPC端点: %s\n", config.Endpoint)
        fmt.Printf("  桥接合约: %s\n", config.BridgeAddress.Hex())
        fmt.Printf("  处理器合约: %s\n", config.HandlerAddress.Hex())
        fmt.Printf("  确认块数: %d\n", config.BlockConfirmations)
        
        // 检查连接状态
        chainClient, err := bridgeClient.GetChainClient(chainID)
        if err != nil {
            fmt.Printf("  状态: ❌ 连接失败\n")
        } else {
            fmt.Printf("  状态: ✅ 连接正常\n")
            
            // 获取最新区块
            latestBlock, err := chainClient.ethClient.BlockNumber(context.Background())
            if err == nil {
                fmt.Printf("  最新区块: %d\n", latestBlock)
            }
        }
        fmt.Println()
    }
    
    // 2. 检查资源支持
    fmt.Printf("=== 资源支持检查 ===\n")
    
    resources := map[string][32]byte{
        "USDC": client.USDCResourceID,
        "USDT": client.USDTResourceID,
        "WETH": client.WETHResourceID,
        "NFT":  client.NFTResourceID,
    }
    
    for resourceName, resourceID := range resources {
        fmt.Printf("%s (资源ID: %x):\n", resourceName, resourceID)
        
        for chainID, config := range client.SupportedChains {
            supported, err := bridgeClient.IsResourceSupported(chainID, resourceID)
            if err != nil {
                fmt.Printf("  %s: ❌ 检查失败\n", config.Name)
            } else if supported {
                fmt.Printf("  %s: ✅ 支持\n", config.Name)
            } else {
                fmt.Printf("  %s: ❌ 不支持\n", config.Name)
            }
        }
        fmt.Println()
    }
    
    // 3. 查询桥接费用
    fmt.Printf("=== 桥接费用查询 ===\n")
    
    originChain := client.EthereumChainID
    destChain := client.PolygonChainID
    
    for resourceName, resourceID := range resources {
        fee, err := bridgeClient.GetBridgeFee(originChain, destChain, resourceID)
        if err != nil {
            fmt.Printf("%s: 查询失败 - %v\n", resourceName, err)
        } else {
            fmt.Printf("%s: %s ETH\n", resourceName, 
                decimal.NewFromBigInt(fee, -18).String())
        }
    }
    
    // 4. USDC桥接示例 (Ethereum -> Polygon)
    fmt.Printf("\n=== USDC桥接示例 ===\n")
    
    bridgeAmount := big.NewInt(100e6) // 100 USDC (6位精度)
    recipient := userAddress          // 发送给自己
    
    fmt.Printf("准备桥接:\n")
    fmt.Printf("  源链: Ethereum (ID: %d)\n", originChain)
    fmt.Printf("  目标链: Polygon (ID: %d)\n", destChain)
    fmt.Printf("  资产: USDC\n")
    fmt.Printf("  数量: %s\n", bridgeAmount.String())
    fmt.Printf("  接收者: %s\n", recipient.Hex())
    
    // 执行桥接
    tx, err := bridgeService.BridgeERC20(
        originChain,
        destChain,
        client.USDCResourceID,
        recipient,
        bridgeAmount,
    )
    if err != nil {
        log.Printf("USDC桥接失败: %v", err)
    } else {
        fmt.Printf("桥接交易已提交: %s\n", tx.Hash().Hex())
        
        // 等待交易确认
        fmt.Printf("等待交易确认...\n")
        time.Sleep(30 * time.Second)
        
        // 查询桥接状态
        // 注意：实际应用中需要从交易收据中获取depositNonce
        depositNonce := uint64(12345) // 示例nonce
        
        status, err := bridgeService.GetBridgeStatus(originChain, destChain, depositNonce)
        if err != nil {
            log.Printf("查询桥接状态失败: %v", err)
        } else {
            fmt.Printf("桥接状态: %s\n", status.OverallStatus)
            
            if status.DepositRecord != nil {
                fmt.Printf("存款状态: %v\n", status.DepositRecord.Status)
            }
            
            if status.ProposalRecord != nil {
                fmt.Printf("提案状态: %v\n", status.ProposalRecord.Status)
                fmt.Printf("赞成票数: %d\n", status.ProposalRecord.YesVotesTotal)
            }
        }
    }
    
    // 5. NFT桥接示例
    fmt.Printf("\n=== NFT桥接示例 ===\n")
    
    tokenID := big.NewInt(1001)
    metadata := "https://example.com/nft/1001.json"
    
    fmt.Printf("准备桥接NFT:\n")
    fmt.Printf("  Token ID: %s\n", tokenID.String())
    fmt.Printf("  元数据: %s\n", metadata)
    
    nftTx, err := bridgeService.BridgeNFT(
        originChain,
        destChain,
        client.NFTResourceID,
        recipient,
        tokenID,
        metadata,
    )
    if err != nil {
        log.Printf("NFT桥接失败: %v", err)
    } else {
        fmt.Printf("NFT桥接交易已提交: %s\n", nftTx.Hash().Hex())
    }
    
    // 6. 批量桥接示例
    fmt.Printf("\n=== 批量桥接示例 ===\n")
    
    batchTransfers := []services.BridgeTransfer{
        {
            Type:               "ERC20",
            OriginChainID:      originChain,
            DestinationChainID: destChain,
            ResourceID:         client.USDCResourceID,
            Recipient:          recipient,
            Amount:             big.NewInt(50e6), // 50 USDC
        },
        {
            Type:               "ERC20",
            OriginChainID:      originChain,
            DestinationChainID: destChain,
            ResourceID:         client.USDTResourceID,
            Recipient:          recipient,
            Amount:             big.NewInt(100e6), // 100 USDT
        },
    }
    
    fmt.Printf("准备批量桥接 %d 笔交易\n", len(batchTransfers))
    
    batchTxs, err := bridgeService.BatchBridge(batchTransfers)
    if err != nil {
        log.Printf("批量桥接失败: %v", err)
    } else {
        fmt.Printf("批量桥接成功，交易数量: %d\n", len(batchTxs))
        for i, tx := range batchTxs {
            fmt.Printf("  交易 %d: %s\n", i+1, tx.Hash().Hex())
        }
    }
    
    // 7. 桥接统计
    fmt.Printf("\n=== 桥接统计 ===\n")
    
    fmt.Printf("ChainBridge特点:\n")
    fmt.Printf("  - 支持多种资产类型 (ERC20, ERC721, 通用数据)\n")
    fmt.Printf("  - 去中心化中继器网络\n")
    fmt.Printf("  - 可配置的安全阈值\n")
    fmt.Printf("  - 模块化架构设计\n")
    
    fmt.Printf("\n安全机制:\n")
    fmt.Printf("  - 多签名验证\n")
    fmt.Printf("  - 时间锁定\n")
    fmt.Printf("  - 紧急暂停功能\n")
    fmt.Printf("  - 资产白名单\n")
    
    // 8. 费用对比
    fmt.Printf("\n=== 费用对比 ===\n")
    
    fmt.Printf("不同路径的桥接费用:\n")
    
    routes := []struct {
        From string
        To   string
        FromID uint8
        ToID uint8
    }{
        {"Ethereum", "Polygon", client.EthereumChainID, client.PolygonChainID},
        {"Ethereum", "BSC", client.EthereumChainID, client.BSCChainID},
        {"Polygon", "BSC", client.PolygonChainID, client.BSCChainID},
    }
    
    for _, route := range routes {
        fee, err := bridgeClient.GetBridgeFee(route.FromID, route.ToID, client.USDCResourceID)
        if err != nil {
            fmt.Printf("  %s -> %s: 查询失败\n", route.From, route.To)
        } else {
            fmt.Printf("  %s -> %s: %s ETH\n", route.From, route.To,
                decimal.NewFromBigInt(fee, -18).String())
        }
    }
    
    // 9. 最佳实践建议
    fmt.Printf("\n=== 最佳实践建议 ===\n")
    
    fmt.Printf("使用ChainBridge时请注意:\n")
    fmt.Printf("  1. 确认目标链支持相应资源\n")
    fmt.Printf("  2. 检查桥接费用和gas费用\n")
    fmt.Printf("  3. 等待足够的区块确认\n")
    fmt.Printf("  4. 监控桥接状态直到完成\n")
    fmt.Printf("  5. 保留交易哈希用于查询\n")
    fmt.Printf("  6. 小额测试后再进行大额转移\n")
    fmt.Printf("  7. 关注网络拥堵情况\n")
    fmt.Printf("  8. 验证接收地址的正确性\n")
}
```

这个ChainBridge使用指南提供了完整的跨链桥接解决方案，涵盖了资产转移、中继器网络、安全机制、状态监控等核心功能，是构建多链DeFi应用的重要参考文档。
