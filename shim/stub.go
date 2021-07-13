package shim

import (
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
	readMap  map[string][]byte
	writeMap map[string][]byte
}

func NewCMStub(handler *Handler, args map[string]string, contractName string) *CMStub {

	stub := &CMStub{
		args:         args,
		Handler:      handler,
		contractName: contractName,
		readMap:      make(map[string][]byte, MapSize),
		writeMap:     make(map[string][]byte, MapSize),
	}

	return stub
}

func (s *CMStub) GetArgs() map[string]string {
	return s.args
}

func (s *CMStub) GetState(key []byte) ([]byte, error) {
	Logger.Debugf("get state for [%s]", string(key))
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

	//getStateKey := s.constructKey(s.contractName, key)
	_ = s.Handler.SendGetStateReq(key, responseCh)

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
	s.writeMap[string(key)] = value
	Logger.Debugf("put key[%s] - value[%s] into write set\n", string(key), string(value))
}

func (s *CMStub) getFromWriteSet(key []byte) ([]byte, bool) {
	Logger.Debugf("get key[%s] from write set\n", string(key))
	if txWrite, ok := s.writeMap[string(key)]; ok {
		return txWrite, true
	}
	return nil, false
}

func (s *CMStub) getFromReadSet(key []byte) ([]byte, bool) {
	Logger.Debugf("get key[%s] from read set\n", string(key))
	if txRead, ok := s.readMap[string(key)]; ok {
		return txRead, true
	}
	return nil, false
}

func (s *CMStub) putIntoReadSet(key []byte, value []byte) {
	s.readMap[string(key)] = value
	Logger.Debugf("put key[%s] - value[%s] into read set\n", string(key), string(value))
}

func (s *CMStub) constructKey(contractName string, key []byte) string {
	return contractName + string(key)
}

func (s *CMStub) GetWriteMap() map[string][]byte {
	return s.writeMap
}
