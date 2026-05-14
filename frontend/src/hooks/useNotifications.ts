import { useState, useEffect, useCallback } from 'react';
import { notificationsAPI } from '../services/api';

interface Notification {
  id: number;
  user_id: number;
  page_id?: number;
  type: string;
  platform: string;
  title: string;
  body: string;
  data?: any;
  is_read: boolean;
  created_at: string;
}

interface NotificationResponse {
  data: Notification[];
  meta: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
}

interface UnreadCountResponse {
  data: {
    count: number;
  };
}

export function useNotifications() {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [filter, setFilter] = useState({ platform: '', type: '' });

  const fetchUnreadCount = useCallback(async () => {
    try {
      const res = await notificationsAPI.getUnreadCount();
      setUnreadCount(res.data.data.count);
    } catch (err) {
      console.error('Failed to fetch unread count:', err);
    }
  }, []);

  const fetchNotifications = useCallback(async (pageNum = 1) => {
    setLoading(true);
    try {
      const res = await notificationsAPI.list({
        page: pageNum,
        page_size: 20,
        platform: filter.platform || undefined,
        type: filter.type || undefined,
      });
      const data = res.data as NotificationResponse;
      setNotifications(prev => pageNum === 1 ? data.data : [...prev, ...data.data]);
      setPage(data.meta.page);
      setTotalPages(data.meta.total_pages);
    } catch (err) {
      console.error('Failed to fetch notifications:', err);
    } finally {
      setLoading(false);
    }
  }, [filter]);

  const markAsRead = useCallback(async (ids: number[]) => {
    try {
      await notificationsAPI.markRead(ids);
      setNotifications(prev =>
        prev.map(n => ids.includes(n.id) ? { ...n, is_read: true } : n)
      );
      setUnreadCount(prev => Math.max(0, prev - ids.length));
    } catch (err) {
      console.error('Failed to mark as read:', err);
    }
  }, []);

  const markAllAsRead = useCallback(async () => {
    try {
      await notificationsAPI.markAllRead();
      setNotifications(prev => prev.map(n => ({ ...n, is_read: true })));
      setUnreadCount(0);
    } catch (err) {
      console.error('Failed to mark all as read:', err);
    }
  }, []);

  const loadMore = useCallback(() => {
    if (page < totalPages && !loading) {
      fetchNotifications(page + 1);
    }
  }, [page, totalPages, loading, fetchNotifications]);

  useEffect(() => {
    fetchUnreadCount();
    fetchNotifications(1);
  }, [fetchUnreadCount, fetchNotifications]);

  return {
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
    refresh: () => {
      fetchUnreadCount();
      fetchNotifications(1);
    },
  };
}