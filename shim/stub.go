/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package shim

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"chainmaker.org/chainmaker/chainmaker-contract-sdk-docker-go/logger"
	"chainmaker.org/chainmaker/chainmaker-contract-sdk-docker-go/pb/protogo"
	"chainmaker.org/chainmaker/common/v2/bytehelper"
	"chainmaker.org/chainmaker/common/v2/serialize"
	"chainmaker.org/chainmaker/protocol/v2"
	"github.com/golang/protobuf/proto"
	"go.uber.org/zap"
)

type ECKeyType = string
type Bool int32

const (
	MapSize = 8

	// special parameters passed to contract

	ContractParamCreatorOrgId = "__creator_org_id__"
	ContractParamCreatorRole  = "__creator_role__"
	ContractParamCreatorPk    = "__creator_pk__"
	ContractParamSenderOrgId  = "__sender_org_id__"
	ContractParamSenderRole   = "__sender_role__"
	ContractParamSenderPk     = "__sender_pk__"
	ContractParamBlockHeight  = "__block_height__"
	ContractParamTxId         = "__tx_id__"
	ContractParamTxTimeStamp  = "__tx_time_stamp__"

	// common easyCodec key

	EC_KEY_TYPE_KEY          ECKeyType = "key"
	EC_KEY_TYPE_FIELD        ECKeyType = "field"
	EC_KEY_TYPE_VALUE        ECKeyType = "value"
	EC_KEY_TYPE_TX_ID        ECKeyType = "currentTxId"
	EC_KEY_TYPE_BLOCK_HEITHT ECKeyType = "blockHeight"
	EC_KEY_TYPE_IS_DELETE    ECKeyType = "isDelete"
	EC_KEY_TYPE_TIMESTAMP    ECKeyType = "timestamp"

	// stateKvIterator method

	FuncKvIteratorCreate    = "createKvIterator"
	FuncKvPreIteratorCreate = "createKvPreIterator"
	FuncKvIteratorHasNext   = "kvIteratorHasNext"
	FuncKvIteratorNext      = "kvIteratorNext"
	FuncKvIteratorClose     = "kvIteratorClose"

	// keyHistoryKvIterator method

	FuncKeyHistoryIterHasNext = "keyHistoryIterHasNext"
	FuncKeyHistoryIterNext    = "keyHistoryIterNext"
	FuncKeyHistoryIterClose   = "keyHistoryIterClose"

	// int32 representation of bool

	BoolTrue  Bool = 1
	BoolFalse Bool = 0
)

type CMStub struct {
	args    map[string][]byte
	Handler *Handler

	// cache
	readMap  map[string][]byte
	writeMap map[string][]byte
	// contract parameters
	creatorOrgId string
	creatorRole  string
	creatorPk    string
	senderOrgId  string
	senderRole   string
	senderPk     string
	blockHeight  string
	txId         string
	txTimeStamp  string
	// events
	contractName    string
	contractVersion string
	events          []*protogo.Event
	// logger
	logger *zap.SugaredLogger
}

func initStubContractParam(args map[string][]byte, key string) string {
	if value, ok := args[key]; ok {
		delete(args, key)
		return string(value)
	} else {
		Logger.Errorf("init contract parameter [%v] failed", key)
		return ""
	}
}

func NewCMStub(handler *Handler, args map[string][]byte, contractName, contractVersion string) *CMStub {

	logLevel := os.Args[4]
	var events []*protogo.Event

	stub := &CMStub{
		args:            args,
		Handler:         handler,
		readMap:         make(map[string][]byte, MapSize),
		writeMap:        make(map[string][]byte, MapSize),
		creatorOrgId:    initStubContractParam(args, ContractParamCreatorOrgId),
		creatorRole:     initStubContractParam(args, ContractParamCreatorRole),
		creatorPk:       initStubContractParam(args, ContractParamCreatorPk),
		senderOrgId:     initStubContractParam(args, ContractParamSenderOrgId),
		senderRole:      initStubContractParam(args, ContractParamSenderRole),
		senderPk:        initStubContractParam(args, ContractParamSenderPk),
		blockHeight:     initStubContractParam(args, ContractParamBlockHeight),
		txId:            initStubContractParam(args, ContractParamTxId),
		txTimeStamp:     initStubContractParam(args, ContractParamTxTimeStamp),
		logger:          logger.NewDockerLogger("[Contract]", logLevel),
		events:          events,
		contractName:    contractName,
		contractVersion: contractVersion,
	}

	return stub
}

func (s *CMStub) GetArgs() map[string][]byte {
	return s.args
}

func (s *CMStub) GetState(key, field string) (string, error) {
	Logger.Debugf("[%s] get state for key: %s, field: %s", s.Handler.currentTxId, key, field)
	// get from write set
	if value, done := s.getFromWriteSet(key, field); done {
		s.putIntoReadSet(key, field, value)
		return string(value), nil
	}

	// get from read set
	if value, done := s.getFromReadSet(key, field); done {
		return string(value), nil
	}

	// get from chain maker
	value, err := s.getState(key, field)
	if err != nil {
		return "", err
	}
	Logger.Debugf("[%s] get state finished for key: %s, field: %s, value: %s", s.Handler.currentTxId, key,
		field, string(value))
	return string(value), nil
}

func (s *CMStub) GetStateByte(key, field string) ([]byte, error) {
	Logger.Debugf("get state for [%s#%s]", key, field)
	// get from write set
	if value, done := s.getFromWriteSet(key, field); done {
		s.putIntoReadSet(key, field, value)
		return value, nil
	}

	// get from read set
	if value, done := s.getFromReadSet(key, field); done {
		return value, nil
	}

	// get from chain maker
	return s.getState(key, field)
}

func (s *CMStub) GetStateFromKey(key string) (string, error) {
	Logger.Debugf("get state for [%s#%s]", key, "")
	// get from write set
	if value, done := s.getFromWriteSet(key, ""); done {
		s.putIntoReadSet(key, "", value)
		return string(value), nil
	}

	// get from read set
	if value, done := s.getFromReadSet(key, ""); done {
		return string(value), nil
	}

	// get from chain maker
	value, err := s.getState(key, "")
	if err != nil {
		return "", err
	}
	return string(value), nil
}

func (s *CMStub) GetStateFromKeyByte(key string) ([]byte, error) {
	Logger.Debugf("get state for [%s#%s]", key, "")
	// get from write set
	if value, done := s.getFromWriteSet(key, ""); done {
		s.putIntoReadSet(key, "", value)
		return value, nil
	}

	// get from read set
	if value, done := s.getFromReadSet(key, ""); done {
		return value, nil
	}

	// get from chain maker
	return s.getState(key, "")
}

func (s *CMStub) getState(key, field string) ([]byte, error) {
	responseCh := make(chan *protogo.DMSMessage, 1)

	getStateKey := s.constructKey(key, field)
	err := s.Handler.SendGetStateReq([]byte(getStateKey), responseCh)
	if err != nil {
		return nil, err
	}

	result := <-responseCh

	if result.ResultCode == protocol.ContractSdkSignalResultFail {
		return nil, errors.New(result.Message)
	}

	value := result.Payload
	s.putIntoReadSet(key, field, value)
	return value, nil
}

func (s *CMStub) PutState(key, field string, value string) error {
	s.putIntoWriteSet(key, field, []byte(value))
	return nil
}

func (s *CMStub) PutStateByte(key, field string, value []byte) error {
	s.putIntoWriteSet(key, field, value)
	return nil
}

func (s *CMStub) PutStateFromKey(key string, value string) error {
	s.putIntoWriteSet(key, "", []byte(value))
	return nil
}

func (s *CMStub) PutStateFromKeyByte(key string, value []byte) error {
	s.putIntoWriteSet(key, "", value)
	return nil
}

func (s *CMStub) DelState(key, field string) error {
	s.putIntoWriteSet(key, field, nil)
	return nil
}

func (s *CMStub) DelStateFromKey(key string) error {
	s.putIntoWriteSet(key, "", nil)
	return nil
}

func (s *CMStub) getFromWriteSet(key, field string) ([]byte, bool) {
	contractKey := s.constructKey(key, field)
	Logger.Debugf("get key[%s] from write set\n", contractKey)
	if txWrite, ok := s.writeMap[contractKey]; ok {
		return txWrite, true
	}
	return nil, false
}

func (s *CMStub) getFromReadSet(key, field string) ([]byte, bool) {
	contractKey := s.constructKey(key, field)
	Logger.Debugf("get key[%s] from read set\n", contractKey)
	if txRead, ok := s.readMap[contractKey]; ok {
		return txRead, true
	}
	return nil, false
}

func (s *CMStub) putIntoWriteSet(key, field string, value []byte) {
	contractKey := s.constructKey(key, field)
	s.writeMap[contractKey] = value
	Logger.Debugf("put key[%s] - value[%s] into write set\n", contractKey, string(value))
}

func (s *CMStub) putIntoReadSet(key, field string, value []byte) {
	contractKey := s.constructKey(key, field)
	s.readMap[contractKey] = value
	Logger.Debugf("put key[%s] - value[%s] into read set\n", string(key), string(value))
}

func (s *CMStub) constructKey(key, field string) string {
	if len(field) == 0 {
		return s.Handler.contractName + "#" + key
	}
	return s.Handler.contractName + "#" + key + "#" + field
}

func (s *CMStub) GetWriteMap() map[string][]byte {
	return s.writeMap
}

func (s *CMStub) GetReadMap() map[string][]byte {
	return s.readMap
}

func (s *CMStub) GetCreatorOrgId() (string, error) {
	if len(s.creatorOrgId) == 0 {
		return s.creatorOrgId, fmt.Errorf("can not get creator org id")
	} else {
		return s.creatorOrgId, nil
	}
}

func (s *CMStub) GetCreatorRole() (string, error) {
	if len(s.creatorRole) == 0 {
		return s.creatorRole, fmt.Errorf("can not get creator role")
	} else {
		return s.creatorRole, nil
	}
}

func (s *CMStub) GetCreatorPk() (string, error) {
	if len(s.creatorPk) == 0 {
		return s.creatorPk, fmt.Errorf("can not get creator pk")
	} else {
		return s.creatorPk, nil
	}
}

func (s *CMStub) GetSenderOrgId() (string, error) {
	if len(s.senderOrgId) == 0 {
		return s.senderOrgId, fmt.Errorf("can not get sender org id")
	} else {
		return s.senderOrgId, nil
	}
}

func (s *CMStub) GetSenderRole() (string, error) {
	if len(s.senderRole) == 0 {
		return s.senderRole, fmt.Errorf("can not get sender role")
	} else {
		return s.senderRole, nil
	}
}

func (s *CMStub) GetSenderPk() (string, error) {
	if len(s.senderPk) == 0 {
		return s.senderPk, fmt.Errorf("can not get sender pk")
	} else {
		return s.senderPk, nil
	}
}

func (s *CMStub) GetBlockHeight() (int, error) {
	if len(s.blockHeight) == 0 {
		return 0, fmt.Errorf("can not get block height")
	}
	if res, err := strconv.Atoi(s.blockHeight); err != nil {
		return 0, fmt.Errorf("block height [%v] can not convert to type int", s.blockHeight)
	} else {
		return res, nil
	}
}

func (s *CMStub) GetTxId() (string, error) {
	if len(s.txId) == 0 {
		return s.txId, fmt.Errorf("can not get tx id")
	} else {
		return s.txId, nil
	}
}

func (s *CMStub) GetTxTimeStamp() (string, error) {
	if len(s.txTimeStamp) == 0 {
		return s.txTimeStamp, fmt.Errorf("can not get tx timestamp")
	}

	return s.txTimeStamp, nil
}

func (s *CMStub) EmitEvent(topic string, data []string) {
	newEvent := &protogo.Event{
		Topic:           topic,
		ContractName:    s.contractName,
		ContractVersion: s.contractVersion,
		Data:            data,
	}
	s.events = append(s.events, newEvent)
}

func (s *CMStub) GetEvents() []*protogo.Event {
	return s.events
}

func (s *CMStub) Log(message string) {
	s.logger.Debugf(message)
}

func (s *CMStub) CallContract(contractName, contractVersion string, args map[string][]byte) protogo.Response {
	Logger.Debugf("[%s] call contract start, called contract name: %s", s.Handler.currentTxId, contractName)
	defer Logger.Debugf("[%s] call contract finished, called contract name: %s", s.Handler.currentTxId,
		contractName)
	// get contract result from docker manager
	responseCh := make(chan *protogo.DMSMessage, 1)

	initialArgs := map[string][]byte{
		ContractParamCreatorOrgId: []byte(s.creatorOrgId),
		ContractParamCreatorRole:  []byte(s.creatorRole),
		ContractParamCreatorPk:    []byte(s.creatorPk),
		ContractParamSenderOrgId:  []byte(s.senderOrgId),
		ContractParamSenderRole:   []byte(s.senderRole),
		ContractParamSenderPk:     []byte(s.senderPk),
		ContractParamBlockHeight:  []byte(s.blockHeight),
		ContractParamTxId:         []byte(s.txId),
		ContractParamTxTimeStamp:  []byte(s.txTimeStamp),
	}

	// add user defined args
	for key, value := range args {
		initialArgs[key] = value
	}

	callContractPayloadStruct := &protogo.CallContractRequest{
		ContractName:    contractName,
		ContractVersion: contractVersion,
		Args:            initialArgs,
	}

	constructErrorCallContractResponse := func(err error) protogo.Response {
		return protogo.Response{
			Status:  1,
			Message: err.Error(),
			Payload: nil,
		}
	}

	callContractPayload, err := proto.Marshal(callContractPayloadStruct)
	if err != nil {
		return constructErrorCallContractResponse(err)
	}

	err = s.Handler.SendCallContract(callContractPayload, responseCh)
	if err != nil {
		return constructErrorCallContractResponse(err)
	}

	result := <-responseCh

	var contractResponse protogo.ContractResponse
	err = proto.Unmarshal(result.Payload, &contractResponse)
	if err != nil {
		return constructErrorCallContractResponse(err)
	}

	if contractResponse.Response.Status != OK {
		return *contractResponse.Response
	}

	// merge read write map
	for key, value := range contractResponse.ReadMap {
		s.readMap[key] = value
	}
	for key, value := range contractResponse.WriteMap {
		s.writeMap[key] = value
	}

	// merge events
	s.events = append(s.events, contractResponse.Events...)

	// return result
	return *contractResponse.Response
}

func (s *CMStub) NewIterator(startKey string, limitKey string) (ResultSetKV, error) {
	return s.newIterator(FuncKvIteratorCreate, startKey, "", limitKey, "")
}

func (s *CMStub) NewIteratorWithField(key string, startField string, limitField string) (ResultSetKV, error) {
	return s.newIterator(FuncKvIteratorCreate, key, startField, key, limitField)
}

func (s *CMStub) NewIteratorPrefixWithKeyField(startKey string, startField string) (ResultSetKV, error) {
	return s.newIterator(FuncKvPreIteratorCreate, startKey, startField, "", "")
}

func (s *CMStub) NewIteratorPrefixWithKey(key string) (ResultSetKV, error) {
	return s.NewIteratorPrefixWithKeyField(key, "")
}

func (s *CMStub) newIterator(iteratorFuncName, startKey string, startField string, limitKey string,
	limitField string) (
	ResultSetKV, error) {

	responseCh := make(chan *protogo.DMSMessage, 1)
	writeMap := s.GetWriteMap()
	wMBytes, err := json.Marshal(writeMap)
	if err != nil {
		return nil, err
	}

	createKvIteratorKey := func() []byte {
		str :=
			s.Handler.contractName + "#" +
				iteratorFuncName + "#" +
				startKey + "#" +
				startField + "#" +
				limitKey + "#" +
				limitField + "#" +
				string(wMBytes)
		return []byte(str)
	}()

	// reset writeMap
	s.writeMap = make(map[string][]byte, MapSize)

	err = s.Handler.SendCreateKvIteratorReq(createKvIteratorKey, responseCh)
	if err != nil {
		return nil, err
	}

	result := <-responseCh

	if result.ResultCode == protocol.ContractSdkSignalResultFail {
		return nil, errors.New(result.Message)
	}

	value := result.Payload

	index, err := bytehelper.BytesToInt(value)
	if err != nil {
		return nil, fmt.Errorf("get iterator index failed, %s", err.Error())
	}

	return &ResultSetKvImpl{s: s, index: index}, nil

}

func (s *CMStub) NewHistoryKvIterForKey(key, field string) (KeyHistoryKvIter, error) {
	responseCh := make(chan *protogo.DMSMessage, 1)
	writeMap := s.GetWriteMap()
	wMapBytes, err := json.Marshal(writeMap)
	if err != nil {
		return nil, err
	}

	createHistoryKvIterKey := func() []byte {
		str :=
			s.Handler.contractName + "#" +
				key + "#" +
				field + "#" +
				string(wMapBytes)
		return []byte(str)
	}()

	// reset writeMap
	s.writeMap = make(map[string][]byte, MapSize)

	err = s.Handler.SendCreateKeyHistoryKvIterReq(createHistoryKvIterKey, responseCh)
	if err != nil {
		return nil, err
	}

	result := <-responseCh

	if result.ResultCode == protocol.ContractSdkSignalResultFail {
		return nil, errors.New(result.Message)
	}

	value := result.Payload

	index, err := bytehelper.BytesToInt(value)
	if err != nil {
		return nil, fmt.Errorf("get history iterator index failed, %s", err.Error())
	}

	return &KeyHistoryKvIterImpl{
		key:   key,
		field: field,
		index: index,
		s:     s,
	}, nil
}

func (s *CMStub) GetSenderAddr() (string, error) {
	responseCh := make(chan *protogo.DMSMessage, 1)

	err := s.Handler.SendGetSenderAddrReq(nil, responseCh)
	if err != nil {
		return "", err
	}

	result := <-responseCh

	if result.ResultCode == protocol.ContractSdkSignalResultFail {
		return "", errors.New(result.Message)
	}

	return string(result.Payload), nil
}

// ResultSetKvImpl iterator query result KVdb
type ResultSetKvImpl struct {
	s *CMStub

	index int32
}

func (r *ResultSetKvImpl) HasNext() bool {

	responseCh := make(chan *protogo.DMSMessage, 1)

	hasNextKey := func() []byte {
		str := FuncKvIteratorHasNext + "#" + string(bytehelper.IntToBytes(r.index))
		return []byte(str)
	}()

	err := r.s.Handler.SendConsumeKvIteratorReq(hasNextKey, responseCh)
	if err != nil {
		return false
	}

	result := <-responseCh

	if result.ResultCode == protocol.ContractSdkSignalResultFail {
		return false
	}

	has, err := bytehelper.BytesToInt(result.Payload)
	if err != nil {
		return false
	}

	if has == 0 {
		return false
	}

	return true
}

func (r *ResultSetKvImpl) Close() (bool, error) {
	responseCh := make(chan *protogo.DMSMessage, 1)

	closeKey := func() []byte {
		str := FuncKvIteratorClose + "#" + string(bytehelper.IntToBytes(r.index))
		return []byte(str)
	}()

	err := r.s.Handler.SendConsumeKvIteratorReq(closeKey, responseCh)
	if err != nil {
		return false, err
	}

	result := <-responseCh

	if result.ResultCode == protocol.ContractSdkSignalResultFail {
		return false, errors.New(result.Message)
	} else if result.ResultCode == protocol.ContractSdkSignalResultSuccess {
		return true, nil
	}

	return true, nil
}

func (r *ResultSetKvImpl) NextRow() (*serialize.EasyCodec, error) {
	responseCh := make(chan *protogo.DMSMessage, 1)

	nextRowKey := func() []byte {
		str := FuncKvIteratorNext + "#" + string(bytehelper.IntToBytes(r.index))
		return []byte(str)
	}()

	err := r.s.Handler.SendConsumeKvIteratorReq(nextRowKey, responseCh)
	if err != nil {
		return nil, err
	}

	result := <-responseCh

	if result.ResultCode == protocol.ContractSdkSignalResultFail {
		return nil, errors.New(result.Message)
	}

	payload := strings.Split(string(result.Payload), "#")
	key := payload[0]
	field := payload[1]
	value := payload[2]

	ec := serialize.NewEasyCodec()
	ec.AddString(EC_KEY_TYPE_KEY, key)
	ec.AddString(EC_KEY_TYPE_FIELD, field)
	ec.AddBytes(EC_KEY_TYPE_VALUE, []byte(value))

	return ec, nil
}

func (r *ResultSetKvImpl) Next() (string, string, []byte, error) {
	ec, err := r.NextRow()
	if err != nil {
		return "", "", nil, err
	}
	key, _ := ec.GetString(EC_KEY_TYPE_KEY)
	field, _ := ec.GetString(EC_KEY_TYPE_FIELD)
	v, _ := ec.GetBytes(EC_KEY_TYPE_VALUE)

	return key, field, v, nil
}

type KeyHistoryKvIterImpl struct {
	s *CMStub

	key   string
	field string
	index int32
}

func (k *KeyHistoryKvIterImpl) HasNext() bool {

	responseCh := make(chan *protogo.DMSMessage, 1)

	hasNextKey := func() []byte {
		str := FuncKeyHistoryIterHasNext + "#" + string(bytehelper.IntToBytes(k.index))
		return []byte(str)
	}()

	err := k.s.Handler.SendConsumeKeyHistoryKvIterReq(hasNextKey, responseCh)
	if err != nil {
		return false
	}

	result := <-responseCh

	if result.ResultCode == protocol.ContractSdkSignalResultFail {
		return false
	}

	has, err := bytehelper.BytesToInt(result.Payload)
	if err != nil {
		return false
	}

	if Bool(has) == BoolFalse {
		return false
	}

	return true
}

func (k *KeyHistoryKvIterImpl) NextRow() (*serialize.EasyCodec, error) {
	responseCh := make(chan *protogo.DMSMessage, 1)

	nextRowKey := func() []byte {
		str := FuncKeyHistoryIterNext + "#" + string(bytehelper.IntToBytes(k.index))
		return []byte(str)
	}()

	err := k.s.Handler.SendConsumeKeyHistoryKvIterReq(nextRowKey, responseCh)
	if err != nil {
		return nil, err
	}

	result := <-responseCh

	if result.ResultCode == protocol.ContractSdkSignalResultFail {
		return nil, errors.New(result.Message)
	}

	/*
		| index | desc        |
		| ---   | ---         |
		| 0     | currentTxId        |
		| 1     | blockHeight |
		| 2     | value       |
		| 3     | isDelete    |
		| 4     | timestamp   |
	*/
	payload := strings.Split(string(result.Payload), "#")
	txId := payload[0]
	blockHeightStr := payload[1]
	value := payload[2]
	isDeleteStr := payload[3]
	timestamp := payload[4]

	ec := serialize.NewEasyCodec()
	ec.AddBytes(EC_KEY_TYPE_VALUE, []byte(value))
	ec.AddString(EC_KEY_TYPE_TX_ID, txId)

	blockHeight, err := bytehelper.BytesToInt([]byte(blockHeightStr))
	if err != nil {
		return nil, err
	}
	ec.AddInt32(EC_KEY_TYPE_BLOCK_HEITHT, blockHeight)

	ec.AddString(EC_KEY_TYPE_TIMESTAMP, timestamp)

	isDelete, err := bytehelper.BytesToInt([]byte(isDeleteStr))
	if err != nil {
		return nil, err
	}
	ec.AddInt32(EC_KEY_TYPE_IS_DELETE, isDelete)

	ec.AddString(EC_KEY_TYPE_KEY, k.key)
	ec.AddString(EC_KEY_TYPE_FIELD, k.field)

	return ec, nil
}

func (k *KeyHistoryKvIterImpl) Close() (bool, error) {
	responseCh := make(chan *protogo.DMSMessage, 1)

	closeKey := func() []byte {
		str := FuncKeyHistoryIterClose + "#" + string(bytehelper.IntToBytes(k.index))
		return []byte(str)
	}()

	err := k.s.Handler.SendConsumeKeyHistoryKvIterReq(closeKey, responseCh)
	if err != nil {
		return false, err
	}

	result := <-responseCh

	if result.ResultCode == protocol.ContractSdkSignalResultFail {
		return false, errors.New(result.Message)
	} else if result.ResultCode == protocol.ContractSdkSignalResultSuccess {
		return true, nil
	}

	return true, nil
}

func (k *KeyHistoryKvIterImpl) Next() (*KeyModification, error) {
	ec, err := k.NextRow()
	if err != nil {
		return nil, err
	}

	value, _ := ec.GetBytes(EC_KEY_TYPE_VALUE)
	txId, _ := ec.GetString(EC_KEY_TYPE_TX_ID)
	blockHeight, _ := ec.GetInt32(EC_KEY_TYPE_BLOCK_HEITHT)
	isDeleteBool, _ := ec.GetInt32(EC_KEY_TYPE_IS_DELETE)
	isDelete := false
	if Bool(isDeleteBool) == BoolTrue {
		isDelete = true
	}

	timestamp, _ := ec.GetString(EC_KEY_TYPE_TIMESTAMP)

	return &KeyModification{
		Key:         k.key,
		Field:       k.field,
		Value:       value,
		TxId:        txId,
		BlockHeight: int(blockHeight),
		IsDelete:    isDelete,
		Timestamp:   timestamp,
	}, nil
}

type KeyModification struct {
	Key         string
	Field       string
	Value       []byte
	TxId        string
	BlockHeight int
	IsDelete    bool
	Timestamp   string
}
