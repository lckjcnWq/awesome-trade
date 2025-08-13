# Flow Blockchain 详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [环境准备](#环境准备)
3. [账户系统](#账户系统)
4. [Cadence智能合约](#cadence智能合约)
5. [NFT标准](#nft标准)
6. [交易处理](#交易处理)
7. [资源管理](#资源管理)
8. [生态集成](#生态集成)
9. [实际应用](#实际应用)

## 基础概念

### 1.1 Flow Blockchain 简介

Flow 是专为NFT和游戏设计的区块链，采用独特的多角色架构和资源导向编程模型，提供高性能、低费用和开发者友好的环境。

```bash
# 安装Flow相关依赖
go get github.com/onflow/flow-go-sdk
go get github.com/onflow/cadence/runtime
go get github.com/shopspring/decimal
```

### 1.2 核心概念

```go
// 主要包导入
import (
    "context"
    "encoding/hex"
    "fmt"
    
    "github.com/onflow/flow-go-sdk"
    "github.com/onflow/flow-go-sdk/access/grpc"
    "github.com/onflow/flow-go-sdk/crypto"
    "github.com/onflow/flow-go-sdk/templates"
    "github.com/onflow/cadence"
    "github.com/shopspring/decimal"
)

// Flow 网络配置
var (
    // Mainnet
    MainnetHost = "access.mainnet.nodes.onflow.org:9000"
    
    // Testnet
    TestnetHost = "access.devnet.nodes.onflow.org:9000"
    
    // Emulator
    EmulatorHost = "127.0.0.1:3569"
)

// Flow 网络信息
type NetworkInfo struct {
    NetworkID   string
    NetworkName string
    Host        string
    ChainID     flow.ChainID
    ExplorerURL string
    Currency    string
    Decimals    int
}

var (
    MainnetInfo = NetworkInfo{
        NetworkID:   "mainnet",
        NetworkName: "Flow Mainnet",
        Host:        MainnetHost,
        ChainID:     flow.Mainnet,
        ExplorerURL: "https://flowscan.org",
        Currency:    "FLOW",
        Decimals:    8,
    }
    
    TestnetInfo = NetworkInfo{
        NetworkID:   "testnet",
        NetworkName: "Flow Testnet",
        Host:        TestnetHost,
        ChainID:     flow.Testnet,
        ExplorerURL: "https://testnet.flowscan.org",
        Currency:    "FLOW",
        Decimals:    8,
    }
)

// Flow 核心合约地址
var (
    // 系统合约
    FlowTokenAddress     = "0x1654653399040a61" // Mainnet
    FungibleTokenAddress = "0xf233dcee88fe0abe" // Mainnet
    NonFungibleTokenAddress = "0x1d7e57aa55817448" // Mainnet
    
    // 知名NFT合约
    TopShotAddress       = "0x0b2a3299cc857e29" // NBA Top Shot
    CryptoKittiesAddress = "0xd796ff17107bbff6" // CryptoKitties
    FlowtyAddress        = "0x5c57f79c6694797f" // Flowty Marketplace
)

// 账户信息
type Account struct {
    Address     flow.Address
    Balance     uint64
    Keys        []AccountKey
    Contracts   map[string][]byte
    Storage     AccountStorage
}

type AccountKey struct {
    Index          int
    PublicKey      crypto.PublicKey
    SigAlgo        crypto.SignatureAlgorithm
    HashAlgo       crypto.HashAlgorithm
    Weight         int
    SequenceNumber uint64
    Revoked        bool
}

type AccountStorage struct {
    Used      uint64
    Capacity  uint64
    Available uint64
}

// NFT资源
type NFTResource struct {
    ID          uint64
    UUID        uint64
    Type        string
    Owner       flow.Address
    Metadata    map[string]cadence.Value
    Royalties   []Royalty
    Collection  string
}

type Royalty struct {
    Receiver flow.Address
    Cut      decimal.Decimal
}

// 集合信息
type Collection struct {
    Address     flow.Address
    Name        string
    Description string
    ExternalURL string
    SquareImage string
    BannerImage string
    SocialMedia map[string]string
    Metadata    map[string]cadence.Value
}

// 交易信息
type Transaction struct {
    ID              flow.Identifier
    Script          []byte
    Arguments       []cadence.Value
    ReferenceBlockID flow.Identifier
    GasLimit        uint64
    ProposalKey     flow.ProposalKey
    Payer           flow.Address
    Authorizers     []flow.Address
    PayloadSignatures []flow.TransactionSignature
    EnvelopeSignatures []flow.TransactionSignature
    Status          flow.TransactionStatus
    StatusCode      uint
    ErrorMessage    string
    Events          []flow.Event
}

// 事件信息
type Event struct {
    Type             string
    TransactionID    flow.Identifier
    TransactionIndex int
    EventIndex       int
    Value            cadence.Event
}

// 市场信息
type MarketListing struct {
    ID          uint64
    NFTType     string
    NFTID       uint64
    Seller      flow.Address
    Price       uint64
    Currency    string
    Status      string
    CreatedAt   uint64
    UpdatedAt   uint64
}

// 拍卖信息
type Auction struct {
    ID              uint64
    NFTType         string
    NFTID           uint64
    Seller          flow.Address
    StartPrice      uint64
    ReservePrice    uint64
    CurrentBid      uint64
    CurrentBidder   flow.Address
    Currency        string
    StartTime       uint64
    EndTime         uint64
    Status          string
    BidHistory      []Bid
}

type Bid struct {
    Bidder    flow.Address
    Amount    uint64
    Timestamp uint64
}

// 项目信息
type Project struct {
    Name            string
    Description     string
    Website         string
    Twitter         string
    Discord         string
    ContractAddress flow.Address
    TotalSupply     uint64
    FloorPrice      uint64
    Volume24h       uint64
    Owners          uint64
}
```

## 环境准备

### 2.1 Flow客户端设置

```go
// client/flow_client.go
package client

import (
    "context"
    "fmt"
    
    "github.com/onflow/flow-go-sdk"
    "github.com/onflow/flow-go-sdk/access/grpc"
    "github.com/onflow/flow-go-sdk/crypto"
    "github.com/onflow/cadence"
)

type FlowClient struct {
    client      *grpc.Client
    networkInfo *NetworkInfo
}

func NewFlowClient(networkInfo *NetworkInfo) (*FlowClient, error) {
    // 连接到Flow节点
    client, err := grpc.NewClient(networkInfo.Host)
    if err != nil {
        return nil, err
    }
    
    return &FlowClient{
        client:      client,
        networkInfo: networkInfo,
    }, nil
}

// 关闭连接
func (c *FlowClient) Close() error {
    return c.client.Close()
}

// 获取网络信息
func (c *FlowClient) GetNetworkInfo() *NetworkInfo {
    return c.networkInfo
}

// 获取账户信息
func (c *FlowClient) GetAccount(address flow.Address) (*Account, error) {
    account, err := c.client.GetAccount(context.Background(), address)
    if err != nil {
        return nil, err
    }
    
    // 转换账户密钥
    var keys []AccountKey
    for i, key := range account.Keys {
        accountKey := AccountKey{
            Index:          i,
            PublicKey:      key.PublicKey,
            SigAlgo:        key.SigAlgo,
            HashAlgo:       key.HashAlgo,
            Weight:         key.Weight,
            SequenceNumber: key.SequenceNumber,
            Revoked:        key.Revoked,
        }
        keys = append(keys, accountKey)
    }
    
    return &Account{
        Address:   account.Address,
        Balance:   account.Balance,
        Keys:      keys,
        Contracts: account.Contracts,
        Storage: AccountStorage{
            Used:      account.StorageUsed,
            Capacity:  account.StorageCapacity,
            Available: account.StorageCapacity - account.StorageUsed,
        },
    }, nil
}

// 获取最新区块
func (c *FlowClient) GetLatestBlock() (*flow.Block, error) {
    block, err := c.client.GetLatestBlock(context.Background(), true)
    if err != nil {
        return nil, err
    }
    
    return block, nil
}

// 执行脚本
func (c *FlowClient) ExecuteScript(script []byte, arguments []cadence.Value) (cadence.Value, error) {
    result, err := c.client.ExecuteScriptAtLatestBlock(context.Background(), script, arguments)
    if err != nil {
        return nil, err
    }
    
    return result, nil
}

// 发送交易
func (c *FlowClient) SendTransaction(tx *flow.Transaction) error {
    err := c.client.SendTransaction(context.Background(), *tx)
    if err != nil {
        return err
    }
    
    return nil
}

// 获取交易结果
func (c *FlowClient) GetTransactionResult(txID flow.Identifier) (*flow.TransactionResult, error) {
    result, err := c.client.GetTransactionResult(context.Background(), txID)
    if err != nil {
        return nil, err
    }
    
    return result, nil
}

// 等待交易确认
func (c *FlowClient) WaitForTransaction(txID flow.Identifier) (*flow.TransactionResult, error) {
    for {
        result, err := c.GetTransactionResult(txID)
        if err != nil {
            return nil, err
        }
        
        if result.Status == flow.TransactionStatusSealed {
            return result, nil
        }
        
        if result.Status == flow.TransactionStatusExpired {
            return nil, fmt.Errorf("交易已过期")
        }
        
        // 等待1秒后重试
        time.Sleep(1 * time.Second)
    }
}

// 获取事件
func (c *FlowClient) GetEvents(eventType string, startHeight, endHeight uint64) ([]flow.BlockEvents, error) {
    events, err := c.client.GetEventsForHeightRange(
        context.Background(),
        eventType,
        startHeight,
        endHeight,
    )
    if err != nil {
        return nil, err
    }
    
    return events, nil
}

// 获取集合信息
func (c *FlowClient) GetCollection(address flow.Address) (*Collection, error) {
    // 执行脚本获取集合信息
    script := []byte(`
        import NonFungibleToken from 0x1d7e57aa55817448
        
        pub fun main(address: Address): {String: String}? {
            let account = getAccount(address)
            // 这里需要根据具体的NFT合约实现
            return nil
        }
    `)
    
    args := []cadence.Value{cadence.NewAddress(address)}
    
    result, err := c.ExecuteScript(script, args)
    if err != nil {
        return nil, err
    }
    
    // 解析结果
    collection := &Collection{
        Address: address,
        Name:    "Unknown Collection",
    }
    
    return collection, nil
}

// 获取NFT信息
func (c *FlowClient) GetNFT(ownerAddress flow.Address, nftID uint64) (*NFTResource, error) {
    // 执行脚本获取NFT信息
    script := []byte(`
        import NonFungibleToken from 0x1d7e57aa55817448
        
        pub fun main(address: Address, id: UInt64): NFT? {
            let account = getAccount(address)
            // 这里需要根据具体的NFT合约实现
            return nil
        }
    `)
    
    args := []cadence.Value{
        cadence.NewAddress(ownerAddress),
        cadence.NewUInt64(nftID),
    }
    
    result, err := c.ExecuteScript(script, args)
    if err != nil {
        return nil, err
    }
    
    // 解析结果
    nft := &NFTResource{
        ID:       nftID,
        Owner:    ownerAddress,
        Metadata: make(map[string]cadence.Value),
    }
    
    return nft, nil
}

// 获取账户NFT列表
func (c *FlowClient) GetAccountNFTs(address flow.Address) ([]*NFTResource, error) {
    // 执行脚本获取账户所有NFT
    script := []byte(`
        import NonFungibleToken from 0x1d7e57aa55817448
        
        pub fun main(address: Address): [UInt64] {
            let account = getAccount(address)
            // 这里需要根据具体的NFT合约实现
            return []
        }
    `)
    
    args := []cadence.Value{cadence.NewAddress(address)}
    
    result, err := c.ExecuteScript(script, args)
    if err != nil {
        return nil, err
    }
    
    // 解析结果
    var nfts []*NFTResource
    
    if arrayValue, ok := result.(cadence.Array); ok {
        for _, value := range arrayValue.Values {
            if idValue, ok := value.(cadence.UInt64); ok {
                nft := &NFTResource{
                    ID:    uint64(idValue),
                    Owner: address,
                }
                nfts = append(nfts, nft)
            }
        }
    }
    
    return nfts, nil
}
```

## Cadence智能合约

### 3.1 NFT合约服务

```go
// services/nft_service.go
package services

import (
    "context"
    "fmt"
    
    "github.com/onflow/flow-go-sdk"
    "github.com/onflow/flow-go-sdk/crypto"
    "github.com/onflow/flow-go-sdk/templates"
    "github.com/onflow/cadence"
)

type FlowNFTService struct {
    client     *FlowClient
    privateKey crypto.PrivateKey
    signer     crypto.Signer
}

func NewFlowNFTService(client *FlowClient, privateKey crypto.PrivateKey) *FlowNFTService {
    signer, _ := crypto.NewInMemorySigner(privateKey, crypto.SHA3_256)
    
    return &FlowNFTService{
        client:     client,
        privateKey: privateKey,
        signer:     signer,
    }
}

// 部署NFT合约
func (s *FlowNFTService) DeployNFTContract(
    payerAddress flow.Address,
    contractName string,
    contractCode []byte,
) (*flow.TransactionResult, error) {
    // 创建部署合约交易
    tx := flow.NewTransaction().
        SetScript(templates.UpdateAccountContractTemplate).
        SetGasLimit(1000).
        SetProposalKey(payerAddress, 0, 0).
        SetPayer(payerAddress).
        AddAuthorizer(payerAddress)
    
    // 添加参数
    err := tx.AddArgument(cadence.String(contractName))
    if err != nil {
        return nil, err
    }
    
    err = tx.AddArgument(cadence.NewBytes(contractCode))
    if err != nil {
        return nil, err
    }
    
    // 设置参考区块
    latestBlock, err := s.client.GetLatestBlock()
    if err != nil {
        return nil, err
    }
    tx.SetReferenceBlockID(latestBlock.ID)
    
    // 签名交易
    err = tx.SignEnvelope(payerAddress, 0, s.signer)
    if err != nil {
        return nil, err
    }
    
    // 发送交易
    err = s.client.SendTransaction(tx)
    if err != nil {
        return nil, err
    }
    
    // 等待交易确认
    return s.client.WaitForTransaction(tx.ID())
}

// 铸造NFT
func (s *FlowNFTService) MintNFT(
    minterAddress flow.Address,
    recipientAddress flow.Address,
    metadata map[string]string,
) (*flow.TransactionResult, error) {
    // NFT铸造脚本
    script := []byte(`
        import NonFungibleToken from 0x1d7e57aa55817448
        import ExampleNFT from 0xf8d6e0586b0a20c7
        
        transaction(recipient: Address, metadata: {String: String}) {
            let minter: &ExampleNFT.NFTMinter
            
            prepare(signer: AuthAccount) {
                self.minter = signer.borrow<&ExampleNFT.NFTMinter>(from: ExampleNFT.MinterStoragePath)
                    ?? panic("Could not borrow a reference to the NFT minter")
            }
            
            execute {
                let recipient = getAccount(recipient)
                let receiver = recipient
                    .getCapability(ExampleNFT.CollectionPublicPath)!
                    .borrow<&{NonFungibleToken.CollectionPublic}>()
                    ?? panic("Could not get receiver reference to the NFT Collection")
                
                self.minter.mintNFT(recipient: receiver, metadata: metadata)
            }
        }
    `)
    
    // 创建交易
    tx := flow.NewTransaction().
        SetScript(script).
        SetGasLimit(1000).
        SetProposalKey(minterAddress, 0, 0).
        SetPayer(minterAddress).
        AddAuthorizer(minterAddress)
    
    // 添加参数
    err := tx.AddArgument(cadence.NewAddress(recipientAddress))
    if err != nil {
        return nil, err
    }
    
    // 转换metadata为Cadence字典
    metadataDict := make([]cadence.KeyValuePair, 0, len(metadata))
    for key, value := range metadata {
        pair := cadence.KeyValuePair{
            Key:   cadence.String(key),
            Value: cadence.String(value),
        }
        metadataDict = append(metadataDict, pair)
    }
    
    err = tx.AddArgument(cadence.NewDictionary(metadataDict))
    if err != nil {
        return nil, err
    }
    
    // 设置参考区块
    latestBlock, err := s.client.GetLatestBlock()
    if err != nil {
        return nil, err
    }
    tx.SetReferenceBlockID(latestBlock.ID)
    
    // 签名交易
    err = tx.SignEnvelope(minterAddress, 0, s.signer)
    if err != nil {
        return nil, err
    }
    
    // 发送交易
    err = s.client.SendTransaction(tx)
    if err != nil {
        return nil, err
    }
    
    // 等待交易确认
    return s.client.WaitForTransaction(tx.ID())
}

// 转移NFT
func (s *FlowNFTService) TransferNFT(
    fromAddress flow.Address,
    toAddress flow.Address,
    nftID uint64,
) (*flow.TransactionResult, error) {
    // NFT转移脚本
    script := []byte(`
        import NonFungibleToken from 0x1d7e57aa55817448
        import ExampleNFT from 0xf8d6e0586b0a20c7
        
        transaction(recipient: Address, withdrawID: UInt64) {
            prepare(signer: AuthAccount) {
                let recipient = getAccount(recipient)
                let receiverRef = recipient
                    .getCapability(ExampleNFT.CollectionPublicPath)!
                    .borrow<&{NonFungibleToken.CollectionPublic}>()
                    ?? panic("Could not borrow a reference to the receiver's collection")
                
                let collectionRef = signer
                    .borrow<&ExampleNFT.Collection>(from: ExampleNFT.CollectionStoragePath)!
                
                let nft <- collectionRef.withdraw(withdrawID: withdrawID)
                receiverRef.deposit(token: <-nft)
            }
        }
    `)
    
    // 创建交易
    tx := flow.NewTransaction().
        SetScript(script).
        SetGasLimit(1000).
        SetProposalKey(fromAddress, 0, 0).
        SetPayer(fromAddress).
        AddAuthorizer(fromAddress)
    
    // 添加参数
    err := tx.AddArgument(cadence.NewAddress(toAddress))
    if err != nil {
        return nil, err
    }
    
    err = tx.AddArgument(cadence.NewUInt64(nftID))
    if err != nil {
        return nil, err
    }
    
    // 设置参考区块
    latestBlock, err := s.client.GetLatestBlock()
    if err != nil {
        return nil, err
    }
    tx.SetReferenceBlockID(latestBlock.ID)
    
    // 签名交易
    err = tx.SignEnvelope(fromAddress, 0, s.signer)
    if err != nil {
        return nil, err
    }
    
    // 发送交易
    err = s.client.SendTransaction(tx)
    if err != nil {
        return nil, err
    }
    
    // 等待交易确认
    return s.client.WaitForTransaction(tx.ID())
}

// 设置NFT销售
func (s *FlowNFTService) ListNFTForSale(
    sellerAddress flow.Address,
    nftID uint64,
    price uint64,
) (*flow.TransactionResult, error) {
    // NFT销售脚本
    script := []byte(`
        import FungibleToken from 0xf233dcee88fe0abe
        import NonFungibleToken from 0x1d7e57aa55817448
        import FlowToken from 0x1654653399040a61
        import ExampleNFT from 0xf8d6e0586b0a20c7
        import NFTStorefront from 0x4eb8a10cb9f87357
        
        transaction(saleItemID: UInt64, saleItemPrice: UFix64) {
            let flowReceiver: Capability<&FlowToken.Vault{FungibleToken.Receiver}>
            let exampleNFTProvider: Capability<&ExampleNFT.Collection{NonFungibleToken.Provider, NonFungibleToken.CollectionPublic}>
            let storefront: &NFTStorefront.Storefront
            
            prepare(signer: AuthAccount) {
                self.flowReceiver = signer.getCapability<&FlowToken.Vault{FungibleToken.Receiver}>(/public/flowTokenReceiver)!
                assert(self.flowReceiver.borrow() != nil, message: "Missing or mis-typed FlowToken receiver")
                
                self.exampleNFTProvider = signer.getCapability<&ExampleNFT.Collection{NonFungibleToken.Provider, NonFungibleToken.CollectionPublic}>(ExampleNFT.CollectionPrivatePath)!
                assert(self.exampleNFTProvider.borrow() != nil, message: "Missing or mis-typed ExampleNFT.Collection provider")
                
                self.storefront = signer.borrow<&NFTStorefront.Storefront>(from: NFTStorefront.StorefrontStoragePath)
                    ?? panic("Missing or mis-typed NFTStorefront Storefront")
            }
            
            execute {
                let saleCut = NFTStorefront.SaleCut(
                    receiver: self.flowReceiver,
                    amount: saleItemPrice
                )
                self.storefront.createListing(
                    nftProviderCapability: self.exampleNFTProvider,
                    nftType: Type<@ExampleNFT.NFT>(),
                    nftID: saleItemID,
                    salePaymentVaultType: Type<@FlowToken.Vault>(),
                    saleCuts: [saleCut]
                )
            }
        }
    `)
    
    // 创建交易
    tx := flow.NewTransaction().
        SetScript(script).
        SetGasLimit(1000).
        SetProposalKey(sellerAddress, 0, 0).
        SetPayer(sellerAddress).
        AddAuthorizer(sellerAddress)
    
    // 添加参数
    err := tx.AddArgument(cadence.NewUInt64(nftID))
    if err != nil {
        return nil, err
    }
    
    // 将价格转换为UFix64 (Flow的固定点数类型)
    priceValue, err := cadence.NewUFix64(fmt.Sprintf("%.8f", float64(price)/100000000))
    if err != nil {
        return nil, err
    }
    
    err = tx.AddArgument(priceValue)
    if err != nil {
        return nil, err
    }
    
    // 设置参考区块
    latestBlock, err := s.client.GetLatestBlock()
    if err != nil {
        return nil, err
    }
    tx.SetReferenceBlockID(latestBlock.ID)
    
    // 签名交易
    err = tx.SignEnvelope(sellerAddress, 0, s.signer)
    if err != nil {
        return nil, err
    }
    
    // 发送交易
    err = s.client.SendTransaction(tx)
    if err != nil {
        return nil, err
    }
    
    // 等待交易确认
    return s.client.WaitForTransaction(tx.ID())
}

// 购买NFT
func (s *FlowNFTService) PurchaseNFT(
    buyerAddress flow.Address,
    listingResourceID uint64,
    storefrontAddress flow.Address,
) (*flow.TransactionResult, error) {
    // NFT购买脚本
    script := []byte(`
        import FungibleToken from 0xf233dcee88fe0abe
        import NonFungibleToken from 0x1d7e57aa55817448
        import FlowToken from 0x1654653399040a61
        import ExampleNFT from 0xf8d6e0586b0a20c7
        import NFTStorefront from 0x4eb8a10cb9f87357
        
        transaction(listingResourceID: UInt64, storefrontAddress: Address) {
            let paymentVault: @FungibleToken.Vault
            let exampleNFTCollection: &ExampleNFT.Collection{NonFungibleToken.Receiver}
            let storefront: &NFTStorefront.Storefront{NFTStorefront.StorefrontPublic}
            let listing: &NFTStorefront.Listing{NFTStorefront.ListingPublic}
            
            prepare(signer: AuthAccount) {
                self.storefront = getAccount(storefrontAddress)
                    .getCapability<&NFTStorefront.Storefront{NFTStorefront.StorefrontPublic}>(
                        NFTStorefront.StorefrontPublicPath
                    )!
                    .borrow()
                    ?? panic("Could not borrow Storefront from provided address")
                
                self.listing = self.storefront.borrowListing(listingResourceID: listingResourceID)
                    ?? panic("No Offer with that ID in Storefront")
                let price = self.listing.getDetails().salePrice
                
                let mainFlowVault = signer.borrow<&FlowToken.Vault>(from: /storage/flowTokenVault)
                    ?? panic("Cannot borrow FlowToken vault from signer storage")
                self.paymentVault <- mainFlowVault.withdraw(amount: price)
                
                self.exampleNFTCollection = signer.borrow<&ExampleNFT.Collection{NonFungibleToken.Receiver}>(
                        from: ExampleNFT.CollectionStoragePath
                    ) ?? panic("Cannot borrow NFT collection receiver from account")
            }
            
            execute {
                let item <- self.listing.purchase(
                    payment: <-self.paymentVault
                )
                
                self.exampleNFTCollection.deposit(token: <-item)
            }
        }
    `)
    
    // 创建交易
    tx := flow.NewTransaction().
        SetScript(script).
        SetGasLimit(1000).
        SetProposalKey(buyerAddress, 0, 0).
        SetPayer(buyerAddress).
        AddAuthorizer(buyerAddress)
    
    // 添加参数
    err := tx.AddArgument(cadence.NewUInt64(listingResourceID))
    if err != nil {
        return nil, err
    }
    
    err = tx.AddArgument(cadence.NewAddress(storefrontAddress))
    if err != nil {
        return nil, err
    }
    
    // 设置参考区块
    latestBlock, err := s.client.GetLatestBlock()
    if err != nil {
        return nil, err
    }
    tx.SetReferenceBlockID(latestBlock.ID)
    
    // 签名交易
    err = tx.SignEnvelope(buyerAddress, 0, s.signer)
    if err != nil {
        return nil, err
    }
    
    // 发送交易
    err = s.client.SendTransaction(tx)
    if err != nil {
        return nil, err
    }
    
    // 等待交易确认
    return s.client.WaitForTransaction(tx.ID())
}

// 批量铸造NFT
func (s *FlowNFTService) BatchMintNFTs(
    minterAddress flow.Address,
    recipients []flow.Address,
    metadataList []map[string]string,
) ([]*flow.TransactionResult, error) {
    if len(recipients) != len(metadataList) {
        return nil, fmt.Errorf("接收者数量与元数据数量不匹配")
    }
    
    var results []*flow.TransactionResult
    
    for i, recipient := range recipients {
        result, err := s.MintNFT(minterAddress, recipient, metadataList[i])
        if err != nil {
            return nil, err
        }
        
        results = append(results, result)
    }
    
    return results, nil
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
    
    "github.com/onflow/flow-go-sdk"
    "github.com/onflow/flow-go-sdk/crypto"
    "github.com/shopspring/decimal"
    
    "your-project/client"
    "your-project/services"
)

func main() {
    // 创建Flow客户端 (使用测试网)
    flowClient, err := client.NewFlowClient(&client.TestnetInfo)
    if err != nil {
        log.Fatal("创建Flow客户端失败:", err)
    }
    defer flowClient.Close()
    
    // 加载私钥
    privateKey, err := crypto.DecodePrivateKeyHex(crypto.ECDSA_P256, "your_private_key_here")
    if err != nil {
        log.Fatal("加载私钥失败:", err)
    }
    
    // 获取用户地址
    publicKey := privateKey.PublicKey()
    userAddress := flow.HexToAddress("0x01cf0e2f2f715450") // 示例地址
    
    // 创建服务
    nftService := services.NewFlowNFTService(flowClient, privateKey)
    
    // 1. 获取网络信息
    config := flowClient.GetNetworkInfo()
    fmt.Printf("=== Flow网络信息 ===\n")
    fmt.Printf("网络名称: %s\n", config.NetworkName)
    fmt.Printf("网络ID: %s\n", config.NetworkID)
    fmt.Printf("节点地址: %s\n", config.Host)
    fmt.Printf("链ID: %s\n", config.ChainID.String())
    fmt.Printf("浏览器: %s\n", config.ExplorerURL)
    fmt.Printf("原生代币: %s\n", config.Currency)
    
    // 2. 获取账户信息
    fmt.Printf("\n=== 账户信息 ===\n")
    
    account, err := flowClient.GetAccount(userAddress)
    if err != nil {
        log.Fatal("获取账户信息失败:", err)
    }
    
    fmt.Printf("地址: %s\n", account.Address.Hex())
    fmt.Printf("FLOW余额: %s (%.8f FLOW)\n", 
        fmt.Sprintf("%d", account.Balance),
        decimal.NewFromInt(int64(account.Balance)).Div(decimal.NewFromInt(100000000)).InexactFloat64())
    fmt.Printf("存储使用: %d bytes\n", account.Storage.Used)
    fmt.Printf("存储容量: %d bytes\n", account.Storage.Capacity)
    fmt.Printf("可用存储: %d bytes\n", account.Storage.Available)
    fmt.Printf("账户密钥数量: %d\n", len(account.Keys))
    fmt.Printf("部署合约数量: %d\n", len(account.Contracts))
    
    // 显示账户密钥信息
    for i, key := range account.Keys {
        fmt.Printf("  密钥 %d:\n", i)
        fmt.Printf("    权重: %d\n", key.Weight)
        fmt.Printf("    序列号: %d\n", key.SequenceNumber)
        fmt.Printf("    已撤销: %t\n", key.Revoked)
        fmt.Printf("    签名算法: %s\n", key.SigAlgo.String())
        fmt.Printf("    哈希算法: %s\n", key.HashAlgo.String())
    }
    
    // 3. 获取最新区块信息
    fmt.Printf("\n=== 最新区块信息 ===\n")
    
    latestBlock, err := flowClient.GetLatestBlock()
    if err != nil {
        log.Fatal("获取最新区块失败:", err)
    }
    
    fmt.Printf("区块高度: %d\n", latestBlock.Height)
    fmt.Printf("区块ID: %s\n", latestBlock.ID.Hex())
    fmt.Printf("区块时间: %s\n", latestBlock.Timestamp.String())
    fmt.Printf("父区块ID: %s\n", latestBlock.ParentID.Hex())
    fmt.Printf("集合保证数量: %d\n", len(latestBlock.CollectionGuarantees))
    
    // 4. 获取用户NFT
    fmt.Printf("\n=== 用户NFT资产 ===\n")
    
    nfts, err := flowClient.GetAccountNFTs(userAddress)
    if err != nil {
        log.Printf("获取用户NFT失败: %v", err)
    } else {
        fmt.Printf("拥有NFT数量: %d\n", len(nfts))
        
        for i, nft := range nfts {
            fmt.Printf("  NFT %d:\n", i+1)
            fmt.Printf("    ID: %d\n", nft.ID)
            fmt.Printf("    UUID: %d\n", nft.UUID)
            fmt.Printf("    类型: %s\n", nft.Type)
            fmt.Printf("    所有者: %s\n", nft.Owner.Hex())
            fmt.Printf("    集合: %s\n", nft.Collection)
            
            if len(nft.Royalties) > 0 {
                fmt.Printf("    版税信息:\n")
                for j, royalty := range nft.Royalties {
                    fmt.Printf("      版税 %d: %s -> %s%%\n", 
                        j+1, 
                        royalty.Receiver.Hex(), 
                        royalty.Cut.Mul(decimal.NewFromInt(100)).String())
                }
            }
        }
    }
    
    // 5. NFT铸造示例
    fmt.Printf("\n=== NFT铸造示例 ===\n")
    
    // 准备NFT元数据
    metadata := map[string]string{
        "name":        "Flow Epic Sword",
        "description": "A legendary sword forged on Flow blockchain",
        "image":       "https://example.com/sword.png",
        "rarity":      "Legendary",
        "attack":      "150",
        "defense":     "50",
        "element":     "Fire",
    }
    
    fmt.Printf("准备铸造NFT:\n")
    fmt.Printf("  名称: %s\n", metadata["name"])
    fmt.Printf("  描述: %s\n", metadata["description"])
    fmt.Printf("  稀有度: %s\n", metadata["rarity"])
    fmt.Printf("  攻击力: %s\n", metadata["attack"])
    fmt.Printf("  防御力: %s\n", metadata["defense"])
    fmt.Printf("  元素: %s\n", metadata["element"])
    
    // 执行铸造 (注释掉实际执行)
    // result, err := nftService.MintNFT(userAddress, userAddress, metadata)
    // if err != nil {
    //     log.Printf("铸造NFT失败: %v", err)
    // } else {
    //     fmt.Printf("NFT铸造成功:\n")
    //     fmt.Printf("  交易ID: %s\n", result.TransactionID.Hex())
    //     fmt.Printf("  状态: %s\n", result.Status.String())
    //     fmt.Printf("  状态码: %d\n", result.StatusCode)
    //     
    //     if len(result.Events) > 0 {
    //         fmt.Printf("  事件数量: %d\n", len(result.Events))
    //         for i, event := range result.Events {
    //             fmt.Printf("    事件 %d: %s\n", i+1, event.Type)
    //         }
    //     }
    // }
    
    // 6. 批量铸造示例
    fmt.Printf("\n=== 批量铸造示例 ===\n")
    
    recipients := []flow.Address{
        userAddress,
        flow.HexToAddress("0x179b6b1cb6755e31"), // 示例地址
        flow.HexToAddress("0xf3fcd2c1a78f5eee"), // 示例地址
    }
    
    metadataList := []map[string]string{
        {
            "name":   "Fire Sword #1",
            "rarity": "Common",
            "attack": "100",
        },
        {
            "name":   "Ice Shield #1",
            "rarity": "Rare",
            "defense": "120",
        },
        {
            "name":   "Lightning Staff #1",
            "rarity": "Epic",
            "magic": "200",
        },
    }
    
    fmt.Printf("准备批量铸造 %d 个NFT\n", len(recipients))
    
    for i, recipient := range recipients {
        fmt.Printf("  NFT %d -> %s: %s\n", 
            i+1, 
            recipient.Hex(), 
            metadataList[i]["name"])
    }
    
    // 执行批量铸造 (注释掉实际执行)
    // results, err := nftService.BatchMintNFTs(userAddress, recipients, metadataList)
    // if err != nil {
    //     log.Printf("批量铸造失败: %v", err)
    // } else {
    //     fmt.Printf("批量铸造成功，交易数量: %d\n", len(results))
    //     for i, result := range results {
    //         fmt.Printf("  交易 %d: %s (状态: %s)\n", 
    //             i+1, 
    //             result.TransactionID.Hex(), 
    //             result.Status.String())
    //     }
    // }
    
    // 7. NFT转移示例
    fmt.Printf("\n=== NFT转移示例 ===\n")
    
    if len(nfts) > 0 {
        nftToTransfer := nfts[0]
        recipientAddress := flow.HexToAddress("0x179b6b1cb6755e31")
        
        fmt.Printf("准备转移NFT:\n")
        fmt.Printf("  NFT ID: %d\n", nftToTransfer.ID)
        fmt.Printf("  从: %s\n", userAddress.Hex())
        fmt.Printf("  到: %s\n", recipientAddress.Hex())
        
        // 执行转移 (注释掉实际执行)
        // result, err := nftService.TransferNFT(userAddress, recipientAddress, nftToTransfer.ID)
        // if err != nil {
        //     log.Printf("转移NFT失败: %v", err)
        // } else {
        //     fmt.Printf("NFT转移成功: %s\n", result.TransactionID.Hex())
        // }
    } else {
        fmt.Printf("没有NFT可供转移\n")
    }
    
    // 8. NFT市场交易示例
    fmt.Printf("\n=== NFT市场交易示例 ===\n")
    
    if len(nfts) > 0 {
        nftToSell := nfts[0]
        salePrice := uint64(100000000) // 1.0 FLOW
        
        fmt.Printf("准备上架NFT:\n")
        fmt.Printf("  NFT ID: %d\n", nftToSell.ID)
        fmt.Printf("  售价: %.8f FLOW\n", 
            decimal.NewFromInt(int64(salePrice)).Div(decimal.NewFromInt(100000000)).InexactFloat64())
        
        // 执行上架 (注释掉实际执行)
        // result, err := nftService.ListNFTForSale(userAddress, nftToSell.ID, salePrice)
        // if err != nil {
        //     log.Printf("上架NFT失败: %v", err)
        // } else {
        //     fmt.Printf("NFT上架成功: %s\n", result.TransactionID.Hex())
        // }
    }
    
    // 9. Flow特性总结
    fmt.Printf("\n=== Flow特性总结 ===\n")
    
    fmt.Printf("Flow优势:\n")
    fmt.Printf("  - 资源导向编程模型\n")
    fmt.Printf("  - 多角色架构 (收集、共识、执行、验证)\n")
    fmt.Printf("  - 开发者友好的Cadence语言\n")
    fmt.Printf("  - 低费用和高性能\n")
    fmt.Printf("  - 内置升级机制\n")
    
    fmt.Printf("\nCadence语言特点:\n")
    fmt.Printf("  - 资源导向: 数字资产作为一等公民\n")
    fmt.Printf("  - 类型安全: 强类型系统\n")
    fmt.Printf("  - 能力安全: 基于能力的访问控制\n")
    fmt.Printf("  - 易于审计: 清晰的代码结构\n")
    
    fmt.Printf("\n生态项目:\n")
    fmt.Printf("  - NBA Top Shot: 体育收藏品\n")
    fmt.Printf("  - CryptoKitties: 数字宠物\n")
    fmt.Printf("  - Flowty: NFT市场\n")
    fmt.Printf("  - Blocto: 钱包服务\n")
    fmt.Printf("  - Dapper Labs: 游戏开发\n")
    
    // 10. 最佳实践建议
    fmt.Printf("\n=== 最佳实践建议 ===\n")
    
    fmt.Printf("使用Flow时请注意:\n")
    fmt.Printf("  1. 理解资源导向编程概念\n")
    fmt.Printf("  2. 合理设计NFT元数据结构\n")
    fmt.Printf("  3. 利用Cadence的类型安全特性\n")
    fmt.Printf("  4. 实现适当的访问控制\n")
    fmt.Printf("  5. 考虑合约升级策略\n")
    fmt.Printf("  6. 优化存储使用\n")
    fmt.Printf("  7. 测试网充分验证\n")
    fmt.Printf("  8. 关注社区最佳实践\n")
}
```

这个Flow Blockchain使用指南提供了完整的NFT专用区块链集成方案，涵盖了资源导向编程、Cadence智能合约、NFT标准、市场交易、批量操作等核心功能，是构建Flow生态NFT和游戏应用的重要参考文档。
