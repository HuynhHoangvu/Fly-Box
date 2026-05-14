package zalo

// --- OAuth ---

// TokenResponse represents the response from Zalo OAuth token exchange.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	Error        int    `json:"error,omitempty"`
	Message      string `json:"message,omitempty"`
}

// --- Webhook ---

// WebhookPayload represents the top-level Zalo OA webhook payload.
type WebhookPayload struct {
	AppID     string       `json:"app_id"`
	OAId      string       `json:"oa_id"`
	UserIDByApp string     `json:"user_id_by_app"`
	EventName string       `json:"event_name"`
	Timestamp string       `json:"timestamp"`
	Sender    SenderInfo   `json:"sender"`
	Recipient RecipientInfo `json:"recipient"`
	Message   *MessageInfo `json:"message,omitempty"`
	Follower  *FollowerInfo `json:"follower,omitempty"`
	Info      *UserInfo    `json:"info,omitempty"`
}

// SenderInfo represents the sender in a Zalo webhook event.
type SenderInfo struct {
	ID string `json:"id"`
}

// RecipientInfo represents the recipient (OA) in a Zalo webhook event.
type RecipientInfo struct {
	ID string `json:"id"`
}

// MessageInfo represents the message content in a Zalo webhook event.
type MessageInfo struct {
	MsgID       string        `json:"msg_id"`
	Text        string        `json:"text,omitempty"`
	Attachments []Attachment  `json:"attachments,omitempty"`
}

// Attachment represents a media attachment in a Zalo message.
type Attachment struct {
	Type    string            `json:"type"` // image, file, sticker, gif, video, audio, link, location
	Payload AttachmentPayload `json:"payload"`
}

// AttachmentPayload contains the URL or data for an attachment.
type AttachmentPayload struct {
	Thumbnail string `json:"thumbnail,omitempty"`
	URL       string `json:"url,omitempty"`
	Size      string `json:"size,omitempty"`
	Name      string `json:"name,omitempty"`
	Type      string `json:"type,omitempty"`
	// Location fields
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
	// Sticker fields
	StickerID string `json:"sticker_id,omitempty"`
}

// FollowerInfo represents follow/unfollow event data.
type FollowerInfo struct {
	ID string `json:"id"`
}

// UserInfo represents user info returned in some events.
type UserInfo struct {
	Address  string `json:"address,omitempty"`
	Phone    string `json:"phone,omitempty"`
	City     string `json:"city,omitempty"`
	District string `json:"district,omitempty"`
	Name     string `json:"name,omitempty"`
}

// --- Send Message API ---

// SendMessageRequest represents the request body for sending a CS message via Zalo OA API.
type SendMessageRequest struct {
	Recipient SendRecipient `json:"recipient"`
	Message   SendMessage   `json:"message"`
}

// SendRecipient represents the recipient of a message.
type SendRecipient struct {
	UserID string `json:"user_id"`
}

// SendMessage represents the message content to send.
type SendMessage struct {
	Text string `json:"text,omitempty"`
}

// SendMessageResponse represents the response from Zalo send message API.
type SendMessageResponse struct {
	Error   int    `json:"error"`
	Message string `json:"message"`
	Data    *struct {
		MessageID string `json:"message_id"`
	} `json:"data,omitempty"`
}

// --- User Profile API ---

// UserProfileResponse represents the response from Zalo get user profile API.
type UserProfileResponse struct {
	Error   int    `json:"error"`
	Message string `json:"message"`
	Data    *struct {
		UserID      string `json:"user_id"`
		DisplayName string `json:"display_name"`
		Avatar      string `json:"avatar"`
		AvatarSmall string `json:"avatars>small,omitempty"`
	} `json:"data,omitempty"`
}

// --- OA Info API ---

// OAInfoResponse represents the response from Zalo get OA info API.
type OAInfoResponse struct {
	Error   int    `json:"error"`
	Message string `json:"message"`
	Data    *struct {
		OAID        string `json:"oa_id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Avatar      string `json:"avatar"`
	} `json:"data,omitempty"`
}
