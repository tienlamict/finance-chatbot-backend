package mysql

import (
	"context"
	"finance-chatbot/addon/core"
	"finance-chatbot/microservice/task/entity"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func (repo *mysqlRepo) GetTaskById(ctx context.Context, id int) (*entity.Task, error) {
	var data entity.Task

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
