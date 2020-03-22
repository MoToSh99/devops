package types

//RequestData A struct for carrying information to the timeline HTML template.
type RequestData struct {
	Title           string
	RequestEndpoint string
	Messages        []MessageViewData
	IsLoggedIn      bool
	SessionUser     string
	UserProfile     string
	Followed        bool
	HasError        bool
	ErrorMsg        string
}
