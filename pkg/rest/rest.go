package rest

import (
	"encoding/json"
	"log"
	"net/http"
)

func PostRequest(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
		} else {
			h.ServeHTTP(w, r)
		}
	})
}

type errorResponse struct {
	Error string `json:"error"`
}

func Error(w http.ResponseWriter, r *http.Request, err error, code int) {
	message := err.Error()
	log.Printf(`error: %s "%s %s" %d: %s`, r.RemoteAddr, r.Method, r.URL, code, message)
	JSON(w, errorResponse{Error: message}, code)
}

func JSON(w http.ResponseWriter, v interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		log.Printf("failed to write response: %s", err)
	}
}
