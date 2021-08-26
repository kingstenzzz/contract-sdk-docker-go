package main

import (
	"fmt"
	"log"

	"chainmaker.org/chainmaker-contract-sdk-docker-go/pb/protogo"
	"chainmaker.org/chainmaker-contract-sdk-docker-go/shim"
)

type TestContract struct {
}

func (t *TestContract) InitContract(stub shim.CMStubInterface) protogo.Response {

	return shim.Success([]byte("Init Success"))

}

func (t *TestContract) InvokeContract(stub shim.CMStubInterface) protogo.Response {

	return shim.Success([]byte("Invoke Success"))
}

func main() {

	err := shim.Start(new(TestContract))
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
}
