package instagram

// ExchangeTokenResponse represents the response from Meta OAuth token exchange.
type ExchangeTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// IGAccount represents an Instagram Business Account linked to a Facebook Page.
type IGAccount struct {
	ID         string `json:"id"`
	Username   string `json:"username"`
	Name       string `json:"name"`
	ProfilePic string `json:"profile_picture_url"`
}

// SendMessageResponse represents the response from sending a message via Instagram Messaging API.
type SendMessageResponse struct {
	RecipientID string `json:"recipient_id"`
	MessageID   string `json:"message_id"`
}

// WebhookPayload represents the top-level Meta webhook payload for Instagram.
// Note: Instagram uses the same webhook structure as Facebook, distinguished by the "object" field.
type WebhookPayload struct {
	Object string         `json:"object"`
	Entry  []WebhookEntry `json:"entry"`
}

// WebhookEntry represents a single entry in the webhook payload.
type WebhookEntry struct {
	ID        string           `json:"id"`
	Time      int64            `json:"time"`
	Messaging []MessagingEvent `json:"messaging"`
	Changes   []FeedChange     `json:"changes,omitempty"`
}

// MessagingEvent represents a single messaging event (DM) for Instagram.
type MessagingEvent struct {
	Sender    IDField           `json:"sender"`
	Recipient IDField           `json:"recipient"`
	Timestamp int64             `json:"timestamp"`
	Message   *MessagePayload   `json:"message,omitempty"`
	Postback  *PostbackPayload  `json:"postback,omitempty"`
	Reaction  *ReactionPayload  `json:"reaction,omitempty"`
	Read      *ReadPayload      `json:"read,omitempty"`
	Delivery  *DeliveryPayload  `json:"delivery,omitempty"`
}

// IDField represents a sender or recipient with an ID.
type IDField struct {
	ID string `json:"id"`
}

// MessagePayload represents the message content in a messaging event.
type MessagePayload struct {
	Mid         string       `json:"mid"`
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments,omitempty"`
	IsEcho      bool         `json:"is_echo,omitempty"`
}

// Attachment represents a media attachment in a message.
type Attachment struct {
	Type    string            `json:"type"` // image, video, audio, file
	Payload AttachmentPayload `json:"payload"`
}

// AttachmentPayload contains the URL for an attachment.
type AttachmentPayload struct {
	URL string `json:"url,omitempty"`
}

// PostbackPayload represents a postback event (e.g., button clicks).
type PostbackPayload struct {
	Title   string `json:"title"`
	Payload string `json:"payload"`
}

// ReactionPayload represents a message reaction event.
type ReactionPayload struct {
	Reaction string `json:"reaction"`
	Emoji    string `json:"emoji"`
	Action   string `json:"action"` // react, unreact
	Mid      string `json:"mid"`
}

// ReadPayload represents a read receipt event.
type ReadPayload struct {
	Mid string `json:"mid"`
}

// DeliveryPayload represents a delivery receipt event.
type DeliveryPayload struct {
	Mids      []string `json:"mids"`
	Watermark int64    `json:"watermark"`
}

// FeedChange represents a change in the Instagram media (comments, mentions, etc.).
type FeedChange struct {
	Field string                 `json:"field"`
	Value map[string]interface{} `json:"value"`
}

// GraphAPIError represents a Meta Graph API error response.
type GraphAPIError struct {
	Error struct {
		Message   string `json:"message"`
		Type      string `json:"type"`
		Code      int    `json:"code"`
		FBTraceID string `json:"fbtrace_id"`
	} `json:"error"`
}

// InstagramConnectResult holds the result of Instagram OAuth flow.
type InstagramConnectResult struct {
	UserAccessToken string
	Pages           []PageAccount
	IGAccounts      []IGAccount
}