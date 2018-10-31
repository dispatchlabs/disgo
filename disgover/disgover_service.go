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
	"time"
	"os"
	"io/ioutil"
	"strings"
)

var disGoverServiceInstance *DisGoverService
var disGoverServiceOnce sync.Once

// GetDisGoverService
func GetDisGoverService() *DisGoverService {
	disGoverServiceOnce.Do(func() {
		disGoverServiceInstance = &DisGoverService{
			ThisNode: &types.Node{
				Address:      types.GetAccount().Address,
				GrpcEndpoint: types.GetConfig().GrpcEndpoint,
				HttpEndpoint: types.GetConfig().HttpEndpoint,
				Type:         types.TypeNode,
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
	for _, seed := range types.GetConfig().Seeds {
		if seed.Address == types.GetAccount().Address {
			this.ThisNode.Type = types.TypeSeed
			break
		}
	}
	if types.GetConfig().Seeds == nil || len(types.GetConfig().Seeds) == 0 {
		this.ThisNode.Type = types.TypeSeed
	}

	// Cache delegates?
	if this.ThisNode.Type != types.TypeSeed {
		delegates, err := this.peerPingSeedGrpc()
		if err != nil {
			utils.Error(err)
			services.GetDbService().Close()
			seeds := types.GetConfig().Seeds
			utils.Fatal(fmt.Sprintf("unable to connect to seed node (%s:%d)...please try again later", seeds[0].GrpcEndpoint.Host, seeds[0].GrpcEndpoint.Port))
		}
		for _, delegate := range delegates {
			delegate.Cache(services.GetCache())
			if delegate.Address == this.ThisNode.Address {
				this.ThisNode.Type = delegate.Type
			}
		}
	}

	// Start update thread.
	if this.ThisNode.Type == types.TypeSeed {
		go this.updateWorker()
	}

	utils.Info(fmt.Sprintf("running as %s", this.ThisNode.Type))
	utils.Events().Raise(types.Events.DisGoverServiceInitFinished)
}

// updateWorker
func (this DisGoverService) updateWorker() {
	for {
		timer := time.NewTimer(30 * time.Second)
		select {
		case <-timer.C:

			// Any files to update?
			updateDirectory := "." + string(os.PathSeparator) + "update"
			files, err := ioutil.ReadDir(updateDirectory)
			if err != nil {
				continue
			}

			// Has lock file?
			if hasLockFile(files) {
				continue
			}

			for _, file := range files {
				// Read file?
				fileName := updateDirectory + string(os.PathSeparator) + file.Name()
				bytes, err := ioutil.ReadFile(fileName)
				if err != nil {
					utils.Error(fmt.Sprintf("unable to read file %s", file.Name()), err)
					continue
				}
				utils.Info(fmt.Sprintf("found software to update [file=%s]", fileName))

				// Update software.
				this.peerUpdateSoftwareGrpc(file.Name(), bytes)

				// Delete file.
				err = os.Remove(fileName)
				if err != nil {
					utils.Warn(fmt.Sprintf("unable to delete file %s", fileName), err)
				}
			}
		}
	}
}

// hasLockFile
func hasLockFile(files []os.FileInfo) bool {
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".LCK") {
			utils.Info(fmt.Sprintf("waiting for update file to upload [lockFile=%s]", file.Name()))
			return true
		}
	}
	return false;
}
