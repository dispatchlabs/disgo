package helper

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"reflect"
	"strings"

	"github.com/dispatchlabs/disgo/commons/crypto"
	"github.com/dispatchlabs/disgo/commons/types"
	"github.com/dispatchlabs/disgo/commons/utils"
	"github.com/dispatchlabs/disgo/dvm/ethereum/abi"
	"github.com/pkg/errors"
)

func GetConvertedParams(tx *types.Transaction) ([]interface{}, error) {
	utils.Info("GetConvertedParams --> ", tx.Params)
	params, err := tx.ToParams()
	if err != nil {
		return nil, err
	}
	theABI, err := GetABI(tx.Abi)
	if err != nil {
		return nil, err
	}
	var result []interface{}
	found := false
	for k, v := range theABI.Methods {
		if k == tx.Method {
			found = true
			if params == nil || len(params) == 0 {
				return params, nil
			}
			if len(v.Inputs) != len(params) {
				return nil, errors.New(fmt.Sprintf("The method %s, requires %d parameters and %d are provided", tx.Method, len(v.Inputs), len(params)))
			}
			for i := 0; i < len(v.Inputs); i++ {
				arg := v.Inputs[i]
				if arg.Type.T == abi.SliceTy || arg.Type.T == abi.ArrayTy {
					value, valErr := getValues(arg, params[i].([]interface{}))
					if valErr != nil {
						msg := fmt.Sprintf("Invalid value provided for method %s: %v", tx.Method, valErr.Error())
						return nil, errors.New(msg)
					}
					result = append(result, value)
				} else if arg.Type.T == abi.AddressTy {
					addressAsString, valErr := getValue(arg, params[i])
					addressAsByteArray := crypto.GetAddressBytes(addressAsString.(string))
					if len(addressAsByteArray) < 0 {
						msg := fmt.Sprintf("Invalid value provided for method %s: %v", tx.Method, valErr.Error())
						return nil, errors.New(msg)
					}
					result = append(result, addressAsByteArray)
				} else if arg.Type.T == abi.BytesTy{
					//params, valErr := base64.StdEncoding.DecodeString(params[i].(string))
					str := params[i].(string)
					value := []byte(str)

					if err != nil{
						return nil, err
					}
					result = append(result, value)
				} else if arg.Type.T == abi.FixedBytesTy {
					str := params[i].(string)
					value := []byte(str)
					if err != nil {
						return nil, err
					}
					result = appendFixedBytesArray(arg, result, value)
				} else {
					value, valErr := getValue(arg, params[i])
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
		return nil, errors.New(fmt.Sprintf("This method '%s' is not valid for this contract", tx.Method))
	}
	return result, nil
}

func getValues(arg abi.Argument, values []interface{}) (interface{}, error) {
	var result interface{}
	dataTypeString := arg.Type.String()[0 : len(arg.Type.String())-2]
	if strings.HasPrefix(dataTypeString, "int") || strings.HasPrefix(dataTypeString, "uint") {
		for _, value := range values {
			_, isNumber := value.(float64)
			if !isNumber {
				return nil, errors.Errorf("only number value required in input array, a provided value is '%v'", value)
			}
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
		if !isNumber {
			return nil, errors.Errorf("number value required, provided value is '%v'", value)
		}
		return big.NewInt(int64(value.(float64))), nil
	}
	switch arg.Type.Kind {
	case reflect.Int:
		if !isNumber {
			return nil, errors.Errorf("number value required, provided value is '%v'", value)
		}
		return big.NewInt(int64(nbrValue)), nil
	case reflect.Int8:
		if !isNumber {
			return nil, errors.Errorf("number value required, provided value is '%v'", value)
		}
		return int8(nbrValue), nil
	case reflect.Int16:
		if !isNumber {
			return nil, errors.Errorf("number value required, provided value is '%v'", value)
		}
		return int16(nbrValue), nil
	case reflect.Int32:
		if !isNumber {
			return nil, errors.Errorf("number value required, provided value is '%v'", value)
		}
		return int32(nbrValue), nil
	case reflect.Int64:
		if !isNumber {
			return nil, errors.Errorf("number value required, provided value is '%v'", value)
		}
		return int64(nbrValue), nil
	case reflect.Uint:
		if !isNumber {
			return nil, errors.Errorf("number value required, provided value is '%v'", value)
		}
		return big.NewInt(int64(nbrValue)), nil
	case reflect.Uint8:
		if !isNumber {
			return nil, errors.Errorf("number value required, provided value is '%v'", value)
		}
		return uint8(nbrValue), nil
	case reflect.Uint16:
		if !isNumber {
			return nil, errors.Errorf("number value required, provided value is '%v'", value)
		}
		return uint16(nbrValue), nil
	case reflect.Uint32:
		if !isNumber {
			return nil, errors.Errorf("number value required, provided value is '%v'", value)
		}
		return uint32(nbrValue), nil
	case reflect.Uint64:
		if !isNumber {
			return nil, errors.Errorf("number value required, provided value is '%v'", value)
		}
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

func appendFixedBytesArray(arg abi.Argument, result []interface{}, value []byte) []interface{} {
	switch arg.Type.Size {
	case 1:
		byte1 := new([1]byte)
		copy(byte1[:], value)
		result = append(result, byte1)
	case 2:
		byte2 := new([2]byte)
		copy(byte2[:], value)
		result = append(result, byte2)
	case 3:
		byte3 := new([3]byte)
		copy(byte3[:], value)
		result = append(result, byte3)
	case 4:
		byte4 := new([4]byte)
		copy(byte4[:], value)
		result = append(result, byte4)
	case 5:
		byte5 := new([5]byte)
		copy(byte5[:], value)
		result = append(result, byte5)
	case 6:
		byte6 := new([6]byte)
		copy(byte6[:], value)
		result = append(result, byte6)
	case 7:
		byte7 := new([7]byte)
		copy(byte7[:], value)
		result = append(result, byte7)
	case 8:
		byte8 := new([8]byte)
		copy(byte8[:], value)
		result = append(result, byte8)
	case 9:
		byte9 := new([9]byte)
		copy(byte9[:], value)
		result = append(result, byte9)
	case 10:
		byte10 := new([10]byte)
		copy(byte10[:], value)
		result = append(result, byte10)
	case 11:
		byte11 := new([11]byte)
		copy(byte11[:], value)
		result = append(result, byte11)
	case 12:
		byte12 := new([12]byte)
		copy(byte12[:], value)
		result = append(result, byte12)
	case 13:
		byte13 := new([13]byte)
		copy(byte13[:], value)
		result = append(result, byte13)
	case 14:
		byte14 := new([14]byte)
		copy(byte14[:], value)
		result = append(result, byte14)
	case 15:
		byte15 := new([15]byte)
		copy(byte15[:], value)
		result = append(result, byte15)
	case 16:
		byte16 := new([16]byte)
		copy(byte16[:], value)
		result = append(result, byte16)
	case 17:
		byte17 := new([17]byte)
		copy(byte17[:], value)
		result = append(result, byte17)
	case 18:
		byte18 := new([18]byte)
		copy(byte18[:], value)
		result = append(result, byte18)
	case 19:
		byte19 := new([19]byte)
		copy(byte19[:], value)
		result = append(result, byte19)
	case 20:
		byte20 := new([20]byte)
		copy(byte20[:], value)
		result = append(result, byte20)
	case 21:
		byte21 := new([21]byte)
		copy(byte21[:], value)
		result = append(result, byte21)
	case 22:
		byte22 := new([22]byte)
		copy(byte22[:], value)
		result = append(result, byte22)
	case 23:
		byte23 := new([23]byte)
		copy(byte23[:], value)
		result = append(result, byte23)
	case 24:
		byte24 := new([24]byte)
		copy(byte24[:], value)
		result = append(result, byte24)
	case 25:
		byte25 := new([25]byte)
		copy(byte25[:], value)
		result = append(result, byte25)
	case 26:
		byte26 := new([26]byte)
		copy(byte26[:], value)
		result = append(result, byte26)
	case 27:
		byte27 := new([27]byte)
		copy(byte27[:], value)
		result = append(result, byte27)
	case 28:
		byte28 := new([28]byte)
		copy(byte28[:], value)
		result = append(result, byte28)
	case 29:
		byte29 := new([29]byte)
		copy(byte29[:], value)
		result = append(result, byte29)
	case 30:
		byte30 := new([30]byte)
		copy(byte30[:], value)
		result = append(result, byte30)
	case 31:
		byte31 := new([31]byte)
		copy(byte31[:], value)
		result = append(result, byte31)
	case 32:
		byte32 := new([32]byte)
		copy(byte32[:], value)
		result = append(result, byte32)
	}
	return result
}

func GetABI(data string) (*abi.ABI, error) {
	//runes := []rune(data)
	// ... Convert back into a string from rune slice.
	//safeSubstring := string(runes[0:10])
	//utils.Info("GetAbi %s\n%s\n", safeSubstring, utils.GetCallStackWithFileAndLineNumber())
	bytes, err := hex.DecodeString(data)
	var abi abi.ABI
	err = abi.UnmarshalJSON(bytes)
	if err != nil {
		utils.Error(err, utils.GetCallStackWithFileAndLineNumber())
		return nil, errors.New(fmt.Sprintf("The ABI provided is not a valid ABI structure: %s", string(bytes)))
	}
	return &abi, nil
}

