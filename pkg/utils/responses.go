package utils

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

func ResponseJSON(w http.ResponseWriter, code int, status bool, message string, data, errors any) {
	response := Response{
		Status:  status,
		Message: message,
		Data:    data,
		Errors:  errors,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}

// Success responses
func ResponseSuccess(w http.ResponseWriter, message string, data any) {
	ResponseJSON(w, http.StatusOK, true, message, data, nil)
}

func ResponseCreated(w http.ResponseWriter, message string, data any) {
	ResponseJSON(w, http.StatusCreated, true, message, data, nil)
}

// Error responses
func ResponseBadRequest(w http.ResponseWriter, message string, errors any) {
	ResponseJSON(w, http.StatusBadRequest, false, message, nil, errors)
}

func ResponseUnauthorized(w http.ResponseWriter, message string) {
	ResponseJSON(w, http.StatusUnauthorized, false, message, nil, nil)
}

func ResponseForbidden(w http.ResponseWriter, message string) {
	ResponseJSON(w, http.StatusForbidden, false, message, nil, nil)
}

func ResponseNotFound(w http.ResponseWriter, message string) {
	ResponseJSON(w, http.StatusNotFound, false, message, nil, nil)
}

func ResponseInternalError(w http.ResponseWriter, message string) {
	ResponseJSON(w, http.StatusInternalServerError, false, message, nil, nil)
}
