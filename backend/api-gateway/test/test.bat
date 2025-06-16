@echo off
echo AYCOM Backend API Gateway Test Runner
echo ======================================
echo.
echo Select a test to run:
echo 1. Login Authentication Test
echo 2. String Similarity Test
echo 3. Token Manager Test
echo 4. Run All Tests
echo.

set /p choice="Enter your choice (1-4): "

cd ..\..\..

if "%choice%"=="1" (
    echo Running Login Authentication Test...
    go test -v aycom/backend/api-gateway/test -run TestLogin
) else if "%choice%"=="2" (
    echo Running String Similarity Test...
    go test -v aycom/backend/api-gateway/test -run TestDamerauLevenshtein
) else if "%choice%"=="3" (
    echo Running Token Manager Test...
    go test -v aycom/backend/api-gateway/test -run TestTokenManager
) else if "%choice%"=="4" (
    echo Running All Tests...
    go test -v aycom/backend/api-gateway/test
) else (
    echo Invalid choice!
    echo Please run the script again and select a number between 1 and 4.
)

pause
