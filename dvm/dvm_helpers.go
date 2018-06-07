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
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/dispatchlabs/disgo/commons/crypto"
	commonTypes "github.com/dispatchlabs/disgo/commons/types"
	"github.com/dispatchlabs/disgo/commons/utils"
	"github.com/dispatchlabs/disgo/dvm/badgerwrapper"
	"github.com/dispatchlabs/disgo/dvm/ethereum"
	"github.com/dispatchlabs/disgo/dvm/ethereum/params"
	"github.com/dispatchlabs/disgo/dvm/ethereum/rlp"
	ethTypes "github.com/dispatchlabs/disgo/dvm/ethereum/types"
	"github.com/dispatchlabs/disgo/dvm/ethereum/vm"
)

var (
	chainID            = big.NewInt(1)
	gasLimit           = big.NewInt(1000000000)
	txMetaSuffix       = []byte{0x01}
	receiptsPrefix     = []byte("receipts-")
	headTxKey          = []byte("LastTx")
	acctounStatePrefix = []byte("AccountState-")
	MIPMapLevels       = []uint64{1000000, 500000, 100000, 50000, 1000}
	isDemo             = false

	_defaultValue    = big.NewInt(0)
	_defaultGas      = big.NewInt(1000000)
	_defaultGasPrice = big.NewInt(0)
	_defaultGasLimit = 1000000000
	_defaultDivvy    = int64(0)
)

func (self *DVMService) applyTransaction(tx *commonTypes.Transaction, stateHelper *VMStateHelper) error {
	price := big.NewInt(int64(0))

	context := vm.Context{
		CanTransfer: ethereum.CanTransfer,
		Transfer:    ethereum.Transfer,
		GetHash:     func(uint64) crypto.HashBytes { return crypto.GetHashBytes(tx.Hash) },
		// Message information
		Origin:      crypto.GetAddressBytes(tx.From),
		GasLimit:    gasLimit.Uint64(),
		GasPrice:    price,
		BlockNumber: big.NewInt(0), //the vm has a dependency on this..
	}

	// Prepare the ethState with transaction Hash so that it can be used in emitted logs
	var txIndex = 0
	stateHelper.ethStateDB.Prepare(crypto.GetHashBytes(tx.Hash), crypto.GetHashBytes(tx.Hash), txIndex)

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
		stateHelper.ethStateDB,
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
	_, contractAddress, gas, failed, err := ethereum.ApplyMessage(vmenv, msg, stateHelper.gp)
	if err != nil {
		utils.Error(fmt.Sprintf("%s Applying transaction to WAS", err))
		return err
	}
	stateHelper.totalUsedGas.Add(stateHelper.totalUsedGas, big.NewInt(0).SetUint64(gas))

	// Create a new receipt for the transaction, storing the intermediate root and gas used by the tx
	// based on the eip phase, we're passing wether the root touch-delete accounts.
	root := stateHelper.ethStateDB.IntermediateRoot(true) //this has side effects. It updates StateObjects (SmartContract memory)

	receipt := ethTypes.NewReceipt(root.Bytes(), failed, stateHelper.totalUsedGas.Uint64())
	receipt.TxHash = crypto.GetHashBytes(tx.Hash)
	receipt.GasUsed = gas
	// if the transaction created a contract, store the creation address in the receipt.
	if msg.To() == nil {
		receipt.ContractAddress = contractAddress

		stateHelper.to = receipt.ContractAddress
	}
	// Set the receipt logs and create a bloom for filtering
	receipt.Logs = stateHelper.ethStateDB.GetLogs(crypto.GetHashBytes(tx.Hash))
	//receipt.Logs = s.was.state.Logs()
	receipt.Bloom = ethTypes.CreateBloom(ethTypes.Receipts{receipt})

	stateHelper.txIndex++
	stateHelper.transactions = append(stateHelper.transactions, tx)
	stateHelper.receipts = append(stateHelper.receipts, receipt)
	stateHelper.allLogs = append(stateHelper.allLogs, receipt.Logs...)

	utils.Debug(fmt.Sprintf("%s Applied tx to WAS", tx.Hash))

	// DEMO-Today
	if isDemo {
		// logsAsJSON, _ := json.Marshal(stateHelper.allLogs)
		logsAsJSON, _ := json.Marshal(vmLogger.StructLogs())
		utils.Debug(string(logsAsJSON))
	}
	// self.evaluateContract(crypto.GetAddressBytes(tx.From), receipt.ContractAddress, root)

	return nil
}

func (self *DVMService) call(callMsg ethTypes.Message, stateHelper *VMStateHelper) ([]byte, error) {
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
		stateHelper.ethStateDB,
		&params.ChainConfig{
			ChainId: chainID,
		},
		vm.Config{
			Debug:  isDemo,
			Tracer: vmLogger,
		},
	)

	// Apply the transaction to the current state (included in the env)
	res, _, _, _, err := ethereum.ApplyMessage(vmenv, callMsg, stateHelper.gp)
	if err != nil {
		utils.Error(fmt.Sprintf("%s Executing Call on WAS", err))
		return nil, err
	}

	// DEMO-Today
	if isDemo {
		// logsAsJSON, _ := json.Marshal(stateHelper.allLogs)
		logsAsJSON, _ := json.Marshal(vmLogger.StructLogs())
		utils.Debug(string(logsAsJSON))
	}

	return res, err
}

func (self *DVMService) getReceipt(txHash []byte) (*ethTypes.Receipt, error) {
	utils.Debug(fmt.Sprintf("receipts- [%v]", crypto.Encode(receiptsPrefix)))
	data, err := badgerwrapper.GetBadgerDatabase().Get(append(receiptsPrefix, txHash[:]...))
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
