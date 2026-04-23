package postgres

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/Ultra-Smork/Subscription-service/internals/model"
	"github.com/Ultra-Smork/Subscription-service/internals/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SubscriptionRepo struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func NewSubscriptionRepository(pool *pgxpool.Pool, logger *slog.Logger) repository.SubscriptionRepository {
	return &SubscriptionRepo{pool: pool, logger: logger}
}

func (r *SubscriptionRepo) Create(ctx context.Context, sub *model.Subscription) error {
	r.logger.Debug("inserting subscription", "layer", "repository", "service_name", sub.ServiceName, "user_id", sub.UserID, "start_date", sub.StartDate)

	query := `
        INSERT INTO subscriptions (service_name, price, user_id, start_date)
        VALUES ($1, $2, $3, $4)
        RETURNING id`
	row := r.pool.QueryRow(ctx, query,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		sub.StartDate,
	)

	if err := row.Scan(&sub.ID); err != nil {
		r.logger.Error("insert failed", "error", err, "layer", "repository")
		return err
	}

	r.logger.Debug("subscription inserted", "layer", "repository", "id", sub.ID)
	return nil
}

func (r *SubscriptionRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	r.logger.Debug("fetching subscription by id", "layer", "repository", "id", id)

	query := `
        SELECT id, service_name, price, user_id, start_date
        FROM subscriptions
        WHERE id = $1`

	sub := &model.Subscription{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&sub.ID,
		&sub.ServiceName,
		&sub.Price,
		&sub.UserID,
		&sub.StartDate,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Debug("subscription not found", "layer", "repository", "id", id)
			return nil, repository.ErrNotFound
		}
		r.logger.Error("fetch failed", "error", err, "layer", "repository")
		return nil, err
	}
	return sub, nil
}

func (r *SubscriptionRepo) Update(ctx context.Context, sub *model.Subscription) error {
	r.logger.Debug("updating subscription", "layer", "repository", "id", sub.ID)

	query := `
        UPDATE subscriptions
        SET service_name = $1,
            price = $2,
            start_date = $3
        WHERE id = $4`

	_, err := r.pool.Exec(ctx, query,
		&sub.ServiceName,
		&sub.Price,
		&sub.StartDate,
		&sub.ID,
	)
	if err != nil {
		r.logger.Error("update failed", "error", err, "layer", "repository")
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.ErrNotFound
		}
		return err
	}
	r.logger.Debug("subscription updated", "layer", "repository", "id", sub.ID)
	return nil
}

func (r *SubscriptionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	r.logger.Debug("deleting subscription", "layer", "repository", "id", id)

	query := `DELETE FROM subscriptions WHERE id = $1`
	res, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		r.logger.Error("delete failed", "error", err, "layer", "repository")
		return err
	}
	if res.RowsAffected() == 0 {
		r.logger.Debug("subscription not found", "layer", "repository", "id", id)
		return repository.ErrNotFound
	}
	r.logger.Debug("subscription deleted", "layer", "repository", "id", id)
	return nil
}

func (r *SubscriptionRepo) List(ctx context.Context, userID *uuid.UUID, serviceName *string) ([]*model.Subscription, error) {
	r.logger.Debug("listing subscriptions", "layer", "repository", "user_id", userID, "service_name", serviceName)

	query := `
        SELECT id, service_name, price, user_id, start_date
        FROM subscriptions
        WHERE ($1::UUID IS NULL OR user_id = $1)
          AND ($2::TEXT IS NULL OR service_name = $2)`

	rows, err := r.pool.Query(ctx, query, userID, serviceName)
	if err != nil {
		r.logger.Error("list query failed", "error", err, "layer", "repository")
		return nil, err
	}
	defer rows.Close()

	var subs []*model.Subscription
	for rows.Next() {
		sub := &model.Subscription{}
		err := rows.Scan(
			&sub.ID,
			&sub.ServiceName,
			&sub.Price,
			&sub.UserID,
			&sub.StartDate,
		)
		if err != nil {
			r.logger.Error("scan failed", "error", err, "layer", "repository")
			return nil, err
		}
		subs = append(subs, sub)
	}
	if err := rows.Err(); err != nil {
		r.logger.Error("rows error", "error", err, "layer", "repository")
		return nil, err
	}
	r.logger.Debug("subscriptions listed", "layer", "repository", "count", len(subs))
	return subs, nil
}

func (r *SubscriptionRepo) SumCost(ctx context.Context, userID *uuid.UUID, serviceName *string, startDate time.Time) (int, error) {
	r.logger.Debug("calculating sum cost", "layer", "repository", "user_id", userID, "service_name", serviceName, "start_date", startDate)

	query := `
        SELECT COALESCE(SUM(price), 0)
        FROM subscriptions
        WHERE start_date >= $1
		  AND ($2::UUID IS NULL OR user_id = $2)
          AND ($3::TEXT IS NULL OR service_name = $3)`

	var total int
	err := r.pool.QueryRow(ctx, query, startDate, userID, serviceName).Scan(&total)
	if err != nil {
		r.logger.Error("sum cost query failed", "error", err, "layer", "repository")
		return 0, err
	}
	r.logger.Debug("sum cost calculated", "layer", "repository", "total", total)
	return total, nil
}