# Alchemy Address Activity 配置指南

## 1. Alchemy 配置确认

### ✅ Address Activity 是正确的选择
- **用途**: 监控特定地址的所有活动
- **触发条件**: 当监控地址作为发送方或接收方时触发
- **数据完整性**: 包含完整的交易详情

### Alchemy Dashboard 配置步骤
```
1. 创建 Notify Webhook
2. 选择 "Address Activity" 类型
3. 添加要监控的地址列表
4. 设置 Webhook URL: https://your-domain.com/webhooks/alchemy
5. 配置签名密钥
```

## 2. Alchemy 真实数据格式

### Address Activity Event Payload
```json
{
  "webhookId": "wh_abc123def456",
  "id": "whevt_789xyz012",
  "createdAt": "2023-10-01T12:00:00.000Z",
  "type": "ADDRESS_ACTIVITY",
  "event": {
    "network": "ETH_MAINNET",
    "activity": [
      {
        "fromAddress": "0x742d35cc6634c0532925a3b8d4c9db96c4b4d8b6",
        "toAddress": "0xd8da6bf26964af9d7eed9e03e53415d37aa96045",
        "blockNum": "0x1234567",
        "hash": "0xabcdef...",
        "value": 1.5,
        "asset": "ETH",
        "category": "external",
        "rawContract": {
          "value": "0x14d1120d7b160000",
          "address": null,
          "decimals": 18
        }
      }
    ]
  }
}
```

### ERC20 Token Transfer
```json
{
  "activity": [
    {
      "fromAddress": "0x...",
      "toAddress": "0x...",
      "blockNum": "0x1234567",
      "hash": "0x...",
      "value": 1000,
      "asset": "USDC",
      "category": "erc20",
      "rawContract": {
        "value": "0x3b9aca00",
        "address": "0xa0b86a33e6441b8c4505b8c4505b8c4505b8c450",
        "decimals": 6
      }
    }
  ]
}
```

## 3. 高吞吐量优化策略

### 3.1 异步处理架构
```
Webhook 接收 (同步) → 立即返回 200 OK
    ↓
后台处理 (异步) → 解析 + 批量存储
```

### 3.2 批量处理优化
- **批量大小**: 100 条交易/批次
- **最大等待**: 5 秒自动刷新
- **内存缓冲**: 减少数据库 I/O

### 3.3 数据库优化
```sql
-- 批量插入优化
INSERT INTO transactions (...) VALUES 
  (...), (...), (...) -- 批量插入
ON CONFLICT (tx_hash) DO NOTHING;

-- 索引优化
CREATE INDEX CONCURRENTLY idx_tx_hash ON transactions(tx_hash);
CREATE INDEX CONCURRENTLY idx_addresses ON transactions(from_address, to_address);
```

### 3.4 连接池配置
```yaml
database:
  max_open_conns: 50    # 增加连接数
  max_idle_conns: 10    # 保持空闲连接
  conn_max_lifetime: 1h # 连接复用
```

## 4. 性能监控指标

### 关键指标
- **吞吐量**: 每秒处理 Webhook 数量
- **延迟**: 从接收到存储的时间
- **错误率**: 处理失败的比例
- **队列深度**: 待处理的交易数量

### 监控端点
```bash
# 获取性能统计
curl http://localhost:8080/monitoring/stats

# 响应示例
{
  "requests_total": 1500,
  "errors_total": 3,
  "avg_processing_time": "15ms",
  "error_rate": 0.002
}
```

## 5. 生产环境部署建议

### 5.1 负载均衡
```
Alchemy → Load Balancer → Multiple App Instances
                       → Shared Database
```

### 5.2 容错机制
- **重试策略**: Alchemy 自动重试失败的 Webhook
- **死信队列**: 处理失败的交易进入重试队列
- **熔断器**: 数据库故障时暂停处理

### 5.3 扩展性
- **水平扩展**: 多个应用实例
- **数据库分片**: 按时间或地址分片
- **缓存层**: Redis 缓存热点数据

## 6. 预期性能指标

### 单实例性能
- **吞吐量**: 1000+ Webhooks/秒
- **延迟**: < 50ms (P95)
- **内存使用**: < 100MB
- **CPU 使用**: < 30%

### 集群性能
- **吞吐量**: 10000+ Webhooks/秒
- **可用性**: 99.9%+
- **数据一致性**: 强一致性保证
