package models

type Message struct {
	Type     string `json:"type"`
	ClientId string `json:"clientId"`
	Content  string `json:"content"`
}
