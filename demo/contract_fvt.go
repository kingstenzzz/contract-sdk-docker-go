package demo

import (
	"fmt"
	"log"
	"time"

	"chainmaker.org/chainmaker/common/v2/bytehelper"

	"chainmaker.org/chainmaker-contract-sdk-docker-go/pb/protogo"
	"chainmaker.org/chainmaker-contract-sdk-docker-go/shim"
)

type TestContract struct {
}

func (t *TestContract) InitContract(stub shim.CMStubInterface) protogo.Response {

	return shim.Success([]byte("Init Success"))
}

func (t *TestContract) InvokeContract(stub shim.CMStubInterface) protogo.Response {

	args := stub.GetArgs()

	method := string(args["method"])
	switch method {
	case "display":
		return t.display()
	case "put_state":
		return t.putState(stub)
	case "put_state_byte":
		return t.putStateByte(stub)
	case "put_state_from_key":
		return t.putStateFromKey(stub)
	case "put_state_from_key_byte":
		return t.putStateFromKeyByte(stub)
	case "get_state":
		return t.getState(stub)
	case "get_state_byte":
		return t.getStateByte(stub)
	case "get_state_from_key":
		return t.getStateFromKey(stub)
	case "get_state_from_key_byte":
		return t.getStateFromKeyByte(stub)
	case "del_state":
		return t.delState(stub)
	case "time_out":
		return t.timeOut(stub)
	case "out_of_range":
		return t.outOfRange()
	case "cross_contract":
		return t.crossContract(stub)

		// KVIterator
		// create
	case "new_iterator":
		return t.newIterator(stub)
	case "new_iterator_with_field":
		return t.newIteratorWithField(stub)
	case "new_iterator_prefix_with_key":
		return t.newIteratorPrefixWithKey(stub)
	case "new_iterator_prefix_with_key_field":
		return t.newIteratorPrefixWithKeyField(stub)
	// consume
	case "iterator_has_next":
		return t.iteratorHasNext(stub)
	case "iterator_next":
		return t.iteratorNext(stub)
	case "iterator_next_row":
		return t.iteratorNextRow(stub)
	case "iterator_release":
		return t.iteratorRelease(stub)

	default:
		return shim.Error("unknown method")
	}
}

func (t *TestContract) display() protogo.Response {
	return shim.Success([]byte("display successful"))
}

func (t *TestContract) putState(stub shim.CMStubInterface) protogo.Response {
	args := stub.GetArgs()

	getKey := string(args["key"])
	getField := string(args["field"])
	getValue := string(args["value"])

	err := stub.PutState(getKey, getField, getValue)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte("put state successfully"))
}

func (t *TestContract) putStateByte(stub shim.CMStubInterface) protogo.Response {
	args := stub.GetArgs()

	getKey := string(args["key"])
	getField := string(args["field"])
	getValue := args["value"]

	err := stub.PutStateByte(getKey, getField, getValue)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte("put state successfully"))
}

func (t *TestContract) putStateFromKey(stub shim.CMStubInterface) protogo.Response {
	args := stub.GetArgs()

	getKey := string(args["key"])
	getValue := string(args["value"])

	err := stub.PutStateFromKey(getKey, getValue)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte("put state successfully"))
}

func (t *TestContract) putStateFromKeyByte(stub shim.CMStubInterface) protogo.Response {
	args := stub.GetArgs()

	getKey := string(args["key"])
	getValue := args["value"]

	err := stub.PutStateFromKeyByte(getKey, getValue)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte("put state successfully"))
}

func (t *TestContract) getState(stub shim.CMStubInterface) protogo.Response {

	args := stub.GetArgs()

	getKey := string(args["key"])
	field := string(args["field"])

	result, err := stub.GetState(getKey, field)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(result))
}

func (t *TestContract) getStateByte(stub shim.CMStubInterface) protogo.Response {

	args := stub.GetArgs()

	getKey := string(args["key"])
	field := string(args["field"])

	result, err := stub.GetStateByte(getKey, field)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(result)
}

func (t *TestContract) getStateFromKey(stub shim.CMStubInterface) protogo.Response {

	args := stub.GetArgs()

	getKey := string(args["key"])

	result, err := stub.GetStateFromKey(getKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(result))
}

func (t *TestContract) getStateFromKeyByte(stub shim.CMStubInterface) protogo.Response {
	args := stub.GetArgs()

	getKey := string(args["key"])

	result, err := stub.GetStateFromKeyByte(getKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(result)
}

func (t *TestContract) delState(stub shim.CMStubInterface) protogo.Response {
	args := stub.GetArgs()

	getKey := string(args["key"])
	getField := string(args["field"])

	err := stub.DelState(getKey, getField)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte("delete successfully"))
}

func (t *TestContract) timeOut(stub shim.CMStubInterface) protogo.Response {
	time.Sleep(5 * time.Second)
	return shim.Success([]byte("success finish timeout"))
}

func (t *TestContract) outOfRange() protogo.Response {
	var group []string
	group[0] = "abc"
	fmt.Println(group[0])
	return shim.Success([]byte("exit out of range"))
}

func (t *TestContract) crossContract(stub shim.CMStubInterface) protogo.Response {

	args := stub.GetArgs()

	contractName := string(args["contract_name"])
	contractVersion := string(args["contract_version"])

	calledMethod := string(args["contract_method"])

	crossContractArgs := make(map[string][]byte)
	crossContractArgs["method"] = []byte(calledMethod)

	// response could be correct or error
	response := stub.CallContract(contractName, contractVersion, crossContractArgs)
	stub.EmitEvent("cross contract", []string{"success"})
	return response
}

// create iterator
func (t *TestContract) newIterator(stub shim.CMStubInterface) protogo.Response {
	paramMap := stub.GetArgs()
	key := paramMap["key"]
	limit := paramMap["limit"]
	iterator, err := stub.NewIterator(string(key), string(limit))
	if err != nil {
		msg := "NewIterator failed"
		stub.Log(msg)
		return shim.Error(msg)
	}

	kviImpl, _ := iterator.(*shim.ResultSetKvImpl)

	// 返回缓存迭代器的索引
	return shim.Success(bytehelper.IntToBytes(kviImpl.GetIteratorCacheIndex()))
}

func (t *TestContract) newIteratorWithField(stub shim.CMStubInterface) protogo.Response {
	paramMap := stub.GetArgs()
	key := paramMap["key"]
	field := paramMap["field"]
	limit := paramMap["limit"]
	iterator, err := stub.NewIteratorWithField(string(key), string(field), string(limit))
	if err != nil {
		msg := "NewIteratorWithField failed"
		stub.Log(msg)
		return shim.Error(msg)
	}

	kviImpl, _ := iterator.(*shim.ResultSetKvImpl)

	// 返回缓存迭代器的索引
	return shim.Success(bytehelper.IntToBytes(kviImpl.GetIteratorCacheIndex()))
}

func (t *TestContract) newIteratorPrefixWithKey(stub shim.CMStubInterface) protogo.Response {
	paramMap := stub.GetArgs()
	key := paramMap["key"]
	iterator, err := stub.NewIteratorPrefixWithKey(string(key))
	if err != nil {
		msg := "NewIteratorPrefixWithKey failed"
		stub.Log(msg)
		return shim.Error(msg)
	}

	kviImpl, _ := iterator.(*shim.ResultSetKvImpl)

	// 返回缓存迭代器的索引
	return shim.Success(bytehelper.IntToBytes(kviImpl.GetIteratorCacheIndex()))
}

func (t *TestContract) newIteratorPrefixWithKeyField(stub shim.CMStubInterface) protogo.Response {
	paramMap := stub.GetArgs()
	key := paramMap["key"]
	field := paramMap["field"]
	iterator, err := stub.NewIteratorPrefixWithKeyField(string(key), string(field))
	if err != nil {
		msg := "NewIteratorPrefixWithKey failed"
		stub.Log(msg)
		return shim.Error(msg)
	}

	kviImpl, _ := iterator.(*shim.ResultSetKvImpl)

	// 返回缓存迭代器的索引
	return shim.Success(bytehelper.IntToBytes(kviImpl.GetIteratorCacheIndex()))
}

// consume iterator
func (t *TestContract) iteratorHasNext(stub shim.CMStubInterface) protogo.Response {
	paramMap := stub.GetArgs()
	indexBytes := paramMap["index"]
	index, _ := bytehelper.BytesToInt(indexBytes)
	kvi := new(shim.ResultSetKvImpl)
	kvi.SetIteratorCacheIndex(index)
	s, _ := stub.(*shim.CMStub)
	kvi.SetIteratorCacheStub(s)

	if kvi.HasNext() {
		return shim.Success(bytehelper.IntToBytes(1))
	}

	return shim.Success(bytehelper.IntToBytes(0))
}

func (t *TestContract) iteratorNext(stub shim.CMStubInterface) protogo.Response {
	paramMap := stub.GetArgs()
	indexBytes := paramMap["index"]
	index, _ := bytehelper.BytesToInt(indexBytes)
	kvi := new(shim.ResultSetKvImpl)
	kvi.SetIteratorCacheIndex(index)
	s, _ := stub.(*shim.CMStub)
	kvi.SetIteratorCacheStub(s)

	key, field, value, err := kvi.Next()
	if err != nil {
		msg := "iterator failed to get the next element"
		stub.Log(msg)
		return shim.Error(msg)
	}

	return shim.Success([]byte(key + "#" + field + "#" + string(value)))
}

func (t *TestContract) iteratorNextRow(stub shim.CMStubInterface) protogo.Response {
	paramMap := stub.GetArgs()
	indexBytes := paramMap["index"]
	index, _ := bytehelper.BytesToInt(indexBytes)
	kvi := new(shim.ResultSetKvImpl)
	kvi.SetIteratorCacheIndex(index)
	s, _ := stub.(*shim.CMStub)
	kvi.SetIteratorCacheStub(s)

	row, err := kvi.NextRow()
	if err != nil {
		msg := "iterator failed to get the next element"
		stub.Log(msg)
		return shim.Error(msg)
	}

	return shim.Success(row.Marshal())
}

func (t *TestContract) iteratorRelease(stub shim.CMStubInterface) protogo.Response {
	paramMap := stub.GetArgs()
	indexBytes := paramMap["index"]
	index, _ := bytehelper.BytesToInt(indexBytes)
	kvi := new(shim.ResultSetKvImpl)
	kvi.SetIteratorCacheIndex(index)
	s, _ := stub.(*shim.CMStub)
	kvi.SetIteratorCacheStub(s)

	ok, err := kvi.Close()
	if err != nil || !ok {
		msg := "close failed"
		stub.Log(msg)
		return shim.Error(msg)
	}

	return shim.Success([]byte("SUCCESS"))
}

func main() {
	err := shim.Start(new(TestContract))
	if err != nil {
		log.Fatal(err)
	}
}
