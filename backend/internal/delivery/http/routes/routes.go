package routes

import (
	"fly-box/backend/internal/delivery/http/controllers"
	"fly-box/backend/internal/delivery/http/middlewares"

	"github.com/casbin/casbin/v3"
	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine, ctl *controllers.Controller, jwtMgr *middlewares.JWTManager, enforcer *casbin.Enforcer) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.GET("/webhooks/facebook", ctl.VerifyFacebookWebhook)
	r.POST("/webhooks/facebook", ctl.FacebookWebhook)
	r.POST("/webhooks/zalo", ctl.ZaloWebhook)
	r.GET("/webhooks/tiktok", ctl.VerifyTikTokWebhook)
	r.POST("/webhooks/tiktok", ctl.TikTokWebhook)
	r.GET("/webhooks/instagram", ctl.VerifyInstagramWebhook)
	r.POST("/webhooks/instagram", ctl.InstagramWebhook)
	r.GET("/webhooks/shopee", ctl.VerifyShopeeWebhook)
	r.POST("/webhooks/shopee", ctl.ShopeeWebhook)

	// Platform OAuth callbacks (public, no auth required)
	r.GET("/api/v1/pages/connect/zalo/callback", ctl.ZaloConnectCallback)
	r.GET("/api/v1/pages/connect/tiktok/callback", ctl.TikTokConnectCallback)
	r.GET("/api/v1/pages/connect/instagram/callback", ctl.InstagramConnectCallback)
	r.GET("/api/v1/pages/connect/shopee/callback", ctl.ShopeeConnectCallback)

	r.GET("/ws", func(c *gin.Context) {
		ctl.Hub.HandleWS(c.Writer, c.Request)
	})

	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/login", ctl.Login)
			auth.POST("/register", ctl.Register)
		}

		v1.GET("/pages/connect/callback", ctl.FacebookConnectCallback)

		protected := v1.Group("")
		protected.Use(middlewares.AuthRequired(jwtMgr))
		protected.Use(middlewares.CasbinMiddleware(enforcer))
		{
			protected.GET("/users/me", ctl.Me)

			protected.GET("/pages", ctl.ListPages)
			protected.POST("/pages/connect", ctl.ConnectPage)
			protected.POST("/pages/connect/complete", ctl.CompletePageConnection)

			protected.GET("/conversations", ctl.ListConversations)
			protected.GET("/conversations/:id/messages", ctl.ListMessages)
			protected.POST("/conversations/:id/messages", ctl.SendMessage)

			protected.GET("/auto-replies", ctl.ListAutoReplies)
			protected.POST("/auto-replies", ctl.CreateAutoReply)
			protected.PUT("/auto-replies/:id", ctl.UpdateAutoReply)

			// Notifications
			protected.GET("/notifications", ctl.ListNotifications)
			protected.GET("/notifications/unread-count", ctl.GetUnreadNotificationCount)
			protected.POST("/notifications/mark-read", ctl.MarkNotificationsRead)
			protected.PUT("/notifications/:id/mark-read", ctl.MarkNotificationRead)
			protected.POST("/notifications/mark-all-read", ctl.MarkAllNotificationsRead)

			// Conversations
			protected.POST("/conversations/:id/mark-read", ctl.MarkConversationRead)
		}
	}
}
