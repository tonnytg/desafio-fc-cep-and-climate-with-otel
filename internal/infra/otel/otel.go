package otel

import (
	"context"
	"log"

	"go.opentelemetry.io/otel/attribute"
)

type LogMessage struct {
	Message string `json:"message"`
}

func Logger(ctx context.Context, message string) {
	tr := GetTracer()
	_, span := tr.Start(ctx, "Logger")
	defer span.End()

	span.SetAttributes(attribute.String("log", message))
	log.Printf("Log: %s", message)
}
