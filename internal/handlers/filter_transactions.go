package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/schema"
	"github.com/joelvarghese6/mitigate-dust-attacks/api"

	"github.com/joelvarghese6/mitigate-dust-attacks/internal/tools"

	log "github.com/sirupsen/logrus"
)

func FilterSuspiciousTransfers(transfers []tools.TransferInfo) []tools.TransferInfo {
	var filtered []tools.TransferInfo
	for _, t := range transfers {
		if t.Dust && t.MultipleAccounts {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

func FilterTransactions(w http.ResponseWriter, r *http.Request) {
	
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

	var allTransfers []tools.TransferInfo

	for _, sig := range signatures {
		infos, err := tools.AnalyzeSystemTransferToAddress(sig, address)

		if err != nil {
			continue
		}
	
		allTransfers = append(allTransfers, infos...)
		time.Sleep(50 * time.Millisecond)
	}

	allTransfers = FilterSuspiciousTransfers(allTransfers)

	var labeledTransfers []map[string]interface{}

	for _, tx := range allTransfers {
		labeledTransfers = append(labeledTransfers, map[string]interface{}{
			"Other": tx.OtherAccount,
			"Amount": tx.Amount,
			"Dust": tx.Dust,
			"Multi": tx.MultipleAccounts,
			"Sig": tx.Signature,
		})
	}
	
	var response = api.CheckDustedResponse {
		Code: http.StatusOK,
		Items: labeledTransfers,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}
}