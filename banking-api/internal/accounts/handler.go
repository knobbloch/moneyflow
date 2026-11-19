package accounts

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"unicode/utf8"

	"moneyflow/banking-api/internal/httpx"
)

const maxInitialBalance = 1_000_000_000_000

var validCurrency = map[string]bool{"RUB": true, "USD": true, "EUR": true}

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/accounts", h.create)
	mux.HandleFunc("GET /api/v1/accounts/{id}", h.get)
	mux.HandleFunc("GET /api/v1/accounts/{id}/balance", h.balance)
	mux.HandleFunc("PATCH /api/v1/accounts/{id}/block", h.block)
}

type createRequest struct {
	OwnerName      string `json:"owner_name"`
	Currency       string `json:"currency"`
	InitialBalance int64  `json:"initial_balance"`
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "request body is not valid JSON")
		return
	}

	req.OwnerName = strings.TrimSpace(req.OwnerName)
	if n := utf8.RuneCountInString(req.OwnerName); n < 1 || n > 100 {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "owner_name must be 1..100 characters")
		return
	}
	if !validCurrency[req.Currency] {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "currency must be one of RUB, USD, EUR")
		return
	}
	if req.InitialBalance < 0 || req.InitialBalance > maxInitialBalance {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "initial_balance must be between 0 and 1000000000000")
		return
	}

	acc, err := h.repo.Create(r.Context(), req.OwnerName, req.Currency, req.InitialBalance)
	if err != nil {
		internalError(w, err)
		return
	}
	w.Header().Set("Location", "/api/v1/accounts/"+acc.ID)
	httpx.WriteJSON(w, http.StatusCreated, acc)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	acc, err := h.repo.Get(r.Context(), id)
	if err != nil {
		h.fail(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, acc)
}

func (h *Handler) balance(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	acc, err := h.repo.Get(r.Context(), id)
	if err != nil {
		h.fail(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"account_id": acc.ID,
		"balance":    acc.Balance,
		"currency":   acc.Currency,
	})
}

func (h *Handler) block(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	acc, err := h.repo.Block(r.Context(), id)
	if err != nil {
		h.fail(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, acc)
}

func pathID(w http.ResponseWriter, r *http.Request) (string, bool) {
	id := r.PathValue("id")
	if !httpx.IsUUID(id) {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "account id must be a valid UUID")
		return "", false
	}
	return id, true
}

func (h *Handler) fail(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, http.StatusNotFound, "ACCOUNT_NOT_FOUND", "account not found")
	case errors.Is(err, ErrAlreadyBlocked):
		httpx.WriteError(w, http.StatusConflict, "ACCOUNT_ALREADY_BLOCKED", "account is already blocked")
	default:
		internalError(w, err)
	}
}

func internalError(w http.ResponseWriter, err error) {
	log.Printf("internal error: %v", err)
	httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
}