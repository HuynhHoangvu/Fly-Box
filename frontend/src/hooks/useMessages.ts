import { useState, useEffect, useCallback, useRef } from 'react';
import { conversationsAPI } from '../services/api';
import { Message, MessageStatus, SenderType } from '../types/messaging';

export const useMessages = (conversationId: number | null) => {
  const [messages, setMessages] = useState<Message[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const abortControllerRef = useRef<AbortController | null>(null);

  const loadMessages = useCallback(async (id: number) => {
    // Cancel previous request to avoid race conditions
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }
    
    abortControllerRef.current = new AbortController();
    setIsLoading(true);
    setError(null);

    try {
      const res = await conversationsAPI.getMessages(String(id));
      const data = (res?.data?.data || []) as Message[];
      
      // Map basic message structure from backend to our rich Message type
      const normalized = data.map(m => ({
        ...m,
        status: MessageStatus.SENT, // Existing messages are already sent
        sender_type: m.sender_type as SenderType
      }));

      setMessages(normalized);
    } catch (err: any) {
      if (err.name !== 'CanceledError') {
        setError('Không thể tải tin nhắn');
        setMessages([]);
      }
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    if (conversationId) {
      loadMessages(conversationId);
    } else {
      setMessages([]);
    }
    
    return () => {
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }
    };
  }, [conversationId, loadMessages]);

  const addMessage = useCallback((msg: Message) => {
    setMessages((prev) => {
      // Deduplicate: check if message with same ID or client_id already exists
      const isDuplicate = prev.some(m => 
        (m.id && m.id === msg.id) || 
        (msg.client_id && m.client_id === msg.client_id)
      );
      
      if (isDuplicate) {
        // If it was an optimistic message, update its status to SENT and its final ID
        if (msg.client_id) {
          return prev.map(m => m.client_id === msg.client_id ? { ...msg, status: MessageStatus.SENT } : m);
        }
        return prev;
      }
      
      return [...prev, msg];
    });
  }, []);

  const sendNewMessage = useCallback(async (content: string) => {
    if (!conversationId) return;

    const clientId = `tmp_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    
    // 1. Optimistic Update
    const optimisticMsg: Message = {
      id: 0,
      client_id: clientId,
      conversation_id: conversationId,
      sender_type: SenderType.STAFF,
      content,
      status: MessageStatus.SENDING,
      created_at: new Date().toISOString(),
    };

    setMessages(prev => [...prev, optimisticMsg]);

    try {
      const res = await conversationsAPI.sendMessage(String(conversationId), content, clientId);
      const sentMsg = res.data.data as Message;
      
      // 2. Update optimistic message with real data
      setMessages(prev => prev.map(m => 
        m.client_id === clientId ? { ...sentMsg, status: MessageStatus.SENT } : m
      ));
    } catch (err) {
      // 3. Handle failure
      setMessages(prev => prev.map(m => 
        m.client_id === clientId ? { ...m, status: MessageStatus.FAILED } : m
      ));
      throw err;
    }
  }, [conversationId]);

  return {
    messages,
    isLoading,
    error,
    sendNewMessage,
    addMessage
  };
};
