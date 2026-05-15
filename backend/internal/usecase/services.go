package usecase

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"fly-box/backend/internal/models"
		fb "fly-box/backend/internal/platform/facebook"
	"fly-box/backend/internal/platform/instagram"
	"fly-box/backend/internal/platform/shopee"
	"fly-box/backend/internal/platform/tiktok"
	"fly-box/backend/internal/platform/zalo"
	"fly-box/backend/internal/repository"
)

type Service struct {
	Repo          *repository.Repository
	FBClient      *fb.Client
	FBAppID       string
	FBAppSecret   string
	FBRedirectURI string

	// Legacy compatibility: keep FBGraphAPI as alias for FBClient
	FBGraphAPI *fb.Client

	// Zalo OA client
	ZaloClient      *zalo.Client
	ZaloAppID       string
	ZaloAppSecret   string
	ZaloRedirectURI string

	// TikTok Shop client
	TikTokClient      *tiktok.Client
	TikTokAppKey      string
	TikTokAppSecret   string
	TikTokRedirectURI string

	// Instagram client
	InstagramClient *instagram.Client

	// Shopee client
	ShopeeClient      *shopee.Client
	ShopeePartnerID   string
	ShopeePartnerKey  string
	ShopeeRedirectURI string
}

func New(repo *repository.Repository, fbAppID, fbAppSecret, fbRedirectURI, zaloAppID, zaloAppSecret, zaloOASecretKey, zaloRedirectURI, tiktokAppKey, tiktokAppSecret, tiktokRedirectURI, shopeePartnerID, shopeePartnerKey, shopeeRedirectURI string) *Service {
	fbClient := fb.NewClient(fbAppID, fbAppSecret)
	zaloClient := zalo.NewClient(zaloAppID, zaloAppSecret, zaloOASecretKey)
	tiktokClient := tiktok.NewClient(tiktokAppKey, tiktokAppSecret)
	instagramClient := instagram.NewClient(fbAppID, fbAppSecret)
	shopeeClient := shopee.NewClient(shopeePartnerID, shopeePartnerKey, "")
	return &Service{
		Repo:              repo,
		FBClient:          fbClient,
		FBGraphAPI:        fbClient,
		FBAppID:           fbAppID,
		FBAppSecret:       fbAppSecret,
		FBRedirectURI:     fbRedirectURI,
		ZaloClient:        zaloClient,
		ZaloAppID:         zaloAppID,
		ZaloAppSecret:     zaloAppSecret,
		ZaloRedirectURI:   zaloRedirectURI,
		TikTokClient:      tiktokClient,
		TikTokAppKey:      tiktokAppKey,
		TikTokAppSecret:   tiktokAppSecret,
		TikTokRedirectURI: tiktokRedirectURI,
		InstagramClient:   instagramClient,
		ShopeeClient:      shopeeClient,
		ShopeePartnerID:   shopeePartnerID,
		ShopeePartnerKey:  shopeePartnerKey,
		ShopeeRedirectURI: shopeeRedirectURI,
	}
}

// ExchangeFacebookCode exchanges authorization code for page tokens
func (s *Service) ExchangeFacebookCode(code string) (*FacebookConnectResult, error) {
	log.Printf("[ExchangeFacebookCode] Starting exchange with code length: %d", len(code))
	log.Printf("[ExchangeFacebookCode] Redirect URI: %s", s.FBRedirectURI)
	
	// Step 1: Exchange code for user access token
	userTokenResp, err := s.FBClient.ExchangeCodeForToken(code, s.FBRedirectURI)
	if err != nil {
		log.Printf("[ExchangeFacebookCode] Exchange user token failed: %v", err)
		return nil, fmt.Errorf("exchange user token failed: %w", err)
	}
	log.Printf("[ExchangeFacebookCode] Got user access token, length: %d", len(userTokenResp.AccessToken))

	// Step 2: Get long-lived token (60 days)
	longLivedResp, err := s.FBClient.GetLongLivedToken(userTokenResp.AccessToken)
	if err != nil {
		log.Printf("[ExchangeFacebookCode] Long-lived token exchange failed, using short-lived: %v", err)
		// Fallback: use the short-lived token if long-lived exchange fails
		longLivedResp = userTokenResp
	}

	// Step 3: Get managed pages
	pages, err := s.FBClient.GetManagedPages(longLivedResp.AccessToken)
	if err != nil {
		log.Printf("[ExchangeFacebookCode] Get managed pages failed: %v", err)
		return nil, fmt.Errorf("get managed pages failed: %w", err)
	}
	
	log.Printf("[ExchangeFacebookCode] Found %d pages", len(pages))
	for i, p := range pages {
		log.Printf("[ExchangeFacebookCode] Page %d: ID=%s, Name=%s, Tasks=%v", i+1, p.ID, p.Name, p.Tasks)
	}

	return &FacebookConnectResult{
		UserAccessToken: longLivedResp.AccessToken,
		Pages:           pages,
	}, nil
}

// FacebookConnectResult holds the result of Facebook OAuth flow
type FacebookConnectResult struct {
	UserAccessToken string
	Pages           []fb.PageAccount
}

// SaveFacebookPage saves a Facebook page with real tokens to database
func (s *Service) SaveFacebookPage(page fb.PageAccount, userAccessToken string) (*models.SocialPage, error) {
	// Determine permission level based on tasks
	permissionLevel := "manage"
	isAdmin := false
	for _, task := range page.Tasks {
		if task == "ADMINISTER" {
			isAdmin = true
			break
		}
	}
	if isAdmin {
		permissionLevel = "admin"
	}

	// Check if page already exists
	existing, err := s.Repo.GetPageByPageID("facebook", page.ID)
	if err == nil && existing != nil {
		// Update existing page
		existing.PageName = page.Name
		existing.AccessToken = page.AccessToken
		existing.ConnectionStatus = "connected"
		existing.PermissionLevel = permissionLevel
		existing.RequiresReauth = false
		existing.SupportsAdvancedTools = isAdmin
		if !isAdmin {
			existing.WarningMessage = "Chỉ có quyền Quản trị viên trang Facebook mới sử dụng được: lời chào, danh sách ủy quyền website, tin nhắn mở đầu, câu hỏi thường gặp, menu chính."
		} else {
			existing.WarningMessage = ""
		}
		if err := s.Repo.SavePage(existing); err != nil {
			return nil, fmt.Errorf("update page failed: %w", err)
		}
		return existing, nil
	}

	// Create new page
	newPage := &models.SocialPage{
		Platform:              "facebook",
		ExternalPageID:        page.ID,
		PageName:              page.Name,
		AccessToken:           page.AccessToken,
		Status:                "active",
		ConnectionStatus:      "connected",
		PermissionLevel:       permissionLevel,
		RequiresReauth:        false,
		SupportsAdvancedTools: isAdmin,
	}
	if !isAdmin {
		newPage.WarningMessage = "Chi co quyen Quan tri vien trang Facebook moi su dung duoc: loi chao, danh sach uy quyen website, tin nhan mo dau, cau hoi thuong gap, menu chinh."
	}

	if err := s.Repo.CreatePage(newPage); err != nil {
		return nil, fmt.Errorf("create page failed: %w", err)
	}

	return newPage, nil
}

// SubscribePageToWebhook subscribes the app to receive webhooks from a page
func (s *Service) SubscribePageToWebhook(page *models.SocialPage) error {
	if page.Platform != "facebook" || page.AccessToken == "" {
		return fmt.Errorf("invalid page for webhook subscription")
	}

	return s.FBClient.SubscribeApp(page.ExternalPageID, page.AccessToken)
}

// SendFacebookMessage sends a real message via Facebook Messenger API
func (s *Service) SendFacebookMessage(conversationID uint, customerPlatformID, messageText string) (*models.Message, error) {
	// Get conversation to get page info
	conv, err := s.Repo.GetConversationByID(conversationID)
	if err != nil {
		return nil, fmt.Errorf("conversation not found: %w", err)
	}

	// Get page to get access token
	page, err := s.Repo.GetPageByID(conv.PageID)
	if err != nil {
		return nil, fmt.Errorf("page not found: %w", err)
	}

	if page.Platform != "facebook" {
		return nil, fmt.Errorf("page is not a Facebook page")
	}

	if page.AccessToken == "" {
		return nil, fmt.Errorf("page has no access token")
	}

	// Send message via Facebook API
	messageID, err := s.FBClient.SendTextMessage(page.AccessToken, customerPlatformID, messageText)
	if err != nil {
		return nil, fmt.Errorf("failed to send Facebook message: %w", err)
	}

	// Create message record in database
	msg := &models.Message{
		ConversationID:  conversationID,
		SenderType:      "page",
		ContentType:     "text",
		Content:         messageText,
		SocialMessageID: messageID,
		CreatedAt:       time.Now(),
	}

	if err := s.Repo.CreateMessage(msg); err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	// Update conversation
	conv.LastMessage = messageText
	conv.UpdatedAt = time.Now()
	_ = s.Repo.SaveConversation(conv)

	return msg, nil
}

func (s *Service) MatchAutoReply(pageID uint, incoming string) (*models.AutoReplyRule, error) {
	rules, err := s.Repo.ListAutoReplyRules(pageID)
	if err != nil {
		return nil, err
	}

	var defaultRule *models.AutoReplyRule
	content := strings.TrimSpace(strings.ToLower(incoming))

	for i := range rules {
		r := rules[i]
		if r.RuleType == "default" {
			defaultRule = &r
			continue
		}

		var keywords []string
		_ = json.Unmarshal(r.Keywords, &keywords)

		for _, kw := range keywords {
			kwNorm := strings.TrimSpace(strings.ToLower(kw))
			if kwNorm == "" {
				continue
			}

			switch r.RuleType {
			case "exact_match":
				if content == kwNorm {
					return &r, nil
				}
			case "contains":
				if strings.Contains(content, kwNorm) {
					return &r, nil
				}
			}
		}
	}

	return defaultRule, nil
}

func (s *Service) CreateAgentMessage(conversationID uint, content string) (*models.Message, error) {
	msg := &models.Message{
		ConversationID: conversationID,
		SenderType:     "page",
		ContentType:    "text",
		Content:        content,
		CreatedAt:      time.Now(),
	}
	if err := s.Repo.CreateMessage(msg); err != nil {
		return nil, err
	}

	conv, err := s.Repo.GetConversationByID(conversationID)
	if err == nil {
		conv.LastMessage = content
		conv.UpdatedAt = time.Now()
		_ = s.Repo.SaveConversation(conv)
	}

	return msg, nil
}

func (s *Service) CreateSystemMessage(conversationID uint, content string) (*models.Message, error) {
	msg := &models.Message{
		ConversationID: conversationID,
		SenderType:     "system",
		ContentType:    "text",
		Content:        content,
		CreatedAt:      time.Now(),
	}
	if err := s.Repo.CreateMessage(msg); err != nil {
		return nil, err
	}

	conv, err := s.Repo.GetConversationByID(conversationID)
	if err == nil {
		conv.LastMessage = content
		conv.UpdatedAt = time.Now()
		_ = s.Repo.SaveConversation(conv)
	}

	return msg, nil
}

// --- Zalo OA methods ---

// ExchangeZaloCode exchanges a Zalo authorization code for OA tokens.
func (s *Service) ExchangeZaloCode(code string) (*ZaloConnectResult, error) {
	tokenResp, err := s.ZaloClient.ExchangeCodeForToken(code, s.ZaloRedirectURI)
	if err != nil {
		return nil, fmt.Errorf("exchange zalo token failed: %w", err)
	}

	// Get OA info to retrieve OA name and ID
	oaInfo, err := s.ZaloClient.GetOAInfo(tokenResp.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("get zalo OA info failed: %w", err)
	}

	return &ZaloConnectResult{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		OAId:         oaInfo.Data.OAID,
		OAName:       oaInfo.Data.Name,
	}, nil
}

// ZaloConnectResult holds the result of Zalo OAuth flow.
type ZaloConnectResult struct {
	AccessToken  string
	RefreshToken string
	OAId         string
	OAName       string
}

// SaveZaloPage saves a Zalo OA page with real tokens to database.
func (s *Service) SaveZaloPage(oaID, oaName, accessToken, refreshToken string) (*models.SocialPage, error) {
	existing, err := s.Repo.GetPageByPageID("zalo", oaID)
	if err == nil && existing != nil {
		existing.PageName = oaName
		existing.AccessToken = accessToken
		existing.RefreshToken = refreshToken
		existing.ConnectionStatus = "connected"
		existing.RequiresReauth = false
		existing.PermissionLevel = "admin"
		existing.SupportsAdvancedTools = true
		existing.WarningMessage = ""
		if err := s.Repo.SavePage(existing); err != nil {
			return nil, fmt.Errorf("update zalo page failed: %w", err)
		}
		return existing, nil
	}

	newPage := &models.SocialPage{
		Platform:              "zalo",
		ExternalPageID:        oaID,
		PageName:              oaName,
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		Status:                "active",
		ConnectionStatus:      "connected",
		PermissionLevel:       "admin",
		RequiresReauth:        false,
		SupportsAdvancedTools: true,
	}

	if err := s.Repo.CreatePage(newPage); err != nil {
		return nil, fmt.Errorf("create zalo page failed: %w", err)
	}

	return newPage, nil
}

// SendZaloMessage sends a real message via Zalo OA API.
func (s *Service) SendZaloMessage(conversationID uint, customerPlatformID, messageText string) (*models.Message, error) {
	conv, err := s.Repo.GetConversationByID(conversationID)
	if err != nil {
		return nil, fmt.Errorf("conversation not found: %w", err)
	}

	page, err := s.Repo.GetPageByID(conv.PageID)
	if err != nil {
		return nil, fmt.Errorf("page not found: %w", err)
	}

	if page.Platform != "zalo" {
		return nil, fmt.Errorf("page is not a Zalo page")
	}

	if page.AccessToken == "" {
		return nil, fmt.Errorf("page has no access token")
	}

	messageID, err := s.ZaloClient.SendTextMessage(page.AccessToken, customerPlatformID, messageText)
	if err != nil {
		return nil, fmt.Errorf("failed to send Zalo message: %w", err)
	}

	msg := &models.Message{
		ConversationID:  conversationID,
		SenderType:      "page",
		ContentType:     "text",
		Content:         messageText,
		SocialMessageID: messageID,
		CreatedAt:       time.Now(),
	}

	if err := s.Repo.CreateMessage(msg); err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	conv.LastMessage = messageText
	conv.UpdatedAt = time.Now()
	_ = s.Repo.SaveConversation(conv)

	return msg, nil
}

// --- TikTok Shop methods ---

// TikTokConnectResult holds the result of TikTok OAuth flow.
type TikTokConnectResult struct {
	AccessToken  string
	RefreshToken string
	ShopID       string
	ShopName     string
	SellerName   string
}

// ExchangeTikTokCode exchanges a TikTok authorization code for shop tokens.
func (s *Service) ExchangeTikTokCode(code string) (*TikTokConnectResult, error) {
	tokenResp, err := s.TikTokClient.ExchangeCodeForToken(code)
	if err != nil {
		return nil, fmt.Errorf("exchange tiktok token failed: %w", err)
	}

	if tokenResp.Data == nil {
		return nil, fmt.Errorf("no token data returned from tiktok")
	}

	result := &TikTokConnectResult{
		AccessToken:  tokenResp.Data.AccessToken,
		RefreshToken: tokenResp.Data.RefreshToken,
		SellerName:   tokenResp.Data.SellerName,
	}

	// Try to get shop info
	shopInfo, err := s.TikTokClient.GetShopInfo(tokenResp.Data.AccessToken)
	if err == nil && shopInfo.Data != nil {
		result.ShopID = shopInfo.Data.ShopID
		result.ShopName = shopInfo.Data.ShopName
	} else {
		// Fallback: use open_id as shop identifier
		result.ShopID = tokenResp.Data.OpenID
		result.ShopName = tokenResp.Data.SellerName
	}

	return result, nil
}

// SaveTikTokShop saves a TikTok Shop page with real tokens to database.
func (s *Service) SaveTikTokShop(shopID, shopName, accessToken, refreshToken string) (*models.SocialPage, error) {
	existing, err := s.Repo.GetPageByPageID("tiktok", shopID)
	if err == nil && existing != nil {
		existing.PageName = shopName
		existing.AccessToken = accessToken
		existing.RefreshToken = refreshToken
		existing.ConnectionStatus = "connected"
		existing.RequiresReauth = false
		existing.PermissionLevel = "admin"
		existing.SupportsAdvancedTools = true
		existing.WarningMessage = ""
		if err := s.Repo.SavePage(existing); err != nil {
			return nil, fmt.Errorf("update tiktok shop failed: %w", err)
		}
		return existing, nil
	}

	newPage := &models.SocialPage{
		Platform:              "tiktok",
		ExternalPageID:        shopID,
		PageName:              shopName,
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		Status:                "active",
		ConnectionStatus:      "connected",
		PermissionLevel:       "admin",
		RequiresReauth:        false,
		SupportsAdvancedTools: true,
	}

	if err := s.Repo.CreatePage(newPage); err != nil {
		return nil, fmt.Errorf("create tiktok shop failed: %w", err)
	}

	return newPage, nil
}

// SendTikTokMessage sends a real message via TikTok Shop seller chat API.
func (s *Service) SendTikTokMessage(conversationID uint, customerPlatformID, messageText string) (*models.Message, error) {
	conv, err := s.Repo.GetConversationByID(conversationID)
	if err != nil {
		return nil, fmt.Errorf("conversation not found: %w", err)
	}

	page, err := s.Repo.GetPageByID(conv.PageID)
	if err != nil {
		return nil, fmt.Errorf("page not found: %w", err)
	}

	if page.Platform != "tiktok" {
		return nil, fmt.Errorf("page is not a TikTok page")
	}

	if page.AccessToken == "" {
		return nil, fmt.Errorf("page has no access token")
	}

	messageID, err := s.TikTokClient.SendTextMessage(page.AccessToken, customerPlatformID, messageText)
	if err != nil {
		return nil, fmt.Errorf("failed to send TikTok message: %w", err)
	}

	msg := &models.Message{
		ConversationID:  conversationID,
		SenderType:      "page",
		ContentType:     "text",
		Content:         messageText,
		SocialMessageID: messageID,
		CreatedAt:       time.Now(),
	}

	if err := s.Repo.CreateMessage(msg); err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	conv.LastMessage = messageText
	conv.UpdatedAt = time.Now()
	_ = s.Repo.SaveConversation(conv)

	return msg, nil
}

// --- Instagram methods ---

// InstagramConnectResult holds the result of Instagram OAuth flow.
type InstagramConnectResult struct {
	UserAccessToken string
	IGAccounts      []instagram.IGAccount
}

// ExchangeInstagramCode exchanges an Instagram authorization code for IG accounts.
// Instagram uses the same Meta OAuth as Facebook, but we need to fetch IG Business Accounts.
func (s *Service) ExchangeInstagramCode(code string) (*InstagramConnectResult, error) {
	// Step 1: Exchange code for user access token (same as Facebook)
	userTokenResp, err := s.InstagramClient.ExchangeCodeForToken(code, s.FBRedirectURI)
	if err != nil {
		return nil, fmt.Errorf("exchange user token failed: %w", err)
	}

	// Step 2: Get long-lived token (60 days)
	longLivedResp, err := s.InstagramClient.GetLongLivedToken(userTokenResp.AccessToken)
	if err != nil {
		// Fallback: use the short-lived token if long-lived exchange fails
		longLivedResp = userTokenResp
	}

	// Step 3: Get managed pages
	pages, err := s.InstagramClient.GetManagedPages(longLivedResp.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("get managed pages failed: %w", err)
	}

	// Step 4: Find Instagram Business Accounts linked to these pages
	var igAccounts []instagram.IGAccount
	for _, page := range pages {
		igAccts, err := s.InstagramClient.GetInstagramAccounts(page.AccessToken, page.ID)
		if err != nil {
			continue
		}
		igAccounts = append(igAccounts, igAccts...)
	}

	return &InstagramConnectResult{
		UserAccessToken: longLivedResp.AccessToken,
		IGAccounts:      igAccounts,
	}, nil
}

// SaveInstagramPage saves an Instagram Business Account page with real tokens to database.
func (s *Service) SaveInstagramPage(igAccount instagram.IGAccount, pageAccessToken string) (*models.SocialPage, error) {
	// Determine display name
	pageName := igAccount.Name
	if pageName == "" {
		pageName = igAccount.Username
	}

	existing, err := s.Repo.GetPageByPageID("instagram", igAccount.ID)
	if err == nil && existing != nil {
		existing.PageName = pageName
		existing.AccessToken = pageAccessToken
		existing.ConnectionStatus = "connected"
		existing.RequiresReauth = false
		existing.PermissionLevel = "admin"
		existing.SupportsAdvancedTools = true
		existing.WarningMessage = ""
		if err := s.Repo.SavePage(existing); err != nil {
			return nil, fmt.Errorf("update instagram page failed: %w", err)
		}
		return existing, nil
	}

	newPage := &models.SocialPage{
		Platform:              "instagram",
		ExternalPageID:        igAccount.ID,
		PageName:              pageName,
		AccessToken:           pageAccessToken,
		Status:                "active",
		ConnectionStatus:      "connected",
		PermissionLevel:       "admin",
		RequiresReauth:        false,
		SupportsAdvancedTools: true,
	}

	if err := s.Repo.CreatePage(newPage); err != nil {
		return nil, fmt.Errorf("create instagram page failed: %w", err)
	}

	return newPage, nil
}

// SubscribeInstagramToWebhook subscribes the app to receive webhooks from an IG account.
func (s *Service) SubscribeInstagramToWebhook(page *models.SocialPage) error {
	if page.Platform != "instagram" || page.AccessToken == "" {
		return fmt.Errorf("invalid page for webhook subscription")
	}

	return s.InstagramClient.SubscribeAppToIG(page.ExternalPageID, page.AccessToken)
}

// SendInstagramMessage sends a real message via Instagram Messaging API.
func (s *Service) SendInstagramMessage(conversationID uint, customerPlatformID, messageText string) (*models.Message, error) {
	conv, err := s.Repo.GetConversationByID(conversationID)
	if err != nil {
		return nil, fmt.Errorf("conversation not found: %w", err)
	}

	page, err := s.Repo.GetPageByID(conv.PageID)
	if err != nil {
		return nil, fmt.Errorf("page not found: %w", err)
	}

	if page.Platform != "instagram" {
		return nil, fmt.Errorf("page is not an Instagram page")
	}

	if page.AccessToken == "" {
		return nil, fmt.Errorf("page has no access token")
	}

	messageID, err := s.InstagramClient.SendTextMessage(page.AccessToken, customerPlatformID, messageText)
	if err != nil {
		return nil, fmt.Errorf("failed to send Instagram message: %w", err)
	}

	msg := &models.Message{
		ConversationID:  conversationID,
		SenderType:      "page",
		ContentType:     "text",
		Content:         messageText,
		SocialMessageID: messageID,
		CreatedAt:       time.Now(),
	}

	if err := s.Repo.CreateMessage(msg); err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	conv.LastMessage = messageText
	conv.UpdatedAt = time.Now()
	_ = s.Repo.SaveConversation(conv)

	return msg, nil
}

// --- Shopee methods ---

// ShopeeConnectResult holds the result of Shopee OAuth flow.
type ShopeeConnectResult struct {
	ShopID      string
	ShopName    string
	AccessToken string
}

// ExchangeShopeeCode exchanges a Shopee authorization code for shop tokens.
// Note: Shopee uses a different OAuth flow - the code is obtained after user authorizes on Shopee's page.
func (s *Service) ExchangeShopeeCode(code, shopID string) (*ShopeeConnectResult, error) {
	// Shopee's authorization flow is different from OAuth2
	// The code is obtained from the redirect URL after user authorizes
	// For now, we'll use the code directly as the access token (simplified flow)
	// In production, you would exchange this code for a real access token
	
	result := &ShopeeConnectResult{
		ShopID:      shopID,
		AccessToken: code,
	}

	// Try to get shop info
	if shopID != "" && code != "" {
		shopInfo, err := s.ShopeeClient.GetShopInfo(code, shopID)
		if err == nil && shopInfo.ShopName != "" {
			result.ShopName = shopInfo.ShopName
		}
	}

	return result, nil
}

// SaveShopeePage saves a Shopee shop page with tokens to database.
func (s *Service) SaveShopeePage(shopID, shopName, accessToken string) (*models.SocialPage, error) {
	existing, err := s.Repo.GetPageByPageID("shopee", shopID)
	if err == nil && existing != nil {
		existing.PageName = shopName
		existing.AccessToken = accessToken
		existing.ConnectionStatus = "connected"
		existing.RequiresReauth = false
		existing.PermissionLevel = "admin"
		existing.SupportsAdvancedTools = true
		existing.WarningMessage = ""
		if err := s.Repo.SavePage(existing); err != nil {
			return nil, fmt.Errorf("update shopee page failed: %w", err)
		}
		return existing, nil
	}

	newPage := &models.SocialPage{
		Platform:              "shopee",
		ExternalPageID:        shopID,
		PageName:              shopName,
		AccessToken:           accessToken,
		Status:                "active",
		ConnectionStatus:      "connected",
		PermissionLevel:       "admin",
		RequiresReauth:        false,
		SupportsAdvancedTools: true,
	}

	if err := s.Repo.CreatePage(newPage); err != nil {
		return nil, fmt.Errorf("create shopee page failed: %w", err)
	}

	return newPage, nil
}

// SendShopeeMessage sends a real message via Shopee seller chat API.
func (s *Service) SendShopeeMessage(conversationID uint, customerPlatformID, messageText string) (*models.Message, error) {
	conv, err := s.Repo.GetConversationByID(conversationID)
	if err != nil {
		return nil, fmt.Errorf("conversation not found: %w", err)
	}

	page, err := s.Repo.GetPageByID(conv.PageID)
	if err != nil {
		return nil, fmt.Errorf("page not found: %w", err)
	}

	if page.Platform != "shopee" {
		return nil, fmt.Errorf("page is not a Shopee page")
	}

	if page.AccessToken == "" {
		return nil, fmt.Errorf("page has no access token")
	}

	// For Shopee, combine access token with shop_id for the API call
	accessTokenWithShopID := page.AccessToken + ":" + page.ExternalPageID

	messageID, err := s.ShopeeClient.SendTextMessage(accessTokenWithShopID, customerPlatformID, messageText)
	if err != nil {
		return nil, fmt.Errorf("failed to send Shopee message: %w", err)
	}

	msg := &models.Message{
		ConversationID:  conversationID,
		SenderType:      "page",
		ContentType:     "text",
		Content:         messageText,
		SocialMessageID: messageID,
		CreatedAt:       time.Now(),
	}

	if err := s.Repo.CreateMessage(msg); err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	conv.LastMessage = messageText
	conv.UpdatedAt = time.Now()
	_ = s.Repo.SaveConversation(conv)

	return msg, nil
}
