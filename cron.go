package expressparser

import (
	"strings"
)

type ExpressionType int

const (
	StandardCron ExpressionType = 5
	ExtendedCron ExpressionType = 6
)

type Expression struct {
	Raw               string
	Type              ExpressionType
	Second            *Field
	Minute            *Field
	Hour              *Field
	DayOfMonth        *Field
	Month             *Field
	DayOfWeek         *Field
	HasLastDayOfMonth bool
	HasLastWeekday    bool
	HasNearestWeekday bool
	HasNthDayOfWeek   bool
	HasLastDayOfWeek  bool
}

var predefinedExpressions = map[string]string{
	"@yearly":   "0 0 1 1 *",
	"@annually": "0 0 1 1 *",
	"@monthly":  "0 0 1 * *",
	"@weekly":   "0 0 * * 0",
	"@daily":    "0 0 * * *",
	"@midnight": "0 0 * * *",
	"@hourly":   "0 * * * *",
}

type cronParser struct {
	seconds bool
}

type ParserOption func(*cronParser)

func WithSeconds() ParserOption {
	return func(p *cronParser) {
		p.seconds = true
	}
}

func parseCron(expr string, opts ...ParserOption) (*Expression, error) {
	parser := &cronParser{seconds: false}
	for _, opt := range opts {
		opt(parser)
	}

	expr = strings.TrimSpace(expr)

	if expr == "" {
		return nil, ErrEmptyExpression
	}

	if strings.HasPrefix(expr, "@") {
		predefined, ok := predefinedExpressions[strings.ToLower(expr)]
		if !ok {
			return nil, NewParseError(expr, "", "", "unknown predefined expression")
		}
		expr = predefined
	}

	fields := strings.Fields(expr)
	fieldCount := len(fields)

	if fieldCount < 5 || fieldCount > 6 {
		return nil, ErrInvalidFieldCount
	}

	result := &Expression{Raw: expr}

	var secondExpr, minuteExpr, hourExpr, domExpr, monthExpr, dowExpr string

	if fieldCount == 6 {
		result.Type = ExtendedCron
		secondExpr = fields[0]
		minuteExpr = fields[1]
		hourExpr = fields[2]
		domExpr = fields[3]
		monthExpr = fields[4]
		dowExpr = fields[5]
	} else {
		result.Type = StandardCron
		secondExpr = "0"
		minuteExpr = fields[0]
		hourExpr = fields[1]
		domExpr = fields[2]
		monthExpr = fields[3]
		dowExpr = fields[4]
	}

	var err error

	result.Second, err = NewFieldParser(FieldSecond).Parse(secondExpr)
	if err != nil {
		return nil, err
	}

	result.Minute, err = NewFieldParser(FieldMinute).Parse(minuteExpr)
	if err != nil {
		return nil, err
	}

	result.Hour, err = NewFieldParser(FieldHour).Parse(hourExpr)
	if err != nil {
		return nil, err
	}

	result.DayOfMonth, err = NewFieldParser(FieldDayOfMonth).Parse(domExpr)
	if err != nil {
		return nil, err
	}

	result.Month, err = NewFieldParser(FieldMonth).Parse(monthExpr)
	if err != nil {
		return nil, err
	}

	result.DayOfWeek, err = NewFieldParser(FieldDayOfWeek).Parse(dowExpr)
	if err != nil {
		return nil, err
	}

	result.detectSpecialFlags()

	return result, nil
}

func (e *Expression) detectSpecialFlags() {
	for v := range e.DayOfMonth.Values {
		if v == 32 {
			e.HasLastDayOfMonth = true
		}
		if v == 33 {
			e.HasLastWeekday = true
		}
		if v >= 101 && v <= 131 {
			e.HasNearestWeekday = true
		}
	}

	for v := range e.DayOfWeek.Values {
		if v >= 10 && v <= 16 {
			e.HasLastDayOfWeek = true
		}
		if v >= 21 && v <= 75 {
			e.HasNthDayOfWeek = true
		}
	}
}

func (e *Expression) HasSpecialDayHandling() bool {
	return e.HasLastDayOfMonth || e.HasLastWeekday || e.HasNearestWeekday ||
		e.HasNthDayOfWeek || e.HasLastDayOfWeek
}

func (e *Expression) Matches(second, minute, hour, day, month, weekday int) bool {
	if !e.Second.Contains(second) {
		return false
	}
	if !e.Minute.Contains(minute) {
		return false
	}
	if !e.Hour.Contains(hour) {
		return false
	}
	if !e.Month.Contains(month) {
		return false
	}

	domMatch := e.DayOfMonth.Contains(day)
	dowMatch := e.DayOfWeek.Contains(weekday)

	if e.DayOfMonth.IsAll() && e.DayOfWeek.IsAll() {
		return true
	}
	if e.DayOfMonth.IsAll() {
		return dowMatch
	}
	if e.DayOfWeek.IsAll() {
		return domMatch
	}
	return domMatch || dowMatch
}

func (e *Expression) GetSeconds() []int { return e.Second.All() }
func (e *Expression) GetMinutes() []int { return e.Minute.All() }
func (e *Expression) GetHours() []int   { return e.Hour.All() }
func (e *Expression) GetMonths() []int  { return e.Month.All() }

func (e *Expression) GetDaysOfMonth() []int {
	result := make([]int, 0)
	for _, v := range e.DayOfMonth.All() {
		if v >= 1 && v <= 31 {
			result = append(result, v)
		}
	}
	return result
}

func (e *Expression) GetDaysOfWeek() []int {
	result := make([]int, 0)
	for _, v := range e.DayOfWeek.All() {
		if v >= 0 && v <= 6 {
			result = append(result, v)
		}
	}
	return result
}

func (e *Expression) String() string {
	if e.Type == ExtendedCron {
		return strings.Join([]string{
			e.Second.Raw, e.Minute.Raw, e.Hour.Raw,
			e.DayOfMonth.Raw, e.Month.Raw, e.DayOfWeek.Raw,
		}, " ")
	}
	return strings.Join([]string{
		e.Minute.Raw, e.Hour.Raw, e.DayOfMonth.Raw,
		e.Month.Raw, e.DayOfWeek.Raw,
	}, " ")
}

func (e *Expression) IsStandard() bool { return e.Type == StandardCron }
func (e *Expression) IsExtended() bool { return e.Type == ExtendedCron }

func (e *Expression) FieldStrings() []string {
	if e.Type == ExtendedCron {
		return []string{
			e.Second.Raw, e.Minute.Raw, e.Hour.Raw,
			e.DayOfMonth.Raw, e.Month.Raw, e.DayOfWeek.Raw,
		}
	}
	return []string{
		e.Minute.Raw, e.Hour.Raw, e.DayOfMonth.Raw,
		e.Month.Raw, e.DayOfWeek.Raw,
	}
}
