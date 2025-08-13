# dYdX 协议 详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [环境准备](#环境准备)
3. [现货交易](#现货交易)
4. [永续合约](#永续合约)
5. [杠杆交易](#杠杆交易)
6. [流动性挖矿](#流动性挖矿)
7. [治理代币](#治理代币)
8. [风险管理](#风险管理)
9. [实际应用](#实际应用)

## 基础概念

### 1.1 dYdX 简介

dYdX 是领先的去中心化衍生品交易所，提供现货交易、永续合约、杠杆交易等功能，是 DeFi 衍生品领域的重要基础设施。

```bash
# 安装dYdX相关依赖
go get github.com/ethereum/go-ethereum
go get github.com/ethereum/go-ethereum/accounts/abi
go get github.com/ethereum/go-ethereum/accounts/abi/bind
go get github.com/shopspring/decimal
go get github.com/gorilla/websocket
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
    "github.com/gorilla/websocket"
)

// dYdX 核心合约地址 (Mainnet)
var (
    // Solo Margin (V1)
    SoloMarginAddress      = common.HexToAddress("0x1E0447b19BB6EcFdAe1e4AE1694b0C3659614e4e")
    
    // Perpetual (V2)
    PerpetualAddress       = common.HexToAddress("0xD54f502e184B6B739d7D27a6410a67dc462D69c8")
    
    // DYDX Token
    DYDXTokenAddress       = common.HexToAddress("0x92D6C1e31e14520e676a687F0a93788B716BEff5")
    
    // Safety Module
    SafetyModuleAddress    = common.HexToAddress("0x65f7BA4Ec257AF7c55fd5854E5f6356bBd0fb8EC")
    
    // Liquidity Staking
    LiquidityStakingAddress = common.HexToAddress("0x5Aa653A076c1dbB47cec8C1B4d152444CAD91941")
    
    // Merkle Distributor
    MerkleDistributorAddress = common.HexToAddress("0x64c7d40c07EFAbec2AafdC243bF59eaF2195c6dc")
)

// API端点
const (
    MainnetAPIURL = "https://api.dydx.exchange"
    TestnetAPIURL = "https://api.stage.dydx.exchange"
    WebSocketURL  = "wss://api.dydx.exchange/v3/ws"
)

// 市场信息
type Market struct {
    Market                string          `json:"market"`
    Status                string          `json:"status"`
    BaseAsset             string          `json:"baseAsset"`
    QuoteAsset            string          `json:"quoteAsset"`
    StepSize              decimal.Decimal `json:"stepSize"`
    TickSize              decimal.Decimal `json:"tickSize"`
    IndexPrice            decimal.Decimal `json:"indexPrice"`
    OraclePrice           decimal.Decimal `json:"oraclePrice"`
    PriceChange24H        decimal.Decimal `json:"priceChange24H"`
    NextFundingRate       decimal.Decimal `json:"nextFundingRate"`
    NextFundingAt         time.Time       `json:"nextFundingAt"`
    MinOrderSize          decimal.Decimal `json:"minOrderSize"`
    Type                  string          `json:"type"`
    InitialMarginFraction decimal.Decimal `json:"initialMarginFraction"`
    MaintenanceMarginFraction decimal.Decimal `json:"maintenanceMarginFraction"`
}

// 订单信息
type Order struct {
    ID               string          `json:"id"`
    ClientID         string          `json:"clientId"`
    AccountID        string          `json:"accountId"`
    Market           string          `json:"market"`
    Side             string          `json:"side"`
    Size             decimal.Decimal `json:"size"`
    RemainingSize    decimal.Decimal `json:"remainingSize"`
    Price            decimal.Decimal `json:"price"`
    TriggerPrice     decimal.Decimal `json:"triggerPrice,omitempty"`
    TrailingPercent  decimal.Decimal `json:"trailingPercent,omitempty"`
    Type             string          `json:"type"`
    Status           string          `json:"status"`
    TimeInForce      string          `json:"timeInForce"`
    PostOnly         bool            `json:"postOnly"`
    ReduceOnly       bool            `json:"reduceOnly"`
    CreatedAt        time.Time       `json:"createdAt"`
    UnfillableAt     time.Time       `json:"unfillableAt,omitempty"`
    ExpiresAt        time.Time       `json:"expiresAt,omitempty"`
}

// 持仓信息
type Position struct {
    Market          string          `json:"market"`
    Status          string          `json:"status"`
    Side            string          `json:"side"`
    Size            decimal.Decimal `json:"size"`
    MaxSize         decimal.Decimal `json:"maxSize"`
    EntryPrice      decimal.Decimal `json:"entryPrice"`
    ExitPrice       decimal.Decimal `json:"exitPrice,omitempty"`
    UnrealizedPnl   decimal.Decimal `json:"unrealizedPnl"`
    RealizedPnl     decimal.Decimal `json:"realizedPnl"`
    CreatedAt       time.Time       `json:"createdAt"`
    ClosedAt        time.Time       `json:"closedAt,omitempty"`
    SumOpen         decimal.Decimal `json:"sumOpen"`
    SumClose        decimal.Decimal `json:"sumClose"`
    NetFunding      decimal.Decimal `json:"netFunding"`
}

// 账户信息
type Account struct {
    ID                    string          `json:"id"`
    AccountNumber         string          `json:"accountNumber"`
    StarkKey              string          `json:"starkKey"`
    PositionID            string          `json:"positionId"`
    Equity                decimal.Decimal `json:"equity"`
    FreeCollateral        decimal.Decimal `json:"freeCollateral"`
    PendingDeposits       decimal.Decimal `json:"pendingDeposits"`
    PendingWithdrawals    decimal.Decimal `json:"pendingWithdrawals"`
    OpenPositions         map[string]*Position `json:"openPositions"`
    AccountNumber         string          `json:"accountNumber"`
    QuoteBalance          decimal.Decimal `json:"quoteBalance"`
    TakerFeeRate          decimal.Decimal `json:"takerFeeRate"`
    MakerFeeRate          decimal.Decimal `json:"makerFeeRate"`
    TakerVolumeShare      decimal.Decimal `json:"takerVolumeShare"`
    MakerVolumeShare      decimal.Decimal `json:"makerVolumeShare"`
}
```

## 环境准备

### 2.1 API客户端设置

```go
// client/dydx_client.go
package client

import (
    "bytes"
    "crypto/hmac"
    "crypto/sha256"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "strconv"
    "time"
    
    "github.com/shopspring/decimal"
)

type DYDXClient struct {
    baseURL    string
    apiKey     string
    apiSecret  string
    passphrase string
    httpClient *http.Client
}

func NewDYDXClient(apiKey, apiSecret, passphrase string, isMainnet bool) *DYDXClient {
    baseURL := TestnetAPIURL
    if isMainnet {
        baseURL = MainnetAPIURL
    }
    
    return &DYDXClient{
        baseURL:    baseURL,
        apiKey:     apiKey,
        apiSecret:  apiSecret,
        passphrase: passphrase,
        httpClient: &http.Client{Timeout: 30 * time.Second},
    }
}

// 生成签名
func (c *DYDXClient) generateSignature(timestamp, method, requestPath, body string) string {
    message := timestamp + method + requestPath + body
    h := hmac.New(sha256.New, []byte(c.apiSecret))
    h.Write([]byte(message))
    return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// 发送HTTP请求
func (c *DYDXClient) sendRequest(method, endpoint string, params interface{}) ([]byte, error) {
    var body []byte
    var err error
    
    if params != nil {
        body, err = json.Marshal(params)
        if err != nil {
            return nil, err
        }
    }
    
    url := c.baseURL + endpoint
    req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
    if err != nil {
        return nil, err
    }
    
    // 添加认证头
    timestamp := strconv.FormatInt(time.Now().Unix(), 10)
    signature := c.generateSignature(timestamp, method, endpoint, string(body))
    
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("DYDX-API-KEY", c.apiKey)
    req.Header.Set("DYDX-SIGNATURE", signature)
    req.Header.Set("DYDX-TIMESTAMP", timestamp)
    req.Header.Set("DYDX-PASSPHRASE", c.passphrase)
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    respBody, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API请求失败: %s", string(respBody))
    }
    
    return respBody, nil
}

// 获取市场信息
func (c *DYDXClient) GetMarkets() (map[string]*Market, error) {
    respBody, err := c.sendRequest("GET", "/v3/markets", nil)
    if err != nil {
        return nil, err
    }
    
    var response struct {
        Markets map[string]*Market `json:"markets"`
    }
    
    if err := json.Unmarshal(respBody, &response); err != nil {
        return nil, err
    }
    
    return response.Markets, nil
}

// 获取账户信息
func (c *DYDXClient) GetAccount(accountID string) (*Account, error) {
    endpoint := fmt.Sprintf("/v3/accounts/%s", accountID)
    respBody, err := c.sendRequest("GET", endpoint, nil)
    if err != nil {
        return nil, err
    }
    
    var response struct {
        Account *Account `json:"account"`
    }
    
    if err := json.Unmarshal(respBody, &response); err != nil {
        return nil, err
    }
    
    return response.Account, nil
}

// 获取持仓信息
func (c *DYDXClient) GetPositions(accountID string) (map[string]*Position, error) {
    endpoint := fmt.Sprintf("/v3/positions?accountId=%s", accountID)
    respBody, err := c.sendRequest("GET", endpoint, nil)
    if err != nil {
        return nil, err
    }
    
    var response struct {
        Positions map[string]*Position `json:"positions"`
    }
    
    if err := json.Unmarshal(respBody, &response); err != nil {
        return nil, err
    }
    
    return response.Positions, nil
}
```

## 现货交易

### 3.1 现货交易服务

```go
// services/spot_trading_service.go
package services

import (
    "encoding/json"
    "fmt"
    "time"
    
    "github.com/shopspring/decimal"
)

type SpotTradingService struct {
    client *DYDXClient
}

func NewSpotTradingService(client *DYDXClient) *SpotTradingService {
    return &SpotTradingService{
        client: client,
    }
}

// 下单参数
type PlaceOrderParams struct {
    Market       string          `json:"market"`
    Side         string          `json:"side"`         // BUY, SELL
    Type         string          `json:"type"`         // MARKET, LIMIT, STOP_LIMIT, TAKE_PROFIT
    Size         decimal.Decimal `json:"size"`
    Price        decimal.Decimal `json:"price,omitempty"`
    TriggerPrice decimal.Decimal `json:"triggerPrice,omitempty"`
    TimeInForce  string          `json:"timeInForce"`  // GTT, FOK, IOC
    PostOnly     bool            `json:"postOnly"`
    ReduceOnly   bool            `json:"reduceOnly"`
    ExpiresAt    *time.Time      `json:"expiresAt,omitempty"`
    ClientID     string          `json:"clientId,omitempty"`
}

// 下单
func (s *SpotTradingService) PlaceOrder(params *PlaceOrderParams) (*Order, error) {
    // 生成客户端ID
    if params.ClientID == "" {
        params.ClientID = fmt.Sprintf("order_%d", time.Now().UnixNano())
    }
    
    respBody, err := s.client.sendRequest("POST", "/v3/orders", params)
    if err != nil {
        return nil, err
    }
    
    var response struct {
        Order *Order `json:"order"`
    }
    
    if err := json.Unmarshal(respBody, &response); err != nil {
        return nil, err
    }
    
    return response.Order, nil
}

// 取消订单
func (s *SpotTradingService) CancelOrder(orderID string) (*Order, error) {
    endpoint := fmt.Sprintf("/v3/orders/%s", orderID)
    respBody, err := s.client.sendRequest("DELETE", endpoint, nil)
    if err != nil {
        return nil, err
    }
    
    var response struct {
        CancelOrder *Order `json:"cancelOrder"`
    }
    
    if err := json.Unmarshal(respBody, &response); err != nil {
        return nil, err
    }
    
    return response.CancelOrder, nil
}

// 取消所有订单
func (s *SpotTradingService) CancelAllOrders(market string) ([]Order, error) {
    params := map[string]string{}
    if market != "" {
        params["market"] = market
    }
    
    respBody, err := s.client.sendRequest("DELETE", "/v3/orders", params)
    if err != nil {
        return nil, err
    }
    
    var response struct {
        CancelOrders []Order `json:"cancelOrders"`
    }
    
    if err := json.Unmarshal(respBody, &response); err != nil {
        return nil, err
    }
    
    return response.CancelOrders, nil
}

// 获取订单列表
func (s *SpotTradingService) GetOrders(accountID, market, status string) ([]Order, error) {
    endpoint := fmt.Sprintf("/v3/orders?accountId=%s", accountID)
    if market != "" {
        endpoint += "&market=" + market
    }
    if status != "" {
        endpoint += "&status=" + status
    }
    
    respBody, err := s.client.sendRequest("GET", endpoint, nil)
    if err != nil {
        return nil, err
    }
    
    var response struct {
        Orders []Order `json:"orders"`
    }
    
    if err := json.Unmarshal(respBody, &response); err != nil {
        return nil, err
    }
    
    return response.Orders, nil
}

// 市价买入
func (s *SpotTradingService) MarketBuy(market string, size decimal.Decimal) (*Order, error) {
    params := &PlaceOrderParams{
        Market:      market,
        Side:        "BUY",
        Type:        "MARKET",
        Size:        size,
        TimeInForce: "IOC",
    }
    
    return s.PlaceOrder(params)
}

// 市价卖出
func (s *SpotTradingService) MarketSell(market string, size decimal.Decimal) (*Order, error) {
    params := &PlaceOrderParams{
        Market:      market,
        Side:        "SELL",
        Type:        "MARKET",
        Size:        size,
        TimeInForce: "IOC",
    }
    
    return s.PlaceOrder(params)
}

// 限价买入
func (s *SpotTradingService) LimitBuy(market string, size, price decimal.Decimal) (*Order, error) {
    params := &PlaceOrderParams{
        Market:      market,
        Side:        "BUY",
        Type:        "LIMIT",
        Size:        size,
        Price:       price,
        TimeInForce: "GTT",
        PostOnly:    true,
    }
    
    return s.PlaceOrder(params)
}

// 限价卖出
func (s *SpotTradingService) LimitSell(market string, size, price decimal.Decimal) (*Order, error) {
    params := &PlaceOrderParams{
        Market:      market,
        Side:        "SELL",
        Type:        "LIMIT",
        Size:        size,
        Price:       price,
        TimeInForce: "GTT",
        PostOnly:    true,
    }
    
    return s.PlaceOrder(params)
}

// 止损订单
func (s *SpotTradingService) StopLoss(market, side string, size, triggerPrice decimal.Decimal) (*Order, error) {
    params := &PlaceOrderParams{
        Market:       market,
        Side:         side,
        Type:         "STOP_LIMIT",
        Size:         size,
        TriggerPrice: triggerPrice,
        Price:        triggerPrice.Mul(decimal.NewFromFloat(0.99)), // 触发后以99%价格成交
        TimeInForce:  "IOC",
        ReduceOnly:   true,
    }
    
    return s.PlaceOrder(params)
}

// 止盈订单
func (s *SpotTradingService) TakeProfit(market, side string, size, triggerPrice decimal.Decimal) (*Order, error) {
    params := &PlaceOrderParams{
        Market:       market,
        Side:         side,
        Type:         "TAKE_PROFIT",
        Size:         size,
        TriggerPrice: triggerPrice,
        Price:        triggerPrice.Mul(decimal.NewFromFloat(1.01)), // 触发后以101%价格成交
        TimeInForce:  "IOC",
        ReduceOnly:   true,
    }
    
    return s.PlaceOrder(params)
}
```

## 永续合约

### 4.1 永续合约服务

```go
// services/perpetual_service.go
package services

import (
    "encoding/json"
    "fmt"
    
    "github.com/shopspring/decimal"
)

type PerpetualService struct {
    client *DYDXClient
}

func NewPerpetualService(client *DYDXClient) *PerpetualService {
    return &PerpetualService{
        client: client,
    }
}

// 资金费率信息
type FundingRate struct {
    Market           string          `json:"market"`
    Rate             decimal.Decimal `json:"rate"`
    Price            decimal.Decimal `json:"price"`
    EffectiveAt      time.Time       `json:"effectiveAt"`
}

// 获取资金费率
func (s *PerpetualService) GetFundingRates(market string) ([]FundingRate, error) {
    endpoint := "/v3/historical-funding"
    if market != "" {
        endpoint += "?market=" + market
    }
    
    respBody, err := s.client.sendRequest("GET", endpoint, nil)
    if err != nil {
        return nil, err
    }
    
    var response struct {
        HistoricalFunding []FundingRate `json:"historicalFunding"`
    }
    
    if err := json.Unmarshal(respBody, &response); err != nil {
        return nil, err
    }
    
    return response.HistoricalFunding, nil
}

// 开多仓
func (s *PerpetualService) OpenLong(market string, size, price decimal.Decimal, leverage int) (*Order, error) {
    // 计算保证金
    margin := size.Mul(price).Div(decimal.NewFromInt(int64(leverage)))
    
    params := &PlaceOrderParams{
        Market:      market,
        Side:        "BUY",
        Type:        "LIMIT",
        Size:        size,
        Price:       price,
        TimeInForce: "GTT",
    }
    
    return s.placePositionOrder(params, margin)
}

// 开空仓
func (s *PerpetualService) OpenShort(market string, size, price decimal.Decimal, leverage int) (*Order, error) {
    // 计算保证金
    margin := size.Mul(price).Div(decimal.NewFromInt(int64(leverage)))
    
    params := &PlaceOrderParams{
        Market:      market,
        Side:        "SELL",
        Type:        "LIMIT",
        Size:        size,
        Price:       price,
        TimeInForce: "GTT",
    }
    
    return s.placePositionOrder(params, margin)
}

// 平仓
func (s *PerpetualService) ClosePosition(market string, size decimal.Decimal, isLong bool) (*Order, error) {
    side := "SELL"
    if !isLong {
        side = "BUY"
    }
    
    params := &PlaceOrderParams{
        Market:      market,
        Side:        side,
        Type:        "MARKET",
        Size:        size,
        TimeInForce: "IOC",
        ReduceOnly:  true,
    }
    
    respBody, err := s.client.sendRequest("POST", "/v3/orders", params)
    if err != nil {
        return nil, err
    }
    
    var response struct {
        Order *Order `json:"order"`
    }
    
    if err := json.Unmarshal(respBody, &response); err != nil {
        return nil, err
    }
    
    return response.Order, nil
}

// 调整保证金
func (s *PerpetualService) AdjustMargin(market string, amount decimal.Decimal, isAdd bool) error {
    action := "ADD"
    if !isAdd {
        action = "REMOVE"
    }
    
    params := map[string]interface{}{
        "market": market,
        "amount": amount,
        "action": action,
    }
    
    _, err := s.client.sendRequest("POST", "/v3/positions/adjust-margin", params)
    return err
}

// 设置止损止盈
func (s *PerpetualService) SetStopLossAndTakeProfit(
    market string,
    positionSize decimal.Decimal,
    isLong bool,
    stopLossPrice, takeProfitPrice decimal.Decimal,
) ([]*Order, error) {
    var orders []*Order
    
    // 确定平仓方向
    closeSide := "SELL"
    if !isLong {
        closeSide = "BUY"
    }
    
    // 设置止损
    if !stopLossPrice.IsZero() {
        stopLossOrder := &PlaceOrderParams{
            Market:       market,
            Side:         closeSide,
            Type:         "STOP_LIMIT",
            Size:         positionSize,
            TriggerPrice: stopLossPrice,
            Price:        stopLossPrice.Mul(decimal.NewFromFloat(0.995)), // 稍微低于触发价
            TimeInForce:  "IOC",
            ReduceOnly:   true,
        }
        
        order, err := s.placeOrder(stopLossOrder)
        if err != nil {
            return nil, err
        }
        orders = append(orders, order)
    }
    
    // 设置止盈
    if !takeProfitPrice.IsZero() {
        takeProfitOrder := &PlaceOrderParams{
            Market:       market,
            Side:         closeSide,
            Type:         "TAKE_PROFIT",
            Size:         positionSize,
            TriggerPrice: takeProfitPrice,
            Price:        takeProfitPrice.Mul(decimal.NewFromFloat(1.005)), // 稍微高于触发价
            TimeInForce:  "IOC",
            ReduceOnly:   true,
        }
        
        order, err := s.placeOrder(takeProfitOrder)
        if err != nil {
            return nil, err
        }
        orders = append(orders, order)
    }
    
    return orders, nil
}

// 计算清算价格
func (s *PerpetualService) CalculateLiquidationPrice(
    entryPrice decimal.Decimal,
    size decimal.Decimal,
    margin decimal.Decimal,
    isLong bool,
    maintenanceMarginRate decimal.Decimal,
) decimal.Decimal {
    // 简化的清算价格计算
    // 实际计算需要考虑资金费率、未实现盈亏等因素
    
    if isLong {
        // 多仓清算价格 = 入场价格 * (1 - 保证金率 + 维持保证金率)
        factor := decimal.NewFromInt(1).Sub(margin.Div(size.Mul(entryPrice))).Add(maintenanceMarginRate)
        return entryPrice.Mul(factor)
    } else {
        // 空仓清算价格 = 入场价格 * (1 + 保证金率 - 维持保证金率)
        factor := decimal.NewFromInt(1).Add(margin.Div(size.Mul(entryPrice))).Sub(maintenanceMarginRate)
        return entryPrice.Mul(factor)
    }
}

// 辅助函数
func (s *PerpetualService) placePositionOrder(params *PlaceOrderParams, margin decimal.Decimal) (*Order, error) {
    // 这里应该包含保证金检查和设置逻辑
    respBody, err := s.client.sendRequest("POST", "/v3/orders", params)
    if err != nil {
        return nil, err
    }
    
    var response struct {
        Order *Order `json:"order"`
    }
    
    if err := json.Unmarshal(respBody, &response); err != nil {
        return nil, err
    }
    
    return response.Order, nil
}

func (s *PerpetualService) placeOrder(params *PlaceOrderParams) (*Order, error) {
    respBody, err := s.client.sendRequest("POST", "/v3/orders", params)
    if err != nil {
        return nil, err
    }
    
    var response struct {
        Order *Order `json:"order"`
    }
    
    if err := json.Unmarshal(respBody, &response); err != nil {
        return nil, err
    }
    
    return response.Order, nil
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
    
    "github.com/shopspring/decimal"
    
    "your-project/client"
    "your-project/services"
)

func main() {
    // 创建dYdX客户端
    dydxClient := client.NewDYDXClient(
        "your_api_key",
        "your_api_secret", 
        "your_passphrase",
        false, // 使用测试网
    )
    
    // 创建服务
    spotService := services.NewSpotTradingService(dydxClient)
    perpService := services.NewPerpetualService(dydxClient)
    
    // 1. 获取市场信息
    markets, err := dydxClient.GetMarkets()
    if err != nil {
        log.Fatal("获取市场信息失败:", err)
    }
    
    fmt.Printf("可用市场:\n")
    for marketName, market := range markets {
        fmt.Printf("  %s: %s/%s, 价格: %s\n", 
            marketName, 
            market.BaseAsset, 
            market.QuoteAsset, 
            market.IndexPrice.String())
    }
    
    // 2. 获取账户信息
    accountID := "your_account_id"
    account, err := dydxClient.GetAccount(accountID)
    if err != nil {
        log.Fatal("获取账户信息失败:", err)
    }
    
    fmt.Printf("账户信息:\n")
    fmt.Printf("  权益: %s\n", account.Equity.String())
    fmt.Printf("  可用保证金: %s\n", account.FreeCollateral.String())
    fmt.Printf("  报价余额: %s\n", account.QuoteBalance.String())
    
    // 3. 现货交易示例
    fmt.Printf("\n=== 现货交易示例 ===\n")
    
    // 限价买入BTC
    buyOrder, err := spotService.LimitBuy(
        "BTC-USD",
        decimal.NewFromFloat(0.01), // 0.01 BTC
        decimal.NewFromFloat(45000), // $45,000
    )
    if err != nil {
        log.Printf("下买单失败: %v", err)
    } else {
        fmt.Printf("买单已提交: %s\n", buyOrder.ID)
    }
    
    // 查询订单状态
    orders, err := spotService.GetOrders(accountID, "BTC-USD", "OPEN")
    if err != nil {
        log.Printf("查询订单失败: %v", err)
    } else {
        fmt.Printf("当前开放订单数量: %d\n", len(orders))
        for _, order := range orders {
            fmt.Printf("  订单 %s: %s %s %s @ %s\n",
                order.ID,
                order.Side,
                order.Size.String(),
                order.Market,
                order.Price.String())
        }
    }
    
    // 4. 永续合约交易示例
    fmt.Printf("\n=== 永续合约交易示例 ===\n")
    
    // 开多仓
    longOrder, err := perpService.OpenLong(
        "BTC-USD",
        decimal.NewFromFloat(0.1),  // 0.1 BTC
        decimal.NewFromFloat(46000), // $46,000
        10, // 10倍杠杆
    )
    if err != nil {
        log.Printf("开多仓失败: %v", err)
    } else {
        fmt.Printf("多仓订单已提交: %s\n", longOrder.ID)
    }
    
    // 获取持仓信息
    positions, err := dydxClient.GetPositions(accountID)
    if err != nil {
        log.Printf("获取持仓失败: %v", err)
    } else {
        fmt.Printf("当前持仓:\n")
        for market, position := range positions {
            if position.Status == "OPEN" {
                fmt.Printf("  %s: %s %s, 入场价: %s, 未实现盈亏: %s\n",
                    market,
                    position.Side,
                    position.Size.String(),
                    position.EntryPrice.String(),
                    position.UnrealizedPnl.String())
                
                // 计算清算价格
                liquidationPrice := perpService.CalculateLiquidationPrice(
                    position.EntryPrice,
                    position.Size,
                    decimal.NewFromFloat(1000), // 假设保证金
                    position.Side == "LONG",
                    decimal.NewFromFloat(0.05), // 5%维持保证金率
                )
                fmt.Printf("    清算价格: %s\n", liquidationPrice.String())
            }
        }
    }
    
    // 5. 设置止损止盈
    if len(positions) > 0 {
        for market, position := range positions {
            if position.Status == "OPEN" && position.Side == "LONG" {
                // 设置止损在入场价下5%，止盈在入场价上10%
                stopLoss := position.EntryPrice.Mul(decimal.NewFromFloat(0.95))
                takeProfit := position.EntryPrice.Mul(decimal.NewFromFloat(1.10))
                
                orders, err := perpService.SetStopLossAndTakeProfit(
                    market,
                    position.Size,
                    true, // 多仓
                    stopLoss,
                    takeProfit,
                )
                if err != nil {
                    log.Printf("设置止损止盈失败: %v", err)
                } else {
                    fmt.Printf("已设置止损止盈，订单数量: %d\n", len(orders))
                }
                break
            }
        }
    }
    
    // 6. 获取资金费率
    fundingRates, err := perpService.GetFundingRates("BTC-USD")
    if err != nil {
        log.Printf("获取资金费率失败: %v", err)
    } else {
        fmt.Printf("\nBTC-USD 最近资金费率:\n")
        for i, rate := range fundingRates {
            if i >= 5 { // 只显示最近5个
                break
            }
            fmt.Printf("  %s: %s%% @ %s\n",
                rate.EffectiveAt.Format("2006-01-02 15:04"),
                rate.Rate.Mul(decimal.NewFromInt(100)).String(),
                rate.Price.String())
        }
    }
}
```

这个dYdX使用指南提供了完整的去中心化衍生品交易所集成方案，涵盖了现货交易、永续合约、杠杆交易、风险管理等核心功能，是DeFi衍生品交易的重要参考文档。
