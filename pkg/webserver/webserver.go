package webserver

import (
	"encoding/json"
	"fmt"
	"github.com/tonnytg/desafio-fc-cep-and-climate-with-otel/internal/domain"
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

	w.Header().Set("Content-Type", "application/json")

	var data struct {
		CEP string `json:"cep"`
	}

	err := json.NewDecoder(r.Body).Decode(&data)
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

	repo := domain.NewLocationRepository()
	serv := domain.NewLocationService(repo)

	err = serv.Execute(l)
	if err != nil {

		errorCode := err.Error()

		if errorCode == "422" {
			log.Println("invalid zipcode", l)
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

	var responseData = domain.Location{}

	byteResponseData, err := json.Marshal(responseData)
	if err != nil {
		_ = ReplyRequest(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(byteResponseData)

	return
}

func Start() {

	mux := http.NewServeMux()

	mux.HandleFunc("/", handlerIndex)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Start webserver listen in port:", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Panicf("error to start http server")
	}
}
