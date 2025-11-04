package database

import (
	"context"

	"gorm.io/gorm"
)

// CreateEntity creates a record for the provided entity type.
func CreateEntity[T any](ctx context.Context, entity *T) error {
	db, err := GetDB()
	if err != nil {
		return err
	}
	return db.WithContext(ctx).Create(entity).Error
}

// GetEntityByID returns a single record of type T by its primary key id.
func GetEntityByID[T any, ID comparable](ctx context.Context, id ID) (*T, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	var out T
	if err := db.WithContext(ctx).First(&out, id).Error; err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateEntityByID updates columns of type T where primary key equals id.
// Pass a non-empty updates map; values set to nil will be written as NULL.
func UpdateEntityByID[T any, ID comparable](ctx context.Context, id ID, updates map[string]interface{}) error {
	db, err := GetDB()
	if err != nil {
		return err
	}
	var zero T
	return db.WithContext(ctx).Model(&zero).Where("id = ?", id).Updates(updates).Error
}

// DeleteEntityByID deletes a record of type T by its primary key id.
func DeleteEntityByID[T any, ID comparable](ctx context.Context, id ID) error {
	db, err := GetDB()
	if err != nil {
		return err
	}
	var zero T
	return db.WithContext(ctx).Where("id = ?", id).Delete(&zero).Error
}

// WithTx allows running a function within a transaction using the shared DB.
func WithTx(ctx context.Context, fn func(tx *gorm.DB) error) error {
	db, err := GetDB()
	if err != nil {
		return err
	}
	return db.WithContext(ctx).Transaction(fn)
}
