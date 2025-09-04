package database

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Config struct {
	User     string
	Password string
	Host     string
	Port     string
	DBName   string
}

var DB *gorm.DB

// Kết nối với MySQL từ cấu hình.
// Cấu hình bao gồm thông tin người dùng, mật khẩu, máy chủ, cổng và tên cơ sở dữ liệu.
// Trả về một đối tượng *gorm.DB để tương tác với cơ sở dữ liệu.
func Connect(cfg Config) *gorm.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to MySQL: %v", err)
	}

	DB = db
	return DB
}

// AutoMigrate tự động tạo bảng và cập nhật cấu trúc dựa trên các mô hình đã cho.
// Các mô hình này phải được định nghĩa trước khi gọi hàm này.
func AutoMigrate(models ...interface{}) {
	err := DB.AutoMigrate(models...)
	if err != nil {
		log.Fatalf("auto migration failed: %v", err)
	}
}
