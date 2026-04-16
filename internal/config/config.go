package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env        string         `yaml:"env" env-default:"local"`
	HTTPServer HTTPServer     `yaml:"http_server"`
	Postgres   PostgresConfig `yaml:"postgres"`
}

type PostgresConfig struct {
	Addr string `yaml:"addr"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8081"`
	Timeout     time.Duration `yaml:"timeout"`
	IdleTimeout time.Duration `yaml:"idle_timeout"`
	Retry       int           `yaml:"retry" env-default:"5"`
}

func MustLoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG env variable not set")
	}

	var cfg Config

	err = cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		log.Fatal("cannot find the config: ", err)
	}

	cfg.Postgres.Addr = os.Getenv("PG_ADDR")

	return &cfg
}
