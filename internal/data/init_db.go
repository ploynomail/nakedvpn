package data

import (
	"NakedVPN/internal/biz"

	"gorm.io/gorm"
)

func MergeInitData(db *gorm.DB) error {
	{

		if err := db.AutoMigrate(&biz.Organize{}); err != nil {
			return err
		}
		// if err := db.Create(&biz.Organize{
		// 	ID:         1,
		// 	Name:       "org1",
		// 	AccessKey:  "accesskey1",
		// 	SubnetCIDR: "10.0.1.0/24",
		// }).Error; err != nil {
		// 	return err
		// }
	}
	return nil
}
