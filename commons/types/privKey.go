package types

import (

	"github.com/dispatchlabs/disgo/commons/crypto"

	"github.com/dispatchlabs/disgo/commons/utils"
	"os"
	"io/ioutil"
	"bufio"
	"fmt"
	"log"
	"strings"
	"golang.org/x/crypto/ssh/terminal"
	"syscall"
)


var key string

// Getkey - Returns the singleton instance of the current private key
func GetKey() string {
	key := GetAccount().PrivateKey
	if key != ""{
		return key
	}
	return readKeyFile()
}

func DecryptKey(bytes []byte,password string) (*crypto.Key, error){

	key, err := crypto.DecryptKey(bytes,password)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func createPKey(password string, name string) error{

	pKey,err := crypto.NewKey()
	if err != nil {
		return err
	}
	veryLightScryptN := 2
	veryLightScryptP := 1
	keystore, err := crypto.EncryptKey(pKey,password,veryLightScryptN, veryLightScryptP)
	if err != nil{
		return err
	}

	writeFile(keystore, name)

	return nil
}

func createFromKey(key, password string, name string) error{

	ECDSAKey,err := crypto.HexToECDSA(key)
	if err !=nil{
		return err
	}

	pkey, err := crypto.NewKeyFromECDSAKey(ECDSAKey)

	veryLightScryptN := 2
	veryLightScryptP := 1
	keystore, err := crypto.EncryptKey(pkey,password,veryLightScryptN, veryLightScryptP)
	if err != nil{
		return err
	}

	writeFile(keystore, name)

	return nil
}

func readKeyFile() string {

	fileName := GetConfig().KeyLocation
	if !utils.Exists(fileName) {
		for {
			pass := getPass("enter a password to secure your key\n")
			conf := getPass("confirm password\n")
			if pass == conf && pass != ""{
				createPKey(pass,fileName)
				break
			}
			fmt.Printf("did not match")
		}
	}
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		utils.Fatal("unable to read %s",fileName, err)
	}

	key, err := DecryptKey(bytes, getPass("enter your private key Password\n"))
	if err != nil {
		utils.Fatal("unable to read %s", fileName, err)
	}
	return key.GetPrivateKeyString()
}

func askForConfirmation(s string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/n]: ", s)

		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}

func getPass(s string)string{
	fmt.Print(s)
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return ""
	}
	password := string(bytePassword)
	return password
}

// writeFile -
func writeFile(bytes []byte, path string) {

	file, err := os.Create(path)
	defer file.Close()
	if err != nil {
		utils.Fatal(fmt.Sprintf("unable to write %s", path), err)
	}
	fmt.Fprintf(file, string(bytes))
}