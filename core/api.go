package core

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"

	httpService "github.com/dispatchlabs/disgo_commons/services"
	"github.com/dispatchlabs/disgo_commons/types"
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
	this.router.HandleFunc("/v1/wallet/{wallet_address}", this.retrieveBalanceHandler).Methods("GET")
	this.router.HandleFunc("/v1/transactions/{wallet_address}", this.retrieveTransactionsHandler).Methods("GET")
	this.router.HandleFunc("/v1/transactions", this.createTransactionHandler).Methods("POST")
	return &this

}

// retrieveBalanceHandler
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
	transaction := &types.Transaction{}
	error = json.Unmarshal(body, transaction)
	if error != nil {
		log.WithFields(log.Fields{
			"method": "Api.createTransactionHandler",
		}).Error("JSON parse error [error=" + error.Error() + "]")
		http.Error(responseWriter, `{"status":"JSON_PARSE_ERROR"}`, http.StatusBadRequest)
		return
	}

	// Verify?
	if !transaction.Verify() {
		log.WithFields(log.Fields{
			"method": "Api.createTransactionHandler",
		}).Error("invalid transaction")
		http.Error(responseWriter, `{"status":"INVALID_TRANSACTION"}`, http.StatusBadRequest)
		return
	}

	// TODO: Remove (just for flushing out API).
	// _, error = this.getService(&dapos.DAPoSService{}).(*dapos.DAPoSService).CreateTransaction(transaction, nil)
	log.WithFields(log.Fields{
		"method": "Api.createTransactionHandler",
	}).Info("valid transaction")
	responseWriter.Write([]byte(`{"status":"OK"}`))
}

// retrieveTransactionsHandler
func (this *Api) retrieveTransactionsHandler(responseWriter http.ResponseWriter, request *http.Request) {
	//vars := mux.Vars(request)
}

// getService
func (this *Api) getService(serviceInterface interface{}) types.IService {
	for _, service := range this.services {
		if reflect.TypeOf(service) == reflect.TypeOf(serviceInterface) {
			return service
		}
	}
	return nil
}
