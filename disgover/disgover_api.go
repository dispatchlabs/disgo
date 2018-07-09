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

// Package disgover is the Dispatch KDHT based node discovery engine.
//
// It is a distributed, node discovery mechanism that enables locating any
// entity (server, worker, drone, actor) based on node id.
//
// The intent is to not be a data storage/distribution mechanism.
// Meaning we implement only `PING` and `FIND` rpc.
//
// One `DisGover` instance in the node:
// - stores info about numerous nodes
// - functions as a gateway to outside local network
package disgover

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/dispatchlabs/disgo/commons/services"
	"github.com/dispatchlabs/disgo/commons/types"
	"github.com/dispatchlabs/disgo/commons/utils"
	"github.com/libp2p/go-libp2p-peer"
	cache "github.com/patrickmn/go-cache"
)

// Find - Finds a node on the network, check internally then asks the peers if not found
func (this *DisGoverService) Find(address string) (*types.Node, error) {

	//first check cache
	value, ok := services.GetCache().Get(address)
	if ok {
		node := value.(*types.Node)
		return node, nil
	}

	//now check badger
	txn := services.NewTxn(true)
	defer txn.Discard()
	node, _ := types.ToNodeByAddress(txn, address)
	if node != nil {
		return node, nil
	}

	// Find node from peer seeds.
	peer := this.kdht.Find(peer.ID(address))
	if peer != "" {
		id := peer.Pretty()
		//id := kbucket.ConvertPeerID(peer)
		node, err := types.ToNodeByAddress(txn, id)
		return node, err
	}

	//peerID := kbucket.ConvertPeerID(peer.ID(this.ThisNode.Address))
	//nearestpeer := this.kdht.NearestPeer(peerID)
	//// Find node from peer seeds.
	//for _, seedContact := range this.seedNodes {
	//	if seedContact.Address == this.ThisNode.Address {
	//		continue
	//	}
	//	node := this.peerFindGrpc(seedContact, address)
	//	if node == nil {
	//		continue
	//	}
	//}
	err := errors.New("could not find")
	return nil, err
}

// FindByType
func (this *DisGoverService) FindByType(tipe string) ([]*types.Node, error) {
	utils.Info(fmt.Sprintf("finding %s nodes...", strings.ToLower(tipe)))

	// TODO: We should put this in node.go and use table- and key- style keys.

	var nodes []*types.Node

	// 1st - Ask SEEDs for more stuff and merge/replace the result - done bye `addOrUpdatePeer()`
	var haveToQueryForUpdates = false
	var cachedItems = services.GetCache().Items()
	for _, value := range cachedItems {
		if reflect.TypeOf(value.Object) != reflect.TypeOf(&types.Node{}) {
			continue
		}
		node := value.Object.(*types.Node)
		if node.Type == types.TypeDelegate {
			var fixedTempAddress = fmt.Sprintf("%s:%d", node.Endpoint.Host, node.Endpoint.Port)

			if node.Address == fixedTempAddress {
				haveToQueryForUpdates = true
				break
			}
		}
	}

	if haveToQueryForUpdates {
		for _, seedNode := range this.seedNodes {
			if seedNode.Address == this.ThisNode.Address {
				continue
			}

			peerNodes, err := this.peerFindByTypeGrpc(seedNode, tipe)
			if err != nil {
				utils.Error(err)
				continue
			}
			for _, node := range peerNodes {
				this.addOrUpdatePeer(node)
			}
		}
	}

	// 2nd - See what is in the cache - this includes recent data from SEEDs
	cachedItems = services.GetCache().Items()
	for _, value := range cachedItems {
		if reflect.TypeOf(value.Object) != reflect.TypeOf(&types.Node{}) {
			continue
		}
		node := value.Object.(*types.Node)
		if node.Type == types.TypeDelegate {
			if !containsNodeByEndpoint(nodes, node.Endpoint) {
				nodes = append(nodes, node)
			}
		}
	}

	utils.Info(fmt.Sprintf("found %d %s nodes...", len(nodes), strings.ToLower(tipe)))

	return nodes, nil
}

// containsNodeByAddress
func containsNodeByAddress(nodes []*types.Node, address string) bool {
	for _, n := range nodes {
		if n.Address == address {
			return true
		}
	}
	return false
}

func containsNodeByEndpoint(nodes []*types.Node, endpoint *types.Endpoint) bool {
	for _, n := range nodes {
		if n.Endpoint.Host == endpoint.Host && n.Endpoint.Port == endpoint.Port {
			return true
		}
	}
	return false
}

/* TODO: Commented out for the time until we need it.
func (this *DisGoverService) findViaPeers(idToFind string, sender *types.Node) (*types.Node, error) {
	log.WithFields(log.Fields{
		"method": utils.GetCallingFuncName() + fmt.Sprintf(" -> %s", idToFind),
	}).Info("find using peers")

	peerIDs := this.kdht.NearestPeers([]byte(this.ThisNode.Address), len(this.Nodes))

	for _, peerID := range peerIDs {
		peerIDAsString := string(peerID)
		if (peerIDAsString == this.ThisNode.Address) || (peerIDAsString == sender.Address) {
			continue
		}

		peerToAsk := this.Nodes[peerIDAsString]
		foundContact := this.peerFindGrpc(peerToAsk, idToFind)

		if foundContact != nil {
			fmt.Println(fmt.Sprintf(" %s, on [%s : %d]", foundContact.Address, foundContact.Endpoint.Host, foundContact.Endpoint.Port))
			go this.addOrUpdate(foundContact)
			return foundContact, nil
		}
	}

	fmt.Println("       NOT FOUND")
	return nil, nil
}
*/

func (this *DisGoverService) addHelper(node *types.Node) (bool bool, err error) {

	if len(strings.TrimSpace(node.Address)) <= 0 {
		utils.Error(fmt.Sprintf("address is empty for: [%s : %d]", node.Endpoint.Host, node.Endpoint.Port))
		return false, nil
	}

	services.GetCache().Set(node.Address, node, cache.NoExpiration)
	txn := services.NewTxn(true)
	defer txn.Discard()
	err = node.Set(txn)
	if err != nil {
		return false, err
	}
	this.kdht.Update(peer.ID(node.Address))

	return true, nil
}

func (this *DisGoverService) deletePeer(node *types.Node) (bool bool, err error) {
	services.GetCache().Delete(node.Address)
	txn := services.NewTxn(true)
	defer txn.Discard()
	err = node.Delete(txn)
	if err != nil {
		return false, err
	}
	this.kdht.Remove(peer.ID(node.Address))
	return true, nil
}

func (this *DisGoverService) addOrUpdatePeer(node *types.Node) (bool bool, err error) {
	var fixedTempAddress = fmt.Sprintf("%s:%d", node.Endpoint.Host, node.Endpoint.Port)

	// Delete 1
	oldNode, _ := this.Find(fixedTempAddress)
	if oldNode != nil {
		ok, err := this.deletePeer(oldNode)
		if !ok {
			return false, err
		}
	}

	// Delete 1
	oldNode, _ = this.Find(node.Address)
	if oldNode != nil {
		ok, err := this.deletePeer(oldNode)
		if !ok {
			return false, err
		}
	}

	if len(strings.TrimSpace(node.Address)) <= 0 {
		node.Address = fixedTempAddress
	}

	ok, err := this.addHelper(node)
	if !ok {
		return false, err
	}

	return true, nil
}
