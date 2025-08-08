# Hyperledger Fabric 详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [环境准备](#环境准备)
3. [网络架构](#网络架构)
4. [智能合约开发](#智能合约开发)
5. [客户端SDK](#客户端SDK)
6. [权限管理](#权限管理)
7. [通道管理](#通道管理)
8. [监控和运维](#监控和运维)
9. [实际应用](#实际应用)

## 基础概念

### 1.1 Hyperledger Fabric简介

Hyperledger Fabric是企业级联盟链平台，支持权限管理、隐私保护和模块化架构。

```bash
# 安装Fabric SDK
go get github.com/hyperledger/fabric-sdk-go
go get github.com/hyperledger/fabric-contract-api-go
go get github.com/hyperledger/fabric-chaincode-go
```

### 1.2 核心组件

```go
// 主要包导入
import (
    "github.com/hyperledger/fabric-contract-api-go/contractapi"
    "github.com/hyperledger/fabric-chaincode-go/shim"
    "github.com/hyperledger/fabric-sdk-go/pkg/gateway"
    "github.com/hyperledger/fabric-sdk-go/pkg/core/config"
)
```

## 环境准备

### 2.1 网络配置

```yaml
# network-config.yaml
name: "fabric-network"
version: "1.0"

client:
  organization: Org1
  logging:
    level: info
  peer:
    timeout:
      connection: 10s
      response: 180s
      discovery:
        greylistExpiry: 10s
  eventService:
    timeout:
      registrationResponse: 15s
  orderer:
    timeout:
      connection: 15s
      response: 15s

channels:
  mychannel:
    orderers:
      - orderer.example.com
    peers:
      peer0.org1.example.com:
        endorsingPeer: true
        chaincodeQuery: true
        ledgerQuery: true
        eventSource: true

organizations:
  Org1:
    mspid: Org1MSP
    peers:
      - peer0.org1.example.com
    certificateAuthorities:
      - ca.org1.example.com

orderers:
  orderer.example.com:
    url: grpcs://orderer.example.com:7050
    tlsCACerts:
      path: crypto-config/ordererOrganizations/example.com/tlsca/tlsca.example.com-cert.pem

peers:
  peer0.org1.example.com:
    url: grpcs://peer0.org1.example.com:7051
    tlsCACerts:
      path: crypto-config/peerOrganizations/org1.example.com/tlsca/tlsca.org1.example.com-cert.pem

certificateAuthorities:
  ca.org1.example.com:
    url: https://ca.org1.example.com:7054
    tlsCACerts:
      path: crypto-config/peerOrganizations/org1.example.com/ca/ca.org1.example.com-cert.pem
```

### 2.2 连接配置

```go
// config/fabric.go
package config

import (
    "path/filepath"
    
    "github.com/hyperledger/fabric-sdk-go/pkg/core/config"
    "github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

type FabricConfig struct {
    ConfigPath     string
    WalletPath     string
    UserName       string
    ChannelName    string
    ChaincodeName  string
    OrgName        string
    PeerEndpoint   string
}

func DefaultConfig() *FabricConfig {
    return &FabricConfig{
        ConfigPath:    "network-config.yaml",
        WalletPath:    "wallet",
        UserName:      "appUser",
        ChannelName:   "mychannel",
        ChaincodeName: "mycontract",
        OrgName:       "Org1",
        PeerEndpoint:  "peer0.org1.example.com",
    }
}

// 创建网关连接
func CreateGateway(config *FabricConfig) (*gateway.Gateway, error) {
    // 加载网络配置
    configProvider := config.FromFile(filepath.Clean(config.ConfigPath))

    // 创建钱包
    wallet, err := gateway.NewFileSystemWallet(config.WalletPath)
    if err != nil {
        return nil, err
    }

    // 检查用户身份
    if !wallet.Exists(config.UserName) {
        return nil, fmt.Errorf("用户 %s 不存在于钱包中", config.UserName)
    }

    // 创建网关
    gw, err := gateway.Connect(
        gateway.WithConfig(configProvider),
        gateway.WithIdentity(wallet, config.UserName),
    )
    if err != nil {
        return nil, err
    }

    return gw, nil
}
```

## 网络架构

### 3.1 组织和节点管理

```go
// network/organization.go
package network

import (
    "crypto/x509"
    
    "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
    "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
    "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

type Organization struct {
    Name     string
    MSPID    string
    Peers    []string
    CAName   string
    AdminKey string
    AdminCert string
}

type NetworkManager struct {
    organizations map[string]*Organization
    channels      map[string]*Channel
}

func NewNetworkManager() *NetworkManager {
    return &NetworkManager{
        organizations: make(map[string]*Organization),
        channels:      make(map[string]*Channel),
    }
}

// 添加组织
func (nm *NetworkManager) AddOrganization(org *Organization) {
    nm.organizations[org.Name] = org
}

// 获取组织
func (nm *NetworkManager) GetOrganization(name string) (*Organization, bool) {
    org, exists := nm.organizations[name]
    return org, exists
}

// 注册用户
func (nm *NetworkManager) RegisterUser(orgName, userName, userType string) error {
    org, exists := nm.GetOrganization(orgName)
    if !exists {
        return fmt.Errorf("组织 %s 不存在", orgName)
    }

    // 创建MSP客户端
    mspClient, err := msp.New(context.Background())
    if err != nil {
        return err
    }

    // 注册用户
    secret, err := mspClient.Register(&msp.RegistrationRequest{
        Name:   userName,
        Type:   userType,
        Affiliation: org.Name,
    })
    if err != nil {
        return err
    }

    // 注册用户
    err = mspClient.Enroll(userName, msp.WithSecret(secret))
    if err != nil {
        return err
    }

    return nil
}
```

### 3.2 通道管理

```go
// network/channel.go
package network

import (
    "github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
    "github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
    "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

type Channel struct {
    Name         string
    Organizations []string
    Peers        []string
    Orderers     []string
}

// 创建通道
func (nm *NetworkManager) CreateChannel(channelName string, channelConfig []byte, ordererEndpoint string) error {
    // 创建资源管理客户端
    resMgmtClient, err := resmgmt.New(context.Background())
    if err != nil {
        return err
    }

    // 创建通道请求
    req := resmgmt.SaveChannelRequest{
        ChannelID:         channelName,
        ChannelConfigPath: string(channelConfig),
        SigningIdentities: []msp.SigningIdentity{},
    }

    // 创建通道
    _, err = resMgmtClient.SaveChannel(req, resmgmt.WithOrdererEndpoint(ordererEndpoint))
    if err != nil {
        return err
    }

    return nil
}

// 加入通道
func (nm *NetworkManager) JoinChannel(channelName, peerEndpoint string) error {
    resMgmtClient, err := resmgmt.New(context.Background())
    if err != nil {
        return err
    }

    // 加入通道
    err = resMgmtClient.JoinChannel(channelName, resmgmt.WithTargetEndpoints(peerEndpoint))
    if err != nil {
        return err
    }

    return nil
}

// 安装链码
func (nm *NetworkManager) InstallChaincode(chaincodeName, version, path string, peers []string) error {
    resMgmtClient, err := resmgmt.New(context.Background())
    if err != nil {
        return err
    }

    // 安装链码请求
    req := resmgmt.InstallCCRequest{
        Name:    chaincodeName,
        Path:    path,
        Version: version,
        Package: &resource.CCPackage{Type: pb.ChaincodeSpec_GOLANG, Code: path},
    }

    // 安装链码
    _, err = resMgmtClient.InstallCC(req, resmgmt.WithTargetEndpoints(peers...))
    if err != nil {
        return err
    }

    return nil
}
```

## 智能合约开发

### 4.1 合约结构

```go
// chaincode/asset.go
package main

import (
    "encoding/json"
    "fmt"
    "log"

    "github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// 资产结构
type Asset struct {
    ID             string `json:"ID"`
    Color          string `json:"color"`
    Size           int    `json:"size"`
    Owner          string `json:"owner"`
    AppraisedValue int    `json:"appraisedValue"`
}

// 智能合约结构
type SmartContract struct {
    contractapi.Contract
}

// 初始化账本
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
    assets := []Asset{
        {ID: "asset1", Color: "blue", Size: 5, Owner: "Tomoko", AppraisedValue: 300},
        {ID: "asset2", Color: "red", Size: 5, Owner: "Brad", AppraisedValue: 400},
        {ID: "asset3", Color: "green", Size: 10, Owner: "Jin Soo", AppraisedValue: 500},
    }

    for _, asset := range assets {
        assetJSON, err := json.Marshal(asset)
        if err != nil {
            return err
        }

        err = ctx.GetStub().PutState(asset.ID, assetJSON)
        if err != nil {
            return fmt.Errorf("failed to put to world state. %v", err)
        }
    }

    return nil
}

// 创建资产
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, id string, color string, size int, owner string, appraisedValue int) error {
    exists, err := s.AssetExists(ctx, id)
    if err != nil {
        return err
    }
    if exists {
        return fmt.Errorf("the asset %s already exists", id)
    }

    asset := Asset{
        ID:             id,
        Color:          color,
        Size:           size,
        Owner:          owner,
        AppraisedValue: appraisedValue,
    }
    assetJSON, err := json.Marshal(asset)
    if err != nil {
        return err
    }

    return ctx.GetStub().PutState(id, assetJSON)
}

// 读取资产
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, id string) (*Asset, error) {
    assetJSON, err := ctx.GetStub().GetState(id)
    if err != nil {
        return nil, fmt.Errorf("failed to read from world state: %v", err)
    }
    if assetJSON == nil {
        return nil, fmt.Errorf("the asset %s does not exist", id)
    }

    var asset Asset
    err = json.Unmarshal(assetJSON, &asset)
    if err != nil {
        return nil, err
    }

    return &asset, nil
}

// 更新资产
func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, id string, color string, size int, owner string, appraisedValue int) error {
    exists, err := s.AssetExists(ctx, id)
    if err != nil {
        return err
    }
    if !exists {
        return fmt.Errorf("the asset %s does not exist", id)
    }

    // 覆写原始资产
    asset := Asset{
        ID:             id,
        Color:          color,
        Size:           size,
        Owner:          owner,
        AppraisedValue: appraisedValue,
    }
    assetJSON, err := json.Marshal(asset)
    if err != nil {
        return err
    }

    return ctx.GetStub().PutState(id, assetJSON)
}

// 删除资产
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, id string) error {
    exists, err := s.AssetExists(ctx, id)
    if err != nil {
        return err
    }
    if !exists {
        return fmt.Errorf("the asset %s does not exist", id)
    }

    return ctx.GetStub().DelState(id)
}

// 检查资产是否存在
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
    assetJSON, err := ctx.GetStub().GetState(id)
    if err != nil {
        return false, fmt.Errorf("failed to read from world state: %v", err)
    }

    return assetJSON != nil, nil
}

// 转移资产
func (s *SmartContract) TransferAsset(ctx contractapi.TransactionContextInterface, id string, newOwner string) error {
    asset, err := s.ReadAsset(ctx, id)
    if err != nil {
        return err
    }

    asset.Owner = newOwner
    assetJSON, err := json.Marshal(asset)
    if err != nil {
        return err
    }

    return ctx.GetStub().PutState(id, assetJSON)
}

// 获取所有资产
func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*Asset, error) {
    resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
    if err != nil {
        return nil, err
    }
    defer resultsIterator.Close()

    var assets []*Asset
    for resultsIterator.HasNext() {
        queryResponse, err := resultsIterator.Next()
        if err != nil {
            return nil, err
        }

        var asset Asset
        err = json.Unmarshal(queryResponse.Value, &asset)
        if err != nil {
            return nil, err
        }
        assets = append(assets, &asset)
    }

    return assets, nil
}

func main() {
    assetChaincode, err := contractapi.NewChaincode(&SmartContract{})
    if err != nil {
        log.Panicf("Error creating asset-transfer-basic chaincode: %v", err)
    }

    if err := assetChaincode.Start(); err != nil {
        log.Panicf("Error starting asset-transfer-basic chaincode: %v", err)
    }
}
```

## 客户端SDK

### 5.1 客户端操作

```go
// client/fabric_client.go
package client

import (
    "fmt"
    "log"

    "github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

type FabricClient struct {
    gateway  *gateway.Gateway
    network  *gateway.Network
    contract *gateway.Contract
}

func NewFabricClient(configPath, walletPath, userName, channelName, chaincodeName string) (*FabricClient, error) {
    // 创建网关
    gw, err := CreateGateway(&FabricConfig{
        ConfigPath:    configPath,
        WalletPath:    walletPath,
        UserName:      userName,
        ChannelName:   channelName,
        ChaincodeName: chaincodeName,
    })
    if err != nil {
        return nil, err
    }

    // 获取网络
    network, err := gw.GetNetwork(channelName)
    if err != nil {
        return nil, err
    }

    // 获取合约
    contract := network.GetContract(chaincodeName)

    return &FabricClient{
        gateway:  gw,
        network:  network,
        contract: contract,
    }, nil
}

// 关闭连接
func (fc *FabricClient) Close() {
    fc.gateway.Close()
}

// 创建资产
func (fc *FabricClient) CreateAsset(id, color string, size int, owner string, appraisedValue int) error {
    _, err := fc.contract.SubmitTransaction("CreateAsset", id, color, fmt.Sprintf("%d", size), owner, fmt.Sprintf("%d", appraisedValue))
    if err != nil {
        return fmt.Errorf("failed to submit transaction: %v", err)
    }

    log.Printf("资产 %s 创建成功", id)
    return nil
}

// 读取资产
func (fc *FabricClient) ReadAsset(id string) ([]byte, error) {
    result, err := fc.contract.EvaluateTransaction("ReadAsset", id)
    if err != nil {
        return nil, fmt.Errorf("failed to evaluate transaction: %v", err)
    }

    return result, nil
}

// 更新资产
func (fc *FabricClient) UpdateAsset(id, color string, size int, owner string, appraisedValue int) error {
    _, err := fc.contract.SubmitTransaction("UpdateAsset", id, color, fmt.Sprintf("%d", size), owner, fmt.Sprintf("%d", appraisedValue))
    if err != nil {
        return fmt.Errorf("failed to submit transaction: %v", err)
    }

    log.Printf("资产 %s 更新成功", id)
    return nil
}

// 转移资产
func (fc *FabricClient) TransferAsset(id, newOwner string) error {
    _, err := fc.contract.SubmitTransaction("TransferAsset", id, newOwner)
    if err != nil {
        return fmt.Errorf("failed to submit transaction: %v", err)
    }

    log.Printf("资产 %s 转移给 %s 成功", id, newOwner)
    return nil
}

// 获取所有资产
func (fc *FabricClient) GetAllAssets() ([]byte, error) {
    result, err := fc.contract.EvaluateTransaction("GetAllAssets")
    if err != nil {
        return nil, fmt.Errorf("failed to evaluate transaction: %v", err)
    }

    return result, nil
}

// 查询资产历史
func (fc *FabricClient) GetAssetHistory(id string) ([]byte, error) {
    result, err := fc.contract.EvaluateTransaction("GetAssetHistory", id)
    if err != nil {
        return nil, fmt.Errorf("failed to evaluate transaction: %v", err)
    }

    return result, nil
}
```

## 权限管理

### 6.1 访问控制

```go
// chaincode/access_control.go
package main

import (
    "fmt"
    
    "github.com/hyperledger/fabric-chaincode-go/pkg/cid"
    "github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// 权限检查
func (s *SmartContract) checkPermission(ctx contractapi.TransactionContextInterface, requiredRole string) error {
    // 获取客户端身份
    clientID, err := cid.GetID(ctx.GetStub())
    if err != nil {
        return fmt.Errorf("failed to get client identity: %v", err)
    }

    // 获取客户端MSPID
    mspID, err := cid.GetMSPID(ctx.GetStub())
    if err != nil {
        return fmt.Errorf("failed to get MSP ID: %v", err)
    }

    // 检查属性
    hasRole, err := cid.HasOUValue(ctx.GetStub(), requiredRole)
    if err != nil {
        return fmt.Errorf("failed to check role: %v", err)
    }

    if !hasRole {
        return fmt.Errorf("client %s from MSP %s does not have required role %s", clientID, mspID, requiredRole)
    }

    return nil
}

// 需要管理员权限的创建资产
func (s *SmartContract) CreateAssetWithPermission(ctx contractapi.TransactionContextInterface, id string, color string, size int, owner string, appraisedValue int) error {
    // 检查管理员权限
    err := s.checkPermission(ctx, "admin")
    if err != nil {
        return err
    }

    return s.CreateAsset(ctx, id, color, size, owner, appraisedValue)
}

// 基于组织的权限检查
func (s *SmartContract) checkOrganization(ctx contractapi.TransactionContextInterface, allowedMSPs []string) error {
    mspID, err := cid.GetMSPID(ctx.GetStub())
    if err != nil {
        return fmt.Errorf("failed to get MSP ID: %v", err)
    }

    for _, allowedMSP := range allowedMSPs {
        if mspID == allowedMSP {
            return nil
        }
    }

    return fmt.Errorf("MSP %s is not authorized", mspID)
}

// 获取调用者信息
func (s *SmartContract) GetCallerInfo(ctx contractapi.TransactionContextInterface) (map[string]string, error) {
    info := make(map[string]string)

    // 获取客户端ID
    clientID, err := cid.GetID(ctx.GetStub())
    if err != nil {
        return nil, err
    }
    info["clientID"] = clientID

    // 获取MSPID
    mspID, err := cid.GetMSPID(ctx.GetStub())
    if err != nil {
        return nil, err
    }
    info["mspID"] = mspID

    // 获取属性
    attrs, err := cid.GetAttributeValue(ctx.GetStub(), "role")
    if err == nil {
        info["role"] = attrs
    }

    return info, nil
}
```

## 通道管理

### 7.1 私有数据

```go
// chaincode/private_data.go
package main

import (
    "encoding/json"
    "fmt"

    "github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// 私有资产数据
type AssetPrivateDetails struct {
    ID             string `json:"assetID"`
    AppraisedValue int    `json:"appraisedValue"`
}

// 创建私有资产
func (s *SmartContract) CreateAssetPrivateData(ctx contractapi.TransactionContextInterface, collection string) error {
    // 获取瞬态数据
    transientMap, err := ctx.GetStub().GetTransient()
    if err != nil {
        return fmt.Errorf("error getting transient: %v", err)
    }

    // 获取私有数据
    assetPrivateDetailsJSON, ok := transientMap["asset_private_details"]
    if !ok {
        return fmt.Errorf("asset_private_details key not found in the transient map")
    }

    var assetPrivateDetails AssetPrivateDetails
    err = json.Unmarshal(assetPrivateDetailsJSON, &assetPrivateDetails)
    if err != nil {
        return fmt.Errorf("failed to unmarshal JSON: %v", err)
    }

    // 检查资产是否存在
    exists, err := s.AssetExists(ctx, assetPrivateDetails.ID)
    if err != nil {
        return fmt.Errorf("error checking if asset exists: %v", err)
    }
    if !exists {
        return fmt.Errorf("asset %s does not exist", assetPrivateDetails.ID)
    }

    // 将私有数据存储到私有数据集合
    err = ctx.GetStub().PutPrivateData(collection, assetPrivateDetails.ID, assetPrivateDetailsJSON)
    if err != nil {
        return fmt.Errorf("failed to put private data: %v", err)
    }

    return nil
}

// 读取私有数据
func (s *SmartContract) ReadAssetPrivateData(ctx contractapi.TransactionContextInterface, collection string, assetID string) (*AssetPrivateDetails, error) {
    privateDataJSON, err := ctx.GetStub().GetPrivateData(collection, assetID)
    if err != nil {
        return nil, fmt.Errorf("failed to read private data: %v", err)
    }
    if privateDataJSON == nil {
        return nil, fmt.Errorf("private data for asset %s does not exist", assetID)
    }

    var assetPrivateDetails AssetPrivateDetails
    err = json.Unmarshal(privateDataJSON, &assetPrivateDetails)
    if err != nil {
        return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
    }

    return &assetPrivateDetails, nil
}

// 验证私有数据哈希
func (s *SmartContract) VerifyAssetPrivateData(ctx contractapi.TransactionContextInterface, collection string, assetID string, expectedHash string) (bool, error) {
    privateDataHash, err := ctx.GetStub().GetPrivateDataHash(collection, assetID)
    if err != nil {
        return false, fmt.Errorf("failed to get private data hash: %v", err)
    }

    if privateDataHash == nil {
        return false, fmt.Errorf("private data hash for asset %s does not exist", assetID)
    }

    return fmt.Sprintf("%x", privateDataHash) == expectedHash, nil
}
```

## 监控和运维

### 8.1 事件监听

```go
// client/event_listener.go
package client

import (
    "context"
    "fmt"
    "log"

    "github.com/hyperledger/fabric-sdk-go/pkg/client/event"
    "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

type EventListener struct {
    eventClient *event.Client
}

func NewEventListener(channelProvider context.ChannelProvider) (*EventListener, error) {
    eventClient, err := event.New(channelProvider)
    if err != nil {
        return nil, err
    }

    return &EventListener{
        eventClient: eventClient,
    }, nil
}

// 监听链码事件
func (el *EventListener) ListenChaincodeEvents(chaincodeName string, eventFilter string) error {
    reg, notifier, err := el.eventClient.RegisterChaincodeEvent(chaincodeName, eventFilter)
    if err != nil {
        return err
    }
    defer el.eventClient.Unregister(reg)

    for {
        select {
        case ccEvent := <-notifier:
            log.Printf("收到链码事件: %s - %s", ccEvent.EventName, string(ccEvent.Payload))
        case <-context.Background().Done():
            return nil
        }
    }
}

// 监听区块事件
func (el *EventListener) ListenBlockEvents() error {
    reg, notifier, err := el.eventClient.RegisterBlockEvent()
    if err != nil {
        return err
    }
    defer el.eventClient.Unregister(reg)

    for {
        select {
        case blockEvent := <-notifier:
            log.Printf("收到区块事件: 区块号 %d", blockEvent.Block.Header.Number)
            // 处理区块中的交易
            for _, txData := range blockEvent.Block.Data.Data {
                log.Printf("交易数据: %x", txData)
            }
        case <-context.Background().Done():
            return nil
        }
    }
}

// 监听交易事件
func (el *EventListener) ListenTransactionEvents(txID string) error {
    reg, notifier, err := el.eventClient.RegisterTxStatusEvent(txID)
    if err != nil {
        return err
    }
    defer el.eventClient.Unregister(reg)

    select {
    case txEvent := <-notifier:
        log.Printf("交易 %s 状态: %s", txEvent.TxID, txEvent.TxValidationCode)
        return nil
    case <-context.Background().Done():
        return fmt.Errorf("context cancelled")
    }
}
```

## 实际应用

### 9.1 完整应用示例

```go
// main.go
package main

import (
    "encoding/json"
    "fmt"
    "log"

    "your-project/client"
    "your-project/config"
)

func main() {
    // 创建Fabric客户端
    fabricClient, err := client.NewFabricClient(
        "network-config.yaml",
        "wallet",
        "appUser",
        "mychannel",
        "asset-transfer-basic",
    )
    if err != nil {
        log.Fatalf("创建Fabric客户端失败: %v", err)
    }
    defer fabricClient.Close()

    // 创建资产
    err = fabricClient.CreateAsset("asset1", "blue", 5, "Tom", 300)
    if err != nil {
        log.Printf("创建资产失败: %v", err)
    }

    // 读取资产
    result, err := fabricClient.ReadAsset("asset1")
    if err != nil {
        log.Printf("读取资产失败: %v", err)
    } else {
        fmt.Printf("资产信息: %s\n", string(result))
    }

    // 转移资产
    err = fabricClient.TransferAsset("asset1", "Jerry")
    if err != nil {
        log.Printf("转移资产失败: %v", err)
    }

    // 获取所有资产
    allAssets, err := fabricClient.GetAllAssets()
    if err != nil {
        log.Printf("获取所有资产失败: %v", err)
    } else {
        var assets []map[string]interface{}
        json.Unmarshal(allAssets, &assets)
        fmt.Printf("所有资产: %+v\n", assets)
    }
}
```
