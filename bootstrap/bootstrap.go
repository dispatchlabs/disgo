/*
 *    This file is part of Disgo library.
 *
 *    The Disgo library is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *
 *    The Disgo library is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *
 *    You should have received a copy of the GNU General Public License
 *    along with the Disgo library.  If not, see <http://www.gnu.org/licenses/>.
 */
package bootstrap

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"encoding/gob"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sync"

	"github.com/dispatchlabs/disgo/commons/services"
	"github.com/dispatchlabs/disgo/commons/types"
	"github.com/dispatchlabs/disgo/commons/utils"
	"github.com/dispatchlabs/disgo/dapos"
	"github.com/dispatchlabs/disgo/disgover"
	"github.com/dispatchlabs/disgo/dvm"
)

const (
	Version = "1.0.0"
)

// Server -
type Server struct {
	services []types.IService
}

// NewServer -
func NewServer() *Server {
	// Load Keys
	if _, _, err := loadKeys(); err != nil {
		utils.Error("unable to keys: " + err.Error())
	}

	return &Server{}
}

// Go
func (server *Server) Go() {
	utils.Info("booting Disgo v" + Version + "...")

	// Add services.
	server.services = append(server.services, dvm.GetDVMService())
	server.services = append(server.services, services.GetDbService())
	server.services = append(server.services, disgover.GetDisGoverService().WithGrpc().WithHttp())
	server.services = append(server.services, dapos.GetDAPoSService().WithGrpc().WithHttp())
	server.services = append(server.services, services.GetHttpService())
	server.services = append(server.services, services.GetGrpcService())

	// Run services.
	var waitGroup sync.WaitGroup
	for _, service := range server.services {
		utils.Info("starting " + utils.GetStructName(service) + "...")
		go service.Go(&waitGroup)
		waitGroup.Add(1)
	}
	waitGroup.Wait()
}

// Keys Helpers
// ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~
func saveGobKey(fileName string, key interface{}) {
	outFile, err := os.Create(fileName)
	checkError(err)
	defer outFile.Close()

	encoder := gob.NewEncoder(outFile)
	err = encoder.Encode(key)
	checkError(err)
}

func savePEMKey(fileName string, key *rsa.PrivateKey) {
	outFile, err := os.Create(fileName)
	checkError(err)
	defer outFile.Close()

	var privateKey = &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	err = pem.Encode(outFile, privateKey)
	checkError(err)
}

func savePublicPEMKey(fileName string, pubkey rsa.PublicKey) {
	asn1Bytes, err := asn1.Marshal(pubkey)
	checkError(err)

	var pemkey = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: asn1Bytes,
	}

	pemfile, err := os.Create(fileName)
	checkError(err)
	defer pemfile.Close()

	err = pem.Encode(pemfile, pemkey)
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}

func loadKeys() ([]byte, []byte, error) {
	privateKeyFile := utils.GetDisgoDir() + string(os.PathSeparator) + "key"
	publicKeyFile := utils.GetDisgoDir() + string(os.PathSeparator) + "key.pub"

	if _, err := os.Stat(privateKeyFile); os.IsNotExist(err) {
		reader := rand.Reader
		if types.GetConfig().UseQuantumEntropy {
			utils.Info("generating keys using Quantum Entropy")

			reader = NewQuantumEntropyReader()
		} else {
			utils.Info("generating keys")
		}

		bitSize := 2048

		key, keyErr := rsa.GenerateKey(reader, bitSize)
		if keyErr != nil {
			return nil, nil, err
		}

		publicKey := key.PublicKey

		saveGobKey(privateKeyFile, key)
		savePEMKey(privateKeyFile+".pem", key)

		saveGobKey(publicKeyFile, publicKey)
		savePublicPEMKey(publicKeyFile+".pem", publicKey)
	} else {
		utils.Info("loadKeys")
	}

	privateKey, error1 := ioutil.ReadFile(privateKeyFile)
	if error1 != nil {
		return nil, nil, errors.New("unable to load " + privateKeyFile)
	}

	publicKey, error2 := ioutil.ReadFile(publicKeyFile)
	if error2 != nil {
		return nil, nil, errors.New("unable to load " + publicKeyFile)
	}

	return privateKey, publicKey, nil
}

// Quantum Entropy
// ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~
// type Reader interface {
// 	Read(p []byte) (n int, err error)
// }

type QuantumEntropyReader struct{}

func NewQuantumEntropyReader() *QuantumEntropyReader {
	return &QuantumEntropyReader{}
}

func (r *QuantumEntropyReader) Read(p []byte) (n int, err error) {
	url := "http://qosmicparticles.io:8080"
	json := fmt.Sprintf(`{"api_key": "AhWvymr2HbpVo1379JbAc1bWxz0ZCWxlgUdXbPEGbJMTX4I9nslBjtqXmffA361e", "entropy_size": %d, "action": "request_quantum_entropy_gob"}`, len(p))

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(json)))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, io.EOF
	}

	body, _ := ioutil.ReadAll(resp.Body)

	copy(p[0:len(p)], body[0:len(p)])

	return len(p), nil
}
