package tests

import (
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/dispatchlabs/commons/services"
	"github.com/dispatchlabs/commons/types"
	"github.com/dispatchlabs/commons/utils"
	"github.com/dispatchlabs/dapos"
	"github.com/dispatchlabs/disgo/core"
	"github.com/dispatchlabs/disgover"
	"github.com/dispatchlabs/dvm"
)

// Test2_PublishContract -
func Test2_PublishContract(t *testing.T) {
	utils.InitMainPackagePath()

	utils.Events().On(services.Events.DbServiceInitFinished, test2_AllServicesInitFinished)
	utils.Events().On(services.Events.GrpcServiceInitFinished, test2_AllServicesInitFinished)
	utils.Events().On(services.Events.HttpServiceInitFinished, test2_AllServicesInitFinished)

	utils.Events().On(disgover.Events.DisGoverServiceInitFinished, test2_AllServicesInitFinished)
	utils.Events().On(dapos.Events.DAPoSServiceInitFinished, test2_AllServicesInitFinished)
	utils.Events().On(dvm.Events.DVMServiceInitFinished, test2_AllServicesInitFinished)

	utils.Info(fmt.Sprintf("NR of services left to be started: %d", nrOfServices))

	server := core.NewServer()
	server.Go()
}

func test2_AllServicesInitFinished() {
	nrOfServices--
	utils.Info(fmt.Sprintf("NR of services left to be started: %d", nrOfServices))

	if nrOfServices > 0 {
		return
	}

	// Taken from Genesis
	var privateKey = "0f86ea981203b26b5b8244c8f661e30e5104555068a4bd168d3e3015db9bb25a"
	// var tipe byte = 0
	var from = "3ed25f42484d517cdfc72cafb7ebc9e8baa52c2c"
	// var to = "d70613f93152c84050e7826c4e2b0cc02c1c3b99"
	var to = "c3be1a3a5c6134cca51896fadf032c4c61bc355e"
	var value int64 = 10
	// var hertz int64 = 0
	var code = `[{
		"type":"function",
		"name":"test",
		"constant":true,
		"inputs":[{
			"name":"i","type":"uint256"
		}],
		"outputs":[{
			"name":"",
			"type":"uint256"
		}],
		"payable":false,
		"stateMutability":"view"
	},{
		"type":"function",
		"name":"testAsync",
		"constant":false,
		"inputs":[{
			"name":"i",
			"type":"uint256"
		}],
		"outputs":[],
		"payable":false,
		"stateMutability":"nonpayable"
	},{
		"type":"event",
		"name":"LocalChange",
		"anonymous":false,
		"inputs":[{
			"indexed":false,
			"name":"",
			"type":"uint256"
		}]
	}]"`

	var theTime int64 = utils.ToMilliSeconds(time.Now())
	var method = "test"

	var tx, _ = types.NewContractCallTransaction(
		privateKey,
		from,
		to,
		hex.EncodeToString([]byte(code)),
		theTime,
		method,
		value,
	)

	var fakeReceipt = &types.Receipt{
		Id:                  "fake2",
		Type:                "fake2",
		Status:              "fake2",
		HumanReadableStatus: "fake2",
	}
	services.GetCache().Set(fakeReceipt.Id, fakeReceipt, types.ReceiptCacheTTL)

	var fakeGossip = &types.Gossip{
		ReceiptId:   fakeReceipt.Id,
		Transaction: *tx,
	}
	dapos.GetDAPoSService().ProcessTransaction(fakeGossip)
}
