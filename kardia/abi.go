/*
 *  Copyright 2020 KardiaChain
 *  This file is part of the go-kardia library.
 *
 *  The go-kardia library is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU Lesser General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  The go-kardia library is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *  GNU Lesser General Public License for more details.
 *
 *  You should have received a copy of the GNU Lesser General Public License
 *  along with the go-kardia library. If not, see <http://www.gnu.org/licenses/>.
 */

package kardia

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"

	"github.com/kardiachain/go-kardia/lib/abi"
	"github.com/kardiachain/go-kardia/lib/common"

	"github.com/kardiachain/kardia-explorer-backend/types"
)

// DecodeInputData returns decoded transaction input data if it match any function in staking and validator contract.
func (ec *Client) DecodeInputData(to string, input string) (*types.FunctionCall, error) {
	// return nil if input data is too short
	if len(input) <= 2 {
		return nil, nil
	}
	data, err := hex.DecodeString(strings.TrimLeft(input, "0x"))
	if err != nil {
		return nil, err
	}
	sig := data[0:4] // get the function signature (first 4 bytes of input data)
	var (
		a      *abi.ABI
		method *abi.Method
	)
	// check if the to address is staking contract, then we search for staking method in staking contract ABI
	if ec.stakingUtil.ContractAddress.Equal(common.HexToAddress(to)) {
		a = ec.stakingUtil.Abi
		method, err = ec.stakingUtil.Abi.MethodById(sig)
		if err != nil {
			return nil, err
		}
	} else if ec.paramsUtil.ContractAddress.Equal(common.HexToAddress(to)) { // if not, search for a params method
		a = ec.paramsUtil.Abi
		method, err = ec.paramsUtil.Abi.MethodById(sig)
		if err != nil {
			return nil, err
		}
	} else { // otherwise, search for a validator method
		a = ec.validatorUtil.Abi
		method, err = ec.validatorUtil.Abi.MethodById(sig)
		if err != nil {
			return nil, err
		}
	}
	// exclude the function signature, only decode and unpack the arguments
	var body []byte
	if len(data) <= 4 {
		body = []byte{}
	} else {
		body = data[4:]
	}
	args, err := ec.getInputArguments(a, method.Name, body)
	if err != nil {
		return nil, err
	}
	arguments := make(map[string]interface{})
	err = args.UnpackIntoMap(arguments, body)
	if err != nil {
		return nil, err
	}
	// convert address, bytes and string arguments into their hex representations
	for i, arg := range arguments {
		arguments[i] = parseBytesArrayIntoString(arg)
	}
	return &types.FunctionCall{
		Function:   method.String(),
		MethodID:   "0x" + hex.EncodeToString(sig),
		MethodName: method.Name,
		Arguments:  arguments,
	}, nil
}

// UnpackLog returns a log detail
func (ec *Client) UnpackLog(log *types.Log) (*types.Log, error) {
	var a *abi.ABI
	// check if the to address is staking contract, then we search for an event in staking contract ABI
	if ec.stakingUtil.ContractAddress.Equal(common.HexToAddress(log.ContractAddress)) {
		a = ec.stakingUtil.Abi
	} else if ec.paramsUtil.ContractAddress.Equal(common.HexToAddress(log.ContractAddress)) {
		a = ec.paramsUtil.Abi
	} else { // otherwise, search for a validator contract event
		a = ec.validatorUtil.Abi
	}
	event, err := a.EventByID(common.HexToHash(log.Topics[0]))
	if err != nil {
		return nil, err
	}
	argumentsValue := make(map[string]interface{})
	err = unpackLogIntoMap(a, argumentsValue, event.RawName, *log)
	if err != nil {
		return nil, err
	}
	// convert address, bytes and string arguments into their hex representations
	for i, arg := range argumentsValue {
		argumentsValue[i] = parseBytesArrayIntoString(arg)
	}
	// append unpacked data
	log.Arguments = argumentsValue
	log.Name = event.RawName + "("
	order := int64(1)
	for _, arg := range event.Inputs {
		if arg.Indexed {
			log.Name += "index_topic_" + strconv.FormatInt(order, 10) + " "
			order++
		}
		log.Name += arg.Type.String() + " " + arg.Name + ", "
	}
	log.Name = strings.TrimRight(log.Name, ", ") + ")"
	return log, nil
}

// UnpackLogIntoMap unpacks a retrieved log into the provided map.
func unpackLogIntoMap(a *abi.ABI, out map[string]interface{}, eventName string, log types.Log) error {
	data, err := hex.DecodeString(log.Data)
	if err != nil {
		return err
	}
	// unpacking unindexed arguments
	if len(data) > 0 {
		if err := a.UnpackIntoMap(out, eventName, data); err != nil {
			return err
		}
	}
	// unpacking indexed arguments
	var indexed abi.Arguments
	for _, arg := range a.Events[eventName].Inputs {
		if arg.Indexed {
			indexed = append(indexed, arg)
		}
	}
	topics := make([]common.Hash, len(log.Topics)-1)
	for i, topic := range log.Topics[1:] { // exclude the eventID (log.Topic[0])
		topics[i] = common.HexToHash(topic)
	}
	return abi.ParseTopicsIntoMap(out, indexed, topics)
}

// getInputArguments get input arguments of a contract call
func (ec *Client) getInputArguments(a *abi.ABI, name string, data []byte) (abi.Arguments, error) {
	var args abi.Arguments
	if method, ok := a.Methods[name]; ok {
		if len(data)%32 != 0 {
			return nil, fmt.Errorf("abi: improperly formatted output: %s - Bytes: [%+v]", string(data), data)
		}
		args = method.Inputs
	}
	if args == nil {
		return nil, ErrMethodNotFound
	}
	return args, nil
}

// parseBytesArrayIntoString is a utility function. It converts address, bytes and string arguments into their hex representation.
func parseBytesArrayIntoString(v interface{}) interface{} {
	if reflect.TypeOf(v).Kind() == reflect.Array {
		arr, ok := v.([32]byte)
		if !ok {
			return v
		}
		slice := arr[:]
		// convert any array of uint8 into a hex string
		if reflect.TypeOf(slice).Elem().Kind() == reflect.Uint8 {
			return common.Bytes(slice).String()
		} else {
			// otherwise recursively check other arguments
			return parseBytesArrayIntoString(v)
		}
	} else if reflect.TypeOf(v).Kind() == reflect.Ptr {
		// convert big.Int to string to avoid overflowing
		if value, ok := v.(*big.Int); ok {
			return value.String()
		}
	}
	return v
}
