package data

import (
	"NakedVPN/internal/conf"
	"NakedVPN/internal/utils"
	"os"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewOrganizeRepo)

// Data .
type Data struct {
	db *gorm.DB
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	_log := &utils.CustomGORMLogger{
		Clog: *log.NewHelper(logger),
	}
	db, err := gorm.Open(sqlite.Open(c.Database.Source), &gorm.Config{
		Logger: _log,
	})
	if err != nil {
		log.Fatalf("open db failed: %v", err)
		os.Exit(1)
	}

	// Migrate the schema
	err = MergeInitData(db)
	if err != nil {
		log.Fatalf("migrate schema failed: %v", err)
		os.Exit(1)
	}
	cleanup := func() {
		defer func() {
			if sqlDB, err := db.DB(); err == nil {
				sqlDB.Close()
			}
		}()
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{
		db: db,
	}, cleanup, nil

}
