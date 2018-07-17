package sdk

import (
	"github.com/dispatchlabs/disgo/commons/types"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/mitchellh/mapstructure"

	"github.com/dispatchlabs/disgo/commons/utils"
	"time"
	"bytes"
	"fmt"
)

// GetDelegates
func GetDelegates() ([]types.Node, error) {

	response, err := http.Get("http://seed.dispatchlabs.io:1975/v1/delegates")
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var jsonMap map[string]interface{}
	err = json.Unmarshal(body, &jsonMap)
	if err != nil {
		return nil, err
	}
	if jsonMap["data"] == nil {
		return nil, errors.Errorf("'data' is missing from response")
	}

	var nodes []types.Node
	err = mapstructure.Decode(jsonMap["data"], &nodes)
	if err != nil {
		return nil, err
	}

	return nodes, nil
}

// TransferTokens
func TransferTokens(delegateNode types.Node, privateKey string, from string, to string, tokens int64) (*types.Receipt, error) {

	transaction, err := types.NewTransferTokensTransaction(privateKey, from, to, tokens, 0, utils.ToMilliSeconds(time.Now()))
	if err != nil {
		return nil, err
	}

	response, err := http.Post(fmt.Sprintf("http://%s:1975/v1/transactions", delegateNode.Endpoint.Host), "application/json", bytes.NewBuffer([]byte(transaction.String())))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var receipt *types.Receipt
	err = json.Unmarshal(body, &receipt)
	if err != nil {
		return nil, err
	}

	return receipt, nil
}

// DeploySmartContract
func DeploySmartContract(delegateNode types.Node, privateKey string, from string, code string, abi string) (*types.Receipt, error) {

	transaction, err := types.NewDeployContractTransaction(privateKey, from, code, abi, utils.ToMilliSeconds(time.Now()))
	if err != nil {
		return nil, err
	}

	response, err := http.Post(fmt.Sprintf("http://%s:1975/v1/transactions", delegateNode.Endpoint.Host), "application/json", bytes.NewBuffer([]byte(transaction.String())))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var receipt *types.Receipt
	err = json.Unmarshal(body, &receipt)
	if err != nil {
		return nil, err
	}

	return receipt, nil
}

// ExecuteSmartContractTransaction
func ExecuteSmartContractTransaction(delegateNode types.Node, privateKey string, from string, to string, method string, params []interface{}) (*types.Receipt, error) {

	transaction, err := types.NewExecuteContractTransaction(privateKey, from, to, method, params, utils.ToMilliSeconds(time.Now()))
	if err != nil {
		return nil, err
	}

	response, err := http.Post(fmt.Sprintf("http://%s:1975/v1/transactions", delegateNode.Endpoint.Host), "application/json", bytes.NewBuffer([]byte(transaction.String())))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var receipt *types.Receipt
	err = json.Unmarshal(body, &receipt)
	if err != nil {
		return nil, err
	}

	return receipt, nil
}

// GetStatus
func GetStatus(delegateNode types.Node, id string) (*types.Receipt, error) {

	response, err := http.Get(fmt.Sprintf("http://%s:1975/v1/statuses/%s", delegateNode.Endpoint.Host, id))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var receipt *types.Receipt
	err = json.Unmarshal(body, &receipt)
	if err != nil {
		return nil, err
	}

	return receipt, nil
}

// GetTransactionsSent
func GetTransactionsSent(delegateNode types.Node, address string) ([]types.Transaction, error) {

	response, err := http.Get(fmt.Sprintf("http://%s:1975/v1/transactions/from/%s", delegateNode.Endpoint.Host, address))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var jsonMap map[string]interface{}
	err = json.Unmarshal(body, &jsonMap)
	if err != nil {
		return nil, err
	}
	if jsonMap["data"] == nil {
		return nil, errors.Errorf("'data' is missing from response")
	}

	var transactions []types.Transaction
	err = mapstructure.Decode(jsonMap["data"], &transactions)
	if err != nil {
		return nil, err
	}

	return transactions, nil
}

// GetTransactionsReceived
func GetTransactionsReceived(delegateNode types.Node, address string) ([]types.Transaction, error) {

	response, err := http.Get(fmt.Sprintf("http://%s:1975/v1/transactions/to/%s", delegateNode.Endpoint.Host, address))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var jsonMap map[string]interface{}
	err = json.Unmarshal(body, &jsonMap)
	if err != nil {
		return nil, err
	}
	if jsonMap["data"] == nil {
		return nil, errors.Errorf("'data' is missing from response")
	}

	var transactions []types.Transaction
	err = mapstructure.Decode(jsonMap["data"], &transactions)
	if err != nil {
		return nil, err
	}

	return transactions, nil
}
