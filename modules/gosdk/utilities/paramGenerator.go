package utilities

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
        "github.com/google/uuid"
	"hfrd/modules/gosdk/common"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const (
	LITERAL_PARAM     = "literal"
        UUID_PARAM        = "uuid"
	STRING_PATTERN    = "stringPattern"
	INTEGER_RANGE     = "intRange"
	PAYLOAD_RANGE     = "payloadRange"
	SEQUENTIAL_STRING = "sequentialString"
	TRANSIENT_MAP     = "transientMap"

	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

type Literal struct {
	Value string
}

type StringPattern struct {
	Regex string
}

type SequentialString struct {
	Value string
}
type IntegerRange struct {
	Min string
	Max string
}

type PayloadRange struct {
	Min string
	Max string
}

type UUID struct {
}

func (p *UUID) GetValue() string {
        rand.Seed(time.Now().UnixNano())
        id, _ := uuid.NewUUID()
        return strings.Replace(id.String(), "-", "", -1)
}

func (p *Literal) GetValue() string {
	return p.Value
}

func (p *StringPattern) GetValue() (string, error) {
	regexString, err := GenerateRegexString(p.Regex, 1)
	if err != nil {
		return "", err
	}
	return regexString, nil
}

func (p *SequentialString) GetValue(loopIndex int) string {
	loopIndexStr := strconv.Itoa(loopIndex)
	if strings.ContainsAny(p.Value, "*") {
		return strings.Replace(p.Value, "*", loopIndexStr, -1)
	}
	return p.Value + loopIndexStr
}

func (p *IntegerRange) GetValue() (string, error) {
	min, err := strconv.Atoi(p.Min)
	if err != nil {
		return "", err
	}
	max, err := strconv.Atoi(p.Max)
	if err != nil {
		return "", err
	}
	if min > max {
		return "", errors.Errorf("IntegerRange error: integerMix is bigger than integerMax")
	}
	if min == max {
		return strconv.Itoa(min), nil
	}
	return strconv.Itoa(min + rand.Intn(max-min)), nil
}

func (p *PayloadRange) GetValue() (string, error) {
	payloadMin, err := strconv.Atoi(p.Min)
	if err != nil {
		return "", err
	}
	payloadMax, err := strconv.Atoi(p.Max)
	if err != nil {
		return "", err
	}
	if payloadMin > payloadMax {
		return "", errors.Errorf("PayloadRange error: payloadMin is bigger than payloadMax")
	}
	var payloadLength int
	if payloadMin == payloadMax {
		payloadLength = payloadMin
	} else {
		payloadLength = payloadMin + rand.Intn(payloadMax-payloadMin)
	}
	payload := make([]byte, payloadLength)
	for i := range payload {
		payload[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(payload), nil
}

func GetComplexArgs(complexParams []string, loopIndex int) ([]string, error) {
	if len(complexParams) == 0 {
		return nil, nil
	}
	if complexParams[len(complexParams)-1] == "" {
		complexParams = complexParams[:len(complexParams)-1]
	}
	var chaincodeArgs []string
	var err error
	rand.Seed(time.Now().UnixNano())
	for _, param := range complexParams {
		var arg string
		paramKV := strings.Split(param, "~~~")
		switch paramKV[0] {
		case LITERAL_PARAM:
			if len(paramKV) != 2 {
				return nil, errors.Errorf("Literal type should contains 2 params")
			}
			literal := &Literal{paramKV[1]}
			arg = literal.GetValue()
                case UUID_PARAM:
                        if len(paramKV) != 1 {
                                return nil, errors.Errorf("uuid type should contains 1 param")
                        }
                        uuidgt := &UUID{}
                        arg = uuidgt.GetValue()
		case STRING_PATTERN:
			if len(paramKV) != 2 {
				return nil, errors.Errorf("stringPattern type should contains 2 params")
			}
			stringPattern := &StringPattern{paramKV[1]}
			arg, err = stringPattern.GetValue()
			if err != nil {
				return nil, err
			}
		case SEQUENTIAL_STRING:
			if len(paramKV) != 2 {
				return nil, errors.Errorf("sequentialString type should contains 2 params")
			}
			sequentialString := &SequentialString{paramKV[1]}
			arg = sequentialString.GetValue(loopIndex)
		case INTEGER_RANGE:
			if len(paramKV) != 3 {
				return nil, errors.Errorf("stringPattern type should contains 3 params")
			}

			integerRange := &IntegerRange{paramKV[1], paramKV[2]}
			arg, err = integerRange.GetValue()
			if err != nil {
				return nil, err
			}
		case PAYLOAD_RANGE:
			if len(paramKV) != 3 {
				return nil, errors.Errorf("stringPattern type should contains 3 params")
			}
			payloadRange := &PayloadRange{paramKV[1], paramKV[2]}
			arg, err = payloadRange.GetValue()
			if err != nil {
				return nil, err
			}
		default:
			return nil, errors.Errorf("GetComplexArgs error: invalid chaincode param type:" + paramKV[0] + "\n" )
		}
		chaincodeArgs = append(chaincodeArgs, arg)
	}
	return chaincodeArgs, nil
}

func GetTransientMap(complexParams []string, loopIndex int) ([]byte, error) {
	if len(complexParams) == 0 {
		return nil, nil
	}
	if complexParams[len(complexParams)-1] == "" {
		complexParams = complexParams[:len(complexParams)-1]
	}
	var transientMapV = make(map[string]interface{})
	var err error
	rand.Seed(time.Now().UnixNano())
	for _, param := range complexParams {
		var arg string
		paramKV := strings.Split(param, "~~~")
		mapKey := paramKV[1]
		switch paramKV[0] {
		case LITERAL_PARAM:
			if len(paramKV) != 3 {
				return nil, errors.Errorf("Literal type should contains 2 params")
			}
			literal := &Literal{paramKV[2]}
			arg = literal.GetValue()
                case UUID_PARAM:
                        if len(paramKV) != 1 {
                                return nil, errors.Errorf("uuid type should contains 1 param")
                        }
                        uuidgt := &UUID{}
                        arg = uuidgt.GetValue()
		case STRING_PATTERN:
			if len(paramKV) != 3 {
				return nil, errors.Errorf("stringPattern type should contains 2 params")
			}
			stringPattern := &StringPattern{paramKV[2]}
			arg, err = stringPattern.GetValue()
			if err != nil {
				return nil, err
			}
		case SEQUENTIAL_STRING:
			if len(paramKV) != 3 {
				return nil, errors.Errorf("sequentialString type should contains 2 params")
			}
			sequentialString := &SequentialString{paramKV[2]}
			arg = sequentialString.GetValue(loopIndex)
		case INTEGER_RANGE:
			if len(paramKV) != 4 {
				return nil, errors.Errorf("stringPattern type should contains 3 params")
			}
			integerRange := &IntegerRange{paramKV[2], paramKV[3]}
			arg, err = integerRange.GetValue()
			if err != nil {
				return nil, err
			}
		case PAYLOAD_RANGE:
			if len(paramKV) != 4 {
				return nil, errors.Errorf("stringPattern type should contains 3 params")
			}
			payloadRange := &PayloadRange{paramKV[2], paramKV[3]}
			arg, err = payloadRange.GetValue()
			if err != nil {
				return nil, err
			}
		default:
			return nil, errors.Errorf("GetTransientMap error: invalid param type:" + paramKV[0] + "\n" )
		}
		if argInt, err := strconv.ParseInt(arg, 10, 64); err == nil {
			transientMapV[mapKey] = argInt
		} else {
			transientMapV[mapKey] = arg
		}

	}
	transientMapJson, _ := json.Marshal(transientMapV)
	return transientMapJson, nil
}

// Generate chaincode arguments (common chaincode) and transientMap (private data chaincode)
func GenerateChaincodeParams(chaincodeParamArray []string, transientMap string, dynamicTransientMapKs []string, dynamicTransientMapVs []string, iterationIndex int) (chaincodeArgs []string, transientStaticMap, transientDynamicMap map[string][]byte, err error) {
	// Get common parameters
	chaincodeArgs, err = GetComplexArgs(chaincodeParamArray, iterationIndex)
	if err != nil {
		return nil, nil, nil, err
	}
	common.Logger.Debug(fmt.Sprintf("ChaincodeArgs: %s", chaincodeArgs))
	if len(chaincodeArgs) < 1 {
		err = errors.New("required args should not be empty")
		return nil, nil, nil, err
	}
	// Get static transient map
	var tStaticMap = make(map[string][]byte)
	if transientMap != "" {
		if err = json.Unmarshal([]byte(transientMap), &tStaticMap); err != nil {
			return nil, nil, nil, errors.New("Generate static transient map failed,due to error " + err.Error())
		}
		common.Logger.Debug(fmt.Sprintf("static transient map: %s", tStaticMap))
	}
	// Get dynamic transient map
	var dynamicTransientMap = make(map[string][]byte)
	if len(dynamicTransientMapKs) != 0 && len(dynamicTransientMapKs) == len(dynamicTransientMapVs) {
		if dynamicTransientMapKs[len(dynamicTransientMapKs)-1] == "" {
			dynamicTransientMapKs = dynamicTransientMapKs[:len(dynamicTransientMapKs)-1]
		}
		for index, dynamicTransientMapK := range dynamicTransientMapKs {
			var paramArray []string
			dynamicTransientMapV := dynamicTransientMapVs[index]
			if dynamicTransientMapV != "" {
				paramArray = strings.Split(dynamicTransientMapV, "#")
				common.Logger.Debug(fmt.Sprintf("chaincode dynamic transient map key: %s , value: %s", dynamicTransientMapV, paramArray))
			}
			transientMapValueJson, err := GetTransientMap(paramArray, iterationIndex)
			if err != nil {
				return nil, nil, nil, errors.New("Generate dynamic transient map failed,due to error " + err.Error())
			}
			dynamicTransientMap[dynamicTransientMapK] = transientMapValueJson
		}

		common.Logger.Debug(fmt.Sprintf("dynamic transient map: %s", dynamicTransientMap))
	}
	return chaincodeArgs, tStaticMap, dynamicTransientMap, nil
}
