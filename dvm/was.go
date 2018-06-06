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
	db         ethdb.Database    // Storate
	ethStateDB *ethState.StateDB // Trie aka Merkle

	txIndex      int
	transactions []*types.Transaction
	receipts     []*ethTypes.Receipt
	allLogs      []*ethTypes.Log

	totalUsedGas *big.Int
	gp           *ethereum.GasPool

	account *types.Account
}

func (was *WriteAheadState) Commit() (crypto.HashBytes, error) {
	utils.Info(fmt.Sprintf("WAS-Commit"))

	// Commit all state changes to the database
	hashOfTrieRootNode, err := was.ethStateDB.Commit(false)

	was.ethStateDB.Database().TrieDB().Commit(hashOfTrieRootNode, true)

	// STORE STATE per account
	accuntAddressAsBytes := crypto.GetAddressBytes(was.account.Address).Bytes()
	was.db.Put(append(acctounStatePrefix, accuntAddressAsBytes...), hashOfTrieRootNode.Bytes())

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

	return hashOfTrieRootNode, nil
}

func (was *WriteAheadState) writeHead() error {
	utils.Info(fmt.Sprintf("WAS-writeHead: TX count %d", len(was.transactions)))

	headTx := &types.Transaction{}
	if len(was.transactions) > 0 {
		headTx = was.transactions[len(was.transactions)-1]
	}

	utils.Info(fmt.Sprintf("WAS-writeHead: 'LastTx' == '%v' + %v", crypto.Encode(headTxKey), crypto.Encode(crypto.GetHashBytes(headTx.Hash).Bytes())))
	return was.db.Put(headTxKey, crypto.GetHashBytes(headTx.Hash).Bytes())
}

func (was *WriteAheadState) writeTransactions() error {
	utils.Info(fmt.Sprintf("WAS-writeTransactions: TX count %d", len(was.transactions)))

	batch := was.db.NewBatch()

	for _, tx := range was.transactions {
		data, err := tx.MarshalJSON()
		if err != nil {
			return err
		}
		if err := batch.Put(crypto.GetHashBytes(tx.Hash).Bytes(), data); err != nil {
			return err
		}
	}

	// Write the scheduled data into the database
	return batch.Write()
}

func (was *WriteAheadState) writeReceipts() error { // hashOfTrieRootNode crypto.HashBytes
	utils.Info(fmt.Sprintf("WAS-writeReceipts: TX count %d", len(was.transactions)))

	batch := was.db.NewBatch()

	for _, receipt := range was.receipts {
		storageReceipt := (*ethTypes.ReceiptForStorage)(receipt)
		data, err := rlp.EncodeToBytes(storageReceipt)
		if err != nil {
			return err
		}

		utils.Info(fmt.Sprintf("receipts- [%v]", crypto.Encode(receiptsPrefix)))

		var key = append(receiptsPrefix, receipt.TxHash.Bytes()...)
		var val = data

		utils.Info(fmt.Sprintf("WAS-writeReceipts-KEY: %v", crypto.Encode(key)))
		utils.Info(fmt.Sprintf("WAS-writeReceipts-VAL: %v", crypto.Encode(val)))

		utils.Info(fmt.Sprintf("WAS-writeReceipts-KEY-RAW: %v", key))
		utils.Info(fmt.Sprintf("WAS-writeReceipts-VAL-RAW: %v", val))

		if err := batch.Put(key, data); err != nil {
			return err
		}
	}

	return batch.Write()
}

func (was *WriteAheadState) initState() error {

	hashOfTrieRootNode := crypto.HashBytes{}

	// READ STATE per account
	utils.Info(fmt.Sprintf("acctounStatePrefix [%v]", crypto.Encode(acctounStatePrefix)))
	accuntAddressAsBytes := crypto.GetAddressBytes(was.account.Address).Bytes()
	data, err := was.db.Get(append(acctounStatePrefix, accuntAddressAsBytes...))
	if err != nil {
		// hashOfTrieRootNode = crypto.HashBytes{}
	} else {
		hashOfTrieRootNode = crypto.BytesToHash(data)
	}

	// use root to initialise the state
	// was.ethStateDB, err = ethState.New(rootHash, ethState.NewNonCacheDatabase(was.db))
	was.ethStateDB, err = ethState.New(hashOfTrieRootNode, ethState.NewDatabase(was.db))

	return err
}

/*
func (s *WriteAheadState) getReceipt(txHash crypto.HashBytes) (*ethTypes.Receipt, error) {
	utils.Info(fmt.Sprintf("receipts- [%v]", crypto.Encode(receiptsPrefix)))
	data, err := s.db.Get(append(receiptsPrefix, txHash[:]...))
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
*/

func LoadOrInitNewState(db ethdb.Database, account *types.Account) (*WriteAheadState, error) {
	was := &WriteAheadState{
		db: db,
		// ethState:     // will be set in `initState`
		txIndex:      0,
		totalUsedGas: big.NewInt(0),
		gp:           new(ethereum.GasPool).AddGas(gasLimit.Uint64()),
		account:      account,
	}

	if err := was.initState(); err != nil {
		return nil, err
	}

	return was, nil
}
