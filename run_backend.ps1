[Environment]::SetEnvironmentVariable("NGROK_URL", "https://clubbed-devotedly-pedometer.ngrok-free.dev", "Process")
[Environment]::SetEnvironmentVariable("FACEBOOK_APP_ID", "2323107644886195", "Process")
[Environment]::SetEnvironmentVariable("FACEBOOK_APP_SECRET", "6e7282c73fb6f63d97d2e30eb2f5e7bd", "Process")
[Environment]::SetEnvironmentVariable("FACEBOOK_VERIFY_TOKEN", "verify-token", "Process")
[Environment]::SetEnvironmentVariable("FACEBOOK_REDIRECT_URI", "https://clubbed-devotedly-pedometer.ngrok-free.dev/api/v1/pages/connect/callback", "Process")
[Environment]::SetEnvironmentVariable("FRONTEND_URL", "https://clubbed-devotedly-pedometer.ngrok-free.dev", "Process")
Start-Process -FilePath "d:\Fly_Visa\Fly-Box\Fly-Box\backend\tmp_bin.exe" -WorkingDirectory "d:\Fly_Visa\Fly-Box\Fly-Box\backend"
