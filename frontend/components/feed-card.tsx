'use client';

import { Card } from '@/components/ui/card';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';

interface FeedItem {
  id: string;
  type: 'transfer' | 'swap' | 'nft';
  from: string;
  to: string;
  amount?: string;
  token?: string;
  timestamp: number;
  summary?: string;
}

export function FeedCard({ item }: { item: FeedItem }) {
  return (
    <Card className="p-4">
      <div className="flex gap-3">
        <Avatar>
          <AvatarFallback>{item.from.slice(2, 4).toUpperCase()}</AvatarFallback>
        </Avatar>
        <div className="flex-1">
          <div className="flex items-center gap-2 mb-1">
            <span className="font-medium">{item.from.slice(0, 6)}...{item.from.slice(-4)}</span>
            <Badge variant="secondary">{item.type}</Badge>
          </div>
          <p className="text-sm text-muted-foreground">
            {item.summary || `Transferred ${item.amount} ${item.token} to ${item.to.slice(0, 6)}...${item.to.slice(-4)}`}
          </p>
          <span className="text-xs text-muted-foreground">
            {new Date(item.timestamp * 1000).toLocaleString('zh-CN')}
          </span>
        </div>
      </div>
    </Card>
  );
}
