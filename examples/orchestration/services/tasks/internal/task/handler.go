package task

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/aquamarinepk/aqm"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Handler wires HTTP routes for the tasks aggregate.
type Handler struct {
	service *Service
	logger  aqm.Logger
	cfg     *aqm.Config
}

func NewHandler(service *Service, logger aqm.Logger, cfg *aqm.Config) *Handler {
	if logger == nil {
		logger = aqm.NewNoopLogger()
	}
	return &Handler{service: service, logger: logger, cfg: cfg}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/tasks", func(r chi.Router) {
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
	tasks, err := h.service.List(r.Context())
	if err != nil {
		aqm.Error(w, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	aqm.Respond(w, http.StatusOK, tasks, nil)
}

func (h *Handler) handleGet(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		aqm.Error(w, http.StatusBadRequest, "invalid_id", err.Error())
		return
	}

	task, err := h.service.Get(r.Context(), id)
	if err != nil {
		h.handleDomainError(w, err, "get_failed")
		return
	}
	aqm.Respond(w, http.StatusOK, task, nil)
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

	task, err := h.service.Create(r.Context(), payload.Title, payload.Description)
	if err != nil {
		h.handleDomainError(w, err, "create_failed")
		return
	}
	aqm.Respond(w, http.StatusCreated, task, nil)
}

func (h *Handler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		aqm.Error(w, http.StatusBadRequest, "invalid_id", err.Error())
		return
	}

	var update UpdateTask
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		aqm.Error(w, http.StatusBadRequest, "invalid_payload", "Malformed JSON payload")
		return
	}

	task, err := h.service.Update(r.Context(), id, update)
	if err != nil {
		h.handleDomainError(w, err, "update_failed")
		return
	}
	aqm.Respond(w, http.StatusOK, task, nil)
}

func (h *Handler) handleReplace(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		aqm.Error(w, http.StatusBadRequest, "invalid_id", err.Error())
		return
	}

	var payload struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Completed   bool   `json:"completed"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		aqm.Error(w, http.StatusBadRequest, "invalid_payload", "Malformed JSON payload")
		return
	}

	task, err := h.service.Update(r.Context(), id, UpdateTask{
		Title:       &payload.Title,
		Description: &payload.Description,
		Completed:   &payload.Completed,
	})
	if err != nil {
		h.handleDomainError(w, err, "replace_failed")
		return
	}
	aqm.Respond(w, http.StatusOK, task, nil)
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		aqm.Error(w, http.StatusBadRequest, "invalid_id", err.Error())
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		h.handleDomainError(w, err, "delete_failed")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleComplete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		aqm.Error(w, http.StatusBadRequest, "invalid_id", err.Error())
		return
	}

	task, err := h.service.Complete(r.Context(), id)
	if err != nil {
		h.handleDomainError(w, err, "complete_failed")
		return
	}
	aqm.Respond(w, http.StatusOK, task, nil)
}

func (h *Handler) handleUncomplete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		aqm.Error(w, http.StatusBadRequest, "invalid_id", err.Error())
		return
	}

	task, err := h.service.Uncomplete(r.Context(), id)
	if err != nil {
		h.handleDomainError(w, err, "uncomplete_failed")
		return
	}
	aqm.Respond(w, http.StatusOK, task, nil)
}

func (h *Handler) handleDomainError(w http.ResponseWriter, err error, code string) {
	switch {
	case errors.Is(err, ErrNotFound):
		aqm.Error(w, http.StatusNotFound, "not_found", "task not found")
	case errors.Is(err, ErrValidationTitleNeeded):
		aqm.Error(w, http.StatusBadRequest, "validation_error", err.Error())
	default:
		aqm.Error(w, http.StatusInternalServerError, code, err.Error())
	}
}

func parseID(r *http.Request) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, "id"))
}
