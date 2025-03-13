package data

import (
	"NakedVPN/internal/biz"

	"gorm.io/gorm"
)

func MergeInitData(db *gorm.DB) error {
	if err := db.AutoMigrate(&biz.Organize{}); err != nil {
		return err
	}
	return nil
}
