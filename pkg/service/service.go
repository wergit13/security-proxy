package service

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sc-proxy/pkg/models"
	"sc-proxy/pkg/repository"
)

type Service struct {
	Repository *repository.Repository
}

func NewService(repo *repository.Repository) *Service {
	return &Service{repo}
}

func (s *Service) SavePair(ctx context.Context, rq *http.Request, rp *http.Response) error {
	request, err := requestToJson(rq)
	if err != nil {
		return err
	}
	var responce *[]byte
	if rp != nil {
		resp, err := responceToJson(rp)
		if err != nil {
			return err
		}
		responce = &resp
	}

	_, err = s.Repository.Save(ctx, &request, responce)
	return err
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
	body, err := io.ReadAll(rp.Body)
	if err != nil {
		return []byte{}, err
	}
	responce := models.Responce{
		Code:    rp.StatusCode,
		Message: rp.Status,
		Body:    string(body),
	}

	responce.Headers = make(map[string]any)
	for key, value := range rp.Header {
		responce.Headers[key] = value
	}
	return json.Marshal(responce)
}
