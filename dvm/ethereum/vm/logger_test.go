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

// import (
// 	"math/big"
// 	"testing"

// 	"github.com/dispatchlabs/disgo/dvm/ethereum/common"
// 	"github.com/dispatchlabs/disgo/dvm/ethereum/params"
// 	"github.com/dispatchlabs/disgo/dvm/ethereum/types"
// )

// type dummyContractRef struct {
// 	calledForEach bool
// }

// func (dummyContractRef) ReturnGas(*big.Int)          {}
// func (dummyContractRef) Address() common.Address     { return common.Address{} }
// func (dummyContractRef) Value() *big.Int             { return new(big.Int) }
// func (dummyContractRef) SetCode(common.Hash, []byte) {}
// func (d *dummyContractRef) ForEachStorage(callback func(key, value common.Hash) bool) {
// 	d.calledForEach = true
// }
// func (d *dummyContractRef) SubBalance(amount *big.Int) {}
// func (d *dummyContractRef) AddBalance(amount *big.Int) {}
// func (d *dummyContractRef) SetBalance(*big.Int)        {}
// func (d *dummyContractRef) SetNonce(uint64)            {}
// func (d *dummyContractRef) Balance() *big.Int          { return new(big.Int) }

// type dummyStateDB struct {
// 	NoopStateDB
// 	ref *dummyContractRef
// }

// func TestStoreCapture(t *testing.T) {
// 	var (
// 		env      = NewEVM(Context{}, nil, params.TestChainConfig, Config{})
// 		logger   = NewStructLogger(nil)
// 		mem      = NewMemory()
// 		stack    = newstack()
// 		contract = NewContract(&dummyContractRef{}, &dummyContractRef{}, new(big.Int), 0)
// 	)
// 	stack.push(big.NewInt(1))
// 	stack.push(big.NewInt(0))

// 	var index common.Hash

// 	logger.CaptureState(env, 0, SSTORE, 0, 0, mem, stack, contract, 0, nil)
// 	if len(logger.changedValues[contract.Address()]) == 0 {
// 		t.Fatalf("expected exactly 1 changed value on address %x, got %d", contract.Address(), len(logger.changedValues[contract.Address()]))
// 	}
// 	exp := common.BigToHash(big.NewInt(1))
// 	if logger.changedValues[contract.Address()][index] != exp {
// 		t.Errorf("expected %x, got %x", exp, logger.changedValues[contract.Address()][index])
// 	}
// }

// type NoopStateDB struct{}

// func (NoopStateDB) CreateAccount(common.Address)                                       {}
// func (NoopStateDB) SubBalance(common.Address, *big.Int)                                {}
// func (NoopStateDB) AddBalance(common.Address, *big.Int)                                {}
// func (NoopStateDB) GetBalance(common.Address) *big.Int                                 { return nil }
// func (NoopStateDB) GetNonce(common.Address) uint64                                     { return 0 }
// func (NoopStateDB) SetNonce(common.Address, uint64)                                    {}
// func (NoopStateDB) GetCodeHash(common.Address) common.Hash                             { return common.Hash{} }
// func (NoopStateDB) GetCode(common.Address) []byte                                      { return nil }
// func (NoopStateDB) SetCode(common.Address, []byte)                                     {}
// func (NoopStateDB) GetCodeSize(common.Address) int                                     { return 0 }
// func (NoopStateDB) AddRefund(uint64)                                                   {}
// func (NoopStateDB) GetRefund() uint64                                                  { return 0 }
// func (NoopStateDB) GetState(common.Address, common.Hash) common.Hash                   { return common.Hash{} }
// func (NoopStateDB) SetState(common.Address, common.Hash, common.Hash)                  {}
// func (NoopStateDB) Suicide(common.Address) bool                                        { return false }
// func (NoopStateDB) HasSuicided(common.Address) bool                                    { return false }
// func (NoopStateDB) Exist(common.Address) bool                                          { return false }
// func (NoopStateDB) Empty(common.Address) bool                                          { return false }
// func (NoopStateDB) RevertToSnapshot(int)                                               {}
// func (NoopStateDB) Snapshot() int                                                      { return 0 }
// func (NoopStateDB) AddLog(*types.Log)                                                  {}
// func (NoopStateDB) AddPreimage(common.Hash, []byte)                                    {}
// func (NoopStateDB) ForEachStorage(common.Address, func(common.Hash, common.Hash) bool) {}
