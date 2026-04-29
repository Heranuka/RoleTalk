// Package config handles environment-specific configuration loading.
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

var ErrMissingEnvVar = errors.New("missing required environment variable")

type Config struct {
	Env           string        `mapstructure:"env"`
	App           App           `mapstructure:"app"`
	HTTP          HTTP          `mapstructure:"http"`
	Observability Observability `mapstructure:"observability"`
	RateLimit     RateLimit     `mapstructure:"rate_limit"`
	Auth          Auth          `mapstructure:"auth"`
	OAuth         OAuth         `mapstructure:"oauth"`
	SMTP          SMTP          `mapstructure:"smtp"`
	Postgres      Postgres      `mapstructure:"postgres"`
	MinIO         MinIO         `mapstructure:"minio"`
	Client        Client        `mapstructure:"client"`
	API           API           `mapstructure:"api"`
	Logging       Logging       `mapstructure:"logging"`
}

type App struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
}

type HTTP struct {
	Port            string        `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type Observability struct {
	Enabled      bool    `mapstructure:"enabled"`
	OTLPEndpoint string  `mapstructure:"otlp_endpoint"`
	LokiURL      string  `mapstructure:"loki_url"`
	SampleRate   float64 `mapstructure:"sample_rate"`
}

type RateLimit struct {
	Enabled         bool          `mapstructure:"enabled"`
	CleanupInterval time.Duration `mapstructure:"cleanup_interval"`
	Global          LimitConfig   `mapstructure:"global"`
	Auth            LimitConfig   `mapstructure:"auth"`
}

type LimitConfig struct {
	Limit float64 `mapstructure:"limit"`
	Burst int     `mapstructure:"burst"`
}

type Auth struct {
	AccessTokenTTL       time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL      time.Duration `mapstructure:"refresh_token_ttl"`
	EmailVerificationTTL time.Duration `mapstructure:"email_verification_ttl"`
	PasswordResetTTL     time.Duration `mapstructure:"password_reset_ttl"`
	Secret               string        `mapstructure:"-"`
}

type OAuth struct {
	Google GoogleOAuth `mapstructure:"google"`
}

type GoogleOAuth struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
}

type SMTP struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	Username    string `mapstructure:"-"`
	Password    string `mapstructure:"-"`
	FromAddress string `mapstructure:"from_address"`
}

type MinIO struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Bucket    string `mapstructure:"bucket"`
	UseSSL    bool   `mapstructure:"use_ssl"`
}

type Postgres struct {
	ConnectionURL string     `mapstructure:"connection_url"`
	Host          string     `mapstructure:"host"`
	Port          int        `mapstructure:"port"`
	SSLMode       string     `mapstructure:"ssl_mode"`
	Pool          PoolConfig `mapstructure:"pool"`
	User          string     `mapstructure:"-"`
	Password      string     `mapstructure:"-"`
	Database      string     `mapstructure:"-"`
}

type PoolConfig struct {
	MaxConns        int32         `mapstructure:"max_conns"`
	MinConns        int32         `mapstructure:"min_conns"`
	MaxConnLifetime time.Duration `mapstructure:"max_conn_lifetime"`
}

type Client struct {
	URL string `mapstructure:"url"`
}

type API struct {
	URL string `mapstructure:"url"`
}

type Logging struct {
	Level string `mapstructure:"level"`
}

func Load() (*Config, error) {
	v := viper.New()

	// 1. Initial Setup
	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 2. Load .env file explicitly from several possible locations
	v.SetConfigFile(".env")
	v.AddConfigPath(".")    // Current dir
	v.AddConfigPath("./..") // Parent dir (if running from cmd/app)
	v.AddConfigPath("./../..")

	if err := v.ReadInConfig(); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			fmt.Printf("Error reading .env: %v\n", err)
		}
	}

	// 3. Merge config.yml
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.AddConfigPath(".")
	if err := v.MergeInConfig(); err != nil {
		return nil, fmt.Errorf("merge config.yml: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	// 4. Force bind and load sensitive values
	// If Unmarshal skipped them because they aren't in YAML, we set them here.
	loadSensitiveValues(v, &cfg)

	// 5. Validation
	if err := validateRequired(&cfg); err != nil {
		return nil, err
	}

	if err := validatePostgres(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func loadSensitiveValues(v *viper.Viper, cfg *Config) {
	// We use the keys exactly as they appear in .env (minus APP_ prefix)
	cfg.Auth.Secret = v.GetString("JWT_SECRET")

	cfg.OAuth.Google.ClientID = v.GetString("OAUTH_GOOGLE_CLIENT_ID")
	cfg.OAuth.Google.ClientSecret = v.GetString("OAUTH_GOOGLE_CLIENT_SECRET")

	cfg.SMTP.Username = v.GetString("SMTP_USERNAME")
	cfg.SMTP.Password = v.GetString("SMTP_PASSWORD")

	cfg.Postgres.User = v.GetString("POSTGRES_USER")
	cfg.Postgres.Password = v.GetString("POSTGRES_PASSWORD")
	cfg.Postgres.Database = v.GetString("POSTGRES_DB")

	cfg.MinIO.AccessKey = v.GetString("MINIO_ACCESS_KEY")
	cfg.MinIO.SecretKey = v.GetString("MINIO_SECRET_KEY")

	cfg.Client.URL = v.GetString("CLIENT_URL")
	cfg.API.URL = v.GetString("API_URL")

	if dbURL := v.GetString("DATABASE_URL"); dbURL != "" {
		cfg.Postgres.ConnectionURL = dbURL
	}
}

func validateRequired(cfg *Config) error {
	required := map[string]string{
		"JWT Secret":           cfg.Auth.Secret,
		"Postgres DB name":     cfg.Postgres.Database,
		"Postgres User":        cfg.Postgres.User,
		"Postgres Password":    cfg.Postgres.Password,
		"Client URL":           cfg.Client.URL,
		"Google Client Secret": cfg.OAuth.Google.ClientSecret,
		"SMTP Password":        cfg.SMTP.Password,
	}

	for name, val := range required {
		if val == "" {
			return fmt.Errorf("%w: %s", ErrMissingEnvVar, name)
		}
	}
	return nil
}

func validatePostgres(cfg *Config) error {
	if cfg.Postgres.ConnectionURL == "" &&
		(cfg.Postgres.Host == "" || cfg.Postgres.Database == "") {
		return errors.New("postgres connection info is incomplete")
	}
	return nil
}
