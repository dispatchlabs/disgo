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
func GetDelegates() ([]types.Node, error) {

	// Get delegates.
	httpResponse, err := http.Get("http://seed.dispatchlabs.io:1975/v1/delegates")
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

// TransferTokens - Send tokens FROM TO
func TransferTokens(delegateNode types.Node, privateKey string, from string, to string, tokens int64) error {

	// Create transfer tokens transaction.
	transaction, err := types.NewTransferTokensTransaction(privateKey, from, to, tokens, 0, utils.ToMilliSeconds(time.Now()))
	if err != nil {
		return err
	}

	// Post transaction.
	httpResponse, err := http.Post(fmt.Sprintf("http://%s:%d/v1/transactions", delegateNode.HttpEndpoint.Host, delegateNode.HttpEndpoint.Port), "application/json", bytes.NewBuffer([]byte(transaction.String())))
	if err != nil {
		return err
	}
	defer httpResponse.Body.Close()

	// Read body.
	body, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return err
	}

	// Unmarshal response.
	var response *types.Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}

	// Status?
	if response.Status != types.StatusPending {
		return errors.New(fmt.Sprintf("%s: %s", response.Status, response.HumanReadableStatus))
	}

	return nil
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
func ExecuteSmartContractTransaction(delegateNode types.Node, privateKey string, from string, to string, abi string, method string, params []interface{}) (string, error) {

	// Create execute smart contract transaction.
	transaction, err := types.NewExecuteContractTransaction(privateKey, from, to, abi, method, params, utils.ToMilliSeconds(time.Now()))
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

// GetTransaction -  TODO
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

	// Unmarshal account.
	var transaction *types.Transaction
	err = json.Unmarshal(jsonMap["data"], &transaction)
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

// GetReceipt - Get details about a transaction base on a TX hash
func GetReceipt(delegateNode types.Node, hash string) (*types.Receipt, error) {

	// Get status.
	httpResponse, err := http.Get(fmt.Sprintf("http://%s:%d/v1/receipts/%s", delegateNode.HttpEndpoint.Host, delegateNode.HttpEndpoint.Port, hash))
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

	// Unmarshal account.
	var receipt *types.Receipt
	err = json.Unmarshal(jsonMap["data"], &receipt)
	if err != nil {
		return nil, err
	}

	return receipt, nil
}

// GetTransactionsSent - Get details about sent transactions for a node
func GetTransactionsSent(delegateNode types.Node, address string) ([]types.Transaction, error) {

	// Get sent transaction.
	httpResponse, err := http.Get(fmt.Sprintf("http://%s:%d/v1/transactions/from/%s", delegateNode.HttpEndpoint.Host, delegateNode.HttpEndpoint.Port, address))
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
func GetTransactionsReceived(delegateNode types.Node, address string) ([]types.Transaction, error) {

	// Get received transactions.
	httpResponse, err := http.Get(fmt.Sprintf("http://%s:%d/v1/transactions/to/%s", delegateNode.HttpEndpoint.Host, delegateNode.HttpEndpoint.Port, address))
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
