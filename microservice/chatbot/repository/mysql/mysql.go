package mysql

import (
	"github.com/google/uuid"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"finance-chatbot/microservice/chatbot/entity"
)

type ChatRepo struct{ db *gorm.DB }

func OpenMySQL(dsn string) (*gorm.DB, error) {
	return gorm.Open(mysql.Open(dsn), &gorm.Config{})
}

func NewMySQLRepository(db *gorm.DB) *ChatRepo { return &ChatRepo{db: db} }

func (r *ChatRepo) AutoMigrate() error {
	return r.db.AutoMigrate(&entity.ChatMessage{})
}

func (r *ChatRepo) Create(msg *entity.ChatMessage) error {
	if msg.ID == "" {
		msg.ID = uuid.NewString()
	}
	return r.db.Create(msg).Error
}

func (r *ChatRepo) ListByUser(userID string, limit int) ([]entity.ChatMessage, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var out []entity.ChatMessage
	err := r.db.Where("user_id = ?", userID).Order("created_at desc").Limit(limit).Find(&out).Error
	return out, err
}
