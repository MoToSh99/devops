package types

//FollowRequest A struct to carry POST data for the simulator follow request.
type FollowRequest struct {
	Follow   string `json:"follow"`
	Unfollow string `json:"unfollow"`
}
