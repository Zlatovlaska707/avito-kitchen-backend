package config

import (
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	App         AppConfig         `yaml:"app"`
	HTTP        HTTPConfig        `yaml:"http"`
	Postgres    PostgresConfig    `yaml:"postgres"`
	Migrate     MigrateConfig     `yaml:"migrate"`
	Order       OrderConfig       `yaml:"order"`
	Cart        CartConfig        `yaml:"cart"`
	Idempotency IdempotencyConfig `yaml:"idempotency"`
	MenuSync    MenuSyncConfig    `yaml:"menu_sync"`
	Webhook     WebhookConfig     `yaml:"webhook"`
	Outbox      OutboxConfig      `yaml:"outbox"`
	Workers     WorkersConfig     `yaml:"workers"`
	Rate        RateConfig        `yaml:"rate"`
	Delivery    DeliveryConfig    `yaml:"delivery"`
	Log         LogConfig         `yaml:"log"`
	OpenAPI     OpenAPIConfig     `yaml:"openapi"`
	Swagger     SwaggerConfig     `yaml:"swagger"`

	PostgresDSN   string `env:"POSTGRES_DSN" env-required:"true"`
	AdminAPIToken string `env:"ADMIN_API_TOKEN" env-required:"true"`
}

type AppConfig struct {
	Env             string `yaml:"env" env:"APP_ENV" env-default:"local"`
	DefaultTimezone string `yaml:"default_timezone" env:"DEFAULT_TIMEZONE" env-default:"Europe/Moscow"`
}

type HTTPConfig struct {
	Address         string        `yaml:"address" env:"HTTP_ADDRESS" env-default:":8080"`
	ReadTimeout     time.Duration `yaml:"read_timeout" env:"HTTP_READ_TIMEOUT" env-default:"5s"`
	WriteTimeout    time.Duration `yaml:"write_timeout" env:"HTTP_WRITE_TIMEOUT" env-default:"10s"`
	IdleTimeout     time.Duration `yaml:"idle_timeout" env:"HTTP_IDLE_TIMEOUT" env-default:"60s"`
	HandlerTimeout  time.Duration `yaml:"handler_timeout" env:"HTTP_HANDLER_TIMEOUT" env-default:"5s"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env:"HTTP_SHUTDOWN_TIMEOUT" env-default:"30s"`
	MaxBodyBytes    int64         `yaml:"max_body_bytes" env:"HTTP_MAX_BODY_BYTES" env-default:"1048576"`
	MaxBodyPublic   int64         `yaml:"max_body_bytes_public" env:"HTTP_MAX_BODY_BYTES_PUBLIC" env-default:"262144"`
}

type PostgresConfig struct {
	MaxConns        int32         `yaml:"max_conns" env:"POSTGRES_MAX_CONNS" env-default:"20"`
	MinConns        int32         `yaml:"min_conns" env:"POSTGRES_MIN_CONNS" env-default:"2"`
	MaxConnLifetime time.Duration `yaml:"max_conn_lifetime" env:"POSTGRES_MAX_CONN_LIFETIME" env-default:"1h"`
	MaxConnIdleTime time.Duration `yaml:"max_conn_idle_time" env:"POSTGRES_MAX_CONN_IDLE_TIME" env-default:"15m"`
	QueryTimeout    time.Duration `yaml:"query_timeout" env:"POSTGRES_QUERY_TIMEOUT" env-default:"3s"`
	WorkerMaxConns  int32         `yaml:"worker_max_conns" env:"POSTGRES_WORKER_MAX_CONNS" env-default:"5"`
}

type MigrateConfig struct {
	AutoRun bool   `yaml:"auto_run" env:"MIGRATIONS_AUTO_RUN" env-default:"false"`
	Dir     string `yaml:"dir" env:"MIGRATIONS_DIR" env-default:"./migrations"`
}

type OrderConfig struct {
	AcceptTimeout time.Duration `yaml:"accept_timeout" env:"ORDER_ACCEPT_TIMEOUT" env-default:"10m"`
	CancelGrace   time.Duration `yaml:"cancel_grace" env:"ORDER_CANCEL_GRACE" env-default:"120s"`
}

type CartConfig struct {
	TTL      time.Duration `yaml:"ttl" env:"CART_TTL" env-default:"24h"`
	MaxItems int           `yaml:"max_items" env:"CART_MAX_ITEMS" env-default:"100"`
}

type IdempotencyConfig struct {
	TTL time.Duration `yaml:"ttl" env:"IDEMPOTENCY_TTL" env-default:"24h"`
}

type MenuSyncConfig struct {
	MaxItems      int `yaml:"max_items" env:"MENU_SYNC_MAX_ITEMS" env-default:"2000"`
	MaxCategories int `yaml:"max_categories" env:"MENU_SYNC_MAX_CATEGORIES" env-default:"200"`
}

type WebhookConfig struct {
	Timeout         time.Duration `yaml:"timeout" env:"WEBHOOK_TIMEOUT" env-default:"3s"`
	MaxAttempts     int           `yaml:"max_attempts" env:"WEBHOOK_MAX_ATTEMPTS" env-default:"5"`
	AllowPrivateIPs bool          `yaml:"allow_private_ips" env:"WEBHOOK_ALLOW_PRIVATE_IPS" env-default:"false"`
}

type OutboxConfig struct {
	PollInterval time.Duration `yaml:"poll_interval" env:"OUTBOX_POLL_INTERVAL" env-default:"1s"`
	BatchSize    int           `yaml:"batch_size" env:"OUTBOX_BATCH_SIZE" env-default:"50"`
}

type WorkersConfig struct {
	Enabled bool `yaml:"enabled" env:"WORKERS_ENABLED" env-default:"true"`
}

type RateConfig struct {
	RPSPublic      int `yaml:"rps_public" env:"RATE_LIMIT_RPS_PUBLIC" env-default:"100"`
	RPSPublicWrite int `yaml:"rps_public_write" env:"RATE_LIMIT_RPS_PUBLIC_WRITE" env-default:"20"`
	RPSPartner     int `yaml:"rps_partner" env:"RATE_LIMIT_RPS_PARTNER" env-default:"50"`
}

type DeliveryConfig struct {
	SimulationEnabled bool          `yaml:"simulation_enabled" env:"DELIVERY_SIMULATION_ENABLED" env-default:"true"`
	AssignDelay       time.Duration `yaml:"assign_delay" env:"DELIVERY_ASSIGN_DELAY" env-default:"30s"`
	TransitDelay      time.Duration `yaml:"transit_delay" env:"DELIVERY_TRANSIT_DELAY" env-default:"60s"`
	BaseMinutes       int           `yaml:"base_minutes" env:"DELIVERY_BASE_MINUTES" env-default:"10"`
	MinutesPerKm      float64       `yaml:"minutes_per_km" env:"DELIVERY_MINUTES_PER_KM" env-default:"3.0"`
}

type LogConfig struct {
	Level  string `yaml:"level" env:"LOG_LEVEL" env-default:"info"`
	Format string `yaml:"format" env:"LOG_FORMAT" env-default:"json"`
}

type OpenAPIConfig struct {
	ValidateRequests  bool `yaml:"validate_requests" env:"OPENAPI_VALIDATE_REQUESTS" env-default:"true"`
	ValidateResponses bool `yaml:"validate_responses" env:"OPENAPI_VALIDATE_RESPONSES" env-default:"false"`
}

type SwaggerConfig struct {
	Enabled bool `yaml:"enabled" env:"SWAGGER_ENABLED" env-default:"true"`
}

func Load(path string) (Config, error) {
	var cfg Config
	if path != "" {
		if err := cleanenv.ReadConfig(path, &cfg); err != nil {
			return Config{}, fmt.Errorf("read config %q: %w", path, err)
		}
	}
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return Config{}, fmt.Errorf("read env: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func MustLoad(path string) Config {
	cfg, err := Load(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}
	return cfg
}

// MigratorConfig holds only settings required by cmd/migrator.
type MigratorConfig struct {
	PostgresDSN string        `env:"POSTGRES_DSN" env-required:"true"`
	Migrate     MigrateConfig `yaml:"migrate"`
}

func LoadMigrator(path string) (MigratorConfig, error) {
	var cfg MigratorConfig
	if path != "" {
		if err := cleanenv.ReadConfig(path, &cfg); err != nil {
			return MigratorConfig{}, fmt.Errorf("read config %q: %w", path, err)
		}
	}
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return MigratorConfig{}, fmt.Errorf("read env: %w", err)
	}
	if cfg.PostgresDSN == "" {
		return MigratorConfig{}, fmt.Errorf("POSTGRES_DSN must not be empty")
	}
	if cfg.Migrate.Dir == "" {
		cfg.Migrate.Dir = "./migrations"
	}
	return cfg, nil
}

func MustLoadMigrator(path string) MigratorConfig {
	cfg, err := LoadMigrator(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}
	return cfg
}

func (c Config) Validate() error {
	if c.PostgresDSN == "" {
		return fmt.Errorf("POSTGRES_DSN must not be empty")
	}
	if len(c.AdminAPIToken) < 16 {
		return fmt.Errorf("ADMIN_API_TOKEN must be at least 16 characters")
	}
	if _, err := time.LoadLocation(c.App.DefaultTimezone); err != nil {
		return fmt.Errorf("DEFAULT_TIMEZONE: %w", err)
	}
	if c.HTTP.ReadTimeout <= 0 || c.HTTP.WriteTimeout <= 0 || c.HTTP.IdleTimeout <= 0 ||
		c.HTTP.HandlerTimeout <= 0 || c.HTTP.ShutdownTimeout <= 0 {
		return fmt.Errorf("HTTP timeouts must be positive")
	}
	if c.Postgres.MaxConnLifetime <= 0 || c.Postgres.MaxConnIdleTime <= 0 || c.Postgres.QueryTimeout <= 0 {
		return fmt.Errorf("postgres durations must be positive")
	}
	if c.Order.AcceptTimeout <= 0 || c.Order.CancelGrace <= 0 {
		return fmt.Errorf("order durations must be positive")
	}
	if c.Cart.TTL <= 0 || c.Idempotency.TTL <= 0 {
		return fmt.Errorf("cart and idempotency TTL must be positive")
	}
	if c.Webhook.Timeout <= 0 || c.Outbox.PollInterval <= 0 {
		return fmt.Errorf("webhook and outbox durations must be positive")
	}
	if c.Delivery.AssignDelay <= 0 || c.Delivery.TransitDelay <= 0 {
		return fmt.Errorf("delivery delays must be positive")
	}
	if c.Postgres.MaxConns < c.Postgres.MinConns {
		return fmt.Errorf("POSTGRES_MAX_CONNS must be >= POSTGRES_MIN_CONNS")
	}
	return nil
}

func (c Config) DefaultLocation() *time.Location {
	loc, err := time.LoadLocation(c.App.DefaultTimezone)
	if err != nil {
		loc, _ = time.LoadLocation("Europe/Moscow")
	}
	return loc
}
