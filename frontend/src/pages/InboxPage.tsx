import { useState, useEffect, useCallback } from 'react';
import { useWebSocket } from '../hooks/useWebSocket';
import { ChannelConnect } from '../components/Connect/ChannelConnect';
import { InboxList } from '../components/Inbox/InboxList';
import { MessagePanel } from '../components/Inbox/MessagePanel';
import { useAuthStore } from '../store/useAuthStore';
import { pagesAPI, conversationsAPI } from '../services/api';
import './InboxPage.css';

interface Customer {
  id: number;
  name: string;
  platform: string;
  avatar?: string;
}

interface Conversation {
  id: number;
  page_id: number;
  customer: Customer;
  last_message: string;
  unread_count: number;
  updated_at: string;
}

interface Message {
  id: number;
  sender_type: string;
  content_type: string;
  content: string;
  created_at: string;
}

export function InboxPage() {
  const [selectedConvId, setSelectedConvId] = useState<number | null>(null);
  const [loading, setLoading] = useState(true);
  const [hasPages, setHasPages] = useState(false);
  const user = useAuthStore(s => s.user);
  const userId = user ? Number(user.id) : 0;

  useEffect(() => {
    async function fetchData() {
      try {
        const pagesRes = await pagesAPI.list();
        const pageList = pagesRes.data.data || [];
        setHasPages(pageList.length > 0);
      } catch (err) {
        console.error('Failed to fetch inbox data:', err);
      } finally {
        setLoading(false);
      }
    }
    fetchData();
  }, []);

  if (loading) {
    return <div className="inbox-loading">Đang tải...</div>;
  }

  if (!hasPages && !loading) {
    return <ChannelConnect />;
  }

  return (
    <div className="inbox-layout">
      <InboxList 
        activeConversationId={selectedConvId} 
        onSelectConversation={setSelectedConvId} 
      />
      <MessagePanel conversationId={selectedConvId} />
    </div>
  );
}