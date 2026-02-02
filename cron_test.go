// cron_test.go - Tests for core cron expression parsing

package expressparser

import (
	"testing"
)

func TestParseCron_StandardExpressions(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		wantErr  bool
		exprType ExpressionType
	}{
		{
			name:     "every minute",
			expr:     "* * * * *",
			wantErr:  false,
			exprType: StandardCron,
		},
		{
			name:     "every hour",
			expr:     "0 * * * *",
			wantErr:  false,
			exprType: StandardCron,
		},
		{
			name:     "every day at midnight",
			expr:     "0 0 * * *",
			wantErr:  false,
			exprType: StandardCron,
		},
		{
			name:     "every monday at 9am",
			expr:     "0 9 * * 1",
			wantErr:  false,
			exprType: StandardCron,
		},
		{
			name:     "weekdays at 9am",
			expr:     "0 9 * * 1-5",
			wantErr:  false,
			exprType: StandardCron,
		},
		{
			name:     "first of month at midnight",
			expr:     "0 0 1 * *",
			wantErr:  false,
			exprType: StandardCron,
		},
		{
			name:     "every 15 minutes",
			expr:     "*/15 * * * *",
			wantErr:  false,
			exprType: StandardCron,
		},
		{
			name:     "complex expression",
			expr:     "0,30 9-17 * * 1-5",
			wantErr:  false,
			exprType: StandardCron,
		},
		{
			name:     "with month name",
			expr:     "0 0 1 JAN *",
			wantErr:  false,
			exprType: StandardCron,
		},
		{
			name:     "with day name",
			expr:     "0 9 * * MON",
			wantErr:  false,
			exprType: StandardCron,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parseCron(tt.expr)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if expr.Type != tt.exprType {
				t.Errorf("expected type %v, got %v", tt.exprType, expr.Type)
			}

			if expr.Raw != tt.expr {
				t.Errorf("expected raw %q, got %q", tt.expr, expr.Raw)
			}
		})
	}
}

func TestParseCron_ExtendedExpressions(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		wantErr  bool
		exprType ExpressionType
	}{
		{
			name:     "every second",
			expr:     "* * * * * *",
			wantErr:  false,
			exprType: ExtendedCron,
		},
		{
			name:     "every minute at 30 seconds",
			expr:     "30 * * * * *",
			wantErr:  false,
			exprType: ExtendedCron,
		},
		{
			name:     "specific time with seconds",
			expr:     "0 0 9 * * 1-5",
			wantErr:  false,
			exprType: ExtendedCron,
		},
		{
			name:     "every 10 seconds",
			expr:     "*/10 * * * * *",
			wantErr:  false,
			exprType: ExtendedCron,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parseCron(tt.expr)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if expr.Type != tt.exprType {
				t.Errorf("expected type %v, got %v", tt.exprType, expr.Type)
			}
		})
	}
}

func TestParseCron_PredefinedExpressions(t *testing.T) {
	tests := []struct {
		name           string
		expr           string
		expectedMinute int
		expectedHour   int
		expectedDOM    int
		expectedMonth  int
		expectedDOW    bool // true if all days
		wantErr        bool
	}{
		{
			name:           "@yearly",
			expr:           "@yearly",
			expectedMinute: 0,
			expectedHour:   0,
			expectedDOM:    1,
			expectedMonth:  1,
			expectedDOW:    true,
			wantErr:        false,
		},
		{
			name:           "@annually",
			expr:           "@annually",
			expectedMinute: 0,
			expectedHour:   0,
			expectedDOM:    1,
			expectedMonth:  1,
			expectedDOW:    true,
			wantErr:        false,
		},
		{
			name:           "@monthly",
			expr:           "@monthly",
			expectedMinute: 0,
			expectedHour:   0,
			expectedDOM:    1,
			expectedMonth:  0, // all months
			expectedDOW:    true,
			wantErr:        false,
		},
		{
			name:           "@weekly",
			expr:           "@weekly",
			expectedMinute: 0,
			expectedHour:   0,
			expectedDOM:    0, // all days
			expectedMonth:  0, // all months
			expectedDOW:    false,
			wantErr:        false,
		},
		{
			name:           "@daily",
			expr:           "@daily",
			expectedMinute: 0,
			expectedHour:   0,
			expectedDOM:    0, // all days
			expectedMonth:  0, // all months
			expectedDOW:    true,
			wantErr:        false,
		},
		{
			name:           "@midnight",
			expr:           "@midnight",
			expectedMinute: 0,
			expectedHour:   0,
			expectedDOM:    0, // all days
			expectedMonth:  0, // all months
			expectedDOW:    true,
			wantErr:        false,
		},
		{
			name:           "@hourly",
			expr:           "@hourly",
			expectedMinute: 0,
			expectedHour:   0, // all hours
			expectedDOM:    0, // all days
			expectedMonth:  0, // all months
			expectedDOW:    true,
			wantErr:        false,
		},
		{
			name:    "@invalid",
			expr:    "@invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parseCron(tt.expr)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Check minute
			if tt.expectedMinute > 0 {
				if !expr.Minute.Contains(tt.expectedMinute) {
					t.Errorf("expected minute %d", tt.expectedMinute)
				}
			}

			// Check hour
			if tt.expectedHour > 0 {
				if !expr.Hour.Contains(tt.expectedHour) {
					t.Errorf("expected hour %d", tt.expectedHour)
				}
			}

			// Check day of month
			if tt.expectedDOM > 0 {
				if !expr.DayOfMonth.Contains(tt.expectedDOM) {
					t.Errorf("expected day of month %d", tt.expectedDOM)
				}
			}

			// Check month
			if tt.expectedMonth > 0 {
				if !expr.Month.Contains(tt.expectedMonth) {
					t.Errorf("expected month %d", tt.expectedMonth)
				}
			}
		})
	}
}

func TestParseCron_InvalidExpressions(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		wantErr bool
	}{
		{
			name:    "empty expression",
			expr:    "",
			wantErr: true,
		},
		{
			name:    "too few fields",
			expr:    "* * *",
			wantErr: true,
		},
		{
			name:    "too many fields",
			expr:    "* * * * * * *",
			wantErr: true,
		},
		{
			name:    "invalid minute",
			expr:    "60 * * * *",
			wantErr: true,
		},
		{
			name:    "invalid hour",
			expr:    "* 24 * * *",
			wantErr: true,
		},
		{
			name:    "invalid day of month",
			expr:    "* * 32 * *",
			wantErr: true,
		},
		{
			name:    "invalid month",
			expr:    "* * * 13 *",
			wantErr: true,
		},
		{
			name:    "invalid day of week",
			expr:    "* * * * 7",
			wantErr: true,
		},
		{
			name:    "invalid characters",
			expr:    "* * * * abc",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			expr:    "   ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseCron(tt.expr)

			if tt.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestParseCron_SpecialDayHandling(t *testing.T) {
	tests := []struct {
		name                  string
		expr                  string
		hasLastDayOfMonth     bool
		hasLastWeekday        bool
		hasNearestWeekday     bool
		hasNthDayOfWeek       bool
		hasLastDayOfWeek      bool
		hasSpecialDayHandling bool
	}{
		{
			name:                  "last day of month",
			expr:                  "0 0 L * *",
			hasLastDayOfMonth:     true,
			hasSpecialDayHandling: true,
		},
		{
			name:                  "last weekday of month",
			expr:                  "0 0 LW * *",
			hasLastWeekday:        true,
			hasSpecialDayHandling: true,
		},
		{
			name:                  "nearest weekday",
			expr:                  "0 0 15W * *",
			hasNearestWeekday:     true,
			hasSpecialDayHandling: true,
		},
		{
			name:                  "nth day of week",
			expr:                  "0 0 * * 1#3",
			hasNthDayOfWeek:       true,
			hasSpecialDayHandling: true,
		},
		{
			name:                  "last day of week",
			expr:                  "0 0 * * 5L",
			hasLastDayOfWeek:      true,
			hasSpecialDayHandling: true,
		},
		{
			name:                  "no special handling",
			expr:                  "0 0 * * *",
			hasSpecialDayHandling: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parseCron(tt.expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if expr.HasLastDayOfMonth != tt.hasLastDayOfMonth {
				t.Errorf("HasLastDayOfMonth = %v, want %v", expr.HasLastDayOfMonth, tt.hasLastDayOfMonth)
			}

			if expr.HasLastWeekday != tt.hasLastWeekday {
				t.Errorf("HasLastWeekday = %v, want %v", expr.HasLastWeekday, tt.hasLastWeekday)
			}

			if expr.HasNearestWeekday != tt.hasNearestWeekday {
				t.Errorf("HasNearestWeekday = %v, want %v", expr.HasNearestWeekday, tt.hasNearestWeekday)
			}

			if expr.HasNthDayOfWeek != tt.hasNthDayOfWeek {
				t.Errorf("HasNthDayOfWeek = %v, want %v", expr.HasNthDayOfWeek, tt.hasNthDayOfWeek)
			}

			if expr.HasLastDayOfWeek != tt.hasLastDayOfWeek {
				t.Errorf("HasLastDayOfWeek = %v, want %v", expr.HasLastDayOfWeek, tt.hasLastDayOfWeek)
			}

			if expr.HasSpecialDayHandling() != tt.hasSpecialDayHandling {
				t.Errorf("HasSpecialDayHandling() = %v, want %v", expr.HasSpecialDayHandling(), tt.hasSpecialDayHandling)
			}
		})
	}
}

func TestExpression_Matches(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		second  int
		minute  int
		hour    int
		day     int
		month   int
		weekday int
		matches bool
	}{
		{
			name:    "every minute matches",
			expr:    "* * * * *",
			second:  0,
			minute:  30,
			hour:    12,
			day:     15,
			month:   6,
			weekday: 3,
			matches: true,
		},
		{
			name:    "specific time matches",
			expr:    "30 9 * * *",
			second:  0,
			minute:  30,
			hour:    9,
			day:     15,
			month:   6,
			weekday: 3,
			matches: true,
		},
		{
			name:    "specific time does not match minute",
			expr:    "30 9 * * *",
			second:  0,
			minute:  31,
			hour:    9,
			day:     15,
			month:   6,
			weekday: 3,
			matches: false,
		},
		{
			name:    "specific time does not match hour",
			expr:    "30 9 * * *",
			second:  0,
			minute:  30,
			hour:    10,
			day:     15,
			month:   6,
			weekday: 3,
			matches: false,
		},
		{
			name:    "weekday matches",
			expr:    "0 9 * * 1-5",
			second:  0,
			minute:  0,
			hour:    9,
			day:     15,
			month:   6,
			weekday: 3, // Wednesday
			matches: true,
		},
		{
			name:    "weekday does not match (Sunday)",
			expr:    "0 9 * * 1-5",
			second:  0,
			minute:  0,
			hour:    9,
			day:     15,
			month:   6,
			weekday: 0, // Sunday
			matches: false,
		},
		{
			name:    "specific month matches",
			expr:    "0 0 1 6 *",
			second:  0,
			minute:  0,
			hour:    0,
			day:     1,
			month:   6,
			weekday: 3,
			matches: true,
		},
		{
			name:    "specific month does not match",
			expr:    "0 0 1 6 *",
			second:  0,
			minute:  0,
			hour:    0,
			day:     1,
			month:   7,
			weekday: 3,
			matches: false,
		},
		{
			name:    "day of month OR day of week (DOM matches)",
			expr:    "0 0 15 * 1",
			second:  0,
			minute:  0,
			hour:    0,
			day:     15,
			month:   6,
			weekday: 3, // Wednesday, not Monday
			matches: true,
		},
		{
			name:    "day of month OR day of week (DOW matches)",
			expr:    "0 0 15 * 1",
			second:  0,
			minute:  0,
			hour:    0,
			day:     10, // Not 15th
			month:   6,
			weekday: 1, // Monday
			matches: true,
		},
		{
			name:    "day of month OR day of week (neither matches)",
			expr:    "0 0 15 * 1",
			second:  0,
			minute:  0,
			hour:    0,
			day:     10, // Not 15th
			month:   6,
			weekday: 3, // Wednesday, not Monday
			matches: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parseCron(tt.expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			got := expr.Matches(tt.second, tt.minute, tt.hour, tt.day, tt.month, tt.weekday)
			if got != tt.matches {
				t.Errorf("Matches() = %v, want %v", got, tt.matches)
			}
		})
	}
}

func TestExpression_GetValues(t *testing.T) {
	expr, err := parseCron("0,30 9-17 1,15 6,12 1-5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check minutes
	expectedMinutes := []int{0, 30}
	if !intSliceEqual(expr.GetMinutes(), expectedMinutes) {
		t.Errorf("GetMinutes() = %v, want %v", expr.GetMinutes(), expectedMinutes)
	}

	// Check hours
	expectedHours := []int{9, 10, 11, 12, 13, 14, 15, 16, 17}
	if !intSliceEqual(expr.GetHours(), expectedHours) {
		t.Errorf("GetHours() = %v, want %v", expr.GetHours(), expectedHours)
	}

	// Check days of month
	expectedDays := []int{1, 15}
	if !intSliceEqual(expr.GetDaysOfMonth(), expectedDays) {
		t.Errorf("GetDaysOfMonth() = %v, want %v", expr.GetDaysOfMonth(), expectedDays)
	}

	// Check months
	expectedMonths := []int{6, 12}
	if !intSliceEqual(expr.GetMonths(), expectedMonths) {
		t.Errorf("GetMonths() = %v, want %v", expr.GetMonths(), expectedMonths)
	}

	// Check days of week
	expectedDOW := []int{1, 2, 3, 4, 5}
	if !intSliceEqual(expr.GetDaysOfWeek(), expectedDOW) {
		t.Errorf("GetDaysOfWeek() = %v, want %v", expr.GetDaysOfWeek(), expectedDOW)
	}
}

func TestExpression_String(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		expected string
	}{
		{
			name:     "standard expression",
			expr:     "0 9 * * 1-5",
			expected: "0 9 * * 1-5",
		},
		{
			name:     "every minute",
			expr:     "* * * * *",
			expected: "* * * * *",
		},
		{
			name:     "extended expression",
			expr:     "30 0 9 * * 1-5",
			expected: "30 0 9 * * 1-5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parseCron(tt.expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if expr.String() != tt.expected {
				t.Errorf("String() = %q, want %q", expr.String(), tt.expected)
			}
		})
	}
}

func TestExpression_IsStandardExtended(t *testing.T) {
	standardExpr, _ := parseCron("* * * * *")
	if !standardExpr.IsStandard() {
		t.Errorf("expected IsStandard() to be true")
	}
	if standardExpr.IsExtended() {
		t.Errorf("expected IsExtended() to be false")
	}

	extendedExpr, _ := parseCron("* * * * * *")
	if extendedExpr.IsStandard() {
		t.Errorf("expected IsStandard() to be false")
	}
	if !extendedExpr.IsExtended() {
		t.Errorf("expected IsExtended() to be true")
	}
}

func TestExpression_FieldStrings(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		expected []string
	}{
		{
			name:     "standard expression",
			expr:     "0 9 * * 1-5",
			expected: []string{"0", "9", "*", "*", "1-5"},
		},
		{
			name:     "extended expression",
			expr:     "30 0 9 * * 1-5",
			expected: []string{"30", "0", "9", "*", "*", "1-5"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parseCron(tt.expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			got := expr.FieldStrings()
			if len(got) != len(tt.expected) {
				t.Errorf("FieldStrings() length = %d, want %d", len(got), len(tt.expected))
				return
			}

			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("FieldStrings()[%d] = %q, want %q", i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestParseCron_WhitespaceHandling(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		wantErr bool
	}{
		{
			name:    "leading whitespace",
			expr:    "  * * * * *",
			wantErr: false,
		},
		{
			name:    "trailing whitespace",
			expr:    "* * * * *  ",
			wantErr: false,
		},
		{
			name:    "extra whitespace between fields",
			expr:    "*  *  *  *  *",
			wantErr: false,
		},
		{
			name:    "tabs between fields",
			expr:    "*\t*\t*\t*\t*",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseCron(tt.expr)

			if tt.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func BenchmarkParseCron_Simple(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = parseCron("0 9 * * *")
	}
}

func BenchmarkParseCron_Complex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = parseCron("0,15,30,45 9-17 1,15 1-6 1-5")
	}
}

func BenchmarkParseCron_WithSteps(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = parseCron("*/15 */2 * * *")
	}
}

func BenchmarkExpression_Matches(b *testing.B) {
	expr, _ := parseCron("0 9 * * 1-5")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		expr.Matches(0, 0, 9, 15, 6, 3)
	}
}
