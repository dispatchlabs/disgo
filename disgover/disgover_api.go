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

