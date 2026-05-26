package main

import (
	"context"
	"log"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("ultrasearch-engine")

// InitOtel safely initializes global OpenTelemetry.
// Returns a shutdown function that flushes metrics gracefully.
func InitOtel(serviceName string) (func(context.Context) error, error) {
	noopShutdown := func(context.Context) error { return nil }

	f, err := os.OpenFile("otel_traces.jsonl", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("⚠️ OTel tracing disabled: could not open log file: %v", err)
		return noopShutdown, err
	}

	exporter, err := stdouttrace.New(
		stdouttrace.WithWriter(f),
	)
	if err != nil {
		log.Printf("⚠️ OTel tracing disabled: exporter creation failed: %v", err)
		return noopShutdown, err
	}

	res, err := resource.New(context.Background(), resource.WithAttributes(
		semconv.ServiceName(serviceName),
	))
	if err != nil {
		log.Printf("⚠️ OTel tracing disabled: resource creation failed: %v", err)
		return noopShutdown, err
	}

	// Memory Management: Use BatchSpanProcessor to prevent main thread blocking and manage memory
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter, sdktrace.WithMaxExportBatchSize(512)),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	
	// Graceful shutdown wrapper
	shutdown := func(ctx context.Context) error {
		return tp.Shutdown(ctx)
	}
	return shutdown, nil
}

// safeContext ensures we never pass a strict nil context, which is the only thing that crashes OTel.
func safeContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

func TraceQueryStart(ctx context.Context, workerID int, query string, aiMode string) (context.Context, trace.Span) {
	ctx, span := tracer.Start(safeContext(ctx), "Worker.ProcessQuery")
	span.SetAttributes(
		attribute.Int("worker.id", workerID),
		attribute.String("query.text", query),
		attribute.String("query.aimode", aiMode),
	)
	return ctx, span
}

func TraceAttemptStart(ctx context.Context, workerID int, attempt int) (context.Context, trace.Span) {
	ctx, span := tracer.Start(safeContext(ctx), "Worker.Attempt")
	span.SetAttributes(
		attribute.Int("worker.id", workerID),
		attribute.Int("attempt.number", attempt),
	)
	return ctx, span
}

func TraceBrowserSpawn(ctx context.Context, workerID int, tempDir string) {
	_, span := tracer.Start(safeContext(ctx), "Worker.BrowserSpawn")
	span.SetAttributes(
		attribute.Int("worker.id", workerID),
		attribute.String("browser.profile_dir", tempDir),
	)
	span.End() // Immediate end, acts as an event
}

func TraceBrowserKill(ctx context.Context, workerID int) {
	_, span := tracer.Start(safeContext(ctx), "Worker.BrowserKill")
	span.SetAttributes(attribute.Int("worker.id", workerID))
	span.End()
}

func TraceQuerySuccess(ctx context.Context, workerID int, resultsCount int, duration time.Duration) {
	_, span := tracer.Start(safeContext(ctx), "Worker.QuerySuccess")
	span.SetAttributes(
		attribute.Int("worker.id", workerID),
		attribute.Int("results.count", resultsCount),
		attribute.String("duration", duration.String()),
	)
	span.End()
}

func TraceQueryFailed(ctx context.Context, workerID int, errMsg string, duration time.Duration) {
	_, span := tracer.Start(safeContext(ctx), "Worker.QueryFailed")
	span.SetAttributes(
		attribute.Int("worker.id", workerID),
		attribute.String("error.message", errMsg),
		attribute.String("duration", duration.String()),
	)
	span.End()
}

func TraceCaptchaDetected(ctx context.Context, workerID int, attempt int, screenshotPath string) {
	_, span := tracer.Start(safeContext(ctx), "Worker.CaptchaDetected")
	span.SetAttributes(
		attribute.Int("worker.id", workerID),
		attribute.Int("attempt.number", attempt),
		attribute.String("screenshot.path", screenshotPath),
	)
	span.End()
}

func TraceNavigate(ctx context.Context, url string, d time.Duration) {
	_, span := tracer.Start(safeContext(ctx), "Browser.Navigate")
	span.SetAttributes(
		attribute.String("url", url),
		attribute.String("duration", d.String()),
	)
	span.End()
}

func TracePollDuration(ctx context.Context, d time.Duration) {
	_, span := tracer.Start(safeContext(ctx), "Browser.DOMPoll")
	span.SetAttributes(attribute.String("duration", d.String()))
	span.End()
}

func TraceSessionInjected(ctx context.Context, sessionID string, cookieCount int) {
	_, span := tracer.Start(safeContext(ctx), "Session.Injected")
	span.SetAttributes(
		attribute.String("session.id", sessionID),
		attribute.Int("session.cookie_count", cookieCount),
	)
	span.End()
}

func TraceSessionPoolEmpty(ctx context.Context) {
	_, span := tracer.Start(safeContext(ctx), "Session.PoolEmpty")
	span.End()
}
