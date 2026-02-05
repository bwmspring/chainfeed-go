import { LoginForm } from './login-form';
import { Logo } from '@/components/logo';

export default function LoginPage() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-background">
      <div className="w-full max-w-md space-y-6 px-4">
        <div className="text-center space-y-3">
          <div className="flex justify-center">
            <Logo className="w-16 h-16" />
          </div>
          <h1 className="text-3xl font-bold">ChainFeed</h1>
          <p className="text-muted-foreground mt-2">连接钱包开始使用</p>
        </div>
        <LoginForm />
      </div>
    </div>
  );
}
