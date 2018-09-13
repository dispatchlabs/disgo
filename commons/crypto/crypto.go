/*
 *    This file is part of Disgo-Commons library.
 *
 *    The Disgo-Commons library is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *
 *    The Disgo-Commons library is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *
 *    You should have received a copy of the GNU General Public License
 *    along with the Disgo-Commons library.  If not, see <http://www.gnu.org/licenses/>.
 */
package crypto

import (
	"crypto/ecdsa"
	"crypto/rand"
	"math/big"

	"crypto/elliptic"

	"github.com/dispatchlabs/disgo/commons/crypto/secp256k1"
	"github.com/dispatchlabs/disgo/commons/math"
	"github.com/dispatchlabs/disgo/commons/utils"
	"github.com/ebfe/keccak"
)

const (
	AddressLength   = 20
	HashLength      = 32
	SignatureLength = 65
)

// NewHash
func NewHash(bytes ...[]byte) (digest HashBytes) {
	hash := keccak.New256()
	for _, b := range bytes {
		hash.Write(b)
	}
	hash.Sum(digest[:0])
	return digest
}

// GenerateKeyPair
func GenerateKeyPair() (publicKey, privateKey []byte) {
	key, error := ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
	if error != nil {
		panic(error)
	}

	// 	return secp256k1.CompressPubkey(key.PublicKey.X, key.PublicKey.Y), math.PaddedBigBytes(key.D, 32)
	return elliptic.Marshal(secp256k1.S256(), key.PublicKey.X, key.PublicKey.Y), math.PaddedBigBytes(key.D, 32)
}

// - GenerateKeyPair
func GeneratePrivateKey() (*ecdsa.PrivateKey, error) {
	key, err := ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// ToPublicKey
func ToPublicKey(hash []byte, signature []byte) ([]byte, error) {
	publicKey, err := secp256k1.RecoverPubkey(hash, signature)
	if err != nil {
		utils.Error(err)
		return nil, err
	}
	return publicKey, nil
}

// ToAddress
func ToAddress(publicKey []byte) []byte {
	hash := NewHash(publicKey[1:])
	address := hash[12:]
	return address
}

// NewSignature
func NewSignature(privateKey []byte, hash []byte) ([]byte, error) {
	signature, err := secp256k1.Sign(hash, privateKey)
	if err != nil {
		return make([]byte, 0), err
	}
	return signature, nil
}

// VerifyHashAndSignature
func VerifySignature(publicKey []byte, hash []byte, signature []byte) bool {
	return secp256k1.VerifySignature(publicKey, hash, signature[:len(signature)-1])
}

func ToBytesFromECDSAPublicKey(pub *ecdsa.PublicKey) []byte {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}
	return elliptic.Marshal(secp256k1.S256(), pub.X, pub.Y)
}

/*
func (fs FrontierSigner) SignatureValues(tx *Transaction, sig []byte) (r, s, v *big.Int, err error) {
	if len(sig) != 65 {
		panic(fmt.Sprintf("wrong size for signature: got %d, want 65", len(sig)))
	}
	r = new(big.Int).SetBytes(sig[:32])
	s = new(big.Int).SetBytes(sig[32:64])
	v = new(big.Int).SetBytes([]byte{sig[64] + 27})
	return r, s, v, nil
}
*/

/*
The below code is copied from the Ethereum library.
I expect us to not use this in the future,
but leaving it here so we can compile and run for now
*/

var (
	secp256k1_N, _  = new(big.Int).SetString("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141", 16)
	secp256k1_halfN = new(big.Int).Div(secp256k1_N, big.NewInt(2))
)

// ValidateSignatureValues verifies whether the signature values are valid with
// the given chain rules. The v value is assumed to be either 0 or 1.
func ValidateSignatureValues(v byte, r, s *big.Int, homestead bool) bool {
	if r.Cmp(big.NewInt(1)) < 0 || s.Cmp(big.NewInt(1)) < 0 {
		return false
	}
	// reject upper range of s values (ECDSA malleability)
	// see discussion in secp256k1/libsecp256k1/include/secp256k1.h
	if homestead && s.Cmp(secp256k1_halfN) > 0 {
		return false
	}
	// Frontier: allow s to be in full N range
	return r.Cmp(secp256k1_N) < 0 && s.Cmp(secp256k1_N) < 0 && (v == 0 || v == 1)
}
