# Chainlink Go 详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [环境准备](#环境准备)
3. [价格预言机](#价格预言机)
4. [VRF随机数](#VRF随机数)
5. [Automation自动化](#Automation自动化)
6. [Functions计算](#Functions计算)
7. [CCIP跨链](#CCIP跨链)
8. [自定义预言机](#自定义预言机)
9. [实际应用](#实际应用)

## 基础概念

### 1.1 Chainlink简介

Chainlink是去中心化预言机网络，为智能合约提供可靠的外部数据源、随机数生成、自动化执行等服务。

```bash
# 安装Chainlink相关依赖
go get github.com/ethereum/go-ethereum
go get github.com/smartcontractkit/chainlink
go get github.com/shopspring/decimal
```

### 1.2 核心服务

```go
// 主要包导入
import (
    "context"
    "math/big"
    "strings"
    
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
)

// 核心服务：
// - Price Feeds: 价格数据预言机
// - VRF: 可验证随机函数
// - Automation: 智能合约自动化
// - Functions: 去中心化计算
// - CCIP: 跨链互操作协议
```

## 环境准备

### 2.1 Chainlink配置

```go
// config/chainlink.go
package config

import (
    "math/big"
    "time"
    
    "github.com/ethereum/go-ethereum/common"
)

type ChainlinkConfig struct {
    // 网络配置
    NetworkName    string
    ChainID        *big.Int
    EthereumRPC    string
    
    // 合约地址
    PriceFeedRegistry common.Address
    VRFCoordinator   common.Address
    AutomationRegistry common.Address
    FunctionsRouter  common.Address
    CCIPRouter       common.Address
    
    // VRF配置
    VRFKeyHash       [32]byte
    VRFSubscriptionID uint64
    
    // 超时配置
    Timeout          time.Duration
    ConfirmationBlocks uint64
    
    // 费用配置
    GasLimit         uint64
    GasPrice         *big.Int
}

func DefaultChainlinkConfig() *ChainlinkConfig {
    return &ChainlinkConfig{
        NetworkName:    "ethereum",
        ChainID:        big.NewInt(1),
        EthereumRPC:    "https://mainnet.infura.io/v3/YOUR_PROJECT_ID",
        
        // 以太坊主网合约地址
        PriceFeedRegistry: common.HexToAddress("0x47Fb2585D2C56Fe188D0E6ec628a38b74fceeedf"),
        VRFCoordinator:   common.HexToAddress("0x271682DEB8C4E0901D1a1550aD2e64D568E69909"),
        AutomationRegistry: common.HexToAddress("0x02777053d6764996e594c3E88AF1D58D5363a2e6"),
        FunctionsRouter:  common.HexToAddress("0x65C939B26b8b2A8b8C8c8c8c8c8c8c8c8c8c8c8c"),
        CCIPRouter:       common.HexToAddress("0x80226fc0Ee2b096224EeAc085Bb9a8cba1146f7D"),
        
        Timeout:          30 * time.Second,
        ConfirmationBlocks: 3,
        GasLimit:         500000,
        GasPrice:         big.NewInt(20000000000), // 20 Gwei
    }
}

// Polygon配置
func PolygonChainlinkConfig() *ChainlinkConfig {
    cfg := DefaultChainlinkConfig()
    cfg.NetworkName = "polygon"
    cfg.ChainID = big.NewInt(137)
    cfg.EthereumRPC = "https://polygon-rpc.com/"
    cfg.PriceFeedRegistry = common.HexToAddress("0xAB594600376Ec9fD91F8e885dADF0CE036862dE0")
    cfg.VRFCoordinator = common.HexToAddress("0xAE975071Be8F8eE67addBC1A82488F1C24858067")
    return cfg
}

// 常用价格对
var (
    ETH_USD = "ETH/USD"
    BTC_USD = "BTC/USD"
    LINK_USD = "LINK/USD"
    MATIC_USD = "MATIC/USD"
    USDC_USD = "USDC/USD"
)
```

## 价格预言机

### 3.1 价格数据获取

```go
// pricefeeds/client.go
package pricefeeds

import (
    "context"
    "fmt"
    "math/big"
    "strings"
    "time"

    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/shopspring/decimal"
    
    "your-project/config"
)

// 价格聚合器ABI（简化版）
const AggregatorV3ABI = `[
    {
        "inputs": [],
        "name": "latestRoundData",
        "outputs": [
            {"name": "roundId", "type": "uint80"},
            {"name": "answer", "type": "int256"},
            {"name": "startedAt", "type": "uint256"},
            {"name": "updatedAt", "type": "uint256"},
            {"name": "answeredInRound", "type": "uint80"}
        ],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [],
        "name": "decimals",
        "outputs": [{"name": "", "type": "uint8"}],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [],
        "name": "description",
        "outputs": [{"name": "", "type": "string"}],
        "stateMutability": "view",
        "type": "function"
    }
]`

type PriceFeedClient struct {
    client *ethclient.Client
    config *config.ChainlinkConfig
    feeds  map[string]common.Address
}

func NewPriceFeedClient(cfg *config.ChainlinkConfig) (*PriceFeedClient, error) {
    client, err := ethclient.Dial(cfg.EthereumRPC)
    if err != nil {
        return nil, fmt.Errorf("连接以太坊节点失败: %v", err)
    }

    // 初始化价格对地址映射
    feeds := make(map[string]common.Address)
    
    // 以太坊主网价格对
    if cfg.ChainID.Cmp(big.NewInt(1)) == 0 {
        feeds["ETH/USD"] = common.HexToAddress("0x5f4eC3Df9cbd43714FE2740f5E3616155c5b8419")
        feeds["BTC/USD"] = common.HexToAddress("0xF4030086522a5bEEa4988F8cA5B36dbC97BeE88c")
        feeds["LINK/USD"] = common.HexToAddress("0x2c1d072e956AFFC0D435Cb7AC38EF18d24d9127c")
        feeds["USDC/USD"] = common.HexToAddress("0x8fFfFfd4AfB6115b954Bd326cbe7B4BA576818f6")
    }
    
    // Polygon价格对
    if cfg.ChainID.Cmp(big.NewInt(137)) == 0 {
        feeds["MATIC/USD"] = common.HexToAddress("0xAB594600376Ec9fD91F8e885dADF0CE036862dE0")
        feeds["ETH/USD"] = common.HexToAddress("0xF9680D99D6C9589e2a93a78A04A279e509205945")
        feeds["BTC/USD"] = common.HexToAddress("0xc907E116054Ad103354f2D350FD2514433D57F6f")
    }

    return &PriceFeedClient{
        client: client,
        config: cfg,
        feeds:  feeds,
    }, nil
}

// 获取最新价格
func (pfc *PriceFeedClient) GetLatestPrice(pair string) (*PriceData, error) {
    feedAddress, exists := pfc.feeds[pair]
    if !exists {
        return nil, fmt.Errorf("不支持的价格对: %s", pair)
    }

    // 解析ABI
    parsedABI, err := abi.JSON(strings.NewReader(AggregatorV3ABI))
    if err != nil {
        return nil, fmt.Errorf("解析ABI失败: %v", err)
    }

    // 创建合约实例
    contract := bind.NewBoundContract(feedAddress, parsedABI, pfc.client, pfc.client, pfc.client)

    // 获取最新轮次数据
    var result []interface{}
    err = contract.Call(nil, &result, "latestRoundData")
    if err != nil {
        return nil, fmt.Errorf("获取价格数据失败: %v", err)
    }

    roundId := result[0].(*big.Int)
    answer := result[1].(*big.Int)
    startedAt := result[2].(*big.Int)
    updatedAt := result[3].(*big.Int)
    answeredInRound := result[4].(*big.Int)

    // 获取小数位数
    var decimalsResult []interface{}
    err = contract.Call(nil, &decimalsResult, "decimals")
    if err != nil {
        return nil, fmt.Errorf("获取小数位数失败: %v", err)
    }
    decimals := decimalsResult[0].(uint8)

    // 获取描述
    var descResult []interface{}
    err = contract.Call(nil, &descResult, "description")
    if err != nil {
        return nil, fmt.Errorf("获取描述失败: %v", err)
    }
    description := descResult[0].(string)

    // 转换价格
    price := decimal.NewFromBigInt(answer, -int32(decimals))

    return &PriceData{
        Pair:            pair,
        Price:           price,
        RoundId:         roundId,
        UpdatedAt:       time.Unix(updatedAt.Int64(), 0),
        StartedAt:       time.Unix(startedAt.Int64(), 0),
        AnsweredInRound: answeredInRound,
        Decimals:        decimals,
        Description:     description,
        FeedAddress:     feedAddress,
    }, nil
}

// 批量获取价格
func (pfc *PriceFeedClient) GetMultiplePrices(pairs []string) (map[string]*PriceData, error) {
    results := make(map[string]*PriceData)
    
    for _, pair := range pairs {
        priceData, err := pfc.GetLatestPrice(pair)
        if err != nil {
            fmt.Printf("获取 %s 价格失败: %v\n", pair, err)
            continue
        }
        results[pair] = priceData
    }

    return results, nil
}

// 获取历史价格
func (pfc *PriceFeedClient) GetHistoricalPrice(pair string, roundId *big.Int) (*PriceData, error) {
    feedAddress, exists := pfc.feeds[pair]
    if !exists {
        return nil, fmt.Errorf("不支持的价格对: %s", pair)
    }

    // 这里需要实现getRoundData方法调用
    // 简化实现，返回最新价格
    return pfc.GetLatestPrice(pair)
}

// 监控价格变化
func (pfc *PriceFeedClient) WatchPriceUpdates(pair string, threshold decimal.Decimal, callback func(*PriceData)) error {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    var lastPrice decimal.Decimal

    for {
        select {
        case <-ticker.C:
            priceData, err := pfc.GetLatestPrice(pair)
            if err != nil {
                fmt.Printf("获取价格失败: %v\n", err)
                continue
            }

            // 检查价格变化
            if !lastPrice.IsZero() {
                change := priceData.Price.Sub(lastPrice).Div(lastPrice).Abs()
                if change.GreaterThan(threshold) {
                    callback(priceData)
                }
            }

            lastPrice = priceData.Price
        }
    }
}

// 计算价格变化率
func (pfc *PriceFeedClient) CalculatePriceChange(pair string, duration time.Duration) (*PriceChange, error) {
    currentPrice, err := pfc.GetLatestPrice(pair)
    if err != nil {
        return nil, err
    }

    // 这里需要实现历史价格获取
    // 简化实现，返回模拟数据
    previousPrice := currentPrice.Price.Mul(decimal.NewFromFloat(0.95))

    change := currentPrice.Price.Sub(previousPrice)
    changePercent := change.Div(previousPrice).Mul(decimal.NewFromInt(100))

    return &PriceChange{
        Pair:           pair,
        CurrentPrice:   currentPrice.Price,
        PreviousPrice:  previousPrice,
        Change:         change,
        ChangePercent:  changePercent,
        Duration:       duration,
    }, nil
}

type PriceData struct {
    Pair            string
    Price           decimal.Decimal
    RoundId         *big.Int
    UpdatedAt       time.Time
    StartedAt       time.Time
    AnsweredInRound *big.Int
    Decimals        uint8
    Description     string
    FeedAddress     common.Address
}

type PriceChange struct {
    Pair          string
    CurrentPrice  decimal.Decimal
    PreviousPrice decimal.Decimal
    Change        decimal.Decimal
    ChangePercent decimal.Decimal
    Duration      time.Duration
}
```

## VRF随机数

### 4.1 VRF客户端

```go
// vrf/client.go
package vrf

import (
    "context"
    "fmt"
    "math/big"
    "strings"

    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/ethclient"
    
    "your-project/config"
    "your-project/wallet"
)

// VRF协调器ABI（简化版）
const VRFCoordinatorABI = `[
    {
        "inputs": [
            {"name": "keyHash", "type": "bytes32"},
            {"name": "subId", "type": "uint64"},
            {"name": "minimumRequestConfirmations", "type": "uint16"},
            {"name": "callbackGasLimit", "type": "uint32"},
            {"name": "numWords", "type": "uint32"}
        ],
        "name": "requestRandomWords",
        "outputs": [{"name": "requestId", "type": "uint256"}],
        "stateMutability": "nonpayable",
        "type": "function"
    }
]`

type VRFClient struct {
    client     *ethclient.Client
    config     *config.ChainlinkConfig
    wallet     *wallet.WalletManager
    coordinator *bind.BoundContract
}

func NewVRFClient(cfg *config.ChainlinkConfig, wallet *wallet.WalletManager) (*VRFClient, error) {
    client, err := ethclient.Dial(cfg.EthereumRPC)
    if err != nil {
        return nil, fmt.Errorf("连接以太坊节点失败: %v", err)
    }

    // 解析VRF协调器ABI
    parsedABI, err := abi.JSON(strings.NewReader(VRFCoordinatorABI))
    if err != nil {
        return nil, fmt.Errorf("解析VRF ABI失败: %v", err)
    }

    // 创建协调器合约实例
    coordinator := bind.NewBoundContract(
        cfg.VRFCoordinator,
        parsedABI,
        client,
        client,
        client,
    )

    return &VRFClient{
        client:      client,
        config:      cfg,
        wallet:      wallet,
        coordinator: coordinator,
    }, nil
}

// 请求随机数
func (vc *VRFClient) RequestRandomWords(numWords uint32, callbackGasLimit uint32) (*RandomRequest, error) {
    // 创建交易选项
    auth, err := vc.wallet.CreateTransactOpts(vc.config.ChainID)
    if err != nil {
        return nil, fmt.Errorf("创建交易选项失败: %v", err)
    }

    auth.GasLimit = vc.config.GasLimit
    auth.GasPrice = vc.config.GasPrice

    // 请求随机数
    tx, err := vc.coordinator.Transact(
        auth,
        "requestRandomWords",
        vc.config.VRFKeyHash,
        vc.config.VRFSubscriptionID,
        uint16(vc.config.ConfirmationBlocks),
        callbackGasLimit,
        numWords,
    )
    if err != nil {
        return nil, fmt.Errorf("请求随机数失败: %v", err)
    }

    return &RandomRequest{
        TxHash:            tx.Hash(),
        NumWords:          numWords,
        CallbackGasLimit:  callbackGasLimit,
        SubscriptionID:    vc.config.VRFSubscriptionID,
        KeyHash:           vc.config.VRFKeyHash,
    }, nil
}

// 创建VRF消费者合约
func (vc *VRFClient) DeployVRFConsumer() (*types.Transaction, common.Address, error) {
    // 这里需要部署VRF消费者合约
    // 简化实现，返回模拟数据
    
    return &types.Transaction{}, common.Address{}, fmt.Errorf("需要实现VRF消费者合约部署")
}

// 监听随机数结果
func (vc *VRFClient) WatchRandomnessFulfilled(requestId *big.Int, callback func(*RandomResult)) error {
    // 这里需要监听RandomWordsFulfilled事件
    // 简化实现
    
    return fmt.Errorf("需要实现随机数结果监听")
}

// 验证随机数
func (vc *VRFClient) VerifyRandomness(proof []byte, seed *big.Int) (bool, error) {
    // VRF验证逻辑
    // 简化实现
    
    return true, nil
}

type RandomRequest struct {
    TxHash           common.Hash
    NumWords         uint32
    CallbackGasLimit uint32
    SubscriptionID   uint64
    KeyHash          [32]byte
}

type RandomResult struct {
    RequestId    *big.Int
    RandomWords  []*big.Int
    Fulfilled    bool
    BlockNumber  uint64
    TxHash       common.Hash
}
```

## Automation自动化

### 5.1 自动化客户端

```go
// automation/client.go
package automation

import (
    "context"
    "fmt"
    "math/big"
    "strings"

    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/ethclient"
    
    "your-project/config"
    "your-project/wallet"
)

// 自动化注册表ABI（简化版）
const AutomationRegistryABI = `[
    {
        "inputs": [
            {"name": "name", "type": "string"},
            {"name": "encryptedEmail", "type": "bytes"},
            {"name": "upkeepContract", "type": "address"},
            {"name": "gasLimit", "type": "uint32"},
            {"name": "adminAddress", "type": "address"},
            {"name": "checkData", "type": "bytes"},
            {"name": "amount", "type": "uint96"},
            {"name": "source", "type": "uint8"}
        ],
        "name": "registerUpkeep",
        "outputs": [{"name": "id", "type": "uint256"}],
        "stateMutability": "nonpayable",
        "type": "function"
    }
]`

type AutomationClient struct {
    client   *ethclient.Client
    config   *config.ChainlinkConfig
    wallet   *wallet.WalletManager
    registry *bind.BoundContract
}

func NewAutomationClient(cfg *config.ChainlinkConfig, wallet *wallet.WalletManager) (*AutomationClient, error) {
    client, err := ethclient.Dial(cfg.EthereumRPC)
    if err != nil {
        return nil, fmt.Errorf("连接以太坊节点失败: %v", err)
    }

    // 解析自动化注册表ABI
    parsedABI, err := abi.JSON(strings.NewReader(AutomationRegistryABI))
    if err != nil {
        return nil, fmt.Errorf("解析Automation ABI失败: %v", err)
    }

    // 创建注册表合约实例
    registry := bind.NewBoundContract(
        cfg.AutomationRegistry,
        parsedABI,
        client,
        client,
        client,
    )

    return &AutomationClient{
        client:   client,
        config:   cfg,
        wallet:   wallet,
        registry: registry,
    }, nil
}

// 注册自动化任务
func (ac *AutomationClient) RegisterUpkeep(params UpkeepParams) (*UpkeepRegistration, error) {
    // 创建交易选项
    auth, err := ac.wallet.CreateTransactOpts(ac.config.ChainID)
    if err != nil {
        return nil, fmt.Errorf("创建交易选项失败: %v", err)
    }

    auth.GasLimit = ac.config.GasLimit
    auth.GasPrice = ac.config.GasPrice

    // 注册自动化任务
    tx, err := ac.registry.Transact(
        auth,
        "registerUpkeep",
        params.Name,
        params.EncryptedEmail,
        params.UpkeepContract,
        params.GasLimit,
        params.AdminAddress,
        params.CheckData,
        params.Amount,
        params.Source,
    )
    if err != nil {
        return nil, fmt.Errorf("注册自动化任务失败: %v", err)
    }

    return &UpkeepRegistration{
        TxHash: tx.Hash(),
        Params: params,
    }, nil
}

// 创建自动化合约
func (ac *AutomationClient) DeployAutomationContract(contractCode []byte) (*types.Transaction, common.Address, error) {
    // 部署自动化合约
    // 简化实现
    
    return &types.Transaction{}, common.Address{}, fmt.Errorf("需要实现自动化合约部署")
}

// 监控自动化执行
func (ac *AutomationClient) WatchUpkeepPerformed(upkeepId *big.Int, callback func(*UpkeepExecution)) error {
    // 监听UpkeepPerformed事件
    // 简化实现
    
    return fmt.Errorf("需要实现自动化执行监听")
}

type UpkeepParams struct {
    Name            string
    EncryptedEmail  []byte
    UpkeepContract  common.Address
    GasLimit        uint32
    AdminAddress    common.Address
    CheckData       []byte
    Amount          *big.Int
    Source          uint8
}

type UpkeepRegistration struct {
    TxHash common.Hash
    Params UpkeepParams
}

type UpkeepExecution struct {
    UpkeepId     *big.Int
    Success      bool
    GasUsed      uint64
    BlockNumber  uint64
    TxHash       common.Hash
}
```

## Functions计算

### 6.1 Functions客户端

```go
// functions/client.go
package functions

import (
    "context"
    "fmt"
    "math/big"
    "strings"

    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/ethclient"
    
    "your-project/config"
    "your-project/wallet"
)

// Functions路由器ABI（简化版）
const FunctionsRouterABI = `[
    {
        "inputs": [
            {"name": "subscriptionId", "type": "uint64"},
            {"name": "data", "type": "bytes"},
            {"name": "dataVersion", "type": "uint16"},
            {"name": "callbackGasLimit", "type": "uint32"},
            {"name": "donId", "type": "bytes32"}
        ],
        "name": "sendRequest",
        "outputs": [{"name": "requestId", "type": "bytes32"}],
        "stateMutability": "nonpayable",
        "type": "function"
    }
]`

type FunctionsClient struct {
    client *ethclient.Client
    config *config.ChainlinkConfig
    wallet *wallet.WalletManager
    router *bind.BoundContract
}

func NewFunctionsClient(cfg *config.ChainlinkConfig, wallet *wallet.WalletManager) (*FunctionsClient, error) {
    client, err := ethclient.Dial(cfg.EthereumRPC)
    if err != nil {
        return nil, fmt.Errorf("连接以太坊节点失败: %v", err)
    }

    // 解析Functions路由器ABI
    parsedABI, err := abi.JSON(strings.NewReader(FunctionsRouterABI))
    if err != nil {
        return nil, fmt.Errorf("解析Functions ABI失败: %v", err)
    }

    // 创建路由器合约实例
    router := bind.NewBoundContract(
        cfg.FunctionsRouter,
        parsedABI,
        client,
        client,
        client,
    )

    return &FunctionsClient{
        client: client,
        config: cfg,
        wallet: wallet,
        router: router,
    }, nil
}

// 发送Functions请求
func (fc *FunctionsClient) SendRequest(params FunctionRequest) (*FunctionExecution, error) {
    // 创建交易选项
    auth, err := fc.wallet.CreateTransactOpts(fc.config.ChainID)
    if err != nil {
        return nil, fmt.Errorf("创建交易选项失败: %v", err)
    }

    auth.GasLimit = fc.config.GasLimit
    auth.GasPrice = fc.config.GasPrice

    // 发送Functions请求
    tx, err := fc.router.Transact(
        auth,
        "sendRequest",
        params.SubscriptionId,
        params.Data,
        params.DataVersion,
        params.CallbackGasLimit,
        params.DonId,
    )
    if err != nil {
        return nil, fmt.Errorf("发送Functions请求失败: %v", err)
    }

    return &FunctionExecution{
        TxHash:  tx.Hash(),
        Request: params,
    }, nil
}

// 创建JavaScript代码
func (fc *FunctionsClient) CreateJavaScriptCode(apiUrl, apiKey string) string {
    return fmt.Sprintf(`
        const apiResponse = await Functions.makeHttpRequest({
            url: "%s",
            headers: {
                "Authorization": "Bearer %s"
            }
        });
        
        if (apiResponse.error) {
            throw Error("Request failed");
        }
        
        const data = apiResponse.data;
        return Functions.encodeUint256(Math.round(data.price * 100));
    `, apiUrl, apiKey)
}

// 监听Functions响应
func (fc *FunctionsClient) WatchFunctionResponse(requestId [32]byte, callback func(*FunctionResponse)) error {
    // 监听ResponseReceived事件
    // 简化实现
    
    return fmt.Errorf("需要实现Functions响应监听")
}

type FunctionRequest struct {
    SubscriptionId    uint64
    Data              []byte
    DataVersion       uint16
    CallbackGasLimit  uint32
    DonId             [32]byte
}

type FunctionExecution struct {
    TxHash  common.Hash
    Request FunctionRequest
}

type FunctionResponse struct {
    RequestId [32]byte
    Response  []byte
    Error     []byte
    Success   bool
}
```

## CCIP跨链

### 7.1 CCIP客户端

```go
// ccip/client.go
package ccip

import (
    "context"
    "fmt"
    "math/big"
    "strings"

    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/ethclient"
    
    "your-project/config"
    "your-project/wallet"
)

// CCIP路由器ABI（简化版）
const CCIPRouterABI = `[
    {
        "inputs": [
            {"name": "destinationChainSelector", "type": "uint64"},
            {"name": "message", "type": "tuple", "components": [
                {"name": "receiver", "type": "bytes"},
                {"name": "data", "type": "bytes"},
                {"name": "tokenAmounts", "type": "tuple[]", "components": [
                    {"name": "token", "type": "address"},
                    {"name": "amount", "type": "uint256"}
                ]},
                {"name": "feeToken", "type": "address"},
                {"name": "extraArgs", "type": "bytes"}
            ]}
        ],
        "name": "ccipSend",
        "outputs": [{"name": "messageId", "type": "bytes32"}],
        "stateMutability": "payable",
        "type": "function"
    }
]`

type CCIPClient struct {
    client *ethclient.Client
    config *config.ChainlinkConfig
    wallet *wallet.WalletManager
    router *bind.BoundContract
}

func NewCCIPClient(cfg *config.ChainlinkConfig, wallet *wallet.WalletManager) (*CCIPClient, error) {
    client, err := ethclient.Dial(cfg.EthereumRPC)
    if err != nil {
        return nil, fmt.Errorf("连接以太坊节点失败: %v", err)
    }

    // 解析CCIP路由器ABI
    parsedABI, err := abi.JSON(strings.NewReader(CCIPRouterABI))
    if err != nil {
        return nil, fmt.Errorf("解析CCIP ABI失败: %v", err)
    }

    // 创建路由器合约实例
    router := bind.NewBoundContract(
        cfg.CCIPRouter,
        parsedABI,
        client,
        client,
        client,
    )

    return &CCIPClient{
        client: client,
        config: cfg,
        wallet: wallet,
        router: router,
    }, nil
}

// 发送跨链消息
func (cc *CCIPClient) SendCrossChainMessage(params CrossChainMessage) (*MessageExecution, error) {
    // 创建交易选项
    auth, err := cc.wallet.CreateTransactOpts(cc.config.ChainID)
    if err != nil {
        return nil, fmt.Errorf("创建交易选项失败: %v", err)
    }

    auth.GasLimit = cc.config.GasLimit
    auth.GasPrice = cc.config.GasPrice
    auth.Value = params.Fee

    // 构建消息结构
    message := struct {
        Receiver     []byte
        Data         []byte
        TokenAmounts []struct {
            Token  common.Address
            Amount *big.Int
        }
        FeeToken   common.Address
        ExtraArgs  []byte
    }{
        Receiver:     params.Receiver,
        Data:         params.Data,
        TokenAmounts: params.TokenAmounts,
        FeeToken:     params.FeeToken,
        ExtraArgs:    params.ExtraArgs,
    }

    // 发送跨链消息
    tx, err := cc.router.Transact(
        auth,
        "ccipSend",
        params.DestinationChainSelector,
        message,
    )
    if err != nil {
        return nil, fmt.Errorf("发送跨链消息失败: %v", err)
    }

    return &MessageExecution{
        TxHash:  tx.Hash(),
        Message: params,
    }, nil
}

// 监听跨链消息
func (cc *CCIPClient) WatchCrossChainMessages(callback func(*CrossChainEvent)) error {
    // 监听CCIPSendRequested事件
    // 简化实现
    
    return fmt.Errorf("需要实现跨链消息监听")
}

type CrossChainMessage struct {
    DestinationChainSelector uint64
    Receiver                 []byte
    Data                     []byte
    TokenAmounts            []struct {
        Token  common.Address
        Amount *big.Int
    }
    FeeToken   common.Address
    ExtraArgs  []byte
    Fee        *big.Int
}

type MessageExecution struct {
    TxHash  common.Hash
    Message CrossChainMessage
}

type CrossChainEvent struct {
    MessageId [32]byte
    Source    uint64
    Receiver  common.Address
    Data      []byte
}
```

## 自定义预言机

### 8.1 预言机节点

```go
// oracle/node.go
package oracle

import (
    "context"
    "fmt"
    "math/big"
    "time"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/ethclient"
    
    "your-project/config"
    "your-project/wallet"
)

type OracleNode struct {
    client     *ethclient.Client
    config     *config.ChainlinkConfig
    wallet     *wallet.WalletManager
    dataFeeds  map[string]DataFeed
    isRunning  bool
}

type DataFeed struct {
    Name        string
    Source      string
    UpdateInterval time.Duration
    LastUpdate  time.Time
    Value       *big.Int
}

func NewOracleNode(cfg *config.ChainlinkConfig, wallet *wallet.WalletManager) (*OracleNode, error) {
    client, err := ethclient.Dial(cfg.EthereumRPC)
    if err != nil {
        return nil, fmt.Errorf("连接以太坊节点失败: %v", err)
    }

    return &OracleNode{
        client:    client,
        config:    cfg,
        wallet:    wallet,
        dataFeeds: make(map[string]DataFeed),
        isRunning: false,
    }, nil
}

// 添加数据源
func (on *OracleNode) AddDataFeed(feed DataFeed) {
    on.dataFeeds[feed.Name] = feed
}

// 启动预言机节点
func (on *OracleNode) Start() error {
    if on.isRunning {
        return fmt.Errorf("预言机节点已在运行")
    }

    on.isRunning = true
    
    // 为每个数据源启动更新goroutine
    for name, feed := range on.dataFeeds {
        go on.updateDataFeed(name, feed)
    }

    return nil
}

// 停止预言机节点
func (on *OracleNode) Stop() {
    on.isRunning = false
}

// 更新数据源
func (on *OracleNode) updateDataFeed(name string, feed DataFeed) {
    ticker := time.NewTicker(feed.UpdateInterval)
    defer ticker.Stop()

    for on.isRunning {
        select {
        case <-ticker.C:
            // 获取外部数据
            value, err := on.fetchExternalData(feed.Source)
            if err != nil {
                fmt.Printf("获取外部数据失败 %s: %v\n", name, err)
                continue
            }

            // 更新链上数据
            err = on.updateOnChainData(name, value)
            if err != nil {
                fmt.Printf("更新链上数据失败 %s: %v\n", name, err)
                continue
            }

            // 更新本地记录
            feed.Value = value
            feed.LastUpdate = time.Now()
            on.dataFeeds[name] = feed

            fmt.Printf("数据源 %s 已更新: %s\n", name, value.String())
        }
    }
}

// 获取外部数据
func (on *OracleNode) fetchExternalData(source string) (*big.Int, error) {
    // 这里实现从外部API获取数据的逻辑
    // 简化实现，返回模拟数据
    
    return big.NewInt(time.Now().Unix()), nil
}

// 更新链上数据
func (on *OracleNode) updateOnChainData(feedName string, value *big.Int) error {
    // 这里实现更新智能合约数据的逻辑
    // 需要调用预言机合约的更新方法
    
    return nil
}

// 获取数据源状态
func (on *OracleNode) GetDataFeedStatus(name string) (*DataFeedStatus, error) {
    feed, exists := on.dataFeeds[name]
    if !exists {
        return nil, fmt.Errorf("数据源不存在: %s", name)
    }

    return &DataFeedStatus{
        Name:           feed.Name,
        Value:          feed.Value,
        LastUpdate:     feed.LastUpdate,
        UpdateInterval: feed.UpdateInterval,
        IsActive:       on.isRunning,
    }, nil
}

type DataFeedStatus struct {
    Name           string
    Value          *big.Int
    LastUpdate     time.Time
    UpdateInterval time.Duration
    IsActive       bool
}
```

## 实际应用

### 9.1 完整Chainlink应用

```go
// main.go
package main

import (
    "fmt"
    "log"
    "math/big"
    "time"

    "github.com/shopspring/decimal"

    "your-project/config"
    "your-project/pricefeeds"
    "your-project/vrf"
    "your-project/automation"
    "your-project/functions"
    "your-project/ccip"
    "your-project/oracle"
    "your-project/wallet"
)

func main() {
    // 创建Chainlink配置
    cfg := config.DefaultChainlinkConfig()

    // 创建钱包
    walletManager, err := wallet.NewWalletManager("your_private_key_here")
    if err != nil {
        log.Fatal("创建钱包失败:", err)
    }

    fmt.Printf("钱包地址: %s\n", walletManager.GetAddress().Hex())

    // 价格预言机示例
    fmt.Println("=== 价格预言机示例 ===")
    
    priceFeedClient, err := pricefeeds.NewPriceFeedClient(cfg)
    if err != nil {
        log.Fatal("创建价格预言机客户端失败:", err)
    }

    // 获取单个价格
    ethPrice, err := priceFeedClient.GetLatestPrice("ETH/USD")
    if err != nil {
        log.Printf("获取ETH价格失败: %v", err)
    } else {
        fmt.Printf("ETH/USD价格: $%s\n", ethPrice.Price.String())
        fmt.Printf("更新时间: %s\n", ethPrice.UpdatedAt.Format("2006-01-02 15:04:05"))
        fmt.Printf("轮次ID: %s\n", ethPrice.RoundId.String())
    }

    // 批量获取价格
    pairs := []string{"ETH/USD", "BTC/USD", "LINK/USD"}
    prices, err := priceFeedClient.GetMultiplePrices(pairs)
    if err != nil {
        log.Printf("批量获取价格失败: %v", err)
    } else {
        fmt.Println("批量价格数据:")
        for pair, priceData := range prices {
            fmt.Printf("  %s: $%s\n", pair, priceData.Price.String())
        }
    }

    // 价格变化监控
    go func() {
        threshold := decimal.NewFromFloat(0.01) // 1%变化阈值
        err := priceFeedClient.WatchPriceUpdates("ETH/USD", threshold, func(priceData *pricefeeds.PriceData) {
            fmt.Printf("ETH价格显著变化: $%s (时间: %s)\n", 
                priceData.Price.String(), 
                priceData.UpdatedAt.Format("15:04:05"))
        })
        if err != nil {
            log.Printf("监控价格变化失败: %v", err)
        }
    }()

    // VRF随机数示例
    fmt.Println("\n=== VRF随机数示例 ===")
    
    vrfClient, err := vrf.NewVRFClient(cfg, walletManager)
    if err != nil {
        log.Printf("创建VRF客户端失败: %v", err)
    } else {
        // 请求随机数
        randomRequest, err := vrfClient.RequestRandomWords(3, 100000)
        if err != nil {
            log.Printf("请求随机数失败: %v", err)
        } else {
            fmt.Printf("随机数请求已发送:\n")
            fmt.Printf("  交易哈希: %s\n", randomRequest.TxHash.Hex())
            fmt.Printf("  请求数量: %d\n", randomRequest.NumWords)
            fmt.Printf("  订阅ID: %d\n", randomRequest.SubscriptionID)
        }
    }

    // Automation自动化示例
    fmt.Println("\n=== Automation自动化示例 ===")
    
    automationClient, err := automation.NewAutomationClient(cfg, walletManager)
    if err != nil {
        log.Printf("创建Automation客户端失败: %v", err)
    } else {
        // 注册自动化任务
        upkeepParams := automation.UpkeepParams{
            Name:           "Price Monitor",
            EncryptedEmail: []byte("encrypted_email"),
            UpkeepContract: walletManager.GetAddress(),
            GasLimit:       500000,
            AdminAddress:   walletManager.GetAddress(),
            CheckData:      []byte("check_data"),
            Amount:         big.NewInt(1000000000000000000), // 1 ETH
            Source:         0,
        }

        upkeepReg, err := automationClient.RegisterUpkeep(upkeepParams)
        if err != nil {
            log.Printf("注册自动化任务失败: %v", err)
        } else {
            fmt.Printf("自动化任务注册成功:\n")
            fmt.Printf("  交易哈希: %s\n", upkeepReg.TxHash.Hex())
            fmt.Printf("  任务名称: %s\n", upkeepReg.Params.Name)
        }
    }

    // Functions计算示例
    fmt.Println("\n=== Functions计算示例 ===")
    
    functionsClient, err := functions.NewFunctionsClient(cfg, walletManager)
    if err != nil {
        log.Printf("创建Functions客户端失败: %v", err)
    } else {
        // 创建JavaScript代码
        jsCode := functionsClient.CreateJavaScriptCode(
            "https://api.coinbase.com/v2/exchange-rates?currency=ETH",
            "your_api_key",
        )
        fmt.Printf("JavaScript代码:\n%s\n", jsCode)

        // 发送Functions请求
        functionRequest := functions.FunctionRequest{
            SubscriptionId:   1,
            Data:             []byte(jsCode),
            DataVersion:      1,
            CallbackGasLimit: 300000,
            DonId:            [32]byte{},
        }

        functionExec, err := functionsClient.SendRequest(functionRequest)
        if err != nil {
            log.Printf("发送Functions请求失败: %v", err)
        } else {
            fmt.Printf("Functions请求已发送: %s\n", functionExec.TxHash.Hex())
        }
    }

    // CCIP跨链示例
    fmt.Println("\n=== CCIP跨链示例 ===")
    
    ccipClient, err := ccip.NewCCIPClient(cfg, walletManager)
    if err != nil {
        log.Printf("创建CCIP客户端失败: %v", err)
    } else {
        // 发送跨链消息
        crossChainMsg := ccip.CrossChainMessage{
            DestinationChainSelector: 5009297550715157269, // Polygon链选择器
            Receiver:                 walletManager.GetAddress().Bytes(),
            Data:                     []byte("Hello from Ethereum!"),
            TokenAmounts:            []struct {
                Token  common.Address
                Amount *big.Int
            }{},
            FeeToken:  common.Address{}, // 使用ETH支付费用
            ExtraArgs: []byte{},
            Fee:       big.NewInt(100000000000000000), // 0.1 ETH
        }

        msgExec, err := ccipClient.SendCrossChainMessage(crossChainMsg)
        if err != nil {
            log.Printf("发送跨链消息失败: %v", err)
        } else {
            fmt.Printf("跨链消息已发送: %s\n", msgExec.TxHash.Hex())
        }
    }

    // 自定义预言机示例
    fmt.Println("\n=== 自定义预言机示例 ===")
    
    oracleNode, err := oracle.NewOracleNode(cfg, walletManager)
    if err != nil {
        log.Printf("创建预言机节点失败: %v", err)
    } else {
        // 添加数据源
        weatherFeed := oracle.DataFeed{
            Name:           "Weather_Temperature",
            Source:         "https://api.weather.com/temperature",
            UpdateInterval: 10 * time.Minute,
        }
        oracleNode.AddDataFeed(weatherFeed)

        stockFeed := oracle.DataFeed{
            Name:           "Stock_AAPL",
            Source:         "https://api.stocks.com/AAPL",
            UpdateInterval: 5 * time.Minute,
        }
        oracleNode.AddDataFeed(stockFeed)

        // 启动预言机节点
        err = oracleNode.Start()
        if err != nil {
            log.Printf("启动预言机节点失败: %v", err)
        } else {
            fmt.Println("自定义预言机节点已启动")
            
            // 运行一段时间后检查状态
            time.Sleep(5 * time.Second)
            
            status, err := oracleNode.GetDataFeedStatus("Weather_Temperature")
            if err != nil {
                log.Printf("获取数据源状态失败: %v", err)
            } else {
                fmt.Printf("天气数据源状态:\n")
                fmt.Printf("  名称: %s\n", status.Name)
                fmt.Printf("  活跃: %t\n", status.IsActive)
                fmt.Printf("  更新间隔: %s\n", status.UpdateInterval.String())
                if status.Value != nil {
                    fmt.Printf("  当前值: %s\n", status.Value.String())
                }
            }
        }
    }

    fmt.Println("\nChainlink服务演示运行中...")
    fmt.Println("按Ctrl+C退出")

    // 保持程序运行以观察价格监控
    time.Sleep(30 * time.Second)

    fmt.Println("Chainlink操作演示完成!")
}
```
