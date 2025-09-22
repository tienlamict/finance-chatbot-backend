package mysql

import (
	"context"
	"finance-chatbot/addon/core"
	"finance-chatbot/microservice/chatbot/entity"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func (repo *ChatRepoMySQL) GetMessageByUserId(ctx context.Context, id int) (*entity.ChatMessage, error) {
	var data entity.ChatMessage

	if err := repo.db.
		Table(data.TableName()).
		Where("id = ?", id).
		First(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, core.ErrRecordNotFound
		}

		return nil, errors.WithStack(err)
	}

	return &data, nil
}
