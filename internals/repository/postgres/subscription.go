package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/Ultra-Smork/Subscription-service/internals/model"
	"github.com/Ultra-Smork/Subscription-service/internals/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SubscriptionRepo struct {
	pool *pgxpool.Pool
}

func NewSubscriptionRepository(pool *pgxpool.Pool) repository.SubscriptionRepository {
	return &SubscriptionRepo{pool: pool}
}

func (r *SubscriptionRepo) Create(ctx context.Context, sub *model.Subscription) error {
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

	return row.Scan(&sub.ID)
}

func (r *SubscriptionRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
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
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return sub, nil
}
func (r *SubscriptionRepo) Update(ctx context.Context, sub *model.Subscription) error {
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
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.ErrNotFound
		}
		return err
	}
	return nil
}

func (r *SubscriptionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM subscriptions WHERE id = $1`
	res, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *SubscriptionRepo) List(ctx context.Context, userID *uuid.UUID, serviceName *string) ([]*model.Subscription, error) {
	query := `
        SELECT id, service_name, price, user_id, start_date
        FROM subscriptions
        WHERE ($1::UUID IS NULL OR user_id = $1)
          AND ($2::TEXT IS NULL OR service_name = $2)`

	rows, err := r.pool.Query(ctx, query, userID, serviceName)
	if err != nil {
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
			return nil, err
		}
		subs = append(subs, sub)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return subs, nil
}

func (r *SubscriptionRepo) SumCost(ctx context.Context, userID *uuid.UUID, serviceName *string, startDate time.Time) (int, error) {
	query := `
        SELECT COALESCE(SUM(price), 0)
        FROM subscriptions
        WHERE start_date >= $1
		  AND ($2::UUID IS NULL OR user_id = $2)
          AND ($3::TEXT IS NULL OR service_name = $3)`

	var total int
	err := r.pool.QueryRow(ctx, query, startDate, userID, serviceName).Scan(&total)
	return total, err
}
