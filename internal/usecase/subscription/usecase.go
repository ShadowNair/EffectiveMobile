package subscription

import (
	"context"
	modelsub "test_task/internal/domain/models/subscription"

	"github.com/google/uuid"
)

type RepoI interface {
	Create(ctx context.Context, s modelsub.Subscription) (modelsub.Subscription, error)
	GetSub(ctx context.Context, id uuid.UUID) (modelsub.Subscription, error)
	UpdateSub(ctx context.Context, id uuid.UUID, s modelsub.Subscription) (modelsub.Subscription, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, f modelsub.ListFilter) ([]modelsub.Subscription, int, error)
	Summary(ctx context.Context, f modelsub.SummaryFilter) (int64, error)
}

type Usecase struct {
	repo RepoI
}

func New(repo RepoI) *Usecase {
	return &Usecase{
		repo: repo,
	}
}

func (u *Usecase) Create(ctx context.Context, s modelsub.Subscription) (modelsub.Subscription, error) {
	return u.repo.Create(ctx, s)
}

func (u *Usecase) GetSub(ctx context.Context, id uuid.UUID) (modelsub.Subscription, error) {
	return u.repo.GetSub(ctx, id)
}

func (u *Usecase) UpdateSub(ctx context.Context, id uuid.UUID, s modelsub.Subscription) (modelsub.Subscription, error) {
	return u.repo.UpdateSub(ctx, id, s)
}

func (u *Usecase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.repo.Delete(ctx, id)
}

func (u *Usecase) List(ctx context.Context, f modelsub.ListFilter) ([]modelsub.Subscription, int, error) {
	return u.repo.List(ctx, f)
}

func (u *Usecase) Summary(ctx context.Context, f modelsub.SummaryFilter) (int64, error) {
	return u.repo.Summary(ctx, f)
}
