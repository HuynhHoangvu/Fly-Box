package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"fly-box/backend/internal/delivery/http/middlewares"
	ws "fly-box/backend/internal/delivery/websocket"
	"fly-box/backend/internal/models"
	fb "fly-box/backend/internal/platform/facebook"
	igplatform "fly-box/backend/internal/platform/instagram"
	shopeeplatform "fly-box/backend/internal/platform/shopee"
	tiktokplatform "fly-box/backend/internal/platform/tiktok"
 zaloplatform "fly-box/backend/internal/platform/zalo"
	"fly-box/backend/internal/repository"
	"fly-box/backend/internal/usecase"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Controller struct {
	Repo                  *repository.Repository
	Svc                   *usecase.Service
	JWT                   *middlewares.JWTManager
	Hub                   *ws.Hub
	DB                    *gorm.DB
	FBToken               string
	TTToken               string
	IGToken               string
	FacebookAppID         string
	FacebookAppSecret     string
	FacebookRedirectURI   string
	FrontendURL           string
	ZaloAppID             string
	ZaloAppSecret         string
	ZaloOASecretKey       string
	ZaloRedirectURI       string
	TikTokAppKey          string
	TikTokAppSecret       string
	TikTokRedirectURI     string
	ShopeePartnerID       string
	ShopeePartnerKey      string
	ShopeeRedirectURI     string
	ShopeeVerifyToken     string
}

func (ctl *Controller) VerifyInstagramWebhook(c *gin.Context) {
	mode := c.Query("hub.mode")
	token := c.Query("hub.verify_token")
	challenge := c.Query("hub.challenge")

	if mode == "subscribe" && token == ctl.IGToken {
		c.String(http.StatusOK, challenge)
		return
	}
	c.JSON(http.StatusForbidden, gin.H{"error": "invalid verify token"})
}

func (ctl *Controller) InstagramWebhook(c *gin.Context) {
	// Read raw body for signature verification
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read body"})
		return
	}

	// Verify webhook signature (X-Hub-Signature-256)
	signature := c.GetHeader("X-Hub-Signature-256")
	if signature != "" {
		igClient := igplatform.NewClient(ctl.FacebookAppID, ctl.FacebookAppSecret)
		if !igClient.VerifyWebhookSignature(rawBody, signature) {
			log.Printf("[instagram-webhook] signature verification failed")
			c.JSON(http.StatusForbidden, gin.H{"error": "invalid signature"})
			return
		}
	}

	// Parse webhook payload
	payload, err := igplatform.ParseWebhookPayload(rawBody)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	// Extract normalized messages
	messages := igplatform.ExtractMessages(payload)

	processed := 0
	for _, m := range messages {
		if m.PagePlatformID == "" || m.SenderPlatformID == "" {
			continue
		}

		page, err := ctl.Repo.GetPageByPageID("instagram", m.PagePlatformID)
		if err != nil || page == nil {
			continue
		}

		customer, err := ctl.Repo.GetOrCreateCustomer("instagram", m.SenderPlatformID)
		if err != nil || customer == nil {
			continue
		}

		// Enrich customer name from Instagram profile if still using platform ID as name
		if customer.Name == m.SenderPlatformID && page.AccessToken != "" {
			go ctl.enrichIGCustomerName(page.AccessToken, customer)
		}

		conv, err := ctl.Repo.GetOrCreateConversation(page.ID, customer.ID)
		if err != nil || conv == nil {
			continue
		}

		// Deduplicate by social message ID
		if m.MessageID != "" {
			if _, err := ctl.Repo.GetMessageBySocialMessageID(m.MessageID); err == nil {
				continue
			}
		}

		content := strings.TrimSpace(m.Text)
		if content == "" {
			content = "[Tin nhắn không xác định]"
		}

		msg := &models.Message{
			ConversationID:  conv.ID,
			SenderType:      "customer",
			ContentType:     m.ContentType,
			Content:         content,
			SocialMessageID: m.MessageID,
			CreatedAt:       time.Now(),
		}
		if err := ctl.Repo.CreateMessage(msg); err != nil {
			continue
		}

		conv.LastMessage = content
		conv.UnreadCount = conv.UnreadCount + 1
		conv.UpdatedAt = time.Now()
		_ = ctl.Repo.SaveConversation(conv)

		ctl.Hub.Broadcast(page.ID, ws.Event{
			Type:   "NEW_MESSAGE",
			PageID: page.ID,
			Data:   msg,
		})

		// Auto-reply check (async)
		go func(pageID uint, convID uint, customerPID string, pageAccessToken string, incomingText string) {
			rule, err := ctl.Svc.MatchAutoReply(pageID, incomingText)
			if err != nil || rule == nil {
				return
			}

			// Send auto-reply via Instagram API
			if pageAccessToken != "" {
				_, err := ctl.Svc.InstagramClient.SendTextMessage(pageAccessToken, customerPID, rule.ReplyContent)
				if err != nil {
					log.Printf("[auto-reply-instagram] failed to send via IG API: %v", err)
				}
			}

			// Save as system message
			sysMsg, err := ctl.Svc.CreateSystemMessage(convID, rule.ReplyContent)
			if err != nil {
				log.Printf("[auto-reply-instagram] failed to save system message: %v", err)
				return
			}

			ctl.Hub.Broadcast(pageID, ws.Event{
				Type:   "NEW_MESSAGE",
				PageID: pageID,
				Data:   sysMsg,
			})
		}(page.ID, conv.ID, m.SenderPlatformID, page.AccessToken, content)

		processed++
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "received",
		"platform":  "instagram",
		"processed": processed,
	})
}

// enrichIGCustomerName fetches the customer's real name from Instagram and updates the record.
func (ctl *Controller) enrichIGCustomerName(pageAccessToken string, customer *models.Customer) {
	igClient := igplatform.NewClient(ctl.FacebookAppID, ctl.FacebookAppSecret)
	profile, err := igClient.GetUserProfile(pageAccessToken, customer.PlatformID)
	if err != nil {
		log.Printf("[enrich-instagram] failed to get profile for %s: %v", customer.PlatformID, err)
		return
	}

	if profile.Name != "" && profile.Name != customer.Name {
		customer.Name = profile.Name
		if profile.Avatar != "" {
			customer.Avatar = profile.Avatar
		}
		customer.UpdatedAt = time.Now()
		_ = ctl.Repo.SaveCustomer(customer)
	}
}

func (ctl *Controller) VerifyTikTokWebhook(c *gin.Context) {
	challenge := c.Query("hub.challenge")
	token := c.Query("hub.verify_token")

	if token == ctl.TTToken {
		c.String(http.StatusOK, challenge)
		return
	}
	c.JSON(http.StatusForbidden, gin.H{"error": "invalid verify token"})
}

func (ctl *Controller) TikTokWebhook(c *gin.Context) {
	// Read raw body for signature verification
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read body"})
		return
	}

	// Verify webhook signature (Authorization header)
	signature := c.GetHeader("Authorization")
	if signature != "" && ctl.TikTokAppSecret != "" {
		ttClient := tiktokplatform.NewClient(ctl.TikTokAppKey, ctl.TikTokAppSecret)
		if !ttClient.VerifyWebhookSignature(rawBody, signature) {
			log.Printf("[tiktok-webhook] signature verification failed")
			c.JSON(http.StatusForbidden, gin.H{"error": "invalid signature"})
			return
		}
	}

	// Parse webhook payload
	payload, err := tiktokplatform.ParseWebhookPayload(rawBody)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	// Extract normalized messages
	messages := tiktokplatform.ExtractMessages(payload)

	processed := 0
	for _, m := range messages {
		if m.PagePlatformID == "" || m.SenderPlatformID == "" {
			continue
		}

		page, err := ctl.Repo.GetPageByPageID("tiktok", m.PagePlatformID)
		if err != nil || page == nil {
			continue
		}

		customer, err := ctl.Repo.GetOrCreateCustomer("tiktok", m.SenderPlatformID)
		if err != nil || customer == nil {
			continue
		}

		conv, err := ctl.Repo.GetOrCreateConversation(page.ID, customer.ID)
		if err != nil || conv == nil {
			continue
		}

		// Deduplicate by social message ID
		if m.MessageID != "" {
			if _, err := ctl.Repo.GetMessageBySocialMessageID(m.MessageID); err == nil {
				continue
			}
		}

		content := strings.TrimSpace(m.Text)
		if content == "" {
			content = "[Tin nhắn không xác định]"
		}

		msg := &models.Message{
			ConversationID:  conv.ID,
			SenderType:      "customer",
			ContentType:     m.ContentType,
			Content:         content,
			SocialMessageID: m.MessageID,
			CreatedAt:       time.Now(),
		}
		if err := ctl.Repo.CreateMessage(msg); err != nil {
			continue
		}

		conv.LastMessage = content
		conv.UnreadCount = conv.UnreadCount + 1
		conv.UpdatedAt = time.Now()
		_ = ctl.Repo.SaveConversation(conv)

		ctl.Hub.Broadcast(page.ID, ws.Event{
			Type:   "NEW_MESSAGE",
			PageID: page.ID,
			Data:   msg,
		})

		// Auto-reply check (async)
		go func(pageID uint, convID uint, customerPID string, pageAccessToken string, incomingText string) {
			rule, err := ctl.Svc.MatchAutoReply(pageID, incomingText)
			if err != nil || rule == nil {
				return
			}

			// Send auto-reply via TikTok Shop API
			if pageAccessToken != "" {
				_, err := ctl.Svc.TikTokClient.SendTextMessage(pageAccessToken, customerPID, rule.ReplyContent)
				if err != nil {
					log.Printf("[auto-reply-tiktok] failed to send via TikTok API: %v", err)
				}
			}

			sysMsg, err := ctl.Svc.CreateSystemMessage(convID, rule.ReplyContent)
			if err != nil {
				log.Printf("[auto-reply-tiktok] failed to save system message: %v", err)
				return
			}

			ctl.Hub.Broadcast(pageID, ws.Event{
				Type:   "NEW_MESSAGE",
				PageID: pageID,
				Data:   sysMsg,
			})
		}(page.ID, conv.ID, m.SenderPlatformID, page.AccessToken, content)

		processed++
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "received",
		"platform":  "tiktok",
		"processed": processed,
	})
}

func New(repo *repository.Repository, svc *usecase.Service, jwt *middlewares.JWTManager, hub *ws.Hub, db *gorm.DB, fbToken, ttToken, igToken, fbAppID, fbAppSecret, fbRedirectURI, frontendURL, zaloAppID, zaloAppSecret, zaloOASecretKey, zaloRedirectURI, tiktokAppKey, tiktokAppSecret, tiktokRedirectURI, shopeePartnerID, shopeePartnerKey, shopeeRedirectURI, shopeeVerifyToken string) *Controller {
	return &Controller{
		Repo:                  repo,
		Svc:                   svc,
		JWT:                   jwt,
		Hub:                   hub,
		DB:                    db,
		FBToken:               fbToken,
		TTToken:               ttToken,
		IGToken:               igToken,
		FacebookAppID:         fbAppID,
		FacebookAppSecret:     fbAppSecret,
		FacebookRedirectURI:   fbRedirectURI,
		FrontendURL:           frontendURL,
		ZaloAppID:             zaloAppID,
		ZaloAppSecret:         zaloAppSecret,
		ZaloOASecretKey:       zaloOASecretKey,
		ZaloRedirectURI:       zaloRedirectURI,
		TikTokAppKey:          tiktokAppKey,
		TikTokAppSecret:       tiktokAppSecret,
		TikTokRedirectURI:     tiktokRedirectURI,
		ShopeePartnerID:       shopeePartnerID,
		ShopeePartnerKey:      shopeePartnerKey,
		ShopeeRedirectURI:     shopeeRedirectURI,
		ShopeeVerifyToken:     shopeeVerifyToken,
	}
}

func (ctl *Controller) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		IDToken  string `json:"id_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || (req.Email == "" && req.IDToken == "") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	// Google OAuth flow: auto-create user if needed
	if req.IDToken != "" {
		email := strings.TrimSpace(strings.ToLower(req.Email))
		if email == "" {
			email = fmt.Sprintf("google_%d@flybox.local", time.Now().UnixNano())
		}

		user, err := ctl.Repo.GetUserByEmail(email)
		if err != nil {
			user = &models.User{
				Email:        email,
				PasswordHash: "",
				Role:         "staff",
				CreatedAt:    time.Now(),
			}
			if err := ctl.DB.Create(user).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot create user"})
				return
			}
		}

		token, err := ctl.JWT.Generate(user.ID, user.Email, user.Role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token": token,
			"user": gin.H{
				"id":    user.ID,
				"email": user.Email,
				"name":  user.Email,
				"role":  user.Role,
			},
		})
		return
	}

	// Email + password flow
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email and password are required"})
		return
	}

	user, err := ctl.Repo.GetUserByEmail(email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		// Fallback: support legacy plaintext passwords and auto-migrate to bcrypt
		if user.PasswordHash != req.Password {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}
		// Password matched as plaintext — upgrade to bcrypt hash
		if hashed, hashErr := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost); hashErr == nil {
			user.PasswordHash = string(hashed)
			_ = ctl.Repo.SaveUser(user)
		}
	}

	token, err := ctl.JWT.Generate(user.ID, user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Email,
			"role":  user.Role,
		},
	})
}

func (ctl *Controller) Register(c *gin.Context) {
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}

	existing, err := ctl.Repo.GetUserByEmail(email)
	if err == nil && existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot hash password"})
		return
	}

	user := &models.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
		Role:         "staff",
		CreatedAt:    time.Now(),
	}

	if err := ctl.DB.Create(user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot create user"})
		return
	}

	token, err := ctl.JWT.Generate(user.ID, user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot generate token"})
		return
	}

	displayName := strings.TrimSpace(req.Name)
	if displayName == "" {
		displayName = user.Email
	}

	c.JSON(http.StatusCreated, gin.H{
		"token": token,
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
			"name":  displayName,
			"role":  user.Role,
		},
	})
}

func (ctl *Controller) Me(c *gin.Context) {
	claims := middlewares.GetClaims(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := claims.UserID

	user, err := ctl.Repo.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Email,
			"role":  user.Role,
		},
	})
}

func (ctl *Controller) VerifyFacebookWebhook(c *gin.Context) {
	mode := c.Query("hub.mode")
	token := c.Query("hub.verify_token")
	challenge := c.Query("hub.challenge")

	if mode == "subscribe" && token == ctl.FBToken {
		c.String(http.StatusOK, challenge)
		return
	}
	c.JSON(http.StatusForbidden, gin.H{"error": "invalid verify token"})
}

func (ctl *Controller) FacebookWebhook(c *gin.Context) {
	// Read raw body for signature verification
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read body"})
		return
	}

	// Verify webhook signature (X-Hub-Signature-256)
	signature := c.GetHeader("X-Hub-Signature-256")
	if signature != "" {
		fbClient := fb.NewClient(ctl.FacebookAppID, ctl.FacebookAppSecret)
		if !fbClient.VerifyWebhookSignature(rawBody, signature) {
			log.Printf("[facebook-webhook] signature verification failed")
			c.JSON(http.StatusForbidden, gin.H{"error": "invalid signature"})
			return
		}
	}

	// Parse webhook payload using platform types
	payload, err := fb.ParseWebhookPayload(rawBody)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	// Extract normalized messages
	messages := fb.ExtractMessages(payload)

	processed := 0
	for _, m := range messages {
		if m.PagePlatformID == "" || m.SenderPlatformID == "" {
			continue
		}

		page, err := ctl.Repo.GetPageByPageID("facebook", m.PagePlatformID)
		if err != nil || page == nil {
			continue
		}

		customer, err := ctl.Repo.GetOrCreateCustomer("facebook", m.SenderPlatformID)
		if err != nil || customer == nil {
			continue
		}

		// Enrich customer name from Facebook profile if still using platform ID as name
		if customer.Name == m.SenderPlatformID && page.AccessToken != "" {
			go ctl.enrichCustomerName(page.AccessToken, customer)
		}

		conv, err := ctl.Repo.GetOrCreateConversation(page.ID, customer.ID)
		if err != nil || conv == nil {
			continue
		}

		// Deduplicate by social message ID
		if m.MessageID != "" {
			if _, err := ctl.Repo.GetMessageBySocialMessageID(m.MessageID); err == nil {
				continue
			}
		}

		content := strings.TrimSpace(m.Text)
		if content == "" {
			content = "[Tin nhắn không xác định]"
		}

		msg := &models.Message{
			ConversationID:  conv.ID,
			SenderType:      "customer",
			ContentType:     m.ContentType,
			Content:         content,
			SocialMessageID: m.MessageID,
			CreatedAt:       time.Now(),
		}
		if err := ctl.Repo.CreateMessage(msg); err != nil {
			continue
		}

		conv.LastMessage = content
		conv.UnreadCount = conv.UnreadCount + 1
		conv.UpdatedAt = time.Now()
		_ = ctl.Repo.SaveConversation(conv)

		ctl.Hub.Broadcast(page.ID, ws.Event{
			Type:   "NEW_MESSAGE",
			PageID: page.ID,
			Data:   msg,
		})

		// Auto-reply check (async to not block webhook response)
		go func(pageID uint, convID uint, customerPID string, pageAccessToken string, incomingText string) {
			rule, err := ctl.Svc.MatchAutoReply(pageID, incomingText)
			if err != nil || rule == nil {
				return
			}

			// Send auto-reply via Facebook API
			if pageAccessToken != "" {
				_, err := ctl.Svc.FBClient.SendTextMessage(pageAccessToken, customerPID, rule.ReplyContent)
				if err != nil {
					log.Printf("[auto-reply] failed to send via FB API: %v", err)
				}
			}

			// Save as system message
			sysMsg, err := ctl.Svc.CreateSystemMessage(convID, rule.ReplyContent)
			if err != nil {
				log.Printf("[auto-reply] failed to save system message: %v", err)
				return
			}

			// Broadcast auto-reply to connected clients
			ctl.Hub.Broadcast(pageID, ws.Event{
				Type:   "NEW_MESSAGE",
				PageID: pageID,
				Data:   sysMsg,
			})
		}(page.ID, conv.ID, m.SenderPlatformID, page.AccessToken, content)

		processed++
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "received",
		"platform":  "facebook",
		"processed": processed,
	})
}

// enrichCustomerName fetches the customer's real name from Facebook and updates the record.
func (ctl *Controller) enrichCustomerName(pageAccessToken string, customer *models.Customer) {
	fbClient := fb.NewClient(ctl.FacebookAppID, ctl.FacebookAppSecret)
	profile, err := fbClient.GetUserProfile(pageAccessToken, customer.PlatformID)
	if err != nil {
		log.Printf("[enrich] failed to get profile for %s: %v", customer.PlatformID, err)
		return
	}

	if profile.Name != "" && profile.Name != customer.Name {
		customer.Name = profile.Name
		if profile.Avatar != "" {
			customer.Avatar = profile.Avatar
		}
		customer.UpdatedAt = time.Now()
		_ = ctl.Repo.SaveCustomer(customer)
	}
}

func (ctl *Controller) ZaloWebhook(c *gin.Context) {
	// Read raw body for signature verification
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read body"})
		return
	}

	// Verify webhook signature (X-ZEvent-Signature header)
	signature := c.GetHeader("X-ZEvent-Signature")
	if signature != "" && ctl.ZaloOASecretKey != "" {
		zaloClient := zaloplatform.NewClient(ctl.ZaloAppID, ctl.ZaloAppSecret, ctl.ZaloOASecretKey)
		if !zaloClient.VerifyWebhookSignature(rawBody, signature) {
			log.Printf("[zalo-webhook] signature verification failed")
			c.JSON(http.StatusForbidden, gin.H{"error": "invalid signature"})
			return
		}
	}

	// Parse webhook payload
	payload, err := zaloplatform.ParseWebhookPayload(rawBody)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	// Handle follow/unfollow events (no message to process)
	if payload.EventName == "follow" || payload.EventName == "unfollow" {
		log.Printf("[zalo-webhook] %s event from user %s", payload.EventName, payload.Sender.ID)
		c.JSON(http.StatusOK, gin.H{"status": "received", "platform": "zalo", "event": payload.EventName})
		return
	}

	// Extract normalized messages
	messages := zaloplatform.ExtractMessages(payload)

	processed := 0
	for _, m := range messages {
		if m.PagePlatformID == "" || m.SenderPlatformID == "" {
			continue
		}

		page, err := ctl.Repo.GetPageByPageID("zalo", m.PagePlatformID)
		if err != nil || page == nil {
			continue
		}

		customer, err := ctl.Repo.GetOrCreateCustomer("zalo", m.SenderPlatformID)
		if err != nil || customer == nil {
			continue
		}

		// Enrich customer name from Zalo profile if still using platform ID as name
		if customer.Name == m.SenderPlatformID && page.AccessToken != "" {
			go ctl.enrichZaloCustomerName(page.AccessToken, customer)
		}

		conv, err := ctl.Repo.GetOrCreateConversation(page.ID, customer.ID)
		if err != nil || conv == nil {
			continue
		}

		// Deduplicate by social message ID
		if m.MessageID != "" {
			if _, err := ctl.Repo.GetMessageBySocialMessageID(m.MessageID); err == nil {
				continue
			}
		}

		content := strings.TrimSpace(m.Text)
		if content == "" {
			content = "[Tin nhắn không xác định]"
		}

		msg := &models.Message{
			ConversationID:  conv.ID,
			SenderType:      "customer",
			ContentType:     m.ContentType,
			Content:         content,
			SocialMessageID: m.MessageID,
			CreatedAt:       time.Now(),
		}
		if err := ctl.Repo.CreateMessage(msg); err != nil {
			continue
		}

		conv.LastMessage = content
		conv.UnreadCount = conv.UnreadCount + 1
		conv.UpdatedAt = time.Now()
		_ = ctl.Repo.SaveConversation(conv)

		ctl.Hub.Broadcast(page.ID, ws.Event{
			Type:   "NEW_MESSAGE",
			PageID: page.ID,
			Data:   msg,
		})

		// Auto-reply check (async)
		go func(pageID uint, convID uint, customerPID string, pageAccessToken string, incomingText string) {
			rule, err := ctl.Svc.MatchAutoReply(pageID, incomingText)
			if err != nil || rule == nil {
				return
			}

			// Send auto-reply via Zalo OA API
			if pageAccessToken != "" {
				_, err := ctl.Svc.ZaloClient.SendTextMessage(pageAccessToken, customerPID, rule.ReplyContent)
				if err != nil {
					log.Printf("[auto-reply-zalo] failed to send via Zalo API: %v", err)
				}
			}

			// Save as system message
			sysMsg, err := ctl.Svc.CreateSystemMessage(convID, rule.ReplyContent)
			if err != nil {
				log.Printf("[auto-reply-zalo] failed to save system message: %v", err)
				return
			}

			ctl.Hub.Broadcast(pageID, ws.Event{
				Type:   "NEW_MESSAGE",
				PageID: pageID,
				Data:   sysMsg,
			})
		}(page.ID, conv.ID, m.SenderPlatformID, page.AccessToken, content)

		processed++
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "received",
		"platform":  "zalo",
		"processed": processed,
	})
}

// enrichZaloCustomerName fetches the customer's real name from Zalo and updates the record.
func (ctl *Controller) enrichZaloCustomerName(oaAccessToken string, customer *models.Customer) {
	profile, err := ctl.Svc.ZaloClient.GetUserProfile(oaAccessToken, customer.PlatformID)
	if err != nil {
		log.Printf("[enrich-zalo] failed to get profile for %s: %v", customer.PlatformID, err)
		return
	}

	if profile.Name != "" && profile.Name != customer.Name {
		customer.Name = profile.Name
		if profile.Avatar != "" {
			customer.Avatar = profile.Avatar
		}
		customer.UpdatedAt = time.Now()
		_ = ctl.Repo.SaveCustomer(customer)
	}
}

func (ctl *Controller) ListPages(c *gin.Context) {
	pages, err := ctl.Repo.ListPages()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot list pages"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": pages})
}

func (ctl *Controller) ConnectPage(c *gin.Context) {
	var req struct {
		Platform string `json:"platform"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Platform == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	platform := strings.ToLower(strings.TrimSpace(req.Platform))
	switch platform {
	case "facebook", "zalo", "tiktok", "instagram":
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported platform"})
		return
	}

	state := fmt.Sprintf("%s-%d", platform, time.Now().Unix())

	switch platform {
	case "facebook":
		if strings.TrimSpace(ctl.FacebookAppID) == "" || strings.TrimSpace(ctl.FacebookRedirectURI) == "" {
			// DEV MOCK MODE: Bypass real Facebook OAuth to allow rapid testing of the UI
			baseURL := strings.TrimRight(ctl.FrontendURL, "/")
			if baseURL == "" {
				baseURL = "http://localhost:5173" // Default fallback
			}
			
			// Redirect back to frontend callback with a special mock code
			authURL := fmt.Sprintf("%s/connect/callback?code=mock_fb_code&state=%s", baseURL, url.QueryEscape(state))
			
			c.JSON(http.StatusOK, gin.H{
				"data": gin.H{
					"platform": platform,
					"auth_url": authURL,
					"state":    state,
				},
			})
			return
		}

		// Request all necessary permissions for Facebook Pages
		scope := "pages_show_list,pages_manage_metadata,pages_messaging,pages_read_engagement,pages_manage_posts,pages_read_user_content"
		authURL := fmt.Sprintf("https://www.facebook.com/v19.0/dialog/oauth?client_id=%s&redirect_uri=%s&scope=%s&state=%s&response_type=code",
			url.QueryEscape(ctl.FacebookAppID),
			url.QueryEscape(ctl.FacebookRedirectURI),
			url.QueryEscape(scope),
			url.QueryEscape(state),
		)

		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"platform": platform,
				"auth_url": authURL,
				"state":    state,
			},
		})

	case "zalo":
		if strings.TrimSpace(ctl.ZaloAppID) == "" {
			c.JSON(http.StatusOK, gin.H{
				"data": gin.H{
					"platform": platform,
					"auth_url": "",
					"state":    state,
					"message":  "zalo oauth not configured - use manual connection",
					"manual":   true,
				},
			})
			return
		}

		authURL := fmt.Sprintf("https://oauth.zaloapp.com/v4/oa/permission?app_id=%s&redirect_uri=%s&state=%s",
			url.QueryEscape(ctl.ZaloAppID),
			url.QueryEscape(ctl.ZaloRedirectURI),
			url.QueryEscape(state),
		)

		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"platform": platform,
				"auth_url": authURL,
				"state":    state,
			},
		})

	case "tiktok":
		if strings.TrimSpace(ctl.TikTokAppKey) == "" {
			c.JSON(http.StatusOK, gin.H{
				"data": gin.H{
					"platform": platform,
					"auth_url": "",
					"state":    state,
					"message":  "tiktok oauth not configured - use manual connection",
					"manual":   true,
				},
			})
			return
		}

		authURL := fmt.Sprintf("https://auth.tiktok-shops.com/oauth/authorize?app_key=%s&redirect_uri=%s&state=%s",
			url.QueryEscape(ctl.TikTokAppKey),
			url.QueryEscape(ctl.TikTokRedirectURI),
			url.QueryEscape(state),
		)

		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"platform": platform,
				"auth_url": authURL,
				"state":    state,
			},
		})

	case "instagram":
		// Instagram uses the same OAuth as Facebook (Meta ecosystem)
		if strings.TrimSpace(ctl.FacebookAppID) == "" || strings.TrimSpace(ctl.FacebookRedirectURI) == "" {
			c.JSON(http.StatusOK, gin.H{
				"data": gin.H{
					"platform": platform,
					"auth_url": "",
					"state":    state,
					"message":  "instagram oauth not configured - use manual connection",
					"manual":   true,
				},
			})
			return
		}

		// Instagram requires additional scopes for messaging
		scope := "pages_show_list,pages_manage_metadata,instagram_basic,instagram_manage_messages"
		authURL := fmt.Sprintf("https://www.facebook.com/v19.0/dialog/oauth?client_id=%s&redirect_uri=%s&scope=%s&state=%s",
			url.QueryEscape(ctl.FacebookAppID),
			url.QueryEscape(ctl.FacebookRedirectURI),
			url.QueryEscape(scope),
			url.QueryEscape(state),
		)

		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"platform": platform,
				"auth_url": authURL,
				"state":    state,
			},
		})

	case "shopee":
		// Shopee uses partner-based authentication
		if strings.TrimSpace(ctl.ShopeePartnerID) == "" || strings.TrimSpace(ctl.ShopeeRedirectURI) == "" {
			c.JSON(http.StatusOK, gin.H{
				"data": gin.H{
					"platform": platform,
					"auth_url": "",
					"state":    state,
					"message":  "shopee oauth not configured - use manual connection",
					"manual":   true,
				},
			})
			return
		}

		authURL := fmt.Sprintf("https://partner.shopeemobile.com/api/v2/shop/auth_partner?partner_id=%s&redirect=%s&state=%s",
			url.QueryEscape(ctl.ShopeePartnerID),
			url.QueryEscape(ctl.ShopeeRedirectURI),
			url.QueryEscape(state),
		)

		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"platform": platform,
				"auth_url": authURL,
				"state":    state,
			},
		})

	default:
		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"platform": platform,
				"auth_url": "",
				"state":    state,
				"message":  platform + " oauth not yet implemented",
				"manual":   true,
			},
		})
	}
}

func (ctl *Controller) FacebookConnectCallback(c *gin.Context) {
	errParam := strings.TrimSpace(c.Query("error"))
	if errParam != "" {
		msg := strings.TrimSpace(c.Query("error_description"))
		if msg == "" {
			msg = errParam
		}
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, `<!DOCTYPE html>
<html lang="vi">
<head><meta charset="UTF-8"><title>Fly-Box - Kết nối Facebook</title>
<style>
  body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; display: flex; justify-content: center; align-items: center; min-height: 100vh; margin: 0; background: #f5f5f5; }
  .card { background: white; padding: 40px; border-radius: 12px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); text-align: center; max-width: 400px; }
  .error { color: #dc3545; } .success { color: #28a745; }
  h2 { margin-bottom: 20px; } p { color: #666; margin: 20px 0; }
  .btn { display: inline-block; padding: 12px 24px; background: #1877f2; color: white; text-decoration: none; border-radius: 6px; font-weight: 500; }
</style>
</head>
<body>
<div class="card">
  <h2 class="error">❌ Kết nối thất bại</h2>
  <p class="error">`+msg+`</p>
  <p>Vui lòng thử lại hoặc liên hệ hỗ trợ.</p>
  <a href="/" class="btn"> Quay về Fly-Box</a>
</div>
</body></html>`)
		return
	}

	code := strings.TrimSpace(c.Query("code"))
	state := strings.TrimSpace(c.Query("state"))
	if code == "" || state == "" {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, `<!DOCTYPE html>
<html lang="vi">
<head><meta charset="UTF-8"><title>Fly-Box - Kết nối Facebook</title>
<style>
  body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; display: flex; justify-content: center; align-items: center; min-height: 100vh; margin: 0; background: #f5f5f5; }
  .card { background: white; padding: 40px; border-radius: 12px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); text-align: center; max-width: 400px; }
  .error { color: #dc3545; }
  h2 { margin-bottom: 20px; } p { color: #666; margin: 20px 0; }
  .btn { display: inline-block; padding: 12px 24px; background: #1877f2; color: white; text-decoration: none; border-radius: 6px; font-weight: 500; }
</style>
</head>
<body>
<div class="card">
  <h2 class="error">⚠️ Kết nối không hợp lệ</h2>
  <p class="error">Thiếu thông tin từ Facebook. Vui lòng thử lại.</p>
  <a href="/" class="btn"> Quay về Fly-Box</a>
</div>
</body></html>`)
		return
	}

	parts := strings.SplitN(state, "-", 2)
	if len(parts) < 2 || parts[0] != "facebook" {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, `<!DOCTYPE html>
<html lang="vi">
<head><meta charset="UTF-8"><title>Fly-Box - Kết nối Facebook</title>
<style>
  body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; display: flex; justify-content: center; align-items: center; min-height: 100vh; margin: 0; background: #f5f5f5; }
  .card { background: white; padding: 40px; border-radius: 12px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); text-align: center; max-width: 400px; }
  .error { color: #dc3545; }
  h2 { margin-bottom: 20px; } p { color: #666; margin: 20px 0; }
  .btn { display: inline-block; padding: 12px 24px; background: #1877f2; color: white; text-decoration: none; border-radius: 6px; font-weight: 500; }
</style>
</head>
<body>
<div class="card">
  <h2 class="error">⚠️ OAuth State không hợp lệ</h2>
  <p class="error">Phiên kết nối không hợp lệ. Vui lòng thử lại.</p>
  <a href="/" class="btn"> Quay về Fly-Box</a>
</div>
</body></html>`)
		return
	}

	// Return HTML page with code for frontend to process
	// Redirect to frontend callback page with code and state
	frontendBaseURL := strings.TrimRight(ctl.FrontendURL, "/")
	if frontendBaseURL == "" {
		frontendBaseURL = "http://localhost:5173"
	}
	
	redirectURL := fmt.Sprintf("%s/connect/callback?code=%s&state=%s", 
		frontendBaseURL, 
		url.QueryEscape(code), 
		url.QueryEscape(state))
	
	c.Redirect(http.StatusFound, redirectURL)
}

func (ctl *Controller) ZaloConnectCallback(c *gin.Context) {
	baseURL := strings.TrimRight(ctl.FrontendURL, "/")

	errParam := strings.TrimSpace(c.Query("error"))
	if errParam != "" {
		msg := strings.TrimSpace(c.Query("error_description"))
		if msg == "" {
			msg = errParam
		}
		c.Redirect(http.StatusFound, baseURL+"/connect/callback?status=error&message="+url.QueryEscape(msg))
		return
	}

	code := strings.TrimSpace(c.Query("code"))
	if code == "" {
		c.Redirect(http.StatusFound, baseURL+"/connect/callback?status=error&message="+url.QueryEscape("missing code from zalo"))
		return
	}

	state := "zalo-" + fmt.Sprintf("%d", time.Now().Unix())

	redirectURL := baseURL + "/connect/callback?status=success&code=" + url.QueryEscape(code) + "&state=" + url.QueryEscape(state)
	c.Redirect(http.StatusFound, redirectURL)
}

func (ctl *Controller) TikTokConnectCallback(c *gin.Context) {
	baseURL := strings.TrimRight(ctl.FrontendURL, "/")

	errParam := strings.TrimSpace(c.Query("error"))
	if errParam != "" {
		msg := strings.TrimSpace(c.Query("error_description"))
		if msg == "" {
			msg = errParam
		}
		c.Redirect(http.StatusFound, baseURL+"/connect/callback?status=error&message="+url.QueryEscape(msg))
		return
	}

	code := strings.TrimSpace(c.Query("code"))
	if code == "" {
		c.Redirect(http.StatusFound, baseURL+"/connect/callback?status=error&message="+url.QueryEscape("missing code from tiktok"))
		return
	}

	state := "tiktok-" + fmt.Sprintf("%d", time.Now().Unix())

	redirectURL := baseURL + "/connect/callback?status=success&code=" + url.QueryEscape(code) + "&state=" + url.QueryEscape(state)
	c.Redirect(http.StatusFound, redirectURL)
}

func (ctl *Controller) InstagramConnectCallback(c *gin.Context) {
	baseURL := strings.TrimRight(ctl.FrontendURL, "/")

	errParam := strings.TrimSpace(c.Query("error"))
	if errParam != "" {
		msg := strings.TrimSpace(c.Query("error_description"))
		if msg == "" {
			msg = errParam
		}
		c.Redirect(http.StatusFound, baseURL+"/connect/callback?status=error&message="+url.QueryEscape(msg))
		return
	}

	code := strings.TrimSpace(c.Query("code"))
	if code == "" {
		c.Redirect(http.StatusFound, baseURL+"/connect/callback?status=error&message="+url.QueryEscape("missing code from instagram"))
		return
	}

	state := "instagram-" + fmt.Sprintf("%d", time.Now().Unix())

	redirectURL := baseURL + "/connect/callback?status=success&code=" + url.QueryEscape(code) + "&state=" + url.QueryEscape(state)
	c.Redirect(http.StatusFound, redirectURL)
}

func (ctl *Controller) VerifyShopeeWebhook(c *gin.Context) {
	token := c.Query("hub.verify_token")
	challenge := c.Query("hub.challenge")

	if token == ctl.ShopeeVerifyToken {
		c.String(http.StatusOK, challenge)
		return
	}
	c.JSON(http.StatusForbidden, gin.H{"error": "invalid verify token"})
}

func (ctl *Controller) ShopeeWebhook(c *gin.Context) {
	// Read raw body for signature verification
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read body"})
		return
	}

	// Verify webhook signature
	signature := c.GetHeader("X-Shopee-Signature")
	if signature != "" {
		shopeeClient := shopeeplatform.NewClient(ctl.ShopeePartnerID, ctl.ShopeePartnerKey, "")
		if !shopeeClient.VerifyWebhookSignature(rawBody, signature) {
			log.Printf("[shopee-webhook] signature verification failed")
			c.JSON(http.StatusForbidden, gin.H{"error": "invalid signature"})
			return
		}
	}

	payload, err := shopeeplatform.ParseWebhookPayload(rawBody)
	if err != nil {
		log.Printf("[shopee-webhook] parse error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	messages := shopeeplatform.ExtractMessages(payload)

	processed := 0
	for _, m := range messages {
		if m.PagePlatformID == "" || m.SenderPlatformID == "" {
			continue
		}

		page, err := ctl.Repo.GetPageByPageID("shopee", m.PagePlatformID)
		if err != nil || page == nil {
			continue
		}

		customer, err := ctl.Repo.GetOrCreateCustomer("shopee", m.SenderPlatformID)
		if err != nil || customer == nil {
			continue
		}

		conv, err := ctl.Repo.GetOrCreateConversation(page.ID, customer.ID)
		if err != nil || conv == nil {
			continue
		}

		// Deduplicate by social message ID
		if m.MessageID != "" {
			if _, err := ctl.Repo.GetMessageBySocialMessageID(m.MessageID); err == nil {
				continue
			}
		}

		content := strings.TrimSpace(m.Text)
		if content == "" {
			content = "[Tin nhắn không xác định]"
		}

		msg := &models.Message{
			ConversationID:  conv.ID,
			SenderType:      "customer",
			ContentType:     m.ContentType,
			Content:         content,
			SocialMessageID: m.MessageID,
			CreatedAt:       time.Now(),
		}
		if err := ctl.Repo.CreateMessage(msg); err != nil {
			continue
		}

		conv.LastMessage = content
		conv.UnreadCount = conv.UnreadCount + 1
		conv.UpdatedAt = time.Now()
		_ = ctl.Repo.SaveConversation(conv)

		ctl.Hub.Broadcast(page.ID, ws.Event{
			Type:   "NEW_MESSAGE",
			PageID: page.ID,
			Data:   msg,
		})

		// Auto-reply check (async)
		go func(pageID uint, convID uint, customerPID string, pageAccessToken string, incomingText string) {
			rule, err := ctl.Svc.MatchAutoReply(pageID, incomingText)
			if err != nil || rule == nil {
				return
			}

			// Send auto-reply via Shopee API
			if pageAccessToken != "" {
				accessTokenWithShopID := pageAccessToken + ":" + page.ExternalPageID
				_, err := ctl.Svc.ShopeeClient.SendTextMessage(accessTokenWithShopID, customerPID, rule.ReplyContent)
				if err != nil {
					log.Printf("[auto-reply-shopee] failed to send via Shopee API: %v", err)
				}
			}

			// Save as system message
			sysMsg, err := ctl.Svc.CreateSystemMessage(convID, rule.ReplyContent)
			if err != nil {
				log.Printf("[auto-reply-shopee] failed to save system message: %v", err)
				return
			}

			ctl.Hub.Broadcast(pageID, ws.Event{
				Type:   "NEW_MESSAGE",
				PageID: pageID,
				Data:   sysMsg,
			})
		}(page.ID, conv.ID, m.SenderPlatformID, page.AccessToken, content)

		processed++
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "received",
		"platform":  "shopee",
		"processed": processed,
	})
}

func (ctl *Controller) ShopeeConnectCallback(c *gin.Context) {
	baseURL := strings.TrimRight(ctl.FrontendURL, "/")

	errParam := strings.TrimSpace(c.Query("error"))
	if errParam != "" {
		msg := strings.TrimSpace(c.Query("error_description"))
		if msg == "" {
			msg = errParam
		}
		c.Redirect(http.StatusFound, baseURL+"/connect/callback?status=error&message="+url.QueryEscape(msg))
		return
	}

	code := strings.TrimSpace(c.Query("code"))
	shopID := strings.TrimSpace(c.Query("shop_id"))

	state := "shopee-" + fmt.Sprintf("%d", time.Now().Unix())

	redirectURL := baseURL + "/connect/callback?status=success&code=" + url.QueryEscape(code) + "&shop_id=" + url.QueryEscape(shopID) + "&state=" + url.QueryEscape(state)
	c.Redirect(http.StatusFound, redirectURL)
}

func (ctl *Controller) CompletePageConnection(c *gin.Context) {
	var req struct {
		Platform        string `json:"platform"`
		Code            string `json:"code"`
		State           string `json:"state"`
		PageID          string `json:"page_id"`
		PageName        string `json:"page_name"`
		PermissionLevel string `json:"permission_level"`
		ConnectedShop   string `json:"connected_shop_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Platform == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	platform := strings.ToLower(strings.TrimSpace(req.Platform))

	// For Facebook, exchange code for real tokens
	if platform == "facebook" && req.Code != "" {
		var savedPages []models.SocialPage

// DEV MOCK MODE
		if req.Code == "mock_fb_code" {
			mockPage := &models.SocialPage{
				Platform:        "facebook",
				ExternalPageID: "mock_page_123",
				PageName:       "Fly-Box Demo Fanpage",
				AccessToken:    "mock_token",
				Status:         "active",
			}
			err := ctl.Repo.DB.Create(mockPage).Error
			if err == nil {
				savedPages = append(savedPages, *mockPage)
			}
			
			// Auto-generate a dummy conversation so user sees something in Inbox
			mockCustomer := &models.Customer{
				Platform:   "facebook",
				PlatformID: "mock_cust_456",
				Name:       "Nguyễn Văn A (Khách test Facebook)",
				Avatar:     "https://ui-avatars.com/api/?name=NVA&background=random",
			}
			ctl.Repo.DB.Create(mockCustomer)
			
			conv := &models.Conversation{
				PageID:      mockPage.ID,
				CustomerID:  mockCustomer.ID,
				LastMessage: "Xin chào, sản phẩm này còn không shop?",
				UnreadCount: 1,
			}
			ctl.Repo.DB.Create(conv)
			
			msg := &models.Message{
				ConversationID: conv.ID,
				SenderType:     "customer",
				Content:        "Xin chào, sản phẩm này còn không shop?",
			}
			ctl.Repo.DB.Create(msg)

			c.JSON(http.StatusOK, gin.H{
				"data":    savedPages,
				"message": "connected via sandbox mode",
			})
			return
		}

		result, err := ctl.Svc.ExchangeFacebookCode(req.Code)
		if err != nil {
			log.Printf("[facebook-connect] ExchangeFacebookCode failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("facebook oauth failed: %v", err)})
			return
		}

		log.Printf("[facebook-connect] Found %d pages", len(result.Pages))

		for _, page := range result.Pages {
			log.Printf("[facebook-connect] Processing page: ID=%s, Name=%s", page.ID, page.Name)
			savedPage, err := ctl.Svc.SaveFacebookPage(page, result.UserAccessToken)
			if err != nil {
				log.Printf("[facebook-connect] Failed to save page %s: %v", page.ID, err)
				continue
			}

			// Subscribe to webhook for this page
			if err := ctl.Svc.SubscribePageToWebhook(savedPage); err != nil {
				log.Printf("[facebook-connect] Failed to subscribe webhook for page %s: %v", page.ID, err)
			}

			savedPages = append(savedPages, *savedPage)
		}

		if len(savedPages) == 0 {
			log.Printf("[facebook-connect] No pages saved. Total pages found: %d", len(result.Pages))
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("no pages found or failed to save pages. Found %d pages from Facebook.", len(result.Pages))})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":    savedPages,
			"message": "connected",
			"pages":   result.Pages,
		})
		return
	}

	// For Zalo, exchange code for OA tokens
	if platform == "zalo" && req.Code != "" {
		result, err := ctl.Svc.ExchangeZaloCode(req.Code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("zalo oauth failed: %v", err)})
			return
		}

		savedPage, err := ctl.Svc.SaveZaloPage(result.OAId, result.OAName, result.AccessToken, result.RefreshToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("save zalo page failed: %v", err)})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":    []models.SocialPage{*savedPage},
			"message": "connected",
		})
		return
	}

	// For TikTok, exchange code for shop tokens
	if platform == "tiktok" && req.Code != "" {
		result, err := ctl.Svc.ExchangeTikTokCode(req.Code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("tiktok oauth failed: %v", err)})
			return
		}

		savedPage, err := ctl.Svc.SaveTikTokShop(result.ShopID, result.ShopName, result.AccessToken, result.RefreshToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("save tiktok shop failed: %v", err)})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":    []models.SocialPage{*savedPage},
			"message": "connected",
		})
		return
	}

	// For Instagram, exchange code for IG accounts
	if platform == "instagram" && req.Code != "" {
		result, err := ctl.Svc.ExchangeInstagramCode(req.Code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("instagram oauth failed: %v", err)})
			return
		}

		if len(result.IGAccounts) == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "no instagram business accounts found. Make sure your Facebook Page is linked to an Instagram Business account."})
			return
		}

		savedPages := make([]models.SocialPage, 0)
		for _, igAccount := range result.IGAccounts {
			savedPage, err := ctl.Svc.SaveInstagramPage(igAccount, result.UserAccessToken)
			if err != nil {
				continue
			}

			// Subscribe to webhook for this IG account
			_ = ctl.Svc.SubscribeInstagramToWebhook(savedPage)

			savedPages = append(savedPages, *savedPage)
		}

		if len(savedPages) == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save instagram accounts"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":    savedPages,
			"message": "connected",
		})
		return
	}

	// For Shopee, exchange code for shop tokens
	if platform == "shopee" && req.Code != "" {
		shopID := req.PageID
		if shopID == "" {
			shopID = c.Query("shop_id")
		}
		result, err := ctl.Svc.ExchangeShopeeCode(req.Code, shopID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("shopee oauth failed: %v", err)})
			return
		}

		savedPage, err := ctl.Svc.SaveShopeePage(result.ShopID, result.ShopName, result.AccessToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("save shopee shop failed: %v", err)})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":    []models.SocialPage{*savedPage},
			"message": "connected",
		})
		return
	}

	// Fallback: manual page connection (for testing or other platforms)
	if req.PageName == "" {
		req.PageName = "Facebook Fanpage"
	}
	if req.PermissionLevel == "" {
		req.PermissionLevel = "admin"
	}

	existing, err := ctl.Repo.GetPageByPageID(platform, req.PageID)
	if err == nil && existing != nil {
		existing.PageName = req.PageName
		existing.ConnectionStatus = "connected"
		existing.PermissionLevel = req.PermissionLevel
		existing.RequiresReauth = false
		existing.ConnectedShopName = req.ConnectedShop
		existing.SupportsAdvancedTools = req.PermissionLevel == "admin"
		if req.PermissionLevel != "admin" {
			existing.WarningMessage = "Chi co quyen Quan tri vien trang Facebook moi su dung duoc: loi chao, danh sach uy quyen website, tin nhan mo dau, cau hoi thuong gap, menu chinh."
		} else {
			existing.WarningMessage = ""
		}
		existing.AccessToken = "facebook-access-token-from-" + req.Code
		if err := ctl.Repo.SavePage(existing); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot update page"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": existing, "message": "connected"})
		return
	}

	page := models.SocialPage{
		Platform:              platform,
		ExternalPageID:        req.PageID,
		PageName:              req.PageName,
		AccessToken:           "facebook-access-token-from-" + req.Code,
		Status:                "active",
		ConnectionStatus:      "connected",
		PermissionLevel:       req.PermissionLevel,
		RequiresReauth:        false,
		ConnectedShopName:     req.ConnectedShop,
		SupportsAdvancedTools: req.PermissionLevel == "admin",
	}
	if req.PermissionLevel != "admin" {
		page.WarningMessage = "Chi co quyen Quan tri vien trang Facebook moi su dung duoc: loi chao, danh sach uy quyen website, tin nhan mo dau, cau hoi thuong gap, menu chinh."
	}

	if err := ctl.Repo.CreatePage(&page); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot create page"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": page, "message": "connected"})
}

func (ctl *Controller) ListConversations(c *gin.Context) {
	pageIDParam := c.Query("page_id")
	var pageID uint
	if pageIDParam != "" {
		id, _ := strconv.ParseUint(pageIDParam, 10, 32)
		pageID = uint(id)
	}

	convs, err := ctl.Repo.ListConversations(pageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot list conversations"})
		return
	}

	// Enrich conversations with customer and page info
	type convResponse struct {
		ID               uint   `json:"id"`
		CustomerName     string `json:"customer_name"`
		CustomerPlatform string `json:"customer_platform"`
		PageName         string `json:"page_name"`
		Platform         string `json:"platform"`
		LastMessage      string `json:"last_message"`
		UnreadCount      int    `json:"unread_count"`
		UpdatedAt        string `json:"updated_at"`
	}

	data := make([]convResponse, 0, len(convs))
	for _, conv := range convs {
		// Get customer info
		customer, _ := ctl.Repo.GetCustomerByID(conv.CustomerID)
		customerName := "Khách vô danh"
		customerPlatform := ""
		if customer != nil {
			customerName = customer.Name
			customerPlatform = customer.Platform
		}

		// Get page info
		pageName := ""
		platform := ""
		if conv.PageID != 0 {
			page, _ := ctl.Repo.GetPageByID(conv.PageID)
			if page != nil {
				pageName = page.PageName
				platform = page.Platform
			}
		}

		data = append(data, convResponse{
			ID:               conv.ID,
			CustomerName:     customerName,
			CustomerPlatform: customerPlatform,
			PageName:         pageName,
			Platform:         platform,
			LastMessage:      conv.LastMessage,
			UnreadCount:      conv.UnreadCount,
			UpdatedAt:        conv.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": data})
}

func (ctl *Controller) ListMessages(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation id"})
		return
	}
	data, err := ctl.Repo.ListMessages(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot list messages"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func (ctl *Controller) SendMessage(c *gin.Context) {
	idParam := c.Param("id")
	convID64, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation id"})
		return
	}
	convID := uint(convID64)

	var req struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	// Get conversation to determine platform
	conv, err := ctl.Repo.GetConversationByID(convID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}

	// Get customer to get platform ID
	customer, err := ctl.Repo.GetCustomerByID(conv.CustomerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "customer not found"})
		return
	}

	var msg *models.Message

	// Try to send via platform API based on the page's platform
	if conv.PageID != 0 {
		page, _ := ctl.Repo.GetPageByID(conv.PageID)
		if page != nil && page.AccessToken != "" {
			switch page.Platform {
			case "facebook":
				msg, err = ctl.Svc.SendFacebookMessage(convID, customer.PlatformID, req.Content)
				if err != nil {
					log.Printf("[send-message] FB API failed, saving locally: %v", err)
					msg, _ = ctl.Svc.CreateAgentMessage(convID, req.Content)
				}
			case "zalo":
				msg, err = ctl.Svc.SendZaloMessage(convID, customer.PlatformID, req.Content)
				if err != nil {
					log.Printf("[send-message] Zalo API failed, saving locally: %v", err)
					msg, _ = ctl.Svc.CreateAgentMessage(convID, req.Content)
				}
			case "tiktok":
				msg, err = ctl.Svc.SendTikTokMessage(convID, customer.PlatformID, req.Content)
				if err != nil {
					log.Printf("[send-message] TikTok API failed, saving locally: %v", err)
					msg, _ = ctl.Svc.CreateAgentMessage(convID, req.Content)
				}
			case "instagram":
				msg, err = ctl.Svc.SendInstagramMessage(convID, customer.PlatformID, req.Content)
				if err != nil {
					log.Printf("[send-message] Instagram API failed, saving locally: %v", err)
					msg, _ = ctl.Svc.CreateAgentMessage(convID, req.Content)
				}
			case "shopee":
				msg, err = ctl.Svc.SendShopeeMessage(convID, customer.PlatformID, req.Content)
				if err != nil {
					log.Printf("[send-message] Shopee API failed, saving locally: %v", err)
					msg, _ = ctl.Svc.CreateAgentMessage(convID, req.Content)
				}
			default:
				msg, err = ctl.Svc.CreateAgentMessage(convID, req.Content)
			}
		} else {
			msg, err = ctl.Svc.CreateAgentMessage(convID, req.Content)
		}
	} else {
		msg, err = ctl.Svc.CreateAgentMessage(convID, req.Content)
	}

	if err != nil || msg == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot create message"})
		return
	}

	// Broadcast to all connected clients
	ctl.Hub.Broadcast(conv.PageID, ws.Event{
		Type:   "NEW_MESSAGE",
		PageID: conv.PageID,
		Data:   msg,
	})

	c.JSON(http.StatusCreated, gin.H{"data": msg})
}

func (ctl *Controller) ListAutoReplies(c *gin.Context) {
	pageIDParam := c.Query("page_id")
	var pageID uint
	if pageIDParam != "" {
		id, _ := strconv.ParseUint(pageIDParam, 10, 32)
		pageID = uint(id)
	}
	data, err := ctl.Repo.ListAutoReplyRules(pageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot list auto replies"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func (ctl *Controller) CreateAutoReply(c *gin.Context) {
	var req struct {
		PageID       uint     `json:"page_id"`
		RuleType     string   `json:"rule_type"`
		Keywords     []string `json:"keywords"`
		ReplyContent string   `json:"reply_content"`
		IsActive     *bool    `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.PageID == 0 || req.RuleType == "" || req.ReplyContent == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	kb, _ := json.Marshal(req.Keywords)
	rule := models.AutoReplyRule{
		PageID:       req.PageID,
		RuleType:     req.RuleType,
		Keywords:     datatypes.JSON(kb),
		ReplyContent: req.ReplyContent,
		IsActive:     true,
	}
	if req.IsActive != nil {
		rule.IsActive = *req.IsActive
	}

	if err := ctl.Repo.CreateAutoReplyRule(&rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot create rule"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": rule})
}

func (ctl *Controller) UpdateAutoReply(c *gin.Context) {
	idParam := c.Param("id")
	ruleID64, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	rule, err := ctl.Repo.GetAutoReplyRuleByID(uint(ruleID64))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "rule not found"})
		return
	}

	var req struct {
		RuleType     *string  `json:"rule_type"`
		Keywords     []string `json:"keywords"`
		ReplyContent *string  `json:"reply_content"`
		IsActive     *bool    `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	if req.RuleType != nil {
		rule.RuleType = *req.RuleType
	}
	if req.ReplyContent != nil {
		rule.ReplyContent = *req.ReplyContent
	}
	if req.IsActive != nil {
		rule.IsActive = *req.IsActive
	}
	if req.Keywords != nil {
		kb, _ := json.Marshal(req.Keywords)
		rule.Keywords = datatypes.JSON(kb)
	}

	if err := ctl.Repo.UpdateAutoReplyRule(rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot update rule"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rule})
}

// --- Notification handlers ---

func (ctl *Controller) ListNotifications(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	platform := c.Query("platform")
	notifType := c.Query("type")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	notifications, total, err := ctl.Repo.ListNotifications(userID, page, pageSize, platform, notifType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot list notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": notifications,
		"meta": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

func (ctl *Controller) GetUnreadNotificationCount(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	count, err := ctl.Repo.GetUnreadNotificationCount(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot get count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"count": count}})
}

func (ctl *Controller) MarkNotificationRead(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idParam := c.Param("id")
	notifID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := ctl.Repo.MarkNotificationRead(uint(notifID), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot mark read"})
		return
	}

	// Broadcast badge update
	count, _ := ctl.Repo.GetUnreadNotificationCount(userID)
	ctl.Hub.BroadcastBadgeUpdate(userID, count)

	c.JSON(http.StatusOK, gin.H{"message": "marked as read"})
}

func (ctl *Controller) MarkAllNotificationsRead(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if err := ctl.Repo.MarkAllNotificationsRead(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot mark all read"})
		return
	}

	// Broadcast badge update
	ctl.Hub.BroadcastBadgeUpdate(userID, 0)

	c.JSON(http.StatusOK, gin.H{"message": "all marked as read"})
}

func (ctl *Controller) MarkNotificationsRead(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		NotificationIDs []uint `json:"notification_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.NotificationIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	if err := ctl.Repo.MarkNotificationsRead(req.NotificationIDs, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot mark read"})
		return
	}

	// Broadcast badge update
	count, _ := ctl.Repo.GetUnreadNotificationCount(userID)
	ctl.Hub.BroadcastBadgeUpdate(userID, count)

	c.JSON(http.StatusOK, gin.H{"message": "marked as read"})
}

func (ctl *Controller) MarkConversationRead(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idParam := c.Param("id")
	convID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	conv, err := ctl.Repo.GetConversationByID(uint(convID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}

	// Reset unread count
	conv.UnreadCount = 0
	_ = ctl.Repo.SaveConversation(conv)

	c.JSON(http.StatusOK, gin.H{"message": "marked as read"})
}
