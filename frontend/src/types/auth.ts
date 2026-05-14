export interface User {
  id: number;
  email: string;
  full_name: string;
  avatar_url?: string;
  role: string;
}

export interface AuthState {
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

export interface Page {
  id: string;
  name: string;
  platform: 'facebook' | 'messenger' | 'tiktok' | 'shopee' | 'instagram' | 'zalo';
  avatar_url?: string;
  connected_at: string;
  status: 'active' | 'expired' | 'disconnected';
}
