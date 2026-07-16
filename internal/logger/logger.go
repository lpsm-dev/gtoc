package logger

import (
	"io"
	"os"
	"time"

	"charm.land/log/v2"
)

var (
	// Global logger instance.
	defaultLogger *log.Logger

	// Default configuration values.
	defaultLevel      = log.WarnLevel
	defaultTimeFormat = time.Kitchen
	defaultPrefix     = "gtoc"
)

// Options configures the logger.
type Options struct {
	Level        log.Level
	TimeFormat   string
	Output       io.Writer
	ShowCaller   bool
	CallerOffset int
	Prefix       string
}

// Init initializes the global logger with the given options.
func Init(opts *Options) {
	if opts == nil {
		opts = &Options{}
	}

	// Apply default values for unset fields.
	if opts.Level == 0 {
		opts.Level = defaultLevel
	}
	if opts.TimeFormat == "" {
		opts.TimeFormat = defaultTimeFormat
	}
	if opts.Output == nil {
		opts.Output = os.Stderr
	}
	if opts.Prefix == "" {
		opts.Prefix = defaultPrefix
	}

	// Create and configure the logger.
	defaultLogger = log.NewWithOptions(opts.Output, log.Options{
		Level:        opts.Level,
		TimeFormat:   opts.TimeFormat,
		ReportCaller: opts.ShowCaller,
		CallerOffset: opts.CallerOffset,
		Prefix:       opts.Prefix,
	})
}

// GetLogger returns the global logger instance.
func GetLogger() *log.Logger {
	if defaultLogger == nil {
		// Initialize with default values if not yet initialized.
		Init(nil)
	}
	return defaultLogger
}

// SetLevel changes the log level.
func SetLevel(level log.Level) {
	GetLogger().SetLevel(level)
}

// SetOutput changes the log output destination.
func SetOutput(output io.Writer) {
	GetLogger().SetOutput(output)
}

// SetTimeFormat changes the time format used in logs.
func SetTimeFormat(format string) {
	GetLogger().SetTimeFormat(format)
}

// SetPrefix changes the log prefix.
func SetPrefix(prefix string) {
	GetLogger().SetPrefix(prefix)
}

// Helper functions for convenient logger usage.

// Debug logs a debug message.
func Debug(msg any, keyvals ...any) {
	GetLogger().Debug(msg, keyvals...)
}

// Info logs an informational message.
func Info(msg any, keyvals ...any) {
	GetLogger().Info(msg, keyvals...)
}

// Warn logs a warning message.
func Warn(msg any, keyvals ...any) {
	GetLogger().Warn(msg, keyvals...)
}

// Error logs an error message.
func Error(msg any, keyvals ...any) {
	GetLogger().Error(msg, keyvals...)
}

// Fatal logs a fatal message and terminates the application.
func Fatal(msg any, keyvals ...any) {
	GetLogger().Fatal(msg, keyvals...)
}

// With returns a new logger with the given key-value pairs added.
func With(keyvals ...any) *log.Logger {
	return GetLogger().With(keyvals...)
}
