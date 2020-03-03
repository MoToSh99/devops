package database

import (
	"github.com/jinzhu/gorm"
	"github.com/matt035343/devops/app/src/types"
)

func (d *Database) GetLatest() (l types.LatestResponse, err error) {
	err = d.db.Model(&types.LatestResponse{}).First(&l).Error
	if err == gorm.ErrRecordNotFound {
		l = types.LatestResponse{Latest: 0}
		err = d.db.Create(&l).Error
	}
	return l, err
}

func (d *Database) SetLatest(latest int64) (err error) {
	l, err := d.GetLatest()
	if err != nil {
		return err
	}
	l.Latest = latest
	return d.db.Save(&l).Error
}
