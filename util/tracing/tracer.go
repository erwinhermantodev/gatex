package tracing

import (
	"context"

	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/database"
)

type contextKey string

const RequestIDKey contextKey = "request_id"

// Trace records a granular event for the current request
func Trace(ctx context.Context, level, component, message string) {
	requestID, ok := ctx.Value(RequestIDKey).(string)
	if !ok || requestID == "" {
		return
	}

	go func() {
		db := database.GetDB()
		db.Create(&database.TraceLog{
			RequestID: requestID,
			Level:     level,
			Component: component,
			Message:   message,
		})
	}()
}

func Info(ctx context.Context, component, message string) {
	Trace(ctx, "INFO", component, message)
}

func Warn(ctx context.Context, component, message string) {
	Trace(ctx, "WARN", component, message)
}

func Error(ctx context.Context, component, message string) {
	Trace(ctx, "ERROR", component, message)
}
