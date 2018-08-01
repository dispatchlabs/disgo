package sdk

import (
	"github.com/dispatchlabs/disgo/commons/types"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"github.com/pkg/errors"

	"github.com/dispatchlabs/disgo/commons/utils"
	"time"
	"bytes"
	"fmt"
)

// GetDelegates
func GetDelegates() ([]types.Node, error) {

	// Get delegates.
	response, err := http.Get("http://seed.dispatchlabs.io:1975/v1/delegates")
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// Read body.
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal to RawMessage.
	var jsonMap map[string]json.RawMessage
	err = json.Unmarshal(body, &jsonMap)
	if err != nil {
		return nil, err
	}

	// Data?
	if jsonMap["data"] == nil {
		return nil, errors.Errorf("'data' is missing from response")
	}

	// Unmarshal nodes.
	var nodes []types.Node
	err = json.Unmarshal(jsonMap["data"], &nodes)
	if err != nil {
		return nil, err
	}

	return nodes, nil
}

// GetAccount
func GetAccount(delegateNode types.Node, address string) (*types.Account, error) {

	// Get account
	response, err := http.Get(fmt.Sprintf("http://%s:%d/v1/accounts/%s", delegateNode.HttpEndpoint.Host, delegateNode.HttpEndpoint.Port, address))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// Read body.
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal to RawMessage.
	var jsonMap map[string]json.RawMessage
	err = json.Unmarshal(body, &jsonMap)
	if err != nil {
		return nil, err
	}

	// Data?
	if jsonMap["data"] == nil {
		return nil, errors.Errorf("'data' is missing from response")
	}

	// Unmarshal account.
	var account types.Account
	err = json.Unmarshal(jsonMap["data"], &account)
	if err != nil {
		return nil, err
	}

	return &account, nil
}

// TransferTokens
func TransferTokens(delegateNode types.Node, privateKey string, from string, to string, tokens int64) (*types.Receipt, error) {

	// Create transfer tokens transaction.
	transaction, err := types.NewTransferTokensTransaction(privateKey, from, to, tokens, 0, utils.ToMilliSeconds(time.Now()))
	if err != nil {
		return nil, err
	}

	// Post transaction.
	response, err := http.Post(fmt.Sprintf("http://%s:%d/v1/transactions", delegateNode.HttpEndpoint.Host, delegateNode.HttpEndpoint.Port), "application/json", bytes.NewBuffer([]byte(transaction.String())))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// Read body.
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal receipt.
	var receipt *types.Receipt
	err = json.Unmarshal(body, &receipt)
	if err != nil {
		return nil, err
	}

	return receipt, nil
}

// DeploySmartContract
func DeploySmartContract(delegateNode types.Node, privateKey string, from string, code string, abi string) (*types.Receipt, error) {

	// Create deploy smart contract transaction.
	transaction, err := types.NewDeployContractTransaction(privateKey, from, code, abi, utils.ToMilliSeconds(time.Now()))
	if err != nil {
		return nil, err
	}

	// Post transaction.
	response, err := http.Post(fmt.Sprintf("http://%s:%d/v1/transactions", delegateNode.HttpEndpoint.Host, delegateNode.HttpEndpoint.Port), "application/json", bytes.NewBuffer([]byte(transaction.String())))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// Ready body.
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal receipt.
	var receipt *types.Receipt
	err = json.Unmarshal(body, &receipt)
	if err != nil {
		return nil, err
	}

	return receipt, nil
}

// ExecuteSmartContractTransaction
func ExecuteSmartContractTransaction(delegateNode types.Node, privateKey string, from string, to string, abi string, method string, params []interface{}) (*types.Receipt, error) {

	// Create execute smart contract transaction.
	transaction, err := types.NewExecuteContractTransaction(privateKey, from, to, abi, method, params, utils.ToMilliSeconds(time.Now()))
	if err != nil {
		return nil, err
	}

	// pos transaction.
	response, err := http.Post(fmt.Sprintf("http://%s:%d/v1/transactions", delegateNode.HttpEndpoint.Host, delegateNode.HttpEndpoint.Port), "application/json", bytes.NewBuffer([]byte(transaction.String())))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// Read body.
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal receipt.
	var receipt *types.Receipt
	err = json.Unmarshal(body, &receipt)
	if err != nil {
		return nil, err
	}

	return receipt, nil
}

// GetReceipt
func GetReceipt(delegateNode types.Node, hash string) (*types.Receipt, error) {

	// Get status.
	response, err := http.Get(fmt.Sprintf("http://%s:%d/v1/receipts/%s", delegateNode.HttpEndpoint.Host, delegateNode.HttpEndpoint.Port, hash))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// Ready body.
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal receipt.
	var receipt *types.Receipt
	err = json.Unmarshal(body, &receipt)
	if err != nil {
		return nil, err
	}

	return receipt, nil
}

// GetTransactionsSent
func GetTransactionsSent(delegateNode types.Node, address string) ([]types.Transaction, error) {

	// Get sent transaction.
	response, err := http.Get(fmt.Sprintf("http://%s:%d/v1/transactions/from/%s", delegateNode.HttpEndpoint.Host, delegateNode.HttpEndpoint.Port, address))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// Read body.
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal to RawMessage.
	var jsonMap map[string]json.RawMessage
	err = json.Unmarshal(body, &jsonMap)
	if err != nil {
		return nil, err
	}

	// Data?
	if jsonMap["data"] == nil {
		return nil, errors.Errorf("'data' is missing from response")
	}

	// Unmarshal transactions.
	var transactions []types.Transaction
	err = json.Unmarshal(jsonMap["data"], &transactions)
	if err != nil {
		return nil, err
	}

	return transactions, nil
}

// GetTransactionsReceived
func GetTransactionsReceived(delegateNode types.Node, address string) ([]types.Transaction, error) {

	// Get received transactions.
	response, err := http.Get(fmt.Sprintf("http://%s:%d/v1/transactions/to/%s", delegateNode.HttpEndpoint.Host, delegateNode.HttpEndpoint.Port, address))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// Read body.
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal to RawMessage.
	var jsonMap map[string]json.RawMessage
	err = json.Unmarshal(body, &jsonMap)
	if err != nil {
		return nil, err
	}

	// Data?
	if jsonMap["data"] == nil {
		return nil, errors.Errorf("'data' is missing from response")
	}

	// Unmarshal transactions.
	var transactions []types.Transaction
	err = json.Unmarshal(jsonMap["data"], &transactions)
	if err != nil {
		return nil, err
	}

	return transactions, nil
}
