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

	if methodName == "getTxId" {
		return t.GetTxIdFunc(stub)
	}

	if methodName == "getContractInfo"{
		return t.GetContractInfo(stub)
	}


	return shim.Error("unknown method")
}

func (t *TestContract) Sum(stub shim.CMStubInterface) protogo.Response {

	args := stub.GetArgs()

	a, _ := strconv.Atoi(args["arg1"])
	b, _ := strconv.Atoi(args["arg2"])
	//key1Value, _ := stub.GetState([]byte("key1"))
	//key1Int, _ := strconv.Atoi(string(key1Value))

	c := a + b

	strc := strconv.Itoa(c)

	return shim.Success([]byte(strc))
}

func (t *TestContract) GetTxIdFunc(stub shim.CMStubInterface) protogo.Response {
	if txId,ok:=stub.GetTxId();ok!=nil{
		return shim.Error("not found")
	}else{
		return shim.Success([]byte(txId))
	}
}

func (t *TestContract) GetContractInfo(stub shim.CMStubInterface) protogo.Response{
	creatorOrgId,_ := stub.GetCreatorOrgId()
	creatorRole,_ := stub.GetCreatorRole()
	creatorPk,_ := stub.GetCreatorPk()
	senderOrgId,_ :=stub.GetSenderOrgId()
	senderRole,_ :=stub.GetSenderRole()
	senderPk,_:=stub.GetSenderPk()
	blockHeight,_:=stub.GetBlockHeight()
	txId,_:=stub.GetTxId()
	info1 := fmt.Sprintf("creatorOrgId:[%v],creatorRole:[%v],creatorPk:[%v] ",creatorOrgId,creatorRole,creatorPk)
	info2 := fmt.Sprintf("senderOrgId:[%v],senderRole:[%v],senderPk:[%v] ",senderOrgId,senderRole,senderPk)
	info3 := fmt.Sprintf("blockHeight:[%v],txId:[%v]",blockHeight,txId)
	info := info1+info2+info3
	return shim.Success([]byte(info))
}


func main() {

	//err := errors.New("sand box test err")
	err := shim.Start(new(TestContract))
	if err != nil {
		log.Fatal(err)
	}
}
