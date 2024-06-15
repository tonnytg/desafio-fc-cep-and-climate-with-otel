package otel

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer

func InitTracer(serviceName string, attrs ...attribute.KeyValue) func() {
	ctx := context.Background()

	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint("otel-collector:4317"), otlptracegrpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to create exporter: %v", err)
	}

	// Adiciona o nome do serviço aos atributos
	defaultAttrs := []attribute.KeyValue{
		semconv.ServiceNameKey.String(serviceName),
	}

	// Combina os atributos padrão com os atributos adicionais fornecidos
	allAttrs := append(defaultAttrs, attrs...)

	resources := resource.NewWithAttributes(
		semconv.SchemaURL,
		allAttrs...,
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resources),
	)

	otel.SetTracerProvider(tp)
	tracer = tp.Tracer(serviceName)

	return func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Fatalf("failed to shutdown TracerProvider: %v", err)
		}
	}
}

func GetTracer() trace.Tracer {
	return tracer
}
