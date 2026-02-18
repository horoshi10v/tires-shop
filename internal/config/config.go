package config

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string `yaml:"env" env:"ENV" env-default:"local"`
	HTTPServer `yaml:"http_server"`
	DB         `yaml:"db"`
}

type HTTPServer struct {
	Address string `yaml:"address" env:"HTTP_PORT" env-default:"8080"`
}

type DB struct {
	Host     string `yaml:"host" env:"POSTGRES_HOST" env-default:"localhost"`
	Port     string `yaml:"port" env:"POSTGRES_PORT" env-default:"5432"`
	User     string `yaml:"user" env:"POSTGRES_USER" env-default:"user"`
	Password string `yaml:"password" env:"POSTGRES_PASSWORD"`
	Name     string `yaml:"name" env:"POSTGRES_DB" env-default:"tires_shop"`
}

func MustLoad() *Config {
	configPath := ".env"

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Printf("config file %s does not exist, reading from env variables", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			log.Fatalf("cannot read config: %s", err)
		}
	}

	return &cfg
}
