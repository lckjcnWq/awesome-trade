# libp2p P2P网络框架详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [环境准备](#环境准备)
3. [节点创建](#节点创建)
4. [网络发现](#网络发现)
5. [协议处理](#协议处理)
6. [流管理](#流管理)
7. [DHT和路由](#DHT和路由)
8. [安全和加密](#安全和加密)
9. [实际应用](#实际应用)

## 基础概念

### 1.1 libp2p简介

libp2p是模块化网络栈，提供NAT穿透、多传输协议支持、模块化网络栈，用于构建P2P应用。

```bash
# 安装libp2p
go get github.com/libp2p/go-libp2p
go get github.com/libp2p/go-libp2p/core/host
go get github.com/libp2p/go-libp2p/core/network
go get github.com/libp2p/go-libp2p/core/peer
```

### 1.2 核心概念

```go
// 主要包导入
import (
    "github.com/libp2p/go-libp2p"
    "github.com/libp2p/go-libp2p/core/host"
    "github.com/libp2p/go-libp2p/core/network"
    "github.com/libp2p/go-libp2p/core/peer"
    "github.com/libp2p/go-libp2p/core/protocol"
    "github.com/libp2p/go-libp2p/p2p/discovery/mdns"
    "github.com/libp2p/go-libp2p/p2p/net/swarm"
)
```

## 环境准备

### 2.1 基础配置

```go
// config/p2p.go
package config

import (
    "crypto/rand"
    "fmt"
    
    "github.com/libp2p/go-libp2p"
    "github.com/libp2p/go-libp2p/core/crypto"
    "github.com/libp2p/go-libp2p/core/host"
    "github.com/libp2p/go-libp2p/core/peer"
    "github.com/libp2p/go-libp2p/p2p/security/noise"
    "github.com/libp2p/go-libp2p/p2p/transport/tcp"
    "github.com/libp2p/go-libp2p/p2p/muxer/yamux"
)

type P2PConfig struct {
    Port        int
    PrivateKey  crypto.PrivKey
    ProtocolID  protocol.ID
    RendezvousString string
}

func DefaultConfig() *P2PConfig {
    return &P2PConfig{
        Port:             0, // 随机端口
        ProtocolID:       "/chat/1.0.0",
        RendezvousString: "meet-me-here",
    }
}

// 生成密钥对
func GenerateKeyPair() (crypto.PrivKey, crypto.PubKey, error) {
    priv, pub, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
    if err != nil {
        return nil, nil, err
    }
    return priv, pub, nil
}

// 创建libp2p主机
func CreateHost(config *P2PConfig) (host.Host, error) {
    var opts []libp2p.Option

    // 设置监听地址
    if config.Port != 0 {
        opts = append(opts, libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", config.Port)))
    } else {
        opts = append(opts, libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"))
    }

    // 设置身份
    if config.PrivateKey != nil {
        opts = append(opts, libp2p.Identity(config.PrivateKey))
    }

    // 设置传输协议
    opts = append(opts, libp2p.Transport(tcp.NewTCPTransport))

    // 设置多路复用器
    opts = append(opts, libp2p.Muxer("/yamux/1.0.0", yamux.DefaultTransport))

    // 设置安全传输
    opts = append(opts, libp2p.Security(noise.ID, noise.New))

    // 创建主机
    h, err := libp2p.New(opts...)
    if err != nil {
        return nil, err
    }

    return h, nil
}
```

## 节点创建

### 3.1 基本节点

```go
// node/p2p_node.go
package node

import (
    "context"
    "fmt"
    "log"
    "sync"

    "github.com/libp2p/go-libp2p/core/host"
    "github.com/libp2p/go-libp2p/core/network"
    "github.com/libp2p/go-libp2p/core/peer"
    "github.com/libp2p/go-libp2p/core/protocol"
)

type P2PNode struct {
    Host       host.Host
    ctx        context.Context
    cancel     context.CancelFunc
    protocols  map[protocol.ID]network.StreamHandler
    peers      map[peer.ID]*PeerInfo
    peersMutex sync.RWMutex
}

type PeerInfo struct {
    ID        peer.ID
    Addresses []string
    Connected bool
    Protocols []protocol.ID
}

func NewP2PNode(h host.Host) *P2PNode {
    ctx, cancel := context.WithCancel(context.Background())
    
    node := &P2PNode{
        Host:      h,
        ctx:       ctx,
        cancel:    cancel,
        protocols: make(map[protocol.ID]network.StreamHandler),
        peers:     make(map[peer.ID]*PeerInfo),
    }

    // 设置连接事件处理器
    h.Network().Notify(&network.NotifyBundle{
        ConnectedF: func(n network.Network, c network.Conn) {
            node.onPeerConnected(c.RemotePeer())
        },
        DisconnectedF: func(n network.Network, c network.Conn) {
            node.onPeerDisconnected(c.RemotePeer())
        },
    })

    return node
}

// 启动节点
func (n *P2PNode) Start() error {
    log.Printf("P2P节点启动，ID: %s", n.Host.ID())
    log.Printf("监听地址: %v", n.Host.Addrs())

    return nil
}

// 停止节点
func (n *P2PNode) Stop() error {
    n.cancel()
    return n.Host.Close()
}

// 注册协议处理器
func (n *P2PNode) RegisterProtocol(protocolID protocol.ID, handler network.StreamHandler) {
    n.protocols[protocolID] = handler
    n.Host.SetStreamHandler(protocolID, handler)
    log.Printf("注册协议: %s", protocolID)
}

// 连接到对等节点
func (n *P2PNode) ConnectToPeer(peerAddr string) error {
    addr, err := peer.AddrInfoFromString(peerAddr)
    if err != nil {
        return fmt.Errorf("解析对等节点地址失败: %v", err)
    }

    err = n.Host.Connect(n.ctx, *addr)
    if err != nil {
        return fmt.Errorf("连接对等节点失败: %v", err)
    }

    log.Printf("成功连接到对等节点: %s", addr.ID)
    return nil
}

// 对等节点连接事件
func (n *P2PNode) onPeerConnected(peerID peer.ID) {
    n.peersMutex.Lock()
    defer n.peersMutex.Unlock()

    peerInfo := &PeerInfo{
        ID:        peerID,
        Connected: true,
    }

    // 获取对等节点地址
    peerInfo.Addresses = make([]string, 0)
    for _, addr := range n.Host.Peerstore().Addrs(peerID) {
        peerInfo.Addresses = append(peerInfo.Addresses, addr.String())
    }

    // 获取支持的协议
    protocols, err := n.Host.Peerstore().GetProtocols(peerID)
    if err == nil {
        peerInfo.Protocols = protocols
    }

    n.peers[peerID] = peerInfo
    log.Printf("对等节点已连接: %s", peerID)
}

// 对等节点断开事件
func (n *P2PNode) onPeerDisconnected(peerID peer.ID) {
    n.peersMutex.Lock()
    defer n.peersMutex.Unlock()

    if peerInfo, exists := n.peers[peerID]; exists {
        peerInfo.Connected = false
    }

    log.Printf("对等节点已断开: %s", peerID)
}

// 获取连接的对等节点
func (n *P2PNode) GetConnectedPeers() []*PeerInfo {
    n.peersMutex.RLock()
    defer n.peersMutex.RUnlock()

    var connectedPeers []*PeerInfo
    for _, peerInfo := range n.peers {
        if peerInfo.Connected {
            connectedPeers = append(connectedPeers, peerInfo)
        }
    }

    return connectedPeers
}
```

## 网络发现

### 4.1 mDNS发现

```go
// discovery/mdns.go
package discovery

import (
    "context"
    "log"
    "time"

    "github.com/libp2p/go-libp2p/core/host"
    "github.com/libp2p/go-libp2p/core/peer"
    "github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

type MDNSDiscovery struct {
    host    host.Host
    service mdns.Service
    tag     string
}

func NewMDNSDiscovery(h host.Host, serviceTag string) *MDNSDiscovery {
    return &MDNSDiscovery{
        host: h,
        tag:  serviceTag,
    }
}

// 启动mDNS发现
func (d *MDNSDiscovery) Start(ctx context.Context) error {
    service := mdns.NewMdnsService(d.host, d.tag, d)
    err := service.Start()
    if err != nil {
        return err
    }

    d.service = service
    log.Printf("mDNS发现服务已启动，标签: %s", d.tag)
    return nil
}

// 停止mDNS发现
func (d *MDNSDiscovery) Stop() error {
    if d.service != nil {
        return d.service.Close()
    }
    return nil
}

// 实现mdns.Notifee接口
func (d *MDNSDiscovery) HandlePeerFound(pi peer.AddrInfo) {
    log.Printf("通过mDNS发现对等节点: %s", pi.ID)
    
    // 连接到发现的对等节点
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    err := d.host.Connect(ctx, pi)
    if err != nil {
        log.Printf("连接到发现的对等节点失败: %v", err)
        return
    }

    log.Printf("成功连接到发现的对等节点: %s", pi.ID)
}
```

### 4.2 DHT发现

```go
// discovery/dht.go
package discovery

import (
    "context"
    "log"
    "time"

    "github.com/libp2p/go-libp2p/core/host"
    "github.com/libp2p/go-libp2p/core/peer"
    "github.com/libp2p/go-libp2p/p2p/discovery/routing"
    "github.com/libp2p/go-libp2p/p2p/discovery/util"
    dht "github.com/libp2p/go-libp2p-kad-dht"
)

type DHTDiscovery struct {
    host           host.Host
    dht            *dht.IpfsDHT
    routingDiscovery *routing.RoutingDiscovery
    rendezvous     string
}

func NewDHTDiscovery(h host.Host, rendezvous string, bootstrapPeers []peer.AddrInfo) (*DHTDiscovery, error) {
    // 创建DHT
    kademliaDHT, err := dht.New(context.Background(), h)
    if err != nil {
        return nil, err
    }

    // 创建路由发现
    routingDiscovery := routing.NewRoutingDiscovery(kademliaDHT)

    discovery := &DHTDiscovery{
        host:             h,
        dht:              kademliaDHT,
        routingDiscovery: routingDiscovery,
        rendezvous:       rendezvous,
    }

    // 连接到引导节点
    for _, peerAddr := range bootstrapPeers {
        go func(pi peer.AddrInfo) {
            ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
            defer cancel()
            
            err := h.Connect(ctx, pi)
            if err != nil {
                log.Printf("连接引导节点失败: %v", err)
            } else {
                log.Printf("连接引导节点成功: %s", pi.ID)
            }
        }(peerAddr)
    }

    return discovery, nil
}

// 启动DHT
func (d *DHTDiscovery) Start(ctx context.Context) error {
    // 引导DHT
    err := d.dht.Bootstrap(ctx)
    if err != nil {
        return err
    }

    log.Println("DHT引导完成")
    return nil
}

// 广告服务
func (d *DHTDiscovery) Advertise(ctx context.Context) error {
    util.Advertise(ctx, d.routingDiscovery, d.rendezvous)
    log.Printf("开始广告服务: %s", d.rendezvous)
    return nil
}

// 发现对等节点
func (d *DHTDiscovery) FindPeers(ctx context.Context) error {
    log.Printf("开始寻找对等节点: %s", d.rendezvous)

    peerChan, err := d.routingDiscovery.FindPeers(ctx, d.rendezvous)
    if err != nil {
        return err
    }

    go func() {
        for peer := range peerChan {
            if peer.ID == d.host.ID() {
                continue // 跳过自己
            }

            log.Printf("发现对等节点: %s", peer.ID)

            // 连接到对等节点
            ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
            err := d.host.Connect(ctx, peer)
            cancel()

            if err != nil {
                log.Printf("连接对等节点失败: %v", err)
            } else {
                log.Printf("成功连接对等节点: %s", peer.ID)
            }
        }
    }()

    return nil
}

// 停止DHT
func (d *DHTDiscovery) Stop() error {
    return d.dht.Close()
}
```

## 协议处理

### 5.1 自定义协议

```go
// protocol/chat.go
package protocol

import (
    "bufio"
    "context"
    "encoding/json"
    "fmt"
    "log"
    "time"

    "github.com/libp2p/go-libp2p/core/host"
    "github.com/libp2p/go-libp2p/core/network"
    "github.com/libp2p/go-libp2p/core/peer"
    "github.com/libp2p/go-libp2p/core/protocol"
)

const ChatProtocol = "/chat/1.0.0"

type ChatMessage struct {
    From      string    `json:"from"`
    To        string    `json:"to"`
    Content   string    `json:"content"`
    Timestamp time.Time `json:"timestamp"`
    Type      string    `json:"type"`
}

type ChatProtocolHandler struct {
    host         host.Host
    messages     chan ChatMessage
    messageHandlers []func(ChatMessage)
}

func NewChatProtocolHandler(h host.Host) *ChatProtocolHandler {
    handler := &ChatProtocolHandler{
        host:     h,
        messages: make(chan ChatMessage, 100),
    }

    // 注册协议处理器
    h.SetStreamHandler(protocol.ID(ChatProtocol), handler.handleStream)

    return handler
}

// 处理传入的流
func (c *ChatProtocolHandler) handleStream(s network.Stream) {
    log.Printf("收到来自 %s 的新流", s.Conn().RemotePeer())

    // 读取消息
    rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
    
    go c.readData(rw, s.Conn().RemotePeer())
    go c.writeData(rw)
}

// 读取数据
func (c *ChatProtocolHandler) readData(rw *bufio.ReadWriter, remotePeer peer.ID) {
    for {
        str, err := rw.ReadString('\n')
        if err != nil {
            log.Printf("读取数据错误: %v", err)
            return
        }

        if str == "" {
            return
        }

        if str != "\n" {
            var msg ChatMessage
            err := json.Unmarshal([]byte(str), &msg)
            if err != nil {
                log.Printf("解析消息失败: %v", err)
                continue
            }

            log.Printf("收到消息: %s", msg.Content)
            
            // 通知消息处理器
            for _, handler := range c.messageHandlers {
                go handler(msg)
            }

            // 发送到消息通道
            select {
            case c.messages <- msg:
            default:
                log.Println("消息通道已满，丢弃消息")
            }
        }
    }
}

// 写入数据
func (c *ChatProtocolHandler) writeData(rw *bufio.ReadWriter) {
    // 这里可以实现定期发送心跳或其他数据
}

// 发送消息
func (c *ChatProtocolHandler) SendMessage(ctx context.Context, peerID peer.ID, content string) error {
    // 打开流
    s, err := c.host.NewStream(ctx, peerID, protocol.ID(ChatProtocol))
    if err != nil {
        return fmt.Errorf("打开流失败: %v", err)
    }
    defer s.Close()

    // 创建消息
    msg := ChatMessage{
        From:      c.host.ID().String(),
        To:        peerID.String(),
        Content:   content,
        Timestamp: time.Now(),
        Type:      "text",
    }

    // 序列化消息
    msgBytes, err := json.Marshal(msg)
    if err != nil {
        return fmt.Errorf("序列化消息失败: %v", err)
    }

    // 发送消息
    rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
    _, err = rw.WriteString(string(msgBytes) + "\n")
    if err != nil {
        return fmt.Errorf("发送消息失败: %v", err)
    }

    err = rw.Flush()
    if err != nil {
        return fmt.Errorf("刷新缓冲区失败: %v", err)
    }

    log.Printf("消息已发送到 %s: %s", peerID, content)
    return nil
}

// 广播消息
func (c *ChatProtocolHandler) BroadcastMessage(ctx context.Context, content string) {
    peers := c.host.Network().Peers()
    
    for _, peerID := range peers {
        go func(pid peer.ID) {
            err := c.SendMessage(ctx, pid, content)
            if err != nil {
                log.Printf("向 %s 发送消息失败: %v", pid, err)
            }
        }(peerID)
    }
}

// 添加消息处理器
func (c *ChatProtocolHandler) AddMessageHandler(handler func(ChatMessage)) {
    c.messageHandlers = append(c.messageHandlers, handler)
}

// 获取消息通道
func (c *ChatProtocolHandler) GetMessageChannel() <-chan ChatMessage {
    return c.messages
}
```

## 流管理

### 6.1 流控制

```go
// stream/manager.go
package stream

import (
    "context"
    "fmt"
    "sync"
    "time"

    "github.com/libp2p/go-libp2p/core/host"
    "github.com/libp2p/go-libp2p/core/network"
    "github.com/libp2p/go-libp2p/core/peer"
    "github.com/libp2p/go-libp2p/core/protocol"
)

type StreamManager struct {
    host    host.Host
    streams map[string]network.Stream
    mutex   sync.RWMutex
}

func NewStreamManager(h host.Host) *StreamManager {
    return &StreamManager{
        host:    h,
        streams: make(map[string]network.Stream),
    }
}

// 创建流
func (sm *StreamManager) CreateStream(ctx context.Context, peerID peer.ID, protocolID protocol.ID) (network.Stream, error) {
    streamKey := fmt.Sprintf("%s-%s", peerID.String(), protocolID)
    
    sm.mutex.Lock()
    defer sm.mutex.Unlock()

    // 检查是否已存在流
    if existingStream, exists := sm.streams[streamKey]; exists {
        // 检查流是否仍然有效
        select {
        case <-existingStream.Context().Done():
            // 流已关闭，删除并创建新的
            delete(sm.streams, streamKey)
        default:
            // 流仍然有效，返回现有流
            return existingStream, nil
        }
    }

    // 创建新流
    stream, err := sm.host.NewStream(ctx, peerID, protocolID)
    if err != nil {
        return nil, err
    }

    // 存储流
    sm.streams[streamKey] = stream

    // 监听流关闭事件
    go func() {
        <-stream.Context().Done()
        sm.mutex.Lock()
        delete(sm.streams, streamKey)
        sm.mutex.Unlock()
    }()

    return stream, nil
}

// 获取流
func (sm *StreamManager) GetStream(peerID peer.ID, protocolID protocol.ID) (network.Stream, bool) {
    streamKey := fmt.Sprintf("%s-%s", peerID.String(), protocolID)
    
    sm.mutex.RLock()
    defer sm.mutex.RUnlock()

    stream, exists := sm.streams[streamKey]
    if !exists {
        return nil, false
    }

    // 检查流是否仍然有效
    select {
    case <-stream.Context().Done():
        return nil, false
    default:
        return stream, true
    }
}

// 关闭流
func (sm *StreamManager) CloseStream(peerID peer.ID, protocolID protocol.ID) error {
    streamKey := fmt.Sprintf("%s-%s", peerID.String(), protocolID)
    
    sm.mutex.Lock()
    defer sm.mutex.Unlock()

    if stream, exists := sm.streams[streamKey]; exists {
        delete(sm.streams, streamKey)
        return stream.Close()
    }

    return nil
}

// 关闭所有流
func (sm *StreamManager) CloseAllStreams() {
    sm.mutex.Lock()
    defer sm.mutex.Unlock()

    for key, stream := range sm.streams {
        stream.Close()
        delete(sm.streams, key)
    }
}

// 获取活跃流数量
func (sm *StreamManager) GetActiveStreamCount() int {
    sm.mutex.RLock()
    defer sm.mutex.RUnlock()

    count := 0
    for _, stream := range sm.streams {
        select {
        case <-stream.Context().Done():
            // 流已关闭
        default:
            count++
        }
    }

    return count
}
```

## DHT和路由

### 7.1 内容路由

```go
// routing/content.go
package routing

import (
    "context"
    "fmt"
    "log"

    "github.com/libp2p/go-libp2p/core/host"
    "github.com/libp2p/go-libp2p/core/peer"
    dht "github.com/libp2p/go-libp2p-kad-dht"
    "github.com/multiformats/go-multihash"
)

type ContentRouter struct {
    host host.Host
    dht  *dht.IpfsDHT
}

func NewContentRouter(h host.Host, d *dht.IpfsDHT) *ContentRouter {
    return &ContentRouter{
        host: h,
        dht:  d,
    }
}

// 提供内容
func (cr *ContentRouter) ProvideContent(ctx context.Context, content []byte) error {
    // 计算内容哈希
    hash, err := multihash.Sum(content, multihash.SHA2_256, -1)
    if err != nil {
        return fmt.Errorf("计算哈希失败: %v", err)
    }

    // 在DHT中提供内容
    err = cr.dht.Provide(ctx, hash, true)
    if err != nil {
        return fmt.Errorf("提供内容失败: %v", err)
    }

    log.Printf("内容已提供，哈希: %s", hash.String())
    return nil
}

// 查找内容提供者
func (cr *ContentRouter) FindContentProviders(ctx context.Context, contentHash string) ([]peer.AddrInfo, error) {
    // 解析哈希
    hash, err := multihash.FromB58String(contentHash)
    if err != nil {
        return nil, fmt.Errorf("解析哈希失败: %v", err)
    }

    // 查找提供者
    providers := cr.dht.FindProvidersAsync(ctx, hash, 10)
    
    var result []peer.AddrInfo
    for provider := range providers {
        result = append(result, provider)
        log.Printf("找到内容提供者: %s", provider.ID)
    }

    return result, nil
}

// 存储值
func (cr *ContentRouter) PutValue(ctx context.Context, key string, value []byte) error {
    err := cr.dht.PutValue(ctx, key, value)
    if err != nil {
        return fmt.Errorf("存储值失败: %v", err)
    }

    log.Printf("值已存储，键: %s", key)
    return nil
}

// 获取值
func (cr *ContentRouter) GetValue(ctx context.Context, key string) ([]byte, error) {
    value, err := cr.dht.GetValue(ctx, key)
    if err != nil {
        return nil, fmt.Errorf("获取值失败: %v", err)
    }

    log.Printf("值已获取，键: %s", key)
    return value, nil
}

// 查找对等节点
func (cr *ContentRouter) FindPeer(ctx context.Context, peerID peer.ID) (peer.AddrInfo, error) {
    peerInfo, err := cr.dht.FindPeer(ctx, peerID)
    if err != nil {
        return peer.AddrInfo{}, fmt.Errorf("查找对等节点失败: %v", err)
    }

    log.Printf("找到对等节点: %s", peerInfo.ID)
    return peerInfo, nil
}
```

## 安全和加密

### 8.1 身份验证

```go
// security/auth.go
package security

import (
    "crypto/rand"
    "fmt"

    "github.com/libp2p/go-libp2p/core/crypto"
    "github.com/libp2p/go-libp2p/core/peer"
)

type IdentityManager struct {
    privateKey crypto.PrivKey
    publicKey  crypto.PubKey
    peerID     peer.ID
}

func NewIdentityManager() (*IdentityManager, error) {
    // 生成密钥对
    priv, pub, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
    if err != nil {
        return nil, fmt.Errorf("生成密钥对失败: %v", err)
    }

    // 从公钥生成对等节点ID
    peerID, err := peer.IDFromPublicKey(pub)
    if err != nil {
        return nil, fmt.Errorf("生成对等节点ID失败: %v", err)
    }

    return &IdentityManager{
        privateKey: priv,
        publicKey:  pub,
        peerID:     peerID,
    }, nil
}

// 获取私钥
func (im *IdentityManager) GetPrivateKey() crypto.PrivKey {
    return im.privateKey
}

// 获取公钥
func (im *IdentityManager) GetPublicKey() crypto.PubKey {
    return im.publicKey
}

// 获取对等节点ID
func (im *IdentityManager) GetPeerID() peer.ID {
    return im.peerID
}

// 签名数据
func (im *IdentityManager) SignData(data []byte) ([]byte, error) {
    signature, err := im.privateKey.Sign(data)
    if err != nil {
        return nil, fmt.Errorf("签名失败: %v", err)
    }
    return signature, nil
}

// 验证签名
func (im *IdentityManager) VerifySignature(data, signature []byte, publicKey crypto.PubKey) (bool, error) {
    valid, err := publicKey.Verify(data, signature)
    if err != nil {
        return false, fmt.Errorf("验证签名失败: %v", err)
    }
    return valid, nil
}

// 从私钥字节恢复身份
func RestoreIdentityFromBytes(keyBytes []byte) (*IdentityManager, error) {
    priv, err := crypto.UnmarshalPrivateKey(keyBytes)
    if err != nil {
        return nil, fmt.Errorf("解析私钥失败: %v", err)
    }

    pub := priv.GetPublic()
    peerID, err := peer.IDFromPublicKey(pub)
    if err != nil {
        return nil, fmt.Errorf("生成对等节点ID失败: %v", err)
    }

    return &IdentityManager{
        privateKey: priv,
        publicKey:  pub,
        peerID:     peerID,
    }, nil
}

// 导出私钥字节
func (im *IdentityManager) ExportPrivateKey() ([]byte, error) {
    return crypto.MarshalPrivateKey(im.privateKey)
}
```

## 实际应用

### 9.1 完整P2P聊天应用

```go
// main.go
package main

import (
    "bufio"
    "context"
    "fmt"
    "log"
    "os"
    "strings"
    "time"

    "your-project/config"
    "your-project/node"
    "your-project/discovery"
    "your-project/protocol"
)

func main() {
    // 创建配置
    cfg := config.DefaultConfig()

    // 生成密钥对
    priv, _, err := config.GenerateKeyPair()
    if err != nil {
        log.Fatal("生成密钥对失败:", err)
    }
    cfg.PrivateKey = priv

    // 创建主机
    host, err := config.CreateHost(cfg)
    if err != nil {
        log.Fatal("创建主机失败:", err)
    }
    defer host.Close()

    // 创建P2P节点
    p2pNode := node.NewP2PNode(host)
    err = p2pNode.Start()
    if err != nil {
        log.Fatal("启动P2P节点失败:", err)
    }
    defer p2pNode.Stop()

    // 创建聊天协议处理器
    chatHandler := protocol.NewChatProtocolHandler(host)
    
    // 添加消息处理器
    chatHandler.AddMessageHandler(func(msg protocol.ChatMessage) {
        fmt.Printf("[%s] %s: %s\n", msg.Timestamp.Format("15:04:05"), msg.From, msg.Content)
    })

    // 启动mDNS发现
    mdnsDiscovery := discovery.NewMDNSDiscovery(host, "chat-app")
    ctx := context.Background()
    err = mdnsDiscovery.Start(ctx)
    if err != nil {
        log.Printf("启动mDNS发现失败: %v", err)
    }
    defer mdnsDiscovery.Stop()

    fmt.Printf("聊天应用已启动！\n")
    fmt.Printf("节点ID: %s\n", host.ID())
    fmt.Printf("监听地址: %v\n", host.Addrs())
    fmt.Printf("输入消息开始聊天，输入 '/quit' 退出\n")

    // 处理用户输入
    scanner := bufio.NewScanner(os.Stdin)
    for {
        fmt.Print("> ")
        if !scanner.Scan() {
            break
        }

        input := strings.TrimSpace(scanner.Text())
        if input == "" {
            continue
        }

        if input == "/quit" {
            break
        }

        if strings.HasPrefix(input, "/connect ") {
            // 连接到指定节点
            peerAddr := strings.TrimPrefix(input, "/connect ")
            err := p2pNode.ConnectToPeer(peerAddr)
            if err != nil {
                fmt.Printf("连接失败: %v\n", err)
            }
            continue
        }

        if input == "/peers" {
            // 显示连接的对等节点
            peers := p2pNode.GetConnectedPeers()
            fmt.Printf("连接的对等节点 (%d):\n", len(peers))
            for _, peer := range peers {
                fmt.Printf("  - %s\n", peer.ID)
            }
            continue
        }

        // 广播消息
        chatHandler.BroadcastMessage(ctx, input)
    }

    fmt.Println("聊天应用已退出")
}
```
