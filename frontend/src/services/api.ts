import axios from 'axios';

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8081';

const api = axios.create({
  baseURL: API_URL,
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add auth token to requests
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Handle 401 responses
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      if (window.location.pathname !== '/login') {
        window.location.href = '/login';
      }
    }
    return Promise.reject(error);
  }
);

// Auth endpoints
export const authAPI = {
  login: (email: string, password: string) =>
    api.post('/api/v1/auth/login', { email, password }),

  loginWithGoogle: (idToken: string) =>
    api.post('/api/v1/auth/login', { id_token: idToken }),

  register: (name: string, email: string, password: string) =>
    api.post('/api/v1/auth/register', { name, email, password }),

  me: () =>
    api.get('/api/v1/users/me'),
};

// Pages endpoints
export const pagesAPI = {
  list: () =>
    api.get('/api/v1/pages'),

  connect: (platform: string) =>
    api.post('/api/v1/pages/connect', { platform }),

  completeConnect: (data: {
    platform: string;
    code: string;
    state: string;
    page_id?: string;
    page_name?: string;
    permission_level?: string;
    connected_shop_name?: string;
  }) =>
    api.post('/api/v1/pages/connect/complete', data),
};

// Conversations endpoints
export const conversationsAPI = {
  list: (params?: { page_id?: string; status?: string }) => 
    api.get('/api/v1/conversations', { params }),
  
  getMessages: (conversationId: string | number, params?: { before?: string; limit?: number }) => 
    api.get(`/api/v1/conversations/${conversationId}/messages`, { params }),
  
  sendMessage: (conversationId: string | number, content: string) => 
    api.post(`/api/v1/conversations/${conversationId}/messages`, { content }),

  markRead: (conversationId: string | number) =>
    api.post(`/api/v1/conversations/${conversationId}/mark-read`),
};

// Auto replies endpoints
export const autoRepliesAPI = {
  list: (params?: { page_id?: string }) => 
    api.get('/api/v1/auto-replies', { params }),
  
  create: (data: { page_id: number; keyword: string; reply_content: string; is_active?: boolean }) => 
    api.post('/api/v1/auto-replies', data),
  
  update: (id: string | number, data: { keyword?: string; reply_content?: string; is_active?: boolean }) => 
    api.put(`/api/v1/auto-replies/${id}`, data),

  delete: (id: string | number) =>
    api.delete(`/api/v1/auto-replies/${id}`),
};

// Notifications endpoints
export const notificationsAPI = {
  list: (params?: { page?: number; page_size?: number; platform?: string; type?: string }) =>
    api.get('/api/v1/notifications', { params }),

  getUnreadCount: () =>
    api.get('/api/v1/notifications/unread-count'),

  markRead: (notificationIds: number[]) =>
    api.post('/api/v1/notifications/mark-read', { notification_ids: notificationIds }),

  markOneRead: (id: number) =>
    api.put(`/api/v1/notifications/${id}/mark-read`),

  markAllRead: () =>
    api.post('/api/v1/notifications/mark-all-read'),
};

export default api;
