/*
 *    This file is part of Disgover library.
 *
 *    The Disgover library is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *
 *    The Disgover library is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *
 *    You should have received a copy of the GNU General Public License
 *    along with the Disgover library.  If not, see <http://www.gnu.org/licenses/>.
 */
package disgover

import (
	"fmt"

	"time"

	"github.com/dispatchlabs/disgo/commons/services"
	"github.com/dispatchlabs/disgo/commons/types"
	"github.com/dispatchlabs/disgo/commons/utils"
	proto "github.com/dispatchlabs/disgo/disgover/proto"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"github.com/patrickmn/go-cache"
)

// WithGrpc - Runs the DisGover service with GRPC transport
func (this *DisGoverService) WithGrpc() *DisGoverService {
	proto.RegisterDisgoverGrpcServer(services.GetGrpcService().Server, this)
	return this
}

// PingSeedGrpc
func (this *DisGoverService) PingSeedGrpc(ctx context.Context, node *proto.Node) (*proto.NodeList, error) {

	// Is this node a seed?
	if this.ThisNode.Type != types.TypeSeed {
		return nil, errors.New("you pinged a non-seed node")
	}

	// Persist and cache  delegate.
	txn := services.NewTxn(true)
	defer txn.Discard()
	convertToDomain(node).PersistAndCache(txn, services.GetCache())

	// Get cached delegates.
	delegates, err := types.ToNodesByTypeFromCache(services.GetCache(), types.TypeDelegate)
	if err != nil {
		return nil, err
	}

	var nodes = make([]*proto.Node, 0)
	for _, delegate := range delegates {
		nodes = append(nodes, convertToProto(delegate))
	}

	utils.Info(fmt.Sprintf("received ping [address=%s, ip=%s, port=%d, delegates=%d]", node.Address, node.Endpoint.Host, node.Endpoint.Port, len(delegates)))

	// Update all peers.
	this.peerUpdateGrpc()

	return &proto.NodeList{Delegates: nodes}, nil
}

// peerPingSeedGrpc
func (this *DisGoverService) peerPingSeedGrpc() ([]*types.Node, error) {

	var delegates = make([]*types.Node, 0)
	for _, seedEndpoint := range types.GetConfig().SeedEndpoints {
		conn, err := grpc.Dial(fmt.Sprintf("%s:%d", seedEndpoint.Host, seedEndpoint.Port), grpc.WithInsecure())
		if err != nil {
			utils.Fatal(fmt.Sprintf("cannot dial seed [host=%s, port=%d]", seedEndpoint.Host, seedEndpoint.Port), err)
			return nil, err
		}
		defer conn.Close()
		client := proto.NewDisgoverGrpcClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Ping seed.
		response, err := client.PingSeedGrpc(ctx, convertToProto(this.ThisNode))
		if err != nil {
			utils.Error(err)
			return nil, err
		}
		defer conn.Close()

		// Response?
		if response == nil {
			utils.Error("unable to ping seed node")
			return nil, errors.New("unable to ping seed node")
		}

		utils.Info(fmt.Sprintf("pinged seed node [delegates=%d]", len(response.Delegates)))

		for _, delegate := range response.Delegates {
			delegates = append(delegates, convertToDomain(delegate))
		}
		return delegates, nil
	}

	return nil, errors.New("unable to ping any seed delegates")
}

// UpdateGrpc
func (this *DisGoverService) UpdateGrpc(ctx context.Context, nodeList *proto.NodeList) (*proto.Empty, error) {

	// Cache delegates.
	for _, delegate := range nodeList.Delegates {
		convertToDomain(delegate).Cache(services.GetCache(), cache.NoExpiration)
	}

	utils.Info(fmt.Sprintf("delegates updated [count=%d]", len(nodeList.Delegates)))

	return &proto.Empty{}, nil
}

// peerUpdateGrpc
func (this *DisGoverService) peerUpdateGrpc() {

	// Get delegates in cache.
	delegates, err := types.ToNodesByTypeFromCache(services.GetCache(), types.TypeDelegate)
	if err != nil {
		utils.Fatal(err)
	}
	var protoDelegates = make([]*proto.Node, 0)
	for _, delegate := range delegates {
		protoDelegates = append(protoDelegates, convertToProto(delegate))
	}

	for _, delegate := range delegates {
		conn, err := grpc.Dial(fmt.Sprintf("%s:%d", delegate.Endpoint.Host, delegate.Endpoint.Port), grpc.WithInsecure())
		if err != nil {
			utils.Fatal(fmt.Sprintf("cannot dial node [host=%s, port=%d]", delegate.Endpoint.Host, delegate.Endpoint.Port), err)
			continue
		}
		client := proto.NewDisgoverGrpcClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		// Update.
		_, err = client.UpdateGrpc(ctx, &proto.NodeList{Delegates: protoDelegates})
		if err != nil {
			utils.Error(err)
		}
		conn.Close()
		cancel()
	}
}

/*
 *  Simple conversion functions from / to proto generated objects and domain level objects
 */
func convertToDomain(node *proto.Node) *types.Node {
	return &types.Node{
		Address: node.Address,
		Endpoint: &types.Endpoint{
			Host: node.Endpoint.Host,
			Port: node.Endpoint.Port,
		},
		Type: node.Type,
	}
}

// convertToProto
func convertToProto(node *types.Node) *proto.Node {
	if node == nil {
		return nil
	}
	return &proto.Node{
		Address: node.Address,
		Endpoint: &proto.Endpoint{
			Host: node.Endpoint.Host,
			Port: node.Endpoint.Port,
		},
		Type: node.Type,
	}
}
