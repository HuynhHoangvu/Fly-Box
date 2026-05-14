import React, { useMemo } from 'react';
import { Conversation } from '../../types/messaging';
import { AppIcon, IconKey } from '../common/AppIcon';

interface ConversationItemProps {
  conversation: Conversation;
  isSelected: boolean;
  onSelect: (id: number) => void;
}

const PLATFORM_CONFIG: Record<string, { icon: IconKey; color: string }> = {
  facebook: { icon: 'facebook', color: '#1877f2' },
  messenger: { icon: 'messenger', color: '#0068ff' },
  tiktok: { icon: 'tiktok', color: '#000' },
  shopee: { icon: 'shopee', color: '#fa4659' },
};

const formatTime = (dateStr: string) => {
  if (!dateStr) return '';
  const date = new Date(dateStr);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);
  const diffDays = Math.floor(diffMs / 86400000);

  if (diffMins < 1) return 'vừa xong';
  if (diffMins < 60) return `${diffMins}p`;
  if (diffHours < 24) return `${diffHours}h`;
  if (diffDays < 7) return `${diffDays}ngày`;
  return date.toLocaleDateString('vi-VN', { day: '2-digit', month: '2-digit' });
};

export const ConversationItem: React.FC<ConversationItemProps> = React.memo(({ 
  conversation, 
  isSelected, 
  onSelect 
}) => {
  const { icon, color } = PLATFORM_CONFIG[conversation.customer.platform] || PLATFORM_CONFIG.facebook;
  
  const formattedTime = useMemo(() => formatTime(conversation.updated_at), [conversation.updated_at]);
  
  const truncatedMessage = useMemo(() => {
    const msg = conversation.last_message || '(trống)';
    return msg.length > 40 ? msg.slice(0, 37) + '...' : msg;
  }, [conversation.last_message]);

  return (
    <button
      className={`conversation-item ${conversation.unread_count > 0 ? 'unread' : ''} ${isSelected ? 'selected' : ''}`}
      onClick={() => onSelect(conversation.id)}
      aria-selected={isSelected}
      type="button"
    >
      <div className="avatar">
        {conversation.customer.name?.charAt(0).toUpperCase() || '?'}
        <span
          className="platform-icon"
          style={{ background: color }}
        >
          <AppIcon name={icon} size={12} />
        </span>
      </div>
      <div className="conversation-content">
        <div className="conversation-header">
          <span className="customer-name">{conversation.customer.name || 'Khách vô danh'}</span>
          <span className="last-time">{formattedTime}</span>
        </div>
        <div className="last-message-row">
          <div className="last-message">{truncatedMessage}</div>
          {conversation.unread_count > 0 && (
            <span className="unread-badge">{conversation.unread_count > 99 ? '99+' : conversation.unread_count}</span>
          )}
        </div>
      </div>
    </button>
  );
});

ConversationItem.displayName = 'ConversationItem';
