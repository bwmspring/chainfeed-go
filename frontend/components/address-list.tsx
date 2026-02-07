'use client';

import { useEffect, useState } from 'react';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { AddAddressDialog } from '@/components/add-address-dialog';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

interface WatchedAddress {
  id: number;
  address: string;
  label: string;
  ens_name?: string;
  created_at: string;
}

export function AddressList() {
  const [addresses, setAddresses] = useState<WatchedAddress[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [hasFetched, setHasFetched] = useState(false);

  const fetchAddresses = async () => {
    try {
      const token = localStorage.getItem('auth_token');
      console.log('Fetching addresses with token:', token ? 'exists' : 'missing');
      
      const res = await fetch(`${API_BASE}/addresses`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      
      console.log('Fetch addresses response:', res.status);
      const data = await res.json();
      console.log('Fetch addresses data:', data);
      
      if (data.code === 0) {
        setAddresses(data.data || []);
      }
    } catch (error) {
      console.error('Failed to fetch addresses:', error);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    if (!hasFetched) {
      setHasFetched(true);
      fetchAddresses();
    }
  }, [hasFetched]);

  const handleDelete = async (id: number) => {
    if (!confirm('确定要删除这个监控地址吗？')) return;

    try {
      const token = localStorage.getItem('auth_token');
      const res = await fetch(`${API_BASE}/addresses/${id}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${token}` },
      });
      const data = await res.json();
      if (data.code === 0) {
        fetchAddresses();
      }
    } catch (error) {
      console.error('Failed to delete address:', error);
    }
  };

  if (isLoading) {
    return (
      <Card className="backdrop-blur-sm bg-white/80 dark:bg-slate-900/80 border-slate-200 dark:border-slate-800 shadow-xl">
        <div className="p-6 space-y-4">
          <div className="h-8 bg-gradient-to-r from-purple-200 to-blue-200 dark:from-purple-900 dark:to-blue-900 rounded animate-pulse"></div>
          {[1, 2, 3].map((i) => (
            <div key={i} className="p-4 bg-slate-100 dark:bg-slate-800 rounded-lg animate-pulse">
              <div className="h-4 bg-slate-200 dark:bg-slate-700 rounded w-3/4 mb-2"></div>
              <div className="h-3 bg-slate-200 dark:bg-slate-700 rounded w-1/2"></div>
            </div>
          ))}
        </div>
      </Card>
    );
  }

  return (
    <Card className="backdrop-blur-sm bg-white/80 dark:bg-slate-900/80 border-slate-200 dark:border-slate-800 shadow-xl">
      <div className="p-6 space-y-4">
        {/* 标题 */}
        <div className="flex items-center justify-between">
          <h2 className="text-xl font-bold bg-gradient-to-r from-purple-600 to-blue-600 bg-clip-text text-transparent">
            监控地址
          </h2>
          <div className="flex items-center gap-2">
            <div className="w-2 h-2 rounded-full bg-green-500 animate-pulse"></div>
            <span className="text-xs text-muted-foreground">{addresses.length} 个</span>
          </div>
        </div>

        {/* 添加按钮 */}
        <AddAddressDialog onSuccess={fetchAddresses} />

        {/* 地址列表 */}
        {addresses.length === 0 ? (
          <div className="py-12 text-center">
            <div className="text-6xl mb-4">👀</div>
            <div className="text-muted-foreground mb-4">
              还没有监控任何地址
            </div>
            <p className="text-sm text-muted-foreground">
              添加地址开始追踪链上活动
            </p>
          </div>
        ) : (
          <div className="space-y-3">
            {addresses.map((addr) => (
              <div
                key={addr.id}
                className="group p-4 rounded-lg bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-800 dark:to-slate-900 border border-slate-200 dark:border-slate-700 hover:shadow-lg hover:scale-[1.02] transition-all duration-200 cursor-pointer"
              >
                <div className="flex justify-between items-start">
                  <div className="flex-1 min-w-0">
                    {/* 标签和 ENS */}
                    <div className="flex items-center gap-2 mb-2">
                      {addr.label && (
                        <span className="font-semibold text-slate-900 dark:text-slate-100 truncate">
                          {addr.label}
                        </span>
                      )}
                      {addr.ens_name && (
                        <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-purple-100 text-purple-700 dark:bg-purple-900 dark:text-purple-300">
                          {addr.ens_name}
                        </span>
                      )}
                    </div>
                    
                    {/* 地址 */}
                    <div className="flex items-center gap-2">
                      <code className="text-xs text-muted-foreground font-mono bg-slate-200 dark:bg-slate-800 px-2 py-1 rounded">
                        {addr.address.slice(0, 6)}...{addr.address.slice(-4)}
                      </code>
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          navigator.clipboard.writeText(addr.address);
                        }}
                        className="opacity-0 group-hover:opacity-100 transition-opacity text-xs text-blue-600 hover:text-blue-700 dark:text-blue-400"
                      >
                        复制
                      </button>
                    </div>
                  </div>

                  {/* 操作按钮 */}
                  <div className="flex gap-2">
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={(e) => {
                        e.stopPropagation();
                        handleDelete(addr.id);
                      }}
                      className="opacity-0 group-hover:opacity-100 transition-opacity text-red-600 hover:text-red-700 hover:bg-red-50 dark:hover:bg-red-950"
                    >
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                      </svg>
                    </Button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </Card>
  );
}
