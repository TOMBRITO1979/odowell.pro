package models

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// LocalTime is a custom time type that serializes without timezone offset
// This prevents JavaScript from converting to browser's local timezone
type LocalTime struct {
	time.Time
}

// MarshalJSON serializes the time without timezone offset
// Output format: "2006-01-02T15:04:05" (no Z or offset)
func (lt LocalTime) MarshalJSON() ([]byte, error) {
	if lt.IsZero() {
		return []byte("null"), nil
	}
	// Format without timezone offset - JavaScript will treat as local time
	return []byte(fmt.Sprintf("\"%s\"", lt.Format("2006-01-02T15:04:05"))), nil
}

// UnmarshalJSON parses time from JSON
func (lt *LocalTime) UnmarshalJSON(data []byte) error {
	// Remove quotes
	s := string(data)
	if s == "null" || s == "\"\"" {
		lt.Time = time.Time{}
		return nil
	}
	s = s[1 : len(s)-1]

	// Try parsing with various formats
	formats := []string{
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		time.RFC3339,
	}

	var err error
	for _, format := range formats {
		lt.Time, err = time.Parse(format, s)
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("cannot parse time: %s", s)
}

// Value implements driver.Valuer for database storage
func (lt LocalTime) Value() (driver.Value, error) {
	if lt.IsZero() {
		return nil, nil
	}
	return lt.Time, nil
}

// Scan implements sql.Scanner for database retrieval
func (lt *LocalTime) Scan(value interface{}) error {
	if value == nil {
		lt.Time = time.Time{}
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		lt.Time = v
		return nil
	default:
		return fmt.Errorf("cannot scan type %T into LocalTime", value)
	}
}
