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
	"strconv"

	"github.com/dgraph-io/badger"
	"github.com/dispatchlabs/disgo/commons/services"
	"github.com/dispatchlabs/disgo/commons/types"
	"github.com/dispatchlabs/disgo/commons/utils"
	"github.com/dispatchlabs/disgo/disgover"
	"time"
	"github.com/dispatchlabs/disgo/commons/helper"
)

// GetDelegateNodes
func (this *DAPoSService) GetDelegateNodes() *types.Response {

	// Find nodes.
	cDelegates, err := types.ToNodesByTypeFromCache(services.GetCache(), types.TypeDelegate)
	if err != nil {
		utils.Error(err)
		return types.NewResponseWithError(err)
	}

	txn := services.NewTxn(false)
	defer txn.Discard()
	//get stored delegates
	sDelegates, err := types.ToNodesByType(txn, types.TypeDelegate)
	if err != nil {
		utils.Error(err)
		return types.NewResponseWithError(err)
	}

	//merge slices
	sDelegates = append(sDelegates, cDelegates...)

	//only allow unique values
	keys := make(map[string]bool)
	nodes := []*types.Node{}
	for _, entry := range sDelegates {
		if _, value := keys[entry.Address]; !value {
			keys[entry.Address] = true
			nodes = append(nodes, entry)
		}
	}


	// Create response.
	response := types.NewResponse()
	response.Data = nodes
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
			receipt, err = types.ToReceiptFromKey(txn, []byte(fmt.Sprintf("table-receipt-" +transactionHash)))
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
		response.HumanReadableStatus = types.StatusNotDelegateAsHumanReadable
	}
	utils.Info(fmt.Sprintf("GetReceipt [hash=%s, status=%s]", transactionHash, response.Status))

	return response
}

// GetAccount
func (this *DAPoSService) GetRateLimitWindow() *types.Response {
	txn := services.NewTxn(true)
	defer txn.Discard()
	response := types.NewResponse()
	epoch := time.Unix(0, types.DispatchEpoch)
	minutesSinceEpoch := int64(time.Now().Sub(epoch).Minutes())

	window, err := types.ToWindowFromKey(txn, minutesSinceEpoch)
	if err != nil {
		utils.Error(err)
	}
	if window == nil {
		window = types.NewWindow()
		helper.CalcSlopeForWindow(services.GetCache(), window)
	}
	window.TTL = types.GetCurrentTTL(*window).String()
	response.Data = window
	response.Status = types.StatusOk
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
			account.AvailableHertz, err = types.CheckMinimumAvailable(txn, services.GetCache(), account.Address, account.Balance.Uint64())
			if err != nil {
				utils.Error(err)
			}
			response.Data = account
			response.Status = types.StatusOk
		}
	} else {
		response.Status = types.StatusNotDelegate
		response.HumanReadableStatus = types.StatusNotDelegateAsHumanReadable
	}
	utils.Info(fmt.Sprintf("retrieved account [address=%s, status=%s]", address, response.Status))

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
		response.HumanReadableStatus = types.StatusNotDelegateAsHumanReadable
	}

	utils.Debug(fmt.Sprintf("new transaction [hash=%s, status=%s]", transaction.Hash, response.Status))
	return response
}

// GetTransaction
func (this *DAPoSService) GetTransaction(hash string) *types.Response {
	txn := services.NewTxn(false)
	defer txn.Discard()
	response := types.NewResponse()

	// Delegate?
	if disgover.GetDisGoverService().ThisNode.Type == types.TypeDelegate {
		transaction, err := types.ToTransactionByHash(txn, hash)
		if err != nil {
			if err == badger.ErrKeyNotFound {
				tx, _ := types.ToTransactionFromCache(services.GetCache(), hash)
				if tx != nil {
					response.Data = tx
					response.Status = types.StatusOk
				} else {
					response.Status = types.StatusNotFound
				}
			} else {
				response.Status = types.StatusInternalError
			}
		} else {
			response.Data = transaction
			response.Status = types.StatusOk
		}
	} else {
		response.Status = types.StatusNotDelegate
		response.HumanReadableStatus = types.StatusNotDelegateAsHumanReadable
	}
	utils.Debug(fmt.Sprintf("retrieved transaction [hash=%s, status=%s]", hash, response.Status))

	return response
}

// GetTransactions
func (this *DAPoSService) GetTransactions(page,size,start string) *types.Response {
	txn := services.NewTxn(true)
	defer txn.Discard()
	response := types.NewResponse()
	var err error
	pageNumber, err := strconv.Atoi(page)
	if err != nil {
		response.Status = types.StatusInternalError
		response.HumanReadableStatus = err.Error()
		return response
	}
	pageSize, err := strconv.Atoi(size)
	if err != nil {
		response.Status = types.StatusInternalError
		response.HumanReadableStatus = err.Error()
		return response
	}

	// Delegate?
	if disgover.GetDisGoverService().ThisNode.Type == types.TypeDelegate {

		response.Data, response.Paging, err = types.TransactionPaging(txn, start,pageNumber,pageSize)
		if err != nil {
			response.Status = types.StatusInternalError
			response.HumanReadableStatus = err.Error()
		} else {
			response.Status = types.StatusOk
		}
	} else {
		response.Status = types.StatusNotDelegate
		response.HumanReadableStatus = types.StatusNotDelegateAsHumanReadable
	}

	utils.Info(fmt.Sprintf("GetTransactions [status=%s]", response.Status))

	return response
}

// GetTransactionsByFromAddress
func (this *DAPoSService) GetTransactionsByFromAddress(address,page,size,start string) *types.Response {
	txn := services.NewTxn(true)
	defer txn.Discard()
	response := types.NewResponse()
	pageNumber, err := strconv.Atoi(page)
	if err != nil {
		response.Status = types.StatusInternalError
		response.HumanReadableStatus = err.Error()
		return response
	}
	pageSize, err := strconv.Atoi(size)
	if err != nil {
		response.Status = types.StatusInternalError
		response.HumanReadableStatus = err.Error()
		return response
	}

	// Delegate?
	if disgover.GetDisGoverService().ThisNode.Type == types.TypeDelegate {
		response.Data, err = types.ToTransactionsByFromAddress(txn, address, start, pageNumber, pageSize)
		if err != nil {
			response.Status = types.StatusInternalError
			response.HumanReadableStatus = err.Error()
		} else {
			response.Status = types.StatusOk
		}
	} else {
		response.Status = types.StatusNotDelegate
		response.HumanReadableStatus = types.StatusNotDelegateAsHumanReadable
	}

	utils.Info(fmt.Sprintf("retrieved transactions by from address [address=%s, status=%s]", address, response.Status))

	return response
}

// GetTransactionsByToAddress
func (this *DAPoSService) GetTransactionsByToAddress(address,page,size,start string ) *types.Response {
	txn := services.NewTxn(true)
	defer txn.Discard()
	response := types.NewResponse()
	pageNumber, err := strconv.Atoi(page)
	if err != nil {
		response.Status = types.StatusInternalError
		response.HumanReadableStatus = err.Error()
		return response
	}
	pageSize, err := strconv.Atoi(size)
	if err != nil {
		response.Status = types.StatusInternalError
		response.HumanReadableStatus = err.Error()
		return response
	}

	// Delegate?
	if disgover.GetDisGoverService().ThisNode.Type == types.TypeDelegate {
		response.Data, err = types.ToTransactionsByToAddress(txn, address, start,pageNumber,pageSize)
		if err != nil {
			response.Status = types.StatusInternalError
			response.HumanReadableStatus = err.Error()
		} else {
			response.Status = types.StatusOk
		}
	} else {
		response.Status = types.StatusNotDelegate
		response.HumanReadableStatus = types.StatusNotDelegateAsHumanReadable
	}

	utils.Info(fmt.Sprintf("retrieved transactions by to address [address=%s, status=%s]", address, response.Status))

	return response
}

func (this *DAPoSService) DumpQueue() *types.Response {
	response := types.NewResponse()
	response.Data = this.gossipQueue.Dump()
	return response
}

func (this *DAPoSService) ToBeSupported() *types.Response {
	response := types.NewResponse()
	response.Data = types.StatusUnavailableFeature
	return response
}

func (this *DAPoSService) GetAccounts(page, size, start string) *types.Response {
	txn := services.NewTxn(true)
	defer txn.Discard()
	response := types.NewResponse()
	var err error
	pageNumber, err := strconv.Atoi(page)
	if err != nil {
		response.Status = types.StatusInternalError
		response.HumanReadableStatus = err.Error()
		return response
	}
	pageSize, err := strconv.Atoi(size)
	if err != nil {
		response.Status = types.StatusInternalError
		response.HumanReadableStatus = err.Error()
		return response
	}

	// Delegate?
	if disgover.GetDisGoverService().ThisNode.Type == types.TypeDelegate {

		response.Data, err = types.AccountPaging(txn, start, pageNumber, pageSize)
		if err != nil {
			response.Status = types.StatusInternalError
			response.HumanReadableStatus = err.Error()
		} else {
			response.Status = types.StatusOk
		}
	} else {
		response.Status = types.StatusNotDelegate
		response.HumanReadableStatus = types.StatusNotDelegateAsHumanReadable
	}

	utils.Info(fmt.Sprintf("GetAccounts [status=%s]", response.Status))

	return response
}

func (this *DAPoSService) GetGossips(page string) *types.Response {
	txn := services.NewTxn(true)
	defer txn.Discard()
	response := types.NewResponse()
	var err error
	pageNumber, err := strconv.Atoi(page)
	if err != nil {
		response.Status = types.StatusInternalError
		response.HumanReadableStatus = err.Error()
		return response
	}

	// Delegate?
	if disgover.GetDisGoverService().ThisNode.Type == types.TypeDelegate {

		response.Data, err = types.GossipPaging(pageNumber, txn)
		if err != nil {
			response.Status = types.StatusInternalError
			response.HumanReadableStatus = err.Error()
		} else {
			response.Status = types.StatusOk
		}
	} else {
		response.Status = types.StatusNotDelegate
		response.HumanReadableStatus = types.StatusNotDelegateAsHumanReadable
	}

	utils.Info(fmt.Sprintf("GetGossips [status=%s]", response.Status))

	return response
}

// GetGossip
func (this *DAPoSService) GetGossip(hash string) *types.Response {
	txn := services.NewTxn(true)
	defer txn.Discard()
	response := types.NewResponse()

	// Delegate?
	if disgover.GetDisGoverService().ThisNode.Type == types.TypeDelegate {
		gossip, err := types.ToGossipByTransactionHash(txn, hash)
		if err != nil {
			if err == badger.ErrKeyNotFound {
				response.Status = types.StatusNotFound
			} else {
				response.Status = types.StatusInternalError
			}
		} else {
			response.Data = gossip
			response.Status = types.StatusOk
		}
	} else {
		response.Status = types.StatusNotDelegate
		response.HumanReadableStatus = types.StatusNotDelegateAsHumanReadable
	}
	utils.Info(fmt.Sprintf("retrieved Gossip [tx hash=%s, status=%s]", hash, response.Status))

	return response
}

