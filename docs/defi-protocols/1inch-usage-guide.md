# 1inch 聚合器 详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [环境准备](#环境准备)
3. [API集成](#api集成)
4. [交易聚合](#交易聚合)
5. [流动性挖矿](#流动性挖矿)
6. [价格查询](#价格查询)
7. [交易执行](#交易执行)
8. [最佳实践](#最佳实践)
9. [实际应用](#实际应用)

## 基础概念

### 1.1 1inch 聚合器简介

1inch 是领先的 DEX 聚合器，通过智能路由算法在多个去中心化交易所中寻找最优交易路径，为用户提供最佳价格和最低滑点。

```bash
# 安装1inch相关依赖
go get github.com/ethereum/go-ethereum
go get github.com/shopspring/decimal
go get github.com/gorilla/websocket
go get github.com/gin-gonic/gin
```

### 1.2 核心功能

```go
// 主要包导入
import (
    "context"
    "encoding/json"
    "fmt"
    "math/big"
    "net/http"
    "time"
    
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/shopspring/decimal"
)

// 1inch API 基础配置
type OneInchConfig struct {
    APIKey      string
    BaseURL     string
    ChainID     int
    HTTPTimeout time.Duration
}

// 支持的网络
const (
    EthereumMainnet = 1
    BSCMainnet      = 56
    PolygonMainnet  = 137
    ArbitrumOne     = 42161
    OptimismMainnet = 10
    AvalancheC      = 43114
)
```

## 环境准备

### 2.1 API密钥配置

```go
// config/oneinch.go
package config

import (
    "os"
    "time"
)

func NewOneInchConfig() *OneInchConfig {
    return &OneInchConfig{
        APIKey:      os.Getenv("ONEINCH_API_KEY"), // 从环境变量获取
        BaseURL:     "https://api.1inch.dev",
        ChainID:     EthereumMainnet,
        HTTPTimeout: 30 * time.Second,
    }
}

// 获取API密钥的方法
func GetAPIKey() string {
    // 1. 访问 https://portal.1inch.dev/
    // 2. 注册账户并创建API密钥
    // 3. 设置环境变量 ONEINCH_API_KEY
    return os.Getenv("ONEINCH_API_KEY")
}
```

### 2.2 HTTP客户端设置

```go
// client/oneinch_client.go
package client

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

type OneInchClient struct {
    config     *OneInchConfig
    httpClient *http.Client
}

func NewOneInchClient(config *OneInchConfig) *OneInchClient {
    return &OneInchClient{
        config: config,
        httpClient: &http.Client{
            Timeout: config.HTTPTimeout,
        },
    }
}

func (c *OneInchClient) makeRequest(method, endpoint string, params map[string]string) ([]byte, error) {
    url := fmt.Sprintf("%s/swap/v6.0/%d%s", c.config.BaseURL, c.config.ChainID, endpoint)
    
    req, err := http.NewRequest(method, url, nil)
    if err != nil {
        return nil, err
    }
    
    // 设置请求头
    req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
    req.Header.Set("Content-Type", "application/json")
    
    // 添加查询参数
    q := req.URL.Query()
    for key, value := range params {
        q.Add(key, value)
    }
    req.URL.RawQuery = q.Encode()
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API请求失败: %d - %s", resp.StatusCode, string(body))
    }
    
    return body, nil
}
```

## API集成

### 3.1 支持的代币查询

```go
// models/token.go
package models

type Token struct {
    Symbol   string          `json:"symbol"`
    Name     string          `json:"name"`
    Address  string          `json:"address"`
    Decimals int             `json:"decimals"`
    LogoURI  string          `json:"logoURI"`
    Tags     []string        `json:"tags"`
}

type TokensResponse struct {
    Tokens map[string]Token `json:"tokens"`
}

// 获取支持的代币列表
func (c *OneInchClient) GetSupportedTokens() (*TokensResponse, error) {
    data, err := c.makeRequest("GET", "/tokens", nil)
    if err != nil {
        return nil, err
    }
    
    var response TokensResponse
    if err := json.Unmarshal(data, &response); err != nil {
        return nil, err
    }
    
    return &response, nil
}

// 获取特定代币信息
func (c *OneInchClient) GetTokenInfo(tokenAddress string) (*Token, error) {
    tokens, err := c.GetSupportedTokens()
    if err != nil {
        return nil, err
    }
    
    for _, token := range tokens.Tokens {
        if token.Address == tokenAddress {
            return &token, nil
        }
    }
    
    return nil, fmt.Errorf("代币未找到: %s", tokenAddress)
}
```

### 3.2 流动性来源查询

```go
// models/liquidity.go
package models

type LiquiditySource struct {
    ID    string `json:"id"`
    Title string `json:"title"`
    Image string `json:"image"`
}

type LiquiditySourcesResponse struct {
    Protocols []LiquiditySource `json:"protocols"`
}

// 获取流动性来源
func (c *OneInchClient) GetLiquiditySources() (*LiquiditySourcesResponse, error) {
    data, err := c.makeRequest("GET", "/liquidity-sources", nil)
    if err != nil {
        return nil, err
    }
    
    var response LiquiditySourcesResponse
    if err := json.Unmarshal(data, &response); err != nil {
        return nil, err
    }
    
    return &response, nil
}
```

## 交易聚合

### 4.1 价格查询

```go
// models/quote.go
package models

type QuoteRequest struct {
    Src           string `json:"src"`           // 源代币地址
    Dst           string `json:"dst"`           // 目标代币地址
    Amount        string `json:"amount"`        // 交易数量
    Fee           string `json:"fee,omitempty"` // 手续费百分比
    Protocols     string `json:"protocols,omitempty"`
    GasLimit      string `json:"gasLimit,omitempty"`
    ConnectorTokens string `json:"connectorTokens,omitempty"`
}

type QuoteResponse struct {
    Src           Token           `json:"src"`
    Dst           Token           `json:"dst"`
    SrcAmount     string          `json:"srcAmount"`
    DstAmount     string          `json:"dstAmount"`
    Protocols     [][]Protocol    `json:"protocols"`
    EstimatedGas  string          `json:"estimatedGas"`
}

type Protocol struct {
    Name         string `json:"name"`
    Part         int    `json:"part"`
    FromTokenAddress string `json:"fromTokenAddress"`
    ToTokenAddress   string `json:"toTokenAddress"`
}

// 获取交易报价
func (c *OneInchClient) GetQuote(req *QuoteRequest) (*QuoteResponse, error) {
    params := map[string]string{
        "src":    req.Src,
        "dst":    req.Dst,
        "amount": req.Amount,
    }
    
    if req.Fee != "" {
        params["fee"] = req.Fee
    }
    if req.Protocols != "" {
        params["protocols"] = req.Protocols
    }
    if req.GasLimit != "" {
        params["gasLimit"] = req.GasLimit
    }
    
    data, err := c.makeRequest("GET", "/quote", params)
    if err != nil {
        return nil, err
    }
    
    var response QuoteResponse
    if err := json.Unmarshal(data, &response); err != nil {
        return nil, err
    }
    
    return &response, nil
}
```

### 4.2 交易构建

```go
// models/swap.go
package models

type SwapRequest struct {
    Src              string `json:"src"`
    Dst              string `json:"dst"`
    Amount           string `json:"amount"`
    From             string `json:"from"`             // 发送者地址
    Slippage         string `json:"slippage"`         // 滑点容忍度
    Protocols        string `json:"protocols,omitempty"`
    Fee              string `json:"fee,omitempty"`
    GasLimit         string `json:"gasLimit,omitempty"`
    GasPrice         string `json:"gasPrice,omitempty"`
    ConnectorTokens  string `json:"connectorTokens,omitempty"`
    AllowPartialFill bool   `json:"allowPartialFill,omitempty"`
    DisableEstimate  bool   `json:"disableEstimate,omitempty"`
    UsePatching      bool   `json:"usePatching,omitempty"`
}

type SwapResponse struct {
    Src           Token  `json:"src"`
    Dst           Token  `json:"dst"`
    SrcAmount     string `json:"srcAmount"`
    DstAmount     string `json:"dstAmount"`
    Protocols     [][]Protocol `json:"protocols"`
    Tx            TransactionData `json:"tx"`
}

type TransactionData struct {
    From     string `json:"from"`
    To       string `json:"to"`
    Data     string `json:"data"`
    Value    string `json:"value"`
    GasPrice string `json:"gasPrice"`
    Gas      string `json:"gas"`
}

// 构建交易
func (c *OneInchClient) BuildSwap(req *SwapRequest) (*SwapResponse, error) {
    params := map[string]string{
        "src":      req.Src,
        "dst":      req.Dst,
        "amount":   req.Amount,
        "from":     req.From,
        "slippage": req.Slippage,
    }
    
    // 添加可选参数
    if req.Protocols != "" {
        params["protocols"] = req.Protocols
    }
    if req.Fee != "" {
        params["fee"] = req.Fee
    }
    if req.GasLimit != "" {
        params["gasLimit"] = req.GasLimit
    }
    if req.GasPrice != "" {
        params["gasPrice"] = req.GasPrice
    }
    
    data, err := c.makeRequest("GET", "/swap", params)
    if err != nil {
        return nil, err
    }
    
    var response SwapResponse
    if err := json.Unmarshal(data, &response); err != nil {
        return nil, err
    }
    
    return &response, nil
}
```

## 流动性挖矿

### 5.1 流动性池查询

```go
// models/liquidity_pool.go
package models

type LiquidityPool struct {
    Address     string          `json:"address"`
    Token0      Token           `json:"token0"`
    Token1      Token           `json:"token1"`
    Fee         string          `json:"fee"`
    APY         decimal.Decimal `json:"apy"`
    TVL         decimal.Decimal `json:"tvl"`
    Volume24h   decimal.Decimal `json:"volume24h"`
}

type LiquidityPoolsResponse struct {
    Pools []LiquidityPool `json:"pools"`
}

// 获取流动性池信息
func (c *OneInchClient) GetLiquidityPools() (*LiquidityPoolsResponse, error) {
    data, err := c.makeRequest("GET", "/liquidity-pools", nil)
    if err != nil {
        return nil, err
    }
    
    var response LiquidityPoolsResponse
    if err := json.Unmarshal(data, &response); err != nil {
        return nil, err
    }
    
    return &response, nil
}
```

### 5.2 收益农场

```go
// models/farming.go
package models

type FarmingPool struct {
    PoolAddress   string          `json:"poolAddress"`
    RewardToken   Token           `json:"rewardToken"`
    StakeToken    Token           `json:"stakeToken"`
    APR           decimal.Decimal `json:"apr"`
    TotalStaked   decimal.Decimal `json:"totalStaked"`
    RewardRate    decimal.Decimal `json:"rewardRate"`
    PeriodFinish  int64           `json:"periodFinish"`
}

type FarmingPoolsResponse struct {
    Farms []FarmingPool `json:"farms"`
}

// 获取收益农场信息
func (c *OneInchClient) GetFarmingPools() (*FarmingPoolsResponse, error) {
    data, err := c.makeRequest("GET", "/farming", nil)
    if err != nil {
        return nil, err
    }
    
    var response FarmingPoolsResponse
    if err := json.Unmarshal(data, &response); err != nil {
        return nil, err
    }
    
    return &response, nil
}
```

## 价格查询

### 6.1 实时价格获取

```go
// services/price_service.go
package services

import (
    "fmt"
    "math/big"
    "strconv"
    
    "github.com/shopspring/decimal"
)

type PriceService struct {
    client *OneInchClient
}

func NewPriceService(client *OneInchClient) *PriceService {
    return &PriceService{
        client: client,
    }
}

// 获取代币价格（以ETH计价）
func (s *PriceService) GetTokenPrice(tokenAddress string, amount *big.Int) (*decimal.Decimal, error) {
    // ETH地址
    ethAddress := "0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE"
    
    quote, err := s.client.GetQuote(&QuoteRequest{
        Src:    tokenAddress,
        Dst:    ethAddress,
        Amount: amount.String(),
    })
    if err != nil {
        return nil, err
    }
    
    // 计算价格
    srcAmount, err := decimal.NewFromString(quote.SrcAmount)
    if err != nil {
        return nil, err
    }
    
    dstAmount, err := decimal.NewFromString(quote.DstAmount)
    if err != nil {
        return nil, err
    }
    
    price := dstAmount.Div(srcAmount)
    return &price, nil
}

// 获取最佳交易路径
func (s *PriceService) GetBestRoute(srcToken, dstToken string, amount *big.Int) (*QuoteResponse, error) {
    return s.client.GetQuote(&QuoteRequest{
        Src:    srcToken,
        Dst:    dstToken,
        Amount: amount.String(),
    })
}

// 比较多个DEX的价格
func (s *PriceService) ComparePrices(srcToken, dstToken string, amount *big.Int) (map[string]*decimal.Decimal, error) {
    prices := make(map[string]*decimal.Decimal)
    
    // 获取1inch聚合价格
    quote, err := s.GetBestRoute(srcToken, dstToken, amount)
    if err != nil {
        return nil, err
    }
    
    srcAmount, _ := decimal.NewFromString(quote.SrcAmount)
    dstAmount, _ := decimal.NewFromString(quote.DstAmount)
    oneInchPrice := dstAmount.Div(srcAmount)
    prices["1inch"] = &oneInchPrice
    
    return prices, nil
}
```

## 交易执行

### 7.1 交易管理器

```go
// services/swap_service.go
package services

import (
    "context"
    "crypto/ecdsa"
    "math/big"
    
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/ethclient"
)

type SwapService struct {
    client    *OneInchClient
    ethClient *ethclient.Client
    privateKey *ecdsa.PrivateKey
}

func NewSwapService(client *OneInchClient, ethClient *ethclient.Client, privateKey *ecdsa.PrivateKey) *SwapService {
    return &SwapService{
        client:     client,
        ethClient:  ethClient,
        privateKey: privateKey,
    }
}

// 执行代币交换
func (s *SwapService) ExecuteSwap(srcToken, dstToken string, amount *big.Int, slippage float64) (*types.Transaction, error) {
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 构建交换请求
    swapReq := &SwapRequest{
        Src:      srcToken,
        Dst:      dstToken,
        Amount:   amount.String(),
        From:     fromAddress.Hex(),
        Slippage: fmt.Sprintf("%.1f", slippage),
    }
    
    // 获取交换数据
    swapResp, err := s.client.BuildSwap(swapReq)
    if err != nil {
        return nil, err
    }
    
    // 构建交易
    nonce, err := s.ethClient.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        return nil, err
    }
    
    gasPrice, err := s.ethClient.SuggestGasPrice(context.Background())
    if err != nil {
        return nil, err
    }
    
    gasLimit, _ := strconv.ParseUint(swapResp.Tx.Gas, 10, 64)
    value, _ := new(big.Int).SetString(swapResp.Tx.Value, 10)
    toAddress := common.HexToAddress(swapResp.Tx.To)
    data := common.FromHex(swapResp.Tx.Data)
    
    tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)
    
    // 签名交易
    chainID, err := s.ethClient.NetworkID(context.Background())
    if err != nil {
        return nil, err
    }
    
    signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), s.privateKey)
    if err != nil {
        return nil, err
    }
    
    // 发送交易
    err = s.ethClient.SendTransaction(context.Background(), signedTx)
    if err != nil {
        return nil, err
    }
    
    return signedTx, nil
}

// 检查交易状态
func (s *SwapService) CheckTransactionStatus(txHash common.Hash) (*types.Receipt, error) {
    return s.ethClient.TransactionReceipt(context.Background(), txHash)
}
```

## 最佳实践

### 8.1 错误处理和重试

```go
// utils/retry.go
package utils

import (
    "time"
)

type RetryConfig struct {
    MaxRetries int
    Delay      time.Duration
    Backoff    float64
}

func RetryWithBackoff(fn func() error, config RetryConfig) error {
    var err error
    delay := config.Delay
    
    for i := 0; i <= config.MaxRetries; i++ {
        err = fn()
        if err == nil {
            return nil
        }
        
        if i < config.MaxRetries {
            time.Sleep(delay)
            delay = time.Duration(float64(delay) * config.Backoff)
        }
    }
    
    return err
}

// 使用示例
func (s *SwapService) GetQuoteWithRetry(req *QuoteRequest) (*QuoteResponse, error) {
    var result *QuoteResponse
    var err error
    
    retryErr := RetryWithBackoff(func() error {
        result, err = s.client.GetQuote(req)
        return err
    }, RetryConfig{
        MaxRetries: 3,
        Delay:      time.Second,
        Backoff:    2.0,
    })
    
    if retryErr != nil {
        return nil, retryErr
    }
    
    return result, nil
}
```

### 8.2 滑点保护

```go
// utils/slippage.go
package utils

import (
    "math/big"
    
    "github.com/shopspring/decimal"
)

// 计算最小输出金额
func CalculateMinOutput(expectedOutput *big.Int, slippageTolerance float64) *big.Int {
    expected := decimal.NewFromBigInt(expectedOutput, 0)
    tolerance := decimal.NewFromFloat(slippageTolerance / 100)
    minOutput := expected.Mul(decimal.NewFromFloat(1).Sub(tolerance))
    
    result, _ := minOutput.BigInt()
    return result
}

// 验证滑点
func ValidateSlippage(expectedOutput, actualOutput *big.Int, maxSlippage float64) bool {
    expected := decimal.NewFromBigInt(expectedOutput, 0)
    actual := decimal.NewFromBigInt(actualOutput, 0)
    
    if actual.GreaterThanOrEqual(expected) {
        return true
    }
    
    slippage := expected.Sub(actual).Div(expected).Mul(decimal.NewFromInt(100))
    return slippage.LessThanOrEqual(decimal.NewFromFloat(maxSlippage))
}
```

## 实际应用

### 9.1 完整交易示例

```go
// main.go
package main

import (
    "fmt"
    "log"
    "math/big"
    
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/ethclient"
    
    "your-project/client"
    "your-project/config"
    "your-project/services"
)

func main() {
    // 初始化配置
    cfg := config.NewOneInchConfig()
    
    // 创建1inch客户端
    oneInchClient := client.NewOneInchClient(cfg)
    
    // 连接以太坊节点
    ethClient, err := ethclient.Dial("https://mainnet.infura.io/v3/YOUR_PROJECT_ID")
    if err != nil {
        log.Fatal("连接以太坊节点失败:", err)
    }
    
    // 加载私钥
    privateKey, err := crypto.HexToECDSA("your_private_key_here")
    if err != nil {
        log.Fatal("加载私钥失败:", err)
    }
    
    // 创建服务
    swapService := services.NewSwapService(oneInchClient, ethClient, privateKey)
    priceService := services.NewPriceService(oneInchClient)
    
    // 代币地址
    usdcAddress := "0xA0b86a33E6417c8f4c8c8c8c8c8c8c8c8c8c8c8c"
    daiAddress := "0x6B175474E89094C44Da98b954EedeAC495271d0F"
    
    // 交易金额 (1000 USDC)
    amount := new(big.Int).Mul(big.NewInt(1000), big.NewInt(1e6))
    
    // 获取报价
    quote, err := priceService.GetBestRoute(usdcAddress, daiAddress, amount)
    if err != nil {
        log.Fatal("获取报价失败:", err)
    }
    
    fmt.Printf("交换 %s USDC 可获得 %s DAI\n", quote.SrcAmount, quote.DstAmount)
    
    // 执行交换
    tx, err := swapService.ExecuteSwap(usdcAddress, daiAddress, amount, 1.0) // 1%滑点
    if err != nil {
        log.Fatal("执行交换失败:", err)
    }
    
    fmt.Printf("交易已提交: %s\n", tx.Hash().Hex())
    
    // 等待交易确认
    receipt, err := swapService.CheckTransactionStatus(tx.Hash())
    if err != nil {
        log.Fatal("检查交易状态失败:", err)
    }
    
    if receipt.Status == 1 {
        fmt.Println("交易成功!")
    } else {
        fmt.Println("交易失败!")
    }
}
```

### 9.2 价格监控服务

```go
// services/price_monitor.go
package services

import (
    "context"
    "fmt"
    "math/big"
    "time"
    
    "github.com/shopspring/decimal"
)

type PriceMonitor struct {
    priceService *PriceService
    alerts       map[string]*PriceAlert
}

type PriceAlert struct {
    TokenPair    string
    TargetPrice  decimal.Decimal
    CurrentPrice decimal.Decimal
    Triggered    bool
    Callback     func(alert *PriceAlert)
}

func NewPriceMonitor(priceService *PriceService) *PriceMonitor {
    return &PriceMonitor{
        priceService: priceService,
        alerts:       make(map[string]*PriceAlert),
    }
}

// 添加价格警报
func (pm *PriceMonitor) AddAlert(id, srcToken, dstToken string, targetPrice decimal.Decimal, callback func(*PriceAlert)) {
    pm.alerts[id] = &PriceAlert{
        TokenPair:   fmt.Sprintf("%s/%s", srcToken, dstToken),
        TargetPrice: targetPrice,
        Triggered:   false,
        Callback:    callback,
    }
}

// 开始监控
func (pm *PriceMonitor) Start(ctx context.Context, interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            pm.checkAlerts()
        }
    }
}

func (pm *PriceMonitor) checkAlerts() {
    for id, alert := range pm.alerts {
        if alert.Triggered {
            continue
        }
        
        // 这里需要解析token地址并获取价格
        // 简化示例，实际需要更复杂的逻辑
        
        if alert.CurrentPrice.GreaterThanOrEqual(alert.TargetPrice) {
            alert.Triggered = true
            if alert.Callback != nil {
                alert.Callback(alert)
            }
        }
    }
}
```

这个1inch使用指南涵盖了从基础配置到高级功能的完整实现，包括API集成、交易执行、价格监控等核心功能。通过这些示例，开发者可以快速集成1inch聚合器到自己的DeFi应用中。
