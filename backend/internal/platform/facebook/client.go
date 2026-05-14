package facebook

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

const graphAPIVersion = "v19.0"
const graphAPIBase = "https://graph.facebook.com/" + graphAPIVersion

// Client handles all Facebook Graph API calls and implements platform.MessagingClient.
type Client struct {
	AppID     string
	AppSecret string
	httpClient *http.Client
}

// NewClient creates a new Facebook Graph API client.
func NewClient(appID, appSecret string) *Client {
	return &Client{
		AppID:      appID,
		AppSecret:  appSecret,
		httpClient: &http.Client{},
	}
}

// PlatformName returns the platform identifier.
func (c *Client) PlatformName() string {
	return "facebook"
}

// --- platform.MessagingClient interface ---

// SendTextMessage sends a text message via the Page Messaging API.
// Returns the platform message ID.
func (c *Client) SendTextMessage(pageAccessToken, recipientID, text string) (string, error) {
	apiURL := graphAPIBase + "/me/messages"

	payload := map[string]interface{}{
		"recipient": map[string]string{"id": recipientID},
		"message":   map[string]string{"text": text},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal payload: %w", err)
	}

	params := url.Values{}
	params.Set("access_token", pageAccessToken)

	req, err := http.NewRequest("POST", apiURL+"?"+params.Encode(), bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	body, err := c.doRequest(req)
	if err != nil {
		return "", fmt.Errorf("send message: %w", err)
	}

	var result SendMessageResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	return result.MessageID, nil
}

// GetUserProfile retrieves user profile info from Facebook.
func (c *Client) GetUserProfile(pageAccessToken, userID string) (*platform.UserProfile, error) {
	apiURL := graphAPIBase + "/" + userID

	params := url.Values{}
	params.Set("access_token", pageAccessToken)
	params.Set("fields", "name,first_name,last_name,profile_pic")

	resp, err := c.httpClient.Get(apiURL + "?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("get user profile: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get user profile failed (%d): %s", resp.StatusCode, NormalizeError(body))
	}

	var raw struct {
		Name       string `json:"name"`
		FirstName  string `json:"first_name"`
		LastName   string `json:"last_name"`
		ProfilePic string `json:"profile_pic"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return &platform.UserProfile{
		ID:        userID,
		Name:      raw.Name,
		FirstName: raw.FirstName,
		LastName:  raw.LastName,
		Avatar:    raw.ProfilePic,
		Platform:  "facebook",
	}, nil
}

// VerifyWebhookSignature verifies the X-Hub-Signature-256 header from Facebook webhooks.
func (c *Client) VerifyWebhookSignature(rawBody []byte, signature string) bool {
	if c.AppSecret == "" || signature == "" {
		return false
	}

	// Signature format: "sha256=<hex>"
	prefix := "sha256="
	if !strings.HasPrefix(signature, prefix) {
		return false
	}
	sigHex := signature[len(prefix):]

	mac := hmac.New(sha256.New, []byte(c.AppSecret))
	mac.Write(rawBody)
	expected := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expected), []byte(sigHex))
}

// --- Facebook-specific methods (not part of MessagingClient interface) ---

// ExchangeCodeForToken exchanges an authorization code for a user access token.
func (c *Client) ExchangeCodeForToken(code, redirectURI string) (*ExchangeTokenResponse, error) {
	apiURL := graphAPIBase + "/oauth/access_token"

	params := url.Values{}
	params.Set("client_id", c.AppID)
	params.Set("client_secret", c.AppSecret)
	params.Set("redirect_uri", redirectURI)
	params.Set("code", code)

	resp, err := http.PostForm(apiURL, params)
	if err != nil {
		return nil, fmt.Errorf("exchange code: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("exchange failed (%d): %s", resp.StatusCode, NormalizeError(body))
	}

	var result ExchangeTokenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return &result, nil
}

// GetLongLivedToken exchanges a short-lived token for a long-lived token (60 days).
func (c *Client) GetLongLivedToken(shortLivedToken string) (*ExchangeTokenResponse, error) {
	apiURL := graphAPIBase + "/oauth/access_token"

	params := url.Values{}
	params.Set("client_id", c.AppID)
	params.Set("client_secret", c.AppSecret)
	params.Set("grant_type", "fb_exchange_token")
	params.Set("fb_exchange_token", shortLivedToken)

	resp, err := c.httpClient.Get(apiURL + "?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("get long-lived token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("long-lived token exchange failed: %s", NormalizeError(body))
	}

	var result ExchangeTokenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return &result, nil
}

// GetManagedPages returns the list of pages managed by the user.
func (c *Client) GetManagedPages(userAccessToken string) ([]PageAccount, error) {
	apiURL := graphAPIBase + "/me/accounts"

	params := url.Values{}
	params.Set("access_token", userAccessToken)

	resp, err := c.httpClient.Get(apiURL + "?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("get pages: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get pages failed (%d): %s", resp.StatusCode, NormalizeError(body))
	}

	var result struct {
		Data []PageAccount `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return result.Data, nil
}

// SubscribeApp subscribes the app to a page to receive webhooks.
func (c *Client) SubscribeApp(pageID, pageAccessToken string) error {
	apiURL := fmt.Sprintf("%s/%s/subscribed_apps", graphAPIBase, pageID)

	params := url.Values{}
	params.Set("access_token", pageAccessToken)
	params.Set("subscribed_fields", "messages,messaging_postbacks,message_reads,message_deliveries,feed")

	req, err := http.NewRequest("POST", apiURL, bytes.NewBufferString(params.Encode()))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	body, err := c.doRequest(req)
	if err != nil {
		return fmt.Errorf("subscribe app: %w", err)
	}

	var result struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal(body, &result); err != nil || !result.Success {
		return fmt.Errorf("subscribe failed: %s", string(body))
	}

	return nil
}

// GetPageInfo retrieves page information.
func (c *Client) GetPageInfo(pageID, pageAccessToken string) (map[string]interface{}, error) {
	apiURL := graphAPIBase + "/" + pageID

	params := url.Values{}
	params.Set("access_token", pageAccessToken)
	params.Set("fields", "id,name,category,picture,username")

	resp, err := c.httpClient.Get(apiURL + "?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("get page info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get page info failed (%d): %s", resp.StatusCode, NormalizeError(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return result, nil
}

// IsPageSubscribed checks if the app is subscribed to the page.
func (c *Client) IsPageSubscribed(pageID, pageAccessToken string) (bool, error) {
	apiURL := fmt.Sprintf("%s/%s/subscribed_apps", graphAPIBase, pageID)

	params := url.Values{}
	params.Set("access_token", pageAccessToken)

	resp, err := c.httpClient.Get(apiURL + "?" + params.Encode())
	if err != nil {
		return false, fmt.Errorf("check subscription: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("check subscription failed: %s", NormalizeError(body))
	}

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return false, fmt.Errorf("parse response: %w", err)
	}

	for _, app := range result.Data {
		if app.ID == c.AppID {
			return true, nil
		}
	}

	return false, nil
}

// DebugToken debugs an access token.
func (c *Client) DebugToken(accessToken string) (map[string]interface{}, error) {
	apiURL := graphAPIBase + "/debug_token"

	params := url.Values{}
	params.Set("input_token", accessToken)
	params.Set("access_token", c.AppID+"|"+c.AppSecret)

	resp, err := c.httpClient.Get(apiURL + "?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("debug token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("debug token failed: %s", NormalizeError(body))
	}

	var result struct {
		Data map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return result.Data, nil
}

// doRequest executes an HTTP request and returns the response body.
func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed (%d): %s", resp.StatusCode, NormalizeError(body))
	}

	return body, nil
}

// ParseWebhookPayload parses a raw webhook body into a structured WebhookPayload.
func ParseWebhookPayload(rawBody []byte) (*WebhookPayload, error) {
	var payload WebhookPayload
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		return nil, fmt.Errorf("parse webhook payload: %w", err)
	}
	return &payload, nil
}

// ExtractMessages extracts normalized WebhookMessages from a Facebook webhook payload.
func ExtractMessages(payload *WebhookPayload) []platform.WebhookMessage {
	var messages []platform.WebhookMessage

	for _, entry := range payload.Entry {
		for _, m := range entry.Messaging {
			if m.Message == nil {
				continue
			}
			// Skip echo messages (messages sent by the page itself)
			if m.Message.IsEcho {
				continue
			}

			msg := platform.WebhookMessage{
				Platform:         "facebook",
				PagePlatformID:   strings.TrimSpace(m.Recipient.ID),
				SenderPlatformID: strings.TrimSpace(m.Sender.ID),
				MessageID:        strings.TrimSpace(m.Message.Mid),
				Timestamp:        m.Timestamp,
			}

			// Determine content type and extract content
			if m.Message.Text != "" {
				msg.ContentType = "text"
				msg.Text = m.Message.Text
			} else if len(m.Message.Attachments) > 0 {
				att := m.Message.Attachments[0]
				msg.ContentType = att.Type
				msg.MediaURL = att.Payload.URL
				if att.Payload.StickerID != 0 {
					msg.ContentType = "sticker"
				}
				// Set a descriptive text for non-text messages
				switch att.Type {
				case "image":
					msg.Text = "[Hình ảnh]"
				case "video":
					msg.Text = "[Video]"
				case "audio":
					msg.Text = "[Âm thanh]"
				case "file":
					msg.Text = "[Tệp đính kèm]"
				default:
					msg.Text = "[Đa phương tiện]"
				}
			} else {
				msg.ContentType = "text"
				msg.Text = "[Tin nhắn không xác định]"
			}

			messages = append(messages, msg)
		}
	}

	return messages
}

// NormalizeError handles common Facebook API errors and returns user-friendly messages.
func NormalizeError(body []byte) string {
	var errResp GraphAPIError
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error.Message != "" {
		msg := errResp.Error.Message
		if strings.Contains(msg, "does not exist") {
			return "Người dùng không tồn tại hoặc đã bị chặn."
		}
		if strings.Contains(msg, "permissions") {
			return "Thiếu quyền truy cập. Vui lòng kiểm tra lại phân quyền trang."
		}
		if strings.Contains(msg, "Invalid parameter") {
			return "Tham số không hợp lệ."
		}
		return msg
	}
	return string(body)
}

// Compile-time check that Client implements platform.MessagingClient.
var _ platform.MessagingClient = (*Client)(nil)
