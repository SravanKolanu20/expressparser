# expressparser

`expressparser` is a cron expression parser and scheduler written in Go.  
It supports standard 5‑field cron expressions, extended 6‑field expressions with seconds, predefined shortcuts, timezone‑aware scheduling, and human‑readable descriptions.

---

## Features

- Parse standard 5‑field cron expressions:
  - `minute hour day-of-month month day-of-week`
- Parse extended 6‑field expressions with seconds:
  - `second minute hour day-of-month month day-of-week`
- Predefined expressions such as `@yearly`, `@monthly`, `@weekly`, `@daily`, `@hourly`
- Special characters:
  - `*` (any), `,` (list), `-` (range), `/` (step), `?`, `L`, `W`, `#`
- Timezone‑aware scheduling via `Scheduler`
- Human‑readable descriptions via `Descriptor`
- Compute next and previous run times

---

## Installation

```bash
go get github.com/SravanKolanu20/expressparser
```

---

## Quick Start

```go
package main

import (
    "fmt"
    "time"

    "github.com/SravanKolanu20/expressparser"
)

func main() {
    // Parse expression
    expr, err := expressparser.Parse("0 9 * * 1-5") // 9 AM on weekdays
    if err != nil {
        panic(err)
    }

    // Create scheduler (UTC by default)
    s := expressparser.NewScheduler(expr)

    next, err := s.Next(time.Now())
    if err != nil {
        panic(err)
    }
    fmt.Println("Next run:", next)

    // Human‑readable description
    desc := expressparser.Describe(expr)
    fmt.Println("Description:", desc)
}
```

---

## Timezone Example

```go
expr, _ := expressparser.Parse("0 9 * * 1-5")

tzOpt, err := expressparser.WithTimezone("America/New_York")
if err != nil {
    panic(err)
}

scheduler := expressparser.NewScheduler(expr, tzOpt)
next, err := scheduler.Next(time.Now())
if err != nil {
    panic(err)
}

fmt.Println("Next run in New York:", next)
```

---

## Convenience APIs

You can skip manual parsing and use helpers in `parser.go`:

```go
next, err := expressparser.Next("0 9 * * *", time.Now())
times, err := expressparser.NextN("0 9 * * 1-5", time.Now(), 5)

due, err := expressparser.IsDue("0 9 * * *", time.Now())
desc, _ := func() (string, error) {
    e, err := expressparser.Parse("0 9 * * 1-5")
    if err != nil {
        return "", err
    }
    return expressparser.Describe(e), nil
}()
```

---

## Error Handling

The package exposes structured error types for better inspection:

- Sentinel errors: `ErrEmptyExpression`, `ErrInvalidFieldCount`, `ErrInvalidTimezone`, etc.
- Detailed types: `ParseError`, `FieldError`, `RangeError`, `StepError`

Example:

```go
expr, err := expressparser.Parse("invalid expression")
if err != nil {
    if errors.Is(err, expressparser.ErrInvalidFieldCount) {
        // handle wrong field count
    }

    var fe *expressparser.FieldError
    if errors.As(err, &fe) {
        // fe.Field, fe.Value, fe.Min, fe.Max give detailed info
    }
}
```
---
## License

This project is licensed under the **MIT License**.

```text
MIT License

Copyright (c) 2026 Sravan Kumar Kolanu

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the “Software”), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
```