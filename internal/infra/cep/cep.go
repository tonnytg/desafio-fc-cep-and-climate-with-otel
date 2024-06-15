package cep

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type ViaCepResponse struct {
	Cep        string `json:"cep"`
	Localidade string `json:"localidade"`
}

func GetCity(cep string) (string, error) {

	url := fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("error to build request")
		return "", fmt.Errorf("internal error")
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error to do request to get cep:%v - error:%v\n", cep, err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v\n", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error to get cep data")
	}

	var viacepResponse ViaCepResponse

	err = json.Unmarshal(body, &viacepResponse)
	if err != nil {
		log.Println("Error unmarshalling JSON:", err)
		return "", fmt.Errorf("error decode json")
	}

	return viacepResponse.Localidade, nil
}
