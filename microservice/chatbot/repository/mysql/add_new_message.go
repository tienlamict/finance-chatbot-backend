package mysql

import (
	"context"
	"finance-chatbot/microservice/chatbot/entity"

	"github.com/pkg/errors"
)

func (repo *mysqlRepo) SendNewMessage(ctx context.Context, data *entity.ChatDataCreation) error {
	if err := repo.db.Create(data).Error; err != nil {
		return errors.WithStack(err)
	}
	return nil
}
