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
	"fmt"
	"math/rand"

	"github.com/dgraph-io/badger"
	"github.com/dispatchlabs/disgo/commons/services"
	"github.com/dispatchlabs/disgo/commons/types"
	"github.com/dispatchlabs/disgo/commons/utils"
	"github.com/dispatchlabs/disgo/disgover"

	"encoding/hex"
	"math/big"
	"strings"
	"time"

	"github.com/dispatchlabs/disgo/dvm"
	"github.com/dispatchlabs/disgo/dvm/ethereum/abi"
)

// startGossiping
func (this *DAPoSService) startGossiping(transaction *types.Transaction) *types.Receipt {
	txn := services.NewTxn(false)
	defer txn.Discard()
	receipt := types.NewReceipt(types.RequestNewTransaction)

	// Verify?
	if !transaction.Verify() {
		utils.Info(fmt.Sprintf("invalid transaction [hash=%s]", transaction.Hash))
		receipt.Status = types.StatusInvalidTransaction
		return receipt
	}

	// Duplicate transaction?
	_, err := txn.Get([]byte(transaction.Key()))
	if err == nil {
		utils.Info(fmt.Sprintf("duplicate transaction [hash=%s]", transaction.Hash))
		receipt.Status = types.StatusDuplicateTransaction
		return receipt
	}
	if err != badger.ErrKeyNotFound {
		utils.Error(err)
		receipt.Status = types.StatusInternalError
		receipt.HumanReadableStatus = err.Error()
		return receipt
	}

	// TODO: Check minimum hertz, balance, and negative value!!!!!

	// Are we already gossiping about this transaction?
	_, err = types.ToTransactionFromCache(services.GetCache(),transaction.Hash)
	if err == nil{
		utils.Info(fmt.Sprintf("already processing this transaction [hash=%s]", transaction.Hash))
		receipt.Status = types.StatusAlreadyProcessingTransaction
		return receipt
	}

	// Cache receipt.
	receipt.Cache(services.GetCache())

	// Cache gossip with my rumor.
	gossip := types.NewGossip(*transaction, *receipt)
	rumor := types.NewRumor(types.GetAccount().PrivateKey, types.GetAccount().Address, transaction.Hash)
	gossip.Rumors = append(gossip.Rumors, *rumor)
	gossip.Cache(services.GetCache())

	this.gossipChan <- gossip

	return receipt
}

// Temp_ProcessTransaction -
func (this *DAPoSService) Temp_ProcessTransaction(gossip *types.Gossip) {
	go func(g *types.Gossip) {
		this.gossipChan <- g
	}(gossip)
}

// synchronizeGossip
func (this *DAPoSService) synchronizeGossip(gossip *types.Gossip) (*types.Gossip, error) {

	// Get or set receipt?
	_, err := types.ToReceiptFromCache(services.GetCache(),gossip.ReceiptId)
	if err != nil {
		receipt := types.NewReceipt(types.RequestNewTransaction)
		receipt.Id = gossip.ReceiptId
		receipt.Cache(services.GetCache())
	}

	// Set synchronizedGossip.
	var synchronizedGossip *types.Gossip
	ourGossip, err := types.ToGossipFromCache(services.GetCache(),gossip.Transaction.Hash)
	if err != nil {
		synchronizedGossip = gossip
	} else {
		synchronizedGossip = ourGossip
		for _, rumor := range gossip.Rumors {
			if !synchronizedGossip.ContainsRumor(rumor.Address) && rumor.Verify() { // We don't want to propagate cryptographic lies.
				synchronizedGossip.Rumors = append(synchronizedGossip.Rumors, rumor)
			}
		}
	}

	// Did rumor?
	didRumor := false
	for _, rumor := range synchronizedGossip.Rumors {
		if rumor.Address == types.GetAccount().Address {
			didRumor = true
		}
	}
	if !didRumor && gossip.Transaction.Verify() { // We don't want to propagate cryptographic lies.
		synchronizedGossip.Rumors = append(gossip.Rumors, *types.NewRumor(types.GetAccount().PrivateKey, types.GetAccount().Address, gossip.Transaction.Hash))
	}
	return synchronizedGossip, nil
}

// gossipWorker
func (this *DAPoSService) gossipWorker() {
	var gossip *types.Gossip
	for {
		select {
		case gossip = <-this.gossipChan:

			go func(theGossip *types.Gossip) {
				delegateNodes, err := types.ToNodesByTypeFromCache(services.GetCache(),types.TypeDelegate)
				if err != nil {
					utils.Error(err)
				}

				if len(gossip.Rumors) >= len(delegateNodes)*2/3 {
					utils.Debug("inside 2/3rds")
					txn := services.NewTxn(true)
					defer txn.Discard()
					// Has this transaction already been processed?
					_, err := txn.Get([]byte(gossip.Transaction.Key()))
					if err == nil {
						utils.Debug("already in db")
						return
					}
					//execute tx

					this.transactionChan <- gossip //TODO: insert into queue
					go this.broadcast(delegateNodes,gossip)
					//broadcast tx

					return
				}

				// Gossip to random delegate.
				for i := 0; i < 2*len(delegateNodes); i++ { // TODO: the `2 * ...` is a random pick to kind of exaust the list
					node := this.getRandomDelegate(gossip, delegateNodes)
					if node == nil {
						continue
					}

					// Peer gossip.
					peerGossip, err := this.peerGossipGrpc(*node, gossip) // TODO: Maybe this should be a different channel????
					if err != nil {
						utils.Error(err)
						continue
					}
					this.gossipChan <- peerGossip
					break
				}
			}(gossip)
		}
	}
}

// getRandomDelegate
func (this *DAPoSService) getRandomDelegate(gossip *types.Gossip, delegateNodes []*types.Node) *types.Node {
	if len(delegateNodes) == 0 {
		return nil
	}

	// Get delegates that have not rumored?
	delegatesNotRumored := make([]*types.Node, 0)
	for _, node := range delegateNodes {
		if gossip.ContainsRumor(node.Address) || node.Address == disgover.GetDisGoverService().ThisNode.Address {
			continue
		}
		delegatesNotRumored = append(delegatesNotRumored, node)
	}
	if len(delegatesNotRumored) == 0 {
		return nil
	}

	// Find random delegate.
	rand.Seed(time.Now().UTC().UnixNano())
	index := rand.Intn(len(delegatesNotRumored))
	return delegatesNotRumored[index]
}

// gossipWorker - transfer tokens, deploy smart contract, and execution of smart contract.
func (this *DAPoSService) transactionWorker() {

	var gossip *types.Gossip
	for {
		select {
		case gossip = <-this.transactionChan:

			utils.Debug("transactionworker")
			// Get receipt.
			var receipt *types.Receipt
			value, err := types.ToReceiptFromCache(services.GetCache(),gossip.ReceiptId)
			if err != nil {
				utils.Error(fmt.Sprintf("receipt not found [id=%s]", gossip.ReceiptId))
				receipt = types.NewReceipt(types.RequestNewTransaction)
				receipt.Status = types.StatusReceiptNotFound
				receipt.Cache(services.GetCache())
				continue
			}
			receipt = value

			// TODO: Should we thread this?
			// Execute.
			utils.Debug("about to excecute")
			executeTransaction(&gossip.Transaction, receipt, gossip)
		}
	}
}

// executeTransaction
func executeTransaction(transaction *types.Transaction, receipt *types.Receipt, gossip *types.Gossip) {
	services.Lock(transaction.Hash)
	defer services.Unlock(transaction.Hash)
	txn := services.NewTxn(true)
	defer txn.Discard()

	utils.Debug("executing transaction")
	// Has this transaction already been processed?
	_, err := txn.Get([]byte(transaction.Key()))
	if err == nil {
		return
	}

	utils.Debug("writing tx to db")
	err = transaction.Set(txn,services.GetCache())
	if err != nil {
		utils.Error(err)
		receipt.Status = types.StatusInternalError
		receipt.HumanReadableStatus = err.Error()
		receipt.Cache(services.GetCache())
		return
	}
	// TODO: Should we verify the transaction again?

	// Find/create fromAccount?
	now := time.Now()
	fromAccount, err := types.ToAccountByAddress(txn, transaction.From)
	if err != nil {
		if err == badger.ErrKeyNotFound {
			fromAccount = &types.Account{Address: transaction.From, Balance: big.NewInt(0), Created: now}
		} else {
			utils.Error(err)
			receipt.Status = types.StatusInternalError
			receipt.HumanReadableStatus = err.Error()
			receipt.Cache(services.GetCache())
			return
		}
	}

	// Find/create toAccount?
	toAccount, err := types.ToAccountByAddress(txn, transaction.To)
	if err != nil {
		if err == badger.ErrKeyNotFound {
			toAccount = &types.Account{Address: transaction.To, Balance: big.NewInt(0), Created: now}
		} else {
			utils.Error(err)
			receipt.SetInternalErrorWithNewTransaction(services.GetDb(), err)
			return
		}
	}

	if len(strings.TrimSpace(transaction.To)) == 0 &&
		len(strings.TrimSpace(transaction.Code)) != 0 {

		// Deploy Smart Contract
		dvmService := dvm.GetDVMService()
		dvmResult, err := dvmService.DeploySmartContract(transaction)
		if err != nil {
			utils.Error(err, utils.GetCallStackWithFileAndLineNumber())
		}

		err = processDVMResult(transaction, dvmResult, receipt)
		if err != nil {
			utils.Error(err)
			receipt.Status = types.StatusInternalError
			receipt.HumanReadableStatus = err.Error()
			receipt.Cache(services.GetCache())
			return
		}

		// Set contract account.
		contractAccount := &types.Account{Address: hex.EncodeToString(dvmResult.ContractAddress[:]), Balance: big.NewInt(0), Updated: now, Created: now}
		err = contractAccount.Set(txn, services.GetCache())
		if err != nil {
			utils.Error(err)
			receipt.Status = types.StatusInternalError
			receipt.HumanReadableStatus = err.Error()
			receipt.Cache(services.GetCache())
			return
		}
		receipt.ContractAddress = contractAccount.Address
		utils.Info(fmt.Sprintf("deployed contract [receiptId=%s hash=%s, contractAddress=%s]", receipt.Id, transaction.Hash, contractAccount.Address))

	} else if len(strings.TrimSpace(transaction.To)) != 0 &&
		len(strings.TrimSpace(transaction.Abi)) != 0 &&
		len(strings.TrimSpace(transaction.Method)) != 0 {

		// Execute Smart Contract Method
		dvmService := dvm.GetDVMService()
		dvmResult, err1 := dvmService.ExecuteSmartContract(transaction)
		if err1 != nil {
			utils.Error(err, utils.GetCallStackWithFileAndLineNumber())
		}

		err = processDVMResult(transaction, dvmResult, receipt)
		if err != nil {
			utils.Error(err)
			receipt.Status = types.StatusInternalError
			receipt.HumanReadableStatus = err.Error()
			receipt.Cache(services.GetCache())
			return
		}
		receipt.ContractAddress = transaction.To
		utils.Info(fmt.Sprintf("executed contract [receiptId=%s hash=%s, contractAddress=%s]", receipt.Id, transaction.Hash, transaction.To))

	} else {

		//
		// Check for a list of invalid cases
		//

		if len(strings.TrimSpace(transaction.To)) == 0 {
			utils.Error(fmt.Sprintf("invalid transaction data [hash=%s]", transaction.Hash))
			receipt.SetStatusWithNewTransaction(services.GetDb(), types.StatusInvalidTransaction)
			return
		}

		if strings.TrimSpace(transaction.From) == strings.TrimSpace(transaction.To) {
			utils.Error(fmt.Sprintf("invalid transaction data [hash=%s]", transaction.Hash))
			receipt.SetStatusWithNewTransaction(services.GetDb(), types.StatusInvalidTransaction)
			return
		}

		// Sufficient tokens?
		if fromAccount.Balance.Int64() < transaction.Value {
			utils.Error(fmt.Sprintf("insufficient tokens [hash=%s]", transaction.Hash))
			receipt.SetStatusWithNewTransaction(services.GetDb(), types.StatusInsufficientTokens)
			return
		}

		// All seems valid - do the account adjustments
		fromAccount.Balance.SetInt64(fromAccount.Balance.Int64() - transaction.Value)
		toAccount.Balance.SetInt64(toAccount.Balance.Int64() + transaction.Value)
		utils.Info(fmt.Sprintf("transferred tokens [receiptId=%s hash=%s, rumors=%d]", receipt.Id, transaction.Hash, len(gossip.Rumors)))
	}

	// Save fromAccount.
	fromAccount.Updated = now
	err = fromAccount.Set(txn,services.GetCache())
	if err != nil {
		utils.Error(err)
		receipt.Status = types.StatusInternalError
		receipt.HumanReadableStatus = err.Error()
		receipt.Cache(services.GetCache())
		return
	}

	// Save toAccount.
	toAccount.Updated = now
	err = toAccount.Set(txn,services.GetCache())
	if err != nil {
		utils.Error(err)
		receipt.Status = types.StatusInternalError
		receipt.HumanReadableStatus = err.Error()
		receipt.Cache(services.GetCache())
		return
	}


	// Save receipt.
	receipt.Status = types.StatusOk
	receipt.Cache(services.GetCache())
	err = receipt.Set(txn,services.GetCache())
	if err != nil {
		utils.Error(err)
		receipt.Status = types.StatusInternalError
		receipt.HumanReadableStatus = err.Error()
		receipt.Cache(services.GetCache())
		return
	}

	// Save gossip.
	err = gossip.Set(txn,services.GetCache())
	if err != nil {
		utils.Error(err)
		receipt.Status = types.StatusInternalError
		receipt.HumanReadableStatus = err.Error()
		receipt.Cache(services.GetCache())
		return
	}

	// Commit.
	err = txn.Commit(nil)
	if err != nil {
		if err == badger.ErrConflict { // Another thread already committed this transaction. This will happen, which is ok.
			return
		}
		utils.Error(err)
		receipt.Status = types.StatusInternalError
		receipt.HumanReadableStatus = err.Error()
		receipt.Cache(services.GetCache())
		return
	}
}

//TODO: implement if useful
//func commit(transaction *types.Transaction) {}
// processDVMResult
func processDVMResult(transaction *types.Transaction, dvmResult *dvm.DVMResult, receipt *types.Receipt) error {
	utils.Info("######### DUMPING-DVMResult #########")
	utils.Info(dvmResult)

	if dvmResult.ContractMethodExecError != nil {
		utils.Error(dvmResult.ContractMethodExecError)
		return dvmResult.ContractMethodExecError
	}

	var errorToReturn error

	// Try read the execution result
	if len(strings.TrimSpace(dvmResult.ABI)) > 0 {
		fromHexAsByteArray, _ := hex.DecodeString(dvmResult.ABI)
		abiAsString := string(fromHexAsByteArray)
		jsonABI, err := abi.JSON(strings.NewReader(abiAsString))
		if err == nil {

			if method, ok := jsonABI.Methods[dvmResult.ContractMethod]; ok {
				marshalledValues, err := method.Outputs.UnpackValues(dvmResult.ContractMethodExecResult)
				if err == nil {
					utils.Info(fmt.Sprintf("CONTRACT-CALL-RES: %v", marshalledValues))
					receipt.ContractResult = marshalledValues
				} else {
					errorToReturn = err
					utils.Error(err)
				}
			}

			// var parsedRes []interface{}
			// var parsedRes = make([]interface{}, 3)
			// err = jsonABI.Unpack(&parsedRes, transaction.Method, dvmResult.ContractMethodExecResult)
			// if err == nil {
			// 	utils.Info(fmt.Sprintf("CONTRACT-CALL-RES: %s", parsedRes))
			// 	receipt.ContractResult = parsedRes
			// } else {
			// 	errorToReturn = err
			// 	utils.Error(err)
			// }
		} else {
			errorToReturn = err
			utils.Error(err)
		}
	}

	return errorToReturn
}

func (this *DAPoSService)broadcast(delegateNodes []*types.Node,gossip *types.Gossip){
	utils.Debug("broadcasting")
	for _ , deli := range delegateNodes{
		_, err := this.peerGossipGrpc(*deli, gossip)
		if err != nil {
			utils.Error(err)
			continue
		}
	}
}