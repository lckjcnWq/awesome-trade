# IPFS Go 详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [环境准备](#环境准备)
3. [节点连接](#节点连接)
4. [文件操作](#文件操作)
5. [目录管理](#目录管理)
6. [网络发现](#网络发现)
7. [内容寻址](#内容寻址)
8. [高级功能](#高级功能)
9. [实际应用](#实际应用)

## 基础概念

### 1.1 IPFS简介

IPFS (InterPlanetary File System) 是分布式文件系统，使用内容寻址、版本控制和点对点网络来创建持久化的分布式存储网络。

```bash
# 安装IPFS Go客户端
go get github.com/ipfs/go-ipfs-api
go get github.com/ipfs/go-ipfs-files
go get github.com/ipfs/go-cid
go get github.com/multiformats/go-multihash
```

### 1.2 核心概念

```go
// 主要包导入
import (
    "context"
    "io"
    "os"
    
    "github.com/ipfs/go-ipfs-api"
    "github.com/ipfs/go-ipfs-files"
    "github.com/ipfs/go-cid"
    "github.com/multiformats/go-multihash"
)

// 核心概念：
// - CID: 内容标识符，基于内容的唯一标识
// - Hash: 文件内容的加密哈希
// - Node: IPFS网络中的节点
// - Pin: 固定内容，防止垃圾回收
// - Gateway: HTTP网关，提供Web访问
```

## 环境准备

### 2.1 客户端配置

```go
// config/ipfs.go
package config

import (
    "time"
)

type IPFSConfig struct {
    APIAddr     string
    GatewayAddr string
    Timeout     time.Duration
    PinTimeout  time.Duration
    ChunkSize   int64
}

func DefaultIPFSConfig() *IPFSConfig {
    return &IPFSConfig{
        APIAddr:     "localhost:5001",
        GatewayAddr: "localhost:8080",
        Timeout:     30 * time.Second,
        PinTimeout:  60 * time.Second,
        ChunkSize:   1024 * 1024, // 1MB
    }
}

// IPFS网关配置
type GatewayConfig struct {
    PublicGateways []string
    LocalGateway   string
    Timeout        time.Duration
}

func DefaultGatewayConfig() *GatewayConfig {
    return &GatewayConfig{
        PublicGateways: []string{
            "https://ipfs.io",
            "https://gateway.ipfs.io",
            "https://cloudflare-ipfs.com",
            "https://dweb.link",
        },
        LocalGateway: "http://localhost:8080",
        Timeout:      10 * time.Second,
    }
}
```

## 节点连接

### 3.1 IPFS客户端

```go
// client/ipfs_client.go
package client

import (
    "context"
    "fmt"
    "time"

    shell "github.com/ipfs/go-ipfs-api"
    
    "your-project/config"
)

type IPFSClient struct {
    shell   *shell.Shell
    config  *config.IPFSConfig
    ctx     context.Context
    cancel  context.CancelFunc
}

func NewIPFSClient(cfg *config.IPFSConfig) (*IPFSClient, error) {
    // 创建IPFS Shell连接
    sh := shell.NewShell(cfg.APIAddr)
    
    // 设置超时
    sh.SetTimeout(cfg.Timeout)

    // 测试连接
    ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
    defer cancel()

    _, err := sh.ID()
    if err != nil {
        return nil, fmt.Errorf("连接IPFS节点失败: %v", err)
    }

    clientCtx, clientCancel := context.WithCancel(context.Background())

    return &IPFSClient{
        shell:  sh,
        config: cfg,
        ctx:    clientCtx,
        cancel: clientCancel,
    }, nil
}

// 获取节点信息
func (c *IPFSClient) GetNodeInfo() (*NodeInfo, error) {
    id, err := c.shell.ID()
    if err != nil {
        return nil, fmt.Errorf("获取节点ID失败: %v", err)
    }

    version, _, err := c.shell.Version()
    if err != nil {
        return nil, fmt.Errorf("获取版本信息失败: %v", err)
    }

    return &NodeInfo{
        ID:        id,
        Version:   version,
        Addresses: []string{}, // 可以通过其他API获取
    }, nil
}

// 检查节点状态
func (c *IPFSClient) IsOnline() bool {
    ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
    defer cancel()

    _, err := c.shell.WithContext(ctx).ID()
    return err == nil
}

// 获取对等节点
func (c *IPFSClient) GetPeers() ([]PeerInfo, error) {
    peers, err := c.shell.SwarmPeers(c.ctx)
    if err != nil {
        return nil, fmt.Errorf("获取对等节点失败: %v", err)
    }

    var peerInfos []PeerInfo
    for _, peer := range peers.Peers {
        peerInfos = append(peerInfos, PeerInfo{
            ID:      peer.Peer,
            Address: peer.Addr,
        })
    }

    return peerInfos, nil
}

// 连接到对等节点
func (c *IPFSClient) ConnectToPeer(peerAddr string) error {
    err := c.shell.SwarmConnect(c.ctx, peerAddr)
    if err != nil {
        return fmt.Errorf("连接对等节点失败: %v", err)
    }
    return nil
}

// 关闭客户端
func (c *IPFSClient) Close() {
    c.cancel()
}

type NodeInfo struct {
    ID        string
    Version   string
    Addresses []string
}

type PeerInfo struct {
    ID      string
    Address string
}
```

## 文件操作

### 4.1 文件上传下载

```go
// files/manager.go
package files

import (
    "context"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "strings"

    "github.com/ipfs/go-ipfs-files"
    
    "your-project/client"
)

type FileManager struct {
    client *client.IPFSClient
}

func NewFileManager(client *client.IPFSClient) *FileManager {
    return &FileManager{
        client: client,
    }
}

// 上传文件
func (fm *FileManager) UploadFile(filePath string) (*UploadResult, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return nil, fmt.Errorf("打开文件失败: %v", err)
    }
    defer file.Close()

    // 获取文件信息
    fileInfo, err := file.Stat()
    if err != nil {
        return nil, fmt.Errorf("获取文件信息失败: %v", err)
    }

    // 上传到IPFS
    hash, err := fm.client.shell.Add(file)
    if err != nil {
        return nil, fmt.Errorf("上传文件失败: %v", err)
    }

    return &UploadResult{
        Hash:     hash,
        Name:     filepath.Base(filePath),
        Size:     fileInfo.Size(),
        MimeType: getMimeType(filePath),
    }, nil
}

// 上传字节数据
func (fm *FileManager) UploadBytes(data []byte, filename string) (*UploadResult, error) {
    reader := strings.NewReader(string(data))
    
    hash, err := fm.client.shell.Add(reader)
    if err != nil {
        return nil, fmt.Errorf("上传数据失败: %v", err)
    }

    return &UploadResult{
        Hash:     hash,
        Name:     filename,
        Size:     int64(len(data)),
        MimeType: getMimeType(filename),
    }, nil
}

// 上传字符串
func (fm *FileManager) UploadString(content, filename string) (*UploadResult, error) {
    return fm.UploadBytes([]byte(content), filename)
}

// 下载文件
func (fm *FileManager) DownloadFile(hash, outputPath string) error {
    reader, err := fm.client.shell.Cat(hash)
    if err != nil {
        return fmt.Errorf("获取文件内容失败: %v", err)
    }
    defer reader.Close()

    // 创建输出目录
    dir := filepath.Dir(outputPath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return fmt.Errorf("创建目录失败: %v", err)
    }

    // 创建输出文件
    outFile, err := os.Create(outputPath)
    if err != nil {
        return fmt.Errorf("创建输出文件失败: %v", err)
    }
    defer outFile.Close()

    // 复制内容
    _, err = io.Copy(outFile, reader)
    if err != nil {
        return fmt.Errorf("写入文件失败: %v", err)
    }

    return nil
}

// 下载为字节数组
func (fm *FileManager) DownloadBytes(hash string) ([]byte, error) {
    reader, err := fm.client.shell.Cat(hash)
    if err != nil {
        return nil, fmt.Errorf("获取文件内容失败: %v", err)
    }
    defer reader.Close()

    data, err := io.ReadAll(reader)
    if err != nil {
        return nil, fmt.Errorf("读取内容失败: %v", err)
    }

    return data, nil
}

// 下载为字符串
func (fm *FileManager) DownloadString(hash string) (string, error) {
    data, err := fm.DownloadBytes(hash)
    if err != nil {
        return "", err
    }
    return string(data), nil
}

// 获取文件信息
func (fm *FileManager) GetFileInfo(hash string) (*FileInfo, error) {
    stat, err := fm.client.shell.ObjectStat(hash)
    if err != nil {
        return nil, fmt.Errorf("获取文件统计失败: %v", err)
    }

    return &FileInfo{
        Hash:          hash,
        Size:          int64(stat.CumulativeSize),
        BlockSize:     int64(stat.BlockSize),
        NumLinks:      stat.NumLinks,
        DataSize:      int64(stat.DataSize),
    }, nil
}

// 检查文件是否存在
func (fm *FileManager) FileExists(hash string) bool {
    _, err := fm.client.shell.ObjectStat(hash)
    return err == nil
}

// 获取文件MIME类型
func getMimeType(filename string) string {
    ext := strings.ToLower(filepath.Ext(filename))
    switch ext {
    case ".txt":
        return "text/plain"
    case ".json":
        return "application/json"
    case ".html":
        return "text/html"
    case ".css":
        return "text/css"
    case ".js":
        return "application/javascript"
    case ".png":
        return "image/png"
    case ".jpg", ".jpeg":
        return "image/jpeg"
    case ".gif":
        return "image/gif"
    case ".pdf":
        return "application/pdf"
    case ".mp4":
        return "video/mp4"
    case ".mp3":
        return "audio/mpeg"
    default:
        return "application/octet-stream"
    }
}

type UploadResult struct {
    Hash     string
    Name     string
    Size     int64
    MimeType string
}

type FileInfo struct {
    Hash      string
    Size      int64
    BlockSize int64
    NumLinks  int
    DataSize  int64
}
```

## 目录管理

### 5.1 目录操作

```go
// directory/manager.go
package directory

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/ipfs/go-ipfs-files"
    
    "your-project/client"
)

type DirectoryManager struct {
    client *client.IPFSClient
}

func NewDirectoryManager(client *client.IPFSClient) *DirectoryManager {
    return &DirectoryManager{
        client: client,
    }
}

// 上传目录
func (dm *DirectoryManager) UploadDirectory(dirPath string) (*DirectoryResult, error) {
    // 检查目录是否存在
    if _, err := os.Stat(dirPath); os.IsNotExist(err) {
        return nil, fmt.Errorf("目录不存在: %s", dirPath)
    }

    // 创建文件节点
    stat, err := os.Lstat(dirPath)
    if err != nil {
        return nil, fmt.Errorf("获取目录信息失败: %v", err)
    }

    file, err := files.NewSerialFile(dirPath, false, stat)
    if err != nil {
        return nil, fmt.Errorf("创建文件节点失败: %v", err)
    }

    // 上传目录
    hash, err := dm.client.shell.AddDir(dirPath)
    if err != nil {
        return nil, fmt.Errorf("上传目录失败: %v", err)
    }

    // 获取目录大小
    dirSize, err := dm.calculateDirectorySize(dirPath)
    if err != nil {
        dirSize = 0
    }

    return &DirectoryResult{
        Hash: hash,
        Name: filepath.Base(dirPath),
        Size: dirSize,
    }, nil
}

// 列出目录内容
func (dm *DirectoryManager) ListDirectory(hash string) ([]DirectoryEntry, error) {
    links, err := dm.client.shell.List(hash)
    if err != nil {
        return nil, fmt.Errorf("列出目录内容失败: %v", err)
    }

    var entries []DirectoryEntry
    for _, link := range links {
        entry := DirectoryEntry{
            Name: link.Name,
            Hash: link.Hash,
            Size: int64(link.Size),
            Type: getEntryType(link.Type),
        }
        entries = append(entries, entry)
    }

    return entries, nil
}

// 下载目录
func (dm *DirectoryManager) DownloadDirectory(hash, outputPath string) error {
    // 创建输出目录
    if err := os.MkdirAll(outputPath, 0755); err != nil {
        return fmt.Errorf("创建输出目录失败: %v", err)
    }

    // 获取目录内容
    entries, err := dm.ListDirectory(hash)
    if err != nil {
        return fmt.Errorf("获取目录内容失败: %v", err)
    }

    // 递归下载每个条目
    for _, entry := range entries {
        entryPath := filepath.Join(outputPath, entry.Name)
        
        if entry.Type == "directory" {
            // 递归下载子目录
            err := dm.DownloadDirectory(entry.Hash, entryPath)
            if err != nil {
                return fmt.Errorf("下载子目录失败: %v", err)
            }
        } else {
            // 下载文件
            err := dm.downloadFile(entry.Hash, entryPath)
            if err != nil {
                return fmt.Errorf("下载文件失败: %v", err)
            }
        }
    }

    return nil
}

// 创建虚拟目录
func (dm *DirectoryManager) CreateVirtualDirectory(files map[string]string) (*DirectoryResult, error) {
    // 创建临时目录
    tempDir, err := os.MkdirTemp("", "ipfs_virtual_dir")
    if err != nil {
        return nil, fmt.Errorf("创建临时目录失败: %v", err)
    }
    defer os.RemoveAll(tempDir)

    // 创建文件
    for filename, content := range files {
        filePath := filepath.Join(tempDir, filename)
        
        // 创建子目录（如果需要）
        dir := filepath.Dir(filePath)
        if err := os.MkdirAll(dir, 0755); err != nil {
            return nil, fmt.Errorf("创建子目录失败: %v", err)
        }

        // 写入文件内容
        if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
            return nil, fmt.Errorf("写入文件失败: %v", err)
        }
    }

    // 上传目录
    return dm.UploadDirectory(tempDir)
}

// 计算目录大小
func (dm *DirectoryManager) calculateDirectorySize(dirPath string) (int64, error) {
    var size int64
    
    err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if !info.IsDir() {
            size += info.Size()
        }
        return nil
    })

    return size, err
}

// 下载单个文件
func (dm *DirectoryManager) downloadFile(hash, outputPath string) error {
    reader, err := dm.client.shell.Cat(hash)
    if err != nil {
        return err
    }
    defer reader.Close()

    outFile, err := os.Create(outputPath)
    if err != nil {
        return err
    }
    defer outFile.Close()

    _, err = outFile.ReadFrom(reader)
    return err
}

// 获取条目类型
func getEntryType(linkType int) string {
    switch linkType {
    case 1:
        return "directory"
    case 2:
        return "file"
    default:
        return "unknown"
    }
}

type DirectoryResult struct {
    Hash string
    Name string
    Size int64
}

type DirectoryEntry struct {
    Name string
    Hash string
    Size int64
    Type string
}
```

## 网络发现

### 6.1 内容发现

```go
// discovery/content.go
package discovery

import (
    "context"
    "fmt"
    "time"

    "your-project/client"
)

type ContentDiscovery struct {
    client *client.IPFSClient
}

func NewContentDiscovery(client *client.IPFSClient) *ContentDiscovery {
    return &ContentDiscovery{
        client: client,
    }
}

// 查找内容提供者
func (cd *ContentDiscovery) FindProviders(hash string, maxProviders int) ([]ProviderInfo, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    providers, err := cd.client.shell.FindProvs(hash)
    if err != nil {
        return nil, fmt.Errorf("查找提供者失败: %v", err)
    }

    var providerInfos []ProviderInfo
    count := 0
    
    for provider := range providers {
        if count >= maxProviders {
            break
        }
        
        providerInfos = append(providerInfos, ProviderInfo{
            ID:      provider.ID,
            Address: provider.Addrs,
        })
        count++
    }

    return providerInfos, nil
}

// 宣告内容
func (cd *ContentDiscovery) Provide(hash string) error {
    err := cd.client.shell.DhtProvide(hash)
    if err != nil {
        return fmt.Errorf("宣告内容失败: %v", err)
    }
    return nil
}

// 查找对等节点
func (cd *ContentDiscovery) FindPeer(peerID string) (*PeerLocation, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer cancel()

    peerInfo, err := cd.client.shell.WithContext(ctx).DhtFindPeer(peerID)
    if err != nil {
        return nil, fmt.Errorf("查找对等节点失败: %v", err)
    }

    return &PeerLocation{
        ID:        peerInfo.ID,
        Addresses: peerInfo.Addrs,
    }, nil
}

// 获取路由表信息
func (cd *ContentDiscovery) GetRoutingTable() (*RoutingTableInfo, error) {
    // 这里需要使用更底层的API来获取路由表信息
    // 简化实现，返回基本信息
    peers, err := cd.client.GetPeers()
    if err != nil {
        return nil, err
    }

    return &RoutingTableInfo{
        PeerCount: len(peers),
        Peers:     peers,
    }, nil
}

// 解析内容
func (cd *ContentDiscovery) ResolvePath(path string) (string, error) {
    resolved, err := cd.client.shell.Resolve(path)
    if err != nil {
        return "", fmt.Errorf("解析路径失败: %v", err)
    }
    return resolved, nil
}

type ProviderInfo struct {
    ID      string
    Address []string
}

type PeerLocation struct {
    ID        string
    Addresses []string
}

type RoutingTableInfo struct {
    PeerCount int
    Peers     []client.PeerInfo
}
```

## 内容寻址

### 7.1 CID管理

```go
// cid/manager.go
package cid

import (
    "fmt"
    "strings"

    "github.com/ipfs/go-cid"
    "github.com/multiformats/go-multihash"
    
    "your-project/client"
)

type CIDManager struct {
    client *client.IPFSClient
}

func NewCIDManager(client *client.IPFSClient) *CIDManager {
    return &CIDManager{
        client: client,
    }
}

// 解析CID
func (cm *CIDManager) ParseCID(cidStr string) (*CIDInfo, error) {
    c, err := cid.Decode(cidStr)
    if err != nil {
        return nil, fmt.Errorf("解析CID失败: %v", err)
    }

    return &CIDInfo{
        CID:       c,
        Version:   int(c.Version()),
        Codec:     c.Type().String(),
        Multihash: c.Hash(),
    }, nil
}

// 创建CID
func (cm *CIDManager) CreateCID(data []byte, version int) (*CIDInfo, error) {
    // 计算哈希
    hash, err := multihash.Sum(data, multihash.SHA2_256, -1)
    if err != nil {
        return nil, fmt.Errorf("计算哈希失败: %v", err)
    }

    // 创建CID
    var c cid.Cid
    if version == 0 {
        c = cid.NewCidV0(hash)
    } else {
        c = cid.NewCidV1(cid.Raw, hash)
    }

    return &CIDInfo{
        CID:       c,
        Version:   version,
        Codec:     c.Type().String(),
        Multihash: hash,
    }, nil
}

// 转换CID版本
func (cm *CIDManager) ConvertCIDVersion(cidStr string, targetVersion int) (string, error) {
    c, err := cid.Decode(cidStr)
    if err != nil {
        return "", fmt.Errorf("解析CID失败: %v", err)
    }

    var newCID cid.Cid
    if targetVersion == 0 && c.Version() == 1 {
        newCID = cid.NewCidV0(c.Hash())
    } else if targetVersion == 1 && c.Version() == 0 {
        newCID = cid.NewCidV1(cid.DagProtobuf, c.Hash())
    } else {
        return cidStr, nil // 已经是目标版本
    }

    return newCID.String(), nil
}

// 验证CID
func (cm *CIDManager) ValidateCID(cidStr string) bool {
    _, err := cid.Decode(cidStr)
    return err == nil
}

// 获取CID信息
func (cm *CIDManager) GetCIDInfo(cidStr string) (*DetailedCIDInfo, error) {
    cidInfo, err := cm.ParseCID(cidStr)
    if err != nil {
        return nil, err
    }

    // 获取内容统计
    stat, err := cm.client.shell.ObjectStat(cidStr)
    if err != nil {
        return nil, fmt.Errorf("获取内容统计失败: %v", err)
    }

    return &DetailedCIDInfo{
        CIDInfo:        *cidInfo,
        Size:           int64(stat.CumulativeSize),
        BlockSize:      int64(stat.BlockSize),
        NumLinks:       stat.NumLinks,
        DataSize:       int64(stat.DataSize),
    }, nil
}

// 比较CID
func (cm *CIDManager) CompareCIDs(cid1, cid2 string) (*CIDComparison, error) {
    c1, err := cid.Decode(cid1)
    if err != nil {
        return nil, fmt.Errorf("解析CID1失败: %v", err)
    }

    c2, err := cid.Decode(cid2)
    if err != nil {
        return nil, fmt.Errorf("解析CID2失败: %v", err)
    }

    return &CIDComparison{
        Equal:          c1.Equals(c2),
        SameHash:       c1.Hash().Equal(c2.Hash()),
        SameVersion:    c1.Version() == c2.Version(),
        SameCodec:      c1.Type() == c2.Type(),
    }, nil
}

// 从路径提取CID
func (cm *CIDManager) ExtractCIDFromPath(path string) (string, error) {
    // 处理 /ipfs/CID 格式的路径
    if strings.HasPrefix(path, "/ipfs/") {
        parts := strings.Split(path, "/")
        if len(parts) >= 3 {
            cidStr := parts[2]
            if cm.ValidateCID(cidStr) {
                return cidStr, nil
            }
        }
    }

    // 直接验证是否为CID
    if cm.ValidateCID(path) {
        return path, nil
    }

    return "", fmt.Errorf("无法从路径提取有效CID: %s", path)
}

type CIDInfo struct {
    CID       cid.Cid
    Version   int
    Codec     string
    Multihash multihash.Multihash
}

type DetailedCIDInfo struct {
    CIDInfo
    Size      int64
    BlockSize int64
    NumLinks  int
    DataSize  int64
}

type CIDComparison struct {
    Equal       bool
    SameHash    bool
    SameVersion bool
    SameCodec   bool
}
```

## 高级功能

### 8.1 固定管理

```go
// pin/manager.go
package pin

import (
    "context"
    "fmt"
    "time"

    "your-project/client"
)

type PinManager struct {
    client *client.IPFSClient
}

func NewPinManager(client *client.IPFSClient) *PinManager {
    return &PinManager{
        client: client,
    }
}

// 固定内容
func (pm *PinManager) Pin(hash string, recursive bool) error {
    var err error
    if recursive {
        err = pm.client.shell.Pin(hash)
    } else {
        err = pm.client.shell.PinWithMode(hash, "direct")
    }

    if err != nil {
        return fmt.Errorf("固定内容失败: %v", err)
    }
    return nil
}

// 取消固定
func (pm *PinManager) Unpin(hash string, recursive bool) error {
    var err error
    if recursive {
        err = pm.client.shell.Unpin(hash)
    } else {
        err = pm.client.shell.UnpinWithMode(hash, "direct")
    }

    if err != nil {
        return fmt.Errorf("取消固定失败: %v", err)
    }
    return nil
}

// 列出固定的内容
func (pm *PinManager) ListPins() ([]PinInfo, error) {
    pins, err := pm.client.shell.Pins()
    if err != nil {
        return nil, fmt.Errorf("获取固定列表失败: %v", err)
    }

    var pinInfos []PinInfo
    for hash, pinType := range pins {
        pinInfos = append(pinInfos, PinInfo{
            Hash: hash,
            Type: pinType,
        })
    }

    return pinInfos, nil
}

// 检查内容是否被固定
func (pm *PinManager) IsPinned(hash string) (bool, string, error) {
    pins, err := pm.client.shell.Pins()
    if err != nil {
        return false, "", fmt.Errorf("获取固定列表失败: %v", err)
    }

    pinType, exists := pins[hash]
    return exists, pinType, nil
}

// 垃圾回收
func (pm *PinManager) GarbageCollect() (*GCResult, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    // 执行垃圾回收
    gcOutput, err := pm.client.shell.WithContext(ctx).Request("repo/gc").Send()
    if err != nil {
        return nil, fmt.Errorf("垃圾回收失败: %v", err)
    }

    // 解析垃圾回收结果
    var removedCount int
    var reclaimedSpace int64

    // 这里需要解析gcOutput来获取实际的统计信息
    // 简化实现，返回基本信息
    return &GCResult{
        RemovedObjects: removedCount,
        ReclaimedSpace: reclaimedSpace,
    }, nil
}

// 获取仓库统计
func (pm *PinManager) GetRepoStats() (*RepoStats, error) {
    stat, err := pm.client.shell.RepoStat()
    if err != nil {
        return nil, fmt.Errorf("获取仓库统计失败: %v", err)
    }

    return &RepoStats{
        RepoSize:   int64(stat.RepoSize),
        StorageMax: int64(stat.StorageMax),
        NumObjects: int64(stat.NumObjects),
        RepoPath:   stat.RepoPath,
        Version:    stat.Version,
    }, nil
}

type PinInfo struct {
    Hash string
    Type string
}

type GCResult struct {
    RemovedObjects int
    ReclaimedSpace int64
}

type RepoStats struct {
    RepoSize   int64
    StorageMax int64
    NumObjects int64
    RepoPath   string
    Version    string
}
```

## 实际应用

### 9.1 完整IPFS应用

```go
// main.go
package main

import (
    "fmt"
    "log"
    "os"
    "path/filepath"

    "your-project/config"
    "your-project/client"
    "your-project/files"
    "your-project/directory"
    "your-project/discovery"
    "your-project/cid"
    "your-project/pin"
)

func main() {
    // 创建IPFS配置
    cfg := config.DefaultIPFSConfig()

    // 创建IPFS客户端
    ipfsClient, err := client.NewIPFSClient(cfg)
    if err != nil {
        log.Fatal("创建IPFS客户端失败:", err)
    }
    defer ipfsClient.Close()

    // 获取节点信息
    nodeInfo, err := ipfsClient.GetNodeInfo()
    if err != nil {
        log.Printf("获取节点信息失败: %v", err)
    } else {
        fmt.Printf("IPFS节点信息:\n")
        fmt.Printf("  ID: %s\n", nodeInfo.ID)
        fmt.Printf("  版本: %s\n", nodeInfo.Version)
    }

    // 检查节点状态
    if ipfsClient.IsOnline() {
        fmt.Println("IPFS节点在线")
    } else {
        fmt.Println("IPFS节点离线")
        return
    }

    // 文件操作示例
    fileManager := files.NewFileManager(ipfsClient)

    // 上传文本文件
    testContent := "Hello, IPFS! 这是一个测试文件。"
    uploadResult, err := fileManager.UploadString(testContent, "test.txt")
    if err != nil {
        log.Printf("上传文件失败: %v", err)
    } else {
        fmt.Printf("文件上传成功:\n")
        fmt.Printf("  哈希: %s\n", uploadResult.Hash)
        fmt.Printf("  名称: %s\n", uploadResult.Name)
        fmt.Printf("  大小: %d 字节\n", uploadResult.Size)
        fmt.Printf("  MIME类型: %s\n", uploadResult.MimeType)

        // 下载文件
        downloadedContent, err := fileManager.DownloadString(uploadResult.Hash)
        if err != nil {
            log.Printf("下载文件失败: %v", err)
        } else {
            fmt.Printf("下载内容: %s\n", downloadedContent)
        }

        // 获取文件信息
        fileInfo, err := fileManager.GetFileInfo(uploadResult.Hash)
        if err != nil {
            log.Printf("获取文件信息失败: %v", err)
        } else {
            fmt.Printf("文件详细信息:\n")
            fmt.Printf("  累计大小: %d 字节\n", fileInfo.Size)
            fmt.Printf("  数据大小: %d 字节\n", fileInfo.DataSize)
            fmt.Printf("  链接数: %d\n", fileInfo.NumLinks)
        }
    }

    // 目录操作示例
    dirManager := directory.NewDirectoryManager(ipfsClient)

    // 创建虚拟目录
    virtualFiles := map[string]string{
        "readme.txt":      "这是一个README文件",
        "config.json":     `{"name": "test", "version": "1.0"}`,
        "docs/guide.md":   "# 使用指南\n\n这是使用指南。",
    }

    dirResult, err := dirManager.CreateVirtualDirectory(virtualFiles)
    if err != nil {
        log.Printf("创建虚拟目录失败: %v", err)
    } else {
        fmt.Printf("虚拟目录创建成功:\n")
        fmt.Printf("  哈希: %s\n", dirResult.Hash)
        fmt.Printf("  名称: %s\n", dirResult.Name)
        fmt.Printf("  大小: %d 字节\n", dirResult.Size)

        // 列出目录内容
        entries, err := dirManager.ListDirectory(dirResult.Hash)
        if err != nil {
            log.Printf("列出目录内容失败: %v", err)
        } else {
            fmt.Printf("目录内容:\n")
            for _, entry := range entries {
                fmt.Printf("  %s (%s) - %s - %d 字节\n", 
                    entry.Name, entry.Type, entry.Hash, entry.Size)
            }
        }
    }

    // CID管理示例
    cidManager := cid.NewCIDManager(ipfsClient)

    if uploadResult != nil {
        // 解析CID
        cidInfo, err := cidManager.ParseCID(uploadResult.Hash)
        if err != nil {
            log.Printf("解析CID失败: %v", err)
        } else {
            fmt.Printf("CID信息:\n")
            fmt.Printf("  版本: %d\n", cidInfo.Version)
            fmt.Printf("  编解码器: %s\n", cidInfo.Codec)
        }

        // 转换CID版本
        if cidInfo.Version == 0 {
            v1CID, err := cidManager.ConvertCIDVersion(uploadResult.Hash, 1)
            if err != nil {
                log.Printf("转换CID版本失败: %v", err)
            } else {
                fmt.Printf("CIDv1: %s\n", v1CID)
            }
        }
    }

    // 固定管理示例
    pinManager := pin.NewPinManager(ipfsClient)

    if uploadResult != nil {
        // 固定文件
        err := pinManager.Pin(uploadResult.Hash, false)
        if err != nil {
            log.Printf("固定文件失败: %v", err)
        } else {
            fmt.Printf("文件已固定: %s\n", uploadResult.Hash)
        }

        // 检查是否被固定
        isPinned, pinType, err := pinManager.IsPinned(uploadResult.Hash)
        if err != nil {
            log.Printf("检查固定状态失败: %v", err)
        } else {
            fmt.Printf("固定状态: %t, 类型: %s\n", isPinned, pinType)
        }
    }

    // 列出所有固定的内容
    pins, err := pinManager.ListPins()
    if err != nil {
        log.Printf("获取固定列表失败: %v", err)
    } else {
        fmt.Printf("固定的内容数量: %d\n", len(pins))
        for i, pin := range pins {
            if i < 5 { // 只显示前5个
                fmt.Printf("  %s (%s)\n", pin.Hash, pin.Type)
            }
        }
    }

    // 获取仓库统计
    repoStats, err := pinManager.GetRepoStats()
    if err != nil {
        log.Printf("获取仓库统计失败: %v", err)
    } else {
        fmt.Printf("仓库统计:\n")
        fmt.Printf("  仓库大小: %d 字节\n", repoStats.RepoSize)
        fmt.Printf("  最大存储: %d 字节\n", repoStats.StorageMax)
        fmt.Printf("  对象数量: %d\n", repoStats.NumObjects)
        fmt.Printf("  版本: %s\n", repoStats.Version)
    }

    // 内容发现示例
    contentDiscovery := discovery.NewContentDiscovery(ipfsClient)

    if uploadResult != nil {
        // 宣告内容
        err := contentDiscovery.Provide(uploadResult.Hash)
        if err != nil {
            log.Printf("宣告内容失败: %v", err)
        } else {
            fmt.Printf("内容已宣告: %s\n", uploadResult.Hash)
        }

        // 查找内容提供者
        providers, err := contentDiscovery.FindProviders(uploadResult.Hash, 5)
        if err != nil {
            log.Printf("查找提供者失败: %v", err)
        } else {
            fmt.Printf("找到 %d 个内容提供者\n", len(providers))
            for i, provider := range providers {
                fmt.Printf("  提供者 %d: %s\n", i+1, provider.ID)
            }
        }
    }

    // 获取对等节点信息
    peers, err := ipfsClient.GetPeers()
    if err != nil {
        log.Printf("获取对等节点失败: %v", err)
    } else {
        fmt.Printf("连接的对等节点数量: %d\n", len(peers))
        for i, peer := range peers {
            if i < 3 { // 只显示前3个
                fmt.Printf("  节点 %d: %s\n", i+1, peer.ID)
            }
        }
    }

    fmt.Println("IPFS操作演示完成!")
}
```
