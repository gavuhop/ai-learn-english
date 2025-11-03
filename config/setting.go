package config

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type ServerConfig struct {
	Port int    `koanf:"port"`
	Mode string `koanf:"mode"`
}

type LogLevel string

const (
	DEBUG LogLevel = "debug"
	INFO  LogLevel = "info"
	WARN  LogLevel = "warn"
	ERROR LogLevel = "error"
	FATAL LogLevel = "fatal"
	PANIC LogLevel = "panic"
)

type DatabaseConfig struct {
	Host     string `koanf:"host"`
	Port     int    `koanf:"port"`
	User     string `koanf:"user"`
	Password string `koanf:"password"`
	Name     string `koanf:"name"`
}

type OpenAIConfig struct {
	Key   string `koanf:"key"`
	Model string `koanf:"model"`
}

type GeminiConfig struct {
	Key   string `koanf:"key"`
	Model string `koanf:"model"`
}

type Config struct {
	Server   ServerConfig   `koanf:"server"`
	Database DatabaseConfig `koanf:"database"`
	OpenAI   OpenAIConfig   `koanf:"openai"`
	Gemini   GeminiConfig   `koanf:"gemini"`
	LogLevel LogLevel       `koanf:"log_level"`
	Dns      string         `koanf:"dns"`
}

func buildMySQLDSN(cfg DatabaseConfig) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
	)
}

var defaultConfig = Config{
	Server: ServerConfig{
		Port: 8000,
		Mode: "release",
	},
	Database: DatabaseConfig{
		Host:     "127.0.0.1",
		Port:     5432,
		User:     "root",
		Password: "",
		Name:     "testdb",
	},
	OpenAI: OpenAIConfig{
		Key:   "",
		Model: "default",
	},
	Gemini: GeminiConfig{
		Key:   "",
		Model: "default",
	},
	LogLevel: INFO,
}

var (
	Cfg  = defaultConfig
	once sync.Once
)

func Init(path string) error {
	var err error

	once.Do(func() {
		k := koanf.New(".")

		// defaults
		Cfg = defaultConfig

		// file
		if e := k.Load(file.Provider(path), yaml.Parser()); e != nil && !os.IsNotExist(e) {
			err = e
			return
		}

		// env APP_SERVER_PORT
		if e := k.Load(env.Provider("APP_", ".", func(s string) string {
			return strings.ToLower(strings.TrimPrefix(s, "APP_"))
		}), nil); e != nil {
			err = e
			return
		}

		// bind
		if e := k.Unmarshal("", &Cfg); e != nil {
			err = e
			return
		}

		if Cfg.Dns == "" {
			Cfg.Dns = buildMySQLDSN(Cfg.Database)
		}
	})

	return err
}
