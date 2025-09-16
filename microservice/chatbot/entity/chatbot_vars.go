package entity

import (
	"finance-chatbot/addon/core"
)

// ChatDataCreation use for inserting data into database, we don't need all data fields
type ChatDataCreation struct {
	core.SQLModel
	Content string `json:"content" gorm:"column:content;" db:"content"`
	// Do not allow client set these fields
	UserID int    `json:"-" gorm:"column:user_id" db:"user_id"`
	Role   string `json:"-" gorm:"column:role;" db:"role"`
}

func (ChatDataCreation) TableName() string { return ChatMessage{}.TableName() }

func (t *ChatDataCreation) Prepare(userID int, role string) {
	t.SQLModel = core.NewSQLModel()
	t.UserID = userID
	t.Role = role
}
