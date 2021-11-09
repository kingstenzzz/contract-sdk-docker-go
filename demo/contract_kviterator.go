package demo

import (
	"fmt"
	"log"

	"chainmaker.org/chainmaker-contract-sdk-docker-go/pb/protogo"
	"chainmaker.org/chainmaker-contract-sdk-docker-go/shim"
)

type KvIteratorContract struct {
}

func (t *KvIteratorContract) InitContract(stub shim.CMStubInterface) protogo.Response {

	return shim.Success([]byte("Init Success"))
}

func (t *KvIteratorContract) InvokeContract(stub shim.CMStubInterface) protogo.Response {

	args := stub.GetArgs()

	method := string(args["method"])
	switch method {
	case "put_state":
		return t.putState(stub)
	case "get_state":
		return t.getState(stub)
	case "construct_data":
		return t.constructData(stub)
	case "iterator_test":
		return t.kvIterator(stub)
	case "cross_contract":
		return t.crossContract(stub)
	default:
		return shim.Error("unknown method")
	}
}

func (t *KvIteratorContract) putState(stub shim.CMStubInterface) protogo.Response {
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

func (t *KvIteratorContract) getState(stub shim.CMStubInterface) protogo.Response {

	args := stub.GetArgs()

	getKey := string(args["key"])
	field := string(args["field"])

	result, err := stub.GetState(getKey, field)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(result))
}

func (t *KvIteratorContract) constructData(stub shim.CMStubInterface) protogo.Response {

	err := stub.PutState("key1", "field1", "val")
	if err != nil {
		msg := fmt.Sprintf("constructData failed, %s", err.Error())
		stub.Log(msg)
		return shim.Error(msg)
	}
	err = stub.PutState("key1", "field2", "val")
	if err != nil {
		msg := fmt.Sprintf("constructData failed, %s", err.Error())
		stub.Log(msg)
		return shim.Error(msg)
	}
	err = stub.PutState("key1", "field23", "val")
	if err != nil {
		msg := fmt.Sprintf("constructData failed, %s", err.Error())
		stub.Log(msg)
		return shim.Error(msg)
	}
	err = stub.PutState("key1", "field3", "val")
	if err != nil {
		msg := fmt.Sprintf("constructData failed, %s", err.Error())
		stub.Log(msg)
		return shim.Error(msg)
	}

	err = stub.PutState("key2", "field1", "val")
	if err != nil {
		msg := fmt.Sprintf("constructData failed, %s", err.Error())
		stub.Log(msg)
		return shim.Error(msg)
	}
	err = stub.PutState("key3", "field2", "val")
	if err != nil {
		msg := fmt.Sprintf("constructData failed, %s", err.Error())
		stub.Log(msg)
		return shim.Error(msg)
	}
	err = stub.PutState("key33", "field2", "val")
	if err != nil {
		msg := fmt.Sprintf("constructData failed, %s", err.Error())
		stub.Log(msg)
		return shim.Error(msg)
	}
	err = stub.PutState("key33", "field2", "val")
	if err != nil {
		msg := fmt.Sprintf("constructData failed, %s", err.Error())
		stub.Log(msg)
		return shim.Error(msg)
	}
	err = stub.PutState("key4", "field3", "val")
	if err != nil {
		msg := fmt.Sprintf("constructData failed, %s", err.Error())
		stub.Log(msg)
		return shim.Error(msg)
	}

	return shim.Success([]byte("construct success!"))
}

func (t *KvIteratorContract) kvIterator(stub shim.CMStubInterface) protogo.Response {
	stub.Log("==============================kvIterator start==============================")

	// 能查询出来 field1, field2, field23 三条数据
	iteratorWithField, err := stub.NewIteratorWithField("key1", "field1", "field3")
	if err != nil {
		// msg := "create with " + string(key1) + string(field1) + string(field3) + " failed"
		msg := "create with " + "key1" + "field1" + "field3" + " failed"
		stub.Log(msg)
		return shim.Error(msg)
	}
	stub.Log("==============================NewIteratorWithField==============================")

	for iteratorWithField.HasNext() {
		stub.Log("HasNext Success!!!")
		row, err := iteratorWithField.NextRow()
		if err != nil {
			msg := "iterator failed to get the next element"
			stub.Log(msg)
			return shim.Error(msg)
		}

		key, err := row.GetString("key")
		if err != nil {
			msg := fmt.Sprintf("get key failed, %s", err.Error())
			stub.Log(msg)
			return shim.Error(msg)
		}

		field, err := row.GetString("field")
		if err != nil {
			msg := fmt.Sprintf("get field failed, %s", err.Error())
			stub.Log(msg)
			return shim.Error(msg)
		}

		value, err := row.GetBytes("value")
		if err != nil {
			msg := fmt.Sprintf("get key value, %s", err.Error())
			stub.Log(msg)
			return shim.Error(msg)
		}
		stub.Log(fmt.Sprintf("=========================[NewIteratorWithField]========================="))
		stub.Log(fmt.Sprintf("=========================[key: %s]=========================!", key))
		stub.Log(fmt.Sprintf("=========================[field: %s]=========================!", field))
		stub.Log(fmt.Sprintf("=========================[value: %s]=========================!", value))
	}

	// 能超讯出来 key3, key33 两条数据
	preWithKeyIterator, err := stub.NewIteratorPrefixWithKey("key3")
	if err != nil {
		msg := "NewIteratorPrefixWithKey failed"
		stub.Log(msg)
		return shim.Error(msg)
	}

	stub.Log("==============================NewIteratorPrefixWithKey==============================")

	for preWithKeyIterator.HasNext() {
		stub.Log("HasNext Success!!!")
		row, err := preWithKeyIterator.NextRow()
		if err != nil {
			msg := "iterator failed to get the next element"
			stub.Log(msg)
			return shim.Error(msg)
		}

		key, err := row.GetString("key")
		if err != nil {
			msg := fmt.Sprintf("get key failed, %s", err.Error())
			stub.Log(msg)
			return shim.Error(msg)
		}

		field, err := row.GetString("field")
		if err != nil {
			msg := fmt.Sprintf("get field failed, %s", err.Error())
			stub.Log(msg)
			return shim.Error(msg)
		}

		value, err := row.GetBytes("value")
		if err != nil {
			msg := fmt.Sprintf("get key value, %s", err.Error())
			stub.Log(msg)
			return shim.Error(msg)
		}
		stub.Log(fmt.Sprintf("=============================[NewIteratorPrefixWithKey]========================="))
		stub.Log(fmt.Sprintf("=============================[key: %s]=========================", key))
		stub.Log(fmt.Sprintf("=============================[field: %s]=========================", field))
		stub.Log(fmt.Sprintf("=============================[value: %s]=========================", value))
	}

	// 能查询出来 field2, field23 三条数据
	preWithFieldIterator, err := stub.NewIteratorPrefixWithKeyField("key1", "field2")
	if err != nil {
		msg := "NewIteratorPrefixWithKeyField failed"
		stub.Log(msg)
		return shim.Error(msg)
	}
	stub.Log("==============================NewIteratorPrefixWithKeyField==============================")

	for preWithFieldIterator.HasNext() {
		stub.Log("HasNext Success!!!")
		row, err := preWithFieldIterator.NextRow()
		if err != nil {
			msg := "iterator failed to get the next element"
			stub.Log(msg)
			return shim.Error(msg)
		}

		key, err := row.GetString("key")
		if err != nil {
			msg := fmt.Sprintf("get key failed, %s", err.Error())
			stub.Log(msg)
			return shim.Error(msg)
		}

		field, err := row.GetString("field")
		if err != nil {
			msg := fmt.Sprintf("get field failed, %s", err.Error())
			stub.Log(msg)
			return shim.Error(msg)
		}

		value, err := row.GetBytes("value")
		if err != nil {
			msg := fmt.Sprintf("get key value, %s", err.Error())
			stub.Log(msg)
			return shim.Error(msg)
		}
		stub.Log(fmt.Sprintf("=========================[NewIteratorPrefixWithKeyField]========================="))
		stub.Log(fmt.Sprintf("=========================[key: %s]=========================!", key))
		stub.Log(fmt.Sprintf("=========================[field: %s]=========================!", field))
		stub.Log(fmt.Sprintf("=========================[value: %s]=========================!", value))
	}

	// 能查询出来 key2, key3, key33 三条数据
	iterator2, err := stub.NewIterator("key2", "key4")
	// iterator2, err := stub.NewIterator("a", "zzzzzzzzzzzzzzzzzz")
	if err != nil {
		msg := "NewIterator failed"
		stub.Log(msg)
		return shim.Error(msg)
	}
	stub.Log("==============================NewIterator==============================")

	for iterator2.HasNext() {
		stub.Log("HasNext Success!!!")
		row, err := iterator2.NextRow()
		if err != nil {
			msg := "iterator failed to get the next element"
			stub.Log(msg)
			return shim.Error(msg)
		}

		key, err := row.GetString("key")
		if err != nil {
			msg := fmt.Sprintf("get key failed, %s", err.Error())
			stub.Log(msg)
			return shim.Error(msg)
		}

		field, err := row.GetString("field")
		if err != nil {
			msg := fmt.Sprintf("get field failed, %s", err.Error())
			stub.Log(msg)
			return shim.Error(msg)
		}

		value, err := row.GetBytes("value")
		if err != nil {
			msg := fmt.Sprintf("get key value, %s", err.Error())
			stub.Log(msg)
			return shim.Error(msg)
		}
		stub.Log(fmt.Sprintf("=========================[NewIterator]========================="))
		stub.Log(fmt.Sprintf("=========================[key: %s]=========================!", key))
		stub.Log(fmt.Sprintf("=========================[field: %s]=========================!", field))
		stub.Log(fmt.Sprintf("=========================[value: %s]=========================!", value))
	}

	closed1, err := iteratorWithField.Close()
	if err != nil {
		msg := fmt.Sprintf("iteratorWithField close failed, %s", err.Error())
		stub.Log(msg)
		return shim.Error(msg)
	}

	closed2, err := iterator2.Close()
	if err != nil {
		msg := fmt.Sprintf("iterator close failed, %s", err.Error())
		stub.Log(msg)
		return shim.Error(msg)
	}

	closed3, err := preWithFieldIterator.Close()
	if err != nil {
		msg := fmt.Sprintf("preWithFieldIterator close failed, %s", err.Error())
		stub.Log(msg)
		return shim.Error(msg)
	}

	closed4, err := preWithKeyIterator.Close()
	if err != nil {
		msg := fmt.Sprintf("preWithKeyIterator close failed, %s", err.Error())
		stub.Log(msg)
		return shim.Error(msg)
	}

	if closed1 && closed2 && closed3 && closed4 {
		return shim.Success([]byte("SUCCESS!"))
	}

	return shim.Error("ERROR!")
}

func (t *KvIteratorContract) crossContract(stub shim.CMStubInterface) protogo.Response {

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

func main() {

	err := shim.Start(new(KvIteratorContract))
	if err != nil {
		log.Fatal(err)
	}
}
