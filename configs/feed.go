package configs

type MusicFeedConfiguration struct {
	FeedLimit int `json:"FeedLimit"`
	FeedTTL   int `json:"FeedTTL"` //seconds
}
