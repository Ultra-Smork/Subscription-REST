package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/Ultra-Smork/Subscription-service/internals/handler/dto"
	"github.com/Ultra-Smork/Subscription-service/internals/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type SubscriptionHandler struct {
	service *service.SubscriptionService
	logger  *slog.Logger
}

func NewSubscriptionHandler(svc *service.SubscriptionService, logger *slog.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		service: svc,
		logger:  logger,
	}
}

func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if err := req.Validate(); err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	sub, err := h.service.Create(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			h.respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusCreated, dto.ToSubscriptionResponse(sub))
}

func (h *SubscriptionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid UUID format")
		return
	}

	sub, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			h.respondError(w, http.StatusNotFound, "subscription not found")
			return
		}
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, dto.ToSubscriptionResponse(sub))
}

func (h *SubscriptionHandler) List(w http.ResponseWriter, r *http.Request) {
	var userID *uuid.UUID
	if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
		id, err := uuid.Parse(userIDStr)
		if err != nil {
			h.respondError(w, http.StatusBadRequest, "invalid user_id UUID")
			return
		}
		userID = &id
	}

	var serviceName *string
	if sn := r.URL.Query().Get("service_name"); sn != "" {
		serviceName = &sn
	}

	subs, err := h.service.List(r.Context(), userID, serviceName)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, dto.ToSubscriptionResponseList(subs))
}

func (h *SubscriptionHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid UUID format")
		return
	}

	var req dto.UpdateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if err := req.Validate(); err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	sub, err := h.service.Update(r.Context(), id, req)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			h.respondError(w, http.StatusNotFound, "subscription not found")
			return
		}
		if errors.Is(err, service.ErrInvalidInput) {
			h.respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, dto.ToSubscriptionResponse(sub))
}

func (h *SubscriptionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid UUID format")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			h.respondError(w, http.StatusNotFound, "subscription not found")
			return
		}
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *SubscriptionHandler) GetTotalCost(w http.ResponseWriter, r *http.Request) {
	var userID *uuid.UUID
	if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
		id, err := uuid.Parse(userIDStr)
		if err != nil {
			h.respondError(w, http.StatusBadRequest, "invalid user_id UUID")
			return
		}
		userID = &id
	}

	var serviceName *string
	if sn := r.URL.Query().Get("service_name"); sn != "" {
		serviceName = &sn
	}

	startDate := r.URL.Query().Get("start_date")
	if startDate == "" {
		h.respondError(w, http.StatusBadRequest, "start_date parameter is required")
		return
	}

	std, err := parseMonthYear(startDate)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid from format, use MM-YYYY")
		return
	}

	total, err := h.service.CalculateCost(r.Context(), userID, serviceName, std)
	if err != nil {
		if errors.Is(err, service.ErrInvalidPeriod) {
			h.respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, dto.TotalCostResponse{TotalCost: total})
}

func (h *SubscriptionHandler) respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
	}
}

func (h *SubscriptionHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, dto.ErrorResponse{Error: message})
}

func parseMonthYear(s string) (time.Time, error) {
	return time.Parse("01-2006", s)
}
