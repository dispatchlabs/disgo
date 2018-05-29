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
	"encoding/hex"
	"encoding/json"
	"github.com/pborman/uuid"
	"io"
	"github.com/dispatchlabs/commons/math"
)

// Version 4 "random" for unique id not derived from key data
// to simplify lookups we also store the address
// we only store privkey as pubkey/address can be derived from it
// privkey in this struct is always in plaintext
type Key struct {
	Id         uuid.UUID
	Address    string
	PrivateKey *ecdsa.PrivateKey
}

type plainKeyJSON struct {
	Address    string `json:"address"`
	PrivateKey string `json:"privatekey"`
	Id         string `json:"id"`
	Version    int    `json:"version"`
}

func (k *Key) MarshalJSON() (j []byte, err error) {
	jStruct := plainKeyJSON{
		k.Address,
		hex.EncodeToString(FromECDSA(k.PrivateKey)),
		k.Id.String(),
		version,
	}
	j, err = json.Marshal(jStruct)
	return j, err
}

func (k *Key) ToPrettyJSON() (j []byte, err error) {
	jStruct := plainKeyJSON{
		k.Address,
		hex.EncodeToString(FromECDSA(k.PrivateKey)),
		k.Id.String(),
		version,
	}
	j, err = json.MarshalIndent(jStruct, "", "\t")
	return j, err
}

func (k *Key) UnmarshalJSON(j []byte) (err error) {
	keyJSON := new(plainKeyJSON)
	err = json.Unmarshal(j, &keyJSON)
	if err != nil {
		return err
	}

	u := new(uuid.UUID)
	*u = uuid.Parse(keyJSON.Id)
	k.Id = *u
	privkey, err := HexToECDSA(keyJSON.PrivateKey)
	if err != nil {
		return err
	}

	k.Address = keyJSON.Address
	k.PrivateKey = privkey

	return nil
}

func NewKey(rand io.Reader) (*Key, error) {
	privateKeyECDSA, err := GeneratePrivateKey()
	if err != nil {
		return nil, err
	}
	id := uuid.NewRandom()
	address := PubkeyToAddress(privateKeyECDSA.PublicKey)

	key := &Key{
		Id:         id,
		Address:    address,
		PrivateKey: privateKeyECDSA,
	}
	return key, nil
}

func (k *Key) GetPrivateKeyBytes() []byte {
	return math.PaddedBigBytes(k.PrivateKey.D, 32)
}

func (k *Key) GetPrivateKeyString() string {
	return hex.EncodeToString(k.GetPrivateKeyBytes())
}

func (k *Key) GetAddressBytes() AddressBytes {
	return GetAddressBytes(k.Address)
}

func AddressBytesToAddressString(addressBytes AddressBytes) string {
	return hex.EncodeToString(addressBytes[:])
}

func HashBytesToHashString(bytes HashBytes) string {
	return hex.EncodeToString(bytes[:])
}
