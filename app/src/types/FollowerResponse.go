package types

//FollowerResponse A struct to carry data for the simulator GET follower request
type FollowerResponse struct {
	Follows []string `json:"follows"`
}
