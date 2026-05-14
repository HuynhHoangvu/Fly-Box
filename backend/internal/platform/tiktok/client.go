package tiktok

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
	"sort"
	"strings"
	"time"

	"fly-box/backend/internal/platform"
)

const (
	authBaseURL = "https://auth.tiktok-shops.com"
	apiBaseURL  = "https://open-api.tiktokglobalshop.com"
)

// Client handles all TikTok Shop API calls and implements platform.MessagingClient.
type Client struct {
	AppKey     string
	AppSecret  string
	httpClient *http.Client
}

// NewClient creates a new TikTok Shop API client.
func NewClient(appKey, appSecret string) *Client {
	return &Client{
		AppKey:     appKey,
		AppSecret:  appSecret,
		httpClient: &http.Client{},
	}
}

// PlatformName returns the platform identifier.
func (c *Client) PlatformName() string {
	return "tiktok"
}

// --- platform.MessagingClient interface ---

// SendTextMessage sends a text message to a buyer via TikTok Shop seller chat API.
// accessToken is the shop access token, recipientID is the conversation_id.
func (c *Client) SendTextMessage(accessToken, recipientID, text string) (string, error) {
	path := "/api/customer_service/conversations/messages/send"

	payload := SendMessageRequest{
		ConversationID: recipientID,
		Type:           "TEXT",
		Content:        text,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal payload: %w", err)
	}

	body, err := c.doSignedRequest("POST", path, accessToken, "", jsonPayload)
	if err != nil {
		return "", fmt.Errorf("send message: %w", err)
	}

	var result SendMessageResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	if result.Code != 0 {
		return "", fmt.Errorf("tiktok API error (%d): %s", result.Code, result.Message)
	}

	messageID := ""
	if result.Data != nil {
		messageID = result.Data.MessageID
	}

	return messageID, nil
}

// GetUserProfile retrieves buyer profile info from a TikTok Shop conversation.
// For TikTok Shop, we get buyer info from the conversation list.
func (c *Client) GetUserProfile(accessToken, userID string) (*platform.UserProfile, error) {
	// TikTok Shop doesn't have a direct user profile API.
	// Buyer info comes from conversation metadata.
	// We return a basic profile with the user ID.
	return &platform.UserProfile{
		ID:       userID,
		Name:     userID,
		Platform: "tiktok",
	}, nil
}

// VerifyWebhookSignature verifies the authenticity of an incoming TikTok Shop webhook.
// TikTok uses HMAC-SHA256 with the app secret.
// The signature is sent in the Authorization header.
func (c *Client) VerifyWebhookSignature(rawBody []byte, signature string) bool {
	if c.AppSecret == "" || signature == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(c.AppSecret))
	mac.Write(rawBody)
	expected := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expected), []byte(signature))
}

// --- TikTok Shop-specific methods ---

// ExchangeCodeForToken exchanges an authorization code for shop access tokens.
func (c *Client) ExchangeCodeForToken(code string) (*TokenResponse, error) {
	apiURL := authBaseURL + "/api/v2/token/get"

	payload := map[string]string{
		"app_key":    c.AppKey,
		"app_secret": c.AppSecret,
		"auth_code":  code,
		"grant_type": "authorized_code",
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("exchange code: %w", err)
	}

	var result TokenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("tiktok OAuth error (%d): %s", result.Code, result.Message)
	}

	return &result, nil
}

// RefreshToken refreshes an expired access token.
func (c *Client) RefreshToken(refreshToken string) (*TokenResponse, error) {
	apiURL := authBaseURL + "/api/v2/token/refresh"

	payload := map[string]string{
		"app_key":       c.AppKey,
		"app_secret":    c.AppSecret,
		"refresh_token": refreshToken,
		"grant_type":    "refresh_token",
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	body, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("refresh token: %w", err)
	}

	var result TokenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("tiktok refresh error (%d): %s", result.Code, result.Message)
	}

	return &result, nil
}

// GetShopInfo retrieves the shop info using the access token.
func (c *Client) GetShopInfo(accessToken string) (*ShopInfoResponse, error) {
	path := "/api/shop/get_authorized_shop"

	body, err := c.doSignedRequest("GET", path, accessToken, "", nil)
	if err != nil {
		return nil, fmt.Errorf("get shop info: %w", err)
	}

	var result ShopInfoResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("tiktok API error (%d): %s", result.Code, result.Message)
	}

	return &result, nil
}

// --- Webhook parsing ---

// ParseWebhookPayload parses raw JSON bytes into a TikTok WebhookPayload.
func ParseWebhookPayload(rawBody []byte) (*WebhookPayload, error) {
	var payload WebhookPayload
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		return nil, fmt.Errorf("parse tiktok webhook: %w", err)
	}
	return &payload, nil
}

// ExtractMessages normalizes a TikTok Shop webhook payload into platform.WebhookMessage slice.
func ExtractMessages(payload *WebhookPayload) []platform.WebhookMessage {
	if payload == nil || payload.Data == "" {
		return nil
	}

	var msgData WebhookMessageData
	if err := json.Unmarshal([]byte(payload.Data), &msgData); err != nil {
		return nil
	}

	// Only process buyer messages
	if msgData.SenderType != "" && msgData.SenderType != "BUYER" {
		return nil
	}

	if msgData.ConversationID == "" {
		return nil
	}

	contentType := "text"
	text := msgData.Content
	mediaURL := ""

	switch strings.ToUpper(msgData.MessageType) {
	case "IMAGE":
		contentType = "image"
		text = "[Hình ảnh]"
		mediaURL = msgData.MediaURL
	case "VIDEO":
		contentType = "video"
		text = "[Video]"
		mediaURL = msgData.MediaURL
	case "ORDER_CARD":
		contentType = "text"
		text = "[Thẻ đơn hàng]"
	case "PRODUCT_CARD":
		contentType = "text"
		text = "[Thẻ sản phẩm]"
	case "TEXT", "":
		contentType = "text"
		if text == "" {
			text = "[Tin nhắn không xác định]"
		}
	default:
		contentType = "text"
		if text == "" {
			text = fmt.Sprintf("[%s]", msgData.MessageType)
		}
	}

	messages := []platform.WebhookMessage{
		{
			Platform:         "tiktok",
			PagePlatformID:   payload.ShopID,
			SenderPlatformID: msgData.BuyerUserID,
			MessageID:        msgData.MessageID,
			Text:             text,
			ContentType:      contentType,
			MediaURL:         mediaURL,
			Timestamp:        msgData.CreateTime,
		},
	}

	return messages
}

// --- Internal helpers ---

// generateSign creates the HMAC-SHA256 signature for TikTok Shop API requests.
// Sign = HMAC-SHA256(app_secret, path + sorted_query_params + body)
func (c *Client) generateSign(path string, params url.Values, body []byte) string {
	// Sort parameter keys
	keys := make([]string, 0, len(params))
	for k := range params {
		// Exclude access_token and sign from signature
		if k == "access_token" || k == "sign" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build base string: path + key1value1key2value2... + body
	var baseString strings.Builder
	baseString.WriteString(path)
	for _, k := range keys {
		baseString.WriteString(k)
		baseString.WriteString(params.Get(k))
	}
	if len(body) > 0 {
		baseString.Write(body)
	}

	// Wrap with app_secret
	signBase := c.AppSecret + baseString.String() + c.AppSecret

	mac := hmac.New(sha256.New, []byte(c.AppSecret))
	mac.Write([]byte(signBase))
	return hex.EncodeToString(mac.Sum(nil))
}

// doSignedRequest performs an authenticated API request with TikTok Shop signature.
func (c *Client) doSignedRequest(method, path, accessToken, shopID string, body []byte) ([]byte, error) {
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	params := url.Values{}
	params.Set("app_key", c.AppKey)
	params.Set("timestamp", timestamp)
	if accessToken != "" {
		params.Set("access_token", accessToken)
	}
	if shopID != "" {
		params.Set("shop_id", shopID)
	}

	sign := c.generateSign(path, params, body)
	params.Set("sign", sign)

	fullURL := apiBaseURL + path + "?" + params.Encode()

	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewBuffer(body)
	}

	req, err := http.NewRequest(method, fullURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.doRequest(req)
}

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
