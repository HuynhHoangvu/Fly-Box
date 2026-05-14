# Kế Hoạch Kiến Trúc Backend: Hệ Thống Quản Lý Tin Nhắn Omni-channel (HaraSocial Clone)

Hệ thống này được thiết kế để tập trung tin nhắn, bình luận từ nhiều nền tảng (Facebook, Zalo) về một dashboard duy nhất, cho phép nhân viên trực chat, cấu hình trả lời tự động và phân quyền chi tiết.

## 1. Tech Stack
* **Ngôn ngữ:** Golang
* **Web Framework:** Gin Gonic (Xử lý HTTP requests, Webhooks siêu tốc)
* **Database:** PostgreSQL (Lưu trữ dữ liệu có cấu trúc, quan hệ phức tạp)
* **ORM:** GORM (Tương tác với DB dễ dàng, auto-migration)
* **Authentication:** JWT (Xác thực API & WebSocket)
* **Authorization:** Casbin (RBAC/ABAC phân quyền chi tiết tới từng page/chức năng)
* **Real-time:** Gorilla WebSocket (Đẩy tin nhắn realtime xuống client)

---

## 2. Kiến Trúc Tổng Thể (Architecture)

1. **Webhook Layer:** Nhận event (tin nhắn, comment mới) từ Facebook Graph API, Zalo ZaloA API.
2. **Core Service Layer:** 
   * Xử lý/Chuẩn hóa payload từ các nền tảng khác nhau về một chuẩn chung của hệ thống.
   * Kiểm tra điều kiện **Auto-reply** (Trả lời tự động) dựa trên keyword hoặc rule.
   * Lưu trữ dữ liệu vào PostgreSQL.
3. **Realtime Layer:** Pub/Sub hoặc broadcast trực tiếp qua **Gorilla WebSocket** tới các Frontend Client (React/Vue/Angular) đang online.
4. **API Layer:** Cung cấp API cho Frontend (lấy lịch sử tin nhắn, cấu hình auto-reply, phân quyền).

---

## 3. Thiết Kế Cơ Sở Dữ Liệu (PostgreSQL + GORM)

Dưới đây là các Model chính cần có:

### a. Phân quyền & Người dùng (Auth & RBAC)
* **`users`**: `id`, `email`, `password_hash`, `role`, `created_at`
* *Lưu ý: Bảng policies của Casbin sẽ do Casbin tự động tạo qua GORM Adapter.*

### b. Quản lý Kênh/Trang (Channels)
* **`social_pages`** (Trang FB, Zalo OA đã kết nối):
  * `id`, `platform` (facebook/zalo), `page_id` (ID nền tảng), `page_name`, `access_token`, `refresh_token`, `status`.

### c. Quản lý Tin nhắn & Bình luận
* **`customers`** (Khách hàng từ MXH): 
  * `id`, `platform_id`, `name`, `avatar`, `platform` (facebook/zalo).
* **`conversations`** (Cuộc hội thoại/Box chat):
  * `id`, `page_id`, `customer_id`, `type` (message/comment), `last_message`, `unread_count`, `updated_at`.
* **`messages`** (Chi tiết tin nhắn/bình luận):
  * `id`, `conversation_id`, `sender_type` (customer/page/system), `content_type` (text/image/video), `content`, `social_message_id`, `created_at`.

### d. Trả lời tự động (Auto-Reply)
* **`auto_reply_rules`**:
  * `id`, `page_id`, `rule_type` (exact_match/contains/default), `keyword` (mảng từ khóa), `reply_content`, `is_active`.

---

## 4. Thiết Kế API Endpoints (Gin Routing)

### Auth & User
* `POST /api/v1/auth/login` (Trả về JWT token)
* `GET /api/v1/users/me`

### Webhooks (Rất quan trọng - Cần Public)
* `GET /webhooks/facebook` (Xác thực verify_token từ FB)
* `POST /webhooks/facebook` (Nhận payload tin nhắn/comment từ FB)
* `POST /webhooks/zalo` (Nhận payload từ Zalo)

### Social Pages
* `GET /api/v1/pages` (Lấy danh sách page đã kết nối)
* `POST /api/v1/pages/connect` (Tích hợp page mới)

### Conversations & Messages
* `GET /api/v1/conversations?page_id=X` (Danh sách chat)
* `GET /api/v1/conversations/:id/messages` (Lịch sử tin nhắn)
* `POST /api/v1/conversations/:id/messages` (Nhân viên gửi tin nhắn -> Gọi API của FB/Zalo)

### Auto Reply Settings
* `GET /api/v1/auto-replies?page_id=X`
* `POST /api/v1/auto-replies`
* `PUT /api/v1/auto-replies/:id`

---

## 5. Flow Xử Lý Kỹ Thuật Chi Tiết

### Flow 1: Nhận tin nhắn & Đẩy Real-time
1. Khách hàng nhắn tin vào FB Fanpage.
2. FB gọi `POST /webhooks/facebook` của hệ thống.
3. **Golang (Gin)** nhận request, parse JSON.
4. Lưu User (nếu chưa có), lưu Conversation, lưu Message vào **PostgreSQL**.
5. Kích hoạt module **Gorilla WebSocket**: 
   * Gửi event `NEW_MESSAGE` đến tất cả nhân viên (client) đang subscribe vào `page_id` đó.
6. (Đồng thời) Chạy background job kiểm tra **Auto Reply**.

### Flow 2: Auto-Reply (Trả lời tự động)
1. Trong bước 5 ở trên, sau khi parse tin nhắn gốc, hệ thống query bảng `auto_reply_rules`.
2. Nếu `message.content` match với `keyword`.
3. Hệ thống tạo một message mới với `sender_type = system`.
4. Gọi API của Facebook/Zalo (VD: Graph API `/me/messages`) để gửi câu trả lời lại cho khách.
5. Lưu message này vào DB và broadcast qua WebSocket để màn hình nhân viên cập nhật.

### Flow 3: Gửi tin nhắn từ Nhân viên (Frontend)
1. Nhân viên gõ tin nhắn trên web, frontend gọi `POST /api/v1/conversations/:id/messages` (kèm JWT).
2. Backend kiểm tra quyền bằng **Casbin** (Nhân viên có quyền chat trên page này không?).
3. Gọi API của FB/Zalo để gửi tin.
4. Thành công -> Lưu DB -> Broadcast WebSocket cập nhật UI cho các tab khác.

### Flow 4: WebSocket Connection
* Endpoint: `ws://domain.com/ws?token=JWT_TOKEN`
* Khi kết nối, middleware phân giải JWT để lấy UserID.
* Tổ chức Connection Manager dạng: `map[page_id][]*websocket.Conn` để biết user nào đang trực page nào mà gửi data cho đúng.

---

## 6. Các Bước Triển Khai (Roadmap)

* **Giai đoạn 1: Base & Auth**
  * Khởi tạo project Golang chuẩn (Clean Architecture hoặc MVC).
  * Setup PostgreSQL, GORM migration.
  * Tích hợp JWT Auth và setup Casbin GORM Adapter.
* **Giai đoạn 2: Webhooks & Social API Integration**
  * Đăng ký App Facebook (Meta Developer), setup Zalo OA.
  * Viết Webhook handlers để nhận chuẩn data.
  * Viết module Service để gọi API gửi tin nhắn ra ngoài (Send API).
* **Giai đoạn 3: Realtime & WebSocket**
  * Tích hợp Gorilla WebSocket.
  * Viết Hub/Client manager để quản lý các socket connections.
* **Giai đoạn 4: Auto-Reply & Business Logic**
  * Xây dựng API CRUD cho rule auto-reply.
  * Viết parser kiểm tra chuỗi (Regex, Contains) để kích hoạt auto-reply ngay khi nhận webhook.