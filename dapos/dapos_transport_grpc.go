/*
 *    This file is part of DAPoS library.
 *
 *    The DAPoS library is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *
 *    The DAPoS library is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *
 *    You should have received a copy of the GNU General Public License
 *    along with the DAPoS library.  If not, see <http://www.gnu.org/licenses/>.
 */
package dapos

import (
	"fmt"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/dispatchlabs/disgo/commons/services"
	"github.com/dispatchlabs/disgo/commons/types"
	"github.com/dispatchlabs/disgo/commons/utils"
	"github.com/dispatchlabs/disgo/dapos/proto"
	"github.com/dispatchlabs/disgo/disgover"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"strings"
	"github.com/dispatchlabs/disgo/commons/helper"
	"encoding/json"
	"math/big"
	"strconv"
)

// TODO: Should we GZIP the response from remote call?

var transactionMap map[int64][]*proto.Transaction
//var accountMap map[int64][]*proto.Account
var gossipMap map[int64][]*proto.Gossip
var refreshTimestamp = time.Now()

// WithGrpc -
func (this *DAPoSService) WithGrpc() *DAPoSService {
	proto.RegisterDAPoSGrpcServer(services.GetGrpcService().Server, this)
	return this
}

//func (this *DAPoSService) SynchronizeAccountsGrpc(context context.Context, request *proto.SynchronizeRequest) (*proto.SynchronizeAccountsResponse, error) {
//	utils.Info("synchronizing accounts with a delegate...")
//
//	if transactionMap == nil { //todo add OR if timestamp is old
//		loadMaps()
//	}
//	if transactionMap[request.Index] == nil {
//		utils.Info("DB synchronized Account")
//		transactionMap[request.Index] = make([]*proto.Transaction, 0)
//		fmt.Printf("\nCountMap: %v\n", helper.GetCounts().ToPrettyJson())
//	}
//	return &proto.SynchronizeAccountsResponse{Accounts: accountMap[request.Index]}, nil
//}

// SyncTransactions - PROTO - Called when a peer asks to sync missed TX, usually this happens at node boot
func (this *DAPoSService) SynchronizeTransactionsGrpc(context context.Context, request *proto.SynchronizeRequest) (*proto.SynchronizeTransactionsResponse, error) {
	utils.Info("synchronizing index", request.Index, " of transactions to a delegate...")


	if transactionMap == nil {
		utils.Info("loadingMaps")
		loadMaps()
	}
	if transactionMap[request.Index] == nil {
		utils.Info("DB synchronized transactions")
		transactionMap[request.Index] = make([]*proto.Transaction, 0)
		utils.Info("Is this even where CountMap is happening?")
		utils.Info("\nCountMap: %v\n", helper.GetCounts().ToPrettyJson())
		fmt.Printf("\nCountMap: %v\n", helper.GetCounts().ToPrettyJson())
	}
	return &proto.SynchronizeTransactionsResponse{Transactions: transactionMap[request.Index]}, nil
}

func (this *DAPoSService) SynchronizeGossipGrpc(context context.Context, request *proto.SynchronizeRequest) (*proto.SynchronizeGossipResponse, error) {
	utils.Info("synchronizing index", request.Index, " of gossips to a delegate...")

	if gossipMap == nil {
		loadMaps()
	}
	if gossipMap[request.Index] == nil {
		utils.Info("DB synchronized gossip")
		gossipMap[request.Index] = make([]*proto.Gossip, 0)
		fmt.Printf("\nCountMap: %v\n", helper.GetCounts().ToPrettyJson())
	}
	return &proto.SynchronizeGossipResponse{Gossips: gossipMap[request.Index]}, nil
}

/*Creates 3 maps: 1) Accounts, TXs, and Gossips
Each Map goes from an int page number to an array of *proto buff defined Bytes
Each Map has a page count that iterates every 50 map entries
 */
func loadMaps() error {
	//accountMap = map[int64][]*proto.Account{}
	transactionMap = map[int64][]*proto.Transaction{}
	gossipMap = map[int64][]*proto.Gossip{}


	var txPage int64 = 0
	//var acctPage int64 = 0
	var gossipPage int64 = 0

	err := services.GetDb().View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 100
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.Key()
			value, err := item.Value()
			if err != nil {
				return err
			}
			keyString := string(key)
			if !strings.HasPrefix(keyString, "table") {
				continue
			}
			if strings.HasPrefix(keyString, "table-transaction") {
				utils.Info(keyString)
				addTx(txPage, value)
				if len(transactionMap[txPage]) == 50 {
					txPage++
				}
			//} else if strings.HasPrefix(keyString, "table-account") {
			//	addAccount(acctPage, value)
			//	if len(accountMap[acctPage]) == 50 {
			//		acctPage++
			//	}
			} else if strings.HasPrefix(keyString, "table-gossip") {
				addGossip(gossipPage, value)
				if len(gossipMap[gossipPage]) == 50 {
					gossipPage++
				}
			} else {
				utils.Info("other prefix", keyString)
			}
		}
		return nil //no error
	})
	if err != nil {
		return err
	}
	return nil
}

/*The following addAccount, addTx, and addGossip functions are
called from the loadMaps() function. If the object unmarshals without error,
it is converted to its proto version and added to its appropriate Map.
 */
//func addAccount(acctPage int64, value []byte) {
//	if accountMap[acctPage] == nil {
//		accountMap[acctPage] = make([]*proto.Account, 0)
//	}
//	account := &types.Account{}
//	if err := json.Unmarshal(value, account); err != nil {
//		utils.Error(err)
//	} else {
//		pacct := convertToProtoAccount(account)
//		accountMap[acctPage] = append(accountMap[acctPage], pacct)
//
//		if !helper.ValidateAccountSync(account) {
//			utils.Error("Error validating this account")
//		}
//	}
//}

func addTx(txPage int64, value []byte) {
	if transactionMap[txPage] == nil {
		transactionMap[txPage] = make([]*proto.Transaction, 0)
	}
	transaction := &types.Transaction{}
	if err := json.Unmarshal(value, transaction); err != nil {
		utils.Error(err)
		//helper.HandleInvalidTransaction()
	} else {
		ptx := convertToProtoTransaction(transaction)
		transactionMap[txPage] = append(transactionMap[txPage], ptx)

		if !helper.ValidateTxSync(transaction) {
			utils.Error("Error validating this transaction")
		}
	}
}

func addGossip(gossipPage int64, value []byte) {
	if gossipMap[gossipPage] == nil {
		gossipMap[gossipPage] = make([]*proto.Gossip, 0)
	}
	gossip := &types.Gossip{}
	if err := json.Unmarshal(value, gossip); err != nil {
		utils.Error(err)
	} else {
		pgossip := convertToProtoGossip(gossip)
		gossipMap[gossipPage] = append(gossipMap[gossipPage], pgossip)

		if !helper.ValidateGossipSync(gossip) {
			utils.Error("Error validating this Gossip")
		}
	}
}

// SynchronizeGrpc -- Deprecated
func (this *DAPoSService) SynchronizeGrpc(constext context.Context, request *proto.SynchronizeRequest) (*proto.SynchronizeResponse, error) {
	utils.Info("synchronizing DB with a delegate...")
	var items = make([]*proto.Item, 0)
	err := services.GetDb().View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 100
		it := txn.NewIterator(opts)
		defer it.Close()
		var i int64 = 0
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.Key()
			value, err := item.Value()
			if err != nil {
				return err
			}
			keyString := string(key)
			if !strings.HasPrefix(keyString, "table-") && !strings.HasPrefix(keyString, "key-") && !strings.HasPrefix(keyString, "AccountState-") {
				continue
			}
			//fmt.Printf("Key: %s\n", keyString)

			//really inefficient fix this crap
			if i < request.Index {
				i++
				continue
			}

			if helper.ValidateSync(key, value) {
				if err != nil {
					utils.Error(err)
				}
			}
			items = append(items, &proto.Item{Key: keyString, Value: value})
			i++
			if len(items) == 50 {
				break
			}
		}
		if len(items) == 0 {
			utils.Info("DB synchronized")
			fmt.Printf("\nCountMap: %v\n", helper.GetCounts().ToPrettyJson())
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &proto.SynchronizeResponse{Items: items}, nil
}

// peerDelegateExecuteGrpc //Caller
func (this *DAPoSService) peerSynchronize() {
	utils.Info("synchronizing DB with peer delegate...")


	// Find delegate nodes.
	delegates, err := types.ToNodesByTypeFromCache(services.GetCache(),types.TypeDelegate)
	if err != nil {
		utils.Error(err)
		return
	}
	if len(delegates) == 0 {
		utils.Warn("unable to find a delegate to synchronize with")
		return
	}

	utils.Info("number of delegates is ", len(delegates))

	//This isn't actually looping through all of the delegates there is a return in the loop
	for _, delegate := range delegates {

		fmt.Printf("going to try to connect to delegate %s\n", delegate.Address)

		// Is this me?
		if delegate.Address == disgover.GetDisGoverService().ThisNode.Address {
			fmt.Println("This delegate is me")
			continue
		}
		// Connect to delegate.
		conn, err := grpc.Dial(fmt.Sprintf("%s:%d", delegate.GrpcEndpoint.Host, delegate.GrpcEndpoint.Port), grpc.WithInsecure())
		if err != nil {
			utils.Warn(fmt.Sprintf("unable to connect to delegate [host=%s, port=%d]", delegate.GrpcEndpoint.Host, delegate.GrpcEndpoint.Port), err)
			continue
		} else {
			utils.Info("connecting to delegate %s worked", delegate.Address)
		}
		defer conn.Close()
		client := proto.NewDAPoSGrpcClient(conn)
		contextWithTimeout, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Synchronize
		var txIndex int64 = 0
		var gossipIndex int64 = 0
		count := 0
		//This For loop indexes through the sending delegates TXs
		for {
			//Open up a new badger db tx
			txn := services.NewTxn(true)
			defer txn.Discard()

			//Synchronizing Transactions First
			txResponse, err := client.SynchronizeTransactionsGrpc(contextWithTimeout, &proto.SynchronizeRequest{Index: txIndex})
			if err != nil {
				utils.Warn(fmt.Sprintf("unable to synchronize with delegate [host=%s, port=%d]", delegate.GrpcEndpoint.Host, delegate.GrpcEndpoint.Port), err)
				continue
			}
			if len(txResponse.Transactions) == 0 {
				break
			}


			/*
			 * Adding code to avoid issues we are having with corrupt data coming from sync
			 * Concerned about the case where current delegate might have an old version of a key value
			 * At the moment this is so much less of a concern than getting an invalid JSON
			 * that I'm willing to let this be the case until we get the Merkle tree implemented. (B.S.)
			 */
			//for _, item := range response.Items {
			//	exists, _ := txn.Get([]byte(item.Key))
			//	if exists == nil {
			//		if helper.ValidateSync([]byte(item.Key), item.Value) {
			//			err = txn.Set([]byte(item.Key), item.Value)
			//			if err != nil {
			//				utils.Error(err)
			//			}
			//		}
			//	} else {
			//		utils.Info(fmt.Sprintf("skipping key: %s ", item.Key))
			//	}
			//}
			//index += int64(len(response.Items))

			//Takes newly received transactions and persists them to badger
			for _, ptx := range txResponse.Transactions {
				count++
				tx := convertToDomainTransaction(ptx)
				exists, _ := txn.Get([]byte(tx.Key()))
				if exists == nil {
					err = tx.Persist(txn)
					if err != nil {
						utils.Error(err)
					}
				}
				if helper.ValidateTxSync(tx) {
					if err != nil {
						utils.Error(err)
					}
				}
			}
			//index += int64(len(response.Transactions))
			txIndex++
			err = txn.Commit(nil)
			if err != nil {
				utils.Error(err)
			}
		}

		//this for loop indexes through the sending delegate's gossips
		for {
			//Open up a new Badger db TX
			txn := services.NewTxn(true)
			defer txn.Discard()

			//Call the next index of gossips from the sender
			gossipResponse, err := client.SynchronizeGossipGrpc(contextWithTimeout, &proto.SynchronizeRequest{Index: gossipIndex})
			if err != nil {
				utils.Warn(fmt.Sprintf("unable to synchronize with delegate [host=%s, port=%d]", delegate.GrpcEndpoint.Host, delegate.GrpcEndpoint.Port), err)
				continue
			}
			if len(gossipResponse.Gossips) == 0 {
				break
			}


			//Takes newly received Gossips and Persists them to badger
			for _, pg := range gossipResponse.Gossips {
				count++
				gossip := convertToDomainGossip(pg)
				exists, _ := txn.Get([]byte(gossip.Key()))
				if exists == nil {
					err = gossip.Persist(txn)
					if err != nil {
						utils.Error(err)
					}
				}
				if helper.ValidateGossipSync(gossip) {
					if err != nil {
						utils.Error(err)
					}
				}
			}
			//index += int64(len(response.Transactions))
			gossipIndex++
			err = txn.Commit(nil)
			if err != nil {
				utils.Error(err)
			}

		}

		utils.Info("this is where the Last CountMap gets printed")
		utils.Info("\nCountMap: %v\n", helper.GetCounts().ToPrettyJson())
		fmt.Printf("\nCountMap: %v\n", helper.GetCounts().ToPrettyJson())
		utils.Info(fmt.Sprintf("synchronized %d records from peer delegate %s's DB", count, delegate.Address))

		return
		//This return means that db is only synced with the first (non-self) delegate returned by the seed node
		//That's good to know because basically the for loop is nearly pointless
	}
}

// Gossip
func (this *DAPoSService) GossipGrpc(context context.Context, request *proto.Request) (*proto.Response, error) {
	gossip, err := types.ToGossipFromJson([]byte(request.Payload))
	if err != nil {
		utils.Error(err)
		return nil, err
	}

	// Synchronize gossip.
	synchronizedGossip, err, addToChan := this.synchronizeGossip(gossip)
	if err != nil {
		utils.Error(err)
		return nil, err
	}

	// Gossip what we got from our peer delegate.
	if(addToChan) {
		this.gossipChan <- gossip
	}

	return &proto.Response{Payload: synchronizedGossip.String()}, nil
}

// peerGossipGrpc
func (this *DAPoSService) peerGossipGrpc(node types.Node, gossip *types.Gossip) (*types.Gossip, error) {
	utils.Debug(fmt.Sprintf("attempting to gossip with delegate [address=%s]", node.Address))

	conn, err := services.GetGrpcConnection(node.Address, node.GrpcEndpoint.Host, node.GrpcEndpoint.Port)
	if err != nil {
		utils.Error(fmt.Sprintf("cannot dial seed [host=%s, port=%d]",  node.GrpcEndpoint.Host,  node.GrpcEndpoint.Port), err)
		return nil, err
	}
	client := proto.NewDAPoSGrpcClient(conn)

	contextWithTimeout, cancel := context.WithTimeout(context.Background(), 20000*time.Millisecond)
	defer cancel()

	// Remote gossip.
	response, err := client.GossipGrpc(contextWithTimeout, &proto.Request{Payload: gossip.String()})
	if err != nil {
		utils.Error(fmt.Sprintf("cannot connect to node [host=%s, port=%d]", node.GrpcEndpoint.Host, node.GrpcEndpoint.Port), err)

		txn := services.NewTxn(true)
		defer txn.Discard()
		node.Status = types.StatusNodeUnavailable
		node.StatusTime = time.Now()

		setErr := node.Set(txn, services.GetCache())
		if setErr != nil {
			utils.Error(setErr)
		}
		return nil, err
	}
	remoteGossip, err := types.ToGossipFromJson([]byte(response.Payload))
	if err != nil {
		utils.Error(err)
		return nil, err
	}
	utils.Debug(fmt.Sprintf("sent gossip [hash=%s] to delegate [Port %d] [address=%s]", gossip.Transaction.Hash, node.HttpEndpoint.Port, node.Address))
	remoteGossip.CacheSentDelegate(services.GetCache(), gossip.Transaction.Hash, node.Address)

	return remoteGossip, err
}

func convertToProtoAccount(acct *types.Account) *proto.Account {
	return &proto.Account{
		Address:			acct.Address,
		Name:				acct.Name,
		Balance:			acct.Balance.String(),
		HertzAvailable:		acct.HertzAvailable,
		TransactionHash:	acct.TransactionHash,
		Created:			utils.ToMilliSeconds(acct.Created),
		Updated:			utils.ToMilliSeconds(acct.Updated),
		Nonce:				acct.Nonce,
	}
}

func convertToDomainAccount(pacct *proto.Account) *types.Account {
	bFloat, err := strconv.ParseFloat(pacct.Balance, 64)
	if err != nil {
		utils.Error("failed to parse balance")
	}
	return &types.Account{
		Address:         pacct.Address,
		Name:            pacct.Name,
		Balance:         big.NewInt(int64(bFloat)),
		HertzAvailable:  pacct.HertzAvailable,
		TransactionHash: pacct.TransactionHash,
		Created:         utils.ToTimeFromMilliseconds(pacct.Created),
		Updated:         utils.ToTimeFromMilliseconds(pacct.Updated),
		Nonce:           pacct.Nonce,
	}
}

func convertToProtoTransaction(tx *types.Transaction) *proto.Transaction {
	return &proto.Transaction{
		Hash:		tx.Hash,
		Type:		int32(tx.Type),
		From:      	tx.From,
		To:        	tx.To,
		Value:     	tx.Value,
		Code:		tx.Code,
		Abi:		tx.Abi,
		Method:		tx.Method,
		Params:		tx.Params,
		Time:      	tx.Time,
		Signature: 	tx.Signature,
		Hertz:		tx.Hertz,
		FromName:	tx.FromName,
		ToName:		tx.ToName,
	}
}

func convertToDomainTransaction(ptx *proto.Transaction) *types.Transaction {
	return &types.Transaction{
		Hash:      	ptx.Hash,
		Type:		byte(ptx.Type),
		From:      	ptx.From,
		To:        	ptx.To,
		Value:     	ptx.Value,
		Code:		ptx.Code,
		Abi:		ptx.Abi,
		Method:		ptx.Method,
		Params:		ptx.Params,
		Time:      	ptx.Time,
		Signature: 	ptx.Signature,
		Hertz:		ptx.Hertz,
		FromName:	ptx.FromName,
		ToName:		ptx.ToName,
	}
}

func convertToProtoRumor(rumor *types.Rumor) *proto.Rumor {
	return &proto.Rumor{
		Hash: 				rumor.Hash,
		Address:			rumor.Address,
		TransactionHash:	rumor.TransactionHash,
		Time:				rumor.Time,
		Signature:			rumor.Signature,
	}
}

func convertToDomainRumor(prumor *proto.Rumor) *types.Rumor {
	return &types.Rumor{
		Hash:            prumor.Hash,
		Address:         prumor.Address,
		TransactionHash: prumor.TransactionHash,
		Time:            prumor.Time,
		Signature:       prumor.Signature,
	}
}

func convertToProtoGossip(gossip *types.Gossip) *proto.Gossip {
	prumors := make([]*proto.Rumor, 0)
	for _, rumor := range gossip.Rumors {
		prumors = append(prumors, convertToProtoRumor(&rumor))
	}
	return &proto.Gossip{
		TxHash:		gossip.Transaction.Hash,
		Rumors: 	prumors,
	}
}

func convertToDomainGossip(pgossip *proto.Gossip) *types.Gossip {
	rumors := make([]types.Rumor, 0)
	for _, prumor := range pgossip.Rumors {
		rumors = append(rumors, *convertToDomainRumor(prumor))
	}
	return &types.Gossip{
		Transaction: 	types.Transaction{Hash: pgossip.TxHash},
		Rumors:			rumors,
	}
}
