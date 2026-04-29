// Package config provides application configuration loading and validation.
// It reads configuration from YAML files, environment variables, and defaults,
// and exposes a fully populated Config struct for use across the application.
package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// ErrMissingEnvVar is returned when a required environment variable is not set.
var ErrMissingEnvVar = errors.New("missing required environment variable")

// Config holds all application configuration parameters.
type Config struct {
	Env string `mapstructure:"env"`

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

// App contains basic application metadata.
type App struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
}

// HTTP contains settings for the HTTP server.
type HTTP struct {
	Port            string        `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

// Observability contains settings for Tracing (Tempo) and Logs (Loki).
type Observability struct {
	Enabled      bool    `mapstructure:"enabled"`
	OTLPEndpoint string  `mapstructure:"otlp_endpoint"` // e.g. "tempo:4318" or "otel-collector:4318"
	LokiURL      string  `mapstructure:"loki_url"`      // e.g. "http://loki:3100/loki/api/v1/push"
	SampleRate   float64 `mapstructure:"sample_rate"`   // percentage of traces to collect (0.0 to 1.0)
}

// RateLimit contains settings for the rate limiter middleware.
type RateLimit struct {
	Enabled         bool          `mapstructure:"enabled"`
	CleanupInterval time.Duration `mapstructure:"cleanup_interval"`
	Global          LimitConfig   `mapstructure:"global"`
	Auth            LimitConfig   `mapstructure:"auth"`
}

// LimitConfig represents specific limits for a rate limiter instance.
type LimitConfig struct {
	Limit float64 `mapstructure:"limit"`
	Burst int     `mapstructure:"burst"`
}

// Auth contains JWT and token verification settings.
type Auth struct {
	AccessTokenTTL       time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL      time.Duration `mapstructure:"refresh_token_ttl"`
	EmailVerificationTTL time.Duration `mapstructure:"email_verification_ttl"`
	PasswordResetTTL     time.Duration `mapstructure:"password_reset_ttl"`
	Secret               string        `mapstructure:"-"`
}

// OAuth contains settings for third-party identity providers.
type OAuth struct {
	Google GoogleOAuth `mapstructure:"google"`
}

// GoogleOAuth holds Google App credentials.
type GoogleOAuth struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
}

// SMTP contains settings for the transactional email sender.
type SMTP struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	Username    string `mapstructure:"-"`
	Password    string `mapstructure:"-"`
	FromAddress string `mapstructure:"from_address"`
}

// MinIO contains S3-compatible storage settings.
type MinIO struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Bucket    string `mapstructure:"bucket"`
	UseSSL    bool   `mapstructure:"use_ssl"`
}

// Postgres contains PostgreSQL connection and pool settings.
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

// PoolConfig contains PostgreSQL connection pool settings.
type PoolConfig struct {
	MaxConns        int32         `mapstructure:"max_conns"`
	MinConns        int32         `mapstructure:"min_conns"`
	MaxConnLifetime time.Duration `mapstructure:"max_conn_lifetime"`
}

// Client contains frontend application URLs.
type Client struct {
	URL string `mapstructure:"url"`
}

// API contains the backend API URLs.
type API struct {
	URL string `mapstructure:"url"`
}

// Logging contains logging configuration.
type Logging struct {
	Level string `mapstructure:"level"`
}

// Load reads configuration from file, environment variables and defaults.
func Load() (*Config, error) {
	v := initViper()
	if err := readConfigFile(v); err != nil {
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("cannot unmarshal config: %w", err)
	}

	loadSensitiveValues(v, &cfg)

	if err := validateRequired(&cfg); err != nil {
		return nil, err
	}

	if err := validatePostgres(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func initViper() *viper.Viper {
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.AddConfigPath(".")

	// Defaults.
	v.SetDefault("env", "development")
	v.SetDefault("app.name", "RoleTalk-API")
	v.SetDefault("app.version", "1.0.0")

	v.SetDefault("http.port", ":8080")
	v.SetDefault("http.read_timeout", "10s")
	v.SetDefault("http.write_timeout", "10s")
	v.SetDefault("http.idle_timeout", "60s")
	v.SetDefault("http.shutdown_timeout", "8s")

	// Observability defaults
	v.SetDefault("observability.enabled", false)
	v.SetDefault("observability.otlp_endpoint", "localhost:4318")
	v.SetDefault("observability.loki_url", "http://localhost:3100/loki/api/v1/push")
	v.SetDefault("observability.sample_rate", 1.0)

	v.SetDefault("rate_limit.enabled", true)
	v.SetDefault("rate_limit.cleanup_interval", "3m")
	v.SetDefault("rate_limit.global.limit", 10.0)
	v.SetDefault("rate_limit.global.burst", 20)
	v.SetDefault("rate_limit.auth.limit", 1.0)
	v.SetDefault("rate_limit.auth.burst", 5)

	v.SetDefault("auth.access_token_ttl", "30m")
	v.SetDefault("auth.refresh_token_ttl", "168h")
	v.SetDefault("auth.email_verification_ttl", "24h")
	v.SetDefault("auth.password_reset_ttl", "1h")

	v.SetDefault("postgres.host", "localhost")
	v.SetDefault("postgres.port", 5432)
	v.SetDefault("postgres.ssl_mode", "disable")
	v.SetDefault("postgres.pool.max_conns", 15)
	v.SetDefault("postgres.pool.min_conns", 2)
	v.SetDefault("postgres.pool.max_conn_lifetime", "30m")

	v.SetDefault("logging.level", "info")

	// Environment variable settings.
	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	return v
}

func readConfigFile(v *viper.Viper) error {
	if err := v.ReadInConfig(); err != nil {
		var cfgNotFound viper.ConfigFileNotFoundError
		if !errors.As(err, &cfgNotFound) {
			return fmt.Errorf("cannot read config file: %w", err)
		}
	}
	return nil
}

func loadSensitiveValues(v *viper.Viper, cfg *Config) {
	cfg.Auth.Secret = v.GetString("jwt_secret")

	cfg.OAuth.Google.ClientID = v.GetString("google_client_id")
	cfg.OAuth.Google.ClientSecret = v.GetString("google_client_secret")

	cfg.SMTP.Username = v.GetString("smtp_username")
	cfg.SMTP.Password = v.GetString("smtp_password")

	cfg.Postgres.User = v.GetString("postgres_user")
	cfg.Postgres.Password = v.GetString("postgres_password")
	cfg.Postgres.Database = v.GetString("postgres_db")

	cfg.MinIO.AccessKey = v.GetString("minio_access_key")
	cfg.MinIO.SecretKey = v.GetString("minio_secret_key")

	cfg.Client.URL = v.GetString("client_url")
	cfg.API.URL = v.GetString("api_url")

	if dbURL := v.GetString("database_url"); dbURL != "" {
		cfg.Postgres.ConnectionURL = dbURL
	}
}

func validateRequired(cfg *Config) error {
	// Only validate strictly required secrets for production-like environments
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
		return errors.New("postgres: need either connection_url or host+database+user+password")
	}
	return nil
}
