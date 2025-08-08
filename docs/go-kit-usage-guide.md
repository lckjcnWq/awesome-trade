# Go-kit微服务框架详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [服务定义](#服务定义)
3. [传输层](#传输层)
4. [中间件](#中间件)
5. [服务发现](#服务发现)
6. [负载均衡](#负载均衡)
7. [熔断器](#熔断器)
8. [监控和日志](#监控和日志)
9. [完整示例](#完整示例)

## 基础概念

### 1.1 Go-kit简介

Go-kit是微服务工具包，架构清晰，微服务工具包，提供了构建大型分布式系统所需的组件。

```bash
# 安装Go-kit
go get github.com/go-kit/kit
go get github.com/go-kit/log
```

### 1.2 核心概念

```go
// Go-kit的三层架构：
// 1. Transport Layer (传输层) - HTTP, gRPC, Thrift等
// 2. Endpoint Layer (端点层) - 业务逻辑的抽象
// 3. Service Layer (服务层) - 业务逻辑实现
```

## 服务定义

### 2.1 服务接口

```go
// service/user.go
package service

import (
    "context"
    "errors"
)

// 用户信息
type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
    Age   int    `json:"age"`
}

// 用户服务接口
type UserService interface {
    CreateUser(ctx context.Context, name, email string, age int) (User, error)
    GetUser(ctx context.Context, id int) (User, error)
    UpdateUser(ctx context.Context, id int, name, email string, age int) (User, error)
    DeleteUser(ctx context.Context, id int) error
    ListUsers(ctx context.Context, page, limit int) ([]User, error)
}

// 错误定义
var (
    ErrUserNotFound    = errors.New("用户不存在")
    ErrInvalidArgument = errors.New("无效参数")
    ErrUserExists      = errors.New("用户已存在")
)
```

### 2.2 服务实现

```go
// service/user_impl.go
package service

import (
    "context"
    "sync"
)

type userService struct {
    users  map[int]User
    nextID int
    mutex  sync.RWMutex
}

// 创建用户服务实例
func NewUserService() UserService {
    return &userService{
        users:  make(map[int]User),
        nextID: 1,
    }
}

func (s *userService) CreateUser(ctx context.Context, name, email string, age int) (User, error) {
    if name == "" || email == "" {
        return User{}, ErrInvalidArgument
    }

    s.mutex.Lock()
    defer s.mutex.Unlock()

    // 检查邮箱是否已存在
    for _, user := range s.users {
        if user.Email == email {
            return User{}, ErrUserExists
        }
    }

    user := User{
        ID:    s.nextID,
        Name:  name,
        Email: email,
        Age:   age,
    }

    s.users[s.nextID] = user
    s.nextID++

    return user, nil
}

func (s *userService) GetUser(ctx context.Context, id int) (User, error) {
    s.mutex.RLock()
    defer s.mutex.RUnlock()

    user, exists := s.users[id]
    if !exists {
        return User{}, ErrUserNotFound
    }

    return user, nil
}

func (s *userService) UpdateUser(ctx context.Context, id int, name, email string, age int) (User, error) {
    s.mutex.Lock()
    defer s.mutex.Unlock()

    user, exists := s.users[id]
    if !exists {
        return User{}, ErrUserNotFound
    }

    if name != "" {
        user.Name = name
    }
    if email != "" {
        user.Email = email
    }
    if age > 0 {
        user.Age = age
    }

    s.users[id] = user
    return user, nil
}

func (s *userService) DeleteUser(ctx context.Context, id int) error {
    s.mutex.Lock()
    defer s.mutex.Unlock()

    if _, exists := s.users[id]; !exists {
        return ErrUserNotFound
    }

    delete(s.users, id)
    return nil
}

func (s *userService) ListUsers(ctx context.Context, page, limit int) ([]User, error) {
    s.mutex.RLock()
    defer s.mutex.RUnlock()

    if page <= 0 {
        page = 1
    }
    if limit <= 0 {
        limit = 10
    }

    start := (page - 1) * limit
    end := start + limit

    var users []User
    i := 0
    for _, user := range s.users {
        if i >= start && i < end {
            users = append(users, user)
        }
        i++
        if i >= end {
            break
        }
    }

    return users, nil
}
```

## 传输层

### 3.1 端点定义

```go
// endpoint/user.go
package endpoint

import (
    "context"
    
    "github.com/go-kit/kit/endpoint"
    "your-project/service"
)

// 请求和响应结构体
type CreateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
    Age   int    `json:"age"`
}

type CreateUserResponse struct {
    User service.User `json:"user,omitempty"`
    Err  string       `json:"error,omitempty"`
}

type GetUserRequest struct {
    ID int `json:"id"`
}

type GetUserResponse struct {
    User service.User `json:"user,omitempty"`
    Err  string       `json:"error,omitempty"`
}

// 端点集合
type Endpoints struct {
    CreateUserEndpoint endpoint.Endpoint
    GetUserEndpoint    endpoint.Endpoint
    UpdateUserEndpoint endpoint.Endpoint
    DeleteUserEndpoint endpoint.Endpoint
    ListUsersEndpoint  endpoint.Endpoint
}

// 创建端点
func MakeEndpoints(s service.UserService) Endpoints {
    return Endpoints{
        CreateUserEndpoint: makeCreateUserEndpoint(s),
        GetUserEndpoint:    makeGetUserEndpoint(s),
        UpdateUserEndpoint: makeUpdateUserEndpoint(s),
        DeleteUserEndpoint: makeDeleteUserEndpoint(s),
        ListUsersEndpoint:  makeListUsersEndpoint(s),
    }
}

// 创建用户端点
func makeCreateUserEndpoint(s service.UserService) endpoint.Endpoint {
    return func(ctx context.Context, request interface{}) (interface{}, error) {
        req := request.(CreateUserRequest)
        user, err := s.CreateUser(ctx, req.Name, req.Email, req.Age)
        if err != nil {
            return CreateUserResponse{Err: err.Error()}, nil
        }
        return CreateUserResponse{User: user}, nil
    }
}

// 获取用户端点
func makeGetUserEndpoint(s service.UserService) endpoint.Endpoint {
    return func(ctx context.Context, request interface{}) (interface{}, error) {
        req := request.(GetUserRequest)
        user, err := s.GetUser(ctx, req.ID)
        if err != nil {
            return GetUserResponse{Err: err.Error()}, nil
        }
        return GetUserResponse{User: user}, nil
    }
}
```

### 3.2 HTTP传输

```go
// transport/http.go
package transport

import (
    "context"
    "encoding/json"
    "net/http"
    "strconv"

    "github.com/gorilla/mux"
    "github.com/go-kit/kit/transport/http"
    "github.com/go-kit/log"
    
    "your-project/endpoint"
)

// 创建HTTP处理器
func NewHTTPHandler(endpoints endpoint.Endpoints, logger log.Logger) http.Handler {
    r := mux.NewRouter()

    options := []http.ServerOption{
        http.ServerErrorHandler(http.NewLogErrorHandler(logger)),
    }

    // 创建用户
    r.Methods("POST").Path("/users").Handler(http.NewServer(
        endpoints.CreateUserEndpoint,
        decodeCreateUserRequest,
        encodeResponse,
        options...,
    ))

    // 获取用户
    r.Methods("GET").Path("/users/{id}").Handler(http.NewServer(
        endpoints.GetUserEndpoint,
        decodeGetUserRequest,
        encodeResponse,
        options...,
    ))

    // 更新用户
    r.Methods("PUT").Path("/users/{id}").Handler(http.NewServer(
        endpoints.UpdateUserEndpoint,
        decodeUpdateUserRequest,
        encodeResponse,
        options...,
    ))

    // 删除用户
    r.Methods("DELETE").Path("/users/{id}").Handler(http.NewServer(
        endpoints.DeleteUserEndpoint,
        decodeDeleteUserRequest,
        encodeResponse,
        options...,
    ))

    // 获取用户列表
    r.Methods("GET").Path("/users").Handler(http.NewServer(
        endpoints.ListUsersEndpoint,
        decodeListUsersRequest,
        encodeResponse,
        options...,
    ))

    return r
}

// 解码创建用户请求
func decodeCreateUserRequest(_ context.Context, r *http.Request) (interface{}, error) {
    var req endpoint.CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        return nil, err
    }
    return req, nil
}

// 解码获取用户请求
func decodeGetUserRequest(_ context.Context, r *http.Request) (interface{}, error) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        return nil, err
    }
    return endpoint.GetUserRequest{ID: id}, nil
}

// 编码响应
func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    return json.NewEncoder(w).Encode(response)
}
```

## 中间件

### 4.1 日志中间件

```go
// middleware/logging.go
package middleware

import (
    "context"
    "time"

    "github.com/go-kit/log"
    "your-project/service"
)

// 日志中间件
func LoggingMiddleware(logger log.Logger) service.Middleware {
    return func(next service.UserService) service.UserService {
        return &loggingMiddleware{
            next:   next,
            logger: logger,
        }
    }
}

type loggingMiddleware struct {
    next   service.UserService
    logger log.Logger
}

func (mw loggingMiddleware) CreateUser(ctx context.Context, name, email string, age int) (user service.User, err error) {
    defer func(begin time.Time) {
        mw.logger.Log(
            "method", "CreateUser",
            "name", name,
            "email", email,
            "age", age,
            "took", time.Since(begin),
            "err", err,
        )
    }(time.Now())
    return mw.next.CreateUser(ctx, name, email, age)
}

func (mw loggingMiddleware) GetUser(ctx context.Context, id int) (user service.User, err error) {
    defer func(begin time.Time) {
        mw.logger.Log(
            "method", "GetUser",
            "id", id,
            "took", time.Since(begin),
            "err", err,
        )
    }(time.Now())
    return mw.next.GetUser(ctx, id)
}
```

### 4.2 指标中间件

```go
// middleware/metrics.go
package middleware

import (
    "context"
    "time"

    "github.com/go-kit/kit/metrics"
    "your-project/service"
)

// 指标中间件
func MetricsMiddleware(counter metrics.Counter, latency metrics.Histogram) service.Middleware {
    return func(next service.UserService) service.UserService {
        return &metricsMiddleware{
            next:    next,
            counter: counter,
            latency: latency,
        }
    }
}

type metricsMiddleware struct {
    next    service.UserService
    counter metrics.Counter
    latency metrics.Histogram
}

func (mw metricsMiddleware) CreateUser(ctx context.Context, name, email string, age int) (user service.User, err error) {
    defer func(begin time.Time) {
        lvs := []string{"method", "CreateUser", "error", fmt.Sprint(err != nil)}
        mw.counter.With(lvs...).Add(1)
        mw.latency.With(lvs...).Observe(time.Since(begin).Seconds())
    }(time.Now())
    return mw.next.CreateUser(ctx, name, email, age)
}
```

## 服务发现

### 5.1 Consul服务发现

```go
// discovery/consul.go
package discovery

import (
    "io"
    "time"

    "github.com/go-kit/kit/sd"
    "github.com/go-kit/kit/sd/consul"
    "github.com/go-kit/log"
    consulapi "github.com/hashicorp/consul/api"
)

// 创建Consul客户端
func NewConsulClient(consulAddr string, logger log.Logger) (consul.Client, error) {
    consulConfig := consulapi.DefaultConfig()
    consulConfig.Address = consulAddr
    
    consulClient, err := consulapi.NewClient(consulConfig)
    if err != nil {
        return nil, err
    }

    client := consul.NewClient(consulClient)
    return client, nil
}

// 服务注册
func RegisterService(client consul.Client, serviceName, serviceAddr string, port int, logger log.Logger) (io.Closer, error) {
    asr := consul.NewRegistrar(client, &consulapi.AgentServiceRegistration{
        ID:      serviceName + "-" + serviceAddr,
        Name:    serviceName,
        Address: serviceAddr,
        Port:    port,
        Check: &consulapi.AgentServiceCheck{
            HTTP:     "http://" + serviceAddr + ":" + fmt.Sprintf("%d", port) + "/health",
            Interval: "10s",
            Timeout:  "3s",
        },
    }, logger)

    asr.Register()
    return asr, nil
}

// 服务发现
func DiscoverService(client consul.Client, serviceName string, logger log.Logger) sd.Instancer {
    tags := []string{}
    passingOnly := true
    instancer := consul.NewInstancer(client, logger, serviceName, tags, passingOnly)
    return instancer
}
```

## 负载均衡

### 6.1 负载均衡器

```go
// loadbalancer/lb.go
package loadbalancer

import (
    "context"
    "time"

    "github.com/go-kit/kit/endpoint"
    "github.com/go-kit/kit/sd"
    "github.com/go-kit/kit/sd/lb"
    "github.com/go-kit/log"
)

// 创建负载均衡端点
func NewLoadBalancedEndpoint(instancer sd.Instancer, factory sd.Factory, logger log.Logger) endpoint.Endpoint {
    endpointer := sd.NewEndpointer(instancer, factory, logger)
    balancer := lb.NewRoundRobin(endpointer)
    retry := lb.Retry(3, 3*time.Second, balancer)
    return retry
}

// 端点工厂
func MakeEndpointFactory(method string) sd.Factory {
    return func(instance string) (endpoint.Endpoint, io.Closer, error) {
        service, err := makeHTTPClient(instance)
        if err != nil {
            return nil, nil, err
        }

        var endpoint endpoint.Endpoint
        switch method {
        case "CreateUser":
            endpoint = makeCreateUserEndpoint(service)
        case "GetUser":
            endpoint = makeGetUserEndpoint(service)
        default:
            return nil, nil, fmt.Errorf("unknown method: %s", method)
        }

        return endpoint, nil, nil
    }
}
```

## 熔断器

### 7.1 熔断器中间件

```go
// circuitbreaker/hystrix.go
package circuitbreaker

import (
    "context"

    "github.com/afex/hystrix-go/hystrix"
    "github.com/go-kit/kit/endpoint"
)

// Hystrix熔断器
func Hystrix(commandName string) endpoint.Middleware {
    return func(next endpoint.Endpoint) endpoint.Endpoint {
        return func(ctx context.Context, request interface{}) (interface{}, error) {
            var resp interface{}
            var err error

            if err := hystrix.Do(commandName, func() error {
                resp, err = next(ctx, request)
                return err
            }, func(err error) error {
                // 降级处理
                resp = struct {
                    Error string `json:"error"`
                }{
                    Error: "服务暂时不可用，请稍后重试",
                }
                return nil
            }); err != nil {
                return nil, err
            }

            return resp, err
        }
    }
}

// 配置熔断器
func ConfigureHystrix() {
    hystrix.ConfigureCommand("user-service", hystrix.CommandConfig{
        Timeout:                int(30 * time.Second / time.Millisecond), // 超时时间
        MaxConcurrentRequests:  100,                                       // 最大并发请求数
        RequestVolumeThreshold: 20,                                        // 请求量阈值
        SleepWindow:            int(5 * time.Second / time.Millisecond),   // 熔断器打开后的休眠时间
        ErrorPercentThreshold:  50,                                        // 错误百分比阈值
    })
}
```

## 监控和日志

### 8.1 Prometheus指标

```go
// metrics/prometheus.go
package metrics

import (
    "github.com/go-kit/kit/metrics"
    "github.com/go-kit/kit/metrics/prometheus"
    stdprometheus "github.com/prometheus/client_golang/prometheus"
)

// 创建Prometheus指标
func NewPrometheusMetrics() (counter metrics.Counter, latency metrics.Histogram) {
    counter = prometheus.NewCounterFrom(stdprometheus.CounterOpts{
        Namespace: "user_service",
        Subsystem: "requests",
        Name:      "total",
        Help:      "Total number of requests received.",
    }, []string{"method", "error"})

    latency = prometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
        Namespace: "user_service",
        Subsystem: "requests",
        Name:      "latency_microseconds",
        Help:      "Total duration of requests in microseconds.",
    }, []string{"method", "error"})

    return
}
```

## 完整示例

### 9.1 主程序

```go
// main.go
package main

import (
    "fmt"
    "net/http"
    "os"
    "os/signal"
    "syscall"

    "github.com/go-kit/log"
    "github.com/go-kit/log/level"
    
    "your-project/service"
    "your-project/endpoint"
    "your-project/transport"
    "your-project/middleware"
    "your-project/metrics"
)

func main() {
    // 创建日志器
    var logger log.Logger
    logger = log.NewLogfmtLogger(os.Stderr)
    logger = log.With(logger, "ts", log.DefaultTimestampUTC)
    logger = log.With(logger, "caller", log.DefaultCaller)

    // 创建指标
    counter, latency := metrics.NewPrometheusMetrics()

    // 创建服务
    var svc service.UserService
    svc = service.NewUserService()
    svc = middleware.LoggingMiddleware(logger)(svc)
    svc = middleware.MetricsMiddleware(counter, latency)(svc)

    // 创建端点
    endpoints := endpoint.MakeEndpoints(svc)

    // 创建HTTP处理器
    httpHandler := transport.NewHTTPHandler(endpoints, logger)

    // 启动HTTP服务器
    errs := make(chan error)
    go func() {
        c := make(chan os.Signal, 1)
        signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
        errs <- fmt.Errorf("%s", <-c)
    }()

    go func() {
        level.Info(logger).Log("transport", "HTTP", "addr", ":8080")
        errs <- http.ListenAndServe(":8080", httpHandler)
    }()

    level.Error(logger).Log("exit", <-errs)
}
```
