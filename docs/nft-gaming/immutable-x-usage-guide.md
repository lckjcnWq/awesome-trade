# Immutable X 详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [环境准备](#环境准备)
3. [账户系统](#账户系统)
4. [NFT铸造](#nft铸造)
5. [NFT交易](#nft交易)
6. [批量操作](#批量操作)
7. [游戏集成](#游戏集成)
8. [市场功能](#市场功能)
9. [实际应用](#实际应用)

## 基础概念

### 1.1 Immutable X 简介

Immutable X 是专为NFT设计的Layer2解决方案，基于StarkEx技术，提供零Gas费用的NFT铸造和交易，支持即时确认和大规模游戏应用。

```bash
# 安装Immutable X相关依赖
go get github.com/ethereum/go-ethereum
go get github.com/ethereum/go-ethereum/accounts/abi
go get github.com/shopspring/decimal
```

### 1.2 核心概念

```go
// 主要包导入
import (
    "context"
    "encoding/json"
    "math/big"
    "net/http"
    "strings"
    
    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/shopspring/decimal"
)

// Immutable X 网络配置
var (
    // Mainnet
    MainnetAPIURL = "https://api.x.immutable.com"
    MainnetLinkURL = "https://link.x.immutable.com"
    
    // Testnet (Ropsten)
    TestnetAPIURL = "https://api.ropsten.x.immutable.com"
    TestnetLinkURL = "https://link.ropsten.x.immutable.com"
    
    // Sandbox
    SandboxAPIURL = "https://api.sandbox.x.immutable.com"
    SandboxLinkURL = "https://link.sandbox.x.immutable.com"
)

// Immutable X 核心合约地址 (Mainnet)
var (
    // Core Contract
    CoreContractAddress = common.HexToAddress("0x5FDCCA53617f4d2b9134B29090C87D01058e27e9")
    
    // Registration Contract
    RegistrationContractAddress = common.HexToAddress("0x72a06bf2a1CE5e39cBA06c0CAb824960B587d64c")
    
    // IMX Token
    IMXTokenAddress = common.HexToAddress("0xF57e7e7C23978C3cAEC3C3548E3D615c346e79fF")
    
    // ETH Token (for trading)
    ETHTokenAddress = common.HexToAddress("0x0000000000000000000000000000000000000000")
)

// 网络信息
type NetworkConfig struct {
    NetworkID   string
    NetworkName string
    APIURL      string
    LinkURL     string
    ChainID     int
    ExplorerURL string
}

var (
    MainnetConfig = NetworkConfig{
        NetworkID:   "1",
        NetworkName: "Immutable X Mainnet",
        APIURL:      MainnetAPIURL,
        LinkURL:     MainnetLinkURL,
        ChainID:     1,
        ExplorerURL: "https://immutascan.io",
    }
    
    TestnetConfig = NetworkConfig{
        NetworkID:   "3",
        NetworkName: "Immutable X Testnet",
        APIURL:      TestnetAPIURL,
        LinkURL:     TestnetLinkURL,
        ChainID:     3,
        ExplorerURL: "https://ropsten.immutascan.io",
    }
)

// 用户信息
type User struct {
    EthAddress    string `json:"eth_address"`
    StarkKey      string `json:"stark_key"`
    StarkSignature string `json:"stark_signature"`
    Accounts      []string `json:"accounts"`
}

// NFT资产
type Asset struct {
    TokenAddress string            `json:"token_address"`
    TokenID      string            `json:"token_id"`
    ID           string            `json:"id"`
    User         string            `json:"user"`
    Status       string            `json:"status"`
    URI          string            `json:"uri"`
    Name         string            `json:"name"`
    Description  string            `json:"description"`
    ImageURL     string            `json:"image_url"`
    Metadata     map[string]interface{} `json:"metadata"`
    Collection   *Collection       `json:"collection"`
    CreatedAt    string            `json:"created_at"`
    UpdatedAt    string            `json:"updated_at"`
}

// NFT集合
type Collection struct {
    Address         string            `json:"address"`
    Name            string            `json:"name"`
    Description     string            `json:"description"`
    IconURL         string            `json:"icon_url"`
    CollectionImageURL string         `json:"collection_image_url"`
    ProjectID       int               `json:"project_id"`
    ProjectOwnerAddress string        `json:"project_owner_address"`
    MetadataAPIURL  string            `json:"metadata_api_url"`
    CreatedAt       string            `json:"created_at"`
    UpdatedAt       string            `json:"updated_at"`
}

// 订单信息
type Order struct {
    OrderID       int               `json:"order_id"`
    Status        string            `json:"status"`
    User          string            `json:"user"`
    SellToken     *Token            `json:"sell"`
    BuyToken      *Token            `json:"buy"`
    AmountSold    string            `json:"amount_sold"`
    AmountBought  string            `json:"amount_bought"`
    Timestamp     string            `json:"timestamp"`
    UpdatedTimestamp string         `json:"updated_timestamp"`
    ExpirationTimestamp string      `json:"expiration_timestamp"`
}

// 代币信息
type Token struct {
    Type         string            `json:"type"`
    Data         *TokenData        `json:"data"`
}

type TokenData struct {
    TokenAddress string            `json:"token_address"`
    TokenID      string            `json:"token_id"`
    Quantity     string            `json:"quantity"`
    Decimals     int               `json:"decimals"`
}

// 交易信息
type Trade struct {
    TransactionID int               `json:"transaction_id"`
    Status        string            `json:"status"`
    User          string            `json:"user"`
    Receiver      string            `json:"receiver"`
    TokenType     string            `json:"token_type"`
    TokenData     *TokenData        `json:"token_data"`
    Timestamp     string            `json:"timestamp"`
}

// 转账信息
type Transfer struct {
    TransactionID int               `json:"transaction_id"`
    Status        string            `json:"status"`
    User          string            `json:"user"`
    Receiver      string            `json:"receiver"`
    Token         *Token            `json:"token"`
    Timestamp     string            `json:"timestamp"`
}

// 铸造请求
type MintRequest struct {
    EthAddress   string            `json:"eth_address"`
    Tokens       []MintToken       `json:"tokens"`
}

type MintToken struct {
    ID           string            `json:"id"`
    Blueprint    string            `json:"blueprint"`
    RoyaltyRecipient string        `json:"royalty_recipient"`
    RoyaltyPercentage decimal.Decimal `json:"royalty_percentage"`
}

// 项目信息
type Project struct {
    ID                  int               `json:"id"`
    Name                string            `json:"name"`
    CompanyName         string            `json:"company_name"`
    ContactEmail        string            `json:"contact_email"`
    PublicAPIURL        string            `json:"public_api_url"`
    CollectionLimitExpiresAt string       `json:"collection_limit_expires_at"`
    CollectionMonthlyLimit int            `json:"collection_monthly_limit"`
    CollectionRemaining int               `json:"collection_remaining"`
    MintLimitExpiresAt  string            `json:"mint_limit_expires_at"`
    MintMonthlyLimit    int               `json:"mint_monthly_limit"`
    MintRemaining       int               `json:"mint_remaining"`
}

// 余额信息
type Balance struct {
    Token         *Token            `json:"token"`
    Balance       string            `json:"balance"`
    PreparingWithdrawal string      `json:"preparing_withdrawal"`
    Withdrawable  string            `json:"withdrawable"`
}
```

## 环境准备

### 2.1 Immutable X客户端设置

```go
// client/immutable_client.go
package client

import (
    "bytes"
    "context"
    "crypto/ecdsa"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    
    "github.com/ethereum/go-ethereum/crypto"
)

type ImmutableClient struct {
    config     *NetworkConfig
    httpClient *http.Client
    apiKey     string
    privateKey *ecdsa.PrivateKey
}

func NewImmutableClient(config *NetworkConfig, apiKey string, privateKey *ecdsa.PrivateKey) *ImmutableClient {
    return &ImmutableClient{
        config:     config,
        httpClient: &http.Client{},
        apiKey:     apiKey,
        privateKey: privateKey,
    }
}

// 获取网络配置
func (c *ImmutableClient) GetNetworkConfig() *NetworkConfig {
    return c.config
}

// 获取用户信息
func (c *ImmutableClient) GetUser(ethAddress string) (*User, error) {
    url := fmt.Sprintf("%s/v1/users/%s", c.config.APIURL, ethAddress)
    
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Content-Type", "application/json")
    if c.apiKey != "" {
        req.Header.Set("x-api-key", c.apiKey)
    }
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API请求失败: %d", resp.StatusCode)
    }
    
    var user User
    err = json.NewDecoder(resp.Body).Decode(&user)
    if err != nil {
        return nil, err
    }
    
    return &user, nil
}

// 获取用户资产
func (c *ImmutableClient) GetAssets(ethAddress string, cursor string, pageSize int) ([]*Asset, string, error) {
    url := fmt.Sprintf("%s/v1/assets", c.config.APIURL)
    
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, "", err
    }
    
    q := req.URL.Query()
    if ethAddress != "" {
        q.Add("user", ethAddress)
    }
    if cursor != "" {
        q.Add("cursor", cursor)
    }
    if pageSize > 0 {
        q.Add("page_size", fmt.Sprintf("%d", pageSize))
    }
    req.URL.RawQuery = q.Encode()
    
    req.Header.Set("Content-Type", "application/json")
    if c.apiKey != "" {
        req.Header.Set("x-api-key", c.apiKey)
    }
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, "", err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, "", fmt.Errorf("API请求失败: %d", resp.StatusCode)
    }
    
    var result struct {
        Result []*Asset `json:"result"`
        Cursor string   `json:"cursor"`
    }
    
    err = json.NewDecoder(resp.Body).Decode(&result)
    if err != nil {
        return nil, "", err
    }
    
    return result.Result, result.Cursor, nil
}

// 获取单个资产
func (c *ImmutableClient) GetAsset(tokenAddress, tokenID string) (*Asset, error) {
    url := fmt.Sprintf("%s/v1/assets/%s/%s", c.config.APIURL, tokenAddress, tokenID)
    
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Content-Type", "application/json")
    if c.apiKey != "" {
        req.Header.Set("x-api-key", c.apiKey)
    }
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API请求失败: %d", resp.StatusCode)
    }
    
    var asset Asset
    err = json.NewDecoder(resp.Body).Decode(&asset)
    if err != nil {
        return nil, err
    }
    
    return &asset, nil
}

// 获取集合信息
func (c *ImmutableClient) GetCollection(address string) (*Collection, error) {
    url := fmt.Sprintf("%s/v1/collections/%s", c.config.APIURL, address)
    
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Content-Type", "application/json")
    if c.apiKey != "" {
        req.Header.Set("x-api-key", c.apiKey)
    }
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API请求失败: %d", resp.StatusCode)
    }
    
    var collection Collection
    err = json.NewDecoder(resp.Body).Decode(&collection)
    if err != nil {
        return nil, err
    }
    
    return &collection, nil
}

// 获取订单
func (c *ImmutableClient) GetOrders(status, user string, cursor string, pageSize int) ([]*Order, string, error) {
    url := fmt.Sprintf("%s/v1/orders", c.config.APIURL)
    
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, "", err
    }
    
    q := req.URL.Query()
    if status != "" {
        q.Add("status", status)
    }
    if user != "" {
        q.Add("user", user)
    }
    if cursor != "" {
        q.Add("cursor", cursor)
    }
    if pageSize > 0 {
        q.Add("page_size", fmt.Sprintf("%d", pageSize))
    }
    req.URL.RawQuery = q.Encode()
    
    req.Header.Set("Content-Type", "application/json")
    if c.apiKey != "" {
        req.Header.Set("x-api-key", c.apiKey)
    }
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, "", err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, "", fmt.Errorf("API请求失败: %d", resp.StatusCode)
    }
    
    var result struct {
        Result []*Order `json:"result"`
        Cursor string   `json:"cursor"`
    }
    
    err = json.NewDecoder(resp.Body).Decode(&result)
    if err != nil {
        return nil, "", err
    }
    
    return result.Result, result.Cursor, nil
}

// 获取交易历史
func (c *ImmutableClient) GetTrades(user string, cursor string, pageSize int) ([]*Trade, string, error) {
    url := fmt.Sprintf("%s/v1/trades", c.config.APIURL)
    
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, "", err
    }
    
    q := req.URL.Query()
    if user != "" {
        q.Add("party_a_token_address", user)
    }
    if cursor != "" {
        q.Add("cursor", cursor)
    }
    if pageSize > 0 {
        q.Add("page_size", fmt.Sprintf("%d", pageSize))
    }
    req.URL.RawQuery = q.Encode()
    
    req.Header.Set("Content-Type", "application/json")
    if c.apiKey != "" {
        req.Header.Set("x-api-key", c.apiKey)
    }
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, "", err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, "", fmt.Errorf("API请求失败: %d", resp.StatusCode)
    }
    
    var result struct {
        Result []*Trade `json:"result"`
        Cursor string   `json:"cursor"`
    }
    
    err = json.NewDecoder(resp.Body).Decode(&result)
    if err != nil {
        return nil, "", err
    }
    
    return result.Result, result.Cursor, nil
}

// 获取用户余额
func (c *ImmutableClient) GetBalances(ethAddress string) ([]*Balance, error) {
    url := fmt.Sprintf("%s/v1/balances/%s", c.config.APIURL, ethAddress)
    
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Content-Type", "application/json")
    if c.apiKey != "" {
        req.Header.Set("x-api-key", c.apiKey)
    }
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API请求失败: %d", resp.StatusCode)
    }
    
    var result struct {
        Result []*Balance `json:"result"`
    }
    
    err = json.NewDecoder(resp.Body).Decode(&result)
    if err != nil {
        return nil, err
    }
    
    return result.Result, nil
}

// 创建项目
func (c *ImmutableClient) CreateProject(name, companyName, contactEmail string) (*Project, error) {
    url := fmt.Sprintf("%s/v1/projects", c.config.APIURL)
    
    projectData := map[string]interface{}{
        "name":          name,
        "company_name":  companyName,
        "contact_email": contactEmail,
    }
    
    jsonData, err := json.Marshal(projectData)
    if err != nil {
        return nil, err
    }
    
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Content-Type", "application/json")
    if c.apiKey != "" {
        req.Header.Set("x-api-key", c.apiKey)
    }
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusCreated {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("创建项目失败: %d, %s", resp.StatusCode, string(body))
    }
    
    var project Project
    err = json.NewDecoder(resp.Body).Decode(&project)
    if err != nil {
        return nil, err
    }
    
    return &project, nil
}

// 创建集合
func (c *ImmutableClient) CreateCollection(
    name, description, iconURL, metadataAPIURL string,
    projectID int,
) (*Collection, error) {
    url := fmt.Sprintf("%s/v1/collections", c.config.APIURL)
    
    collectionData := map[string]interface{}{
        "name":             name,
        "description":      description,
        "icon_url":         iconURL,
        "metadata_api_url": metadataAPIURL,
        "project_id":       projectID,
    }
    
    jsonData, err := json.Marshal(collectionData)
    if err != nil {
        return nil, err
    }
    
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Content-Type", "application/json")
    if c.apiKey != "" {
        req.Header.Set("x-api-key", c.apiKey)
    }
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusCreated {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("创建集合失败: %d, %s", resp.StatusCode, string(body))
    }
    
    var collection Collection
    err = json.NewDecoder(resp.Body).Decode(&collection)
    if err != nil {
        return nil, err
    }
    
    return &collection, nil
}
```

## NFT铸造

### 3.1 铸造服务

```go
// services/mint_service.go
package services

import (
    "bytes"
    "crypto/ecdsa"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/shopspring/decimal"
)

type ImmutableMintService struct {
    client     *ImmutableClient
    privateKey *ecdsa.PrivateKey
}

func NewImmutableMintService(client *ImmutableClient, privateKey *ecdsa.PrivateKey) *ImmutableMintService {
    return &ImmutableMintService{
        client:     client,
        privateKey: privateKey,
    }
}

// 批量铸造NFT
func (s *ImmutableMintService) MintNFTs(
    ethAddress string,
    tokens []MintToken,
) ([]*Asset, error) {
    url := fmt.Sprintf("%s/v1/mints", s.client.config.APIURL)
    
    mintRequest := MintRequest{
        EthAddress: ethAddress,
        Tokens:     tokens,
    }
    
    jsonData, err := json.Marshal(mintRequest)
    if err != nil {
        return nil, err
    }
    
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Content-Type", "application/json")
    if s.client.apiKey != "" {
        req.Header.Set("x-api-key", s.client.apiKey)
    }
    
    // 添加签名认证
    signature, err := s.signRequest(jsonData)
    if err != nil {
        return nil, err
    }
    req.Header.Set("x-signature", signature)
    
    resp, err := s.client.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusCreated {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("铸造NFT失败: %d, %s", resp.StatusCode, string(body))
    }
    
    var result struct {
        Result []*Asset `json:"result"`
    }
    
    err = json.NewDecoder(resp.Body).Decode(&result)
    if err != nil {
        return nil, err
    }
    
    return result.Result, nil
}

// 单个NFT铸造
func (s *ImmutableMintService) MintSingleNFT(
    ethAddress string,
    tokenID string,
    blueprint string,
    royaltyRecipient string,
    royaltyPercentage decimal.Decimal,
) (*Asset, error) {
    tokens := []MintToken{
        {
            ID:                tokenID,
            Blueprint:         blueprint,
            RoyaltyRecipient:  royaltyRecipient,
            RoyaltyPercentage: royaltyPercentage,
        },
    }
    
    assets, err := s.MintNFTs(ethAddress, tokens)
    if err != nil {
        return nil, err
    }
    
    if len(assets) == 0 {
        return nil, fmt.Errorf("铸造失败，未返回资产")
    }
    
    return assets[0], nil
}

// 游戏道具铸造
func (s *ImmutableMintService) MintGameItems(
    ethAddress string,
    items []GameItem,
) ([]*Asset, error) {
    var tokens []MintToken
    
    for _, item := range items {
        // 构建游戏道具的blueprint
        blueprint := s.buildGameItemBlueprint(item)
        
        token := MintToken{
            ID:                item.TokenID,
            Blueprint:         blueprint,
            RoyaltyRecipient:  item.RoyaltyRecipient,
            RoyaltyPercentage: item.RoyaltyPercentage,
        }
        
        tokens = append(tokens, token)
    }
    
    return s.MintNFTs(ethAddress, tokens)
}

// 批量铸造收藏品
func (s *ImmutableMintService) MintCollectibles(
    ethAddress string,
    collectionAddress string,
    collectibles []Collectible,
) ([]*Asset, error) {
    var tokens []MintToken
    
    for _, collectible := range collectibles {
        // 构建收藏品的blueprint
        blueprint := s.buildCollectibleBlueprint(collectible, collectionAddress)
        
        token := MintToken{
            ID:                collectible.TokenID,
            Blueprint:         blueprint,
            RoyaltyRecipient:  collectible.Creator,
            RoyaltyPercentage: collectible.RoyaltyPercentage,
        }
        
        tokens = append(tokens, token)
    }
    
    return s.MintNFTs(ethAddress, tokens)
}

// 获取铸造状态
func (s *ImmutableMintService) GetMintStatus(mintID string) (*MintStatus, error) {
    url := fmt.Sprintf("%s/v1/mints/%s", s.client.config.APIURL, mintID)
    
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Content-Type", "application/json")
    if s.client.apiKey != "" {
        req.Header.Set("x-api-key", s.client.apiKey)
    }
    
    resp, err := s.client.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("获取铸造状态失败: %d", resp.StatusCode)
    }
    
    var status MintStatus
    err = json.NewDecoder(resp.Body).Decode(&status)
    if err != nil {
        return nil, err
    }
    
    return &status, nil
}

// 计算铸造费用
func (s *ImmutableMintService) CalculateMintCost(tokenCount int) (*MintCost, error) {
    // Immutable X 铸造是免费的，但可能有其他成本
    return &MintCost{
        TokenCount:    tokenCount,
        GasFee:        decimal.Zero,
        PlatformFee:   decimal.Zero,
        TotalCost:     decimal.Zero,
        Currency:      "ETH",
    }, nil
}

// 辅助函数
func (s *ImmutableMintService) signRequest(data []byte) (string, error) {
    hash := crypto.Keccak256Hash(data)
    signature, err := crypto.Sign(hash.Bytes(), s.privateKey)
    if err != nil {
        return "", err
    }
    
    return fmt.Sprintf("0x%x", signature), nil
}

func (s *ImmutableMintService) buildGameItemBlueprint(item GameItem) string {
    // 构建游戏道具的元数据blueprint
    blueprint := map[string]interface{}{
        "name":        item.Name,
        "description": item.Description,
        "image":       item.ImageURL,
        "attributes": []map[string]interface{}{
            {"trait_type": "Rarity", "value": item.Rarity},
            {"trait_type": "Type", "value": item.Type},
            {"trait_type": "Level", "value": item.Level},
            {"trait_type": "Attack", "value": item.Attack},
            {"trait_type": "Defense", "value": item.Defense},
            {"trait_type": "Speed", "value": item.Speed},
        },
        "game_data": item.GameData,
    }
    
    jsonData, _ := json.Marshal(blueprint)
    return string(jsonData)
}

func (s *ImmutableMintService) buildCollectibleBlueprint(collectible Collectible, collectionAddress string) string {
    // 构建收藏品的元数据blueprint
    blueprint := map[string]interface{}{
        "name":        collectible.Name,
        "description": collectible.Description,
        "image":       collectible.ImageURL,
        "external_url": collectible.ExternalURL,
        "attributes":  collectible.Attributes,
        "collection":  collectionAddress,
    }
    
    jsonData, _ := json.Marshal(blueprint)
    return string(jsonData)
}

// 数据结构
type GameItem struct {
    TokenID           string
    Name              string
    Description       string
    ImageURL          string
    Rarity            string
    Type              string
    Level             int
    Attack            int
    Defense           int
    Speed             int
    GameData          map[string]interface{}
    RoyaltyRecipient  string
    RoyaltyPercentage decimal.Decimal
}

type Collectible struct {
    TokenID           string
    Name              string
    Description       string
    ImageURL          string
    ExternalURL       string
    Attributes        []map[string]interface{}
    Creator           string
    RoyaltyPercentage decimal.Decimal
}

type MintStatus struct {
    MintID    string `json:"mint_id"`
    Status    string `json:"status"`
    Assets    []*Asset `json:"assets"`
    CreatedAt string `json:"created_at"`
    UpdatedAt string `json:"updated_at"`
}

type MintCost struct {
    TokenCount  int
    GasFee      decimal.Decimal
    PlatformFee decimal.Decimal
    TotalCost   decimal.Decimal
    Currency    string
}
```

## NFT交易

### 4.1 交易服务

```go
// services/trading_service.go
package services

import (
    "bytes"
    "crypto/ecdsa"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
    
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/shopspring/decimal"
)

type ImmutableTradingService struct {
    client     *ImmutableClient
    privateKey *ecdsa.PrivateKey
}

func NewImmutableTradingService(client *ImmutableClient, privateKey *ecdsa.PrivateKey) *ImmutableTradingService {
    return &ImmutableTradingService{
        client:     client,
        privateKey: privateKey,
    }
}

// 创建卖单
func (s *ImmutableTradingService) CreateSellOrder(
    ethAddress string,
    tokenAddress string,
    tokenID string,
    sellTokenType string,
    sellQuantity string,
    buyTokenType string,
    buyQuantity string,
    expirationTimestamp time.Time,
) (*Order, error) {
    url := fmt.Sprintf("%s/v1/orders", s.client.config.APIURL)
    
    orderData := map[string]interface{}{
        "user": ethAddress,
        "sell": map[string]interface{}{
            "type": sellTokenType,
            "data": map[string]interface{}{
                "token_address": tokenAddress,
                "token_id":      tokenID,
                "quantity":      sellQuantity,
            },
        },
        "buy": map[string]interface{}{
            "type": buyTokenType,
            "data": map[string]interface{}{
                "quantity": buyQuantity,
            },
        },
        "expiration_timestamp": expirationTimestamp.Unix(),
    }
    
    jsonData, err := json.Marshal(orderData)
    if err != nil {
        return nil, err
    }
    
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Content-Type", "application/json")
    if s.client.apiKey != "" {
        req.Header.Set("x-api-key", s.client.apiKey)
    }
    
    // 添加签名认证
    signature, err := s.signRequest(jsonData)
    if err != nil {
        return nil, err
    }
    req.Header.Set("x-signature", signature)
    
    resp, err := s.client.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusCreated {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("创建卖单失败: %d, %s", resp.StatusCode, string(body))
    }
    
    var order Order
    err = json.NewDecoder(resp.Body).Decode(&order)
    if err != nil {
        return nil, err
    }
    
    return &order, nil
}

// 创建买单
func (s *ImmutableTradingService) CreateBuyOrder(
    ethAddress string,
    tokenAddress string,
    tokenID string,
    buyTokenType string,
    buyQuantity string,
    sellTokenType string,
    sellQuantity string,
    expirationTimestamp time.Time,
) (*Order, error) {
    url := fmt.Sprintf("%s/v1/orders", s.client.config.APIURL)
    
    orderData := map[string]interface{}{
        "user": ethAddress,
        "buy": map[string]interface{}{
            "type": buyTokenType,
            "data": map[string]interface{}{
                "token_address": tokenAddress,
                "token_id":      tokenID,
                "quantity":      buyQuantity,
            },
        },
        "sell": map[string]interface{}{
            "type": sellTokenType,
            "data": map[string]interface{}{
                "quantity": sellQuantity,
            },
        },
        "expiration_timestamp": expirationTimestamp.Unix(),
    }
    
    jsonData, err := json.Marshal(orderData)
    if err != nil {
        return nil, err
    }
    
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Content-Type", "application/json")
    if s.client.apiKey != "" {
        req.Header.Set("x-api-key", s.client.apiKey)
    }
    
    // 添加签名认证
    signature, err := s.signRequest(jsonData)
    if err != nil {
        return nil, err
    }
    req.Header.Set("x-signature", signature)
    
    resp, err := s.client.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusCreated {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("创建买单失败: %d, %s", resp.StatusCode, string(body))
    }
    
    var order Order
    err = json.NewDecoder(resp.Body).Decode(&order)
    if err != nil {
        return nil, err
    }
    
    return &order, nil
}

// 取消订单
func (s *ImmutableTradingService) CancelOrder(orderID int) error {
    url := fmt.Sprintf("%s/v1/orders/%d", s.client.config.APIURL, orderID)
    
    req, err := http.NewRequest("DELETE", url, nil)
    if err != nil {
        return err
    }
    
    req.Header.Set("Content-Type", "application/json")
    if s.client.apiKey != "" {
        req.Header.Set("x-api-key", s.client.apiKey)
    }
    
    resp, err := s.client.httpClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("取消订单失败: %d, %s", resp.StatusCode, string(body))
    }
    
    return nil
}

// 执行交易
func (s *ImmutableTradingService) ExecuteTrade(
    orderID int,
    ethAddress string,
) (*Trade, error) {
    url := fmt.Sprintf("%s/v1/trades", s.client.config.APIURL)
    
    tradeData := map[string]interface{}{
        "order_id": orderID,
        "user":     ethAddress,
    }
    
    jsonData, err := json.Marshal(tradeData)
    if err != nil {
        return nil, err
    }
    
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Content-Type", "application/json")
    if s.client.apiKey != "" {
        req.Header.Set("x-api-key", s.client.apiKey)
    }
    
    // 添加签名认证
    signature, err := s.signRequest(jsonData)
    if err != nil {
        return nil, err
    }
    req.Header.Set("x-signature", signature)
    
    resp, err := s.client.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusCreated {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("执行交易失败: %d, %s", resp.StatusCode, string(body))
    }
    
    var trade Trade
    err = json.NewDecoder(resp.Body).Decode(&trade)
    if err != nil {
        return nil, err
    }
    
    return &trade, nil
}

// 转移NFT
func (s *ImmutableTradingService) TransferNFT(
    fromAddress string,
    toAddress string,
    tokenAddress string,
    tokenID string,
) (*Transfer, error) {
    url := fmt.Sprintf("%s/v1/transfers", s.client.config.APIURL)
    
    transferData := map[string]interface{}{
        "sender":   fromAddress,
        "receiver": toAddress,
        "token": map[string]interface{}{
            "type": "ERC721",
            "data": map[string]interface{}{
                "token_address": tokenAddress,
                "token_id":      tokenID,
            },
        },
    }
    
    jsonData, err := json.Marshal(transferData)
    if err != nil {
        return nil, err
    }
    
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Content-Type", "application/json")
    if s.client.apiKey != "" {
        req.Header.Set("x-api-key", s.client.apiKey)
    }
    
    // 添加签名认证
    signature, err := s.signRequest(jsonData)
    if err != nil {
        return nil, err
    }
    req.Header.Set("x-signature", signature)
    
    resp, err := s.client.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusCreated {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("转移NFT失败: %d, %s", resp.StatusCode, string(body))
    }
    
    var transfer Transfer
    err = json.NewDecoder(resp.Body).Decode(&transfer)
    if err != nil {
        return nil, err
    }
    
    return &transfer, nil
}

// 批量转移NFT
func (s *ImmutableTradingService) BatchTransferNFTs(
    fromAddress string,
    transfers []NFTTransfer,
) ([]*Transfer, error) {
    var results []*Transfer
    
    for _, transfer := range transfers {
        result, err := s.TransferNFT(
            fromAddress,
            transfer.ToAddress,
            transfer.TokenAddress,
            transfer.TokenID,
        )
        if err != nil {
            return nil, err
        }
        
        results = append(results, result)
        
        // 添加延迟避免API限制
        time.Sleep(100 * time.Millisecond)
    }
    
    return results, nil
}

// 获取市场价格
func (s *ImmutableTradingService) GetMarketPrice(
    tokenAddress string,
    tokenID string,
) (*MarketPrice, error) {
    // 获取该NFT的最近交易
    trades, _, err := s.client.GetTrades("", "", 10)
    if err != nil {
        return nil, err
    }
    
    var recentPrices []decimal.Decimal
    for _, trade := range trades {
        if trade.TokenData.TokenAddress == tokenAddress && 
           trade.TokenData.TokenID == tokenID {
            price, _ := decimal.NewFromString(trade.TokenData.Quantity)
            recentPrices = append(recentPrices, price)
        }
    }
    
    if len(recentPrices) == 0 {
        return &MarketPrice{
            TokenAddress: tokenAddress,
            TokenID:      tokenID,
            LastPrice:    decimal.Zero,
            AveragePrice: decimal.Zero,
            FloorPrice:   decimal.Zero,
            CeilingPrice: decimal.Zero,
        }, nil
    }
    
    // 计算价格统计
    lastPrice := recentPrices[0]
    total := decimal.Zero
    minPrice := recentPrices[0]
    maxPrice := recentPrices[0]
    
    for _, price := range recentPrices {
        total = total.Add(price)
        if price.LessThan(minPrice) {
            minPrice = price
        }
        if price.GreaterThan(maxPrice) {
            maxPrice = price
        }
    }
    
    averagePrice := total.Div(decimal.NewFromInt(int64(len(recentPrices))))
    
    return &MarketPrice{
        TokenAddress: tokenAddress,
        TokenID:      tokenID,
        LastPrice:    lastPrice,
        AveragePrice: averagePrice,
        FloorPrice:   minPrice,
        CeilingPrice: maxPrice,
    }, nil
}

// 辅助函数
func (s *ImmutableTradingService) signRequest(data []byte) (string, error) {
    hash := crypto.Keccak256Hash(data)
    signature, err := crypto.Sign(hash.Bytes(), s.privateKey)
    if err != nil {
        return "", err
    }
    
    return fmt.Sprintf("0x%x", signature), nil
}

// 数据结构
type NFTTransfer struct {
    ToAddress    string
    TokenAddress string
    TokenID      string
}

type MarketPrice struct {
    TokenAddress string
    TokenID      string
    LastPrice    decimal.Decimal
    AveragePrice decimal.Decimal
    FloorPrice   decimal.Decimal
    CeilingPrice decimal.Decimal
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
    
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/shopspring/decimal"
    
    "your-project/client"
    "your-project/services"
)

func main() {
    // 创建Immutable X客户端 (使用测试网)
    immutableClient := client.NewImmutableClient(
        &client.TestnetConfig,
        "your_api_key_here",
        nil, // 私钥稍后设置
    )
    
    // 加载私钥
    privateKey, err := crypto.HexToECDSA("your_private_key_here")
    if err != nil {
        log.Fatal("加载私钥失败:", err)
    }
    
    immutableClient.privateKey = privateKey
    
    // 获取用户地址
    userAddress := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
    
    // 创建服务
    mintService := services.NewImmutableMintService(immutableClient, privateKey)
    tradingService := services.NewImmutableTradingService(immutableClient, privateKey)
    
    // 1. 获取网络信息
    config := immutableClient.GetNetworkConfig()
    fmt.Printf("=== Immutable X网络信息 ===\n")
    fmt.Printf("网络名称: %s\n", config.NetworkName)
    fmt.Printf("API URL: %s\n", config.APIURL)
    fmt.Printf("Link URL: %s\n", config.LinkURL)
    fmt.Printf("链ID: %d\n", config.ChainID)
    fmt.Printf("浏览器: %s\n", config.ExplorerURL)
    
    // 2. 获取用户信息
    fmt.Printf("\n=== 用户信息 ===\n")
    
    user, err := immutableClient.GetUser(userAddress)
    if err != nil {
        log.Printf("获取用户信息失败: %v", err)
    } else {
        fmt.Printf("ETH地址: %s\n", user.EthAddress)
        fmt.Printf("Stark密钥: %s\n", user.StarkKey)
        fmt.Printf("账户数量: %d\n", len(user.Accounts))
    }
    
    // 3. 获取用户资产
    fmt.Printf("\n=== 用户NFT资产 ===\n")
    
    assets, cursor, err := immutableClient.GetAssets(userAddress, "", 10)
    if err != nil {
        log.Printf("获取用户资产失败: %v", err)
    } else {
        fmt.Printf("拥有NFT数量: %d\n", len(assets))
        
        for i, asset := range assets {
            fmt.Printf("  NFT %d:\n", i+1)
            fmt.Printf("    合约: %s\n", asset.TokenAddress)
            fmt.Printf("    Token ID: %s\n", asset.TokenID)
            fmt.Printf("    名称: %s\n", asset.Name)
            fmt.Printf("    描述: %s\n", asset.Description)
            fmt.Printf("    状态: %s\n", asset.Status)
            
            if asset.Collection != nil {
                fmt.Printf("    集合: %s\n", asset.Collection.Name)
            }
        }
        
        if cursor != "" {
            fmt.Printf("还有更多资产，游标: %s\n", cursor)
        }
    }
    
    // 4. 获取用户余额
    fmt.Printf("\n=== 用户余额 ===\n")
    
    balances, err := immutableClient.GetBalances(userAddress)
    if err != nil {
        log.Printf("获取用户余额失败: %v", err)
    } else {
        fmt.Printf("余额信息:\n")
        
        for _, balance := range balances {
            if balance.Token.Type == "ETH" {
                fmt.Printf("  ETH余额: %s\n", balance.Balance)
                fmt.Printf("  可提取: %s\n", balance.Withdrawable)
                fmt.Printf("  准备提取: %s\n", balance.PreparingWithdrawal)
            }
        }
    }
    
    // 5. 创建项目和集合示例
    fmt.Printf("\n=== 创建项目和集合示例 ===\n")
    
    // 创建项目
    project, err := immutableClient.CreateProject(
        "My Game Project",
        "Game Studio Inc",
        "contact@gamestudio.com",
    )
    if err != nil {
        log.Printf("创建项目失败: %v", err)
    } else {
        fmt.Printf("项目创建成功:\n")
        fmt.Printf("  项目ID: %d\n", project.ID)
        fmt.Printf("  项目名称: %s\n", project.Name)
        fmt.Printf("  公司名称: %s\n", project.CompanyName)
        fmt.Printf("  月度铸造限制: %d\n", project.MintMonthlyLimit)
        fmt.Printf("  剩余铸造次数: %d\n", project.MintRemaining)
        
        // 创建集合
        collection, err := immutableClient.CreateCollection(
            "Epic Game Items",
            "Rare and legendary items from our epic game",
            "https://example.com/icon.png",
            "https://api.example.com/metadata",
            project.ID,
        )
        if err != nil {
            log.Printf("创建集合失败: %v", err)
        } else {
            fmt.Printf("集合创建成功:\n")
            fmt.Printf("  集合地址: %s\n", collection.Address)
            fmt.Printf("  集合名称: %s\n", collection.Name)
            fmt.Printf("  描述: %s\n", collection.Description)
        }
    }
    
    // 6. NFT铸造示例
    fmt.Printf("\n=== NFT铸造示例 ===\n")
    
    // 单个游戏道具铸造
    gameItem := services.GameItem{
        TokenID:           "sword_001",
        Name:              "Legendary Fire Sword",
        Description:       "A powerful sword forged in dragon fire",
        ImageURL:          "https://example.com/sword.png",
        Rarity:            "Legendary",
        Type:              "Weapon",
        Level:             50,
        Attack:            150,
        Defense:           20,
        Speed:             30,
        GameData:          map[string]interface{}{"element": "fire", "durability": 100},
        RoyaltyRecipient:  userAddress,
        RoyaltyPercentage: decimal.NewFromFloat(0.05), // 5%
    }
    
    gameItems := []services.GameItem{gameItem}
    
    mintedAssets, err := mintService.MintGameItems(userAddress, gameItems)
    if err != nil {
        log.Printf("铸造游戏道具失败: %v", err)
    } else {
        fmt.Printf("游戏道具铸造成功:\n")
        for _, asset := range mintedAssets {
            fmt.Printf("  资产ID: %s\n", asset.ID)
            fmt.Printf("  Token ID: %s\n", asset.TokenID)
            fmt.Printf("  名称: %s\n", asset.Name)
            fmt.Printf("  状态: %s\n", asset.Status)
        }
    }
    
    // 7. NFT交易示例
    fmt.Printf("\n=== NFT交易示例 ===\n")
    
    if len(assets) > 0 {
        asset := assets[0]
        
        // 创建卖单
        sellOrder, err := tradingService.CreateSellOrder(
            userAddress,
            asset.TokenAddress,
            asset.TokenID,
            "ERC721",
            "1",
            "ETH",
            "100000000000000000", // 0.1 ETH
            time.Now().Add(24*time.Hour), // 24小时后过期
        )
        if err != nil {
            log.Printf("创建卖单失败: %v", err)
        } else {
            fmt.Printf("卖单创建成功:\n")
            fmt.Printf("  订单ID: %d\n", sellOrder.OrderID)
            fmt.Printf("  状态: %s\n", sellOrder.Status)
            fmt.Printf("  卖出: %s (Token ID: %s)\n", 
                sellOrder.SellToken.Data.TokenAddress,
                sellOrder.SellToken.Data.TokenID)
            fmt.Printf("  价格: %s ETH\n", sellOrder.BuyToken.Data.Quantity)
            
            // 获取市场价格
            marketPrice, err := tradingService.GetMarketPrice(asset.TokenAddress, asset.TokenID)
            if err != nil {
                log.Printf("获取市场价格失败: %v", err)
            } else {
                fmt.Printf("市场价格信息:\n")
                fmt.Printf("  最新价格: %s ETH\n", marketPrice.LastPrice.String())
                fmt.Printf("  平均价格: %s ETH\n", marketPrice.AveragePrice.String())
                fmt.Printf("  地板价: %s ETH\n", marketPrice.FloorPrice.String())
                fmt.Printf("  天花板价: %s ETH\n", marketPrice.CeilingPrice.String())
            }
        }
    }
    
    // 8. 获取订单和交易历史
    fmt.Printf("\n=== 订单和交易历史 ===\n")
    
    // 获取用户订单
    orders, orderCursor, err := immutableClient.GetOrders("active", userAddress, "", 5)
    if err != nil {
        log.Printf("获取订单失败: %v", err)
    } else {
        fmt.Printf("活跃订单数量: %d\n", len(orders))
        
        for i, order := range orders {
            fmt.Printf("  订单 %d:\n", i+1)
            fmt.Printf("    订单ID: %d\n", order.OrderID)
            fmt.Printf("    状态: %s\n", order.Status)
            fmt.Printf("    创建时间: %s\n", order.Timestamp)
            fmt.Printf("    过期时间: %s\n", order.ExpirationTimestamp)
        }
    }
    
    // 获取交易历史
    trades, tradeCursor, err := immutableClient.GetTrades(userAddress, "", 5)
    if err != nil {
        log.Printf("获取交易历史失败: %v", err)
    } else {
        fmt.Printf("交易历史数量: %d\n", len(trades))
        
        for i, trade := range trades {
            fmt.Printf("  交易 %d:\n", i+1)
            fmt.Printf("    交易ID: %d\n", trade.TransactionID)
            fmt.Printf("    状态: %s\n", trade.Status)
            fmt.Printf("    时间: %s\n", trade.Timestamp)
            fmt.Printf("    接收者: %s\n", trade.Receiver)
        }
    }
    
    // 9. Immutable X特性总结
    fmt.Printf("\n=== Immutable X特性总结 ===\n")
    
    fmt.Printf("Immutable X优势:\n")
    fmt.Printf("  - 零Gas费用的NFT铸造和交易\n")
    fmt.Printf("  - 即时交易确认\n")
    fmt.Printf("  - 以太坊级别的安全性\n")
    fmt.Printf("  - 大规模游戏应用支持\n")
    fmt.Printf("  - 完整的NFT市场功能\n")
    
    fmt.Printf("\n主要功能:\n")
    fmt.Printf("  - NFT铸造: 批量零费用铸造\n")
    fmt.Printf("  - NFT交易: 即时买卖交易\n")
    fmt.Printf("  - NFT转移: 免费转移\n")
    fmt.Printf("  - 市场集成: 内置交易市场\n")
    fmt.Printf("  - 游戏集成: 专为游戏优化\n")
    
    fmt.Printf("\n适用场景:\n")
    fmt.Printf("  - 区块链游戏\n")
    fmt.Printf("  - NFT收藏品\n")
    fmt.Printf("  - 数字艺术品\n")
    fmt.Printf("  - 虚拟资产交易\n")
    fmt.Printf("  - 大规模NFT应用\n")
    
    // 10. 最佳实践建议
    fmt.Printf("\n=== 最佳实践建议 ===\n")
    
    fmt.Printf("使用Immutable X时请注意:\n")
    fmt.Printf("  1. 合理设计NFT元数据结构\n")
    fmt.Printf("  2. 优化游戏道具属性设计\n")
    fmt.Printf("  3. 设置合理的版税比例\n")
    fmt.Printf("  4. 监控铸造和交易限制\n")
    fmt.Printf("  5. 实现适当的错误处理\n")
    fmt.Printf("  6. 关注API调用频率限制\n")
    fmt.Printf("  7. 保护好私钥和API密钥\n")
    fmt.Printf("  8. 测试网充分验证后再上主网\n")
}
```

这个Immutable X使用指南提供了完整的NFT Layer2解决方案集成方案，涵盖了零Gas费用铸造、即时交易、批量操作、游戏集成、市场功能等核心特性，是构建大规模NFT和区块链游戏应用的重要参考文档。
