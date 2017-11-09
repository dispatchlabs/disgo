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
	"net/http"
	"strconv"

	"github.com/spf13/cobra"
)

var port int


// daemonCmd represents the daemon command
var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Starts listening for pings",
	Long: `Currently just starts a listener for pings on it's own network.
In the future this daemon will also listen to all the contracts and 
messages getting passed around`,
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println("disgo node running on port " + strconv.Itoa(port))
		fmt.Println(port)
		http.HandleFunc("/", hello)
		http.ListenAndServe(":" + strconv.Itoa(port), nil)
	},
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Println("root called")
	io.WriteString(w, "Hello world!")
}

func init() {
	RootCmd.AddCommand(daemonCmd)


	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// daemonCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	daemonCmd.Flags().IntVarP(&port, "port", "p", 8000, "port to serve from")
}
