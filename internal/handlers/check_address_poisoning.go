package handlers

import (
	"encoding/json"
	"fmt"
	"time"

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

	//tools.PrintInstructionTypes(signatures[3])

	for _, sig := range signatures {
		tools.PrintInstructionTypes(sig)
		fmt.Println(sig)
		time.Sleep(1000 * time.Millisecond)
	}
	

	// if err != nil {
	// 	log.Error(err)
	// 	api.InternalErrorHandler(w)
	// 	return
	// }

	var response = api.CheckAddressPoisoningResponse {
		Code: http.StatusOK,
		Poisoned: false,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}
}