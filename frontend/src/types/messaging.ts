export enum SenderType {
  CUSTOMER = 'customer',
  STAFF = 'staff',
  SYSTEM = 'system',
}

export enum MessageStatus {
  SENDING = 'sending',
  SENT = 'sent',
  FAILED = 'failed',
  READ = 'read',
}

export interface Message {
  id: string | number;
  client_id?: string; // For optimistic updates
  conversation_id: number;
  sender_id?: number;
  sender_type: SenderType;
  content: string;
  content_type?: string;
  status: MessageStatus;
  created_at: string;
  metadata?: Record<string, any>;
}

export interface WebSocketMessage {
  type: 'message' | 'typing' | 'read_receipt';
  payload: any;
}

export interface Conversation {
  id: number;
  page_id: number;
  customer_id: number;
  customer: {
    id?: number;
    name: string;
    avatar?: string;
    avatar_url?: string;
    platform: string;
  };
  page?: {
    id: number;
    platform: string;
    page_name: string;
  };
  last_message: string;
  unread_count: number;
  status: 'open' | 'closed' | 'pending';
  updated_at: string;
}
