# WebSocket 实时推送系统

## 在 ChainFeed 中的作用

ChainFeed 的核心功能是**实时追踪链上活动**，WebSocket 是实现这一功能的关键技术。

### 业务场景

```
用户关注地址 → 该地址发生交易 → 实时推送到用户前端 → 像刷 Twitter 一样查看
```

### 完整数据流（Redis Streams 架构）

```
┌─────────────┐
│  Ethereum   │ 链上交易发生
└──────┬──────┘
       │
       ↓ Webhook
┌─────────────┐
│   Alchemy   │ 监听链上事件
└──────┬──────┘
       │
       ↓ HTTP POST
┌─────────────┐
│   Backend   │ 解析交易，存入数据库
│  /webhooks  │
└──────┬──────┘
       │
       ↓ XADD (发布到 Stream)
┌─────────────┐
│   Redis     │ 消息队列（持久化）
│   Streams   │ • 消息确认（ACK）
│             │ • 自动重试（3次）
│             │ • 消费者组
└──────┬──────┘
       │
       ↓ XREADGROUP (消费消息)
┌─────────────┐
│ WebSocket   │ 推送给在线用户
│    Hub      │ • 按 user_id 路由
│             │ • 支持多连接
└──────┬──────┘
       │
       ↓ Push
┌─────────────┐
│  Frontend   │ 实时显示 Feed
│   /feed     │
└─────────────┘
```

## 技术实现

### 后端架构

#### 1. WebSocket Hub（连接管理）

```go
type Hub struct {
    clients    map[int64]map[*Client]bool  // user_id -> clients
    broadcast  chan *Message                // 广播通道
    Register   chan *Client                 // 注册通道
    Unregister chan *Client                 // 注销通道
}
```

**特点**：
- 支持一个用户多个连接（多标签页）
- 使用 channel 实现并发安全
- 自动清理断开的连接
- 按 user_id 精准路由消息

#### 2. Redis Streams（消息队列）

**为什么从 Pub/Sub 迁移到 Streams？**

| 特性 | Pub/Sub | Streams |
|------|---------|---------|
| 消息持久化 | ❌ 内存中，断线丢失 | ✅ 持久化到磁盘 |
| 消息确认 | ❌ 无 | ✅ ACK 机制 |
| 重试机制 | ❌ 无 | ✅ 自动重试 |
| 消费者组 | ❌ 无 | ✅ 支持 |
| 负载均衡 | ❌ 广播模式 | ✅ 自动分配 |
| 消息回溯 | ❌ 无 | ✅ 支持 |

**核心实现**：

```go
// 发布消息到 Stream
func (bp *BatchProcessor) publishFeedUpdate(ctx context.Context, userID int64, tx *Transaction) {
    values := map[string]interface{}{
        "user_id": userID,
        "type":    "new_transaction",
        "payload": jsonData,
    }
    
    bp.redis.XAdd(ctx, &redis.XAddArgs{
        Stream: "feed:stream",
        Values: values,
    })
}

// 消费消息
func (s *StreamService) Consume(ctx context.Context) error {
    // 创建消费者组
    s.redis.XGroupCreateMkStream(ctx, "feed:stream", "feed:consumers", "0")
    
    for {
        // 批量读取消息（每次10条）
        streams, _ := s.redis.XReadGroup(ctx, &redis.XReadGroupArgs{
            Group:    "feed:consumers",
            Consumer: "consumer-1",
            Streams:  []string{"feed:stream", ">"},
            Count:    10,
            Block:    time.Second,
        }).Result()
        
        for _, stream := range streams {
            for _, message := range stream.Messages {
                // 处理消息
                s.processMessage(ctx, message)
                
                // 确认消息
                s.redis.XAck(ctx, "feed:stream", "feed:consumers", message.ID)
            }
        }
        
        // 处理超时未确认的消息（30秒后重试）
        s.claimPendingMessages(ctx)
    }
}
```

**消息重试机制**：

```go
func (s *StreamService) claimPendingMessages(ctx context.Context) {
    // 查询待处理消息
    pending, _ := s.redis.XPendingExt(ctx, &redis.XPendingExtArgs{
        Stream: "feed:stream",
        Group:  "feed:consumers",
        Start:  "-",
        End:    "+",
        Count:  10,
    }).Result()
    
    for _, msg := range pending {
        // 超过最大重试次数（3次），丢弃消息
        if msg.RetryCount >= 3 {
            s.logger.Warn("message exceeded max retries, discarding",
                zap.String("message_id", msg.ID))
            s.redis.XAck(ctx, "feed:stream", "feed:consumers", msg.ID)
            continue
        }
        
        // 认领超时消息（30秒未确认）
        if msg.Idle >= 30*time.Second {
            claimed, _ := s.redis.XClaim(ctx, &redis.XClaimArgs{
                Stream:   "feed:stream",
                Group:    "feed:consumers",
                Consumer: "consumer-1",
                MinIdle:  30 * time.Second,
                Messages: []string{msg.ID},
            }).Result()
            
            // 重新处理
            for _, claimedMsg := range claimed {
                s.processMessage(ctx, claimedMsg)
            }
        }
    }
}
```

**消息路由**：

```go
func (h *Hub) Run() {
    for {
        select {
        case message := <-h.broadcast:
            // 只推送给目标用户的所有连接
            h.mu.RLock()
            clients := h.clients[message.UserID]
            h.mu.RUnlock()
            
            data, _ := json.Marshal(message)
            
            for client := range clients {
                select {
                case client.Send <- data:
                    // 发送成功
                default:
                    // 发送失败，关闭连接
                    close(client.Send)
                    delete(clients, client)
                }
            }
        }
    }
}
```

#### 3. 认证机制

WebSocket 无法设置自定义 Header，通过 **Query 参数** 传递 JWT：

```
ws://localhost:8080/ws?token=eyJhbGciOiJIUzI1NiIs...
```

后端中间件支持两种方式：
- REST API：`Authorization: Bearer <token>`
- WebSocket：`?token=<token>`

### 前端实现

#### 1. useWebSocket Hook

```typescript
const { isConnected } = useWebSocket({
  url: 'ws://localhost:8080/ws',
  token: authToken,
  onMessage: (data) => {
    // 收到新交易，更新 Feed
    setFeeds((prev) => [data, ...prev]);
  },
});
```

**特性**：
- 自动重连（3 秒间隔）
- Token 过期自动停止重连
- 连接状态实时显示

#### 2. Feed 页面

```tsx
<div className="flex items-center gap-2">
  <div className={`w-2 h-2 rounded-full ${isConnected ? 'bg-green-500' : 'bg-gray-400'}`} />
  <span>{isConnected ? '已连接' : '未连接'}</span>
</div>
```

## Redis Streams 技术细节

### 核心概念

**Stream（流）**：
- 类似 Kafka 的 Topic
- 消息持久化存储
- 支持消息回溯

**Consumer Group（消费者组）**：
- 多个消费者协同工作
- 自动负载均衡
- 消息只被消费一次

**Pending List（待处理列表）**：
- 已读取但未确认的消息
- 支持超时重试
- 防止消息丢失

### 消息生命周期

```
1. 生产者 XADD 发布消息
   ↓
2. 消息进入 Stream（持久化）
   ↓
3. 消费者 XREADGROUP 读取消息
   ↓
4. 消息进入 Pending List
   ↓
5. 处理成功，XACK 确认
   ↓
6. 消息从 Pending List 移除

如果步骤 5 失败：
   ↓
7. 30秒后，消息被 XCLAIM 认领
   ↓
8. 重新处理（最多3次）
   ↓
9. 超过重试次数，记录日志并丢弃
```

### 监控命令

```bash
# 查看 Stream 信息
docker-compose exec redis redis-cli XINFO STREAM feed:stream

# 输出示例：
# length: 1234                    # 消息总数
# first-entry: 1675234567890-0    # 最早消息ID
# last-entry: 1675234567999-0     # 最新消息ID

# 查看消费者组
docker-compose exec redis redis-cli XINFO GROUPS feed:stream

# 输出示例：
# name: feed:consumers
# consumers: 1
# pending: 5                      # 待处理消息数
# last-delivered-id: 1675234567995-0

# 查看待处理消息
docker-compose exec redis redis-cli XPENDING feed:stream feed:consumers

# 输出示例：
# 1) (integer) 5                  # 待处理消息数
# 2) "1675234567890-0"            # 最早待处理消息ID
# 3) "1675234567894-0"            # 最新待处理消息ID
# 4) 1) 1) "consumer-1"           # 消费者名称
#       2) "5"                    # 该消费者待处理数

# 查看待处理消息详情
docker-compose exec redis redis-cli XPENDING feed:stream feed:consumers - + 10

# 输出示例：
# 1) 1) "1675234567890-0"         # 消息ID
#    2) "consumer-1"              # 消费者
#    3) (integer) 35000           # 空闲时间（毫秒）
#    4) (integer) 2               # 重试次数
```

### 性能优化

**1. 批量消费**：
```go
// 每次读取 10 条消息
Count: 10
```

**2. Stream 清理**：
```bash
# 保留最近 10000 条消息（定期执行）
docker-compose exec redis redis-cli XTRIM feed:stream MAXLEN ~ 10000
```

**3. 多消费者部署**：
```go
// 使用主机名作为消费者名称
ConsumerName = os.Hostname()  // consumer-1, consumer-2, ...
```

## 常见问题

### 1. WebSocket disconnected (401)

**原因**：未携带 token 或 token 过期

**解决**：
```typescript
// 从 localStorage 获取 token
const token = localStorage.getItem('auth_token');

// 传递给 WebSocket
useWebSocket({ url, token });
```

### 2. 连接成功但收不到消息

**检查清单**：
1. 用户是否添加了监控地址？
2. 监控的地址是否有交易发生？
3. Redis Streams 是否正常工作？
4. Webhook 是否配置正确？

**调试**：
```bash
# 查看 WebSocket 连接
docker-compose logs app | grep "websocket client connected"

# 查看 Stream 消息数量
docker-compose exec redis redis-cli XLEN feed:stream

# 查看待处理消息
docker-compose exec redis redis-cli XPENDING feed:stream feed:consumers

# 手动发送测试消息
docker-compose exec redis redis-cli XADD feed:stream '*' user_id 1 type test payload '{"test":"data"}'
```

### 3. 消息堆积

**现象**：
```bash
docker-compose exec redis redis-cli XLEN feed:stream
# 输出：(integer) 50000  # 消息过多
```

**原因**：
- 消费速度慢于生产速度
- 消费者卡住或宕机

**解决**：
```bash
# 1. 检查消费者状态
docker-compose logs app | grep "consumer group initialized"

# 2. 查看待处理消息
docker-compose exec redis redis-cli XPENDING feed:stream feed:consumers

# 3. 增加消费者实例（水平扩展）
docker-compose up -d --scale app=3

# 4. 清理旧消息
docker-compose exec redis redis-cli XTRIM feed:stream MAXLEN ~ 10000
```

### 4. 消息重复消费

**原因**：
- 处理成功但 ACK 失败
- 消费者在 ACK 前宕机

**解决**：
```go
// 在数据库层面保证幂等性
INSERT INTO feed_items (user_id, transaction_id)
VALUES ($1, $2)
ON CONFLICT (user_id, transaction_id) DO NOTHING
```

### 5. 多标签页同时连接

**正常行为**：一个用户可以有多个 WebSocket 连接

```
user_id=123 | user_connections=2 | total_connections=150
```

每个标签页都会收到推送。

## 性能优化

### 1. 连接池管理

```go
// 限制单用户最大连接数
const maxConnectionsPerUser = 5

if len(h.clients[userID]) >= maxConnectionsPerUser {
    // 拒绝新连接或关闭最旧的连接
}
```

### 2. 心跳优化

```go
// 60 秒心跳检测
const pongWait = 60 * time.Second
const pingPeriod = 54 * time.Second  // 90% of pongWait
```

### 3. 消息压缩

```go
// 对大消息进行 gzip 压缩
if len(data) > 1024 {
    data = gzipCompress(data)
}
```

## 监控指标

### 关键指标

```
# WebSocket
websocket_active_connections{} 150           # 当前连接数
websocket_messages_sent_total{} 10000        # 发送消息总数
websocket_connection_duration_seconds{} 300  # 平均连接时长
websocket_errors_total{type="auth"} 5        # 认证失败次数

# Redis Streams
redis_stream_length{stream="feed:stream"} 1234        # Stream 长度
redis_stream_pending{group="feed:consumers"} 5        # 待处理消息数
redis_stream_consumer_lag_seconds{} 2.5               # 消费延迟
redis_stream_retry_count_total{} 10                   # 重试次数
```

### 告警规则

```yaml
# 连接数异常
- alert: TooManyWebSocketConnections
  expr: websocket_active_connections > 10000
  
# 认证失败率高
- alert: HighWebSocketAuthFailure
  expr: rate(websocket_errors_total{type="auth"}[5m]) > 10

# Stream 消息堆积
- alert: RedisStreamBacklog
  expr: redis_stream_length{stream="feed:stream"} > 10000

# 待处理消息过多
- alert: RedisStreamPendingHigh
  expr: redis_stream_pending{group="feed:consumers"} > 1000

# 消费延迟过高
- alert: RedisStreamConsumerLag
  expr: redis_stream_consumer_lag_seconds > 60
```

## 测试

### 手动测试

```bash
# 使用 wscat 测试
npm install -g wscat

# 连接（需要真实 token）
wscat -c "ws://localhost:8080/ws?token=YOUR_TOKEN"

# 应该看到
Connected (press CTRL+C to quit)
```

### 压力测试

```bash
# 使用 k6 测试
k6 run websocket-load-test.js
```

## 总结

### WebSocket + Redis Streams 架构优势

✅ **实时性**：秒级推送链上交易  
✅ **可靠性**：消息持久化，不会丢失  
✅ **可扩展**：支持多实例部署，自动负载均衡  
✅ **容错性**：自动重试，超时认领  
✅ **可观测**：完善的监控指标和日志  
✅ **用户体验**：像刷 Twitter 一样流畅

### 核心价值

**让用户第一时间感知链上动态，零延迟，零丢失。**

### 技术选型对比

| 方案 | 优点 | 缺点 | 适用场景 |
|------|------|------|----------|
| **轮询** | 简单 | 延迟高，浪费资源 | 低频更新 |
| **长轮询** | 延迟低 | 服务器压力大 | 中频更新 |
| **SSE** | 单向推送简单 | 不支持双向通信 | 单向推送 |
| **WebSocket + Pub/Sub** | 实时，双向 | 消息可能丢失 | 实时性要求高 |
| **WebSocket + Streams** ✅ | 实时，可靠，可扩展 | 实现复杂 | 生产环境 |

ChainFeed 选择 **WebSocket + Redis Streams** 方案，在实时性和可靠性之间取得最佳平衡。
