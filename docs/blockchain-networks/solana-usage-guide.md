# Solana Go SDK 详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [环境准备](#环境准备)
3. [RPC连接](#rpc连接)
4. [账户管理](#账户管理)
5. [交易处理](#交易处理)
6. [程序交互](#程序交互)
7. [SPL代币](#spl代币)
8. [最佳实践](#最佳实践)
9. [实际应用](#实际应用)

## 基础概念

### 1.1 Solana 简介

Solana是高性能区块链，支持每秒数万笔交易，采用独特的Proof of History共识机制。

```bash
# 安装Solana Go SDK
go get github.com/gagliardetto/solana-go
go get github.com/gagliardetto/solana-go/rpc
go get github.com/gagliardetto/solana-go/programs/token
```

### 1.2 核心概念

```go
// 主要包导入
import (
    "context"
    "crypto/ed25519"
    "fmt"
    
    "github.com/gagliardetto/solana-go"
    "github.com/gagliardetto/solana-go/rpc"
    "github.com/gagliardetto/solana-go/programs/token"
    "github.com/gagliardetto/solana-go/programs/system"
)

// Solana网络配置
const (
    MainnetRPC = "https://api.mainnet-beta.solana.com"
    DevnetRPC  = "https://api.devnet.solana.com"
    TestnetRPC = "https://api.testnet.solana.com"
)

// 基础配置
type SolanaConfig struct {
    RPCURL     string
    Commitment rpc.CommitmentType
}
```

## 环境准备

### 2.1 RPC客户端设置

```go
// client/solana_client.go
package client

import (
    "context"
    "time"
    
    "github.com/gagliardetto/solana-go/rpc"
)

type SolanaClient struct {
    rpcClient *rpc.Client
    config    *SolanaConfig
}

func NewSolanaClient(config *SolanaConfig) *SolanaClient {
    rpcClient := rpc.New(config.RPCURL)
    
    return &SolanaClient{
        rpcClient: rpcClient,
        config:    config,
    }
}

// 获取集群信息
func (c *SolanaClient) GetClusterNodes() (*rpc.GetClusterNodesResult, error) {
    return c.rpcClient.GetClusterNodes(context.Background())
}

// 获取版本信息
func (c *SolanaClient) GetVersion() (*rpc.GetVersionResult, error) {
    return c.rpcClient.GetVersion(context.Background())
}

// 获取最新区块高度
func (c *SolanaClient) GetSlot() (uint64, error) {
    return c.rpcClient.GetSlot(context.Background(), c.config.Commitment)
}
```

### 2.2 密钥管理

```go
// wallet/keypair.go
package wallet

import (
    "crypto/ed25519"
    "crypto/rand"
    "encoding/base58"
    "encoding/json"
    "io/ioutil"
    
    "github.com/gagliardetto/solana-go"
)

type Keypair struct {
    privateKey ed25519.PrivateKey
    publicKey  ed25519.PublicKey
}

// 生成新密钥对
func GenerateKeypair() (*Keypair, error) {
    publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
    if err != nil {
        return nil, err
    }
    
    return &Keypair{
        privateKey: privateKey,
        publicKey:  publicKey,
    }, nil
}

// 从私钥创建密钥对
func NewKeypairFromPrivateKey(privateKey ed25519.PrivateKey) *Keypair {
    publicKey := privateKey.Public().(ed25519.PublicKey)
    
    return &Keypair{
        privateKey: privateKey,
        publicKey:  publicKey,
    }
}

// 从文件加载密钥对
func LoadKeypairFromFile(filepath string) (*Keypair, error) {
    data, err := ioutil.ReadFile(filepath)
    if err != nil {
        return nil, err
    }
    
    var keyData []byte
    if err := json.Unmarshal(data, &keyData); err != nil {
        return nil, err
    }
    
    privateKey := ed25519.PrivateKey(keyData)
    return NewKeypairFromPrivateKey(privateKey), nil
}

// 获取公钥地址
func (k *Keypair) PublicKey() solana.PublicKey {
    return solana.PublicKeyFromBytes(k.publicKey)
}

// 获取私钥
func (k *Keypair) PrivateKey() ed25519.PrivateKey {
    return k.privateKey
}

// 签名数据
func (k *Keypair) Sign(data []byte) []byte {
    return ed25519.Sign(k.privateKey, data)
}
```

## RPC连接

### 3.1 账户查询

```go
// services/account_service.go
package services

import (
    "context"
    
    "github.com/gagliardetto/solana-go"
    "github.com/gagliardetto/solana-go/rpc"
)

type AccountService struct {
    client *SolanaClient
}

func NewAccountService(client *SolanaClient) *AccountService {
    return &AccountService{
        client: client,
    }
}

// 获取账户信息
func (s *AccountService) GetAccountInfo(pubkey solana.PublicKey) (*rpc.GetAccountInfoResult, error) {
    return s.client.rpcClient.GetAccountInfo(
        context.Background(),
        pubkey,
    )
}

// 获取账户余额
func (s *AccountService) GetBalance(pubkey solana.PublicKey) (uint64, error) {
    result, err := s.client.rpcClient.GetBalance(
        context.Background(),
        pubkey,
        s.client.config.Commitment,
    )
    if err != nil {
        return 0, err
    }
    
    return result.Value, nil
}

// 获取多个账户信息
func (s *AccountService) GetMultipleAccounts(pubkeys []solana.PublicKey) (*rpc.GetMultipleAccountsResult, error) {
    return s.client.rpcClient.GetMultipleAccounts(
        context.Background(),
        pubkeys...,
    )
}

// 获取程序账户
func (s *AccountService) GetProgramAccounts(programID solana.PublicKey) (*rpc.GetProgramAccountsResult, error) {
    return s.client.rpcClient.GetProgramAccounts(
        context.Background(),
        programID,
    )
}
```

### 3.2 区块和交易查询

```go
// services/block_service.go
package services

import (
    "context"
    
    "github.com/gagliardetto/solana-go"
    "github.com/gagliardetto/solana-go/rpc"
)

type BlockService struct {
    client *SolanaClient
}

func NewBlockService(client *SolanaClient) *BlockService {
    return &BlockService{
        client: client,
    }
}

// 获取区块信息
func (s *BlockService) GetBlock(slot uint64) (*rpc.GetBlockResult, error) {
    return s.client.rpcClient.GetBlock(
        context.Background(),
        slot,
    )
}

// 获取交易信息
func (s *BlockService) GetTransaction(signature solana.Signature) (*rpc.GetTransactionResult, error) {
    return s.client.rpcClient.GetTransaction(
        context.Background(),
        signature,
        &rpc.GetTransactionOpts{
            Encoding: solana.EncodingJSON,
        },
    )
}

// 获取交易历史
func (s *BlockService) GetSignaturesForAddress(address solana.PublicKey, limit int) (*rpc.GetSignaturesForAddressResult, error) {
    return s.client.rpcClient.GetSignaturesForAddress(
        context.Background(),
        address,
        &rpc.GetSignaturesForAddressOpts{
            Limit: &limit,
        },
    )
}

// 获取确认的区块
func (s *BlockService) GetConfirmedBlocks(startSlot, endSlot uint64) ([]uint64, error) {
    return s.client.rpcClient.GetConfirmedBlocks(
        context.Background(),
        startSlot,
        &endSlot,
        s.client.config.Commitment,
    )
}
```

## 账户管理

### 4.1 账户创建和管理

```go
// services/wallet_service.go
package services

import (
    "context"
    "fmt"
    
    "github.com/gagliardetto/solana-go"
    "github.com/gagliardetto/solana-go/programs/system"
    "github.com/gagliardetto/solana-go/rpc"
)

type WalletService struct {
    client  *SolanaClient
    keypair *Keypair
}

func NewWalletService(client *SolanaClient, keypair *Keypair) *WalletService {
    return &WalletService{
        client:  client,
        keypair: keypair,
    }
}

// 创建账户
func (s *WalletService) CreateAccount(newAccount *Keypair, space uint64, owner solana.PublicKey) (*solana.Transaction, error) {
    // 获取最低租金
    rentExemption, err := s.client.rpcClient.GetMinimumBalanceForRentExemption(
        context.Background(),
        space,
        s.client.config.Commitment,
    )
    if err != nil {
        return nil, err
    }
    
    // 创建账户指令
    instruction := system.NewCreateAccountInstruction(
        rentExemption,
        space,
        owner,
        s.keypair.PublicKey(),
        newAccount.PublicKey(),
    ).Build()
    
    // 获取最新区块哈希
    recentBlockhash, err := s.client.rpcClient.GetRecentBlockhash(
        context.Background(),
        s.client.config.Commitment,
    )
    if err != nil {
        return nil, err
    }
    
    // 构建交易
    tx, err := solana.NewTransaction(
        []solana.Instruction{instruction},
        recentBlockhash.Value.Blockhash,
        solana.TransactionPayer(s.keypair.PublicKey()),
    )
    if err != nil {
        return nil, err
    }
    
    // 签名交易
    _, err = tx.Sign(
        func(key solana.PublicKey) *solana.PrivateKey {
            if key.Equals(s.keypair.PublicKey()) {
                return &solana.PrivateKey{s.keypair.PrivateKey()}
            }
            if key.Equals(newAccount.PublicKey()) {
                return &solana.PrivateKey{newAccount.PrivateKey()}
            }
            return nil
        },
    )
    if err != nil {
        return nil, err
    }
    
    return tx, nil
}

// 转账SOL
func (s *WalletService) TransferSOL(to solana.PublicKey, amount uint64) (*solana.Signature, error) {
    // 创建转账指令
    instruction := system.NewTransferInstruction(
        amount,
        s.keypair.PublicKey(),
        to,
    ).Build()
    
    // 获取最新区块哈希
    recentBlockhash, err := s.client.rpcClient.GetRecentBlockhash(
        context.Background(),
        s.client.config.Commitment,
    )
    if err != nil {
        return nil, err
    }
    
    // 构建交易
    tx, err := solana.NewTransaction(
        []solana.Instruction{instruction},
        recentBlockhash.Value.Blockhash,
        solana.TransactionPayer(s.keypair.PublicKey()),
    )
    if err != nil {
        return nil, err
    }
    
    // 签名交易
    _, err = tx.Sign(
        func(key solana.PublicKey) *solana.PrivateKey {
            if key.Equals(s.keypair.PublicKey()) {
                return &solana.PrivateKey{s.keypair.PrivateKey()}
            }
            return nil
        },
    )
    if err != nil {
        return nil, err
    }
    
    // 发送交易
    signature, err := s.client.rpcClient.SendTransaction(
        context.Background(),
        tx,
    )
    if err != nil {
        return nil, err
    }
    
    return &signature, nil
}
```

## 交易处理

### 5.1 交易构建和发送

```go
// services/transaction_service.go
package services

import (
    "context"
    "time"
    
    "github.com/gagliardetto/solana-go"
    "github.com/gagliardetto/solana-go/rpc"
)

type TransactionService struct {
    client *SolanaClient
}

func NewTransactionService(client *SolanaClient) *TransactionService {
    return &TransactionService{
        client: client,
    }
}

// 发送并确认交易
func (s *TransactionService) SendAndConfirmTransaction(
    tx *solana.Transaction,
    signers []solana.PrivateKey,
) (*solana.Signature, error) {
    // 签名交易
    _, err := tx.Sign(
        func(key solana.PublicKey) *solana.PrivateKey {
            for _, signer := range signers {
                if key.Equals(signer.PublicKey()) {
                    return &signer
                }
            }
            return nil
        },
    )
    if err != nil {
        return nil, err
    }
    
    // 发送交易
    signature, err := s.client.rpcClient.SendTransaction(
        context.Background(),
        tx,
    )
    if err != nil {
        return nil, err
    }
    
    // 等待确认
    err = s.waitForConfirmation(signature, 30*time.Second)
    if err != nil {
        return nil, err
    }
    
    return &signature, nil
}

// 等待交易确认
func (s *TransactionService) waitForConfirmation(signature solana.Signature, timeout time.Duration) error {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()
    
    ticker := time.NewTicker(2 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return fmt.Errorf("交易确认超时")
        case <-ticker.C:
            result, err := s.client.rpcClient.GetSignatureStatuses(
                context.Background(),
                true,
                signature,
            )
            if err != nil {
                continue
            }
            
            if len(result.Value) > 0 && result.Value[0] != nil {
                status := result.Value[0]
                if status.Err != nil {
                    return fmt.Errorf("交易失败: %v", status.Err)
                }
                if status.ConfirmationStatus != nil {
                    return nil // 交易已确认
                }
            }
        }
    }
}

// 模拟交易
func (s *TransactionService) SimulateTransaction(tx *solana.Transaction) (*rpc.SimulateTransactionResult, error) {
    return s.client.rpcClient.SimulateTransaction(
        context.Background(),
        tx,
    )
}
```

## 程序交互

### 6.1 自定义程序调用

```go
// programs/custom_program.go
package programs

import (
    "github.com/gagliardetto/solana-go"
)

// 自定义程序ID
var CustomProgramID = solana.MustPublicKeyFromBase58("YourProgramIDHere")

// 程序指令数据结构
type CustomInstruction struct {
    InstructionType uint8
    Data           []byte
}

// 创建自定义指令
func NewCustomInstruction(
    instructionType uint8,
    data []byte,
    accounts []*solana.AccountMeta,
) *solana.GenericInstruction {
    instructionData := append([]byte{instructionType}, data...)
    
    return solana.NewInstruction(
        CustomProgramID,
        accounts,
        instructionData,
    )
}

// 程序账户元数据
func NewAccountMeta(pubkey solana.PublicKey, isSigner, isWritable bool) *solana.AccountMeta {
    return &solana.AccountMeta{
        PublicKey:  pubkey,
        IsSigner:   isSigner,
        IsWritable: isWritable,
    }
}
```

### 6.2 程序部署

```go
// services/program_service.go
package services

import (
    "context"
    "io/ioutil"
    
    "github.com/gagliardetto/solana-go"
    "github.com/gagliardetto/solana-go/programs/loader"
)

type ProgramService struct {
    client *SolanaClient
}

func NewProgramService(client *SolanaClient) *ProgramService {
    return &ProgramService{
        client: client,
    }
}

// 部署程序
func (s *ProgramService) DeployProgram(
    programPath string,
    payer *Keypair,
) (*solana.PublicKey, error) {
    // 读取程序字节码
    programData, err := ioutil.ReadFile(programPath)
    if err != nil {
        return nil, err
    }
    
    // 生成程序账户
    programKeypair, err := GenerateKeypair()
    if err != nil {
        return nil, err
    }
    
    // 创建程序账户
    programSize := uint64(len(programData))
    rentExemption, err := s.client.rpcClient.GetMinimumBalanceForRentExemption(
        context.Background(),
        programSize,
        s.client.config.Commitment,
    )
    if err != nil {
        return nil, err
    }
    
    // 这里需要实现完整的程序部署逻辑
    // 包括创建账户、写入数据、标记为可执行等步骤
    
    programPubkey := programKeypair.PublicKey()
    return &programPubkey, nil
}
```

## SPL代币

### 7.1 代币操作

```go
// services/token_service.go
package services

import (
    "context"
    
    "github.com/gagliardetto/solana-go"
    "github.com/gagliardetto/solana-go/programs/token"
)

type TokenService struct {
    client *SolanaClient
}

func NewTokenService(client *SolanaClient) *TokenService {
    return &TokenService{
        client: client,
    }
}

// 创建代币铸造账户
func (s *TokenService) CreateMint(
    payer *Keypair,
    mintAuthority solana.PublicKey,
    freezeAuthority *solana.PublicKey,
    decimals uint8,
) (*solana.PublicKey, error) {
    // 生成铸造账户密钥对
    mintKeypair, err := GenerateKeypair()
    if err != nil {
        return nil, err
    }
    
    // 获取租金豁免金额
    rentExemption, err := s.client.rpcClient.GetMinimumBalanceForRentExemption(
        context.Background(),
        token.MINT_SIZE,
        s.client.config.Commitment,
    )
    if err != nil {
        return nil, err
    }
    
    // 创建账户指令
    createAccountInstruction := system.NewCreateAccountInstruction(
        rentExemption,
        token.MINT_SIZE,
        token.ProgramID,
        payer.PublicKey(),
        mintKeypair.PublicKey(),
    ).Build()
    
    // 初始化铸造指令
    initializeMintInstruction := token.NewInitializeMintInstruction(
        decimals,
        mintAuthority,
        freezeAuthority,
        mintKeypair.PublicKey(),
    ).Build()
    
    // 构建和发送交易
    instructions := []solana.Instruction{
        createAccountInstruction,
        initializeMintInstruction,
    }
    
    // 这里需要构建完整的交易并发送
    
    mintPubkey := mintKeypair.PublicKey()
    return &mintPubkey, nil
}

// 创建代币账户
func (s *TokenService) CreateTokenAccount(
    payer *Keypair,
    owner solana.PublicKey,
    mint solana.PublicKey,
) (*solana.PublicKey, error) {
    // 生成代币账户密钥对
    tokenAccountKeypair, err := GenerateKeypair()
    if err != nil {
        return nil, err
    }
    
    // 获取租金豁免金额
    rentExemption, err := s.client.rpcClient.GetMinimumBalanceForRentExemption(
        context.Background(),
        token.ACCOUNT_SIZE,
        s.client.config.Commitment,
    )
    if err != nil {
        return nil, err
    }
    
    // 创建账户指令
    createAccountInstruction := system.NewCreateAccountInstruction(
        rentExemption,
        token.ACCOUNT_SIZE,
        token.ProgramID,
        payer.PublicKey(),
        tokenAccountKeypair.PublicKey(),
    ).Build()
    
    // 初始化代币账户指令
    initializeAccountInstruction := token.NewInitializeAccountInstruction(
        tokenAccountKeypair.PublicKey(),
        mint,
        owner,
    ).Build()
    
    // 构建和发送交易
    instructions := []solana.Instruction{
        createAccountInstruction,
        initializeAccountInstruction,
    }
    
    // 这里需要构建完整的交易并发送
    
    tokenAccountPubkey := tokenAccountKeypair.PublicKey()
    return &tokenAccountPubkey, nil
}

// 铸造代币
func (s *TokenService) MintTokens(
    mint solana.PublicKey,
    destination solana.PublicKey,
    authority *Keypair,
    amount uint64,
) error {
    // 创建铸造指令
    mintInstruction := token.NewMintToInstruction(
        amount,
        mint,
        destination,
        authority.PublicKey(),
        []solana.PublicKey{},
    ).Build()
    
    // 构建和发送交易
    // 这里需要完整的交易构建逻辑
    
    return nil
}

// 转移代币
func (s *TokenService) TransferTokens(
    source solana.PublicKey,
    destination solana.PublicKey,
    owner *Keypair,
    amount uint64,
) error {
    // 创建转移指令
    transferInstruction := token.NewTransferInstruction(
        amount,
        source,
        destination,
        owner.PublicKey(),
        []solana.PublicKey{},
    ).Build()
    
    // 构建和发送交易
    // 这里需要完整的交易构建逻辑
    
    return nil
}
```

## 最佳实践

### 8.1 错误处理

```go
// utils/error_handler.go
package utils

import (
    "fmt"
    "strings"
)

// Solana错误类型
type SolanaError struct {
    Code    int
    Message string
    Data    interface{}
}

func (e *SolanaError) Error() string {
    return fmt.Sprintf("Solana错误 %d: %s", e.Code, e.Message)
}

// 解析RPC错误
func ParseRPCError(err error) *SolanaError {
    errStr := err.Error()
    
    // 常见错误模式匹配
    if strings.Contains(errStr, "insufficient funds") {
        return &SolanaError{
            Code:    1001,
            Message: "余额不足",
            Data:    err,
        }
    }
    
    if strings.Contains(errStr, "blockhash not found") {
        return &SolanaError{
            Code:    1002,
            Message: "区块哈希过期",
            Data:    err,
        }
    }
    
    return &SolanaError{
        Code:    9999,
        Message: "未知错误",
        Data:    err,
    }
}
```

### 8.2 性能优化

```go
// utils/performance.go
package utils

import (
    "context"
    "sync"
    "time"
    
    "github.com/gagliardetto/solana-go"
    "github.com/gagliardetto/solana-go/rpc"
)

// 批量账户查询
func BatchGetAccountInfo(
    client *rpc.Client,
    pubkeys []solana.PublicKey,
    batchSize int,
) ([]*rpc.Account, error) {
    var results []*rpc.Account
    var mu sync.Mutex
    var wg sync.WaitGroup
    
    for i := 0; i < len(pubkeys); i += batchSize {
        end := i + batchSize
        if end > len(pubkeys) {
            end = len(pubkeys)
        }
        
        wg.Add(1)
        go func(batch []solana.PublicKey) {
            defer wg.Done()
            
            result, err := client.GetMultipleAccounts(
                context.Background(),
                batch...,
            )
            if err != nil {
                return
            }
            
            mu.Lock()
            results = append(results, result.Value...)
            mu.Unlock()
        }(pubkeys[i:end])
    }
    
    wg.Wait()
    return results, nil
}

// 连接池管理
type ConnectionPool struct {
    clients []*rpc.Client
    current int
    mu      sync.Mutex
}

func NewConnectionPool(endpoints []string) *ConnectionPool {
    var clients []*rpc.Client
    for _, endpoint := range endpoints {
        clients = append(clients, rpc.New(endpoint))
    }
    
    return &ConnectionPool{
        clients: clients,
        current: 0,
    }
}

func (p *ConnectionPool) GetClient() *rpc.Client {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    client := p.clients[p.current]
    p.current = (p.current + 1) % len(p.clients)
    
    return client
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
    
    "your-project/client"
    "your-project/services"
    "your-project/wallet"
)

func main() {
    // 创建配置
    config := &SolanaConfig{
        RPCURL:     DevnetRPC,
        Commitment: rpc.CommitmentConfirmed,
    }
    
    // 创建客户端
    solanaClient := client.NewSolanaClient(config)
    
    // 生成密钥对
    keypair, err := wallet.GenerateKeypair()
    if err != nil {
        log.Fatal("生成密钥对失败:", err)
    }
    
    fmt.Printf("公钥地址: %s\n", keypair.PublicKey().String())
    
    // 创建服务
    accountService := services.NewAccountService(solanaClient)
    walletService := services.NewWalletService(solanaClient, keypair)
    
    // 查询余额
    balance, err := accountService.GetBalance(keypair.PublicKey())
    if err != nil {
        log.Fatal("查询余额失败:", err)
    }
    
    fmt.Printf("账户余额: %d lamports\n", balance)
    
    // 如果有余额，进行转账测试
    if balance > 1000000 { // 0.001 SOL
        targetKeypair, _ := wallet.GenerateKeypair()
        
        signature, err := walletService.TransferSOL(
            targetKeypair.PublicKey(),
            500000, // 0.0005 SOL
        )
        if err != nil {
            log.Fatal("转账失败:", err)
        }
        
        fmt.Printf("转账成功，交易签名: %s\n", signature.String())
    }
}
```

这个Solana使用指南提供了完整的Go语言集成方案，涵盖了从基础连接到高级功能的所有核心操作。
