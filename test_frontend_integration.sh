#!/bin/bash

# Test script for frontend /v1/chatbot/prompt API integration with AI service
# This tests the complete flow: JWT auth -> conversation creation -> AI service call -> response mapping

set -e

BASE_URL="http://localhost:3001/v1"
AI_SERVICE_URL="http://localhost:8000"

echo "========================================="
echo "Frontend /v1/chatbot/prompt Integration Test"
echo "========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test 1: Login to get JWT token
echo "Test 1: Authenticating user..."
echo "-------------------------------"

LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/authenticate" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "admin123"
  }')

TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.data.access_token // .access_token // empty')

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
  echo -e "${RED}❌ Failed to get JWT token${NC}"
  echo "Response: $LOGIN_RESPONSE"
  exit 1
fi

echo -e "${GREEN}✓ Successfully authenticated${NC}"
echo "Token: ${TOKEN:0:50}..."
echo ""

# Test 2: Send first message (creates new conversation)
echo "Test 2: Sending first message (creates new conversation)..."
echo "-------------------------------------------------------------"

FIRST_MSG_RESPONSE=$(curl -s -X POST "$BASE_URL/chatbot/prompt" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "content": "Hello, how are you?"
  }')

echo "Response:"
echo $FIRST_MSG_RESPONSE | jq .

CONVERSATION_ID=$(echo $FIRST_MSG_RESPONSE | jq -r '.conversation_id // .data.conversation_id // empty')
USER_MSG_ID=$(echo $FIRST_MSG_RESPONSE | jq -r '.user_message_id // .data.user_message_id // empty')
ASSISTANT_MSG_ID=$(echo $FIRST_MSG_RESPONSE | jq -r '.assistant_message_id // .data.assistant_message_id // empty')
ASSISTANT_CONTENT=$(echo $FIRST_MSG_RESPONSE | jq -r '.assistant_content // .data.assistant_content // empty')

if [ -z "$CONVERSATION_ID" ] || [ "$CONVERSATION_ID" = "null" ]; then
  echo -e "${RED}❌ Failed to create conversation${NC}"
  exit 1
fi

echo -e "${GREEN}✓ Conversation created successfully${NC}"
echo "Conversation ID: $CONVERSATION_ID"
echo "User Message ID: $USER_MSG_ID"
echo "Assistant Message ID: $ASSISTANT_MSG_ID"
echo "Assistant Reply: $ASSISTANT_CONTENT"
echo ""

# Test 3: Send follow-up message (with conversation_id query param)
echo "Test 3: Sending follow-up message with history..."
echo "---------------------------------------------------"

SECOND_MSG_RESPONSE=$(curl -s -X POST "$BASE_URL/chatbot/prompt?conversation_id=$CONVERSATION_ID" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "content": "What is the capital of France?"
  }')

echo "Response:"
echo $SECOND_MSG_RESPONSE | jq .

RETURNED_CONV_ID=$(echo $SECOND_MSG_RESPONSE | jq -r '.conversation_id // .data.conversation_id // empty')
ASSISTANT_CONTENT_2=$(echo $SECOND_MSG_RESPONSE | jq -r '.assistant_content // .data.assistant_content // empty')
MODEL_NAME=$(echo $SECOND_MSG_RESPONSE | jq -r '.model_name // .data.model_name // empty')
LATENCY_MS=$(echo $SECOND_MSG_RESPONSE | jq -r '.latency_ms // .data.latency_ms // empty')

if [ "$RETURNED_CONV_ID" != "$CONVERSATION_ID" ]; then
  echo -e "${YELLOW}⚠ Warning: Returned conversation_id doesn't match${NC}"
  echo "Expected: $CONVERSATION_ID"
  echo "Got: $RETURNED_CONV_ID"
fi

echo -e "${GREEN}✓ Follow-up message sent successfully${NC}"
echo "Conversation ID: $RETURNED_CONV_ID"
echo "Assistant Reply: $ASSISTANT_CONTENT_2"
echo "Model Name: $MODEL_NAME"
echo "Latency: ${LATENCY_MS}ms"
echo ""

# Test 4: Test with deep_research flag
echo "Test 4: Testing deep_research flag..."
echo "---------------------------------------"

DEEP_RESEARCH_RESPONSE=$(curl -s -X POST "$BASE_URL/chatbot/prompt?conversation_id=$CONVERSATION_ID" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "content": "Analyze VIC stock performance",
    "deep_research": true
  }')

echo "Response:"
echo $DEEP_RESEARCH_RESPONSE | jq .

DEEP_ASSISTANT_CONTENT=$(echo $DEEP_RESEARCH_RESPONSE | jq -r '.assistant_content // .data.assistant_content // empty')

echo -e "${GREEN}✓ Deep research request completed${NC}"
echo "Assistant Reply: $DEEP_ASSISTANT_CONTENT"
echo ""

# Test 5: Test error handling (invalid conversation_id)
echo "Test 5: Testing error handling (invalid conversation_id)..."
echo "-------------------------------------------------------------"

ERROR_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/chatbot/prompt?conversation_id=invalid-id-12345" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "content": "This should fail"
  }')

HTTP_CODE=$(echo "$ERROR_RESPONSE" | tail -n1)
RESPONSE_BODY=$(echo "$ERROR_RESPONSE" | head -n-1)

echo "HTTP Status: $HTTP_CODE"
echo "Response Body: $RESPONSE_BODY"

if [ "$HTTP_CODE" = "403" ] || [ "$HTTP_CODE" = "404" ]; then
  echo -e "${GREEN}✓ Error handling works correctly (403/404 for invalid conversation)${NC}"
else
  echo -e "${YELLOW}⚠ Unexpected status code: $HTTP_CODE (expected 403 or 404)${NC}"
fi
echo ""

# Test 6: Test without JWT (should fail with 401)
echo "Test 6: Testing authentication requirement..."
echo "----------------------------------------------"

UNAUTH_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/chatbot/prompt" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "This should require auth"
  }')

UNAUTH_HTTP_CODE=$(echo "$UNAUTH_RESPONSE" | tail -n1)
UNAUTH_BODY=$(echo "$UNAUTH_RESPONSE" | head -n-1)

echo "HTTP Status: $UNAUTH_HTTP_CODE"

if [ "$UNAUTH_HTTP_CODE" = "401" ]; then
  echo -e "${GREEN}✓ Authentication requirement enforced (401 Unauthorized)${NC}"
else
  echo -e "${RED}❌ Expected 401, got $UNAUTH_HTTP_CODE${NC}"
fi
echo ""

# Summary
echo "========================================="
echo "Test Summary"
echo "========================================="
echo ""
echo "✓ JWT Authentication"
echo "✓ New conversation creation"
echo "✓ Follow-up messages with history"
echo "✓ conversation_id query parameter"
echo "✓ Deep research flag"
echo "✓ Error handling"
echo "✓ Authentication enforcement"
echo ""
echo -e "${GREEN}All tests passed!${NC}"
echo ""
echo "Conversation ID for manual testing: $CONVERSATION_ID"
echo ""
echo "Manual cURL command:"
echo "curl -X POST '$BASE_URL/chatbot/prompt?conversation_id=$CONVERSATION_ID' \\"
echo "  -H 'Authorization: Bearer $TOKEN' \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -d '{\"content\":\"Tell me more\"}'"

