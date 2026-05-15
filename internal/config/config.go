package config

import (
	"log"

	"github.com/spf13/viper" // Viper 是一个非常流行的 Go 配置库，支持多种格式（JSON, YAML, TOML 等）和环境变量，非常适合我们的需求
)

// AppConfig 是总配置结构体
type AppConfig struct {
	Server   ServerConfig
	Database DatabaseConfig
	Jwt      JWTConfig
}

type ServerConfig struct {
	Port int
}
type DatabaseConfig struct {
	DSN string
}
type JWTConfig struct {
	Secret string
	Expire int
}

// Conf 是一个全局变量，包含解析后的全部配置
var Conf *AppConfig

// InitConfig 将在 main.go 中被调用，读取上面写的 config.yaml
func InitConfig() {
	viper.SetConfigFile("config/config.yaml") // 告诉 Viper 去哪找文件
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}

	Conf = &AppConfig{}
	if err := viper.Unmarshal(Conf); err != nil {
		log.Fatalf("解析配置到结构体失败: %v", err)
	}
	log.Println("配置加载成功！")
}
