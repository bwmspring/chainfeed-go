'use client';

import { useState, useEffect } from 'react';
import { Header } from '@/components/header';
import { FeedCard } from '@/components/feed-card';
import { useWebSocket } from '@/hooks/use-websocket';

export function FeedList() {
  const [feeds, setFeeds] = useState<any[]>([]);
  const [token, setToken] = useState<string | null>(null);

  useEffect(() => {
    // 从 localStorage 获取 token
    const authToken = localStorage.getItem('auth_token');
    setToken(authToken);
  }, []);
  
  const { isConnected } = useWebSocket({
    url: process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080/ws',
    token: token || undefined,
    onMessage: (data) => {
      setFeeds((prev) => [data, ...prev].slice(0, 50));
    },
  });

  return (
    <>
      <Header />
      <main className="container mx-auto px-4 py-6 max-w-2xl">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold">实时动态</h2>
          <div className="flex items-center gap-2">
            <div className={`w-2 h-2 rounded-full ${isConnected ? 'bg-green-500' : 'bg-gray-400'}`} />
            <span className="text-sm text-muted-foreground">
              {isConnected ? '已连接' : '未连接'}
            </span>
          </div>
        </div>
        
        <div className="space-y-3">
          {feeds.length === 0 ? (
            <div className="text-center py-12 text-muted-foreground">
              <p>暂无动态</p>
              <p className="text-sm mt-2">添加监控地址开始使用</p>
            </div>
          ) : (
            feeds.map((feed) => <FeedCard key={feed.id} item={feed} />)
          )}
        </div>
      </main>
    </>
  );
}
