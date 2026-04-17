@echo off
setlocal enabledelayedexpansion

:: --- 設定 ---
set POSTGRES_USER=user
set POSTGRES_DB=fitbit_db

:: --- コンテナIDの取得 ---
for /f "tokens=*" %%i in ('docker ps -qf "ancestor=postgres:15-alpine"') do set CONTAINER_ID=%%i

if "%CONTAINER_ID%"=="" (
    echo [ERROR] Postgres container is not running.
    pause
    exit /b
)

echo ==========================================
echo   🏃 Latest 5 Activities
echo ==========================================
docker exec -i %CONTAINER_ID% psql -U %POSTGRES_USER% -d %POSTGRES_DB% -c "SELECT date, steps, calories, sleep_minutes, updated_at FROM daily_activities ORDER BY date DESC LIMIT 5;"

echo.
echo ==========================================
echo   🔑 Token Status
echo ==========================================
docker exec -i %CONTAINER_ID% psql -U %POSTGRES_USER% -d %POSTGRES_DB% -c "SELECT id, expiry, (expiry > now()) as is_valid FROM fitbit_auths;"

echo.
pause
