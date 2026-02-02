// descriptor.go - Human-readable description generator for cron expressions

package expressparser

import (
	"fmt"
	"strings"
)

// DescriptionOptions configures how descriptions are generated
type DescriptionOptions struct {
	// Use24HourTime uses 24-hour format instead of 12-hour with AM/PM
	Use24HourTime bool

	// Verbose generates more detailed descriptions
	Verbose bool

	// Locale for localization (future use)
	Locale string
}

// DefaultDescriptionOptions returns the default description options
func DefaultDescriptionOptions() DescriptionOptions {
	return DescriptionOptions{
		Use24HourTime: false,
		Verbose:       false,
		Locale:        "en",
	}
}

// Descriptor generates human-readable descriptions of cron expressions
type Descriptor struct {
	expr *Expression
	opts DescriptionOptions
}

// NewDescriptor creates a new descriptor for the given expression
func NewDescriptor(expr *Expression, opts ...DescriptionOptions) *Descriptor {
	options := DefaultDescriptionOptions()
	if len(opts) > 0 {
		options = opts[0]
	}

	return &Descriptor{
		expr: expr,
		opts: options,
	}
}

// Describe returns a human-readable description of the cron expression
func (d *Descriptor) Describe() string {
	parts := make([]string, 0)

	// Describe time (second, minute, hour)
	timePart := d.describeTime()
	if timePart != "" {
		parts = append(parts, timePart)
	}

	// Describe day of month
	domPart := d.describeDayOfMonth()
	if domPart != "" {
		parts = append(parts, domPart)
	}

	// Describe month
	monthPart := d.describeMonth()
	if monthPart != "" {
		parts = append(parts, monthPart)
	}

	// Describe day of week
	dowPart := d.describeDayOfWeek()
	if dowPart != "" {
		parts = append(parts, dowPart)
	}

	if len(parts) == 0 {
		return "Every minute"
	}

	result := strings.Join(parts, ", ")
	return capitalizeFirst(result)
}

// describeTime generates description for second, minute, and hour fields
func (d *Descriptor) describeTime() string {
	secondAll := d.expr.Second.IsAll()
	minuteAll := d.expr.Minute.IsAll()
	hourAll := d.expr.Hour.IsAll()

	// Every second
	if secondAll && minuteAll && hourAll && d.expr.Type == ExtendedCron {
		return "every second"
	}

	// Every minute
	if minuteAll && hourAll {
		if d.expr.Type == ExtendedCron && !secondAll {
			seconds := d.expr.GetSeconds()
			return fmt.Sprintf("at second %s of every minute", d.formatList(seconds))
		}
		return "every minute"
	}

	// Every hour at specific minute
	if hourAll && !minuteAll {
		minutes := d.expr.GetMinutes()
		if len(minutes) == 1 {
			if minutes[0] == 0 {
				return "every hour"
			}
			return fmt.Sprintf("at %d minute(s) past every hour", minutes[0])
		}
		return fmt.Sprintf("at minute %s of every hour", d.formatList(minutes))
	}

	// Specific times
	hours := d.expr.GetHours()
	minutes := d.expr.GetMinutes()
	seconds := d.expr.GetSeconds()

	// Single specific time
	if len(hours) == 1 && len(minutes) == 1 {
		timeStr := d.formatTime(hours[0], minutes[0])
		if d.expr.Type == ExtendedCron && len(seconds) == 1 && seconds[0] != 0 {
			return fmt.Sprintf("at %s and %d second(s)", timeStr, seconds[0])
		}
		return fmt.Sprintf("at %s", timeStr)
	}

	// Multiple specific hours, single minute
	if len(minutes) == 1 {
		hourStrs := make([]string, len(hours))
		for i, h := range hours {
			hourStrs[i] = d.formatTime(h, minutes[0])
		}
		return fmt.Sprintf("at %s", strings.Join(hourStrs, ", "))
	}

	// Multiple times
	if len(hours) <= 3 && len(minutes) <= 3 {
		return fmt.Sprintf("at minute %s of %s", d.formatList(minutes), d.formatHours(hours))
	}

	// Complex time specification
	return fmt.Sprintf("at minute %s, during hour %s", d.formatList(minutes), d.formatList(hours))
}

// describeDayOfMonth generates description for day-of-month field
func (d *Descriptor) describeDayOfMonth() string {
	if d.expr.DayOfMonth.IsAll() {
		return ""
	}

	// Handle special values
	if d.expr.HasLastDayOfMonth {
		if d.expr.DayOfMonth.Values[32] {
			return "on the last day of the month"
		}
	}

	if d.expr.HasLastWeekday {
		if d.expr.DayOfMonth.Values[33] {
			return "on the last weekday of the month"
		}
	}

	// Handle L-N (offset from last day)
	for v := range d.expr.DayOfMonth.Values {
		if v > 32 && v < 100 {
			offset := v - 32
			if offset == 1 {
				return "on the day before the last day of the month"
			}
			return fmt.Sprintf("on the %d days before the last day of the month", offset)
		}
	}

	// Handle NW (nearest weekday)
	if d.expr.HasNearestWeekday {
		for v := range d.expr.DayOfMonth.Values {
			if v >= 101 && v <= 131 {
				day := v - 100
				return fmt.Sprintf("on the weekday nearest to day %d of the month", day)
			}
		}
	}

	// Regular days
	days := d.expr.GetDaysOfMonth()
	if len(days) == 0 {
		return ""
	}

	if len(days) == 1 {
		return fmt.Sprintf("on day %d of the month", days[0])
	}

	// Check for range
	if isConsecutive(days) {
		return fmt.Sprintf("on days %d through %d of the month", days[0], days[len(days)-1])
	}

	return fmt.Sprintf("on day %s of the month", d.formatOrdinalList(days))
}

// describeMonth generates description for month field
func (d *Descriptor) describeMonth() string {
	if d.expr.Month.IsAll() {
		return ""
	}

	months := d.expr.GetMonths()
	if len(months) == 0 {
		return ""
	}

	monthNames := make([]string, len(months))
	for i, m := range months {
		monthNames[i] = monthToName(m)
	}

	if len(months) == 1 {
		return fmt.Sprintf("in %s", monthNames[0])
	}

	// Check for consecutive months
	if isConsecutive(months) {
		return fmt.Sprintf("from %s through %s", monthNames[0], monthNames[len(monthNames)-1])
	}

	return fmt.Sprintf("in %s", strings.Join(monthNames, ", "))
}

// describeDayOfWeek generates description for day-of-week field
func (d *Descriptor) describeDayOfWeek() string {
	if d.expr.DayOfWeek.IsAll() {
		return ""
	}

	// Handle special Nth day of week
	if d.expr.HasNthDayOfWeek {
		for v := range d.expr.DayOfWeek.Values {
			if v >= 21 && v <= 75 {
				encoded := v - 20
				weekday := encoded / 10
				occurrence := encoded % 10
				return fmt.Sprintf("on the %s %s of the month",
					ordinal(occurrence), dayToName(weekday))
			}
		}
	}

	// Handle last day of week in month
	if d.expr.HasLastDayOfWeek {
		for v := range d.expr.DayOfWeek.Values {
			if v >= 10 && v <= 16 {
				weekday := v - 10
				return fmt.Sprintf("on the last %s of the month", dayToName(weekday))
			}
		}
	}

	// Regular days of week
	days := d.expr.GetDaysOfWeek()
	if len(days) == 0 {
		return ""
	}

	dayNames := make([]string, len(days))
	for i, day := range days {
		dayNames[i] = dayToName(day)
	}

	// Check for weekdays (Mon-Fri)
	if len(days) == 5 && isConsecutive(days) && days[0] == 1 && days[4] == 5 {
		return "on weekdays"
	}

	// Check for weekend
	if len(days) == 2 && days[0] == 0 && days[1] == 6 {
		return "on weekends"
	}

	if len(days) == 1 {
		return fmt.Sprintf("on %s", dayNames[0])
	}

	// Check for consecutive days
	if isConsecutive(days) {
		return fmt.Sprintf("from %s through %s", dayNames[0], dayNames[len(dayNames)-1])
	}

	return fmt.Sprintf("on %s", strings.Join(dayNames, ", "))
}

// Helper methods

func (d *Descriptor) formatTime(hour, minute int) string {
	if d.opts.Use24HourTime {
		return fmt.Sprintf("%02d:%02d", hour, minute)
	}

	period := "AM"
	displayHour := hour

	if hour == 0 {
		displayHour = 12
	} else if hour == 12 {
		period = "PM"
	} else if hour > 12 {
		displayHour = hour - 12
		period = "PM"
	}

	return fmt.Sprintf("%d:%02d %s", displayHour, minute, period)
}

func (d *Descriptor) formatHours(hours []int) string {
	if len(hours) == 0 {
		return ""
	}

	hourStrs := make([]string, len(hours))
	for i, h := range hours {
		if d.opts.Use24HourTime {
			hourStrs[i] = fmt.Sprintf("%02d:00", h)
		} else {
			period := "AM"
			displayHour := h
			if h == 0 {
				displayHour = 12
			} else if h == 12 {
				period = "PM"
			} else if h > 12 {
				displayHour = h - 12
				period = "PM"
			}
			hourStrs[i] = fmt.Sprintf("%d %s", displayHour, period)
		}
	}

	return strings.Join(hourStrs, ", ")
}

func (d *Descriptor) formatList(values []int) string {
	if len(values) == 0 {
		return ""
	}

	// Check for step pattern
	if len(values) > 2 {
		step := values[1] - values[0]
		isStep := step > 0
		for i := 2; i < len(values) && isStep; i++ {
			if values[i]-values[i-1] != step {
				isStep = false
			}
		}
		if isStep && step > 1 {
			return fmt.Sprintf("every %d starting at %d", step, values[0])
		}
	}

	// Check for range
	if isConsecutive(values) && len(values) > 2 {
		return fmt.Sprintf("%d through %d", values[0], values[len(values)-1])
	}

	// List individual values
	strs := make([]string, len(values))
	for i, v := range values {
		strs[i] = fmt.Sprintf("%d", v)
	}

	if len(strs) == 2 {
		return strs[0] + " and " + strs[1]
	}

	if len(strs) > 2 {
		return strings.Join(strs[:len(strs)-1], ", ") + ", and " + strs[len(strs)-1]
	}

	return strs[0]
}

func (d *Descriptor) formatOrdinalList(values []int) string {
	if len(values) == 0 {
		return ""
	}

	strs := make([]string, len(values))
	for i, v := range values {
		strs[i] = ordinal(v)
	}

	if len(strs) == 2 {
		return strs[0] + " and " + strs[1]
	}

	if len(strs) > 2 {
		return strings.Join(strs[:len(strs)-1], ", ") + ", and " + strs[len(strs)-1]
	}

	return strs[0]
}

// Utility functions

func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(string(s[0])) + s[1:]
}

func isConsecutive(values []int) bool {
	if len(values) < 2 {
		return true
	}
	for i := 1; i < len(values); i++ {
		if values[i]-values[i-1] != 1 {
			return false
		}
	}
	return true
}

func ordinal(n int) string {
	suffix := "th"
	switch n % 10 {
	case 1:
		if n%100 != 11 {
			suffix = "st"
		}
	case 2:
		if n%100 != 12 {
			suffix = "nd"
		}
	case 3:
		if n%100 != 13 {
			suffix = "rd"
		}
	}
	return fmt.Sprintf("%d%s", n, suffix)
}

func monthToName(month int) string {
	names := []string{
		"", "January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December",
	}
	if month >= 1 && month <= 12 {
		return names[month]
	}
	return fmt.Sprintf("Month %d", month)
}

func dayToName(day int) string {
	names := []string{
		"Sunday", "Monday", "Tuesday", "Wednesday",
		"Thursday", "Friday", "Saturday",
	}
	if day >= 0 && day <= 6 {
		return names[day]
	}
	return fmt.Sprintf("Day %d", day)
}

// Describe is a convenience function to describe an expression
func Describe(expr *Expression) string {
	return NewDescriptor(expr).Describe()
}

// DescribeWithOptions describes an expression with custom options
func DescribeWithOptions(expr *Expression, opts DescriptionOptions) string {
	return NewDescriptor(expr, opts).Describe()
}
