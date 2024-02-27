package models

type Request struct {
	Id         int
	Method     string
	Url        string
	Host       string
	Scheme     string
	Headers    map[string][]string
	Cookies    map[string]string
	GetParams  map[string][]string
	PostParams map[string][]string
	Body       string
}
