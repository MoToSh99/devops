package types

//Follower The data model for a user (who) that follows another user (whom).
type Follower struct {
	WhoID  int
	WhomID int
}

//IsValidRelation Checks that the current contents of the Follower struct is valid.
func (f *Follower) IsValidRelation() bool {
	return f.WhoID != 0 && f.WhomID != 0
}
