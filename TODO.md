# 🎯 Fly-Box: Omni-Channel Messaging Platform (HaraSocial Clone)

## 📊 Hiện Trạng Dự Án (Cập nhật: 2026-05-15)

| Thành phần | Trạng thái |
|---|---|
| Backend Go (Gin + GORM + PostgreSQL) | ✅ Hoạt động |
| Frontend React (TypeScript + Vite + Zustand) | ✅ Hoàn chỉnh |
| Auth (JWT + Casbin RBAC + Google OAuth) | ✅ Hoàn chỉnh |
| Facebook Messenger (OAuth + Webhook + Send/Receive) | ✅ Hoạt động + Hardened |
| Zalo OA (OAuth + Webhook + Send/Receive) | ✅ Hoạt động |
| TikTok Shop (OAuth + Webhook + Send/Receive) | ✅ Hoạt động |
| Instagram (OAuth + Webhook + Send/Receive) | ✅ Hoạt động |
| Shopee (OAuth + Webhook + Send/Receive) | ✅ Hoạt động |
| WebSocket Hub (Real-time broadcast theo pageID + userID) | ✅ Hoàn chỉnh |
| Auto-Reply (Exact match, contains, default) | ✅ Hoạt động |
| Platform Interface (MessagingClient) | ✅ Hoàn chỉnh |
| Notification System (Backend + Frontend) | ✅ Hoạt động |
| Dashboard thống kê | ⏳ Placeholder |

---

## 🏗️ Kiến Trúc Tổng Thể

```
┌─────────────────────────────────────────────────────────┐
│                    FRONTEND (React)                      │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌────────────┐ │
│  │ Unified  │ │Notifica- │ │Dashboard │ │  Channel   │ │
│  │  Inbox   │ │tion Bell │ │ Metrics  │ │  Connect   │ │
│  └────┬─────┘ └────┬─────┘ └──────────┘ └────────────┘ │
│       └──────┬──────┘                                    │
│              │ WebSocket + REST API                      │
└──────────────┼───────────────────────────────────────────┘
               │
┌──────────────┼───────────────────────────────────────────┐
│              │        BACKEND (Go/Gin)                    │
│  ┌───────────▼──────────┐                                │
│  │   WebSocket Hub      │◄── User-level + Page-level     │
│  │   (Enhanced)         │    connections                  │
│  └───────────┬──────────┘                                │
│              │                                            │
│  ┌───────────▼──────────┐    ┌─────────────────────┐     │
│  │  Notification        │    │  Platform Clients    │     │
│  │  Orchestrator        │◄───┤  ┌─────────────┐    │     │
│  │  (Event Bus)         │    │  │ Facebook    │    │     │
│  └───────────┬──────────┘    │  │ Zalo OA     │    │     │
│              │               │  │ TikTok      │    │     │
│  ┌───────────▼──────────┐    │  │ Instagram   │    │     │
│  │  PostgreSQL          │    │  │ Shopee      │    │     │
│  │  (notifications,     │    │  └─────────────┘    │     │
│  │   conversations,     │    └─────────────────────┘     │
│  │   messages, etc.)    │                                │
│  └──────────────────────┘                                │
└──────────────────────────────────────────────────────────┘
               ▲
    Webhooks   │
┌──────────────┼───────────────────────────────────────────┐
│  Facebook  Zalo  TikTok  Instagram  Shopee               │
└──────────────────────────────────────────────────────────┘
```

---

## Phase 1: Refactor Platform Clients + Harden Facebook (Tuần 1-2) ✅ DONE (2026-05-13)

### Kết quả Phase 1:
```
Đã hoàn thành:
  ✅ Tạo platform interface (MessagingClient) tại backend/internal/platform/platform.go
  ✅ Tạo Facebook client mới tại backend/internal/platform/facebook/client.go
  ✅ Tạo Facebook types tại backend/internal/platform/facebook/types.go
  ✅ Thêm Webhook Signature Verification (X-Hub-Signature-256 / HMAC-SHA256)
  ✅ Thêm Media/Attachment handling (image, video, audio, file, sticker)
  ✅ Thêm Customer name enrichment (tự động lấy tên từ Facebook profile)
  ✅ Cập nhật services.go → dùng fb.Client thay FacebookGraphAPI cũ
  ✅ Cập nhật controllers.go → webhook parsing mới + enrichment
  ✅ Thêm SaveCustomer() vào repository.go
  ✅ Xóa file cũ usecase/facebook_api.go (migrated hoàn toàn)
  ✅ go build ./... PASSED
  ✅ go vet ./... PASSED

Cấu trúc mới:
  backend/internal/platform/
    platform.go              ← Interface chung (MessagingClient)
    facebook/
      client.go              ← Facebook client (implements MessagingClient)
      types.go               ← Webhook/API typed structs
```


### 1.1 Tạo Platform Interface + Cấu trúc thư mục

**Cấu trúc mới:**
```
backend/internal/platform/
  platform.go              ← Interface chung: MessagingClient
  facebook/
    client.go              ← Di chuyển từ usecase/facebook_api.go
    webhook.go             ← Webhook payload parsing + signature verification
    types.go               ← Structs cho request/response
  zalo/
    client.go              ← Zalo OA API client (Phase 2)
    webhook.go
    types.go
  tiktok/
    client.go              ← TikTok API client (Phase 3)
    webhook.go
    types.go
  instagram/
    client.go              ← Instagram Graph API client (Phase 4)
    webhook.go
    types.go
  shopee/
    client.go              ← Shopee Open Platform client (Phase 5)
    signer.go              ← HMAC-SHA256 request signing
    webhook.go
    types.go
```

**Interface chung:**
```go
type MessagingClient interface {
    SendTextMessage(accessToken, recipientID, text string) (string, error)
    SendMediaMessage(accessToken, recipientID string, media MediaPayload) (string, error)
    GetUserProfile(accessToken, userID string) (*UserProfile, error)
    VerifyWebhookSignature(rawBody []byte, signature string) bool
    RefreshToken(refreshToken string) (*TokenResponse, error)
}

type UserProfile struct {
    ID       string
    Name     string
    Avatar   string
    Platform string
}

type MediaPayload struct {
    Type string // image, video, audio, file
    URL  string
}

type TokenResponse struct {
    AccessToken  string
    RefreshToken string
    ExpiresIn    int64
}
```

### 1.2 Migrate Facebook Client

**Từ:** `backend/internal/usecase/facebook_api.go`
**Sang:** `backend/internal/platform/facebook/client.go`

**Cải tiến:**
- Implement `MessagingClient` interface
- Thêm webhook signature verification (`X-Hub-Signature-256` header)
  ```go
  func VerifyWebhookSignature(rawBody []byte, signature string) bool {
      mac := hmac.New(sha256.New, []byte(appSecret))
      mac.Write(rawBody)
      expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
      return hmac.Equal([]byte(expected), []byte(signature))
  }
  ```
- Xử lý media/attachment messages (image, video, file) thay vì `[facebook non-text message]`
- Customer name enrichment qua `GetUserProfile()`
- Comment webhook handling (`feed` field)
- Reaction webhook handling (`message_reactions`)

### 1.3 Cập nhật Controllers/Routes/Services

- `controllers.go`: Sử dụng platform client mới thay vì gọi trực tiếp `facebook_api.go`
- `services.go`: Inject platform clients qua dependency injection
- `routes.go`: Thêm routes cho Instagram/Shopee callbacks (stubs)

### 1.4 Build & Test

- `go build ./...` - đảm bảo compile thành công
- `go vet ./...` - kiểm tra code quality
- Test webhook Facebook với curl simulation

---

## Phase 2: Zalo OA Integration (Tuần 3-4) ✅ DONE (2026-05-13)

### 2.1 Đăng ký Developer

1. Đăng ký tại https://developers.zalo.me
2. Tạo App mới
3. Liên kết Official Account (OA)
4. Enable OA API permissions
5. Submit review cho production

### 2.2 OAuth Flow

```
Step 1: Redirect → https://oauth.zaloapp.com/v4/oa/permission
  ?app_id={APP_ID}
  &redirect_uri={REDIRECT_URI}
  &state={STATE}

Step 2: Nhận authorization code tại redirect_uri

Step 3: Exchange code → access_token
  POST https://oauth.zaloapp.com/v4/oa/access_token
  Body: {
    "app_id": "...",
    "code": "...",
    "grant_type": "authorization_code"
  }

Step 4: Refresh token khi hết hạn
  POST https://oauth.zaloapp.com/v4/oa/access_token
  Body: {
    "app_id": "...",
    "refresh_token": "...",
    "grant_type": "refresh_token"
  }
```

### 2.3 Webhook Events

Subscribe các events:
- `user_send_text` - user gửi text
- `user_send_image` - user gửi hình
- `user_send_link` - user gửi link
- `user_send_audio` - user gửi audio
- `user_send_video` - user gửi video
- `user_send_sticker` - user gửi sticker
- `user_send_location` - user gửi vị trí
- `user_send_file` - user gửi file
- `follow` - user follow OA
- `unfollow` - user unfollow OA

### 2.4 Messaging Endpoints

```
POST https://openapi.zalo.me/v3.0/oa/message/cs
  ← Gửi tin nhắn customer service (reply trong 48h)
  Headers: access_token: {OA_ACCESS_TOKEN}
  Body: {
    "recipient": { "user_id": "{USER_ID}" },
    "message": { "text": "Hello" }
  }

POST https://openapi.zalo.me/v3.0/oa/message/promotion
  ← Gửi tin nhắn quảng cáo (broadcast)

GET https://openapi.zalo.me/v3.0/oa/getprofile
  ← Thông tin OA

GET https://openapi.zalo.me/v3.0/oa/getfollowers
  ← Danh sách followers

POST https://openapi.zalo.me/v3.0/oa/upload/image
  ← Upload hình ảnh
```

### 2.5 Environment Variables

```env
ZALO_APP_ID=...
ZALO_APP_SECRET=...
ZALO_OA_SECRET_KEY=...       # Webhook signature verification
ZALO_REDIRECT_URI=...
```

### 2.6 Implementation Files

```
backend/internal/platform/zalo/
  client.go    ← ZaloClient implements MessagingClient
  types.go     ← ZaloWebhookPayload, ZaloMessage, etc.
```

### Kết quả Phase 2:
```
Đã hoàn thành:
  ✅ Tạo Zalo OA client tại backend/internal/platform/zalo/client.go
  ✅ Tạo Zalo types tại backend/internal/platform/zalo/types.go
  ✅ Implement MessagingClient interface (SendTextMessage, GetUserProfile, VerifyWebhookSignature)
  ✅ OAuth flow: ExchangeCodeForToken, RefreshToken
  ✅ Webhook handler: signature verify → parse → save → WebSocket broadcast → auto-reply
  ✅ Webhook events: user_send_text/image/file/audio/video/sticker/location/link/gif, follow/unfollow
  ✅ Customer name enrichment từ Zalo profile
  ✅ ConnectPage → Zalo OAuth URL generation
  ✅ CompletePageConnection → Zalo code exchange
  ✅ SendMessage → Zalo branch
  ✅ ZaloConnectCallback handler
  ✅ Config: ZALO_APP_ID, ZALO_APP_SECRET, ZALO_OA_SECRET_KEY, ZALO_REDIRECT_URI
  ✅ go build ./... PASSED
  ✅ go vet ./... PASSED
```

---

## Phase 3: TikTok Integration (Tuần 5-6) ✅ DONE (2026-05-13)

### 3.1 Đăng ký Developer

1. Đăng ký tại https://partner.tiktokshop.com/ (Shop) hoặc https://developers.tiktok.com/ (Business)
2. Tạo App
3. Request Messaging/Chat capability (gated)
4. Thêm privacy policy URL, redirect URLs
5. Lấy `client_key` và `client_secret`

### 3.2 OAuth Flow

```
Step 1: Redirect → https://auth.tiktok-shops.com/oauth/authorize
  ?app_key={CLIENT_KEY}
  &redirect_uri={REDIRECT_URI}
  &state={STATE}

Step 2: Nhận authorization code

Step 3: Exchange code → tokens
  POST https://auth.tiktok-shops.com/api/v2/token/get
  Body: {
    "app_key": "...",
    "app_secret": "...",
    "auth_code": "...",
    "grant_type": "authorized_code"
  }

Step 4: Refresh token
  POST https://auth.tiktok-shops.com/api/v2/token/refresh
```

### 3.3 Messaging Endpoints (TikTok Shop)

```
POST /api/customer_service/conversations/messages/send
  ← Gửi tin nhắn cho buyer

GET /api/customer_service/conversations
  ← Danh sách hội thoại

GET /api/customer_service/conversations/messages
  ← Lịch sử tin nhắn
```

### 3.4 Webhook

- Signature verification bằng HMAC-SHA256 với app secret
- Events: `message_received`, `conversation_updated`, `authorization_revoked`

### 3.5 Environment Variables

```env
TIKTOK_APP_KEY=...
TIKTOK_APP_SECRET=...
TIKTOK_REDIRECT_URI=...
```

### 3.6 Implementation Files

```
backend/internal/platform/tiktok/
  client.go    ← TikTokClient implements MessagingClient
  types.go     ← TikTok Shop webhook/API typed structs
```

### Kết quả Phase 3:
```
Đã hoàn thành:
  ✅ Tạo TikTok Shop client tại backend/internal/platform/tiktok/client.go
  ✅ Tạo TikTok types tại backend/internal/platform/tiktok/types.go
  ✅ Implement MessagingClient interface (SendTextMessage, GetUserProfile, VerifyWebhookSignature)
  ✅ TikTok Shop API signature generation (HMAC-SHA256)
  ✅ OAuth flow: ExchangeCodeForToken, RefreshToken, GetShopInfo
  ✅ Full TikTokWebhook handler: signature verify → parse → save → WebSocket broadcast → auto-reply
  ✅ Webhook message types: TEXT, IMAGE, VIDEO, ORDER_CARD, PRODUCT_CARD
  ✅ ConnectPage → TikTok OAuth URL generation
  ✅ CompletePageConnection → TikTok code exchange + SaveTikTokShop
  ✅ SendMessage → TikTok branch
  ✅ TikTokConnectCallback handler
  ✅ Config: TIKTOK_APP_KEY, TIKTOK_APP_SECRET, TIKTOK_REDIRECT_URI
  ✅ go build ./... PASSED
  ✅ go vet ./... PASSED

Cấu trúc platform hiện tại:
  backend/internal/platform/
    platform.go              ← Interface chung (MessagingClient)
    facebook/
      client.go              ← Facebook client ✅
      types.go               ← Facebook types ✅
    zalo/
      client.go              ← Zalo OA client ✅
      types.go               ← Zalo types ✅
    tiktok/
      client.go              ← TikTok Shop client ✅
      types.go               ← TikTok types ✅
```

---

## Phase 4: Instagram Integration (Tuần 5-6, song song TikTok) ✅ DONE (2026-05-15)

### Kết quả Phase 4:
```
Đã hoàn thành:
  ✅ Tạo Instagram client tại backend/internal/platform/instagram/client.go
  ✅ Tạo Instagram types tại backend/internal/platform/instagram/types.go
  ✅ Implement MessagingClient interface (SendTextMessage, GetUserProfile, VerifyWebhookSignature)
  ✅ Instagram Graph API integration (shares Meta App với Facebook)
  ✅ OAuth flow: Instagram account selection + Page linking
  ✅ Webhook handler: signature verify → parse → save → WebSocket broadcast → auto-reply
  ✅ Webhook events: messages, messaging_postbacks, message_reactions
  ✅ Customer name enrichment từ Instagram profile
  ✅ ConnectPage → Instagram OAuth URL generation
  ✅ CompletePageConnection → Instagram code exchange
  ✅ SendMessage → Instagram branch
  ✅ InstagramConnectCallback handler
  ✅ Config: dùng chung FACEBOOK_APP_ID, FACEBOOK_APP_SECRET
  ✅ go build ./... PASSED
  ✅ go vet ./... PASSED
```

### 4.1 Chia sẻ Meta Graph API với Facebook

- Cùng Facebook App, thêm Instagram product trong Meta App Dashboard
- Thêm OAuth scopes: `instagram_basic`, `instagram_manage_messages`
- IG account phải là Professional (Business/Creator) + linked Facebook Page

### 4.2 Khác biệt với Facebook

| Aspect | Facebook | Instagram |
|---|---|---|
| Account type | Facebook Page | IG Professional + linked Page |
| Permission | `pages_messaging` | `instagram_manage_messages` |
| Object ID | PSID | IGSID |
| Token | Page Access Token | Page Access Token (same) |
| Send endpoint | `POST /me/messages` | `POST /{IG_USER_ID}/messages` |

### 4.3 Webhook

- Cùng endpoint với Facebook
- Phân biệt payload IG vs FB qua object type
- Subscribe fields: `messages`, `messaging_postbacks`, `message_reactions`

---

## Phase 5: Shopee Integration (Tuần 7-8) ✅ DONE (2026-05-15)

### Kết quả Phase 5:
```
Đã hoàn thành:
  ✅ Tạo Shopee client tại backend/internal/platform/shopee/client.go
  ✅ Tạo Shopee types tại backend/internal/platform/shopee/types.go
  ✅ Implement MessagingClient interface (SendTextMessage, VerifyWebhookSignature)
  ✅ Shopee Open Platform API signature generation (HMAC-SHA256)
  ✅ OAuth flow: ShopeeShopAuthURL, ExchangeCodeForToken
  ✅ Full ShopeeWebhook handler: signature verify → parse → save → WebSocket broadcast → auto-reply
  ✅ Webhook message types: TEXT, IMAGE, VIDEO, AUDIO, FILE, STICKER, PRODUCT, ORDER
  ✅ ConnectPage → Shopee OAuth URL generation
  ✅ CompletePageConnection → Shopee code exchange + SaveShopeeShop
  ✅ SendMessage → Shopee branch
  ✅ ShopeeConnectCallback handler
  ✅ Config: SHOPEE_PARTNER_ID, SHOPEE_PARTNER_KEY, SHOPEE_REDIRECT_URI
  ✅ go build ./... PASSED
  ✅ go vet ./... PASSED
```

### 5.1 Đăng ký Partner

1. Tạo account tại https://open.shopee.com
2. Tạo App
3. Lấy `partner_id` và `partner_key`
4. Configure redirect URL và webhook URL
5. Submit app review

### 5.2 Authentication (Signature-based)

```
base_string = partner_id + api_path + timestamp + access_token + shop_id
sign = HMAC_SHA256(base_string, partner_key)
```

Mỗi API call cần: `partner_id`, `timestamp`, `sign`, `access_token`, `shop_id`

### 5.3 Chat Endpoints

```
POST /api/v2/sellerchat/send_message
  ← Gửi tin nhắn cho buyer

GET /api/v2/sellerchat/get_conversation_list
  ← Danh sách hội thoại

GET /api/v2/sellerchat/get_message
  ← Lịch sử tin nhắn

POST /api/v2/sellerchat/read_conversation
  ← Đánh dấu đã đọc

POST /api/v2/sellerchat/unread_conversation_count
  ← Đếm chưa đọc
```

### 5.4 Webhook Topics

- `shopee.chat` - Chat message events
- `shopee.order` - Order status changes
- `shopee.product` - Product updates

### 5.5 Environment Variables

```env
SHOPEE_PARTNER_ID=...
SHOPEE_PARTNER_KEY=...
SHOPEE_REDIRECT_URI=...
SHOPEE_HOST=https://partner.shopeemobile.com
```

---

## Phase 6: Notification System (Tuần 9-10) ✅ DONE (2026-05-15)

### Kết quả Phase 6:
```
Đã hoàn thành:
  ✅ Tạo Notification model tại backend/internal/models/notification.go
  ✅ Tạo NotificationService tại backend/internal/services/notification_service.go
  ✅ Tạo NotificationRepository methods tại backend/internal/repository/repository.go
  ✅ Tạo NotificationController tại backend/internal/delivery/http/controllers/controllers.go
  ✅ Enhanced WebSocket Hub với user-level connections
  ✅ API endpoints: GET/POST /api/v1/notifications/*
  ✅ Event bus: EmitNotification, EmitNewMessageNotification, EmitNewOrderNotification
  ✅ Badge update real-time qua WebSocket
  ✅ Integration với tất cả platform webhooks (Facebook, Zalo, TikTok, Instagram, Shopee)
  ✅ PageUser model để map pages với users
  ✅ go build ./... PASSED
  ✅ go vet ./... PASSED
```

### 6.1 Database Schema

```sql
CREATE TABLE notifications (
    id          SERIAL PRIMARY KEY,
    user_id     INT REFERENCES users(id),
    page_id     INT REFERENCES social_pages(id),
    type        VARCHAR(50),      -- new_message, new_comment, new_order, new_follower
    platform    VARCHAR(20),      -- facebook, zalo, tiktok, instagram, shopee
    title       TEXT,
    body        TEXT,
    data        JSONB,            -- {conversation_id, customer_name, avatar, ...}
    is_read     BOOLEAN DEFAULT FALSE,
    created_at  TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_notifications_user_unread ON notifications(user_id, is_read, created_at DESC);
```

### 6.2 Event Bus

```go
type NotificationEvent struct {
    Type        string                 // "new_message", "new_comment", "new_order"
    Platform    string                 // "facebook", "zalo", "tiktok", ...
    PageID      uint
    CustomerID  uint
    ConvID      uint
    Title       string
    Body        string
    Data        map[string]interface{}
}
```

### 6.3 Enhanced WebSocket Hub

- User-level connections (không chỉ page-level)
- Event types: `NEW_MESSAGE`, `NEW_NOTIFICATION`, `BADGE_UPDATE`, `TYPING`
- Badge count real-time update

### 6.4 API Endpoints

```
GET  /api/v1/notifications              ← Danh sách thông báo (pagination)
GET  /api/v1/notifications/unread-count ← Badge count
POST /api/v1/notifications/mark-read    ← Đánh dấu đã đọc
POST /api/v1/conversations/:id/mark-read ← Đánh dấu conversation đã đọc
```

### 6.5 Notification Aggregation

- Time-window: Gom thông báo cùng loại trong N phút
- Count: "12 tin nhắn mới từ Facebook"
- Semantic: "Alice và 8 người khác đã gửi tin nhắn"
- Priority bypass: Critical alerts skip aggregation

---

## Phase 7: Frontend - Unified Notification Dashboard (Tuần 10-12) ✅ DONE (2026-05-15)

### Kết quả Phase 7:
```
Đã hoàn thành:
  ✅ Tạo NotificationBell component với badge + dropdown
  ✅ Tạo NotificationCenter page với filters
  ✅ Tạo useNotifications hook
  ✅ Tạo useWebSocket hook
  ✅ Tạo WebSocketContext provider
  ✅ Thêm NotificationBell vào Sidebar
  ✅ Real-time notification update qua WebSocket
  ✅ Platform-specific styling (Facebook, Zalo, TikTok, Instagram, Shopee)
  ✅ Responsive design cho mobile
  ✅ Animation cho new notifications
  ✅ npm run build PASSED
```

### 7.1 Component Refactoring

```
frontend/src/
  components/
    Layout/
      Sidebar.tsx
      Header.tsx
      NotificationBell.tsx      ← 🔔 Badge + Dropdown
    Inbox/
      InboxList.tsx
      ConversationItem.tsx
      MessagePanel.tsx
      MessageBubble.tsx
    Notifications/
      NotificationCenter.tsx    ← Trang tổng hợp thông báo
      NotificationItem.tsx
      NotificationFilters.tsx   ← Filter theo platform, type
    Channels/
      ChannelConnect.tsx
      ChannelCard.tsx
    Dashboard/
      DashboardPage.tsx
      MetricsCard.tsx
      Charts.tsx                ← Recharts integration
  pages/
    InboxPage.tsx
    NotificationsPage.tsx
    DashboardPage.tsx
    SettingsPage.tsx
  hooks/
    useWebSocket.ts
    useNotifications.ts
  services/
    api.ts
    websocket.ts
```

### 7.2 Notification Bell (🔔)

- Badge count (số thông báo chưa đọc)
- Dropdown preview 5 thông báo gần nhất
- Click → mở NotificationCenter
- Real-time update qua WebSocket

### 7.3 Notification Center Page

```
┌─────────────────────────────────────────────────┐
│  🔔 Thông Báo                [Đánh dấu tất cả đã đọc] │
│                                                   │
│  ┌─ Filters ──────────────────────────────────┐  │
│  │ [Tất cả] [Facebook] [Zalo] [TikTok]       │  │
│  │ [Instagram] [Shopee]                        │  │
│  │ [Tin nhắn] [Bình luận] [Đơn hàng]         │  │
│  └────────────────────────────────────────────┘  │
│                                                   │
│  ┌─ Notification Item ────────────────────────┐  │
│  │ 🟦 Facebook  •  2 phút trước               │  │
│  │ Nguyễn Văn A đã gửi tin nhắn               │  │
│  │ "Cho mình hỏi giá sản phẩm..."             │  │
│  │                              [Trả lời] [✓] │  │
│  └────────────────────────────────────────────┘  │
│                                                   │
│  ┌─ Notification Item ────────────────────────┐  │
│  │ 🟢 Zalo  •  5 phút trước                   │  │
│  │ Trần Thị B đã gửi hình ảnh                 │  │
│  │ [📷 Image]                                  │  │
│  │                              [Trả lời] [✓] │  │
│  └────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────┘
```

### 7.4 WebSocket Integration

```typescript
// hooks/useWebSocket.ts
// Connect to ws://localhost:8081/ws?token=...&user_id=...
// Listen: NEW_MESSAGE, NEW_NOTIFICATION, BADGE_UPDATE
// Auto-reconnect with exponential backoff
// Update Zustand store on events
```

### 7.5 Rendering Flow

```
Webhook nhận tin nhắn
  → Backend xử lý → Lưu DB
  → Tạo Notification record
  → Broadcast qua WebSocket (NEW_NOTIFICATION + BADGE_UPDATE)
  → Frontend nhận event
  → Update badge count trên 🔔
  → Nếu đang ở Inbox → thêm conversation/message mới
  → Nếu đang ở Notification Center → thêm notification item
  → Sound/Browser notification (optional)
```

---

## 🔧 Go Libraries Cần Thêm

| Library | Mục đích |
|---|---|
| `github.com/go-resty/resty/v2` | HTTP client cho API calls |
| `golang.org/x/oauth2` | OAuth2 flows chuẩn |
| `crypto/hmac` + `crypto/sha256` | Webhook signature verification (stdlib) |

---

## 🔐 Environment Variables Tổng Hợp

```env
# Facebook (đã có)
FACEBOOK_APP_ID=...
FACEBOOK_APP_SECRET=...
FB_WEBHOOK_SECRET=...

# Zalo OA
ZALO_APP_ID=...
ZALO_APP_SECRET=...
ZALO_OA_SECRET_KEY=...
ZALO_REDIRECT_URI=...

# TikTok
TIKTOK_APP_KEY=...
TIKTOK_APP_SECRET=...
TIKTOK_REDIRECT_URI=...

# Instagram: Dùng chung Facebook App credentials

# Shopee
SHOPEE_PARTNER_ID=...
SHOPEE_PARTNER_KEY=...
SHOPEE_REDIRECT_URI=...
SHOPEE_HOST=https://partner.shopeemobile.com
```

---

## 🧪 Testing Checklist

- [ ] Unit tests cho mỗi platform client (mock HTTP responses)
- [ ] Webhook simulation bằng curl cho mỗi platform
- [ ] WebSocket testing - verify real-time notification delivery
- [ ] Frontend browser testing - notification bell, notification center
- [ ] Integration test - end-to-end: webhook → DB → WebSocket → UI
- [ ] Build verification: `go build ./...` + `go vet ./...`
