package main

import (
	"chainmaker.org/chainmaker-contract-sdk-docker-go/pb/protogo"
	"chainmaker.org/chainmaker-contract-sdk-docker-go/shim"
	"fmt"
	"log"
)

type TestContract struct {
}

func (t *TestContract) Init(stub shim.CMStubInterface) protogo.Response {
	fmt.Println("sandbox - init has been invoke")

	return shim.Success([]byte("Init Success"))
}

func (t *TestContract) Invoke(stub shim.CMStubInterface) protogo.Response {
	fmt.Println("sandbox - invoke has been invoke")

	return shim.Success([]byte("Invoke Success"))

}

func main() {

	err := shim.Start(new(TestContract))
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
}
