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

const (
	defaultVersion = "1.0.0"
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
			utils.Info("Version file contents: ", string(file))
			versionInstance = &Version{}
			err = json.Unmarshal(file, versionInstance)
			if err != nil {
				utils.Error(err)
			}
			utils.Info(fmt.Sprintf("loaded version file %s\n%v", versionFileName, versionInstance))
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
		Version: defaultVersion,
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