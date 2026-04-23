package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Ultra-Smork/Subscription-service/internals/handler/dto"
	"github.com/Ultra-Smork/Subscription-service/internals/model"
	"github.com/Ultra-Smork/Subscription-service/internals/repository"
	"github.com/google/uuid"
)

var (
	ErrNotFound      = errors.New("subscription not found")
	ErrInvalidInput  = errors.New("invalid input")
	ErrInvalidPeriod = errors.New("invalid period")
)

type SubscriptionService struct {
	repo   repository.SubscriptionRepository
	logger *slog.Logger
}

func NewSubscriptionService(repo repository.SubscriptionRepository, logger *slog.Logger) *SubscriptionService {
	return &SubscriptionService{repo: repo, logger: logger}
}

func (s *SubscriptionService) Create(ctx context.Context, req dto.CreateSubscriptionRequest) (*model.Subscription, error) {
	s.logger.Debug("creating subscription", "layer", "service", "user_id", req.UserID, "service_name", req.ServiceName)

	startDate, err := parseMonthYear(req.StartDate)
	if err != nil {
		s.logger.Error("invalid start_date", "error", err, "layer", "service")
		return nil, fmt.Errorf("%w: invalid start_date: %v", ErrInvalidInput, err)
	}
	sub := &model.Subscription{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   startDate,
	}

	if err := s.repo.Create(ctx, sub); err != nil {
		s.logger.Error("create failed", "error", err, "layer", "service")
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	s.logger.Debug("subscription created", "layer", "service", "id", sub.ID)
	return sub, nil
}

func (s *SubscriptionService) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	s.logger.Debug("getting subscription by id", "layer", "service", "id", id)

	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			s.logger.Debug("subscription not found", "layer", "service", "id", id)
			return nil, ErrNotFound
		}
		s.logger.Error("get failed", "error", err, "layer", "service")
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}
	return sub, nil
}

func (s *SubscriptionService) List(ctx context.Context, userID *uuid.UUID, serviceName *string) ([]*model.Subscription, error) {
	s.logger.Debug("listing subscriptions", "layer", "service", "user_id", userID, "service_name", serviceName)

	subs, err := s.repo.List(ctx, userID, serviceName)
	if err != nil {
		s.logger.Error("list failed", "error", err, "layer", "service")
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}
	return subs, nil
}

func (s *SubscriptionService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateSubscriptionRequest) (*model.Subscription, error) {
	s.logger.Debug("updating subscription", "layer", "service", "id", id)

	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			s.logger.Debug("subscription not found", "layer", "service", "id", id)
			return nil, ErrNotFound
		}
		s.logger.Error("get for update failed", "error", err, "layer", "service")
		return nil, fmt.Errorf("failed to get subscription for update: %w", err)
	}

	if req.ServiceName != nil {
		sub.ServiceName = *req.ServiceName
	}
	if req.Price != nil {
		if *req.Price < 0 {
			s.logger.Error("price cannot be negative", "layer", "service", "price", *req.Price)
			return nil, fmt.Errorf("%w: price cannot be negative", ErrInvalidInput)
		}
		sub.Price = *req.Price
	}
	if req.StartDate != nil && *req.StartDate != "" {
		startDate, err := parseMonthYear(*req.StartDate)
		if err != nil {
			s.logger.Error("invalid start_date", "error", err, "layer", "service")
			return nil, fmt.Errorf("%w: invalid start_date: %v", ErrInvalidInput, err)
		}
		sub.StartDate = startDate
	}

	if err := s.repo.Update(ctx, sub); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			s.logger.Debug("subscription not found", "layer", "service", "id", id)
			return nil, ErrNotFound
		}
		s.logger.Error("update failed", "error", err, "layer", "service")
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}

	s.logger.Debug("subscription updated", "layer", "service", "id", sub.ID)
	return sub, nil
}

func (s *SubscriptionService) Delete(ctx context.Context, id uuid.UUID) error {
	s.logger.Debug("deleting subscription", "layer", "service", "id", id)

	err := s.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			s.logger.Debug("subscription not found", "layer", "service", "id", id)
			return ErrNotFound
		}
		s.logger.Error("delete failed", "error", err, "layer", "service")
		return fmt.Errorf("failed to delete subscription: %w", err)
	}
	return nil
}

func (s *SubscriptionService) CalculateCost(ctx context.Context, userID *uuid.UUID, serviceName *string, startDate time.Time) (int, error) {
	s.logger.Debug("calculating cost", "layer", "service", "user_id", userID, "service_name", serviceName, "start_date", startDate)

	total, err := s.repo.SumCost(ctx, userID, serviceName, startDate)
	if err != nil {
		s.logger.Error("calculate cost failed", "error", err, "layer", "service")
		return 0, fmt.Errorf("failed to calculate cost: %w", err)
	}
	return total, nil
}

func parseMonthYear(s string) (time.Time, error) {
	t, err := time.Parse("01-2006", s)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC), nil
}