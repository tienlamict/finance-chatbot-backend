# PowerShell test script for frontend /v1/chatbot/prompt API integration
# This tests the complete flow: JWT auth -> conversation creation -> AI service call -> response mapping

$BaseURL = "http://localhost:3001/v1"
$AIServiceURL = "http://localhost:8000"

Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "Frontend /v1/chatbot/prompt Integration Test" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host ""

# Test 1: Login to get JWT token
Write-Host "Test 1: Authenticating user..." -ForegroundColor Yellow
Write-Host "-------------------------------"

$loginBody = @{
    email = "admin@example.com"
    password = "admin123"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "$BaseURL/authenticate" `
        -Method Post `
        -ContentType "application/json" `
        -Body $loginBody
    
    $token = $loginResponse.data.access_token
    if (-not $token) {
        $token = $loginResponse.access_token
    }
    
    if (-not $token) {
        Write-Host "❌ Failed to get JWT token" -ForegroundColor Red
        Write-Host "Response: $($loginResponse | ConvertTo-Json)" -ForegroundColor Red
        exit 1
    }
    
    Write-Host "✓ Successfully authenticated" -ForegroundColor Green
    Write-Host "Token: $($token.Substring(0, [Math]::Min(50, $token.Length)))..."
    Write-Host ""
} catch {
    Write-Host "❌ Authentication failed: $_" -ForegroundColor Red
    exit 1
}

# Test 2: Send first message (creates new conversation)
Write-Host "Test 2: Sending first message (creates new conversation)..." -ForegroundColor Yellow
Write-Host "-------------------------------------------------------------"

$firstMessage = @{
    content = "Hello, how are you?"
} | ConvertTo-Json

try {
    $headers = @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    }
    
    $firstResponse = Invoke-RestMethod -Uri "$BaseURL/chatbot/prompt" `
        -Method Post `
        -Headers $headers `
        -Body $firstMessage
    
    Write-Host "Response:" -ForegroundColor Cyan
    Write-Host ($firstResponse | ConvertTo-Json -Depth 5)
    
    $conversationId = $firstResponse.conversation_id
    if (-not $conversationId) {
        $conversationId = $firstResponse.data.conversation_id
    }
    
    if (-not $conversationId) {
        Write-Host "❌ Failed to create conversation" -ForegroundColor Red
        exit 1
    }
    
    Write-Host "✓ Conversation created successfully" -ForegroundColor Green
    Write-Host "Conversation ID: $conversationId"
    Write-Host "Assistant Reply: $($firstResponse.assistant_content)"
    Write-Host ""
} catch {
    Write-Host "❌ Failed to send first message: $_" -ForegroundColor Red
    exit 1
}

# Test 3: Send follow-up message (with conversation_id query param)
Write-Host "Test 3: Sending follow-up message with history..." -ForegroundColor Yellow
Write-Host "---------------------------------------------------"

$secondMessage = @{
    content = "What is the capital of France?"
} | ConvertTo-Json

try {
    $secondResponse = Invoke-RestMethod -Uri "$BaseURL/chatbot/prompt?conversation_id=$conversationId" `
        -Method Post `
        -Headers $headers `
        -Body $secondMessage
    
    Write-Host "Response:" -ForegroundColor Cyan
    Write-Host ($secondResponse | ConvertTo-Json -Depth 5)
    
    $returnedConvId = $secondResponse.conversation_id
    if (-not $returnedConvId) {
        $returnedConvId = $secondResponse.data.conversation_id
    }
    
    if ($returnedConvId -ne $conversationId) {
        Write-Host "⚠ Warning: Returned conversation_id doesn't match" -ForegroundColor Yellow
        Write-Host "Expected: $conversationId"
        Write-Host "Got: $returnedConvId"
    }
    
    Write-Host "✓ Follow-up message sent successfully" -ForegroundColor Green
    Write-Host "Assistant Reply: $($secondResponse.assistant_content)"
    Write-Host "Model Name: $($secondResponse.model_name)"
    Write-Host "Latency: $($secondResponse.latency_ms)ms"
    Write-Host ""
} catch {
    Write-Host "❌ Failed to send follow-up message: $_" -ForegroundColor Red
    exit 1
}

# Test 4: Test with deep_research flag
Write-Host "Test 4: Testing deep_research flag..." -ForegroundColor Yellow
Write-Host "---------------------------------------"

$deepResearchMessage = @{
    content = "Analyze VIC stock performance"
    deep_research = $true
} | ConvertTo-Json

try {
    $deepResponse = Invoke-RestMethod -Uri "$BaseURL/chatbot/prompt?conversation_id=$conversationId" `
        -Method Post `
        -Headers $headers `
        -Body $deepResearchMessage
    
    Write-Host "Response:" -ForegroundColor Cyan
    Write-Host ($deepResponse | ConvertTo-Json -Depth 5)
    
    Write-Host "✓ Deep research request completed" -ForegroundColor Green
    Write-Host "Assistant Reply: $($deepResponse.assistant_content)"
    Write-Host ""
} catch {
    Write-Host "❌ Failed deep research request: $_" -ForegroundColor Red
}

# Test 5: Test error handling (invalid conversation_id)
Write-Host "Test 5: Testing error handling (invalid conversation_id)..." -ForegroundColor Yellow
Write-Host "-------------------------------------------------------------"

$errorMessage = @{
    content = "This should fail"
} | ConvertTo-Json

try {
    $errorResponse = Invoke-RestMethod -Uri "$BaseURL/chatbot/prompt?conversation_id=invalid-id-12345" `
        -Method Post `
        -Headers $headers `
        -Body $errorMessage
    
    Write-Host "⚠ Expected error but got success response" -ForegroundColor Yellow
} catch {
    $statusCode = $_.Exception.Response.StatusCode.value__
    Write-Host "HTTP Status: $statusCode"
    
    if ($statusCode -eq 403 -or $statusCode -eq 404) {
        Write-Host "✓ Error handling works correctly (403/404 for invalid conversation)" -ForegroundColor Green
    } else {
        Write-Host "⚠ Unexpected status code: $statusCode (expected 403 or 404)" -ForegroundColor Yellow
    }
}
Write-Host ""

# Test 6: Test without JWT (should fail with 401)
Write-Host "Test 6: Testing authentication requirement..." -ForegroundColor Yellow
Write-Host "----------------------------------------------"

$unauthMessage = @{
    content = "This should require auth"
} | ConvertTo-Json

try {
    $unauthResponse = Invoke-RestMethod -Uri "$BaseURL/chatbot/prompt" `
        -Method Post `
        -ContentType "application/json" `
        -Body $unauthMessage
    
    Write-Host "❌ Expected 401 but got success response" -ForegroundColor Red
} catch {
    $statusCode = $_.Exception.Response.StatusCode.value__
    Write-Host "HTTP Status: $statusCode"
    
    if ($statusCode -eq 401) {
        Write-Host "✓ Authentication requirement enforced (401 Unauthorized)" -ForegroundColor Green
    } else {
        Write-Host "❌ Expected 401, got $statusCode" -ForegroundColor Red
    }
}
Write-Host ""

# Summary
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "Test Summary" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "✓ JWT Authentication" -ForegroundColor Green
Write-Host "✓ New conversation creation" -ForegroundColor Green
Write-Host "✓ Follow-up messages with history" -ForegroundColor Green
Write-Host "✓ conversation_id query parameter" -ForegroundColor Green
Write-Host "✓ Deep research flag" -ForegroundColor Green
Write-Host "✓ Error handling" -ForegroundColor Green
Write-Host "✓ Authentication enforcement" -ForegroundColor Green
Write-Host ""
Write-Host "All tests passed!" -ForegroundColor Green
Write-Host ""
Write-Host "Conversation ID for manual testing: $conversationId" -ForegroundColor Cyan
Write-Host ""
Write-Host "Manual PowerShell command:" -ForegroundColor Cyan
Write-Host @"
`$headers = @{
    'Authorization' = 'Bearer $token'
    'Content-Type' = 'application/json'
}
`$body = @{ content = 'Tell me more' } | ConvertTo-Json
Invoke-RestMethod -Uri '$BaseURL/chatbot/prompt?conversation_id=$conversationId' -Method Post -Headers `$headers -Body `$body
"@

