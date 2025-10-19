package mysql

import (
	"context"
	"finance-chatbot/addon/core"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type adminAuthRepo struct {
	db *gorm.DB
}

func NewAdminAuthRepository(db *gorm.DB) *adminAuthRepo {
	return &adminAuthRepo{db: db}
}

// AuthRecord represents authentication data
type AuthRecord struct {
	ID       int    `gorm:"column:id;primaryKey"`
	UserID   int    `gorm:"column:user_id"`
	AuthType string `gorm:"column:auth_type"`
	Email    string `gorm:"column:email"`
	Salt     string `gorm:"column:salt"`
	Password string `gorm:"column:password"`
}

func (AuthRecord) TableName() string { return "auths" }

// CreateAuthForUser creates authentication record for a user
func (repo *adminAuthRepo) CreateAuthForUser(ctx context.Context, userID int, email, salt, hashedPassword string) error {
	auth := AuthRecord{
		UserID:   userID,
		Email:    email,
		Salt:     salt,
		Password: hashedPassword,
		AuthType: "email_password",
	}

	if err := repo.db.Table(auth.TableName()).Create(&auth).Error; err != nil {
		return errors.Wrap(err, "failed to create auth record")
	}

	return nil
}

// UpdateUserPassword updates user's password
func (repo *adminAuthRepo) UpdateUserPassword(ctx context.Context, userID int, salt, hashedPassword string) error {
	updates := map[string]interface{}{
		"salt":     salt,
		"password": hashedPassword,
	}

	if err := repo.db.Table(AuthRecord{}.TableName()).
		Where("user_id = ?", userID).
		Updates(updates).Error; err != nil {
		return errors.Wrap(err, "failed to update password")
	}

	return nil
}

// GetAuthByUserID gets auth record by user ID
func (repo *adminAuthRepo) GetAuthByUserID(ctx context.Context, userID int) (interface{}, error) {
	var auth AuthRecord

	if err := repo.db.Table(auth.TableName()).
		Where("user_id = ?", userID).
		First(&auth).Error; err != nil {
		if err.Error() == "record not found" {
			return nil, core.ErrRecordNotFound
		}
		return nil, errors.Wrap(err, "failed to get auth record")
	}

	return &auth, nil
}

// GetAuthByEmail gets auth record by email
func (repo *adminAuthRepo) GetAuthByEmail(ctx context.Context, email string) (*AuthRecord, error) {
	var auth AuthRecord

	if err := repo.db.Table(auth.TableName()).
		Where("email = ?", email).
		First(&auth).Error; err != nil {
		if err.Error() == "record not found" {
			return nil, core.ErrRecordNotFound
		}
		return nil, errors.Wrap(err, "failed to get auth record")
	}

	return &auth, nil
}
