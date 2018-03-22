package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/dispatchlabs/dapos"
	daposCore "github.com/dispatchlabs/dapos/core"

	httpService "github.com/dispatchlabs/commons/services"
	"github.com/dispatchlabs/disgover"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/dispatchlabs/commons/utils"
)



// registerHttpHandlers
func registerHttpHandlers() {
	httpService.GetHttpRouter().HandleFunc("/v1/ping", pingPongHandler).Methods("POST")
	httpService.GetHttpRouter().HandleFunc("/v1/balance/{address}", retrieveBalanceHandler).Methods("GET")
	httpService.GetHttpRouter().HandleFunc("/v1/sync_transactions", syncTransactionsHandler).Methods("GET")
	httpService.GetHttpRouter().HandleFunc("/v1/transactions/{address}", retrieveTransactionsHandler).Methods("GET")
	httpService.GetHttpRouter().HandleFunc("/v1/transactions", createTransactionHandler).Methods("POST")

}

// retrieveBalanceHandler
func retrieveBalanceHandler(responseWriter http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	balance, error := dapos.GetDAPoS().GetBalance(vars["address"])
	if error != nil {
		responseWriter.Write([]byte(`{"status":"INTERNAL_SERVER_ERROR"}`))
		return
	}
	bytes, error := json.Marshal(struct {
		Status  string `json:"status,omitempty"`
		Balance int64  `json:"balance,omitempty"`
	}{
		Status:  "OK",
		Balance: balance,
	})
	if error != nil {
		log.WithFields(log.Fields{
			"method": utils.GetCallingFuncName(),
		}).Error("JSON parse error [error=" + error.Error() + "]")
		http.Error(responseWriter, `{"status":"JSON_PARSE_ERROR"}`, http.StatusBadRequest)
		return
	}
	responseWriter.Write(bytes)
}

// createTransactionHandler
func createTransactionHandler(responseWriter http.ResponseWriter, request *http.Request) {
	body, error := ioutil.ReadAll(request.Body)
	if error != nil {
		log.WithFields(log.Fields{
			"method": utils.GetCallingFuncName(),
		}).Error("unable to read HTTP body of request [error=" + error.Error() + "]")
		http.Error(responseWriter, `{"status":"INTERNAL_SERVER_ERROR"}`, http.StatusInternalServerError)
		return
	}

	// Unmarshal transaction?
	transaction := &daposCore.Transaction{}
	error = json.Unmarshal(body, transaction)
	if error != nil {
		log.WithFields(log.Fields{
			"method": utils.GetCallingFuncName(),
		}).Error("JSON parse error [error=" + error.Error() + "]")
		http.Error(responseWriter, `{"status":"JSON_PARSE_ERROR"}`, http.StatusBadRequest)
		return
	}
	dapos.GetDAPoS().ProcessTx(transaction)
	responseWriter.Write([]byte(`{"status":"OK"}`))
}

//
func syncTransactionsHandler(responseWriter http.ResponseWriter, request *http.Request) {
	dapos.GetDAPoS().SynchronizeTransactions()
}

// retrieveTransactionsHandler
func retrieveTransactionsHandler(responseWriter http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	transactions := dapos.GetDAPoS().GetTransactions(vars["address"])
	bytes, error := json.Marshal(struct {
		Status       string                  `json:"status,omitempty"`
		Transactions []daposCore.Transaction `json:"transactions,omitempty"`
	}{
		Status:       "OK",
		Transactions: transactions,
	})
	if error != nil {
		log.WithFields(log.Fields{
			"method": utils.GetCallingFuncName(),
		}).Error("JSON parse error [error=" + error.Error() + "]")
		http.Error(responseWriter, `{"status":"JSON_PARSE_ERROR"}`, http.StatusBadRequest)
		return
	}
	responseWriter.Write(bytes)
}

func pingPongHandler(responseWriter http.ResponseWriter, request *http.Request) {
	body, _ := ioutil.ReadAll(request.Body)

	fmt.Println(string(body))

	responseWriter.Write([]byte(fmt.Sprintf(
		"PONG-From: %s @ %s:%d",
		disgover.GetDisgover().ThisContact.Address,
		disgover.GetDisgover().ThisContact.Endpoint.Host,
		disgover.GetDisgover().ThisContact.Endpoint.Port,
	)))
}
