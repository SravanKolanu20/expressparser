package expressparser

import (
	"strconv"
	"strings"
)

// Field represents a parsed cron field with all valid values
type Field struct {
	Type   FieldType
	Values map[int]bool
	Raw    string
}

func NewField(fieldType FieldType) *Field {
	return &Field{
		Type:   fieldType,
		Values: make(map[int]bool),
	}
}

func (f *Field) Contains(value int) bool {
	return f.Values[value]
}

func (f *Field) Min() int {
	min := -1
	for v := range f.Values {
		if min == -1 || v < min {
			min = v
		}
	}
	return min
}

func (f *Field) Max() int {
	max := -1
	for v := range f.Values {
		if v > max {
			max = v
		}
	}
	return max
}

func (f *Field) All() []int {
	bounds := fieldBounds[f.Type]
	result := make([]int, 0, len(f.Values))
	for i := bounds.min; i <= bounds.max; i++ {
		if f.Values[i] {
			result = append(result, i)
		}
	}
	return result
}

func (f *Field) IsAll() bool {
	bounds := fieldBounds[f.Type]
	expected := bounds.max - bounds.min + 1
	return len(f.Values) == expected
}

// FieldParser handles parsing of cron field expressions
type FieldParser struct {
	fieldType FieldType
	min       int
	max       int
}

func NewFieldParser(fieldType FieldType) *FieldParser {
	bounds := fieldBounds[fieldType]
	return &FieldParser{
		fieldType: fieldType,
		min:       bounds.min,
		max:       bounds.max,
	}
}

func (p *FieldParser) Parse(expr string) (*Field, error) {
	field := NewField(p.fieldType)
	field.Raw = expr

	if expr == "" {
		return nil, NewFieldError(p.fieldType, expr, "field cannot be empty")
	}

	if expr == "*" || expr == "?" {
		p.addRange(field, p.min, p.max, 1)
		return field, nil
	}

	parts := strings.Split(expr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if err := p.parsePart(field, part); err != nil {
			return nil, err
		}
	}

	if len(field.Values) == 0 {
		return nil, NewFieldError(p.fieldType, expr, "no valid values found")
	}

	return field, nil
}

func (p *FieldParser) parsePart(field *Field, part string) error {
	// Check for step value first (e.g., "*/5" or "10-20/2")
	if strings.Contains(part, "/") {
		return p.parseStep(field, part)
	}

	// Check for range (e.g., "10-20" or "MON-FRI")
	// But first check if it's a special "L-N" pattern
	if strings.Contains(part, "-") {
		upperPart := strings.ToUpper(part)
		// Handle L-N pattern for day-of-month
		if strings.HasPrefix(upperPart, "L-") && p.fieldType == FieldDayOfMonth {
			return p.parseSpecial(field, part)
		}
		return p.parseRange(field, part)
	}

	// Check if it's a pure named value (like JAN, MON, etc.) BEFORE checking for special chars
	if p.isNamedValue(part) {
		return p.parseSingle(field, part)
	}

	// Check for special characters (L, W, #) - only for non-named values
	if p.hasSpecialChars(part) {
		return p.parseSpecial(field, part)
	}

	// Single value
	return p.parseSingle(field, part)
}

// isNamedValue checks if the part is a named month or day value
func (p *FieldParser) isNamedValue(part string) bool {
	upper := strings.ToUpper(strings.TrimSpace(part))

	if p.fieldType == FieldMonth {
		_, ok := monthNames[upper]
		return ok
	}
	if p.fieldType == FieldDayOfWeek {
		_, ok := dayNames[upper]
		return ok
	}
	return false
}

// hasSpecialChars checks if the part contains special cron characters
func (p *FieldParser) hasSpecialChars(part string) bool {
	upper := strings.ToUpper(part)

	// Only check for special chars in appropriate field types
	switch p.fieldType {
	case FieldDayOfMonth:
		// L, LW, NW patterns
		return upper == "L" || upper == "LW" || strings.HasSuffix(upper, "W") || strings.HasPrefix(upper, "L")
	case FieldDayOfWeek:
		// NL, N#M patterns
		return strings.HasSuffix(upper, "L") || strings.Contains(upper, "#")
	default:
		return false
	}
}

func (p *FieldParser) parseSingle(field *Field, part string) error {
	value, err := p.parseValue(part)
	if err != nil {
		return err
	}
	if err := p.validateValue(value, part); err != nil {
		return err
	}
	field.Values[value] = true
	return nil
}

func (p *FieldParser) parseRange(field *Field, part string) error {
	rangeParts := strings.SplitN(part, "-", 2)
	if len(rangeParts) != 2 {
		return NewFieldError(p.fieldType, part, "invalid range format")
	}

	start, err := p.parseValue(rangeParts[0])
	if err != nil {
		return err
	}
	end, err := p.parseValue(rangeParts[1])
	if err != nil {
		return err
	}

	if err := p.validateValue(start, rangeParts[0]); err != nil {
		return err
	}
	if err := p.validateValue(end, rangeParts[1]); err != nil {
		return err
	}

	if start > end {
		return &RangeError{Field: p.fieldType, Start: start, End: end}
	}

	p.addRange(field, start, end, 1)
	return nil
}

func (p *FieldParser) parseStep(field *Field, part string) error {
	stepParts := strings.SplitN(part, "/", 2)
	if len(stepParts) != 2 {
		return NewFieldError(p.fieldType, part, "invalid step format")
	}

	step, err := strconv.Atoi(stepParts[1])
	if err != nil {
		return NewFieldError(p.fieldType, stepParts[1], "step must be a number")
	}
	if step <= 0 {
		return &StepError{Field: p.fieldType, Step: step}
	}

	baseExpr := stepParts[0]

	if baseExpr == "*" {
		p.addRange(field, p.min, p.max, step)
		return nil
	}

	if strings.Contains(baseExpr, "-") {
		rangeParts := strings.SplitN(baseExpr, "-", 2)
		start, err := p.parseValue(rangeParts[0])
		if err != nil {
			return err
		}
		end, err := p.parseValue(rangeParts[1])
		if err != nil {
			return err
		}
		if err := p.validateValue(start, rangeParts[0]); err != nil {
			return err
		}
		if err := p.validateValue(end, rangeParts[1]); err != nil {
			return err
		}
		if start > end {
			return &RangeError{Field: p.fieldType, Start: start, End: end}
		}
		p.addRange(field, start, end, step)
		return nil
	}

	start, err := p.parseValue(baseExpr)
	if err != nil {
		return err
	}
	if err := p.validateValue(start, baseExpr); err != nil {
		return err
	}
	p.addRange(field, start, p.max, step)
	return nil
}

func (p *FieldParser) parseSpecial(field *Field, part string) error {
	upperPart := strings.ToUpper(part)

	switch p.fieldType {
	case FieldDayOfMonth:
		return p.parseDayOfMonthSpecial(field, upperPart)
	case FieldDayOfWeek:
		return p.parseDayOfWeekSpecial(field, upperPart)
	default:
		return NewFieldError(p.fieldType, part, "special characters L, W, # not allowed in this field")
	}
}

func (p *FieldParser) parseDayOfMonthSpecial(field *Field, part string) error {
	if part == "L" {
		field.Values[32] = true
		return nil
	}
	if part == "LW" {
		field.Values[33] = true
		return nil
	}
	if strings.HasPrefix(part, "L-") {
		offset, err := strconv.Atoi(part[2:])
		if err != nil || offset < 0 || offset > 30 {
			return NewFieldError(p.fieldType, part, "invalid L-N format")
		}
		field.Values[32+offset] = true
		return nil
	}
	if strings.HasSuffix(part, "W") {
		dayStr := part[:len(part)-1]
		day, err := strconv.Atoi(dayStr)
		if err != nil {
			return NewFieldError(p.fieldType, part, "invalid NW format")
		}
		if day < 1 || day > 31 {
			return NewFieldError(p.fieldType, part, "day must be between 1 and 31")
		}
		field.Values[100+day] = true
		return nil
	}
	return NewFieldError(p.fieldType, part, "unrecognized special character combination")
}

func (p *FieldParser) parseDayOfWeekSpecial(field *Field, part string) error {
	if strings.HasSuffix(part, "L") {
		dayStr := part[:len(part)-1]
		day, err := p.parseValue(dayStr)
		if err != nil {
			return err
		}
		if day < 0 || day > 6 {
			return NewFieldError(p.fieldType, part, "day must be between 0 and 6")
		}
		field.Values[10+day] = true
		return nil
	}
	if strings.Contains(part, "#") {
		parts := strings.SplitN(part, "#", 2)
		if len(parts) != 2 {
			return NewFieldError(p.fieldType, part, "invalid N#M format")
		}
		day, err := p.parseValue(parts[0])
		if err != nil {
			return err
		}
		if day < 0 || day > 6 {
			return NewFieldError(p.fieldType, part, "day must be between 0 and 6")
		}
		occurrence, err := strconv.Atoi(parts[1])
		if err != nil || occurrence < 1 || occurrence > 5 {
			return NewFieldError(p.fieldType, part, "occurrence must be between 1 and 5")
		}
		field.Values[20+day*10+occurrence] = true
		return nil
	}
	return NewFieldError(p.fieldType, part, "unrecognized special character combination")
}

func (p *FieldParser) parseValue(s string) (int, error) {
	s = strings.TrimSpace(s)
	s = strings.ToUpper(s)

	if p.fieldType == FieldMonth {
		if v, ok := monthNames[s]; ok {
			return v, nil
		}
	}
	if p.fieldType == FieldDayOfWeek {
		if v, ok := dayNames[s]; ok {
			return v, nil
		}
	}

	value, err := strconv.Atoi(s)
	if err != nil {
		return 0, NewFieldError(p.fieldType, s, "invalid value")
	}
	return value, nil
}

func (p *FieldParser) validateValue(value int, original string) error {
	if value < p.min || value > p.max {
		return NewFieldError(p.fieldType, original, "value out of range")
	}
	return nil
}

func (p *FieldParser) addRange(field *Field, start, end, step int) {
	for i := start; i <= end; i += step {
		field.Values[i] = true
	}
}

var monthNames = map[string]int{
	"JAN": 1, "JANUARY": 1,
	"FEB": 2, "FEBRUARY": 2,
	"MAR": 3, "MARCH": 3,
	"APR": 4, "APRIL": 4,
	"MAY": 5,
	"JUN": 6, "JUNE": 6,
	"JUL": 7, "JULY": 7,
	"AUG": 8, "AUGUST": 8,
	"SEP": 9, "SEPTEMBER": 9,
	"OCT": 10, "OCTOBER": 10,
	"NOV": 11, "NOVEMBER": 11,
	"DEC": 12, "DECEMBER": 12,
}

var dayNames = map[string]int{
	"SUN": 0, "SUNDAY": 0,
	"MON": 1, "MONDAY": 1,
	"TUE": 2, "TUESDAY": 2,
	"WED": 3, "WEDNESDAY": 3,
	"THU": 4, "THURSDAY": 4,
	"FRI": 5, "FRIDAY": 5,
	"SAT": 6, "SATURDAY": 6,
}
