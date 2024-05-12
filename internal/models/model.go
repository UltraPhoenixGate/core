package models

import (
	"ultraphx-core/internal/config"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func init() {
	Setup()
}

// Setup initializes the database instance
func Setup() {
	dbConf := config.GetDataBaseConfig()
	db, err := gorm.Open(sqlite.Open(dbConf.File), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	Migration(db)
	DB = db
}

func migrate(db *gorm.DB, models ...interface{}) {
	err := db.AutoMigrate(models...)
	if err != nil {
		panic("failed to migrate database")
	}
}

func AutoMigrate(models ...interface{}) {
	migrate(DB, models...)
}

// Migration migrate the schema
func Migration(db *gorm.DB) {
	// Migrate the schema
	migrate(db, &Client{}, &Permission{}) // client
}
