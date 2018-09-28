package types

import (

	"github.com/dispatchlabs/disgo/commons/crypto"

)

func Decrypt(bytes []byte,password string) (*crypto.Key, error){

	key, err := crypto.DecryptKey(bytes,password)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func create(password string, name_optional ...string) error{

	name := "my_key.json"
	if len(name_optional) > 0 {
		name = name_optional[0]
	}

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

	writeAccountFile(keystore, name)

	return nil
}

func createFromKey(key, password string, name_optional ...string) error{
	name := "my_key.json"
	if len(name_optional) > 0 {
		name = name_optional[0]
	}

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

	writeAccountFile(keystore, name)

	return nil
}