package helper

import (
	"github.com/dispatchlabs/disgo/dvm/ethereum/abi"
	"github.com/dispatchlabs/disgo/commons/utils"
	"encoding/hex"
	"reflect"
	"math/big"
	"github.com/pkg/errors"
	"fmt"
)

func GetConvertedParams(jsonMap map[string]interface{}) ([]interface{}, error) {
	params, ok := jsonMap["params"].([]interface{})
	if !ok {
		return nil, errors.Errorf("value for field 'params' must be an array")
	}
	if params == nil || len(params) == 0 {
		return params, nil
	}
	theAbi := GetABI(jsonMap["abi"].(string))
	method, _ := jsonMap["method"].(string)
	var result []interface{}
	found := false
	for k, v := range theAbi.Methods {
		//fmt.Printf("Method: %v\n", k)
		if k == method {
			found = true
			if len(v.Inputs) != len(params) {
				return nil, errors.New(fmt.Sprintf("This method requires %d parameters and %d are provided", len(v.Inputs), len(params)))
			}
			for i := 0; i < len(v.Inputs); i++ {
				arg := v.Inputs[i]
				if arg.Type.T == abi.SliceTy || arg.Type.T == abi.ArrayTy {
					result = append(result, getValues(arg, params[i].([]interface{})))
				} else {
					result = append(result, getValue(arg, params[i]))
				}
			}
		}
	}
	if !found {
		return nil, errors.New(fmt.Sprintf("This method %s is not valid for this contract", method))
	}
	return result, nil
}

func getValues(arg abi.Argument, values []interface{}) interface{} {
	var result interface{}
	dataTypeString := arg.Type.String()[0:len(arg.Type.String())-2]

	switch dataTypeString {
	case "int256", "uint256", "int", "uint":
		dynarrin := make([]*big.Int, 0)
		for _, value := range values {
			dynarrin = append(dynarrin, big.NewInt(int64(value.(float64))))
		}
		result = dynarrin
		break
	case "int8":
		array := make([]int8, 0)
		for _, value := range values {
			array = append(array, int8(value.(float64)))
		}
		result = array
		break
	case "int16":
		array := make([]int16, 0)
		for _, value := range values {
			array = append(array, int16(value.(float64)))
		}
		result = array
		break
	case "int32":
		array := make([]int32, 0)
		for _, value := range values {
			array = append(array, int32(value.(float64)))
		}
		result = array
		break
	case "int64":
		array := make([]int64, 0)
		for _, value := range values {
			array = append(array, int64(value.(float64)))
		}
		result = array
		break
	case "uint8":
		array := make([]uint8, 0)
		for _, value := range values {
			array = append(array, uint8(value.(float64)))
		}
		result = array
		break
	case "uint16":
		array := make([]uint16, 0)
		for _, value := range values {
			array = append(array, uint16(value.(float64)))
		}
		result = array
		break
	case "uint32":
		array := make([]uint32, 0)
		for _, value := range values {
			array = append(array, uint32(value.(float64)))
		}
		result = array
		break
	case "uint64":
		array := make([]uint64, 0)
		for _, value := range values {
			array = append(array, uint64(value.(float64)))
		}
		result = array
		break
	default:
		array := make([]interface{}, 0)
		for _, value := range values {
			array = append(array, value)
		}
		result = array
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
		return uint8(value.(float64))
	case reflect.Uint16:
		return uint16(value.(float64))
	case reflect.Uint32:
		return uint32(value.(float64))
	case reflect.Uint64:
		return uint64(value.(float64))
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
