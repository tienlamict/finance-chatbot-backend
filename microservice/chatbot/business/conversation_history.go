// finance-chatbot-backend/microservice/chatbot/business/conversation_history.go
package business

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"finance-chatbot/addon/core"
	"finance-chatbot/microservice/chatbot/entity"
)

// GetConversationHistory retrieves conversation history with advanced filtering and access control
func (uc *chatUsecase) GetConversationHistory(ctx context.Context, req entity.ConversationHistoryRequest, userID string) (*entity.ConversationHistoryResponse, error) {
	// Validate request
	if err := uc.validateConversationHistoryRequest(req); err != nil {
		return nil, core.ErrBadRequest.WithError(err.Error())
	}

	// Check conversation access
	//hasAccess, err := uc.sql.CheckConversationAccess(ctx, req.ConversationID, userID)
	_, err := uc.sql.CheckConversationAccess(ctx, req.ConversationID, userID)
	if err != nil {
		return nil, core.ErrInternalServerError.WithDebug(err.Error())
	}
	// if !hasAccess {
	// 	return nil, core.ErrForbidden.WithError("access denied to conversation")
	// }

	// Get conversation info
	_, err = uc.sql.FindByID(ctx, req.ConversationID)
	if err != nil {
		return nil, core.ErrNotFound.WithError("conversation not found")
	}

	// Get messages with filtering
	messages, err := uc.sql.GetConversationHistory(ctx, req)
	if err != nil {
		return nil, core.ErrInternalServerError.WithDebug(err.Error())
	}

	// Get total count for pagination
	total, err := uc.sql.CountConversationHistory(ctx, req)
	if err != nil {
		return nil, core.ErrInternalServerError.WithDebug(err.Error())
	}

	// Build response with pagination cursors
	response := &entity.ConversationHistoryResponse{
		ConversationID: req.ConversationID,
		Total:          total,
		Items:          messages,
		HasMore:        len(messages) == req.Limit,
	}

	// Set pagination cursors
	if len(messages) > 0 {
		if req.Order == "desc" {
			// For descending order, next cursor is the oldest message
			response.NextCursor = &messages[len(messages)-1].ID
			response.PreviousCursor = &messages[0].ID
		} else {
			// For ascending order, next cursor is the newest message
			response.NextCursor = &messages[len(messages)-1].ID
			response.PreviousCursor = &messages[0].ID
		}
	}

	return response, nil
}

// GetUserConversations retrieves conversations for a user with filtering
func (uc *chatUsecase) GetUserConversations(ctx context.Context, req entity.UserConversationsRequest, requestingUserID string) (*entity.UserConversationsResponse, error) {
	// Validate request
	if err := uc.validateUserConversationsRequest(req); err != nil {
		return nil, core.ErrBadRequest.WithError(err.Error())
	}

	// Check if user can access conversations (self or admin)
	uid, _ := core.FromBase58(req.UserID)

	struserID := strconv.Itoa(int(uid.GetLocalID()))
	if struserID != requestingUserID {
		fmt.Println("Requesting user ID:", requestingUserID)
		fmt.Println("Target user ID:", struserID)
		// TODO: Add admin/superadmin check here
		// For now, only allow self-access
		return nil, core.ErrForbidden.WithError("access denied to user conversations")
	}

	// Get conversations with filtering
	conversations, err := uc.sql.GetUserConversations(ctx, req)
	if err != nil {
		return nil, core.ErrInternalServerError.WithDebug(err.Error())
	}

	// Get total count
	total, err := uc.sql.CountUserConversations(ctx, req)
	if err != nil {
		return nil, core.ErrInternalServerError.WithDebug(err.Error())
	}

	// Build response
	response := &entity.UserConversationsResponse{
		UserID:     req.UserID,
		Total:      total,
		Items:      conversations,
		HasMore:    len(conversations) == req.Limit,
		NextOffset: req.Offset + len(conversations),
	}

	return response, nil
}

// GetConversationSummary retrieves a conversation with its summary information
func (uc *chatUsecase) GetConversationSummary(ctx context.Context, conversationID, userID string) (*entity.ConversationSummary, error) {
	// Check conversation access
	hasAccess, err := uc.sql.CheckConversationAccess(ctx, conversationID, userID)
	if err != nil {
		return nil, core.ErrInternalServerError.WithDebug(err.Error())
	}
	if !hasAccess {
		return nil, core.ErrForbidden.WithError("access denied to conversation")
	}

	// Get conversation with last message
	summary, err := uc.sql.GetConversationWithLastMessage(ctx, conversationID)
	if err != nil {
		return nil, core.ErrNotFound.WithError("conversation not found")
	}

	return summary, nil
}

// Validate conversation history request
func (uc *chatUsecase) validateConversationHistoryRequest(req entity.ConversationHistoryRequest) error {
	if req.ConversationID == "" {
		return fmt.Errorf("conversation_id is required")
	}

	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		return fmt.Errorf("limit cannot exceed 100")
	}

	if req.Order != "" && req.Order != "asc" && req.Order != "desc" {
		return fmt.Errorf("order must be 'asc' or 'desc'")
	}

	// Validate cursor parameters
	if req.Before != nil && req.After != nil {
		return fmt.Errorf("cannot specify both 'before' and 'after' parameters")
	}

	return nil
}

// Validate user conversations request
func (uc *chatUsecase) validateUserConversationsRequest(req entity.UserConversationsRequest) error {
	if req.UserID == "" {
		return fmt.Errorf("user_id is required")
	}

	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		return fmt.Errorf("limit cannot exceed 100")
	}

	if req.Offset < 0 {
		req.Offset = 0
	}

	if req.Status != nil {
		validStatuses := []string{"active", "archived", "deleted"}
		statusValid := false
		for _, validStatus := range validStatuses {
			if *req.Status == validStatus {
				statusValid = true
				break
			}
		}
		if !statusValid {
			return fmt.Errorf("status must be one of: %s", strings.Join(validStatuses, ", "))
		}
	}

	return nil
}

// Helper function to parse cursor (message ID or timestamp)
func (uc *chatUsecase) parseCursor(cursor string) (messageID string, timestamp string, err error) {
	// Try to parse as timestamp first (ISO 8601 format)
	if _, parseErr := strconv.ParseInt(cursor, 10, 64); parseErr == nil {
		// It's a timestamp
		return "", cursor, nil
	}

	// Assume it's a message ID
	return cursor, "", nil
}
