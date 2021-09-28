package models

type Message struct {
	Type     string `json:"type"`
	ClientId string `json:"client_id"`
	Content  string `json:"content"`
}
