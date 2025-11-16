package application

import (
	"flag"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"github.com/zxchelik/avito-test-task/pkg/logger"
	"log"
	"os"
	"time"
)

type Config struct {
	Env      logger.EnvString `yaml:"env" env-default:"local" env-required:"true"`
	Postgres `yaml:"postgres"`
	Server   `yaml:"server"`
}

type Postgres struct {
	Host              string        `yaml:"host" env:"POSTGRES_HOST" env-required:"true"`
	Port              string        `yaml:"port" env:"POSTGRES_PORT" env-required:"true"`
	Username          string        `yaml:"user" env:"POSTGRES_USER" env-required:"true"`
	Password          string        `yaml:"password" env:"POSTGRES_PASSWORD" env-required:"true"`
	Database          string        `yaml:"database" env:"POSTGRES_DB" env-required:"true"`
	MaxConns          int32         `yaml:"maxConns" env-required:"true"`
	MinConns          int32         `yaml:"minConns" env-required:"true"`
	MaxConnLifetime   time.Duration `yaml:"maxConnLifetime" env-required:"true"`
	MaxConnIdleTime   time.Duration `yaml:"maxConnIdleTime" env-required:"true"`
	HealthCheckPeriod time.Duration `yaml:"healthCheckPeriod" env-required:"true"`
}

func (p *Postgres) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", p.Username, p.Password, p.Host, p.Port, p.Database)
}

type Server struct {
	Host            string        `yaml:"host" env:"HOST" env-required:"true"`
	Port            int           `yaml:"port" env:"PORT"  env-required:"true"`
	Timeout         time.Duration `yaml:"timeout"`
	IdleTimeout     time.Duration `yaml:"idle_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

func (s *Server) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

func MustLoad() *Config {
	_ = godotenv.Load()

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/server/default.yaml"
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("CONFIG_PATH does not exist: %s", configPath)
	}

	var config Config

	// Чтение YAML и ENV
	if err := cleanenv.ReadConfig(configPath, &config); err != nil {
		log.Fatalf("Error loading config: %s", err)
	}
	_ = cleanenv.ReadEnv(&config)

	// Парсинг CLI флагов
	portFlag := flag.Int("port", config.Server.Port, "HTTP server port")
	hostFlag := flag.String("host", config.Server.Host, "HTTP server host")
	flag.Parse()

	config.Server.Port = *portFlag
	config.Server.Host = *hostFlag

	return &config
}
