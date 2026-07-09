package logger

import (
	"context"
	"time"

	"github.com/addpur/go-polykit"
	"go.uber.org/zap"
)

// Logger wraps zap.SugaredLogger to provide context-aware logging.
type Logger struct {
	*zap.SugaredLogger
}

// NewLogger creates a new Logger instance wrapping the provided zap.SugaredLogger.
func NewLogger(sugar *zap.SugaredLogger) *Logger {
	return &Logger{SugaredLogger: sugar}
}

// Info logs an informational message.
func (l *Logger) Info(ctx context.Context, args ...interface{}) {
	l.SugaredLogger.Info(args...)
}

// Infof logs a formatted informational message.
func (l *Logger) Infof(ctx context.Context, template string, args ...interface{}) {
	l.SugaredLogger.Infof(template, args...)
}

// Error logs an error message.
func (l *Logger) Error(ctx context.Context, args ...interface{}) {
	l.SugaredLogger.Error(args...)
}

// Errorf logs a formatted error message.
func (l *Logger) Errorf(ctx context.Context, template string, args ...interface{}) {
	l.SugaredLogger.Errorf(template, args...)
}

// Debug logs a debug message.
func (l *Logger) Debug(ctx context.Context, args ...interface{}) {
	l.SugaredLogger.Debug(args...)
}

// Debugf logs a formatted debug message.
func (l *Logger) Debugf(ctx context.Context, template string, args ...interface{}) {
	l.SugaredLogger.Debugf(template, args...)
}

// LoggingMiddleware creates a polykit middleware that logs request details using the provided Zap Logger.
func LoggingMiddleware(logger *Logger, endpointName string) polykit.Middleware {
	return func(next polykit.Endpoint) polykit.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			logger.Infof(ctx, "entering endpoint %s", endpointName)
			logger.Debugf(ctx, "request: %+v", request)

			defer func(begin time.Time) {
				logger.Infof(ctx, "exiting endpoint %s, duration: %v", endpointName, time.Since(begin))
				if err != nil {
					logger.Errorf(ctx, "endpoint %s error: %v", endpointName, err)
				}
			}(time.Now())

			response, err = next(ctx, request)
			if err == nil {
				logger.Debugf(ctx, "response: %+v", response)
			}

			return response, err
		}
	}
}
