package types

//MessageViewData A struct for transferring information about a message to the HTML template.
type MessageViewData struct {
	Text          string
	Email         string
	GravatarURL   string
	Username      string
	PublishedDate string
}
