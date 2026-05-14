import React, { useState, useRef, useLayoutEffect, useCallback, useMemo } from 'react';
import { useAuthStore } from '../../store/useAuthStore';
import { AppIcon } from '../common/AppIcon';
import { useMessages } from '../../hooks/useMessages';
import { useConversations } from '../../hooks/useConversations';
import { MessageStatus, SenderType } from '../../types/messaging';
import { Image, Paperclip, Smile, ShoppingBag } from 'lucide-react';
import './MessagePanel.css';

interface MessagePanelProps {
  conversationId: number | null;
}

const PLATFORM_ICONS: Record<string, string> = {
  facebook: 'https://upload.wikimedia.org/wikipedia/commons/b/b8/2021_Facebook_icon.svg',
  instagram: 'https://upload.wikimedia.org/wikipedia/commons/e/e7/Instagram_logo_2016.svg',
  zalo: 'https://upload.wikimedia.org/wikipedia/commons/9/91/Icon_of_Zalo.svg',
  shopee: 'https://upload.wikimedia.org/wikipedia/commons/f/fe/Shopee_logo.svg',
  tiktok: 'https://upload.wikimedia.org/wikipedia/en/a/a9/TikTok_logo.svg'
};

export const MessagePanel: React.FC<MessagePanelProps> = ({ conversationId }) => {
  const [inputValue, setInputValue] = useState('');
  const user = useAuthStore(s => s.user);
  
  const { conversations } = useConversations();
  const { messages, isLoading, error, sendNewMessage } = useMessages(conversationId);
  
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const scrollContainerRef = useRef<HTMLDivElement>(null);
  const isNearBottomRef = useRef(true);

  // Find active conversation details
  const activeConversation = useMemo(() => {
    return conversations.find(c => c.id === conversationId);
  }, [conversations, conversationId]);

  const scrollToBottom = useCallback((behavior: ScrollBehavior = 'smooth') => {
    messagesEndRef.current?.scrollIntoView({ behavior });
  }, []);

  const handleScroll = useCallback(() => {
    if (!scrollContainerRef.current) return;
    const { scrollTop, scrollHeight, clientHeight } = scrollContainerRef.current;
    const isAtBottom = scrollHeight - scrollTop - clientHeight < 100;
    isNearBottomRef.current = isAtBottom;
  }, []);

  useLayoutEffect(() => {
    if (isNearBottomRef.current) {
      scrollToBottom('auto');
    }
  }, [messages, scrollToBottom]);

  const handleSend = async () => {
    if (!inputValue.trim() || !conversationId) return;
    
    const content = inputValue.trim();
    setInputValue('');
    
    try {
      await sendNewMessage(content);
      isNearBottomRef.current = true;
      scrollToBottom();
    } catch (err) {
      console.error('Failed to send message:', err);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      if (e.nativeEvent.isComposing) return;
      e.preventDefault();
      handleSend();
    }
  };

  if (!conversationId) {
    return (
      <div className="message-panel empty">
        <div className="empty-state">
          <AppIcon name="inbox" size={48} />
          <p>Chọn một cuộc hội thoại từ danh sách để bắt đầu</p>
        </div>
      </div>
    );
  }

  const platform = activeConversation?.page?.platform || 'facebook';
  const platformIcon = PLATFORM_ICONS[platform] || PLATFORM_ICONS['facebook'];
  const customerName = activeConversation?.customer?.name || 'Khách hàng';
  const customerAvatar = activeConversation?.customer?.avatar || `https://ui-avatars.com/api/?name=${encodeURIComponent(customerName)}`;

  return (
    <div className="message-panel">
      <div className="message-header">
        <div className="user-info">
          <div className="avatar-small">
            <img src={customerAvatar} alt={customerName} />
            <img src={platformIcon} className="header-platform-badge" alt={platform} />
          </div>
          <div className="name-status">
            <span className="name">{customerName}</span>
            <span className="status online">Đang hoạt động trên {platform}</span>
          </div>
        </div>
        <div className="header-actions">
          <button className="action-btn" title="Tìm kiếm tin nhắn"><AppIcon name="search" size={18} /></button>
          <button className="action-btn" title="Thông tin khách hàng"><AppIcon name="info" size={18} /></button>
        </div>
      </div>

      <div 
        className="message-list" 
        ref={scrollContainerRef}
        onScroll={handleScroll}
      >
        {isLoading && messages.length === 0 && (
          <div className="message-loading">Đang tải cuộc trò chuyện...</div>
        )}
        {error && <div className="message-error">{error}</div>}
        
        {messages.map((msg, idx) => {
          const isStaff = msg.sender_type === SenderType.STAFF;
          const showTime = idx === 0 || 
            new Date(msg.created_at).getTime() - new Date(messages[idx-1].created_at).getTime() > 300000;

          return (
            <React.Fragment key={msg.id || msg.client_id}>
              {showTime && (
                <div className="time-divider">
                  {new Date(msg.created_at).toLocaleTimeString('vi-VN', { hour: '2-digit', minute: '2-digit' })}
                </div>
              )}
              <div className={`message-bubble ${isStaff ? 'sent' : 'received'}`}>
                {!isStaff && (
                  <img src={customerAvatar} className="message-avatar" alt="customer" />
                )}
                <div className="bubble-content">
                  {msg.content}
                  <div className="bubble-footer">
                    {isStaff && (
                      <span className={`status-icon ${msg.status}`}>
                        {msg.status === MessageStatus.SENDING ? '...' : <AppIcon name="check" size={10} />}
                      </span>
                    )}
                  </div>
                </div>
              </div>
            </React.Fragment>
          );
        })}
        <div ref={messagesEndRef} />
      </div>

      <div className="message-input-area">
        <div className="omni-toolbar">
          <button className="omni-tool-btn" title="Gửi ảnh"><Image size={18} /></button>
          <button className="omni-tool-btn" title="Đính kèm file"><Paperclip size={18} /></button>
          <button className="omni-tool-btn" title="Gửi nhãn dán"><Smile size={18} /></button>
          {platform === 'shopee' || platform === 'tiktok' ? (
             <button className="omni-tool-btn highlight" title="Gửi sản phẩm"><ShoppingBag size={18} /></button>
          ) : null}
        </div>
        <div className="input-wrapper">
          <textarea
            value={inputValue}
            onChange={(e) => setInputValue(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder={`Trả lời ${customerName}...`}
            rows={1}
            onInput={(e) => {
              const target = e.target as HTMLTextAreaElement;
              target.style.height = 'auto';
              target.style.height = `${Math.min(target.scrollHeight, 120)}px`;
            }}
          />
          <button 
            className="send-btn" 
            onClick={handleSend}
            disabled={!inputValue.trim()}
            title="Gửi tin nhắn (Enter)"
          >
            <AppIcon name="send" size={20} />
          </button>
        </div>
      </div>
    </div>
  );
};

