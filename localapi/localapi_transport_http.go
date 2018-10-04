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
package localapi

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/dispatchlabs/disgo/dapos"
	"github.com/dispatchlabs/disgo/disgover"

	"fmt"

	"github.com/dispatchlabs/disgo/commons/services"
	"github.com/dispatchlabs/disgo/commons/types"
	"github.com/dispatchlabs/disgo/commons/utils"
	"github.com/dispatchlabs/disgo/sdk"
	"time"
	"strings"
	"encoding/base64"
	"math/big"
	"github.com/dispatchlabs/disgo/commons/crypto"
)

// WithHttp -
func (this *LocalAPIService) WithHttp() *LocalAPIService {
	services.GetHttpRouter().HandleFunc("/v1/local/transfer", this.tranferHandler).Methods("POST")
	services.GetHttpRouter().HandleFunc("/v1/local/deploy", this.deployHandler).Methods("POST")
	services.GetHttpRouter().HandleFunc("/v1/local/execute", this.executeHandler).Methods("POST")
	services.GetHttpRouter().HandleFunc("/v1/local/packageTx", this.getPackageTxHandler).Methods("POST")
	services.GetHttpRouter().HandleFunc("/v1/local/getAccount", this.getAccountHandler).Methods("GET")
	services.GetHttpRouter().HandleFunc("/v1/local/getNewAccount", this.createAccountHandler).Methods("GET")

	return this
}

// TODO: Is there more generally way todo this ?
func setHeaders(responseWriter *http.ResponseWriter) {
	(*responseWriter).Header().Set("content-type", "application/json")
}


func checkAuth(w http.ResponseWriter, r *http.Request) bool {
	s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(s) != 2 { return false }

	b, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil { return false }

	pair := strings.SplitN(string(b), ":", 2)
	if len(pair) != 2 { return false }

	return pair[0] == "Disgo" && pair[1] == "Dance"
}


func (this *LocalAPIService) tranferHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if !checkAuth(responseWriter, request) {
		responseWriter.Header().Set("WWW-Authenticate", `realm="Dispatch Local"`)
		responseWriter.WriteHeader(401)
		responseWriter.Write([]byte("401 Unauthorized\n"))
		return
	}
	// Read Object from payload
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		utils.Error("unable to read HTTP body of request", err)
		services.Error(responseWriter, fmt.Sprintf(`{"status":"%s: %v"}`, types.StatusInternalError, err), http.StatusInternalServerError)
		return
	}

	transfer := &Transfer{}
	err = json.Unmarshal(body, transfer)
	if err != nil {
		utils.Error("unable to read HTTP body of request", err)
		services.Error(responseWriter, fmt.Sprintf(`{"status":"%s: %v"}`, types.StatusInternalError, err), http.StatusInternalServerError)
		return
	}

	// Invoke SDK
	var delegates = dapos.GetDAPoSService().GetDelegateNodes().Data.([]*types.Node)
	if len(delegates) <= 0 {
		utils.Error("no delegates found")
		services.Error(responseWriter, fmt.Sprintf(`{"status":"no delegates found"}`), http.StatusInternalServerError)
		return
	}

	response, err := sdk.TransferTokens(
		*delegates[0],
		types.GetKey(),
		types.GetAccount().Address,
		transfer.To,
		transfer.Amount,
	)

	// Send Reply
	if err != nil {
		utils.Error("error executing Local API", err)
		services.Error(responseWriter, fmt.Sprintf(`{"status":"%s: %v"}`, types.StatusInternalError, err), http.StatusInternalServerError)
		return
	}

	setHeaders(&responseWriter)
	responseWriter.Write([]byte(response))
}

func (this *LocalAPIService) deployHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if !checkAuth(responseWriter, request) {
		responseWriter.Header().Set("WWW-Authenticate", `realm="Dispatch Local"`)
		responseWriter.WriteHeader(401)
		responseWriter.Write([]byte("401 Unauthorized\n"))
		return
	}
	// Read Object from payload
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		utils.Error("unable to read HTTP body of request", err)
		services.Error(responseWriter, fmt.Sprintf(`{"status":"%s: %v"}`, types.StatusInternalError, err), http.StatusInternalServerError)
		return
	}

	deploy := &Deploy{}
	err = json.Unmarshal(body, deploy)
	if err != nil {
		utils.Error("unable to read HTTP body of request", err)
		services.Error(responseWriter, fmt.Sprintf(`{"status":"%s: %v"}`, types.StatusInternalError, err), http.StatusInternalServerError)
		return
	}

	// Invoke SDK
	var delegates = dapos.GetDAPoSService().GetDelegateNodes().Data.([]*types.Node)
	if len(delegates) <= 0 {
		utils.Error("no delegates found")
		services.Error(responseWriter, fmt.Sprintf(`{"status":"no delegates found"}`), http.StatusInternalServerError)
		return
	}

	response, err := sdk.DeploySmartContract(
		*delegates[0],
		types.GetKey(),
		disgover.GetDisGoverService().ThisNode.Address,
		deploy.ByteCode,
		hex.EncodeToString([]byte(deploy.Abi)),
	)

	// Send Reply
	if err != nil {
		utils.Error("error executing Local API", err)
		services.Error(responseWriter, fmt.Sprintf(`{"status":"%s: %v"}`, types.StatusInternalError, err), http.StatusInternalServerError)
		return
	}

	setHeaders(&responseWriter)
	responseWriter.Write([]byte(response))
}

func (this *LocalAPIService) executeHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if !checkAuth(responseWriter, request) {
		responseWriter.Header().Set("WWW-Authenticate", `realm="Dispatch Local"`)
		responseWriter.WriteHeader(401)
		responseWriter.Write([]byte("401 Unauthorized\n"))
		return
	}
	// Read Object from payload
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		utils.Error("unable to read HTTP body of request", err)
		services.Error(responseWriter, fmt.Sprintf(`{"status":"%s: %v"}`, types.StatusInternalError, err), http.StatusInternalServerError)
		return
	}

	execute := &Execute{}
	err = json.Unmarshal(body, execute)
	if err != nil {
		utils.Error("unable to read HTTP body of request", err)
		services.Error(responseWriter, fmt.Sprintf(`{"status":"%s: %v"}`, types.StatusInternalError, err), http.StatusInternalServerError)
		return
	}

	// Invoke SDK
	var delegates = dapos.GetDAPoSService().GetDelegateNodes().Data.([]*types.Node)
	if len(delegates) <= 0 {
		utils.Error("no delegates found")
		services.Error(responseWriter, fmt.Sprintf(`{"status":"no delegates found"}`), http.StatusInternalServerError)
		return
	}

	response, err := sdk.ExecuteSmartContractTransaction(
		*delegates[0],
		types.GetKey(),
		disgover.GetDisGoverService().ThisNode.Address,
		execute.ContractAddress,
		execute.Method,
		execute.Params,
	)

	// Send Reply
	if err != nil {
		utils.Error("error executing Local API", err)
		services.Error(responseWriter, fmt.Sprintf(`{"status":"%s: %v"}`, types.StatusInternalError, err), http.StatusInternalServerError)
		return
	}

	setHeaders(&responseWriter)
	responseWriter.Write([]byte(response))
}

// getDelegatesHandler
func (this *LocalAPIService) getPackageTxHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if !checkAuth(responseWriter, request) {
		responseWriter.Header().Set("WWW-Authenticate", `realm="Dispatch Local"`)
		responseWriter.WriteHeader(401)
		responseWriter.Write([]byte("401 Unauthorized\n"))
		return
	}
	response := types.NewResponse()
	// Read Object from payload
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		utils.Error("unable to read HTTP body of request", err)
		services.Error(responseWriter, fmt.Sprintf(`{"status":"%s: %v"}`, types.StatusInternalError, err), http.StatusInternalServerError)
		return
	}

	pack := &Package{}
	err = json.Unmarshal(body, pack)
	if err != nil {
		utils.Error("unable to read HTTP body of request", err)
		services.Error(responseWriter, fmt.Sprintf(`{"status":"%s: %v"}`, types.StatusInternalError, err), http.StatusInternalServerError)
		return
	}

	tx, err := sdk.PackageTx(pack.To, pack.Amount, pack.Time)
	if err != nil {
		response.Status = types.StatusInternalError
	} else {
		response.Data = tx
		response.Status = types.StatusOk
	}
	setHeaders(&responseWriter)
	responseWriter.Write([]byte(response.String()))
}

// getDelegatesHandler
func (this *LocalAPIService) getAccountHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if !checkAuth(responseWriter, request) {
		responseWriter.Header().Set("WWW-Authenticate", `realm="Dispatch Local"`)
		responseWriter.WriteHeader(401)
		responseWriter.Write([]byte("401 Unauthorized\n"))
		return
	}
	txn := services.NewTxn(true)
	defer txn.Discard()

	response  :=  types.GetAccount()

	setHeaders(&responseWriter)
	responseWriter.Write([]byte(response.String()))
}

// getDelegatesHandler
func (this *LocalAPIService) createAccountHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if !checkAuth(responseWriter, request) {
		responseWriter.Header().Set("WWW-Authenticate", `realm="Dispatch Local"`)
		responseWriter.WriteHeader(401)
		responseWriter.Write([]byte("401 Unauthorized\n"))
		return
	}
	response := types.NewResponse()

	publicKey, privateKey := crypto.GenerateKeyPair()
	address := crypto.ToAddress(publicKey)
	account := &types.Account{}
	account.Address = hex.EncodeToString(address)
	account.PrivateKey = hex.EncodeToString(privateKey)
	account.Balance = big.NewInt(0)
	account.Name = ""
	now := time.Now()
	account.Created = now
	account.Updated = now

	response.Data  = account

	setHeaders(&responseWriter)
	responseWriter.Write([]byte(response.String()))
}