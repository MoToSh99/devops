package types

//LatestResponse The data model storing the latest sequence number given to the simulator.
type LatestResponse struct {
	ID     int   `gorm:"primary_key" json:"-"`
	Latest int64 `json:"latest"`
}
