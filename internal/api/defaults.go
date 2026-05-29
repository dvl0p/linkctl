package api

import (
	"encoding/json"
	"net/http"
)

func handlerReplyWithError(w http.ResponseWriter, code int, message string) {

	type errorResponse struct {
		Error string `json:"error"`
	}

	handlerReplyWithJSON(w, code, errorResponse{
		Error: message,
	})
}

func handlerReplyWithJSON(w http.ResponseWriter,
	code int, payload interface{}) {
	reply, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(reply)
}
