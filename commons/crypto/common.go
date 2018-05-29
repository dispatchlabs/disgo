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
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"math/big"

	"github.com/dispatchlabs/commons/crypto/secp256k1"
	"github.com/dispatchlabs/commons/math"
	"github.com/dispatchlabs/commons/utils"
)

type AddressBytes [AddressLength]byte
type HashBytes [HashLength]byte

func (h HashBytes) Str() string   { return string(h[:]) }
func (h HashBytes) Bytes() []byte { return h[:] }
func (h HashBytes) Big() *big.Int { return new(big.Int).SetBytes(h[:]) }
func (h HashBytes) Hex() string   { return Encode(h[:]) }

func HexToHash(s string) HashBytes   { return BytesToHash(FromHex(s)) }
func BigToHash(b *big.Int) HashBytes { return BytesToHash(b.Bytes()) }
func EmptyHash(h HashBytes) bool     { return h == HashBytes{} }

func (a AddressBytes) Big() *big.Int { return new(big.Int).SetBytes(a[:]) }
func (a AddressBytes) Bytes() []byte { return a[:] }

// Encode encodes b as a hex string with 0x prefix.
func Encode(b []byte) string {
	enc := make([]byte, len(b)*2+2)
	copy(enc, "0x")
	hex.Encode(enc[2:], b)
	return string(enc)
}

func (h *HashBytes) SetBytes(b []byte) {
	if len(b) > len(h) {
		b = b[len(b)-HashLength:]
	}
	copy(h[HashLength-len(b):], b)
}

func BytesToHash(b []byte) HashBytes {
	var h HashBytes
	h.SetBytes(b)
	return h
}

func FromHex(s string) []byte {
	if len(s) > 1 {
		if s[0:2] == "0x" || s[0:2] == "0X" {
			s = s[2:]
		}
	}
	if len(s)%2 == 1 {
		s = "0" + s
	}
	return Hex2Bytes(s)
}

func Hex2Bytes(str string) []byte {
	h, _ := hex.DecodeString(str)

	return h
}

func GetAddressBytes(value string) AddressBytes {
	bytes, err := hex.DecodeString(value)
	if err != nil {
		utils.Error("unable to decode value", err)
	}
	var addressBytes AddressBytes
	bytes = bytes[len(bytes)-AddressLength:]
	copy(addressBytes[AddressLength-len(bytes):], bytes)
	return addressBytes
}

func GetHashBytes(value string) HashBytes {
	bytes, err := hex.DecodeString(value)
	if err != nil {
		utils.Error("unable to decode value", err)
	}
	var hashBytes HashBytes
	if len(bytes) > len(hashBytes) {
		bytes = bytes[len(bytes)-HashLength:]
	}
	copy(hashBytes[HashLength-len(bytes):], bytes)
	return hashBytes
}

func PubkeyToAddress(p ecdsa.PublicKey) string {
	pubBytes := FromECDSAPub(&p)
	hash := NewHash(pubBytes[1:]).Bytes()
	address := hash[12:]
	return hex.EncodeToString(address[:])
}

func FromECDSAPub(pub *ecdsa.PublicKey) []byte {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}
	return elliptic.Marshal(secp256k1.S256(), pub.X, pub.Y)
}

// FromECDSA exports a private key into a binary dump.
func FromECDSA(priv *ecdsa.PrivateKey) []byte {
	if priv == nil {
		return nil
	}
	return math.PaddedBigBytes(priv.D, priv.Params().BitSize/8)
}

// HexToECDSA parses a secp256k1 private key.
func HexToECDSA(hexkey string) (*ecdsa.PrivateKey, error) {
	b, err := hex.DecodeString(hexkey)
	if err != nil {
		return nil, err
	}
	return toECDSA(b, true)
}

func aesCTRXOR(key, inText, iv []byte) ([]byte, error) {
	// AES-128 is selected due to size of encryptKey.
	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	stream := cipher.NewCTR(aesBlock, iv)
	outText := make([]byte, len(inText))
	stream.XORKeyStream(outText, inText)
	return outText, err
}
