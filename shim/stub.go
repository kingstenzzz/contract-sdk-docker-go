package shim

import (
	"chainmaker.org/chainmaker/protocol/v2"
	"errors"
	"fmt"
	"os"
	"strconv"

	"chainmaker.org/chainmaker-contract-sdk-docker-go/logger"
	"chainmaker.org/chainmaker-contract-sdk-docker-go/pb/protogo"
	"github.com/golang/protobuf/proto"
	"go.uber.org/zap"
)

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
	Logger.Debugf("get state for [%s#%s]", key, field)
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
	_ = s.Handler.SendGetStateReq([]byte(getStateKey), responseCh)

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

	callContractPayload, _ := proto.Marshal(callContractPayloadStruct)

	_ = s.Handler.SendCallContract(callContractPayload, responseCh)

	result := <-responseCh
	callContractResponsePayload := result.Payload

	var contractResponse protogo.ContractResponse
	_ = proto.Unmarshal(callContractResponsePayload, &contractResponse)

	// merge read write map
	for key, value := range contractResponse.ReadMap {
		s.readMap[key] = value
	}
	for key, value := range contractResponse.WriteMap {
		s.writeMap[key] = value
	}

	// merge events
	for _, event := range contractResponse.Events {
		s.events = append(s.events, event)
	}

	// return result
	return *contractResponse.Response
}

//func (s *CMStub) newIterator(startKey string, startField string, limitKey string, limitField string) (ResultSet, error) {
//	// construct request
//
//	// send request
//
//	// get error or index
//	index := 0
//	return &ResultSetKvImpl{index: int32(index)}, nil
//}
//
//func (s *CMStub) NewIterator(key string, limit string) (ResultSet, error) {
//	return s.newIterator(key, "", limit, "")
//}
//
//// ResultSetKvImpl implementation of ResultSet
//type ResultSetKvImpl struct { //为kv查询后的上下文
//	index int32 // 链的句柄的index
//}
//
//func (r *ResultSetKvImpl) HasNext() bool {
//	return false
//}
//
//func (r *ResultSetKvImpl) Close() (bool, error) {
//
//}
//
//func (r *ResultSetKvImpl) Next() (string, string, []byte, error) {
//
//}
