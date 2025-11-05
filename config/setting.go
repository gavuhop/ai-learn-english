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

type serverConfig struct {
	Port        int    `koanf:"port"`
	Mode        string `koanf:"mode"`
	Concurrency int    `koanf:"concurrency"`
	BodyLimit   int    `koanf:"body_limit"`
	AppName     string `koanf:"app_name"`
}

type logLevel string

const (
	Debug logLevel = "debug"
	Info  logLevel = "info"
	Warn  logLevel = "warn"
	Error logLevel = "error"
	Fatal logLevel = "fatal"
	Panic logLevel = "panic"
)

type databaseConfig struct {
	Host         string `koanf:"host"`
	Port         int    `koanf:"port"`
	User         string `koanf:"user"`
	Password     string `koanf:"password"`
	Name         string `koanf:"name"`
	MaxIdleConns int    `koanf:"max_idle_conns"`
	MaxOpenConns int    `koanf:"max_open_conns"`
	MaxLifetime  int    `koanf:"max_lifetime"`
}

type openaiConfig struct {
	Key   string `koanf:"key"`
	Model string `koanf:"model"`
	EmbeddingModel string `koanf:"embedding_model"`
}

type geminiConfig struct {
	Key   string `koanf:"key"`
	Model string `koanf:"model"`
}

type corsConfig struct {
	AllowOrigins []string `koanf:"allow_origins"`
	AllowMethods []string `koanf:"allow_methods"`
	AllowHeaders []string `koanf:"allow_headers"`
}

type milvusConfig struct {
	Address    string `koanf:"address"`
	Collection string `koanf:"collection"`
}

type config struct {
	Server   serverConfig   `koanf:"server"`
	Database databaseConfig `koanf:"database"`
	OpenAI   openaiConfig   `koanf:"openai"`
	Gemini   geminiConfig   `koanf:"gemini"`
	LogLevel logLevel       `koanf:"log_level"`
	Dns      string         `koanf:"dns"`
	S3       s3Config       `koanf:"s3"`
	Cors     corsConfig     `koanf:"cors"`
	Milvus   milvusConfig   `koanf:"milvus"`
	Ingest   ingestConfig   `koanf:"ingest"`
}

type s3Config struct {
	Endpoint  string `koanf:"endpoint"`
	AccessKey string `koanf:"access_key"`
	SecretKey string `koanf:"secret_key"`
	Region    string `koanf:"region"`
	UseSSL    bool   `koanf:"use_ssl"`
	Bucket    string `koanf:"bucket"`
}

type ingestConfig struct {
	ChunkTokens  int `koanf:"chunk_tokens"`
	ChunkOverlap int `koanf:"chunk_overlap"`
}

func buildMySQLDSN(cfg databaseConfig) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
	)
}

var defaultConfig = config{
	Server: serverConfig{
		Port: 8000,
		Mode: "release",
	},
	Database: databaseConfig{
		Host:     "127.0.0.1",
		Port:     5432,
		User:     "root",
		Password: "",
		Name:     "testdb",
	},
	OpenAI: openaiConfig{
		Key:   "",
		Model: "default",
	},
	Gemini: geminiConfig{
		Key:   "",
		Model: "default",
	},
	LogLevel: Info,
	S3: s3Config{
		Endpoint:  "http://localhost:9000",
		AccessKey: "minioadmin",
		SecretKey: "minioadmin",
		Region:    "us-east-1",
		UseSSL:    false,
		Bucket:    "uploads",
	},
	Milvus: milvusConfig{
		Address:    "localhost:19530",
		Collection: "chunks",
	},
	Ingest: ingestConfig{
		ChunkTokens:  600,
		ChunkOverlap: 80,
	},
}

var (
	Cfg  = defaultConfig
	once sync.Once
)

func init() {
	path := "config.yaml"

	once.Do(func() {
		k := koanf.New(".")

		// defaults
		Cfg = defaultConfig

		// file
		if e := k.Load(file.Provider(path), yaml.Parser()); e != nil && !os.IsNotExist(e) {
			return
		}

		// env APP_SERVER_PORT
		if e := k.Load(env.Provider("APP_", ".", func(s string) string {
			return strings.ToLower(strings.TrimPrefix(s, "APP_"))
		}), nil); e != nil {
			return
		}

		// bind
		if e := k.Unmarshal("", &Cfg); e != nil {
			return
		}

		if Cfg.Dns == "" {
			Cfg.Dns = buildMySQLDSN(Cfg.Database)
		}
	})

}
