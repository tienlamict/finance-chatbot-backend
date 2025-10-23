package mysql

import (
	"context"
	"finance-chatbot/microservice/chatbot/entity"
)

type ConversationStore interface {
	Create(ctx context.Context, conv *entity.Conversation) error
	FindByID(ctx context.Context, id string) (*entity.Conversation, error)

	// Enhanced conversation methods
	GetUserConversations(ctx context.Context, req entity.UserConversationsRequest) ([]entity.ConversationSummary, error)
	CountUserConversations(ctx context.Context, req entity.UserConversationsRequest) (int64, error)
	GetConversationWithLastMessage(ctx context.Context, conversationID string) (*entity.ConversationSummary, error)
	CheckConversationAccess(ctx context.Context, conversationID, userID string) (bool, error)
}

func (r *MySQLRepo) Create(ctx context.Context, conv *entity.Conversation) error {
	return r.db.WithContext(ctx).Create(conv).Error
}

func (r *MySQLRepo) FindByID(ctx context.Context, id string) (*entity.Conversation, error) {
	var conv entity.Conversation
	if err := r.db.WithContext(ctx).First(&conv, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &conv, nil
}

// GetUserConversations retrieves conversations for a user with filtering
func (r *MySQLRepo) GetUserConversations(ctx context.Context, req entity.UserConversationsRequest) ([]entity.ConversationSummary, error) {
	var conversations []entity.Conversation
	q := r.db.WithContext(ctx).Where("user_id = ?", req.UserID)

	// Apply status filter
	if req.Status != nil {
		q = q.Where("status = ?", *req.Status)
	}

	// Apply time range filters
	if req.From != nil {
		q = q.Where("created_at >= ?", *req.From)
	}
	if req.To != nil {
		q = q.Where("created_at <= ?", *req.To)
	}

	// Apply search filter
	if req.Search != nil && *req.Search != "" {
		q = q.Where("title LIKE ?", "%"+*req.Search+"%")
	}

	// Apply ordering and pagination
	q = q.Order("updated_at DESC")
	if req.Limit > 0 {
		q = q.Limit(req.Limit).Offset(req.Offset)
	}

	if err := q.Find(&conversations).Error; err != nil {
		return nil, err
	}

	// Convert to ConversationSummary
	result := make([]entity.ConversationSummary, len(conversations))
	for i, conv := range conversations {
		result[i] = entity.ConversationSummary{
			Conversation: conv,
		}

		// Load last message if requested
		if req.IncludeLastMessage {
			lastMessage, err := r.getLastMessageForConversation(ctx, conv.ID)
			if err == nil {
				result[i].LastMessage = lastMessage
			}
		}

		// Load message counts
		totalMessages, err := r.countMessagesForConversation(ctx, conv.ID)
		if err == nil {
			result[i].TotalMessages = totalMessages
		}

		// TODO: Implement unread count logic if needed
		result[i].UnreadCount = 0
	}

	return result, nil
}

// CountUserConversations counts conversations matching the filter criteria
func (r *MySQLRepo) CountUserConversations(ctx context.Context, req entity.UserConversationsRequest) (int64, error) {
	var count int64
	q := r.db.WithContext(ctx).Model(&entity.Conversation{}).Where("user_id = ?", req.UserID)

	// Apply same filters as GetUserConversations
	if req.Status != nil {
		q = q.Where("status = ?", *req.Status)
	}
	if req.From != nil {
		q = q.Where("created_at >= ?", *req.From)
	}
	if req.To != nil {
		q = q.Where("created_at <= ?", *req.To)
	}
	if req.Search != nil && *req.Search != "" {
		q = q.Where("title LIKE ?", "%"+*req.Search+"%")
	}

	return count, q.Count(&count).Error
}

// GetConversationWithLastMessage retrieves a conversation with its last message
func (r *MySQLRepo) GetConversationWithLastMessage(ctx context.Context, conversationID string) (*entity.ConversationSummary, error) {
	conv, err := r.FindByID(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	result := &entity.ConversationSummary{
		Conversation: *conv,
	}

	// Load last message
	lastMessage, err := r.getLastMessageForConversation(ctx, conversationID)
	if err == nil {
		result.LastMessage = lastMessage
	}

	// Load message counts
	totalMessages, err := r.countMessagesForConversation(ctx, conversationID)
	if err == nil {
		result.TotalMessages = totalMessages
	}

	result.UnreadCount = 0 // TODO: Implement unread count logic

	return result, nil
}

// CheckConversationAccess verifies if a user has access to a conversation
func (r *MySQLRepo) CheckConversationAccess(ctx context.Context, conversationID, userID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Conversation{}).
		Where("id = ? AND user_id = ?", conversationID, userID).
		Count(&count).Error

	return count > 0, err
}

// Helper methods

// getLastMessageForConversation retrieves the last message for a conversation
func (r *MySQLRepo) getLastMessageForConversation(ctx context.Context, conversationID string) (*entity.MessagePreview, error) {
	var msg entity.Message
	err := r.db.WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		Order("created_at DESC").
		First(&msg).Error

	if err != nil {
		return nil, err
	}

	return &entity.MessagePreview{
		ID:        msg.ID,
		Role:      msg.Role,
		Content:   msg.Content,
		CreatedAt: msg.CreatedAt,
	}, nil
}

// countMessagesForConversation counts total messages in a conversation
func (r *MySQLRepo) countMessagesForConversation(ctx context.Context, conversationID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Message{}).
		Where("conversation_id = ?", conversationID).
		Count(&count).Error

	return count, err
}
