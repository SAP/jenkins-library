package middlewares

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/golang-app/pkg/http/handlers"
)

func TestRecoverMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		handler    http.HandlerFunc
		statusCode int
		response   string
	}{
		{
			"test 1",
			handlers.ThrowPanicHandler,
			http.StatusInternalServerError,
			fmt.Sprintln(http.StatusText(http.StatusInternalServerError)),
		}, {
			"test 2",
			func(w http.ResponseWriter, r *http.Request) {
				panic("some other panic")
			},
			http.StatusInternalServerError,
			fmt.Sprintln(http.StatusText(http.StatusInternalServerError)),
		}, {
			"test 3",
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("text"))
			},
			http.StatusOK,
			"text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(MultipleMiddleware(tt.handler, RecoverMiddleware)))
			defer ts.Close()

			res, err := http.Get(ts.URL)
			if err != nil {
				t.Fatal(err)
			}

			actualStatusCode := res.StatusCode
			if actualStatusCode != tt.statusCode {
				t.Errorf("\nactual: %v\nexpected: %v\n", actualStatusCode, tt.statusCode)
			}

			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			res.Body.Close()

			actualBody := string(body)

			if tt.response != actualBody {
				t.Errorf("\nactual: %q\nexpected: %q\n", actualBody, tt.response)
			}
		})
	}
}
