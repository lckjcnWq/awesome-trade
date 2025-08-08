# Redis 详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [环境准备](#环境准备)
3. [连接管理](#连接管理)
4. [基本数据类型](#基本数据类型)
5. [高级数据结构](#高级数据结构)
6. [发布订阅](#发布订阅)
7. [事务和管道](#事务和管道)
8. [集群和哨兵](#集群和哨兵)
9. [实际应用](#实际应用)

## 基础概念

### 1.1 Redis简介

Redis是高性能的内存数据库，支持多种数据结构，广泛用于缓存、会话存储、消息队列等场景。

```bash
# 安装Redis Go客户端
go get github.com/redis/go-redis/v9
go get github.com/gomodule/redigo/redis
```

### 1.2 核心特性

```go
// 主要包导入
import (
    "context"
    "time"
    
    "github.com/redis/go-redis/v9"
    "github.com/gomodule/redigo/redis"
)
```

## 环境准备

### 2.1 连接配置

```go
// config/redis.go
package config

import (
    "time"
    
    "github.com/redis/go-redis/v9"
)

type RedisConfig struct {
    Host         string
    Port         int
    Password     string
    DB           int
    PoolSize     int
    MinIdleConns int
    MaxRetries   int
    DialTimeout  time.Duration
    ReadTimeout  time.Duration
    WriteTimeout time.Duration
}

func DefaultRedisConfig() *RedisConfig {
    return &RedisConfig{
        Host:         "localhost",
        Port:         6379,
        Password:     "", // 无密码
        DB:           0,  // 默认数据库
        PoolSize:     10,
        MinIdleConns: 5,
        MaxRetries:   3,
        DialTimeout:  5 * time.Second,
        ReadTimeout:  3 * time.Second,
        WriteTimeout: 3 * time.Second,
    }
}

// 创建Redis客户端
func NewRedisClient(config *RedisConfig) *redis.Client {
    rdb := redis.NewClient(&redis.Options{
        Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
        Password:     config.Password,
        DB:           config.DB,
        PoolSize:     config.PoolSize,
        MinIdleConns: config.MinIdleConns,
        MaxRetries:   config.MaxRetries,
        DialTimeout:  config.DialTimeout,
        ReadTimeout:  config.ReadTimeout,
        WriteTimeout: config.WriteTimeout,
    })

    return rdb
}

// 测试连接
func TestConnection(rdb *redis.Client) error {
    ctx := context.Background()
    _, err := rdb.Ping(ctx).Result()
    return err
}
```

## 连接管理

### 3.1 基本连接

```go
// client/redis_client.go
package client

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
)

type RedisClient struct {
    client *redis.Client
    ctx    context.Context
}

func NewRedisClient(rdb *redis.Client) *RedisClient {
    return &RedisClient{
        client: rdb,
        ctx:    context.Background(),
    }
}

// 关闭连接
func (r *RedisClient) Close() error {
    return r.client.Close()
}

// 获取连接状态
func (r *RedisClient) GetStats() *redis.PoolStats {
    return r.client.PoolStats()
}

// 选择数据库
func (r *RedisClient) SelectDB(db int) error {
    return r.client.Do(r.ctx, "SELECT", db).Err()
}

// 清空当前数据库
func (r *RedisClient) FlushDB() error {
    return r.client.FlushDB(r.ctx).Err()
}

// 清空所有数据库
func (r *RedisClient) FlushAll() error {
    return r.client.FlushAll(r.ctx).Err()
}
```

## 基本数据类型

### 4.1 字符串操作

```go
// operations/string.go
package operations

import (
    "context"
    "time"

    "github.com/redis/go-redis/v9"
)

type StringOperations struct {
    client *redis.Client
    ctx    context.Context
}

func NewStringOperations(client *redis.Client) *StringOperations {
    return &StringOperations{
        client: client,
        ctx:    context.Background(),
    }
}

// 设置字符串值
func (s *StringOperations) Set(key, value string, expiration time.Duration) error {
    return s.client.Set(s.ctx, key, value, expiration).Err()
}

// 获取字符串值
func (s *StringOperations) Get(key string) (string, error) {
    return s.client.Get(s.ctx, key).Result()
}

// 设置多个键值对
func (s *StringOperations) MSet(pairs ...interface{}) error {
    return s.client.MSet(s.ctx, pairs...).Err()
}

// 获取多个键的值
func (s *StringOperations) MGet(keys ...string) ([]interface{}, error) {
    return s.client.MGet(s.ctx, keys...).Result()
}

// 递增操作
func (s *StringOperations) Incr(key string) (int64, error) {
    return s.client.Incr(s.ctx, key).Result()
}

// 按指定值递增
func (s *StringOperations) IncrBy(key string, value int64) (int64, error) {
    return s.client.IncrBy(s.ctx, key, value).Result()
}

// 递减操作
func (s *StringOperations) Decr(key string) (int64, error) {
    return s.client.Decr(s.ctx, key).Result()
}

// 追加字符串
func (s *StringOperations) Append(key, value string) (int64, error) {
    return s.client.Append(s.ctx, key, value).Result()
}

// 获取字符串长度
func (s *StringOperations) StrLen(key string) (int64, error) {
    return s.client.StrLen(s.ctx, key).Result()
}

// 设置键的过期时间
func (s *StringOperations) Expire(key string, expiration time.Duration) (bool, error) {
    return s.client.Expire(s.ctx, key, expiration).Result()
}

// 获取键的剩余生存时间
func (s *StringOperations) TTL(key string) (time.Duration, error) {
    return s.client.TTL(s.ctx, key).Result()
}
```

### 4.2 哈希操作

```go
// operations/hash.go
package operations

import (
    "context"

    "github.com/redis/go-redis/v9"
)

type HashOperations struct {
    client *redis.Client
    ctx    context.Context
}

func NewHashOperations(client *redis.Client) *HashOperations {
    return &HashOperations{
        client: client,
        ctx:    context.Background(),
    }
}

// 设置哈希字段
func (h *HashOperations) HSet(key string, values ...interface{}) (int64, error) {
    return h.client.HSet(h.ctx, key, values...).Result()
}

// 获取哈希字段值
func (h *HashOperations) HGet(key, field string) (string, error) {
    return h.client.HGet(h.ctx, key, field).Result()
}

// 获取所有哈希字段和值
func (h *HashOperations) HGetAll(key string) (map[string]string, error) {
    return h.client.HGetAll(h.ctx, key).Result()
}

// 获取多个哈希字段值
func (h *HashOperations) HMGet(key string, fields ...string) ([]interface{}, error) {
    return h.client.HMGet(h.ctx, key, fields...).Result()
}

// 删除哈希字段
func (h *HashOperations) HDel(key string, fields ...string) (int64, error) {
    return h.client.HDel(h.ctx, key, fields...).Result()
}

// 检查哈希字段是否存在
func (h *HashOperations) HExists(key, field string) (bool, error) {
    return h.client.HExists(h.ctx, key, field).Result()
}

// 获取哈希字段数量
func (h *HashOperations) HLen(key string) (int64, error) {
    return h.client.HLen(h.ctx, key).Result()
}

// 获取所有哈希字段名
func (h *HashOperations) HKeys(key string) ([]string, error) {
    return h.client.HKeys(h.ctx, key).Result()
}

// 获取所有哈希字段值
func (h *HashOperations) HVals(key string) ([]string, error) {
    return h.client.HVals(h.ctx, key).Result()
}

// 哈希字段递增
func (h *HashOperations) HIncrBy(key, field string, incr int64) (int64, error) {
    return h.client.HIncrBy(h.ctx, key, field, incr).Result()
}
```

### 4.3 列表操作

```go
// operations/list.go
package operations

import (
    "context"
    "time"

    "github.com/redis/go-redis/v9"
)

type ListOperations struct {
    client *redis.Client
    ctx    context.Context
}

func NewListOperations(client *redis.Client) *ListOperations {
    return &ListOperations{
        client: client,
        ctx:    context.Background(),
    }
}

// 从左侧推入元素
func (l *ListOperations) LPush(key string, values ...interface{}) (int64, error) {
    return l.client.LPush(l.ctx, key, values...).Result()
}

// 从右侧推入元素
func (l *ListOperations) RPush(key string, values ...interface{}) (int64, error) {
    return l.client.RPush(l.ctx, key, values...).Result()
}

// 从左侧弹出元素
func (l *ListOperations) LPop(key string) (string, error) {
    return l.client.LPop(l.ctx, key).Result()
}

// 从右侧弹出元素
func (l *ListOperations) RPop(key string) (string, error) {
    return l.client.RPop(l.ctx, key).Result()
}

// 阻塞式从左侧弹出元素
func (l *ListOperations) BLPop(timeout time.Duration, keys ...string) ([]string, error) {
    return l.client.BLPop(l.ctx, timeout, keys...).Result()
}

// 阻塞式从右侧弹出元素
func (l *ListOperations) BRPop(timeout time.Duration, keys ...string) ([]string, error) {
    return l.client.BRPop(l.ctx, timeout, keys...).Result()
}

// 获取列表长度
func (l *ListOperations) LLen(key string) (int64, error) {
    return l.client.LLen(l.ctx, key).Result()
}

// 获取列表范围内的元素
func (l *ListOperations) LRange(key string, start, stop int64) ([]string, error) {
    return l.client.LRange(l.ctx, key, start, stop).Result()
}

// 根据索引获取元素
func (l *ListOperations) LIndex(key string, index int64) (string, error) {
    return l.client.LIndex(l.ctx, key, index).Result()
}

// 根据索引设置元素值
func (l *ListOperations) LSet(key string, index int64, value interface{}) error {
    return l.client.LSet(l.ctx, key, index, value).Err()
}

// 移除列表元素
func (l *ListOperations) LRem(key string, count int64, value interface{}) (int64, error) {
    return l.client.LRem(l.ctx, key, count, value).Result()
}

// 修剪列表
func (l *ListOperations) LTrim(key string, start, stop int64) error {
    return l.client.LTrim(l.ctx, key, start, stop).Err()
}
```

## 高级数据结构

### 5.1 集合操作

```go
// operations/set.go
package operations

import (
    "context"

    "github.com/redis/go-redis/v9"
)

type SetOperations struct {
    client *redis.Client
    ctx    context.Context
}

func NewSetOperations(client *redis.Client) *SetOperations {
    return &SetOperations{
        client: client,
        ctx:    context.Background(),
    }
}

// 添加集合成员
func (s *SetOperations) SAdd(key string, members ...interface{}) (int64, error) {
    return s.client.SAdd(s.ctx, key, members...).Result()
}

// 获取集合所有成员
func (s *SetOperations) SMembers(key string) ([]string, error) {
    return s.client.SMembers(s.ctx, key).Result()
}

// 检查成员是否存在
func (s *SetOperations) SIsMember(key string, member interface{}) (bool, error) {
    return s.client.SIsMember(s.ctx, key, member).Result()
}

// 获取集合成员数量
func (s *SetOperations) SCard(key string) (int64, error) {
    return s.client.SCard(s.ctx, key).Result()
}

// 移除集合成员
func (s *SetOperations) SRem(key string, members ...interface{}) (int64, error) {
    return s.client.SRem(s.ctx, key, members...).Result()
}

// 随机弹出成员
func (s *SetOperations) SPop(key string) (string, error) {
    return s.client.SPop(s.ctx, key).Result()
}

// 随机获取成员
func (s *SetOperations) SRandMember(key string) (string, error) {
    return s.client.SRandMember(s.ctx, key).Result()
}

// 集合交集
func (s *SetOperations) SInter(keys ...string) ([]string, error) {
    return s.client.SInter(s.ctx, keys...).Result()
}

// 集合并集
func (s *SetOperations) SUnion(keys ...string) ([]string, error) {
    return s.client.SUnion(s.ctx, keys...).Result()
}

// 集合差集
func (s *SetOperations) SDiff(keys ...string) ([]string, error) {
    return s.client.SDiff(s.ctx, keys...).Result()
}
```

### 5.2 有序集合操作

```go
// operations/zset.go
package operations

import (
    "context"

    "github.com/redis/go-redis/v9"
)

type ZSetOperations struct {
    client *redis.Client
    ctx    context.Context
}

func NewZSetOperations(client *redis.Client) *ZSetOperations {
    return &ZSetOperations{
        client: client,
        ctx:    context.Background(),
    }
}

// 添加有序集合成员
func (z *ZSetOperations) ZAdd(key string, members ...*redis.Z) (int64, error) {
    return z.client.ZAdd(z.ctx, key, members...).Result()
}

// 获取成员分数
func (z *ZSetOperations) ZScore(key, member string) (float64, error) {
    return z.client.ZScore(z.ctx, key, member).Result()
}

// 获取成员排名
func (z *ZSetOperations) ZRank(key, member string) (int64, error) {
    return z.client.ZRank(z.ctx, key, member).Result()
}

// 获取成员逆序排名
func (z *ZSetOperations) ZRevRank(key, member string) (int64, error) {
    return z.client.ZRevRank(z.ctx, key, member).Result()
}

// 按排名范围获取成员
func (z *ZSetOperations) ZRange(key string, start, stop int64) ([]string, error) {
    return z.client.ZRange(z.ctx, key, start, stop).Result()
}

// 按排名范围获取成员（逆序）
func (z *ZSetOperations) ZRevRange(key string, start, stop int64) ([]string, error) {
    return z.client.ZRevRange(z.ctx, key, start, stop).Result()
}

// 按分数范围获取成员
func (z *ZSetOperations) ZRangeByScore(key string, opt *redis.ZRangeBy) ([]string, error) {
    return z.client.ZRangeByScore(z.ctx, key, opt).Result()
}

// 获取有序集合成员数量
func (z *ZSetOperations) ZCard(key string) (int64, error) {
    return z.client.ZCard(z.ctx, key).Result()
}

// 获取分数范围内的成员数量
func (z *ZSetOperations) ZCount(key, min, max string) (int64, error) {
    return z.client.ZCount(z.ctx, key, min, max).Result()
}

// 移除有序集合成员
func (z *ZSetOperations) ZRem(key string, members ...interface{}) (int64, error) {
    return z.client.ZRem(z.ctx, key, members...).Result()
}

// 按排名范围移除成员
func (z *ZSetOperations) ZRemRangeByRank(key string, start, stop int64) (int64, error) {
    return z.client.ZRemRangeByRank(z.ctx, key, start, stop).Result()
}

// 按分数范围移除成员
func (z *ZSetOperations) ZRemRangeByScore(key, min, max string) (int64, error) {
    return z.client.ZRemRangeByScore(z.ctx, key, min, max).Result()
}

// 增加成员分数
func (z *ZSetOperations) ZIncrBy(key string, increment float64, member string) (float64, error) {
    return z.client.ZIncrBy(z.ctx, key, increment, member).Result()
}
```

## 发布订阅

### 6.1 发布订阅模式

```go
// pubsub/pubsub.go
package pubsub

import (
    "context"
    "log"

    "github.com/redis/go-redis/v9"
)

type PubSubManager struct {
    client *redis.Client
    ctx    context.Context
}

func NewPubSubManager(client *redis.Client) *PubSubManager {
    return &PubSubManager{
        client: client,
        ctx:    context.Background(),
    }
}

// 发布消息
func (p *PubSubManager) Publish(channel string, message interface{}) error {
    return p.client.Publish(p.ctx, channel, message).Err()
}

// 订阅频道
func (p *PubSubManager) Subscribe(channels ...string) *redis.PubSub {
    return p.client.Subscribe(p.ctx, channels...)
}

// 模式订阅
func (p *PubSubManager) PSubscribe(patterns ...string) *redis.PubSub {
    return p.client.PSubscribe(p.ctx, patterns...)
}

// 消息处理器
func (p *PubSubManager) HandleMessages(pubsub *redis.PubSub, handler func(channel, message string)) {
    defer pubsub.Close()

    ch := pubsub.Channel()
    for msg := range ch {
        handler(msg.Channel, msg.Payload)
    }
}

// 示例：聊天室实现
type ChatRoom struct {
    pubsub  *PubSubManager
    channel string
}

func NewChatRoom(pubsub *PubSubManager, roomName string) *ChatRoom {
    return &ChatRoom{
        pubsub:  pubsub,
        channel: "chat:" + roomName,
    }
}

// 发送消息
func (c *ChatRoom) SendMessage(user, message string) error {
    fullMessage := user + ": " + message
    return c.pubsub.Publish(c.channel, fullMessage)
}

// 加入聊天室
func (c *ChatRoom) Join(handler func(message string)) {
    pubsub := c.pubsub.Subscribe(c.channel)
    
    go c.pubsub.HandleMessages(pubsub, func(channel, message string) {
        handler(message)
    })
}
```

## 事务和管道

### 7.1 事务处理

```go
// transaction/transaction.go
package transaction

import (
    "context"
    "errors"

    "github.com/redis/go-redis/v9"
)

type TransactionManager struct {
    client *redis.Client
    ctx    context.Context
}

func NewTransactionManager(client *redis.Client) *TransactionManager {
    return &TransactionManager{
        client: client,
        ctx:    context.Background(),
    }
}

// 执行事务
func (t *TransactionManager) ExecuteTransaction(fn func(*redis.Tx) error, keys ...string) error {
    return t.client.Watch(t.ctx, func(tx *redis.Tx) error {
        return fn(tx)
    }, keys...)
}

// 示例：转账操作
func (t *TransactionManager) Transfer(fromAccount, toAccount string, amount int64) error {
    return t.ExecuteTransaction(func(tx *redis.Tx) error {
        // 获取账户余额
        fromBalance, err := tx.Get(t.ctx, fromAccount).Int64()
        if err != nil {
            return err
        }

        if fromBalance < amount {
            return errors.New("余额不足")
        }

        // 开始事务
        _, err = tx.TxPipelined(t.ctx, func(pipe redis.Pipeliner) error {
            pipe.DecrBy(t.ctx, fromAccount, amount)
            pipe.IncrBy(t.ctx, toAccount, amount)
            return nil
        })

        return err
    }, fromAccount, toAccount)
}

// 管道操作
func (t *TransactionManager) ExecutePipeline(commands func(redis.Pipeliner)) ([]redis.Cmder, error) {
    pipe := t.client.Pipeline()
    commands(pipe)
    return pipe.Exec(t.ctx)
}

// 批量设置键值对
func (t *TransactionManager) BatchSet(pairs map[string]interface{}) error {
    _, err := t.ExecutePipeline(func(pipe redis.Pipeliner) {
        for key, value := range pairs {
            pipe.Set(t.ctx, key, value, 0)
        }
    })
    return err
}
```

## 集群和哨兵

### 8.1 集群配置

```go
// cluster/cluster.go
package cluster

import (
    "context"
    "time"

    "github.com/redis/go-redis/v9"
)

type ClusterManager struct {
    client *redis.ClusterClient
    ctx    context.Context
}

func NewClusterManager(addrs []string, password string) *ClusterManager {
    rdb := redis.NewClusterClient(&redis.ClusterOptions{
        Addrs:        addrs,
        Password:     password,
        DialTimeout:  5 * time.Second,
        ReadTimeout:  3 * time.Second,
        WriteTimeout: 3 * time.Second,
        PoolSize:     10,
        MinIdleConns: 5,
    })

    return &ClusterManager{
        client: rdb,
        ctx:    context.Background(),
    }
}

// 获取集群信息
func (c *ClusterManager) GetClusterInfo() (string, error) {
    return c.client.ClusterInfo(c.ctx).Result()
}

// 获取集群节点
func (c *ClusterManager) GetClusterNodes() (string, error) {
    return c.client.ClusterNodes(c.ctx).Result()
}

// 关闭集群连接
func (c *ClusterManager) Close() error {
    return c.client.Close()
}

// 哨兵配置
type SentinelManager struct {
    client *redis.Client
    ctx    context.Context
}

func NewSentinelManager(masterName string, sentinelAddrs []string, password string) *SentinelManager {
    rdb := redis.NewFailoverClient(&redis.FailoverOptions{
        MasterName:    masterName,
        SentinelAddrs: sentinelAddrs,
        Password:      password,
        DialTimeout:   5 * time.Second,
        ReadTimeout:   3 * time.Second,
        WriteTimeout:  3 * time.Second,
    })

    return &SentinelManager{
        client: rdb,
        ctx:    context.Background(),
    }
}

// 获取主节点地址
func (s *SentinelManager) GetMasterAddr() ([]string, error) {
    return s.client.Do(s.ctx, "SENTINEL", "get-master-addr-by-name", "mymaster").StringSlice()
}
```

## 实际应用

### 9.1 缓存管理器

```go
// cache/manager.go
package cache

import (
    "context"
    "encoding/json"
    "time"

    "github.com/redis/go-redis/v9"
)

type CacheManager struct {
    client *redis.Client
    ctx    context.Context
    prefix string
}

func NewCacheManager(client *redis.Client, prefix string) *CacheManager {
    return &CacheManager{
        client: client,
        ctx:    context.Background(),
        prefix: prefix,
    }
}

// 构建缓存键
func (c *CacheManager) buildKey(key string) string {
    if c.prefix != "" {
        return c.prefix + ":" + key
    }
    return key
}

// 设置缓存
func (c *CacheManager) Set(key string, value interface{}, expiration time.Duration) error {
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }
    return c.client.Set(c.ctx, c.buildKey(key), data, expiration).Err()
}

// 获取缓存
func (c *CacheManager) Get(key string, dest interface{}) error {
    data, err := c.client.Get(c.ctx, c.buildKey(key)).Result()
    if err != nil {
        return err
    }
    return json.Unmarshal([]byte(data), dest)
}

// 删除缓存
func (c *CacheManager) Delete(key string) error {
    return c.client.Del(c.ctx, c.buildKey(key)).Err()
}

// 检查缓存是否存在
func (c *CacheManager) Exists(key string) (bool, error) {
    count, err := c.client.Exists(c.ctx, c.buildKey(key)).Result()
    return count > 0, err
}

// 设置缓存过期时间
func (c *CacheManager) Expire(key string, expiration time.Duration) error {
    return c.client.Expire(c.ctx, c.buildKey(key), expiration).Err()
}

// 获取或设置缓存（缓存穿透保护）
func (c *CacheManager) GetOrSet(key string, dest interface{}, fetcher func() (interface{}, error), expiration time.Duration) error {
    // 尝试从缓存获取
    err := c.Get(key, dest)
    if err == nil {
        return nil
    }

    // 缓存未命中，从数据源获取
    data, err := fetcher()
    if err != nil {
        return err
    }

    // 设置缓存
    if err := c.Set(key, data, expiration); err != nil {
        return err
    }

    // 将数据复制到目标变量
    dataBytes, _ := json.Marshal(data)
    return json.Unmarshal(dataBytes, dest)
}
```

### 9.2 完整应用示例

```go
// main.go
package main

import (
    "fmt"
    "log"
    "time"

    "your-project/config"
    "your-project/client"
    "your-project/cache"
    "your-project/operations"
)

type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func main() {
    // 创建Redis配置
    cfg := config.DefaultRedisConfig()

    // 创建Redis客户端
    rdb := config.NewRedisClient(cfg)
    defer rdb.Close()

    // 测试连接
    if err := config.TestConnection(rdb); err != nil {
        log.Fatal("Redis连接失败:", err)
    }

    fmt.Println("Redis连接成功!")

    // 创建操作实例
    stringOps := operations.NewStringOperations(rdb)
    hashOps := operations.NewHashOperations(rdb)
    listOps := operations.NewListOperations(rdb)

    // 字符串操作示例
    err := stringOps.Set("user:1:name", "张三", time.Hour)
    if err != nil {
        log.Printf("设置字符串失败: %v", err)
    }

    name, err := stringOps.Get("user:1:name")
    if err != nil {
        log.Printf("获取字符串失败: %v", err)
    } else {
        fmt.Printf("用户名: %s\n", name)
    }

    // 哈希操作示例
    _, err = hashOps.HSet("user:1", "name", "张三", "email", "zhangsan@example.com", "age", "25")
    if err != nil {
        log.Printf("设置哈希失败: %v", err)
    }

    userInfo, err := hashOps.HGetAll("user:1")
    if err != nil {
        log.Printf("获取哈希失败: %v", err)
    } else {
        fmt.Printf("用户信息: %+v\n", userInfo)
    }

    // 列表操作示例
    _, err = listOps.LPush("messages", "消息1", "消息2", "消息3")
    if err != nil {
        log.Printf("推入列表失败: %v", err)
    }

    messages, err := listOps.LRange("messages", 0, -1)
    if err != nil {
        log.Printf("获取列表失败: %v", err)
    } else {
        fmt.Printf("消息列表: %+v\n", messages)
    }

    // 缓存管理器示例
    cacheManager := cache.NewCacheManager(rdb, "app")

    user := User{ID: 1, Name: "张三", Email: "zhangsan@example.com"}
    err = cacheManager.Set("user:1", user, time.Hour)
    if err != nil {
        log.Printf("设置缓存失败: %v", err)
    }

    var cachedUser User
    err = cacheManager.Get("user:1", &cachedUser)
    if err != nil {
        log.Printf("获取缓存失败: %v", err)
    } else {
        fmt.Printf("缓存用户: %+v\n", cachedUser)
    }

    fmt.Println("Redis操作演示完成!")
}
```
