# OpenSea API 详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [环境准备](#环境准备)
3. [API认证](#API认证)
4. [NFT查询](#NFT查询)
5. [集合管理](#集合管理)
6. [交易操作](#交易操作)
7. [市场数据](#市场数据)
8. [事件监听](#事件监听)
9. [实际应用](#实际应用)

## 基础概念

### 1.1 OpenSea简介

OpenSea是全球最大的NFT交易平台，提供完整的API接口用于NFT查询、交易、集合管理等功能。

```bash
# 安装OpenSea相关依赖
go get github.com/go-resty/resty/v2
go get github.com/shopspring/decimal
go get github.com/ethereum/go-ethereum
```

### 1.2 API配置

```go
// config/opensea.go
package config

import (
    "time"
)

type OpenSeaConfig struct {
    APIKey      string
    BaseURL     string
    TestnetURL  string
    Timeout     time.Duration
    RateLimit   int
    RetryCount  int
    UserAgent   string
}

func DefaultOpenSeaConfig() *OpenSeaConfig {
    return &OpenSeaConfig{
        APIKey:     "", // 需要从OpenSea获取
        BaseURL:    "https://api.opensea.io/api/v1",
        TestnetURL: "https://testnets-api.opensea.io/api/v1",
        Timeout:    30 * time.Second,
        RateLimit:  4, // 每秒4个请求
        RetryCount: 3,
        UserAgent:  "Go-OpenSea-Client/1.0",
    }
}

// API端点
const (
    // NFT相关
    EndpointAssets      = "/assets"
    EndpointAsset       = "/asset"
    EndpointCollections = "/collections"
    EndpointCollection  = "/collection"

    // 交易相关
    EndpointOrders      = "/orders"
    EndpointEvents      = "/events"

    // 用户相关
    EndpointAccount     = "/account"
)

// 网络配置
type NetworkConfig struct {
    ChainName string
    ChainID   int
}

var (
    Ethereum = NetworkConfig{ChainName: "ethereum", ChainID: 1}
    Polygon  = NetworkConfig{ChainName: "matic", ChainID: 137}
    Klaytn   = NetworkConfig{ChainName: "klaytn", ChainID: 8217}
    BSC      = NetworkConfig{ChainName: "bsc", ChainID: 56}
)
```

## 环境准备

### 2.1 HTTP客户端

```go
// client/opensea_client.go
package client

import (
    "fmt"
    "time"

    "github.com/go-resty/resty/v2"

    "your-project/config"
)

type OpenSeaClient struct {
    client *resty.Client
    config *config.OpenSeaConfig
    apiKey string
}

func NewOpenSeaClient(cfg *config.OpenSeaConfig, apiKey string) *OpenSeaClient {
    client := resty.New()

    // 设置基础配置
    client.SetBaseURL(cfg.BaseURL)
    client.SetTimeout(cfg.Timeout)
    client.SetRetryCount(cfg.RetryCount)

    // 设置请求头
    client.SetHeaders(map[string]string{
        "Accept":       "application/json",
        "User-Agent":   cfg.UserAgent,
        "X-API-KEY":    apiKey,
    })

    // 设置重试条件
    client.AddRetryCondition(func(r *resty.Response, err error) bool {
        return r.StatusCode() == 429 || r.StatusCode() >= 500
    })

    return &OpenSeaClient{
        client: client,
        config: cfg,
        apiKey: apiKey,
    }
}

// 切换到测试网
func (c *OpenSeaClient) UseTestnet() {
    c.client.SetBaseURL(c.config.TestnetURL)
}

// 切换到主网
func (c *OpenSeaClient) UseMainnet() {
    c.client.SetBaseURL(c.config.BaseURL)
}

// 执行GET请求
func (c *OpenSeaClient) Get(endpoint string, params map[string]string) (*resty.Response, error) {
    req := c.client.R()

    if params != nil {
        req.SetQueryParams(params)
    }

    resp, err := req.Get(endpoint)
    if err != nil {
        return nil, fmt.Errorf("GET请求失败: %v", err)
    }

    if resp.StatusCode() != 200 {
        return nil, fmt.Errorf("API请求失败: %d - %s", resp.StatusCode(), resp.String())
    }

    return resp, nil
}

// 执行POST请求
func (c *OpenSeaClient) Post(endpoint string, body interface{}) (*resty.Response, error) {
    req := c.client.R()

    if body != nil {
        req.SetBody(body)
        req.SetHeader("Content-Type", "application/json")
    }

    resp, err := req.Post(endpoint)
    if err != nil {
        return nil, fmt.Errorf("POST请求失败: %v", err)
    }

    if resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
        return nil, fmt.Errorf("API请求失败: %d - %s", resp.StatusCode(), resp.String())
    }

    return resp, nil
}

// 速率限制控制
func (c *OpenSeaClient) RateLimit() {
    time.Sleep(time.Second / time.Duration(c.config.RateLimit))
}
```

## API认证

### 3.1 认证管理

```go
// auth/manager.go
package auth

import (
    "crypto/ecdsa"
    "fmt"
    "time"

    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/common/hexutil"
)

type AuthManager struct {
    apiKey     string
    privateKey *ecdsa.PrivateKey
    address    string
}

func NewAuthManager(apiKey, privateKeyHex string) (*AuthManager, error) {
    var privateKey *ecdsa.PrivateKey
    var address string

    if privateKeyHex != "" {
        pk, err := crypto.HexToECDSA(privateKeyHex)
        if err != nil {
            return nil, fmt.Errorf("解析私钥失败: %v", err)
        }
        privateKey = pk
        address = crypto.PubkeyToAddress(pk.PublicKey).Hex()
    }

    return &AuthManager{
        apiKey:     apiKey,
        privateKey: privateKey,
        address:    address,
    }, nil
}

// 获取API密钥
func (am *AuthManager) GetAPIKey() string {
    return am.apiKey
}

// 获取钱包地址
func (am *AuthManager) GetAddress() string {
    return am.address
}

// 签名消息（用于某些需要签名的操作）
func (am *AuthManager) SignMessage(message string) (string, error) {
    if am.privateKey == nil {
        return "", fmt.Errorf("未设置私钥")
    }

    // 添加以太坊消息前缀
    prefixedMessage := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
    hash := crypto.Keccak256Hash([]byte(prefixedMessage))

    signature, err := crypto.Sign(hash.Bytes(), am.privateKey)
    if err != nil {
        return "", fmt.Errorf("签名失败: %v", err)
    }

    // 调整v值
    signature[64] += 27

    return hexutil.Encode(signature), nil
}

// 生成时间戳签名（某些API需要）
func (am *AuthManager) GenerateTimestampSignature() (string, string, error) {
    timestamp := fmt.Sprintf("%d", time.Now().Unix())
    signature, err := am.SignMessage(timestamp)
    if err != nil {
        return "", "", err
    }
    return timestamp, signature, nil
}
```

## NFT查询

### 4.1 资产查询

```go
// nft/assets.go
package nft

import (
    "encoding/json"
    "fmt"
    "strconv"

    "your-project/client"
    "your-project/config"
)

type AssetService struct {
    client *client.OpenSeaClient
}

func NewAssetService(client *client.OpenSeaClient) *AssetService {
    return &AssetService{
        client: client,
    }
}

// 获取资产列表
func (as *AssetService) GetAssets(params AssetQueryParams) (*AssetsResponse, error) {
    queryParams := make(map[string]string)

    if params.Owner != "" {
        queryParams["owner"] = params.Owner
    }
    if params.Collection != "" {
        queryParams["collection"] = params.Collection
    }
    if params.TokenIDs != nil && len(params.TokenIDs) > 0 {
        for i, tokenID := range params.TokenIDs {
            queryParams[fmt.Sprintf("token_ids[%d]", i)] = tokenID
        }
    }
    if params.AssetContractAddress != "" {
        queryParams["asset_contract_address"] = params.AssetContractAddress
    }
    if params.Limit > 0 {
        queryParams["limit"] = strconv.Itoa(params.Limit)
    }
    if params.Offset > 0 {
        queryParams["offset"] = strconv.Itoa(params.Offset)
    }
    if params.OrderBy != "" {
        queryParams["order_by"] = params.OrderBy
    }
    if params.OrderDirection != "" {
        queryParams["order_direction"] = params.OrderDirection
    }

    resp, err := as.client.Get(config.EndpointAssets, queryParams)
    if err != nil {
        return nil, err
    }

    var assetsResp AssetsResponse
    err = json.Unmarshal(resp.Body(), &assetsResp)
    if err != nil {
        return nil, fmt.Errorf("解析响应失败: %v", err)
    }

    return &assetsResp, nil
}

// 获取单个资产
func (as *AssetService) GetAsset(contractAddress, tokenID string) (*Asset, error) {
    endpoint := fmt.Sprintf("%s/%s/%s", config.EndpointAsset, contractAddress, tokenID)

    resp, err := as.client.Get(endpoint, nil)
    if err != nil {
        return nil, err
    }

    var asset Asset
    err = json.Unmarshal(resp.Body(), &asset)
    if err != nil {
        return nil, fmt.Errorf("解析资产数据失败: %v", err)
    }

    return &asset, nil
}

// 获取资产历史
func (as *AssetService) GetAssetHistory(contractAddress, tokenID string, params HistoryParams) (*EventsResponse, error) {
    queryParams := map[string]string{
        "asset_contract_address": contractAddress,
        "token_id":              tokenID,
    }

    if params.EventType != "" {
        queryParams["event_type"] = params.EventType
    }
    if params.Limit > 0 {
        queryParams["limit"] = strconv.Itoa(params.Limit)
    }
    if params.Offset > 0 {
        queryParams["offset"] = strconv.Itoa(params.Offset)
    }

    resp, err := as.client.Get(config.EndpointEvents, queryParams)
    if err != nil {
        return nil, err
    }

    var eventsResp EventsResponse
    err = json.Unmarshal(resp.Body(), &eventsResp)
    if err != nil {
        return nil, fmt.Errorf("解析事件数据失败: %v", err)
    }

    return &eventsResp, nil
}

// 搜索资产
func (as *AssetService) SearchAssets(query string, params SearchParams) (*AssetsResponse, error) {
    queryParams := map[string]string{
        "search": query,
    }

    if params.Collection != "" {
        queryParams["collection"] = params.Collection
    }
    if params.Limit > 0 {
        queryParams["limit"] = strconv.Itoa(params.Limit)
    }

    resp, err := as.client.Get(config.EndpointAssets, queryParams)
    if err != nil {
        return nil, err
    }

    var assetsResp AssetsResponse
    err = json.Unmarshal(resp.Body(), &assetsResp)
    if err != nil {
        return nil, fmt.Errorf("解析搜索结果失败: %v", err)
    }

    return &assetsResp, nil
}

// 数据结构定义
type AssetQueryParams struct {
    Owner                string
    Collection           string
    TokenIDs             []string
    AssetContractAddress string
    Limit                int
    Offset               int
    OrderBy              string
    OrderDirection       string
}

type HistoryParams struct {
    EventType string
    Limit     int
    Offset    int
}

type SearchParams struct {
    Collection string
    Limit      int
}

type AssetsResponse struct {
    Assets []Asset `json:"assets"`
}

type Asset struct {
    ID                   int             `json:"id"`
    TokenID              string          `json:"token_id"`
    NumSales             int             `json:"num_sales"`
    BackgroundColor      string          `json:"background_color"`
    ImageURL             string          `json:"image_url"`
    ImagePreviewURL      string          `json:"image_preview_url"`
    ImageThumbnailURL    string          `json:"image_thumbnail_url"`
    ImageOriginalURL     string          `json:"image_original_url"`
    AnimationURL         string          `json:"animation_url"`
    AnimationOriginalURL string          `json:"animation_original_url"`
    Name                 string          `json:"name"`
    Description          string          `json:"description"`
    ExternalLink         string          `json:"external_link"`
    AssetContract        AssetContract   `json:"asset_contract"`
    Permalink            string          `json:"permalink"`
    Collection           Collection      `json:"collection"`
    Decimals             int             `json:"decimals"`
    TokenMetadata        string          `json:"token_metadata"`
    Owner                Account         `json:"owner"`
    SellOrders           []Order         `json:"sell_orders"`
    Creator              Account         `json:"creator"`
    Traits               []Trait         `json:"traits"`
    LastSale             *Sale           `json:"last_sale"`
    TopBid               *Order          `json:"top_bid"`
    ListingDate          string          `json:"listing_date"`
    IsPresale            bool            `json:"is_presale"`
    TransferFeePayment   *Payment        `json:"transfer_fee_payment"`
    TransferFee          string          `json:"transfer_fee"`
}

type AssetContract struct {
    Address                     string `json:"address"`
    AssetContractType           string `json:"asset_contract_type"`
    CreatedDate                 string `json:"created_date"`
    Name                        string `json:"name"`
    NftVersion                  string `json:"nft_version"`
    OpenseaVersion              string `json:"opensea_version"`
    Owner                       int    `json:"owner"`
    SchemaName                  string `json:"schema_name"`
    Symbol                      string `json:"symbol"`
    TotalSupply                 string `json:"total_supply"`
    Description                 string `json:"description"`
    ExternalLink                string `json:"external_link"`
    ImageURL                    string `json:"image_url"`
    DefaultToFiat               bool   `json:"default_to_fiat"`
    DevBuyerFeeBasisPoints      int    `json:"dev_buyer_fee_basis_points"`
    DevSellerFeeBasisPoints     int    `json:"dev_seller_fee_basis_points"`
    OnlyProxiedTransfers        bool   `json:"only_proxied_transfers"`
    OpenseaBuyerFeeBasisPoints  int    `json:"opensea_buyer_fee_basis_points"`
    OpenseaSellerFeeBasisPoints int    `json:"opensea_seller_fee_basis_points"`
    BuyerFeeBasisPoints         int    `json:"buyer_fee_basis_points"`
    SellerFeeBasisPoints        int    `json:"seller_fee_basis_points"`
    PayoutAddress               string `json:"payout_address"`
}
```

type Collection struct {
    BannerImageURL          string            `json:"banner_image_url"`
    ChatURL                 string            `json:"chat_url"`
    CreatedDate             string            `json:"created_date"`
    DefaultToFiat           bool              `json:"default_to_fiat"`
    Description             string            `json:"description"`
    DevBuyerFeeBasisPoints  string            `json:"dev_buyer_fee_basis_points"`
    DevSellerFeeBasisPoints string            `json:"dev_seller_fee_basis_points"`
    DiscordURL              string            `json:"discord_url"`
    DisplayData             map[string]string `json:"display_data"`
    ExternalURL             string            `json:"external_url"`
    Featured                bool              `json:"featured"`
    FeaturedImageURL        string            `json:"featured_image_url"`
    Hidden                  bool              `json:"hidden"`
    SafelistRequestStatus   string            `json:"safelist_request_status"`
    ImageURL                string            `json:"image_url"`
    IsSubjectToWhitelist    bool              `json:"is_subject_to_whitelist"`
    LargeImageURL           string            `json:"large_image_url"`
    MediumUsername          string            `json:"medium_username"`
    Name                    string            `json:"name"`
    OnlyProxiedTransfers    bool              `json:"only_proxied_transfers"`
    OpenseaBuyerFeeBasisPoints  string        `json:"opensea_buyer_fee_basis_points"`
    OpenseaSellerFeeBasisPoints string        `json:"opensea_seller_fee_basis_points"`
    PayoutAddress           string            `json:"payout_address"`
    RequireEmail            bool              `json:"require_email"`
    ShortDescription        string            `json:"short_description"`
    Slug                    string            `json:"slug"`
    TelegramURL             string            `json:"telegram_url"`
    TwitterUsername         string            `json:"twitter_username"`
    InstagramUsername       string            `json:"instagram_username"`
    WikiURL                 string            `json:"wiki_url"`
}

type Account struct {
    User         *User  `json:"user"`
    ProfileImgURL string `json:"profile_img_url"`
    Address      string `json:"address"`
    Config       string `json:"config"`
}

type User struct {
    Username string `json:"username"`
    ID       int    `json:"id"`
}

type Order struct {
    CreatedDate         string   `json:"created_date"`
    ClosingDate         string   `json:"closing_date"`
    ClosingExtendable   bool     `json:"closing_extendable"`
    ExpirationTime      int64    `json:"expiration_time"`
    ListingTime         int64    `json:"listing_time"`
    OrderHash           string   `json:"order_hash"`
    Metadata            Metadata `json:"metadata"`
    Exchange            string   `json:"exchange"`
    Maker               Account  `json:"maker"`
    Taker               Account  `json:"taker"`
    CurrentPrice        string   `json:"current_price"`
    CurrentBounty       string   `json:"current_bounty"`
    BountyMultiple      string   `json:"bounty_multiple"`
    MakerRelayerFee     string   `json:"maker_relayer_fee"`
    TakerRelayerFee     string   `json:"taker_relayer_fee"`
    MakerProtocolFee    string   `json:"maker_protocol_fee"`
    TakerProtocolFee    string   `json:"taker_protocol_fee"`
    MakerReferrerFee    string   `json:"maker_referrer_fee"`
    FeeRecipient        Account  `json:"fee_recipient"`
    FeeMethod           int      `json:"fee_method"`
    Side                int      `json:"side"`
    SaleKind            int      `json:"sale_kind"`
    Target              string   `json:"target"`
    HowToCall           int      `json:"how_to_call"`
    Calldata            string   `json:"calldata"`
    ReplacementPattern  string   `json:"replacement_pattern"`
    StaticTarget        string   `json:"static_target"`
    StaticExtradata     string   `json:"static_extradata"`
    PaymentToken        string   `json:"payment_token"`
    PaymentTokenContract PaymentTokenContract `json:"payment_token_contract"`
    BasePrice           string   `json:"base_price"`
    Extra               string   `json:"extra"`
    Quantity            string   `json:"quantity"`
    Salt                string   `json:"salt"`
    V                   int      `json:"v"`
    R                   string   `json:"r"`
    S                   string   `json:"s"`
    ApprovedOnChain     bool     `json:"approved_on_chain"`
    Cancelled           bool     `json:"cancelled"`
    Finalized           bool     `json:"finalized"`
    MarkedInvalid       bool     `json:"marked_invalid"`
    PrefixedHash        string   `json:"prefixed_hash"`
}

type Metadata struct {
    Asset  Asset  `json:"asset"`
    Schema string `json:"schema"`
}

type PaymentTokenContract struct {
    ID       int    `json:"id"`
    Symbol   string `json:"symbol"`
    Address  string `json:"address"`
    ImageURL string `json:"image_url"`
    Name     string `json:"name"`
    Decimals int    `json:"decimals"`
    ETHPrice string `json:"eth_price"`
    USDPrice string `json:"usd_price"`
}

type Trait struct {
    TraitType   string      `json:"trait_type"`
    Value       interface{} `json:"value"`
    DisplayType string      `json:"display_type"`
    MaxValue    interface{} `json:"max_value"`
    TraitCount  int         `json:"trait_count"`
    Order       interface{} `json:"order"`
}

type Sale struct {
    Asset               Asset                `json:"asset"`
    AssetBundle         interface{}          `json:"asset_bundle"`
    EventType           string               `json:"event_type"`
    EventTimestamp      string               `json:"event_timestamp"`
    AuctionType         string               `json:"auction_type"`
    TotalPrice          string               `json:"total_price"`
    PaymentToken        PaymentTokenContract `json:"payment_token"`
    Transaction         Transaction          `json:"transaction"`
    CreatedDate         string               `json:"created_date"`
    Quantity            string               `json:"quantity"`
}

type Transaction struct {
    BlockHash        string `json:"block_hash"`
    BlockNumber      string `json:"block_number"`
    FromAccount      Account `json:"from_account"`
    ID               int    `json:"id"`
    Timestamp        string `json:"timestamp"`
    ToAccount        Account `json:"to_account"`
    TransactionHash  string `json:"transaction_hash"`
    TransactionIndex string `json:"transaction_index"`
}

type Payment struct {
    Token    PaymentTokenContract `json:"token"`
    TokenID  string               `json:"token_id"`
    Quantity string               `json:"quantity"`
}

type EventsResponse struct {
    AssetEvents []AssetEvent `json:"asset_events"`
}

type AssetEvent struct {
    Asset           Asset                `json:"asset"`
    AssetBundle     interface{}          `json:"asset_bundle"`
    EventType       string               `json:"event_type"`
    EventTimestamp  string               `json:"event_timestamp"`
    AuctionType     string               `json:"auction_type"`
    TotalPrice      string               `json:"total_price"`
    PaymentToken    PaymentTokenContract `json:"payment_token"`
    Transaction     Transaction          `json:"transaction"`
    CreatedDate     string               `json:"created_date"`
    Quantity        string               `json:"quantity"`
    ApprovedAccount Account              `json:"approved_account"`
    BidAmount       string               `json:"bid_amount"`
    CollectionSlug  string               `json:"collection_slug"`
    ContractAddress string               `json:"contract_address"`
    CustomEventName string               `json:"custom_event_name"`
    DevFeePayment   Payment              `json:"dev_fee_payment"`
    DevSellerFee    string               `json:"dev_seller_fee"`
    Duration        string               `json:"duration"`
    EndingPrice     string               `json:"ending_price"`
    FromAccount     Account              `json:"from_account"`
    ID              int                  `json:"id"`
    IsPrivate       bool                 `json:"is_private"`
    OwnerAccount    Account              `json:"owner_account"`
    Seller          Account              `json:"seller"`
    StartingPrice   string               `json:"starting_price"`
    ToAccount       Account              `json:"to_account"`
    WinnerAccount   Account              `json:"winner_account"`
}
```