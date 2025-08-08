# Cosmos SDK 详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [环境准备](#环境准备)
3. [应用链架构](#应用链架构)
4. [模块开发](#模块开发)
5. [状态管理](#状态管理)
6. [交易处理](#交易处理)
7. [共识和验证](#共识和验证)
8. [IBC跨链](#IBC跨链)
9. [实际应用](#实际应用)

## 基础概念

### 1.1 Cosmos SDK简介

Cosmos SDK是用于构建应用特定区块链的模块化框架，支持Tendermint共识算法和IBC跨链协议。

```bash
# 安装Cosmos SDK和相关工具
go install github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor@latest
go install github.com/cosmos/cosmos-sdk/tools/cosmovisor@latest

# 安装Ignite CLI (原Starport)
curl https://get.ignite.com/cli@v0.27.1! | bash
```

### 1.2 核心架构

```go
// 主要包导入
import (
    "github.com/cosmos/cosmos-sdk/baseapp"
    "github.com/cosmos/cosmos-sdk/client"
    "github.com/cosmos/cosmos-sdk/codec"
    "github.com/cosmos/cosmos-sdk/server"
    "github.com/cosmos/cosmos-sdk/types/module"
    "github.com/cosmos/cosmos-sdk/x/auth"
    "github.com/cosmos/cosmos-sdk/x/bank"
    "github.com/cosmos/cosmos-sdk/x/staking"
)
```

## 环境准备

### 2.1 项目初始化

```bash
# 使用Ignite CLI创建新链
ignite scaffold chain mychain --address-prefix mychain

# 项目结构
mychain/
├── app/           # 应用程序定义
├── cmd/           # CLI命令
├── proto/         # Protocol Buffers定义
├── x/             # 自定义模块
├── testutil/      # 测试工具
└── docs/          # 文档
```

### 2.2 基础配置

```go
// app/app.go
package app

import (
    "encoding/json"
    "io"
    "os"

    "github.com/cosmos/cosmos-sdk/baseapp"
    "github.com/cosmos/cosmos-sdk/client"
    "github.com/cosmos/cosmos-sdk/codec"
    "github.com/cosmos/cosmos-sdk/codec/types"
    "github.com/cosmos/cosmos-sdk/server/api"
    "github.com/cosmos/cosmos-sdk/server/config"
    servertypes "github.com/cosmos/cosmos-sdk/server/types"
    "github.com/cosmos/cosmos-sdk/store/streaming"
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/cosmos/cosmos-sdk/types/module"
    "github.com/cosmos/cosmos-sdk/version"
    "github.com/cosmos/cosmos-sdk/x/auth"
    "github.com/cosmos/cosmos-sdk/x/bank"
    "github.com/cosmos/cosmos-sdk/x/staking"
)

const (
    AccountAddressPrefix = "mychain"
    Name                 = "mychain"
)

// App 应用程序结构
type App struct {
    *baseapp.BaseApp

    cdc               *codec.LegacyAmino
    appCodec          codec.Codec
    interfaceRegistry types.InterfaceRegistry

    // 模块管理器
    mm *module.Manager

    // 模块keeper
    AccountKeeper auth.AccountKeeper
    BankKeeper    bank.Keeper
    StakingKeeper staking.Keeper

    // 其他组件
    scopedIBCKeeper      capabilitykeeper.ScopedKeeper
    scopedTransferKeeper capabilitykeeper.ScopedKeeper
}

// NewApp 创建新的应用程序实例
func NewApp(
    logger log.Logger,
    db dbm.DB,
    traceStore io.Writer,
    loadLatest bool,
    skipUpgradeHeights map[int64]bool,
    homePath string,
    invCheckPeriod uint,
    encodingConfig appparams.EncodingConfig,
    appOpts servertypes.AppOptions,
    baseAppOptions ...func(*baseapp.BaseApp),
) *App {
    
    appCodec := encodingConfig.Marshaler
    cdc := encodingConfig.Amino
    interfaceRegistry := encodingConfig.InterfaceRegistry

    bApp := baseapp.NewBaseApp(Name, logger, db, encodingConfig.TxConfig.TxDecoder(), baseAppOptions...)
    bApp.SetCommitMultiStoreTracer(traceStore)
    bApp.SetVersion(version.Version)
    bApp.SetInterfaceRegistry(interfaceRegistry)

    app := &App{
        BaseApp:           bApp,
        cdc:               cdc,
        appCodec:          appCodec,
        interfaceRegistry: interfaceRegistry,
    }

    return app
}
```

## 应用链架构

### 3.1 模块管理

```go
// app/modules.go
package app

import (
    "github.com/cosmos/cosmos-sdk/types/module"
    "github.com/cosmos/cosmos-sdk/x/auth"
    "github.com/cosmos/cosmos-sdk/x/bank"
    "github.com/cosmos/cosmos-sdk/x/staking"
    "github.com/cosmos/cosmos-sdk/x/gov"
    "github.com/cosmos/cosmos-sdk/x/mint"
    "github.com/cosmos/cosmos-sdk/x/distribution"
)

// ModuleBasics 定义应用程序的基础模块管理器
var ModuleBasics = module.NewBasicManager(
    auth.AppModuleBasic{},
    bank.AppModuleBasic{},
    staking.AppModuleBasic{},
    mint.AppModuleBasic{},
    distribution.AppModuleBasic{},
    gov.NewAppModuleBasic(
        paramsclient.ProposalHandler,
        distrclient.ProposalHandler,
        upgradeclient.ProposalHandler,
        upgradeclient.CancelProposalHandler,
    ),
)

// 模块账户权限
func maccPerms() map[string][]string {
    return map[string][]string{
        auth.FeeCollectorName:     nil,
        distribution.ModuleName:   nil,
        mint.ModuleName:           {auth.Minter},
        staking.BondedPoolName:    {auth.Burner, auth.Staking},
        staking.NotBondedPoolName: {auth.Burner, auth.Staking},
        gov.ModuleName:            {auth.Burner},
    }
}

// 初始化模块管理器
func (app *App) setupModuleManager() {
    app.mm = module.NewManager(
        auth.NewAppModule(app.appCodec, app.AccountKeeper, nil),
        bank.NewAppModule(app.appCodec, app.BankKeeper, app.AccountKeeper),
        staking.NewAppModule(app.appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper),
        mint.NewAppModule(app.appCodec, app.MintKeeper, app.AccountKeeper),
        distribution.NewAppModule(app.appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper),
        gov.NewAppModule(app.appCodec, app.GovKeeper, app.AccountKeeper, app.BankKeeper),
    )

    // 设置模块顺序
    app.mm.SetOrderBeginBlockers(
        mint.ModuleName,
        distribution.ModuleName,
        staking.ModuleName,
    )

    app.mm.SetOrderEndBlockers(
        staking.ModuleName,
        gov.ModuleName,
    )

    app.mm.SetOrderInitGenesis(
        auth.ModuleName,
        bank.ModuleName,
        distribution.ModuleName,
        staking.ModuleName,
        mint.ModuleName,
        gov.ModuleName,
    )
}
```

## 模块开发

### 4.1 自定义模块结构

```go
// x/blog/types/keys.go
package types

const (
    // ModuleName 模块名称
    ModuleName = "blog"

    // StoreKey 存储键
    StoreKey = ModuleName

    // RouterKey 路由键
    RouterKey = ModuleName

    // QuerierRoute 查询路由
    QuerierRoute = ModuleName

    // MemStoreKey 内存存储键
    MemStoreKey = "mem_blog"
)

// 存储键前缀
var (
    PostKey = []byte{0x01}
)

func GetPostKey(id uint64) []byte {
    return append(PostKey, sdk.Uint64ToBigEndian(id)...)
}
```

### 4.2 消息类型定义

```go
// x/blog/types/messages.go
package types

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// 博客文章结构
type Post struct {
    Id      uint64 `json:"id"`
    Title   string `json:"title"`
    Content string `json:"content"`
    Author  string `json:"author"`
}

// 创建文章消息
type MsgCreatePost struct {
    Creator string `json:"creator"`
    Title   string `json:"title"`
    Content string `json:"content"`
}

// 实现sdk.Msg接口
func (msg MsgCreatePost) Route() string { return RouterKey }
func (msg MsgCreatePost) Type() string  { return "create_post" }

func (msg MsgCreatePost) GetSigners() []sdk.AccAddress {
    creator, err := sdk.AccAddressFromBech32(msg.Creator)
    if err != nil {
        panic(err)
    }
    return []sdk.AccAddress{creator}
}

func (msg MsgCreatePost) GetSignBytes() []byte {
    bz := ModuleCdc.MustMarshalJSON(&msg)
    return sdk.MustSortJSON(bz)
}

func (msg MsgCreatePost) ValidateBasic() error {
    _, err := sdk.AccAddressFromBech32(msg.Creator)
    if err != nil {
        return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
    }

    if len(msg.Title) == 0 {
        return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "title cannot be empty")
    }

    if len(msg.Content) == 0 {
        return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "content cannot be empty")
    }

    return nil
}
```

### 4.3 Keeper实现

```go
// x/blog/keeper/keeper.go
package keeper

import (
    "encoding/binary"

    "github.com/cosmos/cosmos-sdk/codec"
    "github.com/cosmos/cosmos-sdk/store/prefix"
    sdk "github.com/cosmos/cosmos-sdk/types"
    
    "mychain/x/blog/types"
)

type Keeper struct {
    cdc      codec.BinaryCodec
    storeKey sdk.StoreKey
    memKey   sdk.StoreKey
}

func NewKeeper(cdc codec.BinaryCodec, storeKey, memKey sdk.StoreKey) *Keeper {
    return &Keeper{
        cdc:      cdc,
        storeKey: storeKey,
        memKey:   memKey,
    }
}

// 创建文章
func (k Keeper) CreatePost(ctx sdk.Context, post types.Post) uint64 {
    store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PostKey)
    
    // 获取下一个ID
    id := k.GetNextPostID(ctx)
    post.Id = id

    // 序列化并存储
    bz := k.cdc.MustMarshal(&post)
    store.Set(types.GetPostKey(id), bz)

    // 更新计数器
    k.SetNextPostID(ctx, id+1)

    return id
}

// 获取文章
func (k Keeper) GetPost(ctx sdk.Context, id uint64) (types.Post, bool) {
    store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PostKey)
    
    bz := store.Get(types.GetPostKey(id))
    if bz == nil {
        return types.Post{}, false
    }

    var post types.Post
    k.cdc.MustUnmarshal(bz, &post)
    return post, true
}

// 获取所有文章
func (k Keeper) GetAllPosts(ctx sdk.Context) []types.Post {
    store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PostKey)
    
    iterator := sdk.KVStorePrefixIterator(store, []byte{})
    defer iterator.Close()

    var posts []types.Post
    for ; iterator.Valid(); iterator.Next() {
        var post types.Post
        k.cdc.MustUnmarshal(iterator.Value(), &post)
        posts = append(posts, post)
    }

    return posts
}

// 获取下一个文章ID
func (k Keeper) GetNextPostID(ctx sdk.Context) uint64 {
    store := ctx.KVStore(k.storeKey)
    bz := store.Get([]byte("next_post_id"))
    if bz == nil {
        return 1
    }
    return binary.BigEndian.Uint64(bz)
}

// 设置下一个文章ID
func (k Keeper) SetNextPostID(ctx sdk.Context, id uint64) {
    store := ctx.KVStore(k.storeKey)
    bz := make([]byte, 8)
    binary.BigEndian.PutUint64(bz, id)
    store.Set([]byte("next_post_id"), bz)
}
```

## 状态管理

### 5.1 状态存储

```go
// x/blog/keeper/grpc_query.go
package keeper

import (
    "context"

    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"

    "github.com/cosmos/cosmos-sdk/store/prefix"
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/cosmos/cosmos-sdk/types/query"
    
    "mychain/x/blog/types"
)

// 查询服务器
type queryServer struct {
    Keeper
}

func NewQueryServerImpl(keeper Keeper) types.QueryServer {
    return &queryServer{Keeper: keeper}
}

// 查询文章
func (k queryServer) Post(c context.Context, req *types.QueryGetPostRequest) (*types.QueryGetPostResponse, error) {
    if req == nil {
        return nil, status.Error(codes.InvalidArgument, "invalid request")
    }

    ctx := sdk.UnwrapSDKContext(c)
    post, found := k.GetPost(ctx, req.Id)
    if !found {
        return nil, status.Error(codes.NotFound, "post not found")
    }

    return &types.QueryGetPostResponse{Post: post}, nil
}

// 查询所有文章（分页）
func (k queryServer) PostAll(c context.Context, req *types.QueryAllPostRequest) (*types.QueryAllPostResponse, error) {
    if req == nil {
        return nil, status.Error(codes.InvalidArgument, "invalid request")
    }

    ctx := sdk.UnwrapSDKContext(c)
    store := ctx.KVStore(k.storeKey)
    postStore := prefix.NewStore(store, types.PostKey)

    var posts []types.Post
    pageRes, err := query.Paginate(postStore, req.Pagination, func(key []byte, value []byte) error {
        var post types.Post
        if err := k.cdc.Unmarshal(value, &post); err != nil {
            return err
        }
        posts = append(posts, post)
        return nil
    })

    if err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }

    return &types.QueryAllPostResponse{Post: posts, Pagination: pageRes}, nil
}
```

## 交易处理

### 6.1 消息处理器

```go
// x/blog/keeper/msg_server.go
package keeper

import (
    "context"

    sdk "github.com/cosmos/cosmos-sdk/types"
    "mychain/x/blog/types"
)

type msgServer struct {
    Keeper
}

func NewMsgServerImpl(keeper Keeper) types.MsgServer {
    return &msgServer{Keeper: keeper}
}

// 处理创建文章消息
func (k msgServer) CreatePost(goCtx context.Context, msg *types.MsgCreatePost) (*types.MsgCreatePostResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)

    // 验证消息
    if err := msg.ValidateBasic(); err != nil {
        return nil, err
    }

    // 创建文章对象
    post := types.Post{
        Title:   msg.Title,
        Content: msg.Content,
        Author:  msg.Creator,
    }

    // 存储文章
    id := k.CreatePost(ctx, post)

    // 发出事件
    ctx.EventManager().EmitEvent(
        sdk.NewEvent(
            types.EventTypeCreatePost,
            sdk.NewAttribute(types.AttributeKeyPostID, fmt.Sprintf("%d", id)),
            sdk.NewAttribute(types.AttributeKeyAuthor, msg.Creator),
        ),
    )

    return &types.MsgCreatePostResponse{Id: id}, nil
}

// 处理更新文章消息
func (k msgServer) UpdatePost(goCtx context.Context, msg *types.MsgUpdatePost) (*types.MsgUpdatePostResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)

    // 获取现有文章
    post, found := k.GetPost(ctx, msg.Id)
    if !found {
        return nil, types.ErrPostNotFound
    }

    // 验证权限
    if post.Author != msg.Creator {
        return nil, types.ErrUnauthorized
    }

    // 更新文章
    post.Title = msg.Title
    post.Content = msg.Content

    // 保存更新
    k.SetPost(ctx, post)

    // 发出事件
    ctx.EventManager().EmitEvent(
        sdk.NewEvent(
            types.EventTypeUpdatePost,
            sdk.NewAttribute(types.AttributeKeyPostID, fmt.Sprintf("%d", msg.Id)),
            sdk.NewAttribute(types.AttributeKeyAuthor, msg.Creator),
        ),
    )

    return &types.MsgUpdatePostResponse{}, nil
}
```

## 共识和验证

### 7.1 验证器管理

```go
// x/blog/keeper/validator.go
package keeper

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// 验证器相关操作
func (k Keeper) GetValidator(ctx sdk.Context, addr sdk.ValAddress) (validator types.Validator, found bool) {
    store := ctx.KVStore(k.storeKey)
    value := store.Get(types.GetValidatorKey(addr))
    if value == nil {
        return validator, false
    }

    validator = types.MustUnmarshalValidator(k.cdc, value)
    return validator, true
}

// 设置验证器
func (k Keeper) SetValidator(ctx sdk.Context, validator types.Validator) {
    store := ctx.KVStore(k.storeKey)
    bz := types.MustMarshalValidator(k.cdc, &validator)
    store.Set(types.GetValidatorKey(validator.GetOperator()), bz)
}

// 委托操作
func (k Keeper) Delegate(ctx sdk.Context, delAddr sdk.AccAddress, bondAmt sdk.Int, tokenSrc types.BondStatus, validator types.Validator, subtractAccount bool) (newShares sdk.Dec, err error) {
    // 验证委托金额
    if bondAmt.IsNegative() {
        return sdk.ZeroDec(), types.ErrBadDelegationAmount
    }

    // 获取或创建委托记录
    delegation, found := k.GetDelegation(ctx, delAddr, validator.GetOperator())
    if !found {
        delegation = types.NewDelegation(delAddr, validator.GetOperator(), sdk.ZeroDec())
    }

    // 计算新的份额
    newShares = validator.AddTokensFromDel(bondAmt)
    delegation.Shares = delegation.Shares.Add(newShares)

    // 更新验证器和委托记录
    k.SetValidator(ctx, validator)
    k.SetDelegation(ctx, delegation)

    return newShares, nil
}
```

## IBC跨链

### 8.1 IBC模块集成

```go
// x/blog/ibc.go
package blog

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
    porttypes "github.com/cosmos/ibc-go/v3/modules/core/05-port/types"
    host "github.com/cosmos/ibc-go/v3/modules/core/24-host"
    ibcexported "github.com/cosmos/ibc-go/v3/modules/core/exported"
)

// IBC包处理
func (am AppModule) OnRecvPacket(
    ctx sdk.Context,
    packet channeltypes.Packet,
    relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
    
    var data types.BlogPacketData
    if err := types.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
        return channeltypes.NewErrorAcknowledgement(err.Error())
    }

    // 根据包类型处理
    switch packet := data.Packet.(type) {
    case *types.BlogPacketData_CreatePostPacket:
        return am.keeper.OnRecvCreatePostPacket(ctx, packet.CreatePostPacket)
    default:
        return channeltypes.NewErrorAcknowledgement("unknown packet type")
    }
}

// 处理跨链创建文章包
func (k Keeper) OnRecvCreatePostPacket(ctx sdk.Context, packet types.CreatePostPacketData) channeltypes.Acknowledgement {
    // 创建文章
    post := types.Post{
        Title:   packet.Title,
        Content: packet.Content,
        Author:  packet.Author,
    }

    id := k.CreatePost(ctx, post)

    // 返回确认
    return channeltypes.NewResultAcknowledgement(
        types.ModuleCdc.MustMarshalJSON(&types.CreatePostPacketAck{
            PostId: id,
        }),
    )
}

// 发送跨链包
func (k Keeper) TransmitCreatePostPacket(
    ctx sdk.Context,
    packetData types.CreatePostPacketData,
    sourcePort,
    sourceChannel string,
    timeoutHeight clienttypes.Height,
    timeoutTimestamp uint64,
) error {
    
    sourceChannelEnd, found := k.ChannelKeeper.GetChannel(ctx, sourcePort, sourceChannel)
    if !found {
        return sdkerrors.Wrapf(channeltypes.ErrChannelNotFound, "port ID (%s) channel ID (%s)", sourcePort, sourceChannel)
    }

    destinationPort := sourceChannelEnd.GetCounterparty().GetPortID()
    destinationChannel := sourceChannelEnd.GetCounterparty().GetChannelID()

    // 创建包数据
    data := types.BlogPacketData{
        Packet: &types.BlogPacketData_CreatePostPacket{
            CreatePostPacket: &packetData,
        },
    }

    packet := channeltypes.NewPacket(
        data.GetBytes(),
        k.ChannelKeeper.GetNextSequenceSend(ctx, sourcePort, sourceChannel),
        sourcePort,
        sourceChannel,
        destinationPort,
        destinationChannel,
        timeoutHeight,
        timeoutTimestamp,
    )

    return k.ChannelKeeper.SendPacket(ctx, packet)
}
```

## 实际应用

### 9.1 完整应用示例

```go
// cmd/mychaind/main.go
package main

import (
    "os"

    "github.com/cosmos/cosmos-sdk/server"
    svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
    
    "mychain/app"
    "mychain/cmd/mychaind/cmd"
)

func main() {
    rootCmd, _ := cmd.NewRootCmd()

    if err := svrcmd.Execute(rootCmd, app.DefaultNodeHome); err != nil {
        switch e := err.(type) {
        case server.ErrorCode:
            os.Exit(e.Code)
        default:
            os.Exit(1)
        }
    }
}

// 启动命令示例
// mychaind init mynode --chain-id mychain-1
// mychaind keys add alice
// mychaind add-genesis-account alice 100000000stake
// mychaind gentx alice 1000000stake --chain-id mychain-1
// mychaind collect-gentxs
// mychaind start
```

### 9.2 客户端交互

```go
// client/tx.go
package client

import (
    "context"
    
    "github.com/cosmos/cosmos-sdk/client"
    "github.com/cosmos/cosmos-sdk/client/tx"
    sdk "github.com/cosmos/cosmos-sdk/types"
    
    "mychain/x/blog/types"
)

// 创建文章交易
func CreatePostTx(clientCtx client.Context, from string, title, content string) error {
    msg := &types.MsgCreatePost{
        Creator: from,
        Title:   title,
        Content: content,
    }

    if err := msg.ValidateBasic(); err != nil {
        return err
    }

    return tx.GenerateOrBroadcastTxCLI(clientCtx, clientCtx.GetFromAddress(), msg)
}

// 查询文章
func QueryPost(clientCtx client.Context, id uint64) (*types.Post, error) {
    queryClient := types.NewQueryClient(clientCtx)
    
    res, err := queryClient.Post(context.Background(), &types.QueryGetPostRequest{
        Id: id,
    })
    if err != nil {
        return nil, err
    }

    return &res.Post, nil
}
```
