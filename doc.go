// Package expressparser is a comprehensive cron expression parser with timezone support
// and human-readable descriptions.
//
// # Features
//
//   - Parse standard 5-field cron expressions (minute, hour, day, month, weekday)
//   - Parse extended 6-field cron expressions with seconds
//   - Support for special characters: * , - / ? L W #
//   - Predefined schedules: @yearly, @monthly, @weekly, @daily, @hourly
//   - Timezone-aware scheduling
//   - Human-readable description generation
//   - Calculate next/previous run times
//
// # Quick Start
//
// Parse a cron expression and get the next run time:
//
//	expr, err := expressparser.Parse("0 9 * * 1-5")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	scheduler := expressparser.NewScheduler(expr)
//	next, _ := scheduler.Next(time.Now())
//	fmt.Println("Next run:", next)
//
// Or use the convenience functions:
//
//	next, err := expressparser.Next("0 9 * * *", time.Now())
//	description := expressparser.Describe(expr)
//
// # Cron Expression Format
//
// Standard format (5 fields):
//
//	┌───────────── minute (0-59)
//	│ ┌───────────── hour (0-23)
//	│ │ ┌───────────── day of month (1-31)
//	│ │ │ ┌───────────── month (1-12 or JAN-DEC)
//	│ │ │ │ ┌───────────── day of week (0-6 or SUN-SAT, 0=Sunday)
//	│ │ │ │ │
//	* * * * *
//
// Extended format (6 fields):
//
//	┌───────────── second (0-59)
//	│ ┌───────────── minute (0-59)
//	│ │ ┌───────────── hour (0-23)
//	│ │ │ ┌───────────── day of month (1-31)
//	│ │ │ │ ┌───────────── month (1-12 or JAN-DEC)
//	│ │ │ │ │ ┌───────────── day of week (0-6 or SUN-SAT)
//	│ │ │ │ │ │
//	* * * * * *
//
// # Special Characters
//
// The following special characters are supported:
//
//   - Any value
//     ,       Value list separator (e.g., "1,3,5")
//   - Range of values (e.g., "1-5")
//     /       Step values (e.g., "*/15" means every 15)
//     ?       Any value (same as *, typically used in day fields)
//     L       Last day of month or last occurrence of weekday
//     W       Nearest weekday to given day of month
//     #       Nth occurrence of weekday in month (e.g., "1#3" = third Monday)
//
// # Special Character Examples
//
// Last day of month:
//
//	"0 0 L * *"     // Midnight on last day of every month
//
// Last weekday of month:
//
//	"0 0 LW * *"    // Midnight on last weekday of every month
//
// Nearest weekday:
//
//	"0 0 15W * *"   // Midnight on weekday nearest to 15th
//
// Last specific weekday:
//
//	"0 0 * * 5L"    // Midnight on last Friday of every month
//
// Nth weekday of month:
//
//	"0 0 * * 1#2"   // Midnight on second Monday of every month
//
// # Predefined Schedules
//
// The following predefined schedules are supported:
//
//	@yearly     Run once a year at midnight on January 1st (0 0 1 1 *)
//	@annually   Same as @yearly
//	@monthly    Run once a month at midnight on the 1st (0 0 1 * *)
//	@weekly     Run once a week at midnight on Sunday (0 0 * * 0)
//	@daily      Run once a day at midnight (0 0 * * *)
//	@midnight   Same as @daily
//	@hourly     Run once an hour at the beginning (0 * * * *)
//
// # Timezone Support
//
// Create a scheduler with timezone support:
//
//	expr, _ := expressparser.Parse("0 9 * * *")
//
//	// Using timezone name
//	tzOpt, _ := expressparser.WithTimezone("America/New_York")
//	scheduler := expressparser.NewScheduler(expr, tzOpt)
//
//	// Or using time.Location
//	loc, _ := time.LoadLocation("Europe/London")
//	scheduler := expressparser.NewScheduler(expr, expressparser.WithLocation(loc))
//
//	// Next run will be in the specified timezone
//	next, _ := scheduler.Next(time.Now())
//
// # Human-Readable Descriptions
//
// Generate descriptions of cron expressions:
//
//	expr, _ := expressparser.Parse("0 9 * * 1-5")
//	desc := expressparser.Describe(expr)
//	// Output: "At 9:00 AM, on weekdays"
//
//	// With 24-hour time format
//	opts := expressparser.DescriptionOptions{Use24HourTime: true}
//	desc = expressparser.DescribeWithOptions(expr, opts)
//	// Output: "At 09:00, on weekdays"
//
// # Error Handling
//
// The package provides detailed error types for better error handling:
//
//	expr, err := expressparser.Parse("invalid")
//	if err != nil {
//	    // Check for specific error types
//	    if errors.Is(err, expressparser.ErrInvalidFieldCount) {
//	        fmt.Println("Wrong number of fields")
//	    }
//
//	    // Get detailed field error information
//	    var fieldErr *expressparser.FieldError
//	    if errors.As(err, &fieldErr) {
//	        fmt.Printf("Invalid %s field: %s (allowed: %d-%d)\n",
//	            fieldErr.Field, fieldErr.Value, fieldErr.Min, fieldErr.Max)
//	    }
//	}
//
// # Schedule Object
//
// For convenience, use the Schedule type which combines parsing, scheduling,
// and description:
//
//	schedule, err := expressparser.NewScheduleInTimezone("0 9 * * 1-5", "America/New_York")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	fmt.Println(schedule.Describe())              // Human-readable description
//	fmt.Println(schedule.Next(time.Now()))        // Next run time
//	fmt.Println(schedule.Timezone())              // Timezone location
//	fmt.Println(schedule.IsDue(time.Now()))       // Check if due now
//
// # Thread Safety
//
// All types in this package are safe for concurrent use. The Expression and
// Scheduler types are immutable after creation.
//
// # Performance
//
// The parser is designed for efficiency:
//   - Expressions are parsed once and can be reused
//   - Next/Previous calculations use optimized algorithms
//   - No regular expressions are used in parsing
//
// For best performance, parse expressions once and reuse the Scheduler:
//
//	// Do this (parse once)
//	scheduler, _ := expressparser.NewSchedulerFromConfig(config)
//	for {
//	    next, _ := scheduler.Next(time.Now())
//	    // use next
//	}
//
//	// Not this (parse every time)
//	for {
//	    next, _ := expressparser.Next("0 9 * * *", time.Now())
//	    // use next
//	}
package expressparser
