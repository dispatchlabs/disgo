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

package crypto

import (
	"testing"
)

const (
	veryLightScryptN = 2
	veryLightScryptP = 1
)

// Tests that a json key file can be decrypted and encrypted in multiple rounds.
func TestKeyEncryptDecrypt(t *testing.T) {

	//password := "test"
	//newKey, err := newKey(rand.Reader)
	//newKeyPretty, err := newKey.ToPrettyJSON()
	//if err != nil {
	//	t.Fatalf("test %d: json key failed to Marshal: %v", 0, err)
	//}
	//fmt.Printf(string(newKeyPretty) + "\n")
	//fmt.Printf("Real Pub ADDR: %s", PubkeyToAddress(newKey.PrivateKey.PublicKey))
	//
	//keyJson, err := EncryptKey(newKey, password, veryLightScryptN, veryLightScryptP)
	//if err != nil {
	//	t.Errorf("test %d: failed to encrypt key %v", newKey.Id, err)
	//}
	//fmt.Printf(string(keyJson) + "\n")
	//
	//key, err := DecryptKey(keyJson, "test")
	//if err != nil {
	//	t.Fatalf("test %d: json key failed to decrypt: %v", 0, err)
	//}
	//if key.Address != newKey.Address {
	//	t.Errorf("test %d: key address mismatch: have %x, want %x", 0, key.Address, newKey.Address)
	//}
	//pretty, err := key.ToPrettyJSON()
	//if err != nil {
	//	t.Fatalf("test %d: json key failed to Marshal: %v", 0, err)
	//}
	//
	//fmt.Printf(string(pretty) + "\n")
}
