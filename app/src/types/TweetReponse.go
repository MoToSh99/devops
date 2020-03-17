package types

//TweetResponse A struct for holding the HTTP response for a tweet request from the simulator.
type TweetResponse struct {
	Content string
	PubDate string `json:"Pub_date"`
	User    string
}

//ConvertToTweetResponse converts a slice of MessageViewData to a slice of TweetResponse.
func ConvertToTweetResponse(tweets []MessageViewData) []TweetResponse {
	var convertedMsgs []TweetResponse

	for _, msg := range tweets {
		convertedMsgs = append(convertedMsgs, TweetResponse{Content: msg.Text, User: msg.Username, PubDate: msg.PublishedDate})
	}
	return convertedMsgs
}
