package mysql

import (
	"context"
	core "finance-chatbot/addon/core"
	"finance-chatbot/microservice/auth/entity"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type mySqlRepo struct {
	db *gorm.DB
}

func NewMySQLRepository(db *gorm.DB) *mySqlRepo {
	return &mySqlRepo{db: db}
}

func (repo *mySqlRepo) AddNewAuth(ctx context.Context, data *entity.Auth) error {
	if err := repo.db.Table(data.TableName()).Create(data).Error; err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (repo *mySqlRepo) GetAuth(ctx context.Context, email string) (*entity.Auth, error) {
	var data entity.Auth

	if err := repo.db.
		Table(data.TableName()).
		Where("email = ?", email).
		First(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, core.ErrRecordNotFound
		}

		return nil, errors.WithStack(err)
	}

	return &data, nil
}
