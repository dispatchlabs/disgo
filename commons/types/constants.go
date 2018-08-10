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
package types

import (
	"errors"
	"time"
)

//TODO: I think we need to convert these timouts and their calculations in code to nano seconds
// Timouts -- currently calculated in milliseconds.
const (
	TxReceiveTimeout    = 10000  //1 second
	GossipQueueTimeout  = time.Second * 5
	GossipTimeout    	= 30000  //300 milliseconds
)

// Requests
const (
	Version = "2.2.0"
)

// Statuses
const (
	StatusPending                      = "Pending"
	StatusOk                           = "Ok"
	StatusNotFound                     = "NotFound"
	StatusReceiptNotFound              = "StatusReceiptNotFound"
	StatusTransactionTimeOut           = "StatusTransactionTimeOut"
	StatusInvalidTransaction           = "InvalidTransaction"
	StatusInsufficientTokens           = "InsufficientTokens"
	StatusDuplicateTransaction         = "DuplicateTransaction"
	StatusNotDelegate                  = "StatusNotDelegate"
	StatusAlreadyProcessingTransaction = "StatusAlreadyProcessingTransaction"
	StatusGossipingTimedOut            = "StatusGossipingTimedOut"
	StatusJsonParseError               = "StatusJsonParseError"
	StatusInternalError                = "InternalError"
	StatusUnavailableFeature 		   = "UnavailableFeature"
)

// Types
const (
	TypeSeed                 = "Seed"
	TypeDelegate             = "Delegate"
	TypeNode                 = "Node"
	TypeTransferTokens       = 0
	TypeDeploySmartContract  = 1
	TypeExecuteSmartContract = 2
)

// Persistence TTLs
const (
	ReceiptTTL     = time.Hour * 24 * 3
	GossipTTL      = time.Hour * 48
	NodeTTL        = time.Hour * 24
	AccountTTL     = time.Hour * 24
	PageTTL        = time.Hour * 24
	TransactionTTL = time.Hour * 48
)

// Cache TTLs
const (
	CacheTTL        = time.Hour
	ReceiptCacheTTL = time.Minute * 30
	GossipCacheTTL  = time.Minute * 5
)

// Errors
var (
	ErrNotFound = errors.New("not found")
	ErrInvalidRequest = errors.New("invalid request")
)
