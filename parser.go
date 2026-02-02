// parser.go - Main public API for expressparser

package expressparser

import (
	"time"
)

// Parse parses a cron expression string and returns an Expression
//
// Supported formats:
//   - Standard 5-field: "minute hour day-of-month month day-of-week"
//   - Extended 6-field: "second minute hour day-of-month month day-of-week"
//   - Predefined: @yearly, @annually, @monthly, @weekly, @daily, @midnight, @hourly
//
// Supported special characters:
//   - * : any value
//   - , : value list separator (e.g., "1,3,5")
//   - - : range of values (e.g., "1-5")
//   - / : step values (e.g., "*/15" or "0-30/5")
//   - ? : any value (same as *)
//   - L : last (day of month or day of week)
//   - W : nearest weekday (day of month only)
//   - # : nth occurrence (day of week only, e.g., "1#3" = third Monday)
//
// Examples:
//
//	Parse("0 9 * * 1-5")     // 9 AM on weekdays
//	Parse("*/15 * * * *")    // Every 15 minutes
//	Parse("0 0 L * *")       // Midnight on last day of month
//	Parse("0 9 * * 1#2")     // 9 AM on second Monday
//	Parse("@daily")          // Midnight every day
func Parse(expr string) (*Expression, error) {
	return parseCron(expr)
}

// ParseWithSeconds parses a 6-field cron expression with seconds
//
// Format: "second minute hour day-of-month month day-of-week"
//
// Example:
//
//	ParseWithSeconds("30 0 9 * * 1-5")  // 9:00:30 AM on weekdays
func ParseWithSeconds(expr string) (*Expression, error) {
	return parseCron(expr, WithSeconds())
}

// MustParse parses a cron expression and panics if it fails
//
// Use this for known-good expressions, typically defined as constants.
//
// Example:
//
//	var dailyBackup = expressparser.MustParse("0 2 * * *")
func MustParse(expr string) *Expression {
	e, err := Parse(expr)
	if err != nil {
		panic(err)
	}
	return e
}

// MustParseWithSeconds parses a 6-field cron expression and panics if it fails
func MustParseWithSeconds(expr string) *Expression {
	e, err := ParseWithSeconds(expr)
	if err != nil {
		panic(err)
	}
	return e
}

// Validate checks if a cron expression is valid without returning the parsed result
//
// Returns nil if valid, error otherwise.
//
// Example:
//
//	if err := expressparser.Validate("0 9 * * *"); err != nil {
//	    log.Fatal("Invalid cron:", err)
//	}
func Validate(expr string) error {
	_, err := Parse(expr)
	return err
}

// ValidateWithSeconds checks if a 6-field cron expression is valid
func ValidateWithSeconds(expr string) error {
	_, err := ParseWithSeconds(expr)
	return err
}

// Next returns the next time the cron expression matches after the given time
//
// Uses UTC timezone by default. For timezone support, use NewScheduler.
//
// Example:
//
//	next, err := expressparser.Next("0 9 * * *", time.Now())
func Next(expr string, from time.Time) (time.Time, error) {
	e, err := Parse(expr)
	if err != nil {
		return time.Time{}, err
	}
	return NewScheduler(e).Next(from)
}

// NextN returns the next n times the cron expression matches
//
// Example:
//
//	times, err := expressparser.NextN("0 9 * * *", time.Now(), 5)
func NextN(expr string, from time.Time, n int) ([]time.Time, error) {
	e, err := Parse(expr)
	if err != nil {
		return nil, err
	}
	return NewScheduler(e).NextNTimes(from, n)
}

// Previous returns the previous time the cron expression matched before the given time
//
// Example:
//
//	prev, err := expressparser.Previous("0 9 * * *", time.Now())
func Previous(expr string, from time.Time) (time.Time, error) {
	e, err := Parse(expr)
	if err != nil {
		return time.Time{}, err
	}
	return NewScheduler(e).Previous(from)
}

// PreviousN returns the previous n times the cron expression matched
//
// Example:
//
//	times, err := expressparser.PreviousN("0 9 * * *", time.Now(), 5)
func PreviousN(expr string, from time.Time, n int) ([]time.Time, error) {
	e, err := Parse(expr)
	if err != nil {
		return nil, err
	}
	return NewScheduler(e).PreviousNTimes(from, n)
}

// NextInTimezone returns the next matching time in the specified timezone
//
// Example:
//
//	next, err := expressparser.NextInTimezone("0 9 * * *", time.Now(), "America/New_York")
func NextInTimezone(expr string, from time.Time, timezone string) (time.Time, error) {
	e, err := Parse(expr)
	if err != nil {
		return time.Time{}, err
	}

	tzOpt, err := WithTimezone(timezone)
	if err != nil {
		return time.Time{}, err
	}

	return NewScheduler(e, tzOpt).Next(from)
}

// IsDue checks if the cron expression matches the given time
//
// Example:
//
//	if expressparser.IsDue("0 9 * * *", time.Now()) {
//	    fmt.Println("It's 9 AM!")
//	}
func IsDue(expr string, t time.Time) (bool, error) {
	e, err := Parse(expr)
	if err != nil {
		return false, err
	}
	return NewScheduler(e).IsDue(t), nil
}

// IsDueNow checks if the cron expression matches the current time
//
// Example:
//
//	if due, _ := expressparser.IsDueNow("0 9 * * *"); due {
//	    fmt.Println("It's 9 AM!")
//	}
func IsDueNow(expr string) (bool, error) {
	return IsDue(expr, time.Now())
}

// Config holds configuration for creating a cron runner
type Config struct {
	// Expression is the cron expression string
	Expression string

	// Timezone is the IANA timezone name (e.g., "America/New_York")
	// Defaults to UTC if empty
	Timezone string

	// Location is the timezone location (alternative to Timezone string)
	// If both Timezone and Location are set, Location takes precedence
	Location *time.Location
}

// NewSchedulerFromConfig creates a scheduler from a Config
//
// Example:
//
//	scheduler, err := expressparser.NewSchedulerFromConfig(expressparser.Config{
//	    Expression: "0 9 * * 1-5",
//	    Timezone:   "America/New_York",
//	})
func NewSchedulerFromConfig(cfg Config) (*Scheduler, error) {
	expr, err := Parse(cfg.Expression)
	if err != nil {
		return nil, err
	}

	var opts []SchedulerOption

	if cfg.Location != nil {
		opts = append(opts, WithLocation(cfg.Location))
	} else if cfg.Timezone != "" {
		tzOpt, err := WithTimezone(cfg.Timezone)
		if err != nil {
			return nil, err
		}
		opts = append(opts, tzOpt)
	}

	return NewScheduler(expr, opts...), nil
}

// Schedule represents a parsed and configured cron schedule
// This is a convenience wrapper combining Expression and Scheduler
type Schedule struct {
	expression *Expression
	scheduler  *Scheduler
}

// NewSchedule creates a new Schedule from an expression string
//
// Example:
//
//	schedule, err := expressparser.NewSchedule("0 9 * * 1-5")
//	next := schedule.Next(time.Now())
func NewSchedule(expr string, opts ...SchedulerOption) (*Schedule, error) {
	e, err := Parse(expr)
	if err != nil {
		return nil, err
	}

	return &Schedule{
		expression: e,
		scheduler:  NewScheduler(e, opts...),
	}, nil
}

// NewScheduleInTimezone creates a new Schedule with timezone support
//
// Example:
//
//	schedule, err := expressparser.NewScheduleInTimezone("0 9 * * 1-5", "America/New_York")
func NewScheduleInTimezone(expr string, timezone string) (*Schedule, error) {
	e, err := Parse(expr)
	if err != nil {
		return nil, err
	}

	tzOpt, err := WithTimezone(timezone)
	if err != nil {
		return nil, err
	}

	return &Schedule{
		expression: e,
		scheduler:  NewScheduler(e, tzOpt),
	}, nil
}

// Expression returns the underlying parsed expression
func (s *Schedule) Expression() *Expression {
	return s.expression
}

// Scheduler returns the underlying scheduler
func (s *Schedule) Scheduler() *Scheduler {
	return s.scheduler
}

// Next returns the next matching time after from
func (s *Schedule) Next(from time.Time) (time.Time, error) {
	return s.scheduler.Next(from)
}

// Previous returns the previous matching time before from
func (s *Schedule) Previous(from time.Time) (time.Time, error) {
	return s.scheduler.Previous(from)
}

// NextN returns the next n matching times
func (s *Schedule) NextN(from time.Time, n int) ([]time.Time, error) {
	return s.scheduler.NextNTimes(from, n)
}

// PreviousN returns the previous n matching times
func (s *Schedule) PreviousN(from time.Time, n int) ([]time.Time, error) {
	return s.scheduler.PreviousNTimes(from, n)
}

// IsDue checks if the schedule matches the given time
func (s *Schedule) IsDue(t time.Time) bool {
	return s.scheduler.IsDue(t)
}

// IsNow checks if the schedule matches the current time
func (s *Schedule) IsNow() bool {
	return s.scheduler.IsNow()
}

// Describe returns a human-readable description
func (s *Schedule) Describe() string {
	return Describe(s.expression)
}

// DescribeWithOptions returns a human-readable description with custom options
func (s *Schedule) DescribeWithOptions(opts DescriptionOptions) string {
	return DescribeWithOptions(s.expression, opts)
}

// String returns the original cron expression
func (s *Schedule) String() string {
	return s.expression.String()
}

// Timezone returns the scheduler's timezone
func (s *Schedule) Timezone() *time.Location {
	return s.scheduler.Location()
}
