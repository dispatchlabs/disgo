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

// TestKeyEncryptDecrypt -
func TestKeyEncryptDecrypt(t *testing.T) {
	utils.InitMainPackagePath()

	utils.Events().On(services.Events.DbServiceInitFinished, allServicesInitFinished)
	utils.Events().On(services.Events.GrpcServiceInitFinished, allServicesInitFinished)
	utils.Events().On(services.Events.HttpServiceInitFinished, allServicesInitFinished)

	utils.Events().On(disgover.Events.DisGoverServiceInitFinished, allServicesInitFinished)
	utils.Events().On(dapos.Events.DAPoSServiceInitFinished, allServicesInitFinished)
	utils.Events().On(dvm.Events.DVMServiceInitFinished, allServicesInitFinished)

	utils.Info(fmt.Sprintf("NR of services left to be started: %d", nrOfServices))

	server := core.NewServer()
	server.Go()
}

func allServicesInitFinished() {
	nrOfServices--
	utils.Info(fmt.Sprintf("NR of services left to be started: %d", nrOfServices))

	if nrOfServices != 0 {
		return
	}

	// {"address":"3ed25f42484d517cdfc72cafb7ebc9e8baa52c2c","privateKey":"0f86ea981203b26b5b8244c8f661e30e5104555068a4bd168d3e3015db9bb25a"}

	var privateKey = "0f86ea981203b26b5b8244c8f661e30e5104555068a4bd168d3e3015db9bb25a"
	var tipe byte = 0
	var from = "3ed25f42484d517cdfc72cafb7ebc9e8baa52c2c"
	var to = "d70613f93152c84050e7826c4e2b0cc02c1c3b99"
	var value int64 = 999
	var hertz int64 = 0
	var theTime int64 = utils.ToMilliSeconds(time.Now())

	t, err := types.NewTransaction(
		privateKey,
		tipe,
		from,
		to,
		value,
		hertz,
		theTime,
	)
	if err != nil {
		utils.Error(err)
	}

	utils.Info(t.Verify())
	utils.Info(t.String())
}
