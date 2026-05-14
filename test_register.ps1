$body = @"
{
  "name": "Admin",
  "email": "admin@harasocial.local",
  "password": "admin123"
}
"@
$response = Invoke-WebRequest -Uri 'http://localhost:8081/api/v1/auth/register' -Method POST -Body $body -ContentType 'application/json' -UseBasicParsing
Write-Host "Status:" $response.StatusCode
Write-Host "Content:" $response.Content
