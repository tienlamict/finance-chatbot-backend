package mysql

import (
	"gorm.io/gorm"
)

type MySQLRepo struct {
	db *gorm.DB
}

func NewMySQLRepo(db *gorm.DB) *MySQLRepo {
	return &MySQLRepo{db: db}
}

func (r *MySQLRepo) DB() *gorm.DB { return r.db }
