# gRPC框架详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [环境准备](#环境准备)
3. [Protocol Buffers](#protocol-buffers)
4. [服务端开发](#服务端开发)
5. [客户端开发](#客户端开发)
6. [流式处理](#流式处理)
7. [中间件和拦截器](#中间件和拦截器)
8. [错误处理](#错误处理)
9. [高级特性](#高级特性)

## 基础概念

### 1.1 gRPC简介

gRPC是高性能、跨语言的RPC框架，基于HTTP/2协议，使用Protocol Buffers作为接口描述语言。

```bash
# 安装gRPC和相关工具
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 安装protoc编译器
# macOS: brew install protobuf
# Ubuntu: apt install protobuf-compiler
# Windows: 下载预编译二进制文件
```

### 1.2 核心概念

```go
// gRPC的四种服务类型：
// 1. 一元RPC (Unary RPC)
// 2. 服务端流式RPC (Server Streaming RPC)
// 3. 客户端流式RPC (Client Streaming RPC)
// 4. 双向流式RPC (Bidirectional Streaming RPC)
```

## 环境准备

### 2.1 项目结构

```
project/
├── proto/
│   └── user.proto
├── server/
│   └── main.go
├── client/
│   └── main.go
├── pb/
│   ├── user.pb.go
│   └── user_grpc.pb.go
└── go.mod
```

### 2.2 依赖安装

```go
// go.mod
module grpc-example

go 1.21

require (
    google.golang.org/grpc v1.58.0
    google.golang.org/protobuf v1.31.0
)
```

## Protocol Buffers

### 3.1 定义服务

```protobuf
// proto/user.proto
syntax = "proto3";

package user;
option go_package = "./pb";

// 用户信息
message User {
  int32 id = 1;
  string name = 2;
  string email = 3;
  int32 age = 4;
  repeated string tags = 5;
}

// 创建用户请求
message CreateUserRequest {
  string name = 1;
  string email = 2;
  int32 age = 3;
}

// 创建用户响应
message CreateUserResponse {
  User user = 1;
  bool success = 2;
  string message = 3;
}

// 获取用户请求
message GetUserRequest {
  int32 id = 1;
}

// 获取用户响应
message GetUserResponse {
  User user = 1;
}

// 用户列表请求
message ListUsersRequest {
  int32 page = 1;
  int32 page_size = 2;
}

// 用户列表响应
message ListUsersResponse {
  repeated User users = 1;
  int32 total = 2;
}

// 用户服务定义
service UserService {
  // 一元RPC：创建用户
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
  
  // 一元RPC：获取用户
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  
  // 服务端流式RPC：获取用户列表
  rpc ListUsers(ListUsersRequest) returns (stream User);
  
  // 客户端流式RPC：批量创建用户
  rpc BatchCreateUsers(stream CreateUserRequest) returns (CreateUserResponse);
  
  // 双向流式RPC：实时聊天
  rpc Chat(stream ChatMessage) returns (stream ChatMessage);
}

// 聊天消息
message ChatMessage {
  int32 user_id = 1;
  string content = 2;
  int64 timestamp = 3;
}
```

### 3.2 生成Go代码

```bash
# 生成Protocol Buffers代码
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/user.proto
```

## 服务端开发

### 4.1 实现服务

```go
// server/main.go
package main

import (
    "context"
    "fmt"
    "io"
    "log"
    "net"
    "sync"
    "time"

    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    
    pb "grpc-example/pb"
)

type userServer struct {
    pb.UnimplementedUserServiceServer
    users map[int32]*pb.User
    mutex sync.RWMutex
    nextID int32
}

func newUserServer() *userServer {
    return &userServer{
        users:  make(map[int32]*pb.User),
        nextID: 1,
    }
}

// 一元RPC：创建用户
func (s *userServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
    // 输入验证
    if req.Name == "" {
        return nil, status.Error(codes.InvalidArgument, "用户名不能为空")
    }
    if req.Email == "" {
        return nil, status.Error(codes.InvalidArgument, "邮箱不能为空")
    }

    s.mutex.Lock()
    defer s.mutex.Unlock()

    // 检查邮箱是否已存在
    for _, user := range s.users {
        if user.Email == req.Email {
            return &pb.CreateUserResponse{
                Success: false,
                Message: "邮箱已存在",
            }, nil
        }
    }

    // 创建新用户
    user := &pb.User{
        Id:    s.nextID,
        Name:  req.Name,
        Email: req.Email,
        Age:   req.Age,
    }
    s.users[s.nextID] = user
    s.nextID++

    log.Printf("创建用户: %v", user)

    return &pb.CreateUserResponse{
        User:    user,
        Success: true,
        Message: "用户创建成功",
    }, nil
}

// 一元RPC：获取用户
func (s *userServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
    s.mutex.RLock()
    defer s.mutex.RUnlock()

    user, exists := s.users[req.Id]
    if !exists {
        return nil, status.Error(codes.NotFound, "用户不存在")
    }

    return &pb.GetUserResponse{
        User: user,
    }, nil
}

// 服务端流式RPC：获取用户列表
func (s *userServer) ListUsers(req *pb.ListUsersRequest, stream pb.UserService_ListUsersServer) error {
    s.mutex.RLock()
    defer s.mutex.RUnlock()

    pageSize := req.PageSize
    if pageSize <= 0 {
        pageSize = 10
    }

    page := req.Page
    if page <= 0 {
        page = 1
    }

    start := (page - 1) * pageSize
    count := int32(0)

    for _, user := range s.users {
        if count < start {
            count++
            continue
        }
        if count >= start+pageSize {
            break
        }

        if err := stream.Send(user); err != nil {
            return err
        }
        count++

        // 模拟延迟
        time.Sleep(100 * time.Millisecond)
    }

    return nil
}

// 客户端流式RPC：批量创建用户
func (s *userServer) BatchCreateUsers(stream pb.UserService_BatchCreateUsersServer) error {
    var createdCount int32

    for {
        req, err := stream.Recv()
        if err == io.EOF {
            // 客户端完成发送
            return stream.SendAndClose(&pb.CreateUserResponse{
                Success: true,
                Message: fmt.Sprintf("批量创建了 %d 个用户", createdCount),
            })
        }
        if err != nil {
            return err
        }

        // 创建用户逻辑
        s.mutex.Lock()
        user := &pb.User{
            Id:    s.nextID,
            Name:  req.Name,
            Email: req.Email,
            Age:   req.Age,
        }
        s.users[s.nextID] = user
        s.nextID++
        s.mutex.Unlock()

        createdCount++
        log.Printf("批量创建用户: %v", user)
    }
}

// 双向流式RPC：实时聊天
func (s *userServer) Chat(stream pb.UserService_ChatServer) error {
    for {
        msg, err := stream.Recv()
        if err == io.EOF {
            return nil
        }
        if err != nil {
            return err
        }

        log.Printf("收到消息: 用户%d: %s", msg.UserId, msg.Content)

        // 回显消息（实际应用中可能是广播给其他用户）
        response := &pb.ChatMessage{
            UserId:    0, // 系统消息
            Content:   fmt.Sprintf("收到来自用户%d的消息: %s", msg.UserId, msg.Content),
            Timestamp: time.Now().Unix(),
        }

        if err := stream.Send(response); err != nil {
            return err
        }
    }
}

func main() {
    // 创建监听器
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("监听失败: %v", err)
    }

    // 创建gRPC服务器
    s := grpc.NewServer()

    // 注册服务
    pb.RegisterUserServiceServer(s, newUserServer())

    log.Println("gRPC服务器启动在 :50051")
    if err := s.Serve(lis); err != nil {
        log.Fatalf("服务启动失败: %v", err)
    }
}
```

## 客户端开发

### 5.1 基本客户端

```go
// client/main.go
package main

import (
    "context"
    "io"
    "log"
    "time"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    
    pb "grpc-example/pb"
)

func main() {
    // 连接到服务器
    conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("连接失败: %v", err)
    }
    defer conn.Close()

    // 创建客户端
    client := pb.NewUserServiceClient(conn)

    // 测试一元RPC
    testUnaryRPC(client)

    // 测试服务端流式RPC
    testServerStreamingRPC(client)

    // 测试客户端流式RPC
    testClientStreamingRPC(client)

    // 测试双向流式RPC
    testBidirectionalStreamingRPC(client)
}

// 测试一元RPC
func testUnaryRPC(client pb.UserServiceClient) {
    log.Println("=== 测试一元RPC ===")

    // 创建用户
    createReq := &pb.CreateUserRequest{
        Name:  "张三",
        Email: "zhangsan@example.com",
        Age:   25,
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    createResp, err := client.CreateUser(ctx, createReq)
    if err != nil {
        log.Printf("创建用户失败: %v", err)
        return
    }

    log.Printf("创建用户响应: %v", createResp)

    // 获取用户
    getReq := &pb.GetUserRequest{Id: createResp.User.Id}
    getResp, err := client.GetUser(ctx, getReq)
    if err != nil {
        log.Printf("获取用户失败: %v", err)
        return
    }

    log.Printf("获取用户响应: %v", getResp.User)
}

// 测试服务端流式RPC
func testServerStreamingRPC(client pb.UserServiceClient) {
    log.Println("=== 测试服务端流式RPC ===")

    req := &pb.ListUsersRequest{
        Page:     1,
        PageSize: 5,
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    stream, err := client.ListUsers(ctx, req)
    if err != nil {
        log.Printf("获取用户列表失败: %v", err)
        return
    }

    for {
        user, err := stream.Recv()
        if err == io.EOF {
            break
        }
        if err != nil {
            log.Printf("接收用户数据失败: %v", err)
            break
        }
        log.Printf("接收到用户: %v", user)
    }
}

// 测试客户端流式RPC
func testClientStreamingRPC(client pb.UserServiceClient) {
    log.Println("=== 测试客户端流式RPC ===")

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    stream, err := client.BatchCreateUsers(ctx)
    if err != nil {
        log.Printf("批量创建用户失败: %v", err)
        return
    }

    // 发送多个用户创建请求
    users := []*pb.CreateUserRequest{
        {Name: "李四", Email: "lisi@example.com", Age: 30},
        {Name: "王五", Email: "wangwu@example.com", Age: 28},
        {Name: "赵六", Email: "zhaoliu@example.com", Age: 35},
    }

    for _, user := range users {
        if err := stream.Send(user); err != nil {
            log.Printf("发送用户数据失败: %v", err)
            return
        }
        log.Printf("发送用户: %v", user)
    }

    // 关闭发送并接收响应
    resp, err := stream.CloseAndRecv()
    if err != nil {
        log.Printf("接收批量创建响应失败: %v", err)
        return
    }

    log.Printf("批量创建响应: %v", resp)
}

// 测试双向流式RPC
func testBidirectionalStreamingRPC(client pb.UserServiceClient) {
    log.Println("=== 测试双向流式RPC ===")

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    stream, err := client.Chat(ctx)
    if err != nil {
        log.Printf("创建聊天流失败: %v", err)
        return
    }

    // 启动接收goroutine
    go func() {
        for {
            msg, err := stream.Recv()
            if err == io.EOF {
                return
            }
            if err != nil {
                log.Printf("接收消息失败: %v", err)
                return
            }
            log.Printf("收到消息: %v", msg)
        }
    }()

    // 发送消息
    messages := []string{"你好", "这是测试消息", "再见"}
    for i, content := range messages {
        msg := &pb.ChatMessage{
            UserId:    int32(i + 1),
            Content:   content,
            Timestamp: time.Now().Unix(),
        }

        if err := stream.Send(msg); err != nil {
            log.Printf("发送消息失败: %v", err)
            return
        }

        time.Sleep(1 * time.Second)
    }

    // 关闭发送
    if err := stream.CloseSend(); err != nil {
        log.Printf("关闭发送失败: %v", err)
    }

    time.Sleep(2 * time.Second) // 等待接收完成
}
```
