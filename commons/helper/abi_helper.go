package helper

import (
	"github.com/dispatchlabs/disgo/dvm/ethereum/abi"
	"github.com/dispatchlabs/disgo/commons/utils"
	"encoding/hex"
	"reflect"
	"math/big"
)

func GetConvertedParams(jsonMap map[string]interface{}) []interface{} {
	params, _ := jsonMap["params"].([]interface{})
	if params == nil || len(params) == 0 {
		return params
	}
	abi := GetABI(jsonMap["abi"].(string))
	method, _ := jsonMap["method"].(string)
	var result []interface{}
	for k, v := range abi.Methods {
		//fmt.Printf("Method: %v\n", k)
		if k == method {
			for i := 0; i < len(v.Inputs); i++ {
				result = append(result, getValue(v.Inputs[i], params[i]))
			}
		}
		//for _, args := range v.Inputs {
		//	fmt.Printf("\tInput Name: %v\n", args.Name)
		//	fmt.Printf("\tInput Type: %v\n", args.Type)
		//}
	}
	return result
}

//numerics from json are always serialized as float64
func getValue(arg abi.Argument, value interface{}) interface{} {
	if arg.Type.String() == "int256" || arg.Type.String() == "uint256" {
		return big.NewInt(int64(value.(float64)))
	}
	switch arg.Type.Kind {
	case reflect.Int:
		return big.NewInt(int64(value.(float64)))
	case reflect.Int8:
		return int8(value.(float64))
	case reflect.Int16:
		return int16(value.(float64))
	case reflect.Int32:
		return int32(value.(float64))
	case reflect.Int64:
		return int64(value.(float64))
	case reflect.Uint:
		return big.NewInt(int64(value.(float64)))
	case reflect.Uint8:
		return int8(value.(float64))
	case reflect.Uint16:
		return int16(value.(float64))
	case reflect.Uint32:
		return int32(value.(float64))
	case reflect.Uint64:
		return int64(value.(float64))
	default:
		return value
	}
}

func GetABI(data string) abi.ABI {
	bytes, err := hex.DecodeString(data)
	var abi abi.ABI
	err = abi.UnmarshalJSON(bytes)
	if err != nil {
		utils.Error(err)
	}
	return abi
}
