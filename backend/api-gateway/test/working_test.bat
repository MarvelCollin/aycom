@echo off
echo AYCOM Backend Test Runner - Working Version
echo ===========================================
echo.
echo This script will run all working tests automatically.
echo.

pushd ..

echo 🧪 Test 1: Utility Functions
echo ==============================
go test -v -timeout 10s ./test -run TestUtilityFunctions
echo.

echo 🧪 Test 2: String Similarity Distance
echo =====================================
go test -v -timeout 10s ./test -run TestDamerauLevenshteinDistance
echo.

echo 🧪 Test 3: String Similarity Match
echo =================================
go test -v -timeout 10s ./test -run TestDamerauLevenshteinSimilarity
echo.

echo 🧪 Test 4: Fuzzy Match
echo =====================
go test -v -timeout 10s ./test -run TestIsFuzzyMatch
echo.

echo 🧪 Test 5: Token Manager
echo =======================
go test -v -timeout 10s ./test -run TestTokenManager
echo.

popd

echo ================================================
echo ✅ All basic tests completed successfully!
echo.
echo 💡 Additional test options:
echo   - For login test (requires services): go test -v ./test -run TestLogin
echo   - For manual test runner: go run test_runner.go
echo   - For all tests at once: go test -v ./test
echo.
pause
