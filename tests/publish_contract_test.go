package tests

import (
	"fmt"
	"testing"

	"github.com/dispatchlabs/commons/services"
	"github.com/dispatchlabs/commons/utils"
	"github.com/dispatchlabs/dapos"
	"github.com/dispatchlabs/disgo/core"
	"github.com/dispatchlabs/disgover"
	"github.com/dispatchlabs/dvm"
)

// //privateKey []byte, tipe byte, from, to string, value, hertz, theTime int64

// t, err := types.NewTransaction("0f86ea981203b26b5b8244c8f661e30e5104555068a4bd168d3e3015db9bb25a", 0, "3ed25f42484d517cdfc72cafb7ebc9e8baa52c2c", "d70613f93152c84050e7826c4e2b0cc02c1c3b99", 999, 0, utils.ToMilliSeconds(time.Now()))
// if err != nil {
// 	utils.Error(err)
// }

// utils.Info(t.Verify())
// utils.Info(t.String())

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
}
