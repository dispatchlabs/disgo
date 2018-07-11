package crypto

import (
	"crypto/ecdsa"
	"fmt"
	"hash"

	dvmCrypto "github.com/dispatchlabs/disgo/commons/crypto"
	"github.com/dispatchlabs/disgo/commons/crypto/secp256k1"
	"github.com/dispatchlabs/disgo/dvm/ethereum/common"
	"github.com/dispatchlabs/disgo/dvm/ethereum/common/math"
	"github.com/dispatchlabs/disgo/dvm/ethereum/rlp"
	"github.com/dispatchlabs/disgo/commons/utils"
)

// Creates an ethereum address given the bytes and the nonce
func CreateAddress(b dvmCrypto.AddressBytes, nonce uint64) dvmCrypto.AddressBytes {
	data, _ := rlp.EncodeToBytes([]interface{}{b, nonce})

	bytes := common.BytesToAddress(Keccak256(data)[12:])
	utils.Info(fmt.Sprintf("***** CreateAddress: %s : @ %s", common.EthAddressToDispatchAddress(bytes), utils.GetCallStackWithFileAndLineNumber()))

	return bytes
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
func Keccak256Hash(data ...[]byte) (h common.Hash) {
	d := NewKeccak256()
	for _, b := range data {
		d.Write(b)
	}
	d.Sum(h[:0])
	return h
}

// Sign calculates an ECDSA signature.
//
// This function is susceptible to chosen plaintext attacks that can leak
// information about the private key that is used for signing. Callers must
// be aware that the given hash cannot be chosen by an adversery. Common
// solution is to hash any input before calculating the signature.
//
// The produced signature is in the [R || S || V] format where V is 0 or 1.
func Sign(hash []byte, prv *ecdsa.PrivateKey) (sig []byte, err error) {
	if len(hash) != 32 {
		return nil, fmt.Errorf("hash is required to be exactly 32 bytes (%d)", len(hash))
	}
	seckey := math.PaddedBigBytes(prv.D, prv.Params().BitSize/8)
	defer zeroBytes(seckey)
	return secp256k1.Sign(hash, seckey)
}
