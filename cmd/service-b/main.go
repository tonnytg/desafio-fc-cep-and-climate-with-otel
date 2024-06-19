// cmd/service-b/main.go
package main

import (
	"encoding/json"
	"fmt"
	"github.com/tonnytg/desafio-fc-cep-and-climate-with-otel/internal/domain"
	"github.com/tonnytg/desafio-fc-cep-and-climate-with-otel/internal/infra/otel_provider"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"log"
	"net/http"
	"os"
)

type ErrorMessage struct {
	Message string `json:"message"`
}

const name = "service-b"

var (
	tracer  = otel.Tracer(name)
	meter   = otel.Meter(name)
	logger  = otelslog.NewLogger(name)
	rollCnt metric.Int64Counter
)

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

	// Configuração do OpenTelemetry
	shutdown, err := otel_provider.SetupOTelSDK(r.Context())
	if err != nil {
		log.Fatalf("failed to setup OpenTelemetry SDK: %v", err)
	}
	defer func() {
		if err := shutdown(r.Context()); err != nil {
			log.Fatalf("failed to shutdown OpenTelemetry SDK: %v", err)
		}
	}()

	tracer := otel.Tracer("service-b")
	propagator := otel.GetTextMapPropagator()

	ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
	ctx, span := tracer.Start(ctx, "service_b-handler: check cep and weather")
	defer span.End()

	span.SetAttributes(attribute.String("service.name", "service-b"))

	w.Header().Set("Content-Type", "application/json")

	var data struct {
		CEP string `json:"cep"`
	}

	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {

		_ = ReplyRequest(w, http.StatusBadRequest, "no zipcode provided")
		return
	}

	location, err := domain.NewLocation(data.CEP)
	if err != nil {

		log.Println(err)
		_ = ReplyRequest(w, http.StatusUnprocessableEntity, "invalid zipcode")
		return
	}

	repo := domain.NewLocationRepository()
	serv := domain.NewLocationService(repo)

	err = serv.Execute(ctx, location)
	if err != nil {
		errorCode := err.Error()

		if errorCode == "422" {

			log.Println("invalid zipcode", location)

			_ = ReplyRequest(w, http.StatusUnprocessableEntity, "invalid zipcode")
		} else if errorCode == "404" {

			log.Println("can not find zipcode")

			_ = ReplyRequest(w, http.StatusNotFound, "can not find zipcode")

		} else {

			log.Println("internal server error")

			_ = ReplyRequest(w, http.StatusInternalServerError, "internal server error")

		}
		return
	}

	byteResponseData, err := json.Marshal(location)
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

	log.Println("Start Weather Collector listen in port:", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Panicf("error to start http server")
	}
}

func main() {
	log.Println("Start Service B")

	StartCepCollector()
}
