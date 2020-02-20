package database

import (
	"github.com/jinzhu/gorm"
	"github.com/matt035343/devops/src/types"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db *gorm.DB
}

func New(gdb *gorm.DB) *Database {
	return &Database{db: gdb}
}

func (d *Database) CloseDatabase() {
	d.db.Close()
}

func ConnectDatabase(databaseDialect, connectionString string) (*Database, error) {
	db, err := gorm.Open(databaseDialect, connectionString)
	autoMigrate(db)
	return New(db), err
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&types.User{}).AutoMigrate(&types.Message{}).AutoMigrate(&types.Follower{}).Error
}
