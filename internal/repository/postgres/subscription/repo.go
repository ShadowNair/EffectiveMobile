package subscription

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	modelsub "test_task/internal/domain/models/subscription"
	myerror "test_task/pkg/global_errors"
	"time"

	"github.com/google/uuid"
)

const (
	sqlTextForCreate = `INSERT INTO subscriptions(service_name, price, user_id, start_date, end_date)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, created_at, updated_at`
	sqlTextForGet = `SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
	FROM subscriptions
	WHERE id = $1`
	sqlTextForUpdate = `UPDATE subscriptions
	SET service_name=$2, price=$3, user_id=$4, start_date=$5, end_date=$6, updated_at=now()
	WHERE id=$1
	RETURNING id, service_name, price, user_id, start_date, end_date, created_at, updated_at`
	sqlTextForDelete = `DELETE FROM subscriptions
	WHERE id=$1`
	sqlTextForCount = `SELECT COUNT(*)
	FROM subscriptions
	WHERE ($1::uuid IS NULL OR user_id = $1)
	AND ($2::text IS NULL OR service_name = $2)`
	sqlTextForList = `SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
	FROM subscriptions
	WHERE ($1::uuid IS NULL OR user_id = $1)
	AND ($2::text IS NULL OR service_name = $2)
	ORDER BY start_date DESC, created_at DESC
	LIMIT $3 OFFSET $4`
	sqlTextForSum = `WITH params AS (
	SELECT $1::date AS p_from, $2::date AS p_to
	),
	filtered AS (
	SELECT
		price,
		GREATEST(start_date, (SELECT p_from FROM params)) AS start_eff,
		LEAST(COALESCE(end_date, (SELECT p_to FROM params)), (SELECT p_to FROM params)) AS end_eff
	FROM subscriptions
	WHERE ($3::uuid IS NULL OR user_id = $3)
		AND ($4::text IS NULL OR service_name = $4)
	)
	SELECT COALESCE(SUM(
	price * (
		(EXTRACT(YEAR FROM end_eff)::int * 12 + EXTRACT(MONTH FROM end_eff)::int) -
		(EXTRACT(YEAR FROM start_eff)::int * 12 + EXTRACT(MONTH FROM start_eff)::int) + 1
	)
	), 0)::bigint AS total
	FROM filtered
	WHERE end_eff >= start_eff;`
)

type DB struct {
	sql *sql.DB
}

func New(sql *sql.DB) *DB {
	return &DB{
		sql: sql,
	}
}

func (r *DB) Create(ctx context.Context, s modelsub.Subscription) (modelsub.Subscription, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var id uuid.UUID
	var created, updated time.Time

	err := r.sql.QueryRowContext(ctx, sqlTextForCreate,
		s.ServiceName,
		s.Price,
		s.UserID,
		s.StartDate,
		s.EndDate,
	).Scan(&id, &created, &updated)
	if err != nil {
		return modelsub.Subscription{}, fmt.Errorf("create subscription: %w", err)
	}

	s.ID = id
	s.CreatedAt = created
	s.UpdatedAt = updated
	return s, nil
}

func (r *DB) GetSub(ctx context.Context, id uuid.UUID) (modelsub.Subscription, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var s modelsub.Subscription
	err := r.sql.QueryRowContext(ctx, sqlTextForGet, id).Scan(
		&s.ID,
		&s.ServiceName,
		&s.Price,
		&s.UserID,
		&s.StartDate,
		&s.EndDate,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return modelsub.Subscription{}, myerror.ErrorNotFound
		}
		return modelsub.Subscription{}, fmt.Errorf("get subscription: %w", err)
	}
	return s, nil
}

func (r *DB) UpdateSub(ctx context.Context, id uuid.UUID, s modelsub.Subscription) (modelsub.Subscription, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var out modelsub.Subscription
	err := r.sql.QueryRowContext(ctx, sqlTextForUpdate,
		id,
		s.ServiceName,
		s.Price,
		s.UserID,
		s.StartDate,
		s.EndDate,
	).Scan(
		&out.ID,
		&out.ServiceName,
		&out.Price,
		&out.UserID,
		&out.StartDate,
		&out.EndDate,
		&out.CreatedAt,
		&out.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return modelsub.Subscription{}, myerror.ErrorNotFound
		}
		return modelsub.Subscription{}, fmt.Errorf("update subscription: %w", err)
	}
	return out, nil
}

func (r *DB) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tag, err := r.sql.ExecContext(ctx, sqlTextForDelete, id)
	if err != nil {
		return fmt.Errorf("delete subscription: %w", err)
	}
	rows, err := tag.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return myerror.ErrorNotFound
	}
	return nil
}

func (r *DB) List(ctx context.Context, f modelsub.ListFilter) ([]modelsub.Subscription, int, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if f.Limit <= 0 || f.Limit > 200 {
		f.Limit = 50
	}
	if f.Offset < 0 {
		f.Offset = 0
	}

	var total int
	if err := r.sql.QueryRowContext(ctx, sqlTextForCount, f.UserID, f.ServiceName).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("list count: %w", err)
	}

	rows, err := r.sql.QueryContext(ctx, sqlTextForList, f.UserID, f.ServiceName, f.Limit, f.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list query: %w", err)
	}
	defer rows.Close()
	out := make([]modelsub.Subscription, 0, f.Limit)
	for rows.Next() {
		var s modelsub.Subscription
		if err := rows.Scan(
			&s.ID,
			&s.ServiceName,
			&s.Price,
			&s.UserID,
			&s.StartDate,
			&s.EndDate,
			&s.CreatedAt,
			&s.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("list scan: %w", err)
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("list rows: %w", err)
	}

	return out, total, nil
}

func (r *DB) Summary(ctx context.Context, f modelsub.SummaryFilter) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 7*time.Second)
	defer cancel()

	var total int64
	if err := r.sql.QueryRowContext(ctx, sqlTextForSum, f.From, f.To, f.UserID, f.ServiceName).Scan(&total); err != nil {
		return 0, fmt.Errorf("summary query: %w", err)
	}
	return total, nil
}
