'use client';

import Link from 'next/link';
import { Button } from '@/components/ui/button';
import { ArrowRight, Activity, Shield, Zap } from 'lucide-react';
import { Logo } from '@/components/logo';

export function Hero() {
  return (
    <div className="min-h-screen bg-gradient-to-b from-background to-muted/20">
      {/* Header */}
      <header className="border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <div className="container mx-auto px-4 h-16 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Logo className="w-8 h-8" />
            <span className="text-xl font-bold">ChainFeed</span>
          </div>
          <Link href="/login">
            <Button>启动应用</Button>
          </Link>
        </div>
      </header>

      {/* Hero Section */}
      <main className="container mx-auto px-4">
        <div className="flex flex-col items-center text-center pt-20 pb-16 space-y-8">
          <div className="space-y-4 max-w-3xl">
            <h1 className="text-5xl md:text-6xl font-bold tracking-tight">
              像刷 Twitter 一样
              <br />
              <span className="bg-gradient-to-r from-blue-600 to-cyan-600 bg-clip-text text-transparent">
                追踪链上活动
              </span>
            </h1>
            <p className="text-xl text-muted-foreground max-w-2xl mx-auto">
              专注于 Web3 链上数据的实时信息流平台，让复杂的链上交易触手可及
            </p>
          </div>

          <div className="flex gap-4">
            <Link href="/login">
              <Button size="lg" className="gap-2">
                立即开始 <ArrowRight className="w-4 h-4" />
              </Button>
            </Link>
          </div>

          {/* Features */}
          <div className="grid md:grid-cols-3 gap-8 pt-16 max-w-5xl">
            <div className="space-y-3">
              <div className="w-12 h-12 rounded-lg bg-blue-500/10 flex items-center justify-center">
                <Zap className="w-6 h-6 text-blue-600" />
              </div>
              <h3 className="text-lg font-semibold">实时推送</h3>
              <p className="text-sm text-muted-foreground">
                WebSocket 长连接，秒级感知链上动态
              </p>
            </div>

            <div className="space-y-3">
              <div className="w-12 h-12 rounded-lg bg-green-500/10 flex items-center justify-center">
                <Activity className="w-6 h-6 text-green-600" />
              </div>
              <h3 className="text-lg font-semibold">AI 解析</h3>
              <p className="text-sm text-muted-foreground">
                LLMs 自动解析复杂交易，生成人类可读摘要
              </p>
            </div>

            <div className="space-y-3">
              <div className="w-12 h-12 rounded-lg bg-purple-500/10 flex items-center justify-center">
                <Shield className="w-6 h-6 text-purple-600" />
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
