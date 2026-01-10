package config

import (
	"Backend_Dorm_PTIT/logger"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Server     ServerConfig         `mapstructure:"server"`
	Database   DatabaseConfig       `mapstructure:"database"`
	CORS       CORSConfig           `mapstructure:"cors"`
	JWT        JWTConfig            `mapstructure:"jwt"`
	Redis      RedisConfig          `mapstructure:"redis"`
	MailGoogle MailGoogleConfig     `mapstructure:"mail_google"`
	Logging    logger.LogConfig     `mapstructure:"logging"`
	Cloudinary CloudinaryConfig     `mapstructure:"cloudinary"`
	WebSocket  WebSocketConfig      `mapstructure:"websocket"`
	APIKey     APIKeyConfig         `mapstructure:"api_key"`
	Chatbot    ChatbotConfig        `mapstructure:"chatbot"`
}

type ServerConfig struct {
	Host    string `mapstructure:"host"`
	Port    string `mapstructure:"port"`
	GinMode string `mapstructure:"gin_mode"` // debug, release, test
}
type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	Refresh_Exp int    `mapstructure:"refresh_exp"`
	Access_Exp  int    `mapstructure:"access_exp"`
}

type RedisConfig struct {
	Address  string `mapstructure:"address"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type CloudinaryConfig struct {
	Apikey    string `mapstructure:"apikey"`
	Secret    string `mapstructure:"secret"`
	CloudName string `mapstructure:"cloudname"`
}
type MailGoogleConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Email    string `mapstructure:"email"`
	Password string `mapstructure:"password"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	UserName string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"sslmode"`
	Schema   string `mapstructure:"schema"`
}

type CORSConfig struct {
	AllowOrigins []string
	AllowMethods string
	AllowHeaders string
	AllowCreds   bool
}
type WebSocketConfig struct {
	WriteWait  int `mapstructure:"write_wait"`  // in seconds
	PongWait   int `mapstructure:"pong_wait"`   // in seconds
	PingPeriod int `mapstructure:"ping_period"` // in seconds
}

type APIKeyConfig struct {
	ChatbotService string `mapstructure:"chatbot_service"`
}

// ChatbotConfig contains config for external chatbot service
type ChatbotConfig struct {
	BaseURL string `mapstructure:"base_url"` // e.g. https://chatbot.example.com
}


func LoadConfig(cfgFile string) (*Config, error) {
	// Use specific config file if provided
	viper.SetConfigFile(cfgFile)

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Allow environment variables to override config file
	// Environment variables should use format: SERVER_HOST, DATABASE_PORT, etc.
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Unmarshal config into struct
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}

func (c *DatabaseConfig) GetDSN() string {
	// Add binary_parameters=yes to force lib/pq to use the simple query protocol
	// This avoids issues with unnamed prepared statements when going through
	// connection poolers like PgBouncer/Neon ("bind message has X result formats...").
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s binary_parameters=yes",
		c.Host, c.Port, c.UserName, c.Password, c.Name, c.SSLMode)
}
