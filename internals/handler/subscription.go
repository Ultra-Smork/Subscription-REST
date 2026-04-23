// @title Subscription Service API
// @version 1.0
// @description REST API for managing subscriptions
// @host localhost:8085
// @BasePath /api/v1
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

// Create subscription
// @Summary Create a new subscription
// @Description Creates a new subscription for a user with service details
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param request body dto.CreateSubscriptionRequest true "Subscription data"
// @Success 201 {object} dto.SubscriptionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/subscriptions [post]
func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("create subscription request received", "layer", "handler")

	var req dto.CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid JSON body", "error", err, "layer", "handler")
		h.respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if err := req.Validate(); err != nil {
		h.logger.Error("validation failed", "error", err, "layer", "handler")
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	sub, err := h.service.Create(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			h.logger.Error("invalid input", "error", err, "layer", "handler")
			h.respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.logger.Error("create failed", "error", err, "layer", "handler")
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.logger.Info("subscription created", "id", sub.ID, "user_id", sub.UserID, "service", sub.ServiceName, "layer", "handler")
	h.respondJSON(w, http.StatusCreated, dto.ToSubscriptionResponse(sub))
}

// Get subscription by ID
// @Summary Get a subscription by ID
// @Description Retrieves a single subscription by its UUID
// @Tags subscriptions
// @Produce json
// @Param id path string true "Subscription UUID"
// @Success 200 {object} dto.SubscriptionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/subscriptions/{id} [get]
func (h *SubscriptionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("get subscription by id request received", "layer", "handler")

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("invalid UUID format", "error", err, "layer", "handler")
		h.respondError(w, http.StatusBadRequest, "invalid UUID format")
		return
	}

	sub, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			h.logger.Error("subscription not found", "id", id, "layer", "handler")
			h.respondError(w, http.StatusNotFound, "subscription not found")
			return
		}
		h.logger.Error("get subscription failed", "error", err, "layer", "handler")
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.logger.Info("subscription retrieved", "id", sub.ID, "layer", "handler")
	h.respondJSON(w, http.StatusOK, dto.ToSubscriptionResponse(sub))
}

// List subscriptions
// @Summary List all subscriptions
// @Description Returns a list of all subscriptions with optional filtering
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "Filter by user UUID"
// @Param service_name query string false "Filter by service name"
// @Success 200 {array} dto.SubscriptionResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/subscriptions [get]
func (h *SubscriptionHandler) List(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("list subscriptions request received", "layer", "handler")

	var userID *uuid.UUID
	if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
		id, err := uuid.Parse(userIDStr)
		if err != nil {
			h.logger.Error("invalid user_id UUID", "error", err, "layer", "handler")
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
		h.logger.Error("list subscriptions failed", "error", err, "layer", "handler")
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.logger.Info("subscriptions listed", "count", len(subs), "layer", "handler")
	h.respondJSON(w, http.StatusOK, dto.ToSubscriptionResponseList(subs))
}

// Update subscription
// @Summary Update a subscription
// @Description Updates an existing subscription by its UUID
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription UUID"
// @Param request body dto.UpdateSubscriptionRequest true "Update data"
// @Success 200 {object} dto.SubscriptionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/subscriptions/{id} [put]
func (h *SubscriptionHandler) Update(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("update subscription request received", "layer", "handler")

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("invalid UUID format", "error", err, "layer", "handler")
		h.respondError(w, http.StatusBadRequest, "invalid UUID format")
		return
	}

	var req dto.UpdateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid JSON body", "error", err, "layer", "handler")
		h.respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if err := req.Validate(); err != nil {
		h.logger.Error("validation failed", "error", err, "layer", "handler")
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	sub, err := h.service.Update(r.Context(), id, req)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			h.logger.Error("subscription not found", "id", id, "layer", "handler")
			h.respondError(w, http.StatusNotFound, "subscription not found")
			return
		}
		if errors.Is(err, service.ErrInvalidInput) {
			h.logger.Error("invalid input", "error", err, "layer", "handler")
			h.respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.logger.Error("update failed", "error", err, "layer", "handler")
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.logger.Info("subscription updated", "id", sub.ID, "layer", "handler")
	h.respondJSON(w, http.StatusOK, dto.ToSubscriptionResponse(sub))
}

// Delete subscription
// @Summary Delete a subscription
// @Description Deletes a subscription by its UUID
// @Tags subscriptions
// @Param id path string true "Subscription UUID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/subscriptions/{id} [delete]
func (h *SubscriptionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("delete subscription request received", "layer", "handler")

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("invalid UUID format", "error", err, "layer", "handler")
		h.respondError(w, http.StatusBadRequest, "invalid UUID format")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			h.logger.Error("subscription not found", "id", id, "layer", "handler")
			h.respondError(w, http.StatusNotFound, "subscription not found")
			return
		}
		h.logger.Error("delete failed", "error", err, "layer", "handler")
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.logger.Info("subscription deleted", "id", id, "layer", "handler")
	w.WriteHeader(http.StatusNoContent)
}

// Get total cost
// @Summary Calculate total subscription cost
// @Description Calculates the sum of all subscription prices from a given start date with optional filters
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "Filter by user UUID"
// @Param service_name query string false "Filter by service name"
// @Param start_date query string true "Start date in MM-YYYY format"
// @Success 200 {object} dto.TotalCostResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/subscriptions/cost [get]
func (h *SubscriptionHandler) GetTotalCost(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("get total cost request received", "layer", "handler")

	var userID *uuid.UUID
	if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
		id, err := uuid.Parse(userIDStr)
		if err != nil {
			h.logger.Error("invalid user_id UUID", "error", err, "layer", "handler")
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
		h.logger.Error("start_date missing", "layer", "handler")
		h.respondError(w, http.StatusBadRequest, "start_date parameter is required")
		return
	}

	std, err := parseMonthYear(startDate)
	if err != nil {
		h.logger.Error("invalid start_date format", "error", err, "layer", "handler")
		h.respondError(w, http.StatusBadRequest, "invalid from format, use MM-YYYY")
		return
	}

	total, err := h.service.CalculateCost(r.Context(), userID, serviceName, std)
	if err != nil {
		if errors.Is(err, service.ErrInvalidPeriod) {
			h.logger.Error("invalid period", "error", err, "layer", "handler")
			h.respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.logger.Error("calculate cost failed", "error", err, "layer", "handler")
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.logger.Info("total cost calculated", "total", total, "layer", "handler")
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