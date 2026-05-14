package shopee

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
	"strconv"
	"strings"
	"time"

	"fly-box/backend/internal/platform"
)

// Client handles all Shopee Open Platform API calls and implements platform.MessagingClient.
type Client struct {
	PartnerID   string
	PartnerKey  string
	Host        string
	httpClient  *http.Client
}

// NewClient creates a new Shopee Open Platform client.
func NewClient(partnerID, partnerKey, host string) *Client {
	if host == "" {
		host = "https://partner.shopeemobile.com"
	}
	return &Client{
		PartnerID:  partnerID,
		PartnerKey: partnerKey,
		Host:       host,
		httpClient: &http.Client{},
	}
}

// PlatformName returns the platform identifier.
func (c *Client) PlatformName() string {
	return "shopee"
}

// --- platform.MessagingClient interface ---

// SendTextMessage sends a text message to a buyer via Shopee seller chat API.
func (c *Client) SendTextMessage(accessToken, buyerID, messageText string) (string, error) {
	// For Shopee, we need shop_id which is stored in the page record
	// This method will be called with accessToken = accessToken + ":" + shopID
	parts := strings.SplitN(accessToken, ":", 2)
	shopID := ""
	if len(parts) == 2 {
		shopID = parts[1]
		accessToken = parts[0]
	}

	apiPath := "/api/v2/sellerchat/send_message"
	timestamp := time.Now().Unix()

	// Build request body
	body := map[string]interface{}{
		"shop_id":  shopID,
		"buyer_id": buyerID,
		"message": map[string]interface{}{
			"type": "text",
			"content": map[string]string{
				"text": messageText,
			},
		},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal body: %w", err)
	}

	// Make signed request
	respBody, err := c.doSignedRequest("POST", apiPath, accessToken, shopID, timestamp, jsonBody)
	if err != nil {
		return "", fmt.Errorf("send message: %w", err)
	}

	var result SendMessageResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	if result.ResponseError.Code != 0 {
		return "", fmt.Errorf("send message failed: %s", result.ResponseError.Message)
	}

	return result.ResponseData.MessageID, nil
}

// GetUserProfile retrieves buyer profile info from Shopee.
func (c *Client) GetUserProfile(accessToken, buyerID string) (*platform.UserProfile, error) {
	// Return basic profile with buyer_id as name
	return &platform.UserProfile{
		ID:       buyerID,
		Name:     buyerID, // Shopee doesn't provide buyer name by default
		Platform: "shopee",
	}, nil
}

// VerifyWebhookSignature verifies the Shopee webhook signature.
// Shopee uses HMAC-SHA256 with partner_key.
func (c *Client) VerifyWebhookSignature(rawBody []byte, signature string) bool {
	if c.PartnerKey == "" || signature == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(c.PartnerKey))
	mac.Write(rawBody)
	expected := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expected), []byte(signature))
}

// --- Shopee-specific methods ---

// GenerateSignature generates HMAC-SHA256 signature for Shopee API requests.
// Base string: partner_id + api_path + timestamp + access_token + shop_id
func (c *Client) GenerateSignature(apiPath, accessToken, shopID string, timestamp int64) string {
	baseString := c.PartnerID + apiPath + strconv.FormatInt(timestamp, 10) + accessToken + shopID
	mac := hmac.New(sha256.New, []byte(c.PartnerKey))
	mac.Write([]byte(baseString))
	return hex.EncodeToString(mac.Sum(nil))
}

// doSignedRequest makes a signed request to Shopee API.
func (c *Client) doSignedRequest(method, apiPath, accessToken, shopID string, timestamp int64, body []byte) ([]byte, error) {
	signature := c.GenerateSignature(apiPath, accessToken, shopID, timestamp)

	reqURL := c.Host + apiPath
	req, err := http.NewRequest(method, reqURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Partner-ID", c.PartnerID)
	req.Header.Set("Timestamp", strconv.FormatInt(timestamp, 10))
	req.Header.Set("Access-Token", accessToken)
	req.Header.Set("Shop-ID", shopID)
	req.Header.Set("Signature", signature)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed (%d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// GetShopInfo retrieves shop information.
func (c *Client) GetShopInfo(accessToken, shopID string) (*ShopInfo, error) {
	apiPath := "/api/v2/shop/get_profile"
	timestamp := time.Now().Unix()

	body := map[string]interface{}{
		"shop_id": shopID,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal body: %w", err)
	}

	respBody, err := c.doSignedRequest("POST", apiPath, accessToken, shopID, timestamp, jsonBody)
	if err != nil {
		return nil, fmt.Errorf("get shop info: %w", err)
	}

	var result ShopInfoResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if result.ResponseError.Code != 0 {
		return nil, fmt.Errorf("get shop info failed: %s", result.ResponseError.Message)
	}

	return &result.ResponseData, nil
}

// GetConversationList retrieves the list of conversations.
func (c *Client) GetConversationList(accessToken, shopID string, pageSize, offset int) (*GetConversationListResponse, error) {
	apiPath := "/api/v2/sellerchat/get_conversation_list"
	timestamp := time.Now().Unix()

	body := map[string]interface{}{
		"shop_id":   shopID,
		"page_size": pageSize,
		"offset":    offset,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal body: %w", err)
	}

	respBody, err := c.doSignedRequest("POST", apiPath, accessToken, shopID, timestamp, jsonBody)
	if err != nil {
		return nil, fmt.Errorf("get conversation list: %w", err)
	}

	var result GetConversationListResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return &result, nil
}

// GetMessages retrieves messages from a conversation.
func (c *Client) GetMessages(accessToken, shopID, buyerID string, pageSize, offset int) (*GetMessagesResponse, error) {
	apiPath := "/api/v2/sellerchat/get_message"
	timestamp := time.Now().Unix()

	body := map[string]interface{}{
		"shop_id":   shopID,
		"buyer_id":  buyerID,
		"page_size": pageSize,
		"offset":    offset,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal body: %w", err)
	}

	respBody, err := c.doSignedRequest("POST", apiPath, accessToken, shopID, timestamp, jsonBody)
	if err != nil {
		return nil, fmt.Errorf("get messages: %w", err)
	}

	var result GetMessagesResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return &result, nil
}

// MarkConversationRead marks a conversation as read.
func (c *Client) MarkConversationRead(accessToken, shopID, buyerID string) error {
	apiPath := "/api/v2/sellerchat/read_conversation"
	timestamp := time.Now().Unix()

	body := map[string]interface{}{
		"shop_id":  shopID,
		"buyer_id": buyerID,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal body: %w", err)
	}

	respBody, err := c.doSignedRequest("POST", apiPath, accessToken, shopID, timestamp, jsonBody)
	if err != nil {
		return fmt.Errorf("mark read: %w", err)
	}

	var result APIResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}

	if result.ResponseError.Code != 0 {
		return fmt.Errorf("mark read failed: %s", result.ResponseError.Message)
	}

	return nil
}

// GetUnreadCount gets the unread conversation count.
func (c *Client) GetUnreadCount(accessToken, shopID string) (int, error) {
	apiPath := "/api/v2/sellerchat/unread_conversation_count"
	timestamp := time.Now().Unix()

	body := map[string]interface{}{
		"shop_id": shopID,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return 0, fmt.Errorf("marshal body: %w", err)
	}

	respBody, err := c.doSignedRequest("POST", apiPath, accessToken, shopID, timestamp, jsonBody)
	if err != nil {
		return 0, fmt.Errorf("get unread count: %w", err)
	}

	var result UnreadCountResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return 0, fmt.Errorf("parse response: %w", err)
	}

	if result.ResponseError.Code != 0 {
		return 0, fmt.Errorf("get unread count failed: %s", result.ResponseError.Message)
	}

	return result.ResponseData.Count, nil
}

// --- OAuth methods (Shopee uses a different OAuth flow) ---

// GetAuthorizationURL returns the Shopee authorization URL.
func (c *Client) GetAuthorizationURL(redirectURI, state string) string {
	baseURL := "https://partner.shopeemobile.com/api/v2/shop/auth_partner"
	params := url.Values{}
	params.Set("partner_id", c.PartnerID)
	params.Set("redirect", redirectURI)
	params.Set("state", state)
	return baseURL + "?" + params.Encode()
}

// Authorize authorizes the app for a shop.
// This is called after the user grants permission on Shopee's page.
func (c *Client) Authorize(shopID, accessToken string) (*AuthorizeResponse, error) {
	apiPath := "/api/v2/shop/auth_partner/cancel"
	timestamp := time.Now().Unix()

	body := map[string]interface{}{
		"shop_id": shopID,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal body: %w", err)
	}

	respBody, err := c.doSignedRequest("POST", apiPath, accessToken, shopID, timestamp, jsonBody)
	if err != nil {
		return nil, fmt.Errorf("authorize: %w", err)
	}

	var result AuthorizeResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return &result, nil
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

// ExtractMessages extracts normalized WebhookMessages from a Shopee webhook payload.
func ExtractMessages(payload *WebhookPayload) []platform.WebhookMessage {
	var messages []platform.WebhookMessage

	for _, event := range payload.Events {
		if event.Topic != "shopee.chat" {
			continue
		}

		data, ok := event.Data.(map[string]interface{})
		if !ok {
			continue
		}

		// Extract conversation info
		shopIDStr, _ := data["shop_id"].(string)
		buyerID, _ := data["buyer_id"].(string)
		messageID, _ := data["message_id"].(string)
		messageType, _ := data["message_type"].(string)
		
		// Extract message content
		var text, mediaURL, contentType string
		
		switch messageType {
		case "text":
			contentType = "text"
			if content, ok := data["content"].(map[string]interface{}); ok {
				text, _ = content["text"].(string)
			}
		case "image":
			contentType = "image"
			text = "[Hình ảnh Shopee]"
			if content, ok := data["content"].(map[string]interface{}); ok {
				mediaURL, _ = content["url"].(string)
			}
		case "video":
			contentType = "video"
			text = "[Video Shopee]"
			if content, ok := data["content"].(map[string]interface{}); ok {
				mediaURL, _ = content["url"].(string)
			}
		case "audio":
			contentType = "audio"
			text = "[Âm thanh Shopee]"
			if content, ok := data["content"].(map[string]interface{}); ok {
				mediaURL, _ = content["url"].(string)
			}
		case "product_card":
			contentType = "product"
			text = "[Thẻ sản phẩm Shopee]"
		case "order_card":
			contentType = "order"
			text = "[Thẻ đơn hàng Shopee]"
		default:
			contentType = "text"
			text = "[Tin nhắn Shopee]"
		}

		if text == "" {
			text = "[Tin nhắn không xác định]"
		}

		// Get timestamp
		var timestamp int64
		if ts, ok := data["create_time"].(float64); ok {
			timestamp = int64(ts)
		}

		messages = append(messages, platform.WebhookMessage{
			Platform:         "shopee",
			PagePlatformID:   shopIDStr,
			SenderPlatformID: buyerID,
			MessageID:        messageID,
			Text:             text,
			ContentType:      contentType,
			MediaURL:         mediaURL,
			Timestamp:        timestamp,
		})
	}

	return messages
}

// NormalizeError handles Shopee API errors and returns user-friendly messages.
func NormalizeError(errResp *ShopeeError) string {
	if errResp == nil {
		return ""
	}
	
	msg := errResp.Message
	if strings.Contains(msg, "invalid_signature") {
		return "Chữ ký API không hợp lệ."
	}
	if strings.Contains(msg, "expired") {
		return "Phiên đăng nhập đã hết hạn."
	}
	if strings.Contains(msg, "permission") {
		return "Không có quyền truy cập API này."
	}
	return msg
}

// Compile-time check that Client implements platform.MessagingClient.
var _ platform.MessagingClient = (*Client)(nil)