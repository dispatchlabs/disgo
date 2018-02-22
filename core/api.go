package core

import (
	"github.com/dispatchlabs/disgo_commons/types"
	"github.com/gorilla/mux"
	"reflect"
	httpService "github.com/dispatchlabs/disgo_commons/services"
	"io/ioutil"
	"net/http"
	log "github.com/sirupsen/logrus"
	dapos "github.com/dispatchlabs/dapos/core"
)

// Api
type Api struct {
	services []types.IService
	router *mux.Router
}

// NewApi
func NewApi(services []types.IService) *Api {
	api := Api{services, httpService.GetHttpRouter()}
	api.router.HandleFunc("/v1/wallet", api.createWalletHandler).Methods("POST")
	api.router.HandleFunc("/v1/wallet/{wallet_address}", api.retrieveWalletHandler).Methods("GET")
	api.router.HandleFunc("/v1/transactions", api.createTransactionHandler).Methods("POST")
	api.router.HandleFunc("/v1/transactions/{wallet_address}", api.retrieveTransactionHandler).Methods("GET")
	return &api

}

// createWalletHandler
func (api *Api) createWalletHandler(responseWriter http.ResponseWriter, request *http.Request) {
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
func (api *Api) retrieveWalletHandler(responseWriter http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	responseWriter.WriteHeader(http.StatusOK)
	log.Info("retrieve wallet [address=" + vars["wallet_address"] + "]")
}

// createTransactionHandler
func (api *Api) createTransactionHandler(responseWriter http.ResponseWriter, request *http.Request) {
	body, error := ioutil.ReadAll(request.Body)
	if error != nil {
		log.WithFields(log.Fields{
			"method": "Server.createTransactionHandler",
		}).Error("unable to read HTTP body of request ", error)
		http.Error(responseWriter, "error reading HTTP body of request", http.StatusBadRequest)
		return
	}

	transaction, error := types.NewTransactionFromJson(body)
	if error != nil {
		log.WithFields(log.Fields{
			"method": "Server.createTransactionHandler",
		}).Error("JSON_PARSE_ERROR ", error) // TODO: Should return JSON!!!
		http.Error(responseWriter, "error reading HTTP body of request", http.StatusBadRequest)
		return
	}

	transaction, error = api.getService(&dapos.DAPoSService{}).(*dapos.DAPoSService).CreateTransaction(transaction, nil)
	if error != nil {
		log.WithFields(log.Fields{
			"method": "Server.createTransactionHandler",
		}).Error("JSON_PARSE_ERROR ", error) // TODO: Should return JSON!!!
		http.Error(responseWriter, "error reading HTTP body of request", http.StatusBadRequest)
		return
	}

	http.Error(responseWriter, "foobar", http.StatusOK)
	log.Info("create transaction")
}

// retrieveTransactionHandler
func (api *Api) retrieveTransactionHandler(responseWriter http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	responseWriter.WriteHeader(http.StatusOK)
	log.Info("retrieve transactions [address=" + vars["wallet_address"] + "]")
}

// getService
func (api *Api) getService(serviceInterface interface{}) types.IService {
	for _, service := range api.services {
		if reflect.TypeOf(service) == reflect.TypeOf(serviceInterface) {
			return service
		}
	}
	return nil
}



