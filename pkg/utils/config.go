package utils

import (
	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Email    EmailConfig
	OTP      OTPConfig
}

type AppConfig struct {
	Name    string
	Port    string
	Debug   bool
	LogPath string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	MaxConns int32
}

type JWTConfig struct {
	Secret      string
	ExpiryHours int
}

type EmailConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
}

type OTPConfig struct {
	ExpiryMinutes int
	Length        int
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	// Set defaults
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("DEBUG", false)
	viper.SetDefault("DB_MAX_CONNS", 10)
	viper.SetDefault("JWT_EXPIRY_HOURS", 24)
	viper.SetDefault("OTP_EXPIRY_MINUTES", 10)
	viper.SetDefault("OTP_LENGTH", 6)
	viper.SetDefault("LOG_PATH", "logs/")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	viper.AutomaticEnv()

	config := &Config{
		App: AppConfig{
			Name:    viper.GetString("APP_NAME"),
			Port:    viper.GetString("PORT"),
			Debug:   viper.GetBool("DEBUG"),
			LogPath: viper.GetString("LOG_PATH"),
		},
		Database: DatabaseConfig{
			Host:     viper.GetString("DB_HOST"),
			Port:     viper.GetString("DB_PORT"),
			Name:     viper.GetString("DB_NAME"),
			User:     viper.GetString("DB_USER"),
			Password: viper.GetString("DB_PASS"),
			MaxConns: viper.GetInt32("DB_MAX_CONNS"),
		},
		JWT: JWTConfig{
			Secret:      viper.GetString("JWT_SECRET"),
			ExpiryHours: viper.GetInt("JWT_EXPIRY_HOURS"),
		},
		Email: EmailConfig{
			Host:     viper.GetString("SMTP_HOST"),
			Port:     viper.GetInt("SMTP_PORT"),
			User:     viper.GetString("SMTP_USER"),
			Password: viper.GetString("SMTP_PASS"),
			From:     viper.GetString("EMAIL_FROM"),
		},
		OTP: OTPConfig{
			ExpiryMinutes: viper.GetInt("OTP_EXPIRY_MINUTES"),
			Length:        viper.GetInt("OTP_LENGTH"),
		},
	}

	return config, nil
}
