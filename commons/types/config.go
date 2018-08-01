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
	DelegateAddresses  []string    `json:"delegateAddresses"`
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
			GenesisTransaction: `{"hash":"56c723b72be5bc3b13b25bbbbe8bae62e3f3793406f1248743d2b011953514dd","type":0,"from":"dbae0d9e9b819c41ab7801a748f9c928fc9cf317","to":"3ed25f42484d517cdfc72cafb7ebc9e8baa52c2c","value":"10000000","time":0,"signature":"d686befb458e45ae1f968475af1bee9a2e1bf01db100b80fa2a08ef855ef7aed174fd9b052083cbcb1c3395663ade9274bf1f4057dd7a0e286e13eb862c978bc00","hertz":0}`,
		}
		var configFileName = utils.GetConfigDir() + string(os.PathSeparator) + "config.json"
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




