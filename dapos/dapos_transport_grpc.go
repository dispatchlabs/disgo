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
)

// TODO: Should we GZIP the response from remote call?

// WithGrpc -
func (this *DAPoSService) WithGrpc() *DAPoSService {
	proto.RegisterDAPoSGrpcServer(services.GetGrpcService().Server, this)
	return this
}

func (this *DAPoSService) SynchronizeAccountsGrpc(constext context.Context, request *proto.SynchronizeRequest) (*proto.SynchronizeAccountsResponse, error) {
	return nil, nil;
}

var txMap map[int64][]*proto.Transaction

// SyncTransactions - PROTO - Called when a peer asks to sync missed TX, usually this happens at node boot
func (this *DAPoSService) SynchronizeTransactionsGrpc(constext context.Context, request *proto.SynchronizeRequest) (*proto.SynchronizeTransactionsResponse, error) {
	utils.Info("synchronizing DB with a delegate...")
	if txMap == nil {
		txMap = map[int64][]*proto.Transaction{}
	}
	var page int64 = 0
	if txMap[0] == nil {
		err := services.GetDb().View(func(txn *badger.Txn) error {
			opts := badger.DefaultIteratorOptions
			opts.PrefetchSize = 100
			it := txn.NewIterator(opts)
			defer it.Close()
			for it.Rewind(); it.Valid(); it.Next() {
				if txMap[page] == nil {
					txMap[page] = make([]*proto.Transaction, 0)
				}
				item := it.Item()
				key := item.Key()
				value, err := item.Value()
				if err != nil {
					return err
				}
				keyString := string(key)
				if !strings.HasPrefix(keyString, "table-transaction") {
					continue
				}

				transaction := &types.Transaction{}
				if err := json.Unmarshal(value, transaction); err == nil {
					utils.Error(err)
				}
				ptx := convertToProto(transaction)
				txMap[page] = append(txMap[page], ptx)

				if helper.ValidateTxSync(transaction) {
					if err != nil {
						utils.Error(err)
					}
				}
				if len(txMap[page]) == 50 {
					page++
				}
			}
			return nil //no error
		})
		if err != nil {
			return nil, err
		}
	}
	if txMap[request.Index] == nil {
		utils.Info("DB synchronized transactions")
		txMap[request.Index] = make([]*proto.Transaction, 0)
		fmt.Printf("\nCountMap: %v\n", helper.GetCounts().ToPrettyJson())
	}
	return &proto.SynchronizeTransactionsResponse{Transactions: txMap[request.Index]}, nil}

// SynchronizeGrpc
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
	//This isn't actually looping through all of the delegates there is a return in the loop
	for _, delegate := range delegates {

		// Is this me?
		if delegate.Address == disgover.GetDisGoverService().ThisNode.Address {
			continue
		}
		// Connect to delegate.
		conn, err := grpc.Dial(fmt.Sprintf("%s:%d", delegate.GrpcEndpoint.Host, delegate.GrpcEndpoint.Port), grpc.WithInsecure())
		if err != nil {
			utils.Warn(fmt.Sprintf("unable to connect to delegate [host=%s, port=%d]", delegate.GrpcEndpoint.Host, delegate.GrpcEndpoint.Port), err)
			continue
		}
		defer conn.Close()
		client := proto.NewDAPoSGrpcClient(conn)
		contextWithTimeout, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Synchronize
		var index int64 = 0
		count := 0
		for {
			txn := services.NewTxn(true)
			defer txn.Discard()

			//response, err := client.SynchronizeGrpc(contextWithTimeout, &proto.SynchronizeRequest{Index: index})
			response, err := client.SynchronizeTransactionsGrpc(contextWithTimeout, &proto.SynchronizeRequest{Index: index})
			if err != nil {
				utils.Warn(fmt.Sprintf("unable to synchronize with delegate [host=%s, port=%d]", delegate.GrpcEndpoint.Host, delegate.GrpcEndpoint.Port), err)
				continue
			}
			if len(response.Transactions) == 0 {
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

			for _, ptx := range response.Transactions {
				count++
				tx := convertToDomain(ptx)
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
			index++
			err = txn.Commit(nil)
			if err != nil {
				utils.Error(err)
			}
		}
		fmt.Printf("\nCountMap: %v\n", helper.GetCounts().ToPrettyJson())
		utils.Info(fmt.Sprintf("synchronized %d records from peer delegate's DB", count))
		return
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

func convertToProto(tx *types.Transaction) *proto.Transaction {
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

func convertToDomain(ptx *proto.Transaction) *types.Transaction {
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
