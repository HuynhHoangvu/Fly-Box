// User types
export interface User {
  id: number;
  email: string;
  name: string;
  avatar_url?: string;
  role: string;
}

// Social Page types
export interface SocialPage {
  id: number;
  platform: string;
  external_page_id: string;
  page_name: string;
  access_token?: string;
  status: string;
  connection_status?: string;
  permission_level?: string;
  warning_message?: string;
  requires_reauth?: boolean;
  supports_advanced_tools?: boolean;
  created_at: string;
  updated_at?: string;
}

// Customer types
export interface Customer {
  id: number;
  platform: string;
  platform_id: string;
  name: string;
  avatar?: string;
  phone?: string;
  email?: string;
  created_at: string;
}

// Conversation types
export interface Conversation {
  id: number;
  page_id: number;
  customer_id: number;
  customer: Customer;
  page?: SocialPage;
  last_message: string;
  unread_count: number;
  status: 'open' | 'closed' | 'pending';
  created_at: string;
  updated_at: string;
}

// Message types
export type SenderType = 'customer' | 'staff' | 'system';
export type MessageStatus = 'sending' | 'sent' | 'failed' | 'read';
export type ContentType = 'text' | 'image' | 'video' | 'audio' | 'file' | 'sticker';

export interface Message {
  id: number;
  client_id?: string;
  conversation_id: number;
  sender_type: SenderType;
  sender_id?: number;
  content_type: ContentType;
  content: string;
  status: MessageStatus;
  social_message_id?: string;
  metadata?: Record<string, any>;
  created_at: string;
}

// Auto Reply types
export interface AutoReply {
  id: number;
  page_id: number;
  keyword: string;
  reply_content: string;
  is_active: boolean;
  created_at: string;
  updated_at?: string;
}

// Notification types
export interface Notification {
  id: number;
  type: string;
  title: string;
  content: string;
  platform?: string;
  page_id?: number;
  conversation_id?: number;
  is_read: boolean;
  created_at: string;
}

// Dashboard stats
export interface DashboardStats {
  total_conversations: number;
  unread_conversations: number;
  total_messages: number;
  messages_today: number;
  by_platform: Record<string, number>;
}

// WebSocket event types
export interface WSEvent {
  type: 'NEW_MESSAGE' | 'NEW_NOTIFICATION' | 'BADGE_UPDATE' | 'CONVERSATION_UPDATE';
  page_id?: number;
  user_id?: number;
  data?: any;
}
