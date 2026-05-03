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

// ErrMissingEnvVar is returned when a required environment variable is not set.
var ErrMissingEnvVar = errors.New("missing required environment variable")

// Config holds the entire application configuration.
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
	AI            AI            `mapstructure:"ai"`
	Redis         Redis         `mapstructure:"redis"`
	RabbitMQ      RabbitMQ      `mapstructure:"rabbitmq"`
	MinIO         MinIO         `mapstructure:"minio"`
	Ollama        Ollama        `mapstructure:"ollama"`
	Client        Client        `mapstructure:"client"`
	API           API           `mapstructure:"api"`
	Logging       Logging       `mapstructure:"logging"`
}

// App contains basic application metadata.
type App struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
}

// HTTP contains server timeout and port settings.
type HTTP struct {
	Port            string        `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

// Ollama contains settings for the local Large Language Model provider.
type Ollama struct {
	URL   string `mapstructure:"url"`
	Model string `mapstructure:"model"`
}

// Observability contains settings for tracing and logging.
type Observability struct {
	Enabled      bool    `mapstructure:"enabled"`
	OTLPEndpoint string  `mapstructure:"otlp_endpoint"`
	LokiURL      string  `mapstructure:"loki_url"`
	SampleRate   float64 `mapstructure:"sample_rate"`
}

// AI contains the address of the AI microservice.
type AI struct {
	Addr string `mapstructure:"addr"`
}

// RateLimit defines the global and auth-specific rate limiting rules.
type RateLimit struct {
	Enabled         bool          `mapstructure:"enabled"`
	CleanupInterval time.Duration `mapstructure:"cleanup_interval"`
	Global          LimitConfig   `mapstructure:"global"`
	Auth            LimitConfig   `mapstructure:"auth"`
}

// LimitConfig defines the burst and rate for a single limiter.
type LimitConfig struct {
	Limit float64 `mapstructure:"limit"`
	Burst int     `mapstructure:"burst"`
}

// Auth contains JWT and token lifespan settings.
type Auth struct {
	AccessTokenTTL       time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL      time.Duration `mapstructure:"refresh_token_ttl"`
	EmailVerificationTTL time.Duration `mapstructure:"email_verification_ttl"`
	PasswordResetTTL     time.Duration `mapstructure:"password_reset_ttl"`
	Secret               string        `mapstructure:"-"`
}

// OAuth contains third-party authentication provider settings.
type OAuth struct {
	Google GoogleOAuth `mapstructure:"google"`
}

// GoogleOAuth contains the client ID and secret for Google Login.
type GoogleOAuth struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
}

// SMTP contains settings for sending transactional emails.
type SMTP struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	Username    string `mapstructure:"-"`
	Password    string `mapstructure:"-"`
	FromAddress string `mapstructure:"from_address"`
}

// MinIO contains S3-compatible storage settings.
type MinIO struct {
	Endpoint       string `mapstructure:"endpoint"`
	PublicEndpoint string `mapstructure:"public_endpoint"`
	AccessKey      string `mapstructure:"access_key"`
	SecretKey      string `mapstructure:"secret_key"`
	Bucket         string `mapstructure:"bucket"`
	UseSSL         bool   `mapstructure:"use_ssl"`
	// Region is the signing region sent to MinIO (e.g. us-east-1). When set, the client skips bucket location HTTP probes.
	Region string `mapstructure:"region"`
}

// Postgres contains database connection and pooling settings.
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

// PoolConfig defines the PostgreSQL connection pool behavior.
type PoolConfig struct {
	MaxConns        int32         `mapstructure:"max_conns"`
	MinConns        int32         `mapstructure:"min_conns"`
	MaxConnLifetime time.Duration `mapstructure:"max_conn_lifetime"`
}

// Redis contains Redis connection settings.
type Redis struct {
	Addr         string        `mapstructure:"addr"`
	Password     string        `mapstructure:"-"`
	DB           int           `mapstructure:"db"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	Pool         RedisPool     `mapstructure:"pool"`
}

// RedisPool defines the Redis connection pool size and retries.
type RedisPool struct {
	PoolSize     int `mapstructure:"pool_size"`
	MinIdleConns int `mapstructure:"min_idle_conns"`
	MaxRetries   int `mapstructure:"max_retries"`
}

// RabbitMQ contains RabbitMQ connection and queue settings.
type RabbitMQ struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"-"`
	Password     string `mapstructure:"-"`
	QueueName    string `mapstructure:"queue_name"`
	DLXName      string `mapstructure:"dlx_name"`
	DLQName      string `mapstructure:"dlq_name"`
	ExchangeName string `mapstructure:"exchange_name"`
}

// Client contains settings for the frontend application.
type Client struct {
	URL string `mapstructure:"url"`
}

// API contains the public URL of the backend API.
type API struct {
	URL string `mapstructure:"url"`
}

// Logging defines the global log level.
type Logging struct {
	Level string `mapstructure:"level"`
}

// Load initializes the configuration by merging YAML files and Environment Variables.
func Load() (*Config, error) {
	v := viper.New()

	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.SetConfigFile(".env")
	v.AddConfigPath(".")
	v.AddConfigPath("./..")
	v.AddConfigPath("./../..")

	if err := v.ReadInConfig(); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			fmt.Printf("Note: .env file not found, proceeding with ENV and YAML\n")
		}
	}

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.AddConfigPath(".")
	if err := v.MergeInConfig(); err != nil {
		return nil, fmt.Errorf("merge config.yml: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal configuration: %w", err)
	}

	loadSensitiveValues(v, &cfg)

	if err := validateRequired(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// loadSensitiveValues explicitly binds secret keys from environment variables to the config struct.
func loadSensitiveValues(v *viper.Viper, cfg *Config) {
	cfg.Auth.Secret = v.GetString("JWT_SECRET")

	cfg.OAuth.Google.ClientID = v.GetString("OAUTH_GOOGLE_CLIENT_ID")
	cfg.OAuth.Google.ClientSecret = v.GetString("OAUTH_GOOGLE_CLIENT_SECRET")

	cfg.SMTP.Username = v.GetString("SMTP_USERNAME")
	cfg.SMTP.Password = v.GetString("SMTP_PASSWORD")

	cfg.Postgres.User = v.GetString("POSTGRES_USER")
	cfg.Postgres.Password = v.GetString("POSTGRES_PASSWORD")
	cfg.Postgres.Database = v.GetString("POSTGRES_DB")

	if pe := strings.TrimSpace(v.GetString("MINIO_PUBLIC_ENDPOINT")); pe != "" {
		cfg.MinIO.PublicEndpoint = pe
	}
	cfg.MinIO.AccessKey = v.GetString("MINIO_ACCESS_KEY")
	cfg.MinIO.SecretKey = v.GetString("MINIO_SECRET_KEY")

	cfg.Redis.Password = v.GetString("REDIS_PASSWORD")

	cfg.RabbitMQ.User = v.GetString("RABBITMQ_USER")
	cfg.RabbitMQ.Password = v.GetString("RABBITMQ_PASSWORD")

	cfg.AI.Addr = v.GetString("AI_SERVICE_ADDR")
	cfg.Ollama.URL = v.GetString("OLLAMA_URL")
	cfg.Ollama.Model = v.GetString("OLLAMA_MODEL")

	cfg.Client.URL = v.GetString("CLIENT_URL")
	cfg.API.URL = v.GetString("API_URL")

	if dbURL := v.GetString("DATABASE_URL"); dbURL != "" {
		cfg.Postgres.ConnectionURL = dbURL
	}
}

// validateRequired ensures the application does not start with missing security or infrastructure secrets.
func validateRequired(cfg *Config) error {
	required := map[string]string{
		"JWT Secret":           cfg.Auth.Secret,
		"Postgres Password":    cfg.Postgres.Password,
		"MinIO Access Key":     cfg.MinIO.AccessKey,
		"MinIO Secret Key":     cfg.MinIO.SecretKey,
		"AI Service Addr":      cfg.AI.Addr,
		"RabbitMQ Password":    cfg.RabbitMQ.Password,
		"Google Client Secret": cfg.OAuth.Google.ClientSecret,
	}

	for name, val := range required {
		if val == "" {
			return fmt.Errorf("%w: %s", ErrMissingEnvVar, name)
		}
	}
	return nil
}
