$json = '{"email":"test@test.com","password":"test123"}'
Invoke-WebRequest -Uri 'http://localhost:8081/api/v1/auth/login' -Method POST -Body $json -ContentType 'application/json' -UseBasicParsing
