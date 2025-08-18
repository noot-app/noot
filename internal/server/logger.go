package server

import (
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strconv"
	"strings"
)

var logger *slog.Logger

// InitLogger must be called after environment variables are loaded
func InitLogger() {
	setupLogger()
}

func setupLogger() {
	logLevel := getLogLevel()

	opts := &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: logLevel == slog.LevelDebug,
	}

	var handler slog.Handler
	if isDebugMode() {
		// Pretty text output for development
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		// JSON output for production
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	logger = slog.New(handler)
	slog.SetDefault(logger)
}

func getLogLevel() slog.Level {
	levelStr := strings.ToUpper(getenv("LOG_LEVEL", "INFO"))
	switch levelStr {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func isDebugMode() bool {
	debug := getenv("DEBUG", "false")
	val, err := strconv.ParseBool(debug)
	return err == nil && val
}

func isDevMode() bool {
	env := strings.ToLower(getenv("ENV", getenv("ENVIRONMENT", "production")))
	return env == "dev" || env == "development" || env == "local"
}

// Enhanced error with stack trace capability
type AppError struct {
	Message    string
	Err        error
	StatusCode int
	Stack      []string
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewAppError(message string, statusCode int, err error) *AppError {
	appErr := &AppError{
		Message:    message,
		Err:        err,
		StatusCode: statusCode,
	}

	// Capture stack trace if debug mode is enabled
	if isDebugMode() || isDevMode() {
		appErr.Stack = captureStack(3) // Skip 3 frames: captureStack, NewAppError, caller
	}

	return appErr
}

func captureStack(skip int) []string {
	var stack []string
	pc := make([]uintptr, 10)
	n := runtime.Callers(skip, pc)
	frames := runtime.CallersFrames(pc[:n])

	for {
		frame, more := frames.Next()
		if !strings.Contains(frame.File, "runtime/") {
			stack = append(stack, fmt.Sprintf("%s:%d %s", frame.File, frame.Line, frame.Function))
		}
		if !more {
			break
		}
	}

	return stack
}

// Log helpers
func LogInfo(msg string, args ...any) {
	logger.Info(msg, args...)
}

func LogError(msg string, err error, args ...any) {
	allArgs := append([]any{"error", err}, args...)
	logger.Error(msg, allArgs...)
}

func LogWarn(msg string, args ...any) {
	logger.Warn(msg, args...)
}

func LogDebug(msg string, args ...any) {
	logger.Debug(msg, args...)
}
