import { useState, useRef, useEffect } from 'react';
import { useNotifications } from '../../hooks/useNotifications';
import { useWebSocket } from '../../hooks/useWebSocket';
import { useAuthStore } from '../../store/useAuthStore';
import './NotificationBell.css';

interface Notification {
  id: number;
  type: string;
  platform: string;
  title: string;
  body: string;
  is_read: boolean;
  created_at: string;
  data?: any;
}

export function NotificationBell() {
  const [open, setOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const user = useAuthStore(s => s.user);
  const userId = user ? Number(user.id) : 0;

  const {
    notifications,
    unreadCount,
    markAsRead,
    markAllAsRead,
    refresh,
  } = useNotifications();

  // WebSocket for real-time updates
  useWebSocket({
    userId,
    onNewNotification: () => {
      refresh();
    },
    onBadgeUpdate: () => {
      refresh();
    },
  });

  // Close dropdown when clicking outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const handleNotificationClick = (notification: Notification) => {
    if (!notification.is_read) {
      markAsRead([notification.id]);
    }
    // Navigate to conversation if available
    if (notification.data && typeof notification.data === 'object' && 'conversation_id' in notification.data) {
      const convId = (notification.data as { conversation_id: number }).conversation_id;
      window.location.href = `/inbox?conversation=${convId}`;
    }
    setOpen(false);
  };

  const getPlatformIcon = (platform: string) => {
    const icons: Record<string, string> = {
      facebook: '📘',
      zalo: '💬',
      tiktok: '🎵',
      instagram: '📷',
      shopee: '🛒',
    };
    return icons[platform] || '📢';
  };

  const formatTime = (dateStr: string) => {
    const date = new Date(dateStr);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMins < 1) return 'Vừa xong';
    if (diffMins < 60) return `${diffMins} phút trước`;
    if (diffHours < 24) return `${diffHours} giờ trước`;
    if (diffDays < 7) return `${diffDays} ngày trước`;
    return date.toLocaleDateString('vi-VN');
  };

  const previewNotifications = notifications.slice(0, 5);

  return (
    <div className="notification-bell" ref={dropdownRef}>
      <button
        className="bell-button"
        onClick={() => setOpen(!open)}
        aria-label="Thông báo"
      >
        <span className="bell-icon">🔔</span>
        {unreadCount > 0 && (
          <span className="badge">{unreadCount > 99 ? '99+' : unreadCount}</span>
        )}
      </button>

      {open && (
        <div className="dropdown">
          <div className="dropdown-header">
            <h3>Thông báo</h3>
            {unreadCount > 0 && (
              <button className="mark-all-btn" onClick={markAllAsRead}>
                Đánh dấu tất cả đã đọc
              </button>
            )}
          </div>

          <div className="dropdown-content">
            {previewNotifications.length === 0 ? (
              <div className="empty-state">
                <span>📭</span>
                <p>Không có thông báo nào</p>
              </div>
            ) : (
              previewNotifications.map((notif) => (
                <div
                  key={notif.id}
                  className={`notification-item ${!notif.is_read ? 'unread' : ''}`}
                  onClick={() => handleNotificationClick(notif)}
                >
                  <div className="notif-icon">{getPlatformIcon(notif.platform)}</div>
                  <div className="notif-content">
                    <div className="notif-title">{notif.title}</div>
                    <div className="notif-body">{notif.body}</div>
                    <div className="notif-time">{formatTime(notif.created_at)}</div>
                  </div>
                  {!notif.is_read && <div className="unread-dot" />}
                </div>
              ))
            )}
          </div>

          <div className="dropdown-footer">
            <a href="/notifications">Xem tất cả thông báo</a>
          </div>
        </div>
      )}
    </div>
  );
}