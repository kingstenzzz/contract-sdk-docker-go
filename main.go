package main

import (
	"chainmaker.org/chainmaker-contract-sdk-docker-go/pb/protogo"
	"chainmaker.org/chainmaker-contract-sdk-docker-go/shim"
	"fmt"
	"log"
	"strconv"
)

type TestContract struct {
}

func (t *TestContract) InitContract(stub shim.CMStubInterface) protogo.Response {

	err := stub.PutState([]byte("key1"), []byte("5"))
	if err != nil {
		return shim.Error("err to put state")
	}

	return shim.Success([]byte("Init Success"))
}

func (t *TestContract) InvokeContract(stub shim.CMStubInterface) protogo.Response {

	args := stub.GetArgs()
	methodName := args["arg0"]
	if methodName == "sum" {
		return t.Sum(stub)
	}

	return shim.Error("unknown method")
}

func (t *TestContract) Sum(stub shim.CMStubInterface) protogo.Response {

	args := stub.GetArgs()

	a, _ := strconv.Atoi(args["arg1"])
	b, _ := strconv.Atoi(args["arg2"])
	key1Value, _ := stub.GetState([]byte("key1"))
	key1Int, _ := strconv.Atoi(string(key1Value))

	c := a + b + key1Int

	strc := strconv.Itoa(c)

	return shim.Success([]byte(strc))
}

func main() {

	err := shim.Start(new(TestContract))
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
}
