# Conversation History API Test Script (PowerShell)
# This script tests the new conversation history API endpoints

Write-Host "Testing Conversation History API..." -ForegroundColor Green
Write-Host "==================================" -ForegroundColor Green

$BASE_URL = "http://localhost:3001"
$JWT_TOKEN = "your-jwt-token-here"  # Replace with actual JWT token

# Test data
$CONVERSATION_ID = "conv-123"
$USER_ID = "user-123"

Write-Host "1. Testing Get Conversation History..." -ForegroundColor Yellow
Write-Host "-------------------------------------" -ForegroundColor Yellow

try {
    $response = Invoke-WebRequest -Uri "$BASE_URL/v1/chatbot/conversations/$CONVERSATION_ID/messages?limit=20&order=asc&include_attachments=true" -Method GET -Headers @{
        "Authorization" = "Bearer $JWT_TOKEN"
        "Content-Type" = "application/json"
    } -UseBasicParsing
    
    Write-Host "✓ Success (HTTP $($response.StatusCode))" -ForegroundColor Green
    Write-Host "Response: $($response.Content | ConvertFrom-Json | ConvertTo-Json -Depth 3)" -ForegroundColor White
} catch {
    Write-Host "✗ Failed: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "2. Testing Get Conversation History with Search..." -ForegroundColor Yellow
Write-Host "--------------------------------------------------" -ForegroundColor Yellow

try {
    $response = Invoke-WebRequest -Uri "$BASE_URL/v1/chatbot/conversations/$CONVERSATION_ID/messages?limit=10&search=finance&order=desc" -Method GET -Headers @{
        "Authorization" = "Bearer $JWT_TOKEN"
        "Content-Type" = "application/json"
    } -UseBasicParsing
    
    Write-Host "✓ Success (HTTP $($response.StatusCode))" -ForegroundColor Green
    Write-Host "Response: $($response.Content | ConvertFrom-Json | ConvertTo-Json -Depth 3)" -ForegroundColor White
} catch {
    Write-Host "✗ Failed: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "3. Testing Get User Conversations..." -ForegroundColor Yellow
Write-Host "------------------------------------" -ForegroundColor Yellow

try {
    $response = Invoke-WebRequest -Uri "$BASE_URL/v1/chatbot/users/$USER_ID/conversations?limit=10&status=active&include_last_message=true" -Method GET -Headers @{
        "Authorization" = "Bearer $JWT_TOKEN"
        "Content-Type" = "application/json"
    } -UseBasicParsing
    
    Write-Host "✓ Success (HTTP $($response.StatusCode))" -ForegroundColor Green
    Write-Host "Response: $($response.Content | ConvertFrom-Json | ConvertTo-Json -Depth 3)" -ForegroundColor White
} catch {
    Write-Host "✗ Failed: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "4. Testing Get Conversation Summary..." -ForegroundColor Yellow
Write-Host "--------------------------------------" -ForegroundColor Yellow

try {
    $response = Invoke-WebRequest -Uri "$BASE_URL/v1/chatbot/conversations/$CONVERSATION_ID" -Method GET -Headers @{
        "Authorization" = "Bearer $JWT_TOKEN"
        "Content-Type" = "application/json"
    } -UseBasicParsing
    
    Write-Host "✓ Success (HTTP $($response.StatusCode))" -ForegroundColor Green
    Write-Host "Response: $($response.Content | ConvertFrom-Json | ConvertTo-Json -Depth 3)" -ForegroundColor White
} catch {
    Write-Host "✗ Failed: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "5. Testing Error Cases..." -ForegroundColor Yellow
Write-Host "-------------------------" -ForegroundColor Yellow

Write-Host "Testing without authentication:" -ForegroundColor Cyan
try {
    $response = Invoke-WebRequest -Uri "$BASE_URL/v1/chatbot/conversations/$CONVERSATION_ID/messages" -Method GET -Headers @{
        "Content-Type" = "application/json"
    } -UseBasicParsing
    
    Write-Host "✓ Success (HTTP $($response.StatusCode))" -ForegroundColor Green
} catch {
    Write-Host "✗ Expected failure: $($_.Exception.Message)" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Testing with invalid conversation ID:" -ForegroundColor Cyan
try {
    $response = Invoke-WebRequest -Uri "$BASE_URL/v1/chatbot/conversations/invalid-id/messages" -Method GET -Headers @{
        "Authorization" = "Bearer $JWT_TOKEN"
        "Content-Type" = "application/json"
    } -UseBasicParsing
    
    Write-Host "✓ Success (HTTP $($response.StatusCode))" -ForegroundColor Green
} catch {
    Write-Host "✗ Expected failure: $($_.Exception.Message)" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Testing with invalid limit:" -ForegroundColor Cyan
try {
    $response = Invoke-WebRequest -Uri "$BASE_URL/v1/chatbot/conversations/$CONVERSATION_ID/messages?limit=200" -Method GET -Headers @{
        "Authorization" = "Bearer $JWT_TOKEN"
        "Content-Type" = "application/json"
    } -UseBasicParsing
    
    Write-Host "✓ Success (HTTP $($response.StatusCode))" -ForegroundColor Green
} catch {
    Write-Host "✗ Expected failure: $($_.Exception.Message)" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Conversation History API Test Complete!" -ForegroundColor Green
Write-Host "=======================================" -ForegroundColor Green
Write-Host ""
Write-Host "Expected Results:" -ForegroundColor White
Write-Host "- Authenticated requests should return 200 OK" -ForegroundColor White
Write-Host "- Unauthenticated requests should return 401 Unauthorized" -ForegroundColor White
Write-Host "- Invalid conversation IDs should return 404 Not Found" -ForegroundColor White
Write-Host "- Invalid parameters should return 400 Bad Request" -ForegroundColor White
Write-Host ""
Write-Host "Note: Replace JWT_TOKEN with a valid token from authentication endpoint" -ForegroundColor Cyan
