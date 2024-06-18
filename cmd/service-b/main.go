package main

import (
	"encoding/json"
	"fmt"
	"github.com/tonnytg/desafio-fc-cep-and-climate-with-otel/internal/domain"
	"github.com/tonnytg/desafio-fc-cep-and-climate-with-otel/internal/infra/otel"
	"go.opentelemetry.io/otel/attribute"
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
	ctxSpanFull, spanFull := tr.Start(r.Context(), "handlerIndex - Start process")
	defer spanFull.End()

	w.Header().Set("Content-Type", "application/json")

	var data struct {
		CEP string `json:"cep"`
	}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {

		_ = ReplyRequest(w, http.StatusBadRequest, "no zipcode provided")

		spanFull.SetAttributes(attribute.String("error", "no zipcode provided"))
		return
	}

	location, err := domain.NewLocation(data.CEP)
	if err != nil {

		log.Println(err)

		spanFull.SetAttributes(attribute.String("error", fmt.Sprintf("invalid zipcode: %v", err)))

		_ = ReplyRequest(w, http.StatusUnprocessableEntity, "invalid zipcode")
		return
	}

	repo := domain.NewLocationRepository()
	serv := domain.NewLocationService(repo)

	ctxSpanExecute, spanExecute := tr.Start(ctxSpanFull, "Preparing Execute - getting city and weather")
	defer spanExecute.End()

	err = serv.Execute(ctxSpanExecute, location)
	if err != nil {
		errorCode := err.Error()

		if errorCode == "422" {

			log.Println("invalid zipcode", location)

			spanExecute.SetAttributes(attribute.String("error", "invalid zipcode"))

			_ = ReplyRequest(w, http.StatusUnprocessableEntity, "invalid zipcode")
		} else if errorCode == "404" {

			log.Println("can not find zipcode")

			spanExecute.SetAttributes(attribute.String("error", "can not find zipcode"))

			_ = ReplyRequest(w, http.StatusNotFound, "can not find zipcode")

		} else {

			log.Println("internal server error")

			spanExecute.SetAttributes(attribute.String("error", "internal server error"))

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

	cleanup := otel.InitTracer(
		"service-b",
		attribute.String("deployment.environment", "production"),
		attribute.String("service.name", "service-b"),
		attribute.String("service.version", "1.0.0"),
		attribute.String("service.instance.id", "instance-456"),
		attribute.String("host.name", "host-def"),
		attribute.String("host.id", "host-id-789"),
		attribute.String("telemetry.sdk.name", "opentelemetry"),
		attribute.String("telemetry.sdk.language", "go"),
		attribute.String("telemetry.sdk.version", "1.0.0"),
		attribute.String("component", "cep-checker"),
		attribute.String("responsibility", "check cep"),
	)
	defer cleanup()

	StartCepCollector()
}
