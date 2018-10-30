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
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"sync"

	"github.com/dispatchlabs/disgo/commons/utils"
	"net/http"
	"time"
)

var configInstance *Config
var configOnce sync.Once

// Config - Is the structure definition for the system properties
type Config struct {
	HttpEndpoint      *Endpoint `json:"httpEndpoint"`
	GrpcEndpoint      *Endpoint `json:"grpcEndpoint"`
	GrpcTimeout       int       `json:"grpcTimeout"`
	LocalHttpApiPort  int       `json:"localHttpApiPort"`
	Seeds             []*Node   `json:"seeds"`
	DelegateAddresses []string  `json:"delegateAddresses"`
	UseQuantumEntropy bool      `json:"useQuantumEntropy"`
	IsBookkeeper      bool      `json:"isBookkeeper"`
	//GenesisTransaction string      `json:"genesisTransaction"`
	RateLimits *RateLimits `json:"rateLimits"`
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

func (this Config) ToPrettyJson() string {
	bytes, err := json.MarshalIndent(this, "", "  ")
	if err != nil {
		utils.Error("unable to marshal config", err)
		return ""
	}
	return string(bytes)
}

// ToAccountFromJson -
func ToConfigFromJson(payload []byte) (*Config, error) {
	config := &Config{}
	err := json.Unmarshal(payload, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// GetConfig - Reads from disk or Returns default config. Saves to disk
func GetConfig() *Config {
	configOnce.Do(func() {
		var configFileName = utils.GetConfigDir() + string(os.PathSeparator) + "config.json"
		if utils.Exists(configFileName) {
			file, err := ioutil.ReadFile(configFileName)
			if err != nil {
				utils.Error(fmt.Sprintf("unable to load config file %s", configFileName), err)
				os.Exit(1)
			}
			tempConfig := &Config{}
			json.Unmarshal(file, tempConfig)
			configInstance = tempConfig
			utils.Info(fmt.Sprintf("loaded config file %s", configFileName))
		} else {
			configInstance = GetDefaultConfig()
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
		temp, _ := json.MarshalIndent(configInstance, "", "   ")
		fmt.Printf(string(temp))

	})

	return configInstance
}

// GetDefaultConfig - Returns default config, does not save to disk
func GetDefaultConfig() *Config {
	address, err := GetSeedAddress()
	if err != nil {
		utils.Error(err)
	}
	return &Config{
		HttpEndpoint: &Endpoint{
			Host: "0.0.0.0",
			Port: 1975,
		},
		GrpcEndpoint: &Endpoint{
			Host: "127.0.0.1",
			Port: 1973,
		},
		GrpcTimeout:      5,
		LocalHttpApiPort: 1971,
		Seeds: []*Node{
			{
				Address: address,
				GrpcEndpoint: &Endpoint{
					Host: "seed.dispatchlabs.io",
					Port: 1973,
				},
				HttpEndpoint: &Endpoint{
					Host: "seed.dispatchlabs.io",
					Port: 1975,
				},
				Type: TypeSeed,
			},
		},
		IsBookkeeper: true,
		//GenesisTransaction: `{"hash":"a48ff2bd1fb99d9170e2bae2f4ed94ed79dbc8c1002986f8054a369655e29276","type":0,"from":"e6098cc0d5c20c6c31c4d69f0201a02975264e94","to":"3ed25f42484d517cdfc72cafb7ebc9e8baa52c2c","value":10000000,"data":"","time":0,"signature":"03c1fdb91cd10aa441e0025dd21def5ebe045762c1eeea0f6a3f7e63b27deb9c40e08b656a744f6c69c55f7cb41751eebd49c1eedfbd10b861834f0352c510b200","hertz":0,"fromName":"","toName":""}`,
		RateLimits: &RateLimits{
			EpochTime:   1538352000000000000,
			NumWindows:  240,
			TxPerMinute: 600,
			AvgHzPerTxn: 13162215217,
		},
	}
}

// GetDefaultConfig - Returns default config, does not save to disk
func GetGenesisAccount() (*Account, error) {
	var fileName = utils.GetConfigDir() + string(os.PathSeparator) + "genesis_account.json"
	genesisAccount := &Account{}
	if utils.Exists(fileName) {
		file, err := ioutil.ReadFile(fileName)
		if err != nil {
			utils.Error(fmt.Sprintf("unable to load address file %s", fileName), err)
			os.Exit(1)
		}
		json.Unmarshal(file, genesisAccount)
		utils.Info(fmt.Sprintf("loaded genesis account file %s", fileName))
	} else {
		genesisAccount = getDefaultGenesisAccount()
	}
	genesisAccount.Name = "Dispatch Labs"
	genesisAccount.Updated = time.Now()
	genesisAccount.Created = time.Now()

	return genesisAccount, nil
}

func getDefaultGenesisAccount() *Account {
	var genesisFileName = utils.GetConfigDir() + string(os.PathSeparator) + "genesis_account.json"

	file, err := os.Create(genesisFileName)
	defer file.Close()
	if err != nil {
		utils.Error(fmt.Sprintf("unable to create genesis account file %s", genesisFileName), err)
		panic(err)
	}
	var genesisString = "{\"address\": \"3ed25f42484d517cdfc72cafb7ebc9e8baa52c2c\",\"balance\": 10000000}"
	genesisAccount := &Account{}
	json.Unmarshal([]byte(genesisString), genesisAccount)
	fmt.Fprintf(file, genesisString)
	utils.Info(fmt.Sprintf("generated default config file %s", genesisString))

	return genesisAccount
}

// GetSeedAddress - Get the address for seed.dispatchlabs.io
func GetSeedAddress(seedUrl_optional ...string) (string, error) {
	seedUrl := "seed.dispatchlabs.io:1975"
	if len(seedUrl_optional) > 0 {
		seedUrl = seedUrl_optional[0]
	}

	// Get delegates.
	httpResponse, err := http.Get(fmt.Sprintf("http://%s/v1/address", seedUrl))
	if err != nil {
		return "", err
	}
	defer httpResponse.Body.Close()

	// Read body.
	body, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return "", err
	}

	// Unmarshal response.
	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}

	// Status?
	if response.Status != StatusOk {
		return "", errors.New(fmt.Sprintf("%s: %s", response.Status, response.HumanReadableStatus))
	}
	address := response.Data.(string)
	return address, nil

}
