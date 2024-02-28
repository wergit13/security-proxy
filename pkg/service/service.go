package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sc-proxy/pkg/models"
	"sc-proxy/pkg/repository"
)

const conLen string = " длину ответа"
const code string = " код ответа"

var symbols = [...]string{`'`, `"`}

type Service struct {
	Repository *repository.Repository
}

func NewService(repo *repository.Repository) *Service {
	return &Service{repo}
}

func (s *Service) SavePair(ctx context.Context, rq *http.Request, rp *http.Response) (int, error) {
	request, err := requestToJson(rq)
	if err != nil {
		return 0, err
	}
	var responce *[]byte
	if rp != nil {
		resp, err := responceToJson(rp)
		if err != nil {
			return 0, err
		}
		responce = &resp
	}

	return s.Repository.Save(ctx, &request, responce)
}

func (s *Service) ClearBase(ctx context.Context) error {
	return s.Repository.ClearBase(ctx)
}

func (s *Service) GetAllRequests(ctx context.Context) ([]models.Request, error) {
	return s.Repository.GetAllRequests(ctx)
}

func (s *Service) GetRequestById(ctx context.Context, id int) (models.Request, error) {
	return s.Repository.GetRequestById(ctx, id)
}

func (s *Service) GetPairById(ctx context.Context, id int) (models.RequestResponce, error) {
	return s.Repository.GetPairById(ctx, id)
}

func (s *Service) GetAllPairs(ctx context.Context) ([]models.RequestResponce, error) {
	return s.Repository.GetAllPairs(ctx)
}

func (s *Service) RepeatRequest(ctx context.Context, request models.Request) (models.RequestResponce, error) {
	req, err := makeHttpRequest(request)
	if err != nil {
		return models.RequestResponce{}, err
	}

	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return models.RequestResponce{}, err
	}
	req, _ = makeHttpRequest(request)
	id, err := s.SavePair(ctx, req, resp)
	if err != nil {
		return models.RequestResponce{}, err
	}
	return s.GetPairById(ctx, id)
}

func sendRequest(request models.Request) (*http.Response, error) {
	req, err := makeHttpRequest(request)
	if err != nil {
		return nil, err
	}

	return http.DefaultTransport.RoundTrip(req)
}

func requestToJson(rq *http.Request) ([]byte, error) {
	err := rq.ParseForm()
	if err != nil {
		log.Println("Failed to parse req form")
		return []byte{}, err
	}

	request := models.Request{
		Method: rq.Method,
		Url:    rq.URL.Path,
		Host:   rq.Host,
		Scheme: rq.URL.Scheme,
	}
	request.Headers = make(map[string][]string)
	for key, value := range rq.Header {
		request.Headers[key] = value
	}
	request.Cookies = make(map[string]string)
	for _, cookie := range rq.Cookies() {
		request.Cookies[cookie.Name] = cookie.Value
	}
	request.GetParams = make(map[string][]string)
	for key, value := range rq.URL.Query() {
		request.GetParams[key] = value
	}

	request.PostParams = make(map[string][]string)
	for key, value := range rq.Form {
		request.PostParams[key] = value
	}

	body, err := io.ReadAll(rq.Body)
	if err != nil {
		return []byte{}, err
	}
	request.Body = string(body)

	return json.Marshal(request)
}

func responceToJson(rp *http.Response) ([]byte, error) {
	model, err := responceToModel(rp)
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(model)
}

func responceToModel(rp *http.Response) (models.Responce, error) {
	bodyBytes, err := io.ReadAll(rp.Body)
	rp.Body.Close()
	rp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	if err != nil {
		return models.Responce{}, err
	}
	responce := models.Responce{
		Code:    rp.StatusCode,
		Message: rp.Status,
		Body:    string(bodyBytes),
	}

	responce.Headers = make(map[string]any)
	for key, value := range rp.Header {
		responce.Headers[key] = value
	}
	return responce, nil
}

func makeHttpRequest(request models.Request) (*http.Request, error) {
	var body io.Reader
	if len(request.PostParams) != 0 {
		form := url.Values{}
		for k, v := range request.PostParams {
			for _, val := range v {
				form.Add(k, val)
			}
		}
		body = bytes.NewReader([]byte(form.Encode()))
	} else if len(request.Body) != 0 {
		body = bytes.NewReader([]byte(request.Body))
	} else {
		body = http.NoBody
	}

	req, err := http.NewRequest(request.Method, request.Url, body)
	if err != nil {
		return nil, err
	}

	req.Host = request.Host
	req.URL.Scheme = request.Scheme
	req.URL.Host = request.Host

	query := req.URL.Query()
	for key, value := range request.GetParams {
		for _, item := range value {
			query.Add(key, item)
		}
	}
	req.URL.RawQuery = query.Encode()

	for k, v := range request.Headers {
		for _, item := range v {
			req.Header.Set(k, fmt.Sprint(item))
		}
	}
	for k, v := range request.Cookies {
		req.AddCookie(&http.Cookie{Name: k, Value: fmt.Sprint(v)})
	}

	return req, nil
}

func (s *Service) ScanRequestOnSqlInjection(ctx context.Context, request models.Request) ([]string, error) {
	resp, err := sendRequest(request)
	if err != nil {
		return nil, fmt.Errorf("Не удалость отпрвить запрос")
	}
	resp.Body.Close()

	contentLen := resp.ContentLength
	status := resp.StatusCode

	var result []string

	//get params
	if len(request.GetParams) != 0 {
		res, err := scanMapOfStrings(request, contentLen, status, "get")
		if err != nil {
			return nil, err
		}
		result = append(result, res...)
	}

	//post
	if len(request.PostParams) != 0 {
		res, err := scanMapOfStrings(request, contentLen, status, "post")
		if err != nil {
			return nil, err
		}
		result = append(result, res...)
	}

	//headers
	if len(request.Headers) != 0 {
		res, err := scanMapOfStrings(request, contentLen, status, "header")
		if err != nil {
			return nil, err
		}
		result = append(result, res...)
	}
	//cookie
	if len(request.Cookies) != 0 {
		res, err := scanCookies(request, contentLen, status)
		if err != nil {
			return nil, err
		}
		result = append(result, res...)
	}
	//body
	if len(request.Body) != 0 {
		res, err := scanBody(request, contentLen, status)
		if err != nil {
			return nil, err
		}
		result = append(result, res...)
	}

	return result, nil
}

func scanMapOfStrings(request models.Request, contentLen int64, status int, scanType string) ([]string, error) {
	var res []string
	var table map[string][]string
	switch scanType {
	case "post":
		table = request.PostParams
	case "get":
		table = request.GetParams
	case "header":
		table = request.Headers
	}
	for _, symbol := range symbols {
		for name, list := range table {
			for i, val := range list {
				r := request
				switch scanType {
				case "post":
					r.PostParams[name][i] = val + symbol
				case "get":
					r.GetParams[name][i] = val + symbol
				case "header":
					r.Headers[name][i] = val + symbol
				}

				resp, err := sendRequest(r)
				if err != nil {
					return nil, fmt.Errorf("Не удалость отпрaвить запрос")
				}
				resp.Body.Close()

				if resp.ContentLength != contentLen || resp.StatusCode != status {
					var message string

					switch scanType {
					case "post":
						message = fmt.Sprintf("post: параметр %s с значением %s потенциально уязвим, добавление %s изменило", name, val, symbol)
					case "get":
						message = fmt.Sprintf("get: параметр %s с значением %s потенциально уязвим, добавление %s изменило", name, val, symbol)
					case "header":
						message = fmt.Sprintf("header: заголовок %s с значением %s потенциально уязвим, добавление %s изменило", name, val, symbol)
					}

					if resp.StatusCode != status {
						message += code
					}
					if resp.ContentLength != contentLen {
						message += conLen
					}
					res = append(res, message)
				}

			}
		}
	}
	return res, nil
}

func scanCookies(request models.Request, contentLen int64, status int) ([]string, error) {
	var res []string
	for name, val := range request.Cookies {
		r := request
		r.Cookies[name] = val + `'`
		resp, err := sendRequest(r)
		if err != nil {
			return nil, fmt.Errorf("Не удалость отпрaвить запрос")
		}
		resp.Body.Close()

		if resp.ContentLength != contentLen || resp.StatusCode != status {
			message := fmt.Sprintf("cookie: кука %s с значением %s уязвим, добавление %s изменило", name, val, `'`)
			if resp.StatusCode != status {
				message += code
			}
			if resp.ContentLength != contentLen {
				message += conLen
			}
			res = append(res, message)
		}

	}
	return res, nil
}

func scanBody(request models.Request, contentLen int64, status int) ([]string, error) {
	var res []string

	for _, symbol := range symbols {
		r := request
		resp, err := sendRequest(r)
		if err != nil {
			return nil, fmt.Errorf("Не удалость отпрaвить запрос")
		}
		resp.Body.Close()
		if resp.ContentLength != contentLen || resp.StatusCode != status {

			message := fmt.Sprintf("body: тело уязвимо, добавление %s изменило", symbol)
			if resp.StatusCode != status {
				message += code
			}
			if resp.ContentLength != contentLen {
				message += conLen
			}
			res = append(res, message)
		}
	}
	return res, nil
}
