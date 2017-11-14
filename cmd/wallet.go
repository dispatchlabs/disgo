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
	//"io/ioutil"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	//"encoding/json"
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
	},
}

type Rankings struct {
    		private_key  string `json:"private_key"`
    		private_key_bytes string `json:"private_key_bytes"`
    		public_key_bytes string `json:"public_key_bytes"`
    	}

func genPPKeys(random io.Reader) (private_key_bytes, public_key_bytes []byte) {
	private_key, _ := ecdsa.GenerateKey(elliptic.P224(), random)
	private_key_bytes, _ = x509.MarshalECPrivateKey(private_key)
	public_key_bytes, _ = x509.MarshalPKIXPublicKey(&private_key.PublicKey)
	fmt.Println("Checking if there's a dispatch folder")
	fmt.Println()
	//Checks if there is a Dispatch folder in the users Home. If not, it creates one.
	usr, _ := user.Current()
	var dir = usr.HomeDir

	if _, err := os.Stat(dir + "/disgo"); os.IsNotExist(err) {
		fmt.Println(os.ModePerm)

		//find user current directory and make the disgo folder inside it
    	os.Mkdir(dir + "/disgo", os.ModePerm)

    	fmt.Println("creating dispatch folder")
    	fmt.Println()
    	fmt.Println("creating file to hold keys")
    	fmt.Println()
    	
    	fmt.Println(dir)
		fmt.Println("/disgo/do_not_touch.json")
    	var _, err = os.Stat(dir + "/disgo/do_not_touch.json")
    	fmt.Println("after os.Stat")
		// create file if not exists
		if os.IsNotExist(err) {
			
			file, err := os.Create(dir + "/disgo/do_not_touch.json")
			fmt.Println(err)

			fmt.Println("after os.create")
			if isError(err) { return }
			fmt.Println("after if isError")
			defer file.Close()
		}
		fmt.Println("==> done creating file", dir + "/disgo/do_not_touch.json")

		//fmt.Println(hex.EncodeToString(private_key[:]))


		//var jsonBlob = []byte(`
        //{"private_key":"`+hex.EncodeToString(private_key)+`", "private_key_bytes":"`+hex.EncodeToString(private_key_bytes[:])+`","public_key_bytes":"`+hex.EncodeToString(public_key_bytes[:])+`"}`)
		// // open file using READ & WRITE permission
		// rankings := Rankings{}
  //  		err = json.Unmarshal(jsonBlob, &rankings)
  //   	if err != nil {
  //      		 //nozzle.printError("opening config file", err.Error())
  //  		 }

  //   	rankingsJson, _ := json.Marshal(rankings)
  //   	err = ioutil.WriteFile("~/dispatch/do_not_touch.json", rankingsJson, 0644)
  //   	fmt.Printf("%+v", rankings)
	
		// fmt.Println("==> done writing to file")
	}

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

func isError(err error) bool {
	if err != nil {
		fmt.Println(err.Error())
	}

	return (err != nil)
}
