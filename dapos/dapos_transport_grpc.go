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
)

// TODO: Should we GZIP the response from remote call?

// WithGrpc -
func (this *DAPoSService) WithGrpc() *DAPoSService {
	proto.RegisterDAPoSGrpcServer(services.GetGrpcService().Server, this)
	return this
}


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
			//fmt.Printf("Key: %s\n", keyString)
			if !strings.HasPrefix(keyString, "table-") && !strings.HasPrefix(keyString, "key-") { //&& !strings.HasPrefix(keyString, "AccountState-") {
				continue
			}
			if i < request.Index {
				i++
				continue
			}
			items = append(items, &proto.Item{Key: keyString, Value: string(value)})
			i++
			if len(items) == 50 {
				break
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	utils.Info("DB synchronized")
	return &proto.SynchronizeResponse{Items: items}, nil
}

// peerDelegateExecuteGrpc
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
		for {
			txn := services.NewTxn(true)
			defer txn.Discard()
			response, err := client.SynchronizeGrpc(contextWithTimeout, &proto.SynchronizeRequest{Index: index})
			if err != nil {
				utils.Warn(fmt.Sprintf("unable to synchronize with delegate [host=%s, port=%d]", delegate.GrpcEndpoint.Host, delegate.GrpcEndpoint.Port), err)
				continue
			}
			if len(response.Items) == 0 {
				break
			}
			for _, item := range response.Items {
				err = txn.Set([]byte(item.Key), []byte(item.Value))
				if err != nil {
					utils.Error(err)
				}
			}
			index += int64(len(response.Items))
			err = txn.Commit(nil)
			if err != nil {
				utils.Error(err)
			}
		}
		utils.Info(fmt.Sprintf("synchronized %d records from peer delegate's DB", index))
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
	synchronizedGossip, err := this.synchronizeGossip(gossip)
	if err != nil {
		utils.Error(err)
		return nil, err
	}

	// Gossip what we got from our peer delegate.
	this.gossipChan <- gossip

	return &proto.Response{Payload: synchronizedGossip.String()}, nil
}

// peerGossipGrpc
func (this *DAPoSService) peerGossipGrpc(node types.Node, gossip *types.Gossip) (*types.Gossip, error) {
	utils.Debug(fmt.Sprintf("attempting to gossip with delegate [address=%s]", node.Address))

	//conn, err := grpc.Dial(fmt.Sprintf("%s:%d", node.GrpcEndpoint.Host, node.GrpcEndpoint.Port), grpc.WithInsecure())
	//if err != nil {
	//	utils.Error(fmt.Sprintf("cannot dial seed [host=%s, port=%d]",  node.GrpcEndpoint.Host,  node.GrpcEndpoint.Port), err)
	//	return nil, err
	//}
	//defer conn.Close()

	conn, err := services.GetGrpcConnection(node.Address, node.GrpcEndpoint.Host, node.GrpcEndpoint.Port)

	client := proto.NewDAPoSGrpcClient(conn)

	contextWithTimeout, cancel := context.WithTimeout(context.Background(), 20000*time.Millisecond)
	defer cancel()

	// Remote gossip.
	response, err := client.GossipGrpc(contextWithTimeout, &proto.Request{Payload: gossip.String()})
	if err != nil {
		utils.Error(fmt.Sprintf("cannot connect to node [host=%s, port=%d]", node.GrpcEndpoint.Host, node.GrpcEndpoint.Port), err)
		//txn := services.NewTxn(true)
		//defer txn.Discard()
		//
		//unsetErr := node.Unset(txn, services.GetCache())
		//if unsetErr != nil {
		//	utils.Error(unsetErr)
		//}
		return nil, err
	}
	remoteGossip, err := types.ToGossipFromJson([]byte(response.Payload))
	if err != nil {
		return nil, err
	}

	return remoteGossip, err
}
