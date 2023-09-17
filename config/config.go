package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"time"
)

type Config struct {
	Env              string `yaml:"env" env-required:"true"`
	StoragePath      string `yaml:"storage_path" env-required:"true"`
	ConnectionString string `yaml:"-"`
	Postgres         `yaml:"db"`
	Kafka            `yaml:"kafka"`
	HTTPServer       `yaml:"http_server"`
	Redis            `yaml:"redis"`
	Log              `yaml:"log"`
}

type Postgres struct {
	Host        string        `yaml:"host" env:"DB_HOST" env-required:"true"`
	DBPort      int           `yaml:"port" env:"DB_PORT" env-required:"true"`
	User        string        `env:"DB_USER" env-required:"true"`
	Password    string        `env:"DB_PASSWORD" env-required:"true"`
	DBName      string        `env:"DB_NAME" env-required:"true"`
	SSLMode     string        `yaml:"sslmode" env:"DB_SSLMODE" env-default:"disable"`
	Timeout     time.Duration `yaml:"timeout" env:"DB_TIMEOUT" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env:"DB_IDLE_TIMEOUT" env-default:"60s"`
	MaxPoolSize int           `yaml:"max_pool_size" env-required:"true"  env:"PG_MAX_POOL_SIZE"`
}

type Kafka struct {
	Brokers  []string `yaml:"brokers" env:"KAFKA_BROKERS" env-required:"true"`
	GroupID  string   `yaml:"group_id" env:"KAFKA_GROUP_ID" env-required:"true"`
	Topic    string   `yaml:"topic" env:"KAFKA_TOPIC" env-required:"true"`
	MinBytes int      `yaml:"min_bytes" env:"KAFKA_MIN_BYTES" env-default:"10000"`
	MaxBytes int      `yaml:"max_bytes" env:"KAFKA_MAX_BYTES" env-default:"10000000"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
	Port        string        `yaml:"server_port" env:"PORT" env-required:"true"`
}

type Redis struct {
	Addr     string `yaml:"address"`
	Password string `env:"REDIS_PASS" env-required:"true"`
}

type Log struct {
	Level string `env-required:"true" yaml:"level" env:"LOG_LEVEL"`
}

func LoadConfig() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatal("config file does not exist: %s", configPath)
	}
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatal("cannot read config: %s", err)
	}
	cfg.ConnectionString = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.DBPort, cfg.User, cfg.Postgres.Password, cfg.DBName, cfg.SSLMode)
	return &cfg
}
