package config

import (
	"os"
)

type Config struct {
	AppPort              string
	DatabaseURL          string
	JWTSecret            string
	FrontendURL          string
	FacebookAppID        string
	FacebookAppSecret   string
	FacebookRedirectURI string
	FacebookVerifyToken string
	TikTokVerifyToken   string
	TikTokAppKey        string
	TikTokAppSecret     string
	TikTokRedirectURI   string
	ZaloAppID           string
	ZaloAppSecret       string
	ZaloOASecretKey     string
	ZaloRedirectURI     string
	InstagramVerifyToken string
	ShopeePartnerID     string
	ShopeePartnerKey    string
	ShopeeRedirectURI   string
	ShopeeHost          string
	ShopeeVerifyToken   string
}

func Load() Config {
	// Use ngrok URL if available, otherwise fall back to localhost
	// Also check common ngrok environment variables
	ngrokURL := os.Getenv("NGROK_URL")
	if ngrokURL == "" {
		ngrokURL = os.Getenv("NGROK_STATIC_SUBDOMAIN")
	}
	if ngrokURL == "" {
		ngrokURL = os.Getenv("NGROK_HOSTNAME")
	}
	if ngrokURL == "" && os.Getenv("NGROK_AUTHTOKEN") != "" {
		// If ngrok is being used but we don't have the URL, use placeholder
		ngrokURL = "https://your-ngrok-url.ngrok-free.dev"
	}

	frontendURL := getEnv("FRONTEND_URL", "http://localhost:5173")
	if ngrokURL != "" {
		frontendURL = ngrokURL
	}

	// Facebook app configuration
	fbAppID := getEnv("FACEBOOK_APP_ID", "")
	fbAppSecret := getEnv("FACEBOOK_APP_SECRET", "")
	if fbAppID == "" {
		// Fallback to hardcoded values for testing
		fbAppID = "2323107644886195"
		fbAppSecret = "6e7282c73fb6f63d97d2e30eb2f5e7bd"
	}

	return Config{
		AppPort:             getEnv("APP_PORT", "8081"),
		DatabaseURL:         getEnv("DATABASE_URL", "host=localhost user=postgres password=123456 dbname=flybox port=5436 sslmode=disable"),
		JWTSecret:           getEnv("JWT_SECRET", "dev-secret"),
		FrontendURL:         frontendURL,
		FacebookAppID:       fbAppID,
		FacebookAppSecret:   fbAppSecret,
		FacebookRedirectURI: getEnv("FACEBOOK_REDIRECT_URI", frontendURL+"/connect/callback"),
		FacebookVerifyToken: getEnv("FACEBOOK_VERIFY_TOKEN", "verify-token"),
		TikTokVerifyToken:   getEnv("TIKTOK_VERIFY_TOKEN", "tiktok-verify-token"),
		TikTokAppKey:        getEnv("TIKTOK_APP_KEY", ""),
		TikTokAppSecret:     getEnv("TIKTOK_APP_SECRET", ""),
		TikTokRedirectURI:   getEnv("TIKTOK_REDIRECT_URI", frontendURL+"/connect/tiktok/callback"),
		ZaloAppID:           getEnv("ZALO_APP_ID", ""),
		ZaloAppSecret:       getEnv("ZALO_APP_SECRET", ""),
		ZaloOASecretKey:     getEnv("ZALO_OA_SECRET_KEY", ""),
		ZaloRedirectURI:     getEnv("ZALO_REDIRECT_URI", frontendURL+"/connect/zalo/callback"),
		InstagramVerifyToken: getEnv("INSTAGRAM_VERIFY_TOKEN", "instagram-verify-token"),
		ShopeePartnerID:     getEnv("SHOPEE_PARTNER_ID", ""),
		ShopeePartnerKey:    getEnv("SHOPEE_PARTNER_KEY", ""),
		ShopeeRedirectURI:   getEnv("SHOPEE_REDIRECT_URI", frontendURL+"/connect/shopee/callback"),
		ShopeeHost:          getEnv("SHOPEE_HOST", "https://partner.shopeemobile.com"),
		ShopeeVerifyToken:   getEnv("SHOPEE_VERIFY_TOKEN", "shopee-verify-token"),
	}
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
