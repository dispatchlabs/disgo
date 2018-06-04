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
	"github.com/dispatchlabs/disgo/commons/types"
	"github.com/dispatchlabs/disgo/dvm/ethereum/rlp"
)

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

// newObject creates a state object.
func newObject(db *StateDB, address crypto.AddressBytes, data types.Account) *stateObject {
	// if data.Balance == nil {
	// 	data.Balance = new(big.Int)
	// }
	if data.CodeHash == nil {
		data.CodeHash = emptyCodeHash
	}
	result := &stateObject{
		db:            db,
		account:       data,
		cachedStorage: make(Storage),
		dirtyStorage:  make(Storage),
	}
	return result
}

// EncodeRLP implements rlp.Encoder.
func (c *stateObject) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, c.account)
}

// setError remembers the first non-nil error it is called with.
func (self *stateObject) setError(err error) {
	if self.dbErr == nil {
		self.dbErr = err
	}
}

func (self *stateObject) markSuicided() {
	self.suicided = true
}

func (c *stateObject) touch() {
	var addressAsBytes = crypto.GetAddressBytes(c.account.Address)

	c.db.journal.append(touchChange{
		account: addressAsBytes,
	})
	if addressAsBytes == ripemd {
		// Explicitly put it in the dirty-cache, which is otherwise generated from
		// flattened journals.
		c.db.journal.dirty(addressAsBytes)
	}
}

func (c *stateObject) getTrie(db Database) Trie {
	if c.trie == nil {
		var err error

		var addressAsBytes = crypto.GetAddressBytes(c.account.Address)
		var addressHash = crypto.NewHash(addressAsBytes[:])

		c.trie, err = db.OpenStorageTrie(addressHash, c.account.Root)
		if err != nil {
			c.trie, _ = db.OpenStorageTrie(addressHash, crypto.HashBytes{})
			c.setError(fmt.Errorf("can't create storage trie: %v", err))
		}
	}
	return c.trie
}

// GetState returns a value in account storage.
func (self *stateObject) GetState(db Database, key crypto.HashBytes) crypto.HashBytes {
	value, exists := self.cachedStorage[key]
	if exists {
		return value
	}
	// Load from DB in case it is missing.
	enc, err := self.getTrie(db).TryGet(key[:])
	if err != nil {
		self.setError(err)
		return crypto.HashBytes{}
	}
	if len(enc) > 0 {
		_, content, _, err := rlp.Split(enc)
		if err != nil {
			self.setError(err)
		}
		value.SetBytes(content)
	}
	if (value != crypto.HashBytes{}) {
		self.cachedStorage[key] = value
	}
	return value
}

// SetState updates a value in account storage.
func (self *stateObject) SetState(db Database, key, value crypto.HashBytes) {
	fmt.Printf("***** DB SetState: %s\n", crypto.HashBytesToHashString(key))
	var addressAsBytes = crypto.GetAddressBytes(self.account.Address)

	self.db.journal.append(storageChange{
		account:  addressAsBytes,
		key:      key,
		prevalue: self.GetState(db, key),
	})
	self.setState(key, value)
}

func (self *stateObject) setState(key, value crypto.HashBytes) {
	self.cachedStorage[key] = value
	self.dirtyStorage[key] = value

}

// updateTrie writes cached storage modifications into the object's storage trie.
func (self *stateObject) updateTrie(db Database) Trie {
	tr := self.getTrie(db)
	for key, value := range self.dirtyStorage {
		delete(self.dirtyStorage, key)
		if (value == crypto.HashBytes{}) {
			self.setError(tr.TryDelete(key[:]))
			continue
		}
		// Encoding []byte cannot fail, ok to ignore the error.
		v, _ := rlp.EncodeToBytes(bytes.TrimLeft(value[:], "\x00"))
		self.setError(tr.TryUpdate(key[:], v))
	}
	return tr
}

// UpdateRoot sets the trie root to the current root hash of
func (self *stateObject) updateRoot(db Database) {
	self.updateTrie(db)
	self.account.Root = self.trie.Hash()
}

// CommitTrie the storage trie of the object to dwb.
// This updates the trie root.
func (self *stateObject) CommitTrie(db Database) error {
	self.updateTrie(db)
	if self.dbErr != nil {
		return self.dbErr
	}
	root, err := self.trie.Commit(nil)
	if err == nil {
		self.account.Root = root
	}
	return err
}

// AddBalance removes amount from c's balance.
// It is used to add funds to the destination account of a transfer.
func (c *stateObject) AddBalance(amount *big.Int) {
	// EIP158: We must check emptiness for the objects such that the account
	// clearing (0,0,0 objects) can take effect.
	if amount.Sign() == 0 {
		if c.empty() {
			c.touch()
		}

		return
	}
	c.SetBalance(new(big.Int).Add(c.Balance(), amount))
}

// SubBalance removes amount from c's balance.
// It is used to remove funds from the origin account of a transfer.
func (c *stateObject) SubBalance(amount *big.Int) {
	if amount.Sign() == 0 {
		return
	}
	c.SetBalance(new(big.Int).Sub(c.Balance(), amount))
}

func (self *stateObject) SetBalance(amount *big.Int) {
	var addressAsBytes = crypto.GetAddressBytes(self.account.Address)

	self.db.journal.append(balanceChange{
		account: addressAsBytes,
		prev:    self.account.Balance,
	})
	self.setBalance(amount)
}

func (self *stateObject) setBalance(amount *big.Int) {
	self.account.Balance = amount
}

// Return the gas back to the origin. Used by the Virtual machine or Closures
func (c *stateObject) ReturnGas(gas *big.Int) {}

func (self *stateObject) deepCopy(db *StateDB) *stateObject {
	var addressAsBytes = crypto.GetAddressBytes(self.account.Address)

	stateObject := newObject(db, addressAsBytes, self.account)
	if self.trie != nil {
		stateObject.trie = db.db.CopyTrie(self.trie)
	}
	stateObject.code = self.code
	stateObject.dirtyStorage = self.dirtyStorage.Copy()
	stateObject.cachedStorage = self.dirtyStorage.Copy()
	stateObject.suicided = self.suicided
	stateObject.dirtyCode = self.dirtyCode
	stateObject.deleted = self.deleted
	return stateObject
}

//
// Attribute accessors
//

// Returns the address of the contract/account
func (c *stateObject) Address() crypto.AddressBytes {
	var addressAsBytes = crypto.GetAddressBytes(c.account.Address)

	return addressAsBytes
}

// Code returns the contract code associated with this object, if any.
func (self *stateObject) Code(db Database) []byte {
	if self.code != nil {
		return self.code
	}
	if bytes.Equal(self.CodeHash(), emptyCodeHash) {
		return nil
	}

	var addressAsBytes = crypto.GetAddressBytes(self.account.Address)
	var addressHash = crypto.NewHash(addressAsBytes[:])

	code, err := db.ContractCode(addressHash, crypto.BytesToHash(self.CodeHash()))
	if err != nil {
		self.setError(fmt.Errorf("can't load code hash %x: %v", self.CodeHash(), err))
	}
	self.code = code
	return code
}

func (self *stateObject) SetCode(codeHash crypto.HashBytes, code []byte) {
	var addressAsBytes = crypto.GetAddressBytes(self.account.Address)

	prevcode := self.Code(self.db.db)
	self.db.journal.append(codeChange{
		account:  addressAsBytes,
		prevhash: self.CodeHash(),
		prevcode: prevcode,
	})
	self.setCode(codeHash, code)
}

func (self *stateObject) setCode(codeHash crypto.HashBytes, code []byte) {
	self.code = code
	self.account.CodeHash = codeHash[:]
	self.dirtyCode = true
}

func (self *stateObject) SetNonce(nonce uint64) {
	var addressAsBytes = crypto.GetAddressBytes(self.account.Address)

	self.db.journal.append(nonceChange{
		account: addressAsBytes,
		prev:    self.account.Nonce,
	})
	self.setNonce(nonce)
}

func (self *stateObject) setNonce(nonce uint64) {
	self.account.Nonce = nonce
}

func (self *stateObject) CodeHash() []byte {
	return self.account.CodeHash
}

func (self *stateObject) Balance() *big.Int {
	return self.account.Balance
}

func (self *stateObject) Nonce() uint64 {
	return self.account.Nonce
}

// Never called, but must be present to allow stateObject to be used
// as a vm.Account interface that also satisfies the vm.ContractRef
// interface. Interfaces are awesome.
func (self *stateObject) Value() *big.Int {
	panic("Value on stateObject should never be called")
}
