package common

import (
	"math/big"
	"github.com/dispatchlabs/disgo/dvm/ethereum/common/hexutil"
	"encoding/hex"
	"github.com/dispatchlabs/disgo/commons/utils"
	"github.com/dispatchlabs/disgo/commons/crypto"
)

const (
	HashLength    = 32
	AddressLength = 20
)

//Address and its methods
type Address [AddressLength]byte

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
func (a Address) Bytes() []byte       { return a[:] }

func (a Address) Big() *big.Int { return new(big.Int).SetBytes(a[:]) }

// Sets the address to the value of b. If b is larger than len(a) it will panic
//func (a *Address) SetBytes(b []byte) {
//	if len(b) > len(a) {
//		b = b[len(b)-AddressLength:]
//	}
//	copy(a[AddressLength-len(b):], b)
//}

// Hash and its methods:

// Hash represents the 32 byte Keccak256 hash of arbitrary data.
type Hash [HashLength]byte

func (h Hash) Str() string   { return string(h[:]) }
func (h Hash) Bytes() []byte { return h[:] }
func (h Hash) Big() *big.Int { return new(big.Int).SetBytes(h[:]) }
func (h Hash) Hex() string   { return hexutil.Encode(h[:]) }

// Sets the hash to the value of b. If b is larger than len(h), 'b' will be cropped (from the left).
func (h *Hash) SetBytes(b []byte) {
	if len(b) > len(h) {
		b = b[len(b)-HashLength:]
	}

	copy(h[HashLength-len(b):], b)
}
func HexToHash(s string) Hash   { return BytesToHash(FromHex(s)) }
func BigToHash(b *big.Int) Hash { return BytesToHash(b.Bytes()) }

func BytesToHash(b []byte) Hash {
	var h Hash
	h.SetBytes(b)
	return h
}
func EmptyHash(h Hash) bool {
	return h == Hash{}
}


