package helper

import (
	"github.com/dispatchlabs/disgo/dvm/ethereum/abi"
	"encoding/hex"
	"reflect"
	"math/big"
	"github.com/pkg/errors"
	"fmt"
	"strings"
	"github.com/dispatchlabs/disgo/commons/crypto"
	"github.com/dispatchlabs/disgo/commons/types"
)

func GetConvertedParams(tx *types.Transaction) ([]interface{}, error) {

	//if tx.Params == nil || len(tx.Params) == 0 {
	//	return tx.Params, nil
	//}
	theABI, err := GetABI(tx.Abi)
	if err != nil {
		return nil, err
	}
	var result []interface{}
	found := false
	for k, v := range theABI.Methods {
		if k == tx.Method {
			found = true
			if len(v.Inputs) != len(tx.Params) {
				return nil, errors.New(fmt.Sprintf("This method %s, requires %d parameters and %d are provided", tx.Method, len(v.Inputs), len(tx.Params)))
			}
			for i := 0; i < len(v.Inputs); i++ {
				arg := v.Inputs[i]
				if arg.Type.T == abi.SliceTy || arg.Type.T == abi.ArrayTy {
					value, valErr := getValues(arg, tx.Params[i].([]interface{}))
					if valErr != nil {
						msg := fmt.Sprintf("Invalid value provided for method %s: %v", tx.Method, valErr.Error())
						return nil, errors.New(msg)
					}
					result = append(result, value)
				} else if arg.Type.T == abi.AddressTy {
					addressAsString, valErr := getValue(arg, tx.Params[i])
					addressAsByteArray := crypto.GetAddressBytes(addressAsString.(string))
					if len(addressAsByteArray) < 0 {
						msg := fmt.Sprintf("Invalid value provided for method %s: %v", tx.Method, valErr.Error())
						return nil, errors.New(msg)
					}
					result = append(result, addressAsByteArray)
				} else {
					value, valErr := getValue(arg, tx.Params[i])
					if valErr != nil {
						msg := fmt.Sprintf("Invalid value provided for method %s: %v", tx.Method, valErr.Error())
						return nil, errors.New(msg)
					}
					result = append(result, value)
				}
			}
		}
	}
	if !found {
		return nil, errors.New(fmt.Sprintf("This method %s is not valid for this contract", tx.Method))
	}
	return result, nil
}

func getValues(arg abi.Argument, values []interface{}) (interface{}, error) {
	var result interface{}
	dataTypeString := arg.Type.String()[0:len(arg.Type.String())-2]
	if strings.HasPrefix(dataTypeString, "int") || strings.HasPrefix(dataTypeString, "uint") {
		for _, value := range values {
			_, isNumber := value.(float64)
			if !isNumber {return nil, errors.Errorf("only number value required in input array, a provided value is '%v'", value)}
		}
	}
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
	return result, nil
}

//numerics from json are always serialized as float64
func getValue(arg abi.Argument, value interface{}) (interface{}, error) {
	nbrValue, isNumber := value.(float64)
	if arg.Type.String() == "int256" || arg.Type.String() == "uint256" {
		if !isNumber {return nil, errors.Errorf("number value required, provided value is '%v'", value)}
		return big.NewInt(int64(value.(float64))), nil
	}
	switch arg.Type.Kind {
	case reflect.Int:
		if !isNumber {return nil, errors.Errorf("number value required, provided value is '%v'", value)}
		return big.NewInt(int64(nbrValue)), nil
	case reflect.Int8:
		if !isNumber {return nil, errors.Errorf("number value required, provided value is '%v'", value)}
		return int8(nbrValue), nil
	case reflect.Int16:
		if !isNumber {return nil, errors.Errorf("number value required, provided value is '%v'", value)}
		return int16(nbrValue), nil
	case reflect.Int32:
		if !isNumber {return nil, errors.Errorf("number value required, provided value is '%v'", value)}
		return int32(nbrValue), nil
	case reflect.Int64:
		if !isNumber {return nil, errors.Errorf("number value required, provided value is '%v'", value)}
		return int64(nbrValue), nil
	case reflect.Uint:
		if !isNumber {return nil, errors.Errorf("number value required, provided value is '%v'", value)}
		return big.NewInt(int64(nbrValue)), nil
	case reflect.Uint8:
		if !isNumber {return nil, errors.Errorf("number value required, provided value is '%v'", value)}
		return uint8(nbrValue), nil
	case reflect.Uint16:
		if !isNumber {return nil, errors.Errorf("number value required, provided value is '%v'", value)}
		return uint16(nbrValue), nil
	case reflect.Uint32:
		if !isNumber {return nil, errors.Errorf("number value required, provided value is '%v'", value)}
		return uint32(nbrValue), nil
	case reflect.Uint64:
		if !isNumber {return nil, errors.Errorf("number value required, provided value is '%v'", value)}
		return uint64(nbrValue), nil
	case reflect.Bool:
		val, ok := value.(bool)
		if !ok {
			return nil, errors.Errorf("boolean value required, provided value is %v", value)
		}
		if val == true || val == false {
			return val, nil
		}
	default:
		return value, nil
	}
	return value, nil
}

func GetABI(data string) (*abi.ABI, error) {
	bytes, err := hex.DecodeString(data)
	var abi abi.ABI
	err = abi.UnmarshalJSON(bytes)
	if err != nil {
		return nil, errors.New("The ABI provided is not a valid ABI structure")
	}
	return &abi, nil
}
