package entity

type Role struct {
	ID          int    `gorm:"column:id;primaryKey"`
	Code        string `gorm:"column:code"`
	Name        string `gorm:"column:name"`
	Description string `gorm:"column:description"`
}

func (Role) TableName() string { return "roles" }

type Permission struct {
	ID          int    `gorm:"column:id;primaryKey"`
	Code        string `gorm:"column:code"`
	Name        string `gorm:"column:name"`
	Description string `gorm:"column:description"`
}

func (Permission) TableName() string { return "permissions" }
