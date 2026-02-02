// scheduler.go - Timezone-aware scheduling engine

package expressparser

import (
	"time"
)

const (
	// DefaultMaxIterations is the maximum number of iterations to find next/prev time
	DefaultMaxIterations = 366 * 24 * 60 // 1 year of minutes

	// DefaultSearchYears is how many years to search for next/prev time
	DefaultSearchYears = 5
)

// Scheduler handles timezone-aware scheduling for cron expressions
type Scheduler struct {
	expr     *Expression
	location *time.Location
}

// SchedulerOption configures the scheduler
type SchedulerOption func(*Scheduler)

// WithTimezone sets the timezone for the scheduler
func WithTimezone(tz string) (SchedulerOption, error) {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return nil, ErrInvalidTimezone
	}
	return func(s *Scheduler) {
		s.location = loc
	}, nil
}

// WithLocation sets the location directly for the scheduler
func WithLocation(loc *time.Location) SchedulerOption {
	return func(s *Scheduler) {
		s.location = loc
	}
}

// NewScheduler creates a new scheduler for the given expression
func NewScheduler(expr *Expression, opts ...SchedulerOption) *Scheduler {
	s := &Scheduler{
		expr:     expr,
		location: time.UTC,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Location returns the scheduler's timezone
func (s *Scheduler) Location() *time.Location {
	return s.location
}

// Expression returns the scheduler's cron expression
func (s *Scheduler) Expression() *Expression {
	return s.expr
}

// Next returns the next time the cron expression matches after the given time
func (s *Scheduler) Next(from time.Time) (time.Time, error) {
	return s.NextN(from, 1)
}

// NextN returns the next N times the cron expression matches after the given time
func (s *Scheduler) NextN(from time.Time, n int) (time.Time, error) {
	if n < 1 {
		n = 1
	}

	// Convert to scheduler's timezone
	t := from.In(s.location)

	// Start from the next second
	t = t.Add(time.Second)
	t = t.Truncate(time.Second)

	maxTime := t.AddDate(DefaultSearchYears, 0, 0)
	iterations := 0

	for t.Before(maxTime) && iterations < DefaultMaxIterations*60 {
		iterations++

		// Find next matching time
		next, found := s.findNextMatch(t, maxTime)
		if !found {
			return time.Time{}, ErrNoNextRun
		}

		n--
		if n <= 0 {
			return next, nil
		}

		t = next.Add(time.Second)
	}

	return time.Time{}, ErrNoNextRun
}

// Previous returns the previous time the cron expression matched before the given time
func (s *Scheduler) Previous(from time.Time) (time.Time, error) {
	return s.PreviousN(from, 1)
}

// PreviousN returns the previous N times the cron expression matched before the given time
func (s *Scheduler) PreviousN(from time.Time, n int) (time.Time, error) {
	if n < 1 {
		n = 1
	}

	// Convert to scheduler's timezone
	t := from.In(s.location)

	// Start from the previous second
	t = t.Add(-time.Second)
	t = t.Truncate(time.Second)

	minTime := t.AddDate(-DefaultSearchYears, 0, 0)
	iterations := 0

	for t.After(minTime) && iterations < DefaultMaxIterations*60 {
		iterations++

		// Find previous matching time
		prev, found := s.findPrevMatch(t, minTime)
		if !found {
			return time.Time{}, ErrNoPreviousRun
		}

		n--
		if n <= 0 {
			return prev, nil
		}

		t = prev.Add(-time.Second)
	}

	return time.Time{}, ErrNoPreviousRun
}

// findNextMatch finds the next matching time starting from t
func (s *Scheduler) findNextMatch(t time.Time, maxTime time.Time) (time.Time, bool) {
	// Align to valid month
	t = s.alignToNextMonth(t)
	if t.After(maxTime) {
		return time.Time{}, false
	}

	for t.Before(maxTime) {
		// Check month
		if !s.expr.Month.Contains(int(t.Month())) {
			t = s.nextMonth(t)
			continue
		}

		// Check day (both day-of-month and day-of-week)
		if !s.matchesDay(t) {
			t = s.nextDay(t)
			continue
		}

		// Check hour
		if !s.expr.Hour.Contains(t.Hour()) {
			t = s.nextHour(t)
			continue
		}

		// Check minute
		if !s.expr.Minute.Contains(t.Minute()) {
			t = s.nextMinute(t)
			continue
		}

		// Check second
		if !s.expr.Second.Contains(t.Second()) {
			t = s.nextSecond(t)
			continue
		}

		// All fields match
		return t, true
	}

	return time.Time{}, false
}

// findPrevMatch finds the previous matching time starting from t
func (s *Scheduler) findPrevMatch(t time.Time, minTime time.Time) (time.Time, bool) {
	for t.After(minTime) {
		// Check month
		if !s.expr.Month.Contains(int(t.Month())) {
			t = s.prevMonth(t)
			continue
		}

		// Check day (both day-of-month and day-of-week)
		if !s.matchesDay(t) {
			t = s.prevDay(t)
			continue
		}

		// Check hour
		if !s.expr.Hour.Contains(t.Hour()) {
			t = s.prevHour(t)
			continue
		}

		// Check minute
		if !s.expr.Minute.Contains(t.Minute()) {
			t = s.prevMinute(t)
			continue
		}

		// Check second
		if !s.expr.Second.Contains(t.Second()) {
			t = s.prevSecond(t)
			continue
		}

		// All fields match
		return t, true
	}

	return time.Time{}, false
}

// matchesDay checks if the day matches considering special handling
func (s *Scheduler) matchesDay(t time.Time) bool {
	day := t.Day()
	weekday := int(t.Weekday())
	month := t.Month()
	year := t.Year()

	// Handle special day-of-month values
	if s.expr.HasSpecialDayHandling() {
		return s.matchesSpecialDay(t)
	}

	// Standard matching
	domMatch := s.expr.DayOfMonth.Contains(day)
	dowMatch := s.expr.DayOfWeek.Contains(weekday)

	// If both are wildcards, any day matches
	if s.expr.DayOfMonth.IsAll() && s.expr.DayOfWeek.IsAll() {
		return true
	}

	// If DOM is wildcard, only check DOW
	if s.expr.DayOfMonth.IsAll() {
		return dowMatch
	}

	// If DOW is wildcard, only check DOM
	if s.expr.DayOfWeek.IsAll() {
		return domMatch
	}

	// Neither is wildcard - OR logic
	_ = month // Used in special handling
	_ = year
	return domMatch || dowMatch
}

// matchesSpecialDay handles L, W, # specifiers
func (s *Scheduler) matchesSpecialDay(t time.Time) bool {
	day := t.Day()
	weekday := int(t.Weekday())
	month := t.Month()
	year := t.Year()

	// Get last day of month
	lastDay := s.lastDayOfMonth(year, month)

	// Check for "L" - last day of month
	if s.expr.HasLastDayOfMonth && s.expr.DayOfMonth.Values[32] {
		if day == lastDay {
			return true
		}
	}

	// Check for "L-N" - Nth day before end of month
	for v := range s.expr.DayOfMonth.Values {
		if v > 32 && v < 100 {
			offset := v - 32
			targetDay := lastDay - offset
			if targetDay > 0 && day == targetDay {
				return true
			}
		}
	}

	// Check for "LW" - last weekday of month
	if s.expr.HasLastWeekday && s.expr.DayOfMonth.Values[33] {
		lastWeekday := s.lastWeekdayOfMonth(year, month)
		if day == lastWeekday {
			return true
		}
	}

	// Check for "NW" - nearest weekday to day N
	if s.expr.HasNearestWeekday {
		for v := range s.expr.DayOfMonth.Values {
			if v >= 101 && v <= 131 {
				targetDay := v - 100
				nearest := s.nearestWeekday(year, month, targetDay)
				if day == nearest {
					return true
				}
			}
		}
	}

	// Check for "NL" - last N day of month
	if s.expr.HasLastDayOfWeek {
		for v := range s.expr.DayOfWeek.Values {
			if v >= 10 && v <= 16 {
				targetWeekday := v - 10
				lastOccurrence := s.lastWeekdayOccurrence(year, month, targetWeekday)
				if day == lastOccurrence && weekday == targetWeekday {
					return true
				}
			}
		}
	}

	// Check for "N#M" - Mth occurrence of day N
	if s.expr.HasNthDayOfWeek {
		for v := range s.expr.DayOfWeek.Values {
			if v >= 21 && v <= 75 {
				encoded := v - 20
				targetWeekday := encoded / 10
				occurrence := encoded % 10
				nthDay := s.nthWeekdayOfMonth(year, month, targetWeekday, occurrence)
				if nthDay > 0 && day == nthDay && weekday == targetWeekday {
					return true
				}
			}
		}
	}

	// Fall back to standard matching if no special matched
	domMatch := s.expr.DayOfMonth.Contains(day)
	dowMatch := s.expr.DayOfWeek.Contains(weekday)

	if s.expr.DayOfMonth.IsAll() && s.expr.DayOfWeek.IsAll() {
		return true
	}
	if s.expr.DayOfMonth.IsAll() {
		return dowMatch
	}
	if s.expr.DayOfWeek.IsAll() {
		return domMatch
	}
	return domMatch || dowMatch
}

// Helper methods for time navigation

func (s *Scheduler) alignToNextMonth(t time.Time) time.Time {
	// If current month is not valid, jump to next valid month
	if !s.expr.Month.Contains(int(t.Month())) {
		return s.nextMonth(t)
	}
	return t
}

func (s *Scheduler) nextMonth(t time.Time) time.Time {
	// Move to first day of next month at 00:00:00
	year := t.Year()
	month := int(t.Month())

	for i := 0; i < 12; i++ {
		month++
		if month > 12 {
			month = 1
			year++
		}
		if s.expr.Month.Contains(month) {
			return time.Date(year, time.Month(month), 1, 0, 0, 0, 0, s.location)
		}
	}

	// No valid month found in next 12 months, move to next year
	return time.Date(year+1, time.January, 1, 0, 0, 0, 0, s.location)
}

func (s *Scheduler) prevMonth(t time.Time) time.Time {
	// Move to last valid day/time of previous month
	year := t.Year()
	month := int(t.Month())

	for i := 0; i < 12; i++ {
		month--
		if month < 1 {
			month = 12
			year--
		}
		if s.expr.Month.Contains(month) {
			lastDay := s.lastDayOfMonth(year, time.Month(month))
			return time.Date(year, time.Month(month), lastDay, 23, 59, 59, 0, s.location)
		}
	}

	return time.Date(year-1, time.December, 31, 23, 59, 59, 0, s.location)
}

func (s *Scheduler) nextDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, s.location)
}

func (s *Scheduler) prevDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day()-1, 23, 59, 59, 0, s.location)
}

func (s *Scheduler) nextHour(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour()+1, 0, 0, 0, s.location)
}

func (s *Scheduler) prevHour(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour()-1, 59, 59, 0, s.location)
}

func (s *Scheduler) nextMinute(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute()+1, 0, 0, s.location)
}

func (s *Scheduler) prevMinute(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute()-1, 59, 0, s.location)
}

func (s *Scheduler) nextSecond(t time.Time) time.Time {
	return t.Add(time.Second)
}

func (s *Scheduler) prevSecond(t time.Time) time.Time {
	return t.Add(-time.Second)
}

// Calendar helper methods

func (s *Scheduler) lastDayOfMonth(year int, month time.Month) int {
	// Get first day of next month, then subtract one day
	firstOfNext := time.Date(year, month+1, 1, 0, 0, 0, 0, s.location)
	lastDay := firstOfNext.AddDate(0, 0, -1)
	return lastDay.Day()
}

func (s *Scheduler) lastWeekdayOfMonth(year int, month time.Month) int {
	lastDay := s.lastDayOfMonth(year, month)
	t := time.Date(year, month, lastDay, 0, 0, 0, 0, s.location)

	// Move back from last day until we find a weekday
	for t.Weekday() == time.Saturday || t.Weekday() == time.Sunday {
		t = t.AddDate(0, 0, -1)
	}
	return t.Day()
}

func (s *Scheduler) nearestWeekday(year int, month time.Month, day int) int {
	lastDay := s.lastDayOfMonth(year, month)
	if day > lastDay {
		day = lastDay
	}

	t := time.Date(year, month, day, 0, 0, 0, 0, s.location)
	weekday := t.Weekday()

	switch weekday {
	case time.Saturday:
		// Move to Friday (day before) unless it's in previous month
		if day > 1 {
			return day - 1
		}
		// Move to Monday (day + 2)
		return day + 2
	case time.Sunday:
		// Move to Monday (day after) unless it exceeds month
		if day < lastDay {
			return day + 1
		}
		// Move to Friday (day - 2)
		return day - 2
	default:
		return day
	}
}

func (s *Scheduler) lastWeekdayOccurrence(year int, month time.Month, targetWeekday int) int {
	lastDay := s.lastDayOfMonth(year, month)
	t := time.Date(year, month, lastDay, 0, 0, 0, 0, s.location)

	// Move back until we find the target weekday
	for int(t.Weekday()) != targetWeekday {
		t = t.AddDate(0, 0, -1)
	}
	return t.Day()
}

func (s *Scheduler) nthWeekdayOfMonth(year int, month time.Month, targetWeekday, n int) int {
	// Start from first day of month
	t := time.Date(year, month, 1, 0, 0, 0, 0, s.location)

	// Find first occurrence of target weekday
	for int(t.Weekday()) != targetWeekday {
		t = t.AddDate(0, 0, 1)
	}

	// Add (n-1) weeks
	t = t.AddDate(0, 0, (n-1)*7)

	// Check if still in same month
	if t.Month() != month {
		return 0 // Nth occurrence doesn't exist
	}

	return t.Day()
}

// NextNTimes returns the next n run times after the given time
func (s *Scheduler) NextNTimes(from time.Time, n int) ([]time.Time, error) {
	if n <= 0 {
		return nil, nil
	}

	results := make([]time.Time, 0, n)
	t := from

	for i := 0; i < n; i++ {
		next, err := s.Next(t)
		if err != nil {
			if len(results) > 0 {
				return results, nil
			}
			return nil, err
		}
		results = append(results, next)
		t = next
	}

	return results, nil
}

// PreviousNTimes returns the previous n run times before the given time
func (s *Scheduler) PreviousNTimes(from time.Time, n int) ([]time.Time, error) {
	if n <= 0 {
		return nil, nil
	}

	results := make([]time.Time, 0, n)
	t := from

	for i := 0; i < n; i++ {
		prev, err := s.Previous(t)
		if err != nil {
			if len(results) > 0 {
				return results, nil
			}
			return nil, err
		}
		results = append(results, prev)
		t = prev
	}

	return results, nil
}

// IsNow checks if the expression matches the current time (within 1 second)
func (s *Scheduler) IsNow() bool {
	now := time.Now().In(s.location)
	return s.expr.Matches(
		now.Second(),
		now.Minute(),
		now.Hour(),
		now.Day(),
		int(now.Month()),
		int(now.Weekday()),
	)
}

// IsDue checks if the expression matches the given time (within 1 second)
func (s *Scheduler) IsDue(t time.Time) bool {
	t = t.In(s.location)
	return s.expr.Matches(
		t.Second(),
		t.Minute(),
		t.Hour(),
		t.Day(),
		int(t.Month()),
		int(t.Weekday()),
	)
}
