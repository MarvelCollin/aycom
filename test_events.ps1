Write-Host "🧪 Testing RabbitMQ Event Publishing" -ForegroundColor Yellow
Write-Host "======================================" -ForegroundColor Yellow

Write-Host "`n1. Testing Thread Like Event" -ForegroundColor Cyan
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8083/api/v1/threads/test-thread-123/like" `
        -Method POST `
        -Headers @{"Authorization"="Bearer test-token"; "Content-Type"="application/json"} `
        -ErrorAction Stop
    Write-Host "✅ Status: $($response.StatusCode)" -ForegroundColor Green
    Write-Host "Response: $($response.Content)"
} catch {
    Write-Host "❌ Error: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`n2. Testing Thread Bookmark Event" -ForegroundColor Cyan
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8083/api/v1/threads/test-thread-123/bookmark" `
        -Method POST `
        -Headers @{"Authorization"="Bearer test-token"; "Content-Type"="application/json"} `
        -ErrorAction Stop
    Write-Host "✅ Status: $($response.StatusCode)" -ForegroundColor Green
    Write-Host "Response: $($response.Content)"
} catch {
    Write-Host "❌ Error: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`n3. Testing User Follow Event" -ForegroundColor Cyan
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8083/api/v1/users/test-user-456/follow" `
        -Method POST `
        -Headers @{"Authorization"="Bearer test-token"; "Content-Type"="application/json"} `
        -ErrorAction Stop
    Write-Host "✅ Status: $($response.StatusCode)" -ForegroundColor Green
    Write-Host "Response: $($response.Content)"
} catch {
    Write-Host "❌ Error: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`n✅ Test complete! Check event bus logs for events:" -ForegroundColor Green
Write-Host "docker-compose logs event_bus --tail=20" -ForegroundColor White
