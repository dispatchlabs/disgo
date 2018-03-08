package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"time"

	"github.com/dispatchlabs/dapos"
	daposCore "github.com/dispatchlabs/dapos/core"

	httpService "github.com/dispatchlabs/disgo_commons/services"
	"github.com/dispatchlabs/disgo_commons/types"
	"github.com/dispatchlabs/disgover"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// Api
type Api struct {
	services []types.IService
	router   *mux.Router
}

// NewApi
func NewApi(services []types.IService) *Api {
	this := Api{services, httpService.GetHttpRouter()}
	this.router.HandleFunc("/v1/ping", this.pingPongHandler).Methods("POST")
	this.router.HandleFunc("/v1/wallet/{wallet_address}", this.retrieveBalanceHandler).Methods("GET")
	this.router.HandleFunc("/v1/transactions/{wallet_address}", this.retrieveTransactionsHandler).Methods("GET")
	this.router.HandleFunc("/v1/transactions/new", this.createTransactionHandler).Methods("POST")
	return &this

}

func (this *Api) retrieveBalanceHandler(responseWriter http.ResponseWriter, request *http.Request) {
	//vars := mux.Vars(request)
}

// createTransactionHandler
func (this *Api) createTransactionHandler(responseWriter http.ResponseWriter, request *http.Request) {
	body, error := ioutil.ReadAll(request.Body)
	if error != nil {
		log.WithFields(log.Fields{
			"method": "Api.createTransactionHandler",
		}).Error("unable to read HTTP body of request [error=" + error.Error() + "]")
		http.Error(responseWriter, `{"status":"INTERNAL_SERVER_ERROR"}`, http.StatusInternalServerError)
		return
	}

	// Unmarshal transaction?
	transaction := &daposCore.Transaction{}
	error = json.Unmarshal(body, transaction)
	if error != nil {
		log.WithFields(log.Fields{
			"method": "Api.createTransactionHandler",
		}).Error("JSON parse error [error=" + error.Error() + "]")
		http.Error(responseWriter, `{"status":"JSON_PARSE_ERROR"}`, http.StatusBadRequest)
		return
	}

	// Temporarily Comment This - Need to run TXes in DAPoS
	// // Verify?
	// if !transaction.Verify() {
	// 	log.WithFields(log.Fields{
	// 		"method": "Api.createTransactionHandler",
	// 	}).Error("invalid transaction")
	// 	http.Error(responseWriter, `{"status":"INVALID_TRANSACTION"}`, http.StatusBadRequest)
	// 	return
	// }

	// Pass TX to DAPoS
	var daposTx = &daposCore.Transaction{
		Hash:  transaction.Hash,
		From:  transaction.From,
		To:    transaction.To,
		Value: transaction.Value,
		Time:  time.Now(),
	}

	dapos.GetDAPoS().ProcessTx(daposTx)

	// TODO: Remove (just for flushing out API).
	// _, error = this.getService(&dapos.DAPoSService{}).(*dapos.DAPoSService).CreateTransaction(transaction, nil)
	log.WithFields(log.Fields{
		"method": "Api.createTransactionHandler",
	}).Info("valid transaction")
	responseWriter.Write([]byte(`{"status":"OK"}`))
}

func (this *Api) retrieveTransactionsHandler(responseWriter http.ResponseWriter, request *http.Request) {
	//vars := mux.Vars(request)
}

func (this *Api) getService(serviceInterface interface{}) types.IService {
	for _, service := range this.services {
		if reflect.TypeOf(service) == reflect.TypeOf(serviceInterface) {
			return service
		}
	}
	return nil
}

func (this *Api) pingPongHandler(responseWriter http.ResponseWriter, request *http.Request) {
	body, _ := ioutil.ReadAll(request.Body)

	fmt.Println(string(body))

	responseWriter.Write([]byte(fmt.Sprintf(
		"PONG-From: %s @ %s:%d",
		disgover.GetDisgover().ThisContact.Address,
		disgover.GetDisgover().ThisContact.Endpoint.Host,
		disgover.GetDisgover().ThisContact.Endpoint.Port,
	)))
}
