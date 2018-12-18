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
	"math/big"
	"strings"
	"time"
	"encoding/hex"

	"github.com/dgraph-io/badger"
	"github.com/dispatchlabs/disgo/commons/helper"
	"github.com/dispatchlabs/disgo/commons/services"
	"github.com/dispatchlabs/disgo/commons/types"
	"github.com/dispatchlabs/disgo/commons/utils"
	"github.com/dispatchlabs/disgo/disgover"
	"github.com/dispatchlabs/disgo/dvm"
	"github.com/dispatchlabs/disgo/dvm/ethereum/abi"
	"github.com/dispatchlabs/disgo/dvm/ethereum/params"
	"encoding/base64"
	"bytes"
)

var delegateMap = map[string]*types.Node{}

// startGossiping
func (this *DAPoSService) startGossiping(transaction *types.Transaction) *types.Response {
	utils.Debug("startGossiping")
	txn := services.NewTxn(false)
	defer txn.Discard()

	// Verify?
	err := transaction.Verify()
	if err != nil {
		utils.Info(fmt.Sprintf("invalid transaction [hash=%s]", transaction.Hash))
		return types.NewResponseWithStatus(types.StatusInvalidTransaction, err.Error())
	}
	elapsedMilliSeconds := utils.ToMilliSeconds(time.Now()) - transaction.Time
	if elapsedMilliSeconds > types.TxReceiveTimeout {
		utils.Error(fmt.Sprintf("Timed out [hash=%s]", transaction.Hash))
		return types.NewResponseWithStatus(types.StatusTransactionTimeOut, "Transaction was received later than 3 second limit")
	}

	// Duplicate transaction?
	_, err = txn.Get([]byte(transaction.Key()))
	if err == nil {
		utils.Info(fmt.Sprintf("duplicate transaction [hash=%s]", transaction.Hash))
		return types.NewResponseWithStatus(types.StatusDuplicateTransaction, "Duplicate transaction")
	}
	if err != badger.ErrKeyNotFound {
		utils.Error(err)
		return types.NewResponseWithError(err)
	}

	// TODO: Check minimum hertz

	// Are we already gossiping about this transaction?
	_, err = types.ToTransactionFromCache(services.GetCache(), transaction.Hash)
	if err == nil {
		utils.Info(fmt.Sprintf("already processing this transaction [hash=%s]", transaction.Hash))
		return types.NewResponseWithStatus(types.StatusAlreadyProcessingTransaction, "Transaction is already being processed")
	}
	// Cache gossip with my rumor.
	gossip := types.NewGossip(*transaction)
	rumor := types.NewRumor(types.GetKey(), types.GetAccount().Address, transaction.Hash)
	gossip.Rumors = append(gossip.Rumors, *rumor)

	this.cacheOnFirstReceive(gossip)
	this.gossipChan <- gossip

	return types.NewResponseWithStatus(types.StatusPending, "Pending")
}

func (this *DAPoSService) cacheOnFirstReceive(gossip *types.Gossip) {
	// Cache receipt.
	utils.Debug(fmt.Sprintf("First receipt of transaction [hash=%s] [Rumors=%d]", gossip.Transaction.Hash, len(gossip.Rumors)))
	receipt := types.NewReceipt(gossip.Transaction.Hash)
	receipt.Cache(services.GetCache())

	// Cache gossip with my rumor.
	gossip.Cache(services.GetCache())

	// transaction.Receipt.Status = types.StatusReceived
	gossip.Transaction.Cache(services.GetCache())

	delegateNodes, err := types.ToNodesByTypeFromCache(services.GetCache(), types.TypeDelegate)
	if err != nil {
		utils.Error(err)
		return
	}
	for _, node := range delegateNodes {
		haveSent := gossip.HaveSent(services.GetCache(), gossip.Transaction.Hash, node.Address)
		isThisAddress := node.Address == disgover.GetDisGoverService().ThisNode.Address

		if !haveSent && !isThisAddress {
			this.peerGossipGrpc(*node, gossip)
		}
	}

}

// Temp_ProcessTransaction -
func (this *DAPoSService) Temp_ProcessTransaction(transaction *types.Transaction) *types.Response {
	// go func(tx *types.Transaction) {

	// Cache receipt.
	receipt := types.NewReceipt(transaction.Hash)
	receipt.Cache(services.GetCache())

	// Cache gossip with my rumor.
	gossip := types.NewGossip(*transaction)
	rumor := types.NewRumor(types.GetKey(), types.GetAccount().Address, transaction.Hash)
	gossip.Rumors = append(gossip.Rumors, *rumor)
	gossip.Cache(services.GetCache())

	this.gossipChan <- gossip

	return types.NewResponseWithStatus(types.StatusPending, "Pending")
	// }(transaction)
}

// synchronizeGossip
func (this *DAPoSService) synchronizeGossip(gossip *types.Gossip) (*types.Gossip, error, bool) {

	// PersistAndCache synchronizedGossip.
	var synchronizedGossip *types.Gossip
	hasAll := false
	ourGossip, err := types.ToGossipFromCache(services.GetCache(), gossip.Transaction.Hash)
	if err != nil {
		synchronizedGossip = gossip
		gossip.Transaction.Cache(services.GetCache())

	} else {
		synchronizedGossip = ourGossip
		for _, rumor := range gossip.Rumors {
			hasAll = true
			if !ourGossip.ContainsRumor(rumor.Address) {
				hasAll = false
			}
			if !synchronizedGossip.ContainsRumor(rumor.Address) && rumor.Verify() { // We don't want to propagate cryptographic lies.
				synchronizedGossip.Rumors = append(synchronizedGossip.Rumors, rumor)
			}
		}
		//we have already seen all of these rumors, so we don't want to put them back into our Gossip worker
	}

	// Did rumor?
	didRumor := false
	for _, rumor := range synchronizedGossip.Rumors {
		if rumor.Address == types.GetAccount().Address {
			didRumor = true
		}
	}
	if !didRumor {

		// We don't want to propagate cryptographic lies.
		err = gossip.Transaction.Verify()
		if err == nil {
			synchronizedGossip.Rumors = append(gossip.Rumors, *types.NewRumor(types.GetKey(), types.GetAccount().Address, gossip.Transaction.Hash))
		} else {
			utils.Error(err)
			return synchronizedGossip, err, true
		}
		//This is the first time receiving this gossip
		this.cacheOnFirstReceive(synchronizedGossip)
	}
	return synchronizedGossip, nil, !hasAll
}

// gossipWorker
func (this *DAPoSService) gossipWorker() {
	var gossip *types.Gossip
	for {
		select {
		case gossip = <-this.gossipChan:

			go func(gossip *types.Gossip) {
				// Find nodes in cache?
				delegateNodes, err := types.ToNodesByTypeFromCache(services.GetCache(), types.TypeDelegate)
				if err != nil {
					utils.Error(err)
					return
				}
				if delegateMap == nil || len(delegateMap) == 0 {
					for _, d := range delegateNodes {
						delegateMap[d.Address] = d
					}
				}

				// Gossip timeout?
				if len(gossip.Rumors) > 1 {
					if !types.ValidateTimeDelta(gossip.Rumors) {
						utils.Warn("The rumors have an invalid time delta (greater than gossip timeout milliseconds")
						updateReceiptStatus(gossip.Transaction.Hash, types.StatusGossipingTimedOut)
						//ignore this gossip's rumors and hopefully still hit 2/3 from well timed gossip, but keep listening
						receipt := types.NewReceipt(gossip.Transaction.Hash)
						receipt.Status = types.StatusGossipingTimedOut
						receipt.Cache(services.GetCache())

						return
					}
				}
				// Do we have 2/3 of rumors?
				if float32(len(gossip.Rumors)) >= float32(len(delegateNodes))*2/3 {
					if !this.gossipQueue.Exists(gossip.Transaction.Hash) {
						//for _, rumor := range gossip.Rumors {
						//	utils.Info(fmt.Sprintf("rumor from: [address=%s] for [tx=%s] with [hash=%s]", rumor.Address, rumor.TransactionHash, rumor.Hash))
						//}
						this.gossipQueue.Push(gossip)

						go func() {
							//adding timeout as a function of tx time.  If tx is in the future, add future delta to the default timeout
							delta := gossip.Transaction.Time - utils.ToMilliSeconds(time.Now())
							totalMilliseconds := (types.GossipTimeout * len(delegateNodes)) + types.TxReceiveTimeout
							timeout := time.Duration(totalMilliseconds) * time.Millisecond
							utils.Debug("Timeout Queue value: ", timeout)
							if delta > 0 {
								timeout = time.Millisecond*time.Duration(delta) + timeout
							}
							time.Sleep(timeout)
							this.timoutChan <- true
						}()
						//for _, node := range delegateNodes {
						//	haveSent := gossip.HaveSent(services.GetCache(), gossip.Transaction.Hash, node.Address)
						//
						//	if !haveSent {
						//		utils.Info(fmt.Sprintf("*********** Last send after 2/3 [hash=%s] to delegate [Port %d] [address=%s]", gossip.Transaction.Hash, node.HttpEndpoint.Port, node.Address))
						//		this.peerGossipGrpc(*node, gossip)
						//	}
						//}
					}
					//No reason to keep gossiping if we are executing the transaction
					return
				}

				// Did we already receive all the delegate's rumors?
				if len(gossip.Rumors) == len(delegateNodes) {
					utils.Debug("already received all rumors from delegates")
					return
				}

				// Get random delegate?
				node := this.getRandomDelegate(gossip, delegateNodes)
				if node == nil {
					utils.Warn("did not find any delegates to rumor with")
					gossip.Cache(services.GetCache())
					updateReceiptStatus(gossip.Transaction.Hash, types.StatusCouldNotReachConsensus)

					//Commented out because if we have no-one left to talk to, why are we continuing?
					//Plus it was causing me all kinds of timeout problems
					if len(gossip.Rumors) != len(delegateNodes) {
						utils.Debug(fmt.Sprintf("Stopped Gossiping when there are %d nodes that don't have a rumor", len(delegateNodes)-len(gossip.Rumors)))
					}

					return
				}
				utils.Debug(fmt.Sprintf("Picked RandomDelegate = [hash=%s] to delegate [Port %d] [address=%s]", gossip.Transaction.Hash, node.HttpEndpoint.Port, node.Address))

				// Peer gossip.
				//peerGossip, err := this.peerGossipGrpc(*node, gossip)
				_, err = this.peerGossipGrpc(*node, gossip)
				if err != nil {
					utils.Error(err)
					this.gossipChan <- gossip
					return
				}
				//this.gossipChan <- peerGossip
			}(gossip)
		}
	}
}

func updateReceiptStatus(txHash, status string) {
	receipt, err := types.ToReceiptFromCache(services.GetCache(), txHash)
	if err != nil {
		utils.Error(err)
	} else {
		receipt.Status = status
		receipt.Cache(services.GetCache())
	}
}

// getRandomDelegate
func (this *DAPoSService) getRandomDelegate(gossip *types.Gossip, delegateNodes []*types.Node) *types.Node {
	if len(delegateNodes) == 0 {
		utils.Error("Delegate Nodes length is 0")
		return nil
	}

	// Get delegates that have not rumored?
	delegatesNotRumored := make([]*types.Node, 0)
	for _, node := range delegateNodes {
		haveSent := gossip.HaveSent(services.GetCache(), gossip.Transaction.Hash, node.Address)
		containsRumor := gossip.ContainsRumor(node.Address)
		isThisAddress := node.Address == disgover.GetDisGoverService().ThisNode.Address

		if !node.IsAvailable() {
			utils.Debug(fmt.Sprintf("Node is not available: [hash=%s] to delegate [Port %d] [address=%s]", gossip.Transaction.Hash, node.HttpEndpoint.Port, node.Address))
		}
		if haveSent {
			utils.Debug(fmt.Sprintf("Have Sent: [hash=%s] to delegate [Port %d] [address=%s]", gossip.Transaction.Hash, node.HttpEndpoint.Port, node.Address))
		}
		if !containsRumor {
			utils.Debug(fmt.Sprintf("Don't have a Rumor for: [hash=%s] to delegate [Port %d] [address=%s]", gossip.Transaction.Hash, node.HttpEndpoint.Port, node.Address))
		}
		if isThisAddress || haveSent || !node.IsAvailable() {
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
	for _, d := range delegatesNotRumored {
		fmt.Printf("Available: %s", d.GrpcEndpoint.Host)
	}
	return delegatesNotRumored[index]
}

// gossipWorker - transfer tokens, deploy smart contract, and execution of smart contract.
func (this *DAPoSService) transactionWorker() {

	for {
		select {
		case <-this.timoutChan:
			this.doWork()
		}
	}
}

func (this *DAPoSService) doWork() {
	var gossip *types.Gossip

	if this.gossipQueue.HasAvailable() {
		gossip = this.gossipQueue.Pop()
		// Get receipt.
		receipt, err := types.ToReceiptFromCache(services.GetCache(), gossip.Transaction.Hash)
		if err != nil {
			utils.Error(fmt.Sprintf("receipt not found [hash=%s]", gossip.Transaction.Hash))
			receipt = types.NewReceipt(gossip.Transaction.Hash)
			receipt.Status = types.StatusReceiptNotFound
			receipt.Cache(services.GetCache())
			return
		}
		initialRcvDuration := gossip.Rumors[0].Time - gossip.Transaction.Time
		utils.Debug("Initial Receive Duration = ", initialRcvDuration, types.TxReceiveTimeout)
		if initialRcvDuration >= types.TxReceiveTimeout {
			utils.Error(fmt.Sprintf("Timed out [hash=%s] %v milliseconds", gossip.Transaction.Hash, initialRcvDuration))
			receipt = types.NewReceipt(gossip.Transaction.Hash)
			receipt.Status = types.StatusTransactionTimeOut
			receipt.Cache(services.GetCache())
			return
		}
		receipt.Created = time.Now()
		if types.GetConfig().IsBookkeeper {
			executeTransaction(&gossip.Transaction, receipt, gossip)
		}
	}
}

// executeTransaction
func executeTransaction(transaction *types.Transaction, receipt *types.Receipt, gossip *types.Gossip) {
	utils.Info("executeTransaction --> ", transaction.Hash)
	services.Lock(transaction.Hash)
	defer services.Unlock(transaction.Hash)

	txn := services.NewTxn(true)
	defer txn.Discard()

	// Has this transaction already been processed?
	_, err := txn.Get([]byte(transaction.Key()))
	if err == nil {
		utils.Info("Already executed this transaction --> ", transaction.Hash)
		return
	}

	//Get Min Hetz you will use (intrinsic hertz)
	minHertzUsed := params.CallValueTransferGas

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
		minHertzUsed += params.CallNewAccountGas

	}

	// Find/create toAccount?
	var toAccount *types.Account
	if transaction.To != "" {
		toAccount, err = types.ToAccountByAddress(txn, transaction.To)
		if err != nil {
			if err == badger.ErrKeyNotFound {
				toAccount = &types.Account{Address: transaction.To, Balance: big.NewInt(0), Created: now}
				minHertzUsed += params.CallNewAccountGas
			} else {
				utils.Error(err)
				receipt.SetInternalErrorWithNewTransaction(services.GetDb(), err)
				return
			}
		}
	}

	//Check to see if there is enough Hertz to execute minimum
	availableHertz, err := types.CheckMinimumAvailable(txn, services.GetCache(), fromAccount.Address, fromAccount.Balance.Uint64())
	if err != nil {
		utils.Error(err)
	}
	if availableHertz < (minHertzUsed * types.HertzMultiplier) {
		msg := fmt.Sprintf("Account %s has a hertz balance of %d\n", fromAccount.Address, availableHertz)
		utils.Error(msg)
		receipt.SetStatusWithNewTransaction(services.GetDb(), types.StatusInsufficientHertz)
		return
	}
	var hertz uint64
	// Execute.
	switch transaction.Type {
	case types.TypeTransferTokens:
		// Sufficient tokens?
		if fromAccount.Balance.Int64() < transaction.Value {
			utils.Error(fmt.Sprintf("insufficient tokens [hash=%s]", transaction.Hash))
			receipt.SetStatusWithNewTransaction(services.GetDb(), types.StatusInsufficientTokens)
			return
		}

		fromAccount.Balance.SetInt64(fromAccount.Balance.Int64() - transaction.Value)
		toAccount.Balance.SetInt64(toAccount.Balance.Int64() + transaction.Value)

		hertz = minHertzUsed
		utils.Info(fmt.Sprintf("transferred tokens [hash=%s, rumors=%d]", transaction.Hash, len(gossip.Rumors)))
		break
	case types.TypeDeploySmartContract:
		dvmService := dvm.GetDVMService()

		// ENCODE to HEX here, the DECODE is happening in GetABI()
		transaction.Abi = hex.EncodeToString([]byte(transaction.Abi))

		dvmResult, err := dvmService.DeploySmartContract(transaction)
		if err != nil {
			utils.Error(err, utils.GetCallStackWithFileAndLineNumber())
			receipt.Status = types.StatusInternalError
			receipt.HumanReadableStatus = err.Error()
			receipt.Cache(services.GetCache())
			return
		}

		err = processDVMResult(transaction, dvmResult, receipt)
		if err != nil {
			utils.Error(err)
			receipt.Status = types.StatusInternalError
			receipt.HumanReadableStatus = err.Error()
			receipt.Cache(services.GetCache())
			return
		}

		// Update contract account.
		smartContractAddress := hex.EncodeToString(dvmResult.ContractAddress[:])
		for _, stateObject := range dvmResult.StorageState.EthStateDB.StateObjects {
			if stateObject.Account().Address == smartContractAddress {
				stateObject.Account().TransactionHash = transaction.Hash
				stateObject.Account().Persist(txn)
				break
			}
		}
		//receipt.ContractAddress = contractAccount.Address
		receipt.ContractAddress = smartContractAddress
		hertz = minHertzUsed + dvmResult.CumulativeHertzUsed
		utils.Info(fmt.Sprintf("deployed contract [hash=%s, contractAddress=%s]", transaction.Hash, smartContractAddress))
		break
	case types.TypeExecuteSmartContract:

		// READ PARAMS
		contractTx, err := types.ToTransactionByAddress(txn, transaction.To)
		if err != nil {
			utils.Error(err, utils.GetCallStackWithFileAndLineNumber())
			receipt.Status = types.StatusInternalError
			receipt.HumanReadableStatus = err.Error()
			receipt.Cache(services.GetCache())
			return
		}

		transaction.Abi = contractTx.Abi
		_, err = helper.GetConvertedParams(transaction)
		if err != nil {
			utils.Error(err, utils.GetCallStackWithFileAndLineNumber())
			receipt.Status = types.StatusInternalError
			receipt.HumanReadableStatus = err.Error()
			receipt.Cache(services.GetCache())
			return
		}

		dvmService := dvm.GetDVMService()
		dvmResult, err1 := dvmService.ExecuteSmartContract(transaction)
		if err1 != nil {
			utils.Error(err, utils.GetCallStackWithFileAndLineNumber())
		}

		hertz = minHertzUsed + dvmResult.CumulativeHertzUsed

		err = processDVMResult(transaction, dvmResult, receipt)
		if err != nil {
			utils.Error(err)
			hertz = hertz * types.HertzMultiplier

			receipt.Status = types.StatusInternalError
			receipt.HumanReadableStatus = err.Error()
			receipt.Cache(services.GetCache())

			rateLimit, err := types.NewRateLimit(transaction.From, transaction.Hash,  hertz)
			if err != nil {
				utils.Error(err)
			}
			window := helper.AddHertz(txn, services.GetCache(), hertz);
			rateLimit.Set(*window, txn, services.GetCache())

			return
		}
		receipt.ContractAddress = transaction.To

		utils.Info(fmt.Sprintf("executed contract [hash=%s, contractAddress=%s]", transaction.Hash, transaction.To))
		break
	default:
		utils.Error(fmt.Sprintf("invalid transaction type [hash=%s]", transaction.Hash))
		receipt.SetStatusWithNewTransaction(services.GetDb(), types.StatusInvalidTransaction)
		return
	}
	hertz = hertz * types.HertzMultiplier
	rateLimit, err := types.NewRateLimit(transaction.From, transaction.Hash,  hertz)
	if err != nil {
		utils.Error(err)
	}
	window := helper.AddHertz(txn, services.GetCache(), hertz);
	rateLimit.Set(*window, txn, services.GetCache())

	if availableHertz < hertz {
		msg := fmt.Sprintf("Account %s has a hertz balance of %d\n", fromAccount.Address, availableHertz)
		utils.Error(msg)
		receipt.SetStatusWithNewTransaction(services.GetDb(), types.StatusInsufficientHertz)
		return
	}

	//Change this to set hertz to the
	transaction.Hertz = hertz
	// Persist transaction
	err = transaction.Persist(txn)
	if err != nil {
		utils.Error(err)
		receipt.Status = types.StatusInternalError
		receipt.HumanReadableStatus = err.Error()
		receipt.Cache(services.GetCache())
		return
	}

	// Save fromAccount.
	fromAccount.Updated = now
	err = fromAccount.Persist(txn)
	if err != nil {
		utils.Error(err)
		receipt.Status = types.StatusInternalError
		receipt.HumanReadableStatus = err.Error()
		receipt.Cache(services.GetCache())
		return
	}

	if toAccount != nil {
		// Save toAccount.
		toAccount.Updated = now
		err = toAccount.Persist(txn)
		if err != nil {
			utils.Error(err)
			receipt.Status = types.StatusInternalError
			receipt.HumanReadableStatus = err.Error()
			receipt.Cache(services.GetCache())
			return
		}
		//Also lock up for the receiver
		//This code is way down here so that the rate limiting works "after" the account is saved.  New accounts don't exist until the above persist.
		//Take the lower value of Hertz for this transaction and the receivers balance (so we don't lock more than they have)
		//No longer doing this and handling it on the request balance side
		//maxToLock := math.Min(float64(toAccount.Balance.Uint64()), float64(hertz))

		rateLimitTo, err := types.NewRateLimit(transaction.To, transaction.Hash, hertz)
		if err != nil {
			utils.Error(err)
		}
		window = helper.AddHertz(txn, services.GetCache(), hertz);
		rateLimitTo.Set(*window, txn, services.GetCache())
	}

	// Save receipt.
	receipt.Status = types.StatusOk
	err = receipt.Set(txn, services.GetCache())
	if err != nil {
		utils.Error(err)
		receipt.Status = types.StatusInternalError
		receipt.HumanReadableStatus = err.Error()
		receipt.Cache(services.GetCache())
		return
	}

	// Save gossip.
	err = gossip.Set(txn, services.GetCache())
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
				if dvmResult.ContractMethodExecResult != nil && len(dvmResult.ContractMethodExecResult) > 0 {
					marshalledValues, err := method.Outputs.UnpackValues(dvmResult.ContractMethodExecResult)
					if err == nil {
						utils.Info(fmt.Sprintf("CONTRACT-CALL-RES: %v", marshalledValues))
						receipt.ContractResult = marshalledValues
					} else {
						errorToReturn = err
						utils.Error(err)
					}
					for i, arg := range method.Outputs {
						if arg.Type.T == abi.BytesTy {
							valBytes := receipt.ContractResult[i].([]byte)
							base64Bytes := make([]byte, base64.StdEncoding.DecodedLen(len(valBytes)))
							_, valErr := base64.StdEncoding.Decode(base64Bytes, valBytes)
							utils.Info(fmt.Sprintf("byteString = %v and base64Text = %v", string(valBytes), string(base64Bytes)))
							if valErr != nil {
								utils.Error(valErr)
							}
							base64Bytes = bytes.Trim(base64Bytes, "\x00")
							receipt.ContractResult[i] = string(base64Bytes)
						}
					}
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

func getAccountFromBadgerByAddress(address string) (*types.Account, error) {
	utils.Debug(fmt.Sprintf("toAccountByAddress: %s", address))

	txn := services.NewTxn(true)
	defer txn.Discard()

	account, err := types.ToAccountByAddress(txn, address)
	if err != nil {
		utils.Error(fmt.Sprintf("toAccountByAddress: %v", err))
		return nil, err
	}

	return account, nil
}
