package monthyear

import (
	"testing"
	"time"
)

func TestParseMonthYear_OK(t *testing.T) {
	got, err := ParseMonthYear("07-2025")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	want := time.Date(2025, time.July, 1, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestParseMonthYear_BadFormat(t *testing.T) {
	_, err := ParseMonthYear("2025-07")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestFormatMonthYear(t *testing.T) {
	in := time.Date(2025, time.December, 31, 23, 59, 59, 0, time.UTC)
	got := FormatMonthYear(in)
	if got != "12-2025" {
		t.Fatalf("want %q, got %q", "12-2025", got)
	}
}
