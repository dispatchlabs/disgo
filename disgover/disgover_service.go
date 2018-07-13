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
	"sync"

	//"github.com/dispatchlabs/disgo/commons/services"
	"github.com/dispatchlabs/disgo/commons/types"
	"github.com/dispatchlabs/disgo/commons/utils"
	"github.com/libp2p/go-libp2p-kbucket"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-peerstore"
	"strconv"
)

var disGoverServiceInstance *DisGoverService
var disGoverServiceOnce sync.Once

type disgoverEvents struct {
	DisGoverServiceInitFinished string
}

var (
	// Events - `disgover` events
	Events = disgoverEvents{
		DisGoverServiceInitFinished: "DisGoverServiceInitFinished",
	}
)

// GetDisGoverService
func GetDisGoverService() *DisGoverService {
	disGoverServiceOnce.Do(func() {
		thisNodeType := types.TypeNode

		// Check if we are a SEED
		for _, endpoint := range types.GetConfig().SeedEndpoints {
			var portAndIP1 = fmt.Sprintf("%s:%d", endpoint.Host, endpoint.Port)
			var portAndIP2 = fmt.Sprintf("%s:%d", types.GetConfig().GrpcEndpoint.Host, types.GetConfig().GrpcEndpoint.Port)

			if portAndIP1 == portAndIP2 {
				thisNodeType = types.TypeSeed
				break
			}
		}

		// If no seeds are specified then we are THE seed
		if thisNodeType == types.TypeNode && len(types.GetConfig().SeedEndpoints) == 0 {
			thisNodeType = types.TypeSeed
		}

		utils.Info(fmt.Sprintf("running as %s", thisNodeType))

		disGoverServiceInstance = &DisGoverService{
			ThisNode: &types.Node{
				Address:  types.GetAccount().Address,
				Endpoint: types.GetConfig().GrpcEndpoint,
				Type:     thisNodeType,
			},
			// lruCache: lCache,
			kdht: kbucket.NewRoutingTable(
				1000,
				kbucket.ConvertPeerID(peer.ID(types.GetAccount().Address)),
				1000,
				peerstore.NewMetrics(),
			),
			running: false,
		}
	})
	return disGoverServiceInstance
}

// DisGoverService
type DisGoverService struct {
	ThisNode  *types.Node
	seedNodes []*types.Node
	kdht      *kbucket.RoutingTable
	running   bool
}

// IsRunning - Returns the status if service is running
func (this *DisGoverService) IsRunning() bool {
	return this.running
}

// Go - Starts, Init and Runs the service
func (this *DisGoverService) Go(waitGroup *sync.WaitGroup) {
	this.running = true
	utils.Info("running")
	if this.ThisNode.Type == types.TypeSeed{
		this.saveDelegatesFromConfigToCache()
	}
	go this.pingSeedNodes()
}

// pingSeedNodes
func (this *DisGoverService) pingSeedNodes() {
	utils.Info("PING seed nodes...")

	// Ping Seed List
	for _, endpoint := range types.GetConfig().SeedEndpoints {
		var seedNode *types.Node
		var portAndIP1 = fmt.Sprintf("%s:%d", endpoint.Host, endpoint.Port)
		var portAndIP2 = fmt.Sprintf("%s:%d", types.GetConfig().GrpcEndpoint.Host, types.GetConfig().GrpcEndpoint.Port)

		// IF - WE are the seed then do nothing
		if portAndIP1 == portAndIP2 {
			continue
		}

		// ELSE - init a new object and query it
		seedNode = &types.Node{
			Address:  "",
			Endpoint: endpoint,
			Type:     types.TypeSeed,
		}

		var err error
		seedNode, err = this.peerPingGrpc(seedNode, this.ThisNode)
		if err != nil {
			utils.Error(err)
			continue
		}
		//add the seed nodes
		this.seedNodes = append(this.seedNodes, seedNode)
		this.addOrUpdatePeer(seedNode)

		//ask them for delegates
		delis, err := this.FindByType(types.TypeDelegate)
		if err != nil{
			utils.Error(err)
			continue
		}
		//check if we are one
		for _, deli := range delis {
			deliPortAndIP1 := fmt.Sprintf("%s:%d", deli.Endpoint.Host, deli.Endpoint.Port)
			if deliPortAndIP1 == portAndIP2 {
				this.ThisNode.Type = types.TypeDelegate
			}
		}
		//if we are
		if this.ThisNode.Type == types.TypeDelegate{
			seedNode, err = this.peerPingGrpc(seedNode, this.ThisNode) // tell the seed
		}
		utils.Info(fmt.Sprintf("pinged seed [address=%s, ip:port=%s:%d]", seedNode.Address, seedNode.Endpoint.Host, seedNode.Endpoint.Port))
	}

	utils.Events().Raise(Events.DisGoverServiceInitFinished)
	return
}

func (this *DisGoverService) saveDelegatesFromConfigToCache() {
	i := 0 //TODO: bad practice to make addresses numbers before we have the actual address
	for _, endpoint := range types.GetConfig().DelegateEndpoints {
		var node = &types.Node{
			Address:  strconv.Itoa(i),
			Endpoint: endpoint,
			Type:     types.TypeDelegate,
		}
		i++
		this.addOrUpdatePeer(node)
	}
	this.addOrUpdatePeer(this.ThisNode)
}
