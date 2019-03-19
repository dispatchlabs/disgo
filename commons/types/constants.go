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


const (
	HertzMultiplier = 1
)
//TODO: I think we need to convert these timouts and their calculations in code to nano seconds
// Timouts -- currently calculated in milliseconds.
const (
	TxReceiveTimeout = 3000 //1 second
	//GossipQueueTimeout = time.Second * 5
	GossipTimeout = 1000 //1 second  //will continue to decrease until we find best value
	TxFutureLimit = time.Minute * 3
	UnavailableNodeTimeout = float64(time.Second * 5)
)


// Statuses
const (
	StatusReceived                     = "Received"
	StatusPending                      = "Pending"
	StatusOk                           = "Ok"
	StatusNotFound                     = "NotFound"
	StatusReceiptNotFound              = "StatusReceiptNotFound"
	StatusTransactionTimeOut           = "StatusTransactionTimeOut"
	StatusInvalidTransaction           = "InvalidTransaction"
	StatusInsufficientTokens           = "InsufficientTokens"
	StatusInsufficientHertz            = "InsufficientHertz"
	StatusDuplicateTransaction         = "DuplicateTransaction"
	StatusNotDelegate                  = "StatusNotDelegate"
	StatusAlreadyProcessingTransaction = "StatusAlreadyProcessingTransaction"
	StatusGossipingTimedOut            = "StatusGossipingTimedOut"
	StatusJsonParseError               = "StatusJsonParseError"
	StatusInternalError                = "InternalError"
	StatusUnavailableFeature           = "UnavailableFeature"
	StatusNodeUnavailable              = "NodeUnavailable"
	StatusCouldNotReachConsensus       = "CouldNotReachConsensus"
)

const (
	StatusNotDelegateAsHumanReadable = "This node is not a delegate. Please select a delegate node."
)

// Types
const (
	TypeSeed                 = "Seed"
	TypeDelegate             = "Delegate"
	TypeNode                 = "Node"
	TypeTransferTokens       = 0
	TypeDeploySmartContract  = 1
	TypeExecuteSmartContract = 2
	TypeReadSmartContract	 = 3
)

// Persistence TTLs
const (
	AccountTTL = time.Hour * 24
	PageTTL    = time.Hour * 24
)

// Cache TTLs
const (
	CacheTTL               = time.Hour
	TransactionCacheTTL    = time.Hour * 48
	ReceiptCacheTTL        = time.Hour * 48
	GossipCacheTTL         = time.Minute * 5
	AuthenticationCacheTTL = time.Minute
	RateLimitAverageTTL    = time.Minute * 240
)

// Errors
var (
	ErrNotFound               = errors.New("not found")
	ErrInvalidRequest         = errors.New("invalid request")
	ErrInvalidRequestPage     = errors.New("invalid request Page")
	ErrInvalidRequestPageSize = errors.New("invalid request Page Size")
	ErrInvalidRequestStartingHash = errors.New("invalid request Starting Hash")
	ErrInvalidRequestHash     = errors.New("invalid request Hash")
)
