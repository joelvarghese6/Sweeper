package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/schema"
	"github.com/joelvarghese6/mitigate-dust-attacks/api"

	"github.com/joelvarghese6/mitigate-dust-attacks/internal/tools"

	log "github.com/sirupsen/logrus"
)

func GenerateDetailedReport(w http.ResponseWriter, r *http.Request) {
	
	var params = api.CheckDustedParams{}
	var decoder *schema.Decoder = schema.NewDecoder()
	var err error

	err = decoder.Decode(&params, r.URL.Query())

	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}

	if params.Publickey == "" || !tools.IsValidSolanaAddress(params.Publickey) {
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}

	var response = api.GenerateFullReportResponse {
		Code: http.StatusOK,
		Details: "Heyyy",
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}
}