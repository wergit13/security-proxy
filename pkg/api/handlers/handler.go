package handlers

import (
	"net/http"
	"sc-proxy/pkg/service"

	_ "sc-proxy/docs"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Handler struct {
	services *service.Service
}

func NewHandler(
	services *service.Service) *Handler {
	return &Handler{
		services: services,
	}
}

func (h *Handler) InitRoutes() http.Handler {
	r := mux.NewRouter()
	r.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8000/swagger/doc.json"),
	))

	api := r.PathPrefix("/api").Subrouter()
	apiRouter := api.PathPrefix("/v1").Subrouter()

	apiRouter.HandleFunc("/requests", h.requests).Methods("GET")
	apiRouter.HandleFunc("/request/{id}", h.requestById).Methods("GET")
	apiRouter.HandleFunc("/pairs", h.pairs).Methods("GET")
	apiRouter.HandleFunc("/pair/{id}", h.pairById).Methods("GET")
	apiRouter.HandleFunc("/repeat/{id}", h.repeatById).Methods("GET")
	// apiRouter.HandleFunc("/scan/{id}", h.scanById).Methods("GET")

	return r
}
