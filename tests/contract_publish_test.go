package tests

import (
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

var nrOfServices = 6

// Test1_PublishContract -
func Test1_PublishContract(t *testing.T) {
	utils.InitMainPackagePath()

	utils.Events().On(services.Events.DbServiceInitFinished, test1_AllServicesInitFinished)
	utils.Events().On(services.Events.GrpcServiceInitFinished, test1_AllServicesInitFinished)
	utils.Events().On(services.Events.HttpServiceInitFinished, test1_AllServicesInitFinished)

	utils.Events().On(disgover.Events.DisGoverServiceInitFinished, test1_AllServicesInitFinished)
	utils.Events().On(dapos.Events.DAPoSServiceInitFinished, test1_AllServicesInitFinished)
	utils.Events().On(dvm.Events.DVMServiceInitFinished, test1_AllServicesInitFinished)

	utils.Info(fmt.Sprintf("NR of services left to be started: %d", nrOfServices))

	server := core.NewServer()
	server.Go()
}

func test1_AllServicesInitFinished() {
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
	// var value int64 = 0
	// var hertz int64 = 0
	var theTime int64 = utils.ToMilliSeconds(time.Now())
	var code = "6060604052600160005534610000575b6101168061001e6000396000f30060606040526000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806329e99f07146046578063cb0d1c76146074575b6000565b34600057605e6004808035906020019091905050608e565b6040518082815260200191505060405180910390f35b34600057608c6004808035906020019091905050609d565b005b6000816000540290505b919050565b806000600082825401925050819055507ffa753cb3413ce224c9858a63f9d3cf8d9d02295bdb4916a594b41499014bb57f6000546040518082815260200191505060405180910390a15b505600a165627a7a723058203f0887849cabeb36c6f72cc345c5ff3521d889356357e6815dd8dbe9f7c41bbe0029"

	var tx, _ = types.NewContractTransaction(
		privateKey,
		from,
		code,
		theTime,
	)

	// makePostReuqest(
	// 	"http://localhost:1975/v1/transactions",
	// 	[]byte(tx.String()),
	// )

	var fakeReceipt = &types.Receipt{
		Id:                  "fake1",
		Type:                "fake1",
		Status:              "fake1",
		HumanReadableStatus: "fake1",
	}
	services.GetCache().Set(fakeReceipt.Id, fakeReceipt, types.ReceiptCacheTTL)

	var fakeGossip = &types.Gossip{
		ReceiptId:   fakeReceipt.Id,
		Transaction: *tx,
	}
	dapos.GetDAPoSService().Temp_ProcessTransaction(fakeGossip)

	test2_AllServicesInitFinished()
}

// func makePostReuqest(url string, data []byte) {
// 	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
// 	req.Header.Set("Content-Type", "application/json")

// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer resp.Body.Close()

// 	responseAsBytes, _ := ioutil.ReadAll(resp.Body)
// 	utils.Info("RESP-Status : ", resp.Status)
// 	utils.Info("RESP-Headers: ", resp.Header)
// 	utils.Info("RESP-Body   : ", string(responseAsBytes))
// }
