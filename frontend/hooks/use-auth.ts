import { useState } from 'react';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

type SignMessageMutate = (
  args: { message: string },
  options?: {
    onSuccess?: (data: string) => void;
    onError?: (error: Error) => void;
  }
) => void;

export function useAuth() {
  const [isLoading, setIsLoading] = useState(false);

  const login = async (address: string, signMessage: SignMessageMutate) => {
    setIsLoading(true);
    try {
      // 1. 获取 nonce
      const nonceRes = await fetch(`${API_BASE}/auth/nonce`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ address }),
      });
      const nonceData = await nonceRes.json();
      if (nonceData.code !== 0) {
        throw new Error(nonceData.message || 'Failed to get nonce');
      }
      const { message } = nonceData.data;

      // 2. 签名 - 使用 Promise 包装
      const signature = await new Promise<string>((resolve, reject) => {
        signMessage(
          { message },
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
      const verifyData = await verifyRes.json();
      if (verifyData.code !== 0) {
        throw new Error(verifyData.message || 'Failed to verify signature');
      }
      const { token } = verifyData.data;

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
