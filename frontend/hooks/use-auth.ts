import { useState } from 'react';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

type SignMessageFn = (args: { message: string }) => void;

export function useAuth() {
  const [isLoading, setIsLoading] = useState(false);

  const login = async (address: string, signMessage: SignMessageFn) => {
    setIsLoading(true);
    try {
      // 1. 获取 nonce
      const nonceRes = await fetch(`${API_BASE}/auth/nonce`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ address }),
      });
      const { nonce } = await nonceRes.json();

      // 2. 签名 - 使用 Promise 包装
      const signature = await new Promise<string>((resolve, reject) => {
        signMessage(
          { message: nonce },
          {
            onSuccess: (data) => resolve(data),
            onError: (error) => reject(error),
          }
        );
      });

      // 3. 验证签名
      const verifyRes = await fetch(`${API_BASE}/auth/verify`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ address, signature }),
      });
      const { token } = await verifyRes.json();

      // 4. 保存 token
      localStorage.setItem('auth_token', token);
    } finally {
      setIsLoading(false);
    }
  };

  const logout = () => {
    localStorage.removeItem('auth_token');
  };

  return { login, logout, isLoading };
}
