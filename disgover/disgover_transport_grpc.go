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
	"io/ioutil"
	"os"
	"github.com/jasonlvhit/gocron"
)

//TODO: we are going to drop delegates if we fail to communicate with them.  The exepctation will be that the seed will tell us when they come back on line
//TODO: Need to add more robust solution to resolve missing delegates and add a notification infrastructure for the rest of the delegates.

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
	authentication := convertToDomainAuthentication(pingSeed.Authentication)
	authenticationAddress, err := authentication.GetAddress()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("unable to authentication you [error=%s]", err.Error()))
	}
	if authenticationAddress != node.Address {
		return nil, errors.New("unable to authentication you")
	}

	// Persist and cache node.
	txn := services.NewTxn(true)
	defer txn.Discard()

	// If delegate addresses is not set all nodes other than seed become a delegate (making it easy for testing and production).
	if len(types.GetConfig().DelegateAddresses) == 0 {
		node.Type = types.TypeDelegate
	} else {
		for _, delegateAddress := range types.GetConfig().DelegateAddresses {

			// Is this a delegate node?
			if delegateAddress == node.Address {

				// Is this an authentic delegate?
				err := authentication.Verify(services.GetCache(), node.Address)
				if err != nil {
					utils.Warn(fmt.Sprintf("unable to authentication delegate [address=%s]", node.Address))
					return nil, errors.New("unable to authentication you as a delegate")
				}
				node.Type = types.TypeDelegate
				break
			}
		}
	}
	node.PersistAndCache(txn, services.GetCache())

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

	// New authentication.
	authentication, err = types.NewAuthentication()
	if err != nil {
		utils.Error(err)
		return nil, err
	}

	// Update all peers.
	go func() {
		time.Sleep(500 * time.Millisecond)
		this.peerUpdateGrpc()
	}()

	return &proto.Update{Authentication: convertToProtoAuthentication(authentication), Delegates: nodes}, nil
}

// peerPingSeedGrpc
func (this *DisGoverService) peerPingSeedGrpc() ([]*types.Node, error) {

	var delegates = make([]*types.Node, 0)
	for _, seedEndpoint := range types.GetConfig().Seeds {
		conn, err := grpc.Dial(fmt.Sprintf("%s:%d", seedEndpoint.GrpcEndpoint.Host, seedEndpoint.GrpcEndpoint.Port), grpc.WithInsecure())
		if err != nil {
			utils.Fatal(fmt.Sprintf("cannot dial seed [host=%s, port=%d]", seedEndpoint.GrpcEndpoint.Host, seedEndpoint.GrpcEndpoint.Port), err)
			return nil, err
		}
		client := proto.NewDisgoverGrpcClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		// New authentication.
		authentication, err := types.NewAuthentication()
		if err != nil {
			return nil, err
		}

		protoNode := convertToProtoNode(this.ThisNode)
		// Ping seed.
		response, err := client.PingSeedGrpc(ctx, &proto.PingSeed{Authentication: convertToProtoAuthentication(authentication), Node: protoNode})
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
		err = this.verifySeedNode(response.Authentication)
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
func (this *DisGoverService) UpdateGrpc(ctx context.Context, update *proto.Update) (*proto.Empty, error) {

	// Verify seed node is authentic?
	err := this.verifySeedNode(update.Authentication)
	if err != nil {
		return &proto.Empty{}, err
	}

	// Cache delegates.
	for _, delegate := range update.Delegates {
		convertToDomainNode(delegate).Cache(services.GetCache())
		utils.Info(fmt.Sprintf("delegates updated [count=%d] %s : %s:%d", len(update.Delegates), delegate.Address, delegate.GrpcEndpoint.Host ,delegate.GrpcEndpoint.Port))
	}
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

	// New authentication.
	authentication, err := types.NewAuthentication()
	if err != nil {
		utils.Error(err)
		return
	}

	txn := services.NewTxn(true)
	defer txn.Discard()

	for _, delegate := range delegates {
		conn, err := grpc.Dial(fmt.Sprintf("%s:%d", delegate.GrpcEndpoint.Host, delegate.GrpcEndpoint.Port), grpc.WithInsecure())
		if err != nil {
			utils.Error(fmt.Sprintf("cannot dial node [host=%s, port=%d]", delegate.GrpcEndpoint.Host, delegate.GrpcEndpoint.Port), err)
			err = delegate.Unset(txn, services.GetCache())
			if err != nil {
				utils.Error(err)
			}
			continue
		}
		client := proto.NewDisgoverGrpcClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		// Update.
		_, err = client.UpdateGrpc(ctx, &proto.Update{Authentication: convertToProtoAuthentication(authentication), Delegates: protoDelegates})
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
	err := this.verifySeedNode(softwareUpdate.Authentication)
	if err != nil {
		return &proto.Empty{}, err
	}

	// Write file.
	fileName := "." + string(os.PathSeparator) + "disgo"
	err = ioutil.WriteFile(fileName, softwareUpdate.Software, 0)
	if err != nil {
		utils.Error(fmt.Sprintf("unable to save file %s", fileName), err)
		return &proto.Empty{}, err
	}

	utils.Info(fmt.Sprintf("software updated from seed node"))

	// Schedule the reboot.
	go func() {
		gocron.Every(1).Day().At(softwareUpdate.ScheduledReboot).Do(func() {
			gocron.Clear()
			services.GetDbService().Close()
			utils.Info("rebooting with new version of disgo...")
			os.Exit(0)
		})
		<-gocron.Start()
	}()

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

	// New authentication.
	authentication, err := types.NewAuthentication()
	if err != nil {
		utils.Error(err)
		return
	}

	txn := services.NewTxn(true)
	defer txn.Discard()
	reboot := time.Now()
	reboot = reboot.Add(2 * time.Minute)

	for _, delegate := range delegates {
		maxSize := 1024 * 1024 * 1024 // TODO: Do we need this (ServerOptions sets this)?
		conn, err := grpc.Dial(fmt.Sprintf("%s:%d", delegate.GrpcEndpoint.Host, delegate.GrpcEndpoint.Port), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxSize), grpc.MaxCallSendMsgSize(maxSize)), grpc.WithInsecure())
		if err != nil {
			utils.Error(fmt.Sprintf("cannot dial node [host=%s, port=%d]", delegate.GrpcEndpoint.Host, delegate.GrpcEndpoint.Port), err)
			err = delegate.Unset(txn, services.GetCache())
			if err != nil {
				utils.Error(err)
			}
			continue
		}
		client := proto.NewDisgoverGrpcClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

		// Update software.
		_, err = client.UpdateSoftwareGrpc(ctx, &proto.SoftwareUpdate{Authentication: convertToProtoAuthentication(authentication), Software: software, ScheduledReboot: reboot.Format("3:04")})
		if err != nil {
			utils.Warn(fmt.Sprintf("unable to update sofware [address=%s, host=%s, port=%d]", delegate.Address, delegate.GrpcEndpoint.Host, delegate.GrpcEndpoint.Port))
			continue
		}
		conn.Close()
		cancel()
	}
}

// verifySeedNode
func (this *DisGoverService) verifySeedNode(protoAuthenticate *proto.Authentication) error {

	if protoAuthenticate == nil {
		utils.Warn("attempted update by a non-authorized seed node")
		return errors.New("you are not an authorized seed node")
	}

	authentication := convertToDomainAuthentication(protoAuthenticate)
	authenticationAddress, err := authentication.GetAddress()
	if err != nil {
		utils.Error(err)
		return err
	}

	for _, SeedNode := range types.GetConfig().Seeds {
		if SeedNode.Address == authenticationAddress {
			err = authentication.Verify(services.GetCache(), SeedNode.Address)
			if err != nil {
				return errors.New("you are not an authorized seed node")
			}
			return nil
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

// convertToDomainAuthentication
func convertToDomainAuthentication(authentication *proto.Authentication) *types.Authentication {
	return &types.Authentication{
		Hash:      authentication.Hash,
		Time:      authentication.Time,
		Signature: authentication.Signature,
	}
}

// convertToProtoAuthentication
func convertToProtoAuthentication(authentication *types.Authentication) *proto.Authentication {
	if authentication == nil {
		return nil
	}
	return &proto.Authentication{
		Hash:      authentication.Hash,
		Time:      authentication.Time,
		Signature: authentication.Signature,
	}
}
