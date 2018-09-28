package types

import (
	"testing"
	"os"
	"github.com/dispatchlabs/disgo/commons/utils"
	"io/ioutil"
)

var testing_key = "5f88ec3b109dfdd5879684dce2e196dbde95e6a527395ad171e94f10e72b6953"
var testing_pass = "DisgoDance"
var testing_key_name = "testingkey.json"
var keyfilepath = "." + string(os.PathSeparator) + "config" + string(os.PathSeparator) + testing_key_name

func destructKey(){
	if utils.Exists(keyfilepath) {
		err := os.RemoveAll(keyfilepath)
		if err != nil {
			utils.Info("Failed to delete "+testing_key_name)
		}
	}
}

//TestDecrypt
func TestDecrypt(t *testing.T) {
	defer destructKey()

	err := createFromKey(testing_key,testing_pass, testing_key_name)
	if err != nil{
		t.Error(err)
	}

	bytes, err := ioutil.ReadFile(keyfilepath)
	if err != nil {
		utils.Fatal("unable to read ", keyfilepath, err)
	}

	key, err := DecryptKey(bytes,testing_pass)

	if key.GetPrivateKeyString() != testing_key{
		t.Error("key value not the same")
	}

}

//TestCreate/Encrypt
func TestCreate(t *testing.T) {
	defer destructKey()

	err := createPKey(testing_pass, testing_key_name)
	if err != nil{
		t.Error(err)
	}
}
