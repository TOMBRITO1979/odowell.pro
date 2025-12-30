package models

import (
	"database/sql/driver"
	"fmt"
	"os"
	"time"
)

// saoPauloLocation returns the São Paulo timezone location
// This is the default timezone for all time operations in the system
func saoPauloLocation() *time.Location {
	tz := os.Getenv("TZ")
	if tz == "" {
		tz = "America/Sao_Paulo"
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		// Fallback to fixed offset for São Paulo (-03:00)
		loc = time.FixedZone("BRT", -3*60*60)
	}
	return loc
}

// LocalTime is a custom time type that always uses São Paulo timezone
// This ensures consistent time handling regardless of server location
type LocalTime struct {
	time.Time
}

// MarshalJSON serializes the time in São Paulo timezone without offset
// Output format: "2006-01-02T15:04:05" (no Z or offset)
// This format is treated as local time by JavaScript
func (lt LocalTime) MarshalJSON() ([]byte, error) {
	if lt.IsZero() {
		return []byte("null"), nil
	}
	// Convert to São Paulo timezone before formatting
	spTime := lt.Time.In(saoPauloLocation())
	return []byte(fmt.Sprintf("\"%s\"", spTime.Format("2006-01-02T15:04:05"))), nil
}

// UnmarshalJSON parses time from JSON and interprets it in São Paulo timezone
func (lt *LocalTime) UnmarshalJSON(data []byte) error {
	// Remove quotes
	s := string(data)
	if s == "null" || s == "\"\"" {
		lt.Time = time.Time{}
		return nil
	}
	s = s[1 : len(s)-1]

	spLoc := saoPauloLocation()

	// Try parsing with timezone information first (RFC3339/ISO8601)
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		// Convert from UTC to São Paulo timezone for storage
		lt.Time = t.In(spLoc)
		return nil
	}

	// Try parsing with Z suffix (UTC)
	if t, err := time.Parse("2006-01-02T15:04:05Z", s); err == nil {
		// Time is in UTC, convert to São Paulo
		lt.Time = t.In(spLoc)
		return nil
	}

	// Parse without timezone - interpret as São Paulo local time
	formats := []string{
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
	}

	for _, format := range formats {
		if t, err := time.ParseInLocation(format, s, spLoc); err == nil {
			lt.Time = t
			return nil
		}
	}

	return fmt.Errorf("cannot parse time: %s", s)
}

// Value implements driver.Valuer for database storage
// Stores time in São Paulo timezone
func (lt LocalTime) Value() (driver.Value, error) {
	if lt.IsZero() {
		return nil, nil
	}
	// Ensure time is in São Paulo timezone before storing
	return lt.Time.In(saoPauloLocation()), nil
}

// Scan implements sql.Scanner for database retrieval
// Ensures retrieved time is in São Paulo timezone
func (lt *LocalTime) Scan(value interface{}) error {
	if value == nil {
		lt.Time = time.Time{}
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		// Ensure time is in São Paulo timezone
		lt.Time = v.In(saoPauloLocation())
		return nil
	default:
		return fmt.Errorf("cannot scan type %T into LocalTime", value)
	}
}
