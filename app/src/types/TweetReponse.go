package types

type TweetResponse struct {
	Content  string
	PubDate string `json:"Pub_date"`
	User     string
}

func ConvertToTweetResponse(tweets []MessageViewData) []TweetResponse {
	var convertedMsgs []TweetResponse

	for _, msg := range tweets {
		convertedMsgs = append(convertedMsgs, TweetResponse{Content: msg.Text, User: msg.Username, PubDate: msg.PublishedDate})
	}
	return convertedMsgs
}
