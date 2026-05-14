import React, { useState, useMemo } from 'react';
import { Search, Loader2 } from 'lucide-react';
import { Conversation } from '../../types/dashboard';
import { useConversations } from '../../hooks/useConversations';
import './InboxList.css';

interface InboxListProps {
  onSelectConversation: (id: number) => void;
  activeConversationId: number | null;
}

const PLATFORM_ICONS: Record<string, string> = {
  facebook: 'https://upload.wikimedia.org/wikipedia/commons/b/b8/2021_Facebook_icon.svg',
  instagram: 'https://upload.wikimedia.org/wikipedia/commons/e/e7/Instagram_logo_2016.svg',
  zalo: 'https://upload.wikimedia.org/wikipedia/commons/9/91/Icon_of_Zalo.svg',
  shopee: 'https://upload.wikimedia.org/wikipedia/commons/f/fe/Shopee_logo.svg',
  tiktok: 'https://upload.wikimedia.org/wikipedia/en/a/a9/TikTok_logo.svg'
};

const formatTime = (dateStr: string) => {
  const d = new Date(dateStr);
  const now = new Date();
  if (d.toDateString() === now.toDateString()) {
    return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  }
  return d.toLocaleDateString([], { day: '2-digit', month: '2-digit' });
};

// Memoized item component for performance
const ConversationItem = React.memo(({ 
  conv, 
  isActive, 
  onClick 
}: { 
  conv: Conversation; 
  isActive: boolean; 
  onClick: (id: number) => void 
}) => {
  // Try to determine platform from page info if available, fallback to facebook visually
  const platform = conv.page?.platform || 'facebook';
  const platformIcon = PLATFORM_ICONS[platform] || PLATFORM_ICONS['facebook'];

  return (
    <button
      className={`conversation-item ${isActive ? 'active' : ''} ${conv.unread_count > 0 ? 'unread' : ''}`}
      onClick={() => onClick(conv.id)}
      aria-selected={isActive}
      role="tab"
    >
      <div className="avatar">
        <img 
          src={conv.customer.avatar || "https://ui-avatars.com/api/?name=" + encodeURIComponent(conv.customer.name)} 
          alt={conv.customer.name} 
        />
        <img src={platformIcon} alt={platform} className="platform-badge" />
      </div>
      <div className="conversation-info">
        <div className="header">
          <span className="customer-name">{conv.customer.name}</span>
          <span className="time">{formatTime(conv.updated_at)}</span>
        </div>
        <div className="preview">
          <span className="last-message" title={conv.last_message}>{conv.last_message}</span>
          {conv.unread_count > 0 && (
            <span className="unread-badge">{conv.unread_count}</span>
          )}
        </div>
      </div>
    </button>
  );
});

ConversationItem.displayName = 'ConversationItem';

export const InboxList: React.FC<InboxListProps> = ({
  onSelectConversation,
  activeConversationId,
}) => {
  const { conversations, loading, error, search, setSearch } = useConversations();
  const [activePlatform, setActivePlatform] = useState<string>('all');

  const filteredConversations = useMemo(() => {
    if (activePlatform === 'all') return conversations;
    return conversations.filter(c => c.page?.platform === activePlatform);
  }, [conversations, activePlatform]);

  if (error) {
    return <div className="inbox-list-error">Error: {error}</div>;
  }

  return (
    <div className="inbox-list-container">
      <div className="inbox-header">
        <div className="search-bar">
          <Search size={18} color="#64748b" />
          <input
            type="text"
            placeholder="Tìm kiếm khách hàng..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
        <div className="platform-filters">
          <button 
            className={`filter-btn ${activePlatform === 'all' ? 'active' : ''}`}
            onClick={() => setActivePlatform('all')}
          >
            Tất cả
          </button>
          <button 
            className={`filter-btn ${activePlatform === 'facebook' ? 'active' : ''}`}
            onClick={() => setActivePlatform('facebook')}
          >
            Facebook
          </button>
          <button 
            className={`filter-btn ${activePlatform === 'zalo' ? 'active' : ''}`}
            onClick={() => setActivePlatform('zalo')}
          >
            Zalo OA
          </button>
        </div>
      </div>

      <div className="conversation-list" role="tablist">
        {loading && conversations.length === 0 ? (
          <div className="loading-state">
            <Loader2 size={24} className="spinner" />
          </div>
        ) : filteredConversations.length === 0 ? (
          <div className="empty-state">
            <p>Không có cuộc trò chuyện nào</p>
          </div>
        ) : (
          filteredConversations.map((conv) => (
            <ConversationItem
              key={conv.id}
              conv={conv}
              isActive={activeConversationId === conv.id}
              onClick={onSelectConversation}
            />
          ))
        )}
      </div>
    </div>
  );
};
