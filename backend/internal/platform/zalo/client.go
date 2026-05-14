package zalo

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"fly-box/backend/internal/platform"
)

const (
	oauthBaseURL = "https://oauth.zaloapp.com/v4/oa"
	openAPIBase  = "https://openapi.zalo.me/v3.0/oa"
)

// Client handles all Zalo OA API calls and implements platform.MessagingClient.
type Client struct {
	AppID       string
	AppSecret   string
	OASecretKey string // Used for webhook signature verification
	httpClient  *http.Client
}

// NewClient creates a new Zalo OA API client.
func NewClient(appID, appSecret, oaSecretKey string) *Client {
	return &Client{
		AppID:       appID,
		AppSecret:   appSecret,
		OASecretKey: oaSecretKey,
		httpClient:  &http.Client{},
	}
}

// PlatformName returns the platform identifier.
func (c *Client) PlatformName() string {
	return "zalo"
}

// --- platform.MessagingClient interface ---

// SendTextMessage sends a customer service text message via Zalo OA API.
// accessToken is the OA access token, recipientID is the Zalo user ID.
func (c *Client) SendTextMessage(accessToken, recipientID, text string) (string, error) {
	apiURL := openAPIBase + "/message/cs"

	payload := SendMessageRequest{
		Recipient: SendRecipient{UserID: recipientID},
		Message:   SendMessage{Text: text},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("access_token", accessToken)

	body, err := c.doRequest(req)
	if err != nil {
		return "", fmt.Errorf("send message: %w", err)
	}

	var result SendMessageResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	if result.Error != 0 {
		return "", fmt.Errorf("zalo API error (%d): %s", result.Error, result.Message)
	}

	messageID := ""
	if result.Data != nil {
		messageID = result.Data.MessageID
	}

	return messageID, nil
}

// GetUserProfile retrieves user profile info from Zalo OA.
func (c *Client) GetUserProfile(accessToken, userID string) (*platform.UserProfile, error) {
	apiURL := openAPIBase + "/getprofile"

	params := url.Values{}
	params.Set("data", fmt.Sprintf(`{"user_id":"%s"}`, userID))

	req, err := http.NewRequest("GET", apiURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("access_token", accessToken)

	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("get user profile: %w", err)
	}

	var result UserProfileResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if result.Error != 0 {
		return nil, fmt.Errorf("zalo API error (%d): %s", result.Error, result.Message)
	}

	profile := &platform.UserProfile{
		ID:       userID,
		Platform: "zalo",
	}
	if result.Data != nil {
		profile.Name = result.Data.DisplayName
		profile.Avatar = result.Data.Avatar
	}

	return profile, nil
}

// VerifyWebhookSignature verifies the authenticity of an incoming Zalo webhook payload.
// Zalo uses HMAC-SHA256 with the OA Secret Key.
// The signature is sent in the X-ZEvent-Signature header as "mac=<hex>".
func (c *Client) VerifyWebhookSignature(rawBody []byte, signature string) bool {
	if c.OASecretKey == "" || signature == "" {
		return false
	}

	// Signature format: "mac=<hex>"
	prefix := "mac="
	if !strings.HasPrefix(signature, prefix) {
		return false
	}
	sigHex := signature[len(prefix):]

	mac := hmac.New(sha256.New, []byte(c.OASecretKey))
	mac.Write(rawBody)
	expected := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expected), []byte(sigHex))
}

// --- Zalo-specific methods ---

// ExchangeCodeForToken exchanges an authorization code for an OA access token.
func (c *Client) ExchangeCodeForToken(code, redirectURI string) (*TokenResponse, error) {
	apiURL := oauthBaseURL + "/access_token"

	formData := url.Values{}
	formData.Set("app_id", c.AppID)
	formData.Set("code", code)
	formData.Set("grant_type", "authorization_code")

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("secret_key", c.AppSecret)

	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("exchange code: %w", err)
	}

	var result TokenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if result.Error != 0 {
		return nil, fmt.Errorf("zalo OAuth error (%d): %s", result.Error, result.Message)
	}

	return &result, nil
}

// RefreshToken refreshes an expired OA access token.
func (c *Client) RefreshToken(refreshToken string) (*TokenResponse, error) {
	apiURL := oauthBaseURL + "/access_token"

	formData := url.Values{}
	formData.Set("app_id", c.AppID)
	formData.Set("refresh_token", refreshToken)
	formData.Set("grant_type", "refresh_token")

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("secret_key", c.AppSecret)

	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("refresh token: %w", err)
	}

	var result TokenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if result.Error != 0 {
		return nil, fmt.Errorf("zalo refresh error (%d): %s", result.Error, result.Message)
	}

	return &result, nil
}

// GetOAInfo retrieves the Official Account info.
func (c *Client) GetOAInfo(accessToken string) (*OAInfoResponse, error) {
	apiURL := openAPIBase + "/getoa"

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("access_token", accessToken)

	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("get OA info: %w", err)
	}

	var result OAInfoResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if result.Error != 0 {
		return nil, fmt.Errorf("zalo API error (%d): %s", result.Error, result.Message)
	}

	return &result, nil
}

// --- Webhook parsing ---

// ParseWebhookPayload parses raw JSON bytes into a ZaloWebhookPayload.
func ParseWebhookPayload(rawBody []byte) (*WebhookPayload, error) {
	var payload WebhookPayload
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		return nil, fmt.Errorf("parse zalo webhook: %w", err)
	}
	return &payload, nil
}

// ExtractMessages normalizes a Zalo webhook payload into platform.WebhookMessage slice.
func ExtractMessages(payload *WebhookPayload) []platform.WebhookMessage {
	if payload == nil {
		return nil
	}

	var messages []platform.WebhookMessage

	switch payload.EventName {
	case "user_send_text":
		if payload.Message != nil {
			messages = append(messages, platform.WebhookMessage{
				Platform:         "zalo",
				PagePlatformID:   payload.Recipient.ID,
				SenderPlatformID: payload.Sender.ID,
				MessageID:        payload.Message.MsgID,
				Text:             payload.Message.Text,
				ContentType:      "text",
				Timestamp:        parseTimestamp(payload.Timestamp),
			})
		}

	case "user_send_image":
		if payload.Message != nil {
			mediaURL := ""
			if len(payload.Message.Attachments) > 0 {
				mediaURL = payload.Message.Attachments[0].Payload.URL
				if mediaURL == "" {
					mediaURL = payload.Message.Attachments[0].Payload.Thumbnail
				}
			}
			messages = append(messages, platform.WebhookMessage{
				Platform:         "zalo",
				PagePlatformID:   payload.Recipient.ID,
				SenderPlatformID: payload.Sender.ID,
				MessageID:        payload.Message.MsgID,
				Text:             "[Hình ảnh]",
				ContentType:      "image",
				MediaURL:         mediaURL,
				Timestamp:        parseTimestamp(payload.Timestamp),
			})
		}

	case "user_send_file":
		if payload.Message != nil {
			mediaURL := ""
			fileName := ""
			if len(payload.Message.Attachments) > 0 {
				mediaURL = payload.Message.Attachments[0].Payload.URL
				fileName = payload.Message.Attachments[0].Payload.Name
			}
			text := "[Tệp tin]"
			if fileName != "" {
				text = fmt.Sprintf("[Tệp tin: %s]", fileName)
			}
			messages = append(messages, platform.WebhookMessage{
				Platform:         "zalo",
				PagePlatformID:   payload.Recipient.ID,
				SenderPlatformID: payload.Sender.ID,
				MessageID:        payload.Message.MsgID,
				Text:             text,
				ContentType:      "file",
				MediaURL:         mediaURL,
				Timestamp:        parseTimestamp(payload.Timestamp),
			})
		}

	case "user_send_audio":
		if payload.Message != nil {
			mediaURL := ""
			if len(payload.Message.Attachments) > 0 {
				mediaURL = payload.Message.Attachments[0].Payload.URL
			}
			messages = append(messages, platform.WebhookMessage{
				Platform:         "zalo",
				PagePlatformID:   payload.Recipient.ID,
				SenderPlatformID: payload.Sender.ID,
				MessageID:        payload.Message.MsgID,
				Text:             "[Tin nhắn thoại]",
				ContentType:      "audio",
				MediaURL:         mediaURL,
				Timestamp:        parseTimestamp(payload.Timestamp),
			})
		}

	case "user_send_video":
		if payload.Message != nil {
			mediaURL := ""
			if len(payload.Message.Attachments) > 0 {
				mediaURL = payload.Message.Attachments[0].Payload.URL
				if mediaURL == "" {
					mediaURL = payload.Message.Attachments[0].Payload.Thumbnail
				}
			}
			messages = append(messages, platform.WebhookMessage{
				Platform:         "zalo",
				PagePlatformID:   payload.Recipient.ID,
				SenderPlatformID: payload.Sender.ID,
				MessageID:        payload.Message.MsgID,
				Text:             "[Video]",
				ContentType:      "video",
				MediaURL:         mediaURL,
				Timestamp:        parseTimestamp(payload.Timestamp),
			})
		}

	case "user_send_sticker":
		if payload.Message != nil {
			mediaURL := ""
			if len(payload.Message.Attachments) > 0 {
				mediaURL = payload.Message.Attachments[0].Payload.URL
				if mediaURL == "" {
					mediaURL = payload.Message.Attachments[0].Payload.Thumbnail
				}
			}
			messages = append(messages, platform.WebhookMessage{
				Platform:         "zalo",
				PagePlatformID:   payload.Recipient.ID,
				SenderPlatformID: payload.Sender.ID,
				MessageID:        payload.Message.MsgID,
				Text:             "[Sticker]",
				ContentType:      "sticker",
				MediaURL:         mediaURL,
				Timestamp:        parseTimestamp(payload.Timestamp),
			})
		}

	case "user_send_location":
		if payload.Message != nil {
			lat, lng := 0.0, 0.0
			if len(payload.Message.Attachments) > 0 {
				lat = payload.Message.Attachments[0].Payload.Latitude
				lng = payload.Message.Attachments[0].Payload.Longitude
			}
			text := fmt.Sprintf("[Vị trí: %.6f, %.6f]", lat, lng)
			messages = append(messages, platform.WebhookMessage{
				Platform:         "zalo",
				PagePlatformID:   payload.Recipient.ID,
				SenderPlatformID: payload.Sender.ID,
				MessageID:        payload.Message.MsgID,
				Text:             text,
				ContentType:      "location",
				Timestamp:        parseTimestamp(payload.Timestamp),
			})
		}

	case "user_send_link":
		if payload.Message != nil {
			linkURL := ""
			if len(payload.Message.Attachments) > 0 {
				linkURL = payload.Message.Attachments[0].Payload.URL
			}
			text := "[Liên kết]"
			if linkURL != "" {
				text = fmt.Sprintf("[Liên kết: %s]", linkURL)
			}
			messages = append(messages, platform.WebhookMessage{
				Platform:         "zalo",
				PagePlatformID:   payload.Recipient.ID,
				SenderPlatformID: payload.Sender.ID,
				MessageID:        payload.Message.MsgID,
				Text:             text,
				ContentType:      "text",
				MediaURL:         linkURL,
				Timestamp:        parseTimestamp(payload.Timestamp),
			})
		}

	case "user_send_gif":
		if payload.Message != nil {
			mediaURL := ""
			if len(payload.Message.Attachments) > 0 {
				mediaURL = payload.Message.Attachments[0].Payload.URL
				if mediaURL == "" {
					mediaURL = payload.Message.Attachments[0].Payload.Thumbnail
				}
			}
			messages = append(messages, platform.WebhookMessage{
				Platform:         "zalo",
				PagePlatformID:   payload.Recipient.ID,
				SenderPlatformID: payload.Sender.ID,
				MessageID:        payload.Message.MsgID,
				Text:             "[GIF]",
				ContentType:      "image",
				MediaURL:         mediaURL,
				Timestamp:        parseTimestamp(payload.Timestamp),
			})
		}

		// follow/unfollow events don't produce messages but could be logged
	}

	return messages
}

// --- Internal helpers ---

func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func parseTimestamp(ts string) int64 {
	var t int64
	fmt.Sscanf(ts, "%d", &t)
	return t
}
