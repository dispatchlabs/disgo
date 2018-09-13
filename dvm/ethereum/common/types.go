// Copyright 2015 The go-ethereum Authors
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

package common

import (
	"encoding/hex"
	"math/big"

	"github.com/dispatchlabs/disgo/commons/crypto"
	"github.com/dispatchlabs/disgo/commons/utils"
)

// Lengths of hashes and addresses in bytes.
const (
	// HashLength is the expected length of the hash
	HashLength = 32
	// AddressLength is the expected length of the address
	AddressLength = 20
)

type Message struct {
	to         crypto.AddressBytes
	from       crypto.AddressBytes
	nonce      uint64
	amount     *big.Int
	gasLimit   uint64
	gasPrice   *big.Int
	data       []byte
	checkNonce bool
}

func DispatchAddressToEthAddress(address string) crypto.AddressBytes {
	bytes, err := hex.DecodeString(address)
	if err != nil {
		utils.Error("unable to decode address", err)
	}
	var addressBytes crypto.AddressBytes
	if len(bytes) > len(addressBytes) {
		bytes = bytes[len(bytes)-crypto.AddressLength:]
	}
	copy(addressBytes[crypto.AddressLength-len(bytes):], bytes)
	return addressBytes
}

func EthAddressToDispatchAddress(address crypto.AddressBytes) string {
	return hex.EncodeToString(address[:])
}

func BytesToAddress(bytes []byte) crypto.AddressBytes {
	var addressBytes crypto.AddressBytes
	if len(bytes) > len(addressBytes) {
		bytes = bytes[len(bytes)-crypto.AddressLength:]
	}
	copy(addressBytes[crypto.AddressLength-len(bytes):], bytes)
	return addressBytes
}

func BigToAddress(b *big.Int) crypto.AddressBytes { return BytesToAddress(b.Bytes()) }
func HexToAddress(s string) crypto.AddressBytes   { return BytesToAddress(FromHex(s)) }
