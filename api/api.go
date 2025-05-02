package api

import (
	"encoding/json"
	"net/http"
	//"github.com/joelvarghese6/mitigate-dust-attacks/internal/tools"
)

type CoinBalanceParams struct {
	Username string
}

type CheckDustedParams struct {
	Publickey string
}

type CheckDustedResponse struct {
	Code int
	Items []map[string]interface{}
}


type CheckAddressPoisoningResponse struct {
	Code     int                             `json:"code"`
	Poisoned bool                            `json:"poisoned"`
	Matches  []AddressPoisoningMatchDetails  `json:"matches"`
}

type AddressPoisoningMatchDetails struct {
	Signature      string `json:"signature"`
	FromAddress    string `json:"from_address"`
	SimilarAddress string `json:"similar_address"`
	Amount         uint64 `json:"amount"`
}

type GenerateFullReportResponse struct {
	Code int
	Details string
}

type CoinBalanceResponse struct {
	Code int
	Balance int64
}

type Error struct {
	Code int
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