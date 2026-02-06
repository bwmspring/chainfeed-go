'use client';

import { useState, useEffect } from 'react';
import { Card } from '@/components/ui/card';
import { FeedCard } from '@/components/feed-card';
import { useWebSocket } from '@/hooks/use-websocket';

export function FeedList() {
  const [feeds, setFeeds] = useState<any[]>([]);
  const [token, setToken] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [hasFetched, setHasFetched] = useState(false);

  useEffect(() => {
    if (hasFetched) return;
    setHasFetched(true);

    const authToken = localStorage.getItem('auth_token');
    setToken(authToken);
    
    // 获取初始 feed 数据
    if (authToken) {
      fetch(`${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1'}/feed`, {
        headers: { 'Authorization': `Bearer ${authToken}` }
      })
        .then(async res => {
          const text = await res.text();
          console.log('[FeedList] API response:', text);
          try {
            const data = JSON.parse(text);
            if (data.code === 0 && data.data?.items) {
              setFeeds(data.data.items);
            }
          } catch (e) {
            console.error('[FeedList] JSON parse error:', e, 'Response:', text);
          }
        })
        .catch(console.error)
        .finally(() => setLoading(false));
    } else {
      setLoading(false);
    }
  }, [hasFetched]);
  
  const { isConnected } = useWebSocket({
    url: process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080/ws',
    token: token || undefined,
    onMessage: (data) => {
      setFeeds((prev) => [data, ...prev].slice(0, 50));
    },
  });

  return (
    <div className="space-y-6">
      {/* 标题栏 */}
      <Card className="backdrop-blur-sm bg-white/80 dark:bg-slate-900/80 border-slate-200 dark:border-slate-800 shadow-xl">
        <div className="p-6">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-2xl font-bold bg-gradient-to-r from-purple-600 to-blue-600 bg-clip-text text-transparent">
                实时动态
              </h2>
              <p className="text-sm text-muted-foreground mt-1">
                追踪监控地址的链上活动
              </p>
            </div>
            <div className="flex items-center gap-3">
              <div className={`w-3 h-3 rounded-full ${isConnected ? 'bg-green-500 animate-pulse' : 'bg-gray-400'}`} />
              <span className="text-sm font-medium">
                {isConnected ? '实时连接' : '未连接'}
              </span>
            </div>
          </div>
        </div>
      </Card>

      {/* Feed 列表 */}
      <div className="space-y-4">
        {loading ? (
          <Card className="backdrop-blur-sm bg-white/80 dark:bg-slate-900/80 border-slate-200 dark:border-slate-800 shadow-xl">
            <div className="p-12 text-center">
              <div className="text-6xl mb-4 animate-pulse">⏳</div>
              <p className="text-muted-foreground">加载中...</p>
            </div>
          </Card>
        ) : feeds.length === 0 ? (
          <Card className="backdrop-blur-sm bg-white/80 dark:bg-slate-900/80 border-slate-200 dark:border-slate-800 shadow-xl">
            <div className="p-12 text-center">
              <div className="text-6xl mb-4">📡</div>
              <h3 className="text-lg font-semibold mb-2">等待链上活动</h3>
              <p className="text-muted-foreground mb-4">
                当监控的地址有交易时，会实时显示在这里
              </p>
              <div className="inline-flex items-center gap-2 text-sm text-muted-foreground">
                <div className={`w-2 h-2 rounded-full ${isConnected ? 'bg-green-500 animate-pulse' : 'bg-gray-400'}`} />
                <span>{isConnected ? 'WebSocket 已连接' : 'WebSocket 未连接'}</span>
              </div>
            </div>
          </Card>
        ) : (
          feeds.map((feed, index) => (
            <div
              key={feed.id || index}
              className="animate-in fade-in slide-in-from-top-4 duration-500"
              style={{ animationDelay: `${index * 50}ms` }}
            >
              <FeedCard item={feed} />
            </div>
          ))
        )}
      </div>
    </div>
  );
}
