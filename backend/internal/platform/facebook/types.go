package facebook

// ExchangeTokenResponse represents the response from Facebook OAuth token exchange.
type ExchangeTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// PageAccount represents a Facebook page from /me/accounts.
type PageAccount struct {
	ID          string   `json:"id"`
	AccessToken string   `json:"access_token"`
	Name        string   `json:"name"`
	Category    string   `json:"category"`
	Tasks       []string `json:"tasks,omitempty"`
}

// SendMessageResponse represents the response from sending a message via Messenger API.
type SendMessageResponse struct {
	RecipientID string `json:"recipient_id"`
	MessageID   string `json:"message_id"`
}

// WebhookPayload represents the top-level Facebook webhook payload.
type WebhookPayload struct {
	Object string         `json:"object"`
	Entry  []WebhookEntry `json:"entry"`
}

// WebhookEntry represents a single entry in the webhook payload.
type WebhookEntry struct {
	ID        string             `json:"id"`
	Time      int64              `json:"time"`
	Messaging []MessagingEvent   `json:"messaging"`
	Changes   []FeedChange       `json:"changes,omitempty"`
}

// MessagingEvent represents a single messaging event (DM).
type MessagingEvent struct {
	Sender    IDField          `json:"sender"`
	Recipient IDField          `json:"recipient"`
	Timestamp int64            `json:"timestamp"`
	Message   *MessagePayload  `json:"message,omitempty"`
	Postback  *PostbackPayload `json:"postback,omitempty"`
	Reaction  *ReactionPayload `json:"reaction,omitempty"`
	Read      *ReadPayload     `json:"read,omitempty"`
	Delivery  *DeliveryPayload `json:"delivery,omitempty"`
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
	AppID       int64        `json:"app_id,omitempty"`
}

// Attachment represents a media attachment in a message.
type Attachment struct {
	Type    string            `json:"type"` // image, video, audio, file, template, fallback
	Payload AttachmentPayload `json:"payload"`
}

// AttachmentPayload contains the URL or other data for an attachment.
type AttachmentPayload struct {
	URL      string `json:"url,omitempty"`
	Title    string `json:"title,omitempty"`
	StickerID int64 `json:"sticker_id,omitempty"`
}

// PostbackPayload represents a postback event (button tap).
type PostbackPayload struct {
	Title   string `json:"title"`
	Payload string `json:"payload"`
}

// ReactionPayload represents a message reaction event.
type ReactionPayload struct {
	Reaction  string `json:"reaction"` // smile, angry, sad, wow, love, like, dislike, other
	Emoji     string `json:"emoji"`
	Action    string `json:"action"` // react, unreact
	Mid       string `json:"mid"`
}

// ReadPayload represents a read receipt event.
type ReadPayload struct {
	Watermark int64 `json:"watermark"`
}

// DeliveryPayload represents a delivery receipt event.
type DeliveryPayload struct {
	Mids      []string `json:"mids"`
	Watermark int64    `json:"watermark"`
}

// FeedChange represents a change in the page feed (comments, posts).
type FeedChange struct {
	Field string                 `json:"field"`
	Value map[string]interface{} `json:"value"`
}

// GraphAPIError represents a Facebook Graph API error response.
type GraphAPIError struct {
	Error struct {
		Message   string `json:"message"`
		Type      string `json:"type"`
		Code      int    `json:"code"`
		FBTraceID string `json:"fbtrace_id"`
	} `json:"error"`
}
