package main

import (
	"ai-learn-english/config"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func main() {
	if err := config.Init("config.yaml"); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	dsn := config.Cfg.Dns

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	g := gen.NewGenerator(gen.Config{
		OutPath:        "internal/database/query",
		ModelPkgPath:   "internal/database/model",
		Mode:           gen.WithDefaultQuery | gen.WithQueryInterface,
		FieldNullable:  true,
		FieldCoverable: true,
	})

	g.UseDB(db)

	// Generate models for all tables found in the connected database
	g.ApplyBasic(g.GenerateAllTable()...)

	g.Execute()
}
