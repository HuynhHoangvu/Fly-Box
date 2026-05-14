package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// NotificationType represents the type of notification.
type NotificationType string

const (
	NotificationTypeNewMessage  NotificationType = "new_message"
	NotificationTypeNewComment  NotificationType = "new_comment"
	NotificationTypeNewOrder   NotificationType = "new_order"
	NotificationTypeNewFollower NotificationType = "new_follower"
	NotificationTypeSystem     NotificationType = "system"
)

// Notification represents a user notification.
type Notification struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"index;not null" json:"user_id"`
	PageID    *uint          `gorm:"index" json:"page_id,omitempty"`
	Type      NotificationType `gorm:"size:50;not null" json:"type"`
	Platform  string          `gorm:"size:20" json:"platform"`
	Title     string         `gorm:"type:text" json:"title"`
	Body      string         `gorm:"type:text" json:"body"`
	Data      datatypes.JSON `json:"data,omitempty"`
	IsRead    bool           `gorm:"default:false;index" json:"is_read"`
	CreatedAt time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations - removed to avoid GORM FK confusion
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	// Page relation intentionally removed - use PageID as a plain field
}

// TableName returns the table name for Notification.
func (Notification) TableName() string {
	return "notifications"
}

// NotificationData holds structured data for the notification.
type NotificationData struct {
	ConversationID uint   `json:"conversation_id,omitempty"`
	CustomerID     uint   `json:"customer_id,omitempty"`
	CustomerName   string `json:"customer_name,omitempty"`
	CustomerAvatar string `json:"customer_avatar,omitempty"`
	MessagePreview string `json:"message_preview,omitempty"`
	PageName       string `json:"page_name,omitempty"`
	Link           string `json:"link,omitempty"`
}

// NotificationEvent represents an event that triggers a notification.
type NotificationEvent struct {
	Type       NotificationType       `json:"type"`
	Platform   string                 `json:"platform"`
	UserID     uint                   `json:"user_id"`
	PageID     uint                   `json:"page_id"`
	CustomerID uint                   `json:"customer_id,omitempty"`
	ConvID     uint                   `json:"conv_id,omitempty"`
	Title      string                 `json:"title"`
	Body       string                 `json:"body"`
	Data       map[string]interface{} `json:"data,omitempty"`
}

// ToNotification converts an event to a Notification model.
func (e *NotificationEvent) ToNotification() *Notification {
	return &Notification{
		UserID:   e.UserID,
		PageID:   &e.PageID,
		Type:     e.Type,
		Platform: e.Platform,
		Title:    e.Title,
		Body:     e.Body,
		IsRead:   false,
	}
}