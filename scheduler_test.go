package expressparser

import (
	"testing"
	"time"
)

func mustParseExpr(t *testing.T, expr string) *Expression {
	t.Helper()
	e, err := Parse(expr)
	if err != nil {
		t.Fatalf("Parse(%q) error = %v", expr, err)
	}
	return e
}

func TestScheduler_Next_SimpleDaily(t *testing.T) {
	expr := mustParseExpr(t, "0 9 * * *") // 09:00 every day
	s := NewScheduler(expr)

	from := time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC)
	got, err := s.Next(from)
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}

	want := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("Next() = %v, want %v", got, want)
	}
}

func TestScheduler_Next_FirstOfMonth(t *testing.T) {
	expr := mustParseExpr(t, "0 0 1 * *") // midnight on 1st of month
	s := NewScheduler(expr)

	from := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	got, err := s.Next(from)
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}

	want := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("Next() = %v, want %v", got, want)
	}
}

func TestScheduler_Previous_SimpleDaily(t *testing.T) {
	expr := mustParseExpr(t, "0 9 * * *") // 09:00 every day
	s := NewScheduler(expr)

	from := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	got, err := s.Previous(from)
	if err != nil {
		t.Fatalf("Previous() error = %v", err)
	}

	want := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("Previous() = %v, want %v", got, want)
	}
}

func TestScheduler_Timezone_Next(t *testing.T) {
	expr := mustParseExpr(t, "0 9 * * *") // 09:00 every day (local tz)
	tzOpt, err := WithTimezone("America/New_York")
	if err != nil {
		t.Fatalf("WithTimezone error = %v", err)
	}
	s := NewScheduler(expr, tzOpt)

	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("LoadLocation error = %v", err)
	}

	from := time.Date(2024, 1, 1, 8, 0, 0, 0, loc)
	got, err := s.Next(from)
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}

	gotLocal := got.In(loc)
	if gotLocal.Hour() != 9 || gotLocal.Minute() != 0 || gotLocal.Second() != 0 {
		t.Errorf("Next() local time = %v, want 09:00:00 in %v", gotLocal, loc)
	}
}

func TestScheduler_NextNTimes(t *testing.T) {
	expr := mustParseExpr(t, "0 9 * * 1-5") // 09:00 on weekdays
	s := NewScheduler(expr)

	from := time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC) // Monday
	n := 3
	times, err := s.NextNTimes(from, n)
	if err != nil {
		t.Fatalf("NextNTimes() error = %v", err)
	}
	if len(times) != n {
		t.Fatalf("NextNTimes() len = %d, want %d", len(times), n)
	}

	for i, tt := range times {
		if !expr.Matches(
			tt.Second(),
			tt.Minute(),
			tt.Hour(),
			tt.Day(),
			int(tt.Month()),
			int(tt.Weekday()),
		) {
			t.Errorf("NextNTimes()[%d] = %v does not match expression", i, tt)
		}
		if i > 0 && !tt.After(times[i-1]) {
			t.Errorf("times not strictly increasing: %v >= %v", times[i-1], tt)
		}
	}
}

func TestScheduler_PreviousNTimes(t *testing.T) {
	expr := mustParseExpr(t, "0 9 * * 1-5") // 09:00 on weekdays
	s := NewScheduler(expr)

	from := time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC) // Wednesday
	n := 3
	times, err := s.PreviousNTimes(from, n)
	if err != nil {
		t.Fatalf("PreviousNTimes() error = %v", err)
	}
	if len(times) != n {
		t.Fatalf("PreviousNTimes() len = %d, want %d", len(times), n)
	}

	for i, tt := range times {
		if !expr.Matches(
			tt.Second(),
			tt.Minute(),
			tt.Hour(),
			tt.Day(),
			int(tt.Month()),
			int(tt.Weekday()),
		) {
			t.Errorf("PreviousNTimes()[%d] = %v does not match expression", i, tt)
		}
		if i > 0 && !tt.Before(times[i-1]) {
			t.Errorf("times not strictly decreasing: %v <= %v", times[i-1], tt)
		}
	}
}

func TestScheduler_IsDue(t *testing.T) {
	expr := mustParseExpr(t, "0 9 * * 1-5") // 09:00 on weekdays
	s := NewScheduler(expr)

	// Monday 09:00 UTC
	t1 := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC) // Monday
	if !s.IsDue(t1) {
		t.Errorf("IsDue(%v) = false, want true", t1)
	}

	// Sunday 09:00 UTC
	t2 := time.Date(2024, 1, 7, 9, 0, 0, 0, time.UTC) // Sunday
	if s.IsDue(t2) {
		t.Errorf("IsDue(%v) = true, want false", t2)
	}
}
