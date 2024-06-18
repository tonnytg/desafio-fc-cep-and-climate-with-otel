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
	ctxFull, spanFull := tr.Start(r.Context(), "getting information about cep")
	defer spanFull.End()

	w.Header().Set("Content-Type", "application/json")

	var data struct {
		CEP string `json:"cep"`
	}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		_ = ReplyRequest(w, http.StatusBadRequest, "no zipcode provided")

		spanFull.SetAttributes(attribute.String("error", "no zipcode provided"))
		otel.Logger(ctxFull, "no zipcode provided")

		return
	}

	l, err := domain.NewLocation(data.CEP)
	if err != nil {

		log.Println(err)

		spanFull.SetAttributes(attribute.String("error", fmt.Sprintf("invalid zipcode: %v", err)))
		otel.Logger(ctxFull, fmt.Sprintf("invalid zipcode: %v", err))

		_ = ReplyRequest(w, http.StatusUnprocessableEntity, "invalid zipcode")
		return
	}

	err = l.Validate()
	if err != nil {

		log.Println("invalid zipcode", l)

		spanFull.SetAttributes(attribute.String("error", fmt.Sprintf("invalid zipcode: %v", l)))
		otel.Logger(ctxFull, fmt.Sprintf("invalid zipcode: %v", l))

		_ = ReplyRequest(w, http.StatusUnprocessableEntity, "invalid zipcode")
		return
	}

	serviceB := "http://service-b:8080"

	jsonData, err := json.Marshal(data)
	if err != nil {

		log.Println("error marshaling data:", err)

		spanFull.SetAttributes(attribute.String("error", fmt.Sprintf("error marshaling data: %v", err)))
		otel.Logger(ctxFull, fmt.Sprintf("error marshaling data: %v", err))

		_ = ReplyRequest(w, http.StatusInternalServerError, "internal server error")
		return
	}

	log.Println("start request to service b passing cep:", l.GetCEP())

	ctxRequestServiceB, spanRequestServiceB := tr.Start(ctxFull, "getting information to service b")
	defer spanRequestServiceB.End()
	spanRequestServiceB.SetAttributes(attribute.String("start request", l.GetCEP()))

	resp, err := http.Post(serviceB, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {

		log.Println("error sending data to service-b:", err)

		spanRequestServiceB.SetAttributes(attribute.String("error", fmt.Sprintf("error sending data to service-b: %v", err)))
		otel.Logger(ctxRequestServiceB, fmt.Sprintf("error sending data to service-b: %v", err))

		_ = ReplyRequest(w, http.StatusInternalServerError, "internal server error")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {

		log.Println("service-b returned non-OK status:", resp.Status)

		spanRequestServiceB.SetAttributes(attribute.String("error", fmt.Sprintf("service-b returned non-OK status: %v", resp.Status)))
		otel.Logger(ctxRequestServiceB, fmt.Sprintf("service-b returned non-OK status: %v", resp.Status))

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

		spanRequestServiceB.SetAttributes(attribute.String("error", fmt.Sprintf("error to io.ReadAll resp.Body: %v", err)))
		otel.Logger(ctxRequestServiceB, fmt.Sprintf("error to io.ReadAll resp.Body: %v", err))

		_ = ReplyRequest(w, http.StatusInternalServerError, "internal server error")
		return
	}

	var responseData struct {
		City  string  `json:"city"`
		TempC float64 `json:"temp_c"`
		TempF float64 `json:"temp_f"`
		TempK float64 `json:"temp_k"`
	}

	err = json.Unmarshal(body, &responseData)
	if err != nil {
		log.Println("error to unMarshall resp.Body")
		spanRequestServiceB.SetAttributes(attribute.String("error", fmt.Sprintf("error to unMarshall resp.Body: %v", err)))
		otel.Logger(ctxRequestServiceB, fmt.Sprintf("error to unMarshall resp.Body: %v", err))
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

	StartCepCollector()
}
