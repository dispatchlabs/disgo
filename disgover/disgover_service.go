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
		tipe := types.TypeDelegate
		for _, endpoint := range types.GetConfig().SeedEndpoints {
			if fmt.Sprintf("%s:%d", endpoint.Host, endpoint.Port) == fmt.Sprintf("%s:%d", types.GetConfig().GrpcEndpoint.Host, types.GetConfig().GrpcEndpoint.Port) {
				tipe = types.TypeSeed
			}
		}

		// lCache, _ := lru.New(0), // FROM-AVERY

		disGoverServiceInstance = &DisGoverService{
			ThisNode: &types.Node{
				Address:  types.GetAccount().Address,
				Endpoint: types.GetConfig().GrpcEndpoint,
				Type:     tipe,
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
	kdht    *kbucket.RoutingTable
	running bool
}

// IsRunning - Returns the status if service is running
func (this *DisGoverService) IsRunning() bool {
	return this.running
}

// Go - Starts, Init and Runs the service
func (this *DisGoverService) Go(waitGroup *sync.WaitGroup) {
	this.running = true
	utils.Info("running")
	go func() {
		go this.pingSeedNodes()
	}()
}

// pingSeedNodes
func (this *DisGoverService) pingSeedNodes() {
	utils.Info("pinging other seed nodes...")

	// Ping Seed List
	for _, endpoint := range types.GetConfig().SeedEndpoints {
		var seedNode *types.Node
		if fmt.Sprintf("%s:%d", endpoint.Host, endpoint.Port) == fmt.Sprintf("%s:%d", types.GetConfig().GrpcEndpoint.Host, types.GetConfig().GrpcEndpoint.Port) {
			seedNode = this.ThisNode

			this.addPeer(*seedNode)
			continue
		} else {
			seedNode = &types.Node{
				Address:  "",
				Endpoint: endpoint,
				Type:     types.TypeSeed,
			}
		}
		var err error
		seedNode, err = this.peerPingGrpc(seedNode, this.ThisNode)
		if err != nil {
			utils.Error(err)
			continue
		}
		this.seedNodes = append(this.seedNodes, seedNode)

		this.addPeer(*seedNode)
		utils.Info(fmt.Sprintf("pinged seed [address=%s, ip=%s]", seedNode.Address, seedNode.Endpoint.Host))
	}

	utils.Events().Raise(Events.DisGoverServiceInitFinished)

}
