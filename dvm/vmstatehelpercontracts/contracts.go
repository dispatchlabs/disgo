/*
 *    This file is part of DVM library.
 *
 *    The DVM library is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *
 *    The DVM library is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *
 *    You should have received a copy of the GNU General Public License
 *    along with the DVM library.  If not, see <http://www.gnu.org/licenses/>.
 */
package vmstatehelpercontracts

import (
	"github.com/dispatchlabs/disgo/commons/crypto"
	ethState "github.com/dispatchlabs/disgo/dvm/ethereum/state"
)

// VMStateQueryHelper - Decouples and helps load code/size from
// the persistence layer + loads the Particia Merkle Trie
type VMStateQueryHelper interface {
	GetCode(smartContractAddress crypto.AddressBytes) []byte
	GetCodeSize(executingContractAddress crypto.AddressBytes, callerAddress crypto.AddressBytes, toBeExecutedContractAddress crypto.AddressBytes) int
	NewEthStateLoader(smartContractAddress crypto.AddressBytes) *ethState.StateDB
}
