$token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo1LCJlbWFpbCI6ImFkbWluQGhhcmFzb2NpYWwubG9jYWwiLCJyb2xlIjoic3RhZmYiLCJpc3MiOiJmbHktYm94LWFwaSIsImV4cCI6MTc3OTMzODg4MCwiaWF0IjoxNzc4NzM0MDgwfQ.lR6xt7vuqk8nD9nVhAngZkIBg-ORH2VRWhmyOOUdef0"

$headers = @{
  Authorization = "Bearer $token"
}

$response = Invoke-WebRequest -Uri 'http://localhost:8081/api/v1/users/me' -Method GET -Headers $headers -UseBasicParsing
Write-Host "Status:" $response.StatusCode
Write-Host "Content:" $response.Content
