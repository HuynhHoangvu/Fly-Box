package shopee

// APIResponse is the base response structure for Shopee API.
type APIResponse struct {
	ResponseError ShopeeError `json:"error"`
}

// ShopeeError represents a Shopee API error.
type ShopeeError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ShopInfoResponse represents the response from get_profile API.
type ShopInfoResponse struct {
	ResponseError ShopeeError `json:"error"`
	ResponseData  ShopInfo   `json:"data"`
}

// ShopInfo represents Shopee shop information.
type ShopInfo struct {
	ShopID       int64  `json:"shop_id"`
	Country      string `json:"country"`
	ShopName     string `json:"shop_name"`
	ShopLogo     string `json:"shop_logo"`
	IsEnterprise bool   `json:"is_enterprise"`
}

// SendMessageResponse represents the response from send_message API.
type SendMessageResponse struct {
	ResponseError ShopeeError `json:"error"`
	ResponseData   struct {
		MessageID string `json:"message_id"`
	} `json:"data"`
}

// GetConversationListResponse represents the response from get_conversation_list API.
type GetConversationListResponse struct {
	ResponseError ShopeeError       `json:"error"`
	ResponseData   ConversationList `json:"data"`
}

// ConversationList represents a list of conversations.
type ConversationList struct {
	Conversations []Conversation `json:"conversation_list"`
	TotalCount    int            `json:"total_count"`
	HasMore       bool           `json:"has_more"`
}

// Conversation represents a single conversation.
type Conversation struct {
	ShopID           int64  `json:"shop_id"`
	BuyerID          string `json:"buyer_id"`
	BuyerName        string `json:"buyer_name"`
	LastMessage      string `json:"last_message"`
	LastMessageType  string `json:"last_message_type"`
	LastMessageTime  int64  `json:"last_message_time"`
	UnreadCount      int    `json:"unread_count"`
}

// GetMessagesResponse represents the response from get_message API.
type GetMessagesResponse struct {
	ResponseError ShopeeError `json:"error"`
	ResponseData   MessageList `json:"data"`
}

// MessageList represents a list of messages.
type MessageList struct {
	Messages []Message `json:"message_list"`
	HasMore  bool      `json:"has_more"`
}

// Message represents a single message in a conversation.
type Message struct {
	MessageID   string `json:"message_id"`
	ShopID      int64  `json:"shop_id"`
	BuyerID     string `json:"buyer_id"`
	MessageType string `json:"message_type"`
	Content     struct {
		Text string `json:"text"`
		URL  string `json:"url,omitempty"`
	} `json:"content"`
	CreateTime int64 `json:"create_time"`
}

// UnreadCountResponse represents the response from unread_conversation_count API.
type UnreadCountResponse struct {
	ResponseError ShopeeError `json:"error"`
	ResponseData   struct {
		Count int `json:"count"`
	} `json:"data"`
}

// AuthorizeResponse represents the response from authorize API.
type AuthorizeResponse struct {
	ResponseError ShopeeError `json:"error"`
	ResponseData   struct {
		Success bool `json:"success"`
	} `json:"data"`
}

// WebhookPayload represents the Shopee webhook payload.
type WebhookPayload struct {
	Events []WebhookEvent `json:"events"`
}

// WebhookEvent represents a single webhook event.
type WebhookEvent struct {
	Topic   string      `json:"topic"`
	Data    interface{} `json:"data"`
	TraceID string      `json:"trace_id"`
}

// ShopeeConnectResult holds the result of Shopee OAuth flow.
type ShopeeConnectResult struct {
	ShopID       string
	ShopName     string
	AccessToken  string
}

// ShopAccount represents a Shopee shop account.
type ShopAccount struct {
	ShopID   string `json:"shop_id"`
	ShopName string `json:"shop_name"`
}