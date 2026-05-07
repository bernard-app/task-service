package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env        string         `yaml:"env" env-default:"local"`
	HTTPServer HTTPServer     `yaml:"http_server"`
	Postgres   PostgresConfig `yaml:"postgres"`
	Redis      RedisConfig    `yaml:"redis"`
}

type RedisConfig struct {
	Addr string `yaml:"addr"`
	Password string `yaml:"password"`
	DB int `yaml:"db"`
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

	cfg.Postgres.Addr = os.Getenv("POSTGRES_URL")

	cfg.Redis.Password = os.Getenv("REDIS_PASSWORD")
	cfg.Redis.DB, err = strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		log.Fatal("Redis config db value error")
	}
	
	return &cfg
}
