package types

type RequestData struct {
	Title           string
	RequestEndpoint string
	Messages        []MessageViewData
	IsLoggedIn      bool
	SessionUser     string
	UserProfile     string
	Followed        bool
}
