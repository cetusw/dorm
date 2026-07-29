package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"
)

type CronConfig struct {
	WeekStart string `json:"weekStart"`
	SyncStart string `json:"syncStart"`
}

type WebPushConfig struct {
	Enabled    bool
	PublicKey  string
	PrivateKey string
	Subject    string
}

type AppConfig struct {
	DBUser     string
	DBPassword string
	DBHost     string
	DBPort     string
	DBName     string
	TZ         string
	DBDriver   string
	AuthSecret string

	Cron                         CronConfig
	BotToken                     string
	GoogleCredentials            string
	WebPush                      WebPushConfig
	NotificationSchedulerEnabled bool
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

	cfg.getAuthConfig()
	cfg.getWebPushConfig()
	cfg.getNotificationSchedulerConfig()

	if err := cfg.WebPush.Validate(); err != nil {
		return nil, err
	}
	if err := cfg.ValidateTimezone(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func LoadDatabaseConfig() (*AppConfig, error) {
	cfg := &AppConfig{}
	if err := cfg.getDBConfig(); err != nil {
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

func (cfg *AppConfig) getAuthConfig() {
	cfg.AuthSecret = os.Getenv("AUTH_SECRET")
	if cfg.AuthSecret != "" {
		return
	}

	if cfg.DBPassword != "" {
		cfg.AuthSecret = cfg.DBPassword
		log.Println("Warning: AUTH_SECRET not set, using DB_PASSWORD as auth secret fallback.")
		return
	}

	cfg.AuthSecret = "dev-auth-secret"
	log.Println("Warning: AUTH_SECRET not set, using insecure development fallback secret.")
}

func (cfg *AppConfig) getWebPushConfig() {
	cfg.WebPush.Enabled = strings.EqualFold(strings.TrimSpace(os.Getenv("WEB_PUSH_ENABLED")), "true")
	cfg.WebPush.PublicKey = strings.TrimSpace(os.Getenv("WEB_PUSH_PUBLIC_KEY"))
	cfg.WebPush.PrivateKey = strings.TrimSpace(os.Getenv("WEB_PUSH_PRIVATE_KEY"))
	cfg.WebPush.Subject = strings.TrimSpace(os.Getenv("WEB_PUSH_SUBJECT"))
}

func (cfg *AppConfig) getNotificationSchedulerConfig() {
	cfg.NotificationSchedulerEnabled = strings.EqualFold(
		strings.TrimSpace(os.Getenv("NOTIFICATION_SCHEDULER_ENABLED")),
		"true",
	)
}

func (cfg WebPushConfig) Validate() error {
	if !cfg.Enabled {
		return nil
	}

	if cfg.PublicKey == "" {
		return errors.New("WEB_PUSH_PUBLIC_KEY is required when WEB_PUSH_ENABLED=true")
	}
	if cfg.PrivateKey == "" {
		return errors.New("WEB_PUSH_PRIVATE_KEY is required when WEB_PUSH_ENABLED=true")
	}
	if cfg.Subject == "" {
		return errors.New("WEB_PUSH_SUBJECT is required when WEB_PUSH_ENABLED=true")
	}

	parsedSubject, err := url.Parse(cfg.Subject)
	if err != nil || parsedSubject == nil {
		return errors.New("WEB_PUSH_SUBJECT must be a valid mailto: or https URL")
	}

	switch parsedSubject.Scheme {
	case "mailto":
		if parsedSubject.Opaque == "" {
			return errors.New("WEB_PUSH_SUBJECT mailto: value must include an email address")
		}
	case "https":
		if parsedSubject.Host == "" {
			return errors.New("WEB_PUSH_SUBJECT https URL must include a host")
		}
	default:
		return errors.New("WEB_PUSH_SUBJECT must start with mailto: or https://")
	}

	return nil
}

func (cfg *AppConfig) ValidateTimezone() error {
	if strings.TrimSpace(cfg.TZ) == "" {
		return errors.New("TZ environment variable is required")
	}

	if _, err := time.LoadLocation(cfg.TZ); err != nil {
		return fmt.Errorf("invalid TZ value %q: %w", cfg.TZ, err)
	}

	return nil
}
