package types

type LatestResponse struct {
	ID     int   `gorm:"primary_key" json:"-"`
	Latest int64 `json:"latest"`
}
