# ENS 服务完善说明

## ✅ 已完成

ENS 服务已使用 `go-ens` 库完整实现，支持：

### 功能
1. **正向解析**：ENS 名称 → 以太坊地址
   - 例如：`vitalik.eth` → `0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045`

2. **反向解析**：以太坊地址 → ENS 名称
   - 例如：`0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045` → `vitalik.eth`

### 依赖
```bash
github.com/ethereum/go-ethereum v1.16.8
github.com/wealdtech/go-ens/v3 v3.6.0
```

## 使用方式

### 在监控地址中使用 ENS

#### 1. 添加 ENS 域名
```bash
curl -X POST http://localhost:8080/api/v1/addresses \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "address": "vitalik.eth",
    "label": "Vitalik Buterin"
  }'
```

**响应：**
```json
{
  "data": {
    "id": 1,
    "user_id": 1,
    "address": "0xd8da6bf26964af9d7eed9e03e53415d37aa96045",
    "label": "Vitalik Buterin",
    "ens_name": "vitalik.eth",
    "created_at": "2026-02-05T10:20:00Z"
  }
}
```

#### 2. 添加以太坊地址（自动反向解析 ENS）
```bash
curl -X POST http://localhost:8080/api/v1/addresses \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "address": "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045",
    "label": "Vitalik"
  }'
```

**响应：**
```json
{
  "data": {
    "id": 2,
    "user_id": 1,
    "address": "0xd8da6bf26964af9d7eed9e03e53415d37aa96045",
    "label": "Vitalik",
    "ens_name": "vitalik.eth",  // 自动解析
    "created_at": "2026-02-05T10:21:00Z"
  }
}
```

## 配置要求

确保 `config/config.yaml` 中配置了有效的 Ethereum RPC：

```yaml
ethereum:
  rpc_url: "https://eth-mainnet.g.alchemy.com/v2/YOUR_API_KEY"
  chain_id: 1
  network: "mainnet"
```

### 推荐的 RPC 提供商

1. **Alchemy**（推荐）
   - 免费额度：每月 300M CU
   - 注册：https://www.alchemy.com/
   - RPC: `https://eth-mainnet.g.alchemy.com/v2/YOUR_KEY`

2. **Infura**
   - 免费额度：每天 100k 请求
   - 注册：https://infura.io/
   - RPC: `https://mainnet.infura.io/v3/YOUR_KEY`

3. **QuickNode**
   - 免费试用
   - 注册：https://www.quicknode.com/

## 实现细节

### 代码结构

```go
// internal/service/ens.go
type ENSService struct {
    client *ethclient.Client
}

// 正向解析
func (s *ENSService) Resolve(ctx context.Context, ensName string) (string, error)

// 反向解析
func (s *ENSService) ReverseResolve(ctx context.Context, address string) (string, error)
```

### 在 Handler 中的使用

```go
// internal/handler/watched_address.go
func (h *WatchedAddressHandler) Add(c *gin.Context) {
    // 判断输入是地址还是 ENS
    if common.IsHexAddress(req.Address) {
        // 是地址，尝试反向解析 ENS
        address = common.HexToAddress(req.Address).Hex()
        if name, err := h.ensService.ReverseResolve(ctx, address); err == nil {
            ensName = name
        }
    } else {
        // 是 ENS，正向解析为地址
        resolvedAddr, err := h.ensService.Resolve(ctx, req.Address)
        if err != nil {
            return error
        }
        address = resolvedAddr
        ensName = req.Address
    }
}
```

## 性能优化建议

### 1. 添加 Redis 缓存

```go
func (s *ENSService) ResolveWithCache(ctx context.Context, ensName string) (string, error) {
    // 检查缓存
    cacheKey := "ens:resolve:" + ensName
    if cached, err := s.redis.Get(ctx, cacheKey).Result(); err == nil {
        return cached, nil
    }

    // 调用 ENS 解析
    address, err := s.Resolve(ctx, ensName)
    if err != nil {
        return "", err
    }

    // 缓存 24 小时
    s.redis.Set(ctx, cacheKey, address, 24*time.Hour)
    return address, nil
}
```

### 2. 批量解析

如果需要批量解析多个 ENS，可以使用 goroutine 并发处理：

```go
func (s *ENSService) ResolveBatch(ctx context.Context, names []string) (map[string]string, error) {
    results := make(map[string]string)
    var mu sync.Mutex
    var wg sync.WaitGroup

    for _, name := range names {
        wg.Add(1)
        go func(n string) {
            defer wg.Done()
            if addr, err := s.Resolve(ctx, n); err == nil {
                mu.Lock()
                results[n] = addr
                mu.Unlock()
            }
        }(name)
    }

    wg.Wait()
    return results, nil
}
```

## 错误处理

### ENS 不存在
```go
// 正向解析失败 - 返回错误
address, err := ensService.Resolve(ctx, "nonexistent.eth")
// err: "failed to resolve ENS name: ..."
```

### 地址没有设置 ENS
```go
// 反向解析失败 - 返回空字符串（不是错误）
name, err := ensService.ReverseResolve(ctx, "0x123...")
// name: ""
// err: nil
```

## 测试

### 手动测试

```bash
# 1. 启动服务
go run cmd/server/main.go

# 2. 获取认证 token（参考 phase-1.3-quickstart.md）

# 3. 测试 ENS 解析
curl -X POST http://localhost:8080/api/v1/addresses \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"address": "vitalik.eth", "label": "Vitalik"}'

# 4. 测试反向解析
curl -X POST http://localhost:8080/api/v1/addresses \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"address": "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045", "label": "Test"}'
```

### 单元测试（待添加）

```go
// internal/service/ens_test.go
func TestENSResolve(t *testing.T) {
    service, _ := NewENSService("https://eth-mainnet.g.alchemy.com/v2/...")
    
    address, err := service.Resolve(context.Background(), "vitalik.eth")
    assert.NoError(t, err)
    assert.Equal(t, "0xd8da6bf26964af9d7eed9e03e53415d37aa96045", strings.ToLower(address))
}
```

## 常见问题

### Q: ENS 解析很慢怎么办？
A: 添加 Redis 缓存，ENS 记录通常不会频繁变化。

### Q: 支持其他域名后缀吗（如 .crypto）？
A: 当前只支持 .eth，其他域名系统需要额外集成。

### Q: RPC 调用会产生费用吗？
A: ENS 解析是读操作，不消耗 gas，但会占用 RPC 配额。

### Q: 如何处理 ENS 解析失败？
A: 
- 正向解析失败：返回错误给用户
- 反向解析失败：静默处理，ens_name 字段为空

## 下一步优化

- [ ] 添加 Redis 缓存层
- [ ] 实现批量解析
- [ ] 添加单元测试
- [ ] 监控 RPC 调用次数
- [ ] 支持 ENS 头像解析
- [ ] 支持 ENS 文本记录（Twitter、GitHub 等）

## 相关资源

- [ENS 官方文档](https://docs.ens.domains/)
- [go-ens GitHub](https://github.com/wealdtech/go-ens)
- [go-ethereum 文档](https://geth.ethereum.org/docs)
