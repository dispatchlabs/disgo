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
	"github.com/pkg/errors"
	"github.com/processout/grpc-go-pool"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// TODO: Should we GZIP the response from remote call?

// WithGrpc -
func (this *DAPoSService) WithGrpc() *DAPoSService {
	proto.RegisterDAPoSGrpcServer(services.GetGrpcService().Server, this)
	return this
}

// Synchronize
func (this *DAPoSService) SynchronizeGrpc(context.Context, *proto.Empty) (*proto.SynchronizeResponse, error) {
	utils.Info("synchronizing DB with a delegate...")
	var items = make([]*proto.Item, 0)
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
			items = append(items, &proto.Item{Key: string(key), Value: string(value)})
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
	utils.Info("synchronizing DB with a delegate...")

	// TODO: This should be done during elections.
	// Find delegate nodes.
	nodes, err := types.ToNodesByTypeFromCache(services.GetCache(),types.TypeDelegate)
	if err != nil {
		utils.Error(err)
		return
	}
	if len(nodes) == 0 {
		utils.Warn("unable to find a delegate to synchronize with")
		return
	}
	for _, node := range nodes {

		// Check if it is the same node
		if node.Address == disgover.GetDisGoverService().ThisNode.Address {
			continue
		}

		foundNode, err := disgover.GetDisGoverService().Find(node.Address)
		if err != nil {
			utils.Warn(fmt.Sprintf("'%s' with [host=%s, port=%d]", node.Address, node.Endpoint.Host, node.Endpoint.Port), err)
			continue
		}

		if foundNode == nil {
			continue
		}

		// Connect to delegate.
		conn, err := grpc.Dial(fmt.Sprintf("%s:%d", foundNode.Endpoint.Host, foundNode.Endpoint.Port), grpc.WithInsecure())
		if err != nil {
			utils.Warn(fmt.Sprintf("unable to connect to delegate [host=%s, port=%d]", foundNode.Endpoint.Host, foundNode.Endpoint.Port), err)
			continue
		}
		defer conn.Close()
		client := proto.NewDAPoSGrpcClient(conn)
		contextWithTimeout, cancel := context.WithTimeout(context.Background(), 2000*time.Millisecond)
		defer cancel()

		// Synchronize
		response, err := client.SynchronizeGrpc(contextWithTimeout, &proto.Empty{})
		if err != nil {
			utils.Warn(fmt.Sprintf("unable to synchronize with delegate [host=%s, port=%d]", foundNode.Endpoint.Host, foundNode.Endpoint.Port), err)
			continue
		}
		txn := services.NewTxn(true)
		defer txn.Discard()
		for _, item := range response.Items {
			err = txn.Set([]byte(item.Key), []byte(item.Value))
			if err != nil {
				utils.Error(err)
			}
		}
		txn.Commit(nil)
		utils.Info("DB synchronized from peer delegate")
		return
	}
}

// Execute
func (this *DAPoSService) ExecuteGrpc(context context.Context, request *proto.Request) (*proto.Response, error) {

	// Type?
	switch request.Type {
	case types.RequestGetAccount:
		return &proto.Response{Payload: this.GetAccount(request.Payload).String()}, nil
	case types.RequestGetStatus:
		return &proto.Response{Payload: this.GetStatus(request.Payload).String()}, nil
	case types.RequestNewTransaction:
		transaction, err := types.ToTransactionFromJson([]byte(request.Payload))
		if err != nil {
			return &proto.Response{Payload: types.NewReceiptWithError(types.RequestNewTransaction, err).String()}, nil
		}
		return &proto.Response{Payload: this.NewTransaction(transaction).String()}, nil
	case types.RequestGetTransactionsByToAddress:
		return &proto.Response{Payload: this.GetTransactionsByToAddress(request.Payload).String()}, nil
	case types.RequestGetTransactions:
		return &proto.Response{Payload: this.GetTransactions().String()}, nil
	}
	return &proto.Response{Payload: types.NewReceiptWithStatus(request.Type, types.StatusInvalidRequest, "invalid request").String()}, nil
}

// peerDelegateExecuteGrpc
func (this *DAPoSService) peerDelegateExecuteGrpc(tipe string, payload string) *types.Receipt {

	// TODO: This should be done during elections.
	// Find delegate nodes.
	nodes, err := types.ToNodesByTypeFromCache(services.GetCache(),types.TypeDelegate)
	if err != nil {
		utils.Error(err)
		return types.NewReceiptWithError(tipe, err)
	}
	if len(nodes) == 0 {
		utils.Warn("unable to find a delegate to execute request")
		return types.NewReceiptWithStatus(tipe, types.StatusUnableToFindDelegates, "unable to find any delegates")
	}

	for _, node := range nodes {

		// Check if it is the same node
		if node.Address == disgover.GetDisGoverService().ThisNode.Address {
			continue
		}

		foundNode, err := disgover.GetDisGoverService().Find(node.Address)
		if err != nil {
			utils.Warn(fmt.Sprintf("'%s' with [host=%s, port=%d]", node.Address, node.Endpoint.Host, node.Endpoint.Port), err)
			continue
		}

		if foundNode == nil {
			continue
		}

		return this.peerExecuteGrpc(*foundNode, tipe, payload)
	}

	return types.NewReceiptWithStatus(tipe, types.StatusUnableToConnectToDelegate, "unable to connect to a delegate")
}

// remoteExecuteWithContact
func (this *DAPoSService) peerExecuteGrpc(node types.Node, tipe string, payload string) *types.Receipt {

	// Connect.
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", node.Endpoint.Host, node.Endpoint.Port), grpc.WithInsecure())
	if err != nil {
		return types.NewReceiptWithStatus(tipe, types.StatusUnableToConnectToDelegate, "unable to connect to a delegate")
	}
	defer conn.Close()
	client := proto.NewDAPoSGrpcClient(conn)
	contextWithTimeout, cancel := context.WithTimeout(context.Background(), 2000*time.Millisecond)
	defer cancel()

	// Remote execute.
	response, err := client.ExecuteGrpc(contextWithTimeout, &proto.Request{Type: tipe, Payload: payload})
	if err != nil {
		return types.NewReceiptWithError(tipe, err)
	}
	receipt, err := types.ToReceiptFromJson([]byte(response.Payload))
	if err != nil {
		return types.NewReceiptWithStatus(tipe, types.StatusInternalError, err.Error())
	}
	return receipt
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

	value, ok := services.GetCache().Get(fmt.Sprintf("dapos-grpc-pool-%s", node.Address))
	// IF not found then setup one
	if !ok {
		this.setupConnectionPoolForPeer(&node)
	}
	value, ok = services.GetCache().Get(fmt.Sprintf("dapos-grpc-pool-%s", node.Address))
	if !ok {
		return nil, errors.New(fmt.Sprintf("unable to find GRPC pool for this delegate [address=%s]", node.Address))
	}

	pool := value.(*grpcpool.Pool)
	clientConn, err := pool.Get(context.Background())
	defer clientConn.Close()

	client := proto.NewDAPoSGrpcClient(clientConn.ClientConn)
	contextWithTimeout, cancel := context.WithTimeout(context.Background(), 2000*time.Millisecond)
	defer cancel()

	// Remote gossip.
	response, err := client.GossipGrpc(contextWithTimeout, &proto.Request{Payload: gossip.String()})
	if err != nil {
		return nil, err
	}
	remoteGossip, err := types.ToGossipFromJson([]byte(response.Payload))
	if err != nil {
		return nil, err
	}

	return remoteGossip, err
}
