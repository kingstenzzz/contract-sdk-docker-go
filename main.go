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

func (t *TestContract) Init(stub shim.CMStubInterface) protogo.Response {
	fmt.Println("sandbox - init has been invoke")

	return shim.Success([]byte("Init Success"))
}

func (t *TestContract) Invoke(stub shim.CMStubInterface) protogo.Response {
	fmt.Println("sandbox - invoke has been invoke")

	args := stub.GetArgs()

	a, _ := strconv.Atoi(args["arg1"])
	b, _ := strconv.Atoi(args["arg2"])

	c := a + b

	strc := strconv.Itoa(c)

	fmt.Println(c)

	return shim.Success([]byte(strc))

}

func main() {

	err := shim.Start(new(TestContract))
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
}
