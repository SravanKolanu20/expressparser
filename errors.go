// errors.go - Custom error types for expressparser

package expressparser

import (
	"errors"
	"fmt"
)

// Sentinel errors for common error cases
var (
	// ErrEmptyExpression is returned when an empty cron expression is provided
	ErrEmptyExpression = errors.New("cron expression cannot be empty")

	// ErrInvalidFieldCount is returned when the expression has wrong number of fields
	ErrInvalidFieldCount = errors.New("invalid number of fields in cron expression")

	// ErrInvalidTimezone is returned when an invalid timezone is provided
	ErrInvalidTimezone = errors.New("invalid timezone")

	// ErrNoNextRun is returned when no next run time can be calculated
	ErrNoNextRun = errors.New("no next run time found within search range")

	// ErrNoPreviousRun is returned when no previous run time can be calculated
	ErrNoPreviousRun = errors.New("no previous run time found within search range")
)

// ParseError represents an error that occurred during parsing
type ParseError struct {
	Expression string // The original expression that failed to parse
	Field      string // The field that caused the error (minute, hour, etc.)
	Value      string // The value that caused the error
	Reason     string // Human-readable reason for the error
}

// Error implements the error interface
func (e *ParseError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("parse error in %s field: %q - %s", e.Field, e.Value, e.Reason)
	}
	return fmt.Sprintf("parse error: %q - %s", e.Expression, e.Reason)
}

// NewParseError creates a new ParseError
func NewParseError(expression, field, value, reason string) *ParseError {
	return &ParseError{
		Expression: expression,
		Field:      field,
		Value:      value,
		Reason:     reason,
	}
}

// FieldError represents an error in a specific cron field
type FieldError struct {
	Field  FieldType // The type of field (Minute, Hour, etc.)
	Value  string    // The problematic value
	Min    int       // Minimum allowed value
	Max    int       // Maximum allowed value
	Reason string    // Reason for the error
}

// Error implements the error interface
func (e *FieldError) Error() string {
	return fmt.Sprintf("invalid %s field %q: %s (allowed range: %d-%d)",
		e.Field, e.Value, e.Reason, e.Min, e.Max)
}

// NewFieldError creates a new FieldError
func NewFieldError(field FieldType, value, reason string) *FieldError {
	bounds := fieldBounds[field]
	return &FieldError{
		Field:  field,
		Value:  value,
		Min:    bounds.min,
		Max:    bounds.max,
		Reason: reason,
	}
}

// RangeError represents an invalid range error
type RangeError struct {
	Field FieldType
	Start int
	End   int
}

// Error implements the error interface
func (e *RangeError) Error() string {
	return fmt.Sprintf("invalid range in %s field: start (%d) is greater than end (%d)",
		e.Field, e.Start, e.End)
}

// StepError represents an invalid step value error
type StepError struct {
	Field FieldType
	Step  int
}

// Error implements the error interface
func (e *StepError) Error() string {
	return fmt.Sprintf("invalid step value in %s field: %d (must be positive)", e.Field, e.Step)
}

// FieldType represents the type of cron field
type FieldType string

const (
	FieldSecond     FieldType = "second"
	FieldMinute     FieldType = "minute"
	FieldHour       FieldType = "hour"
	FieldDayOfMonth FieldType = "day-of-month"
	FieldMonth      FieldType = "month"
	FieldDayOfWeek  FieldType = "day-of-week"
)

// fieldBound defines the min and max values for a field
type fieldBound struct {
	min int
	max int
}

// fieldBounds maps field types to their allowed ranges
var fieldBounds = map[FieldType]fieldBound{
	FieldSecond:     {0, 59},
	FieldMinute:     {0, 59},
	FieldHour:       {0, 23},
	FieldDayOfMonth: {1, 31},
	FieldMonth:      {1, 12},
	FieldDayOfWeek:  {0, 6}, // 0 = Sunday, 6 = Saturday
}

// IsParseError checks if an error is a ParseError
func IsParseError(err error) bool {
	var parseErr *ParseError
	return errors.As(err, &parseErr)
}

// IsFieldError checks if an error is a FieldError
func IsFieldError(err error) bool {
	var fieldErr *FieldError
	return errors.As(err, &fieldErr)
}
