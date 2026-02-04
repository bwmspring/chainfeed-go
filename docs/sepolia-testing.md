# Sepolia 测试网络配置指南

## 1. 获取测试资金
```bash
# 访问 Sepolia 水龙头获取测试 ETH
https://sepoliafaucet.com/
https://faucet.sepolia.dev/
```

## 2. Alchemy 配置
```
1. 登录 Alchemy Dashboard
2. 创建新的 App，选择 Sepolia 网络
3. 获取 API Key
4. 创建 Notify Webhook:
   - Network: Ethereum Sepolia
   - Webhook URL: https://your-domain.com/webhooks/alchemy
   - Addresses to watch: [测试钱包地址]
```

## 3. 测试地址示例
```
测试发送地址: 0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6
测试接收地址: 0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045
```

## 4. 本地测试命令
```bash
# 启动应用 (使用 Sepolia 配置)
make run

# 使用 ngrok 暴露本地端口 (用于接收 Webhook)
ngrok http 8080

# 更新 Alchemy Webhook URL 为 ngrok 地址
https://abc123.ngrok.io/webhooks/alchemy
```

## 5. 发送测试交易
```javascript
// 使用 MetaMask 或 ethers.js 发送测试交易
const tx = await wallet.sendTransaction({
  to: "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045",
  value: ethers.utils.parseEther("0.01") // 0.01 Sepolia ETH
});
```

## 6. 验证数据接收
```bash
# 检查数据库中的交易记录
docker exec chainfeed-postgres psql -U chainfeed -d chainfeed -c "SELECT * FROM transactions ORDER BY created_at DESC LIMIT 5;"

# 查看应用日志
tail -f app.log
```
