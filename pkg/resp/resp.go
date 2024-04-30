package resp

import (
	"encoding/json"
	"net/http"
)

type H = map[string]interface{}

func Error(w http.ResponseWriter, message string) {
	w.WriteHeader(http.StatusBadRequest)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(H{"error": message})
}

func JSON(w http.ResponseWriter, code int, data interface{}) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func OK(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, data)
}
