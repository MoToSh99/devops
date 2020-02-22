package types

type Follower struct {
	WhoID  int
	WhomID int
}

func (f *Follower) IsValidRelation() bool {
	return f.WhoID != 0 && f.WhomID != 0
}
