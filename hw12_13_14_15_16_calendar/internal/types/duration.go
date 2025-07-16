package types

import (
	"database/sql/driver"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Regular expressions for interval parsing, compiled once at package initialization.
var (
	reYear         = regexp.MustCompile(`(\d+)\s*years?`)
	reMonth        = regexp.MustCompile(`(\d+)\s*mon(th)?s?`)
	reDay          = regexp.MustCompile(`(\d+)\s*days?`)
	reHour         = regexp.MustCompile(`(\d+)\s*hours?`)
	reMinute       = regexp.MustCompile(`(\d+)\s*minutes?`)
	reSecond       = regexp.MustCompile(`(\d+\.?\d*)\s*seconds?`)
	reMillisecond  = regexp.MustCompile(`(\d+)\s*milliseconds?`)
	reMicrosecond  = regexp.MustCompile(`(\d+)\s*microseconds?`)
	reYearExtract  = regexp.MustCompile(`(\d+)y`)
	reMonthExtract = regexp.MustCompile(`(\d+)mo`)
	reDayExtract   = regexp.MustCompile(`(\d+)d`)
)

// Duration represents a time duration with its string representation for SQL compatibility.
type Duration struct {
	value time.Duration // The duration value.
	str   string        // String representation in "%v seconds" format.
}

// NewDuration creates a new Duration from a time.Duration.
func NewDuration(d time.Duration) Duration {
	return Duration{
		value: d,
		str:   fmt.Sprintf("%d seconds", int64(d.Seconds())),
	}
}

// String returns the string representation of the Duration in "%v seconds" format.
func (d Duration) String() string {
	return d.str
}

// ToDuration returns the time.Duration value of the Duration.
func (d Duration) ToDuration() time.Duration {
	return d.value
}

// Scan implements the sql.Scanner interface for converting SQL interval values to Duration.
func (d *Duration) Scan(value any) error {
	var v string
	switch val := value.(type) {
	case string:
		v = val
	case []uint8:
		// Handle byte slice (e.g., from PostgreSQL with sqlx).
		v = string(val)
	case int64:
		// Numeric format (seconds, e.g., in SQLite).
		*d = Duration{
			value: time.Duration(val) * time.Second,
			str:   fmt.Sprintf("%d seconds", val),
		}
		return nil
	case nil:
		*d = Duration{
			value: 0,
			str:   "0 seconds",
		}
		return nil
	default:
		return fmt.Errorf("unsupported scan type for Duration: %T", value)
	}

	v = strings.ToLower(v)
	v = strings.ReplaceAll(v, " ", "")
	replacements := []struct {
		re          *regexp.Regexp
		replacement string
	}{
		{reYear, "${1}y"},         // Years: "1 year" or "1 years" -> "1y".
		{reMonth, "${1}mo"},       // Months: "1 mon" or "1 months" -> "1mo".
		{reDay, "${1}d"},          // Days: "1 day" or "1 days" -> "1d".
		{reHour, "${1}h"},         // Hours: "1 hour" or "1 hours" -> "1h".
		{reMinute, "${1}m"},       // Minutes: "1 minute" or "1 minutes" -> "1m".
		{reSecond, "${1}s"},       // Seconds: "1 second" or "1.234 seconds" -> "1s" or "1.234s".
		{reMillisecond, "${1}ms"}, // Milliseconds: "1 millisecond" -> "1ms".
		{reMicrosecond, "${1}us"}, // Microseconds: "1 microsecond" -> "1us".
	}
	for _, r := range replacements {
		v = r.re.ReplaceAllString(v, r.replacement)
	}
	// Handle "HH:MM:SS" or "HH:MM:SS.ffffff" format.
	//nolint:nestif,nolintlint
	if strings.Contains(v, ":") {
		parts := strings.Split(v, ":")
		if len(parts) == 3 {
			secParts := strings.Split(parts[2], ".") // Handle microseconds.
			seconds := secParts[0]
			micros := "0"
			if len(secParts) > 1 {
				micros = secParts[1]
				if len(micros) > 6 {
					micros = micros[:6] // Limit to microseconds.
				}
				micros = strings.TrimRight(micros, "0") // Remove trailing zeros.
				if micros != "" {
					seconds += "." + micros
				}
			}
			v = fmt.Sprintf("%sh%sm%ss", parts[0], parts[1], seconds)
		}
	}
	// Convert years and months to approximate durations (1 year = 365 days, 1 month = 30 days).
	var totalDuration time.Duration
	if matches := reYearExtract.FindStringSubmatch(v); len(matches) > 1 {
		years, _ := strconv.ParseInt(matches[1], 10, 64)
		totalDuration += time.Duration(years) * 365 * 24 * time.Hour
		v = reYearExtract.ReplaceAllString(v, "")
	}
	if matches := reMonthExtract.FindStringSubmatch(v); len(matches) > 1 {
		months, _ := strconv.ParseInt(matches[1], 10, 64)
		totalDuration += time.Duration(months) * 30 * 24 * time.Hour
		v = reMonthExtract.ReplaceAllString(v, "")
	}
	if matches := reDayExtract.FindStringSubmatch(v); len(matches) > 1 {
		days, _ := strconv.ParseInt(matches[1], 10, 64)
		totalDuration += time.Duration(days) * 24 * time.Hour
		v = reDayExtract.ReplaceAllString(v, "")
	}
	// Parse remaining duration (hours, minutes, seconds, milliseconds, microseconds).
	if v != "" {
		parsed, err := time.ParseDuration(v)
		if err != nil {
			return fmt.Errorf("failed to parse duration: %w", err)
		}
		totalDuration += parsed
	}
	*d = Duration{
		value: totalDuration,
		str:   fmt.Sprintf("%d seconds", int64(totalDuration.Seconds())),
	}
	return nil
}

// Value implements the driver.Valuer interface for converting Duration to SQL-compatible format.
func (d Duration) Value() (driver.Value, error) {
	return d.str, nil
}
