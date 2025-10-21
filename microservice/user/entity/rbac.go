package entity

type Role struct {
	ID          int    `json:"id" gorm:"column:id;primaryKey"`
	Code        string `json:"code" gorm:"column:code"`
	Name        string `json:"name" gorm:"column:name"`
	Description string `json:"description" gorm:"column:description"`
}

func (Role) TableName() string { return "roles" }

type Permission struct {
	ID          int    `json:"id" gorm:"column:id;primaryKey"`
	Code        string `json:"code" gorm:"column:code"`
	Name        string `json:"name" gorm:"column:name"`
	Description string `json:"description" gorm:"column:description"`
}

func (Permission) TableName() string { return "permissions" }
