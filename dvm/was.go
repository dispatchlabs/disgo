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

	"github.com/dispatchlabs/disgo/commons/crypto"
	"github.com/dispatchlabs/disgo/commons/types"
	"github.com/dispatchlabs/disgo/commons/utils"
	"github.com/dispatchlabs/disgo/dvm/ethereum"
	"github.com/dispatchlabs/disgo/dvm/ethereum/ethdb"
	"github.com/dispatchlabs/disgo/dvm/ethereum/rlp"
	ethState "github.com/dispatchlabs/disgo/dvm/ethereum/state"
	ethTypes "github.com/dispatchlabs/disgo/dvm/ethereum/types"
)

// write ahead state, updated with each AppendTx and reset on Commit
type WriteAheadState struct {
	db         ethdb.Database
	ethStateDB *ethState.StateDB

	txIndex      int
	transactions []*types.Transaction
	receipts     []*ethTypes.Receipt
	allLogs      []*ethTypes.Log

	totalUsedGas *big.Int
	gp           *ethereum.GasPool
}

func (was *WriteAheadState) Commit() (crypto.HashBytes, error) {
	utils.Info(fmt.Sprintf("WAS-Commit"))

	//commit all state changes to the database
	hashArray, err := was.ethStateDB.Commit(false)
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
	utils.Info(fmt.Sprintf("WAS-writeHead: TX count %d", len(was.transactions)))

	head := &types.Transaction{}
	if len(was.transactions) > 0 {
		head = was.transactions[len(was.transactions)-1]
	}

	utils.Info(fmt.Sprintf("WAS-writeHead: 'LastTx' == '%v' + %v", crypto.Encode(headTxKey), head.CalculateHash()))
	return was.db.Put(headTxKey, head.CalculateHash())
}

func (was *WriteAheadState) writeTransactions() error {
	utils.Info(fmt.Sprintf("WAS-writeTransactions: TX count %d", len(was.transactions)))

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
	utils.Info(fmt.Sprintf("WAS-writeReceipts: TX count %d", len(was.transactions)))

	batch := was.db.NewBatch()

	for _, receipt := range was.receipts {
		storageReceipt := (*ethTypes.ReceiptForStorage)(receipt)
		data, err := rlp.EncodeToBytes(storageReceipt)
		if err != nil {
			return err
		}

		utils.Info(fmt.Sprintf("receipts- [%v]", crypto.Encode(receiptsPrefix)))
		if err := batch.Put(append(receiptsPrefix, receipt.TxHash.Bytes()...), data); err != nil {
			return err
		}
	}

	return batch.Write()
}

func (was *WriteAheadState) initState() error {

	rootHash := crypto.HashBytes{}

	//get head transaction hash
	headTxHash := crypto.HashBytes{}

	utils.Info(fmt.Sprintf("LastTx [%v]", crypto.Encode(headTxKey)))
	data, _ := was.db.Get(headTxKey)
	if len(data) != 0 {
		headTxHash = crypto.BytesToHash(data)
		utils.Info(fmt.Sprintf("Loading state from existing head - head_tx: %v", headTxHash.Hex()))

		// bytes, _ := hex.DecodeString(tx.Hash)
		// receipt, err := dvm.getReceipt(bytes)

		// if err != nil {
		// 	utils.Error(err)
		// }

		//get head tx receipt
		headTxReceipt, err := was.getReceipt(headTxHash)
		if err != nil {
			utils.Error(fmt.Sprintf("Head transaction receipt missing: %v", err))
			return err
		}

		//extract root from receipt
		if len(headTxReceipt.PostState) != 0 {
			rootHash = crypto.BytesToHash(headTxReceipt.PostState)
			utils.Info(fmt.Sprintf("Head transaction root: %v", rootHash.Hex()))
		}

	}

	//use root to initialise the state
	var err error
	was.ethStateDB, err = ethState.New(rootHash, ethState.NewNonCacheDatabase(was.db))
	return err

	// rootHash := crypto.HashBytes{} // TODO: load this from DB
	// var err error
	// dvmServiceInstance.statedb, err = ethState.New(
	// 	rootHash,
	// 	ethState.NewNonCacheDatabase(dvmServiceInstance.db),
	// )
	// if err != nil {
	// 	utils.Fatal(err)
	// }

}

func (was *WriteAheadState) getReceipt2(txHash []byte) (*ethTypes.Receipt, error) {
	utils.Info(fmt.Sprintf("LastTx [%v]", crypto.Encode(headTxKey)))
	data, err := was.db.Get(append(headTxKey, txHash[:]...))
	if err != nil {
		utils.Error(fmt.Sprintf("%s GetReceipt", err))

		return nil, err
	}
	var receipt ethTypes.ReceiptForStorage
	if err := rlp.DecodeBytes(data, &receipt); err != nil {
		utils.Error(fmt.Sprintf("%s Decoding Receipt", err))

		return nil, err
	}

	return (*ethTypes.Receipt)(&receipt), nil
}

func (s *WriteAheadState) getReceipt(txHash crypto.HashBytes) (*ethTypes.Receipt, error) {
	// utils.Info(fmt.Sprintf("receipts- [%v]", crypto.Encode(receiptsPrefix)))
	// data, err := s.db.Get(append(receiptsPrefix, txHash[:]...))
	data, err := s.db.Get(txHash[:])
	if err != nil {
		utils.Error(err)
		return nil, err
	}
	var receipt ethTypes.ReceiptForStorage
	if err := rlp.DecodeBytes(data, &receipt); err != nil {
		utils.Error(err)
		return nil, err
	}

	return (*ethTypes.Receipt)(&receipt), nil
}

func LoadOrInitNewState(db ethdb.Database) (*WriteAheadState, error) {
	was := &WriteAheadState{
		db: db,
		// ethState:     // will be set in `initState`
		txIndex:      0,
		totalUsedGas: big.NewInt(0),
		gp:           new(ethereum.GasPool).AddGas(gasLimit.Uint64()),
	}

	if err := was.initState(); err != nil {
		return nil, err
	}

	return was, nil
}
