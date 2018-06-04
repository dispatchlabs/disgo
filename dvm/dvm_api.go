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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/dispatchlabs/disgo/commons/crypto"
	commonTypes "github.com/dispatchlabs/disgo/commons/types"
	"github.com/dispatchlabs/disgo/commons/utils"
	"github.com/dispatchlabs/disgo/dvm/ethereum"
	"github.com/dispatchlabs/disgo/dvm/ethereum/abi"
	dvmCrypto "github.com/dispatchlabs/disgo/dvm/ethereum/crypto"
	"github.com/dispatchlabs/disgo/dvm/ethereum/params"
	"github.com/dispatchlabs/disgo/dvm/ethereum/rlp"
	ethTypes "github.com/dispatchlabs/disgo/dvm/ethereum/types"
	"github.com/dispatchlabs/disgo/dvm/ethereum/vm"
	"github.com/dispatchlabs/disgo/dvm/ethereum/trie"
)

var (
	chainID        = big.NewInt(1)
	gasLimit       = big.NewInt(1000000000)
	txMetaSuffix   = []byte{0x01}
	receiptsPrefix = []byte("receipts-")
	MIPMapLevels   = []uint64{1000000, 500000, 100000, 50000, 1000}
	headTxKey      = []byte("LastTx")
	isDemo         = false

	_defaultValue    = big.NewInt(0)
	_defaultGas      = big.NewInt(1000000)
	_defaultGasPrice = big.NewInt(0)
	_defaultGasLimit = 1000000000
	_defaultDivvy    = int64(0)
)

func (dvm *DVMService) DeploySmartContract(tx *commonTypes.Transaction) (*DVMResult, error) {
	if err := dvm.applyTransaction(tx); err != nil {
		return nil, err
	}

	_, err := dvm.commit() // hash
	if err != nil {
		utils.Error(err)
	}

	bytes, _ := hex.DecodeString(tx.Hash)
	receipt, err := dvm.getReceipt(bytes)

	if err != nil {
		utils.Fatal(err)
	}
	return &DVMResult{
		From:                crypto.GetAddressBytes(tx.From),
		To:                  crypto.AddressBytes{},
		Divvy:               _defaultDivvy,
		ContractAddress:     receipt.ContractAddress,
		Status:              receipt.Status,
		HertzCost:           receipt.GasUsed,
		CumulativeHertzUsed: receipt.CumulativeGasUsed,
		Bloom:               receipt.Bloom,
		Logs:                receipt.Logs,
	}, nil
}

func (dvm *DVMService) ExecuteSmartContract(tx *commonTypes.Transaction) (*DVMResult, error) {
	// var expected = big.NewInt(tx.Params)

	fromHex, _ := hex.DecodeString(tx.Abi)
	codeAsString := string(fromHex)
	//jsonABI, err := abi.JSON(strings.NewReader(tx.Abi))
	jsonABI, err := abi.JSON(strings.NewReader(codeAsString))
	// jsonABI, err := abi.JSON(strings.NewReader(tx.Code))
	if err != nil {
		utils.Error(err)
		return nil, err
	}

	// callData, err := jsonABI.Pack(tx.Method, expected)
	callData, err := jsonABI.Pack(tx.Method, tx.Params...)
	if err != nil {
		return nil, err
	}

	toAsBytes := crypto.GetAddressBytes(tx.To)
	callMsg := ethTypes.NewMessage(
		crypto.GetAddressBytes(tx.From),
		&toAsBytes,
		0, // nonce
		_defaultValue,
		_defaultGas.Uint64(),
		_defaultGasPrice,
		callData,
		false,
	)

	if err != nil {
		utils.Error(err)
		return nil, err
	}

	res, err := dvm.call(callMsg)
	if err != nil {
		utils.Error(err)
		return nil, err
	}
	var parsedRes string
	err = jsonABI.Unpack(&parsedRes, "getVar5", res)

	utils.Info(fmt.Sprintf("DEBUG-CONTRACT-CALL res: %s", parsedRes))

	// var parsedRes *big.Int
	// err = jsonABI.Unpack(&parsedRes, "test", res)
	if err != nil {
		utils.Error(err)
	}
	//utils.Info(fmt.Sprintf("parsed res: %v", parsedRes))

	// if parsedRes.Cmp(expected) != 0 {
	// 	utils.Error(fmt.Sprintf("Result should be %v, not %v", expected, parsedRes))
	// 	return nil, err
	// }

	return &DVMResult{
		From:            crypto.GetAddressBytes(tx.From),
		ContractAddress: crypto.GetAddressBytes(tx.To),
		To:              crypto.AddressBytes{},
		// Status:              receipt.Status,
		// HertzCost:           receipt.GasUsed,
		// CumulativeHertzUsed: receipt.CumulativeGasUsed,
		// Bloom:               receipt.Bloom,
		// Logs:                receipt.Logs,
	}, nil

}

func (self *DVMService) applyTransaction(tx *commonTypes.Transaction) error {
	price := big.NewInt(int64(0))

	context := vm.Context{
		CanTransfer: ethereum.CanTransfer,
		Transfer:    ethereum.Transfer,
		GetHash:     func(uint64) crypto.HashBytes { return tx.GetHashBytes() },
		// Message information
		Origin:      crypto.GetAddressBytes(tx.From),
		GasLimit:    gasLimit.Uint64(),
		GasPrice:    price,
		BlockNumber: big.NewInt(0), //the vm has a dependency on this..
	}

	//Prepare the ethState with transaction Hash so that it can be used in emitted
	//logs
	var txIndex = 0
	self.was.ethState.Prepare(tx.GetHashBytes(), tx.GetHashBytes(), txIndex)

	// The EVM should never be reused and is not thread safe.
	vmLogger := vm.NewStructLogger(&vm.LogConfig{
		DisableMemory:  false,
		DisableStack:   false,
		DisableStorage: false,
		Debug:          isDemo,
		Limit:          0,
	})

	vmenv := vm.NewEVM(
		context,
		self.was.ethState,
		&params.ChainConfig{
			ChainId: chainID,
		},
		vm.Config{
			Debug:  isDemo,
			Tracer: vmLogger,
		},
	)

	msg := ethTypes.AsMessage(tx, uint64(_defaultGasLimit))

	// Apply the transaction to the current state (included in the env)
	// GRAB-THIS: gas will be the GAS/Hertz used to execute the TX - for contract creation or execution
	_, gas, failed, err := ethereum.ApplyMessage(vmenv, msg, self.was.gp)
	if err != nil {
		utils.Error(fmt.Sprintf("%s Applying transaction to WAS", err))
		return err
	}
	self.was.totalUsedGas.Add(self.was.totalUsedGas, big.NewInt(0).SetUint64(gas))

	// Create a new receipt for the transaction, storing the intermediate root and gas used by the tx
	// based on the eip phase, we're passing wether the root touch-delete accounts.
	root := self.was.ethState.IntermediateRoot(true) //this has side effects. It updates StateObjects (SmartContract memory)

	receipt := ethTypes.NewReceipt(root.Bytes(), failed, self.was.totalUsedGas.Uint64())
	receipt.TxHash = tx.GetHashBytes()
	receipt.GasUsed = gas
	// if the transaction created a contract, store the creation address in the receipt.
	if msg.To() == nil {
		receipt.ContractAddress = dvmCrypto.CreateAddress(vmenv.Context.Origin, 0)
	}
	// Set the receipt logs and create a bloom for filtering
	receipt.Logs = self.was.ethState.GetLogs(tx.GetHashBytes())
	//receipt.Logs = s.was.state.Logs()
	receipt.Bloom = ethTypes.CreateBloom(ethTypes.Receipts{receipt})

	self.was.txIndex++
	self.was.transactions = append(self.was.transactions, tx)
	self.was.receipts = append(self.was.receipts, receipt)
	self.was.allLogs = append(self.was.allLogs, receipt.Logs...)

	utils.Info(fmt.Sprintf("%s Applied tx to WAS", tx.Hash))

	// DEMO-Today
	if isDemo {
		// logsAsJSON, _ := json.Marshal(self.was.allLogs)
		logsAsJSON, _ := json.Marshal(vmLogger.StructLogs())
		utils.Info(string(logsAsJSON))
	}
	self.evaluateContract(crypto.GetAddressBytes(tx.From), receipt.ContractAddress, root)

	return nil
}

func (self *DVMService) evaluateContract(fromAddress crypto.AddressBytes, contractAddress crypto.AddressBytes, root crypto.HashBytes) {
	state := self.was.ethState
	contractStateObject := state.GetOrNewStateObject(contractAddress)
	contractHash := crypto.NewHash(contractAddress[:])
	stateHash := state.GetState(contractAddress, contractHash)
	trie := state.StorageTrie(contractAddress)

	fmt.Printf("Contract state object --> \n\n" +
		"Address:    %v\n" +
		"Hash:       %v\n" +
		"Nonce:      %v\n" +
		"Code:       %v\n" +
		"Code Hash:  %v\n" +
		"Tree Hash:  %v\n" +
		"Root Hash:  %v\n" +
		"StateHash:  %v\n\n",
		contractStateObject.Address(),
		contractHash,
		contractStateObject.Nonce(),
		contractStateObject.Code(state.Database()),
		contractStateObject.CodeHash(),
		trie.Hash(),
		root,
		stateHash,
	)

	iterateTrie(trie.NodeIterator(root.Bytes()), true)
	fmt.Println("\n")
	//bytes := state.GetState(address, root)
	//s := state.GetOrNewStateObject(root)

}

func iterateTrie(iterator trie.NodeIterator, isRoot bool) {
	path := hex.EncodeToString(iterator.Path())
	if iterator.Leaf() {
		fmt.Printf("\n" +
			"Leaf Node:   %v\n" +
			"With Path:   %v\n"+
			"With Parent: %v\n"+
			"Leaf Key:    %v\n"+
			"Leaf Blob:   %v\n",
			iterator.Hash(), path, iterator.Parent(), iterator.LeafKey(), iterator.LeafBlob())

	} else {
		if isRoot {
			fmt.Printf( "\n" +
				"Root Hash:   %v\n" +
				"With Path:   %v\n",
				iterator.Hash(), path)
		} else {
			fmt.Printf( "\n" +
				"Node Hash:   %v\n" +
				"With Path:   %v\n" +
				"With Parent: %v\n",
				iterator.Hash(), path, iterator.Parent())
		}
		iterator.Next(true)
		iterateTrie(iterator, false)
	}
}

func (self *DVMService) call(callMsg ethTypes.Message) ([]byte, error) {
	context := vm.Context{
		CanTransfer: ethereum.CanTransfer,
		Transfer:    ethereum.Transfer,
		GetHash:     func(uint64) crypto.HashBytes { return crypto.HashBytes{} },
		// Message information
		Origin:   callMsg.From(),
		GasPrice: callMsg.GasPrice(),
	}

	// The EVM should never be reused and is not thread safe.
	// Call is done on a copy of the state...we dont want any changes to be persisted
	// Call is a readonly operation
	vmLogger := vm.NewStructLogger(&vm.LogConfig{
		DisableMemory:  false,
		DisableStack:   false,
		DisableStorage: false,
		Debug:          isDemo,
		Limit:          0,
	})

	vmenv := vm.NewEVM(
		context,
		self.was.ethState.Copy(),
		&params.ChainConfig{
			ChainId: chainID,
		},
		vm.Config{
			Debug:  isDemo,
			Tracer: vmLogger,
		},
	)

	// Apply the transaction to the current state (included in the env)
	res, _, _, err := ethereum.ApplyMessage(vmenv, callMsg, self.was.gp)
	if err != nil {
		utils.Error(fmt.Sprintf("%s Executing Call on WAS", err))
		return nil, err
	}

	// DEMO-Today
	if isDemo {
		// logsAsJSON, _ := json.Marshal(self.was.allLogs)
		logsAsJSON, _ := json.Marshal(vmLogger.StructLogs())
		utils.Info(string(logsAsJSON))
	}

	return res, err
}

func (self *DVMService) commit() (crypto.HashBytes, error) {
	//commit all state changes to the database
	root, err := self.was.Commit()
	if err != nil {
		utils.Error(fmt.Sprintf("%s Committing WAS", err))

		return root, err
	}

	// reset the write ahead state for the next block
	// with the latest eth state
	self.statedb = self.was.ethState
	utils.Info(fmt.Sprintf("root %s Committed", root.Hex()))

	self.resetWAS()

	return root, nil
}

func (self *DVMService) resetWAS() {
	self.was = &WriteAheadState{
		db:           self.db,
		ethState:     self.statedb.Copy(),
		txIndex:      0,
		totalUsedGas: big.NewInt(0),
		gp:           new(ethereum.GasPool).AddGas(gasLimit.Uint64()),
	}
	// utils.Info("Reset Write Ahead state")
}

func (self *DVMService) getReceipt(txHash []byte) (*ethTypes.Receipt, error) {
	data, err := self.db.Get(append(receiptsPrefix, txHash[:]...))
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
