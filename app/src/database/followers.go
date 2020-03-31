package database

import (
	"fmt"

	"github.com/matt035343/devops/app/src/log"
	"github.com/matt035343/devops/app/src/types"
)

//GetFollower Queries the database for the entry having whoID and whomID
func (d *Database) GetFollower(whoID, whomID int) (follower types.Follower, err error) {
	err = d.db.Where(&types.Follower{WhomID: whomID, WhoID: whoID}).First(&follower).Error
	log.WarningErr("Could not find follower relation with whoID %d and whomID %d", err, whoID, whomID)
	return follower, err
}

//GetFollowers Queries the database for who follows userID and returns a maximum of limit entries.
func (d *Database) GetFollowers(userID, limit int) (followers []types.Follower, err error) {
	err = d.db.Where(&types.Follower{WhomID: userID}).Limit(limit).Find(&followers).Error
	log.WarningErr("Could not find followers for userID %d", err, userID)
	return followers, err
}

//GetFollowing Queries the database for whom userID follows and returns a maximum of limit entries.
func (d *Database) GetFollowing(userID, limit int) (following []types.Follower, err error) {
	err = d.db.Where(&types.Follower{WhoID: userID}).Limit(limit).Find(&following).Error
	log.WarningErr("Could not find follows for userID %d", err, userID)
	return following, err
}

//AddFollower Adds a follower entry to the database given whoID and whomID
func (d *Database) AddFollower(whoID, whomID int) error {
	f, _ := d.GetFollower(whoID, whomID)
	if !f.IsValidRelation() {
		err := d.db.Create(&types.Follower{WhoID: whoID, WhomID: whomID}).Error
		log.WarningErr("Could not create follower relation for whoID %d and whomID %d", err, whoID, whomID)
		return err
	}
	return fmt.Errorf("already following this user")
}

//DeleteFollower Deletes entry from database given whoID and whomID
func (d *Database) DeleteFollower(whoID, whomID int) error {
	err := d.db.Delete(&types.Follower{WhoID: whoID, WhomID: whomID}).Error
	log.WarningErr("Could not delete follower relation for whoID %d and whomID %d", err, whoID, whomID)
	return err
}
