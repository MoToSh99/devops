package types

type FollowRequest struct {
	Follow   string `json:"follow"`
	Unfollow string `json:"unfollow"`
}
