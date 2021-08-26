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

	return shim.Success([]byte("Init Success"))

}

func (t *TestContract) InvokeContract(stub shim.CMStubInterface) protogo.Response {

	stub.Log("just testing")

	args := stub.GetArgs()

	val1, _ := strconv.Atoi(args["arg1"])
	val2, _ := strconv.Atoi(args["arg2"])

	val := val1 + val2

	stub.EmitEvent("topic1", []byte("testing"))

	return shim.Success([]byte(strconv.Itoa(val)))
}

func main() {

	err := shim.Start(new(TestContract))
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
}
