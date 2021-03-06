/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package demo

import (
	"fmt"
	"log"

	"github.com/kingstenzzz/contract-sdk-docker-go/pb/protogo"
	"github.com/kingstenzzz/contract-sdk-docker-go/shim"
)

type ContractCut struct {
}

func (t *ContractCut) InitContract(stub shim.CMStubInterface) protogo.Response {

	return shim.Success([]byte("Init Success"))
}

func (t *ContractCut) InvokeContract(stub shim.CMStubInterface) protogo.Response {

	args := stub.GetArgs()

	method := string(args["method"])
	switch method {
	case "save":
		return t.save(stub)
	case "findByFileHash":
		return t.findByFileHash(stub)
	default:
		msg := fmt.Sprintf("unknown method")
		return shim.Error(msg)
	}
}

func (t *ContractCut) save(stub shim.CMStubInterface) protogo.Response {
	key := string(stub.GetArgs()["file_key"])
	name := stub.GetArgs()["file_name"]

	err := stub.PutStateByte(key, "", name)
	if err != nil {
		return shim.Error("fail to save")
	}
	return shim.Success([]byte("success"))
}

func (t *ContractCut) findByFileHash(stub shim.CMStubInterface) protogo.Response {
	key := string(stub.GetArgs()["file_key"])

	_, err := stub.GetStateByte(key, "")
	if err != nil {
		return shim.Error("fail to find")
	}
	return shim.Success([]byte("success"))
}

func main() {
	err := shim.Start(new(ContractCut))
	if err != nil {
		log.Fatal(err)
	}
}
