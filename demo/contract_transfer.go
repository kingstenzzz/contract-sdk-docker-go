package demo

import (
	"chainmaker.org/chainmaker/chainmaker-contract-sdk-docker-go/pb/protogo"
	"chainmaker.org/chainmaker/chainmaker-contract-sdk-docker-go/shim"
	"fmt"
	"log"
	"strconv"
)

type TransferContract struct {
}

func (t *TransferContract) InitContract(stub shim.CMStubInterface) protogo.Response {
	return shim.Success([]byte("Init Success"))
}

func (t *TransferContract) InvokeContract(stub shim.CMStubInterface) protogo.Response {

	args := stub.GetArgs()

	method := string(args["method"])
	switch method {
	case "init":
		return t.init(stub)
	case "transfer":
		return t.transfer(stub)
	default:
		msg := fmt.Sprintf("unknown method")
		return shim.Error(msg)
	}
}

func (t *TransferContract) init(stub shim.CMStubInterface) protogo.Response {
	args := stub.GetArgs()

	accFrom := string(args["accFrom"])
	accTo := string(args["accTo"])
	fromBal := string(args["from_bal"])
	toBal := string(args["to_bal"])
	startIndex := string(args["start_index"]) // 分批创建账户 - start index
	endIndex := string(args["end_index"])     // 分批创建账户 - end index

	if !t.isNumber(fromBal) {
		return shim.Error("from_bal is not a number")
	}

	if !t.isNumber(toBal) {
		return shim.Error("to_bal is not a number")
	}

	start, err := strconv.Atoi(startIndex)
	if err != nil {
		return shim.Error("start index is not a number")
	}

	end, err := strconv.Atoi(endIndex)
	if err != nil {
		return shim.Error("end index is not a number")
	}

	if start > end {
		return shim.Error("start index bigger than end index")
	}

	for i := start; i <= end; i++ {
		newAccFrom := accFrom + strconv.Itoa(i)
		newAccTo := accTo + strconv.Itoa(i)

		err = stub.PutStateFromKey(newAccFrom, fromBal)
		if err != nil {
			return shim.Error(fmt.Sprintf("putState(%s, %s) err: %+v", newAccFrom, fromBal, err))
		}

		err = stub.PutStateFromKey(newAccTo, toBal)
		if err != nil {
			return shim.Error(fmt.Sprintf("putState(%s, %s) err: %+v", newAccTo, toBal, err))
		}
	}

	return shim.Success([]byte("init success"))
}

func (t *TransferContract) transfer(stub shim.CMStubInterface) protogo.Response {
	args := stub.GetArgs()
	accFrom := string(args["acc_from"])
	accTo := string(args["acc_to"])
	amtTrans, err := strconv.Atoi(string(args["amt_trans"]))
	if err != nil {
		return shim.Error("amt_trans is not a number")
	}

	fromBalStr, err := stub.GetStateFromKey(accFrom)
	if err != nil {
		return shim.Error(fmt.Sprintf("getState(%s) error: %+v", accFrom, err))
	}

	fromBal, err := strconv.Atoi(fromBalStr)
	if err != nil {
		return shim.Error("from_bal is not a number")
	}

	toBalStr, err := stub.GetStateFromKey(accTo)
	if err != nil {
		return shim.Error(fmt.Sprintf("getState(%s) error: %+v", accTo, err))
	}
	toBal, err := strconv.Atoi(toBalStr)
	if err != nil {
		return shim.Error("to_bal is not a number")
	}
	if fromBal < amtTrans {
		return shim.Error(fmt.Sprintf("money doesn't enough, from_bal: %d, amt_trans: %d", fromBal, amtTrans))
	}
	fromBal -= amtTrans
	toBal += amtTrans
	err = stub.PutStateFromKey(accFrom, strconv.Itoa(fromBal))
	if err != nil {
		return shim.Error(fmt.Sprintf("putState(%s, %d) err: %+v", accFrom, fromBal, err))
	}
	err = stub.PutStateFromKey(accTo, strconv.Itoa(toBal))
	if err != nil {
		return shim.Error(fmt.Sprintf("putState(%s, %d) err: %+v", accTo, toBal, err))
	}
	return shim.Success([]byte("transfer success"))
}

func (t *TransferContract) isNumber(bal string) bool {
	_, err := strconv.Atoi(bal)
	if err != nil {
		return false
	}
	return true
}

func main() {
	err := shim.Start(new(TransferContract))
	if err != nil {
		log.Fatal(err)
	}
}
