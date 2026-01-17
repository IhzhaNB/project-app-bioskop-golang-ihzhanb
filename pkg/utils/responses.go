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

// ResponseJSON writes JSON response with custom status code
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

// ------------- Success responses -------------

// returns 200 OK
func ResponseSuccess(w http.ResponseWriter, message string, data any) {
	ResponseJSON(w, http.StatusOK, true, message, data, nil)
}

// returns 201 Created
func ResponseCreated(w http.ResponseWriter, message string, data any) {
	ResponseJSON(w, http.StatusCreated, true, message, data, nil)
}

// ------------- Error responses -------------

// returns 400 Bad Request
func ResponseBadRequest(w http.ResponseWriter, message string, errors any) {
	ResponseJSON(w, http.StatusBadRequest, false, message, nil, errors)
}

// returns 401 Unauthorized
func ResponseUnauthorized(w http.ResponseWriter, message string) {
	ResponseJSON(w, http.StatusUnauthorized, false, message, nil, nil)
}

// returns 403 Forbidden
func ResponseForbidden(w http.ResponseWriter, message string) {
	ResponseJSON(w, http.StatusForbidden, false, message, nil, nil)
}

// returns 404 Not Found
func ResponseNotFound(w http.ResponseWriter, message string) {
	ResponseJSON(w, http.StatusNotFound, false, message, nil, nil)
}

// returns 500 Internal Server Error
func ResponseInternalError(w http.ResponseWriter, message string) {
	ResponseJSON(w, http.StatusInternalServerError, false, message, nil, nil)
}
