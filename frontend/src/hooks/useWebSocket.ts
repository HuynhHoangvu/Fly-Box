import { useEffect, useCallback } from 'react';
import { useWebSocketContext } from '../contexts/WebSocketContext';

interface WebSocketMessage {
  type: string;
  page_id?: number;
  user_id?: number;
  data?: unknown;
  notification?: unknown;
  count?: number;
}

interface UseWebSocketOptions<T = any> {
  userId?: number;
  pageId?: number;
  onMessage?: (msg: WebSocketMessage) => void;
  onNewNotification?: (notification: any) => void;
  onBadgeUpdate?: (count: number) => void;
  onNewMessage?: (msg: T) => void;
}

export function useWebSocket<T = any>(options: UseWebSocketOptions<T> = {}) {
  const { onMessage, onNewNotification, onBadgeUpdate, onNewMessage } = options;
  const { connected, subscribe } = useWebSocketContext();

  useEffect(() => {
    const unsubscribe = subscribe((msg: WebSocketMessage) => {
      // Call specific handlers based on message type
      switch (msg.type) {
        case 'NEW_NOTIFICATION':
          onNewNotification?.(msg.notification);
          break;
        case 'BADGE_UPDATE':
          onBadgeUpdate?.(msg.count ?? 0);
          break;
        case 'NEW_MESSAGE':
          onNewMessage?.(msg.data as T);
          break;
        default:
          onMessage?.(msg);
      }
    });

    return unsubscribe;
  }, [subscribe, onMessage, onNewNotification, onBadgeUpdate, onNewMessage]);

  const send = useCallback((data: unknown) => {
    // Use context send if available
    const context = useWebSocketContext;
    // This will be handled by the context
  }, []);

  return {
    connected,
    send,
  };
}
