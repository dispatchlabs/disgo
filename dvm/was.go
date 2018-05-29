/*
 *    This file is part of DVM library.
 *
 *    The DVM library is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *
 *    The DVM library is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *
 *    You should have received a copy of the GNU General Public License
 *    along with the DVM library.  If not, see <http://www.gnu.org/licenses/>.
*/
package dvm

import (
	"fmt"
	"math/big"

	"github.com/dispatchlabs/commons/crypto"
	"github.com/dispatchlabs/commons/types"
	"github.com/dispatchlabs/commons/utils"
	"github.com/dispatchlabs/dvm/ethereum"
	"github.com/dispatchlabs/dvm/ethereum/ethdb"
	"github.com/dispatchlabs/dvm/ethereum/rlp"
	ethState "github.com/dispatchlabs/dvm/ethereum/state"
	ethTypes "github.com/dispatchlabs/dvm/ethereum/types"
)

// write ahead state, updated with each AppendTx and reset on Commit
type WriteAheadState struct {
	db       ethdb.Database
	ethState *ethState.StateDB

	txIndex      int
	transactions []*types.Transaction
	receipts     []*ethTypes.Receipt
	allLogs      []*ethTypes.Log

	totalUsedGas *big.Int
	gp           *ethereum.GasPool
}

func (was *WriteAheadState) Commit() (crypto.HashBytes, error) {
	// utils.Info("TRACE")

	//commit all state changes to the database
	hashArray, err := was.ethState.Commit(false)
	if err != nil {
		utils.Error(fmt.Sprintf("%s Committing state", err))
		return crypto.HashBytes{}, err
	}
	if err := was.writeHead(); err != nil {
		utils.Error(fmt.Sprintf("%s Writing head", err))

		return crypto.HashBytes{}, err
	}
	if err := was.writeTransactions(); err != nil {
		utils.Error(fmt.Sprintf("%s Writing txsd", err))
		return crypto.HashBytes{}, err
	}
	if err := was.writeReceipts(); err != nil {
		utils.Error(fmt.Sprintf("%s Writing receipts", err))
		return crypto.HashBytes{}, err
	}
	return hashArray, nil
}

func (was *WriteAheadState) writeHead() error {
	// utils.Info(fmt.Sprintf("TRACE: WAS len(was.transactions) = %d", len(was.transactions)))

	head := &types.Transaction{}
	if len(was.transactions) > 0 {
		head = was.transactions[len(was.transactions)-1]
	}
	return was.db.Put(headTxKey, head.CalculateHash())
}

func (was *WriteAheadState) writeTransactions() error {
	// utils.Info(fmt.Sprintf("TRACE: WAS len(was.transactions) = %d", len(was.transactions)))

	batch := was.db.NewBatch()

	for _, tx := range was.transactions {
		data, err := tx.MarshalJSON()
		if err != nil {
			return err
		}
		if err := batch.Put(tx.CalculateHash(), data); err != nil {
			return err
		}
	}

	// Write the scheduled data into the database
	return batch.Write()
}

func (was *WriteAheadState) writeReceipts() error {
	// utils.Info(fmt.Sprintf("TRACE: WAS len(was.receipts) = %d", len(was.transactions)))

	batch := was.db.NewBatch()

	for _, receipt := range was.receipts {
		storageReceipt := (*ethTypes.ReceiptForStorage)(receipt)
		data, err := rlp.EncodeToBytes(storageReceipt)
		if err != nil {
			return err
		}
		if err := batch.Put(append(receiptsPrefix, receipt.TxHash.Bytes()...), data); err != nil {
			return err
		}
	}

	return batch.Write()
}
