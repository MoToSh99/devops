package types

type TweetResponse struct {
	Content  string
	PubDate string
	User     string
}

func ConvertToTweetResponse(tweets []MessageViewData) []TweetResponse {
	var convertedMsgs []TweetResponse

	for _, msg := range tweets {
		convertedMsgs = append(convertedMsgs, TweetResponse{Content: msg.Text, User: msg.Username, PubDate: msg.PublishedDate})
	}
	return convertedMsgs
}
