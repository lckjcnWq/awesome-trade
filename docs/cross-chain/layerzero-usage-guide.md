# LayerZero 全链协议 详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [环境准备](#环境准备)
3. [端点架构](#端点架构)
4. [消息传递](#消息传递)
5. [全链应用](#全链应用)
6. [安全机制](#安全机制)
7. [Gas优化](#gas优化)
8. [监控和调试](#监控和调试)
9. [实际应用](#实际应用)

## 基础概念

### 1.1 LayerZero 简介

LayerZero 是全链互操作性协议，通过超轻节点和预言机网络实现不同区块链之间的无缝通信，支持构建真正的全链应用。

```bash
# 安装LayerZero相关依赖
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

// LayerZero 链ID定义
const (
    EthereumChainID   = 101
    BSCChainID        = 102
    AvalancheChainID  = 106
    PolygonChainID    = 109
    ArbitrumChainID   = 110
    OptimismChainID   = 111
    FantomChainID     = 112
    AptosCoreChainID  = 108
    SolanaChainID     = 168
)

// LayerZero 端点地址
var LayerZeroEndpoints = map[uint16]common.Address{
    EthereumChainID:  common.HexToAddress("0x66A71Dcef29A0fFBDBE3c6a460a3B5BC225Cd675"),
    BSCChainID:       common.HexToAddress("0x3c2269811836af69497E5F486A85D7316753cf62"),
    AvalancheChainID: common.HexToAddress("0x3c2269811836af69497E5F486A85D7316753cf62"),
    PolygonChainID:   common.HexToAddress("0x3c2269811836af69497E5F486A85D7316753cf62"),
    ArbitrumChainID:  common.HexToAddress("0x3c2269811836af69497E5F486A85D7316753cf62"),
    OptimismChainID:  common.HexToAddress("0x3c2269811836af69497E5F486A85D7316753cf62"),
    FantomChainID:    common.HexToAddress("0xb6319cC6c8c27A8F5dAF0dD3DF91EA35C4720dd7"),
}

// 网络配置
type ChainConfig struct {
    ChainID     uint16
    Name        string
    RPC         string
    Endpoint    common.Address
    GasLimit    uint64
    GasPrice    *big.Int
}

var SupportedChains = map[uint16]*ChainConfig{
    EthereumChainID: {
        ChainID:  EthereumChainID,
        Name:     "Ethereum",
        RPC:      "https://mainnet.infura.io/v3/YOUR_PROJECT_ID",
        Endpoint: LayerZeroEndpoints[EthereumChainID],
        GasLimit: 200000,
        GasPrice: big.NewInt(20e9), // 20 Gwei
    },
    BSCChainID: {
        ChainID:  BSCChainID,
        Name:     "BSC",
        RPC:      "https://bsc-dataseed.binance.org",
        Endpoint: LayerZeroEndpoints[BSCChainID],
        GasLimit: 200000,
        GasPrice: big.NewInt(5e9), // 5 Gwei
    },
    PolygonChainID: {
        ChainID:  PolygonChainID,
        Name:     "Polygon",
        RPC:      "https://polygon-rpc.com",
        Endpoint: LayerZeroEndpoints[PolygonChainID],
        GasLimit: 200000,
        GasPrice: big.NewInt(30e9), // 30 Gwei
    },
}

// LayerZero 消息类型
type LzMessage struct {
    SrcChainId    uint16
    DstChainId    uint16
    SrcAddress    []byte
    DstAddress    []byte
    Nonce         uint64
    Payload       []byte
    AdapterParams []byte
}

// 消息配置
type MessageConfig struct {
    Version       uint16
    GasLimit      uint64
    NativeForDst  *big.Int
    AdapterParams []byte
}

// 应用配置
type AppConfig struct {
    InboundProofLibraryVersion  uint16
    InboundBlockConfirmations   uint64
    Relayer                     common.Address
    OutboundProofType           uint16
    OutboundBlockConfirmations  uint64
    Oracle                      common.Address
}

// 费用估算
type FeeEstimate struct {
    NativeFee  *big.Int
    ZroFee     *big.Int
    TotalFee   *big.Int
}

// 交易状态
type MessageStatus struct {
    SrcTxHash     common.Hash
    DstTxHash     common.Hash
    Status        string
    SrcChainId    uint16
    DstChainId    uint16
    Nonce         uint64
    Payload       []byte
    Timestamp     uint64
    Confirmations uint64
    Error         string
}

// 全链代币信息
type OmniToken struct {
    Name           string
    Symbol         string
    Decimals       uint8
    TotalSupply    *big.Int
    Deployments    map[uint16]common.Address
    IsOFT          bool
    SharedDecimals uint8
}
```

## 环境准备

### 2.1 LayerZero客户端设置

```go
// client/layerzero_client.go
package client

import (
    "context"
    "math/big"
    
    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
)

type LayerZeroClient struct {
    chains      map[uint16]*ChainClient
    endpointABI abi.ABI
    oftABI      abi.ABI
    oappABI     abi.ABI
}

type ChainClient struct {
    config    *ChainConfig
    ethClient *ethclient.Client
}

func NewLayerZeroClient() (*LayerZeroClient, error) {
    // 加载合约ABI
    endpointABI, err := abi.JSON(strings.NewReader(LayerZeroEndpointABI))
    if err != nil {
        return nil, err
    }
    
    oftABI, err := abi.JSON(strings.NewReader(OFTABI))
    if err != nil {
        return nil, err
    }
    
    oappABI, err := abi.JSON(strings.NewReader(OAppABI))
    if err != nil {
        return nil, err
    }
    
    client := &LayerZeroClient{
        chains:      make(map[uint16]*ChainClient),
        endpointABI: endpointABI,
        oftABI:      oftABI,
        oappABI:     oappABI,
    }
    
    // 初始化支持的链
    for chainID, config := range SupportedChains {
        ethClient, err := ethclient.Dial(config.RPC)
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
func (c *LayerZeroClient) GetChainClient(chainID uint16) (*ChainClient, error) {
    client, exists := c.chains[chainID]
    if !exists {
        return nil, fmt.Errorf("不支持的链ID: %d", chainID)
    }
    return client, nil
}

// 估算跨链费用
func (c *LayerZeroClient) EstimateFees(
    srcChainID uint16,
    dstChainID uint16,
    userApplication common.Address,
    payload []byte,
    payInZRO bool,
    adapterParams []byte,
) (*FeeEstimate, error) {
    srcClient, err := c.GetChainClient(srcChainID)
    if err != nil {
        return nil, err
    }
    
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    // 调用端点合约估算费用
    var result []interface{}
    err = srcClient.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &srcClient.config.Endpoint,
        Data: c.endpointABI.Methods["estimateFees"].ID,
    }, nil)
    if err != nil {
        return nil, err
    }
    
    nativeFee := result[0].(*big.Int)
    zroFee := result[1].(*big.Int)
    
    totalFee := new(big.Int).Add(nativeFee, zroFee)
    
    return &FeeEstimate{
        NativeFee: nativeFee,
        ZroFee:    zroFee,
        TotalFee:  totalFee,
    }, nil
}

// 发送消息
func (c *LayerZeroClient) Send(
    srcChainID uint16,
    dstChainID uint16,
    dstAddress []byte,
    payload []byte,
    refundAddress common.Address,
    zroPaymentAddress common.Address,
    adapterParams []byte,
    nativeFee *big.Int,
) (*types.Transaction, error) {
    srcClient, err := c.GetChainClient(srcChainID)
    if err != nil {
        return nil, err
    }
    
    // 构建发送消息的交易数据
    data, err := c.endpointABI.Pack(
        "send",
        dstChainID,
        dstAddress,
        payload,
        refundAddress,
        zroPaymentAddress,
        adapterParams,
    )
    if err != nil {
        return nil, err
    }
    
    // 创建交易
    tx := types.NewTransaction(
        0, // nonce将在发送时设置
        srcClient.config.Endpoint,
        nativeFee,
        srcClient.config.GasLimit,
        srcClient.config.GasPrice,
        data,
    )
    
    return tx, nil
}

// 获取消息状态
func (c *LayerZeroClient) GetMessageStatus(
    srcChainID uint16,
    srcAddress common.Address,
    dstChainID uint16,
    dstAddress common.Address,
    nonce uint64,
) (*MessageStatus, error) {
    // 查询源链的发送事件
    srcClient, err := c.GetChainClient(srcChainID)
    if err != nil {
        return nil, err
    }
    
    // 查询目标链的接收事件
    dstClient, err := c.GetChainClient(dstChainID)
    if err != nil {
        return nil, err
    }
    
    // 简化实现 - 实际需要查询事件日志
    status := &MessageStatus{
        SrcChainId: srcChainID,
        DstChainId: dstChainID,
        Nonce:      nonce,
        Status:     "PENDING",
    }
    
    return status, nil
}

// 获取应用配置
func (c *LayerZeroClient) GetAppConfig(
    chainID uint16,
    userApplication common.Address,
    remoteChainID uint16,
) (*AppConfig, error) {
    chainClient, err := c.GetChainClient(chainID)
    if err != nil {
        return nil, err
    }
    
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    // 查询应用配置
    var result []interface{}
    err = chainClient.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &chainClient.config.Endpoint,
        Data: c.endpointABI.Methods["getConfig"].ID,
    }, nil)
    if err != nil {
        return nil, err
    }
    
    config := &AppConfig{
        InboundProofLibraryVersion:  result[0].(uint16),
        InboundBlockConfirmations:   result[1].(uint64),
        Relayer:                     result[2].(common.Address),
        OutboundProofType:           result[3].(uint16),
        OutboundBlockConfirmations:  result[4].(uint64),
        Oracle:                      result[5].(common.Address),
    }
    
    return config, nil
}

// 检查路径是否可信
func (c *LayerZeroClient) IsTrustedRemote(
    chainID uint16,
    userApplication common.Address,
    remoteChainID uint16,
    remoteAddress []byte,
) (bool, error) {
    chainClient, err := c.GetChainClient(chainID)
    if err != nil {
        return false, err
    }
    
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    var trusted bool
    err = chainClient.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &userApplication,
        Data: c.oappABI.Methods["isTrustedRemote"].ID,
    }, nil)
    if err != nil {
        return false, err
    }
    
    return trusted, nil
}
```

## 消息传递

### 3.1 消息服务

```go
// services/message_service.go
package services

import (
    "context"
    "crypto/ecdsa"
    "encoding/hex"
    "math/big"
    
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/crypto"
)

type LayerZeroMessageService struct {
    client     *LayerZeroClient
    privateKey *ecdsa.PrivateKey
}

func NewLayerZeroMessageService(client *LayerZeroClient, privateKey *ecdsa.PrivateKey) *LayerZeroMessageService {
    return &LayerZeroMessageService{
        client:     client,
        privateKey: privateKey,
    }
}

// 发送跨链消息
func (s *LayerZeroMessageService) SendMessage(
    srcChainID uint16,
    dstChainID uint16,
    dstAddress common.Address,
    payload []byte,
    gasLimit uint64,
    nativeForDst *big.Int,
) (*types.Transaction, error) {
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 构建适配器参数
    adapterParams := s.buildAdapterParams(1, gasLimit, nativeForDst)
    
    // 估算费用
    feeEstimate, err := s.client.EstimateFees(
        srcChainID,
        dstChainID,
        fromAddress,
        payload,
        false, // 不使用ZRO支付
        adapterParams,
    )
    if err != nil {
        return nil, err
    }
    
    // 发送消息
    tx, err := s.client.Send(
        srcChainID,
        dstChainID,
        dstAddress.Bytes(),
        payload,
        fromAddress,
        common.Address{}, // 不使用ZRO支付
        adapterParams,
        feeEstimate.NativeFee,
    )
    if err != nil {
        return nil, err
    }
    
    return s.signAndSendTransaction(srcChainID, tx)
}

// 发送OFT代币
func (s *LayerZeroMessageService) SendOFT(
    srcChainID uint16,
    dstChainID uint16,
    oftContract common.Address,
    to common.Address,
    amount *big.Int,
    gasLimit uint64,
) (*types.Transaction, error) {
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 构建适配器参数
    adapterParams := s.buildAdapterParams(1, gasLimit, big.NewInt(0))
    
    // 估算OFT发送费用
    feeEstimate, err := s.estimateOFTSendFee(
        srcChainID,
        oftContract,
        dstChainID,
        to.Bytes(),
        amount,
        false,
        adapterParams,
    )
    if err != nil {
        return nil, err
    }
    
    // 获取源链客户端
    srcClient, err := s.client.GetChainClient(srcChainID)
    if err != nil {
        return nil, err
    }
    
    // 获取nonce
    nonce, err := srcClient.ethClient.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        return nil, err
    }
    
    // 构建OFT发送交易数据
    data, err := s.client.oftABI.Pack(
        "sendFrom",
        fromAddress,
        dstChainID,
        to.Bytes(),
        amount,
        fromAddress, // refundAddress
        common.Address{}, // zroPaymentAddress
        adapterParams,
    )
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(
        nonce,
        oftContract,
        feeEstimate.NativeFee,
        gasLimit,
        srcClient.config.GasPrice,
        data,
    )
    
    return s.signAndSendTransaction(srcChainID, tx)
}

// 批量发送消息
func (s *LayerZeroMessageService) BatchSendMessages(messages []CrossChainMessage) ([]*types.Transaction, error) {
    var transactions []*types.Transaction
    
    for _, msg := range messages {
        tx, err := s.SendMessage(
            msg.SrcChainID,
            msg.DstChainID,
            msg.DstAddress,
            msg.Payload,
            msg.GasLimit,
            msg.NativeForDst,
        )
        if err != nil {
            return nil, err
        }
        
        transactions = append(transactions, tx)
        
        // 等待一段时间避免nonce冲突
        time.Sleep(2 * time.Second)
    }
    
    return transactions, nil
}

// 重试失败的消息
func (s *LayerZeroMessageService) RetryMessage(
    srcChainID uint16,
    srcAddress common.Address,
    dstChainID uint16,
    nonce uint64,
    payload []byte,
) (*types.Transaction, error) {
    // 获取目标链客户端
    dstClient, err := s.client.GetChainClient(dstChainID)
    if err != nil {
        return nil, err
    }
    
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 获取nonce
    txNonce, err := dstClient.ethClient.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        return nil, err
    }
    
    // 构建重试交易数据
    data, err := s.client.endpointABI.Pack(
        "retryPayload",
        srcChainID,
        srcAddress.Bytes(),
        payload,
    )
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(
        txNonce,
        dstClient.config.Endpoint,
        big.NewInt(0),
        dstClient.config.GasLimit,
        dstClient.config.GasPrice,
        data,
    )
    
    return s.signAndSendTransaction(dstChainID, tx)
}

// 查询消息状态
func (s *LayerZeroMessageService) TrackMessage(
    srcChainID uint16,
    srcTxHash common.Hash,
) (*MessageStatus, error) {
    // 获取源链客户端
    srcClient, err := s.client.GetChainClient(srcChainID)
    if err != nil {
        return nil, err
    }
    
    // 获取交易收据
    receipt, err := srcClient.ethClient.TransactionReceipt(context.Background(), srcTxHash)
    if err != nil {
        return nil, err
    }
    
    // 解析事件日志获取消息信息
    // 简化实现
    status := &MessageStatus{
        SrcTxHash:  srcTxHash,
        SrcChainId: srcChainID,
        Status:     "SENT",
        Timestamp:  uint64(time.Now().Unix()),
    }
    
    return status, nil
}

// 辅助函数
func (s *LayerZeroMessageService) buildAdapterParams(
    version uint16,
    gasLimit uint64,
    nativeForDst *big.Int,
) []byte {
    // 构建适配器参数
    // 版本1格式: version(2) + gasLimit(32) + nativeForDst(32)
    params := make([]byte, 0)
    
    // 添加版本
    versionBytes := make([]byte, 2)
    binary.BigEndian.PutUint16(versionBytes, version)
    params = append(params, versionBytes...)
    
    // 添加gas限制
    gasLimitBytes := make([]byte, 32)
    big.NewInt(int64(gasLimit)).FillBytes(gasLimitBytes)
    params = append(params, gasLimitBytes...)
    
    // 添加原生代币数量
    nativeBytes := make([]byte, 32)
    if nativeForDst != nil {
        nativeForDst.FillBytes(nativeBytes)
    }
    params = append(params, nativeBytes...)
    
    return params
}

func (s *LayerZeroMessageService) estimateOFTSendFee(
    srcChainID uint16,
    oftContract common.Address,
    dstChainID uint16,
    toAddress []byte,
    amount *big.Int,
    useZro bool,
    adapterParams []byte,
) (*FeeEstimate, error) {
    srcClient, err := s.client.GetChainClient(srcChainID)
    if err != nil {
        return nil, err
    }
    
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    // 调用OFT合约估算费用
    var result []interface{}
    err = srcClient.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &oftContract,
        Data: s.client.oftABI.Methods["estimateSendFee"].ID,
    }, nil)
    if err != nil {
        return nil, err
    }
    
    nativeFee := result[0].(*big.Int)
    zroFee := result[1].(*big.Int)
    
    return &FeeEstimate{
        NativeFee: nativeFee,
        ZroFee:    zroFee,
        TotalFee:  new(big.Int).Add(nativeFee, zroFee),
    }, nil
}

func (s *LayerZeroMessageService) signAndSendTransaction(
    chainID uint16,
    tx *types.Transaction,
) (*types.Transaction, error) {
    chainClient, err := s.client.GetChainClient(chainID)
    if err != nil {
        return nil, err
    }
    
    // 获取网络ID
    networkID, err := chainClient.ethClient.NetworkID(context.Background())
    if err != nil {
        return nil, err
    }
    
    // 签名交易
    signedTx, err := types.SignTx(tx, types.NewEIP155Signer(networkID), s.privateKey)
    if err != nil {
        return nil, err
    }
    
    // 发送交易
    err = chainClient.ethClient.SendTransaction(context.Background(), signedTx)
    if err != nil {
        return nil, err
    }
    
    return signedTx, nil
}

type CrossChainMessage struct {
    SrcChainID   uint16
    DstChainID   uint16
    DstAddress   common.Address
    Payload      []byte
    GasLimit     uint64
    NativeForDst *big.Int
}
```

## 全链应用

### 4.1 OFT代币服务

```go
// services/oft_service.go
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

type OFTService struct {
    client     *LayerZeroClient
    privateKey *ecdsa.PrivateKey
}

func NewOFTService(client *LayerZeroClient, privateKey *ecdsa.PrivateKey) *OFTService {
    return &OFTService{
        client:     client,
        privateKey: privateKey,
    }
}

// 部署OFT代币
func (s *OFTService) DeployOFT(
    chainID uint16,
    name string,
    symbol string,
    sharedDecimals uint8,
) (*types.Transaction, error) {
    chainClient, err := s.client.GetChainClient(chainID)
    if err != nil {
        return nil, err
    }
    
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 获取nonce
    nonce, err := chainClient.ethClient.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        return nil, err
    }
    
    // 构建部署交易 (简化实现)
    // 实际需要合约字节码和构造函数参数
    deployData := []byte{} // OFT合约字节码 + 构造函数参数
    
    tx := types.NewTransaction(
        nonce,
        common.Address{}, // 部署合约地址为空
        big.NewInt(0),
        3000000, // 部署需要更多gas
        chainClient.config.GasPrice,
        deployData,
    )
    
    return s.signAndSendTransaction(chainID, tx)
}

// 设置可信远程地址
func (s *OFTService) SetTrustedRemote(
    chainID uint16,
    oftContract common.Address,
    remoteChainID uint16,
    remoteAddress common.Address,
) (*types.Transaction, error) {
    chainClient, err := s.client.GetChainClient(chainID)
    if err != nil {
        return nil, err
    }
    
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 获取nonce
    nonce, err := chainClient.ethClient.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        return nil, err
    }
    
    // 构建远程路径 (remoteChainID + remoteAddress)
    remotePath := append(
        big.NewInt(int64(remoteChainID)).Bytes(),
        remoteAddress.Bytes()...,
    )
    
    // 构建设置可信远程交易数据
    data, err := s.client.oftABI.Pack("setTrustedRemote", remoteChainID, remotePath)
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(
        nonce,
        oftContract,
        big.NewInt(0),
        200000,
        chainClient.config.GasPrice,
        data,
    )
    
    return s.signAndSendTransaction(chainID, tx)
}

// 跨链转账OFT代币
func (s *OFTService) SendOFT(
    srcChainID uint16,
    dstChainID uint16,
    oftContract common.Address,
    to common.Address,
    amount *big.Int,
    gasLimit uint64,
    nativeForDst *big.Int,
) (*types.Transaction, error) {
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 检查余额
    balance, err := s.GetOFTBalance(srcChainID, oftContract, fromAddress)
    if err != nil {
        return nil, err
    }
    
    if balance.Cmp(amount) < 0 {
        return nil, fmt.Errorf("余额不足: 需要 %s, 拥有 %s", amount.String(), balance.String())
    }
    
    // 构建适配器参数
    adapterParams := s.buildAdapterParams(1, gasLimit, nativeForDst)
    
    // 估算费用
    feeEstimate, err := s.estimateOFTSendFee(
        srcChainID,
        oftContract,
        dstChainID,
        to.Bytes(),
        amount,
        false,
        adapterParams,
    )
    if err != nil {
        return nil, err
    }
    
    // 获取源链客户端
    srcClient, err := s.client.GetChainClient(srcChainID)
    if err != nil {
        return nil, err
    }
    
    // 获取nonce
    nonce, err := srcClient.ethClient.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        return nil, err
    }
    
    // 构建发送交易数据
    data, err := s.client.oftABI.Pack(
        "sendFrom",
        fromAddress,
        dstChainID,
        to.Bytes(),
        amount,
        fromAddress, // refundAddress
        common.Address{}, // zroPaymentAddress
        adapterParams,
    )
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(
        nonce,
        oftContract,
        feeEstimate.NativeFee,
        gasLimit,
        srcClient.config.GasPrice,
        data,
    )
    
    return s.signAndSendTransaction(srcChainID, tx)
}

// 获取OFT余额
func (s *OFTService) GetOFTBalance(
    chainID uint16,
    oftContract common.Address,
    account common.Address,
) (*big.Int, error) {
    chainClient, err := s.client.GetChainClient(chainID)
    if err != nil {
        return nil, err
    }
    
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    var balance *big.Int
    err = chainClient.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &oftContract,
        Data: s.client.oftABI.Methods["balanceOf"].ID,
    }, nil)
    if err != nil {
        return nil, err
    }
    
    return balance, nil
}

// 获取OFT信息
func (s *OFTService) GetOFTInfo(
    chainID uint16,
    oftContract common.Address,
) (*OmniToken, error) {
    chainClient, err := s.client.GetChainClient(chainID)
    if err != nil {
        return nil, err
    }
    
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    // 获取代币基本信息
    var name string
    err = chainClient.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &oftContract,
        Data: s.client.oftABI.Methods["name"].ID,
    }, nil)
    if err != nil {
        return nil, err
    }
    
    var symbol string
    err = chainClient.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &oftContract,
        Data: s.client.oftABI.Methods["symbol"].ID,
    }, nil)
    if err != nil {
        return nil, err
    }
    
    var decimals uint8
    err = chainClient.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &oftContract,
        Data: s.client.oftABI.Methods["decimals"].ID,
    }, nil)
    if err != nil {
        return nil, err
    }
    
    var totalSupply *big.Int
    err = chainClient.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &oftContract,
        Data: s.client.oftABI.Methods["totalSupply"].ID,
    }, nil)
    if err != nil {
        return nil, err
    }
    
    var sharedDecimals uint8
    err = chainClient.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &oftContract,
        Data: s.client.oftABI.Methods["sharedDecimals"].ID,
    }, nil)
    if err != nil {
        return nil, err
    }
    
    token := &OmniToken{
        Name:           name,
        Symbol:         symbol,
        Decimals:       decimals,
        TotalSupply:    totalSupply,
        IsOFT:          true,
        SharedDecimals: sharedDecimals,
        Deployments:    make(map[uint16]common.Address),
    }
    
    token.Deployments[chainID] = oftContract
    
    return token, nil
}

// 批量设置可信远程
func (s *OFTService) BatchSetTrustedRemotes(
    chainID uint16,
    oftContract common.Address,
    remotes map[uint16]common.Address,
) ([]*types.Transaction, error) {
    var transactions []*types.Transaction
    
    for remoteChainID, remoteAddress := range remotes {
        tx, err := s.SetTrustedRemote(chainID, oftContract, remoteChainID, remoteAddress)
        if err != nil {
            return nil, err
        }
        
        transactions = append(transactions, tx)
        
        // 等待一段时间避免nonce冲突
        time.Sleep(1 * time.Second)
    }
    
    return transactions, nil
}

// 计算跨链转账费用
func (s *OFTService) CalculateOFTSendCost(
    srcChainID uint16,
    dstChainID uint16,
    oftContract common.Address,
    amount *big.Int,
    gasLimit uint64,
) (*CrossChainCost, error) {
    // 构建适配器参数
    adapterParams := s.buildAdapterParams(1, gasLimit, big.NewInt(0))
    
    // 估算LayerZero费用
    feeEstimate, err := s.estimateOFTSendFee(
        srcChainID,
        oftContract,
        dstChainID,
        make([]byte, 20), // 占位符地址
        amount,
        false,
        adapterParams,
    )
    if err != nil {
        return nil, err
    }
    
    // 获取源链gas价格
    srcClient, err := s.client.GetChainClient(srcChainID)
    if err != nil {
        return nil, err
    }
    
    gasPrice, err := srcClient.ethClient.SuggestGasPrice(context.Background())
    if err != nil {
        return nil, err
    }
    
    // 计算总成本
    gasCost := new(big.Int).Mul(big.NewInt(int64(gasLimit)), gasPrice)
    totalCost := new(big.Int).Add(feeEstimate.NativeFee, gasCost)
    
    return &CrossChainCost{
        LayerZeroFee: feeEstimate.NativeFee,
        GasCost:      gasCost,
        TotalCost:    totalCost,
        GasLimit:     gasLimit,
        GasPrice:     gasPrice,
    }, nil
}

// 辅助函数
func (s *OFTService) buildAdapterParams(version uint16, gasLimit uint64, nativeForDst *big.Int) []byte {
    // 与MessageService中的实现相同
    params := make([]byte, 0)
    
    versionBytes := make([]byte, 2)
    binary.BigEndian.PutUint16(versionBytes, version)
    params = append(params, versionBytes...)
    
    gasLimitBytes := make([]byte, 32)
    big.NewInt(int64(gasLimit)).FillBytes(gasLimitBytes)
    params = append(params, gasLimitBytes...)
    
    nativeBytes := make([]byte, 32)
    if nativeForDst != nil {
        nativeForDst.FillBytes(nativeBytes)
    }
    params = append(params, nativeBytes...)
    
    return params
}

func (s *OFTService) estimateOFTSendFee(
    srcChainID uint16,
    oftContract common.Address,
    dstChainID uint16,
    toAddress []byte,
    amount *big.Int,
    useZro bool,
    adapterParams []byte,
) (*FeeEstimate, error) {
    srcClient, err := s.client.GetChainClient(srcChainID)
    if err != nil {
        return nil, err
    }
    
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    var result []interface{}
    err = srcClient.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &oftContract,
        Data: s.client.oftABI.Methods["estimateSendFee"].ID,
    }, nil)
    if err != nil {
        return nil, err
    }
    
    nativeFee := result[0].(*big.Int)
    zroFee := result[1].(*big.Int)
    
    return &FeeEstimate{
        NativeFee: nativeFee,
        ZroFee:    zroFee,
        TotalFee:  new(big.Int).Add(nativeFee, zroFee),
    }, nil
}

func (s *OFTService) signAndSendTransaction(chainID uint16, tx *types.Transaction) (*types.Transaction, error) {
    chainClient, err := s.client.GetChainClient(chainID)
    if err != nil {
        return nil, err
    }
    
    networkID, err := chainClient.ethClient.NetworkID(context.Background())
    if err != nil {
        return nil, err
    }
    
    signedTx, err := types.SignTx(tx, types.NewEIP155Signer(networkID), s.privateKey)
    if err != nil {
        return nil, err
    }
    
    err = chainClient.ethClient.SendTransaction(context.Background(), signedTx)
    if err != nil {
        return nil, err
    }
    
    return signedTx, nil
}

type CrossChainCost struct {
    LayerZeroFee *big.Int
    GasCost      *big.Int
    TotalCost    *big.Int
    GasLimit     uint64
    GasPrice     *big.Int
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
    // 创建LayerZero客户端
    lzClient, err := client.NewLayerZeroClient()
    if err != nil {
        log.Fatal("创建LayerZero客户端失败:", err)
    }
    
    // 加载私钥
    privateKey, err := crypto.HexToECDSA("your_private_key_here")
    if err != nil {
        log.Fatal("加载私钥失败:", err)
    }
    
    // 获取用户地址
    userAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
    
    // 创建服务
    messageService := services.NewLayerZeroMessageService(lzClient, privateKey)
    oftService := services.NewOFTService(lzClient, privateKey)
    
    // 1. 查询支持的链
    fmt.Printf("=== LayerZero支持的区块链 ===\n")
    
    for chainID, config := range client.SupportedChains {
        fmt.Printf("链ID %d: %s\n", chainID, config.Name)
        fmt.Printf("  RPC: %s\n", config.RPC)
        fmt.Printf("  端点: %s\n", config.Endpoint.Hex())
        fmt.Printf("  Gas限制: %d\n", config.GasLimit)
        fmt.Printf("  Gas价格: %s Gwei\n", 
            decimal.NewFromBigInt(config.GasPrice, -9).String())
        
        // 检查连接状态
        chainClient, err := lzClient.GetChainClient(chainID)
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
    
    // 2. 费用估算示例
    fmt.Printf("=== 跨链费用估算 ===\n")
    
    routes := []struct {
        SrcChain uint16
        DstChain uint16
        SrcName  string
        DstName  string
    }{
        {client.EthereumChainID, client.BSCChainID, "Ethereum", "BSC"},
        {client.EthereumChainID, client.PolygonChainID, "Ethereum", "Polygon"},
        {client.BSCChainID, client.PolygonChainID, "BSC", "Polygon"},
    }
    
    testPayload := []byte("Hello LayerZero!")
    testAdapterParams := []byte{0x00, 0x01} // 简化的适配器参数
    
    for _, route := range routes {
        feeEstimate, err := lzClient.EstimateFees(
            route.SrcChain,
            route.DstChain,
            userAddress,
            testPayload,
            false,
            testAdapterParams,
        )
        if err != nil {
            fmt.Printf("%s -> %s: 估算失败 - %v\n", route.SrcName, route.DstName, err)
        } else {
            fmt.Printf("%s -> %s:\n", route.SrcName, route.DstName)
            fmt.Printf("  原生代币费用: %s ETH\n", 
                decimal.NewFromBigInt(feeEstimate.NativeFee, -18).String())
            fmt.Printf("  ZRO费用: %s\n", feeEstimate.ZroFee.String())
            fmt.Printf("  总费用: %s ETH\n", 
                decimal.NewFromBigInt(feeEstimate.TotalFee, -18).String())
        }
    }
    
    // 3. 发送跨链消息示例
    fmt.Printf("\n=== 跨链消息发送示例 ===\n")
    
    srcChain := client.EthereumChainID
    dstChain := client.BSCChainID
    message := "Hello from Ethereum to BSC!"
    payload := []byte(message)
    
    fmt.Printf("准备发送消息:\n")
    fmt.Printf("  源链: %s (ID: %d)\n", "Ethereum", srcChain)
    fmt.Printf("  目标链: %s (ID: %d)\n", "BSC", dstChain)
    fmt.Printf("  消息: %s\n", message)
    fmt.Printf("  接收者: %s\n", userAddress.Hex())
    
    // 发送消息
    tx, err := messageService.SendMessage(
        srcChain,
        dstChain,
        userAddress,
        payload,
        200000, // gasLimit
        big.NewInt(0), // nativeForDst
    )
    if err != nil {
        log.Printf("发送消息失败: %v", err)
    } else {
        fmt.Printf("消息发送交易已提交: %s\n", tx.Hash().Hex())
        
        // 跟踪消息状态
        status, err := messageService.TrackMessage(srcChain, tx.Hash())
        if err != nil {
            log.Printf("跟踪消息失败: %v", err)
        } else {
            fmt.Printf("消息状态: %s\n", status.Status)
        }
    }
    
    // 4. OFT代币操作示例
    fmt.Printf("\n=== OFT代币操作示例 ===\n")
    
    // 假设的OFT合约地址
    oftContractEth := common.HexToAddress("0x1234567890123456789012345678901234567890")
    oftContractBSC := common.HexToAddress("0x0987654321098765432109876543210987654321")
    
    // 获取OFT信息
    oftInfo, err := oftService.GetOFTInfo(srcChain, oftContractEth)
    if err != nil {
        log.Printf("获取OFT信息失败: %v", err)
    } else {
        fmt.Printf("OFT代币信息:\n")
        fmt.Printf("  名称: %s\n", oftInfo.Name)
        fmt.Printf("  符号: %s\n", oftInfo.Symbol)
        fmt.Printf("  精度: %d\n", oftInfo.Decimals)
        fmt.Printf("  共享精度: %d\n", oftInfo.SharedDecimals)
        fmt.Printf("  总供应量: %s\n", oftInfo.TotalSupply.String())
    }
    
    // 查询OFT余额
    oftBalance, err := oftService.GetOFTBalance(srcChain, oftContractEth, userAddress)
    if err != nil {
        log.Printf("查询OFT余额失败: %v", err)
    } else {
        fmt.Printf("OFT余额: %s\n", oftBalance.String())
        
        if oftBalance.Cmp(big.NewInt(0)) > 0 {
            // 计算跨链转账费用
            transferAmount := big.NewInt(1e18) // 1个代币
            
            cost, err := oftService.CalculateOFTSendCost(
                srcChain,
                dstChain,
                oftContractEth,
                transferAmount,
                200000,
            )
            if err != nil {
                log.Printf("计算转账费用失败: %v", err)
            } else {
                fmt.Printf("跨链转账费用分析:\n")
                fmt.Printf("  LayerZero费用: %s ETH\n", 
                    decimal.NewFromBigInt(cost.LayerZeroFee, -18).String())
                fmt.Printf("  Gas费用: %s ETH\n", 
                    decimal.NewFromBigInt(cost.GasCost, -18).String())
                fmt.Printf("  总费用: %s ETH\n", 
                    decimal.NewFromBigInt(cost.TotalCost, -18).String())
                
                // 执行OFT跨链转账
                fmt.Printf("准备跨链转账 %s 个代币\n", 
                    decimal.NewFromBigInt(transferAmount, -18).String())
                
                // oftTx, err := oftService.SendOFT(
                //     srcChain,
                //     dstChain,
                //     oftContractEth,
                //     userAddress,
                //     transferAmount,
                //     200000,
                //     big.NewInt(0),
                // )
                // if err != nil {
                //     log.Printf("OFT跨链转账失败: %v", err)
                // } else {
                //     fmt.Printf("OFT转账交易已提交: %s\n", oftTx.Hash().Hex())
                // }
            }
        }
    }
    
    // 5. 应用配置查询
    fmt.Printf("\n=== 应用配置查询 ===\n")
    
    appConfig, err := lzClient.GetAppConfig(srcChain, oftContractEth, dstChain)
    if err != nil {
        log.Printf("获取应用配置失败: %v", err)
    } else {
        fmt.Printf("应用配置:\n")
        fmt.Printf("  入站证明库版本: %d\n", appConfig.InboundProofLibraryVersion)
        fmt.Printf("  入站区块确认数: %d\n", appConfig.InboundBlockConfirmations)
        fmt.Printf("  中继器: %s\n", appConfig.Relayer.Hex())
        fmt.Printf("  出站证明类型: %d\n", appConfig.OutboundProofType)
        fmt.Printf("  出站区块确认数: %d\n", appConfig.OutboundBlockConfirmations)
        fmt.Printf("  预言机: %s\n", appConfig.Oracle.Hex())
    }
    
    // 6. 批量操作示例
    fmt.Printf("\n=== 批量操作示例 ===\n")
    
    // 批量发送消息到多个链
    batchMessages := []services.CrossChainMessage{
        {
            SrcChainID:   srcChain,
            DstChainID:   client.BSCChainID,
            DstAddress:   userAddress,
            Payload:      []byte("Message to BSC"),
            GasLimit:     200000,
            NativeForDst: big.NewInt(0),
        },
        {
            SrcChainID:   srcChain,
            DstChainID:   client.PolygonChainID,
            DstAddress:   userAddress,
            Payload:      []byte("Message to Polygon"),
            GasLimit:     200000,
            NativeForDst: big.NewInt(0),
        },
    }
    
    fmt.Printf("准备批量发送 %d 条消息\n", len(batchMessages))
    
    // batchTxs, err := messageService.BatchSendMessages(batchMessages)
    // if err != nil {
    //     log.Printf("批量发送失败: %v", err)
    // } else {
    //     fmt.Printf("批量发送成功，交易数量: %d\n", len(batchTxs))
    //     for i, tx := range batchTxs {
    //         fmt.Printf("  交易 %d: %s\n", i+1, tx.Hash().Hex())
    //     }
    // }
    
    // 7. LayerZero特性总结
    fmt.Printf("\n=== LayerZero特性总结 ===\n")
    
    fmt.Printf("LayerZero优势:\n")
    fmt.Printf("  - 真正的全链互操作性\n")
    fmt.Printf("  - 超轻节点架构，无需运行完整节点\n")
    fmt.Printf("  - 可配置的安全性和去中心化程度\n")
    fmt.Printf("  - 支持任意消息传递，不仅限于代币\n")
    fmt.Printf("  - 统一的开发体验\n")
    
    fmt.Printf("\n应用场景:\n")
    fmt.Printf("  - 全链代币 (OFT)\n")
    fmt.Printf("  - 跨链DeFi协议\n")
    fmt.Printf("  - 全链NFT\n")
    fmt.Printf("  - 跨链治理\n")
    fmt.Printf("  - 多链数据同步\n")
    
    // 8. 最佳实践建议
    fmt.Printf("\n=== 最佳实践建议 ===\n")
    
    fmt.Printf("使用LayerZero时请注意:\n")
    fmt.Printf("  1. 合理设置gas限制和适配器参数\n")
    fmt.Printf("  2. 验证可信远程地址的正确性\n")
    fmt.Printf("  3. 监控消息传递状态\n")
    fmt.Printf("  4. 考虑网络拥堵对费用的影响\n")
    fmt.Printf("  5. 实现适当的错误处理和重试机制\n")
    fmt.Printf("  6. 测试网充分验证后再部署主网\n")
    fmt.Printf("  7. 关注LayerZero协议升级\n")
    fmt.Printf("  8. 保持预言机和中继器的多样性\n")
}
```

这个LayerZero使用指南提供了完整的全链互操作性协议集成方案，涵盖了消息传递、OFT代币、全链应用开发、安全配置等核心功能，是构建真正全链DeFi应用的重要参考文档。
