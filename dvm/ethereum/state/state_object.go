// Copyright 2014 The go-ethereum Authors
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

package state

import (
	"bytes"
	"fmt"
	"io"
	"math/big"

	"github.com/dispatchlabs/disgo/commons/crypto"
	"github.com/dispatchlabs/disgo/commons/services"
	"github.com/dispatchlabs/disgo/commons/types"
	"github.com/dispatchlabs/disgo/commons/utils"
	"github.com/dispatchlabs/disgo/dvm/ethereum/rlp"
)

// func getAccountByAddressFromBadger(address crypto.AddressBytes) *types.Account {
func getAccountByAddressFromBadger(address string) *types.Account {
	// utils.Debug(fmt.Sprintf("state_object-getAccountByAddressFromBadger: %s", crypto.Encode(address[:])))
	utils.Debug(fmt.Sprintf("state_object-getAccountByAddressFromBadger: %s", address))

	txn := services.NewTxn(true)
	defer txn.Discard()

	// addressAsString := crypto.EncodeNo0x(address[:])

	account, err := types.ToAccountByAddress(txn, address)
	if err != nil {
		utils.Debug(fmt.Sprintf("state_object-getAccountByAddressFromBadger: %v", err))
		return nil
	}

	return account
}

var emptyCodeHash = crypto.NewHash(nil).Bytes()

type Code []byte

func (self Code) String() string {
	return string(self) //strings.Join(Disassemble(self), " ")
}

type Storage map[crypto.HashBytes]crypto.HashBytes

func (self Storage) String() (str string) {
	for key, value := range self {
		str += fmt.Sprintf("%X : %X\n", key, value)
	}

	return
}

func (self Storage) Copy() Storage {
	cpy := make(Storage)
	for key, value := range self {
		cpy[key] = value
	}

	return cpy
}

// stateObject represents an Ethereum account which is being modified.
//
// The usage pattern is as follows:
// First you need to obtain a state object.
// Account values can be accessed and modified through the object.
// Finally, call CommitTrie to write the modified storage trie into a database.
type stateObject struct {
	account types.Account
	db      *StateDB

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error

	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access
	code Code // contract bytecode, which gets set when code is loaded

	cachedStorage Storage // Storage entry cache to avoid duplicate reads
	dirtyStorage  Storage // Storage entries that need to be flushed to disk

	// Cache flags.
	// When an object is marked suicided it will be delete from the trie
	// during the "update" phase of the state transition.
	dirtyCode bool // true if the code was updated
	suicided  bool
	deleted   bool
}

// empty returns whether the account is considered empty.
func (s *stateObject) empty() bool {
	// var accountFromBadger = getAccountByAddressFromBadger(s.account.Address)
	utils.Debug(fmt.Sprintf("stateObject-empty: %s -> %v", s.account.Address, s.account.Balance))

	return s.account.Nonce == 0 && s.account.Balance.Sign() == 0 && bytes.Equal(s.account.CodeHash, emptyCodeHash)
}

// Account is the Ethereum consensus representation of accounts.
// These objects are stored in the main account trie.
// type Account struct {
// 	Nonce    uint64
// 	Balance  *big.Int
// 	Root     crypto.HashBytes // merkle root of the storage trie
// 	CodeHash []byte
// }

// newStateObject creates a state object.
func newStateObject(db *StateDB, address crypto.AddressBytes, data types.Account) *stateObject {
	if data.Balance == nil {
		data.Balance = new(big.Int)
	}
	if data.CodeHash == nil {
		data.CodeHash = emptyCodeHash
	}

	data.Nonce = 0 // sets the object to dirty

	result := &stateObject{
		db:            db,
		account:       data,
		cachedStorage: make(Storage),
		dirtyStorage:  make(Storage),
	}

	return result
}

// EncodeRLP implements rlp.Encoder.
func (s *stateObject) EncodeRLP(w io.Writer) error {
	// var accountFromBadger = getAccountByAddressFromBadger(s.account.Address)
	utils.Debug(fmt.Sprintf("stateObject-EncodeRLP: %s -> %v", s.account.Address, s.account.Balance))

	return rlp.Encode(w, s.account)
}

// setError remembers the first non-nil error it is called with.
func (s *stateObject) setError(err error) {
	// var accountFromBadger = getAccountByAddressFromBadger(s.account.Address)
	utils.Debug(fmt.Sprintf("stateObject-setError: %s -> %v", s.account.Address, s.account.Balance))

	if s.dbErr == nil {
		s.dbErr = err
	}
}

func (s *stateObject) markSuicided() {
	// var accountFromBadger = getAccountByAddressFromBadger(s.account.Address)
	utils.Debug(fmt.Sprintf("stateObject-markSuicided: %s -> %v", s.account.Address, s.account.Balance))

	s.suicided = true
}

func (s *stateObject) touch() {
	// var accountFromBadger = getAccountByAddressFromBadger(s.account.Address)
	utils.Debug(fmt.Sprintf("stateObject-touch: %s -> %v", s.account.Address, s.account.Balance))

	var addressAsBytes = crypto.GetAddressBytes(s.account.Address)

	s.db.journal.append(touchChange{
		account: addressAsBytes,
	})
	if addressAsBytes == ripemd {
		// Explicitly put it in the dirty-cache, which is otherwise generated from
		// flattened journals.
		s.db.journal.dirty(addressAsBytes)
	}
}

func (s *stateObject) getTrie(db Database) Trie {
	// var accountFromBadger = getAccountByAddressFromBadger(s.account.Address)
	utils.Debug(fmt.Sprintf("stateObject-getTrie: %s -> %v", s.account.Address, s.account.Balance))

	if s.trie == nil {
		var err error

		var addressAsBytes = crypto.GetAddressBytes(s.account.Address)
		var addressHash = crypto.NewHash(addressAsBytes[:])

		s.trie, err = db.OpenStorageTrie(addressHash, s.account.Root)
		if err != nil {
			s.trie, _ = db.OpenStorageTrie(addressHash, crypto.HashBytes{})
			s.setError(fmt.Errorf("can't create storage trie: %v", err))
		}
	}
	return s.trie
}

// GetState returns a value in account storage.
func (s *stateObject) GetState(db Database, key crypto.HashBytes) crypto.HashBytes {
	// var accountFromBadger = getAccountByAddressFromBadger(s.account.Address)
	utils.Debug(fmt.Sprintf("stateObject-GetState: %s -> %v", s.account.Address, s.account.Balance))

	value, exists := s.cachedStorage[key]
	if exists {
		return value
	}
	// Load from DB in case it is missing.
	enc, err := s.getTrie(db).TryGet(key[:])
	if err != nil {
		s.setError(err)
		return crypto.HashBytes{}
	}
	if len(enc) > 0 {
		_, content, _, err := rlp.Split(enc)
		if err != nil {
			s.setError(err)
		}
		value.SetBytes(content)
	}
	if (value != crypto.HashBytes{}) {
		s.cachedStorage[key] = value
	}
	return value
}

// SetState updates a value in account storage.
func (s *stateObject) SetState(db Database, key, value crypto.HashBytes) {
	// var accountFromBadger = getAccountByAddressFromBadger(s.account.Address)
	utils.Debug("***** DB SetState: %s\n", crypto.HashBytesToHashString(key))
	utils.Debug(fmt.Sprintf("stateObject-SetState: %s -> %v", s.account.Address, s.account.Balance))

	var addressAsBytes = crypto.GetAddressBytes(s.account.Address)

	s.db.journal.append(storageChange{
		account:  addressAsBytes,
		key:      key,
		prevalue: s.GetState(db, key),
	})
	s.setState(key, value)
}

func (s *stateObject) setState(key, value crypto.HashBytes) {
	// var accountFromBadger = getAccountByAddressFromBadger(s.account.Address)
	utils.Debug(fmt.Sprintf("stateObject-setState: %s -> %v", s.account.Address, s.account.Balance))

	s.cachedStorage[key] = value
	s.dirtyStorage[key] = value
}

// updateTrie writes cached storage modifications into the object's storage trie.
func (s *stateObject) updateTrie(db Database) Trie {
	// var accountFromBadger = getAccountByAddressFromBadger(s.account.Address)
	utils.Debug(fmt.Sprintf("stateObject-updateTrie: %s -> %v", s.account.Address, s.account.Balance))

	tr := s.getTrie(db)
	for key, value := range s.dirtyStorage {
		delete(s.dirtyStorage, key)
		if (value == crypto.HashBytes{}) {
			s.setError(tr.TryDelete(key[:]))
			continue
		}
		// Encoding []byte cannot fail, ok to ignore the error.
		v, _ := rlp.EncodeToBytes(bytes.TrimLeft(value[:], "\x00"))
		s.setError(tr.TryUpdate(key[:], v))
	}
	return tr
}

// UpdateRoot sets the trie root to the current root hash of
func (s *stateObject) updateRoot(db Database) {
	// var accountFromBadger = getAccountByAddressFromBadger(s.account.Address)
	utils.Debug(fmt.Sprintf("stateObject-updateRoot: %s -> %v", s.account.Address, s.account.Balance))

	s.updateTrie(db)
	s.account.Root = s.trie.Hash()
}

// CommitTrie the storage trie of the object to dwb.
// This updates the trie root.
func (s *stateObject) CommitTrie(db Database) error {
	// var accountFromBadger = getAccountByAddressFromBadger(s.account.Address)
	utils.Debug(fmt.Sprintf("stateObject-CommitTrie: %s -> %v", s.account.Address, s.account.Balance))

	s.updateTrie(db)
	if s.dbErr != nil {
		return s.dbErr
	}
	root, err := s.trie.Commit(nil)
	if err == nil {
		s.account.Root = root
	}
	return err
}

// AddBalance removes amount from c's balance.
// It is used to add funds to the destination account of a transfer.
func (s *stateObject) AddBalance(amount *big.Int) {
	// var accountFromBadger = getAccountByAddressFromBadger(s.account.Address)
	utils.Debug(fmt.Sprintf("stateObject-AddBalance: %s -> %v", s.account.Address, s.account.Balance))

	// EIP158: We must check emptiness for the objects such that the account
	// clearing (0,0,0 objects) can take effect.
	if amount.Sign() == 0 {
		if s.empty() {
			s.touch()
		}

		return
	}

	s.SetBalance(new(big.Int).Add(s.account.Balance, amount))
}

// SubBalance removes amount from c's balance.
// It is used to remove funds from the origin account of a transfer.
func (s *stateObject) SubBalance(amount *big.Int) {
	// var accountFromBadger = getAccountByAddressFromBadger(s.account.Address)
	utils.Debug(fmt.Sprintf("stateObject-SubBalance: %s -> %v", s.account.Address, s.account.Balance))

	if amount.Sign() == 0 {
		return
	}
	s.SetBalance(new(big.Int).Sub(s.account.Balance, amount))
}

func (s *stateObject) SetBalance(amount *big.Int) {
	// var accountFromBadger = getAccountByAddressFromBadger(s.account.Address)
	utils.Debug(fmt.Sprintf("stateObject-SetBalance: %s -> %v", s.account.Address, s.account.Balance))

	var addressAsBytes = crypto.GetAddressBytes(s.account.Address)

	s.db.journal.append(balanceChange{
		account: addressAsBytes,
		prev:    s.account.Balance,
	})
	s.setBalance(amount)
}

func (s *stateObject) setBalance(amount *big.Int) {
	// var accountFromBadger = getAccountByAddressFromBadger(s.account.Address)
	utils.Debug(fmt.Sprintf("stateObject-setBalance: %s -> %v", s.account.Address, s.account.Balance))

	s.account.Balance = amount

	// txn := services.NewTxn(true)
	// defer txn.Discard()

	// account.Balance = amount
	// account.Persist(txn)
}

// Return the gas back to the origin. Used by the Virtual machine or Closures
func (s *stateObject) ReturnGas(gas *big.Int) {
	// IMPLEMENT-ME: ???
}

func (s *stateObject) deepCopy(db *StateDB) *stateObject {
	// var accountFromBadger = getAccountByAddressFromBadger(s.account.Address)
	utils.Debug(fmt.Sprintf("stateObject-deepCopy: %s -> %v", s.account.Address, s.account.Balance))

	var addressAsBytes = crypto.GetAddressBytes(s.account.Address)

	stateObject := newStateObject(db, addressAsBytes, s.account)
	if s.trie != nil {
		stateObject.trie = db.db.CopyTrie(s.trie)
	}
	stateObject.code = s.code
	stateObject.dirtyStorage = s.dirtyStorage.Copy()
	stateObject.cachedStorage = s.dirtyStorage.Copy()
	stateObject.suicided = s.suicided
	stateObject.dirtyCode = s.dirtyCode
	stateObject.deleted = s.deleted
	return stateObject
}

//
// Attribute accessors
//
func (s *stateObject) Account() *types.Account {
	// var accountFromBadger = getAccountByAddressFromBadger(s.account.Address)
	utils.Debug(fmt.Sprintf("stateObject-Account: %s -> %v", s.account.Address, s.account.Balance))

	return &s.account
}

// Code returns the contract code associated with this object, if any.
func (s *stateObject) Code(db Database) []byte {
	// var accountFromBadger = getAccountByAddressFromBadger(s.account.Address)
	utils.Debug(fmt.Sprintf("stateObject-Code: %s -> %v", s.account.Address, s.account.Balance))

	if s.code != nil {
		return s.code
	}
	if s.account.CodeHash == nil || len(s.account.CodeHash) <= 0 || bytes.Equal(s.account.CodeHash, emptyCodeHash) {
		return nil
	}

	var addressAsBytes = crypto.GetAddressBytes(s.account.Address)
	var addressHash = crypto.NewHash(addressAsBytes[:])

	code, err := db.ContractCode(addressHash, crypto.BytesToHash(s.account.CodeHash))
	if err != nil {
		s.setError(fmt.Errorf("can't load code hash %x: %v", s.account.CodeHash, err))
	}
	s.code = code
	return code
}

func (s *stateObject) SetCode(codeHash crypto.HashBytes, code []byte) {
	// var accountFromBadger = getAccountByAddressFromBadger(s.account.Address)
	utils.Debug(fmt.Sprintf("stateObject-SetCode: %s -> %v", s.account.Address, s.account.Balance))

	var addressAsBytes = crypto.GetAddressBytes(s.account.Address)

	prevcode := s.Code(s.db.db)
	s.db.journal.append(codeChange{
		account:  addressAsBytes,
		prevhash: s.account.CodeHash,
		prevcode: prevcode,
	})
	s.setCode(codeHash, code)
}

func (s *stateObject) setCode(codeHash crypto.HashBytes, code []byte) {
	// var accountFromBadger = getAccountByAddressFromBadger(s.account.Address)
	utils.Debug(fmt.Sprintf("stateObject-setCode: %s -> %v", s.account.Address, s.account.Balance))

	// txn := services.NewTxn(true)
	// defer txn.Commit(nil)
	// account.CodeHash = codeHash[:]
	// account.Persist(txn)

	s.code = code
	s.account.CodeHash = codeHash[:]
	s.dirtyCode = true
}

func (s *stateObject) SetNonce(nonce uint64) {
	// var accountFromBadger = getAccountByAddressFromBadger(s.account.Address)
	utils.Debug(fmt.Sprintf("stateObject-SetNonce: %s -> %v", s.account.Address, s.account.Balance))

	var addressAsBytes = crypto.GetAddressBytes(s.account.Address)

	s.db.journal.append(nonceChange{
		account: addressAsBytes,
		prev:    s.account.Nonce,
	})
	s.account.Nonce = nonce
}

// Never called, but must be present to allow stateObject to be used
// as a vm.Account interface that also satisfies the vm.ContractRef
// interface. Interfaces are awesome.
func (s *stateObject) Value() *big.Int {
	// var accountFromBadger = getAccountByAddressFromBadger(s.account.Address)
	utils.Debug(fmt.Sprintf("stateObject-Value: %s -> %v", s.account.Address, s.account.Balance))

	panic("Value on stateObject should never be called")
}
