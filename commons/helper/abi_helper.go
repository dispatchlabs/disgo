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
	theAbi := GetABI(jsonMap["abi"].(string))
	method, _ := jsonMap["method"].(string)
	var result []interface{}
	for k, v := range theAbi.Methods {
		//fmt.Printf("Method: %v\n", k)
		if k == method {
			for i := 0; i < len(v.Inputs); i++ {
				arg := v.Inputs[i]
				if arg.Type.T == abi.SliceTy || arg.Type.T == abi.ArrayTy {
					dataTypeString := arg.Type.String()[0:len(arg.Type.String())-2]
					typeString := arg.Type.String()
					utils.Info("New Type: ", typeString)
					argType, err := abi.NewType(typeString)
					if err != nil {
						utils.Error(err)
					}
					var argument abi.Argument
					argument.Type = argType
					result = append(result, getValues(argument, dataTypeString, params[i].([]interface{})))
				} else {
					result = append(result, getValue(arg, params[i]))
				}
			}
		}
	}
	return result
}

func getValues(arg abi.Argument, dataTypeString string, values []interface{}) []interface{} {
	result := make([]interface{}, 0)
	if dataTypeString == "int256" || dataTypeString == "uint256" {
		//var result []*interface{}
		//tmp := make([]*big.Int, 0)
		for _, value := range values {
			result = append(result, new(big.Int).SetUint64(uint64(value.(float64))))
		}
		return result
	}

	for _, value := range values {
		temp := getValue(arg, value)
		result = append(result, temp)
	}
	return result
}

//TODO: Need to add code to handle Arrays of numeric data
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
