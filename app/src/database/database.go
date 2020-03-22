package database

import (
	"github.com/jinzhu/gorm"
	"github.com/matt035343/devops/app/src/types"
)

//Database Wrapper to a GORM database instance.
type Database struct {
	db *gorm.DB
}

//New Creates a new instance of Database given a GORM database instance.
func New(gdb *gorm.DB) *Database {
	return &Database{db: gdb}
}

//CloseDatabase Closes the database connection of the wrapped instance.
func (d *Database) CloseDatabase() {
	err := d.db.Close()
	if err != nil {
		panic(err)
	}
}

//ConnectDatabase Connects to a database given the dialect and connection string.
func ConnectDatabase(databaseDialect, connectionString string) (*Database, error) {
	db, err := gorm.Open(databaseDialect, connectionString)
	if err != nil {
		return nil, err
	}
	err = autoMigrate(db)
	return New(db), err
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&types.User{}).AutoMigrate(&types.Message{}).AutoMigrate(&types.Follower{}).AutoMigrate(&types.LatestResponse{}).Error
}
