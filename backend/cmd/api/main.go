package main

import (
	"log"
	"net/http"

	"fly-box/backend/internal/config"
	"fly-box/backend/internal/delivery/http/controllers"
	"fly-box/backend/internal/delivery/http/middlewares"
	"fly-box/backend/internal/delivery/http/routes"
	ws "fly-box/backend/internal/delivery/websocket"
	"fly-box/backend/internal/repository"
	"fly-box/backend/internal/usecase"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	db := config.InitDB(cfg)

	repo := repository.New(db)
	svc := usecase.New(repo, cfg.FacebookAppID, cfg.FacebookAppSecret, cfg.FacebookRedirectURI, cfg.ZaloAppID, cfg.ZaloAppSecret, cfg.ZaloOASecretKey, cfg.ZaloRedirectURI, cfg.TikTokAppKey, cfg.TikTokAppSecret, cfg.TikTokRedirectURI, cfg.ShopeePartnerID, cfg.ShopeePartnerKey, cfg.ShopeeRedirectURI)
	hub := ws.NewHub()
	jwtMgr := middlewares.NewJWTManager(cfg.JWTSecret)
	enforcer := middlewares.InitCasbin(db)

	ctl := controllers.New(
		repo,
		svc,
		jwtMgr,
		hub,
		db,
		cfg.FacebookVerifyToken,
		cfg.TikTokVerifyToken,
		cfg.InstagramVerifyToken,
		cfg.FacebookAppID,
		cfg.FacebookAppSecret,
		cfg.FacebookRedirectURI,
		cfg.FrontendURL,
		cfg.ZaloAppID,
		cfg.ZaloAppSecret,
		cfg.ZaloOASecretKey,
		cfg.ZaloRedirectURI,
		cfg.TikTokAppKey,
		cfg.TikTokAppSecret,
		cfg.TikTokRedirectURI,
		cfg.ShopeePartnerID,
		cfg.ShopeePartnerKey,
		cfg.ShopeeRedirectURI,
		cfg.ShopeeVerifyToken,
	)

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		// For development: always allow localhost origins
		if origin == "" {
			origin = c.Request.Host
		}
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}
		c.Writer.Header().Set("Vary", "Origin")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})
	routes.Register(r, ctl, jwtMgr, enforcer)

	addr := ":" + cfg.AppPort
	log.Printf("server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
