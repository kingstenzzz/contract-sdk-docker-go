package demo

import (
	"chainmaker.org/chainmaker-contract-sdk-docker-go/pb/protogo"
	"chainmaker.org/chainmaker-contract-sdk-docker-go/shim"
	"fmt"
	"log"
)

type Contract2 struct {
}

func (c *Contract2) InitContract(stub shim.CMStubInterface) protogo.Response {

	return shim.Success([]byte("Init Success"))
}

func (c *Contract2) InvokeContract(stub shim.CMStubInterface) protogo.Response {

	args := stub.GetArgs()

	method := string(args["method"])
	switch method {
	case "display":
		return c.display(stub)
	case "cross_call":
		return c.crossCall(stub)
	default:
		msg := fmt.Sprintf("unknown method")
		return shim.Error(msg)
	}
}

func (c *Contract2) display(stub shim.CMStubInterface) protogo.Response {
	return shim.Success([]byte("successfully display"))
}

func (c *Contract2) crossCall(stub shim.CMStubInterface) protogo.Response {

	contractName := "contract1"
	contractVersion := "1.0.0"

	crossContractArgs := make(map[string][]byte)
	crossContractArgs["method"] = []byte("find")
	crossContractArgs["key"] = []byte("key")

	result := stub.CallContract(contractName, contractVersion, crossContractArgs)
	return result
}

func main() {
	err := shim.Start(new(Contract2))
	if err != nil {
		log.Fatal(err)
	}
}
