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
	"time"
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

func createPKey(password string, name string) error {

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

	WriteFile(keystore, name)

	return nil
}

func CreateFromKey(key, password string) ([]byte, error) {

	ECDSAKey,err := crypto.HexToECDSA(key)
	if err !=nil{
		return nil, err
	}

	pkey, err := crypto.NewKeyFromECDSAKey(ECDSAKey)

	veryLightScryptN := 2
	veryLightScryptP := 1
	keystore, err := crypto.EncryptKey(pkey, password, veryLightScryptN, veryLightScryptP)
	if err != nil{
		return nil, err
	}

	return keystore, nil
}

func readKeyFile() string {

	fileName := GetConfig().KeyLocation
	if !utils.Exists(fileName) {
		for {
			pass := GetPass("enter a password to secure your key")
			conf := GetPass("confirm password")
      
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

	key, err := DecryptKey(bytes, GetPass("enter your private key password"))
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

func GetPass(s string)string{
	//adding the sleep because the threading for logging often makes it show up several lines up and it's confusing.
	time.Sleep(3000)
	fmt.Printf("%s: \n", s)
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return ""
	}
	//bytePassword := []byte("test")
	password := string(bytePassword)
	return password
}

// writeFile -
func WriteFile(bytes []byte, path string) {

	file, err := os.Create(path)
	defer file.Close()
	if err != nil {
		utils.Fatal(fmt.Sprintf("unable to write %s", path), err)
	}
	fmt.Fprintf(file, string(bytes))
}