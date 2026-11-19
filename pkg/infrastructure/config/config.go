package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"strings"
	"time"
)

type CronConfig struct {
	WeekStart string `json:"weekStart"`
}

type WebPushConfig struct {
	Enabled    bool
	PublicKey  string
	PrivateKey string
	Subject    string
}

type AppConfig struct {
	DBUser             string
	DBPassword         string
	DBHost             string
	DBPort             string
	DBName             string
	TZ                 string
	DBDriver           string
	Production         bool
	AuthOrigins        []string
	AuthTrustedProxies []string

	Cron                         CronConfig
	WebPush                      WebPushConfig
	NotificationSchedulerEnabled bool
}

func LoadConfig(configPath string) (*AppConfig, error) {
	cfg := &AppConfig{}

	err := cfg.getCronConfig(configPath)
	if err != nil {
		return nil, err
	}

	err = cfg.getDBConfig()
	if err != nil {
		return nil, err
	}

	if err := cfg.getAuthConfig(); err != nil {
		return nil, err
	}
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
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true&allowNativePasswords=true&loc=%s&timeout=5s&readTimeout=30s&writeTimeout=15s",
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

func (cfg *AppConfig) getAuthConfig() error {
	mode := strings.TrimSpace(os.Getenv("APP_ENV"))
	if mode != "" && mode != "development" && mode != "production" {
		return errors.New("APP_ENV must be development or production")
	}
	cfg.Production = mode == "production"
	origins := []string{}
	if !cfg.Production {
		origins = []string{"http://localhost:8080", "https://localhost:5173", "http://127.0.0.1:8080", "https://127.0.0.1:5173"}
	}
	if configured := os.Getenv("AUTH_ORIGIN"); configured != "" || cfg.Production {
		origins = append(origins, configured)
	}
	for _, value := range origins {
		origin, err := NormalizeOrigin(value)
		if err != nil {
			return fmt.Errorf("AUTH_ORIGIN: %w", err)
		}
		if cfg.Production && !strings.HasPrefix(origin, "https://") {
			return errors.New("production AUTH_ORIGIN must use HTTPS")
		}
		cfg.AuthOrigins = append(cfg.AuthOrigins, origin)
	}
	for _, entry := range strings.Split(os.Getenv("AUTH_TRUSTED_PROXIES"), ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if ip := net.ParseIP(entry); ip == nil {
			_, network, err := net.ParseCIDR(entry)
			if err != nil {
				return errors.New("invalid AUTH_TRUSTED_PROXIES address")
			}
			ones, _ := network.Mask.Size()
			if ones == 0 {
				return errors.New("AUTH_TRUSTED_PROXIES must not trust every address")
			}
		} else if ip.IsUnspecified() {
			return errors.New("AUTH_TRUSTED_PROXIES must not trust unspecified addresses")
		}
		cfg.AuthTrustedProxies = append(cfg.AuthTrustedProxies, entry)
	}
	if cfg.Production && len(cfg.AuthTrustedProxies) == 0 {
		return errors.New("production requires explicit AUTH_TRUSTED_PROXIES and a proxy-provided X-Real-IP")
	}
	return nil
}

func NormalizeOrigin(value string) (string, error) {
	u, err := url.Parse(value)
	if err != nil || u == nil || u.User != nil || u.Opaque != "" || u.Hostname() == "" || u.RawQuery != "" || u.Fragment != "" || u.Path != "" && u.Path != "/" {
		return "", errors.New("must be an HTTP(S) origin without credentials, path or query")
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return "", errors.New("must use HTTP(S)")
	}
	host := strings.ToLower(u.Hostname())
	port := u.Port()
	if port == "80" && scheme == "http" || port == "443" && scheme == "https" {
		port = ""
	}
	if port != "" {
		host = net.JoinHostPort(host, port)
	} else if strings.Contains(host, ":") {
		host = "[" + host + "]"
	}
	return scheme + "://" + host, nil
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
