# Cosmos IBC 跨链协议 详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [环境准备](#环境准备)
3. [IBC协议架构](#ibc协议架构)
4. [跨链转账](#跨链转账)
5. [跨链合约调用](#跨链合约调用)
6. [中继器操作](#中继器操作)
7. [安全机制](#安全机制)
8. [监控和调试](#监控和调试)
9. [实际应用](#实际应用)

## 基础概念

### 1.1 Cosmos IBC 简介

Inter-Blockchain Communication (IBC) 是 Cosmos 生态的跨链通信协议，实现了不同区块链之间的安全、可靠的数据和资产传输。

```bash
# 安装Cosmos IBC相关依赖
go get github.com/cosmos/cosmos-sdk
go get github.com/cosmos/ibc-go/v7
go get github.com/cosmos/relayer/v2
go get github.com/tendermint/tendermint
```

### 1.2 核心概念

```go
// 主要包导入
import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    
    "github.com/cosmos/cosmos-sdk/client"
    "github.com/cosmos/cosmos-sdk/codec"
    "github.com/cosmos/cosmos-sdk/types"
    "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
    "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
    "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
    "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
    transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
)

// IBC 客户端类型
const (
    TendermintClientType = "07-tendermint"
    SoloMachineClientType = "06-solomachine"
    LocalhostClientType = "09-localhost"
)

// IBC 连接状态
const (
    ConnectionStateInit    = "STATE_INIT"
    ConnectionStateTryOpen = "STATE_TRYOPEN"
    ConnectionStateOpen    = "STATE_OPEN"
)

// IBC 通道状态
const (
    ChannelStateInit     = "STATE_INIT"
    ChannelStateTryOpen  = "STATE_TRYOPEN"
    ChannelStateOpen     = "STATE_OPEN"
    ChannelStateClosed   = "STATE_CLOSED"
)

// IBC 客户端信息
type IBCClient struct {
    ClientID    string
    ClientType  string
    ChainID     string
    Height      uint64
    TrustLevel  string
    TrustPeriod time.Duration
    MaxClockDrift time.Duration
    FrozenHeight uint64
    LatestHeight uint64
}

// IBC 连接信息
type IBCConnection struct {
    ConnectionID string
    ClientID     string
    State        string
    Counterparty *ConnectionCounterparty
    DelayPeriod  uint64
}

type ConnectionCounterparty struct {
    ClientID     string
    ConnectionID string
    Prefix       string
}

// IBC 通道信息
type IBCChannel struct {
    ChannelID    string
    PortID       string
    State        string
    Ordering     string
    Counterparty *ChannelCounterparty
    ConnectionHops []string
    Version      string
}

type ChannelCounterparty struct {
    PortID    string
    ChannelID string
}

// IBC 数据包
type IBCPacket struct {
    Sequence           uint64
    SourcePort         string
    SourceChannel      string
    DestinationPort    string
    DestinationChannel string
    Data               []byte
    TimeoutHeight      uint64
    TimeoutTimestamp   uint64
}

// 跨链转账信息
type IBCTransfer struct {
    Sender        string
    Receiver      string
    Token         types.Coin
    SourcePort    string
    SourceChannel string
    TimeoutHeight uint64
    TimeoutTimestamp uint64
    Memo          string
}
```

## 环境准备

### 2.1 IBC客户端设置

```go
// client/ibc_client.go
package client

import (
    "context"
    "encoding/json"
    "fmt"
    
    "github.com/cosmos/cosmos-sdk/client"
    "github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
    "github.com/cosmos/cosmos-sdk/codec"
    "github.com/cosmos/cosmos-sdk/types"
    "github.com/cosmos/cosmos-sdk/types/query"
    clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
    connectiontypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
    channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
    transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
    "google.golang.org/grpc"
)

type IBCClient struct {
    clientCtx     client.Context
    grpcConn      *grpc.ClientConn
    chainID       string
    rpcEndpoint   string
    
    // IBC查询客户端
    clientQuery     clienttypes.QueryClient
    connectionQuery connectiontypes.QueryClient
    channelQuery    channeltypes.QueryClient
    transferQuery   transfertypes.QueryClient
    tmQuery         tmservice.ServiceClient
}

func NewIBCClient(chainID, rpcEndpoint string, cdc codec.Codec) (*IBCClient, error) {
    // 创建客户端上下文
    clientCtx := client.Context{}.
        WithCodec(cdc).
        WithChainID(chainID).
        WithNodeURI(rpcEndpoint)
    
    // 建立gRPC连接
    grpcConn, err := grpc.Dial(rpcEndpoint, grpc.WithInsecure())
    if err != nil {
        return nil, err
    }
    
    return &IBCClient{
        clientCtx:       clientCtx,
        grpcConn:        grpcConn,
        chainID:         chainID,
        rpcEndpoint:     rpcEndpoint,
        clientQuery:     clienttypes.NewQueryClient(grpcConn),
        connectionQuery: connectiontypes.NewQueryClient(grpcConn),
        channelQuery:    channeltypes.NewQueryClient(grpcConn),
        transferQuery:   transfertypes.NewQueryClient(grpcConn),
        tmQuery:         tmservice.NewServiceClient(grpcConn),
    }, nil
}

// 获取IBC客户端信息
func (c *IBCClient) GetClient(clientID string) (*IBCClient, error) {
    req := &clienttypes.QueryClientStateRequest{
        ClientId: clientID,
    }
    
    resp, err := c.clientQuery.ClientState(context.Background(), req)
    if err != nil {
        return nil, err
    }
    
    var clientState clienttypes.ClientState
    err = c.clientCtx.Codec.UnpackAny(resp.ClientState, &clientState)
    if err != nil {
        return nil, err
    }
    
    // 解析客户端状态
    client := &IBCClient{
        ClientID:   clientID,
        ClientType: clientState.ClientType(),
    }
    
    return client, nil
}

// 获取所有IBC客户端
func (c *IBCClient) GetAllClients() ([]*IBCClient, error) {
    req := &clienttypes.QueryClientStatesRequest{
        Pagination: &query.PageRequest{
            Limit: 100,
        },
    }
    
    resp, err := c.clientQuery.ClientStates(context.Background(), req)
    if err != nil {
        return nil, err
    }
    
    var clients []*IBCClient
    for _, clientState := range resp.ClientStates {
        var cs clienttypes.ClientState
        err = c.clientCtx.Codec.UnpackAny(clientState.ClientState, &cs)
        if err != nil {
            continue
        }
        
        client := &IBCClient{
            ClientID:   clientState.ClientId,
            ClientType: cs.ClientType(),
        }
        clients = append(clients, client)
    }
    
    return clients, nil
}

// 获取IBC连接信息
func (c *IBCClient) GetConnection(connectionID string) (*IBCConnection, error) {
    req := &connectiontypes.QueryConnectionRequest{
        ConnectionId: connectionID,
    }
    
    resp, err := c.connectionQuery.Connection(context.Background(), req)
    if err != nil {
        return nil, err
    }
    
    connection := &IBCConnection{
        ConnectionID: connectionID,
        ClientID:     resp.Connection.ClientId,
        State:        resp.Connection.State.String(),
        DelayPeriod:  resp.Connection.DelayPeriod,
    }
    
    if resp.Connection.Counterparty != nil {
        connection.Counterparty = &ConnectionCounterparty{
            ClientID:     resp.Connection.Counterparty.ClientId,
            ConnectionID: resp.Connection.Counterparty.ConnectionId,
            Prefix:       resp.Connection.Counterparty.Prefix.String(),
        }
    }
    
    return connection, nil
}

// 获取所有IBC连接
func (c *IBCClient) GetAllConnections() ([]*IBCConnection, error) {
    req := &connectiontypes.QueryConnectionsRequest{
        Pagination: &query.PageRequest{
            Limit: 100,
        },
    }
    
    resp, err := c.connectionQuery.Connections(context.Background(), req)
    if err != nil {
        return nil, err
    }
    
    var connections []*IBCConnection
    for _, conn := range resp.Connections {
        connection := &IBCConnection{
            ConnectionID: conn.Id,
            ClientID:     conn.ClientId,
            State:        conn.State.String(),
            DelayPeriod:  conn.DelayPeriod,
        }
        
        if conn.Counterparty != nil {
            connection.Counterparty = &ConnectionCounterparty{
                ClientID:     conn.Counterparty.ClientId,
                ConnectionID: conn.Counterparty.ConnectionId,
                Prefix:       conn.Counterparty.Prefix.String(),
            }
        }
        
        connections = append(connections, connection)
    }
    
    return connections, nil
}

// 获取IBC通道信息
func (c *IBCClient) GetChannel(portID, channelID string) (*IBCChannel, error) {
    req := &channeltypes.QueryChannelRequest{
        PortId:    portID,
        ChannelId: channelID,
    }
    
    resp, err := c.channelQuery.Channel(context.Background(), req)
    if err != nil {
        return nil, err
    }
    
    channel := &IBCChannel{
        ChannelID:      channelID,
        PortID:         portID,
        State:          resp.Channel.State.String(),
        Ordering:       resp.Channel.Ordering.String(),
        ConnectionHops: resp.Channel.ConnectionHops,
        Version:        resp.Channel.Version,
    }
    
    if resp.Channel.Counterparty != nil {
        channel.Counterparty = &ChannelCounterparty{
            PortID:    resp.Channel.Counterparty.PortId,
            ChannelID: resp.Channel.Counterparty.ChannelId,
        }
    }
    
    return channel, nil
}

// 获取所有IBC通道
func (c *IBCClient) GetAllChannels() ([]*IBCChannel, error) {
    req := &channeltypes.QueryChannelsRequest{
        Pagination: &query.PageRequest{
            Limit: 100,
        },
    }
    
    resp, err := c.channelQuery.Channels(context.Background(), req)
    if err != nil {
        return nil, err
    }
    
    var channels []*IBCChannel
    for _, ch := range resp.Channels {
        channel := &IBCChannel{
            ChannelID:      ch.ChannelId,
            PortID:         ch.PortId,
            State:          ch.State.String(),
            Ordering:       ch.Ordering.String(),
            ConnectionHops: ch.ConnectionHops,
            Version:        ch.Version,
        }
        
        if ch.Counterparty != nil {
            channel.Counterparty = &ChannelCounterparty{
                PortID:    ch.Counterparty.PortId,
                ChannelID: ch.Counterparty.ChannelId,
            }
        }
        
        channels = append(channels, channel)
    }
    
    return channels, nil
}
```

## 跨链转账

### 3.1 IBC转账服务

```go
// services/ibc_transfer_service.go
package services

import (
    "context"
    "fmt"
    "time"
    
    "github.com/cosmos/cosmos-sdk/client"
    "github.com/cosmos/cosmos-sdk/client/tx"
    "github.com/cosmos/cosmos-sdk/types"
    transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
    clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
)

type IBCTransferService struct {
    client    *IBCClient
    clientCtx client.Context
}

func NewIBCTransferService(client *IBCClient, clientCtx client.Context) *IBCTransferService {
    return &IBCTransferService{
        client:    client,
        clientCtx: clientCtx,
    }
}

// 执行IBC转账
func (s *IBCTransferService) Transfer(
    sender string,
    receiver string,
    token types.Coin,
    sourcePort string,
    sourceChannel string,
    timeoutHeight uint64,
    timeoutTimestamp uint64,
    memo string,
) (*types.TxResponse, error) {
    // 创建转账消息
    msg := transfertypes.NewMsgTransfer(
        sourcePort,
        sourceChannel,
        token,
        sender,
        receiver,
        clienttypes.Height{
            RevisionNumber: 0,
            RevisionHeight: timeoutHeight,
        },
        timeoutTimestamp,
        memo,
    )
    
    // 构建交易
    txBuilder := s.clientCtx.TxConfig.NewTxBuilder()
    err := txBuilder.SetMsgs(msg)
    if err != nil {
        return nil, err
    }
    
    // 设置gas和费用
    txBuilder.SetGasLimit(200000)
    txBuilder.SetFeeAmount(types.NewCoins(types.NewCoin("uatom", types.NewInt(5000))))
    
    // 签名并广播交易
    txBytes, err := tx.Sign(s.clientCtx, txBuilder.GetTx())
    if err != nil {
        return nil, err
    }
    
    // 广播交易
    res, err := s.clientCtx.BroadcastTx(txBytes)
    if err != nil {
        return nil, err
    }
    
    return res, nil
}

// 查询IBC转账历史
func (s *IBCTransferService) GetTransferHistory(
    address string,
    limit int,
) ([]*IBCTransferRecord, error) {
    // 这里需要查询链上的转账事件
    // 实际实现需要解析区块和交易事件
    
    var records []*IBCTransferRecord
    
    // 示例查询逻辑
    // 实际需要从链上事件日志中解析
    
    return records, nil
}

// 计算超时高度
func (s *IBCTransferService) CalculateTimeoutHeight(
    currentHeight uint64,
    blocksToAdd uint64,
) uint64 {
    return currentHeight + blocksToAdd
}

// 计算超时时间戳
func (s *IBCTransferService) CalculateTimeoutTimestamp(
    duration time.Duration,
) uint64 {
    return uint64(time.Now().Add(duration).UnixNano())
}

// 验证IBC路径
func (s *IBCTransferService) ValidateIBCPath(
    sourcePort string,
    sourceChannel string,
    destinationChainID string,
) error {
    // 获取通道信息
    channel, err := s.client.GetChannel(sourcePort, sourceChannel)
    if err != nil {
        return fmt.Errorf("获取通道信息失败: %w", err)
    }
    
    // 检查通道状态
    if channel.State != ChannelStateOpen {
        return fmt.Errorf("通道状态不是OPEN: %s", channel.State)
    }
    
    // 获取连接信息
    if len(channel.ConnectionHops) == 0 {
        return fmt.Errorf("通道没有连接跳数")
    }
    
    connection, err := s.client.GetConnection(channel.ConnectionHops[0])
    if err != nil {
        return fmt.Errorf("获取连接信息失败: %w", err)
    }
    
    // 检查连接状态
    if connection.State != ConnectionStateOpen {
        return fmt.Errorf("连接状态不是OPEN: %s", connection.State)
    }
    
    return nil
}

// 跟踪IBC数据包
func (s *IBCTransferService) TrackPacket(
    sourcePort string,
    sourceChannel string,
    sequence uint64,
) (*PacketStatus, error) {
    // 查询数据包确认
    req := &channeltypes.QueryPacketAcknowledgementRequest{
        PortId:    sourcePort,
        ChannelId: sourceChannel,
        Sequence:  sequence,
    }
    
    resp, err := s.client.channelQuery.PacketAcknowledgement(context.Background(), req)
    if err != nil {
        // 数据包可能还未被确认
        return &PacketStatus{
            Sequence: sequence,
            Status:   "PENDING",
        }, nil
    }
    
    // 解析确认数据
    var ackResult AcknowledgementResult
    err = json.Unmarshal(resp.Acknowledgement, &ackResult)
    if err != nil {
        return nil, err
    }
    
    status := "SUCCESS"
    if ackResult.Error != "" {
        status = "FAILED"
    }
    
    return &PacketStatus{
        Sequence:        sequence,
        Status:          status,
        Acknowledgement: string(resp.Acknowledgement),
        Error:           ackResult.Error,
    }, nil
}

type IBCTransferRecord struct {
    TxHash            string
    Sender            string
    Receiver          string
    Token             types.Coin
    SourcePort        string
    SourceChannel     string
    DestinationPort   string
    DestinationChannel string
    Sequence          uint64
    Status            string
    Timestamp         time.Time
}

type PacketStatus struct {
    Sequence        uint64
    Status          string
    Acknowledgement string
    Error           string
}

type AcknowledgementResult struct {
    Result string `json:"result,omitempty"`
    Error  string `json:"error,omitempty"`
}
```

## 中继器操作

### 4.1 中继器服务

```go
// services/relayer_service.go
package services

import (
    "context"
    "fmt"
    "time"
    
    "github.com/cosmos/relayer/v2/relayer"
    "github.com/cosmos/relayer/v2/relayer/chains/cosmos"
)

type RelayerService struct {
    relayer *relayer.Relayer
    chains  map[string]*cosmos.CosmosProvider
}

func NewRelayerService() *RelayerService {
    return &RelayerService{
        chains: make(map[string]*cosmos.CosmosProvider),
    }
}

// 添加链配置
func (s *RelayerService) AddChain(
    chainID string,
    rpcAddr string,
    grpcAddr string,
    keyName string,
) error {
    // 创建链提供者配置
    config := &cosmos.CosmosProviderConfig{
        ChainID:  chainID,
        RPCAddr:  rpcAddr,
        GRPCAddr: grpcAddr,
        KeyName:  keyName,
    }
    
    // 创建链提供者
    provider, err := cosmos.NewProvider(config)
    if err != nil {
        return err
    }
    
    s.chains[chainID] = provider
    return nil
}

// 创建客户端
func (s *RelayerService) CreateClient(
    srcChainID string,
    dstChainID string,
    trustingPeriod time.Duration,
) error {
    srcChain, exists := s.chains[srcChainID]
    if !exists {
        return fmt.Errorf("源链 %s 未配置", srcChainID)
    }
    
    dstChain, exists := s.chains[dstChainID]
    if !exists {
        return fmt.Errorf("目标链 %s 未配置", dstChainID)
    }
    
    // 创建客户端
    err := srcChain.CreateClient(context.Background(), dstChain, trustingPeriod)
    if err != nil {
        return fmt.Errorf("创建客户端失败: %w", err)
    }
    
    return nil
}

// 创建连接
func (s *RelayerService) CreateConnection(
    srcChainID string,
    dstChainID string,
    srcClientID string,
    dstClientID string,
) error {
    srcChain, exists := s.chains[srcChainID]
    if !exists {
        return fmt.Errorf("源链 %s 未配置", srcChainID)
    }
    
    dstChain, exists := s.chains[dstChainID]
    if !exists {
        return fmt.Errorf("目标链 %s 未配置", dstChainID)
    }
    
    // 创建连接
    err := srcChain.CreateConnection(
        context.Background(),
        dstChain,
        srcClientID,
        dstClientID,
    )
    if err != nil {
        return fmt.Errorf("创建连接失败: %w", err)
    }
    
    return nil
}

// 创建通道
func (s *RelayerService) CreateChannel(
    srcChainID string,
    dstChainID string,
    srcConnectionID string,
    dstConnectionID string,
    srcPortID string,
    dstPortID string,
    version string,
    ordering string,
) error {
    srcChain, exists := s.chains[srcChainID]
    if !exists {
        return fmt.Errorf("源链 %s 未配置", srcChainID)
    }
    
    dstChain, exists := s.chains[dstChainID]
    if !exists {
        return fmt.Errorf("目标链 %s 未配置", dstChainID)
    }
    
    // 创建通道
    err := srcChain.CreateChannel(
        context.Background(),
        dstChain,
        srcConnectionID,
        dstConnectionID,
        srcPortID,
        dstPortID,
        version,
        ordering,
    )
    if err != nil {
        return fmt.Errorf("创建通道失败: %w", err)
    }
    
    return nil
}

// 启动数据包中继
func (s *RelayerService) StartPacketRelay(
    srcChainID string,
    dstChainID string,
    srcChannelID string,
    dstChannelID string,
    srcPortID string,
    dstPortID string,
) error {
    srcChain, exists := s.chains[srcChainID]
    if !exists {
        return fmt.Errorf("源链 %s 未配置", srcChainID)
    }
    
    dstChain, exists := s.chains[dstChainID]
    if !exists {
        return fmt.Errorf("目标链 %s 未配置", dstChainID)
    }
    
    // 启动中继循环
    go s.relayPackets(srcChain, dstChain, srcChannelID, dstChannelID, srcPortID, dstPortID)
    
    return nil
}

// 中继数据包
func (s *RelayerService) relayPackets(
    srcChain *cosmos.CosmosProvider,
    dstChain *cosmos.CosmosProvider,
    srcChannelID string,
    dstChannelID string,
    srcPortID string,
    dstPortID string,
) {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        // 查询未中继的数据包
        packets, err := s.queryUnrelayedPackets(srcChain, srcChannelID, srcPortID)
        if err != nil {
            fmt.Printf("查询未中继数据包失败: %v\n", err)
            continue
        }
        
        // 中继每个数据包
        for _, packet := range packets {
            err := s.relayPacket(srcChain, dstChain, packet)
            if err != nil {
                fmt.Printf("中继数据包失败: %v\n", err)
            }
        }
        
        // 查询未中继的确认
        acks, err := s.queryUnrelayedAcknowledgements(dstChain, dstChannelID, dstPortID)
        if err != nil {
            fmt.Printf("查询未中继确认失败: %v\n", err)
            continue
        }
        
        // 中继每个确认
        for _, ack := range acks {
            err := s.relayAcknowledgement(dstChain, srcChain, ack)
            if err != nil {
                fmt.Printf("中继确认失败: %v\n", err)
            }
        }
    }
}

// 查询未中继的数据包
func (s *RelayerService) queryUnrelayedPackets(
    chain *cosmos.CosmosProvider,
    channelID string,
    portID string,
) ([]*IBCPacket, error) {
    // 实现查询逻辑
    // 这里需要查询链上的未中继数据包
    return nil, nil
}

// 查询未中继的确认
func (s *RelayerService) queryUnrelayedAcknowledgements(
    chain *cosmos.CosmosProvider,
    channelID string,
    portID string,
) ([]*PacketAcknowledgement, error) {
    // 实现查询逻辑
    return nil, nil
}

// 中继单个数据包
func (s *RelayerService) relayPacket(
    srcChain *cosmos.CosmosProvider,
    dstChain *cosmos.CosmosProvider,
    packet *IBCPacket,
) error {
    // 实现数据包中继逻辑
    return nil
}

// 中继确认
func (s *RelayerService) relayAcknowledgement(
    srcChain *cosmos.CosmosProvider,
    dstChain *cosmos.CosmosProvider,
    ack *PacketAcknowledgement,
) error {
    // 实现确认中继逻辑
    return nil
}

type PacketAcknowledgement struct {
    Packet          *IBCPacket
    Acknowledgement []byte
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
    "time"
    
    "github.com/cosmos/cosmos-sdk/client"
    "github.com/cosmos/cosmos-sdk/codec"
    "github.com/cosmos/cosmos-sdk/types"
    
    "your-project/client"
    "your-project/services"
)

func main() {
    // 创建编解码器
    cdc := codec.NewProtoCodec(nil)
    
    // 创建IBC客户端 (Cosmos Hub)
    cosmosClient, err := client.NewIBCClient(
        "cosmoshub-4",
        "https://rpc.cosmos.network:443",
        cdc,
    )
    if err != nil {
        log.Fatal("创建Cosmos IBC客户端失败:", err)
    }
    
    // 创建IBC客户端 (Osmosis)
    osmosisClient, err := client.NewIBCClient(
        "osmosis-1",
        "https://rpc.osmosis.zone:443",
        cdc,
    )
    if err != nil {
        log.Fatal("创建Osmosis IBC客户端失败:", err)
    }
    
    // 创建客户端上下文
    clientCtx := client.Context{}.
        WithCodec(cdc).
        WithChainID("cosmoshub-4")
    
    // 创建服务
    transferService := services.NewIBCTransferService(cosmosClient, clientCtx)
    relayerService := services.NewRelayerService()
    
    // 1. 查询IBC客户端
    fmt.Printf("=== IBC客户端信息 ===\n")
    
    clients, err := cosmosClient.GetAllClients()
    if err != nil {
        log.Printf("获取IBC客户端失败: %v", err)
    } else {
        fmt.Printf("Cosmos Hub IBC客户端数量: %d\n", len(clients))
        for i, client := range clients {
            if i >= 5 { // 只显示前5个
                break
            }
            fmt.Printf("  客户端 %d:\n", i+1)
            fmt.Printf("    ID: %s\n", client.ClientID)
            fmt.Printf("    类型: %s\n", client.ClientType)
        }
    }
    
    // 2. 查询IBC连接
    fmt.Printf("\n=== IBC连接信息 ===\n")
    
    connections, err := cosmosClient.GetAllConnections()
    if err != nil {
        log.Printf("获取IBC连接失败: %v", err)
    } else {
        fmt.Printf("Cosmos Hub IBC连接数量: %d\n", len(connections))
        for i, conn := range connections {
            if i >= 3 { // 只显示前3个
                break
            }
            fmt.Printf("  连接 %d:\n", i+1)
            fmt.Printf("    ID: %s\n", conn.ConnectionID)
            fmt.Printf("    客户端ID: %s\n", conn.ClientID)
            fmt.Printf("    状态: %s\n", conn.State)
            if conn.Counterparty != nil {
                fmt.Printf("    对端连接ID: %s\n", conn.Counterparty.ConnectionID)
            }
        }
    }
    
    // 3. 查询IBC通道
    fmt.Printf("\n=== IBC通道信息 ===\n")
    
    channels, err := cosmosClient.GetAllChannels()
    if err != nil {
        log.Printf("获取IBC通道失败: %v", err)
    } else {
        fmt.Printf("Cosmos Hub IBC通道数量: %d\n", len(channels))
        for i, ch := range channels {
            if i >= 5 { // 只显示前5个
                break
            }
            fmt.Printf("  通道 %d:\n", i+1)
            fmt.Printf("    端口ID: %s\n", ch.PortID)
            fmt.Printf("    通道ID: %s\n", ch.ChannelID)
            fmt.Printf("    状态: %s\n", ch.State)
            fmt.Printf("    排序: %s\n", ch.Ordering)
            if ch.Counterparty != nil {
                fmt.Printf("    对端通道: %s/%s\n", ch.Counterparty.PortID, ch.Counterparty.ChannelID)
            }
        }
    }
    
    // 4. IBC转账示例
    fmt.Printf("\n=== IBC转账示例 ===\n")
    
    // 验证IBC路径
    sourcePort := "transfer"
    sourceChannel := "channel-141" // Cosmos Hub到Osmosis的通道
    
    err = transferService.ValidateIBCPath(sourcePort, sourceChannel, "osmosis-1")
    if err != nil {
        log.Printf("IBC路径验证失败: %v", err)
    } else {
        fmt.Printf("IBC路径验证成功: %s/%s\n", sourcePort, sourceChannel)
        
        // 计算超时参数
        timeoutHeight := transferService.CalculateTimeoutHeight(12345678, 1000)
        timeoutTimestamp := transferService.CalculateTimeoutTimestamp(10 * time.Minute)
        
        fmt.Printf("超时高度: %d\n", timeoutHeight)
        fmt.Printf("超时时间戳: %d\n", timeoutTimestamp)
        
        // 创建转账 (示例，不实际执行)
        token := types.NewCoin("uatom", types.NewInt(1000000)) // 1 ATOM
        
        fmt.Printf("准备转账:\n")
        fmt.Printf("  代币: %s\n", token.String())
        fmt.Printf("  源端口: %s\n", sourcePort)
        fmt.Printf("  源通道: %s\n", sourceChannel)
        
        // 实际转账需要有效的发送者地址和私钥
        // txResp, err := transferService.Transfer(
        //     "cosmos1...", // 发送者
        //     "osmo1...",   // 接收者
        //     token,
        //     sourcePort,
        //     sourceChannel,
        //     timeoutHeight,
        //     timeoutTimestamp,
        //     "IBC transfer memo",
        // )
    }
    
    // 5. 中继器配置示例
    fmt.Printf("\n=== 中继器配置示例 ===\n")
    
    // 添加链配置
    err = relayerService.AddChain(
        "cosmoshub-4",
        "https://rpc.cosmos.network:443",
        "https://grpc.cosmos.network:443",
        "cosmos-key",
    )
    if err != nil {
        log.Printf("添加Cosmos Hub链配置失败: %v", err)
    } else {
        fmt.Printf("Cosmos Hub链配置添加成功\n")
    }
    
    err = relayerService.AddChain(
        "osmosis-1",
        "https://rpc.osmosis.zone:443",
        "https://grpc.osmosis.zone:443",
        "osmosis-key",
    )
    if err != nil {
        log.Printf("添加Osmosis链配置失败: %v", err)
    } else {
        fmt.Printf("Osmosis链配置添加成功\n")
    }
    
    // 6. 数据包跟踪示例
    fmt.Printf("\n=== 数据包跟踪示例 ===\n")
    
    // 跟踪特定数据包
    packetStatus, err := transferService.TrackPacket(
        "transfer",
        "channel-141",
        12345, // 序列号
    )
    if err != nil {
        log.Printf("跟踪数据包失败: %v", err)
    } else {
        fmt.Printf("数据包状态:\n")
        fmt.Printf("  序列号: %d\n", packetStatus.Sequence)
        fmt.Printf("  状态: %s\n", packetStatus.Status)
        if packetStatus.Error != "" {
            fmt.Printf("  错误: %s\n", packetStatus.Error)
        }
    }
    
    // 7. IBC路径发现
    fmt.Printf("\n=== IBC路径发现 ===\n")
    
    // 查找Cosmos Hub到Osmosis的所有可用路径
    transferChannels := []string{}
    for _, ch := range channels {
        if ch.PortID == "transfer" && ch.State == ChannelStateOpen {
            transferChannels = append(transferChannels, ch.ChannelID)
        }
    }
    
    fmt.Printf("可用的转账通道: %v\n", transferChannels)
    
    // 8. 监控IBC活动
    fmt.Printf("\n=== IBC活动监控 ===\n")
    
    fmt.Printf("开始监控IBC活动...\n")
    
    // 模拟监控循环
    for i := 0; i < 3; i++ {
        fmt.Printf("监控周期 %d:\n", i+1)
        
        // 检查新的数据包
        fmt.Printf("  检查新数据包...\n")
        
        // 检查超时的数据包
        fmt.Printf("  检查超时数据包...\n")
        
        // 检查需要中继的确认
        fmt.Printf("  检查待中继确认...\n")
        
        time.Sleep(5 * time.Second)
    }
    
    fmt.Printf("IBC监控完成\n")
}
```

这个Cosmos IBC使用指南提供了完整的跨链通信协议集成方案，涵盖了客户端管理、连接建立、通道操作、跨链转账、中继器配置等核心功能，是构建跨链DeFi应用的重要参考文档。
