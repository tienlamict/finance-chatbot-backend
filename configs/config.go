package configs

import (
	db "finance-chatbot/internal/database"
	"log"

	"github.com/spf13/viper"
)

// Config đại diện cho cấu hình ứng dụng, bao gồm cấu hình cơ sở dữ liệu, máy chủ và Ethereum.
// Các trường được ánh xạ từ file YAML sử dụng thẻ `yaml` và `mapstructure` để đảm bảo đúng tên trường.

type Config struct {
	Database db.Config `yaml:"database"`

	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`
}

var config *Config

func LoadConfig() *Config {
	if config != nil {
		return config
	}

	viper.SetConfigFile("configs/config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	config = &Config{}
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Unable to unmarshal config: %v", err)
	}
	return config
}
