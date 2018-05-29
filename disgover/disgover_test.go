package disgover

import (
	"testing"
	"github.com/dispatchlabs/commons/types"
)

func recoverMe(t *testing.T) {
	if r := recover(); r != nil {
		// fmt.Println ("Recovered from: ", r)
		t.Error("Code Panic!  Test Failed")
	}
}

func TestDisGoverService(t *testing.T) {

	//testing full
	//checking cache
	
	// TODO: Fix Test
	
	// defer recoverMe(t)
	// DS := GetDisGoverService()
	// node := makeTestNode()
	// defer DS.deletePeer(node)
	// DS.addPeer(node)
	// _, err := DS.Find(node.Address)
	// if err != nil {
	// 	t.Errorf("cannot find in cache")
	// }
	}

func TestDisGoverServiceAdd(t *testing.T) {
	// TODO: Fix Test
	
	// defer recoverMe(t)
	// DS := GetDisGoverService()
	// node := makeTestNode()
	// defer DS.deletePeer(node)
	// _, err := DS.Find(node.Address)
	// if err == nil {
	// 	t.Errorf("found non-added node")
	// }
	// DS.addPeer(node)
	// _, err = DS.Find(node.Address)
	// if err != nil {
	// 	t.Errorf("did not find added node")
	// }
}

func TestDisGoverServiceDelete(t *testing.T) {
	defer recoverMe(t)
	DS := GetDisGoverService()
	node := makeTestNode()
	DS.deletePeer(node)
	_, err := DS.Find(node.Address)
	if err == nil {
		t.Errorf("found non-added node")
	}
}

func TestDisGoverServiceUpdate(t *testing.T) {
	// TODO: Fix Test
	
	// //defer recoverMe(t)
	// DS := GetDisGoverService()

	// node := makeTestNode()
	// defer DS.deletePeer(node)

	// _, err := DS.addPeer(node)
	// if err != nil {
	// 	t.Errorf("problem adding node")
	// }

	// node2 := makeChangedTestNode()
	// defer DS.deletePeer(node2)

	// _, err = DS.updatePeer(node2)
	// if err != nil {
	// 	t.Errorf("problem updating node")
	// }
	// testAgainst, err:= DS.Find(node.Address)
	// if testAgainst.Type == node.Type{
	// 	t.Errorf("node did not change")
	// }
}



func makeTestNode() types.Node{

	testNode  := &types.Node{
			Address:  "de3a0dba79b563588b15e38909ce206eb83dd27b53150e53c858036978b23412",
			Endpoint: types.GetConfig().GrpcEndpoint,
			Type:     "0",
		}
	return *testNode

}

func makeChangedTestNode() types.Node{

	testNode  := &types.Node{
		Address:  "de3a0dba79b563588b15e38909ce206eb83dd27b53150e53c858036978b23412",
		Endpoint: types.GetConfig().GrpcEndpoint,
		Type:     "1",
	}
	return *testNode

}