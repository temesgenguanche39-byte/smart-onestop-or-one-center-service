@echo off
echo ===================================================================
echo   Smart One-Stop Municipal Platform - Vercel to Docker Bridge
echo ===================================================================
echo.
echo Starting secure HTTPS tunnel to your local Docker system (port 8080)...
echo.
echo Make sure your Docker container is running:
echo   docker compose up -d
echo.
echo -------------------------------------------------------------------
echo Connect your Vercel Citizen Portal:
echo 1. Copy the https://xxxx.loca.lt URL shown below.
echo 2. Open your Vercel deployment with ?api=https://xxxx.loca.lt
echo    OR click the "Addis Ababa Municipal Cloud" badge to paste it!
echo -------------------------------------------------------------------
echo.
npx -y localtunnel --port 8080
