package middlewares

import "net/http"

type Middleware func(http.HandlerFunc) http.HandlerFunc

func MultipleMiddleware(h http.HandlerFunc, m ...Middleware) http.HandlerFunc {

	if len(m) < 1 {
		return h
	}

	wrapped := h

	for i := range m {
		wrapped = m[i](wrapped)
	}

	return wrapped
}
