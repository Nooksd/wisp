package model

type Message struct {
	Type      string `json:"type"`
	From      string `json:"from"`
	To        string `json:"to"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
	ID        string `json:"id"`
}

type Ack struct {
	Type      string `json:"type"`
	MessageID string `json:"messageId"`
	From      string `json:"from"`
	To        string `json:"to"`
}
