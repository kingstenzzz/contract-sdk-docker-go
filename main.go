package main

import (
	"fmt"
	"log"
	"time"

	"chainmaker.org/chainmaker-contract-sdk-docker-go/pb/protogo"
	"chainmaker.org/chainmaker-contract-sdk-docker-go/shim"
)

type TestContract struct {
}

func (t *TestContract) InitContract(stub shim.CMStubInterface) protogo.Response {

	_ = stub.PutState([]byte("test_key1"), []byte("100"))

	return shim.Success([]byte("Init Success"))
}

func (t *TestContract) InvokeContract(stub shim.CMStubInterface) protogo.Response {

	args := stub.GetArgs()

	method := string(args["method"])
	switch method {
	case "display":
		return t.display()
	case "get_state":
		return t.getState(stub)
	case "put_state":
		return t.putState(stub)
	case "del_state":
		return t.delState(stub)
	case "time_out":
		return t.timeOut(stub)
	case "out_of_range":
		return t.outOfRange()
	case "cross_contract":
		return t.crossContract(stub)
	default:
		return shim.Error("unknown method")
	}
}

func (t *TestContract) display() protogo.Response {
	return shim.Success([]byte("display successful"))
}

func (t *TestContract) putState(stub shim.CMStubInterface) protogo.Response {
	args := stub.GetArgs()

	getKey := args["key"]
	getValue := args["value"]

	err := stub.PutState(getKey, getValue)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte("put state successfully"))
}

func (t *TestContract) getState(stub shim.CMStubInterface) protogo.Response {

	args := stub.GetArgs()

	getKey := args["get_state_key"]

	result, err := stub.GetState(getKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(result)
}

func (t *TestContract) delState(stub shim.CMStubInterface) protogo.Response {
	args := stub.GetArgs()
	getKey := args["get_state_key"]

	err := stub.DelState(getKey)
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

func main() {

	err := shim.Start(new(TestContract))
	if err != nil {
		log.Fatal(err)
	}
}
