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

	peer "github.com/libp2p/go-libp2p-peer"
	log "github.com/sirupsen/logrus"

	"time"

	"github.com/dispatchlabs/commons/services"
	"github.com/dispatchlabs/commons/types"
	"github.com/dispatchlabs/commons/utils"
	proto "github.com/dispatchlabs/disgover/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// WithGrpc - Runs the DisGover service with GRPC transport
func (this *DisGoverService) WithGrpc() *DisGoverService {
	proto.RegisterDisgoverGrpcServer(services.GetGrpcService().Server, this)
	return this
}

// PeerPingGrpc - Called when a PING call was made by a client
func (this *DisGoverService) PingGrpc(ctx context.Context, node *proto.Node) (*proto.Node, error) {
	// FROM-AVERY
	// txn := services.NewTxn(true)
	// defer txn.Discard()
	// domainNode := convertToDomain(node)
	// domainNode.Set(txn)
	// this.kdht.Update(peer.ID(node.Address))
	// utils.Info(fmt.Sprintf("received ping [address=%s, ip=%s, port=%d]", node.Address, node.Endpoint.Host, node.Endpoint.Port))
	// return convertToProto(this.ThisNode), nil

	domainNode := convertToDomain(node)
	services.GetCache().Set(domainNode.Address, domainNode, types.NodeTTL)
	this.kdht.Update(peer.ID(node.Address))
	utils.Info(fmt.Sprintf("received ping [address=%s, ip=%s, port=%d]", node.Address, node.Endpoint.Host, node.Endpoint.Port))
	return convertToProto(this.ThisNode), nil
}

// peerPingGrpc
func (this *DisGoverService) peerPingGrpc(contactToPing *types.Node, sender *types.Node) (*types.Node, error) {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", contactToPing.Endpoint.Host, contactToPing.Endpoint.Port), grpc.WithInsecure())
	if err != nil {
		utils.Fatal("cannot dial peer", err)
		return nil, err
	}
	defer conn.Close()
	client := proto.NewDisgoverGrpcClient(conn)
	contactProto := convertToProto(sender)
	ctx, cancel := context.WithTimeout(context.Background(), 2000*time.Millisecond)
	defer cancel()
	response, err := client.PingGrpc(ctx, contactProto)
	if err != nil {
		utils.Error(err)
		return nil, err
	}
	defer conn.Close()

	if response == nil {
		log.Error(fmt.Sprintf("uanble to pinge peer node [address=%s, ip=%s]", response.Address, response.Endpoint.Host))
		return nil, nil
	}
	log.Info(fmt.Sprintf("pinged peer node [address=%s, ip=%s]", response.Address, response.Endpoint.Host))
	return convertToDomain(response), nil
}

// FindGrpc - Called when peerFindGrpc() is called by the client.
func (this *DisGoverService) FindGrpc(ctx context.Context, request *proto.Request) (*proto.Node, error) {

	// Find node by address?
	node, err := this.Find(request.Payload)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, types.ErrNotFound
	}
	return convertToProto(node), nil
}

// peerFindGrpc
func (this *DisGoverService) peerFindGrpc(peerToAsk *types.Node, address string) *types.Node {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", peerToAsk.Endpoint.Host, peerToAsk.Endpoint.Port), grpc.WithInsecure())
	if err != nil {
		utils.Error(err)
	}
	defer conn.Close()
	contextWithTimeout, cancel := context.WithTimeout(context.Background(), 2000*time.Millisecond)
	defer cancel()

	client := proto.NewDisgoverGrpcClient(conn)
	if err != nil {
		utils.Error(err)
		return nil
	}
	response, err := client.FindGrpc(contextWithTimeout, &proto.Request{Payload: address})
	if err != nil {
		if err != types.ErrNotFound {
			utils.Error(err)
		}
		return nil
	}
	if response == nil {
		return nil
	}
	return convertToDomain(response)
}

// FindByTypeGrpc
func (this *DisGoverService) FindByTypeGrpc(context context.Context, request *proto.Request) (*proto.NodeList, error) {

	// Find nodes.
	nodes, err := this.FindByType(request.Payload)
	if err != nil {
		return nil, err
	}

	// Convert to domain.
	protoNodes := make([]*proto.Node, 0)
	for _, node := range nodes {
		protoNodes = append(protoNodes, convertToProto(node))
	}

	return &proto.NodeList{Nodes: protoNodes}, nil
}

// peerFindByTypeGrpc
func (this *DisGoverService) peerFindByTypeGrpc(peerToAsk *types.Node, tipe string) ([]*types.Node, error) {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", peerToAsk.Endpoint.Host, peerToAsk.Endpoint.Port), grpc.WithInsecure())
	if err != nil {
		utils.Error(err)
	}
	defer conn.Close()
	contextWithTimeout, cancel := context.WithTimeout(context.Background(), 2000*time.Millisecond)
	defer cancel()

	client := proto.NewDisgoverGrpcClient(conn)
	if err != nil {
		utils.Error(err)
		return nil, err
	}
	response, err := client.FindByTypeGrpc(contextWithTimeout, &proto.Request{Payload: tipe})
	if err != nil {
		utils.Error(err)
		return nil, err
	}

	var nodes = make([]*types.Node, 0)
	for _, node := range response.Nodes {
		nodes = append(nodes, convertToDomain(node))
	}
	return nodes, nil
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

// func getIpFromContext(ctx context.Context) string {
// 	// grpcPeer "google.golang.org/grpc/peer"

// 	thePeer, ok := grpcPeer.FromContext(ctx)
// 	if !ok {
// 		return nil, fmt.Errorf("Disgover-TRACE: failed to get peer from ctx")
// 	}
// 	if thePeer.Addr == net.Addr(nil) {
// 		return nil, fmt.Errorf("Disgover-TRACE: failed to get peer address")
// 	}

// 	var peerAddressWithPort = thePeer.Addr.String()

// 	return (peerAddressWithPort[0:strings.Index(peerAddressWithPort, ":")], "")
// }
