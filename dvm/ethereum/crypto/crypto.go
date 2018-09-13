package crypto

import (
	"fmt"
	"hash"

	"github.com/dispatchlabs/disgo/commons/crypto"
	"github.com/dispatchlabs/disgo/commons/utils"
	"github.com/dispatchlabs/disgo/dvm/ethereum/common"
	"github.com/dispatchlabs/disgo/dvm/ethereum/rlp"
)

// Creates an ethereum address given the bytes and the nonce
func CreateAddress(b crypto.AddressBytes, nonce uint64) crypto.AddressBytes {
	data, err := rlp.EncodeToBytes([]interface{}{b, nonce})
	if err != nil {
		utils.Error(err)
	}
	bytes := common.BytesToAddress(Keccak256(data)[12:])
	utils.Info(fmt.Sprintf("***** CreateAddress: %s : @ %s", common.EthAddressToDispatchAddress(bytes), utils.GetCallStackWithFileAndLineNumber()))

	return bytes
}

// CreateAddress2 creates an ethereum address given the address bytes, initial
// contract code and a salt.
func CreateAddress2(b crypto.AddressBytes, salt [32]byte, code []byte) crypto.AddressBytes {
	return common.BytesToAddress(Keccak256([]byte{0xff}, b.Bytes(), salt[:], Keccak256(code))[12:])
}

func NewKeccak256() hash.Hash { return &state{rate: 136, outputLen: 32, dsbyte: 0x01} }

// Keccak256 calculates and returns the Keccak256 hash of the input data.
func Keccak256(data ...[]byte) []byte {
	d := NewKeccak256()
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}

// Keccak256Hash calculates and returns the Keccak256 hash of the input data,
// converting it to an internal Hash data structure.
func Keccak256Hash(data ...[]byte) (h crypto.HashBytes) {
	d := NewKeccak256()
	for _, b := range data {
		d.Write(b)
	}
	d.Sum(h[:0])
	return h
}
