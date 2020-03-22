package types

//RegisterRequest A struct for carrying the POST data from the register request from the simulator.
type RegisterRequest struct {
	Username string
	Email    string
	Pwd      string
}
