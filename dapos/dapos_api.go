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
func (this *DAPoSService) GetDelegateNodes() *types.Receipt {
	txn := services.NewTxn(true)
	defer txn.Discard()
	receipt := types.NewReceipt(types.RequestGetDelegates)
	receipt.Status = types.StatusOk

	// Find nodes.
	nodes, err := disgover.GetDisGoverService().FindByType(types.TypeDelegate)
	if err != nil {
		utils.Error(err)
		receipt.SetInternalErrorWithNewTransaction(services.GetDb(), err)
		return nil
	}
	receipt.Data = nodes
	err = receipt.Set(txn,services.GetCache()) // TODO: Should we store receipts?
	if err != nil {
		utils.Error(err)
		receipt.SetInternalErrorWithNewTransaction(services.GetDb(), err)
		return nil
	}
	err = txn.Commit(nil)
	if err != nil {
		utils.Error(err)
		receipt.SetInternalErrorWithNewTransaction(services.GetDb(), err)
		return nil
	}
	utils.Info(fmt.Sprintf("id=%s, type=%s, status=%s", receipt.Id, receipt.Type, receipt.Status))
	utils.Info(receipt.String())
	return receipt
}

// GetStatus
func (this *DAPoSService) GetStatus(id string) *types.Receipt {
	txn := services.NewTxn(false)
	defer txn.Discard()
	var receipt *types.Receipt

	// Delegate?
	if disgover.GetDisGoverService().ThisNode.Type == types.TypeDelegate {
		var err error
		receipt, err = types.ToReceiptFromCache(services.GetCache(),id)
		if err != nil {
			receipt, err = types.ToReceiptFromId(txn, id)
			if err != nil {
				if err == badger.ErrKeyNotFound {
					receipt = types.NewReceiptWithStatus(types.RequestGetStatus, types.StatusNotFound, fmt.Sprintf("unable to find receipt [id=%s]", id))
				} else {
					receipt = types.NewReceiptWithError(types.RequestGetStatus, err)
				}
			}
		}
	} else {
		receipt = this.peerDelegateExecuteGrpc(types.RequestGetStatus, id)
	}
	utils.Info(fmt.Sprintf("id=%s, type=%s, status=%s", receipt.Id, receipt.Type, receipt.Status))
	return receipt
}

// GetAccount
func (this *DAPoSService) GetAccount(address string) *types.Receipt {
	txn := services.NewTxn(true)
	defer txn.Discard()
	var receipt *types.Receipt

	// Delegate?
	if disgover.GetDisGoverService().ThisNode.Type == types.TypeDelegate {
		receipt = types.NewReceipt(types.RequestGetAccount)
		account, err := types.ToAccountByAddress(txn, address)
		if err != nil {
			if err == badger.ErrKeyNotFound {
				receipt.Status = types.StatusNotFound
			} else {
				receipt.Status = types.StatusInternalError
			}
		} else {
			receipt.Data = account
			receipt.Status = types.StatusOk
		}
	} else {
		receipt = this.peerDelegateExecuteGrpc(types.RequestGetAccount, address)
	}

	// Save receipt.
	err := receipt.Set(txn,services.GetCache())
	if err != nil {
		utils.Error(err)
	}
	err = txn.Commit(nil)
	if err != nil {
		utils.Error(err)
	}
	utils.Info(fmt.Sprintf("id=%s, type=%s, status=%s", receipt.Id, receipt.Type, receipt.Status))
	return receipt
}

// SetAccount
func (this *DAPoSService) SetAccount(account types.Account, hash string, signature string) *types.Receipt {
	txn := services.NewTxn(true)
	defer txn.Discard()
	var receipt *types.Receipt

	/*

			// Delegate?
			if types.GetConfig().IsDelegate {
				receipt = types.NewReceipt(types.RequestSetName)
				persistedAccount, err := types.ToAccountByAddress(txn, account.Address)
				if err != nil {
					if err == badger.ErrKeyNotFound {

					} else {
						receipt.Status = types.StatusInternalError
					}
				} else {
					persistedAccount.Name = account.Name
					err := txn.Set([]byte(persistedAccount.Key()), []byte(persistedAccount.String()))
					if err != nil {
						utils.Error(err)
					}
					receipt.Data = persistedAccount
					receipt.Status = types.StatusOk
				}
			} else {
				receipt = this.peerDelegateExecuteGrpc(types.RequestSetName, account.String())
			}
		} else {
			receipt = types.NewReceiptWithStatus(types.RequestSetName, types.StatusInvalidAddress, "invalid address")
		}

		// Save receipt.
		err := receipt.Set(txn)
		if err != nil {
			utils.Error(err)
		}
		err = txn.Commit(nil)
		if err != nil {
			utils.Error(err)
		}
		utils.Info(fmt.Sprintf("id=%s, type=%s, status=%s", receipt.Id, receipt.Type, receipt.Status))
	*/

	return receipt
}

// NewTransaction
func (this *DAPoSService) NewTransaction(transaction *types.Transaction) *types.Receipt {
	var receipt *types.Receipt

	// Delegate?
	if disgover.GetDisGoverService().ThisNode.Type == types.TypeDelegate {
		receipt = this.startGossiping(transaction)
	} else {
		receipt = this.peerDelegateExecuteGrpc(types.RequestNewTransaction, transaction.String())
	}
	utils.Info(fmt.Sprintf("id=%s, type=%s, status=%s", receipt.Id, receipt.Type, receipt.Status))
	return receipt
}

// GetTransactions
func (this *DAPoSService) GetTransactions() *types.Receipt {
	txn := services.NewTxn(true)
	defer txn.Discard()
	var receipt *types.Receipt

	// Delegate?
	if disgover.GetDisGoverService().ThisNode.Type == types.TypeDelegate {
		receipt = types.NewReceipt(types.RequestGetTransactions)
		receipt.Status = types.StatusOk
		var err error
		receipt.Data, err = types.ToTransactions(txn)
		if err != nil {
			receipt.Status = types.StatusInternalError
			receipt.HumanReadableStatus = err.Error()
		} else {
			receipt.Status = types.StatusOk
		}
	} else {
		receipt = this.peerDelegateExecuteGrpc(types.RequestGetTransactions, "")
	}

	// Save receipt.
	err := receipt.Set(txn,services.GetCache())
	if err != nil {
		utils.Error(err)
	}
	err = txn.Commit(nil)
	if err != nil {
		utils.Error(err)
	}
	utils.Info(fmt.Sprintf("id=%s, type=%s, status=%s", receipt.Id, receipt.Type, receipt.Status))
	return receipt
}

// GetTransactionsByFromAddress
func (this *DAPoSService) GetTransactionsByFromAddress(address string) *types.Receipt {
	txn := services.NewTxn(true)
	defer txn.Discard()
	var receipt *types.Receipt

	// Delegate?
	if disgover.GetDisGoverService().ThisNode.Type == types.TypeDelegate {
		receipt = types.NewReceipt(types.RequestGetTransactionsByFromAddress)
		receipt.Status = types.StatusOk
		var err error
		receipt.Data, err = types.ToTransactionsByFromAddress(txn, address)
		if err != nil {
			receipt.Status = types.StatusInternalError
			receipt.HumanReadableStatus = err.Error()
		} else {
			receipt.Status = types.StatusOk
		}
	} else {
		receipt = this.peerDelegateExecuteGrpc(types.RequestGetTransactionsByFromAddress, address)
	}

	// Save receipt.
	err := receipt.Set(txn,services.GetCache())
	if err != nil {
		utils.Error(err)
	}
	err = txn.Commit(nil)
	if err != nil {
		utils.Error(err)
	}
	utils.Info(fmt.Sprintf("id=%s, type=%s, status=%s", receipt.Id, receipt.Type, receipt.Status))
	return receipt
}

// GetTransactionsByToAddress
func (this *DAPoSService) GetTransactionsByToAddress(address string) *types.Receipt {
	txn := services.NewTxn(true)
	defer txn.Discard()
	var receipt *types.Receipt

	// Delegate?
	if disgover.GetDisGoverService().ThisNode.Type == types.TypeDelegate {
		receipt = types.NewReceipt(types.RequestGetTransactionsByToAddress)
		receipt.Status = types.StatusOk
		var err error
		receipt.Data, err = types.ToTransactionsByToAddress(txn, address)
		if err != nil {
			receipt.Status = types.StatusInternalError
			receipt.HumanReadableStatus = err.Error()
		} else {
			receipt.Status = types.StatusOk
		}
	} else {
		receipt = this.peerDelegateExecuteGrpc(types.RequestGetTransactionsByToAddress, address)
	}

	// Save receipt.
	err := receipt.Set(txn,services.GetCache())
	if err != nil {
		utils.Error(err)
	}
	err = txn.Commit(nil)
	if err != nil {
		utils.Error(err)
	}
	utils.Info(fmt.Sprintf("id=%s, type=%s, status=%s", receipt.Id, receipt.Type, receipt.Status))
	return receipt
}
