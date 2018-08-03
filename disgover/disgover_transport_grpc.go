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
	"io/ioutil"
	"os"
)

// WithGrpc - Runs the DisGover service with GRPC transport
func (this *DisGoverService) WithGrpc() *DisGoverService {
	proto.RegisterDisgoverGrpcServer(services.GetGrpcService().Server, this)
	return this
}

// PingSeedGrpc
func (this *DisGoverService) PingSeedGrpc(ctx context.Context, pingSeed *proto.PingSeed) (*proto.Update, error) {

	// Is this node a seed?
	if this.ThisNode.Type != types.TypeSeed {
		return nil, errors.New("you pinged a non-seed node")
	}

	node := convertToDomainNode(pingSeed.Node)
	authenticate := convertToDomainAuthenticate(pingSeed.Authenticate)
	authenticateAddress, err := authenticate.GetAddress()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("unable to authenticate you [error=%s]", err.Error()))
	}
	if authenticateAddress != node.Address {
		return nil, errors.New("unable to authenticate you")
	}

	// Persist and cache node.
	txn := services.NewTxn(true)
	defer txn.Discard()

	for _, delegateAddress := range types.GetConfig().DelegateAddresses {

		// Is this a delegate node?
		if delegateAddress == node.Address {

			// Is this an authentic delegate?
			err := authenticate.Verify(node.Address)
			if err != nil {
				utils.Warn(fmt.Sprintf("unable to authenticate delegate [address=%s]", node.Address))
				return nil, errors.New("unable to authenticate you as a delegate")
			}
			node.Type = types.TypeDelegate
			break
		}
	}
	node.PersistAndCache(txn, services.GetCache(), cache.NoExpiration)

	// Get cached delegates.
	delegates, err := types.ToNodesByTypeFromCache(services.GetCache(), types.TypeDelegate)
	if err != nil {
		return nil, err
	}

	var nodes = make([]*proto.Node, 0)
	for _, delegate := range delegates {
		nodes = append(nodes, convertToProtoNode(delegate))
	}

	utils.Info(fmt.Sprintf("received ping [address=%s, host=%s, port=%d, delegates=%d]", node.Address, node.GrpcEndpoint.Host, node.GrpcEndpoint.Port, len(delegates)))

	// New authenticate.
	authenticate, err = types.NewAuthenticate()
	if err != nil {
		utils.Error(err)
		return nil, err
	}

	// Update all peers.
	go func() {
		time.Sleep(500 * time.Millisecond)
		this.peerUpdateGrpc()
	}()

	return &proto.Update{Authenticate: convertToProtoAuthenticate(authenticate), Delegates: nodes}, nil
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
		client := proto.NewDisgoverGrpcClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		// New authenticate.
		authenticate, err := types.NewAuthenticate()
		if err != nil {
			return nil, err
		}

		// Ping seed.
		response, err := client.PingSeedGrpc(ctx, &proto.PingSeed{Authenticate: convertToProtoAuthenticate(authenticate), Node: convertToProtoNode(this.ThisNode)})
		if err != nil {
			conn.Close()
			cancel()
			utils.Error(err)
			return nil, err
		}
		conn.Close()
		cancel()

		// Response?
		if response == nil {
			utils.Error("unable to ping seed node")
			return nil, errors.New("unable to ping seed node")
		}

		// Verify seed node is authentic?
		err = this.verifySeedNode(response.Authenticate)
		if err != nil {
			return nil, err
		}

		utils.Info(fmt.Sprintf("pinged seed node [delegates=%d]", len(response.Delegates)))

		for _, delegate := range response.Delegates {
			delegates = append(delegates, convertToDomainNode(delegate))
		}
		return delegates, nil
	}

	return nil, errors.New("unable to ping any seed delegates")
}

// UpdateGrpc
func (this *DisGoverService) UpdateGrpc(ctx context.Context, nodeInfo *proto.Update) (*proto.Empty, error) {

	// Verify seed node is authentic?
	err := this.verifySeedNode(nodeInfo.Authenticate)
	if err != nil {
		return &proto.Empty{}, err
	}

	// Cache delegates.
	for _, delegate := range nodeInfo.Delegates {
		convertToDomainNode(delegate).Cache(services.GetCache(), cache.NoExpiration)
	}

	utils.Info(fmt.Sprintf("delegates updated [count=%d]", len(nodeInfo.Delegates)))

	return &proto.Empty{}, nil
}

// peerUpdateGrpc
func (this *DisGoverService) peerUpdateGrpc() {

	// Get delegates in cache.
	delegates, err := types.ToNodesByTypeFromCache(services.GetCache(), types.TypeDelegate)
	if err != nil {
		utils.Error(err)
		return
	}
	var protoDelegates = make([]*proto.Node, 0)
	for _, delegate := range delegates {
		protoDelegates = append(protoDelegates, convertToProtoNode(delegate))
	}

	// New authenticate.
	authenticate, err := types.NewAuthenticate()
	if err != nil {
		utils.Error(err)
		return
	}

	for _, delegate := range delegates {
		conn, err := grpc.Dial(fmt.Sprintf("%s:%d", delegate.GrpcEndpoint.Host, delegate.GrpcEndpoint.Port), grpc.WithInsecure())
		if err != nil {
			utils.Fatal(fmt.Sprintf("cannot dial node [host=%s, port=%d]", delegate.GrpcEndpoint.Host, delegate.GrpcEndpoint.Port), err)
			continue
		}
		client := proto.NewDisgoverGrpcClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		// Update.
		_, err = client.UpdateGrpc(ctx, &proto.Update{Authenticate: convertToProtoAuthenticate(authenticate), Delegates: protoDelegates})
		if err != nil {
			utils.Error(err)
		}
		conn.Close()
		cancel()
	}
}

// UpdateSoftwareGrpc
func (this *DisGoverService) UpdateSoftwareGrpc(ctx context.Context, softwareUpdate *proto.SoftwareUpdate) (*proto.Empty, error) {

	// Verify seed node is authentic?
	err := this.verifySeedNode(softwareUpdate.Authenticate)
	if err != nil {
		return &proto.Empty{}, err
	}

	// Create directory?
	directoryName := "." + string(os.PathSeparator) + "update"
	if !utils.Exists(directoryName) {
		os.MkdirAll(directoryName, os.ModePerm)
	}

	// Write file.
	fileName := directoryName + string(os.PathSeparator) + "disgo"
	err = ioutil.WriteFile(fileName, softwareUpdate.Software, 0)
	if err != nil {
		utils.Error(fmt.Sprintf("unable to save file %s", fileName), err)
		return &proto.Empty{}, err
	}

	utils.Info(fmt.Sprintf("software updated from seed node"))

	return &proto.Empty{}, nil
}

// peerUpdateSoftwareGrpc
func (this *DisGoverService) peerUpdateSoftwareGrpc(software []byte) {

	// Get delegates in cache.
	delegates, err := types.ToNodesByTypeFromCache(services.GetCache(), types.TypeDelegate)
	if err != nil {
		utils.Error(err)
		return
	}

	// New authenticate.
	authenticate, err := types.NewAuthenticate()
	if err != nil {
		utils.Error(err)
		return
	}

	for _, delegate := range delegates {
		maxSize := 1024 * 1024 * 1024 // TODO: Do we need this (ServerOptions sets this)?
		conn, err := grpc.Dial(fmt.Sprintf("%s:%d", delegate.GrpcEndpoint.Host, delegate.GrpcEndpoint.Port), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxSize), grpc.MaxCallSendMsgSize(maxSize)), grpc.WithInsecure())
		if err != nil {
			utils.Warn(fmt.Sprintf("cannot dial node [host=%s, port=%d]", delegate.GrpcEndpoint.Host, delegate.GrpcEndpoint.Port), err)
			continue
		}
		client := proto.NewDisgoverGrpcClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

		// Update software.
		_, err = client.UpdateSoftwareGrpc(ctx, &proto.SoftwareUpdate{Authenticate: convertToProtoAuthenticate(authenticate), Software: software})
		if err != nil {
			utils.Warn(fmt.Sprintf("unable to update sofware [address=%s, host=%s, port=%d]", delegate.Address, delegate.GrpcEndpoint.Host, delegate.GrpcEndpoint.Port))
			continue
		}
		conn.Close()
		cancel()
	}
}

// verifySeedNode
func (this *DisGoverService) verifySeedNode(protoAuthenticate *proto.Authenticate) error {

	if protoAuthenticate == nil {
		utils.Warn("attempted update by a non-authorized seed node")
		return errors.New("you are not an authorized seed node")
	}

	authenticate := convertToDomainAuthenticate(protoAuthenticate)
	authenticateAddress, err := authenticate.GetAddress()
	if err != nil {
		utils.Error(err)
		return err
	}

	for _, seedAddress := range types.GetConfig().SeedAddresses {
		if seedAddress == authenticateAddress {
			err = authenticate.Verify(seedAddress)
			if err != nil {
				return errors.New("you are not an authorized seed node")
			}
		}
	}

	utils.Warn("attempted update by a non-authorized seed node")
	return errors.New("you are not an authorized seed node")
}

/*
 *  Simple conversion functions from / to proto generated objects and domain level objects
 */
func convertToDomainNode(node *proto.Node) *types.Node {
	return &types.Node{
		Address: node.Address,
		GrpcEndpoint: &types.Endpoint{
			Host: node.GrpcEndpoint.Host,
			Port: node.GrpcEndpoint.Port, //TODO: grpc endpoint?
		},
		HttpEndpoint: &types.Endpoint{
			Host: node.HttpEndpoint.Host,
			Port: node.HttpEndpoint.Port,
		},
		Type: node.Type,
	}
}

// convertToProtoNode
func convertToProtoNode(node *types.Node) *proto.Node {
	if node == nil {
		return nil
	}
	return &proto.Node{
		Address: node.Address,
		GrpcEndpoint: &proto.Endpoint{
			Host: node.GrpcEndpoint.Host,
			Port: node.GrpcEndpoint.Port,
		},
		HttpEndpoint: &proto.Endpoint{
			Host: node.HttpEndpoint.Host,
			Port: node.HttpEndpoint.Port,
		},
		Type: node.Type,
	}
}

// convertToDomainAuthenticate
func convertToDomainAuthenticate(authenticate *proto.Authenticate) *types.Authenticate {
	return &types.Authenticate{
		Hash:      authenticate.Hash,
		Time:      authenticate.Time,
		Signature: authenticate.Signature,
	}
}

// convertToProtoAuthenticate
func convertToProtoAuthenticate(authenticate *types.Authenticate) *proto.Authenticate {
	if authenticate == nil {
		return nil
	}
	return &proto.Authenticate{
		Hash:      authenticate.Hash,
		Time:      authenticate.Time,
		Signature: authenticate.Signature,
	}
}
