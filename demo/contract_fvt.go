package demo

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
	default:
		return shim.Error("unknown method")
	}
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

func main() {
	err := shim.Start(new(TestContract))
	if err != nil {
		log.Fatal(err)
	}
}
