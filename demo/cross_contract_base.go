package demo

import (
	"chainmaker.org/chainmaker-contract-sdk-docker-go/pb/protogo"
	"chainmaker.org/chainmaker-contract-sdk-docker-go/shim"
	"fmt"
	"log"
)

type Contract1 struct {
}

func (c *Contract1) InitContract(stub shim.CMStubInterface) protogo.Response {

	return shim.Success([]byte("Init Success"))
}

func (c *Contract1) InvokeContract(stub shim.CMStubInterface) protogo.Response {

	args := stub.GetArgs()

	method := string(args["method"])
	switch method {
	case "save":
		return c.save(stub)
	case "find":
		return c.find(stub)
	default:
		msg := fmt.Sprintf("unknown method")
		return shim.Error(msg)
	}
}

func (c *Contract1) save(stub shim.CMStubInterface) protogo.Response {
	params := stub.GetArgs()

	key := string(params["key"])
	value := string(params["value"])

	err := stub.PutStateFromKey(key, value)
	if err != nil {
		errMsg := fmt.Sprintf("fail to save key [%s], value [%s]: err: [%s]",
			key, value, err)
		return shim.Error(errMsg)
	}
	return shim.Success([]byte("successfully save"))
}

func (c *Contract1) find(stub shim.CMStubInterface) protogo.Response {
	params := stub.GetArgs()

	key := string(params["key"])
	value, err := stub.GetStateFromKey(key)
	if err != nil {
		errMsg := fmt.Sprintf("fail to get key [%s], value [%s]: err: [%s]",
			key, value, err)
		return shim.Error(errMsg)
	}
	return shim.Success([]byte(value))
}

func main() {
	err := shim.Start(new(Contract1))
	if err != nil {
		log.Fatal(err)
	}
}
