package instagram

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

// Client handles all Instagram Graph API calls and implements platform.MessagingClient.
type Client struct {
	AppID      string
	AppSecret  string
	httpClient *http.Client
}

// NewClient creates a new Instagram Graph API client.
func NewClient(appID, appSecret string) *Client {
	return &Client{
		AppID:      appID,
		AppSecret:  appSecret,
		httpClient: &http.Client{},
	}
}

// PlatformName returns the platform identifier.
func (c *Client) PlatformName() string {
	return "instagram"
}

// --- platform.MessagingClient interface ---

// SendTextMessage sends a text message via the Instagram Messaging API.
// For Instagram, we use the same endpoint as Facebook: /me/messages
// The accessToken should be a Page Access Token that has Instagram messaging permissions.
func (c *Client) SendTextMessage(accessToken, recipientID, text string) (string, error) {
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
	params.Set("access_token", accessToken)

	req, err := http.NewRequest("POST", apiURL+"?"+params.Encode(), bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("send message request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("send message failed (%d): %s", resp.StatusCode, NormalizeError(body))
	}

	var result SendMessageResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	return result.MessageID, nil
}

// GetUserProfile retrieves user profile info from Instagram.
// Note: Instagram uses IGSID (Instagram Graph API ID) for messaging.
func (c *Client) GetUserProfile(accessToken, igUserID string) (*platform.UserProfile, error) {
	apiURL := graphAPIBase + "/" + igUserID

	params := url.Values{}
	params.Set("access_token", accessToken)
	params.Set("fields", "id,username,name,profile_picture_url,account_type,follows_count,followers_count")

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
		ID                string `json:"id"`
		Username          string `json:"username"`
		Name              string `json:"name"`
		ProfilePictureURL string `json:"profile_picture_url"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	name := raw.Name
	if name == "" {
		name = raw.Username
	}

	return &platform.UserProfile{
		ID:       igUserID,
		Name:     name,
		Avatar:   raw.ProfilePictureURL,
		Platform: "instagram",
	}, nil
}

// VerifyWebhookSignature verifies the X-Hub-Signature-256 header.
// Instagram uses the same signature verification as Facebook (same Meta App).
func (c *Client) VerifyWebhookSignature(rawBody []byte, signature string) bool {
	if c.AppSecret == "" || signature == "" {
		return false
	}

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

// --- Instagram-specific OAuth methods ---

// ExchangeCodeForToken exchanges an authorization code for a user access token.
// For Instagram, this is the same as Facebook OAuth since they're in the same Meta ecosystem.
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

// GetLongLivedToken exchanges a short-lived token for a long-lived token.
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

// GetManagedPages returns the list of Facebook Pages managed by the user.
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

// GetInstagramAccounts returns the Instagram Business Accounts linked to a Facebook Page.
// This requires the page access token and the page must have a linked IG account.
func (c *Client) GetInstagramAccounts(pageAccessToken, pageID string) ([]IGAccount, error) {
	apiURL := graphAPIBase + "/" + pageID

	params := url.Values{}
	params.Set("access_token", pageAccessToken)
	params.Set("fields", "instagram_business_account{id,username,name,profile_picture_url,account_type}")

	resp, err := c.httpClient.Get(apiURL + "?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("get instagram accounts: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get instagram accounts failed (%d): %s", resp.StatusCode, NormalizeError(body))
	}

	var result struct {
		IGBusinessAccount *struct {
			ID        string `json:"id"`
			Username  string `json:"username"`
			Name      string `json:"name"`
			ProfilePic string `json:"profile_picture_url"`
		} `json:"instagram_business_account"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	var accounts []IGAccount
	if result.IGBusinessAccount != nil {
		accounts = append(accounts, IGAccount{
			ID:         result.IGBusinessAccount.ID,
			Username:   result.IGBusinessAccount.Username,
			Name:       result.IGBusinessAccount.Name,
			ProfilePic: result.IGBusinessAccount.ProfilePic,
		})
	}

	return accounts, nil
}

// SubscribeAppToIG subscribes the app to receive Instagram webhooks for an IG account.
func (c *Client) SubscribeAppToIG(igUserID, pageAccessToken string) error {
	apiURL := graphAPIBase + "/" + igUserID + "/subscribed_apps"

	params := url.Values{}
	params.Set("access_token", pageAccessToken)
	params.Set("subscribed_fields", "messages,messaging_postbacks,message_reactions,message_reads,message_deliveries")

	req, err := http.NewRequest("POST", apiURL, bytes.NewBufferString(params.Encode()))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("subscribe app: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("subscribe failed (%d): %s", resp.StatusCode, NormalizeError(body))
	}

	return nil
}

// GetIGAccountInfo retrieves detailed information about an Instagram Business Account.
func (c *Client) GetIGAccountInfo(igUserID, accessToken string) (*IGAccount, error) {
	apiURL := graphAPIBase + "/" + igUserID

	params := url.Values{}
	params.Set("access_token", accessToken)
	params.Set("fields", "id,username,name,profile_picture_url,account_type,follows_count,followers_count,biography,website")

	resp, err := c.httpClient.Get(apiURL + "?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("get ig account info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get ig account info failed (%d): %s", resp.StatusCode, NormalizeError(body))
	}

	var raw struct {
		ID                string `json:"id"`
		Username          string `json:"username"`
		Name              string `json:"name"`
		ProfilePictureURL string `json:"profile_picture_url"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return &IGAccount{
		ID:         raw.ID,
		Username:   raw.Username,
		Name:       raw.Name,
		ProfilePic: raw.ProfilePictureURL,
	}, nil
}

// --- Webhook parsing ---

// ParseWebhookPayload parses a raw webhook body into a structured WebhookPayload.
func ParseWebhookPayload(rawBody []byte) (*WebhookPayload, error) {
	var payload WebhookPayload
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		return nil, fmt.Errorf("parse webhook payload: %w", err)
	}
	return &payload, nil
}

// ExtractMessages extracts normalized WebhookMessages from an Instagram webhook payload.
func ExtractMessages(payload *WebhookPayload) []platform.WebhookMessage {
	var messages []platform.WebhookMessage

	for _, entry := range payload.Entry {
		// entry.ID is the Instagram Business Account ID for Instagram webhooks
		igAccountID := strings.TrimSpace(entry.ID)

		for _, m := range entry.Messaging {
			if m.Message == nil {
				continue
			}
			// Skip echo messages (messages sent by the page itself)
			if m.Message.IsEcho {
				continue
			}

			// Determine Page ID: prefer recipient.id for incoming messages, fallback to entry.ID
			finalPageID := strings.TrimSpace(m.Recipient.ID)
			if finalPageID == "" {
				finalPageID = igAccountID
			}

			msg := platform.WebhookMessage{
				Platform:         "instagram",
				PagePlatformID:   finalPageID,
				SenderPlatformID: strings.TrimSpace(m.Sender.ID),
				MessageID:        strings.TrimSpace(m.Message.Mid),
				Timestamp:        m.Timestamp,
			}

			if m.Message.Text != "" {
				msg.ContentType = "text"
				msg.Text = m.Message.Text
			} else if len(m.Message.Attachments) > 0 {
				att := m.Message.Attachments[0]
				msg.ContentType = att.Type
				msg.MediaURL = att.Payload.URL

				switch att.Type {
				case "image":
					msg.Text = "[Hình ảnh Instagram]"
				case "video":
					msg.Text = "[Video Instagram]"
				case "audio":
					msg.Text = "[Âm thanh Instagram]"
				case "file":
					msg.Text = "[Tệp đính kèm Instagram]"
				default:
					msg.Text = "[Đa phương tiện Instagram]"
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

// --- Helper Methods ---

// NormalizeError handles common Meta Graph API errors and returns user-friendly messages.
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

// PageAccount represents a Facebook Page account (used for getting IG accounts).
type PageAccount struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Category    string   `json:"category"`
	Tasks       []string `json:"tasks"`
	AccessToken string   `json:"access_token"`
}