package shim

import (
	"chainmaker.org/chainmaker-contract-sdk-docker-go/pb/protogo"
	"fmt"
)

const (
	MapSize = 8
)

type CMStub struct {
	args         map[string]string
	Handler      *Handler
	contractName string

	// simContext
	readMap  map[string]*protogo.TxRead
	writeMap map[string]*protogo.TxWrite
}

func NewCMStub(handler *Handler, args map[string]string, contractName string) *CMStub {

	stub := &CMStub{
		args:         args,
		Handler:      handler,
		contractName: contractName,
		readMap:      make(map[string]*protogo.TxRead, MapSize),
		writeMap:     make(map[string]*protogo.TxWrite, MapSize),
	}

	return stub
}

func (s *CMStub) GetArgs() map[string]string {
	return s.args
}

func (s *CMStub) GetState(key []byte) ([]byte, error) {

	// get from write set
	if value, done := s.getFromWriteSet(key); done {
		s.putIntoWriteSet(key, value)
		return value, nil
	}

	// get from read set
	if value, done := s.getFromReadSet(key); done {
		return value, nil
	}

	// get from chain maker
	responseCh := make(chan []byte, 1)

	getStateKey := s.constructKey(s.contractName, key)
	_ = s.Handler.SendGetStateReq(getStateKey, responseCh)

	value := <-responseCh
	if len(value) > 0 {
		s.putIntoReadSet(key, value)
		return value, nil
	}

	return nil, fmt.Errorf("fail to get value from chainmaker for [%s]", string(key))
}

func (s *CMStub) PutState(key []byte, value []byte) error {
	s.putIntoWriteSet(key, value)
	return nil
}

func (s *CMStub) DelState(key []byte) error {
	s.putIntoWriteSet(key, nil)
	return nil
}

func (s *CMStub) putIntoWriteSet(key []byte, value []byte) {
	s.writeMap[s.constructKey(s.contractName, key)] = &protogo.TxWrite{
		Key:          key,
		Value:        value,
		ContractName: s.contractName,
	}
}

func (s *CMStub) getFromWriteSet(key []byte) ([]byte, bool) {
	if txWrite, ok := s.writeMap[s.constructKey(s.contractName, key)]; ok {
		return txWrite.Value, true
	}
	return nil, false
}

func (s *CMStub) getFromReadSet(key []byte) ([]byte, bool) {
	if txRead, ok := s.readMap[s.constructKey(s.contractName, key)]; ok {
		return txRead.Value, true
	}
	return nil, false
}

func (s *CMStub) putIntoReadSet(key []byte, value []byte) {
	s.readMap[s.constructKey(s.contractName, key)] = &protogo.TxRead{
		Key:          key,
		Value:        value,
		ContractName: s.contractName,
		Version:      nil,
	}
}

func (s *CMStub) constructKey(contractName string, key []byte) string {
	return contractName + string(key)
}

func (s *CMStub) GetWriteMap() map[string]*protogo.TxWrite {
	return s.writeMap
}
