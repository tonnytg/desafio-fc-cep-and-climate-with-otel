package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/tonnytg/desafio-fc-cep-and-climate-with-otel/internal/domain"
	"github.com/tonnytg/desafio-fc-cep-and-climate-with-otel/internal/infra/otel"
	"go.opentelemetry.io/otel/attribute"
	"io"
	"log"
	"net/http"
	"os"
)

type ErrorMessage struct {
	Message string `json:"message"`
}

func ReplyRequest(w http.ResponseWriter, statusCode int, msg string) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	replyMessage := ErrorMessage{Message: msg}

	err := json.NewEncoder(w).Encode(replyMessage)
	if err != nil {
		w.WriteHeader(statusCode)
		log.Println("")
		return fmt.Errorf("error to try reply request")
	}

	return nil
}

func handlerIndex(w http.ResponseWriter, r *http.Request) {
	tr := otel.GetTracer()
	ctx, span := tr.Start(r.Context(), "handlerIndex")
	defer span.End()

	w.Header().Set("Content-Type", "application/json")

	var data struct {
		CEP string `json:"cep"`
	}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		_ = ReplyRequest(w, http.StatusBadRequest, "no zipcode provided")
		span.SetAttributes(attribute.String("error", "no zipcode provided"))
		otel.Logger(ctx, "no zipcode provided")
		return
	}

	l, err := domain.NewLocation(data.CEP)
	if err != nil {
		log.Println(err)
		span.SetAttributes(attribute.String("error", fmt.Sprintf("invalid zipcode: %v", err)))
		otel.Logger(ctx, fmt.Sprintf("invalid zipcode: %v", err))
		_ = ReplyRequest(w, http.StatusUnprocessableEntity, "invalid zipcode")
		return
	}

	err = l.Validate()
	if err != nil {
		log.Println("invalid zipcode", l)
		span.SetAttributes(attribute.String("error", fmt.Sprintf("invalid zipcode: %v", l)))
		otel.Logger(ctx, fmt.Sprintf("invalid zipcode: %v", l))
		_ = ReplyRequest(w, http.StatusUnprocessableEntity, "invalid zipcode")
		return
	}

	serviceB := "http://service-b:8080"

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Println("error marshaling data:", err)
		span.SetAttributes(attribute.String("error", fmt.Sprintf("error marshaling data: %v", err)))
		otel.Logger(ctx, fmt.Sprintf("error marshaling data: %v", err))
		_ = ReplyRequest(w, http.StatusInternalServerError, "internal server error")
		return
	}

	serviceBSpanCtx, serviceBSpan := tr.Start(ctx, "call-service-b")
	defer serviceBSpan.End()

	log.Println("start request to service b passing cep:", l.GetCEP())
	serviceBSpan.SetAttributes(attribute.String("start request", l.GetCEP()))
	otel.Logger(ctx, fmt.Sprintf("start request to service b passing cep: %s", l.GetCEP()))

	resp, err := http.Post(serviceB, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("error sending data to service-b:", err)
		serviceBSpan.SetAttributes(attribute.String("error", fmt.Sprintf("error sending data to service-b: %v", err)))
		otel.Logger(serviceBSpanCtx, fmt.Sprintf("error sending data to service-b: %v", err))
		_ = ReplyRequest(w, http.StatusInternalServerError, "internal server error")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("service-b returned non-OK status:", resp.Status)
		serviceBSpan.SetAttributes(attribute.String("error", fmt.Sprintf("service-b returned non-OK status: %v", resp.Status)))
		otel.Logger(serviceBSpanCtx, fmt.Sprintf("service-b returned non-OK status: %v", resp.Status))
		if resp.StatusCode == 404 {
			_ = ReplyRequest(w, resp.StatusCode, "can not find zipcode")
		} else if resp.StatusCode == 422 {
			_ = ReplyRequest(w, resp.StatusCode, " invalid zipcode")
		} else {
			_ = ReplyRequest(w, resp.StatusCode, "bad request")
		}
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("error to io.ReadAll resp.Body")
		serviceBSpan.SetAttributes(attribute.String("error", fmt.Sprintf("error to io.ReadAll resp.Body: %v", err)))
		otel.Logger(serviceBSpanCtx, fmt.Sprintf("error to io.ReadAll resp.Body: %v", err))
		_ = ReplyRequest(w, http.StatusInternalServerError, "internal server error")
		return
	}

	log.Println("body:", string(body))
	serviceBSpan.SetAttributes(attribute.String("body", string(body)))
	otel.Logger(serviceBSpanCtx, fmt.Sprintf("body: %s", string(body)))

	var responseData struct {
		City  string  `json:"city"`
		TempC float64 `json:"temp_c"`
		TempF float64 `json:"temp_f"`
		TempK float64 `json:"temp_k"`
	}

	err = json.Unmarshal(body, &responseData)
	if err != nil {
		log.Println("error to unMarshall resp.Body")
		serviceBSpan.SetAttributes(attribute.String("error", fmt.Sprintf("error to unMarshall resp.Body: %v", err)))
		otel.Logger(serviceBSpanCtx, fmt.Sprintf("error to unMarshall resp.Body: %v", err))
		_ = ReplyRequest(w, http.StatusInternalServerError, "internal server error")
		return
	}

	byteResponseData, err := json.Marshal(responseData)
	if err != nil {
		_ = ReplyRequest(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(byteResponseData)
}

func StartCepCollector() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlerIndex)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Start CEP Collector listen in port:", port)
	otel.Logger(context.Background(), fmt.Sprintf("Start CEP Collector listen in port: %s", port))
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		otel.Logger(context.Background(), fmt.Sprintf("error to start http server: %v", err))
		log.Panicf("error to start http server")
	}
}

func main() {
	log.Println("Start Service A")

	cleanup := otel.InitTracer(
		"service-a",
		attribute.String("deployment.environment", "production"),
		attribute.String("service.name", "service-a"),
		attribute.String("service.version", "1.0.0"),
		attribute.String("service.instance.id", "instance-123"),
		attribute.String("host.name", "host-abc"),
		attribute.String("host.id", "host-id-456"),
		attribute.String("telemetry.sdk.name", "opentelemetry"),
		attribute.String("telemetry.sdk.language", "go"),
		attribute.String("telemetry.sdk.version", "1.0.0"),
		attribute.String("component", "cep-checker"),
		attribute.String("responsibility", "check cep"),
	)
	defer cleanup()

	otel.Logger(context.Background(), "Start Service A")
	StartCepCollector()
}
