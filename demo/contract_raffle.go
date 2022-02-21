/*
 Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.
   SPDX-License-Identifier: Apache-2.0
*/

package demo

import (
	"encoding/json"
	"fmt"
	"log"

	"chainmaker.org/chainmaker/chainmaker-contract-sdk-docker-go/pb/protogo"
	"chainmaker.org/chainmaker/chainmaker-contract-sdk-docker-go/shim"
)

type RaffleContract struct {
}

//type Peoples struct {
//	Peoples map[string]int `json:"peoples"`
//}

type Peoples struct {
	Peoples map[int]string `json:"peoples"`
}

func (f *RaffleContract) InitContract(stub shim.CMStubInterface) protogo.Response {
	return shim.Success([]byte("Init Success"))
}

func (f *RaffleContract) InvokeContract(stub shim.CMStubInterface) protogo.Response {
	method := string(stub.GetArgs()["method"])
	switch method {
	//case "register":
	//	return f.register(stub)
	case "registerAll":
		return f.registerAll(stub)
	case "query":
		return f.query(stub)
	//case "raffle":
	//	return f.raffle(stub)
	case "raffle":
		return f.raffle(stub)
	default:
		return shim.Error("invalid method")
	}
}

func (f *RaffleContract) registerAll(stub shim.CMStubInterface) protogo.Response {
	params := stub.GetArgs()

	// 获取参数
	value := params["peoples"]
	var errMsg string
	if len(value) == 0 {
		errMsg = "value should not be empty!"
		stub.Log(errMsg)
		return shim.Error(errMsg)
	}

	var peoples Peoples
	err := json.Unmarshal(value, &peoples)
	for i := 1; i < len(peoples.Peoples); i++ {
		if name, ok := peoples.Peoples[i]; !ok || len(name) == 0 {
			errMsg = fmt.Sprintf("[registerAll] name should not be empty for number %d", i)
			stub.Log(errMsg)
			return shim.Error(errMsg)
		}
	}
	err = stub.PutStateByte("peoples", "", value)
	if err != nil {
		errMsg = fmt.Sprintf("[register] put state bytes failed, %s", err)
		stub.Log(errMsg)
		return shim.Error(errMsg)
	}
	// 返回结果
	return shim.Success([]byte("ok"))
}

func (f *RaffleContract) raffle(stub shim.CMStubInterface) protogo.Response {
	params := stub.GetArgs()

	var errMsg string
	argTimestamp := string(params["timestamp"])
	if len(argTimestamp) == 0 {
		errMsg = "argTimestamp should not be empty!"
		stub.Log(errMsg)
		return shim.Error(errMsg)
	}

	peoplesData, err := stub.GetStateByte("peoples", "")
	if err != nil {
		errMsg = "get peoples data from store failed!"
		stub.Log(errMsg)
		return shim.Error(errMsg)
	}

	var peoples Peoples
	err = json.Unmarshal(peoplesData, &peoples)
	if err != nil {
		errMsg = fmt.Sprintf("unmarshal peoples data failed, %s", err)
		stub.Log(errMsg)
		return shim.Error(errMsg)
	}
	num := f.bkdrHash(argTimestamp)
	num = num % len(peoples.Peoples)

	result := fmt.Sprintf("num: %d, name: %s", num, peoples.Peoples[num])
	delete(peoples.Peoples, num)
	newPeoplesData, err := json.Marshal(peoples.Peoples)
	if err != nil {
		errMsg = fmt.Sprintf("marshal new peoples data failed, %s", err)
		stub.Log(errMsg)
		return shim.Error(errMsg)
	}
	err = stub.PutStateByte("peoples", "", newPeoplesData)
	if err != nil {
		errMsg = fmt.Sprintf("put new peoples data failed, %s", err)
		stub.Log(errMsg)
		return shim.Error(errMsg)
	}

	return shim.Success([]byte(result))
}

func (f *RaffleContract) bkdrHash(timestamp string) int {
	hash := 0
	seed := 131
	for x := range timestamp {
		hash = hash*seed + x
	}
	return hash & 0x7FFFFFFF
}

func (f *RaffleContract) query(stub shim.CMStubInterface) protogo.Response {
	peoplesData, err := stub.GetStateByte("peoples", "")
	if err != nil {
		errMsg := "get peoples data from store failed!"
		stub.Log(errMsg)
		return shim.Error(errMsg)
	}

	return shim.Success(peoplesData)
}

//func (f *RaffleContract) register(stub shim.CMStubInterface) protogo.Response {
//	params := stub.GetArgs()
//
//	// 获取参数
//	name := string(params["name"])
//	if len(name) == 0 {
//		msg := "name should not be empty!"
//		stub.Log(msg)
//		return shim.Error(msg)
//	}
//
//	index := 0
//	var errMsg string
//	result, err := stub.GetState("index", "")
//	if err != nil {
//		errMsg = fmt.Sprintf("[register] get index from store failed, %s", err)
//		stub.Log(errMsg)
//		return shim.Error(errMsg)
//	}
//	if len(result) != 0 {
//		index, err = strconv.Atoi(result)
//	}
//	if err != nil {
//		errMsg = fmt.Sprintf("[register] convert index failed, %s", err)
//		stub.Log(errMsg)
//		return shim.Error(errMsg)
//	}
//	if index == 0 {
//		index = 100
//	} else {
//		index++
//	}
//	resultBytes, err := stub.GetStateByte("peoples", "")
//	if err != nil {
//		errMsg = fmt.Sprintf("[register] get peoples data failed, %s", err)
//		stub.Log(errMsg)
//		return shim.Error(errMsg)
//	}
//	var peoples Peoples
//	if err = json.Unmarshal(resultBytes, &peoples); err != nil {
//		errMsg = fmt.Sprintf("[register] unmarshal peoples failed, %s", err)
//		stub.Log(errMsg)
//		return shim.Error(errMsg)
//	}
//	_, ok := peoples.Peoples[name]
//	if ok {
//		errMsg = fmt.Sprintf("[register] %s has already been register", name)
//		stub.Log(errMsg)
//		return shim.Error(errMsg)
//	}
//	peoples.Peoples[name] = index
//	peoplesBytes, err := json.Marshal(peoples)
//	if err != nil {
//		errMsg = fmt.Sprintf("[register] mashal peoples failed, %s", err)
//		stub.Log(errMsg)
//		return shim.Error(errMsg)
//	}
//	err = stub.PutStateByte("peoples", "", peoplesBytes)
//	if err != nil {
//		errMsg = fmt.Sprintf("[register] put state bytes failed, %s", err)
//		stub.Log(errMsg)
//		return shim.Error(errMsg)
//	}
//	indexStr := strconv.Itoa(index)
//	err = stub.PutState("index", "", indexStr)
//	if err != nil {
//		errMsg = fmt.Sprintf("[register] put state bytes failed, %s", err)
//		stub.Log(errMsg)
//		return shim.Error(errMsg)
//	}
//
//	// 返回结果
//	return shim.Success([]byte(indexStr))
//}

func main() {

	err := shim.Start(new(RaffleContract))
	if err != nil {
		log.Fatal(err)
	}
}
