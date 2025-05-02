package handlers

import (
	"encoding/json"
	//"fmt"

	"net/http"

	"github.com/gorilla/schema"
	"github.com/joelvarghese6/mitigate-dust-attacks/api"

	"github.com/joelvarghese6/mitigate-dust-attacks/internal/tools"

	log "github.com/sirupsen/logrus"
)

func CheckAddressPoisoning(w http.ResponseWriter, r *http.Request) {

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

	address := params.Publickey
	signatures, err := tools.GetSignatures(address)

	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}

	result, err := tools.DetectAddressPoisoning(signatures)

	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}

	response := api.CheckAddressPoisoningResponse{
		Code:     http.StatusOK,
		Poisoned: result.Count > 0,
	}

	for _, match := range result.Matches {
		response.Matches = append(response.Matches, api.AddressPoisoningMatchDetails{
			Signature:      match.Signature,
			FromAddress:    match.FromAddress,
			SimilarAddress: match.SimilarAddress,
			Amount:         match.Amount,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("Encoding response failed: %v", err)
		api.InternalErrorHandler(w)
		return
	}
}