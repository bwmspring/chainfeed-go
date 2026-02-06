# PostgreSQL Docker 容器操作指南

## 容器管理

```bash
# 启动容器
docker-compose up -d

# 停止容器
docker-compose down

# 查看容器状态
docker-compose ps

# 查看容器日志
docker-compose logs postgres
docker-compose logs -f postgres  # 实时查看
```

## 数据库连接

```bash
# 进入 PostgreSQL 容器
docker exec -it chainfeed-postgres psql -U chainfeed -d chainfeed

# 或使用本机 psql 客户端连接
PGPASSWORD=chainfeed psql -h localhost -p 5433 -U chainfeed -d chainfeed
```

## 基本 CRUD 操作

### 查询 (Read)

```sql
-- 查看所有表
\dt

-- 查看表结构
\d users
\d watched_addresses
\d transactions
\d feed_items

-- 查询所有用户
SELECT * FROM users;

-- 查询特定用户的监控地址
SELECT * FROM watched_addresses WHERE user_id = 1;

-- 查询最近的交易
SELECT * FROM transactions ORDER BY block_timestamp DESC LIMIT 10;

-- 联表查询用户的 feed
SELECT 
    u.wallet_address,
    wa.label,
    t.tx_hash,
    t.tx_type,
    t.value
FROM feed_items fi
JOIN users u ON fi.user_id = u.id
JOIN watched_addresses wa ON fi.watched_address_id = wa.id
JOIN transactions t ON fi.transaction_id = t.id
WHERE u.id = 1
ORDER BY fi.created_at DESC
LIMIT 20;
```

### 插入 (Create)

```sql
-- 插入用户
INSERT INTO users (wallet_address, nonce) 
VALUES ('0x1234567890123456789012345678901234567890', 'random_nonce_123');

-- 插入监控地址
INSERT INTO watched_addresses (user_id, address, label, ens_name) 
VALUES (1, '0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045', 'Vitalik', 'vitalik.eth');

-- 插入交易
INSERT INTO transactions (
    tx_hash, 
    block_number, 
    block_timestamp, 
    from_address, 
    to_address, 
    value, 
    tx_type
) VALUES (
    '0xabc123...', 
    18000000, 
    NOW(), 
    '0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045', 
    '0x1234567890123456789012345678901234567890', 
    1000000000000000000, 
    'ETH_TRANSFER'
);
```

### 更新 (Update)

```sql
-- 更新用户 nonce
UPDATE users 
SET nonce = 'new_nonce_456', updated_at = NOW() 
WHERE wallet_address = '0x1234567890123456789012345678901234567890';

-- 更新监控地址标签
UPDATE watched_addresses 
SET label = 'Vitalik Buterin', ens_name = 'vitalik.eth' 
WHERE id = 1;

-- 批量更新
UPDATE watched_addresses 
SET label = CONCAT(label, ' (Verified)') 
WHERE user_id = 1;
```

### 删除 (Delete)

```sql
-- 删除特定监控地址
DELETE FROM watched_addresses WHERE id = 1;

-- 删除用户（会级联删除相关数据）
DELETE FROM users WHERE id = 1;

-- 删除旧交易记录
DELETE FROM transactions 
WHERE block_timestamp < NOW() - INTERVAL '30 days';

-- 清空表（保留结构）
TRUNCATE TABLE feed_items CASCADE;
```

## 数据库维护

```bash
# 备份数据库
docker exec chainfeed-postgres pg_dump -U chainfeed chainfeed > backup.sql

# 恢复数据库
docker exec -i chainfeed-postgres psql -U chainfeed -d chainfeed < backup.sql

# 查看数据库大小
docker exec chainfeed-postgres psql -U chainfeed -d chainfeed -c "SELECT pg_size_pretty(pg_database_size('chainfeed'));"

# 查看表大小
docker exec chainfeed-postgres psql -U chainfeed -d chainfeed -c "SELECT tablename, pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size FROM pg_tables WHERE schemaname = 'public' ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;"
```

## 性能分析

```sql
-- 查看慢查询
SELECT query, calls, total_time, mean_time 
FROM pg_stat_statements 
ORDER BY mean_time DESC 
LIMIT 10;

-- 查看表统计信息
SELECT * FROM pg_stat_user_tables WHERE schemaname = 'public';

-- 查看索引使用情况
SELECT 
    schemaname,
    tablename,
    indexname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
ORDER BY idx_scan DESC;

-- 查看未使用的索引
SELECT 
    schemaname,
    tablename,
    indexname
FROM pg_stat_user_indexes
WHERE idx_scan = 0 
AND schemaname = 'public';
```

## 常用快捷命令

```bash
# psql 内部命令
\l          # 列出所有数据库
\dt         # 列出所有表
\d table    # 查看表结构
\du         # 列出所有用户
\q          # 退出 psql
\?          # 帮助
\timing     # 开启/关闭查询计时

# 执行 SQL 文件
\i /path/to/file.sql

# 导出查询结果到 CSV
\copy (SELECT * FROM users) TO '/tmp/users.csv' CSV HEADER;
```

## Makefile 快捷命令

```bash
# 运行迁移
make migrate

# 回滚迁移
make migrate-down

# 重置数据库
make db-reset
```
