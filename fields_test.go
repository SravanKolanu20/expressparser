// fields_test.go - Tests for field parsing

package expressparser

import (
	"testing"
)

func TestFieldParser_ParseSingleValue(t *testing.T) {
	tests := []struct {
		name      string
		fieldType FieldType
		expr      string
		expected  []int
		wantErr   bool
	}{
		{
			name:      "minute single value",
			fieldType: FieldMinute,
			expr:      "5",
			expected:  []int{5},
			wantErr:   false,
		},
		{
			name:      "hour single value",
			fieldType: FieldHour,
			expr:      "12",
			expected:  []int{12},
			wantErr:   false,
		},
		{
			name:      "day of month single value",
			fieldType: FieldDayOfMonth,
			expr:      "15",
			expected:  []int{15},
			wantErr:   false,
		},
		{
			name:      "month single value",
			fieldType: FieldMonth,
			expr:      "6",
			expected:  []int{6},
			wantErr:   false,
		},
		{
			name:      "day of week single value",
			fieldType: FieldDayOfWeek,
			expr:      "3",
			expected:  []int{3},
			wantErr:   false,
		},
		{
			name:      "minute out of range high",
			fieldType: FieldMinute,
			expr:      "60",
			expected:  nil,
			wantErr:   true,
		},
		{
			name:      "hour out of range high",
			fieldType: FieldHour,
			expr:      "24",
			expected:  nil,
			wantErr:   true,
		},
		{
			name:      "day of month out of range low",
			fieldType: FieldDayOfMonth,
			expr:      "0",
			expected:  nil,
			wantErr:   true,
		},
		{
			name:      "month out of range high",
			fieldType: FieldMonth,
			expr:      "13",
			expected:  nil,
			wantErr:   true,
		},
		{
			name:      "invalid value",
			fieldType: FieldMinute,
			expr:      "abc",
			expected:  nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewFieldParser(tt.fieldType)
			field, err := parser.Parse(tt.expr)

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

			got := field.All()
			if !intSliceEqual(got, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestFieldParser_ParseWildcard(t *testing.T) {
	tests := []struct {
		name      string
		fieldType FieldType
		expr      string
		minVal    int
		maxVal    int
	}{
		{
			name:      "minute wildcard *",
			fieldType: FieldMinute,
			expr:      "*",
			minVal:    0,
			maxVal:    59,
		},
		{
			name:      "hour wildcard *",
			fieldType: FieldHour,
			expr:      "*",
			minVal:    0,
			maxVal:    23,
		},
		{
			name:      "day of month wildcard *",
			fieldType: FieldDayOfMonth,
			expr:      "*",
			minVal:    1,
			maxVal:    31,
		},
		{
			name:      "month wildcard *",
			fieldType: FieldMonth,
			expr:      "*",
			minVal:    1,
			maxVal:    12,
		},
		{
			name:      "day of week wildcard *",
			fieldType: FieldDayOfWeek,
			expr:      "*",
			minVal:    0,
			maxVal:    6,
		},
		{
			name:      "minute wildcard ?",
			fieldType: FieldMinute,
			expr:      "?",
			minVal:    0,
			maxVal:    59,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewFieldParser(tt.fieldType)
			field, err := parser.Parse(tt.expr)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if !field.IsAll() {
				t.Errorf("expected IsAll() to be true")
			}

			if field.Min() != tt.minVal {
				t.Errorf("expected min %d, got %d", tt.minVal, field.Min())
			}

			if field.Max() != tt.maxVal {
				t.Errorf("expected max %d, got %d", tt.maxVal, field.Max())
			}
		})
	}
}

func TestFieldParser_ParseRange(t *testing.T) {
	tests := []struct {
		name      string
		fieldType FieldType
		expr      string
		expected  []int
		wantErr   bool
	}{
		{
			name:      "minute range",
			fieldType: FieldMinute,
			expr:      "10-15",
			expected:  []int{10, 11, 12, 13, 14, 15},
			wantErr:   false,
		},
		{
			name:      "hour range",
			fieldType: FieldHour,
			expr:      "9-17",
			expected:  []int{9, 10, 11, 12, 13, 14, 15, 16, 17},
			wantErr:   false,
		},
		{
			name:      "day of week range",
			fieldType: FieldDayOfWeek,
			expr:      "1-5",
			expected:  []int{1, 2, 3, 4, 5},
			wantErr:   false,
		},
		{
			name:      "single value range",
			fieldType: FieldMinute,
			expr:      "5-5",
			expected:  []int{5},
			wantErr:   false,
		},
		{
			name:      "invalid range start > end",
			fieldType: FieldMinute,
			expr:      "15-10",
			expected:  nil,
			wantErr:   true,
		},
		{
			name:      "range out of bounds",
			fieldType: FieldHour,
			expr:      "20-25",
			expected:  nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewFieldParser(tt.fieldType)
			field, err := parser.Parse(tt.expr)

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

			got := field.All()
			if !intSliceEqual(got, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestFieldParser_ParseStep(t *testing.T) {
	tests := []struct {
		name      string
		fieldType FieldType
		expr      string
		expected  []int
		wantErr   bool
	}{
		{
			name:      "every 15 minutes",
			fieldType: FieldMinute,
			expr:      "*/15",
			expected:  []int{0, 15, 30, 45},
			wantErr:   false,
		},
		{
			name:      "every 6 hours",
			fieldType: FieldHour,
			expr:      "*/6",
			expected:  []int{0, 6, 12, 18},
			wantErr:   false,
		},
		{
			name:      "every 2 from 10-20",
			fieldType: FieldMinute,
			expr:      "10-20/2",
			expected:  []int{10, 12, 14, 16, 18, 20},
			wantErr:   false,
		},
		{
			name:      "every 5 starting from 10",
			fieldType: FieldMinute,
			expr:      "10/5",
			expected:  []int{10, 15, 20, 25, 30, 35, 40, 45, 50, 55},
			wantErr:   false,
		},
		{
			name:      "step of 1 (every value)",
			fieldType: FieldDayOfWeek,
			expr:      "*/1",
			expected:  []int{0, 1, 2, 3, 4, 5, 6},
			wantErr:   false,
		},
		{
			name:      "invalid step 0",
			fieldType: FieldMinute,
			expr:      "*/0",
			expected:  nil,
			wantErr:   true,
		},
		{
			name:      "invalid negative step",
			fieldType: FieldMinute,
			expr:      "*/-5",
			expected:  nil,
			wantErr:   true,
		},
		{
			name:      "invalid step non-numeric",
			fieldType: FieldMinute,
			expr:      "*/abc",
			expected:  nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewFieldParser(tt.fieldType)
			field, err := parser.Parse(tt.expr)

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

			got := field.All()
			if !intSliceEqual(got, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestFieldParser_ParseList(t *testing.T) {
	tests := []struct {
		name      string
		fieldType FieldType
		expr      string
		expected  []int
		wantErr   bool
	}{
		{
			name:      "simple list",
			fieldType: FieldMinute,
			expr:      "0,15,30,45",
			expected:  []int{0, 15, 30, 45},
			wantErr:   false,
		},
		{
			name:      "list with ranges",
			fieldType: FieldHour,
			expr:      "9-12,14-17",
			expected:  []int{9, 10, 11, 12, 14, 15, 16, 17},
			wantErr:   false,
		},
		{
			name:      "list with steps",
			fieldType: FieldMinute,
			expr:      "0-10/2,30-40/2",
			expected:  []int{0, 2, 4, 6, 8, 10, 30, 32, 34, 36, 38, 40},
			wantErr:   false,
		},
		{
			name:      "mixed list",
			fieldType: FieldHour,
			expr:      "0,6,12,18",
			expected:  []int{0, 6, 12, 18},
			wantErr:   false,
		},
		{
			name:      "list with duplicates",
			fieldType: FieldMinute,
			expr:      "5,5,10,10",
			expected:  []int{5, 10},
			wantErr:   false,
		},
		{
			name:      "list with invalid value",
			fieldType: FieldHour,
			expr:      "9,25,12",
			expected:  nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewFieldParser(tt.fieldType)
			field, err := parser.Parse(tt.expr)

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

			got := field.All()
			if !intSliceEqual(got, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestFieldParser_ParseNamedValues(t *testing.T) {
	tests := []struct {
		name      string
		fieldType FieldType
		expr      string
		expected  []int
		wantErr   bool
	}{
		{
			name:      "month name JAN",
			fieldType: FieldMonth,
			expr:      "JAN",
			expected:  []int{1},
			wantErr:   false,
		},
		{
			name:      "month name lowercase jan",
			fieldType: FieldMonth,
			expr:      "jan",
			expected:  []int{1},
			wantErr:   false,
		},
		{
			name:      "month full name January",
			fieldType: FieldMonth,
			expr:      "JANUARY",
			expected:  []int{1},
			wantErr:   false,
		},
		{
			name:      "month range JAN-MAR",
			fieldType: FieldMonth,
			expr:      "JAN-MAR",
			expected:  []int{1, 2, 3},
			wantErr:   false,
		},
		{
			name:      "month list",
			fieldType: FieldMonth,
			expr:      "JAN,APR,JUL,OCT",
			expected:  []int{1, 4, 7, 10},
			wantErr:   false,
		},
		{
			name:      "day name MON",
			fieldType: FieldDayOfWeek,
			expr:      "MON",
			expected:  []int{1},
			wantErr:   false,
		},
		{
			name:      "day name SUN",
			fieldType: FieldDayOfWeek,
			expr:      "SUN",
			expected:  []int{0},
			wantErr:   false,
		},
		{
			name:      "day range MON-FRI",
			fieldType: FieldDayOfWeek,
			expr:      "MON-FRI",
			expected:  []int{1, 2, 3, 4, 5},
			wantErr:   false,
		},
		{
			name:      "day full name Monday",
			fieldType: FieldDayOfWeek,
			expr:      "MONDAY",
			expected:  []int{1},
			wantErr:   false,
		},
		{
			name:      "mixed day list",
			fieldType: FieldDayOfWeek,
			expr:      "MON,WED,FRI",
			expected:  []int{1, 3, 5},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewFieldParser(tt.fieldType)
			field, err := parser.Parse(tt.expr)

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

			got := field.All()
			if !intSliceEqual(got, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestFieldParser_ParseSpecialCharacters(t *testing.T) {
	tests := []struct {
		name      string
		fieldType FieldType
		expr      string
		checkFunc func(*Field) bool
		wantErr   bool
	}{
		{
			name:      "L for last day of month",
			fieldType: FieldDayOfMonth,
			expr:      "L",
			checkFunc: func(f *Field) bool { return f.Values[32] },
			wantErr:   false,
		},
		{
			name:      "LW for last weekday",
			fieldType: FieldDayOfMonth,
			expr:      "LW",
			checkFunc: func(f *Field) bool { return f.Values[33] },
			wantErr:   false,
		},
		{
			name:      "L-3 for 3 days before end",
			fieldType: FieldDayOfMonth,
			expr:      "L-3",
			checkFunc: func(f *Field) bool { return f.Values[35] }, // 32 + 3
			wantErr:   false,
		},
		{
			name:      "15W for nearest weekday to 15th",
			fieldType: FieldDayOfMonth,
			expr:      "15W",
			checkFunc: func(f *Field) bool { return f.Values[115] }, // 100 + 15
			wantErr:   false,
		},
		{
			name:      "5L for last Friday",
			fieldType: FieldDayOfWeek,
			expr:      "5L",
			checkFunc: func(f *Field) bool { return f.Values[15] }, // 10 + 5
			wantErr:   false,
		},
		{
			name:      "1#3 for third Monday",
			fieldType: FieldDayOfWeek,
			expr:      "1#3",
			checkFunc: func(f *Field) bool { return f.Values[33] }, // 20 + 1*10 + 3
			wantErr:   false,
		},
		{
			name:      "2#2 for second Tuesday",
			fieldType: FieldDayOfWeek,
			expr:      "2#2",
			checkFunc: func(f *Field) bool { return f.Values[42] }, // 20 + 2*10 + 2
			wantErr:   false,
		},
		{
			name:      "L not allowed in minute",
			fieldType: FieldMinute,
			expr:      "L",
			checkFunc: nil,
			wantErr:   true,
		},
		{
			name:      "W not allowed in hour",
			fieldType: FieldHour,
			expr:      "15W",
			checkFunc: nil,
			wantErr:   true,
		},
		{
			name:      "# not allowed in month",
			fieldType: FieldMonth,
			expr:      "1#2",
			checkFunc: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewFieldParser(tt.fieldType)
			field, err := parser.Parse(tt.expr)

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

			if tt.checkFunc != nil && !tt.checkFunc(field) {
				t.Errorf("check function failed for field values: %v", field.Values)
			}
		})
	}
}

func TestFieldParser_ParseEmpty(t *testing.T) {
	parser := NewFieldParser(FieldMinute)
	_, err := parser.Parse("")

	if err == nil {
		t.Errorf("expected error for empty expression")
	}
}

func TestField_Contains(t *testing.T) {
	parser := NewFieldParser(FieldMinute)
	field, _ := parser.Parse("0,15,30,45")

	tests := []struct {
		value    int
		expected bool
	}{
		{0, true},
		{15, true},
		{30, true},
		{45, true},
		{1, false},
		{14, false},
		{59, false},
	}

	for _, tt := range tests {
		got := field.Contains(tt.value)
		if got != tt.expected {
			t.Errorf("Contains(%d) = %v, want %v", tt.value, got, tt.expected)
		}
	}
}

func TestField_MinMax(t *testing.T) {
	tests := []struct {
		name        string
		expr        string
		expectedMin int
		expectedMax int
	}{
		{
			name:        "single value",
			expr:        "30",
			expectedMin: 30,
			expectedMax: 30,
		},
		{
			name:        "range",
			expr:        "10-20",
			expectedMin: 10,
			expectedMax: 20,
		},
		{
			name:        "list",
			expr:        "5,15,25,35",
			expectedMin: 5,
			expectedMax: 35,
		},
		{
			name:        "wildcard",
			expr:        "*",
			expectedMin: 0,
			expectedMax: 59,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewFieldParser(FieldMinute)
			field, err := parser.Parse(tt.expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if field.Min() != tt.expectedMin {
				t.Errorf("Min() = %d, want %d", field.Min(), tt.expectedMin)
			}

			if field.Max() != tt.expectedMax {
				t.Errorf("Max() = %d, want %d", field.Max(), tt.expectedMax)
			}
		})
	}
}

// Helper function to compare int slices
func intSliceEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
