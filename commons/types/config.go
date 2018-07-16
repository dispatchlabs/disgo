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
// Package types package contains the Properties structure that defines the system properties
// used by the system to initialize communication and set the state for some objects.
package types

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"github.com/dispatchlabs/disgo/commons/utils"
)

var configInstance *Config
var configOnce sync.Once

// Config - Is the structure definition for the system properties
type Config struct {
	HttpEndpoint       *Endpoint   `json:"httpEndpoint"`
	GrpcEndpoint       *Endpoint   `json:"grpcEndpoint"`
	GrpcTimeout        int         `json:"grpcTimeout"`
	SeedEndpoints      []*Endpoint `json:"seedEndpoints"`
	DelegateEndpoints  []*Endpoint `json:"delegateEndpoints"`
	UseQuantumEntropy  bool        `json:"useQuantumEntropy"`
	GenesisTransaction string      `json:"genesisTransaction"`
}

// String - Implement the `fmt.Stringer` interface
func (this Config) String() string {
	bytes, err := json.Marshal(configInstance)
	if err != nil {
		utils.Error("unable to marshal config", err)
		return ""
	}
	return string(bytes)
}

// GetConfig -
func GetConfig() *Config {
	configOnce.Do(func() {
		configInstance = &Config{
			HttpEndpoint: &Endpoint{
				Host: "0.0.0.0",
				Port: 1975,
			},
			GrpcEndpoint: &Endpoint{
				Host: "127.0.0.1",
				Port: 1973,
			},
			GrpcTimeout: 5,
			SeedEndpoints: []*Endpoint{
				{
					Host: "seed.dispatchlabs.io",
					Port: 1973,
				},
			},
			DelegateEndpoints:  []*Endpoint{},
			GenesisTransaction: `{"hash":"a48ff2bd1fb99d9170e2bae2f4ed94ed79dbc8c1002986f8054a369655e29276","type":0,"from":"e6098cc0d5c20c6c31c4d69f0201a02975264e94","to":"3ed25f42484d517cdfc72cafb7ebc9e8baa52c2c","value":10000000,"data":"","time":0,"signature":"03c1fdb91cd10aa441e0025dd21def5ebe045762c1eeea0f6a3f7e63b27deb9c40e08b656a744f6c69c55f7cb41751eebd49c1eedfbd10b861834f0352c510b200","hertz":0,"fromName":"","toName":""}`,
		}
		var configFileName = utils.GetDisgoDir() + string(os.PathSeparator) + "config.json"
		if utils.Exists(configFileName) {
			file, err := ioutil.ReadFile(configFileName)
			if err != nil {
				utils.Error(fmt.Sprintf("unable to load config file %s", configFileName), err)
				os.Exit(1)
			}
			json.Unmarshal(file, configInstance)
			utils.Info(fmt.Sprintf("loaded config file %s", configFileName))
		} else {
			file, err := os.Create(configFileName)
			defer file.Close()
			if err != nil {
				utils.Error(fmt.Sprintf("unable to create config file %s", configFileName), err)
				panic(err)
			}

			var configAsString = configInstance.String()
			if configAsString == "" {
				utils.Error(fmt.Sprintf("unable to marshal %s", configFileName), err)
				panic(err)
			}
			fmt.Fprintf(file, configAsString)
			utils.Info(fmt.Sprintf("generated default config file %s", configFileName))
		}
		utils.Info(fmt.Sprintf("node configuration: %s", configInstance.String()))
	})

	return configInstance
}




