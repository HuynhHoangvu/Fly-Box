import { create } from 'zustand';
import { authAPI } from '../services/api';

interface User {
  id: number;
  email: string;
  name: string;
  avatar_url?: string;
  role: string;
}

interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  isCheckingAuth: boolean;
  loginWithGoogle: (idToken: string) => Promise<void>;
  loginWithCredentials: (email: string, password: string) => Promise<void>;
  register: (name: string, email: string, password: string) => Promise<void>;
  logout: () => void;
  checkAuth: () => Promise<void>;
}

// Load from localStorage on init
const loadStoredAuth = () => {
  const token = localStorage.getItem('token');
  const userStr = localStorage.getItem('user');
  if (token && userStr) {
    try {
      const user = JSON.parse(userStr) as User;
      return { token, user, isAuthenticated: true };
    } catch {
      // Invalid stored data
    }
  }
  return { token: null, user: null, isAuthenticated: false };
};

const storedAuth = loadStoredAuth();

export const useAuthStore = create<AuthState>((set) => ({
  user: storedAuth.user,
  token: storedAuth.token,
  isAuthenticated: storedAuth.isAuthenticated,
  isLoading: false,
  isCheckingAuth: false,

  loginWithGoogle: async (idToken: string) => {
    set({ isLoading: true });
    try {
      const { data } = await authAPI.loginWithGoogle(idToken);
      const token = data.token;
      const user = data.user;

      console.log('[AuthStore] Login Google Success, User:', user.name);
      localStorage.setItem('token', token);
      localStorage.setItem('user', JSON.stringify(user));
      set({
        user,
        token,
        isAuthenticated: true,
        isLoading: false,
        isCheckingAuth: false
      });
    } catch (error) {
      console.error('[AuthStore] Login Google Failed:', error);
      set({ isLoading: false });
      throw error;
    }
  },

  loginWithCredentials: async (email: string, password: string) => {
    set({ isLoading: true });
    try {
      const { data } = await authAPI.login(email, password);
      const token = data.token;
      const user = data.user;

      console.log('[AuthStore] Login Credentials Success, User:', user.name);
      localStorage.setItem('token', token);
      localStorage.setItem('user', JSON.stringify(user));
      set({
        user,
        token,
        isAuthenticated: true,
        isLoading: false,
        isCheckingAuth: false
      });
    } catch (error) {
      console.error('[AuthStore] Login Credentials Failed:', error);
      set({ isLoading: false });
      throw error;
    }
  },

  register: async (name: string, email: string, password: string) => {
    set({ isLoading: true });
    try {
      const { data } = await authAPI.register(name, email, password);
      const token = data.token;
      const user = data.user;

      console.log('[AuthStore] Register Success, User:', user.name);
      localStorage.setItem('token', token);
      localStorage.setItem('user', JSON.stringify(user));
      set({
        user,
        token,
        isAuthenticated: true,
        isLoading: false,
        isCheckingAuth: false
      });
    } catch (error) {
      console.error('[AuthStore] Register Failed:', error);
      set({ isLoading: false });
      throw error;
    }
  },

  logout: () => {
    console.log('[AuthStore] Logging out...');
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    set({ 
      user: null, 
      token: null, 
      isAuthenticated: false,
      isLoading: false,
      isCheckingAuth: false
    });
  },

  checkAuth: async () => {
    const token = localStorage.getItem('token');
    const userStr = localStorage.getItem('user');
    
    if (!token || !userStr) {
      console.log('[AuthStore] No stored auth, user not authenticated');
      set({ isAuthenticated: false, user: null, token: null });
      return;
    }

    set({ isCheckingAuth: true });
    try {
      // Verify with backend
      const { data } = await authAPI.me();
      console.log('[AuthStore] checkAuth success, user:', data.user.name);
      set({
        user: data.user,
        isAuthenticated: true,
        isCheckingAuth: false
      });
    } catch (error) {
      console.error('[AuthStore] checkAuth failed, clearing auth:', error);
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      set({ 
        user: null, 
        token: null, 
        isAuthenticated: false,
        isCheckingAuth: false 
      });
    }
  },
}));
