// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"io"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/user"
	"github.com/spf13/cobra"
)

// walletCmd represents the wallet command
var walletCmd = &cobra.Command{
	Use:   "wallet",
	Short: "wallet as an disgo node from your address",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:


Cobra wallets when I tell it to.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("wallet called")
		fmt.Println()

		fmt.Println("creating keys")
		private_key, public_key := genPPKeys(rand.Reader)
		fmt.Print("private key: 0x")
		fmt.Println(hex.EncodeToString(private_key[:]))
		fmt.Println()
		fmt.Print("public key: 0x")
		fmt.Println(hex.EncodeToString(public_key[:]))
		fmt.Println()
		storing(hex.EncodeToString(private_key[:]),hex.EncodeToString(public_key[:]))

		
	},
}


func genPPKeys(random io.Reader) (private_key_bytes, public_key_bytes []byte) {
	private_key, _ := ecdsa.GenerateKey(elliptic.P224(), random)
	private_key_bytes, _ = x509.MarshalECPrivateKey(private_key)
	public_key_bytes, _ = x509.MarshalPKIXPublicKey(&private_key.PublicKey)
	
		

	return private_key_bytes, public_key_bytes
}

func init() {
	RootCmd.AddCommand(walletCmd)
	fmt.Println("wallet init")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// walletCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// walletCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func storing(private_key, public_key string) {
	fmt.Println("Checking if there's a dispatch folder")
	fmt.Println()
	//Checks if there is a Dispatch folder in the users Home. If not, it creates one.
	usr, _ := user.Current()
	var dir = usr.HomeDir

	if _, err := os.Stat(dir + "/disgoTest"); os.IsNotExist(err) {

		//find user current directory and make the disgo folder inside it
    	os.Mkdir(dir + "/disgoTest", os.ModePerm)

    	fmt.Println("creating dispatch folder")
    	fmt.Println()
	

	//This is making the JSON file and storing the keys (note that this should be put 
		//into its own function instead of out in the functionality of the command)
		//JSON structure
type Keys struct {
    		Private string `json:"private_key"`
    		Public  string  `json:"public_key"`
    	}
    	
		//making obj to be put into JSON
    	keys := Keys{
    		Private: private_key,
    		Public: public_key,
    	}

    	//Creating Json file
    	fmt.Println("Creating JSON File")
    	jsonFile, err := os.Create(dir + "/disgoTest/do_not_touch.json")
   		if err != nil {
   			//Erroring out
      	fmt.Println("Error creating JSON file:", err)
      	return
   	}

   		//Writing JSON data to JSON file
   		fmt.Println("writing to file")
    	jsonWriter := io.Writer(jsonFile)
   		encoder := json.NewEncoder(jsonWriter)
   		err = encoder.Encode(&keys)
   		if err != nil {
      		//Erroring out
      		fmt.Println("Error encoding JSON to file:", err)
      		return
   		}
	
		fmt.Println("==> done writing to file")
	}
}

func isError(err error) bool {
	if err != nil {
		fmt.Println(err.Error())
	}

	return (err != nil)
}
