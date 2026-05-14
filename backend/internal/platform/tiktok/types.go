package tiktok

// --- OAuth ---

// TokenResponse represents the response from TikTok Shop OAuth token exchange.
type TokenResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    *struct {
		AccessToken          string `json:"access_token"`
		AccessTokenExpireIn  int64  `json:"access_token_expire_in"`
		RefreshToken         string `json:"refresh_token"`
		RefreshTokenExpireIn int64  `json:"refresh_token_expire_in"`
		OpenID               string `json:"open_id"`
		SellerName           string `json:"seller_name"`
		SellerBaseRegion     string `json:"seller_base_region"`
		UserType             int    `json:"user_type"`
	} `json:"data,omitempty"`
}

// --- Webhook ---

// WebhookPayload represents the top-level TikTok Shop webhook payload.
type WebhookPayload struct {
	Type      int    `json:"type"`       // 1 = push notification
	ShopID    string `json:"shop_id"`
	Timestamp int64  `json:"timestamp"`
	Data      string `json:"data"` // JSON string of the actual event data
}

// WebhookMessageData represents the parsed message data from a webhook event.
type WebhookMessageData struct {
	ConversationID string `json:"conversation_id"`
	MessageID      string `json:"message_id,omitempty"`
	// For message_received events
	Content     string `json:"content,omitempty"`
	MessageType string `json:"message_type,omitempty"` // TEXT, IMAGE, VIDEO, etc.
	SenderType  string `json:"sender_type,omitempty"`  // BUYER, SELLER
	BuyerUserID string `json:"buyer_user_id,omitempty"`
	CreateTime  int64  `json:"create_time,omitempty"`
	// For image/video messages
	MediaURL string `json:"media_url,omitempty"`
}

// --- Send Message API ---

// SendMessageRequest represents the request body for sending a message via TikTok Shop.
type SendMessageRequest struct {
	ConversationID string `json:"conversation_id"`
	Type           string `json:"type"` // TEXT, IMAGE
	Content        string `json:"content,omitempty"`
}

// SendMessageResponse represents the response from TikTok Shop send message API.
type SendMessageResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    *struct {
		MessageID string `json:"message_id"`
	} `json:"data,omitempty"`
}

// --- Conversation API ---

// ConversationListResponse represents the response from get conversation list API.
type ConversationListResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    *struct {
		Conversations []ConversationInfo `json:"conversations"`
		NextPageToken string             `json:"next_page_token"`
	} `json:"data,omitempty"`
}

// ConversationInfo represents a single conversation.
type ConversationInfo struct {
	ConversationID string `json:"conversation_id"`
	BuyerUserID    string `json:"buyer_user_id"`
	BuyerNickname  string `json:"buyer_nickname"`
	BuyerAvatar    string `json:"buyer_avatar"`
	LastMessage    string `json:"last_message"`
	LastMessageTime int64 `json:"last_message_time"`
	UnreadCount    int    `json:"unread_count"`
}

// --- Message API ---

// MessageListResponse represents the response from get message list API.
type MessageListResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    *struct {
		Messages      []MessageInfo `json:"messages"`
		NextPageToken string        `json:"next_page_token"`
	} `json:"data,omitempty"`
}

// MessageInfo represents a single message.
type MessageInfo struct {
	MessageID      string `json:"message_id"`
	ConversationID string `json:"conversation_id"`
	Type           string `json:"type"` // TEXT, IMAGE, VIDEO, ORDER_CARD
	Content        string `json:"content"`
	SenderType     string `json:"sender_type"` // BUYER, SELLER, SYSTEM
	CreateTime     int64  `json:"create_time"`
}

// --- Shop Info ---

// ShopInfoResponse represents the response from get shop info API.
type ShopInfoResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    *struct {
		ShopID   string `json:"shop_id"`
		ShopName string `json:"shop_name"`
		Region   string `json:"region"`
	} `json:"data,omitempty"`
}
