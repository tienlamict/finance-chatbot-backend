// finance-chatbot-backend/microservice/chatbot/repository/mysql/message_store.go
package mysql

import (
	"context"

	"finance-chatbot/microservice/chatbot/entity"
)

type MessageStore interface {
	CreateMessage(ctx context.Context, msg *entity.Message) error
	CountByConversation(ctx context.Context, convID string) (int64, error)
	ListByConversation(ctx context.Context, convID string, limit, offset int) ([]entity.Message, error)

	// Enhanced conversation history methods
	GetConversationHistory(ctx context.Context, req entity.ConversationHistoryRequest) ([]entity.MessageWithAttachments, error)
	CountConversationHistory(ctx context.Context, req entity.ConversationHistoryRequest) (int64, error)
	GetMessageByID(ctx context.Context, messageID string) (*entity.Message, error)
	GetMessagesWithAttachments(ctx context.Context, messageIDs []string) ([]entity.MessageWithAttachments, error)
}

func (r *MySQLRepo) CreateMessage(ctx context.Context, msg *entity.Message) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

func (r *MySQLRepo) CountByConversation(ctx context.Context, convID string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&entity.Message{}).
		Where("conversation_id = ?", convID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *MySQLRepo) ListByConversation(ctx context.Context, convID string, limit, offset int) ([]entity.Message, error) {
	var items []entity.Message
	q := r.db.WithContext(ctx).
		Where("conversation_id = ?", convID).
		Order("created_at ASC")
	if limit > 0 {
		q = q.Limit(limit).Offset(offset)
	}
	if err := q.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// GetConversationHistory retrieves conversation history with advanced filtering
func (r *MySQLRepo) GetConversationHistory(ctx context.Context, req entity.ConversationHistoryRequest) ([]entity.MessageWithAttachments, error) {
	var messages []entity.Message
	q := r.db.WithContext(ctx).Where("conversation_id = ?", req.ConversationID)

	// Apply time range filters
	if req.From != nil {
		q = q.Where("created_at >= ?", *req.From)
	}
	if req.To != nil {
		q = q.Where("created_at <= ?", *req.To)
	}

	// Apply cursor-based pagination
	if req.Before != nil {
		// Try to parse as timestamp first, then as message ID
		if msg, err := r.GetMessageByID(ctx, *req.Before); err == nil {
			q = q.Where("created_at < ?", msg.CreatedAt)
		} else {
			q = q.Where("id < ?", *req.Before)
		}
	}
	if req.After != nil {
		if msg, err := r.GetMessageByID(ctx, *req.After); err == nil {
			q = q.Where("created_at > ?", msg.CreatedAt)
		} else {
			q = q.Where("id > ?", *req.After)
		}
	}

	// Apply search filter
	if req.Search != nil && *req.Search != "" {
		q = q.Where("content LIKE ?", "%"+*req.Search+"%")
	}

	// Apply ordering
	order := "ASC"
	if req.Order == "desc" {
		order = "DESC"
	}
	q = q.Order("created_at " + order)

	// Apply limit
	if req.Limit > 0 {
		q = q.Limit(req.Limit)
	}

	if err := q.Find(&messages).Error; err != nil {
		return nil, err
	}

	// Convert to MessageWithAttachments
	result := make([]entity.MessageWithAttachments, len(messages))
	for i, msg := range messages {
		result[i] = entity.MessageWithAttachments{
			Message: msg,
		}

		// Load attachments if requested
		if req.IncludeAttachments {
			attachments, err := r.getAttachmentsForMessage(ctx, msg.ID)
			if err == nil {
				result[i].Attachments = attachments
			}
		}
	}

	return result, nil
}

// CountConversationHistory counts messages matching the filter criteria
func (r *MySQLRepo) CountConversationHistory(ctx context.Context, req entity.ConversationHistoryRequest) (int64, error) {
	var count int64
	q := r.db.WithContext(ctx).Model(&entity.Message{}).Where("conversation_id = ?", req.ConversationID)

	// Apply same filters as GetConversationHistory
	if req.From != nil {
		q = q.Where("created_at >= ?", *req.From)
	}
	if req.To != nil {
		q = q.Where("created_at <= ?", *req.To)
	}
	if req.Search != nil && *req.Search != "" {
		q = q.Where("content LIKE ?", "%"+*req.Search+"%")
	}

	return count, q.Count(&count).Error
}

// GetMessageByID retrieves a single message by ID
func (r *MySQLRepo) GetMessageByID(ctx context.Context, messageID string) (*entity.Message, error) {
	var msg entity.Message
	if err := r.db.WithContext(ctx).First(&msg, "id = ?", messageID).Error; err != nil {
		return nil, err
	}
	return &msg, nil
}

// GetMessagesWithAttachments retrieves multiple messages with their attachments
func (r *MySQLRepo) GetMessagesWithAttachments(ctx context.Context, messageIDs []string) ([]entity.MessageWithAttachments, error) {
	var messages []entity.Message
	if err := r.db.WithContext(ctx).Where("id IN ?", messageIDs).Find(&messages).Error; err != nil {
		return nil, err
	}

	result := make([]entity.MessageWithAttachments, len(messages))
	for i, msg := range messages {
		result[i] = entity.MessageWithAttachments{
			Message: msg,
		}

		// Load attachments
		attachments, err := r.getAttachmentsForMessage(ctx, msg.ID)
		if err == nil {
			result[i].Attachments = attachments
		}
	}

	return result, nil
}

// getAttachmentsForMessage retrieves attachments for a specific message
func (r *MySQLRepo) getAttachmentsForMessage(ctx context.Context, messageID string) ([]entity.AttachmentDTO, error) {
	var attachments []entity.MessageAttachment
	if err := r.db.WithContext(ctx).Where("message_id = ?", messageID).Find(&attachments).Error; err != nil {
		return nil, err
	}

	result := make([]entity.AttachmentDTO, len(attachments))
	for i, att := range attachments {
		result[i] = entity.AttachmentDTO{
			ID:         att.ID,
			Filename:   att.Filename,
			MimeType:   *att.MimeType,
			SHA256:     *att.SHA256,
			Pages:      att.Pages,
			CreatedAt:  att.CreatedAt,
			StorageKey: att.StorageKey,
		}
	}

	return result, nil
}
