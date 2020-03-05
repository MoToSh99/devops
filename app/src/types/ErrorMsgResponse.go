package types

type ErrorMsgResponse struct {
	Status    int
	ErrorMsg string `json:"error_msg"`
}
