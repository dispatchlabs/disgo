package sdk

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"math/big"
	"net/http"

	"github.com/dispatchlabs/disgo/commons/types"
	"github.com/pkg/errors"

	"bytes"
	"fmt"
	"time"

	"github.com/dispatchlabs/disgo/commons/utils"
	"github.com/dispatchlabs/disgo/commons/crypto"
)

// GetDelegates - Get the known delegates at this point in time
func GetDelegates(seedUrl_optional ...string) ([]types.Node, error) {
	seedUrl := "seed.dispatchlabs.io:1975"
	if len(seedUrl_optional) > 0 {
		seedUrl = seedUrl_optional[0]
	}

	// Get delegates.
	httpResponse, err := http.Get(fmt.Sprintf("http://%s/v1/delegates", seedUrl))
	if err != nil {
		return nil, err
	}
	defer httpResponse.Body.Close()

	// Read body.
	body, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal response.
	var response *types.Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	// Status?
	if response.Status != types.StatusOk {
		return nil, errors.New(fmt.Sprintf("%s: %s", response.Status, response.HumanReadableStatus))
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

// CreateAccount - Generate a new account
func CreateAccount(name_optional ...string) (*types.Account, error) {
	name := ""
	if len(name_optional) > 0 {
		name = name_optional[0]
	}
	publicKey, privateKey := crypto.GenerateKeyPair()
	address := crypto.ToAddress(publicKey)
	account := &types.Account{}
	account.Address = hex.EncodeToString(address)
	account.PrivateKey = hex.EncodeToString(privateKey)
	account.Balance = big.NewInt(0)
	account.Name = name
	now := time.Now()
	account.Created = now
	account.Updated = now

	return account, nil;
}

// GetAccount - Get account details
func GetAccount(delegateNode types.Node, address string) (*types.Account, error) {

	// Get account
	httpResponse, err := http.Get(fmt.Sprintf("http://%s:%d/v1/accounts/%s", delegateNode.HttpEndpoint.Host, delegateNode.HttpEndpoint.Port, address))
	if err != nil {
		return nil, err
	}
	defer httpResponse.Body.Close()

	// Read body.
	body, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal response.
	var response *types.Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	// Status?
	if response.Status != types.StatusOk {
		return nil, errors.New(fmt.Sprintf("%s: %s", response.Status, response.HumanReadableStatus))
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
	var account *types.Account
	err = json.Unmarshal(jsonMap["data"], &account)
	if err != nil {
		return nil, err
	}

	return account, nil
}

// PackageTx - Package a Transaction
func PackageTx(to string, tokens int64, time int64 ) (*types.Transaction, error) {

	transaction, err := types.NewTransferTokensTransaction(types.GetAccount().PrivateKey, types.GetAccount().Address, to, tokens, 0, time)
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

// TransferTokens - Send tokens FROM TO
func TransferTokens(delegateNode types.Node, privateKey string, from string, to string, tokens int64) (string, error) {
	// Create transfer tokens transaction.
	transaction, err := types.NewTransferTokensTransaction(privateKey, from, to, tokens, 0, utils.ToMilliSeconds(time.Now()))
	if err != nil {
		return "", err
	}

	// Post transaction.
	httpResponse, err := http.Post(fmt.Sprintf("http://%s:%d/v1/transactions", delegateNode.HttpEndpoint.Host, delegateNode.HttpEndpoint.Port), "application/json", bytes.NewBuffer([]byte(transaction.String())))
	if err != nil {
		return "", err
	}
	defer httpResponse.Body.Close()

	// Read body.
	body, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return "", err
	}

	// Unmarshal response.
	var response *types.Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}

	// Status?
	if response.Status != types.StatusPending {
		return "", errors.New(fmt.Sprintf("%s: %s", response.Status, response.HumanReadableStatus))
	}

	return transaction.Hash, nil
}

// DeploySmartContract - Deploy a smart contract, get the TX hash as result
func DeploySmartContract(delegateNode types.Node, privateKey string, from string, code string, abi string) (string, error) {
	// Create deploy smart contract transaction.
	transaction, err := types.NewDeployContractTransaction(privateKey, from, code, abi, utils.ToMilliSeconds(time.Now()))
	if err != nil {
		return "", err
	}

	// Post transaction.
	httpResponse, err := http.Post(fmt.Sprintf("http://%s:%d/v1/transactions", delegateNode.HttpEndpoint.Host, delegateNode.HttpEndpoint.Port), "application/json", bytes.NewBuffer([]byte(transaction.String())))
	if err != nil {
		return "", err
	}
	defer httpResponse.Body.Close()

	// Ready body.
	body, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return "", err
	}

	// Unmarshal response.
	var response *types.Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}

	// Status?
	if response.Status != types.StatusPending {
		return "", errors.New(fmt.Sprintf("%s: %s", response.Status, response.HumanReadableStatus))
	}

	return transaction.Hash, nil
}

// ExecuteSmartContractTransaction - Execute a smart contract, get the TX hash as result
func ExecuteSmartContractTransaction(delegateNode types.Node, privateKey string, from string, to string, method string, params []interface{}) (string, error) {
	// Create execute smart contract transaction.
	transaction, err := types.NewExecuteContractTransaction(privateKey, from, to, method, params, utils.ToMilliSeconds(time.Now()))
	if err != nil {
		return "", err
	}

	// Post transaction.
	httpResponse, err := http.Post(fmt.Sprintf("http://%s:%d/v1/transactions", delegateNode.HttpEndpoint.Host, delegateNode.HttpEndpoint.Port), "application/json", bytes.NewBuffer([]byte(transaction.String())))
	if err != nil {
		return "", err
	}
	defer httpResponse.Body.Close()

	// Read body.
	body, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return "", err
	}

	// Unmarshal response.
	var response *types.Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}

	// Status?
	if response.Status != types.StatusPending {
		return "", errors.New(fmt.Sprintf("%s: %s", response.Status, response.HumanReadableStatus))
	}

	return transaction.Hash, nil
}

// GetTransaction
func GetTransaction(delegateNode types.Node, hash string) (*types.Transaction, error) {

	// Get transaction.
	httpResponse, err := http.Get(fmt.Sprintf("http://%s:%d/v1/transactions/%s", delegateNode.HttpEndpoint.Host, delegateNode.HttpEndpoint.Port, hash))
	if err != nil {
		return nil, err
	}
	defer httpResponse.Body.Close()

	// Ready body.
	body, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal response.
	var response *types.Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	// Status?
	if response.Status != types.StatusOk {
		return nil, errors.New(fmt.Sprintf("%s: %s", response.Status, response.HumanReadableStatus))
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

	// Unmarshal transaction.
	var transaction *types.Transaction
	err = json.Unmarshal(jsonMap["data"], &transaction)
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

// GetReceipt - Get details about a transaction base on a TX hash
func GetReceipt(delegateNode types.Node, hash string) (*types.Receipt, error) {

	// Get transaction.
	transaction, err := GetTransaction(delegateNode, hash)
	if err != nil {
		return nil, err
	}

	return &transaction.Receipt, nil
}

// GetTransactions - Get details about sent transactions for a node
func GetTransactions(delegateNode types.Node, page, pageStart, pageSize string) ([]types.Transaction, error) {

	// Get sent transaction.
	httpResponse, err := http.Get(fmt.Sprintf("http://%s:%d/v1/transactions?page=%s&pageSize=%s&pageStart=%s,", delegateNode.HttpEndpoint.Host, delegateNode.HttpEndpoint.Port, page,pageSize,pageStart))
	if err != nil {
		return nil, err
	}
	defer httpResponse.Body.Close()

	// Read body.
	body, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal response.
	var response *types.Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	// Status?
	if response.Status != types.StatusOk {
		return nil, errors.New(fmt.Sprintf("%s: %s", response.Status, response.HumanReadableStatus))
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

// GetTransactionsSent - Get details about sent transactions for a node
func GetTransactionsSent(delegateNode types.Node, address, page, pageStart, pageSize string) ([]types.Transaction, error) {

	// Get sent transaction.
	httpResponse, err := http.Get(fmt.Sprintf("http://%s:%d/v1/transactions?from=%s&page=%s&pageSize=%s&pageStart=%s,", delegateNode.HttpEndpoint.Host, delegateNode.HttpEndpoint.Port, address,page,pageSize,pageStart))
	if err != nil {
		return nil, err
	}
	defer httpResponse.Body.Close()

	// Read body.
	body, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal response.
	var response *types.Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	// Status?
	if response.Status != types.StatusOk {
		return nil, errors.New(fmt.Sprintf("%s: %s", response.Status, response.HumanReadableStatus))
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

// GetTransactionsReceived - Get details about received transactions for a node
func GetTransactionsReceived(delegateNode types.Node, address, page, pageStart, pageSize string) ([]types.Transaction, error) {

	// Get received transactions.
	httpResponse, err := http.Get(fmt.Sprintf("http://%s:%d/v1/transactions?to=%s&page=%s&pageSize=%s&pageStart=%s,", delegateNode.HttpEndpoint.Host, delegateNode.HttpEndpoint.Port, address,page,pageSize,pageStart))
	if err != nil {
		return nil, err
	}
	defer httpResponse.Body.Close()

	// Read body.
	body, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal response.
	var response *types.Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	// Status?
	if response.Status != types.StatusOk {
		return nil, errors.New(fmt.Sprintf("%s: %s", response.Status, response.HumanReadableStatus))
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
