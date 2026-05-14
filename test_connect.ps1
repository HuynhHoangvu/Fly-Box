$loginJson = '{"email":"hhoangvu001@gmail.com","password":"12345678"}'
$loginResponse = Invoke-WebRequest -Uri 'http://localhost:8081/api/v1/auth/login' -Method POST -Body $loginJson -ContentType 'application/json' -UseBasicParsing

$token = ($loginResponse.Content | ConvertFrom-Json).token
Write-Host "Token: $token"

$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

$connectResponse = Invoke-WebRequest -Uri 'http://localhost:8081/api/v1/pages/connect' -Method POST -Body '{"platform":"facebook"}' -Headers $headers -UseBasicParsing
Write-Host "Connect Response: $($connectResponse.Content)"
