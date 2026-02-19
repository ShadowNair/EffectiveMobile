package monthyear

import (
	"fmt"
	"time"
)

const MonthYearLayout = "01-2006"

func ParseMonthYear(s string) (time.Time, error) {
	t, err := time.Parse(MonthYearLayout, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date %q, expected MM-YYYY: %w", s, err)
	}
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC), nil
}

func FormatMonthYear(t time.Time) string {
	t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	return t.Format(MonthYearLayout)
}
