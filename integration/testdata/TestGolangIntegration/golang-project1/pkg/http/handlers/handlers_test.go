package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestHelloHandler(t *testing.T) {
	tests := []struct {
		name     string
		who      string
		response string
	}{
		{
			"test 1",
			"cepera",
			"Hello cepera",
		}, {
			"test 2",
			"guest",
			"Hello guest",
		}, {
			"test 3",
			"",
			fmt.Sprintln("404 page not found"),
		},
	}

	router := mux.NewRouter()
	router.HandleFunc("/hello/{who}", HelloHandler).Methods("GET")

	server := httptest.NewServer(router)
	defer server.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/hello/%s", server.URL, tt.who), nil)
			if err != nil {
				t.Error(err)
			}

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Error(err)
			}

			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()

			actual := string(body)

			if tt.response != actual {
				t.Errorf("\nactual: %q\nexpected: %q\n", actual, tt.response)
			}
		})
	}
}
