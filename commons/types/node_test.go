/*
 *    This file is part of Disgo-Commons library.
 *
 *    The Disgo-Commons library is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *
 *    The Disgo-Commons library is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *
 *    You should have received a copy of the GNU General Public License
 *    along with the Disgo-Commons library.  If not, see <http://www.gnu.org/licenses/>.
 */
package types

import (
	"encoding/json"
	"reflect"
	"testing"
)

//recoverMe
func recoverMe(t *testing.T) {
	if r := recover(); r != nil {
		// fmt.Println ("Recovered from: ", r)
		t.Error("Code Panic!  Test Failed")
	}
}

//TestNodeCache
func TestNodeCache(t *testing.T) {
	node1 := &Node{}
	node1.GrpcEndpoint = &Endpoint{}
	node1.GrpcEndpoint.Host = "127.0.0.1"
	node1.GrpcEndpoint.Port = 1973
	node1.HttpEndpoint = &Endpoint{}
	node1.HttpEndpoint.Host = "127.0.0.1"
	node1.HttpEndpoint.Port = 1975
	node1.Address = "123"
	node1.Cache(c)
	testNode, err := ToNodeFromCache(c, node1.Address)
	if err != nil {
		t.Error(err)
	}
	if reflect.DeepEqual(testNode, node1) == false {
		t.Error("node1 not equal to testNode")
	}
}

//TestEndPointSerialization
func TestEndPointSerialization(t *testing.T) {

	var ep Endpoint
	var ep2 Endpoint

	ep.Host = "127.0.0.1"
	ep.Port = 1975

	j, err := ep.MarshalJSON()

	if err != nil {
		t.Error("Error Marshaling Endpiont")
	}

	err = ep2.UnmarshalJSON(j)
	if err != nil {
		t.Error("Error UnMarshaling Endpoint")
	}

	if ep2.Host != ep.Host {
		t.Error("JSON serailizer/deserailizer is not working - HOST")
	}

	if ep2.Port != ep.Port {
		t.Error("JSON seralizer/deserializer is not working - PORT")
	}

}

//TestNodeSerialization
func TestNodeSerialization(t *testing.T) {
	node1 := &Node{}
	node1.GrpcEndpoint = &Endpoint{}
	node1.GrpcEndpoint.Host = "127.0.0.1"
	node1.GrpcEndpoint.Port = 1975
	node1.Address = "123"

	defer recoverMe(t)

	node2 := &Node{}
	err := json.Unmarshal([]byte(node1.String()), node2)
	if err != nil {
		t.Error("Error Marshaling Endpoint")
	}

	if node1.Address != node2.Address {
		t.Error("Marshal/Unmarshal failed = 1 Address")
	}

	if node1.GrpcEndpoint.Host != node2.GrpcEndpoint.Host {
		t.Error("Marshal/Unmarshal failed = 1 Host")
	}

	if node1.GrpcEndpoint.Port != node2.GrpcEndpoint.Port {
		t.Error("Marshal/Unmarshal Failed = 2 Port")
	}
}
