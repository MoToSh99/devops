package types

//ErrorMsgResponse A struct to carry information about a failed HTTP request
type ErrorMsgResponse struct {
	Status   int
	ErrorMsg string `json:"error_msg"`
}
