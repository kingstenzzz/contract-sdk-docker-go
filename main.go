/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"log"
	"strconv"

	"chainmaker.org/chainmaker/chainmaker-contract-sdk-docker-go/pb/protogo"
	"chainmaker.org/chainmaker/chainmaker-contract-sdk-docker-go/shim"
)

type FactContract struct {
}

// 存证对象
type Fact struct {
	FileHash string `json:"FileHash"`
	FileName string `json:"FileName"`
	Time     int32  `json:"time"`
}

// 新建存证对象
func NewFact(FileHash string, FileName string, time int32) *Fact {
	fact := &Fact{
		FileHash: FileHash,
		FileName: FileName,
		Time:     time,
	}
	return fact
}

func (f *FactContract) InitContract(stub shim.CMStubInterface) protogo.Response {

	return shim.Success([]byte("Init Success"))

}

func (f *FactContract) InvokeContract(stub shim.CMStubInterface) protogo.Response {

	// 获取参数
	method := string(stub.GetArgs()["method"])

	switch method {
	case "save":
		return f.save(stub)
	case "findByFileHash":
		return f.findByFileHash(stub)
	default:
		return shim.Error("invalid method")
	}

}

func (f *FactContract) save(stub shim.CMStubInterface) protogo.Response {
	params := stub.GetArgs()

	// 获取参数
	fileHash := string(params["file_hash"])
	fileName := string(params["file_name"])
	timeStr := string(params["time"])
	time, err := strconv.Atoi(timeStr)
	if err != nil {
		msg := "time is [" + timeStr + "] not int"
		stub.Log(msg)
		return shim.Error(msg)
	}

	// 构建结构体
	fact := NewFact(fileHash, fileName, int32(time))

	// 序列化
	factBytes, _ := json.Marshal(fact)

	// 发送事件
	stub.EmitEvent("topic_vx", []string{fact.FileHash, fact.FileName})

	// 存储数据
	err = stub.PutStateByte("fact_bytes", fact.FileHash, factBytes)
	if err != nil {
		return shim.Error("fail to save fact bytes")
	}

	// 记录日志
	stub.Log("[save] FileHash=" + fact.FileHash)
	stub.Log("[save] FileName=" + fact.FileName)

	// 返回结果
	return shim.Success([]byte(fact.FileName + fact.FileHash))

}

func (f *FactContract) findByFileHash(stub shim.CMStubInterface) protogo.Response {
	// 获取参数
	FileHash := string(stub.GetArgs()["file_hash"])

	// 查询结果
	result, err := stub.GetStateByte("fact_bytes", FileHash)
	if err != nil {
		return shim.Error("failed to call get_state")
	}

	// 反序列化
	var fact Fact
	_ = json.Unmarshal(result, &fact)

	// 记录日志
	stub.Log("[find_by_file_hash] FileHash=" + fact.FileHash)
	stub.Log("[find_by_file_hash] FileName=" + fact.FileName)

	// 返回结果
	return shim.Success(result)
}

func main() {

	err := shim.Start(new(FactContract))
	if err != nil {
		log.Fatal(err)
	}
}
