'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAccount, useDisconnect } from 'wagmi';
import { ConnectButton } from '@rainbow-me/rainbowkit';
import { Logo } from './logo';
import { Button } from './ui/button';

export function Header() {
  const router = useRouter();
  const { address, isConnected } = useAccount();
  const { disconnect } = useDisconnect();
  const [language, setLanguage] = useState('zh');
  const [isDark, setIsDark] = useState(false);

  useEffect(() => {
    // 初始化主题状态
    setIsDark(document.documentElement.classList.contains('dark'));
  }, []);

  const handleLogout = () => {
    disconnect();
    localStorage.removeItem('auth_token');
    router.push('/');
  };

  // 监听钱包断开
  useEffect(() => {
    if (!isConnected) {
      localStorage.removeItem('auth_token');
    }
  }, [isConnected]);

  const toggleTheme = () => {
    const newIsDark = !isDark;
    setIsDark(newIsDark);
    if (newIsDark) {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  };

  return (
    <header className="sticky top-0 z-50 w-full border-b border-slate-200 dark:border-slate-800 backdrop-blur-xl bg-white/80 dark:bg-slate-900/80 shadow-sm">
      <div className="container mx-auto px-4">
        <div className="flex h-16 items-center justify-between">
          {/* 左侧：Logo */}
          <div className="flex items-center gap-8">
            <a href="/" className="flex items-center gap-3 hover:opacity-80 transition-opacity">
              <Logo className="w-8 h-8" />
              <span className="text-xl font-bold bg-gradient-to-r from-purple-600 to-blue-600 bg-clip-text text-transparent">
                ChainFeed
              </span>
            </a>

            {/* 导航链接 */}
            <nav className="hidden md:flex items-center gap-6">
              <a href="/feed" className="text-sm font-medium hover:text-purple-600 transition-colors">
                动态
              </a>
              <a href="#" className="text-sm font-medium text-muted-foreground hover:text-purple-600 transition-colors">
                探索
              </a>
            </nav>
          </div>

          {/* 右侧：操作区 */}
          <div className="flex items-center gap-4">
            {/* 主题切换 */}
            <button
              onClick={toggleTheme}
              className="hidden sm:flex items-center justify-center w-10 h-10 rounded-lg bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 transition-colors"
              title={isDark ? '切换到浅色模式' : '切换到深色模式'}
            >
              <span className="text-xl">
                {isDark ? '🌙' : '☀️'}
              </span>
            </button>

            {/* 语言选择 */}
            <select
              value={language}
              onChange={(e) => setLanguage(e.target.value)}
              className="hidden sm:block px-3 py-1.5 rounded-lg bg-slate-100 dark:bg-slate-800 text-sm font-medium border-none focus:ring-2 focus:ring-purple-500 cursor-pointer"
            >
              <option value="zh">🇨🇳 中文</option>
              <option value="en">🇺🇸 English</option>
              <option value="ja">🇯🇵 日本語</option>
            </select>

            {/* 钱包连接 */}
            {isConnected ? (
              <div className="flex items-center gap-2">
                <div className="hidden sm:flex items-center gap-2 px-3 py-1.5 rounded-lg bg-gradient-to-r from-purple-100 to-blue-100 dark:from-purple-900 dark:to-blue-900">
                  <div className="w-2 h-2 rounded-full bg-green-500 animate-pulse"></div>
                  <span className="text-sm font-medium">
                    {address?.slice(0, 6)}...{address?.slice(-4)}
                  </span>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleLogout}
                  className="text-red-600 hover:text-red-700 hover:bg-red-50 dark:hover:bg-red-950"
                >
                  断开
                </Button>
              </div>
            ) : (
              <ConnectButton.Custom>
                {({ openConnectModal }) => (
                  <Button
                    onClick={openConnectModal}
                    className="bg-gradient-to-r from-purple-600 to-blue-600 hover:from-purple-700 hover:to-blue-700"
                  >
                    连接钱包
                  </Button>
                )}
              </ConnectButton.Custom>
            )}
          </div>
        </div>
      </div>
    </header>
  );
}
