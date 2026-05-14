package middlewares

import (
	"log"
	"net/http"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RBAC model definition (request = sub, obj, act)
const casbinModelText = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && keyMatch2(r.obj, p.obj) && regexMatch(r.act, p.act)
`

// InitCasbin creates a Casbin enforcer with GORM adapter and loads default policies
func InitCasbin(db *gorm.DB) *casbin.Enforcer {
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		log.Fatalf("failed to create casbin adapter: %v", err)
	}

	m, err := model.NewModelFromString(casbinModelText)
	if err != nil {
		log.Fatalf("failed to create casbin model: %v", err)
	}

	enforcer, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		log.Fatalf("failed to create casbin enforcer: %v", err)
	}

	if err := enforcer.LoadPolicy(); err != nil {
		log.Fatalf("failed to load casbin policies: %v", err)
	}

	loadDefaultPolicies(enforcer)

	return enforcer
}

// loadDefaultPolicies seeds RBAC policies if they don't exist yet
func loadDefaultPolicies(e *casbin.Enforcer) {
	policies := [][]string{
		// Admin: full access to all API endpoints
		{"admin", "/api/v1/*", "GET|POST|PUT|PATCH|DELETE"},

		// Staff: read own profile
		{"staff", "/api/v1/users/me", "GET"},

		// Staff: pages (list, connect)
		{"staff", "/api/v1/pages", "GET|POST"},
		{"staff", "/api/v1/pages/connect", "POST"},
		{"staff", "/api/v1/pages/connect/complete", "POST"},

		// Staff: conversations & messages (read + send)
		{"staff", "/api/v1/conversations", "GET"},
		{"staff", "/api/v1/conversations/:id/messages", "GET|POST"},

		// Staff: auto-replies (read + create + update)
		{"staff", "/api/v1/auto-replies", "GET|POST|PUT"},

		// Staff: notifications
		{"staff", "/api/v1/notifications", "GET|POST"},
		{"staff", "/api/v1/notifications/*", "GET|POST|PUT"},
	}

	for _, p := range policies {
		hasPolicy, _ := e.HasPolicy(p)
		if !hasPolicy {
			_, _ = e.AddPolicy(p)
		}
	}

	_ = e.SavePolicy()
}

// CasbinMiddleware returns a Gin middleware that enforces RBAC policies
func CasbinMiddleware(enforcer *casbin.Enforcer) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaims(c)
		if claims == nil {
			log.Printf("[Casbin] Missing claims in context for path: %s", c.Request.URL.Path)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden: missing authentication context"})
			return
		}

		role := claims.Role
		if role == "" {
			role = "staff"
		}

		obj := c.Request.URL.Path
		act := c.Request.Method

		allowed, err := enforcer.Enforce(role, obj, act)
		if err != nil {
			log.Printf("[Casbin] Enforce error: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "authorization error"})
			return
		}

		if !allowed {
			log.Printf("[Casbin] Access denied: role=%s, obj=%s, act=%s", role, obj, act)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden: insufficient permissions for " + obj})
			return
		}

		c.Next()
	}
}
