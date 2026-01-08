package config

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
)

type CronConfig struct {
	WeekStart string `json:"weekStart"`
	SyncStart string `json:"syncStart"`
}

type AppConfig struct {
	DBUser     string
	DBPassword string
	DBHost     string
	DBPort     string
	DBName     string
	TZ         string
	DBDriver   string

	Cron              CronConfig
	BotToken          string
	GoogleCredentials string
}

func LoadConfig(configPath string) (*AppConfig, error) {
	cfg := &AppConfig{}

	err := cfg.getCronConfig(configPath)
	if err != nil {
		return nil, err
	}

	cfg.getBotToken()
	err = cfg.getGoogleConfig()
	if err != nil {
		return nil, err
	}

	err = cfg.getDBConfig()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (cfg *AppConfig) GetDBConnectionString() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true&allowNativePasswords=true&loc=%s",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
		url.QueryEscape(cfg.TZ),
	)
}

func (cfg *AppConfig) getCronConfig(configPath string) error {
	file, err := os.Open(configPath)
	if err != nil {
		return fmt.Errorf("failed to open config file '%s': %w", configPath, err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg.Cron); err != nil {
		return fmt.Errorf("failed to decode config file '%s': %w", configPath, err)
	}
	log.Printf("Loaded Cron config: %+v", cfg.Cron)
	return nil
}

func (cfg *AppConfig) getBotToken() {
	cfg.BotToken = os.Getenv("BOT_TOKEN")
	if cfg.BotToken == "" {
		log.Println("Warning: BOT_TOKEN not set in environment variables.")
	}
}

func (cfg *AppConfig) getGoogleConfig() error {
	googleCredentialsPath := os.Getenv("GOOGLE_CREDENTIALS")
	if googleCredentialsPath == "" {
		return fmt.Errorf("GOOGLE_CREDENTIALS environment variable not set")
	}
	b, err := os.ReadFile(googleCredentialsPath)
	if err != nil {
		return fmt.Errorf("failed to read GOOGLE_CREDENTIALS file: %w", err)
	}
	cfg.GoogleCredentials = string(b)

	return nil
}

func (cfg *AppConfig) getDBConfig() error {
	cfg.DBUser = os.Getenv("DB_USER")
	cfg.DBPassword = os.Getenv("DB_PASSWORD")
	cfg.DBHost = os.Getenv("DB_HOST")
	cfg.DBPort = os.Getenv("DB_PORT")
	cfg.DBName = os.Getenv("DB_NAME")
	cfg.TZ = os.Getenv("TZ")
	if driver := os.Getenv("DB_DRIVER"); driver != "" {
		cfg.DBDriver = driver
	}

	if cfg.DBUser == "" || cfg.DBHost == "" || cfg.DBName == "" || cfg.DBPort == "" {
		return fmt.Errorf("database connection parameters are incomplete (user, host, name, port must be set)")
	}
	if cfg.TZ == "" {
		log.Println("Warning: TZ environment variable not set, defaulting may occur.")
	}
	return nil
}
