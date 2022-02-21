/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package demo

import (
	"fmt"
	"log"
	"time"

	"chainmaker.org/chainmaker/chainmaker-contract-sdk-docker-go/pb/protogo"
	"chainmaker.org/chainmaker/chainmaker-contract-sdk-docker-go/shim"
)

type TestContract struct {
}

func (t *TestContract) InitContract(stub shim.CMStubInterface) protogo.Response {

	return shim.Success([]byte("Init Success"))
}

func (t *TestContract) InvokeContract(stub shim.CMStubInterface) protogo.Response {

	args := stub.GetArgs()

	method := string(args["method"])
	switch method {
	case "display":
		return t.display()
	case "put_state":
		return t.putState(stub)
	case "put_state_byte":
		return t.putStateByte(stub)
	case "put_state_from_key":
		return t.putStateFromKey(stub)
	case "put_state_from_key_byte":
		return t.putStateFromKeyByte(stub)
	case "get_state":
		return t.getState(stub)
	case "get_state_byte":
		return t.getStateByte(stub)
	case "get_state_from_key":
		return t.getStateFromKey(stub)
	case "get_state_from_key_byte":
		return t.getStateFromKeyByte(stub)
	case "del_state":
		return t.delState(stub)
	case "time_out":
		return t.timeOut(stub)
	case "out_of_range":
		return t.outOfRange()
	case "cross_contract":
		return t.crossContract(stub)
	case "cross_contract_self":
		return t.callSelf(stub)

	// kvIterator
	case "construct_data":
		return t.constructData(stub)
	case "kv_iterator_test":
		return t.kvIterator(stub)

	// keyHistoryIterator
	case "key_history_kv_iter":
		return t.keyHistoryIter(stub)

	// getSenderAddress
	case "get_sender_address":
		return t.GetSenderAddrTest(stub)

	default:
		msg := fmt.Sprintf("unknown method")
		return shim.Error(msg)
	}
}

func (t *TestContract) GetSenderAddrTest(stub shim.CMStubInterface) protogo.Response {
	senderAddr, err := stub.GetSenderAddr()
	if err != nil {
		msg := "GetSenderAddr failed"
		stub.Log(msg)
		return shim.Error(msg)
	}

	l := len([]byte(senderAddr))
	msg := fmt.Sprintf("=== sender address: [%s] len: %d===", senderAddr, l)
	stub.Log(msg)
	return shim.Success([]byte(senderAddr))
}

func (t *TestContract) keyHistoryIter(stub shim.CMStubInterface) protogo.Response {
	stub.Log("===Key History Iter START===")
	args := stub.GetArgs()
	key := string(args["key"])
	field := string(args["field"])

	iter, err := stub.NewHistoryKvIterForKey(key, field)
	if err != nil {
		msg := "NewHistoryIterForKey failed"
		stub.Log(msg)
		return shim.Error(msg)
	}

	stub.Log("===create iter success===")

	count := 0
	for iter.HasNext() {
		stub.Log("HasNext")
		count++
		km, err := iter.Next()
		if err != nil {
			msg := "iterator failed to get the next element"
			stub.Log(msg)
			return shim.Error(msg)
		}

		stub.Log(fmt.Sprintf("=== Data History [%d] Info:", count))
		stub.Log(fmt.Sprintf("=== Key: [%s]", km.Key))
		stub.Log(fmt.Sprintf("=== Field: [%s]", km.Field))
		stub.Log(fmt.Sprintf("=== Value: [%s]", km.Value))
		stub.Log(fmt.Sprintf("=== TxId: [%s]", km.TxId))
		stub.Log(fmt.Sprintf("=== BlockHeight: [%d]", km.BlockHeight))
		stub.Log(fmt.Sprintf("=== IsDelete: [%t]", km.IsDelete))
		stub.Log(fmt.Sprintf("=== Timestamp: [%s]", km.Timestamp))
	}

	closed, err := iter.Close()
	if !closed || err != nil {
		msg := fmt.Sprintf("iterator close failed, %s", err.Error())
		stub.Log(msg)
		return shim.Error(msg)
	}
	stub.Log("===iter close success===")

	stub.Log("===Key History Iter END===")

	return shim.Success([]byte("get key history successfully"))
}

func (t *TestContract) display() protogo.Response {
	return shim.Success([]byte("display successful"))
}

func (t *TestContract) putState(stub shim.CMStubInterface) protogo.Response {
	args := stub.GetArgs()

	getKey := string(args["key"])
	getField := string(args["field"])
	getValue := string(args["value"])

	err := stub.PutState(getKey, getField, getValue)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte("put state successfully"))
}

func (t *TestContract) putStateByte(stub shim.CMStubInterface) protogo.Response {
	args := stub.GetArgs()

	getKey := string(args["key"])
	getField := string(args["field"])
	getValue := args["value"]

	err := stub.PutStateByte(getKey, getField, getValue)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte("put state successfully"))
}

func (t *TestContract) putStateFromKey(stub shim.CMStubInterface) protogo.Response {
	args := stub.GetArgs()

	getKey := string(args["key"])
	getValue := string(args["value"])

	err := stub.PutStateFromKey(getKey, getValue)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte("put state successfully"))
}

func (t *TestContract) putStateFromKeyByte(stub shim.CMStubInterface) protogo.Response {
	args := stub.GetArgs()

	getKey := string(args["key"])
	getValue := args["value"]

	err := stub.PutStateFromKeyByte(getKey, getValue)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte("put state successfully"))
}

func (t *TestContract) getState(stub shim.CMStubInterface) protogo.Response {

	args := stub.GetArgs()

	getKey := string(args["key"])
	field := string(args["field"])

	result, err := stub.GetState(getKey, field)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(result))
}

func (t *TestContract) getStateByte(stub shim.CMStubInterface) protogo.Response {

	args := stub.GetArgs()

	getKey := string(args["key"])
	field := string(args["field"])

	result, err := stub.GetStateByte(getKey, field)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(result)
}

func (t *TestContract) getStateFromKey(stub shim.CMStubInterface) protogo.Response {

	args := stub.GetArgs()

	getKey := string(args["key"])

	result, err := stub.GetStateFromKey(getKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(result))
}

func (t *TestContract) getStateFromKeyByte(stub shim.CMStubInterface) protogo.Response {
	args := stub.GetArgs()

	getKey := string(args["key"])

	result, err := stub.GetStateFromKeyByte(getKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(result)
}

func (t *TestContract) delState(stub shim.CMStubInterface) protogo.Response {
	args := stub.GetArgs()

	getKey := string(args["key"])
	getField := string(args["field"])

	err := stub.DelState(getKey, getField)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte("delete successfully"))
}

func (t *TestContract) timeOut(stub shim.CMStubInterface) protogo.Response {
	time.Sleep(5 * time.Second)
	return shim.Success([]byte("success finish timeout"))
}

func (t *TestContract) outOfRange() protogo.Response {
	var group []string
	group[0] = "abc"
	fmt.Println(group[0])
	return shim.Success([]byte("exit out of range"))
}

func (t *TestContract) crossContract(stub shim.CMStubInterface) protogo.Response {

	args := stub.GetArgs()

	contractName := string(args["contract_name"])
	contractVersion := string(args["contract_version"])

	calledMethod := string(args["contract_method"])

	crossContractArgs := make(map[string][]byte)
	crossContractArgs["method"] = []byte(calledMethod)

	// response could be correct or error
	response := stub.CallContract(contractName, contractVersion, crossContractArgs)
	stub.EmitEvent("cross contract", []string{"success"})
	return response
}

func (t *TestContract) callSelf(stub shim.CMStubInterface) protogo.Response {

	stub.Log("testing call self")

	contractName := "contract_test03"
	contractVersion := "v1.0.0"

	crossContractArgs := make(map[string][]byte)
	crossContractArgs["method"] = []byte("cross_contract_self")

	response := stub.CallContract(contractName, contractVersion, crossContractArgs)
	return response
}

// constructData 提供Kv迭代器的测试数据
/*
	| Key   | Field   | Value |
	| ---   | ---     | ---   |
	| key1  | field1  | val   |
	| key1  | field2  | val   |
	| key1  | field23 | val   |
	| ey1   | field3  | val   |
	| key2  | field1  | val   |
	| key3  | field2  | val   |
	| key33 | field2  | val   |
	| key4  | field3  | val   |
*/
func (t *TestContract) constructData(stub shim.CMStubInterface) protogo.Response {
	dataList := []struct {
		key   string
		field string
		value string
	}{
		{key: "key1", field: "field1", value: "val"},
		{key: "key1", field: "field2", value: "val"},
		{key: "key1", field: "field23", value: "val"},
		{key: "key1", field: "field3", value: "val"},
		{key: "key2", field: "field1", value: "val"},
		{key: "key3", field: "field2", value: "val"},
		{key: "key33", field: "field2", value: "val"},
		{key: "key33", field: "field2", value: "val"},
		{key: "key4", field: "field3", value: "val"},
	}

	for _, data := range dataList {
		err := stub.PutState(data.key, data.field, data.value)
		if err != nil {
			msg := fmt.Sprintf("constructData failed, %s", err.Error())
			stub.Log(msg)
			return shim.Error(msg)
		}
	}

	return shim.Success([]byte("construct success!"))
}

// kvIterator 前置数据
/*
	| Key   | Field   | Value |
	| ---   | ---     | ---   |
	| key1  | field1  | val   |
	| key1  | field2  | val   |
	| key1  | field23 | val   |
	| ey1   | field3  | val   |
	| key2  | field1  | val   |
	| key3  | field2  | val   |
	| key33 | field2  | val   |
	| key4  | field3  | val   |
*/
func (t *TestContract) kvIterator(stub shim.CMStubInterface) protogo.Response {
	stub.Log("===construct START===")
	dataList := []struct {
		key   string
		field string
		value string
	}{
		{key: "key1", field: "field1", value: "val"},
		{key: "key1", field: "field2", value: "val"},
		{key: "key1", field: "field23", value: "val"},
		{key: "key1", field: "field3", value: "val"},
		{key: "key2", field: "field1", value: "val"},
		{key: "key3", field: "field2", value: "val"},
		{key: "key33", field: "field2", value: "val"},
		{key: "key33", field: "field2", value: "val"},
		{key: "key4", field: "field3", value: "val"},
	}

	for _, data := range dataList {
		err := stub.PutState(data.key, data.field, data.value)
		if err != nil {
			msg := fmt.Sprintf("constructData failed, %s", err.Error())
			stub.Log(msg)
			return shim.Error(msg)
		}
	}
	stub.Log("===construct END===")

	stub.Log("===kvIterator START===")
	iteratorList := make([]shim.ResultSetKV, 4)

	// 能查询出 key2, key3, key33 三条数据
	iterator, err := stub.NewIterator("key2", "key4")
	if err != nil {
		msg := "NewIterator failed"
		stub.Log(msg)
		return shim.Error(msg)
	}
	iteratorList[0] = iterator

	// 能查询出 field1, field2, field23 三条数据
	iteratorWithField, err := stub.NewIteratorWithField("key1", "field1", "field3")
	if err != nil {
		// msg := "create with " + string(key1) + string(field1) + string(field3) + " failed"
		msg := "create with " + "key1" + "field1" + "field3" + " failed"
		stub.Log(msg)
		return shim.Error(msg)
	}
	iteratorList[1] = iteratorWithField

	// 能查询出 key3, key33 两条数据
	preWithKeyIterator, err := stub.NewIteratorPrefixWithKey("key3")
	if err != nil {
		msg := "NewIteratorPrefixWithKey failed"
		stub.Log(msg)
		return shim.Error(msg)
	}
	iteratorList[2] = preWithKeyIterator

	// 能查询出 field2, field23 三条数据
	preWithKeyFieldIterator, err := stub.NewIteratorPrefixWithKeyField("key1", "field2")
	if err != nil {
		msg := "NewIteratorPrefixWithKeyField failed"
		stub.Log(msg)
		return shim.Error(msg)
	}
	iteratorList[3] = preWithKeyFieldIterator

	for index, iter := range iteratorList {
		index++
		stub.Log(fmt.Sprintf("===iterator %d START===", index))
		for iter.HasNext() {
			stub.Log("HasNext Success")
			key, field, value, err := iter.Next()
			if err != nil {
				msg := "iterator failed to get the next element"
				stub.Log(msg)
				return shim.Error(msg)
			}

			stub.Log(fmt.Sprintf("===[key: %s]===", key))
			stub.Log(fmt.Sprintf("===[field: %s]===", field))
			stub.Log(fmt.Sprintf("===[value: %s]===", value))
		}

		closed, err := iter.Close()
		if !closed || err != nil {
			msg := fmt.Sprintf("iterator %d close failed, %s", index, err.Error())
			stub.Log(msg)
			return shim.Error(msg)
		}
		stub.Log(fmt.Sprintf("===iterator %d END===", index))
	}
	stub.Log("===kvIterator END===")

	return shim.Success([]byte("SUCCESS"))
}

func main() {
	err := shim.Start(new(TestContract))
	if err != nil {
		log.Fatal(err)
	}
}
