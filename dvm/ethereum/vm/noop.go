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

	"github.com/dispatchlabs/disgo/commons/crypto"
	"github.com/dispatchlabs/disgo/dvm/ethereum/types"
)

func NoopCanTransfer(db StateDB, from crypto.AddressBytes, balance *big.Int) bool {
	return true
}
func NoopTransfer(db StateDB, from, to crypto.AddressBytes, amount *big.Int) {}

type NoopEVMCallContext struct{}

func (NoopEVMCallContext) Call(caller ContractRef, addr crypto.AddressBytes, data []byte, gas, value *big.Int) ([]byte, error) {
	return nil, nil
}
func (NoopEVMCallContext) CallCode(caller ContractRef, addr crypto.AddressBytes, data []byte, gas, value *big.Int) ([]byte, error) {
	return nil, nil
}
func (NoopEVMCallContext) Create(caller ContractRef, data []byte, gas, value *big.Int) ([]byte, crypto.AddressBytes, error) {
	return nil, crypto.AddressBytes{}, nil
}
func (NoopEVMCallContext) DelegateCall(me ContractRef, addr crypto.AddressBytes, data []byte, gas *big.Int) ([]byte, error) {
	return nil, nil
}

type NoopStateDB struct{}

func (NoopStateDB) CreateAccount(crypto.AddressBytes)                {}
func (NoopStateDB) SubBalance(crypto.AddressBytes, *big.Int)         {}
func (NoopStateDB) AddBalance(crypto.AddressBytes, *big.Int)         {}
func (NoopStateDB) GetBalance(crypto.AddressBytes) *big.Int          { return nil }
func (NoopStateDB) GetNonce(crypto.AddressBytes) uint64              { return 0 }
func (NoopStateDB) SetNonce(crypto.AddressBytes, uint64)             {}
func (NoopStateDB) GetCodeHash(crypto.AddressBytes) crypto.HashBytes { return crypto.HashBytes{} }
func (NoopStateDB) GetCode(crypto.AddressBytes) []byte               { return nil }
func (NoopStateDB) SetCode(crypto.AddressBytes, []byte)              {}
func (NoopStateDB) GetCodeSize(crypto.AddressBytes) int              { return 0 }
func (NoopStateDB) AddRefund(uint64)                                 {}
func (NoopStateDB) GetRefund() uint64                                { return 0 }
func (NoopStateDB) GetState(crypto.AddressBytes, crypto.HashBytes) crypto.HashBytes {
	return crypto.HashBytes{}
}
func (NoopStateDB) SetState(crypto.AddressBytes, crypto.HashBytes, crypto.HashBytes) {}
func (NoopStateDB) Suicide(crypto.AddressBytes) bool                                 { return false }
func (NoopStateDB) HasSuicided(crypto.AddressBytes) bool                             { return false }
func (NoopStateDB) Exist(crypto.AddressBytes) bool                                   { return false }
func (NoopStateDB) Empty(crypto.AddressBytes) bool                                   { return false }
func (NoopStateDB) RevertToSnapshot(int)                                             {}
func (NoopStateDB) Snapshot() int                                                    { return 0 }
func (NoopStateDB) AddLog(*types.Log)                                                {}
func (NoopStateDB) AddPreimage(crypto.HashBytes, []byte)                             {}
func (NoopStateDB) ForEachStorage(crypto.AddressBytes, func(crypto.HashBytes, crypto.HashBytes) bool) {
}
