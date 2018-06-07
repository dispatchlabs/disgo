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
	"github.com/dispatchlabs/disgo/dvm/badgerwrapper"
	"github.com/dispatchlabs/disgo/dvm/ethereum"
	"github.com/dispatchlabs/disgo/dvm/ethereum/ethdb"
	"github.com/dispatchlabs/disgo/dvm/ethereum/rlp"
	ethState "github.com/dispatchlabs/disgo/dvm/ethereum/state"
	ethTypes "github.com/dispatchlabs/disgo/dvm/ethereum/types"
)

// VMStateHelper - Helps load and save Smart Contract storage state
type VMStateHelper struct {
	db           ethdb.Database       // Storage - like disk storage
	ethStateDB   *ethState.StateDB    // Particia Merkle Trie
	txIndex      int                  // TODO: is it used ?
	transactions []*types.Transaction // TXes executed
	receipts     []*ethTypes.Receipt  // Recepts per TX
	allLogs      []*ethTypes.Log      // VM opcodes execetion logs
	totalUsedGas *big.Int             // $$$ used to execute the opcodes and such
	gp           *ethereum.GasPool    // TODO: what is this ?
	from         crypto.AddressBytes  // FROM
	to           crypto.AddressBytes  // TO

	HashOfTrieRootNode crypto.HashBytes
}

// NewVMStateHelper - loads (if any) and returns the state for a Smart Contract
func NewVMStateHelper(from, to crypto.AddressBytes) (*VMStateHelper, error) {
	badgerWrapper, _ := badgerwrapper.NewBadgerDatabase()

	vmStateHelper := &VMStateHelper{
		db:           badgerWrapper,                                   //
		ethStateDB:   nil,                                             // will be set in `initState`
		txIndex:      0,                                               // TODO: is it used ?
		totalUsedGas: big.NewInt(0),                                   // TODO: is it used ?
		gp:           new(ethereum.GasPool).AddGas(gasLimit.Uint64()), // TODO: is it used ?
		from:         from,                                            //
		to:           to,                                              //
	}

	if err := vmStateHelper.initOrLoadState(); err != nil {
		return nil, err
	}

	return vmStateHelper, nil
}

// Commit - Writes all the changes to the actual storage (aka Badger)
func (stateHelper *VMStateHelper) Commit() (crypto.HashBytes, error) {

	// CLEAN up all flags and COMMIT all `Trie` changes to the memory
	var err error
	stateHelper.HashOfTrieRootNode, err = stateHelper.ethStateDB.Commit(false)
	if err != nil {
		utils.Error(fmt.Sprintf("VMStateHelper-Commit: %s", err))
		return crypto.HashBytes{}, err
	}

	// Write all changes to the Physical DB to persist the state
	err = stateHelper.ethStateDB.Database().TrieDB().Commit(stateHelper.HashOfTrieRootNode, true)
	if err != nil {
		utils.Error(fmt.Sprintf("VMStateHelper-Commit: %s", err))
		return crypto.HashBytes{}, err
	}

	// STORE STATE mapping as [FROM_TO]->[TrieRootHash] pair
	var key = append(acctounStatePrefix, stateHelper.from.Bytes()...)
	key = append(key, stateHelper.to.Bytes()...)

	utils.Info(fmt.Sprintf("`acctounStatePrefix` is %v", crypto.Encode(acctounStatePrefix)))
	utils.Info(fmt.Sprintf("`from` is %v", crypto.Encode(stateHelper.from.Bytes())))
	utils.Info(fmt.Sprintf("`to` is %v", crypto.Encode(stateHelper.to.Bytes())))

	var val = stateHelper.HashOfTrieRootNode.Bytes()
	stateHelper.db.Put(key, val)

	// Save the THESE - need to see if needed
	if err := stateHelper.writeHead(); err != nil {
		utils.Error(fmt.Sprintf("%s Writing head", err))

		return crypto.HashBytes{}, err
	}
	if err := stateHelper.writeTransactions(); err != nil {
		utils.Error(fmt.Sprintf("%s Writing txsd", err))
		return crypto.HashBytes{}, err
	}
	if err := stateHelper.writeReceipts(); err != nil {
		utils.Error(fmt.Sprintf("%s Writing receipts", err))
		return crypto.HashBytes{}, err
	}

	return stateHelper.HashOfTrieRootNode, nil
}

func (stateHelper *VMStateHelper) writeHead() error {
	utils.Info(fmt.Sprintf("VMStateHelper-writeHead: TX count %d", len(stateHelper.transactions)))

	headTx := &types.Transaction{}
	if len(stateHelper.transactions) > 0 {
		headTx = stateHelper.transactions[len(stateHelper.transactions)-1]
	}

	utils.Info(fmt.Sprintf("VMStateHelper-writeHead: 'LastTx' == '%v' + %v", crypto.Encode(headTxKey), crypto.Encode(crypto.GetHashBytes(headTx.Hash).Bytes())))
	return stateHelper.db.Put(headTxKey, crypto.GetHashBytes(headTx.Hash).Bytes())
}

func (stateHelper *VMStateHelper) writeTransactions() error {
	utils.Info(fmt.Sprintf("VMStateHelper-writeTransactions: TX count %d", len(stateHelper.transactions)))

	batch := stateHelper.db.NewBatch()

	for _, tx := range stateHelper.transactions {
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

func (stateHelper *VMStateHelper) writeReceipts() error {
	utils.Info(fmt.Sprintf("VMStateHelper-writeReceipts: TX count %d", len(stateHelper.transactions)))

	batch := stateHelper.db.NewBatch()

	for _, receipt := range stateHelper.receipts {
		storageReceipt := (*ethTypes.ReceiptForStorage)(receipt)
		data, err := rlp.EncodeToBytes(storageReceipt)
		if err != nil {
			return err
		}

		utils.Info(fmt.Sprintf("receipts- [%v]", crypto.Encode(receiptsPrefix)))

		var key = append(receiptsPrefix, receipt.TxHash.Bytes()...)
		var val = data

		utils.Info(fmt.Sprintf("VMStateHelper-writeReceipts-KEY: %v", crypto.Encode(key)))
		utils.Info(fmt.Sprintf("VMStateHelper-writeReceipts-VAL: %v", crypto.Encode(val)))

		utils.Info(fmt.Sprintf("VMStateHelper-writeReceipts-KEY-RAW: %v", key))
		utils.Info(fmt.Sprintf("VMStateHelper-writeReceipts-VAL-RAW: %v", val))

		if err := batch.Put(key, data); err != nil {
			return err
		}
	}

	return batch.Write()
}

func (stateHelper *VMStateHelper) initOrLoadState() error {

	stateHelper.HashOfTrieRootNode = crypto.HashBytes{}

	// READ STATE mapping as [FROM_TO]->[TrieRootHash] pair
	var key = append(acctounStatePrefix, stateHelper.from.Bytes()...)
	key = append(key, stateHelper.to.Bytes()...)

	utils.Info(fmt.Sprintf("`acctounStatePrefix` is %v", crypto.Encode(acctounStatePrefix)))
	utils.Info(fmt.Sprintf("`from` is %v", crypto.Encode(stateHelper.from.Bytes())))
	utils.Info(fmt.Sprintf("`to` is %v", crypto.Encode(stateHelper.to.Bytes())))

	data, err := stateHelper.db.Get(key)
	if err != nil {
		// stateHelper.HashOfTrieRootNode = crypto.HashBytes{}
	} else {
		stateHelper.HashOfTrieRootNode = crypto.BytesToHash(data)
	}

	// use root to initialise the state
	// stateHelper.ethStateDB, err = ethState.New(rootHash, ethState.NewNonCacheDatabase(stateHelper.db))
	stateHelper.ethStateDB, err = ethState.New(stateHelper.HashOfTrieRootNode, ethState.NewDatabase(stateHelper.db))

	return err
}
