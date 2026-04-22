package repository

import (
	"context"
	"errors"

	"time"

	"github.com/Ultra-Smork/Subscription-service/internals/model"
	"github.com/google/uuid"
)

var ErrNotFound = errors.New("record not found")

type SubscriptionRepository interface {
	Create(ctx context.Context, sub *model.Subscription) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error)
	Update(ctx context.Context, sub *model.Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, userID *uuid.UUID, serviceName *string) ([]*model.Subscription, error)
	SumCost(ctx context.Context, userID *uuid.UUID, serviceName *string, startDate time.Time) (int, error)
}
