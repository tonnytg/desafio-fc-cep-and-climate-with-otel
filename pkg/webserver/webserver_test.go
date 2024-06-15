package webserver

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWebserverGET(t *testing.T) {

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handlerIndex(w, req)
	res := w.Result()
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("expected error to be nil and got %v", err)
	}

	type Response struct {
		Message string `json:"message"`
	}

	expected := Response{
		Message: "no zipcode provided",
	}

	var received Response
	err = json.Unmarshal(data, &received)
	if err != nil {
		t.Errorf("error to decoding response")
	}

	if expected != received {
		t.Errorf("expected %v but got %v", received, received)
	}
}
