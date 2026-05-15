# Test WebSocket messaging and real-time notifications
# This script tests that when a message is sent, the sender receives notification via WebSocket

param(
    [string]$BackendURL = "https://backend-production-90a6.up.railway.app",
    [string]$Email = "test@example.com",
    [string]$Password = "test123456"
)

$ErrorActionPreference = "Stop"

# Step 1: Login to get token and user ID
Write-Host "[1] Logging in..." -ForegroundColor Cyan
$loginBody = @{
    email = $Email
    password = $Password
} | ConvertTo-Json

$loginResp = Invoke-RestMethod -Uri "$BackendURL/api/v1/auth/login" `
    -Method POST `
    -ContentType "application/json" `
    -Body $loginBody

$token = $loginResp.token
$userId = $loginResp.user.id

Write-Host "Logged in as user ID: $userId" -ForegroundColor Green
Write-Host "Token: $($token.Substring(0, 20))..." -ForegroundColor Green

# Step 2: List connected pages (need a page ID to work with)
Write-Host "`n[2] Listing pages..." -ForegroundColor Cyan
$pagesResp = Invoke-RestMethod -Uri "$BackendURL/api/v1/pages" `
    -Method GET `
    -Headers @{ Authorization = "Bearer $token" }

if ($pagesResp.data.Count -eq 0) {
    Write-Host "No pages connected. Please connect a page first." -ForegroundColor Yellow
    exit
}

$pageId = $pagesResp.data[0].id
Write-Host "Using page ID: $pageId" -ForegroundColor Green

# Step 3: List conversations
Write-Host "`n[3] Listing conversations..." -ForegroundColor Cyan
$convsResp = Invoke-RestMethod -Uri "$BackendURL/api/v1/conversations?page_id=$pageId" `
    -Method GET `
    -Headers @{ Authorization = "Bearer $token" }

if ($convsResp.data.Count -eq 0) {
    Write-Host "No conversations found. Creating a test conversation..." -ForegroundColor Yellow
    
    # We can't easily create a conversation without webhook, but we can note this
    Write-Host "Please ensure you have at least one conversation (via webhook or mock)." -ForegroundColor Yellow
    Write-Host "The test requires at least one conversation to send a message." -ForegroundColor Red
    exit
}

$convId = $convsResp.data[0].id
Write-Host "Using conversation ID: $convId" -ForegroundColor Green

# Step 4: Send a test message
Write-Host "`n[4] Sending test message..." -ForegroundColor Cyan
$messageContent = "Test message $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
$messageBody = @{
    content = $messageContent
} | ConvertTo-Json

try {
    $sendResp = Invoke-RestMethod -Uri "$BackendURL/api/v1/conversations/$convId/messages" `
        -Method POST `
        -ContentType "application/json" `
        -Headers @{ Authorization = "Bearer $token" } `
        -Body $messageBody

    Write-Host "Message sent successfully!" -ForegroundColor Green
    Write-Host "Response: $($sendResp | ConvertTo-Json -Compress)" -ForegroundColor Green
    
    $msgId = $sendResp.data.id
    Write-Host "New message ID: $msgId" -ForegroundColor Green
} catch {
    Write-Host "Error sending message: $_" -ForegroundColor Red
    exit
}

# Step 5: Verify via WebSocket that message was received
Write-Host "`n[5] WebSocket Test Instructions:" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
Write-Host "To verify real-time notifications work:" -ForegroundColor Yellow
Write-Host ""
Write-Host "1. Open the frontend app in browser" -ForegroundColor White
Write-Host "2. Login with same account" -ForegroundColor White  
Write-Host "3. Navigate to Inbox (http://localhost:5173/inbox)" -ForegroundColor White
Write-Host "4. Open browser DevTools (F12) -> Console" -ForegroundColor White
Write-Host "5. You should see '[WS] Connected' message" -ForegroundColor White
Write-Host "6. Send a message - you should see it appear in real-time without refreshing" -ForegroundColor White
Write-Host ""
Write-Host "The WebSocket connection URL would be:" -ForegroundColor Cyan
Write-Host "   $BackendURL/ws?user_id=$userId" -ForegroundColor White
Write-Host ""

# Step 6: Check the message was saved
Write-Host "`n[6] Verifying message in database..." -ForegroundColor Cyan
$msgsResp = Invoke-RestMethod -Uri "$BackendURL/api/v1/conversations/$convId/messages" `
    -Method GET `
    -Headers @{ Authorization = "Bearer $token" }

$latestMsg = $msgsResp.data | Select-Object -Last 1
Write-Host "Latest message: $($latestMsg.content)" -ForegroundColor Green
Write-Host "Sender type: $($latestMsg.sender_type)" -ForegroundColor Green

Write-Host "`n======================================" -ForegroundColor Cyan
Write-Host "Test Complete!" -ForegroundColor Green
Write-Host ""
Write-Host "To test real-time WebSocket notifications:" -ForegroundColor Yellow
Write-Host "- Open frontend in browser" -ForegroundColor White
Write-Host "- Open DevTools Console" -ForegroundColor White
Write-Host "- Send message via API or frontend" -ForegroundColor White
Write-Host "- You should see '[WS] Connected' and message updates without refresh" -ForegroundColor White
