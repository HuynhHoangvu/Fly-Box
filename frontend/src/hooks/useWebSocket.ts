import { useEffect, useRef, useCallback, useState } from 'react';

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
  const { userId, pageId, onMessage, onNewNotification, onBadgeUpdate, onNewMessage } = options;
  const [connected, setConnected] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<any>();

  const handlersRef = useRef({ onMessage, onNewNotification, onBadgeUpdate, onNewMessage });

  useEffect(() => {
    handlersRef.current = { onMessage, onNewNotification, onBadgeUpdate, onNewMessage };
  }, [onMessage, onNewNotification, onBadgeUpdate, onNewMessage]);

  const connect = useCallback(() => {
    const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8081';
    const wsUrl = API_URL.replace(/^http/, 'ws');
    
    let url = `${wsUrl}/ws`;
    const params = new URLSearchParams();
    if (userId) params.set('user_id', String(userId));
    if (pageId) params.set('page_id', String(pageId));
    if (params.toString()) url += `?${params.toString()}`;

    try {
      const ws = new WebSocket(url);
      wsRef.current = ws;

      ws.onopen = () => {
        setConnected(true);
        console.log('[WS] Connected');
      };

      ws.onmessage = (event) => {
        try {
          const msg: WebSocketMessage = JSON.parse(event.data);
          const handlers = handlersRef.current;
          
          switch (msg.type) {
            case 'NEW_NOTIFICATION':
              handlers.onNewNotification?.(msg.notification);
              break;
            case 'BADGE_UPDATE':
              handlers.onBadgeUpdate?.(msg.count ?? 0);
              break;
            case 'NEW_MESSAGE':
              handlers.onNewMessage?.(msg.data as T);
              break;
            default:
              handlers.onMessage?.(msg);
          }
        } catch (err) {
          console.error('[WS] Failed to parse message:', err);
        }
      };

      ws.onclose = () => {
        setConnected(false);
        console.log('[WS] Disconnected, reconnecting in 3s...');
        reconnectTimeoutRef.current = setTimeout(connect, 3000);
      };

      ws.onerror = (err) => {
        console.error('[WS] Error:', err);
      };
    } catch (err) {
      console.error('[WS] Failed to connect:', err);
      reconnectTimeoutRef.current = setTimeout(connect, 3000);
    }
  }, [userId, pageId]);

  useEffect(() => {
    connect();

    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, [connect]);

  const send = useCallback((data: unknown) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(data));
    }
  }, []);

  return {
    connected,
    send,
  };
}