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
package vmstatehelperimplemtations

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
	"github.com/dispatchlabs/disgo/dvm/vmstatehelpercontracts"
	"encoding/hex"
)

var (
	// chainID            = big.NewInt(1)
	GasLimit           = big.NewInt(1000000000000)
	txMetaSuffix       = []byte{0x01}
	ReceiptsPrefix     = []byte("receipts-")
	headTxKey          = []byte("LastTx")
	acctounStatePrefix = []byte("AccountState-")
	MIPMapLevels       = []uint64{1000000, 500000, 100000, 50000, 1000}
	IsDemo             = false

	DefaultValue    = big.NewInt(0)
	DefaultGas      = big.NewInt(5000000)
	DefaultGasPrice = big.NewInt(0)
	DefaultGasLimit = 5000000
	DefaultDivvy    = int64(0)
)

// VMStateHelper - Helps load and save Smart Contract storage state
type VMStateHelper struct {
	db                   ethdb.Database       // Storage - like disk storage
	EthStateDB           *ethState.StateDB    // Particia Merkle Trie
	TxIndex              int                  // TODO: is it used ?
	Transactions         []*types.Transaction // TXes executed
	Receipts             []*ethTypes.Receipt  // Recepts per TX
	AllLogs              []*ethTypes.Log      // VM opcodes execetion logs
	TotalUsedGas         *big.Int             // $$$ used to execute the opcodes and such
	GP                   *ethereum.GasPool    // TODO: what is this ?
	SmartContractAddress crypto.AddressBytes  // Smart Contract

	HashOfTrieRootNode crypto.HashBytes
}

// NewVMStateHelper - loads (if any) and returns the state for a Smart Contract
func NewVMStateHelper(smartContractAddress crypto.AddressBytes) (*VMStateHelper, error) {
	utils.Debug(fmt.Sprintf("NewVMStateHelper-CONTRACT: %s", crypto.Encode(smartContractAddress[:])))
	// debug.PrintStack()

	badgerWrapper, _ := badgerwrapper.NewBadgerDatabase()

	vmStateHelper := &VMStateHelper{
		db:                   badgerWrapper,                                   //
		EthStateDB:           nil,                                             // will be set in `initState`
		TxIndex:              0,                                               // TODO: is it used ?
		TotalUsedGas:         big.NewInt(0),                                   // TODO: is it used ?
		GP:                   new(ethereum.GasPool).AddGas(GasLimit.Uint64()), // TODO: is it used ?
		SmartContractAddress: smartContractAddress,                            //
	}

	if err := vmStateHelper.initOrLoadState(); err != nil {
		return nil, err
	}

	return vmStateHelper, nil
}

// Commit - Writes all the changes to the actual storage (aka Badger)
func (stateHelper *VMStateHelper) Commit() (crypto.HashBytes, error) {
	utils.Debug(fmt.Sprintf("VMStateHelper-Commit-CONTRACT    : %s", crypto.Encode(stateHelper.SmartContractAddress[:])))
	utils.Debug(fmt.Sprintf("VMStateHelper-Commit-TrieRootNode: %s", crypto.Encode(stateHelper.HashOfTrieRootNode[:])))

	// CLEAN up all flags and COMMIT all `Trie` changes to the memory
	var err error
	stateHelper.HashOfTrieRootNode, err = stateHelper.EthStateDB.Commit(false)
	if err != nil {
		utils.Error(fmt.Sprintf("VMStateHelper-Commit: %s", err))
		return crypto.HashBytes{}, err
	}

	// Write all changes to the Physical DB to persist the state
	err = stateHelper.EthStateDB.Database().TrieDB().Commit(stateHelper.HashOfTrieRootNode, true)
	if err != nil {
		utils.Error(fmt.Sprintf("VMStateHelper-Commit: %s", err))
		return crypto.HashBytes{}, err
	}

	// STORE STATE mapping as [FROM_TO]->[TrieRootHash] pair
	smartContractAddress := hex.EncodeToString(stateHelper.SmartContractAddress[:])
	var key = append(acctounStatePrefix, smartContractAddress...)
	// key = append(key, stateHelper.To.Bytes()...)

	utils.Debug(fmt.Sprintf("`acctounStatePrefix` is %v", crypto.Encode(acctounStatePrefix)))
	utils.Debug(fmt.Sprintf("`smartContractAddress` is %v", crypto.Encode(stateHelper.SmartContractAddress.Bytes())))

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
	utils.Debug(fmt.Sprintf("VMStateHelper-writeHead: TX count %d", len(stateHelper.Transactions)))

	headTx := &types.Transaction{}
	if len(stateHelper.Transactions) > 0 {
		headTx = stateHelper.Transactions[len(stateHelper.Transactions)-1]
	}

	utils.Debug(fmt.Sprintf("VMStateHelper-writeHead: 'LastTx' == '%v' + %v", crypto.Encode(headTxKey), crypto.Encode(crypto.GetHashBytes(headTx.Hash).Bytes())))
	return stateHelper.db.Put(headTxKey, crypto.GetHashBytes(headTx.Hash).Bytes())
}

func (stateHelper *VMStateHelper) writeTransactions() error {
	utils.Debug(fmt.Sprintf("VMStateHelper-writeTransactions: TX count %d", len(stateHelper.Transactions)))

	batch := stateHelper.db.NewBatch()

	for _, tx := range stateHelper.Transactions {
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
	utils.Debug(fmt.Sprintf("VMStateHelper-writeReceipts: TX count %d", len(stateHelper.Transactions)))

	batch := stateHelper.db.NewBatch()

	for _, receipt := range stateHelper.Receipts {
		storageReceipt := (*ethTypes.ReceiptForStorage)(receipt)
		data, err := rlp.EncodeToBytes(storageReceipt)
		if err != nil {
			return err
		}

		utils.Debug(fmt.Sprintf("receipts- [%v]", crypto.Encode(ReceiptsPrefix)))

		var key = append(ReceiptsPrefix, receipt.TxHash.Bytes()...)
		var val = data

		utils.Debug(fmt.Sprintf("VMStateHelper-writeReceipts-KEY: %v", crypto.Encode(key)))
		utils.Debug(fmt.Sprintf("VMStateHelper-writeReceipts-VAL: %v", crypto.Encode(val)))

		utils.Debug(fmt.Sprintf("VMStateHelper-writeReceipts-KEY-RAW: %v", key))
		utils.Debug(fmt.Sprintf("VMStateHelper-writeReceipts-VAL-RAW: %v", val))

		if err := batch.Put(key, data); err != nil {
			return err
		}
	}

	return batch.Write()
}

func (stateHelper *VMStateHelper) initOrLoadState() error {

	stateHelper.HashOfTrieRootNode = crypto.HashBytes{}

	// READ STATE mapping as [FROM_TO]->[TrieRootHash] pair
	smartContractAddress := hex.EncodeToString(stateHelper.SmartContractAddress[:])
	var key = append(acctounStatePrefix, smartContractAddress...)
	//var key = append(acctounStatePrefix, stateHelper.SmartContractAddress.Bytes()...)
	// key = append(key, stateHelper.To.Bytes()...)

	utils.Debug(fmt.Sprintf("`acctounStatePrefix` is %v", crypto.Encode(acctounStatePrefix)))
	utils.Debug(fmt.Sprintf("`smartContractAddress` is %v", crypto.Encode(stateHelper.SmartContractAddress.Bytes())))
	// utils.Debug(fmt.Sprintf("`to` is %v", crypto.Encode(stateHelper.To.Bytes())))

	data, err := stateHelper.db.Get(key)
	if err != nil {
		// stateHelper.HashOfTrieRootNode = crypto.HashBytes{}
	} else {
		stateHelper.HashOfTrieRootNode = crypto.BytesToHash(data)
	}

	// use root to initialise the state
	// stateHelper.EthStateDB, err = ethState.New(rootHash, ethState.NewNonCacheDatabase(stateHelper.db))
	stateHelper.EthStateDB, err = ethState.New(stateHelper.HashOfTrieRootNode, ethState.NewDatabase(stateHelper.db))

	return err
}

// VMStateQueryHelper Interface
// ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~

// GetCode - Loads a new "VMStateHelper" then does what "StateDB.GetCodeSize()" does
func (stateHelper *VMStateHelper) GetCode(smartContractAddress crypto.AddressBytes) []byte {
	// NOTE: not used yet

	return []byte{}
}

// GetCodeSize - Loads a new "VMStateHelper" then does what "StateDB.GetCodeSize()" does
func (stateHelper *VMStateHelper) GetCodeSize(executingContractAddress crypto.AddressBytes, callerAddress crypto.AddressBytes, toBeExecutedContractAddress crypto.AddressBytes) int {
	utils.Debug(fmt.Sprintf("VMStateHelper-GetCodeSize: executingContractAddress    -> %s", crypto.Encode(executingContractAddress[:])))
	utils.Debug(fmt.Sprintf("VMStateHelper-GetCodeSize: callerAddress               -> %s", crypto.Encode(callerAddress[:])))
	utils.Debug(fmt.Sprintf("VMStateHelper-GetCodeSize: toBeExecutedContractAddress -> %s", crypto.Encode(toBeExecutedContractAddress[:])))

	stateHelper, err := NewVMStateHelper(toBeExecutedContractAddress)
	if err == nil {
		return stateHelper.EthStateDB.GetCodeSize(toBeExecutedContractAddress)
	}

	return 0
}

func (stateHelper *VMStateHelper) NewEthStateLoader(smartContractAddress crypto.AddressBytes) vmstatehelpercontracts.VMStateQueryHelper {
	newStateHelper, err := NewVMStateHelper(smartContractAddress)
	if err == nil {
		return newStateHelper
	}

	return nil
}

func (stateHelper *VMStateHelper) CommitState() {
	stateHelper.Commit()
}

func (stateHelper *VMStateHelper) GetEthStateDB() *ethState.StateDB {
	return stateHelper.EthStateDB
}
