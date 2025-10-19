// finance-chatbot-backend/microservice/chatbot/repository/mysql/attachment_store.go
package mysql

import (
	"context"

	"finance-chatbot/microservice/chatbot/entity"
)

type AttachmentStore interface {
	CreateAttachment(ctx context.Context, attachment *entity.MessageAttachment) error
	CreateAttachmentsBatch(ctx context.Context, attachments []*entity.MessageAttachment) error
	GetAttachmentsByMessageID(ctx context.Context, messageID string) ([]*entity.MessageAttachment, error)
	GetAttachmentByID(ctx context.Context, id string) (*entity.MessageAttachment, error)
	DeleteAttachment(ctx context.Context, id string) error
}

func (r *MySQLRepo) CreateAttachment(ctx context.Context, attachment *entity.MessageAttachment) error {
	return r.db.WithContext(ctx).Create(attachment).Error
}

func (r *MySQLRepo) CreateAttachmentsBatch(ctx context.Context, attachments []*entity.MessageAttachment) error {
	if len(attachments) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&attachments).Error
}

func (r *MySQLRepo) GetAttachmentsByMessageID(ctx context.Context, messageID string) ([]*entity.MessageAttachment, error) {
	var attachments []*entity.MessageAttachment
	err := r.db.WithContext(ctx).
		Where("message_id = ?", messageID).
		Order("created_at ASC").
		Find(&attachments).Error
	if err != nil {
		return nil, err
	}
	return attachments, nil
}

func (r *MySQLRepo) GetAttachmentByID(ctx context.Context, id string) (*entity.MessageAttachment, error) {
	var attachment entity.MessageAttachment
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&attachment).Error
	if err != nil {
		return nil, err
	}
	return &attachment, nil
}

func (r *MySQLRepo) DeleteAttachment(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&entity.MessageAttachment{}).Error
}
