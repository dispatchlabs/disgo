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
	"github.com/patrickmn/go-cache"
	"github.com/processout/grpc-go-pool"
	"google.golang.org/grpc"

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
	_, ok := services.GetCache().Get(transaction.Hash)
	if ok {
		utils.Info(fmt.Sprintf("already processing this transaction [hash=%s]", transaction.Hash))
		receipt.Status = types.StatusAlreadyProcessingTransaction
		return receipt
	}

	// Cache receipt.
	services.GetCache().Set(receipt.Id, receipt, types.ReceiptCacheTTL)

	// Cache gossip with my rumor.
	gossip := types.NewGossip(*transaction, *receipt)
	rumor := types.NewRumor(types.GetAccount().PrivateKey, types.GetAccount().Address, transaction.Hash)
	gossip.Rumors = append(gossip.Rumors, *rumor)
	services.GetCache().Set(gossip.Transaction.Hash, gossip, types.GossipCacheTTL)

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
	_, ok := services.GetCache().Get(gossip.ReceiptId)
	if !ok {
		receipt := types.NewReceipt(types.RequestNewTransaction)
		receipt.Id = gossip.ReceiptId
		services.GetCache().Set(receipt.Id, receipt, types.ReceiptCacheTTL)
	}

	// Set synchronizedGossip.
	var synchronizedGossip *types.Gossip
	value, ok := services.GetCache().Get(gossip.Transaction.Hash)
	if !ok {
		synchronizedGossip = gossip
	} else {
		synchronizedGossip = value.(*types.Gossip)
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

			// TODO: The following code should be executed during elections!!!!
			if len(this.delegateNodes) == 0 {
				var err error
				this.delegateNodes, err = disgover.GetDisGoverService().FindByType(types.TypeDelegate)
				if err != nil {
					utils.Error(err)
					continue
				}

				// Create delegate connection pools.
				for _, delegateNode := range this.delegateNodes {
					if delegateNode.Address == disgover.GetDisGoverService().ThisNode.Address {
						continue
					}
					pool, err := grpcpool.New(func() (*grpc.ClientConn, error) {
						clientConn, err := grpc.Dial(fmt.Sprintf("%s:%d", delegateNode.Endpoint.Host, delegateNode.Endpoint.Port), grpc.WithInsecure())
						if err != nil {
							utils.Error(err.Error())
							return nil, err
						}
						return clientConn, nil
					}, 10, 10, -1)
					if err != nil {
						utils.Error(err.Error())
					}
					services.GetCache().Set(fmt.Sprintf("dapos-grpc-pool-%s", delegateNode.Address), pool, cache.NoExpiration)
				}
			}
			if len(gossip.Rumors) >= len(this.delegateNodes)*2/3 {
				this.transactionChan <- gossip
			}

			// Gossip to random delegate.
			node := this.getRandomDelegate(gossip)
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
		}
	}
}

// getRandomDelegate
func (this *DAPoSService) getRandomDelegate(gossip *types.Gossip) *types.Node {
	if len(this.delegateNodes) == 0 {
		return nil
	}

	// Get delegates that have not rumored?
	delegatesNotRumored := make([]*types.Node, 0)
	for _, node := range this.delegateNodes {
		if gossip.ContainsRumor(node.Address) || node.Address == disgover.GetDisGoverService().ThisNode.Address {
			continue
		}
		delegatesNotRumored = append(delegatesNotRumored, node)
	}
	if len(delegatesNotRumored) == 0 {
		return nil
	}

	// Find random delegate.
	index := rand.Intn(len(delegatesNotRumored))
	return delegatesNotRumored[index]
}

// gossipWorker - transfer tokens, deploy smart contract, and execution of smart contract.
func (this *DAPoSService) transactionWorker() {

	var gossip *types.Gossip
	for {
		select {
		case gossip = <-this.transactionChan:

			// Get receipt.
			var receipt *types.Receipt
			value, ok := services.GetCache().Get(gossip.ReceiptId)
			if !ok {
				utils.Error(fmt.Sprintf("receipt not found [id=%s]", gossip.ReceiptId))
				receipt = types.NewReceipt(types.RequestNewTransaction)
				receipt.Status = types.StatusReceiptNotFound
				services.GetCache().Set(receipt.Id, receipt, types.ReceiptCacheTTL)
				continue
			}
			receipt = value.(*types.Receipt)

			// TODO: Should we thread this?
			// Execute.
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

	// Has this transaction already been processed?
	_, err := txn.Get([]byte(transaction.Key()))
	if err == nil {
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
			services.GetCache().Set(receipt.Id, receipt, types.ReceiptCacheTTL)
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

	// Execute.
	if len(strings.TrimSpace(transaction.To)) == 0 && len(strings.TrimSpace(transaction.Code)) != 0 {
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
			services.GetCache().Set(receipt.Id, receipt, types.ReceiptCacheTTL)
			return
		}

		// Set contract account.
		contractAccount := &types.Account{Address: hex.EncodeToString(dvmResult.ContractAddress[:]), Balance: big.NewInt(0), Updated: now, Created: now}
		err = contractAccount.Set(txn)
		if err != nil {
			utils.Error(err)
			receipt.Status = types.StatusInternalError
			receipt.HumanReadableStatus = err.Error()
			services.GetCache().Set(receipt.Id, receipt, types.ReceiptCacheTTL)
			return
		}
		receipt.ContractAddress = contractAccount.Address
		utils.Info(fmt.Sprintf("deployed contract [receiptId=%s hash=%s, contractAddress=%s]", receipt.Id, transaction.Hash, contractAccount.Address))
	} else if len(strings.TrimSpace(transaction.To)) != 0 && len(strings.TrimSpace(transaction.Abi)) != 0 && len(strings.TrimSpace(transaction.Method)) != 0 {
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
			services.GetCache().Set(receipt.Id, receipt, types.ReceiptCacheTTL)
			return
		}
		receipt.ContractAddress = transaction.To
		utils.Info(fmt.Sprintf("executed contract [receiptId=%s hash=%s, contractAddress=%s]", receipt.Id, transaction.Hash, transaction.To))
	} else {
		// Sufficient tokens?
		if fromAccount.Balance.Int64() < transaction.Value {
			utils.Error(fmt.Sprintf("insufficient tokens [hash=%s]", transaction.Hash))
			receipt.SetStatusWithNewTransaction(services.GetDb(), types.StatusInsufficientTokens)
			return
		}
		fromAccount.Balance.SetInt64(fromAccount.Balance.Int64() - transaction.Value)
		fromAccount.Balance.SetInt64(fromAccount.Balance.Int64() + transaction.Value)
		utils.Info(fmt.Sprintf("transferred tokens [receiptId=%s hash=%s, rumors=%d]", receipt.Id, transaction.Hash, len(gossip.Rumors)))
	}

	// Save fromAccount.
	fromAccount.Updated = now
	err = fromAccount.Set(txn)
	if err != nil {
		utils.Error(err)
		receipt.Status = types.StatusInternalError
		receipt.HumanReadableStatus = err.Error()
		services.GetCache().Set(receipt.Id, receipt, types.ReceiptCacheTTL)
		return
	}

	// Save toAccount.
	toAccount.Updated = now
	err = toAccount.Set(txn)
	if err != nil {
		utils.Error(err)
		receipt.Status = types.StatusInternalError
		receipt.HumanReadableStatus = err.Error()
		services.GetCache().Set(receipt.Id, receipt, types.ReceiptCacheTTL)
		return
	}

	// Save transaction.
	err = transaction.Set(txn)
	if err != nil {
		utils.Error(err)
		receipt.Status = types.StatusInternalError
		receipt.HumanReadableStatus = err.Error()
		services.GetCache().Set(receipt.Id, receipt, types.ReceiptCacheTTL)
		return
	}

	// Save receipt.
	receipt.Status = types.StatusOk
	services.GetCache().Set(receipt.Id, receipt, types.ReceiptCacheTTL)
	err = receipt.Set(txn)
	if err != nil {
		utils.Error(err)
		receipt.Status = types.StatusInternalError
		receipt.HumanReadableStatus = err.Error()
		services.GetCache().Set(receipt.Id, receipt, types.ReceiptCacheTTL)
		return
	}

	// Save gossip.
	err = gossip.Set(txn)
	if err != nil {
		utils.Error(err)
		receipt.Status = types.StatusInternalError
		receipt.HumanReadableStatus = err.Error()
		services.GetCache().Set(receipt.Id, receipt, types.ReceiptCacheTTL)
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
		services.GetCache().Set(receipt.Id, receipt, types.ReceiptCacheTTL)
		return
	}
}

// processDVMResult
func processDVMResult(transaction *types.Transaction, dvmResult *dvm.DVMResult, receipt *types.Receipt) error {
	utils.Info("######### DUMPING-DVMResult #########")
	utils.Info(dvmResult)

	if dvmResult.ContractMethodExecError != nil {
		utils.Error(dvmResult.ContractMethodExecError)
		return dvmResult.ContractMethodExecError
	}

	// Try read the execution result
	if len(strings.TrimSpace(dvmResult.ABI)) > 0 {
		fromHexAsByteArray, _ := hex.DecodeString(dvmResult.ABI)
		abiAsString := string(fromHexAsByteArray)
		jsonABI, err := abi.JSON(strings.NewReader(abiAsString))
		if err == nil {
			var parsedRes interface{}
			err = jsonABI.Unpack(&parsedRes, transaction.Method, dvmResult.ContractMethodExecResult)
			if err == nil {
				utils.Info(fmt.Sprintf("CONTRACT-CALL-RES: %s", parsedRes))
				receipt.ContractResult = parsedRes
			} else {
				utils.Error(err)
			}
		} else {
			utils.Error(err)
		}
	}
	return nil
}
