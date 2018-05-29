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
	"strings"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/dispatchlabs/disgo/commons/services"
	"github.com/dispatchlabs/disgo/commons/types"
	"github.com/dispatchlabs/disgo/commons/utils"
	"github.com/dispatchlabs/disgo/disgover"
	"github.com/dispatchlabs/disgo/dvm"
	"github.com/processout/grpc-go-pool"
	"google.golang.org/grpc"
	"github.com/patrickmn/go-cache"
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

	// TODO: Check minimum hertz!!!!!

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
						clientConn, err := grpc.Dial( fmt.Sprintf("%s:%d", delegateNode.Endpoint.Host, delegateNode.Endpoint.Port), grpc.WithInsecure())
						if err != nil {
							utils.Error(err.Error())
							return nil, err
						}
						return clientConn, nil
					}, 10, 10, -1)
					if err != nil {
						utils.Error(err.Error())
					}
					services.GetCache().Set(fmt.Sprintf("dapos-grpc-pool-%s",  delegateNode.Address), pool, cache.NoExpiration)
				}
			}
			if len(gossip.Rumors) >= len(this.delegateNodes) * 2/3 {
				this.transactionChan <- gossip
			}

			// Gossip to random delegate.
			node := this.getRandomDelegate(gossip)
			if node == nil {
				continue
			}

			// Peer gossip.
			peerGossip, err := this.peerGossipGrpc(*node, gossip)
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

			// TODO: These lock will go away when we have state management implemented.
			transaction := gossip.Transaction
			//services.Lock(transaction.From)
			//defer services.Unlock(transaction.From)
			//services.Lock(transaction.To)
			//defer services.Unlock(transaction.To)
			txn := services.NewTxn(true)
			defer txn.Discard()

			// TODO: Should we check the receipt status is pending?

			// Has this transaction already been processed?
			_, err := txn.Get([]byte(transaction.Key()))
			if err == nil {
				continue
			}

			// Get receipt.
			var receipt *types.Receipt
			value, ok := services.GetCache().Get(gossip.ReceiptId)
			if !ok {
				utils.Error(fmt.Sprintf("receipt not found [id=%s]", gossip.ReceiptId))
				receipt =  types.NewReceipt(types.RequestNewTransaction)
				receipt.Status = types.StatusReceiptNotFound
				services.GetCache().Set(receipt.Id, receipt, types.ReceiptCacheTTL)
				continue
			}
			receipt = value.(*types.Receipt)

			// Verify?
			if !transaction.Verify() {
				utils.Error(fmt.Sprintf("invalid transaction [hash=%s]", transaction.Hash))
				receipt.Status = types.StatusInvalidTransaction
				services.GetCache().Set(receipt.Id, receipt, types.ReceiptCacheTTL)
				continue
			}

			if len(strings.TrimSpace(transaction.To)) == 0 &&
				len(strings.TrimSpace(transaction.Code)) != 0 {

				// DEPLOY
				dvmService := dvm.GetDVMService()
				dvmResult, err := dvmService.DeploySmartContract(&transaction)
				if err != nil {
					utils.Error(err, utils.GetCallStackWithFileAndLineNumber())
				}

				processDVMResult(dvmResult)

			} else if len(strings.TrimSpace(transaction.To)) != 0 &&
				len(strings.TrimSpace(transaction.Code)) != 0 &&
				len(strings.TrimSpace(transaction.Method)) != 0 {

				// EXECUTE
				dvmService := dvm.GetDVMService()
				dvmResult, err1 := dvmService.ExecuteSmartContract(&transaction)
				if err1 != nil {
					utils.Error(err, utils.GetCallStackWithFileAndLineNumber())
				}

				processDVMResult(dvmResult)
			} else {

				// TRANSFER $$$
				TransferTokens(&transaction, txn, receipt, gossip)
			}
		}
	}
}

// - TransferTokens is only called if it is a simple transfer and has no contract
func TransferTokens(transaction *types.Transaction, txn *badger.Txn, receipt *types.Receipt, gossip *types.Gossip) {
	// Find/create fromAccount?
	now := time.Now()
	var fromAccount *types.Account
	fromAccount, err := types.ToAccountByAddress(txn, transaction.From)
	if err != nil {
		if err == badger.ErrKeyNotFound {
			fromAccount = &types.Account{Address: transaction.From, Balance: 0, Created: now}
		} else {
			utils.Error(err)
			receipt.Status = types.StatusInternalError
			receipt.HumanReadableStatus = err.Error()
			services.GetCache().Set(receipt.Id, receipt, types.ReceiptCacheTTL)
			return
		}
	}

	// Sufficient tokens?
	if fromAccount.Balance < transaction.Value {
		utils.Error(fmt.Sprintf("insufficient tokens [hash=%s]", transaction.Hash))
		receipt.SetStatusWithNewTransaction(services.GetDb(), types.StatusInsufficientTokens)
		return
	}

	// Find/create toAccount?
	var toAccount *types.Account
	toAccount, err = types.ToAccountByAddress(txn, transaction.To)
	if err != nil {
		if err == badger.ErrKeyNotFound {
			toAccount = &types.Account{Address: transaction.To, Balance: 0, Created: now}
		} else {
			utils.Error(err)
			receipt.SetInternalErrorWithNewTransaction(services.GetDb(), err)
			return
		}
	}

	// Save accounts.
	fromAccount.Balance -= transaction.Value
	fromAccount.Updated = now
	err = fromAccount.Set(txn)
	if err != nil {
		utils.Error(err)
		receipt.Status = types.StatusInternalError
		receipt.HumanReadableStatus = err.Error()
		services.GetCache().Set(receipt.Id, receipt, types.ReceiptCacheTTL)
		return
	}
	toAccount.Balance += transaction.Value
	toAccount.Updated = now
	err = toAccount.Set(txn)
	if err != nil {
		utils.Error(err)
		receipt.Status = types.StatusInternalError
		receipt.HumanReadableStatus = err.Error()
		services.GetCache().Set(receipt.Id, receipt, types.ReceiptCacheTTL)
		return
	}

	// Save transaction.  -- actually does save to BadgerDB
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
	utils.Info(fmt.Sprintf("successful transaction [hash=%s, rumors=%d]", transaction.Hash, len(gossip.Rumors)))

}

//TODO: implement if useful
func commit(transaction *types.Transaction) {}

func processDVMResult(result *dvm.DVMResult) error {
	utils.Info("TODO: *** Not doing anything right now")
	utils.Info(result)

	return nil
}
