package transfers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"moneyflow/banking-api/internal/httpx"
)

const maxAmount = 1_000_000_000

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/transfers", h.create)
	mux.HandleFunc("GET /api/v1/transfers/{id}", h.get)
	mux.HandleFunc("GET /api/v1/accounts/{id}/transfers", h.listByAccount)
}

type createRequest struct {
	FromAccountID string `json:"from_account_id"`
	ToAccountID   string `json:"to_account_id"`
	Amount        int64  `json:"amount"`
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if len(key) < 1 || len(key) > 64 {
		httpx.WriteError(w, http.StatusBadRequest, "MISSING_IDEMPOTENCY_KEY", "Idempotency-Key header is required (1..64 characters)")
		return
	}

	var req createRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "request body is not valid JSON")
		return
	}

	if !httpx.IsUUID(req.FromAccountID) || !httpx.IsUUID(req.ToAccountID) {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "from_account_id and to_account_id must be valid UUIDs")
		return
	}
	from := strings.ToLower(req.FromAccountID)
	to := strings.ToLower(req.ToAccountID)
	if from == to {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "cannot transfer to the same account")
		return
	}
	if req.Amount < 1 || req.Amount > maxAmount {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "amount must be between 1 and 1000000000")
		return
	}

	t, replayed, err := h.repo.Create(r.Context(), Input{From: from, To: to, Amount: req.Amount, Key: key})
	if err != nil {
		h.fail(w, err)
		return
	}
	if replayed {
		w.Header().Set("Idempotent-Replay", "true")
		httpx.WriteJSON(w, http.StatusOK, t)
		return
	}
	w.Header().Set("Location", "/api/v1/transfers/"+t.ID)
	httpx.WriteJSON(w, http.StatusCreated, t)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !httpx.IsUUID(id) {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "transfer id must be a valid UUID")
		return
	}
	t, err := h.repo.Get(r.Context(), strings.ToLower(id))
	if err != nil {
		h.fail(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, t)
}

func (h *Handler) listByAccount(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !httpx.IsUUID(id) {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "account id must be a valid UUID")
		return
	}
	limit, ok := intParam(w, r, "limit", 20, 1, 100)
	if !ok {
		return
	}
	offset, ok := intParam(w, r, "offset", 0, 0, 1_000_000_000)
	if !ok {
		return
	}
	items, err := h.repo.ListByAccount(r.Context(), strings.ToLower(id), limit, offset)
	if err != nil {
		h.fail(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items, "limit": limit, "offset": offset})
}

func intParam(w http.ResponseWriter, r *http.Request, name string, def, min, max int) (int, bool) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return def, true
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < min || v > max {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", name+" is out of range or not a number")
		return 0, false
	}
	return v, true
}

func (h *Handler) fail(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, http.StatusNotFound, "TRANSFER_NOT_FOUND", "transfer not found")
	case errors.Is(err, ErrAccountNotFound):
		httpx.WriteError(w, http.StatusNotFound, "ACCOUNT_NOT_FOUND", "account not found")
	case errors.Is(err, ErrAccountBlocked):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "ACCOUNT_BLOCKED", "one of the accounts is blocked")
	case errors.Is(err, ErrCurrencyMismatch):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "CURRENCY_MISMATCH", "accounts have different currencies")
	case errors.Is(err, ErrInsufficientFunds):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "INSUFFICIENT_FUNDS", "insufficient funds")
	case errors.Is(err, ErrIdempotencyConflict):
		httpx.WriteError(w, http.StatusConflict, "IDEMPOTENCY_KEY_CONFLICT", "idempotency key was used with a different request")
	default:
		log.Printf("internal error: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}