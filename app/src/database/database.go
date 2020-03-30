package database

import (
	"github.com/jinzhu/gorm"
	"github.com/matt035343/devops/app/src/log"
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
		log.ErrorErr("Could not close DB", err)
		panic(err)
	}
}

//ConnectDatabase Connects to a database given the dialect and connection string.
func ConnectDatabase(databaseDialect, connectionString string) (*Database, error) {
	db, err := gorm.Open(databaseDialect, connectionString)
	if err != nil {
		log.CriticalErr("Could not connect to %s DB", err, databaseDialect)
		return nil, err
	}
	log.Info("Database connected")

	err = autoMigrate(db)
	if err != nil {
		log.CriticalErr("Could not auto migrate %s db", err, databaseDialect)
	}
	return New(db), err
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&types.User{}).AutoMigrate(&types.Message{}).AutoMigrate(&types.Follower{}).AutoMigrate(&types.LatestResponse{}).Error
}
