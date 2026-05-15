import React, { createContext, useContext, useEffect, useRef, useCallback, useState, ReactNode } from 'react';

interface WebSocketMessage {
  type: string;
  page_id?: number;
  user_id?: number;
  data?: unknown;
  notification?: unknown;
  count?: number;
}

type MessageHandler = (msg: WebSocketMessage) => void;

interface WebSocketContextValue {
  connected: boolean;
  subscribe: (handler: MessageHandler) => () => void;
  send: (data: unknown) => void;
}

const WebSocketContext = createContext<WebSocketContextValue | null>(null);

class WebSocketManager {
  private ws: WebSocket | null = null;
  private handlers: Set<MessageHandler> = new Set();
  private reconnectTimeout: any = null;
  private userId?: number;
  private connecting: boolean = false;

  connect(userId?: number) {
    if (this.ws?.readyState === WebSocket.OPEN || this.connecting) return;
    
    this.userId = userId;
    this.connecting = true;

    const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8081';
    const wsUrl = API_URL.replace(/^http/, 'ws');
    
    let url = `${wsUrl}/ws`;
    if (userId) {
      url += `?user_id=${userId}`;
    }

    try {
      this.ws = new WebSocket(url);

      this.ws.onopen = () => {
        this.connecting = false;
        console.log('[WS] Connected');
      };

      this.ws.onmessage = (event) => {
        try {
          const msg: WebSocketMessage = JSON.parse(event.data);
          this.handlers.forEach(handler => handler(msg));
        } catch (err) {
          console.error('[WS] Failed to parse message:', err);
        }
      };

      this.ws.onclose = () => {
        this.connecting = false;
        console.log('[WS] Disconnected, reconnecting in 3s...');
        this.reconnectTimeout = setTimeout(() => {
          this.connect(this.userId);
        }, 3000);
      };

      this.ws.onerror = (err) => {
        this.connecting = false;
        console.error('[WS] Error:', err);
      };
    } catch (err) {
      this.connecting = false;
      console.error('[WS] Failed to connect:', err);
      this.reconnectTimeout = setTimeout(() => {
        this.connect(this.userId);
      }, 3000);
    }
  }

  disconnect() {
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
    }
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  subscribe(handler: MessageHandler): () => void {
    this.handlers.add(handler);
    return () => {
      this.handlers.delete(handler);
    };
  }

  send(data: unknown) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data));
    }
  }

  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }
}

// Singleton instance
const wsManager = new WebSocketManager();

export function WebSocketProvider({ children, userId }: { children: ReactNode; userId?: number }) {
  const [connected, setConnected] = useState(false);
  const checkIntervalRef = useRef<any>(null);

  useEffect(() => {
    wsManager.connect(userId);

    // Check connection status periodically
    checkIntervalRef.current = setInterval(() => {
      setConnected(wsManager.isConnected());
    }, 1000);

    return () => {
      if (checkIntervalRef.current) {
        clearInterval(checkIntervalRef.current);
      }
    };
  }, [userId]);

  const subscribe = useCallback((handler: MessageHandler) => {
    return wsManager.subscribe(handler);
  }, []);

  const send = useCallback((data: unknown) => {
    wsManager.send(data);
  }, []);

  return (
    <WebSocketContext.Provider value={{ connected, subscribe, send }}>
      {children}
    </WebSocketContext.Provider>
  );
}

export function useWebSocketContext() {
  const context = useContext(WebSocketContext);
  if (!context) {
    throw new Error('useWebSocketContext must be used within WebSocketProvider');
  }
  return context;
}

// Hook for components that only need to listen for specific message types
export function useWebSocketListener<T = any>(
  messageType: string,
  handler: (data: T) => void
) {
  const { subscribe } = useWebSocketContext();

  useEffect(() => {
    const unsubscribe = subscribe((msg) => {
      if (msg.type === messageType) {
        handler(msg.data as T);
      }
    });
    return unsubscribe;
  }, [messageType, handler, subscribe]);
}
