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

	"github.com/dispatchlabs/disgo/commons/services"
	"github.com/dispatchlabs/disgo/commons/types"
	"github.com/dispatchlabs/disgo/commons/utils"
	"github.com/gorilla/mux"
	"fmt"
)

// WithHttp -
func (this *DAPoSService) WithHttp() *DAPoSService {
	services.GetHttpRouter().HandleFunc("/v1/delegates", this.getDelegatesHandler).Methods("GET")
	services.GetHttpRouter().HandleFunc("/v1/statuses/{id}", this.getStatusHandler).Methods("GET")
	services.GetHttpRouter().HandleFunc("/v1/accounts/{address}", this.getAccountHandler).Methods("GET")
	services.GetHttpRouter().HandleFunc("/v1/transactions", this.getTransactionsHandler).Methods("GET")
	services.GetHttpRouter().HandleFunc("/v1/transactions/from/{address}", this.getTransactionsByFromAddressHandler).Methods("GET")
	services.GetHttpRouter().HandleFunc("/v1/transactions/to/{address}", this.getTransactionsByToAddressHandler).Methods("GET")
	services.GetHttpRouter().HandleFunc("/v1/transactions", this.newTransactionHandler).Methods("POST")
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
func (this *DAPoSService) getStatusHandler(responseWriter http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	receipt := this.GetStatus(vars["id"])
	setHeaders(&responseWriter)
	responseWriter.Write([]byte(receipt.String()))
}

// getAccountHandler
func (this *DAPoSService) getAccountHandler(responseWriter http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	receipt := this.GetAccount(vars["address"])
	setHeaders(&responseWriter)
	responseWriter.Write([]byte(receipt.String()))
}

// setAccountHandler
func (this *DAPoSService) setAccountHandler(responseWriter http.ResponseWriter, request *http.Request) {
	// TODO: Call SetAccount.
	setHeaders(&responseWriter)
}

// newTransactionHandler
func (this *DAPoSService) newTransactionHandler(responseWriter http.ResponseWriter, request *http.Request) {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		utils.Error("unable to read HTTP body of request", err)
		http.Error(responseWriter, `{"status":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}

	// New transaction.
	transaction, err := types.ToTransactionFromJson(body)
	if err != nil {
		utils.Error("JSON parse error", err)
		http.Error(responseWriter, fmt.Sprintf(`{"status":"JSON_PARSE_ERROR: %v"}`, err), http.StatusInternalServerError)
		return
	}
	receipt := this.NewTransaction(transaction)
	setHeaders(&responseWriter)
	responseWriter.Write([]byte(receipt.String()))
}

// getTransactionsByFromAddressHandler
func (this *DAPoSService) getTransactionsByFromAddressHandler(responseWriter http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	receipt := this.GetTransactionsByFromAddress(vars["address"])
	setHeaders(&responseWriter)
	responseWriter.Write([]byte(receipt.String()))
}

// getTransactionsByToAddressHandler
func (this *DAPoSService) getTransactionsByToAddressHandler(responseWriter http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	receipt := this.GetTransactionsByToAddress(vars["address"])
	setHeaders(&responseWriter)
	responseWriter.Write([]byte(receipt.String()))
}

// getTransactionsByFromAddressHandler
func (this *DAPoSService) getTransactionsHandler(responseWriter http.ResponseWriter, request *http.Request) {
	receipt := this.GetTransactions()
	setHeaders(&responseWriter)
	responseWriter.Write([]byte(receipt.String()))
}
