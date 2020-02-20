package database

import (
	"fmt"

	"github.com/matt035343/devops/src/types"

	_ "github.com/mattn/go-sqlite3"
)

func (d *Database) GetFollower(whoID, whomID int) (follower types.Follower, err error) {
	err = d.db.Where(&types.Follower{WhomID: whomID, WhoID: whoID}).First(&follower).Error
	return follower, err
}

func (d *Database) GetFollowers(userID, limit int) (followers []types.Follower, err error) {
	err = d.db.Where(&types.Follower{WhomID: userID}).Limit(limit).Find(&followers).Error
	return followers, err
}

func (d *Database) GetFollowing(userID, limit int) (following []types.Follower, err error) {
	err = d.db.Where(&types.Follower{WhoID: userID}).Limit(limit).Find(&following).Error
	return following, err
}

func (d *Database) AddFollower(whoID, whomID int) error {
	f, _ := d.GetFollower(whoID, whomID)
	if !f.IsValidRelation() {
		return d.db.Create(&types.Follower{WhoID: whoID, WhomID: whomID}).Error
	}
	return fmt.Errorf("Already following this user")
}

func (d *Database) DeleteFollower(whoID, whomID int) error {
	return d.db.Delete(&types.Follower{WhoID: whoID, WhomID: whomID}).Error
}
