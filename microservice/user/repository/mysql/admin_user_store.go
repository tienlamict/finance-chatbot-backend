package mysql

import (
	"context"
	"finance-chatbot/addon/core"
	"finance-chatbot/microservice/user/entity"
	"fmt"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// AdminCreateUser creates a new user by admin
func (repo *mysqlRepo) AdminCreateUser(ctx context.Context, req *entity.AdminCreateUserRequest) (*entity.User, error) {
	// Check if email already exists
	var count int64
	if err := repo.db.Model(&entity.User{}).Where("email = ?", req.Email).Count(&count).Error; err != nil {
		return nil, errors.Wrap(err, "failed to check email existence")
	}

	if count > 0 {
		return nil, entity.ErrUserAlreadyExists
	}

	user := entity.User{
		SQLModel:   core.NewSQLModel(),
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Email:      req.Email,
		Phone:      req.Phone,
		Gender:     req.Gender,
		SystemRole: req.SystemRole,
		Status:     req.Status,
	}

	if err := repo.db.Create(&user).Error; err != nil {
		return nil, errors.Wrap(err, entity.ErrCannotCreateUser.Error())
	}

	return &user, nil
}

// AdminUpdateUser updates a user by admin
func (repo *mysqlRepo) AdminUpdateUser(ctx context.Context, userID int, req *entity.AdminUpdateUserRequest) error {
	// Check if user exists
	var user entity.User
	if err := repo.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return entity.ErrUserNotFound
		}
		return errors.Wrap(err, "failed to get user")
	}

	// Check if email already exists (if email is being updated)
	if req.Email != nil && *req.Email != user.Email {
		var count int64
		if err := repo.db.Model(&entity.User{}).Where("email = ? AND id != ?", *req.Email, userID).Count(&count).Error; err != nil {
			return errors.Wrap(err, "failed to check email existence")
		}

		if count > 0 {
			return entity.ErrUserAlreadyExists
		}
	}

	// Build update map
	updateData := make(map[string]interface{})

	if req.FirstName != nil {
		updateData["first_name"] = *req.FirstName
	}
	if req.LastName != nil {
		updateData["last_name"] = *req.LastName
	}
	if req.Email != nil {
		updateData["email"] = *req.Email
	}
	if req.Phone != nil {
		updateData["phone"] = *req.Phone
	}
	if req.Gender != nil {
		updateData["gender"] = *req.Gender
	}
	if req.SystemRole != nil {
		updateData["system_role"] = *req.SystemRole
	}
	if req.Status != nil {
		updateData["status"] = *req.Status
	}

	if len(updateData) == 0 {
		return nil // Nothing to update
	}

	if err := repo.db.Model(&entity.User{}).Where("id = ?", userID).Updates(updateData).Error; err != nil {
		return errors.Wrap(err, entity.ErrCannotUpdateUser.Error())
	}

	return nil
}

// AdminDeleteUser soft deletes a user by admin
func (repo *mysqlRepo) AdminDeleteUser(ctx context.Context, userID int) error {
	result := repo.db.Where("id = ?", userID).Delete(&entity.User{})

	if result.Error != nil {
		return errors.Wrap(result.Error, entity.ErrCannotDeleteUser.Error())
	}

	if result.RowsAffected == 0 {
		return entity.ErrUserNotFound
	}

	return nil
}

// AdminListUsers lists users with pagination and filters
func (repo *mysqlRepo) AdminListUsers(ctx context.Context, filter *entity.ListUsersFilter) ([]entity.User, error) {
	var users []entity.User

	db := repo.db.Model(&entity.User{})

	// Apply filters
	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		db = db.Where("first_name LIKE ? OR last_name LIKE ? OR email LIKE ?",
			searchPattern, searchPattern, searchPattern)
	}

	if filter.Role != "" {
		db = db.Where("system_role = ?", filter.Role)
	}

	if filter.Status != "" {
		db = db.Where("status = ?", filter.Status)
	}

	if filter.Email != "" {
		db = db.Where("email = ?", filter.Email)
	}

	// Count total
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, errors.Wrap(err, "failed to count users")
	}

	filter.Total = total

	// Apply pagination
	offset := (filter.Page - 1) * filter.Limit
	db = db.Offset(offset).Limit(filter.Limit)

	// Order by created_at desc
	db = db.Order("created_at DESC")

	if err := db.Find(&users).Error; err != nil {
		return nil, errors.Wrap(err, entity.ErrCannotGetUsers.Error())
	}

	return users, nil
}

// AdminGetUserByID gets a user by ID for admin
func (repo *mysqlRepo) AdminGetUserByID(ctx context.Context, userID int) (*entity.User, error) {
	var user entity.User

	if err := repo.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, entity.ErrUserNotFound
		}
		return nil, errors.Wrap(err, entity.ErrCannotGetUser.Error())
	}

	return &user, nil
}

// CheckEmailExists checks if an email already exists (excluding a specific user ID if provided)
func (repo *mysqlRepo) CheckEmailExists(ctx context.Context, email string, excludeUserID *int) (bool, error) {
	var count int64
	query := repo.db.Model(&entity.User{}).Where("email = ?", email)

	if excludeUserID != nil {
		query = query.Where("id != ?", *excludeUserID)
	}

	if err := query.Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return count > 0, nil
}
