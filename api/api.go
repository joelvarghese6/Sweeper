package api

import (
	"encoding/json"
	"net/http"
)

type CoinBalanceParams struct {
	Username string
}

type CheckDustedParams struct {
	Publickey string
}

type CheckDustedResponse struct {
	//response code
	Code int

	//whether it is dusted
	Items []map[string]interface{}
}

type CoinBalanceResponse struct {
	// response
	Code int

	//is there dust
	Balance int64
}

type Error struct {
	// Error code
	Code int

	//Error Message
	Message string
}

func writeError(w http.ResponseWriter, message string, code int) {
	resp := Error {
		Code: code,
		Message: message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	json.NewEncoder(w).Encode(resp)
}

var (
	RequestErrorHandler = func(w http.ResponseWriter, err error) {
		writeError(w, err.Error(), http.StatusBadRequest)
	}
	InternalErrorHandler = func(w http.ResponseWriter) {
		writeError(w, "An Unexpected Error Occured.", http.StatusInternalServerError)
	}
)