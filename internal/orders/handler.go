package orders

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/bhumi46/orders-svc/internal/httpx"
)

// Handler serves /api/orders.
type Handler struct {
	store *Store
	users UserValidator
}

// NewHandler returns a Handler that validates user IDs with users.
func NewHandler(store *Store, users UserValidator) *Handler {
	return &Handler{store: store, users: users}
}

// Register adds the orders routes to mux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/orders", h.list)
	mux.HandleFunc("GET /api/orders/{id}", h.get)
	mux.HandleFunc("POST /api/orders", h.create)
}

func (h *Handler) list(w http.ResponseWriter, _ *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, h.store.List())
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	o, err := h.store.Get(r.PathValue("id"))
	if errors.Is(err, ErrNotFound) {
		httpx.WriteError(w, http.StatusNotFound, err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, o)
}

type createRequest struct {
	UserID   string `json:"user_id"`
	Item     string `json:"item"`
	Quantity int    `json:"quantity"`
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	req.UserID = strings.TrimSpace(req.UserID)
	req.Item = strings.TrimSpace(req.Item)
	if req.UserID == "" || req.Item == "" || req.Quantity < 1 {
		httpx.WriteError(w, http.StatusBadRequest, "user_id, item and quantity >= 1 are required")
		return
	}

	switch err := h.users.ValidateUser(r.Context(), req.UserID); {
	case errors.Is(err, ErrUnknownUser):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "user "+req.UserID+" does not exist")
		return
	case err != nil:
		slog.Error("validate user", "user_id", req.UserID, "err", err)
		httpx.WriteError(w, http.StatusBadGateway, "users-svc unavailable")
		return
	}

	o := h.store.Create(Order{UserID: req.UserID, Item: req.Item, Quantity: req.Quantity})
	httpx.WriteJSON(w, http.StatusCreated, o)
}
