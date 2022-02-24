package handlers

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func HelloHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	who := vars["who"]
	response := fmt.Sprintf("Hello %s", who)
	fmt.Fprint(w, response)
}
