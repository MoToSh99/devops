package types

type TweetResponse struct {
	Content  string
	Pub_date string
	User     string
}

func ConvertToTweetResponse(tweets []MessageViewData) []TweetResponse {
	var converted_msgs []TweetResponse

	for _, msg := range tweets {
		converted_msgs = append(converted_msgs, TweetResponse{Content: msg.Text, User: msg.Username, Pub_date: msg.PublishedDate})
	}
	return converted_msgs
}
