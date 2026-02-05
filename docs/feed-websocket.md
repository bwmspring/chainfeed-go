# Phase 1.4: Feed 流与 WebSocket 实现文档

## 功能概述

Phase 1.4 实现了以下核心功能：

- ✅ Feed 流查询 API (分页/筛选)
- ✅ WebSocket 服务实现
- ✅ 连接管理与心跳机制
- ✅ Redis Pub/Sub 订阅与推送
- ✅ 用户级别的消息路由

## 架构设计

### 数据流

```
Webhook → BatchProcessor → Database (transactions + feed_items)
                         ↓
                    Redis Pub/Sub
                         ↓
                   WebSocket Hub
                         ↓
                  Connected Clients
```

### 核心组件

1. **Feed Repository** (`internal/repository/feed.go`)
   - 查询用户的 feed 流数据
   - 支持分页和关联查询

2. **WebSocket Hub** (`internal/websocket/hub.go`)
   - 管理所有 WebSocket 连接
   - 按用户 ID 路由消息
   - 实现心跳机制（ping/pong）

3. **Redis Pub/Sub Service** (`internal/service/pubsub.go`)
   - 订阅 Redis 频道 `feed:updates`
   - 将消息广播到 WebSocket Hub

4. **Batch Processor** (`internal/webhook/batch_processor.go`)
   - 批量处理交易
   - 自动创建 feed_items
   - 发布消息到 Redis

## API 接口

### 1. 获取 Feed 流

**请求**
```bash
GET /api/v1/feed?page=1&page_size=20
Authorization: Bearer <JWT_TOKEN>
```

**响应**
```json
{
  "items": [
    {
      "id": 1,
      "user_id": 1,
      "transaction_id": 100,
      "watched_address_id": 5,
      "created_at": "2026-02-05T15:00:00Z",
      "transaction": {
        "id": 100,
        "tx_hash": "0x...",
        "block_number": 12345678,
        "block_timestamp": "2026-02-05T14:59:00Z",
        "from_address": "0x...",
        "to_address": "0x...",
        "value": "1000000000000000000",
        "tx_type": "ETH",
        "token_address": "",
        "token_id": "",
        "token_symbol": "",
        "token_decimals": 0
      },
      "watched_address": {
        "id": 5,
        "address": "0x...",
        "label": "Vitalik",
        "ens_name": "vitalik.eth"
      }
    }
  ],
  "total_count": 1,
  "page": 1,
  "page_size": 20
}
```

**参数说明**
- `page`: 页码，默认 1
- `page_size`: 每页数量，默认 20，最大 100

### 2. WebSocket 连接

**连接**
```javascript
// 浏览器端示例
const token = "your_jwt_token";
const ws = new WebSocket(`ws://localhost:8080/ws?token=${token}`);

ws.onopen = () => {
  console.log("WebSocket connected");
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log("Received:", message);
  
  // message 格式:
  // {
  //   "user_id": 1,
  //   "type": "new_transaction",
  //   "payload": {
  //     "transaction": {...},
  //     "watched_address": {...}
  //   }
  // }
};

ws.onerror = (error) => {
  console.error("WebSocket error:", error);
};

ws.onclose = () => {
  console.log("WebSocket closed");
};
```

**认证方式**

WebSocket 连接需要通过 JWT token 认证，有两种方式：

1. **Query 参数** (推荐)
   ```
   ws://localhost:8080/ws?token=<JWT_TOKEN>
   ```

2. **Header** (需要客户端支持)
   ```
   Authorization: Bearer <JWT_TOKEN>
   ```

**心跳机制**

- 服务端每 54 秒发送一次 ping
- 客户端需要在 60 秒内响应 pong
- 超时自动断开连接

## 消息类型

### new_transaction

当监控的地址有新交易时推送：

```json
{
  "user_id": 1,
  "type": "new_transaction",
  "payload": {
    "transaction": {
      "id": 100,
      "tx_hash": "0x...",
      "from_address": "0x...",
      "to_address": "0x...",
      "value": "1000000000000000000",
      "tx_type": "ETH"
    },
    "watched_address": {
      "id": 5,
      "address": "0x...",
      "label": "Vitalik",
      "ens_name": "vitalik.eth"
    }
  }
}
```

## 测试流程

### 1. 启动服务

```bash
# 确保 PostgreSQL 和 Redis 运行
docker-compose up -d

# 启动服务
make run
```

### 2. 获取 JWT Token

```bash
# 1. 获取 nonce
curl -X POST http://localhost:8080/api/v1/auth/nonce \
  -H "Content-Type: application/json" \
  -d '{"wallet_address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"}'

# 2. 签名并验证（使用 MetaMask 或其他钱包）
# 3. 获取 JWT token
```

### 3. 添加监控地址

```bash
curl -X POST http://localhost:8080/api/v1/addresses \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "address": "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045",
    "label": "Vitalik"
  }'
```

### 4. 查询 Feed 流

```bash
curl http://localhost:8080/api/v1/feed?page=1&page_size=20 \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

### 5. 连接 WebSocket

使用浏览器控制台或 WebSocket 客户端：

```javascript
const token = "your_jwt_token";
const ws = new WebSocket(`ws://localhost:8080/ws?token=${token}`);

ws.onmessage = (event) => {
  console.log("New transaction:", JSON.parse(event.data));
};
```

### 6. 触发 Webhook

模拟 Alchemy Webhook 推送：

```bash
curl -X POST http://localhost:8080/webhook/alchemy \
  -H "Content-Type: application/json" \
  -H "X-Alchemy-Signature: <signature>" \
  -d @test_webhook.json
```

## 性能优化

### 批量处理

- 交易批量大小：100 条
- 最大等待时间：5 秒
- 自动刷新缓冲区

### 连接管理

- 每个用户可以有多个 WebSocket 连接
- 消息广播到用户的所有连接
- 自动清理断开的连接

### Redis Pub/Sub

- 频道：`feed:updates`
- 消息格式：JSON
- 异步处理，不阻塞主流程

## 监控指标

### 日志

服务会记录以下关键事件：

- WebSocket 连接/断开
- Redis 消息发布/订阅
- 批量处理统计
- 错误和异常

### 健康检查

```bash
curl http://localhost:8080/health
```

返回：
```json
{
  "status": "healthy",
  "time": 1738742400
}
```

## 故障排查

### WebSocket 连接失败

1. 检查 JWT token 是否有效
2. 确认服务端口是否正确
3. 查看服务日志

### 收不到实时消息

1. 确认 Redis 是否运行
2. 检查监控地址是否添加成功
3. 验证 Webhook 是否正常接收

### Feed 流为空

1. 确认已添加监控地址
2. 检查是否有相关交易
3. 查看数据库 feed_items 表

## 下一步计划

- [ ] 添加消息过滤（按交易类型、金额等）
- [ ] 实现消息持久化和离线推送
- [ ] 添加 WebSocket 重连机制
- [ ] 性能测试和优化
- [ ] 添加监控和告警

## 技术细节

### WebSocket 心跳

```go
const (
    writeWait      = 10 * time.Second  // 写超时
    pongWait       = 60 * time.Second  // pong 超时
    pingPeriod     = 54 * time.Second  // ping 间隔
    maxMessageSize = 512               // 最大消息大小
)
```

### Redis 频道

- `feed:updates`: Feed 更新消息

### 数据库索引

```sql
-- feed_items 表索引
CREATE INDEX idx_feed_items_user_id ON feed_items(user_id, created_at DESC);
CREATE INDEX idx_feed_items_transaction_id ON feed_items(transaction_id);

-- watched_addresses 表索引
CREATE INDEX idx_watched_addresses_address ON watched_addresses(address);
```

## 参考资料

- [Gorilla WebSocket](https://github.com/gorilla/websocket)
- [Redis Pub/Sub](https://redis.io/docs/manual/pubsub/)
- [Gin Web Framework](https://gin-gonic.com/)
