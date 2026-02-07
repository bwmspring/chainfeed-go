'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { FeedList } from './feed-list';
import { AddressList } from '@/components/address-list';

export default function FeedPage() {
  const router = useRouter();
  const [isAuthorized, setIsAuthorized] = useState(false);

  useEffect(() => {
    const token = localStorage.getItem('auth_token');
    if (!token) {
      router.replace('/');
      return;
    }
    
    // 简单验证：只检查 token 存在，不发请求
    // 如果 token 过期，API 调用会自然失败，前端可以统一处理 401
    setIsAuthorized(true);
  }, [router]);

  if (!isAuthorized) {
    return null; // 不显示任何内容，直接跳转
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-purple-50 to-blue-50 dark:from-slate-950 dark:via-purple-950 dark:to-blue-950">
      <div className="container mx-auto px-4 py-8 max-w-7xl">
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
          {/* 左侧：监控地址列表 */}
          <div className="lg:col-span-4">
            <div className="sticky top-8">
              <AddressList />
            </div>
          </div>

          {/* 右侧：Feed 流 */}
          <div className="lg:col-span-8">
            <FeedList />
          </div>
        </div>
      </div>
    </div>
  );
}
