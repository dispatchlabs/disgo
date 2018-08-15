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
)

// WithHttp -
func (this *LocalAPIService) WithHttp() *LocalAPIService {
	services.GetHttpRouter().HandleFunc("/v1/localapi/transfer", this.tranferHandler).Methods("POST")
	services.GetHttpRouter().HandleFunc("/v1/localapi/deploy", this.deployHandler).Methods("POST")
	services.GetHttpRouter().HandleFunc("/v1/localapi/execute", this.executeHandler).Methods("POST")

	return this
}

// TODO: Is there more generally way todo this ?
func setHeaders(responseWriter *http.ResponseWriter) {
	(*responseWriter).Header().Set("content-type", "application/json")
}

func (this *LocalAPIService) tranferHandler(responseWriter http.ResponseWriter, request *http.Request) {
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
		types.GetAccount().PrivateKey,
		transfer.From,
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
		types.GetAccount().PrivateKey,
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
		types.GetAccount().PrivateKey,
		disgover.GetDisGoverService().ThisNode.Address,
		execute.ContractAddress,
		hex.EncodeToString([]byte(execute.Abi)),
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
