import { useState, useEffect } from 'react';
import { useNotifications } from '../../hooks/useNotifications';
import { useWebSocket } from '../../hooks/useWebSocket';
import { useAuthStore } from '../../store/useAuthStore';
import './NotificationCenter.css';

export function NotificationCenter() {
  const user = useAuthStore(s => s.user);
  const userId = user ? Number(user.id) : 0;
  const [newNotification, setNewNotification] = useState(false);
  
  const {
    notifications,
    unreadCount,
    loading,
    page,
    totalPages,
    filter,
    setFilter,
    markAsRead,
    markAllAsRead,
    loadMore,
    refresh,
  } = useNotifications();

  // WebSocket for real-time updates
  useWebSocket({
    userId,
    onNewNotification: (notif: any) => {
      setNewNotification(true);
      refresh();
      // Play notification sound (optional)
      try {
        const audio = new Audio('/notification.mp3');
        audio.volume = 0.3;
        audio.play().catch(() => {});
      } catch {}
    },
    onBadgeUpdate: () => {
      refresh();
    },
  });

  // Clear new notification indicator after 3 seconds
  useEffect(() => {
    if (newNotification) {
      const timer = setTimeout(() => setNewNotification(false), 3000);
      return () => clearTimeout(timer);
    }
  }, [newNotification]);

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

  const getPlatformLabel = (platform: string) => {
    const labels: Record<string, string> = {
      facebook: 'Facebook',
      zalo: 'Zalo',
      tiktok: 'TikTok',
      instagram: 'Instagram',
      shopee: 'Shopee',
    };
    return labels[platform] || platform;
  };

  const getTypeLabel = (type: string) => {
    const labels: Record<string, string> = {
      new_message: 'Tin nhắn',
      new_comment: 'Bình luận',
      new_order: 'Đơn hàng',
      new_follower: 'Người theo dõi',
      system: 'Hệ thống',
    };
    return labels[type] || type;
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
    return date.toLocaleDateString('vi-VN', {
      day: 'numeric',
      month: 'short',
      year: date.getFullYear() !== now.getFullYear() ? 'numeric' : undefined,
    });
  };

  const handleNotificationClick = (notification: typeof notifications[0]) => {
    if (!notification.is_read) {
      markAsRead([notification.id]);
    }
    if (notification.data && typeof notification.data === 'object' && 'conversation_id' in notification.data) {
      const convId = (notification.data as { conversation_id: number }).conversation_id;
      window.location.href = `/inbox?conversation=${convId}`;
    }
  };

  const platforms = ['', 'facebook', 'zalo', 'tiktok', 'instagram', 'shopee'];
  const types = ['', 'new_message', 'new_comment', 'new_order', 'new_follower', 'system'];

  return (
    <div className="notification-center">
      <div className="center-header">
        <h1>🔔 Thông báo</h1>
        {unreadCount > 0 && (
          <button className="mark-all-btn" onClick={markAllAsRead}>
            Đánh dấu tất cả đã đọc
          </button>
        )}
      </div>

      <div className="filters">
        <div className="filter-group">
          <label>Nền tảng:</label>
          <select
            value={filter.platform}
            onChange={(e) => setFilter({ ...filter, platform: e.target.value })}
          >
            <option value="">Tất cả</option>
            {platforms.slice(1).map((p) => (
              <option key={p} value={p}>{getPlatformLabel(p)}</option>
            ))}
          </select>
        </div>
        <div className="filter-group">
          <label>Loại:</label>
          <select
            value={filter.type}
            onChange={(e) => setFilter({ ...filter, type: e.target.value })}
          >
            <option value="">Tất cả</option>
            {types.slice(1).map((t) => (
              <option key={t} value={t}>{getTypeLabel(t)}</option>
            ))}
          </select>
        </div>
      </div>

      <div className="notification-list">
        {notifications.length === 0 && !loading ? (
          <div className="empty-state">
            <span>📭</span>
            <h3>Không có thông báo nào</h3>
            <p>Các thông báo mới sẽ xuất hiện ở đây</p>
          </div>
        ) : (
          <>
            {notifications.map((notif) => (
              <div
                key={notif.id}
                className={`notification-card ${!notif.is_read ? 'unread' : ''} ${newNotification ? 'new' : ''}`}
                data-platform={notif.platform}
                data-type={notif.type}
                onClick={() => handleNotificationClick(notif)}
              >
                <div className="card-icon">{getPlatformIcon(notif.platform)}</div>
                <div className="card-content">
                  <div className="card-meta">
                    <span className="platform-badge">{getPlatformLabel(notif.platform)}</span>
                    <span className="type-badge">{getTypeLabel(notif.type)}</span>
                    <span className="time">{formatTime(notif.created_at)}</span>
                  </div>
                  <div className="card-title">{notif.title}</div>
                  <div className="card-body">{notif.body}</div>
                </div>
                {!notif.is_read && <div className="unread-indicator" />}
              </div>
            ))}

            {loading && (
              <div className="loading">Đang tải...</div>
            )}

            {page < totalPages && !loading && (
              <button className="load-more-btn" onClick={loadMore}>
                Xem thêm
              </button>
            )}
          </>
        )}
      </div>
    </div>
  );
}