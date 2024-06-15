package weather

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

type WeatherResponse struct {
	Current CurrentWeather `json:"current"`
}

type CurrentWeather struct {
	FeelsLikeC float64 `json:"feelslike_c"`
}

func GetWeather(city string) (float64, error) {

	apiKey := os.Getenv("WEATHER_API_KEY")

	encodedCity := url.QueryEscape(city)

	url := fmt.Sprintf("http://api.weatherapi.com/v1/current.json?key=%s&q=%s&aqi=no",
		apiKey,
		encodedCity)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("error to build request")
		return 0, fmt.Errorf("internal error")
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error to do request to get city:%v - error:%v\n", city, err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("error reading response body: %v\n", err)
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("response expected 200 but got %v\n", resp.StatusCode)
	}

	var weatherResponse WeatherResponse

	err = json.Unmarshal(body, &weatherResponse)
	if err != nil {
		log.Println("Error unmarshalling JSON:", err)
		return 0, fmt.Errorf("error decode json")
	}

	return weatherResponse.Current.FeelsLikeC, nil
}
