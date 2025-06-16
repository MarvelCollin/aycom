@echo off
echo Testing the new /api/v1/messages/{messageId} route...
echo.

REM Test with a temp message ID (should handle gracefully)
echo Testing temp message unsend...
curl -X DELETE "http://localhost:8083/api/v1/messages/temp_123?chat_id=test-chat" ^
  -H "Authorization: Bearer test-token" ^
  -H "Content-Type: application/json" ^
  -w "\nHTTP Status: %%{http_code}\n" ^
  -s

echo.
echo.
echo Testing non-existent message...
curl -X DELETE "http://localhost:8083/api/v1/messages/nonexistent-123?chat_id=test-chat" ^
  -H "Authorization: Bearer test-token" ^
  -H "Content-Type: application/json" ^
  -w "\nHTTP Status: %%{http_code}\n" ^
  -s

echo.
echo Test completed. Check the responses above.
pause
