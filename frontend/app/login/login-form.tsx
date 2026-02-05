'use client';

import { useRouter } from 'next/navigation';
import { useConnection, useSignMessage } from 'wagmi';
import { ConnectButton } from '@rainbow-me/rainbowkit';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { useAuth } from '@/hooks/use-auth';
import { useEffect } from 'react';

export function LoginForm() {
  const router = useRouter();
  const { address, isConnected } = useConnection();
  const { mutate: signMessage } = useSignMessage();
  const { login, isLoading } = useAuth();

  // 连接钱包后自动触发签名登录
  useEffect(() => {
    if (isConnected && address && signMessage && !isLoading) {
      login(address, signMessage)
        .then(() => router.push('/feed'))
        .catch((error) => console.error('Login failed:', error));
    }
  }, [isConnected, address]);

  return (
    <Card className="p-6 space-y-4">
      <ConnectButton.Custom>
        {({ account, chain, openConnectModal, mounted }) => {
          const ready = mounted;
          const connected = ready && account && chain;

          return (
            <div className="space-y-4">
              {!connected ? (
                <Button
                  onClick={openConnectModal}
                  className="w-full"
                  size="lg"
                >
                  连接钱包
                </Button>
              ) : (
                <div className="text-center space-y-2">
                  <div className="text-sm text-muted-foreground">
                    已连接: {account.address.slice(0, 6)}...{account.address.slice(-4)}
                  </div>
                  <div className="text-sm text-muted-foreground">
                    {isLoading ? '正在登录...' : '登录中...'}
                  </div>
                </div>
              )}
            </div>
          );
        }}
      </ConnectButton.Custom>
    </Card>
  );
}
