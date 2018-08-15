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
	"github.com/dispatchlabs/disgo/commons/services"
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
		disGoverServiceInstance = &DisGoverService{
			ThisNode: &types.Node{
				Address:  types.GetAccount().Address,
				GrpcEndpoint: types.GetConfig().GrpcEndpoint,
				HttpEndpoint: types.GetConfig().HttpEndpoint,
				Type:     types.TypeDelegate,
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
	ThisNode *types.Node
	kdht     *kbucket.RoutingTable
	running  bool
}

// IsRunning - Returns the status if service is running
func (this *DisGoverService) IsRunning() bool {
	return this.running
}

// Go - Starts, Init and Runs the service
func (this *DisGoverService) Go() {
	this.running = true

	// Check if we are a seed.
	for _, seedEndpoint := range types.GetConfig().SeedEndpoints {
		if seedEndpoint.Host == types.GetConfig().GrpcEndpoint.Host && seedEndpoint.Port == types.GetConfig().GrpcEndpoint.Port {
			this.ThisNode.Type = types.TypeSeed
			break
		}
	}
	if types.GetConfig().SeedEndpoints == nil || len(types.GetConfig().SeedEndpoints) == 0 {
		this.ThisNode.Type = types.TypeSeed
	}

	// Cache delegates.
	if this.ThisNode.Type == types.TypeDelegate {
		delegates, err := this.peerPingSeedGrpc()
		if err != nil {
			utils.Fatal(err)
		}

		for _, delegate := range delegates {
			delegate.Cache(services.GetCache())
		}
	}

	utils.Info(fmt.Sprintf("running as %s", this.ThisNode.Type))
	utils.Events().Raise(Events.DisGoverServiceInitFinished)
}
