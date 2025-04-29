package handlers

import (
	"net/http"

	"github.com/gorilla/schema"
	"github.com/joelvarghese6/mitigate-dust-attacks/api"

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
}