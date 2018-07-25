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
	cache "github.com/patrickmn/go-cache"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"github.com/pkg/errors"
)

// WithGrpc - Runs the DisGover service with GRPC transport
func (this *DisGoverService) WithGrpc() *DisGoverService {
	proto.RegisterDisgoverGrpcServer(services.GetGrpcService().Server, this)
	return this
}

// PingSeedGrpc
func (this *DisGoverService) PingSeedGrpc(ctx context.Context, node *proto.Node) (*proto.SeedResponse, error) {

	// Is this node a seed?
	if this.ThisNode.Type != types.TypeSeed {
		return nil, errors.New("you pinged a non-seed node")
	}

	// TODO: TEMP-Fix
	node.Type = types.TypeDelegate

	// Persist NEW Delegate
	txn := services.NewTxn(true)
	defer txn.Discard()

	newOrUpdatedDelegate := convertToDomain(node)
	newOrUpdatedDelegate.Persist(txn)                                   // Save to Badger
	newOrUpdatedDelegate.Cache(services.GetCache(), cache.NoExpiration) // Save to Cache

	// Drop OLD Delegate
	txn2 := services.NewTxn(true)
	defer txn2.Discard()

	newOrUpdatedDelegate2 := convertToDomain(node)
	newOrUpdatedDelegate2.Address = fmt.Sprintf("%s-%d", newOrUpdatedDelegate2.Endpoint.Host, int(newOrUpdatedDelegate2.Endpoint.Port))
	newOrUpdatedDelegate2.Unset(txn2, services.GetCache())

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

	return &proto.SeedResponse{Delegates: nodes}, nil
}

// peerPingSeedGrpc
func (this *DisGoverService) peerPingSeedGrpc() ([]*types.Node, error) {

	var nodes = make([]*types.Node, 0)
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

		for _, node := range response.Delegates {
			nodes = append(nodes, convertToDomain(node))
		}
		return nodes, nil
	}

	return nil, errors.New("unable to ping any seed nodes")
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
