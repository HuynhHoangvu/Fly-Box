package models

import (
	"time"

	"gorm.io/datatypes"
)

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"not null" json:"-"`
	Role         string    `gorm:"not null;default:staff" json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type SocialPage struct {
	ID                    uint      `gorm:"primaryKey" json:"id"`
	UserID                *uint     `gorm:"index" json:"user_id,omitempty"` // Owner of this page (nullable for shared pages)
	Platform              string    `gorm:"not null;index" json:"platform"`
	ExternalPageID        string    `gorm:"column:page_id;not null;uniqueIndex" json:"page_id"`
	PageName              string    `gorm:"not null" json:"page_name"`
	AccessToken           string    `gorm:"not null" json:"-"`
	RefreshToken          string    `json:"-"`
	Status                string    `gorm:"not null;default:active" json:"status"`
	ConnectionStatus      string    `gorm:"not null;default:connected" json:"connection_status"`
	PermissionLevel       string    `gorm:"not null;default:admin" json:"permission_level"`
	RequiresReauth        bool      `gorm:"not null;default:false" json:"requires_reauth"`
	ConnectedShopName     string    `gorm:"default:null" json:"connected_shop_name"`
	SupportsAdvancedTools bool      `gorm:"not null;default:true" json:"supports_advanced_tools"`
	WarningMessage        string    `gorm:"default:null" json:"warning_message"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type Customer struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	PlatformID string    `gorm:"not null;index" json:"platform_id"`
	Name       string    `gorm:"not null" json:"name"`
	Avatar     string    `json:"avatar"`
	Platform   string    `gorm:"not null;index" json:"platform"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Conversation struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	PageID      uint      `gorm:"not null;index" json:"page_id"`
	CustomerID  uint      `gorm:"not null;index" json:"customer_id"`
	Type        string    `gorm:"not null;default:message" json:"type"`
	LastMessage string    `json:"last_message"`
	UnreadCount int       `gorm:"not null;default:0" json:"unread_count"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedAt   time.Time `json:"created_at"`
}

type Message struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	ConversationID  uint      `gorm:"not null;index" json:"conversation_id"`
	SenderType      string    `gorm:"not null" json:"sender_type"`
	ContentType     string    `gorm:"not null;default:text" json:"content_type"`
	Content         string    `gorm:"type:text;not null" json:"content"`
	SocialMessageID string    `gorm:"index" json:"social_message_id"`
	CreatedAt       time.Time `json:"created_at"`
}

type AutoReplyRule struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	PageID       uint           `gorm:"not null;index" json:"page_id"`
	RuleType     string         `gorm:"not null" json:"rule_type"` // exact_match, contains, default
	Keywords     datatypes.JSON `gorm:"type:jsonb" json:"keywords"`
	ReplyContent string         `gorm:"type:text;not null" json:"reply_content"`
	IsActive     bool           `gorm:"not null;default:true" json:"is_active"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// PageUser maps pages to users for notification routing
type PageUser struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	PageID    uint      `gorm:"not null;uniqueIndex:idx_page_user" json:"page_id"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_page_user" json:"user_id"`
	Role      string    `gorm:"not null;default:admin" json:"role"` // admin, editor, viewer
	CreatedAt time.Time `json:"created_at"`
}
