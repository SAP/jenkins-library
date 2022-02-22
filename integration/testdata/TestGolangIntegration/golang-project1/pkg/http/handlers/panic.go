package handlers

import (
	"errors"
	"net/http"
)

var PanicError = errors.New("panic was thrown")

func ThrowPanicHandler(w http.ResponseWriter, r *http.Request) {
	panic(PanicError)
}
