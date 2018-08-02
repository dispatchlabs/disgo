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
	"github.com/dgraph-io/badger"
	"github.com/dispatchlabs/disgo/commons/services"
	"github.com/dispatchlabs/disgo/commons/types"
	"github.com/dispatchlabs/disgo/commons/utils"
	"github.com/dispatchlabs/disgo/disgover"
)

// GetDelegateNodes
func (this *DAPoSService) GetDelegateNodes() *types.Response {

	// Find nodes.
	nodes, err := types.ToNodesByTypeFromCache(services.GetCache(), types.TypeDelegate)
	if err != nil {
		utils.Error(err)
		return types.NewResponseWithError(err)
	}

	var newnodes []*types.FakeNode //TODO: remove this
	for _, node := range nodes{
		var fakie = &types.FakeNode{
			Address: node.Address,
			Endpoint: node.GrpcEndpoint,
			GrpcEndpoint: node.GrpcEndpoint,
			HttpEndpoint: node.HttpEndpoint,
			Type: node.Type}

		newnodes = append(newnodes, fakie)
	}

	// Create response.
	response := types.NewResponse()
	response.Data = newnodes //TODO: newnodes-> nodes
	utils.Info("GetDelegateNodes")

	return response
}

// GetReceipt
func (this *DAPoSService) GetReceipt(transactionHash string) *types.Response {
	txn := services.NewTxn(false)
	defer txn.Discard()
	response := types.NewResponse()

	// Delegate?
	if disgover.GetDisGoverService().ThisNode.Type == types.TypeDelegate {
		receipt, err := types.ToReceiptFromCache(services.GetCache(), transactionHash)
		if err != nil {
			receipt, err = types.ToReceiptFromTransactionHash(txn, transactionHash)
			if err != nil {
				if err == badger.ErrKeyNotFound {
					response.Status = types.StatusNotFound
					response.HumanReadableStatus = fmt.Sprintf("unable to find receipt [hash=%s]", transactionHash)
				} else {
					response.Status = types.StatusInternalError
					response.HumanReadableStatus = err.Error()
				}
			} else {
				response.Data = receipt
			}
		} else {
			response.Data = receipt
		}
	} else {
		response.Status = types.StatusNotDelegate
		response.HumanReadableStatus = "This node is not a delegate. Please select a delegate node."
	}
	utils.Info(fmt.Sprintf("GetAccount [hash=%s, status=%s]", transactionHash, response.Status))

	return response
}

// GetAccount
func (this *DAPoSService) GetAccount(address string) *types.Response {
	txn := services.NewTxn(true)
	defer txn.Discard()
	response := types.NewResponse()

	// Delegate?
	if disgover.GetDisGoverService().ThisNode.Type == types.TypeDelegate {
		account, err := types.ToAccountByAddress(txn, address)
		if err != nil {
			if err == badger.ErrKeyNotFound {
				response.Status = types.StatusNotFound
			} else {
				response.Status = types.StatusInternalError
			}
		} else {
			response.Data = account
			response.Status = types.StatusOk
		}
	} else {
		response.Status = types.StatusNotDelegate
		response.HumanReadableStatus = "This node is not a delegate. Please select a delegate node."
	}
	utils.Info(fmt.Sprintf("GetAccount [address=%s, status=%s]", address, response.Status))

	return response
}

// NewTransaction
func (this *DAPoSService) NewTransaction(transaction *types.Transaction) *types.Response {
	response := types.NewResponse()

	// Delegate?
	if disgover.GetDisGoverService().ThisNode.Type == types.TypeDelegate {
		response = this.startGossiping(transaction)
	} else {
		response.Status = types.StatusNotDelegate
		response.HumanReadableStatus = "This node is not a delegate. Please select a delegate node."
	}

	utils.Info(fmt.Sprintf("NewTransaction [hash=%s, status=%s]", transaction.Hash, response.Status))
	return response
}

// GetTransaction
func (this *DAPoSService) GetTransaction(hash string) *types.Response {
	txn := services.NewTxn(true)
	defer txn.Discard()
	response := types.NewResponse()

	// Delegate?
	if disgover.GetDisGoverService().ThisNode.Type == types.TypeDelegate {
		account, err := types.ToTransactionByKey(txn, []byte(hash))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				response.Status = types.StatusNotFound
			} else {
				response.Status = types.StatusInternalError
			}
		} else {
			response.Data = account
			response.Status = types.StatusOk
		}
	} else {
		response.Status = types.StatusNotDelegate
		response.HumanReadableStatus = "This node is not a delegate. Please select a delegate node."
	}
	utils.Info(fmt.Sprintf("GetTransaction [hash=%s, status=%s]", hash, response.Status))

	return response
}

// GetTransactions
func (this *DAPoSService) GetTransactions() *types.Response {
	txn := services.NewTxn(true)
	defer txn.Discard()
	response := types.NewResponse()

	// Delegate?
	if disgover.GetDisGoverService().ThisNode.Type == types.TypeDelegate {
		var err error
		response.Data, err = types.ToTransactions(txn)
		if err != nil {
			response.Status = types.StatusInternalError
			response.HumanReadableStatus = err.Error()
		} else {
			response.Status = types.StatusOk
		}
	} else {
		response.Status = types.StatusNotDelegate
		response.HumanReadableStatus = "This node is not a delegate. Please select a delegate node."
	}

	utils.Info(fmt.Sprintf("GetTransactions [status=%s]", response.Status))

	return response
}

// GetTransactionsByFromAddress
func (this *DAPoSService) GetTransactionsByFromAddress(address string) *types.Response {
	txn := services.NewTxn(true)
	defer txn.Discard()
	response := types.NewResponse()

	// Delegate?
	if disgover.GetDisGoverService().ThisNode.Type == types.TypeDelegate {
		var err error
		response.Data, err = types.ToTransactionsByFromAddress(txn, address)
		if err != nil {
			response.Status = types.StatusInternalError
			response.HumanReadableStatus = err.Error()
		} else {
			response.Status = types.StatusOk
		}
	} else {
		response.Status = types.StatusNotDelegate
		response.HumanReadableStatus = "This node is not a delegate. Please select a delegate node."
	}

	utils.Info(fmt.Sprintf("GetTransactionsByFromAddress [address=%s, status=%s]", address, response.Status))

	return response
}

// GetTransactionsByToAddress
func (this *DAPoSService) GetTransactionsByToAddress(address string) *types.Response {
	txn := services.NewTxn(true)
	defer txn.Discard()
	response := types.NewResponse()

	// Delegate?
	if disgover.GetDisGoverService().ThisNode.Type == types.TypeDelegate {
		var err error
		response.Data, err = types.ToTransactionsByToAddress(txn, address)
		if err != nil {
			response.Status = types.StatusInternalError
			response.HumanReadableStatus = err.Error()
		} else {
			response.Status = types.StatusOk
		}
	} else {
		response.Status = types.StatusNotDelegate
		response.HumanReadableStatus = "This node is not a delegate. Please select a delegate node."
	}

	utils.Info(fmt.Sprintf("GetTransactionsByToAddress [address=%s, status=%s]", address, response.Status))

	return response
}

func (this *DAPoSService) DumpQueue() *types.Response {
	response := types.NewResponse()
	response.Data = this.gossipQueue.Dump()
	return response
}
