'use client';

import { useRouter } from 'next/navigation';
import { Activity, Shield, Zap } from 'lucide-react';
import { useEffect } from 'react';

export function Hero() {
  const router = useRouter();

  // 检查是否已登录并跳转
  useEffect(() => {
    const token = localStorage.getItem('auth_token');
    if (token) {
      router.push('/feed');
    }
  }, [router]);

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-purple-50 to-blue-50 dark:from-slate-950 dark:via-purple-950 dark:to-blue-950">
      {/* Hero Section */}
      <main className="container mx-auto px-4">
        <div className="flex flex-col items-center text-center pt-20 pb-16 space-y-12">
          {/* 标题和描述 */}
          <div className="space-y-6 max-w-3xl">
            <h1 className="text-5xl md:text-6xl font-bold tracking-tight">
              像刷 Twitter 一样
              <br />
              <span className="bg-gradient-to-r from-purple-600 to-blue-600 bg-clip-text text-transparent">
                追踪链上活动
              </span>
            </h1>
            <p className="text-xl text-muted-foreground max-w-2xl mx-auto">
              专注于 Web3 链上数据的实时信息流平台，让复杂的链上交易触手可及
            </p>
          </div>

          {/* 使用步骤引导 */}
          <div className="w-full max-w-4xl">
            <div className="bg-white/80 dark:bg-slate-900/80 backdrop-blur-sm border border-slate-200 dark:border-slate-800 rounded-2xl p-8 shadow-xl">
              <h2 className="text-2xl font-bold mb-8 bg-gradient-to-r from-purple-600 to-blue-600 bg-clip-text text-transparent">
                三步开始使用
              </h2>

              <div className="grid md:grid-cols-3 gap-6">
                {/* 步骤 1 */}
                <div className="relative">
                  <div className="absolute -top-4 -left-4 w-12 h-12 rounded-full bg-gradient-to-r from-purple-600 to-blue-600 flex items-center justify-center text-white font-bold text-xl shadow-lg">
                    1
                  </div>
                  <div className="pt-6 space-y-3">
                    <div className="text-4xl mb-2">🔗</div>
                    <h3 className="text-lg font-semibold">连接钱包</h3>
                    <p className="text-sm text-muted-foreground">
                      点击右上角"连接钱包"按钮，使用 MetaMask 等钱包登录
                    </p>
                  </div>
                </div>

                {/* 步骤 2 */}
                <div className="relative">
                  <div className="absolute -top-4 -left-4 w-12 h-12 rounded-full bg-gradient-to-r from-purple-600 to-blue-600 flex items-center justify-center text-white font-bold text-xl shadow-lg">
                    2
                  </div>
                  <div className="pt-6 space-y-3">
                    <div className="text-4xl mb-2">👀</div>
                    <h3 className="text-lg font-semibold">添加监控地址</h3>
                    <p className="text-sm text-muted-foreground">
                      输入你想追踪的以太坊地址或 ENS 域名，如 vitalik.eth
                    </p>
                  </div>
                </div>

                {/* 步骤 3 */}
                <div className="relative">
                  <div className="absolute -top-4 -left-4 w-12 h-12 rounded-full bg-gradient-to-r from-purple-600 to-blue-600 flex items-center justify-center text-white font-bold text-xl shadow-lg">
                    3
                  </div>
                  <div className="pt-6 space-y-3">
                    <div className="text-4xl mb-2">📡</div>
                    <h3 className="text-lg font-semibold">实时追踪</h3>
                    <p className="text-sm text-muted-foreground">
                      当监控的地址有交易时，会实时推送到你的动态流
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Features */}
          <div className="grid md:grid-cols-3 gap-8 pt-8 max-w-5xl">
            <div className="space-y-3 p-6 rounded-xl bg-white/80 dark:bg-slate-900/80 backdrop-blur-sm border border-slate-200 dark:border-slate-800 hover:shadow-xl transition-shadow">
              <div className="w-12 h-12 rounded-lg bg-gradient-to-br from-purple-500 to-blue-500 flex items-center justify-center">
                <Zap className="w-6 h-6 text-white" />
              </div>
              <h3 className="text-lg font-semibold">实时推送</h3>
              <p className="text-sm text-muted-foreground">
                WebSocket 长连接，秒级感知链上动态
              </p>
            </div>

            <div className="space-y-3 p-6 rounded-xl bg-white/80 dark:bg-slate-900/80 backdrop-blur-sm border border-slate-200 dark:border-slate-800 hover:shadow-xl transition-shadow">
              <div className="w-12 h-12 rounded-lg bg-gradient-to-br from-green-500 to-emerald-500 flex items-center justify-center">
                <Activity className="w-6 h-6 text-white" />
              </div>
              <h3 className="text-lg font-semibold">AI 解析</h3>
              <p className="text-sm text-muted-foreground">
                LLMs 自动解析复杂交易，生成人类可读摘要
              </p>
            </div>

            <div className="space-y-3 p-6 rounded-xl bg-white/80 dark:bg-slate-900/80 backdrop-blur-sm border border-slate-200 dark:border-slate-800 hover:shadow-xl transition-shadow">
              <div className="w-12 h-12 rounded-lg bg-gradient-to-br from-purple-500 to-pink-500 flex items-center justify-center">
                <Shield className="w-6 h-6 text-white" />
              </div>
              <h3 className="text-lg font-semibold">专注以太坊</h3>
              <p className="text-sm text-muted-foreground">
                深度解析以太坊主网，提供最精准的监控
              </p>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}
