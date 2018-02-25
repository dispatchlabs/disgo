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
	"github.com/dispatchlabs/disgo_commons/crypto"
	"time"
	"encoding/hex"
)

// Api
type Api struct {
	services []types.IService
	router   *mux.Router
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

	// TODO: Remove (just for flushing out API). MAO!
	walletAccount := types.NewWalletAccount()
	walletAccount.Balance = 100

	// Write response.
	response, error := json.Marshal(struct {
		Status string `json:"status,omitempty"`
		WalletAccount *types.WalletAccount `json:"walletAccount,omitempty"`
	}{
		Status: "OK",
		WalletAccount: walletAccount,
	})
	if error != nil {
		log.WithFields(log.Fields{
			"method": "Server.createWalletHandler",
		}).Error("unable to create response JSON [error=", error.Error() + "]")
		http.Error(responseWriter, `{"status":"INTERNAL_SERVER_ERROR"}`, http.StatusInternalServerError)
		return
	}
	responseWriter.Write(response)
	log.WithFields(log.Fields{
		"method": "Server.createWalletHandler",
	}).Info(string(response))
}

// retrieveWalletHandler
func (this *Api) retrieveWalletHandler(responseWriter http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)

	// TODO: Remove (just for flushing out API). MAO!
	walletAccount := types.NewWalletAccount()
	address, _  := hex.DecodeString(vars["wallet_address"])
	copy (walletAccount.Address[:], address)

	// Write response.
	response, error := json.Marshal(struct {
		Status string `json:"status,omitempty"`
		WalletAccount *types.WalletAccount `json:"walletAccount,omitempty"`
	}{
		Status: "OK",
		WalletAccount: walletAccount,
	})
	if error != nil {
		log.WithFields(log.Fields{
			"method": "Server.retrieveWalletHandler",
		}).Error("unable to create response JSON [error=", error.Error() + "]")
		http.Error(responseWriter, `{"status":"INTERNAL_SERVER_ERROR"}`, http.StatusInternalServerError)
		return
	}
	responseWriter.Write(response)
	log.WithFields(log.Fields{
		"method": "Server.retrieveWalletHandler",
	}).Info(string(response))
}

// createTransactionHandler
func (this *Api) createTransactionHandler(responseWriter http.ResponseWriter, request *http.Request) {
	body, error := ioutil.ReadAll(request.Body)
	if error != nil {
		log.WithFields(log.Fields{
			"method": "Server.createTransactionHandler",
		}).Error("unable to read HTTP body of request ", error)
		http.Error(responseWriter, `{"status":"INTERNAL_SERVER_ERROR"}`, http.StatusInternalServerError)
		return
	}

	transaction := &types.Transaction{}
	error = json.Unmarshal(body, transaction)
	if error != nil {
		log.WithFields(log.Fields{
			"method": "Server.createTransactionHandler",
		}).Error("JSON_PARSE_ERROR", error)
		http.Error(responseWriter, `{"status":"JSON_PARSE_ERROR"}`, http.StatusBadRequest)
		return
	}

	// TODO: Remove (just for flushing out API). MAO!
	transaction.Hash = crypto.NewHash()
	transaction.Time = time.Now()

	// Create transaction.
	_, error = this.getService(&dapos.DAPoSService{}).(*dapos.DAPoSService).CreateTransaction(transaction, nil)
	if error != nil {
		log.WithFields(log.Fields{
			"method": "Server.createTransactionHandler",
		}).Error("JSON_PARSE_ERROR [error=", error.Error()+"]")
		http.Error(responseWriter, "error reading HTTP body of request", http.StatusBadRequest)
		return
	}

	// Write response.
	response, error := json.Marshal(struct {
		Status string `json:"status,omitempty"`
		Transaction *types.Transaction `json:"transaction,omitempty"`
	}{
		Status: "OK",
		Transaction: transaction,
	})
	if error != nil {
		log.WithFields(log.Fields{
			"method": "Server.createTransactionHandler",
		}).Error("unable to create response JSON [error=", error.Error() + "]")
		http.Error(responseWriter, `{"status":"INTERNAL_SERVER_ERROR"}`, http.StatusInternalServerError)
		return
	}
	responseWriter.Write(response)
	log.WithFields(log.Fields{
		"method": "Server.createTransactionHandler",
	}).Info(string(response))
}

// retrieveTransactionHandler
func (this *Api) retrieveTransactionHandler(responseWriter http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)

	// TODO: Remove (just for flushing out API). MAO!
	var transactions [2] *types.Transaction
	transaction := types.NewTransaction()
	transaction.Value = 2
	from, _  := hex.DecodeString(vars["wallet_address"])
	copy (transaction.From[:], from)
	to, _  := hex.DecodeString("cc3f682246d4a755833f9cb19e1acc8565a0c2ba")
	copy (transaction.To[:], to)
	transactions[0] = transaction
	transaction = types.NewTransaction()
	transaction.Value = 4
	from, _  = hex.DecodeString(vars["wallet_address"])
	copy (transaction.From[:], from)
	to, _  = hex.DecodeString("cc3f682246d4a755833f9cb19e1acc8565a0c2ba")
	copy (transaction.To[:], to)
	transactions[1] = transaction

	// Write response.
	response, error := json.Marshal(struct {
		Status string `json:"status,omitempty"`
		Transactions [2] *types.Transaction `json:"transactions,omitempty"`
	}{
		Status: "OK",
		Transactions: transactions,
	})
	if error != nil {
		log.WithFields(log.Fields{
			"method": "Server.retrieveTransactionHandler",
		}).Error("unable to create response JSON [error=", error.Error() + "]")
		http.Error(responseWriter, `{"status":"INTERNAL_SERVER_ERROR"}`, http.StatusInternalServerError)
		return
	}
	responseWriter.Write(response)
	log.WithFields(log.Fields{
		"method": "Server.retrieveTransactionHandler",
	}).Info(string(response))
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
