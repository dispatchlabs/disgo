// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package vm

import (
	"math/big"

	"github.com/dispatchlabs/disgo/dvm/ethereum/types"
	"github.com/dispatchlabs/disgo/commons/crypto"
)

// StateDB is an EVM database for full state querying.
type StateDB interface {
	CreateAccount(crypto.AddressBytes)

	SubBalance(crypto.AddressBytes, *big.Int)
	AddBalance(crypto.AddressBytes, *big.Int)
	GetBalance(crypto.AddressBytes) *big.Int

	GetNonce(crypto.AddressBytes) uint64
	SetNonce(crypto.AddressBytes, uint64)

	GetCodeHash(crypto.AddressBytes) crypto.HashBytes
	GetCode(crypto.AddressBytes) []byte
	SetCode(crypto.AddressBytes, []byte)
	GetCodeSize(crypto.AddressBytes) int

	AddRefund(uint64)
	GetRefund() uint64

	GetState(crypto.AddressBytes, crypto.HashBytes) crypto.HashBytes
	SetState(crypto.AddressBytes, crypto.HashBytes, crypto.HashBytes)

	Suicide(crypto.AddressBytes) bool
	HasSuicided(crypto.AddressBytes) bool

	// Exist reports whether the given account exists in state.
	// Notably this should also return true for suicided accounts.
	Exist(crypto.AddressBytes) bool
	// Empty returns whether the given account is empty. Empty
	// is defined according to EIP161 (balance = nonce = code = 0).
	Empty(crypto.AddressBytes) bool

	RevertToSnapshot(int)
	Snapshot() int

	AddLog(*types.Log)
	AddPreimage(crypto.HashBytes, []byte)

	ForEachStorage(crypto.AddressBytes, func(crypto.HashBytes, crypto.HashBytes) bool)
}

// CallContext provides a basic interface for the EVM calling conventions. The EVM EVM
// depends on this context being implemented for doing subcalls and initialising new EVM contracts.
type CallContext interface {
	// Call another contract
	Call(env *EVM, me ContractRef, addr crypto.AddressBytes, data []byte, gas, value *big.Int) ([]byte, error)
	// Take another's contract code and execute within our own context
	CallCode(env *EVM, me ContractRef, addr crypto.AddressBytes, data []byte, gas, value *big.Int) ([]byte, error)
	// Same as CallCode except sender and value is propagated from parent to child scope
	DelegateCall(env *EVM, me ContractRef, addr crypto.AddressBytes, data []byte, gas *big.Int) ([]byte, error)
	// Create a new contract
	Create(env *EVM, me ContractRef, data []byte, gas, value *big.Int) ([]byte, crypto.AddressBytes, error)
}
