package otel

import (
	"context"
	"log"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracer     trace.Tracer
	tracerOnce sync.Once
)

func InitTracerTest() func() {

	var ctx context.Context
	var shutdownFunc func()

	tpTest := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tpTest)

	otel.SetTracerProvider(tpTest)
	tracer = tpTest.Tracer("test-service-execute")

	shutdownFunc = func() {
		if err := tpTest.Shutdown(ctx); err != nil {
			log.Fatalf("failed to shutdown TracerTestProvider: %v", err)
		}
	}

	return shutdownFunc
}

func InitTracer(serviceName string, attrs ...attribute.KeyValue) func() {
	var shutdownFunc func()

	tracerOnce.Do(func() {
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

		shutdownFunc = func() {
			if err := tp.Shutdown(ctx); err != nil {
				log.Fatalf("failed to shutdown TracerProvider: %v", err)
			}
		}
	})

	return shutdownFunc
}

func GetTracer() trace.Tracer {
	return tracer
}
