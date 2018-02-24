package core

import (
	"github.com/dispatchlabs/disgo_commons/types"
	"github.com/gorilla/mux"
	"reflect"
	httpService "github.com/dispatchlabs/disgo_commons/services"
	"io/ioutil"
	"net/http"
	log "github.com/sirupsen/logrus"
	"encoding/json"
	dapos "github.com/dispatchlabs/dapos/core"
)

// Api
type Api struct {
	services []types.IService
	router *mux.Router
}

// NewApi
func NewApi(services []types.IService) *Api {
	this := Api{services, httpService.GetHttpRouter()}
	this.router.HandleFunc("/v1/wallet", this.createWalletHandler).Methods("POST")
	this.router.HandleFunc("/v1/wallet/{wallet_address}", this.retrieveWalletHandler).Methods("GET")
	this.router.HandleFunc("/v1/transactions", this.createTransactionHandler).Methods("POST")
	this.router.HandleFunc("/v1/transactions/{wallet_address}", this.retrieveTransactionHandler).Methods("GET")
	return &this

}

// createWalletHandler
func (this *Api) createWalletHandler(responseWriter http.ResponseWriter, request *http.Request) {
	/*
	body, error := ioutil.ReadAll(request.Body)
	if error != nil {
		log.WithFields(log.Fields{
			"method": "Server.createTransactionHandler",
		}).Error("unable to read HTTP body of request ", error)
		http.Error(responseWriter, "error reading HTTP body of request", http.StatusBadRequest)
		return
	}
	*/
	log.Info("create wallet")
}

// retrieveWalletHandler
func (this *Api) retrieveWalletHandler(responseWriter http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	responseWriter.WriteHeader(http.StatusOK)
	log.Info("retrieve wallet [address=" + vars["wallet_address"] + "]")
}

// createTransactionHandler
func (this *Api) createTransactionHandler(responseWriter http.ResponseWriter, request *http.Request) {
	body, error := ioutil.ReadAll(request.Body)
	if error != nil {
		log.WithFields(log.Fields{
			"method": "Server.createTransactionHandler",
		}).Error("unable to read HTTP body of request ", error)
		http.Error(responseWriter, "error reading HTTP body of request", http.StatusBadRequest)
		return
	}

	transaction := &types.Transaction{}
	error = json.Unmarshal(body, transaction)
	if error != nil {
		log.WithFields(log.Fields{
			"method": "Server.createTransactionHandler",
		}).Error("JSON_PARSE_ERROR ", error) // TODO: Should return JSON!!!
		http.Error(responseWriter, "error reading HTTP body of request", http.StatusBadRequest)
		return
	}

	// Create transaction.
	_, error = this.getService(&dapos.DAPoSService{}).(*dapos.DAPoSService).CreateTransaction(transaction, nil)
	if error != nil {
		log.WithFields(log.Fields{
			"method": "Server.createTransactionHandler",
		}).Error("JSON_PARSE_ERROR ", error) // TODO: Should return JSON!!!
		http.Error(responseWriter, "error reading HTTP body of request", http.StatusBadRequest)
		return
	}

	http.Error(responseWriter, "{}", http.StatusOK)

}

// retrieveTransactionHandler
func (this *Api) retrieveTransactionHandler(responseWriter http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	responseWriter.WriteHeader(http.StatusOK)
	log.Info("retrieve transactions [address=" + vars["wallet_address"] + "]")
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



