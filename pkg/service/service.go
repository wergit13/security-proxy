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
	bodyBytes, err := io.ReadAll(rp.Body)
	rp.Body.Close() //  must close
	rp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	if err != nil {
		return []byte{}, err
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
	return json.Marshal(responce)
}

func makeHttpRequest(request models.Request) (*http.Request, error) {
	var body []byte
	if len(request.PostParams) != 0 {
		form := url.Values{}
		for k, v := range request.PostParams {
			for _, val := range v {
				form.Add(k, val)
			}
		}
		body = []byte(form.Encode())
	} else {
		body = []byte(request.Body)
	}

	req, err := http.NewRequest(request.Method, request.Url, bytes.NewReader(body))
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
