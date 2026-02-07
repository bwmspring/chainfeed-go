'use client';

import { useState, useEffect, useRef } from 'react';
import { Card } from '@/components/ui/card';
import { FeedCard } from '@/components/feed-card';
import { useWebSocket } from '@/hooks/use-websocket';

interface FeedItem {
  id: number;
  created_at: string;
  [key: string]: any;
}

export function FeedList() {
  const [feeds, setFeeds] = useState<FeedItem[]>([]);
  const [token, setToken] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [hasMore, setHasMore] = useState(true);
  const [page, setPage] = useState(1);
  const [hasFetched, setHasFetched] = useState(false);
  const feedIdsRef = useRef(new Set<number>());

  const fetchFeeds = async (pageNum: number, append = false) => {
    const authToken = localStorage.getItem('auth_token');
    if (!authToken) return;

    try {
      const res = await fetch(
        `${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1'}/feed?page=${pageNum}&page_size=10`,
        { headers: { 'Authorization': `Bearer ${authToken}` } }
      );
      
      const text = await res.text();
      const data = JSON.parse(text);
      
      if (data.code === 0 && data.data?.items) {
        const items = data.data.items;
        items.forEach((item: FeedItem) => feedIdsRef.current.add(item.id));
        
        if (append) {
          setFeeds(prev => [...prev, ...items]);
        } else {
          setFeeds(items);
        }
        
        setHasMore(items.length === 10);
      }
    } catch (e) {
      console.error('[FeedList] Fetch error:', e);
    }
  };

  useEffect(() => {
    if (hasFetched) return;
    setHasFetched(true);

    const authToken = localStorage.getItem('auth_token');
    setToken(authToken);

    if (authToken) {
      fetchFeeds(1).finally(() => setLoading(false));
    } else {
      setLoading(false);
    }
  }, [hasFetched]);

  const loadMore = async () => {
    if (loadingMore || !hasMore) return;
    
    setLoadingMore(true);
    const nextPage = page + 1;
    await fetchFeeds(nextPage, true);
    setPage(nextPage);
    setLoadingMore(false);
  };

  const { isConnected } = useWebSocket({
    url: process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080/ws',
    token: token || undefined,
    onMessage: (data: FeedItem) => {
      console.log('[FeedList] WebSocket message received:', data);
      
      // 去重：如果已存在则忽略
      if (feedIdsRef.current.has(data.id)) {
        console.log('[FeedList] Duplicate feed item ignored:', data.id);
        return;
      }

      feedIdsRef.current.add(data.id);

      setFeeds((prev) => {
        const updated = [data, ...prev];
        // 按 created_at 降序排序（最新在前）
        return updated
          .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
          .slice(0, 50); // 只保留最新 50 条
      });
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
                添加监控地址后，会自动加载最近 20 笔交易
              </p>
              <div className="inline-flex items-center gap-2 text-sm text-muted-foreground">
                <div className={`w-2 h-2 rounded-full ${isConnected ? 'bg-green-500 animate-pulse' : 'bg-gray-400'}`} />
                <span>{isConnected ? 'WebSocket 已连接' : 'WebSocket 未连接'}</span>
              </div>
            </div>
          </Card>
        ) : (
          <>
            {feeds.map((feed, index) => (
              <div
                key={feed.id}
                className="animate-in fade-in slide-in-from-top-4 duration-500"
                style={{ animationDelay: `${index * 50}ms` }}
              >
                <FeedCard item={feed} />
              </div>
            ))}
            
            {/* 加载更多按钮 */}
            {hasMore && (
              <div className="flex justify-center pt-4">
                <button
                  onClick={loadMore}
                  disabled={loadingMore}
                  className="px-6 py-3 rounded-lg bg-gradient-to-r from-purple-600 to-blue-600 hover:from-purple-700 hover:to-blue-700 text-white font-medium disabled:opacity-50 disabled:cursor-not-allowed transition-all"
                >
                  {loadingMore ? '加载中...' : '加载更多'}
                </button>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}
