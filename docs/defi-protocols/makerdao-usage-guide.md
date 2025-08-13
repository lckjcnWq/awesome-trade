# MakerDAO 协议 详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [环境准备](#环境准备)
3. [CDP系统](#cdp系统)
4. [DAI稳定币](#dai稳定币)
5. [抵押品管理](#抵押品管理)
6. [清算机制](#清算机制)
7. [治理系统](#治理系统)
8. [储蓄利率](#储蓄利率)
9. [实际应用](#实际应用)

## 基础概念

### 1.1 MakerDAO 简介

MakerDAO 是以太坊上最重要的去中心化稳定币协议，通过超额抵押机制发行 DAI 稳定币，是 DeFi 生态的基础设施之一。

```bash
# 安装MakerDAO相关依赖
go get github.com/ethereum/go-ethereum
go get github.com/ethereum/go-ethereum/accounts/abi
go get github.com/ethereum/go-ethereum/accounts/abi/bind
go get github.com/shopspring/decimal
```

### 1.2 核心概念

```go
// 主要包导入
import (
    "context"
    "math/big"
    "strings"
    
    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/shopspring/decimal"
)

// MakerDAO 核心合约地址 (Mainnet)
var (
    // CDP Manager
    CDPManagerAddress    = common.HexToAddress("0x5ef30b9986345249bc32d8928B7ee64DE9435E39")
    
    // DAI Token
    DAITokenAddress      = common.HexToAddress("0x6B175474E89094C44Da98b954EedeAC495271d0F")
    
    // MKR Token
    MKRTokenAddress      = common.HexToAddress("0x9f8F72aA9304c8B593d555F12eF6589cC3A579A2")
    
    // Vat (核心会计系统)
    VatAddress           = common.HexToAddress("0x35D1b3F3D7966A1DFe207aa4514C12a259A0492B")
    
    // Jug (稳定费率)
    JugAddress           = common.HexToAddress("0x19c0976f590D67707E62397C87829d896Dc0f1F1")
    
    // Pot (DAI储蓄利率)
    PotAddress           = common.HexToAddress("0x197E90f9FAD81970bA7976f33CbD77088E5D7cf7")
    
    // Spotter (价格预言机)
    SpotterAddress       = common.HexToAddress("0x65C79fcB50Ca1594B025960e539eD7A9a6D434A3")
    
    // Cat (清算系统)
    CatAddress           = common.HexToAddress("0xa5679C04fc3d9d8b0AaB1F0ab83555b301cA70Ea")
    
    // Proxy Registry
    ProxyRegistryAddress = common.HexToAddress("0x4678f0a6958e4D2Bc4F1BAF7Bc52E8F3564f3fE4")
)

// 抵押品类型
type CollateralType struct {
    Ilk         [32]byte  // 抵押品标识符
    Name        string    // 抵押品名称 (如 "ETH-A")
    Gem         common.Address // 抵押品代币地址
    Join        common.Address // Join适配器地址
    Flip        common.Address // 拍卖合约地址
    Pip         common.Address // 价格预言机地址
}

// CDP信息
type CDP struct {
    ID           *big.Int
    Owner        common.Address
    Ilk          [32]byte
    Collateral   *big.Int  // 抵押品数量 (WAD)
    Debt         *big.Int  // 债务数量 (WAD)
    CollateralUSD *big.Int // 抵押品美元价值
    DebtUSD      *big.Int  // 债务美元价值
    Ratio        decimal.Decimal // 抵押率
}

// 系统参数
type SystemParams struct {
    GlobalDebt     *big.Int // 全局债务上限
    GlobalDebtUsed *big.Int // 已使用的全局债务
    BaseRate       *big.Int // 基础利率
    GlobalStabilityFee *big.Int // 全局稳定费
}
```

## 环境准备

### 2.1 合约ABI定义

```go
// contracts/maker_abi.go
package contracts

// CDP Manager ABI (简化版)
const CDPManagerABI = `[
    {
        "constant": false,
        "inputs": [
            {"name": "ilk", "type": "bytes32"},
            {"name": "usr", "type": "address"}
        ],
        "name": "open",
        "outputs": [{"name": "cdp", "type": "uint256"}],
        "type": "function"
    },
    {
        "constant": true,
        "inputs": [{"name": "cdp", "type": "uint256"}],
        "name": "urns",
        "outputs": [{"name": "urn", "type": "address"}],
        "type": "function"
    },
    {
        "constant": true,
        "inputs": [{"name": "cdp", "type": "uint256"}],
        "name": "owns",
        "outputs": [{"name": "owner", "type": "address"}],
        "type": "function"
    },
    {
        "constant": true,
        "inputs": [{"name": "cdp", "type": "uint256"}],
        "name": "ilks",
        "outputs": [{"name": "ilk", "type": "bytes32"}],
        "type": "function"
    }
]`

// Vat ABI (简化版)
const VatABI = `[
    {
        "constant": true,
        "inputs": [
            {"name": "ilk", "type": "bytes32"},
            {"name": "urn", "type": "address"}
        ],
        "name": "urns",
        "outputs": [
            {"name": "ink", "type": "uint256"},
            {"name": "art", "type": "uint256"}
        ],
        "type": "function"
    },
    {
        "constant": true,
        "inputs": [{"name": "ilk", "type": "bytes32"}],
        "name": "ilks",
        "outputs": [
            {"name": "Art", "type": "uint256"},
            {"name": "rate", "type": "uint256"},
            {"name": "spot", "type": "uint256"},
            {"name": "line", "type": "uint256"},
            {"name": "dust", "type": "uint256"}
        ],
        "type": "function"
    },
    {
        "constant": false,
        "inputs": [
            {"name": "ilk", "type": "bytes32"},
            {"name": "dink", "type": "int256"},
            {"name": "dart", "type": "int256"}
        ],
        "name": "frob",
        "outputs": [],
        "type": "function"
    }
]`

// Pot ABI (简化版)
const PotABI = `[
    {
        "constant": false,
        "inputs": [{"name": "wad", "type": "uint256"}],
        "name": "join",
        "outputs": [],
        "type": "function"
    },
    {
        "constant": false,
        "inputs": [{"name": "wad", "type": "uint256"}],
        "name": "exit",
        "outputs": [],
        "type": "function"
    },
    {
        "constant": true,
        "inputs": [{"name": "usr", "type": "address"}],
        "name": "pie",
        "outputs": [{"name": "wad", "type": "uint256"}],
        "type": "function"
    },
    {
        "constant": true,
        "inputs": [],
        "name": "dsr",
        "outputs": [{"name": "ray", "type": "uint256"}],
        "type": "function"
    },
    {
        "constant": false,
        "inputs": [],
        "name": "drip",
        "outputs": [{"name": "tmp", "type": "uint256"}],
        "type": "function"
    }
]`
```

### 2.2 客户端设置

```go
// client/maker_client.go
package client

import (
    "context"
    "math/big"
    
    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
)

type MakerClient struct {
    ethClient       *ethclient.Client
    cdpManagerABI   abi.ABI
    vatABI          abi.ABI
    potABI          abi.ABI
    cdpManagerAddr  common.Address
    vatAddr         common.Address
    potAddr         common.Address
}

func NewMakerClient(rpcURL string) (*MakerClient, error) {
    ethClient, err := ethclient.Dial(rpcURL)
    if err != nil {
        return nil, err
    }
    
    cdpManagerABI, err := abi.JSON(strings.NewReader(CDPManagerABI))
    if err != nil {
        return nil, err
    }
    
    vatABI, err := abi.JSON(strings.NewReader(VatABI))
    if err != nil {
        return nil, err
    }
    
    potABI, err := abi.JSON(strings.NewReader(PotABI))
    if err != nil {
        return nil, err
    }
    
    return &MakerClient{
        ethClient:      ethClient,
        cdpManagerABI:  cdpManagerABI,
        vatABI:         vatABI,
        potABI:         potABI,
        cdpManagerAddr: CDPManagerAddress,
        vatAddr:        VatAddress,
        potAddr:        PotAddress,
    }, nil
}

// 获取CDP信息
func (c *MakerClient) GetCDPInfo(cdpID *big.Int) (*CDP, error) {
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    // 获取CDP所有者
    var owner common.Address
    err := c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &c.cdpManagerAddr,
        Data: c.cdpManagerABI.Methods["owns"].ID,
    }, nil)
    if err != nil {
        return nil, err
    }
    
    // 获取抵押品类型
    var ilk [32]byte
    err = c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &c.cdpManagerAddr,
        Data: c.cdpManagerABI.Methods["ilks"].ID,
    }, nil)
    if err != nil {
        return nil, err
    }
    
    // 获取Urn地址
    var urnAddr common.Address
    err = c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &c.cdpManagerAddr,
        Data: c.cdpManagerABI.Methods["urns"].ID,
    }, nil)
    if err != nil {
        return nil, err
    }
    
    // 从Vat获取抵押品和债务信息
    var ink, art *big.Int
    err = c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &c.vatAddr,
        Data: c.vatABI.Methods["urns"].ID,
    }, nil)
    if err != nil {
        return nil, err
    }
    
    return &CDP{
        ID:         cdpID,
        Owner:      owner,
        Ilk:        ilk,
        Collateral: ink,
        Debt:       art,
    }, nil
}
```

## CDP系统

### 3.1 CDP管理服务

```go
// services/cdp_service.go
package services

import (
    "context"
    "crypto/ecdsa"
    "math/big"
    
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/crypto"
)

type CDPService struct {
    client     *MakerClient
    privateKey *ecdsa.PrivateKey
}

func NewCDPService(client *MakerClient, privateKey *ecdsa.PrivateKey) *CDPService {
    return &CDPService{
        client:     client,
        privateKey: privateKey,
    }
}

// 抵押品类型常量
var (
    ETH_A = [32]byte{0x45, 0x54, 0x48, 0x2d, 0x41} // "ETH-A"
    ETH_B = [32]byte{0x45, 0x54, 0x48, 0x2d, 0x42} // "ETH-B"
    ETH_C = [32]byte{0x45, 0x54, 0x48, 0x2d, 0x43} // "ETH-C"
)

// 开启新的CDP
func (s *CDPService) OpenCDP(ilk [32]byte) (*big.Int, error) {
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 获取nonce
    nonce, err := s.client.ethClient.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        return nil, err
    }
    
    // 获取gas价格
    gasPrice, err := s.client.ethClient.SuggestGasPrice(context.Background())
    if err != nil {
        return nil, err
    }
    
    // 构建open交易数据
    data, err := s.client.cdpManagerABI.Pack("open", ilk, fromAddress)
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, s.client.cdpManagerAddr, big.NewInt(0), 200000, gasPrice, data)
    
    // 签名交易
    chainID, err := s.client.ethClient.NetworkID(context.Background())
    if err != nil {
        return nil, err
    }
    
    signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), s.privateKey)
    if err != nil {
        return nil, err
    }
    
    // 发送交易
    err = s.client.ethClient.SendTransaction(context.Background(), signedTx)
    if err != nil {
        return nil, err
    }
    
    // 等待交易确认并获取CDP ID
    receipt, err := s.waitForReceipt(signedTx.Hash())
    if err != nil {
        return nil, err
    }
    
    // 从事件日志中解析CDP ID
    cdpID := s.parseCDPIDFromLogs(receipt.Logs)
    
    return cdpID, nil
}

// 存入抵押品
func (s *CDPService) LockCollateral(
    cdpID *big.Int,
    collateralAmount *big.Int,
    collateralToken common.Address,
) (*types.Transaction, error) {
    // 首先需要授权Join适配器使用抵押品
    joinAdapter, err := s.getJoinAdapter(cdpID)
    if err != nil {
        return nil, err
    }
    
    if err := s.approveToken(collateralToken, joinAdapter, collateralAmount); err != nil {
        return nil, err
    }
    
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 获取nonce
    nonce, err := s.client.ethClient.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        return nil, err
    }
    
    // 获取gas价格
    gasPrice, err := s.client.ethClient.SuggestGasPrice(context.Background())
    if err != nil {
        return nil, err
    }
    
    // 构建lockGem交易数据 (通过代理合约)
    data, err := s.buildLockGemData(cdpID, collateralAmount)
    if err != nil {
        return nil, err
    }
    
    // 获取用户的代理合约地址
    proxyAddr, err := s.getUserProxy(fromAddress)
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, proxyAddr, big.NewInt(0), 300000, gasPrice, data)
    
    // 签名交易
    chainID, err := s.client.ethClient.NetworkID(context.Background())
    if err != nil {
        return nil, err
    }
    
    signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), s.privateKey)
    if err != nil {
        return nil, err
    }
    
    // 发送交易
    err = s.client.ethClient.SendTransaction(context.Background(), signedTx)
    if err != nil {
        return nil, err
    }
    
    return signedTx, nil
}

// 生成DAI
func (s *CDPService) DrawDAI(cdpID *big.Int, daiAmount *big.Int) (*types.Transaction, error) {
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 获取用户的代理合约地址
    proxyAddr, err := s.getUserProxy(fromAddress)
    if err != nil {
        return nil, err
    }
    
    // 获取nonce
    nonce, err := s.client.ethClient.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        return nil, err
    }
    
    // 获取gas价格
    gasPrice, err := s.client.ethClient.SuggestGasPrice(context.Background())
    if err != nil {
        return nil, err
    }
    
    // 构建draw交易数据
    data, err := s.buildDrawData(cdpID, daiAmount)
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, proxyAddr, big.NewInt(0), 250000, gasPrice, data)
    
    // 签名交易
    chainID, err := s.client.ethClient.NetworkID(context.Background())
    if err != nil {
        return nil, err
    }
    
    signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), s.privateKey)
    if err != nil {
        return nil, err
    }
    
    // 发送交易
    err = s.client.ethClient.SendTransaction(context.Background(), signedTx)
    if err != nil {
        return nil, err
    }
    
    return signedTx, nil
}

// 偿还DAI
func (s *CDPService) WipeDAI(cdpID *big.Int, daiAmount *big.Int) (*types.Transaction, error) {
    // 首先授权代理合约使用DAI
    proxyAddr, err := s.getUserProxy(crypto.PubkeyToAddress(s.privateKey.PublicKey))
    if err != nil {
        return nil, err
    }
    
    if err := s.approveToken(DAITokenAddress, proxyAddr, daiAmount); err != nil {
        return nil, err
    }
    
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 获取nonce
    nonce, err := s.client.ethClient.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        return nil, err
    }
    
    // 获取gas价格
    gasPrice, err := s.client.ethClient.SuggestGasPrice(context.Background())
    if err != nil {
        return nil, err
    }
    
    // 构建wipe交易数据
    data, err := s.buildWipeData(cdpID, daiAmount)
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, proxyAddr, big.NewInt(0), 250000, gasPrice, data)
    
    // 签名交易
    chainID, err := s.client.ethClient.NetworkID(context.Background())
    if err != nil {
        return nil, err
    }
    
    signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), s.privateKey)
    if err != nil {
        return nil, err
    }
    
    // 发送交易
    err = s.client.ethClient.SendTransaction(context.Background(), signedTx)
    if err != nil {
        return nil, err
    }
    
    return signedTx, nil
}

// 提取抵押品
func (s *CDPService) FreeCollateral(cdpID *big.Int, collateralAmount *big.Int) (*types.Transaction, error) {
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 获取用户的代理合约地址
    proxyAddr, err := s.getUserProxy(fromAddress)
    if err != nil {
        return nil, err
    }
    
    // 获取nonce
    nonce, err := s.client.ethClient.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        return nil, err
    }
    
    // 获取gas价格
    gasPrice, err := s.client.ethClient.SuggestGasPrice(context.Background())
    if err != nil {
        return nil, err
    }
    
    // 构建freeGem交易数据
    data, err := s.buildFreeGemData(cdpID, collateralAmount)
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, proxyAddr, big.NewInt(0), 250000, gasPrice, data)
    
    // 签名交易
    chainID, err := s.client.ethClient.NetworkID(context.Background())
    if err != nil {
        return nil, err
    }
    
    signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), s.privateKey)
    if err != nil {
        return nil, err
    }
    
    // 发送交易
    err = s.client.ethClient.SendTransaction(context.Background(), signedTx)
    if err != nil {
        return nil, err
    }
    
    return signedTx, nil
}
```

## DAI稳定币

### 4.1 DAI操作服务

```go
// services/dai_service.go
package services

import (
    "context"
    "crypto/ecdsa"
    "math/big"
    
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/shopspring/decimal"
)

type DAIService struct {
    client     *MakerClient
    privateKey *ecdsa.PrivateKey
}

func NewDAIService(client *MakerClient, privateKey *ecdsa.PrivateKey) *DAIService {
    return &DAIService{
        client:     client,
        privateKey: privateKey,
    }
}

// 获取DAI余额
func (s *DAIService) GetDAIBalance(address common.Address) (*big.Int, error) {
    // ERC20 balanceOf调用
    erc20ABI := `[{"constant":true,"inputs":[{"name":"account","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"type":"function"}]`
    
    parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
    if err != nil {
        return nil, err
    }
    
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    var balance *big.Int
    err = s.client.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &DAITokenAddress,
        Data: parsedABI.Methods["balanceOf"].ID,
    }, nil)
    
    if err != nil {
        return nil, err
    }
    
    return balance, nil
}

// 转账DAI
func (s *DAIService) TransferDAI(to common.Address, amount *big.Int) (*types.Transaction, error) {
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 获取nonce
    nonce, err := s.client.ethClient.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        return nil, err
    }
    
    // 获取gas价格
    gasPrice, err := s.client.ethClient.SuggestGasPrice(context.Background())
    if err != nil {
        return nil, err
    }
    
    // ERC20 transfer函数
    erc20ABI := `[{"constant":false,"inputs":[{"name":"to","type":"address"},{"name":"value","type":"uint256"}],"name":"transfer","outputs":[{"name":"","type":"bool"}],"type":"function"}]`
    
    parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
    if err != nil {
        return nil, err
    }
    
    // 构建transfer交易数据
    data, err := parsedABI.Pack("transfer", to, amount)
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, DAITokenAddress, big.NewInt(0), 100000, gasPrice, data)
    
    // 签名交易
    chainID, err := s.client.ethClient.NetworkID(context.Background())
    if err != nil {
        return nil, err
    }
    
    signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), s.privateKey)
    if err != nil {
        return nil, err
    }
    
    // 发送交易
    err = s.client.ethClient.SendTransaction(context.Background(), signedTx)
    if err != nil {
        return nil, err
    }
    
    return signedTx, nil
}

// 获取DAI总供应量
func (s *DAIService) GetDAITotalSupply() (*big.Int, error) {
    erc20ABI := `[{"constant":true,"inputs":[],"name":"totalSupply","outputs":[{"name":"","type":"uint256"}],"type":"function"}]`
    
    parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
    if err != nil {
        return nil, err
    }
    
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    var totalSupply *big.Int
    err = s.client.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &DAITokenAddress,
        Data: parsedABI.Methods["totalSupply"].ID,
    }, nil)
    
    if err != nil {
        return nil, err
    }
    
    return totalSupply, nil
}
```

## 储蓄利率

### 5.1 DSR服务

```go
// services/dsr_service.go
package services

import (
    "context"
    "crypto/ecdsa"
    "math/big"
    
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/shopspring/decimal"
)

type DSRService struct {
    client     *MakerClient
    privateKey *ecdsa.PrivateKey
}

func NewDSRService(client *MakerClient, privateKey *ecdsa.PrivateKey) *DSRService {
    return &DSRService{
        client:     client,
        privateKey: privateKey,
    }
}

// 存入DAI到DSR
func (s *DSRService) JoinDSR(daiAmount *big.Int) (*types.Transaction, error) {
    // 首先授权Pot合约使用DAI
    if err := s.approveToken(DAITokenAddress, s.client.potAddr, daiAmount); err != nil {
        return nil, err
    }
    
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 获取nonce
    nonce, err := s.client.ethClient.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        return nil, err
    }
    
    // 获取gas价格
    gasPrice, err := s.client.ethClient.SuggestGasPrice(context.Background())
    if err != nil {
        return nil, err
    }
    
    // 构建join交易数据
    data, err := s.client.potABI.Pack("join", daiAmount)
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, s.client.potAddr, big.NewInt(0), 150000, gasPrice, data)
    
    // 签名交易
    chainID, err := s.client.ethClient.NetworkID(context.Background())
    if err != nil {
        return nil, err
    }
    
    signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), s.privateKey)
    if err != nil {
        return nil, err
    }
    
    // 发送交易
    err = s.client.ethClient.SendTransaction(context.Background(), signedTx)
    if err != nil {
        return nil, err
    }
    
    return signedTx, nil
}

// 从DSR提取DAI
func (s *DSRService) ExitDSR(daiAmount *big.Int) (*types.Transaction, error) {
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 获取nonce
    nonce, err := s.client.ethClient.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        return nil, err
    }
    
    // 获取gas价格
    gasPrice, err := s.client.ethClient.SuggestGasPrice(context.Background())
    if err != nil {
        return nil, err
    }
    
    // 构建exit交易数据
    data, err := s.client.potABI.Pack("exit", daiAmount)
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, s.client.potAddr, big.NewInt(0), 120000, gasPrice, data)
    
    // 签名交易
    chainID, err := s.client.ethClient.NetworkID(context.Background())
    if err != nil {
        return nil, err
    }
    
    signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), s.privateKey)
    if err != nil {
        return nil, err
    }
    
    // 发送交易
    err = s.client.ethClient.SendTransaction(context.Background(), signedTx)
    if err != nil {
        return nil, err
    }
    
    return signedTx, nil
}

// 获取DSR利率
func (s *DSRService) GetDSR() (decimal.Decimal, error) {
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    var dsr *big.Int
    err := s.client.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &s.client.potAddr,
        Data: s.client.potABI.Methods["dsr"].ID,
    }, nil)
    
    if err != nil {
        return decimal.Zero, err
    }
    
    // 转换RAY精度到年化利率
    dsrDecimal := decimal.NewFromBigInt(dsr, -27) // RAY精度
    secondsPerYear := decimal.NewFromInt(31536000)
    
    // 计算年化利率: (dsr - 1) * secondsPerYear
    apy := dsrDecimal.Sub(decimal.NewFromInt(1)).Mul(secondsPerYear)
    
    return apy.Mul(decimal.NewFromInt(100)), nil // 转换为百分比
}

// 获取用户在DSR中的余额
func (s *DSRService) GetDSRBalance(user common.Address) (*big.Int, error) {
    callOpts := &bind.CallOpts{Context: context.Background()}
    
    var pie *big.Int
    err := s.client.ethClient.CallContract(context.Background(), ethereum.CallMsg{
        To: &s.client.potAddr,
        Data: s.client.potABI.Methods["pie"].ID,
    }, nil)
    
    if err != nil {
        return nil, err
    }
    
    return pie, nil
}

// 更新DSR累积利息
func (s *DSRService) DripDSR() (*types.Transaction, error) {
    // 获取发送者地址
    publicKey := s.privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 获取nonce
    nonce, err := s.client.ethClient.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        return nil, err
    }
    
    // 获取gas价格
    gasPrice, err := s.client.ethClient.SuggestGasPrice(context.Background())
    if err != nil {
        return nil, err
    }
    
    // 构建drip交易数据
    data, err := s.client.potABI.Pack("drip")
    if err != nil {
        return nil, err
    }
    
    tx := types.NewTransaction(nonce, s.client.potAddr, big.NewInt(0), 100000, gasPrice, data)
    
    // 签名交易
    chainID, err := s.client.ethClient.NetworkID(context.Background())
    if err != nil {
        return nil, err
    }
    
    signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), s.privateKey)
    if err != nil {
        return nil, err
    }
    
    // 发送交易
    err = s.client.ethClient.SendTransaction(context.Background(), signedTx)
    if err != nil {
        return nil, err
    }
    
    return signedTx, nil
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
    "github.com/ethereum/go-ethereum/crypto"
    
    "your-project/client"
    "your-project/services"
)

func main() {
    // 创建MakerDAO客户端
    makerClient, err := client.NewMakerClient("https://mainnet.infura.io/v3/YOUR_PROJECT_ID")
    if err != nil {
        log.Fatal("创建MakerDAO客户端失败:", err)
    }
    
    // 加载私钥
    privateKey, err := crypto.HexToECDSA("your_private_key_here")
    if err != nil {
        log.Fatal("加载私钥失败:", err)
    }
    
    // 获取用户地址
    userAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
    
    // 创建服务
    cdpService := services.NewCDPService(makerClient, privateKey)
    daiService := services.NewDAIService(makerClient, privateKey)
    dsrService := services.NewDSRService(makerClient, privateKey)
    
    // 1. 开启新的ETH-A CDP
    cdpID, err := cdpService.OpenCDP(ETH_A)
    if err != nil {
        log.Fatal("开启CDP失败:", err)
    }
    
    fmt.Printf("成功开启CDP，ID: %s\n", cdpID.String())
    
    // 2. 存入1 ETH作为抵押品
    ethAmount := big.NewInt(1e18) // 1 ETH
    ethAddress := common.HexToAddress("0x0000000000000000000000000000000000000000") // ETH
    
    tx, err := cdpService.LockCollateral(cdpID, ethAmount, ethAddress)
    if err != nil {
        log.Fatal("存入抵押品失败:", err)
    }
    
    fmt.Printf("存入抵押品交易已提交: %s\n", tx.Hash().Hex())
    
    // 3. 生成100 DAI
    daiAmount := big.NewInt(100e18) // 100 DAI
    tx, err = cdpService.DrawDAI(cdpID, daiAmount)
    if err != nil {
        log.Fatal("生成DAI失败:", err)
    }
    
    fmt.Printf("生成DAI交易已提交: %s\n", tx.Hash().Hex())
    
    // 4. 查询DAI余额
    daiBalance, err := daiService.GetDAIBalance(userAddress)
    if err != nil {
        log.Fatal("查询DAI余额失败:", err)
    }
    
    fmt.Printf("DAI余额: %s\n", daiBalance.String())
    
    // 5. 将50 DAI存入DSR赚取利息
    dsrAmount := big.NewInt(50e18) // 50 DAI
    tx, err = dsrService.JoinDSR(dsrAmount)
    if err != nil {
        log.Fatal("存入DSR失败:", err)
    }
    
    fmt.Printf("存入DSR交易已提交: %s\n", tx.Hash().Hex())
    
    // 6. 查询DSR利率
    dsrRate, err := dsrService.GetDSR()
    if err != nil {
        log.Fatal("查询DSR利率失败:", err)
    }
    
    fmt.Printf("当前DSR年化利率: %s%%\n", dsrRate.String())
    
    // 7. 查询CDP信息
    cdpInfo, err := makerClient.GetCDPInfo(cdpID)
    if err != nil {
        log.Fatal("查询CDP信息失败:", err)
    }
    
    fmt.Printf("CDP信息:\n")
    fmt.Printf("  抵押品: %s\n", cdpInfo.Collateral.String())
    fmt.Printf("  债务: %s\n", cdpInfo.Debt.String())
    
    // 8. 查询DAI总供应量
    totalSupply, err := daiService.GetDAITotalSupply()
    if err != nil {
        log.Fatal("查询DAI总供应量失败:", err)
    }
    
    fmt.Printf("DAI总供应量: %s\n", totalSupply.String())
}
```

这个MakerDAO使用指南提供了完整的去中心化稳定币协议集成方案，涵盖了CDP管理、DAI操作、DSR储蓄等核心功能，是DeFi稳定币系统的重要参考文档。
