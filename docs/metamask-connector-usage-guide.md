# MetaMask连接器 详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [环境准备](#环境准备)
3. [连接管理](#连接管理)
4. [账户操作](#账户操作)
5. [网络切换](#网络切换)
6. [交易签名](#交易签名)
7. [消息签名](#消息签名)
8. [事件监听](#事件监听)
9. [实际应用](#实际应用)

## 基础概念

### 1.1 MetaMask连接器简介

MetaMask连接器提供Go后端与MetaMask钱包的桥接功能，通过WebSocket或HTTP API实现钱包连接、账户管理、交易签名等功能。

```bash
# 安装MetaMask连接相关依赖
go get github.com/gorilla/websocket
go get github.com/gin-gonic/gin
go get github.com/ethereum/go-ethereum
go get github.com/shopspring/decimal
```

### 1.2 架构设计

```go
// 主要包导入
import (
    "context"
    "encoding/json"
    "net/http"
    
    "github.com/gorilla/websocket"
    "github.com/gin-gonic/gin"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
)

// 核心概念：
// - Connector: MetaMask连接器
// - Session: 用户会话管理
// - Request: 钱包请求处理
// - Response: 钱包响应处理
// - Event: 事件监听和分发
```

## 环境准备

### 2.1 连接器配置

```go
// config/metamask.go
package config

import (
    "time"
)

type MetaMaskConfig struct {
    // 服务器配置
    Host           string
    Port           int
    TLSEnabled     bool
    CertFile       string
    KeyFile        string
    
    // WebSocket配置
    WSPath         string
    WSOrigins      []string
    WSTimeout      time.Duration
    
    // 会话配置
    SessionTimeout time.Duration
    MaxSessions    int
    
    // 安全配置
    AllowedOrigins []string
    APIKey         string
    RateLimit      int
    
    // 日志配置
    LogLevel       string
    LogFile        string
}

func DefaultMetaMaskConfig() *MetaMaskConfig {
    return &MetaMaskConfig{
        Host:           "localhost",
        Port:           8080,
        TLSEnabled:     false,
        WSPath:         "/ws",
        WSOrigins:      []string{"*"},
        WSTimeout:      30 * time.Second,
        SessionTimeout: 24 * time.Hour,
        MaxSessions:    1000,
        AllowedOrigins: []string{"http://localhost:3000", "https://yourdapp.com"},
        RateLimit:      100, // 每分钟请求数
        LogLevel:       "info",
    }
}

// 网络配置
type NetworkConfig struct {
    ChainID      string `json:"chainId"`
    ChainName    string `json:"chainName"`
    NativeCurrency struct {
        Name     string `json:"name"`
        Symbol   string `json:"symbol"`
        Decimals int    `json:"decimals"`
    } `json:"nativeCurrency"`
    RPCUrls      []string `json:"rpcUrls"`
    BlockExplorerUrls []string `json:"blockExplorerUrls"`
}

// 预定义网络
var (
    EthereumMainnet = NetworkConfig{
        ChainID:   "0x1",
        ChainName: "Ethereum Mainnet",
        NativeCurrency: struct {
            Name     string `json:"name"`
            Symbol   string `json:"symbol"`
            Decimals int    `json:"decimals"`
        }{
            Name:     "Ether",
            Symbol:   "ETH",
            Decimals: 18,
        },
        RPCUrls:           []string{"https://mainnet.infura.io/v3/"},
        BlockExplorerUrls: []string{"https://etherscan.io"},
    }
    
    PolygonMainnet = NetworkConfig{
        ChainID:   "0x89",
        ChainName: "Polygon Mainnet",
        NativeCurrency: struct {
            Name     string `json:"name"`
            Symbol   string `json:"symbol"`
            Decimals int    `json:"decimals"`
        }{
            Name:     "MATIC",
            Symbol:   "MATIC",
            Decimals: 18,
        },
        RPCUrls:           []string{"https://polygon-rpc.com/"},
        BlockExplorerUrls: []string{"https://polygonscan.com"},
    }
)
```

## 连接管理

### 3.1 连接器核心

```go
// connector/metamask.go
package connector

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "sync"
    "time"

    "github.com/gorilla/websocket"
    "github.com/gin-gonic/gin"
    
    "your-project/config"
)

type MetaMaskConnector struct {
    config    *config.MetaMaskConfig
    sessions  map[string]*Session
    upgrader  websocket.Upgrader
    mutex     sync.RWMutex
    ctx       context.Context
    cancel    context.CancelFunc
}

func NewMetaMaskConnector(cfg *config.MetaMaskConfig) *MetaMaskConnector {
    ctx, cancel := context.WithCancel(context.Background())
    
    upgrader := websocket.Upgrader{
        CheckOrigin: func(r *http.Request) bool {
            origin := r.Header.Get("Origin")
            for _, allowed := range cfg.AllowedOrigins {
                if allowed == "*" || allowed == origin {
                    return true
                }
            }
            return false
        },
        HandshakeTimeout: cfg.WSTimeout,
    }

    return &MetaMaskConnector{
        config:   cfg,
        sessions: make(map[string]*Session),
        upgrader: upgrader,
        ctx:      ctx,
        cancel:   cancel,
    }
}

// 启动连接器服务
func (mc *MetaMaskConnector) Start() error {
    router := gin.Default()
    
    // 设置CORS
    router.Use(func(c *gin.Context) {
        origin := c.Request.Header.Get("Origin")
        for _, allowed := range mc.config.AllowedOrigins {
            if allowed == "*" || allowed == origin {
                c.Header("Access-Control-Allow-Origin", origin)
                break
            }
        }
        c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
        
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        
        c.Next()
    })

    // WebSocket端点
    router.GET(mc.config.WSPath, mc.handleWebSocket)
    
    // HTTP API端点
    api := router.Group("/api/v1")
    {
        api.POST("/connect", mc.handleConnect)
        api.POST("/disconnect", mc.handleDisconnect)
        api.GET("/sessions", mc.handleGetSessions)
        api.POST("/request", mc.handleRequest)
    }

    // 启动清理goroutine
    go mc.cleanupSessions()

    addr := fmt.Sprintf("%s:%d", mc.config.Host, mc.config.Port)
    
    if mc.config.TLSEnabled {
        return http.ListenAndServeTLS(addr, mc.config.CertFile, mc.config.KeyFile, router)
    }
    
    return http.ListenAndServe(addr, router)
}

// 处理WebSocket连接
func (mc *MetaMaskConnector) handleWebSocket(c *gin.Context) {
    conn, err := mc.upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "WebSocket升级失败"})
        return
    }

    sessionID := generateSessionID()
    session := NewSession(sessionID, conn, mc.config.SessionTimeout)
    
    mc.mutex.Lock()
    mc.sessions[sessionID] = session
    mc.mutex.Unlock()

    // 处理会话
    go mc.handleSession(session)
}

// 处理会话
func (mc *MetaMaskConnector) handleSession(session *Session) {
    defer func() {
        mc.mutex.Lock()
        delete(mc.sessions, session.ID)
        mc.mutex.Unlock()
        session.Close()
    }()

    for {
        select {
        case <-mc.ctx.Done():
            return
        case <-session.ctx.Done():
            return
        default:
            var message Message
            err := session.conn.ReadJSON(&message)
            if err != nil {
                if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                    fmt.Printf("WebSocket错误: %v\n", err)
                }
                return
            }

            // 处理消息
            response := mc.processMessage(session, &message)
            if response != nil {
                session.SendMessage(response)
            }
        }
    }
}

// 处理连接请求
func (mc *MetaMaskConnector) handleConnect(c *gin.Context) {
    var req ConnectRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求格式"})
        return
    }

    // 创建连接会话
    sessionID := generateSessionID()
    session := NewHTTPSession(sessionID, mc.config.SessionTimeout)
    
    mc.mutex.Lock()
    mc.sessions[sessionID] = session
    mc.mutex.Unlock()

    c.JSON(http.StatusOK, gin.H{
        "sessionId": sessionID,
        "status":    "connected",
    })
}

// 清理过期会话
func (mc *MetaMaskConnector) cleanupSessions() {
    ticker := time.NewTicker(time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-mc.ctx.Done():
            return
        case <-ticker.C:
            mc.mutex.Lock()
            for id, session := range mc.sessions {
                if session.IsExpired() {
                    delete(mc.sessions, id)
                    session.Close()
                }
            }
            mc.mutex.Unlock()
        }
    }
}

// 关闭连接器
func (mc *MetaMaskConnector) Close() {
    mc.cancel()
    
    mc.mutex.Lock()
    for _, session := range mc.sessions {
        session.Close()
    }
    mc.sessions = make(map[string]*Session)
    mc.mutex.Unlock()
}
```

## 账户操作

### 4.1 账户管理

```go
// account/manager.go
package account

import (
    "encoding/json"
    "fmt"

    "github.com/ethereum/go-ethereum/common"
    
    "your-project/connector"
)

type AccountManager struct {
    connector *connector.MetaMaskConnector
}

func NewAccountManager(connector *connector.MetaMaskConnector) *AccountManager {
    return &AccountManager{
        connector: connector,
    }
}

// 请求连接账户
func (am *AccountManager) RequestAccounts(sessionID string) (*AccountsResponse, error) {
    request := &connector.Message{
        ID:     generateRequestID(),
        Method: "eth_requestAccounts",
        Params: []interface{}{},
    }

    response, err := am.connector.SendRequest(sessionID, request)
    if err != nil {
        return nil, fmt.Errorf("请求账户失败: %v", err)
    }

    if response.Error != nil {
        return nil, fmt.Errorf("MetaMask错误: %s", response.Error.Message)
    }

    var accounts []string
    err = json.Unmarshal(response.Result, &accounts)
    if err != nil {
        return nil, fmt.Errorf("解析账户响应失败: %v", err)
    }

    return &AccountsResponse{
        Accounts: accounts,
        Primary:  accounts[0],
    }, nil
}

// 获取当前账户
func (am *AccountManager) GetAccounts(sessionID string) (*AccountsResponse, error) {
    request := &connector.Message{
        ID:     generateRequestID(),
        Method: "eth_accounts",
        Params: []interface{}{},
    }

    response, err := am.connector.SendRequest(sessionID, request)
    if err != nil {
        return nil, fmt.Errorf("获取账户失败: %v", err)
    }

    if response.Error != nil {
        return nil, fmt.Errorf("MetaMask错误: %s", response.Error.Message)
    }

    var accounts []string
    err = json.Unmarshal(response.Result, &accounts)
    if err != nil {
        return nil, fmt.Errorf("解析账户响应失败: %v", err)
    }

    if len(accounts) == 0 {
        return &AccountsResponse{
            Accounts: []string{},
            Primary:  "",
        }, nil
    }

    return &AccountsResponse{
        Accounts: accounts,
        Primary:  accounts[0],
    }, nil
}

// 获取账户余额
func (am *AccountManager) GetBalance(sessionID string, address common.Address) (*BalanceResponse, error) {
    request := &connector.Message{
        ID:     generateRequestID(),
        Method: "eth_getBalance",
        Params: []interface{}{address.Hex(), "latest"},
    }

    response, err := am.connector.SendRequest(sessionID, request)
    if err != nil {
        return nil, fmt.Errorf("获取余额失败: %v", err)
    }

    if response.Error != nil {
        return nil, fmt.Errorf("MetaMask错误: %s", response.Error.Message)
    }

    var balanceHex string
    err = json.Unmarshal(response.Result, &balanceHex)
    if err != nil {
        return nil, fmt.Errorf("解析余额响应失败: %v", err)
    }

    return &BalanceResponse{
        Address: address,
        Balance: balanceHex,
    }, nil
}

// 监听账户变化
func (am *AccountManager) WatchAccountsChanged(sessionID string, handler func([]string)) error {
    return am.connector.SubscribeEvent(sessionID, "accountsChanged", func(data interface{}) {
        if accounts, ok := data.([]string); ok {
            handler(accounts)
        }
    })
}

type AccountsResponse struct {
    Accounts []string `json:"accounts"`
    Primary  string   `json:"primary"`
}

type BalanceResponse struct {
    Address common.Address `json:"address"`
    Balance string         `json:"balance"`
}
```

## 网络切换

### 5.1 网络管理

```go
// network/manager.go
package network

import (
    "encoding/json"
    "fmt"

    "your-project/connector"
    "your-project/config"
)

type NetworkManager struct {
    connector *connector.MetaMaskConnector
}

func NewNetworkManager(connector *connector.MetaMaskConnector) *NetworkManager {
    return &NetworkManager{
        connector: connector,
    }
}

// 获取当前网络
func (nm *NetworkManager) GetChainId(sessionID string) (*ChainResponse, error) {
    request := &connector.Message{
        ID:     generateRequestID(),
        Method: "eth_chainId",
        Params: []interface{}{},
    }

    response, err := nm.connector.SendRequest(sessionID, request)
    if err != nil {
        return nil, fmt.Errorf("获取链ID失败: %v", err)
    }

    if response.Error != nil {
        return nil, fmt.Errorf("MetaMask错误: %s", response.Error.Message)
    }

    var chainId string
    err = json.Unmarshal(response.Result, &chainId)
    if err != nil {
        return nil, fmt.Errorf("解析链ID响应失败: %v", err)
    }

    return &ChainResponse{
        ChainId: chainId,
    }, nil
}

// 切换网络
func (nm *NetworkManager) SwitchChain(sessionID string, chainId string) error {
    request := &connector.Message{
        ID:     generateRequestID(),
        Method: "wallet_switchEthereumChain",
        Params: []interface{}{
            map[string]string{
                "chainId": chainId,
            },
        },
    }

    response, err := nm.connector.SendRequest(sessionID, request)
    if err != nil {
        return fmt.Errorf("切换网络失败: %v", err)
    }

    if response.Error != nil {
        // 如果网络不存在，尝试添加网络
        if response.Error.Code == 4902 {
            return fmt.Errorf("网络不存在，请先添加网络")
        }
        return fmt.Errorf("MetaMask错误: %s", response.Error.Message)
    }

    return nil
}

// 添加网络
func (nm *NetworkManager) AddChain(sessionID string, network config.NetworkConfig) error {
    request := &connector.Message{
        ID:     generateRequestID(),
        Method: "wallet_addEthereumChain",
        Params: []interface{}{network},
    }

    response, err := nm.connector.SendRequest(sessionID, request)
    if err != nil {
        return fmt.Errorf("添加网络失败: %v", err)
    }

    if response.Error != nil {
        return fmt.Errorf("MetaMask错误: %s", response.Error.Message)
    }

    return nil
}

// 添加并切换到网络
func (nm *NetworkManager) AddAndSwitchChain(sessionID string, network config.NetworkConfig) error {
    // 先尝试切换
    err := nm.SwitchChain(sessionID, network.ChainID)
    if err != nil {
        // 如果切换失败，尝试添加网络
        err = nm.AddChain(sessionID, network)
        if err != nil {
            return fmt.Errorf("添加网络失败: %v", err)
        }

        // 添加成功后再次尝试切换
        err = nm.SwitchChain(sessionID, network.ChainID)
        if err != nil {
            return fmt.Errorf("切换到新添加的网络失败: %v", err)
        }
    }

    return nil
}

// 监听网络变化
func (nm *NetworkManager) WatchChainChanged(sessionID string, handler func(string)) error {
    return nm.connector.SubscribeEvent(sessionID, "chainChanged", func(data interface{}) {
        if chainId, ok := data.(string); ok {
            handler(chainId)
        }
    })
}

// 预定义网络操作
func (nm *NetworkManager) SwitchToEthereum(sessionID string) error {
    return nm.SwitchChain(sessionID, "0x1")
}

func (nm *NetworkManager) SwitchToPolygon(sessionID string) error {
    return nm.AddAndSwitchChain(sessionID, config.PolygonMainnet)
}

func (nm *NetworkManager) SwitchToBSC(sessionID string) error {
    bscMainnet := config.NetworkConfig{
        ChainID:   "0x38",
        ChainName: "Binance Smart Chain Mainnet",
        NativeCurrency: struct {
            Name     string `json:"name"`
            Symbol   string `json:"symbol"`
            Decimals int    `json:"decimals"`
        }{
            Name:     "Binance Coin",
            Symbol:   "BNB",
            Decimals: 18,
        },
        RPCUrls:           []string{"https://bsc-dataseed1.binance.org/"},
        BlockExplorerUrls: []string{"https://bscscan.com"},
    }
    
    return nm.AddAndSwitchChain(sessionID, bscMainnet)
}

type ChainResponse struct {
    ChainId string `json:"chainId"`
}
```

## 交易签名

### 6.1 交易管理

```go
// transaction/signer.go
package transaction

import (
    "encoding/json"
    "fmt"
    "math/big"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    
    "your-project/connector"
)

type TransactionSigner struct {
    connector *connector.MetaMaskConnector
}

func NewTransactionSigner(connector *connector.MetaMaskConnector) *TransactionSigner {
    return &TransactionSigner{
        connector: connector,
    }
}

// 发送交易
func (ts *TransactionSigner) SendTransaction(sessionID string, tx TransactionRequest) (*TransactionResponse, error) {
    request := &connector.Message{
        ID:     generateRequestID(),
        Method: "eth_sendTransaction",
        Params: []interface{}{tx},
    }

    response, err := ts.connector.SendRequest(sessionID, request)
    if err != nil {
        return nil, fmt.Errorf("发送交易失败: %v", err)
    }

    if response.Error != nil {
        return nil, fmt.Errorf("MetaMask错误: %s", response.Error.Message)
    }

    var txHash string
    err = json.Unmarshal(response.Result, &txHash)
    if err != nil {
        return nil, fmt.Errorf("解析交易哈希失败: %v", err)
    }

    return &TransactionResponse{
        Hash: txHash,
    }, nil
}

// 签名交易（不发送）
func (ts *TransactionSigner) SignTransaction(sessionID string, tx TransactionRequest) (*SignedTransactionResponse, error) {
    request := &connector.Message{
        ID:     generateRequestID(),
        Method: "eth_signTransaction",
        Params: []interface{}{tx},
    }

    response, err := ts.connector.SendRequest(sessionID, request)
    if err != nil {
        return nil, fmt.Errorf("签名交易失败: %v", err)
    }

    if response.Error != nil {
        return nil, fmt.Errorf("MetaMask错误: %s", response.Error.Message)
    }

    var signedTx SignedTransactionData
    err = json.Unmarshal(response.Result, &signedTx)
    if err != nil {
        return nil, fmt.Errorf("解析签名交易失败: %v", err)
    }

    return &SignedTransactionResponse{
        Raw: signedTx.Raw,
        Tx:  signedTx.Tx,
    }, nil
}

// 估算Gas费用
func (ts *TransactionSigner) EstimateGas(sessionID string, tx TransactionRequest) (*GasEstimateResponse, error) {
    request := &connector.Message{
        ID:     generateRequestID(),
        Method: "eth_estimateGas",
        Params: []interface{}{tx},
    }

    response, err := ts.connector.SendRequest(sessionID, request)
    if err != nil {
        return nil, fmt.Errorf("估算Gas失败: %v", err)
    }

    if response.Error != nil {
        return nil, fmt.Errorf("MetaMask错误: %s", response.Error.Message)
    }

    var gasHex string
    err = json.Unmarshal(response.Result, &gasHex)
    if err != nil {
        return nil, fmt.Errorf("解析Gas估算失败: %v", err)
    }

    return &GasEstimateResponse{
        GasLimit: gasHex,
    }, nil
}

// 获取Gas价格
func (ts *TransactionSigner) GetGasPrice(sessionID string) (*GasPriceResponse, error) {
    request := &connector.Message{
        ID:     generateRequestID(),
        Method: "eth_gasPrice",
        Params: []interface{}{},
    }

    response, err := ts.connector.SendRequest(sessionID, request)
    if err != nil {
        return nil, fmt.Errorf("获取Gas价格失败: %v", err)
    }

    if response.Error != nil {
        return nil, fmt.Errorf("MetaMask错误: %s", response.Error.Message)
    }

    var gasPriceHex string
    err = json.Unmarshal(response.Result, &gasPriceHex)
    if err != nil {
        return nil, fmt.Errorf("解析Gas价格失败: %v", err)
    }

    return &GasPriceResponse{
        GasPrice: gasPriceHex,
    }, nil
}

// 发送原始交易
func (ts *TransactionSigner) SendRawTransaction(sessionID string, rawTx string) (*TransactionResponse, error) {
    request := &connector.Message{
        ID:     generateRequestID(),
        Method: "eth_sendRawTransaction",
        Params: []interface{}{rawTx},
    }

    response, err := ts.connector.SendRequest(sessionID, request)
    if err != nil {
        return nil, fmt.Errorf("发送原始交易失败: %v", err)
    }

    if response.Error != nil {
        return nil, fmt.Errorf("MetaMask错误: %s", response.Error.Message)
    }

    var txHash string
    err = json.Unmarshal(response.Result, &txHash)
    if err != nil {
        return nil, fmt.Errorf("解析交易哈希失败: %v", err)
    }

    return &TransactionResponse{
        Hash: txHash,
    }, nil
}

type TransactionRequest struct {
    From     string `json:"from"`
    To       string `json:"to,omitempty"`
    Value    string `json:"value,omitempty"`
    Gas      string `json:"gas,omitempty"`
    GasPrice string `json:"gasPrice,omitempty"`
    Data     string `json:"data,omitempty"`
    Nonce    string `json:"nonce,omitempty"`
}

type TransactionResponse struct {
    Hash string `json:"hash"`
}

type SignedTransactionResponse struct {
    Raw string                `json:"raw"`
    Tx  SignedTransactionData `json:"tx"`
}

type SignedTransactionData struct {
    Nonce    string `json:"nonce"`
    GasPrice string `json:"gasPrice"`
    Gas      string `json:"gas"`
    To       string `json:"to"`
    Value    string `json:"value"`
    Input    string `json:"input"`
    V        string `json:"v"`
    R        string `json:"r"`
    S        string `json:"s"`
    Hash     string `json:"hash"`
    Raw      string `json:"raw"`
}

type GasEstimateResponse struct {
    GasLimit string `json:"gasLimit"`
}

type GasPriceResponse struct {
    GasPrice string `json:"gasPrice"`
}
```

## 消息签名

### 7.1 签名管理

```go
// signature/manager.go
package signature

import (
    "encoding/json"
    "fmt"

    "your-project/connector"
)

type SignatureManager struct {
    connector *connector.MetaMaskConnector
}

func NewSignatureManager(connector *connector.MetaMaskConnector) *SignatureManager {
    return &SignatureManager{
        connector: connector,
    }
}

// 个人签名
func (sm *SignatureManager) PersonalSign(sessionID string, message string, address string) (*SignatureResponse, error) {
    request := &connector.Message{
        ID:     generateRequestID(),
        Method: "personal_sign",
        Params: []interface{}{message, address},
    }

    response, err := sm.connector.SendRequest(sessionID, request)
    if err != nil {
        return nil, fmt.Errorf("个人签名失败: %v", err)
    }

    if response.Error != nil {
        return nil, fmt.Errorf("MetaMask错误: %s", response.Error.Message)
    }

    var signature string
    err = json.Unmarshal(response.Result, &signature)
    if err != nil {
        return nil, fmt.Errorf("解析签名失败: %v", err)
    }

    return &SignatureResponse{
        Signature: signature,
        Message:   message,
        Address:   address,
    }, nil
}

// 类型化数据签名 (EIP-712)
func (sm *SignatureManager) SignTypedData(sessionID string, address string, typedData TypedData) (*SignatureResponse, error) {
    request := &connector.Message{
        ID:     generateRequestID(),
        Method: "eth_signTypedData_v4",
        Params: []interface{}{address, typedData},
    }

    response, err := sm.connector.SendRequest(sessionID, request)
    if err != nil {
        return nil, fmt.Errorf("类型化数据签名失败: %v", err)
    }

    if response.Error != nil {
        return nil, fmt.Errorf("MetaMask错误: %s", response.Error.Message)
    }

    var signature string
    err = json.Unmarshal(response.Result, &signature)
    if err != nil {
        return nil, fmt.Errorf("解析签名失败: %v", err)
    }

    return &SignatureResponse{
        Signature: signature,
        Address:   address,
        TypedData: &typedData,
    }, nil
}

// 以太坊签名
func (sm *SignatureManager) EthSign(sessionID string, address string, dataHash string) (*SignatureResponse, error) {
    request := &connector.Message{
        ID:     generateRequestID(),
        Method: "eth_sign",
        Params: []interface{}{address, dataHash},
    }

    response, err := sm.connector.SendRequest(sessionID, request)
    if err != nil {
        return nil, fmt.Errorf("以太坊签名失败: %v", err)
    }

    if response.Error != nil {
        return nil, fmt.Errorf("MetaMask错误: %s", response.Error.Message)
    }

    var signature string
    err = json.Unmarshal(response.Result, &signature)
    if err != nil {
        return nil, fmt.Errorf("解析签名失败: %v", err)
    }

    return &SignatureResponse{
        Signature: signature,
        DataHash:  dataHash,
        Address:   address,
    }, nil
}

// 验证签名
func (sm *SignatureManager) VerifySignature(signature string, message string, address string) (*VerificationResponse, error) {
    // 这里需要实现签名验证逻辑
    // 可以使用go-ethereum的crypto包进行验证
    
    return &VerificationResponse{
        Valid:     true,
        Signature: signature,
        Message:   message,
        Address:   address,
    }, nil
}

type SignatureResponse struct {
    Signature string     `json:"signature"`
    Message   string     `json:"message,omitempty"`
    DataHash  string     `json:"dataHash,omitempty"`
    Address   string     `json:"address"`
    TypedData *TypedData `json:"typedData,omitempty"`
}

type VerificationResponse struct {
    Valid     bool   `json:"valid"`
    Signature string `json:"signature"`
    Message   string `json:"message"`
    Address   string `json:"address"`
}

type TypedData struct {
    Types       map[string][]TypedDataField `json:"types"`
    PrimaryType string                      `json:"primaryType"`
    Domain      TypedDataDomain             `json:"domain"`
    Message     map[string]interface{}      `json:"message"`
}

type TypedDataField struct {
    Name string `json:"name"`
    Type string `json:"type"`
}

type TypedDataDomain struct {
    Name              string `json:"name"`
    Version           string `json:"version"`
    ChainId           int    `json:"chainId"`
    VerifyingContract string `json:"verifyingContract"`
}
```

## 事件监听

### 8.1 事件管理

```go
// events/manager.go
package events

import (
    "sync"

    "your-project/connector"
)

type EventManager struct {
    connector *connector.MetaMaskConnector
    listeners map[string]map[string][]EventHandler
    mutex     sync.RWMutex
}

type EventHandler func(interface{})

func NewEventManager(connector *connector.MetaMaskConnector) *EventManager {
    return &EventManager{
        connector: connector,
        listeners: make(map[string]map[string][]EventHandler),
    }
}

// 订阅事件
func (em *EventManager) Subscribe(sessionID string, eventName string, handler EventHandler) {
    em.mutex.Lock()
    defer em.mutex.Unlock()

    if em.listeners[sessionID] == nil {
        em.listeners[sessionID] = make(map[string][]EventHandler)
    }

    em.listeners[sessionID][eventName] = append(em.listeners[sessionID][eventName], handler)
}

// 取消订阅
func (em *EventManager) Unsubscribe(sessionID string, eventName string) {
    em.mutex.Lock()
    defer em.mutex.Unlock()

    if em.listeners[sessionID] != nil {
        delete(em.listeners[sessionID], eventName)
    }
}

// 触发事件
func (em *EventManager) Emit(sessionID string, eventName string, data interface{}) {
    em.mutex.RLock()
    defer em.mutex.RUnlock()

    if handlers, exists := em.listeners[sessionID][eventName]; exists {
        for _, handler := range handlers {
            go handler(data)
        }
    }
}

// 清理会话事件
func (em *EventManager) CleanupSession(sessionID string) {
    em.mutex.Lock()
    defer em.mutex.Unlock()

    delete(em.listeners, sessionID)
}

// 预定义事件处理器
func (em *EventManager) OnAccountsChanged(sessionID string, handler func([]string)) {
    em.Subscribe(sessionID, "accountsChanged", func(data interface{}) {
        if accounts, ok := data.([]string); ok {
            handler(accounts)
        }
    })
}

func (em *EventManager) OnChainChanged(sessionID string, handler func(string)) {
    em.Subscribe(sessionID, "chainChanged", func(data interface{}) {
        if chainId, ok := data.(string); ok {
            handler(chainId)
        }
    })
}

func (em *EventManager) OnConnect(sessionID string, handler func(ConnectInfo)) {
    em.Subscribe(sessionID, "connect", func(data interface{}) {
        if connectInfo, ok := data.(ConnectInfo); ok {
            handler(connectInfo)
        }
    })
}

func (em *EventManager) OnDisconnect(sessionID string, handler func(DisconnectInfo)) {
    em.Subscribe(sessionID, "disconnect", func(data interface{}) {
        if disconnectInfo, ok := data.(DisconnectInfo); ok {
            handler(disconnectInfo)
        }
    })
}

type ConnectInfo struct {
    ChainId string `json:"chainId"`
}

type DisconnectInfo struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
}
```

## 实际应用

### 9.1 完整MetaMask连接器应用

```go
// main.go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/ethereum/go-ethereum/common"

    "your-project/config"
    "your-project/connector"
    "your-project/account"
    "your-project/network"
    "your-project/transaction"
    "your-project/signature"
    "your-project/events"
)

func main() {
    // 创建MetaMask配置
    cfg := config.DefaultMetaMaskConfig()

    // 创建连接器
    metaMaskConnector := connector.NewMetaMaskConnector(cfg)

    // 创建管理器
    accountManager := account.NewAccountManager(metaMaskConnector)
    networkManager := network.NewNetworkManager(metaMaskConnector)
    transactionSigner := transaction.NewTransactionSigner(metaMaskConnector)
    signatureManager := signature.NewSignatureManager(metaMaskConnector)
    eventManager := events.NewEventManager(metaMaskConnector)

    // 启动连接器服务
    go func() {
        fmt.Printf("MetaMask连接器启动在 %s:%d\n", cfg.Host, cfg.Port)
        if err := metaMaskConnector.Start(); err != nil {
            log.Fatal("启动连接器失败:", err)
        }
    }()

    // 等待服务启动
    time.Sleep(2 * time.Second)

    // 模拟客户端连接
    sessionID := "demo-session-123"

    // 账户操作示例
    fmt.Println("=== 账户操作示例 ===")
    
    // 请求连接账户
    accounts, err := accountManager.RequestAccounts(sessionID)
    if err != nil {
        log.Printf("请求账户失败: %v", err)
    } else {
        fmt.Printf("连接的账户: %+v\n", accounts)
        
        // 获取主账户余额
        if accounts.Primary != "" {
            address := common.HexToAddress(accounts.Primary)
            balance, err := accountManager.GetBalance(sessionID, address)
            if err != nil {
                log.Printf("获取余额失败: %v", err)
            } else {
                fmt.Printf("账户余额: %s\n", balance.Balance)
            }
        }
    }

    // 网络操作示例
    fmt.Println("\n=== 网络操作示例 ===")
    
    // 获取当前网络
    chainInfo, err := networkManager.GetChainId(sessionID)
    if err != nil {
        log.Printf("获取链ID失败: %v", err)
    } else {
        fmt.Printf("当前链ID: %s\n", chainInfo.ChainId)
    }

    // 切换到Polygon网络
    err = networkManager.SwitchToPolygon(sessionID)
    if err != nil {
        log.Printf("切换到Polygon失败: %v", err)
    } else {
        fmt.Println("成功切换到Polygon网络")
    }

    // 交易操作示例
    fmt.Println("\n=== 交易操作示例 ===")
    
    if accounts != nil && accounts.Primary != "" {
        // 创建交易请求
        txRequest := transaction.TransactionRequest{
            From:  accounts.Primary,
            To:    "0x742d35Cc6634C0532925a3b8D4C9db96c4b4d8b6",
            Value: "0x16345785D8A0000", // 0.1 ETH in wei
        }

        // 估算Gas费用
        gasEstimate, err := transactionSigner.EstimateGas(sessionID, txRequest)
        if err != nil {
            log.Printf("估算Gas失败: %v", err)
        } else {
            fmt.Printf("估算Gas: %s\n", gasEstimate.GasLimit)
            txRequest.Gas = gasEstimate.GasLimit
        }

        // 获取Gas价格
        gasPrice, err := transactionSigner.GetGasPrice(sessionID)
        if err != nil {
            log.Printf("获取Gas价格失败: %v", err)
        } else {
            fmt.Printf("Gas价格: %s\n", gasPrice.GasPrice)
            txRequest.GasPrice = gasPrice.GasPrice
        }

        // 发送交易
        txResponse, err := transactionSigner.SendTransaction(sessionID, txRequest)
        if err != nil {
            log.Printf("发送交易失败: %v", err)
        } else {
            fmt.Printf("交易已发送: %s\n", txResponse.Hash)
        }
    }

    // 签名操作示例
    fmt.Println("\n=== 签名操作示例 ===")
    
    if accounts != nil && accounts.Primary != "" {
        message := "Hello, MetaMask!"
        
        // 个人签名
        sigResponse, err := signatureManager.PersonalSign(sessionID, message, accounts.Primary)
        if err != nil {
            log.Printf("个人签名失败: %v", err)
        } else {
            fmt.Printf("签名结果: %s\n", sigResponse.Signature)
            
            // 验证签名
            verification, err := signatureManager.VerifySignature(
                sigResponse.Signature,
                message,
                accounts.Primary,
            )
            if err != nil {
                log.Printf("验证签名失败: %v", err)
            } else {
                fmt.Printf("签名验证: %t\n", verification.Valid)
            }
        }

        // EIP-712类型化数据签名
        typedData := signature.TypedData{
            Types: map[string][]signature.TypedDataField{
                "EIP712Domain": {
                    {Name: "name", Type: "string"},
                    {Name: "version", Type: "string"},
                    {Name: "chainId", Type: "uint256"},
                },
                "Person": {
                    {Name: "name", Type: "string"},
                    {Name: "wallet", Type: "address"},
                },
            },
            PrimaryType: "Person",
            Domain: signature.TypedDataDomain{
                Name:    "MyDApp",
                Version: "1",
                ChainId: 1,
            },
            Message: map[string]interface{}{
                "name":   "Alice",
                "wallet": accounts.Primary,
            },
        }

        typedSigResponse, err := signatureManager.SignTypedData(sessionID, accounts.Primary, typedData)
        if err != nil {
            log.Printf("类型化数据签名失败: %v", err)
        } else {
            fmt.Printf("类型化数据签名: %s\n", typedSigResponse.Signature)
        }
    }

    // 事件监听示例
    fmt.Println("\n=== 事件监听示例 ===")
    
    // 监听账户变化
    eventManager.OnAccountsChanged(sessionID, func(newAccounts []string) {
        fmt.Printf("账户已变化: %+v\n", newAccounts)
    })

    // 监听网络变化
    eventManager.OnChainChanged(sessionID, func(newChainId string) {
        fmt.Printf("网络已变化: %s\n", newChainId)
    })

    // 监听连接状态
    eventManager.OnConnect(sessionID, func(connectInfo events.ConnectInfo) {
        fmt.Printf("已连接到链: %s\n", connectInfo.ChainId)
    })

    eventManager.OnDisconnect(sessionID, func(disconnectInfo events.DisconnectInfo) {
        fmt.Printf("连接已断开: %s (代码: %d)\n", disconnectInfo.Message, disconnectInfo.Code)
    })

    fmt.Println("\nMetaMask连接器演示运行中...")
    fmt.Println("请在浏览器中访问 http://localhost:8080 进行测试")
    fmt.Println("WebSocket端点: ws://localhost:8080/ws")
    fmt.Println("按Ctrl+C退出")

    // 保持服务运行
    select {}
}
```
