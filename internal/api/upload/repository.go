package upload

import (
	"ai-learn-english/internal/database/model"
	"errors"

	"gorm.io/gorm"
)

// EnsureDocumentsStatusColumn attempts to add documents.status if it doesn't exist.
// MySQL 8.0+ supports IF NOT EXISTS; on older versions, ignore duplicate column errors.
func EnsureDocumentsStatusColumn(db *gorm.DB) error {
	if db == nil {
		return errors.New("nil db")
	}
	res := db.Exec("ALTER TABLE documents ADD COLUMN IF NOT EXISTS status VARCHAR(32) NOT NULL DEFAULT 'uploaded'")
	if res.Error != nil {
		var count int64
		check := db.
			Table("information_schema.COLUMNS").
			Where("TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?", "documents", "status").
			Count(&count)
		if check.Error != nil {
			return res.Error
		}
		if count == 0 {
			return res.Error
		}
	}
	return nil
}

// EnsureDefaultUser finds or creates a default user and returns its ID.
func EnsureDefaultUser(db *gorm.DB) (int64, error) {
	if db == nil {
		return 0, errors.New("nil db")
	}
	const defaultEmail = "default@local"
	var u model.User
	err := db.Where("email = ?", defaultEmail).First(&u).Error
	if err == nil {
		return u.ID, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		newUser := model.User{Email: defaultEmail}
		if e := db.Create(&newUser).Error; e != nil {
			return 0, e
		}
		return newUser.ID, nil
	}
	return 0, err
}
