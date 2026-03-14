package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env                 string `yaml:"env" env:"ENV" env-default:"local"`
	HTTPServer          `yaml:"http_server"`
	DB                  `yaml:"db"`
	Auth                `yaml:"auth"`
	Storage             `yaml:"storage"`
	GoogleSpreadsheetID string `yaml:"google_spreadsheet_id" env:"GOOGLE_SPREADSHEET_ID"`
}

type HTTPServer struct {
	Address string `yaml:"address" env:"HTTP_PORT" env-default:"8083"`
}

type DB struct {
	Host     string `yaml:"host" env:"POSTGRES_HOST" env-default:"localhost"`
	Port     string `yaml:"port" env:"POSTGRES_PORT" env-default:"5432"`
	User     string `yaml:"user" env:"POSTGRES_USER" env-default:"user"`
	Password string `yaml:"password" env:"POSTGRES_PASSWORD"`
	Name     string `yaml:"name" env:"POSTGRES_DB" env-default:"tires_shop"`
	SSLMode  string `yaml:"ssl_mode" env:"POSTGRES_SSLMODE" env-default:"disable"`
}

type Auth struct {
	JWTSecret        string        `env:"JWT_SECRET" env-required:"true"`
	TelegramBotToken string        `env:"TELEGRAM_BOT_TOKEN" env-required:"true"`
	TokenTTL         time.Duration `env:"JWT_TTL" env-default:"72h"`
}

type Storage struct {
	Endpoint   string `yaml:"endpoint" env:"MINIO_ENDPOINT" env-default:"localhost:9000"`
	AccessKey  string `yaml:"access_key" env:"MINIO_ACCESS_KEY" env-required:"true"`
	SecretKey  string `yaml:"secret_key" env:"MINIO_SECRET_KEY" env-required:"true"`
	BucketName string `yaml:"bucket_name" env:"MINIO_BUCKET_NAME" env-default:"tires-shop"`
	PublicURL  string `yaml:"public_url" env:"MINIO_PUBLIC_URL" env-default:"http://localhost:9000"`
	UseSSL     bool   `yaml:"use_ssl" env:"MINIO_USE_SSL" env-default:"false"`
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
