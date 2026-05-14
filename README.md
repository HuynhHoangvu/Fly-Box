# Fly-Box

## Backend

### 1. Start PostgreSQL

```bash
docker-compose up -d
```

### 2. Run Backend

```bash
cd backend && go run cmd/api/main.go
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| APP_PORT | 8081 | Server port |
| DATABASE_URL | host=localhost user=postgres password=123456 dbname=flybox port=5436 sslmode=disable | PostgreSQL connection |
| JWT_SECRET | dev-secret | JWT signing secret |
| FRONTEND_URL | http://localhost:5173 | Frontend URL for CORS |
| FACEBOOK_APP_ID | - | Facebook App ID |
| FACEBOOK_APP_SECRET | - | Facebook App Secret |
| FACEBOOK_REDIRECT_URI | http://localhost:5173/connect/callback | Facebook callback URL |
| FACEBOOK_VERIFY_TOKEN | verify-token | Facebook verify token |
| TIKTOK_VERIFY_TOKEN | tiktok-verify-token | TikTok verify token |

## Frontend

```bash
cd frontend && npm run dev
```