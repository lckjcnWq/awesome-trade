# Polkadot 生态 详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [环境准备](#环境准备)
3. [Substrate连接](#substrate连接)
4. [平行链交互](#平行链交互)
5. [跨链通信](#跨链通信)
6. [治理参与](#治理参与)
7. [质押操作](#质押操作)
8. [最佳实践](#最佳实践)
9. [实际应用](#实际应用)

## 基础概念

### 1.1 Polkadot 生态简介

Polkadot是多链区块链平台，通过中继链连接多个平行链，实现跨链互操作性和共享安全。

```bash
# 安装Polkadot Go SDK
go get github.com/centrifuge/go-substrate-rpc-client/v4
go get github.com/vedhavyas/go-subkey
go get github.com/ChainSafe/gossamer/lib/crypto/sr25519
```

### 1.2 核心架构

```go
// 主要包导入
import (
    "context"
    "fmt"
    "math/big"
    
    gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
    "github.com/centrifuge/go-substrate-rpc-client/v4/config"
    "github.com/centrifuge/go-substrate-rpc-client/v4/signature"
    "github.com/centrifuge/go-substrate-rpc-client/v4/types"
    "github.com/vedhavyas/go-subkey/v2"
)

// 网络配置
const (
    PolkadotMainnet = "wss://rpc.polkadot.io"
    KusamaMainnet   = "wss://kusama-rpc.polkadot.io"
    WestendTestnet  = "wss://westend-rpc.polkadot.io"
)

// Polkadot配置
type PolkadotConfig struct {
    RPCURL    string
    Network   string
    ChainID   uint8
}

// 账户信息
type AccountInfo struct {
    Address    string
    PublicKey  []byte
    PrivateKey []byte
    SS58Format uint8
}
```

## 环境准备

### 2.1 RPC客户端设置

```go
// client/polkadot_client.go
package client

import (
    "context"
    "fmt"
    
    gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
    "github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

type PolkadotClient struct {
    api    *gsrpc.SubstrateAPI
    config *PolkadotConfig
}

func NewPolkadotClient(config *PolkadotConfig) (*PolkadotClient, error) {
    api, err := gsrpc.NewSubstrateAPI(config.RPCURL)
    if err != nil {
        return nil, fmt.Errorf("连接Polkadot节点失败: %w", err)
    }
    
    return &PolkadotClient{
        api:    api,
        config: config,
    }, nil
}

// 获取链信息
func (c *PolkadotClient) GetChainInfo() (*types.ChainProperties, error) {
    return c.api.RPC.System.Properties()
}

// 获取最新区块
func (c *PolkadotClient) GetLatestBlock() (*types.SignedBlock, error) {
    hash, err := c.api.RPC.Chain.GetBlockHashLatest()
    if err != nil {
        return nil, err
    }
    
    return c.api.RPC.Chain.GetBlock(hash)
}

// 获取运行时版本
func (c *PolkadotClient) GetRuntimeVersion() (*types.RuntimeVersion, error) {
    return c.api.RPC.State.GetRuntimeVersionLatest()
}

// 获取元数据
func (c *PolkadotClient) GetMetadata() (*types.Metadata, error) {
    return c.api.RPC.State.GetMetadataLatest()
}

// 关闭连接
func (c *PolkadotClient) Close() {
    // Substrate API会自动处理连接关闭
}
```

### 2.2 密钥管理

```go
// wallet/polkadot_wallet.go
package wallet

import (
    "crypto/rand"
    "fmt"
    
    "github.com/ChainSafe/gossamer/lib/crypto/sr25519"
    "github.com/centrifuge/go-substrate-rpc-client/v4/signature"
    "github.com/centrifuge/go-substrate-rpc-client/v4/types"
    "github.com/vedhavyas/go-subkey/v2"
)

type PolkadotWallet struct {
    keyringPair signature.KeyringPair
    ss58Format  uint8
}

// 创建新钱包
func NewPolkadotWallet(ss58Format uint8) (*PolkadotWallet, error) {
    // 生成随机种子
    seed := make([]byte, 32)
    _, err := rand.Read(seed)
    if err != nil {
        return nil, err
    }
    
    // 从种子创建密钥对
    keyringPair, err := signature.KeyringPairFromSecret(seed, 42) // 42 = sr25519
    if err != nil {
        return nil, err
    }
    
    return &PolkadotWallet{
        keyringPair: keyringPair,
        ss58Format:  ss58Format,
    }, nil
}

// 从助记词创建钱包
func NewPolkadotWalletFromMnemonic(mnemonic string, ss58Format uint8) (*PolkadotWallet, error) {
    // 从助记词生成种子
    seed, err := subkey.DeriveKeyPair(subkey.Sr25519Type, mnemonic, "", "")
    if err != nil {
        return nil, err
    }
    
    // 创建密钥对
    keyringPair, err := signature.KeyringPairFromSecret(seed.Seed(), 42)
    if err != nil {
        return nil, err
    }
    
    return &PolkadotWallet{
        keyringPair: keyringPair,
        ss58Format:  ss58Format,
    }, nil
}

// 从私钥创建钱包
func NewPolkadotWalletFromPrivateKey(privateKey []byte, ss58Format uint8) (*PolkadotWallet, error) {
    keyringPair, err := signature.KeyringPairFromSecret(privateKey, 42)
    if err != nil {
        return nil, err
    }
    
    return &PolkadotWallet{
        keyringPair: keyringPair,
        ss58Format:  ss58Format,
    }, nil
}

// 获取地址
func (w *PolkadotWallet) GetAddress() string {
    return w.keyringPair.Address
}

// 获取公钥
func (w *PolkadotWallet) GetPublicKey() []byte {
    return w.keyringPair.PublicKey
}

// 获取SS58地址
func (w *PolkadotWallet) GetSS58Address() (string, error) {
    return subkey.SS58Address(w.keyringPair.PublicKey, w.ss58Format)
}

// 签名数据
func (w *PolkadotWallet) Sign(data []byte) ([]byte, error) {
    return w.keyringPair.Sign(data)
}

// 验证签名
func (w *PolkadotWallet) Verify(data, signature []byte) bool {
    return w.keyringPair.Verify(data, signature)
}
```

## Substrate连接

### 3.1 账户查询

```go
// services/account_service.go
package services

import (
    "fmt"
    
    "github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

type AccountService struct {
    client *PolkadotClient
}

func NewAccountService(client *PolkadotClient) *AccountService {
    return &AccountService{
        client: client,
    }
}

// 获取账户信息
func (s *AccountService) GetAccountInfo(address string) (*types.AccountInfo, error) {
    // 将地址转换为AccountID
    accountID, err := types.NewAccountID([]byte(address))
    if err != nil {
        return nil, err
    }
    
    // 获取存储键
    key, err := types.CreateStorageKey(s.client.api.Metadata, "System", "Account", accountID[:])
    if err != nil {
        return nil, err
    }
    
    // 查询账户信息
    var accountInfo types.AccountInfo
    ok, err := s.client.api.RPC.State.GetStorageLatest(key, &accountInfo)
    if err != nil {
        return nil, err
    }
    
    if !ok {
        return nil, fmt.Errorf("账户不存在")
    }
    
    return &accountInfo, nil
}

// 获取账户余额
func (s *AccountService) GetBalance(address string) (*types.AccountInfo, error) {
    return s.GetAccountInfo(address)
}

// 获取账户nonce
func (s *AccountService) GetNonce(address string) (types.UCompact, error) {
    accountInfo, err := s.GetAccountInfo(address)
    if err != nil {
        return 0, err
    }
    
    return accountInfo.Nonce, nil
}

// 查询存储
func (s *AccountService) QueryStorage(module, method string, args ...interface{}) (interface{}, error) {
    key, err := types.CreateStorageKey(s.client.api.Metadata, module, method, args...)
    if err != nil {
        return nil, err
    }
    
    var result interface{}
    ok, err := s.client.api.RPC.State.GetStorageLatest(key, &result)
    if err != nil {
        return nil, err
    }
    
    if !ok {
        return nil, fmt.Errorf("存储项不存在")
    }
    
    return result, nil
}
```

### 3.2 交易处理

```go
// services/transaction_service.go
package services

import (
    "fmt"
    
    "github.com/centrifuge/go-substrate-rpc-client/v4/signature"
    "github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

type TransactionService struct {
    client *PolkadotClient
}

func NewTransactionService(client *PolkadotClient) *TransactionService {
    return &TransactionService{
        client: client,
    }
}

// 转账
func (s *TransactionService) Transfer(
    from *PolkadotWallet,
    to string,
    amount types.UCompact,
) (types.Hash, error) {
    // 获取元数据
    meta, err := s.client.api.RPC.State.GetMetadataLatest()
    if err != nil {
        return types.Hash{}, err
    }
    
    // 创建转账调用
    call, err := types.NewCall(meta, "Balances.transfer", to, amount)
    if err != nil {
        return types.Hash{}, err
    }
    
    // 创建外部交易
    ext := types.NewExtrinsic(call)
    
    // 获取创世哈希
    genesisHash, err := s.client.api.RPC.Chain.GetBlockHash(0)
    if err != nil {
        return types.Hash{}, err
    }
    
    // 获取运行时版本
    rv, err := s.client.api.RPC.State.GetRuntimeVersionLatest()
    if err != nil {
        return types.Hash{}, err
    }
    
    // 获取账户信息以获取nonce
    accountInfo, err := s.getAccountInfo(from.GetAddress())
    if err != nil {
        return types.Hash{}, err
    }
    
    // 设置签名选项
    o := types.SignatureOptions{
        BlockHash:          genesisHash,
        Era:                types.ExtrinsicEra{IsMortalEra: false},
        GenesisHash:        genesisHash,
        Nonce:              types.NewUCompactFromUInt(uint64(accountInfo.Nonce)),
        SpecVersion:        rv.SpecVersion,
        Tip:                types.NewUCompactFromUInt(0),
        TransactionVersion: rv.TransactionVersion,
    }
    
    // 签名交易
    err = ext.Sign(from.keyringPair, o)
    if err != nil {
        return types.Hash{}, err
    }
    
    // 提交交易
    hash, err := s.client.api.RPC.Author.SubmitExtrinsic(ext)
    if err != nil {
        return types.Hash{}, err
    }
    
    return hash, nil
}

// 批量转账
func (s *TransactionService) BatchTransfer(
    from *PolkadotWallet,
    transfers []TransferInfo,
) (types.Hash, error) {
    // 获取元数据
    meta, err := s.client.api.RPC.State.GetMetadataLatest()
    if err != nil {
        return types.Hash{}, err
    }
    
    // 创建批量调用
    var calls []types.Call
    for _, transfer := range transfers {
        call, err := types.NewCall(meta, "Balances.transfer", transfer.To, transfer.Amount)
        if err != nil {
            return types.Hash{}, err
        }
        calls = append(calls, call)
    }
    
    // 创建批量调用
    batchCall, err := types.NewCall(meta, "Utility.batch", calls)
    if err != nil {
        return types.Hash{}, err
    }
    
    // 创建外部交易
    ext := types.NewExtrinsic(batchCall)
    
    // 签名和提交逻辑与单笔转账相同
    // ... (省略重复代码)
    
    return types.Hash{}, nil
}

// 转账信息结构
type TransferInfo struct {
    To     string
    Amount types.UCompact
}

// 获取账户信息的辅助方法
func (s *TransactionService) getAccountInfo(address string) (*types.AccountInfo, error) {
    accountID, err := types.NewAccountID([]byte(address))
    if err != nil {
        return nil, err
    }
    
    key, err := types.CreateStorageKey(s.client.api.Metadata, "System", "Account", accountID[:])
    if err != nil {
        return nil, err
    }
    
    var accountInfo types.AccountInfo
    ok, err := s.client.api.RPC.State.GetStorageLatest(key, &accountInfo)
    if err != nil {
        return nil, err
    }
    
    if !ok {
        return nil, fmt.Errorf("账户不存在")
    }
    
    return &accountInfo, nil
}
```

## 平行链交互

### 4.1 平行链管理

```go
// services/parachain_service.go
package services

import (
    "fmt"
    
    "github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

type ParachainService struct {
    client *PolkadotClient
}

func NewParachainService(client *PolkadotClient) *ParachainService {
    return &ParachainService{
        client: client,
    }
}

// 平行链信息
type ParachainInfo struct {
    ID          types.U32
    Head        types.Bytes
    CodeHash    types.Hash
    Validators  []types.AccountID
}

// 获取平行链列表
func (s *ParachainService) GetParachains() ([]types.U32, error) {
    key, err := types.CreateStorageKey(s.client.api.Metadata, "Paras", "Parachains")
    if err != nil {
        return nil, err
    }
    
    var parachains []types.U32
    ok, err := s.client.api.RPC.State.GetStorageLatest(key, &parachains)
    if err != nil {
        return nil, err
    }
    
    if !ok {
        return nil, fmt.Errorf("无法获取平行链列表")
    }
    
    return parachains, nil
}

// 获取平行链信息
func (s *ParachainService) GetParachainInfo(paraID types.U32) (*ParachainInfo, error) {
    // 获取平行链头部
    headKey, err := types.CreateStorageKey(s.client.api.Metadata, "Paras", "Heads", paraID)
    if err != nil {
        return nil, err
    }
    
    var head types.Bytes
    ok, err := s.client.api.RPC.State.GetStorageLatest(headKey, &head)
    if err != nil {
        return nil, err
    }
    
    if !ok {
        return nil, fmt.Errorf("平行链 %d 不存在", paraID)
    }
    
    // 获取代码哈希
    codeKey, err := types.CreateStorageKey(s.client.api.Metadata, "Paras", "CurrentCodeHash", paraID)
    if err != nil {
        return nil, err
    }
    
    var codeHash types.Hash
    s.client.api.RPC.State.GetStorageLatest(codeKey, &codeHash)
    
    return &ParachainInfo{
        ID:       paraID,
        Head:     head,
        CodeHash: codeHash,
    }, nil
}

// 获取平行链验证者
func (s *ParachainService) GetParachainValidators(paraID types.U32) ([]types.AccountID, error) {
    key, err := types.CreateStorageKey(s.client.api.Metadata, "ParaScheduler", "ValidatorGroups")
    if err != nil {
        return nil, err
    }
    
    var validatorGroups [][]types.AccountID
    ok, err := s.client.api.RPC.State.GetStorageLatest(key, &validatorGroups)
    if err != nil {
        return nil, err
    }
    
    if !ok || len(validatorGroups) == 0 {
        return nil, fmt.Errorf("无法获取验证者组")
    }
    
    // 简化处理，返回第一个验证者组
    return validatorGroups[0], nil
}
```

## 跨链通信

### 5.1 XCM消息处理

```go
// services/xcm_service.go
package services

import (
    "fmt"
    
    "github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

type XCMService struct {
    client *PolkadotClient
}

func NewXCMService(client *PolkadotClient) *XCMService {
    return &XCMService{
        client: client,
    }
}

// XCM消息结构
type XCMMessage struct {
    Version     types.U8
    Instructions []XCMInstruction
}

type XCMInstruction struct {
    Type string
    Data interface{}
}

// 跨链资产转移
func (s *XCMService) CrossChainTransfer(
    from *PolkadotWallet,
    destChain types.U32,
    destAccount string,
    asset AssetInfo,
    amount types.U128,
) (types.Hash, error) {
    // 获取元数据
    meta, err := s.client.api.RPC.State.GetMetadataLatest()
    if err != nil {
        return types.Hash{}, err
    }
    
    // 构建XCM消息
    xcmMessage := s.buildTransferMessage(destChain, destAccount, asset, amount)
    
    // 创建XCM调用
    call, err := types.NewCall(meta, "XcmPallet.send", destChain, xcmMessage)
    if err != nil {
        return types.Hash{}, err
    }
    
    // 创建外部交易
    ext := types.NewExtrinsic(call)
    
    // 签名和提交交易
    hash, err := s.signAndSubmit(from, ext)
    if err != nil {
        return types.Hash{}, err
    }
    
    return hash, nil
}

// 构建转移消息
func (s *XCMService) buildTransferMessage(
    destChain types.U32,
    destAccount string,
    asset AssetInfo,
    amount types.U128,
) XCMMessage {
    // 简化的XCM消息构建
    instructions := []XCMInstruction{
        {
            Type: "WithdrawAsset",
            Data: map[string]interface{}{
                "assets": []interface{}{
                    map[string]interface{}{
                        "id":  asset.ID,
                        "fun": map[string]interface{}{"Fungible": amount},
                    },
                },
            },
        },
        {
            Type: "BuyExecution",
            Data: map[string]interface{}{
                "fees": map[string]interface{}{
                    "id":  asset.ID,
                    "fun": map[string]interface{}{"Fungible": types.U128(1000000)}, // 手续费
                },
                "weightLimit": "Unlimited",
            },
        },
        {
            Type: "DepositAsset",
            Data: map[string]interface{}{
                "assets": "All",
                "maxAssets": types.U32(1),
                "beneficiary": map[string]interface{}{
                    "parents": types.U8(0),
                    "interior": map[string]interface{}{
                        "X1": map[string]interface{}{
                            "AccountId32": map[string]interface{}{
                                "network": "Any",
                                "id":      destAccount,
                            },
                        },
                    },
                },
            },
        },
    }
    
    return XCMMessage{
        Version:      types.U8(2), // XCM v2
        Instructions: instructions,
    }
}

// 资产信息
type AssetInfo struct {
    ID   interface{}
    Name string
}

// 签名和提交交易的辅助方法
func (s *XCMService) signAndSubmit(wallet *PolkadotWallet, ext types.Extrinsic) (types.Hash, error) {
    // 获取创世哈希
    genesisHash, err := s.client.api.RPC.Chain.GetBlockHash(0)
    if err != nil {
        return types.Hash{}, err
    }
    
    // 获取运行时版本
    rv, err := s.client.api.RPC.State.GetRuntimeVersionLatest()
    if err != nil {
        return types.Hash{}, err
    }
    
    // 获取账户信息
    accountInfo, err := s.getAccountInfo(wallet.GetAddress())
    if err != nil {
        return types.Hash{}, err
    }
    
    // 设置签名选项
    o := types.SignatureOptions{
        BlockHash:          genesisHash,
        Era:                types.ExtrinsicEra{IsMortalEra: false},
        GenesisHash:        genesisHash,
        Nonce:              types.NewUCompactFromUInt(uint64(accountInfo.Nonce)),
        SpecVersion:        rv.SpecVersion,
        Tip:                types.NewUCompactFromUInt(0),
        TransactionVersion: rv.TransactionVersion,
    }
    
    // 签名交易
    err = ext.Sign(wallet.keyringPair, o)
    if err != nil {
        return types.Hash{}, err
    }
    
    // 提交交易
    return s.client.api.RPC.Author.SubmitExtrinsic(ext)
}

// 获取账户信息的辅助方法
func (s *XCMService) getAccountInfo(address string) (*types.AccountInfo, error) {
    accountID, err := types.NewAccountID([]byte(address))
    if err != nil {
        return nil, err
    }
    
    key, err := types.CreateStorageKey(s.client.api.Metadata, "System", "Account", accountID[:])
    if err != nil {
        return nil, err
    }
    
    var accountInfo types.AccountInfo
    ok, err := s.client.api.RPC.State.GetStorageLatest(key, &accountInfo)
    if err != nil {
        return nil, err
    }
    
    if !ok {
        return nil, fmt.Errorf("账户不存在")
    }
    
    return &accountInfo, nil
}
```

## 治理参与

### 6.1 治理操作

```go
// services/governance_service.go
package services

import (
    "github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

type GovernanceService struct {
    client *PolkadotClient
}

func NewGovernanceService(client *PolkadotClient) *GovernanceService {
    return &GovernanceService{
        client: client,
    }
}

// 提案信息
type ProposalInfo struct {
    Index       types.U32
    Hash        types.Hash
    Proposer    types.AccountID
    Deposit     types.U128
    Description string
}

// 获取公投列表
func (s *GovernanceService) GetReferendums() ([]types.U32, error) {
    key, err := types.CreateStorageKey(s.client.api.Metadata, "Democracy", "ReferendumInfoOf")
    if err != nil {
        return nil, err
    }
    
    // 获取所有公投
    keys, err := s.client.api.RPC.State.GetKeysLatest(key)
    if err != nil {
        return nil, err
    }
    
    var referendumIDs []types.U32
    for _, k := range keys {
        // 从存储键中提取公投ID
        // 这里需要解析存储键的具体实现
        // 简化处理
        referendumIDs = append(referendumIDs, types.U32(len(referendumIDs)))
    }
    
    return referendumIDs, nil
}

// 投票
func (s *GovernanceService) Vote(
    wallet *PolkadotWallet,
    referendumIndex types.U32,
    vote VoteInfo,
) (types.Hash, error) {
    // 获取元数据
    meta, err := s.client.api.RPC.State.GetMetadataLatest()
    if err != nil {
        return types.Hash{}, err
    }
    
    // 创建投票调用
    call, err := types.NewCall(meta, "Democracy.vote", referendumIndex, vote)
    if err != nil {
        return types.Hash{}, err
    }
    
    // 创建外部交易
    ext := types.NewExtrinsic(call)
    
    // 签名和提交
    return s.signAndSubmit(wallet, ext)
}

// 投票信息
type VoteInfo struct {
    Aye      bool
    Conviction types.U8
    Balance  types.U128
}

// 提交提案
func (s *GovernanceService) SubmitProposal(
    wallet *PolkadotWallet,
    proposal types.Call,
    value types.U128,
) (types.Hash, error) {
    // 获取元数据
    meta, err := s.client.api.RPC.State.GetMetadataLatest()
    if err != nil {
        return types.Hash{}, err
    }
    
    // 创建提案调用
    call, err := types.NewCall(meta, "Democracy.propose", proposal, value)
    if err != nil {
        return types.Hash{}, err
    }
    
    // 创建外部交易
    ext := types.NewExtrinsic(call)
    
    // 签名和提交
    return s.signAndSubmit(wallet, ext)
}

// 签名和提交的辅助方法
func (s *GovernanceService) signAndSubmit(wallet *PolkadotWallet, ext types.Extrinsic) (types.Hash, error) {
    // 实现与XCMService中相同的签名逻辑
    // ... (省略重复代码)
    return types.Hash{}, nil
}
```

## 质押操作

### 7.1 提名质押

```go
// services/staking_service.go
package services

import (
    "github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

type StakingService struct {
    client *PolkadotClient
}

func NewStakingService(client *PolkadotClient) *StakingService {
    return &StakingService{
        client: client,
    }
}

// 质押信息
type StakingInfo struct {
    Stash      types.AccountID
    Controller types.AccountID
    Total      types.U128
    Active     types.U128
    Unlocking  []UnlockingInfo
}

type UnlockingInfo struct {
    Value types.U128
    Era   types.U32
}

// 绑定质押
func (s *StakingService) Bond(
    wallet *PolkadotWallet,
    controller string,
    value types.U128,
    payee string,
) (types.Hash, error) {
    // 获取元数据
    meta, err := s.client.api.RPC.State.GetMetadataLatest()
    if err != nil {
        return types.Hash{}, err
    }
    
    // 创建绑定调用
    call, err := types.NewCall(meta, "Staking.bond", controller, value, payee)
    if err != nil {
        return types.Hash{}, err
    }
    
    // 创建外部交易
    ext := types.NewExtrinsic(call)
    
    // 签名和提交
    return s.signAndSubmit(wallet, ext)
}

// 提名验证者
func (s *StakingService) Nominate(
    wallet *PolkadotWallet,
    validators []string,
) (types.Hash, error) {
    // 获取元数据
    meta, err := s.client.api.RPC.State.GetMetadataLatest()
    if err != nil {
        return types.Hash{}, err
    }
    
    // 创建提名调用
    call, err := types.NewCall(meta, "Staking.nominate", validators)
    if err != nil {
        return types.Hash{}, err
    }
    
    // 创建外部交易
    ext := types.NewExtrinsic(call)
    
    // 签名和提交
    return s.signAndSubmit(wallet, ext)
}

// 解绑质押
func (s *StakingService) Unbond(
    wallet *PolkadotWallet,
    value types.U128,
) (types.Hash, error) {
    // 获取元数据
    meta, err := s.client.api.RPC.State.GetMetadataLatest()
    if err != nil {
        return types.Hash{}, err
    }
    
    // 创建解绑调用
    call, err := types.NewCall(meta, "Staking.unbond", value)
    if err != nil {
        return types.Hash{}, err
    }
    
    // 创建外部交易
    ext := types.NewExtrinsic(call)
    
    // 签名和提交
    return s.signAndSubmit(wallet, ext)
}

// 提取解绑资金
func (s *StakingService) WithdrawUnbonded(
    wallet *PolkadotWallet,
    numSlashingSpans types.U32,
) (types.Hash, error) {
    // 获取元数据
    meta, err := s.client.api.RPC.State.GetMetadataLatest()
    if err != nil {
        return types.Hash{}, err
    }
    
    // 创建提取调用
    call, err := types.NewCall(meta, "Staking.withdrawUnbonded", numSlashingSpans)
    if err != nil {
        return types.Hash{}, err
    }
    
    // 创建外部交易
    ext := types.NewExtrinsic(call)
    
    // 签名和提交
    return s.signAndSubmit(wallet, ext)
}

// 获取质押信息
func (s *StakingService) GetStakingInfo(stash string) (*StakingInfo, error) {
    // 获取质押账本
    stashID, err := types.NewAccountID([]byte(stash))
    if err != nil {
        return nil, err
    }
    
    key, err := types.CreateStorageKey(s.client.api.Metadata, "Staking", "Ledger", stashID[:])
    if err != nil {
        return nil, err
    }
    
    var ledger types.StakingLedger
    ok, err := s.client.api.RPC.State.GetStorageLatest(key, &ledger)
    if err != nil {
        return nil, err
    }
    
    if !ok {
        return nil, fmt.Errorf("质押信息不存在")
    }
    
    return &StakingInfo{
        Stash:      ledger.Stash,
        Controller: ledger.Stash, // 简化处理
        Total:      ledger.Total,
        Active:     ledger.Active,
    }, nil
}

// 签名和提交的辅助方法
func (s *StakingService) signAndSubmit(wallet *PolkadotWallet, ext types.Extrinsic) (types.Hash, error) {
    // 实现签名和提交逻辑
    // ... (省略重复代码)
    return types.Hash{}, nil
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

// Polkadot错误类型
type PolkadotError struct {
    Module  string
    Error   string
    Details string
}

func (e *PolkadotError) Error() string {
    return fmt.Sprintf("Polkadot错误 [%s]: %s - %s", e.Module, e.Error, e.Details)
}

// 解析运行时错误
func ParseRuntimeError(err error) *PolkadotError {
    errStr := err.Error()
    
    // 常见错误模式匹配
    if strings.Contains(errStr, "InsufficientBalance") {
        return &PolkadotError{
            Module:  "Balances",
            Error:   "InsufficientBalance",
            Details: "账户余额不足",
        }
    }
    
    if strings.Contains(errStr, "BadOrigin") {
        return &PolkadotError{
            Module:  "System",
            Error:   "BadOrigin",
            Details: "权限不足",
        }
    }
    
    return &PolkadotError{
        Module:  "Unknown",
        Error:   "UnknownError",
        Details: errStr,
    }
}
```

### 8.2 连接管理

```go
// utils/connection_manager.go
package utils

import (
    "sync"
    "time"
    
    gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
)

// 连接池
type ConnectionPool struct {
    endpoints []string
    clients   []*gsrpc.SubstrateAPI
    current   int
    mu        sync.Mutex
}

func NewConnectionPool(endpoints []string) (*ConnectionPool, error) {
    var clients []*gsrpc.SubstrateAPI
    
    for _, endpoint := range endpoints {
        client, err := gsrpc.NewSubstrateAPI(endpoint)
        if err != nil {
            continue // 跳过无法连接的端点
        }
        clients = append(clients, client)
    }
    
    if len(clients) == 0 {
        return nil, fmt.Errorf("无法连接到任何端点")
    }
    
    return &ConnectionPool{
        endpoints: endpoints,
        clients:   clients,
        current:   0,
    }, nil
}

// 获取客户端（轮询）
func (p *ConnectionPool) GetClient() *gsrpc.SubstrateAPI {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    client := p.clients[p.current]
    p.current = (p.current + 1) % len(p.clients)
    
    return client
}

// 健康检查
func (p *ConnectionPool) HealthCheck() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        p.mu.Lock()
        var healthyClients []*gsrpc.SubstrateAPI
        
        for i, client := range p.clients {
            // 简单的健康检查
            _, err := client.RPC.System.Health()
            if err == nil {
                healthyClients = append(healthyClients, client)
            } else {
                // 尝试重新连接
                newClient, err := gsrpc.NewSubstrateAPI(p.endpoints[i])
                if err == nil {
                    healthyClients = append(healthyClients, newClient)
                }
            }
        }
        
        p.clients = healthyClients
        if p.current >= len(p.clients) {
            p.current = 0
        }
        p.mu.Unlock()
    }
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
    
    "github.com/centrifuge/go-substrate-rpc-client/v4/types"
    
    "your-project/client"
    "your-project/services"
    "your-project/wallet"
)

func main() {
    // 创建配置
    config := &PolkadotConfig{
        RPCURL:  WestendTestnet, // 使用测试网
        Network: "westend",
        ChainID: 42,
    }
    
    // 创建客户端
    polkadotClient, err := client.NewPolkadotClient(config)
    if err != nil {
        log.Fatal("创建Polkadot客户端失败:", err)
    }
    defer polkadotClient.Close()
    
    // 创建钱包
    wallet, err := wallet.NewPolkadotWallet(42) // Westend SS58格式
    if err != nil {
        log.Fatal("创建钱包失败:", err)
    }
    
    fmt.Printf("钱包地址: %s\n", wallet.GetAddress())
    
    // 创建服务
    accountService := services.NewAccountService(polkadotClient)
    transactionService := services.NewTransactionService(polkadotClient)
    stakingService := services.NewStakingService(polkadotClient)
    
    // 查询账户信息
    accountInfo, err := accountService.GetAccountInfo(wallet.GetAddress())
    if err != nil {
        log.Printf("查询账户信息失败: %v", err)
    } else {
        fmt.Printf("账户余额: %s\n", accountInfo.Data.Free.String())
    }
    
    // 获取链信息
    chainInfo, err := polkadotClient.GetChainInfo()
    if err != nil {
        log.Printf("获取链信息失败: %v", err)
    } else {
        fmt.Printf("链信息: %+v\n", chainInfo)
    }
    
    // 如果有余额，进行转账测试
    if accountInfo != nil && accountInfo.Data.Free.Int64() > 1000000000000 { // 1 DOT
        targetWallet, _ := wallet.NewPolkadotWallet(42)
        
        hash, err := transactionService.Transfer(
            wallet,
            targetWallet.GetAddress(),
            types.NewUCompactFromUInt(500000000000), // 0.5 DOT
        )
        if err != nil {
            log.Printf("转账失败: %v", err)
        } else {
            fmt.Printf("转账成功，交易哈希: %s\n", hash.Hex())
        }
    }
}
```

这个Polkadot使用指南提供了完整的多链生态集成方案，涵盖了从基础连接到高级跨链功能的所有核心操作。
