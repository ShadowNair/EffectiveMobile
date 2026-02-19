package subscription

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	modelsub "test_task/internal/domain/models/subscription"
	myerror "test_task/pkg/global_errors"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestRepo_Create_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := New(db)

	id := uuid.New()
	userID := uuid.New()
	now := time.Now().UTC()

	s := modelsub.Subscription{
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      userID,
		StartDate:   time.Date(2025, time.July, 1, 0, 0, 0, 0, time.UTC),
		EndDate:     nil,
	}

	mock.ExpectQuery(regexp.QuoteMeta(sqlTextForCreate)).
		WithArgs(s.ServiceName, s.Price, s.UserID, s.StartDate, s.EndDate).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(id.String(), now, now),
		)

	got, err := repo.Create(context.Background(), s)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if got.ID != id {
		t.Fatalf("id mismatch: %v", got.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestRepo_GetSub_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := New(db)

	id := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(sqlTextForGet)).
		WithArgs(id).
		WillReturnError(sql.ErrNoRows)

	_, err = repo.GetSub(context.Background(), id)
	if err == nil {
		t.Fatalf("expected error")
	}
	if err != myerror.ErrorNotFound {
		t.Fatalf("want ErrorNotFound, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestRepo_Summary_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := New(db)

	userID := uuid.New()
	from := time.Date(2025, time.July, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, time.December, 1, 0, 0, 0, 0, time.UTC)

	service := "Yandex Plus"

	mock.ExpectQuery(regexp.QuoteMeta(sqlTextForSum)).
		WithArgs(from, to, &userID, &service).
		WillReturnRows(sqlmock.NewRows([]string{"total"}).AddRow(int64(2400)))

	total, err := repo.Summary(context.Background(), modelsub.SummaryFilter{
		From:        from,
		To:          to,
		UserID:      &userID,
		ServiceName: &service,
	})
	if err != nil {
		t.Fatalf("Summary error: %v", err)
	}
	if total != 2400 {
		t.Fatalf("want 2400, got %d", total)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
