package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

//	@Summary	get all requests
//	@Tags		requests
//	@ID			requests
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	ClientResponseDto[string]
//	@Failure	500	{object}	ClientResponseDto[string]
//	@Router		/api/v1/requests [get]
func (h *Handler) requests(w http.ResponseWriter, r *http.Request) {
	requests, err := h.services.GetAllRequests(r.Context())

	if err != nil {
		log.Println(err)
		NewErrorClientResponseDto(r.Context(), w, http.StatusInternalServerError, "internal server error")
		return
	}

	NewSuccessClientResponseDto(r.Context(), w, requests)
}

//	@Summary	get all requests with responces
//	@Tags		pairs
//	@ID			pairs
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	ClientResponseDto[string]
//	@Failure	500	{object}	ClientResponseDto[string]
//	@Router		/api/v1/pairs [get]
func (h *Handler) pairs(w http.ResponseWriter, r *http.Request) {
	pairs, err := h.services.GetAllPairs(r.Context())

	if err != nil {
		log.Println(err)
		NewErrorClientResponseDto(r.Context(), w, http.StatusInternalServerError, "internal server error")
		return
	}

	NewSuccessClientResponseDto(r.Context(), w, pairs)
}

//	@Summary	get request by id
//	@Tags		request
//	@ID			request
//	@Accept		json
//	@Produce	json
//	@Param		id			path		string	true	"request id"
//	@Success	200			{object}	ClientResponseDto[string]
//	@Failure	400,404,500	{object}	ClientResponseDto[string]
//	@Router		/api/v1/request/{id} [get]
func (h *Handler) requestById(w http.ResponseWriter, r *http.Request) {
	idStr, ok := mux.Vars(r)["id"]
	if !ok {
		NewErrorClientResponseDto(r.Context(), w, http.StatusBadRequest, "invalid params")
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		NewErrorClientResponseDto(r.Context(), w, http.StatusBadRequest, "invalid params")
		return
	}

	request, err := h.services.GetRequestById(r.Context(), id)

	if err != nil {
		NewErrorClientResponseDto(r.Context(), w, http.StatusInternalServerError, "internal server error")
		return
	}

	NewSuccessClientResponseDto(r.Context(), w, request)
}

//	@Summary	get request with responce by id
//	@Tags		pair
//	@ID			pair
//	@Accept		json
//	@Produce	json
//	@Param		id			path		string	true	"request id"
//	@Success	200			{object}	ClientResponseDto[string]
//	@Failure	400,404,500	{object}	ClientResponseDto[string]
//	@Router		/api/v1/pair/{id} [get]
func (h *Handler) pairById(w http.ResponseWriter, r *http.Request) {
	idStr, ok := mux.Vars(r)["id"]
	if !ok {
		NewErrorClientResponseDto(r.Context(), w, http.StatusBadRequest, "invalid params")
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		NewErrorClientResponseDto(r.Context(), w, http.StatusBadRequest, "invalid params")
		return
	}

	pair, err := h.services.GetPairById(r.Context(), id)

	if err != nil {
		NewErrorClientResponseDto(r.Context(), w, http.StatusInternalServerError, "internal server error")
		return
	}

	NewSuccessClientResponseDto(r.Context(), w, pair)
}

//	@Summary	get repeat request by id
//	@Tags		repeat
//	@ID			repeat
//	@Accept		json
//	@Produce	json
//	@Param		id			path		string	true	"request id"
//	@Success	200			{object}	ClientResponseDto[string]
//	@Failure	400,404,500	{object}	ClientResponseDto[string]
//	@Router		/api/v1/repeat/{id} [get]
func (h *Handler) repeatById(w http.ResponseWriter, r *http.Request) {
	idStr, ok := mux.Vars(r)["id"]
	if !ok {
		NewErrorClientResponseDto(r.Context(), w, http.StatusBadRequest, "invalid params")
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		NewErrorClientResponseDto(r.Context(), w, http.StatusBadRequest, "invalid params")
		return
	}

	request, err := h.services.GetRequestById(r.Context(), id)

	if err != nil {
		NewErrorClientResponseDto(r.Context(), w, http.StatusInternalServerError, "internal server error")
		return
	}

	resp, err := h.services.RepeatRequest(r.Context(), request)
	if err != nil {
		NewErrorClientResponseDto(r.Context(), w, http.StatusInternalServerError, "failed to repeat request")
		return
	}

	NewSuccessClientResponseDto(r.Context(), w, resp)
}
