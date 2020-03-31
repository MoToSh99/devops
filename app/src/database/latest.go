package database

import (
	"github.com/jinzhu/gorm"
	"github.com/matt035343/devops/app/src/log"
	"github.com/matt035343/devops/app/src/types"
)

//GetLatest Retrieves the latest sequence ID from the database
func (d *Database) GetLatest() (l types.LatestResponse, err error) {
	err = d.db.Model(&types.LatestResponse{}).First(&l).Error
	if err == gorm.ErrRecordNotFound {
		l = types.LatestResponse{Latest: 0}
		err = d.db.Create(&l).Error
	}
	log.ErrorErr("Could not get latest sequence number", err)
	return l, err
}

//SetLatest Saves the latest sequence ID to the database
func (d *Database) SetLatest(latest int64) (err error) {
	l, err := d.GetLatest()
	if err != nil {
		return err
	}
	l.Latest = latest
	err = d.db.Save(&l).Error
	log.ErrorErr("Could not set latest sequence number", err)
	return err
}
