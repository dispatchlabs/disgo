package types

import (
	"github.com/dispatchlabs/disgo/commons/utils"
	"os"
	"fmt"
	"time"
	"sync"
	"encoding/json"

	"io/ioutil"
)

type Version struct {
	Version		string
	BuildTime	string
}

var versionInstance *Version
var versionOnce sync.Once



func SetVersion(version string, timestamp string) {
	versionOnce.Do(func() {
		versionInstance = &Version{
			Version: version,
			BuildTime: timestamp,
		}
	})
}

func GetVersion() *Version {
	if versionInstance == nil {

		var versionFileName = utils.GetConfigDir() + string(os.PathSeparator) + "version.json"
		if utils.Exists(versionFileName) {
			file, err := ioutil.ReadFile(versionFileName)
			if err != nil {
				utils.Error(fmt.Sprintf("unable to load version file %s", versionFileName), err)
				os.Exit(1)
			}
			json.Unmarshal(file, versionInstance)
			utils.Info(fmt.Sprintf("loaded version file %s", versionFileName))
		} else {
			versionInstance = getDefaultVersion()
			file, err := os.Create(versionFileName)
			defer file.Close()
			if err != nil {
				utils.Error(fmt.Sprintf("unable to create version file %s", versionFileName), err)
				panic(err)
			}
			fmt.Fprintf(file, versionInstance.String())
		}
		utils.Info(fmt.Sprintf("generated default config file %s", versionInstance.String()))
	}
	return versionInstance
}

func getDefaultVersion() *Version {
	t := time.Now()
	tm := fmt.Sprintf("%d-%02d-%02d-%02d:%02d:%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
	return &Version{
		Version: "2.4.0",
		BuildTime: tm,
	}
}

func (this Version) String() string {
	bytes, err := json.Marshal(this)
	if err != nil {
		utils.Error("unable to marshal Window", err)
		return ""
	}
	return string(bytes)
}