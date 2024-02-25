package models

type Responce struct {
	Code    int
	Message string
	Headers map[string]any
	Body    string
}
