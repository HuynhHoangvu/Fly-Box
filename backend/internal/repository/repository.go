package repository

import (
	"errors"
	"fly-box/backend/internal/models"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Repository struct {
	DB *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{DB: db}
}

// Users
func (r *Repository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	if err := r.DB.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) SaveUser(user *models.User) error {
	return r.DB.Save(user).Error
}

// Pages
func (r *Repository) ListPages() ([]models.SocialPage, error) {
	var pages []models.SocialPage
	if err := r.DB.Order("id desc").Find(&pages).Error; err != nil {
		return nil, err
	}
	return pages, nil
}

func (r *Repository) CreatePage(page *models.SocialPage) error {
	return r.DB.Create(page).Error
}

func (r *Repository) GetPageByPlatform(platform string) (*models.SocialPage, error) {
	var page models.SocialPage
	if err := r.DB.Where("platform = ?", platform).First(&page).Error; err != nil {
		return nil, err
	}
	return &page, nil
}

func (r *Repository) GetPageByPageID(platform string, pageID string) (*models.SocialPage, error) {
	var page models.SocialPage
	if err := r.DB.Where("platform = ? AND page_id = ?", strings.ToLower(strings.TrimSpace(platform)), pageID).First(&page).Error; err != nil {
		return nil, err
	}
	return &page, nil
}

func (r *Repository) GetPageByID(id uint) (*models.SocialPage, error) {
	var page models.SocialPage
	if err := r.DB.First(&page, id).Error; err != nil {
		return nil, err
	}
	return &page, nil
}

func (r *Repository) SavePage(page *models.SocialPage) error {
	return r.DB.Save(page).Error
}

// Conversations
func (r *Repository) ListConversations(pageID uint) ([]models.Conversation, error) {
	var convs []models.Conversation
	q := r.DB.Order("updated_at desc")
	if pageID != 0 {
		q = q.Where("page_id = ?", pageID)
	}
	if err := q.Find(&convs).Error; err != nil {
		return nil, err
	}
	return convs, nil
}

func (r *Repository) GetConversationByID(id uint) (*models.Conversation, error) {
	var conv models.Conversation
	if err := r.DB.First(&conv, id).Error; err != nil {
		return nil, err
	}
	return &conv, nil
}

func (r *Repository) SaveConversation(conv *models.Conversation) error {
	return r.DB.Save(conv).Error
}

func (r *Repository) GetOrCreateCustomer(platform, platformID string) (*models.Customer, error) {
	platform = strings.ToLower(strings.TrimSpace(platform))
	platformID = strings.TrimSpace(platformID)
	if platform == "" || platformID == "" {
		return nil, errors.New("invalid customer identity")
	}

	var customer models.Customer
	err := r.DB.Where("platform = ? AND platform_id = ?", platform, platformID).First(&customer).Error
	if err == nil {
		return &customer, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	customer = models.Customer{
		Platform:   platform,
		PlatformID: platformID,
		Name:       platformID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := r.DB.Create(&customer).Error; err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *Repository) GetOrCreateConversation(pageID, customerID uint) (*models.Conversation, error) {
	if pageID == 0 || customerID == 0 {
		return nil, errors.New("invalid conversation identity")
	}

	var conv models.Conversation
	err := r.DB.Where("page_id = ? AND customer_id = ?", pageID, customerID).First(&conv).Error
	if err == nil {
		return &conv, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	now := time.Now()
	conv = models.Conversation{
		PageID:      pageID,
		CustomerID:  customerID,
		Type:        "message",
		LastMessage: "",
		UnreadCount: 0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := r.DB.Create(&conv).Error; err != nil {
		return nil, err
	}
	return &conv, nil
}

// Messages
func (r *Repository) ListMessages(conversationID uint) ([]models.Message, error) {
	var msgs []models.Message
	if err := r.DB.Where("conversation_id = ?", conversationID).Order("id asc").Find(&msgs).Error; err != nil {
		return nil, err
	}
	return msgs, nil
}

func (r *Repository) CreateMessage(msg *models.Message) error {
	return r.DB.Create(msg).Error
}

func (r *Repository) GetMessageBySocialMessageID(socialMessageID string) (*models.Message, error) {
	socialMessageID = strings.TrimSpace(socialMessageID)
	if socialMessageID == "" {
		return nil, gorm.ErrRecordNotFound
	}

	var msg models.Message
	if err := r.DB.Where("social_message_id = ?", socialMessageID).First(&msg).Error; err != nil {
		return nil, err
	}
	return &msg, nil
}

// Auto-reply
func (r *Repository) ListAutoReplyRules(pageID uint) ([]models.AutoReplyRule, error) {
	var rules []models.AutoReplyRule
	q := r.DB.Where("is_active = ?", true).Order("id desc")
	if pageID != 0 {
		q = q.Where("page_id = ?", pageID)
	}
	if err := q.Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

func (r *Repository) CreateAutoReplyRule(rule *models.AutoReplyRule) error {
	return r.DB.Create(rule).Error
}

func (r *Repository) UpdateAutoReplyRule(rule *models.AutoReplyRule) error {
	return r.DB.Save(rule).Error
}

func (r *Repository) GetAutoReplyRuleByID(id uint) (*models.AutoReplyRule, error) {
	var rule models.AutoReplyRule
	if err := r.DB.First(&rule, id).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

// Customers
func (r *Repository) GetCustomerByID(id uint) (*models.Customer, error) {
	var customer models.Customer
	if err := r.DB.First(&customer, id).Error; err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *Repository) GetCustomerByPlatformID(platform, platformID string) (*models.Customer, error) {
	var customer models.Customer
	if err := r.DB.Where("platform = ? AND platform_id = ?", platform, platformID).First(&customer).Error; err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *Repository) SaveCustomer(customer *models.Customer) error {
	return r.DB.Save(customer).Error
}

// Notifications

// CreateNotification creates a new notification.
func (r *Repository) CreateNotification(notification *models.Notification) error {
	return r.DB.Create(notification).Error
}

// ListNotifications returns notifications for a user with pagination.
func (r *Repository) ListNotifications(userID uint, page, pageSize int, platform, notifType string) ([]models.Notification, int64, error) {
	var notifications []models.Notification
	var total int64

	query := r.DB.Model(&models.Notification{}).Where("user_id = ?", userID)

	if platform != "" {
		query = query.Where("platform = ?", platform)
	}
	if notifType != "" {
		query = query.Where("type = ?", notifType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&notifications).Error; err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

// GetUnreadNotificationCount returns the count of unread notifications for a user.
func (r *Repository) GetUnreadNotificationCount(userID uint) (int64, error) {
	var count int64
	if err := r.DB.Model(&models.Notification{}).Where("user_id = ? AND is_read = ?", userID, false).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// MarkNotificationRead marks a single notification as read.
func (r *Repository) MarkNotificationRead(notificationID, userID uint) error {
	return r.DB.Model(&models.Notification{}).Where("id = ? AND user_id = ?", notificationID, userID).Update("is_read", true).Error
}

// MarkNotificationsRead marks multiple notifications as read.
func (r *Repository) MarkNotificationsRead(notificationIDs []uint, userID uint) error {
	if len(notificationIDs) == 0 {
		return nil
	}
	return r.DB.Model(&models.Notification{}).Where("id IN ? AND user_id = ?", notificationIDs, userID).Update("is_read", true).Error
}

// MarkAllNotificationsRead marks all notifications as read for a user.
func (r *Repository) MarkAllNotificationsRead(userID uint) error {
	return r.DB.Model(&models.Notification{}).Where("user_id = ? AND is_read = ?", userID, false).Update("is_read", true).Error
}

// GetNotificationByID returns a notification by ID.
func (r *Repository) GetNotificationByID(id uint) (*models.Notification, error) {
	var notification models.Notification
	if err := r.DB.First(&notification, id).Error; err != nil {
		return nil, err
	}
	return &notification, nil
}

// DeleteNotification deletes a notification.
func (r *Repository) DeleteNotification(id, userID uint) error {
	return r.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Notification{}).Error
}

// PageUser methods

// GetPageUsers returns all users associated with a page.
func (r *Repository) GetPageUsers(pageID uint) ([]models.PageUser, error) {
	var pageUsers []models.PageUser
	if err := r.DB.Where("page_id = ?", pageID).Find(&pageUsers).Error; err != nil {
		return nil, err
	}
	return pageUsers, nil
}

// GetUserIDsForPage returns all user IDs associated with a page.
func (r *Repository) GetUserIDsForPage(pageID uint) ([]uint, error) {
	var userIDs []uint
	if err := r.DB.Model(&models.PageUser{}).Where("page_id = ?", pageID).Pluck("user_id", &userIDs).Error; err != nil {
		return nil, err
	}
	return userIDs, nil
}

// AddPageUser adds a user to a page.
func (r *Repository) AddPageUser(pageID, userID uint, role string) error {
	pageUser := models.PageUser{
		PageID: pageID,
		UserID: userID,
		Role:    role,
	}
	return r.DB.Create(&pageUser).Error
}

// RemovePageUser removes a user from a page.
func (r *Repository) RemovePageUser(pageID, userID uint) error {
	return r.DB.Where("page_id = ? AND user_id = ?", pageID, userID).Delete(&models.PageUser{}).Error
}

// GetPagesForUser returns all pages a user has access to.
func (r *Repository) GetPagesForUser(userID uint) ([]models.SocialPage, error) {
	var pages []models.SocialPage
	if err := r.DB.Joins("JOIN page_users ON page_users.page_id = social_pages.id").
		Where("page_users.user_id = ?", userID).
		Find(&pages).Error; err != nil {
		return nil, err
	}
	return pages, nil
}
