package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string     `yaml:"env"`
	HttpServer httpServer `yaml:"http-server"`
	Postgres   postgres   `yaml:"postgres"`
	Redis      redis      `yaml:"redis"`
	Jwt        jwt        `yaml:"jwt"`
}

type httpServer struct {
	Address     string `yaml:"address"`
	Timeout     string `yaml:"timeout"`
	IdleTimeout string `yaml:"idle-timeout"`
}

type jwt struct {
	SecretKey string        `yaml:"secret-key"`
	TokenTTL  time.Duration `yaml:"token-ttl"`
}

type postgres struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DbName   string `yaml:"db-name"`
	SslMode  string `yaml:"ssl-mode"`
}

type redis struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Password string `yaml:"password"`
	DbName   int    `yaml:"db-name"`
}

func MustLoad() *Config {
	var cfg Config

	configPath := os.Getenv("CONFIG_PATH")

	if configPath == "" {
		panic("CONFIG_PATH isn't set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Panicf("config file doesn't exist: %s", configPath)
	}

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Panicf("can't read config: %s", err)
	}

	return &cfg
}
