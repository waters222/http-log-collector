package model

type Error struct {
	Code string `json:"code"`

	Message string `json:"message"`
}

type LogMessage struct {
	Message string `json:message`
}
