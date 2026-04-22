package service

import (
	"context"
	"errors"
	"fmt"
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
	repo repository.SubscriptionRepository
}

func NewSubscriptionService(repo repository.SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{repo: repo}
}

func (s *SubscriptionService) Create(ctx context.Context, req dto.CreateSubscriptionRequest) (*model.Subscription, error) {
	startDate, err := parseMonthYear(req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid start_date: %v", ErrInvalidInput, err)
	}
	sub := &model.Subscription{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   startDate,
	}

	if err := s.repo.Create(ctx, sub); err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	return sub, nil
}

func (s *SubscriptionService) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}
	return sub, nil
}

func (s *SubscriptionService) List(ctx context.Context, userID *uuid.UUID, serviceName *string) ([]*model.Subscription, error) {
	subs, err := s.repo.List(ctx, userID, serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}
	return subs, nil
}

func (s *SubscriptionService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateSubscriptionRequest) (*model.Subscription, error) {
	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get subscription for update: %w", err)
	}

	if req.ServiceName != nil {
		sub.ServiceName = *req.ServiceName
	}
	if req.Price != nil {
		if *req.Price < 0 {
			return nil, fmt.Errorf("%w: price cannot be negative", ErrInvalidInput)
		}
		sub.Price = *req.Price
	}
	if req.StartDate != nil && *req.StartDate != "" {
		startDate, err := parseMonthYear(*req.StartDate)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid start_date: %v", ErrInvalidInput, err)
		}
		sub.StartDate = startDate
	}

	if err := s.repo.Update(ctx, sub); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}

	return sub, nil
}

func (s *SubscriptionService) Delete(ctx context.Context, id uuid.UUID) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("failed to delete subscription: %w", err)
	}
	return nil
}

func (s *SubscriptionService) CalculateCost(ctx context.Context, userID *uuid.UUID, serviceName *string, startDate time.Time) (int, error) {
	total, err := s.repo.SumCost(ctx, userID, serviceName, startDate)
	if err != nil {
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
