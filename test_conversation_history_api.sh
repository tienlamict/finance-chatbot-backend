#!/bin/bash

# Conversation History API Test Script
# This script tests the new conversation history API endpoints

echo "Testing Conversation History API..."
echo "=================================="

BASE_URL="http://localhost:3001"
JWT_TOKEN="your-jwt-token-here"  # Replace with actual JWT token

# Test data
CONVERSATION_ID="conv-123"
USER_ID="user-123"

echo "1. Testing Get Conversation History..."
echo "-------------------------------------"

curl -X GET \
  "$BASE_URL/v1/chatbot/conversations/$CONVERSATION_ID/messages?limit=20&order=asc&include_attachments=true" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -w "\nHTTP Status: %{http_code}\n" \
  -s

echo ""
echo "2. Testing Get Conversation History with Search..."
echo "--------------------------------------------------"

curl -X GET \
  "$BASE_URL/v1/chatbot/conversations/$CONVERSATION_ID/messages?limit=10&search=finance&order=desc" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -w "\nHTTP Status: %{http_code}\n" \
  -s

echo ""
echo "3. Testing Get User Conversations..."
echo "------------------------------------"

curl -X GET \
  "$BASE_URL/v1/chatbot/users/$USER_ID/conversations?limit=10&status=active&include_last_message=true" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -w "\nHTTP Status: %{http_code}\n" \
  -s

echo ""
echo "4. Testing Get Conversation Summary..."
echo "--------------------------------------"

curl -X GET \
  "$BASE_URL/v1/chatbot/conversations/$CONVERSATION_ID" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -w "\nHTTP Status: %{http_code}\n" \
  -s

echo ""
echo "5. Testing Error Cases..."
echo "-------------------------"

echo "Testing without authentication:"
curl -X GET \
  "$BASE_URL/v1/chatbot/conversations/$CONVERSATION_ID/messages" \
  -H "Content-Type: application/json" \
  -w "\nHTTP Status: %{http_code}\n" \
  -s

echo ""
echo "Testing with invalid conversation ID:"
curl -X GET \
  "$BASE_URL/v1/chatbot/conversations/invalid-id/messages" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -w "\nHTTP Status: %{http_code}\n" \
  -s

echo ""
echo "Testing with invalid limit:"
curl -X GET \
  "$BASE_URL/v1/chatbot/conversations/$CONVERSATION_ID/messages?limit=200" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -w "\nHTTP Status: %{http_code}\n" \
  -s

echo ""
echo "Conversation History API Test Complete!"
echo "======================================="
echo ""
echo "Expected Results:"
echo "- Authenticated requests should return 200 OK"
echo "- Unauthenticated requests should return 401 Unauthorized"
echo "- Invalid conversation IDs should return 404 Not Found"
echo "- Invalid parameters should return 400 Bad Request"
echo ""
echo "Note: Replace JWT_TOKEN with a valid token from authentication endpoint"
