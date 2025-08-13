# Kafka 详细使用指南

## 目录
1. [基础概念](#基础概念)
2. [环境准备](#环境准备)
3. [生产者](#生产者)
4. [消费者](#消费者)
5. [主题管理](#主题管理)
6. [分区和偏移量](#分区和偏移量)
7. [消费者组](#消费者组)
8. [事务处理](#事务处理)
9. [实际应用](#实际应用)

## 基础概念

### 1.1 Kafka简介

Apache Kafka是分布式流处理平台，高性能、跨语言支持，广泛用于消息队列、事件流处理、日志收集等场景。

```bash
# 安装Kafka Go客户端
go get github.com/IBM/sarama
go get github.com/segmentio/kafka-go
```

### 1.2 核心概念

```go
// 主要包导入
import (
    "context"
    "time"
    
    "github.com/IBM/sarama"
    "github.com/segmentio/kafka-go"
)

// 核心概念：
// - Broker: Kafka服务器节点
// - Topic: 消息主题
// - Partition: 主题分区
// - Producer: 消息生产者
// - Consumer: 消息消费者
// - Consumer Group: 消费者组
```

## 环境准备

### 2.1 连接配置

```go
// config/kafka.go
package config

import (
    "time"
    
    "github.com/IBM/sarama"
    "github.com/segmentio/kafka-go"
)

type KafkaConfig struct {
    Brokers          []string
    ClientID         string
    Version          string
    ReturnSuccesses  bool
    ReturnErrors     bool
    RequiredAcks     sarama.RequiredAcks
    Retry            int
    Timeout          time.Duration
    CompressionType  sarama.CompressionCodec
    FlushFrequency   time.Duration
    FlushMessages    int
    FlushBytes       int
}

func DefaultKafkaConfig() *KafkaConfig {
    return &KafkaConfig{
        Brokers:         []string{"localhost:9092"},
        ClientID:        "go-kafka-client",
        Version:         "2.8.0",
        ReturnSuccesses: true,
        ReturnErrors:    true,
        RequiredAcks:    sarama.WaitForAll,
        Retry:           3,
        Timeout:         10 * time.Second,
        CompressionType: sarama.CompressionSnappy,
        FlushFrequency:  100 * time.Millisecond,
        FlushMessages:   100,
        FlushBytes:      1024 * 1024, // 1MB
    }
}

// 创建Sarama配置
func (c *KafkaConfig) ToSaramaConfig() *sarama.Config {
    config := sarama.NewConfig()
    
    // 设置Kafka版本
    version, _ := sarama.ParseKafkaVersion(c.Version)
    config.Version = version
    
    // 生产者配置
    config.Producer.Return.Successes = c.ReturnSuccesses
    config.Producer.Return.Errors = c.ReturnErrors
    config.Producer.RequiredAcks = c.RequiredAcks
    config.Producer.Retry.Max = c.Retry
    config.Producer.Timeout = c.Timeout
    config.Producer.Compression = c.CompressionType
    config.Producer.Flush.Frequency = c.FlushFrequency
    config.Producer.Flush.Messages = c.FlushMessages
    config.Producer.Flush.Bytes = c.FlushBytes
    
    // 消费者配置
    config.Consumer.Return.Errors = true
    config.Consumer.Offsets.Initial = sarama.OffsetNewest
    config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
    
    // 客户端ID
    config.ClientID = c.ClientID
    
    return config
}

// 创建kafka-go配置
func (c *KafkaConfig) ToKafkaGoDialer() *kafka.Dialer {
    return &kafka.Dialer{
        Timeout:   c.Timeout,
        DualStack: true,
    }
}
```

## 生产者

### 3.1 同步生产者

```go
// producer/sync_producer.go
package producer

import (
    "encoding/json"
    "fmt"
    "log"

    "github.com/IBM/sarama"
)

type SyncProducer struct {
    producer sarama.SyncProducer
    config   *sarama.Config
}

func NewSyncProducer(brokers []string, config *sarama.Config) (*SyncProducer, error) {
    producer, err := sarama.NewSyncProducer(brokers, config)
    if err != nil {
        return nil, fmt.Errorf("创建同步生产者失败: %v", err)
    }

    return &SyncProducer{
        producer: producer,
        config:   config,
    }, nil
}

// 发送消息
func (p *SyncProducer) SendMessage(topic, key string, value interface{}) (partition int32, offset int64, err error) {
    // 序列化消息值
    var valueBytes []byte
    switch v := value.(type) {
    case string:
        valueBytes = []byte(v)
    case []byte:
        valueBytes = v
    default:
        valueBytes, err = json.Marshal(v)
        if err != nil {
            return 0, 0, fmt.Errorf("序列化消息失败: %v", err)
        }
    }

    // 创建消息
    msg := &sarama.ProducerMessage{
        Topic: topic,
        Value: sarama.ByteEncoder(valueBytes),
    }

    // 设置消息键
    if key != "" {
        msg.Key = sarama.StringEncoder(key)
    }

    // 发送消息
    partition, offset, err = p.producer.SendMessage(msg)
    if err != nil {
        return 0, 0, fmt.Errorf("发送消息失败: %v", err)
    }

    log.Printf("消息已发送到主题 %s，分区 %d，偏移量 %d", topic, partition, offset)
    return partition, offset, nil
}

// 批量发送消息
func (p *SyncProducer) SendMessages(topic string, messages []MessageData) error {
    var producerMessages []*sarama.ProducerMessage

    for _, msgData := range messages {
        valueBytes, err := json.Marshal(msgData.Value)
        if err != nil {
            return fmt.Errorf("序列化消息失败: %v", err)
        }

        msg := &sarama.ProducerMessage{
            Topic: topic,
            Value: sarama.ByteEncoder(valueBytes),
        }

        if msgData.Key != "" {
            msg.Key = sarama.StringEncoder(msgData.Key)
        }

        producerMessages = append(producerMessages, msg)
    }

    // 批量发送
    err := p.producer.SendMessages(producerMessages)
    if err != nil {
        return fmt.Errorf("批量发送消息失败: %v", err)
    }

    log.Printf("批量发送 %d 条消息到主题 %s", len(messages), topic)
    return nil
}

// 关闭生产者
func (p *SyncProducer) Close() error {
    return p.producer.Close()
}

type MessageData struct {
    Key   string
    Value interface{}
}
```

### 3.2 异步生产者

```go
// producer/async_producer.go
package producer

import (
    "encoding/json"
    "fmt"
    "log"

    "github.com/IBM/sarama"
)

type AsyncProducer struct {
    producer sarama.AsyncProducer
    config   *sarama.Config
}

func NewAsyncProducer(brokers []string, config *sarama.Config) (*AsyncProducer, error) {
    producer, err := sarama.NewAsyncProducer(brokers, config)
    if err != nil {
        return nil, fmt.Errorf("创建异步生产者失败: %v", err)
    }

    asyncProducer := &AsyncProducer{
        producer: producer,
        config:   config,
    }

    // 启动消息处理goroutine
    go asyncProducer.handleSuccesses()
    go asyncProducer.handleErrors()

    return asyncProducer, nil
}

// 处理成功消息
func (p *AsyncProducer) handleSuccesses() {
    for success := range p.producer.Successes() {
        log.Printf("消息发送成功: 主题=%s, 分区=%d, 偏移量=%d", 
            success.Topic, success.Partition, success.Offset)
    }
}

// 处理错误消息
func (p *AsyncProducer) handleErrors() {
    for err := range p.producer.Errors() {
        log.Printf("消息发送失败: %v", err)
    }
}

// 发送消息
func (p *AsyncProducer) SendMessage(topic, key string, value interface{}) error {
    // 序列化消息值
    var valueBytes []byte
    var err error
    
    switch v := value.(type) {
    case string:
        valueBytes = []byte(v)
    case []byte:
        valueBytes = v
    default:
        valueBytes, err = json.Marshal(v)
        if err != nil {
            return fmt.Errorf("序列化消息失败: %v", err)
        }
    }

    // 创建消息
    msg := &sarama.ProducerMessage{
        Topic: topic,
        Value: sarama.ByteEncoder(valueBytes),
    }

    // 设置消息键
    if key != "" {
        msg.Key = sarama.StringEncoder(key)
    }

    // 异步发送消息
    select {
    case p.producer.Input() <- msg:
        return nil
    default:
        return fmt.Errorf("生产者输入通道已满")
    }
}

// 关闭生产者
func (p *AsyncProducer) Close() error {
    p.producer.AsyncClose()
    return nil
}
```

## 消费者

### 4.1 简单消费者

```go
// consumer/simple_consumer.go
package consumer

import (
    "context"
    "encoding/json"
    "fmt"
    "log"

    "github.com/segmentio/kafka-go"
)

type SimpleConsumer struct {
    reader *kafka.Reader
}

func NewSimpleConsumer(brokers []string, topic, groupID string) *SimpleConsumer {
    reader := kafka.NewReader(kafka.ReaderConfig{
        Brokers:  brokers,
        Topic:    topic,
        GroupID:  groupID,
        MinBytes: 10e3, // 10KB
        MaxBytes: 10e6, // 10MB
    })

    return &SimpleConsumer{
        reader: reader,
    }
}

// 消费消息
func (c *SimpleConsumer) ConsumeMessages(ctx context.Context, handler func(kafka.Message) error) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            // 读取消息
            msg, err := c.reader.ReadMessage(ctx)
            if err != nil {
                log.Printf("读取消息失败: %v", err)
                continue
            }

            log.Printf("收到消息: 主题=%s, 分区=%d, 偏移量=%d, 键=%s", 
                msg.Topic, msg.Partition, msg.Offset, string(msg.Key))

            // 处理消息
            if err := handler(msg); err != nil {
                log.Printf("处理消息失败: %v", err)
                continue
            }

            log.Printf("消息处理成功: 偏移量=%d", msg.Offset)
        }
    }
}

// 关闭消费者
func (c *SimpleConsumer) Close() error {
    return c.reader.Close()
}
```

### 4.2 消费者组

```go
// consumer/group_consumer.go
package consumer

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "sync"

    "github.com/IBM/sarama"
)

type GroupConsumer struct {
    consumerGroup sarama.ConsumerGroup
    handler       ConsumerGroupHandler
    topics        []string
    ready         chan bool
}

type ConsumerGroupHandler struct {
    ready    chan bool
    callback func(message *sarama.ConsumerMessage) error
}

func NewGroupConsumer(brokers []string, groupID string, topics []string, config *sarama.Config) (*GroupConsumer, error) {
    consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, config)
    if err != nil {
        return nil, fmt.Errorf("创建消费者组失败: %v", err)
    }

    ready := make(chan bool)
    handler := ConsumerGroupHandler{
        ready: ready,
    }

    return &GroupConsumer{
        consumerGroup: consumerGroup,
        handler:       handler,
        topics:        topics,
        ready:         ready,
    }, nil
}

// 设置消息处理回调
func (c *GroupConsumer) SetMessageHandler(callback func(message *sarama.ConsumerMessage) error) {
    c.handler.callback = callback
}

// 开始消费
func (c *GroupConsumer) Start(ctx context.Context) error {
    wg := &sync.WaitGroup{}
    wg.Add(1)

    go func() {
        defer wg.Done()
        for {
            if err := c.consumerGroup.Consume(ctx, c.topics, &c.handler); err != nil {
                log.Printf("消费者组错误: %v", err)
                return
            }

            if ctx.Err() != nil {
                return
            }

            c.handler.ready = make(chan bool)
        }
    }()

    <-c.ready
    log.Println("消费者组已启动")

    // 处理错误
    go func() {
        for err := range c.consumerGroup.Errors() {
            log.Printf("消费者组错误: %v", err)
        }
    }()

    wg.Wait()
    return nil
}

// 关闭消费者组
func (c *GroupConsumer) Close() error {
    return c.consumerGroup.Close()
}

// 实现sarama.ConsumerGroupHandler接口
func (h *ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
    close(h.ready)
    return nil
}

func (h *ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
    return nil
}

func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
    for {
        select {
        case message := <-claim.Messages():
            if message == nil {
                return nil
            }

            log.Printf("收到消息: 主题=%s, 分区=%d, 偏移量=%d", 
                message.Topic, message.Partition, message.Offset)

            // 调用用户定义的处理函数
            if h.callback != nil {
                if err := h.callback(message); err != nil {
                    log.Printf("处理消息失败: %v", err)
                    continue
                }
            }

            // 标记消息已处理
            session.MarkMessage(message, "")

        case <-session.Context().Done():
            return nil
        }
    }
}
```

## 主题管理

### 5.1 主题操作

```go
// admin/topic_manager.go
package admin

import (
    "fmt"
    "log"

    "github.com/IBM/sarama"
)

type TopicManager struct {
    admin sarama.ClusterAdmin
}

func NewTopicManager(brokers []string, config *sarama.Config) (*TopicManager, error) {
    admin, err := sarama.NewClusterAdmin(brokers, config)
    if err != nil {
        return nil, fmt.Errorf("创建集群管理员失败: %v", err)
    }

    return &TopicManager{
        admin: admin,
    }, nil
}

// 创建主题
func (t *TopicManager) CreateTopic(topicName string, numPartitions int32, replicationFactor int16) error {
    topicDetail := &sarama.TopicDetail{
        NumPartitions:     numPartitions,
        ReplicationFactor: replicationFactor,
        ConfigEntries: map[string]*string{
            "cleanup.policy": &[]string{"delete"}[0],
            "retention.ms":   &[]string{"604800000"}[0], // 7天
        },
    }

    err := t.admin.CreateTopic(topicName, topicDetail, false)
    if err != nil {
        return fmt.Errorf("创建主题失败: %v", err)
    }

    log.Printf("主题 %s 创建成功", topicName)
    return nil
}

// 删除主题
func (t *TopicManager) DeleteTopic(topicName string) error {
    err := t.admin.DeleteTopic(topicName)
    if err != nil {
        return fmt.Errorf("删除主题失败: %v", err)
    }

    log.Printf("主题 %s 删除成功", topicName)
    return nil
}

// 列出所有主题
func (t *TopicManager) ListTopics() (map[string]sarama.TopicDetail, error) {
    metadata, err := t.admin.DescribeTopics(nil)
    if err != nil {
        return nil, fmt.Errorf("获取主题列表失败: %v", err)
    }

    return metadata, nil
}

// 获取主题详情
func (t *TopicManager) GetTopicDetail(topicName string) (*sarama.TopicMetadata, error) {
    metadata, err := t.admin.DescribeTopics([]string{topicName})
    if err != nil {
        return nil, fmt.Errorf("获取主题详情失败: %v", err)
    }

    if topicMeta, exists := metadata[topicName]; exists {
        return &topicMeta, nil
    }

    return nil, fmt.Errorf("主题 %s 不存在", topicName)
}

// 修改主题配置
func (t *TopicManager) AlterTopicConfig(topicName string, configs map[string]*string) error {
    configEntries := make(map[string]*sarama.ConfigEntry)
    for key, value := range configs {
        configEntries[key] = &sarama.ConfigEntry{
            Name:  key,
            Value: *value,
        }
    }

    resource := sarama.ConfigResource{
        Type: sarama.TopicResource,
        Name: topicName,
    }

    err := t.admin.AlterConfig(sarama.TopicResource, topicName, configEntries, false)
    if err != nil {
        return fmt.Errorf("修改主题配置失败: %v", err)
    }

    log.Printf("主题 %s 配置修改成功", topicName)
    return nil
}

// 关闭管理员连接
func (t *TopicManager) Close() error {
    return t.admin.Close()
}
```

## 分区和偏移量

### 6.1 偏移量管理

```go
// offset/manager.go
package offset

import (
    "context"
    "fmt"
    "log"

    "github.com/segmentio/kafka-go"
)

type OffsetManager struct {
    conn *kafka.Conn
}

func NewOffsetManager(brokers []string) (*OffsetManager, error) {
    conn, err := kafka.Dial("tcp", brokers[0])
    if err != nil {
        return nil, fmt.Errorf("连接Kafka失败: %v", err)
    }

    return &OffsetManager{
        conn: conn,
    }, nil
}

// 获取主题分区信息
func (o *OffsetManager) GetPartitions(topic string) ([]kafka.Partition, error) {
    partitions, err := o.conn.ReadPartitions(topic)
    if err != nil {
        return nil, fmt.Errorf("获取分区信息失败: %v", err)
    }

    return partitions, nil
}

// 获取最新偏移量
func (o *OffsetManager) GetLatestOffset(topic string, partition int) (int64, error) {
    conn, err := kafka.DialLeader(context.Background(), "tcp", o.conn.RemoteAddr().String(), topic, partition)
    if err != nil {
        return 0, fmt.Errorf("连接分区leader失败: %v", err)
    }
    defer conn.Close()

    _, high, err := conn.ReadOffsets()
    if err != nil {
        return 0, fmt.Errorf("读取偏移量失败: %v", err)
    }

    return high, nil
}

// 获取最早偏移量
func (o *OffsetManager) GetEarliestOffset(topic string, partition int) (int64, error) {
    conn, err := kafka.DialLeader(context.Background(), "tcp", o.conn.RemoteAddr().String(), topic, partition)
    if err != nil {
        return 0, fmt.Errorf("连接分区leader失败: %v", err)
    }
    defer conn.Close()

    low, _, err := conn.ReadOffsets()
    if err != nil {
        return 0, fmt.Errorf("读取偏移量失败: %v", err)
    }

    return low, nil
}

// 重置消费者组偏移量
func (o *OffsetManager) ResetConsumerGroupOffset(groupID, topic string, partition int, offset int64) error {
    // 这里需要使用Kafka管理API来重置偏移量
    // 实际实现可能需要使用sarama的OffsetManager
    log.Printf("重置消费者组 %s 的偏移量: 主题=%s, 分区=%d, 偏移量=%d", 
        groupID, topic, partition, offset)
    return nil
}

// 关闭连接
func (o *OffsetManager) Close() error {
    return o.conn.Close()
}
```

## 消费者组

### 7.1 消费者组管理

```go
// group/manager.go
package group

import (
    "fmt"
    "log"

    "github.com/IBM/sarama"
)

type GroupManager struct {
    admin sarama.ClusterAdmin
}

func NewGroupManager(brokers []string, config *sarama.Config) (*GroupManager, error) {
    admin, err := sarama.NewClusterAdmin(brokers, config)
    if err != nil {
        return nil, fmt.Errorf("创建集群管理员失败: %v", err)
    }

    return &GroupManager{
        admin: admin,
    }, nil
}

// 列出所有消费者组
func (g *GroupManager) ListConsumerGroups() (map[string]string, error) {
    groups, err := g.admin.ListConsumerGroups()
    if err != nil {
        return nil, fmt.Errorf("获取消费者组列表失败: %v", err)
    }

    return groups, nil
}

// 获取消费者组详情
func (g *GroupManager) DescribeConsumerGroup(groupID string) (*sarama.GroupDescription, error) {
    groups, err := g.admin.DescribeConsumerGroups([]string{groupID})
    if err != nil {
        return nil, fmt.Errorf("获取消费者组详情失败: %v", err)
    }

    if group, exists := groups[groupID]; exists {
        return group, nil
    }

    return nil, fmt.Errorf("消费者组 %s 不存在", groupID)
}

// 删除消费者组
func (g *GroupManager) DeleteConsumerGroup(groupID string) error {
    err := g.admin.DeleteConsumerGroup(groupID)
    if err != nil {
        return fmt.Errorf("删除消费者组失败: %v", err)
    }

    log.Printf("消费者组 %s 删除成功", groupID)
    return nil
}

// 获取消费者组偏移量
func (g *GroupManager) GetConsumerGroupOffsets(groupID string) (map[string]map[int32]int64, error) {
    coordinator, err := g.admin.FindCoordinator(groupID)
    if err != nil {
        return nil, fmt.Errorf("查找协调器失败: %v", err)
    }

    // 这里需要实现获取偏移量的逻辑
    log.Printf("消费者组 %s 的协调器: %s", groupID, coordinator.Addr)
    
    // 返回示例数据
    offsets := make(map[string]map[int32]int64)
    return offsets, nil
}

// 关闭管理器
func (g *GroupManager) Close() error {
    return g.admin.Close()
}
```

## 事务处理

### 8.1 事务生产者

```go
// transaction/producer.go
package transaction

import (
    "fmt"
    "log"

    "github.com/IBM/sarama"
)

type TransactionalProducer struct {
    producer sarama.AsyncProducer
    config   *sarama.Config
}

func NewTransactionalProducer(brokers []string, transactionID string) (*TransactionalProducer, error) {
    config := sarama.NewConfig()
    config.Version = sarama.V2_5_0_0
    config.Producer.Transaction.ID = transactionID
    config.Producer.RequiredAcks = sarama.WaitForAll
    config.Producer.Return.Successes = true
    config.Producer.Return.Errors = true
    config.Producer.Idempotent = true
    config.Net.MaxOpenRequests = 1

    producer, err := sarama.NewAsyncProducer(brokers, config)
    if err != nil {
        return nil, fmt.Errorf("创建事务生产者失败: %v", err)
    }

    return &TransactionalProducer{
        producer: producer,
        config:   config,
    }, nil
}

// 开始事务
func (t *TransactionalProducer) BeginTransaction() error {
    // Sarama目前不直接支持事务API，这里是概念性实现
    log.Println("开始事务")
    return nil
}

// 提交事务
func (t *TransactionalProducer) CommitTransaction() error {
    log.Println("提交事务")
    return nil
}

// 回滚事务
func (t *TransactionalProducer) AbortTransaction() error {
    log.Println("回滚事务")
    return nil
}

// 在事务中发送消息
func (t *TransactionalProducer) SendMessageInTransaction(topic, key string, value []byte) error {
    msg := &sarama.ProducerMessage{
        Topic: topic,
        Key:   sarama.StringEncoder(key),
        Value: sarama.ByteEncoder(value),
    }

    select {
    case t.producer.Input() <- msg:
        return nil
    default:
        return fmt.Errorf("生产者输入通道已满")
    }
}

// 关闭生产者
func (t *TransactionalProducer) Close() error {
    t.producer.AsyncClose()
    return nil
}
```

## 实际应用

### 9.1 消息队列服务

```go
// service/message_service.go
package service

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "time"

    "your-project/config"
    "your-project/producer"
    "your-project/consumer"
)

type MessageService struct {
    syncProducer  *producer.SyncProducer
    asyncProducer *producer.AsyncProducer
    consumer      *consumer.SimpleConsumer
    config        *config.KafkaConfig
}

type Message struct {
    ID        string                 `json:"id"`
    Type      string                 `json:"type"`
    Payload   map[string]interface{} `json:"payload"`
    Timestamp time.Time              `json:"timestamp"`
    Source    string                 `json:"source"`
}

func NewMessageService(cfg *config.KafkaConfig) (*MessageService, error) {
    saramaConfig := cfg.ToSaramaConfig()

    // 创建同步生产者
    syncProd, err := producer.NewSyncProducer(cfg.Brokers, saramaConfig)
    if err != nil {
        return nil, err
    }

    // 创建异步生产者
    asyncProd, err := producer.NewAsyncProducer(cfg.Brokers, saramaConfig)
    if err != nil {
        return nil, err
    }

    return &MessageService{
        syncProducer:  syncProd,
        asyncProducer: asyncProd,
        config:        cfg,
    }, nil
}

// 发送消息（同步）
func (s *MessageService) SendMessage(topic string, msg Message) error {
    msg.Timestamp = time.Now()
    
    msgBytes, err := json.Marshal(msg)
    if err != nil {
        return fmt.Errorf("序列化消息失败: %v", err)
    }

    _, _, err = s.syncProducer.SendMessage(topic, msg.ID, msgBytes)
    return err
}

// 发送消息（异步）
func (s *MessageService) SendMessageAsync(topic string, msg Message) error {
    msg.Timestamp = time.Now()
    
    msgBytes, err := json.Marshal(msg)
    if err != nil {
        return fmt.Errorf("序列化消息失败: %v", err)
    }

    return s.asyncProducer.SendMessage(topic, msg.ID, msgBytes)
}

// 批量发送消息
func (s *MessageService) SendBatchMessages(topic string, messages []Message) error {
    var msgDataList []producer.MessageData
    
    for _, msg := range messages {
        msg.Timestamp = time.Now()
        msgDataList = append(msgDataList, producer.MessageData{
            Key:   msg.ID,
            Value: msg,
        })
    }

    return s.syncProducer.SendMessages(topic, msgDataList)
}

// 消费消息
func (s *MessageService) ConsumeMessages(ctx context.Context, topic, groupID string, handler func(Message) error) error {
    consumer := consumer.NewSimpleConsumer(s.config.Brokers, topic, groupID)
    defer consumer.Close()

    return consumer.ConsumeMessages(ctx, func(kafkaMsg kafka.Message) error {
        var msg Message
        if err := json.Unmarshal(kafkaMsg.Value, &msg); err != nil {
            return fmt.Errorf("反序列化消息失败: %v", err)
        }

        return handler(msg)
    })
}

// 关闭服务
func (s *MessageService) Close() error {
    if err := s.syncProducer.Close(); err != nil {
        log.Printf("关闭同步生产者失败: %v", err)
    }
    
    if err := s.asyncProducer.Close(); err != nil {
        log.Printf("关闭异步生产者失败: %v", err)
    }

    return nil
}
```

### 9.2 完整应用示例

```go
// main.go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "your-project/config"
    "your-project/service"
    "your-project/admin"
)

func main() {
    // 创建Kafka配置
    cfg := config.DefaultKafkaConfig()

    // 创建主题管理器
    topicManager, err := admin.NewTopicManager(cfg.Brokers, cfg.ToSaramaConfig())
    if err != nil {
        log.Fatal("创建主题管理器失败:", err)
    }
    defer topicManager.Close()

    // 创建主题
    topicName := "user-events"
    err = topicManager.CreateTopic(topicName, 3, 1)
    if err != nil {
        log.Printf("创建主题失败: %v", err)
    }

    // 创建消息服务
    msgService, err := service.NewMessageService(cfg)
    if err != nil {
        log.Fatal("创建消息服务失败:", err)
    }
    defer msgService.Close()

    // 发送消息示例
    message := service.Message{
        ID:   "msg-001",
        Type: "user.created",
        Payload: map[string]interface{}{
            "user_id": 12345,
            "name":    "张三",
            "email":   "zhangsan@example.com",
        },
        Source: "user-service",
    }

    err = msgService.SendMessage(topicName, message)
    if err != nil {
        log.Printf("发送消息失败: %v", err)
    } else {
        fmt.Println("消息发送成功!")
    }

    // 批量发送消息示例
    messages := []service.Message{
        {
            ID:   "msg-002",
            Type: "user.updated",
            Payload: map[string]interface{}{
                "user_id": 12345,
                "name":    "张三三",
            },
            Source: "user-service",
        },
        {
            ID:   "msg-003",
            Type: "user.deleted",
            Payload: map[string]interface{}{
                "user_id": 12346,
            },
            Source: "user-service",
        },
    }

    err = msgService.SendBatchMessages(topicName, messages)
    if err != nil {
        log.Printf("批量发送消息失败: %v", err)
    } else {
        fmt.Println("批量消息发送成功!")
    }

    // 消费消息示例
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    fmt.Println("开始消费消息...")
    err = msgService.ConsumeMessages(ctx, topicName, "user-service-group", func(msg service.Message) error {
        fmt.Printf("收到消息: ID=%s, Type=%s, Payload=%+v\n", 
            msg.ID, msg.Type, msg.Payload)
        return nil
    })

    if err != nil && err != context.DeadlineExceeded {
        log.Printf("消费消息失败: %v", err)
    }

    fmt.Println("Kafka操作演示完成!")
}
```
