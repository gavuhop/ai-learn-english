package config

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3/log"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type serverConfig struct {
	Port        int    `koanf:"port" validate:"required"`
	Mode        string `koanf:"mode" validate:"required"`
	Concurrency int    `koanf:"concurrency" validate:"required"`
	BodyLimit   int    `koanf:"body_limit" validate:"required"`
	AppName     string `koanf:"app_name" validate:"required"`
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

type Module string

const (
	ModuleMilvus   Module = "milvus"
	ModuleIngest   Module = "ingest"
	ModuleDatabase Module = "database"
	ModuleOpenAI   Module = "openai"
	ModuleGemini   Module = "gemini"
	ModuleS3       Module = "s3"
	ModuleCors     Module = "cors"
	ModuleServer   Module = "server"
	ModuleSetting  Module = "setting"
	ModuleUpload   Module = "upload"
	ModuleRetriever Module = "retriever"
)

type databaseConfig struct {
	Host         string `koanf:"host" validate:"required"`
	Port         int    `koanf:"port" validate:"required"`
	User         string `koanf:"user" validate:"required"`
	Password     string `koanf:"password" validate:"required"`
	Name         string `koanf:"name" validate:"required"`
	MaxIdleConns int    `koanf:"max_idle_conns" validate:"required"`
	MaxOpenConns int    `koanf:"max_open_conns" validate:"required"`
	MaxLifetime  int    `koanf:"max_lifetime" validate:"required"`
}

type openaiConfig struct {
	Key            string `koanf:"key" validate:"required"`
	Model          string `koanf:"model" validate:"required"`
	EmbeddingModel string `koanf:"embedding_model" validate:"required"`
}

type geminiConfig struct {
	Key   string `koanf:"key" validate:"required"`
	Model string `koanf:"model" validate:"required"`
}

type corsConfig struct {
	AllowOrigins []string `koanf:"allow_origins" validate:"required"`
	AllowMethods []string `koanf:"allow_methods" validate:"required"`
	AllowHeaders []string `koanf:"allow_headers" validate:"required"`
}

type milvusConfig struct {
	Address         string          `koanf:"address" validate:"required"`
	Collection      string          `koanf:"collection" validate:"required"`
	IndexHNSWConfig indexHNSWConfig `koanf:"index_hnsw_config"`
}

type indexHNSWConfig struct {
	MetricType     string `koanf:"metric_type" validate:"required"`
	M              int    `koanf:"m" validate:"required"`
	EfConstruction int    `koanf:"ef_construction" validate:"required"`
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
	Endpoint  string `koanf:"endpoint" validate:"required"`
	AccessKey string `koanf:"access_key" validate:"required"`
	SecretKey string `koanf:"secret_key" validate:"required"`
	Region    string `koanf:"region" validate:"required"`
	UseSSL    bool   `koanf:"use_ssl" validate:"required"`
	Bucket    string `koanf:"bucket" validate:"required"`
}

type ingestConfig struct {
	ChunkTokens  int `koanf:"chunk_tokens" validate:"required"`
	ChunkOverlap int `koanf:"chunk_overlap" validate:"required"`
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

		validate := validator.New()
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
			log.Errorf("failed to unmarshal config: %v", e)
		}

		if Cfg.Dns == "" {
			Cfg.Dns = buildMySQLDSN(Cfg.Database)
		}

		// validate config
		if err := validate.Struct(Cfg); err != nil {
			if errs, ok := err.(validator.ValidationErrors); ok {
				var sb strings.Builder
				sb.WriteString(fmt.Sprintf("%v Config validation failed:\n", ModuleSetting))
	
				for _, e := range errs {
					sb.WriteString(
						fmt.Sprintf("  • %s: failed '%s' (value: %v)\n", e.Field(), e.Tag(), e.Value()),
					)
				}
	
				log.Error(sb.String())
			} else {
				log.Errorf("config validation failed: %v", err)
			}
		}
	})

}
