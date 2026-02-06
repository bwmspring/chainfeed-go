'use client';

import { Card } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';

interface Transaction {
  tx_hash: string;
  from_address: string;
  to_address: string;
  value: string;
  tx_type: string;
  block_timestamp: string;
}

interface WatchedAddress {
  address: string;
  label?: string;
}

interface FeedItem {
  id: number;
  transaction?: Transaction;
  Transaction?: Transaction; // 兼容后端大写
  watched_address?: WatchedAddress;
  WatchedAddress?: WatchedAddress;
}

function formatETH(weiValue: string): string {
  try {
    const wei = BigInt(weiValue);
    const eth = Number(wei) / 1e18;
    return eth.toFixed(4);
  } catch {
    return '0.0000';
  }
}

function getDirection(tx: Transaction, watchedAddr: string): 'send' | 'receive' {
  return tx.from_address.toLowerCase() === watchedAddr.toLowerCase() ? 'send' : 'receive';
}

export function FeedCard({ item }: { item: FeedItem }) {
  const tx = item.transaction || item.Transaction;
  const watched_address = item.watched_address || item.WatchedAddress;
  
  if (!tx || !watched_address) {
    console.error('[FeedCard] Missing data:', { tx, watched_address, item });
    return null;
  }
  
  const direction = getDirection(tx, watched_address.address);
  const ethAmount = formatETH(tx.value);
  
  return (
    <Card className="backdrop-blur-sm bg-white/80 dark:bg-slate-900/80 border-slate-200 dark:border-slate-800 shadow-lg hover:shadow-xl transition-all">
      <div className="p-5">
        <div className="flex items-start gap-4">
          {/* 图标 */}
          <div className="text-3xl">
            {direction === 'send' ? '📤' : '📥'}
          </div>
          
          <div className="flex-1 min-w-0">
            {/* 标题行 */}
            <div className="flex items-center gap-2 mb-2">
              <span className="font-semibold text-lg">
                {direction === 'send' ? '发送' : '接收'} ETH
              </span>
              <Badge variant="secondary" className="text-xs">💎 {tx.tx_type}</Badge>
            </div>
            
            {/* 金额 */}
            <div className="text-2xl font-bold bg-gradient-to-r from-purple-600 to-blue-600 bg-clip-text text-transparent mb-3">
              {ethAmount} ETH
            </div>
            
            {/* 地址信息 */}
            <div className="space-y-1.5 text-sm">
              <div className="flex items-center gap-2">
                <span className="text-muted-foreground w-12">从:</span>
                <code className="text-xs bg-slate-100 dark:bg-slate-800 px-2 py-1 rounded">
                  {tx.from_address.slice(0, 10)}...{tx.from_address.slice(-8)}
                </code>
              </div>
              <div className="flex items-center gap-2">
                <span className="text-muted-foreground w-12">到:</span>
                <code className="text-xs bg-slate-100 dark:bg-slate-800 px-2 py-1 rounded">
                  {tx.to_address.slice(0, 10)}...{tx.to_address.slice(-8)}
                </code>
              </div>
            </div>
            
            {/* 底部信息 */}
            <div className="flex items-center justify-between mt-4 pt-3 border-t border-slate-200 dark:border-slate-700">
              <span className="text-xs text-muted-foreground">
                {tx.block_timestamp.replace('T', ' ').replace('Z', '').slice(0, 19)}
              </span>
              <a
                href={`https://etherscan.io/tx/${tx.tx_hash}`}
                target="_blank"
                rel="noopener noreferrer"
                className="text-xs text-blue-600 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300 font-medium"
              >
                查看详情 →
              </a>
            </div>
          </div>
        </div>
      </div>
    </Card>
  );
}
