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
	"sync"

	"github.com/dgraph-io/badger"
	"github.com/dispatchlabs/disgo/commons/services"
	"github.com/dispatchlabs/disgo/commons/types"
	"github.com/dispatchlabs/disgo/commons/utils"
	"github.com/dispatchlabs/disgo/disgover"
	"github.com/dispatchlabs/disgo/commons/queue"
)

var daposServiceInstance *DAPoSService
var daposServiceOnce sync.Once

// GetDAPoSService
func GetDAPoSService() *DAPoSService {
	daposServiceOnce.Do(func() {
		daposServiceInstance = &DAPoSService{
			running: false,
			gossipChan: make(chan *types.Gossip, 1000),
			queueChan: make(chan *types.Gossip, 1000),
			timoutChan: make(chan bool, 1000),
			gossipQueue: queue.NewGossipQueue(),
		} // TODO: What should this be?
	})
	return daposServiceInstance
}

// DAPoSService -
type DAPoSService struct {
	running         bool
	gossipChan      chan *types.Gossip
	queueChan      	chan *types.Gossip
	timoutChan 		chan bool
	gossipQueue 	*queue.GossipQueue
}

// IsRunning -
func (this *DAPoSService) IsRunning() bool {
	return this.running
}

// Go -
func (this *DAPoSService) Go() {
	this.running = true
	utils.Info("running, waiting for delegates sync")

	utils.Events().On(
		types.Events.DisGoverServiceInitFinished,
		this.disGoverServiceInitFinished,
	)
}

// OnEvent - Event to
func (this *DAPoSService) disGoverServiceInitFinished() {

	if disgover.GetDisGoverService().ThisNode.Type == types.TypeDelegate {
		this.peerSynchronize()
		utils.Info("this prints out after peerSynchronize is called in disGoverServiceInitFinished()")
	}

	// Create genesis account.
	err := this.createGenesisAccount()
	if err != nil {
		services.GetDbService().Close()
		utils.Fatal("unable to create genesis account", err)
	}

	go this.gossipWorker()
	go this.transactionWorker()
	//No longer used.  GossipQueue is part of Gossip Worker
	//go this.queueWorker()

	utils.Events().Raise(types.Events.DAPoSServiceInitFinished)
}

// createGenesisTransactionAndAccount
func (this *DAPoSService) createGenesisAccount() error {
	txn := services.GetDb().NewTransaction(true)
	defer txn.Discard()

	genesisAccount, err := types.GetGenesisAccount()
	if err != nil {
		utils.Error(err)
	}
	_, err = types.ToAccountByAddress(txn, genesisAccount.Address)
	if err != nil {
		if err == badger.ErrKeyNotFound {
			err = genesisAccount.Set(txn, services.GetCache())
			if err != nil {
				return err
			}
		}
	}
	return txn.Commit(nil)
}
