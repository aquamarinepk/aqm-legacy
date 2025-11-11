package todo

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/aquamarinepk/aqm"
	"github.com/go-chi/chi/v5"
)

// Handler wires HTTP routes for the todo aggregate.
type Handler struct {
	logger  aqm.Logger
	cfg     *aqm.Config
	service *Service
}

// NewHandler constructs the HTTP handler.
func NewHandler(service *Service, logger aqm.Logger, cfg *aqm.Config) *Handler {
	if logger == nil {
		logger = aqm.NewNoopLogger()
	}
	if service == nil {
		service = NewService(nil, logger, cfg)
	}
	return &Handler{logger: logger, cfg: cfg, service: service}
}

// RegisterRoutes implements aqm.HTTPModule and sets up routing plus middleware.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/todos", func(r chi.Router) {
		r.Get("/", h.handleList)
		r.Post("/", h.handleCreate)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.handleGet)
			r.Patch("/", h.handleUpdate)
			r.Put("/", h.handleReplace)
			r.Delete("/", h.handleDelete)
			r.Post("/complete", h.handleComplete)
			r.Delete("/complete", h.handleUncomplete)
		})
	})
}

func (h *Handler) handleList(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListItems(r.Context())
	if err != nil {
		aqm.Error(w, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}

	aqm.Respond(w, http.StatusOK, items, nil)
}

func (h *Handler) handleGet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	item, err := h.service.GetItem(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			aqm.Error(w, http.StatusNotFound, "not_found", "todo not found")
			return
		}
		aqm.Error(w, http.StatusInternalServerError, "get_failed", err.Error())
		return
	}

	aqm.Respond(w, http.StatusOK, item, nil)
}

func (h *Handler) handleCreate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var payload struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		aqm.Error(w, http.StatusBadRequest, "invalid_payload", "Malformed JSON payload")
		return
	}

	item, err := h.service.CreateItem(r.Context(), payload.Title, payload.Description)
	if err != nil {
		if errors.Is(err, ErrValidationTitleRequired) {
			aqm.Error(w, http.StatusBadRequest, "validation_error", err.Error())
			return
		}
		aqm.Error(w, http.StatusInternalServerError, "create_failed", err.Error())
		return
	}

	aqm.Respond(w, http.StatusCreated, item, nil)
}

func (h *Handler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	defer r.Body.Close()

	var payload struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
		Completed   *bool   `json:"completed"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		aqm.Error(w, http.StatusBadRequest, "invalid_payload", "Malformed JSON payload")
		return
	}

	updated, err := h.service.UpdateItem(r.Context(), id, UpdateItem{
		Title:       payload.Title,
		Description: payload.Description,
		Completed:   payload.Completed,
	})
	if err != nil {
		if errors.Is(err, ErrValidationTitleRequired) {
			aqm.Error(w, http.StatusBadRequest, "validation_error", err.Error())
			return
		}
		if errors.Is(err, ErrNotFound) {
			aqm.Error(w, http.StatusNotFound, "not_found", "todo not found")
			return
		}
		aqm.Error(w, http.StatusInternalServerError, "update_failed", err.Error())
		return
	}

	aqm.Respond(w, http.StatusOK, updated, nil)
}

func (h *Handler) handleReplace(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	defer r.Body.Close()

	var payload struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Completed   bool   `json:"completed"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		aqm.Error(w, http.StatusBadRequest, "invalid_payload", "Malformed JSON payload")
		return
	}

	updated, err := h.service.UpdateItem(r.Context(), id, UpdateItem{
		Title:       &payload.Title,
		Description: &payload.Description,
		Completed:   &payload.Completed,
	})
	if err != nil {
		if errors.Is(err, ErrValidationTitleRequired) {
			aqm.Error(w, http.StatusBadRequest, "validation_error", err.Error())
			return
		}
		if errors.Is(err, ErrNotFound) {
			aqm.Error(w, http.StatusNotFound, "not_found", "todo not found")
			return
		}
		aqm.Error(w, http.StatusInternalServerError, "replace_failed", err.Error())
		return
	}

	aqm.Respond(w, http.StatusOK, updated, nil)
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.service.DeleteItem(r.Context(), id); err != nil {
		if errors.Is(err, ErrNotFound) {
			aqm.Error(w, http.StatusNotFound, "not_found", "todo not found")
			return
		}
		aqm.Error(w, http.StatusInternalServerError, "delete_failed", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleComplete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	done := true
	updated, err := h.service.UpdateItem(r.Context(), id, UpdateItem{Completed: &done})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			aqm.Error(w, http.StatusNotFound, "not_found", "todo not found")
			return
		}
		aqm.Error(w, http.StatusInternalServerError, "complete_failed", err.Error())
		return
	}
	aqm.Respond(w, http.StatusOK, updated, nil)
}

func (h *Handler) handleUncomplete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	done := false
	updated, err := h.service.UpdateItem(r.Context(), id, UpdateItem{Completed: &done})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			aqm.Error(w, http.StatusNotFound, "not_found", "todo not found")
			return
		}
		aqm.Error(w, http.StatusInternalServerError, "uncomplete_failed", err.Error())
		return
	}
	aqm.Respond(w, http.StatusOK, updated, nil)
}
