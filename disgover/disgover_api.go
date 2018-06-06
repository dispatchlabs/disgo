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
	"fmt"
	"reflect"
	"strings"
	"github.com/dispatchlabs/disgo/commons/services"
	"github.com/dispatchlabs/disgo/commons/types"
	"github.com/dispatchlabs/disgo/commons/utils"
	"github.com/libp2p/go-libp2p-peer"
)

// Find - Finds a node on the network, check internally then asks the peers if not found
func (this *DisGoverService) Find(address string) (*types.Node, error) {

	//first check cache
	node, err := types.ToNodeFromCache(services.GetCache(), address)
	if err == nil{
		return node, nil
	}

	//now check badger
	txn := services.NewTxn(true)
	defer txn.Discard()
	node, _ = types.ToNodeByAddress(txn,address)
	if node != nil{
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
	err = types.ErrNotFound
	return nil, err
}

// FindByType
func (this *DisGoverService) FindByType(tipe string) ([]*types.Node, error) {
	utils.Debug(fmt.Sprintf("finding %s nodes...", strings.ToLower(tipe)))

	// FROM-AVERY
	// var nodes []*types.Node
	// txn := services.NewTxn(true)
	// defer txn.Discard()

	// nodes, err := types.ToNodesByType(txn, tipe)
	// if err != nil { // We dont know any of the type, go ask people
	// 	//utils.Error(err)
	// 	peerID := kbucket.ConvertPeerID(peer.ID(this.ThisNode.Address))
	// 	nearestpeer := this.kdht.NearestPeer(peerID)
	// 	peerIDAsString := string(nearestpeer)
	// 	node, err := types.ToNodeByAddress(txn, peerIDAsString)
	// 	if err != nil {
	// 		utils.Error(err)
	// 	}
	// 	return this.peerFindByTypeGrpc(node, tipe)
	// }

	var nodes []*types.Node
	for _, value := range services.GetCache().Items() {
		if reflect.TypeOf(value.Object) != reflect.TypeOf(&types.Node{}) {
			continue
		}
		node := value.Object.(*types.Node)
		if node.Type == types.TypeDelegate {
			nodes = append(nodes, node)
		}
	}

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
			if !containsNode(nodes, node.Address) {
				nodes = append(nodes, node)
			}
			services.GetCache().Set(node.Address, node, types.NodeTTL)
			this.kdht.Update(peer.ID(node.Address))
		}
	}

	return nodes, nil
}

// containsNode
func containsNode(nodes []*types.Node, address string) bool {
	for _, n := range nodes {
		if n.Address == address {
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

func (this *DisGoverService) addPeer(node types.Node) (bool bool, err error){
	txn := services.NewTxn(true)
	defer txn.Discard()
	err = node.Set(txn,services.GetCache())
	if err != nil {
		return false, err
	}
	this.kdht.Update(peer.ID(node.Address))

	return true, nil
}

func (this *DisGoverService) deletePeer(node types.Node) (bool bool, err error){
	txn := services.NewTxn(true)
	defer txn.Discard()
	err = node.Unset(txn,services.GetCache())
	if err != nil {
		return false, err
	}
	this.kdht.Remove(peer.ID(node.Address))
	return true, nil
}

func (this *DisGoverService) updatePeer(node types.Node) (bool bool, err error){

	oldNode,err := this.Find(node.Address)
	ok, err := this.deletePeer(*oldNode)
	if !ok {
		return false, err
	}
	ok, err = this.addPeer(node)
	if !ok {
		return false, err
	}
	return true, nil
}