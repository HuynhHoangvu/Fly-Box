# Authentication Logic Analysis

## Overview
Tài liệu phân tích luồng đăng nhập/đăng xuất và các path liên quan.

## Frontend (React + Zustand)

### Files:
- `frontend/src/store/useAuthStore.ts` - Auth state management
- `frontend/src/services/api.ts` - API calls
- `frontend/src/pages/auth/LoginPage.tsx` - Login UI

### Luồng Đăng Nhập (Login)

#### 1. Login với Credentials (Email + Password)
```typescript
// useAuthStore.ts
loginWithCredentials: async (email: string, password: string) => {
  const { data } = await authAPI.login(email, password);
  // Lưu token và user vào localStorage
  localStorage.setItem('token', token);
  localStorage.setItem('user', JSON.stringify(user));
  set({ user, token, isAuthenticated: true });
}
```

**API Call:**
- Method: `POST`
- Path: `/api/v1/auth/login`
- Body: `{ email, password }`
- Response: `{ token, user }`

#### 2. Login với Google OAuth
```typescript
loginWithGoogle: async (idToken: string) => {
  const { data } = await authAPI.loginWithGoogle(idToken);
  // ...same flow
}
```

**API Call:**
- Method: `POST`
- Path: `/api/v1/auth/login`
- Body: `{ id_token: string }`
- Response: `{ token, user }`

#### 3. Dev Mode (Luôn được authenticate)
```typescript
// useAuthStore.ts
const initialAuth = getInitialAuth();
// isAuthenticated = true (always for dev)
```

### Luồng Đăng Ký (Register)

**API Call:**
- Method: `POST`
- Path: `/api/v1/auth/register`
- Body: `{ name, email, password }`
- Response: `{ token, user }`

### Luồng Đăng Xuất (Logout)

```typescript
logout: () => {
  localStorage.removeItem('token');
  localStorage.removeItem('user');
  set({ user: null, token: null, isAuthenticated: false });
}
```

### Kiểm Tra Auth (checkAuth)
```typescript
checkAuth: async () => {
  // DEV MODE: Always authenticated, no need to check with server
  return;
}
```

---

## Backend (Go + Gin)

### Files:
- `backend/internal/delivery/http/controllers/controllers.go` - Auth handlers
- `backend/internal/delivery/http/routes/routes.go` - Route definitions
- `backend/internal/delivery/http/middlewares/auth.go` - JWT middleware

### HTTP Paths

| Method | Path | Mô Tả | Auth Required |
|--------|-----|-------|-------------|
| POST | `/api/v1/auth/login` | Login với email/password hoặc Google id_token | ❌ |
| POST | `/api/v1/auth/register` | Register new user | ❌ |
| GET | `/api/v1/users/me` | Get current user info | ✅ |

### Handler: Login (`ctl.Login`)

```go
func (ctl *Controller) Login(c *gin.Context) {
  var req struct {
    Email    string `json:"email"`
    Password string `json:"password"`
    IDToken  string `json:"id_token"`
  }
  
  // 1. Google OAuth flow (id_token != "")
  if req.IDToken != "" {
    // Tạo user mới nếu chưa tồn tại
    // Generate JWT token
    token, _ := ctl.JWT.Generate(user.ID, user.Email, user.Role)
    return { token, user }
  }
  
  // 2. Email + Password flow
  user, _ := ctl.Repo.GetUserByEmail(email)
  bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
  
  // Generate JWT token
  token, _ := ctl.JWT.Generate(user.ID, user.Email, user.Role)
  return { token, user }
}
```

**Features:**
- ✅ Tự động tạo user mới khi login Google lần đầu
- ✅ Hỗ trợ password cũ (plaintext) và tự migrate sang bcrypt
- ✅ Validate input

### Handler: Register (`ctl.Register`)

```go
func (ctl *Controller) Register(c *gin.Context) {
  // Validate email, password
  // Check email exists
  // Hash password với bcrypt
  // Tạo user mới với role "staff"
  // Generate JWT token
  return { token, user }
}
```

### Handler: Me (`ctl.Me`)

```go
func (ctl *Controller) Me(c *gin.Context) {
  claims := middlewares.GetClaims(c)
  user, _ := ctl.Repo.GetUserByID(claims.UserID)
  return { user }
}
```

### JWT Middleware

```go
// middlewares/auth.go
func AuthRequired(jwtMgr *JWTManager) gin.HandlerFunc {
  return func(c *gin.Context) {
    authHeader := c.GetHeader("Authorization")
    tokenString := strings.TrimPrefix(authHeader, "Bearer ")
    
    claims, err := jwtMgr.Validate(tokenString)
    if err != nil {
      c.JSON(401, gin.H{"error": "unauthorized"})
      c.Abort()
      return
    }
    
    c.Set("user_id", claims.UserID)
    c.Next()
  }
}
```

---

## Data Flow

### Login Flow
```
[User] 
  → LoginPage (email/password) 
    → useAuthStore.loginWithCredentials()
      → api.ts → POST /api/v1/auth/login
        → Backend: Login() 
          → Validate credentials
          → Generate JWT
        → Return { token, user }
      → Store to localStorage
    → isAuthenticated = true
```

### Logout Flow
```
[User] 
  → Click Logout
    → useAuthStore.logout()
      → Clear localStorage
      → isAuthenticated = false
      → Redirect to /login
```

### Protected Request Flow
```
[API Call] 
  → Add "Authorization: Bearer <token>" header
    → Backend middleware validates JWT
      → Set user_id in context
    → Handler processes request
```

---

## Security Notes

### ✅ Implemented
1. **JWT Token** - Bearer token in Authorization header
2. **Bcrypt password hashing** - Secure password storage
3. **Casbin RBAC** - Role-based access control
4. **Token validation** - Middleware validates every protected request

### ⚠️ Issues / Cần Cải Thiện
1. **Dev mode bypass** - `isAuthenticated = true` always in dev mode
2. **No token refresh** - Token không có expiration handling
3. **No logout invalidation** - Token bị revoke khi logout (cần blacklisting)

---

## API Endpoints Summary

```
Public:
  POST /api/v1/auth/login      # Login
  POST /api/v1/auth/register   # Register
  
Protected (JWT required):
  GET  /api/v1/users/me       # Get current user
  GET  /api/v1/pages          # List pages
  POST /api/v1/pages/connect  # Connect platform
  ...
```

---

## Test Commands

### Login
```bash
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@harasocial.local","password":"admin123"}'
```

### Register  
```bash
curl -X POST http://localhost:8081/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Test User","email":"test@example.com","password":"password123"}'
```

### Protected (với token)
```bash
curl -X GET http://localhost:8081/api/v1/users/me \
  -H "Authorization: Bearer <token>"
