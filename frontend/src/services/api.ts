import axios from 'axios';

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8081';

const api = axios.create({
  baseURL: API_URL,
  withCredentials: true, // Thêm dòng này để hỗ trợ CORS tốt hơn
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
      // We can't use hooks here, but clearing localStorage is enough 
      // for the next App mount/checkAuth to detect the logout.
      // Optionally redirect to login
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
    page_id: string;
    page_name?: string;
    permission_level?: string;
    connected_shop_name?: string;
  }) =>
    api.post('/api/v1/pages/connect/complete', data),
};

// Conversations endpoints
export const conversationsAPI = {
  list: (pageId?: string) => 
    api.get('/api/v1/conversations', { params: { page_id: pageId } }),
  
  getMessages: (conversationId: string) => 
    api.get(`/api/v1/conversations/${conversationId}/messages`),
  
sendMessage: (conversationId: string, content: string, clientId: string) => 
    api.post(`/api/v1/conversations/${conversationId}/messages`, { 
      content
    }),
};

// Auto replies endpoints
export const autoRepliesAPI = {
  list: () => 
    api.get('/api/v1/auto-replies'),
  
  create: (data: { keyword: string; reply: string; platform: string }) => 
    api.post('/api/v1/auto-replies', data),
  
  update: (id: string, data: { keyword: string; reply: string; platform: string }) => 
    api.put(`/api/v1/auto-replies/${id}`, data),
};

// Notifications endpoints
export const notificationsAPI = {
  list: (params?: { page?: number; page_size?: number; platform?: string; type?: string }) =>
    api.get('/api/v1/notifications', { params }),

  getUnreadCount: () =>
    api.get('/api/v1/notifications/unread-count'),

  markRead: (notificationIds: number[]) =>
    api.post('/api/v1/notifications/mark-read', { notification_ids: notificationIds }),

  markAllRead: () =>
    api.post('/api/v1/notifications/mark-all-read'),
};

export default api;
