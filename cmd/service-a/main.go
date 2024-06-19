package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/tonnytg/desafio-fc-cep-and-climate-with-otel/internal/domain"
	"github.com/tonnytg/desafio-fc-cep-and-climate-with-otel/internal/infra/otel_provider"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"io"
	"log"
	"net/http"
	"os"

	"go.opentelemetry.io/contrib/bridges/otelslog"
)

type ErrorMessage struct {
	Message string `json:"message"`
}

const name = "service-a"

var (
	tracer = otel.Tracer(name)
	meter  = otel.Meter(name)
	logger = otelslog.NewLogger(name)
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

	ctx := context.Background()

	// Configuração do OpenTelemetry
	shutdown, err := otel_provider.SetupOTelSDK(ctx)
	if err != nil {
		log.Fatalf("failed to setup OpenTelemetry SDK: %v", err)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Fatalf("failed to shutdown OpenTelemetry SDK: %v", err)
		}
	}()

	ctx, span := tracer.Start(r.Context(), "check-cep")
	defer span.End()

	span.SetAttributes(attribute.String("service.name", "service-a"))

	w.Header().Set("Content-Type", "application/json")

	var data struct {
		CEP string `json:"cep"`
	}

	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		_ = ReplyRequest(w, http.StatusBadRequest, "no zipcode provided")
		return
	}

	l, err := domain.NewLocation(data.CEP)
	if err != nil {

		log.Println(err)
		_ = ReplyRequest(w, http.StatusUnprocessableEntity, "invalid zipcode")
		return
	}

	err = l.Validate()
	if err != nil {

		log.Println("invalid zipcode", l)
		_ = ReplyRequest(w, http.StatusUnprocessableEntity, "invalid zipcode")
		return
	}

	serviceB := "http://service-b:8080"

	jsonData, err := json.Marshal(data)
	if err != nil {

		log.Println("error marshaling data:", err)
		_ = ReplyRequest(w, http.StatusInternalServerError, "internal server error")
		return
	}

	log.Println("start request to service b passing cep:", l.GetCEP())

	client := &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
	req, err := http.NewRequestWithContext(ctx, "POST", serviceB, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("error creating request:", err)

		_ = ReplyRequest(w, http.StatusInternalServerError, "internal server error")
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println("error making request to service b:", err)
		_ = ReplyRequest(w, http.StatusInternalServerError, "internal server error")
		return
	}
	defer resp.Body.Close()

	// Additional processing of the response can go here

	if resp.StatusCode != http.StatusOK {

		log.Println("service-b returned non-OK status:", resp.Status)

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
		log.Panicf("error to start http server")
	}
}

func main() {
	log.Println("Start Service A")

	StartCepCollector()
}
