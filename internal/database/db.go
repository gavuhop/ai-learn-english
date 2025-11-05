package database

import (
	"ai-learn-english/config"
	"ai-learn-english/pkg/logger"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// connect opens the DB and applies pool configuration
func connect() (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(config.Cfg.Dns), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(config.Cfg.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.Cfg.Database.MaxOpenConns)
	lifetime := time.Duration(config.Cfg.Database.MaxLifetime) * time.Minute
	sqlDB.SetConnMaxIdleTime(lifetime)
	sqlDB.SetConnMaxLifetime(lifetime)

	return db, nil
}

func init() {
	db, err := connect()
	if err != nil {
		logger.Error(err, "database: failed to connect to database")
	}
	DB = db
}

// EnsureConnection verifies DB connectivity and reconnects if needed
func ensureConnection() error {
	// If db is not initialized, connect to the database
	if DB == nil {
		new_db, err := connect()
		if err != nil {
			logger.Error(err, "database: failed to ensure connection")
			return err
		}
		DB = new_db
		return nil
	} else { // If db is initialized, check if it is reachable
		sql_db, err := DB.DB()
		if err != nil {
			logger.Error(err, "database: failed to get database connection")
			return err
		}
		if err := sql_db.Ping(); err != nil {
			// If db is not reachable, try connecting to the database
			new_db, err := connect()
			if err != nil {
				logger.Error(err, "database: failed to connect to database")
				return err
			}
			DB = new_db
			return nil
		}
	}
	return nil
}

// GetDB returns a healthy *gorm.DB, attempting reconnect if necessary
func GetDB() (*gorm.DB, error) {
	if err := ensureConnection(); err != nil {
		logger.Error(err, "database: failed to get database connection")
		return nil, err
	}
	return DB, nil
}
