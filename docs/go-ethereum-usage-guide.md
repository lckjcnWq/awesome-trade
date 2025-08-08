# Go-Ethereum (Geth) 详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [环境准备](#环境准备)
3. [客户端连接](#客户端连接)
4. [账户管理](#账户管理)
5. [交易操作](#交易操作)
6. [智能合约](#智能合约)
7. [事件监听](#事件监听)
8. [高级功能](#高级功能)
9. [实际应用](#实际应用)

## 基础概念

### 1.1 Go-Ethereum简介

Go-Ethereum是以太坊的Go语言实现，提供了完整的以太坊节点实现和EVM执行环境。

```bash
# 安装go-ethereum
go get github.com/ethereum/go-ethereum
go get github.com/ethereum/go-ethereum/ethclient
go get github.com/ethereum/go-ethereum/accounts/abi
```

### 1.2 核心组件

```go
// 主要包导入
import (
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/crypto"
)
```

## 环境准备

### 2.1 连接配置

```go
// config/ethereum.go
package config

import (
    "context"
    "math/big"
    
    "github.com/ethereum/go-ethereum/ethclient"
)

type EthereumConfig struct {
    NodeURL     string
    ChainID     *big.Int
    GasLimit    uint64
    GasPrice    *big.Int
    PrivateKey  string
}

// 默认配置
func DefaultConfig() *EthereumConfig {
    return &EthereumConfig{
        NodeURL:    "https://mainnet.infura.io/v3/YOUR_PROJECT_ID", // 主网
        // NodeURL: "https://goerli.infura.io/v3/YOUR_PROJECT_ID",  // 测试网
        // NodeURL: "http://localhost:8545",                        // 本地节点
        ChainID:    big.NewInt(1),     // 主网链ID
        GasLimit:   21000,             // 基础转账gas限制
        GasPrice:   big.NewInt(20000000000), // 20 Gwei
    }
}

// 创建客户端连接
func NewEthereumClient(nodeURL string) (*ethclient.Client, error) {
    client, err := ethclient.Dial(nodeURL)
    if err != nil {
        return nil, err
    }
    
    // 验证连接
    _, err = client.NetworkID(context.Background())
    if err != nil {
        return nil, err
    }
    
    return client, nil
}
```

### 2.2 工具函数

```go
// utils/ethereum.go
package utils

import (
    "crypto/ecdsa"
    "math/big"
    
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/crypto"
)

// Wei转换为Ether
func WeiToEther(wei *big.Int) *big.Float {
    ether := new(big.Float)
    ether.SetString(wei.String())
    return ether.Quo(ether, big.NewFloat(1e18))
}

// Ether转换为Wei
func EtherToWei(ether *big.Float) *big.Int {
    wei := new(big.Float)
    wei.Mul(ether, big.NewFloat(1e18))
    result, _ := wei.Int(nil)
    return result
}

// 从私钥生成地址
func PrivateKeyToAddress(privateKey *ecdsa.PrivateKey) common.Address {
    publicKey := privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    return crypto.PubkeyToAddress(*publicKeyECDSA)
}

// 从私钥字符串生成私钥对象
func HexToPrivateKey(hexKey string) (*ecdsa.PrivateKey, error) {
    return crypto.HexToECDSA(hexKey)
}
```

## 客户端连接

### 3.1 基本连接

```go
// client/ethereum.go
package client

import (
    "context"
    "fmt"
    "log"
    "math/big"

    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/ethereum/go-ethereum/common"
)

type EthereumClient struct {
    client  *ethclient.Client
    chainID *big.Int
}

func NewEthereumClient(nodeURL string) (*EthereumClient, error) {
    client, err := ethclient.Dial(nodeURL)
    if err != nil {
        return nil, fmt.Errorf("连接以太坊节点失败: %v", err)
    }

    // 获取链ID
    chainID, err := client.NetworkID(context.Background())
    if err != nil {
        return nil, fmt.Errorf("获取网络ID失败: %v", err)
    }

    return &EthereumClient{
        client:  client,
        chainID: chainID,
    }, nil
}

// 获取最新区块号
func (ec *EthereumClient) GetLatestBlockNumber() (uint64, error) {
    header, err := ec.client.HeaderByNumber(context.Background(), nil)
    if err != nil {
        return 0, err
    }
    return header.Number.Uint64(), nil
}

// 获取账户余额
func (ec *EthereumClient) GetBalance(address common.Address) (*big.Int, error) {
    balance, err := ec.client.BalanceAt(context.Background(), address, nil)
    if err != nil {
        return nil, err
    }
    return balance, nil
}

// 获取账户nonce
func (ec *EthereumClient) GetNonce(address common.Address) (uint64, error) {
    nonce, err := ec.client.PendingNonceAt(context.Background(), address)
    if err != nil {
        return 0, err
    }
    return nonce, nil
}

// 获取建议的Gas价格
func (ec *EthereumClient) SuggestGasPrice() (*big.Int, error) {
    gasPrice, err := ec.client.SuggestGasPrice(context.Background())
    if err != nil {
        return nil, err
    }
    return gasPrice, nil
}
```

## 账户管理

### 4.1 账户操作

```go
// account/manager.go
package account

import (
    "crypto/ecdsa"
    "crypto/rand"
    "fmt"
    
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/crypto"
)

type Account struct {
    PrivateKey *ecdsa.PrivateKey
    PublicKey  *ecdsa.PublicKey
    Address    common.Address
}

// 创建新账户
func CreateAccount() (*Account, error) {
    privateKey, err := crypto.GenerateKey()
    if err != nil {
        return nil, err
    }

    publicKey := privateKey.Public()
    publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
    if !ok {
        return nil, fmt.Errorf("无法转换公钥")
    }

    address := crypto.PubkeyToAddress(*publicKeyECDSA)

    return &Account{
        PrivateKey: privateKey,
        PublicKey:  publicKeyECDSA,
        Address:    address,
    }, nil
}

// 从私钥导入账户
func ImportAccount(privateKeyHex string) (*Account, error) {
    privateKey, err := crypto.HexToECDSA(privateKeyHex)
    if err != nil {
        return nil, err
    }

    publicKey := privateKey.Public()
    publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
    if !ok {
        return nil, fmt.Errorf("无法转换公钥")
    }

    address := crypto.PubkeyToAddress(*publicKeyECDSA)

    return &Account{
        PrivateKey: privateKey,
        PublicKey:  publicKeyECDSA,
        Address:    address,
    }, nil
}

// 获取私钥十六进制字符串
func (a *Account) PrivateKeyHex() string {
    return fmt.Sprintf("%x", crypto.FromECDSA(a.PrivateKey))
}

// 获取地址字符串
func (a *Account) AddressHex() string {
    return a.Address.Hex()
}
```

## 交易操作

### 5.1 ETH转账

```go
// transaction/transfer.go
package transaction

import (
    "context"
    "crypto/ecdsa"
    "fmt"
    "math/big"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/ethclient"
)

type TransferManager struct {
    client  *ethclient.Client
    chainID *big.Int
}

func NewTransferManager(client *ethclient.Client, chainID *big.Int) *TransferManager {
    return &TransferManager{
        client:  client,
        chainID: chainID,
    }
}

// ETH转账
func (tm *TransferManager) TransferETH(
    privateKey *ecdsa.PrivateKey,
    to common.Address,
    amount *big.Int,
    gasLimit uint64,
    gasPrice *big.Int,
) (*types.Transaction, error) {
    
    fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
    
    // 获取nonce
    nonce, err := tm.client.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        return nil, err
    }

    // 创建交易
    tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, nil)

    // 签名交易
    signedTx, err := types.SignTx(tx, types.NewEIP155Signer(tm.chainID), privateKey)
    if err != nil {
        return nil, err
    }

    // 发送交易
    err = tm.client.SendTransaction(context.Background(), signedTx)
    if err != nil {
        return nil, err
    }

    return signedTx, nil
}

// 等待交易确认
func (tm *TransferManager) WaitForConfirmation(txHash common.Hash) (*types.Receipt, error) {
    for {
        receipt, err := tm.client.TransactionReceipt(context.Background(), txHash)
        if err != nil {
            // 交易还未被打包
            continue
        }
        return receipt, nil
    }
}

// 获取交易详情
func (tm *TransferManager) GetTransaction(txHash common.Hash) (*types.Transaction, bool, error) {
    tx, isPending, err := tm.client.TransactionByHash(context.Background(), txHash)
    if err != nil {
        return nil, false, err
    }
    return tx, isPending, nil
}
```

## 智能合约

### 5.1 合约交互

```go
// contract/erc20.go
package contract

import (
    "context"
    "crypto/ecdsa"
    "math/big"
    "strings"

    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/ethclient"
)

// ERC20合约ABI（简化版）
const ERC20ABI = `[
    {
        "constant": true,
        "inputs": [{"name": "_owner", "type": "address"}],
        "name": "balanceOf",
        "outputs": [{"name": "balance", "type": "uint256"}],
        "type": "function"
    },
    {
        "constant": false,
        "inputs": [
            {"name": "_to", "type": "address"},
            {"name": "_value", "type": "uint256"}
        ],
        "name": "transfer",
        "outputs": [{"name": "", "type": "bool"}],
        "type": "function"
    },
    {
        "constant": true,
        "inputs": [],
        "name": "decimals",
        "outputs": [{"name": "", "type": "uint8"}],
        "type": "function"
    }
]`

type ERC20Contract struct {
    client   *ethclient.Client
    contract *bind.BoundContract
    address  common.Address
    abi      abi.ABI
}

func NewERC20Contract(client *ethclient.Client, contractAddress common.Address) (*ERC20Contract, error) {
    parsedABI, err := abi.JSON(strings.NewReader(ERC20ABI))
    if err != nil {
        return nil, err
    }

    contract := bind.NewBoundContract(contractAddress, parsedABI, client, client, client)

    return &ERC20Contract{
        client:   client,
        contract: contract,
        address:  contractAddress,
        abi:      parsedABI,
    }, nil
}

// 查询余额
func (erc20 *ERC20Contract) BalanceOf(account common.Address) (*big.Int, error) {
    var result []interface{}
    err := erc20.contract.Call(nil, &result, "balanceOf", account)
    if err != nil {
        return nil, err
    }
    
    balance := result[0].(*big.Int)
    return balance, nil
}

// 转账
func (erc20 *ERC20Contract) Transfer(
    privateKey *ecdsa.PrivateKey,
    to common.Address,
    amount *big.Int,
    gasLimit uint64,
    gasPrice *big.Int,
) (*types.Transaction, error) {
    
    fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
    
    // 获取nonce
    nonce, err := erc20.client.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        return nil, err
    }

    // 创建交易选项
    auth := &bind.TransactOpts{
        From:     fromAddress,
        Nonce:    big.NewInt(int64(nonce)),
        GasLimit: gasLimit,
        GasPrice: gasPrice,
        Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
            if address != fromAddress {
                return nil, fmt.Errorf("not authorized to sign this account")
            }
            chainID, err := erc20.client.NetworkID(context.Background())
            if err != nil {
                return nil, err
            }
            return types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
        },
    }

    // 调用transfer方法
    tx, err := erc20.contract.Transact(auth, "transfer", to, amount)
    if err != nil {
        return nil, err
    }

    return tx, nil
}

// 获取小数位数
func (erc20 *ERC20Contract) Decimals() (uint8, error) {
    var result []interface{}
    err := erc20.contract.Call(nil, &result, "decimals")
    if err != nil {
        return 0, err
    }
    
    decimals := result[0].(uint8)
    return decimals, nil
}
```

### 5.2 合约部署

```go
// contract/deploy.go
package contract

import (
    "context"
    "crypto/ecdsa"
    "math/big"
    "strings"

    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/ethclient"
)

// 部署合约
func DeployContract(
    client *ethclient.Client,
    privateKey *ecdsa.PrivateKey,
    abiJSON string,
    bytecode string,
    gasLimit uint64,
    gasPrice *big.Int,
    params ...interface{},
) (common.Address, *types.Transaction, error) {
    
    fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
    
    // 解析ABI
    parsedABI, err := abi.JSON(strings.NewReader(abiJSON))
    if err != nil {
        return common.Address{}, nil, err
    }

    // 获取nonce
    nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        return common.Address{}, nil, err
    }

    // 创建交易选项
    auth := &bind.TransactOpts{
        From:     fromAddress,
        Nonce:    big.NewInt(int64(nonce)),
        GasLimit: gasLimit,
        GasPrice: gasPrice,
        Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
            if address != fromAddress {
                return nil, fmt.Errorf("not authorized to sign this account")
            }
            chainID, err := client.NetworkID(context.Background())
            if err != nil {
                return nil, err
            }
            return types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
        },
    }

    // 部署合约
    address, tx, _, err := bind.DeployContract(auth, parsedABI, common.FromHex(bytecode), client, params...)
    if err != nil {
        return common.Address{}, nil, err
    }

    return address, tx, nil
}
```

## 事件监听

### 7.1 事件过滤和监听

```go
// event/listener.go
package event

import (
    "context"
    "fmt"
    "log"
    "math/big"

    "github.com/ethereum/go-ethereum"
    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/ethclient"
)

type EventListener struct {
    client *ethclient.Client
}

func NewEventListener(client *ethclient.Client) *EventListener {
    return &EventListener{client: client}
}

// 监听Transfer事件
func (el *EventListener) ListenTransferEvents(contractAddress common.Address, fromBlock *big.Int) error {
    // Transfer事件的签名
    transferSig := []byte("Transfer(address,address,uint256)")
    transferSigHash := crypto.Keccak256Hash(transferSig)

    query := ethereum.FilterQuery{
        FromBlock: fromBlock,
        Addresses: []common.Address{contractAddress},
        Topics:    [][]common.Hash{{transferSigHash}},
    }

    logs := make(chan types.Log)
    sub, err := el.client.SubscribeFilterLogs(context.Background(), query, logs)
    if err != nil {
        return err
    }

    for {
        select {
        case err := <-sub.Err():
            return err
        case vLog := <-logs:
            el.handleTransferEvent(vLog)
        }
    }
}

// 处理Transfer事件
func (el *EventListener) handleTransferEvent(vLog types.Log) {
    // 解析事件数据
    from := common.HexToAddress(vLog.Topics[1].Hex())
    to := common.HexToAddress(vLog.Topics[2].Hex())
    value := new(big.Int).SetBytes(vLog.Data)

    log.Printf("Transfer事件: 从 %s 到 %s，金额 %s", from.Hex(), to.Hex(), value.String())
}

// 获取历史事件
func (el *EventListener) GetPastEvents(contractAddress common.Address, fromBlock, toBlock *big.Int) ([]types.Log, error) {
    transferSig := []byte("Transfer(address,address,uint256)")
    transferSigHash := crypto.Keccak256Hash(transferSig)

    query := ethereum.FilterQuery{
        FromBlock: fromBlock,
        ToBlock:   toBlock,
        Addresses: []common.Address{contractAddress},
        Topics:    [][]common.Hash{{transferSigHash}},
    }

    logs, err := el.client.FilterLogs(context.Background(), query)
    if err != nil {
        return nil, err
    }

    return logs, nil
}
```

## 高级功能

### 8.1 批量操作

```go
// batch/operations.go
package batch

import (
    "context"
    "fmt"
    "sync"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
)

type BatchOperations struct {
    client *ethclient.Client
}

func NewBatchOperations(client *ethclient.Client) *BatchOperations {
    return &BatchOperations{client: client}
}

// 批量查询余额
func (bo *BatchOperations) BatchGetBalances(addresses []common.Address) (map[common.Address]string, error) {
    results := make(map[common.Address]string)
    var mutex sync.Mutex
    var wg sync.WaitGroup

    // 限制并发数
    semaphore := make(chan struct{}, 10)

    for _, addr := range addresses {
        wg.Add(1)
        go func(address common.Address) {
            defer wg.Done()
            semaphore <- struct{}{} // 获取信号量
            defer func() { <-semaphore }() // 释放信号量

            balance, err := bo.client.BalanceAt(context.Background(), address, nil)
            if err != nil {
                fmt.Printf("获取地址 %s 余额失败: %v\n", address.Hex(), err)
                return
            }

            mutex.Lock()
            results[address] = balance.String()
            mutex.Unlock()
        }(addr)
    }

    wg.Wait()
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
    "math/big"

    "github.com/ethereum/go-ethereum/common"
    
    "your-project/client"
    "your-project/account"
    "your-project/transaction"
    "your-project/contract"
)

func main() {
    // 连接以太坊节点
    ethClient, err := client.NewEthereumClient("https://goerli.infura.io/v3/YOUR_PROJECT_ID")
    if err != nil {
        log.Fatal("连接以太坊节点失败:", err)
    }

    // 创建账户
    acc, err := account.CreateAccount()
    if err != nil {
        log.Fatal("创建账户失败:", err)
    }

    fmt.Printf("新账户地址: %s\n", acc.AddressHex())
    fmt.Printf("私钥: %s\n", acc.PrivateKeyHex())

    // 查询余额
    balance, err := ethClient.GetBalance(acc.Address)
    if err != nil {
        log.Fatal("查询余额失败:", err)
    }

    fmt.Printf("账户余额: %s Wei\n", balance.String())

    // 如果有余额，进行转账
    if balance.Cmp(big.NewInt(0)) > 0 {
        transferManager := transaction.NewTransferManager(ethClient.client, ethClient.chainID)
        
        toAddress := common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96c4b4Db93")
        amount := big.NewInt(1000000000000000) // 0.001 ETH
        gasLimit := uint64(21000)
        gasPrice := big.NewInt(20000000000) // 20 Gwei

        tx, err := transferManager.TransferETH(acc.PrivateKey, toAddress, amount, gasLimit, gasPrice)
        if err != nil {
            log.Fatal("转账失败:", err)
        }

        fmt.Printf("交易哈希: %s\n", tx.Hash().Hex())

        // 等待交易确认
        receipt, err := transferManager.WaitForConfirmation(tx.Hash())
        if err != nil {
            log.Fatal("等待交易确认失败:", err)
        }

        fmt.Printf("交易已确认，区块号: %d\n", receipt.BlockNumber.Uint64())
    }
}
```
