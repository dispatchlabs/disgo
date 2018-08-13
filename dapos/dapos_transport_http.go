/*
 *    This file is part of DAPoS library.
 *
 *    The DAPoS library is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *
 *    The DAPoS library is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *
 *    You should have received a copy of the GNU General Public License
 *    along with the DAPoS library.  If not, see <http://www.gnu.org/licenses/>.
 */
package dapos

import (
	"io/ioutil"
	"net/http"

	"fmt"

	"github.com/dispatchlabs/disgo/commons/services"
	"github.com/dispatchlabs/disgo/commons/types"
	"github.com/dispatchlabs/disgo/commons/utils"
	"github.com/gorilla/mux"
)

// WithHttp -
func (this *DAPoSService) WithHttp() *DAPoSService {
	//Accounts
	services.GetHttpRouter().HandleFunc("/v1/accounts/{address}", this.getAccountHandler).Methods("GET")
	services.GetHttpRouter().HandleFunc("/v1/accounts", this.getAccountsHandler).Methods("GET")
	//Transactions
	services.GetHttpRouter().HandleFunc("/v1/transactions/from/{address}", this.getTransactionsByFromAddressHandler).Methods("GET")
	services.GetHttpRouter().HandleFunc("/v1/transactions/to/{address}", this.getTransactionsByToAddressHandler).Methods("GET")
	services.GetHttpRouter().HandleFunc("/v1/transactions", this.newTransactionHandler).Methods("POST")
	services.GetHttpRouter().HandleFunc("/v1/transactions/{hash}", this.getTransactionHandler).Methods("GET") //TODO: support pagination
	services.GetHttpRouter().HandleFunc("/v1/transactions", this.getTransactionsHandler).Methods("GET") //TODO: to be deprecated
	//Artifacts
	services.GetHttpRouter().HandleFunc("/v1/artifacts/{query}", this.unsupportedFunctionHandler).Methods("GET")//TODO: support pagination
	services.GetHttpRouter().HandleFunc("/v1/artifacts/", this.unsupportedFunctionHandler).Methods("POST")
	services.GetHttpRouter().HandleFunc("/v1/artifacts/{hash}", this.unsupportedFunctionHandler).Methods("GET")
	//delegates
	services.GetHttpRouter().HandleFunc("/v1/delegates", this.getDelegatesHandler).Methods("GET")
	services.GetHttpRouter().HandleFunc("/v1/delegates/subscribe", this.unsupportedFunctionHandler).Methods("POST")
	services.GetHttpRouter().HandleFunc("/v1/delegates/unsubscribe", this.unsupportedFunctionHandler).Methods("POST")

	//Page
	services.GetHttpRouter().HandleFunc("/v1/page", this.unsupportedFunctionHandler).Methods("GET") //TODO:only return hashes
	services.GetHttpRouter().HandleFunc("/v1/page/{id}", this.unsupportedFunctionHandler).Methods("GET")
	//analytical
	services.GetHttpRouter().HandleFunc("/v1/queue", this.getQueueHandler).Methods("GET")
	services.GetHttpRouter().HandleFunc("/v1/gossips", this.getGossipsHandler).Methods("GET")
	services.GetHttpRouter().HandleFunc("/v1/gossips/{hash}", this.getGossipHandler).Methods("GET")

	services.GetHttpRouter().HandleFunc("/v1/receipts/{hash}", this.getReceiptHandler).Methods("GET")

	return this
}

// TODO: Is there more generally way todo this ?
func setHeaders(responseWriter *http.ResponseWriter) {
	(*responseWriter).Header().Set("content-type", "application/json")
}

// getDelegatesHandler
func (this *DAPoSService) getDelegatesHandler(responseWriter http.ResponseWriter, request *http.Request) {
	setHeaders(&responseWriter)
	responseWriter.Write([]byte(this.GetDelegateNodes().String()))
}

// getAccountHandler
func (this *DAPoSService) getAccountHandler(responseWriter http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	response := this.GetAccount(vars["address"])
	setHeaders(&responseWriter)
	responseWriter.Write([]byte(response.String()))
}

// getTransactionHandler
func (this *DAPoSService) getTransactionHandler(responseWriter http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	response := this.GetTransaction(vars["hash"])
	setHeaders(&responseWriter)
	responseWriter.Write([]byte(response.String()))
}

// newTransactionHandler
func (this *DAPoSService) newTransactionHandler(responseWriter http.ResponseWriter, request *http.Request) {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		utils.Error("unable to read HTTP body of request", err)
		services.Error(responseWriter, fmt.Sprintf(`{"status":"%s: %v"}`, types.StatusInternalError, err), http.StatusInternalServerError)
		return
	}

	// New transaction.
	transaction, err := types.ToTransactionFromJson(body)
	if err != nil {
		utils.Error("JSON parse error", err)
		services.Error(responseWriter, fmt.Sprintf(`{"status":"%s: %v"}`, types.StatusJsonParseError, err), http.StatusInternalServerError)
		return
	}
	response := this.NewTransaction(transaction)
	setHeaders(&responseWriter)
	responseWriter.Write([]byte(response.String()))
}

// getTransactionsByFromAddressHandler
func (this *DAPoSService) getTransactionsByFromAddressHandler(responseWriter http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	response := this.GetTransactionsByFromAddress(vars["address"])
	setHeaders(&responseWriter)
	responseWriter.Write([]byte(response.String()))
}

// getTransactionsByToAddressHandler
func (this *DAPoSService) getTransactionsByToAddressHandler(responseWriter http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	response := this.GetTransactionsByToAddress(vars["address"])
	setHeaders(&responseWriter)
	responseWriter.Write([]byte(response.String()))
}

func (this *DAPoSService) getTransactionsHandler(responseWriter http.ResponseWriter, request *http.Request) {
	pageNumber := request.URL.Query().Get("page")
	if pageNumber == ""{
		pageNumber = "1"
	}
	response := this.GetTransactions(pageNumber)
	setHeaders(&responseWriter)
	responseWriter.Write([]byte(response.String()))
}

// getQueueHandler
func (this *DAPoSService) getQueueHandler(responseWriter http.ResponseWriter, request *http.Request) {
	response := this.DumpQueue()
	setHeaders(&responseWriter)
	responseWriter.Write([]byte(response.String()))
}

// getArtifactHandler
func (this *DAPoSService) unsupportedFunctionHandler(responseWriter http.ResponseWriter, request *http.Request) {
	response := this.ToBeSupported()
	setHeaders(&responseWriter)
	responseWriter.Write([]byte(response.String()))
}

// getAccountsHandler
func (this *DAPoSService) getAccountsHandler(responseWriter http.ResponseWriter, request *http.Request) {
	pageNumber := request.URL.Query().Get("page")
	if pageNumber == ""{
		pageNumber = "1"
	}
	response := this.GetAccounts(pageNumber)
	setHeaders(&responseWriter)
	responseWriter.Write([]byte(response.String()))
}

// getGossipsHandler
func (this *DAPoSService) getGossipsHandler(responseWriter http.ResponseWriter, request *http.Request) {
	pageNumber := request.URL.Query().Get("page")
	if pageNumber == ""{
		pageNumber = "1"
	}
	response := this.GetGossips(pageNumber)
	setHeaders(&responseWriter)
	responseWriter.Write([]byte(response.String()))
}

// getTransactionHandler
func (this *DAPoSService) getGossipHandler(responseWriter http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	response := this.GetGossip(vars["hash"])
	setHeaders(&responseWriter)
	responseWriter.Write([]byte(response.String()))
}

// getReceiptHandler
func (this *DAPoSService) getReceiptHandler(responseWriter http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	response := this.GetReceipt(vars["hash"])
	setHeaders(&responseWriter)
	responseWriter.Write([]byte(response.String()))
}