package platform

// MessagingClient defines the common interface for all social platform integrations.
// Each platform (Facebook, Zalo, TikTok, Instagram, Shopee) must implement this interface.
type MessagingClient interface {
	// SendTextMessage sends a text message to a recipient and returns the platform message ID.
	SendTextMessage(accessToken, recipientID, text string) (string, error)

	// GetUserProfile retrieves the user profile from the platform.
	GetUserProfile(accessToken, userID string) (*UserProfile, error)

	// VerifyWebhookSignature verifies the authenticity of an incoming webhook payload.
	VerifyWebhookSignature(rawBody []byte, signature string) bool

	// PlatformName returns the platform identifier (e.g., "facebook", "zalo", "tiktok").
	PlatformName() string
}

// UserProfile represents a unified user profile across platforms.
type UserProfile struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Avatar    string `json:"avatar,omitempty"`
	Platform  string `json:"platform"`
}

// TokenResponse represents a unified token response from OAuth flows.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
	ExpiresIn    int64  `json:"expires_in,omitempty"`
}

// WebhookMessage represents a normalized incoming message from any platform.
type WebhookMessage struct {
	Platform          string `json:"platform"`
	PagePlatformID    string `json:"page_platform_id"`
	SenderPlatformID  string `json:"sender_platform_id"`
	MessageID         string `json:"message_id"`
	Text              string `json:"text"`
	ContentType       string `json:"content_type"` // text, image, video, audio, file, sticker, location
	MediaURL          string `json:"media_url,omitempty"`
	Timestamp         int64  `json:"timestamp"`
}

// WebhookEvent represents a normalized webhook event from any platform.
type WebhookEvent struct {
	Platform  string            `json:"platform"`
	EventType string            `json:"event_type"` // message, comment, reaction, follow, unfollow
	Messages  []WebhookMessage  `json:"messages,omitempty"`
	RawData   map[string]interface{} `json:"raw_data,omitempty"`
}
