package config

import (
	"log"

	"fly-box/backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB(cfg Config) *gorm.DB {
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.SocialPage{},
		&models.Customer{},
		&models.Conversation{},
		&models.Message{},
		&models.AutoReplyRule{},
		&models.Notification{},
		&models.PageUser{},
	); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	return db
}
