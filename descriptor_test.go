package expressparser

import "testing"

func TestDescribe_EveryMinute(t *testing.T) {
	expr, err := Parse("* * * * *")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	got := Describe(expr)
	want := "Every minute"

	if got != want {
		t.Errorf("Describe() = %q, want %q", got, want)
	}
}

func TestDescribe_WeekdaysAtNineAM(t *testing.T) {
	expr, err := Parse("0 9 * * 1-5")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	got := Describe(expr)
	want := "At 9:00 AM, on weekdays"

	if got != want {
		t.Errorf("Describe() = %q, want %q", got, want)
	}
}

func TestDescribeWithOptions_24HourTime(t *testing.T) {
	expr, err := Parse("0 9 * * 1-5")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	opts := DescriptionOptions{Use24HourTime: true}
	got := DescribeWithOptions(expr, opts)
	want := "At 09:00, on weekdays"

	if got != want {
		t.Errorf("DescribeWithOptions() = %q, want %q", got, want)
	}
}

func TestDescribe_MonthAndDayOfMonth(t *testing.T) {
	expr, err := Parse("0 0 1 6 *") // 1st of June at midnight
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	got := Describe(expr)
	// Example expected shape, adjust if you change description rules:
	// "At 12:00 AM, on day 1 of the month, in June"
	if got == "" {
		t.Errorf("Describe() returned empty string, want non-empty description")
	}
}
