'use client';

import { ConnectButton } from '@rainbow-me/rainbowkit';
import { Logo } from './logo';

export function Header() {
  return (
    <header className="border-b">
      <div className="container mx-auto px-4 h-16 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Logo className="w-7 h-7" />
          <h1 className="text-xl font-bold">ChainFeed</h1>
        </div>
        <ConnectButton />
      </div>
    </header>
  );
}
