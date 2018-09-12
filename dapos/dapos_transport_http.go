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

	"encoding/hex"

	"github.com/dispatchlabs/disgo/commons/helper"
	"github.com/dispatchlabs/disgo/commons/services"
	"github.com/dispatchlabs/disgo/commons/types"
	"github.com/dispatchlabs/disgo/commons/utils"
	"github.com/gorilla/mux"
)

// WithHttp -
func (this *DAPoSService) WithHttp() *DAPoSService {
	//Accounts
	services.GetHttpRouter().HandleFunc("/v1/accounts/{address}", this.getAccountHandler).Methods("GET")
	services.GetHttpRouter().HandleFunc("/v1/accounts", this.unsupportedFunctionHandler).Methods("GET")
	//Transactions
	services.GetHttpRouter().HandleFunc("/v1/transactions", this.newTransactionHandler).Methods("POST")
	services.GetHttpRouter().HandleFunc("/v1/transactions/{hash}", this.getTransactionHandler).Methods("GET")
	services.GetHttpRouter().HandleFunc("/v1/transactions", this.getTransactionsHandler).Methods("GET")
	//Artifacts
	services.GetHttpRouter().HandleFunc("/v1/artifacts/{query}", this.unsupportedFunctionHandler).Methods("GET") //TODO: support pagination
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

	services.GetHttpRouter().HandleFunc("/v1/receipts/{hash}", this.unsupportedFunctionHandler).Methods("GET")

	return this
}

// TODO: Is there more generally way todo this ?
func setHeaders(response *types.Response, responseWriter *http.ResponseWriter) {
	(*responseWriter).Header().Set("content-type", "application/json")

	// Adjust the HTTP status reply - taken from `disgo/commons/types/constants.go`
	// StatusPending                      = "Pending"
	// StatusOk                           = "Ok"
	// StatusNotFound                     = "NotFound"
	// StatusReceiptNotFound              = "StatusReceiptNotFound"
	// StatusTransactionTimeOut           = "StatusTransactionTimeOut"
	// StatusInvalidTransaction           = "InvalidTransaction"
	// StatusInsufficientTokens           = "InsufficientTokens"
	// StatusDuplicateTransaction         = "DuplicateTransaction"
	// StatusNotDelegate                  = "StatusNotDelegate"
	// StatusAlreadyProcessingTransaction = "StatusAlreadyProcessingTransaction"
	// StatusGossipingTimedOut            = "StatusGossipingTimedOut"
	// StatusJsonParseError               = "StatusJsonParseError"
	// StatusInternalError                = "InternalError"
	// StatusUnavailableFeature           = "UnavailableFeature"

	if response != nil {
		if response.Status == types.StatusOk {
			(*responseWriter).WriteHeader(http.StatusOK)
		} else if response.Status == types.StatusNotFound {
			(*responseWriter).WriteHeader(http.StatusNotFound)
		} else if response.Status == types.StatusPending {
			(*responseWriter).WriteHeader(http.StatusOK)
		} else if response.Status == types.StatusInternalError {
			(*responseWriter).WriteHeader(http.StatusInternalServerError)
		} else if response.Status == types.StatusNotDelegate {
			(*responseWriter).WriteHeader(http.StatusTeapot)
		} else {
			(*responseWriter).WriteHeader(http.StatusBadRequest)
		}
	}
}

// getDelegatesHandler
func (this *DAPoSService) getDelegatesHandler(responseWriter http.ResponseWriter, request *http.Request) {
	setHeaders(nil, &responseWriter)
	responseWriter.Write([]byte(this.GetDelegateNodes().String()))
}

// getAccountHandler
func (this *DAPoSService) getAccountHandler(responseWriter http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	response := this.GetAccount(vars["address"])
	setHeaders(response, &responseWriter)
	responseWriter.Write([]byte(response.String()))
}

// getTransactionHandler
func (this *DAPoSService) getTransactionHandler(responseWriter http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	response := this.GetTransaction(vars["hash"])
	setHeaders(response, &responseWriter)
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
	txn := services.NewTxn(true)
	defer txn.Discard()

	if err != nil {
		if err != nil {
			utils.Error("Paramater type error", err)
			services.Error(responseWriter, fmt.Sprintf(`{"status":"%s: %v"}`, types.StatusJsonParseError, err), http.StatusBadRequest)
			return
		}
	}
	// DUPLICATE by mistake ??
	// txn := services.NewTxn(true)
	// defer txn.Discard()

	if transaction.Type == types.TypeExecuteSmartContract {
		contractTx, err := types.ToTransactionByAddress(txn, transaction.To)

		transaction.Abi = hex.EncodeToString([]byte(contractTx.Abi))
		if err != nil {
			utils.Error(err)
		}
		transaction.Params, err = helper.GetConvertedParams(transaction)
		if err != nil {
			utils.Error("Paramater type error", err)
			services.Error(responseWriter, fmt.Sprintf(`{"status":"%s: %v"}`, types.StatusJsonParseError, err), http.StatusBadRequest)
			return
		}

	}
	response := this.NewTransaction(transaction)
	setHeaders(response, &responseWriter)
	responseWriter.Write([]byte(response.String()))
}

func (this *DAPoSService) getTransactionsHandler(responseWriter http.ResponseWriter, request *http.Request) {
	response := types.NewResponse()
	pageNumber := request.URL.Query().Get("page")
	if pageNumber == "" {
		pageNumber = "1"
	}
	pageLimit := request.URL.Query().Get("pageSize")
	if pageLimit == "" {
		pageLimit = "10"
	}
	startingHash := request.URL.Query().Get("pageStart")
	from := request.URL.Query().Get("from")
	to := request.URL.Query().Get("to")
	if from != "" && to != "" {
		response.Status = http.StatusText(http.StatusBadRequest)
		response.HumanReadableStatus = "\"from\" and \"to\" parameters may not both be provided"
		services.Error(responseWriter, response.String(), http.StatusBadRequest)
		return
	} else if from != "" {
		response = this.GetTransactionsByFromAddress(from, pageNumber, pageLimit, startingHash)
	} else if to != "" {
		response = this.GetTransactionsByToAddress(to, pageNumber, pageLimit, startingHash)
	} else {
		response = this.GetTransactions(pageNumber, pageLimit, startingHash)
	}
	setHeaders(response, &responseWriter)
	responseWriter.Write([]byte(response.String()))
}

// getQueueHandler
func (this *DAPoSService) getQueueHandler(responseWriter http.ResponseWriter, request *http.Request) {
	response := this.DumpQueue()
	setHeaders(response, &responseWriter)
	responseWriter.Write([]byte(response.String()))
}

// getArtifactHandler
func (this *DAPoSService) unsupportedFunctionHandler(responseWriter http.ResponseWriter, request *http.Request) {
	response := this.ToBeSupported()
	setHeaders(response, &responseWriter)
	responseWriter.Write([]byte(response.String()))
}

// getAccountsHandler
// func (this *DAPoSService) getAccountsHandler(responseWriter http.ResponseWriter, request *http.Request) {
// 	pageNumber := request.URL.Query().Get("page")
// 	if pageNumber == "" {
// 		pageNumber = "1"
// 	}
// 	pageLimit := request.URL.Query().Get("pageSize")
// 	if pageLimit == "" {
// 		pageLimit = "10"
// 	}
// 	startingAddress := request.URL.Query().Get("pageStart")
// 	response := this.GetAccounts(pageNumber, pageLimit, startingAddress)
// 	setHeaders(response, &responseWriter)
// 	responseWriter.Write([]byte(response.String()))
// }

// getGossipsHandler
func (this *DAPoSService) getGossipsHandler(responseWriter http.ResponseWriter, request *http.Request) {
	pageNumber := request.URL.Query().Get("page")
	if pageNumber == "" {
		pageNumber = "1"
	}
	response := this.GetGossips(pageNumber)
	setHeaders(response, &responseWriter)
	responseWriter.Write([]byte(response.String()))
}

// getTransactionHandler
func (this *DAPoSService) getGossipHandler(responseWriter http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	response := this.GetGossip(vars["hash"])
	setHeaders(response, &responseWriter)
	responseWriter.Write([]byte(response.String()))
}

// getReceiptHandler
// func (this *DAPoSService) getReceiptHandler(responseWriter http.ResponseWriter, request *http.Request) {
// 	vars := mux.Vars(request)
// 	response := this.GetReceipt(vars["hash"])
// 	setHeaders(response, &responseWriter)
// 	responseWriter.Write([]byte(response.String()))
// }
