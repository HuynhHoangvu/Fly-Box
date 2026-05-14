# Thorough API Testing Plan (cURL) - Omni-channel Backend

Prerequisites:
1. Install Go and ensure `go version` works.
2. Run PostgreSQL and set `DATABASE_URL` correctly.
3. Start server:
   ```bash
   cd Fly-Box/backend
   go mod tidy
   go run ./cmd/api
   ```
4. Base URL:
   - `http://localhost:8080`

---

## 0) Health check

```bash
curl -i http://localhost:8080/health
```

Expected:
- `200 OK`
- body: `{"status":"ok"}`

---

## 1) Auth

### 1.1 Login - happy path
```bash
curl -i -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"staff@example.com\",\"password\":\"123456\"}"
```

Expected:
- `200 OK`
- response has `token`

Save token:
```bash
TOKEN="<paste_token_here>"
```

### 1.2 Login - invalid payload
```bash
curl -i -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d "{}"
```

Expected:
- `400 Bad Request`

### 1.3 Me - authorized
```bash
curl -i http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer $TOKEN"
```

Expected:
- `200 OK`
- body includes decoded claim fields

### 1.4 Me - missing token
```bash
curl -i http://localhost:8080/api/v1/users/me
```

Expected:
- `401 Unauthorized`

### 1.5 Me - invalid token
```bash
curl -i http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer invalid.token.value"
```

Expected:
- `401 Unauthorized`

---

## 2) Facebook/Zalo Webhooks

### 2.1 FB verify token - valid
```bash
curl -i "http://localhost:8080/webhooks/facebook?hub.mode=subscribe&hub.verify_token=verify-token&hub.challenge=12345"
```

Expected:
- `200 OK`
- response body: `12345`

### 2.2 FB verify token - invalid
```bash
curl -i "http://localhost:8080/webhooks/facebook?hub.mode=subscribe&hub.verify_token=wrong&hub.challenge=12345"
```

Expected:
- `403 Forbidden`

### 2.3 FB webhook - valid JSON
```bash
curl -i -X POST http://localhost:8080/webhooks/facebook \
  -H "Content-Type: application/json" \
  -d "{\"object\":\"page\",\"entry\":[]}"
```

Expected:
- `200 OK`
- `{"status":"received","platform":"facebook"}`

### 2.4 FB webhook - invalid JSON
```bash
curl -i -X POST http://localhost:8080/webhooks/facebook \
  -H "Content-Type: application/json" \
  -d "{bad json}"
```

Expected:
- `400 Bad Request`

### 2.5 Zalo webhook - valid JSON
```bash
curl -i -X POST http://localhost:8080/webhooks/zalo \
  -H "Content-Type: application/json" \
  -d "{\"event\":\"message\",\"data\":{}}"
```

Expected:
- `200 OK`
- `{"status":"received","platform":"zalo"}`

### 2.6 Zalo webhook - invalid JSON
```bash
curl -i -X POST http://localhost:8080/webhooks/zalo \
  -H "Content-Type: application/json" \
  -d "{bad json}"
```

Expected:
- `400 Bad Request`

---

## 3) Social Pages APIs

### 3.1 List pages - unauthorized
```bash
curl -i http://localhost:8080/api/v1/pages
```

Expected:
- `401 Unauthorized`

### 3.2 Connect page - valid
```bash
curl -i -X POST http://localhost:8080/api/v1/pages/connect \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"platform\":\"facebook\",
    \"page_id\":\"fb_001\",
    \"page_name\":\"Fly Visa Fanpage\",
    \"access_token\":\"token_abc\",
    \"refresh_token\":\"refresh_xyz\"
  }"
```

Expected:
- `201 Created`

### 3.3 Connect page - invalid payload
```bash
curl -i -X POST http://localhost:8080/api/v1/pages/connect \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"platform\":\"facebook\"}"
```

Expected:
- `400 Bad Request`

### 3.4 List pages - authorized
```bash
curl -i http://localhost:8080/api/v1/pages \
  -H "Authorization: Bearer $TOKEN"
```

Expected:
- `200 OK`
- page list includes created page

---

## 4) Conversations & Messages APIs

Note: Current implementation requires existing conversations in DB for message list/send.

### 4.1 List conversations
```bash
curl -i "http://localhost:8080/api/v1/conversations?page_id=1" \
  -H "Authorization: Bearer $TOKEN"
```

Expected:
- `200 OK`

### 4.2 List messages - invalid conversation id
```bash
curl -i "http://localhost:8080/api/v1/conversations/abc/messages" \
  -H "Authorization: Bearer $TOKEN"
```

Expected:
- `400 Bad Request`

### 4.3 List messages - valid conversation id
```bash
curl -i "http://localhost:8080/api/v1/conversations/1/messages" \
  -H "Authorization: Bearer $TOKEN"
```

Expected:
- `200 OK` (may be empty array if no data)

### 4.4 Send message - invalid payload
```bash
curl -i -X POST "http://localhost:8080/api/v1/conversations/1/messages" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{}"
```

Expected:
- `400 Bad Request`

### 4.5 Send message - valid payload (requires conversation id=1 exists)
```bash
curl -i -X POST "http://localhost:8080/api/v1/conversations/1/messages" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"content\":\"Hello customer\"}"
```

Expected:
- `201 Created` if conversation exists
- if not exists, likely `500` in current scaffold (known gap to improve later)

---

## 5) Auto-reply APIs

### 5.1 Create auto-reply rule - exact_match
```bash
curl -i -X POST http://localhost:8080/api/v1/auto-replies \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"page_id\":1,
    \"rule_type\":\"exact_match\",
    \"keywords\":[\"xin chào\",\"hello\"],
    \"reply_content\":\"Fly Visa xin chào!\",
    \"is_active\":true
  }"
```

Expected:
- `201 Created`

### 5.2 Create auto-reply rule - contains
```bash
curl -i -X POST http://localhost:8080/api/v1/auto-replies \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"page_id\":1,
    \"rule_type\":\"contains\",
    \"keywords\":[\"giá\",\"price\"],
    \"reply_content\":\"Bạn vui lòng để lại SĐT để được báo giá.\"
  }"
```

Expected:
- `201 Created`

### 5.3 Create auto-reply rule - default
```bash
curl -i -X POST http://localhost:8080/api/v1/auto-replies \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"page_id\":1,
    \"rule_type\":\"default\",
    \"keywords\":[],
    \"reply_content\":\"Cảm ơn bạn đã nhắn tin.\"
  }"
```

Expected:
- `201 Created`

### 5.4 List auto-replies
```bash
curl -i "http://localhost:8080/api/v1/auto-replies?page_id=1" \
  -H "Authorization: Bearer $TOKEN"
```

Expected:
- `200 OK`

### 5.5 Update auto-reply - existing id
```bash
curl -i -X PUT "http://localhost:8080/api/v1/auto-replies/1" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"reply_content\":\"Nội dung trả lời mới\",\"is_active\":true}"
```

Expected:
- `200 OK` if id exists
- `404` if not found

### 5.6 Update auto-reply - invalid id
```bash
curl -i -X PUT "http://localhost:8080/api/v1/auto-replies/abc" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"reply_content\":\"x\"}"
```

Expected:
- `400 Bad Request`

---

## 6) WebSocket

Endpoint:
- `ws://localhost:8080/ws?page_id=1&token=<JWT>`

Manual validation checklist:
1. Open 2 websocket clients subscribed to same `page_id`.
2. Trigger `POST /api/v1/conversations/:id/messages`.
3. Verify both clients receive `NEW_MESSAGE` event payload.

CLI example (if `wscat` installed):
```bash
wscat -c "ws://localhost:8080/ws?page_id=1&token=$TOKEN"
```

---

## 7) Error/Edge Coverage Summary

- Missing auth header: should be `401`
- Invalid JWT: should be `401`
- Invalid JSON bodies: should be `400`
- Invalid path params: should be `400`
- Unknown auto-reply id on update: should be `404`
- DB unique constraint (duplicate `page_id`): should return error (currently generic `500`)

---

## 8) Known limitations in current scaffold to improve after tests

1. Casbin not yet wired into middleware checks.
2. Webhook -> conversation/message persistence flow is still stub-level (payload accepted but not fully normalized+saved).
3. `SendMessage` depends on existing conversation; no auto-create fallback yet.
4. WS token currently accepted via query but not validated at hub level (routing includes auth at API, WS endpoint currently open).

Use this document to execute full thorough testing and collect response logs.
