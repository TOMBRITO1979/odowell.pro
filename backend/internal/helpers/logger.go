package helpers

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// LogLevel represents the severity of a log entry
type LogLevel string

const (
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
	LogLevelFatal LogLevel = "FATAL"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp  string      `json:"timestamp"`
	Level      LogLevel    `json:"level"`
	Message    string      `json:"message"`
	RequestID  string      `json:"request_id,omitempty"`
	UserID     uint        `json:"user_id,omitempty"`
	TenantID   uint        `json:"tenant_id,omitempty"`
	Method     string      `json:"method,omitempty"`
	Path       string      `json:"path,omitempty"`
	StatusCode int         `json:"status_code,omitempty"`
	DurationMs float64     `json:"duration_ms,omitempty"`
	IP         string      `json:"ip,omitempty"`
	UserAgent  string      `json:"user_agent,omitempty"`
	Error      string      `json:"error,omitempty"`
	Stack      string      `json:"stack,omitempty"`
	Extra      interface{} `json:"extra,omitempty"`
}

// logJSON outputs a log entry as JSON to stdout
func logJSON(entry LogEntry) {
	entry.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	json.NewEncoder(os.Stdout).Encode(entry)
}

// LogDebug logs a debug message
func LogDebug(message string, extra ...interface{}) {
	entry := LogEntry{
		Level:   LogLevelDebug,
		Message: message,
	}
	if len(extra) > 0 {
		entry.Extra = extra[0]
	}
	logJSON(entry)
}

// LogInfo logs an info message
func LogInfo(message string, extra ...interface{}) {
	entry := LogEntry{
		Level:   LogLevelInfo,
		Message: message,
	}
	if len(extra) > 0 {
		entry.Extra = extra[0]
	}
	logJSON(entry)
}

// LogWarn logs a warning message
func LogWarn(message string, extra ...interface{}) {
	entry := LogEntry{
		Level:   LogLevelWarn,
		Message: message,
	}
	if len(extra) > 0 {
		entry.Extra = extra[0]
	}
	logJSON(entry)
}

// LogError logs an error message
func LogError(message string, err error, extra ...interface{}) {
	entry := LogEntry{
		Level:   LogLevelError,
		Message: message,
	}
	if err != nil {
		entry.Error = err.Error()
	}
	if len(extra) > 0 {
		entry.Extra = extra[0]
	}
	logJSON(entry)
}

// LogFatal logs a fatal message (does not exit - caller should handle)
func LogFatal(message string, err error, extra ...interface{}) {
	entry := LogEntry{
		Level:   LogLevelFatal,
		Message: message,
	}
	if err != nil {
		entry.Error = err.Error()
	}
	if len(extra) > 0 {
		entry.Extra = extra[0]
	}
	logJSON(entry)
}

// LogRequest logs an HTTP request with all relevant details
func LogRequest(entry LogEntry) {
	if entry.Level == "" {
		entry.Level = LogLevelInfo
	}
	logJSON(entry)
}

// LogSecurityEvent logs security-related events (login, logout, permission denied, etc.)
func LogSecurityEvent(event string, userID uint, tenantID uint, success bool, details map[string]interface{}) {
	level := LogLevelInfo
	if !success {
		level = LogLevelWarn
	}

	entry := LogEntry{
		Level:    level,
		Message:  fmt.Sprintf("Security event: %s", event),
		UserID:   userID,
		TenantID: tenantID,
		Extra: map[string]interface{}{
			"event":   event,
			"success": success,
			"details": details,
		},
	}
	logJSON(entry)
}

// LogDatabaseQuery logs slow database queries
func LogDatabaseQuery(query string, durationMs float64, err error) {
	level := LogLevelInfo
	if durationMs > 200 {
		level = LogLevelWarn
	}
	if err != nil {
		level = LogLevelError
	}

	entry := LogEntry{
		Level:      level,
		Message:    "Database query",
		DurationMs: durationMs,
		Extra: map[string]interface{}{
			"query": query,
		},
	}
	if err != nil {
		entry.Error = err.Error()
	}
	logJSON(entry)
}
