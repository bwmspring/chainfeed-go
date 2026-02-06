# Redis Streams 快速参考

## 容器管理

```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 停止所有服务
docker-compose down

# 查看日志
docker-compose logs -f redis
docker-compose logs -f postgres
```

## Redis 监控命令

### 基础操作

```bash
# 进入 Redis 容器
docker-compose exec redis redis-cli

# 测试连接
docker-compose exec redis redis-cli ping
```

### Stream 监控

```bash
# 查看 Stream 基本信息
docker-compose exec redis redis-cli XINFO STREAM feed:stream

# 查看消费者组信息
docker-compose exec redis redis-cli XINFO GROUPS feed:stream

# 查看消费者详情
docker-compose exec redis redis-cli XINFO CONSUMERS feed:stream feed:consumers

# 查看待处理消息（Pending）
docker-compose exec redis redis-cli XPENDING feed:stream feed:consumers

# 查看待处理消息详情
docker-compose exec redis redis-cli XPENDING feed:stream feed:consumers - + 10

# 查看消息总数
docker-compose exec redis redis-cli XLEN feed:stream

# 查看最新 10 条消息
docker-compose exec redis redis-cli XRANGE feed:stream - + COUNT 10

# 查看最旧 10 条消息
docker-compose exec redis redis-cli XRANGE feed:stream + - COUNT 10
```

### 调试命令

```bash
# 手动发送测试消息
docker-compose exec redis redis-cli XADD feed:stream '*' user_id 1 type test payload '{"test":"data"}'

# 删除 Stream（慎用）
docker-compose exec redis redis-cli DEL feed:stream

# 删除消费者组（慎用）
docker-compose exec redis redis-cli XGROUP DESTROY feed:stream feed:consumers

# 重置消费者组到起始位置
docker-compose exec redis redis-cli XGROUP SETID feed:stream feed:consumers 0
```

## 常见问题排查

### 1. 消息堆积

```bash
# 查看消息数量
docker-compose exec redis redis-cli XLEN feed:stream

# 查看待处理消息
docker-compose exec redis redis-cli XPENDING feed:stream feed:consumers

# 如果消息过多，检查消费者是否正常运行
docker-compose logs -f app
```

### 2. 消费者卡住

```bash
# 查看消费者状态
docker-compose exec redis redis-cli XINFO CONSUMERS feed:stream feed:consumers

# 查看 pending 消息的详细信息
docker-compose exec redis redis-cli XPENDING feed:stream feed:consumers - + 10

# 如果有消息长时间 pending，检查应用日志
docker-compose logs -f app | grep "failed to process"
```

### 3. 消息丢失

```bash
# 检查 Stream 是否存在
docker-compose exec redis redis-cli EXISTS feed:stream

# 检查消费者组是否存在
docker-compose exec redis redis-cli XINFO GROUPS feed:stream

# 检查应用是否正常启动
docker-compose logs app | grep "consumer group initialized"
```

## 性能监控

```bash
# 查看 Redis 内存使用
docker-compose exec redis redis-cli INFO memory

# 查看 Redis 统计信息
docker-compose exec redis redis-cli INFO stats

# 实时监控 Redis 命令
docker-compose exec redis redis-cli MONITOR

# 查看慢查询
docker-compose exec redis redis-cli SLOWLOG GET 10
```

## 生产环境建议

### 1. Stream 清理策略

```bash
# 保留最近 10000 条消息（定期执行）
docker-compose exec redis redis-cli XTRIM feed:stream MAXLEN ~ 10000
```

### 2. 监控指标

- Stream 长度：`XLEN feed:stream`
- Pending 消息数：`XPENDING feed:stream feed:consumers`
- 消费者数量：`XINFO CONSUMERS feed:stream feed:consumers`

### 3. 告警阈值

- Stream 长度 > 10000：消费能力不足
- Pending 消息 > 1000：处理速度慢
- 消息 idle 时间 > 60s：可能有消费者卡住

## 日志查看

```bash
# 查看应用日志
docker-compose logs -f app

# 查看最近 100 行
docker-compose logs --tail=100 app

# 查看错误日志
docker-compose logs app | grep ERROR

# 查看 Stream 相关日志
docker-compose logs app | grep "stream\|consumer"
```

## 快速测试

```bash
# 运行测试脚本
./test_streams.sh

# 启动服务
make run

# 查看实时日志
docker-compose logs -f app | grep "request completed"
```
