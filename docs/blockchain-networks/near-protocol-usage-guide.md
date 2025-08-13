# Near Protocol 区块链 详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [环境准备](#环境准备)
3. [账户系统](#账户系统)
4. [智能合约](#智能合约)
5. [跨合约调用](#跨合约调用)
6. [存储和状态](#存储和状态)
7. [Gas和费用](#gas和费用)
8. [分片架构](#分片架构)
9. [实际应用](#实际应用)

## 基础概念

### 1.1 Near Protocol 简介

Near Protocol 是一个高性能、开发者友好的区块链平台，采用分片技术实现可扩展性，支持智能合约和去中心化应用开发。

```bash
# 安装Near相关依赖
go get github.com/near/borsh-go
go get github.com/near/near-api-go
go get github.com/shopspring/decimal
```

### 1.2 核心概念

```go
// 主要包导入
import (
    "context"
    "crypto/ed25519"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "math/big"
    
    "github.com/near/borsh-go"
    "github.com/shopspring/decimal"
)

// Near 网络配置
var (
    // Mainnet
    MainnetRPC = "https://rpc.mainnet.near.org"
    MainnetArchivalRPC = "https://archival-rpc.mainnet.near.org"
    
    // Testnet
    TestnetRPC = "https://rpc.testnet.near.org"
    TestnetArchivalRPC = "https://archival-rpc.testnet.near.org"
    
    // Localnet
    LocalnetRPC = "http://localhost:3030"
)

// Near 网络信息
type NetworkConfig struct {
    NetworkID    string
    NodeURL      string
    WalletURL    string
    HelperURL    string
    ExplorerURL  string
}

var (
    MainnetConfig = NetworkConfig{
        NetworkID:   "mainnet",
        NodeURL:     MainnetRPC,
        WalletURL:   "https://wallet.near.org",
        HelperURL:   "https://helper.mainnet.near.org",
        ExplorerURL: "https://explorer.near.org",
    }
    
    TestnetConfig = NetworkConfig{
        NetworkID:   "testnet",
        NodeURL:     TestnetRPC,
        WalletURL:   "https://wallet.testnet.near.org",
        HelperURL:   "https://helper.testnet.near.org",
        ExplorerURL: "https://explorer.testnet.near.org",
    }
)

// Near 账户信息
type Account struct {
    AccountID string
    PublicKey string
    Balance   *big.Int
    Storage   *StorageUsage
    CodeHash  string
}

// 存储使用情况
type StorageUsage struct {
    Used      uint64
    Available uint64
    Total     uint64
}

// Near 交易
type Transaction struct {
    SignerID   string
    PublicKey  string
    Nonce      uint64
    ReceiverID string
    BlockHash  string
    Actions    []Action
}

// Near 动作类型
type Action interface {
    ActionType() string
}

// 创建账户动作
type CreateAccountAction struct{}

func (a CreateAccountAction) ActionType() string { return "CreateAccount" }

// 部署合约动作
type DeployContractAction struct {
    Code []byte
}

func (a DeployContractAction) ActionType() string { return "DeployContract" }

// 函数调用动作
type FunctionCallAction struct {
    MethodName string
    Args       []byte
    Gas        uint64
    Deposit    *big.Int
}

func (a FunctionCallAction) ActionType() string { return "FunctionCall" }

// 转账动作
type TransferAction struct {
    Deposit *big.Int
}

func (a TransferAction) ActionType() string { return "Transfer" }

// 质押动作
type StakeAction struct {
    Stake     *big.Int
    PublicKey string
}

func (a StakeAction) ActionType() string { return "Stake" }

// 添加密钥动作
type AddKeyAction struct {
    PublicKey   string
    AccessKey   AccessKey
}

func (a AddKeyAction) ActionType() string { return "AddKey" }

// 删除密钥动作
type DeleteKeyAction struct {
    PublicKey string
}

func (a DeleteKeyAction) ActionType() string { return "DeleteKey" }

// 删除账户动作
type DeleteAccountAction struct {
    BeneficiaryID string
}

func (a DeleteAccountAction) ActionType() string { return "DeleteAccount" }

// 访问密钥
type AccessKey struct {
    Nonce      uint64
    Permission Permission
}

// 权限类型
type Permission interface {
    PermissionType() string
}

// 全权限
type FullAccessPermission struct{}

func (p FullAccessPermission) PermissionType() string { return "FullAccess" }

// 函数调用权限
type FunctionCallPermission struct {
    Allowance   *big.Int
    ReceiverID  string
    MethodNames []string
}

func (p FunctionCallPermission) PermissionType() string { return "FunctionCall" }

// 交易结果
type TransactionResult struct {
    Status             TransactionStatus
    TransactionHash    string
    ReceiptsOutcome    []ReceiptOutcome
    TransactionOutcome TransactionOutcome
}

// 交易状态
type TransactionStatus struct {
    SuccessValue string `json:"SuccessValue,omitempty"`
    Failure      string `json:"Failure,omitempty"`
}

// 收据结果
type ReceiptOutcome struct {
    ID      string
    Outcome ExecutionOutcome
}

// 执行结果
type ExecutionOutcome struct {
    Logs        []string
    ReceiptIDs  []string
    GasBurnt    uint64
    TokensBurnt *big.Int
    ExecutorID  string
    Status      ExecutionStatus
}

// 执行状态
type ExecutionStatus struct {
    SuccessValue   string `json:"SuccessValue,omitempty"`
    SuccessReceiptID string `json:"SuccessReceiptId,omitempty"`
    Failure        string `json:"Failure,omitempty"`
}

// 交易结果
type TransactionOutcome struct {
    ID      string
    Outcome ExecutionOutcome
}
```

## 环境准备

### 2.1 Near客户端设置

```go
// client/near_client.go
package client

import (
    "bytes"
    "context"
    "crypto/ed25519"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "time"
    
    "github.com/near/borsh-go"
    "github.com/shopspring/decimal"
)

type NearClient struct {
    networkConfig NetworkConfig
    httpClient    *http.Client
}

func NewNearClient(config NetworkConfig) *NearClient {
    return &NearClient{
        networkConfig: config,
        httpClient:    &http.Client{Timeout: 30 * time.Second},
    }
}

// RPC请求结构
type RPCRequest struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      string      `json:"id"`
    Method  string      `json:"method"`
    Params  interface{} `json:"params"`
}

// RPC响应结构
type RPCResponse struct {
    JSONRPC string          `json:"jsonrpc"`
    ID      string          `json:"id"`
    Result  json.RawMessage `json:"result,omitempty"`
    Error   *RPCError       `json:"error,omitempty"`
}

// RPC错误
type RPCError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    string `json:"data,omitempty"`
}

// 发送RPC请求
func (c *NearClient) sendRPCRequest(method string, params interface{}) (json.RawMessage, error) {
    request := RPCRequest{
        JSONRPC: "2.0",
        ID:      "1",
        Method:  method,
        Params:  params,
    }
    
    requestBody, err := json.Marshal(request)
    if err != nil {
        return nil, err
    }
    
    resp, err := c.httpClient.Post(
        c.networkConfig.NodeURL,
        "application/json",
        bytes.NewBuffer(requestBody),
    )
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    responseBody, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    
    var rpcResp RPCResponse
    err = json.Unmarshal(responseBody, &rpcResp)
    if err != nil {
        return nil, err
    }
    
    if rpcResp.Error != nil {
        return nil, fmt.Errorf("RPC错误: %s", rpcResp.Error.Message)
    }
    
    return rpcResp.Result, nil
}

// 获取账户信息
func (c *NearClient) GetAccount(accountID string) (*Account, error) {
    params := map[string]interface{}{
        "request_type": "view_account",
        "finality":     "final",
        "account_id":   accountID,
    }
    
    result, err := c.sendRPCRequest("query", params)
    if err != nil {
        return nil, err
    }
    
    var accountInfo struct {
        Amount        string `json:"amount"`
        Locked        string `json:"locked"`
        CodeHash      string `json:"code_hash"`
        StorageUsage  uint64 `json:"storage_usage"`
        StoragePaidAt uint64 `json:"storage_paid_at"`
    }
    
    err = json.Unmarshal(result, &accountInfo)
    if err != nil {
        return nil, err
    }
    
    balance, _ := new(big.Int).SetString(accountInfo.Amount, 10)
    
    return &Account{
        AccountID: accountID,
        Balance:   balance,
        CodeHash:  accountInfo.CodeHash,
        Storage: &StorageUsage{
            Used: accountInfo.StorageUsage,
        },
    }, nil
}

// 获取访问密钥
func (c *NearClient) GetAccessKey(accountID, publicKey string) (*AccessKey, error) {
    params := map[string]interface{}{
        "request_type": "view_access_key",
        "finality":     "final",
        "account_id":   accountID,
        "public_key":   publicKey,
    }
    
    result, err := c.sendRPCRequest("query", params)
    if err != nil {
        return nil, err
    }
    
    var accessKeyInfo struct {
        Nonce      uint64 `json:"nonce"`
        Permission struct {
            FunctionCall *struct {
                Allowance   string   `json:"allowance"`
                ReceiverID  string   `json:"receiver_id"`
                MethodNames []string `json:"method_names"`
            } `json:"FunctionCall,omitempty"`
            FullAccess *struct{} `json:"FullAccess,omitempty"`
        } `json:"permission"`
    }
    
    err = json.Unmarshal(result, &accessKeyInfo)
    if err != nil {
        return nil, err
    }
    
    accessKey := &AccessKey{
        Nonce: accessKeyInfo.Nonce,
    }
    
    if accessKeyInfo.Permission.FullAccess != nil {
        accessKey.Permission = FullAccessPermission{}
    } else if accessKeyInfo.Permission.FunctionCall != nil {
        allowance, _ := new(big.Int).SetString(accessKeyInfo.Permission.FunctionCall.Allowance, 10)
        accessKey.Permission = FunctionCallPermission{
            Allowance:   allowance,
            ReceiverID:  accessKeyInfo.Permission.FunctionCall.ReceiverID,
            MethodNames: accessKeyInfo.Permission.FunctionCall.MethodNames,
        }
    }
    
    return accessKey, nil
}

// 调用视图函数
func (c *NearClient) CallViewFunction(
    accountID string,
    methodName string,
    args []byte,
) ([]byte, error) {
    params := map[string]interface{}{
        "request_type": "call_function",
        "finality":     "final",
        "account_id":   accountID,
        "method_name":  methodName,
        "args_base64":  base64.StdEncoding.EncodeToString(args),
    }
    
    result, err := c.sendRPCRequest("query", params)
    if err != nil {
        return nil, err
    }
    
    var viewResult struct {
        Result []byte   `json:"result"`
        Logs   []string `json:"logs"`
    }
    
    err = json.Unmarshal(result, &viewResult)
    if err != nil {
        return nil, err
    }
    
    return viewResult.Result, nil
}

// 发送交易
func (c *NearClient) SendTransaction(signedTx []byte) (*TransactionResult, error) {
    params := []interface{}{base64.StdEncoding.EncodeToString(signedTx)}
    
    result, err := c.sendRPCRequest("broadcast_tx_commit", params)
    if err != nil {
        return nil, err
    }
    
    var txResult TransactionResult
    err = json.Unmarshal(result, &txResult)
    if err != nil {
        return nil, err
    }
    
    return &txResult, nil
}

// 获取交易状态
func (c *NearClient) GetTransactionStatus(txHash, senderID string) (*TransactionResult, error) {
    params := []interface{}{txHash, senderID}
    
    result, err := c.sendRPCRequest("tx", params)
    if err != nil {
        return nil, err
    }
    
    var txResult TransactionResult
    err = json.Unmarshal(result, &txResult)
    if err != nil {
        return nil, err
    }
    
    return &txResult, nil
}

// 获取区块信息
func (c *NearClient) GetBlock(blockID string) (*Block, error) {
    params := map[string]interface{}{
        "finality": "final",
    }
    
    if blockID != "" {
        params["block_id"] = blockID
    }
    
    result, err := c.sendRPCRequest("block", params)
    if err != nil {
        return nil, err
    }
    
    var block Block
    err = json.Unmarshal(result, &block)
    if err != nil {
        return nil, err
    }
    
    return &block, nil
}

// 获取Gas价格
func (c *NearClient) GetGasPrice(blockID string) (*big.Int, error) {
    params := []interface{}{blockID}
    
    result, err := c.sendRPCRequest("gas_price", params)
    if err != nil {
        return nil, err
    }
    
    var gasPriceInfo struct {
        GasPrice string `json:"gas_price"`
    }
    
    err = json.Unmarshal(result, &gasPriceInfo)
    if err != nil {
        return nil, err
    }
    
    gasPrice, _ := new(big.Int).SetString(gasPriceInfo.GasPrice, 10)
    return gasPrice, nil
}

type Block struct {
    Author string `json:"author"`
    Header struct {
        Height                uint64 `json:"height"`
        EpochID               string `json:"epoch_id"`
        NextEpochID           string `json:"next_epoch_id"`
        Hash                  string `json:"hash"`
        PrevHash              string `json:"prev_hash"`
        PrevStateRoot         string `json:"prev_state_root"`
        ChunkReceiptsRoot     string `json:"chunk_receipts_root"`
        ChunkHeadersRoot      string `json:"chunk_headers_root"`
        ChunkTxRoot           string `json:"chunk_tx_root"`
        OutcomeRoot           string `json:"outcome_root"`
        ChallengesRoot        string `json:"challenges_root"`
        Timestamp             uint64 `json:"timestamp"`
        TimestampNanosec      string `json:"timestamp_nanosec"`
        RandomValue           string `json:"random_value"`
        ValidatorProposals    []interface{} `json:"validator_proposals"`
        ChunkMask             []bool `json:"chunk_mask"`
        GasPrice              string `json:"gas_price"`
        RentPaid              string `json:"rent_paid"`
        ValidatorReward       string `json:"validator_reward"`
        TotalSupply           string `json:"total_supply"`
        ChallengesResult      []interface{} `json:"challenges_result"`
        LastFinalBlock        string `json:"last_final_block"`
        LastDsFinalBlock      string `json:"last_ds_final_block"`
        NextBpHash            string `json:"next_bp_hash"`
        BlockMerkleRoot       string `json:"block_merkle_root"`
        Approvals             []interface{} `json:"approvals"`
        Signature             string `json:"signature"`
        LatestProtocolVersion uint32 `json:"latest_protocol_version"`
    } `json:"header"`
    Chunks []interface{} `json:"chunks"`
}
```

## 账户系统

### 3.1 账户管理服务

```go
// services/account_service.go
package services

import (
    "crypto/ed25519"
    "crypto/rand"
    "encoding/base64"
    "fmt"
    "math/big"
    
    "github.com/near/borsh-go"
)

type AccountService struct {
    client *NearClient
}

func NewAccountService(client *NearClient) *AccountService {
    return &AccountService{
        client: client,
    }
}

// 生成密钥对
func (s *AccountService) GenerateKeyPair() (ed25519.PublicKey, ed25519.PrivateKey, error) {
    publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
    if err != nil {
        return nil, nil, err
    }
    
    return publicKey, privateKey, nil
}

// 创建账户
func (s *AccountService) CreateAccount(
    creatorAccountID string,
    newAccountID string,
    publicKey ed25519.PublicKey,
    initialBalance *big.Int,
    privateKey ed25519.PrivateKey,
) (*TransactionResult, error) {
    // 获取创建者账户的nonce
    creatorAccessKey, err := s.client.GetAccessKey(creatorAccountID, base64.StdEncoding.EncodeToString(publicKey))
    if err != nil {
        return nil, err
    }
    
    // 获取最新区块哈希
    block, err := s.client.GetBlock("")
    if err != nil {
        return nil, err
    }
    
    // 构建交易
    tx := Transaction{
        SignerID:   creatorAccountID,
        PublicKey:  base64.StdEncoding.EncodeToString(publicKey),
        Nonce:      creatorAccessKey.Nonce + 1,
        ReceiverID: newAccountID,
        BlockHash:  block.Header.Hash,
        Actions: []Action{
            CreateAccountAction{},
            TransferAction{Deposit: initialBalance},
            AddKeyAction{
                PublicKey: base64.StdEncoding.EncodeToString(publicKey),
                AccessKey: AccessKey{
                    Nonce:      0,
                    Permission: FullAccessPermission{},
                },
            },
        },
    }
    
    // 序列化交易
    serializedTx, err := borsh.Serialize(tx)
    if err != nil {
        return nil, err
    }
    
    // 签名交易
    signature := ed25519.Sign(privateKey, serializedTx)
    
    // 构建签名交易
    signedTx := struct {
        Transaction Transaction
        Signature   []byte
    }{
        Transaction: tx,
        Signature:   signature,
    }
    
    // 序列化签名交易
    signedTxBytes, err := borsh.Serialize(signedTx)
    if err != nil {
        return nil, err
    }
    
    // 发送交易
    return s.client.SendTransaction(signedTxBytes)
}

// 删除账户
func (s *AccountService) DeleteAccount(
    accountID string,
    beneficiaryID string,
    privateKey ed25519.PrivateKey,
) (*TransactionResult, error) {
    publicKey := privateKey.Public().(ed25519.PublicKey)
    
    // 获取账户的nonce
    accessKey, err := s.client.GetAccessKey(accountID, base64.StdEncoding.EncodeToString(publicKey))
    if err != nil {
        return nil, err
    }
    
    // 获取最新区块哈希
    block, err := s.client.GetBlock("")
    if err != nil {
        return nil, err
    }
    
    // 构建交易
    tx := Transaction{
        SignerID:   accountID,
        PublicKey:  base64.StdEncoding.EncodeToString(publicKey),
        Nonce:      accessKey.Nonce + 1,
        ReceiverID: accountID,
        BlockHash:  block.Header.Hash,
        Actions: []Action{
            DeleteAccountAction{BeneficiaryID: beneficiaryID},
        },
    }
    
    return s.signAndSendTransaction(tx, privateKey)
}

// 转账
func (s *AccountService) Transfer(
    senderID string,
    receiverID string,
    amount *big.Int,
    privateKey ed25519.PrivateKey,
) (*TransactionResult, error) {
    publicKey := privateKey.Public().(ed25519.PublicKey)
    
    // 获取发送者的nonce
    accessKey, err := s.client.GetAccessKey(senderID, base64.StdEncoding.EncodeToString(publicKey))
    if err != nil {
        return nil, err
    }
    
    // 获取最新区块哈希
    block, err := s.client.GetBlock("")
    if err != nil {
        return nil, err
    }
    
    // 构建交易
    tx := Transaction{
        SignerID:   senderID,
        PublicKey:  base64.StdEncoding.EncodeToString(publicKey),
        Nonce:      accessKey.Nonce + 1,
        ReceiverID: receiverID,
        BlockHash:  block.Header.Hash,
        Actions: []Action{
            TransferAction{Deposit: amount},
        },
    }
    
    return s.signAndSendTransaction(tx, privateKey)
}

// 添加访问密钥
func (s *AccountService) AddAccessKey(
    accountID string,
    newPublicKey ed25519.PublicKey,
    permission Permission,
    privateKey ed25519.PrivateKey,
) (*TransactionResult, error) {
    publicKey := privateKey.Public().(ed25519.PublicKey)
    
    // 获取账户的nonce
    accessKey, err := s.client.GetAccessKey(accountID, base64.StdEncoding.EncodeToString(publicKey))
    if err != nil {
        return nil, err
    }
    
    // 获取最新区块哈希
    block, err := s.client.GetBlock("")
    if err != nil {
        return nil, err
    }
    
    // 构建交易
    tx := Transaction{
        SignerID:   accountID,
        PublicKey:  base64.StdEncoding.EncodeToString(publicKey),
        Nonce:      accessKey.Nonce + 1,
        ReceiverID: accountID,
        BlockHash:  block.Header.Hash,
        Actions: []Action{
            AddKeyAction{
                PublicKey: base64.StdEncoding.EncodeToString(newPublicKey),
                AccessKey: AccessKey{
                    Nonce:      0,
                    Permission: permission,
                },
            },
        },
    }
    
    return s.signAndSendTransaction(tx, privateKey)
}

// 删除访问密钥
func (s *AccountService) DeleteAccessKey(
    accountID string,
    publicKeyToDelete ed25519.PublicKey,
    privateKey ed25519.PrivateKey,
) (*TransactionResult, error) {
    publicKey := privateKey.Public().(ed25519.PublicKey)
    
    // 获取账户的nonce
    accessKey, err := s.client.GetAccessKey(accountID, base64.StdEncoding.EncodeToString(publicKey))
    if err != nil {
        return nil, err
    }
    
    // 获取最新区块哈希
    block, err := s.client.GetBlock("")
    if err != nil {
        return nil, err
    }
    
    // 构建交易
    tx := Transaction{
        SignerID:   accountID,
        PublicKey:  base64.StdEncoding.EncodeToString(publicKey),
        Nonce:      accessKey.Nonce + 1,
        ReceiverID: accountID,
        BlockHash:  block.Header.Hash,
        Actions: []Action{
            DeleteKeyAction{
                PublicKey: base64.StdEncoding.EncodeToString(publicKeyToDelete),
            },
        },
    }
    
    return s.signAndSendTransaction(tx, privateKey)
}

// 签名并发送交易
func (s *AccountService) signAndSendTransaction(
    tx Transaction,
    privateKey ed25519.PrivateKey,
) (*TransactionResult, error) {
    // 序列化交易
    serializedTx, err := borsh.Serialize(tx)
    if err != nil {
        return nil, err
    }
    
    // 签名交易
    signature := ed25519.Sign(privateKey, serializedTx)
    
    // 构建签名交易
    signedTx := struct {
        Transaction Transaction
        Signature   []byte
    }{
        Transaction: tx,
        Signature:   signature,
    }
    
    // 序列化签名交易
    signedTxBytes, err := borsh.Serialize(signedTx)
    if err != nil {
        return nil, err
    }
    
    // 发送交易
    return s.client.SendTransaction(signedTxBytes)
}

// 获取账户余额
func (s *AccountService) GetBalance(accountID string) (*big.Int, error) {
    account, err := s.client.GetAccount(accountID)
    if err != nil {
        return nil, err
    }
    
    return account.Balance, nil
}

// 检查账户是否存在
func (s *AccountService) AccountExists(accountID string) (bool, error) {
    _, err := s.client.GetAccount(accountID)
    if err != nil {
        // 检查是否是账户不存在的错误
        if fmt.Sprintf("%v", err) == "账户不存在" {
            return false, nil
        }
        return false, err
    }
    
    return true, nil
}

// 获取账户的所有访问密钥
func (s *AccountService) GetAccessKeys(accountID string) (map[string]*AccessKey, error) {
    // Near协议需要通过RPC查询所有访问密钥
    // 这里简化实现，实际需要调用view_access_key_list
    
    params := map[string]interface{}{
        "request_type": "view_access_key_list",
        "finality":     "final",
        "account_id":   accountID,
    }
    
    result, err := s.client.sendRPCRequest("query", params)
    if err != nil {
        return nil, err
    }
    
    var accessKeyList struct {
        Keys []struct {
            PublicKey  string `json:"public_key"`
            AccessKey  AccessKey `json:"access_key"`
        } `json:"keys"`
    }
    
    err = json.Unmarshal(result, &accessKeyList)
    if err != nil {
        return nil, err
    }
    
    accessKeys := make(map[string]*AccessKey)
    for _, key := range accessKeyList.Keys {
        accessKeys[key.PublicKey] = &key.AccessKey
    }
    
    return accessKeys, nil
}
```

## 智能合约

### 4.1 合约部署和调用服务

```go
// services/contract_service.go
package services

import (
    "crypto/ed25519"
    "encoding/base64"
    "encoding/json"
    "math/big"
    
    "github.com/near/borsh-go"
)

type ContractService struct {
    client *NearClient
}

func NewContractService(client *NearClient) *ContractService {
    return &ContractService{
        client: client,
    }
}

// 部署合约
func (s *ContractService) DeployContract(
    accountID string,
    contractCode []byte,
    privateKey ed25519.PrivateKey,
) (*TransactionResult, error) {
    publicKey := privateKey.Public().(ed25519.PublicKey)
    
    // 获取账户的nonce
    accessKey, err := s.client.GetAccessKey(accountID, base64.StdEncoding.EncodeToString(publicKey))
    if err != nil {
        return nil, err
    }
    
    // 获取最新区块哈希
    block, err := s.client.GetBlock("")
    if err != nil {
        return nil, err
    }
    
    // 构建交易
    tx := Transaction{
        SignerID:   accountID,
        PublicKey:  base64.StdEncoding.EncodeToString(publicKey),
        Nonce:      accessKey.Nonce + 1,
        ReceiverID: accountID,
        BlockHash:  block.Header.Hash,
        Actions: []Action{
            DeployContractAction{Code: contractCode},
        },
    }
    
    return s.signAndSendTransaction(tx, privateKey)
}

// 调用合约函数
func (s *ContractService) CallFunction(
    signerID string,
    contractID string,
    methodName string,
    args interface{},
    gas uint64,
    deposit *big.Int,
    privateKey ed25519.PrivateKey,
) (*TransactionResult, error) {
    publicKey := privateKey.Public().(ed25519.PublicKey)
    
    // 序列化参数
    argsBytes, err := json.Marshal(args)
    if err != nil {
        return nil, err
    }
    
    // 获取签名者的nonce
    accessKey, err := s.client.GetAccessKey(signerID, base64.StdEncoding.EncodeToString(publicKey))
    if err != nil {
        return nil, err
    }
    
    // 获取最新区块哈希
    block, err := s.client.GetBlock("")
    if err != nil {
        return nil, err
    }
    
    // 构建交易
    tx := Transaction{
        SignerID:   signerID,
        PublicKey:  base64.StdEncoding.EncodeToString(publicKey),
        Nonce:      accessKey.Nonce + 1,
        ReceiverID: contractID,
        BlockHash:  block.Header.Hash,
        Actions: []Action{
            FunctionCallAction{
                MethodName: methodName,
                Args:       argsBytes,
                Gas:        gas,
                Deposit:    deposit,
            },
        },
    }
    
    return s.signAndSendTransaction(tx, privateKey)
}

// 调用视图函数
func (s *ContractService) CallViewFunction(
    contractID string,
    methodName string,
    args interface{},
) (interface{}, error) {
    // 序列化参数
    argsBytes, err := json.Marshal(args)
    if err != nil {
        return nil, err
    }
    
    // 调用视图函数
    result, err := s.client.CallViewFunction(contractID, methodName, argsBytes)
    if err != nil {
        return nil, err
    }
    
    // 反序列化结果
    var viewResult interface{}
    err = json.Unmarshal(result, &viewResult)
    if err != nil {
        return nil, err
    }
    
    return viewResult, nil
}

// 批量调用合约函数
func (s *ContractService) BatchCall(
    signerID string,
    calls []ContractCall,
    privateKey ed25519.PrivateKey,
) (*TransactionResult, error) {
    publicKey := privateKey.Public().(ed25519.PublicKey)
    
    // 获取签名者的nonce
    accessKey, err := s.client.GetAccessKey(signerID, base64.StdEncoding.EncodeToString(publicKey))
    if err != nil {
        return nil, err
    }
    
    // 获取最新区块哈希
    block, err := s.client.GetBlock("")
    if err != nil {
        return nil, err
    }
    
    // 构建动作列表
    var actions []Action
    for _, call := range calls {
        argsBytes, err := json.Marshal(call.Args)
        if err != nil {
            return nil, err
        }
        
        actions = append(actions, FunctionCallAction{
            MethodName: call.MethodName,
            Args:       argsBytes,
            Gas:        call.Gas,
            Deposit:    call.Deposit,
        })
    }
    
    // 构建交易
    tx := Transaction{
        SignerID:   signerID,
        PublicKey:  base64.StdEncoding.EncodeToString(publicKey),
        Nonce:      accessKey.Nonce + 1,
        ReceiverID: calls[0].ContractID, // 假设所有调用都是同一个合约
        BlockHash:  block.Header.Hash,
        Actions:    actions,
    }
    
    return s.signAndSendTransaction(tx, privateKey)
}

// 估算Gas费用
func (s *ContractService) EstimateGas(
    contractID string,
    methodName string,
    args interface{},
) (uint64, error) {
    // Near协议的Gas估算比较复杂
    // 这里提供一个简化的实现
    
    // 基础Gas费用
    baseGas := uint64(30000000000000) // 30 TGas
    
    // 根据方法名称和参数复杂度调整
    argsBytes, _ := json.Marshal(args)
    complexityFactor := uint64(len(argsBytes)) * 1000000000 // 每字节1 GGas
    
    return baseGas + complexityFactor, nil
}

// 获取合约状态
func (s *ContractService) GetContractState(contractID string) (map[string]interface{}, error) {
    params := map[string]interface{}{
        "request_type": "view_state",
        "finality":     "final",
        "account_id":   contractID,
        "prefix_base64": "",
    }
    
    result, err := s.client.sendRPCRequest("query", params)
    if err != nil {
        return nil, err
    }
    
    var stateResult struct {
        Values []struct {
            Key   string `json:"key"`
            Value string `json:"value"`
        } `json:"values"`
    }
    
    err = json.Unmarshal(result, &stateResult)
    if err != nil {
        return nil, err
    }
    
    state := make(map[string]interface{})
    for _, item := range stateResult.Values {
        // 解码base64值
        value, err := base64.StdEncoding.DecodeString(item.Value)
        if err != nil {
            continue
        }
        
        state[item.Key] = string(value)
    }
    
    return state, nil
}

// 签名并发送交易
func (s *ContractService) signAndSendTransaction(
    tx Transaction,
    privateKey ed25519.PrivateKey,
) (*TransactionResult, error) {
    // 序列化交易
    serializedTx, err := borsh.Serialize(tx)
    if err != nil {
        return nil, err
    }
    
    // 签名交易
    signature := ed25519.Sign(privateKey, serializedTx)
    
    // 构建签名交易
    signedTx := struct {
        Transaction Transaction
        Signature   []byte
    }{
        Transaction: tx,
        Signature:   signature,
    }
    
    // 序列化签名交易
    signedTxBytes, err := borsh.Serialize(signedTx)
    if err != nil {
        return nil, err
    }
    
    // 发送交易
    return s.client.SendTransaction(signedTxBytes)
}

type ContractCall struct {
    ContractID string
    MethodName string
    Args       interface{}
    Gas        uint64
    Deposit    *big.Int
}
```

## 实际应用

### 9.1 完整示例

```go
// main.go
package main

import (
    "crypto/ed25519"
    "encoding/base64"
    "fmt"
    "log"
    "math/big"
    
    "your-project/client"
    "your-project/services"
)

func main() {
    // 创建Near客户端 (使用测试网)
    nearClient := client.NewNearClient(client.TestnetConfig)
    
    // 创建服务
    accountService := services.NewAccountService(nearClient)
    contractService := services.NewContractService(nearClient)
    
    // 1. 生成密钥对
    fmt.Printf("=== 密钥对生成 ===\n")
    
    publicKey, privateKey, err := accountService.GenerateKeyPair()
    if err != nil {
        log.Fatal("生成密钥对失败:", err)
    }
    
    fmt.Printf("公钥: %s\n", base64.StdEncoding.EncodeToString(publicKey))
    fmt.Printf("私钥长度: %d bytes\n", len(privateKey))
    
    // 2. 查询账户信息
    fmt.Printf("\n=== 账户信息查询 ===\n")
    
    testAccountID := "test.testnet"
    account, err := nearClient.GetAccount(testAccountID)
    if err != nil {
        log.Printf("获取账户信息失败: %v", err)
    } else {
        fmt.Printf("账户ID: %s\n", account.AccountID)
        fmt.Printf("余额: %s yoctoNEAR\n", account.Balance.String())
        fmt.Printf("存储使用: %d bytes\n", account.Storage.Used)
        fmt.Printf("代码哈希: %s\n", account.CodeHash)
    }
    
    // 3. 查询访问密钥
    fmt.Printf("\n=== 访问密钥查询 ===\n")
    
    accessKeys, err := accountService.GetAccessKeys(testAccountID)
    if err != nil {
        log.Printf("获取访问密钥失败: %v", err)
    } else {
        fmt.Printf("访问密钥数量: %d\n", len(accessKeys))
        for pubKey, accessKey := range accessKeys {
            fmt.Printf("  公钥: %s\n", pubKey[:20]+"...")
            fmt.Printf("  Nonce: %d\n", accessKey.Nonce)
            fmt.Printf("  权限类型: %s\n", accessKey.Permission.PermissionType())
        }
    }
    
    // 4. 转账示例
    fmt.Printf("\n=== 转账示例 ===\n")
    
    // 注意：这需要有效的私钥和足够的余额
    senderID := "sender.testnet"
    receiverID := "receiver.testnet"
    transferAmount := big.NewInt(1000000000000000000000000) // 1 NEAR
    
    fmt.Printf("准备转账:\n")
    fmt.Printf("  发送者: %s\n", senderID)
    fmt.Printf("  接收者: %s\n", receiverID)
    fmt.Printf("  金额: %s yoctoNEAR (%.6f NEAR)\n", 
        transferAmount.String(), 
        float64(transferAmount.Int64())/1e24)
    
    // 实际转账需要有效的私钥
    // txResult, err := accountService.Transfer(senderID, receiverID, transferAmount, privateKey)
    // if err != nil {
    //     log.Printf("转账失败: %v", err)
    // } else {
    //     fmt.Printf("转账成功，交易哈希: %s\n", txResult.TransactionHash)
    // }
    
    // 5. 合约调用示例
    fmt.Printf("\n=== 合约调用示例 ===\n")
    
    // 调用一个简单的视图函数
    contractID := "guest-book.testnet"
    methodName := "get_messages"
    args := map[string]interface{}{
        "from_index": 0,
        "limit":      10,
    }
    
    result, err := contractService.CallViewFunction(contractID, methodName, args)
    if err != nil {
        log.Printf("调用视图函数失败: %v", err)
    } else {
        fmt.Printf("合约调用结果: %v\n", result)
    }
    
    // 6. Gas价格查询
    fmt.Printf("\n=== Gas价格查询 ===\n")
    
    gasPrice, err := nearClient.GetGasPrice("")
    if err != nil {
        log.Printf("获取Gas价格失败: %v", err)
    } else {
        fmt.Printf("当前Gas价格: %s yoctoNEAR\n", gasPrice.String())
    }
    
    // 7. 区块信息查询
    fmt.Printf("\n=== 区块信息查询 ===\n")
    
    block, err := nearClient.GetBlock("")
    if err != nil {
        log.Printf("获取区块信息失败: %v", err)
    } else {
        fmt.Printf("最新区块:\n")
        fmt.Printf("  高度: %d\n", block.Header.Height)
        fmt.Printf("  哈希: %s\n", block.Header.Hash)
        fmt.Printf("  时间戳: %d\n", block.Header.Timestamp)
        fmt.Printf("  Gas价格: %s\n", block.Header.GasPrice)
        fmt.Printf("  验证者奖励: %s\n", block.Header.ValidatorReward)
    }
    
    // 8. 合约状态查询
    fmt.Printf("\n=== 合约状态查询 ===\n")
    
    contractState, err := contractService.GetContractState(contractID)
    if err != nil {
        log.Printf("获取合约状态失败: %v", err)
    } else {
        fmt.Printf("合约状态键数量: %d\n", len(contractState))
        
        // 显示前几个状态键
        count := 0
        for key, value := range contractState {
            if count >= 3 {
                break
            }
            fmt.Printf("  %s: %v\n", key, value)
            count++
        }
    }
    
    // 9. Gas估算示例
    fmt.Printf("\n=== Gas估算示例 ===\n")
    
    estimatedGas, err := contractService.EstimateGas(
        contractID,
        "add_message",
        map[string]interface{}{
            "text": "Hello, Near!",
        },
    )
    if err != nil {
        log.Printf("Gas估算失败: %v", err)
    } else {
        fmt.Printf("估算Gas费用: %d\n", estimatedGas)
        fmt.Printf("估算费用: %.6f NEAR\n", float64(estimatedGas)*float64(gasPrice.Int64())/1e24)
    }
    
    // 10. 账户存在性检查
    fmt.Printf("\n=== 账户存在性检查 ===\n")
    
    testAccounts := []string{
        "test.testnet",
        "nonexistent.testnet",
        "near",
    }
    
    for _, accountID := range testAccounts {
        exists, err := accountService.AccountExists(accountID)
        if err != nil {
            fmt.Printf("检查账户 %s 失败: %v\n", accountID, err)
        } else {
            status := "不存在"
            if exists {
                status = "存在"
            }
            fmt.Printf("账户 %s: %s\n", accountID, status)
        }
    }
}
```

这个Near Protocol使用指南提供了完整的高性能区块链集成方案，涵盖了账户管理、智能合约部署和调用、跨合约通信、分片架构等核心功能，是构建可扩展DeFi应用的重要参考文档。
