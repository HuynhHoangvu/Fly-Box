import { useState, useEffect, useCallback, useMemo } from 'react';
import { conversationsAPI } from '../services/api';
import { Conversation, Message, SenderType } from '../types/messaging';

export const useConversations = (userId?: number) => {
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [activeTab, setActiveTab] = useState('all');

  const loadConversations = useCallback(async (isSilent = false) => {
    if (!isSilent) setIsLoading(true);
    setError(null);
    try {
      const res = await conversationsAPI.list();
      const data = (res?.data?.data || []) as Conversation[];
      
      // Sort by updated_at descending
      const sorted = [...data].sort((a, b) => 
        new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime()
      );
      
      setConversations(sorted);
    } catch (err: any) {
      setError('Không thể tải danh sách hội thoại');
      setConversations([]);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    loadConversations();
  }, [loadConversations]);

  // Real-time updates handler
  const handleNewMessage = useCallback((msg: Message) => {
    setConversations((prev) => {
      const existingIdx = prev.findIndex(c => c.id === msg.conversation_id);
      
      if (existingIdx !== -1) {
        const updatedList = [...prev];
        const existing = updatedList[existingIdx];
        
        // Update the conversation
        const updated = {
          ...existing,
          last_message: msg.content,
          updated_at: msg.created_at,
          unread_count: msg.sender_type === SenderType.CUSTOMER 
            ? existing.unread_count + 1 
            : existing.unread_count
        };
        
        // Remove from current position and move to top
        updatedList.splice(existingIdx, 1);
        return [updated, ...updatedList];
      }
      
      // If conversation not in list, reload silently
      loadConversations(true);
      return prev;
    });
  }, [loadConversations]);

  // Filtered & Searched list
  const filteredConversations = useMemo(() => {
    return conversations.filter(c => {
      const matchesSearch = c.customer.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
                          c.last_message.toLowerCase().includes(searchQuery.toLowerCase());
      
      if (!matchesSearch) return false;
      
      if (activeTab === 'unassigned') return c.unread_count > 0; // Simple example filter
      return true;
    });
  }, [conversations, searchQuery, activeTab]);

  return {
    conversations: filteredConversations,
    totalCount: conversations.length,
    isLoading,
    error,
    searchQuery,
    setSearchQuery,
    activeTab,
    setActiveTab,
    refresh: loadConversations,
    handleNewMessage
  };
};
