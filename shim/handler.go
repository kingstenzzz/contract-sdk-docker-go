/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package shim

import (
	"errors"
	"fmt"
	"sync"

	"chainmaker.org/chainmaker/chainmaker-contract-sdk-docker-go/pb/protogo"
	"github.com/golang/protobuf/proto"
)

type state string

const (
	created state = "created"

	ready state = "ready"
)

// handler tx state:
// when handler handle such tx, state is occupied
// when handler doesn't handle tx, state is empty
const (
	empty = iota

	occupied
)

type ContactStream interface {
	Send(message *protogo.DMSMessage) error
	Recv() (*protogo.DMSMessage, error)
}

type ClientStream interface {
	ContactStream
	CloseSend() error
}

type Handler struct {
	handlerLock     sync.Mutex
	contactStream   ContactStream
	cmContract      CMContract
	state           state
	processName     string
	contractName    string
	contractVersion string
	responseCh      chan *protogo.DMSMessage

	// related to each tx
	currentTxId     string
	currentTxState  int
	currentTxHeight uint32
}

// NewChaincodeHandler returns a new instance of the shim side handler.
func newHandler(chaincodeStream ContactStream, cmContract CMContract, processName, contractName, contractVersion string) *Handler {
	return &Handler{
		contactStream:   chaincodeStream,
		cmContract:      cmContract,
		state:           created,
		processName:     processName,
		contractName:    contractName,
		contractVersion: contractVersion,
		responseCh:      nil,

		currentTxId:     "",
		currentTxState:  empty,
		currentTxHeight: 0,
	}
}

// SendMessage Send on the gRPC client.
func (h *Handler) SendMessage(msg *protogo.DMSMessage) error {
	h.handlerLock.Lock()
	defer h.handlerLock.Unlock()
	Logger.Debugf("sandbox process [%s] tx [%s] - send message: [%v]", h.processName, msg.TxId, msg)

	return h.contactStream.Send(msg)
}

// handleMessage message handles loop for shim side of chaincode/peer stream.
func (h *Handler) handleMessage(msg *protogo.DMSMessage) error {

	Logger.Debugf("sandbox process [%s] tx [%s] - handle message: [%v]", h.processName, msg.TxId, msg)
	var err error

	switch h.state {
	case created:
		err = h.handleCreated(msg)
	case ready:
		err = h.handleReady(msg)
	default:
		err = fmt.Errorf("invalid handler state: %s", h.state)
	}

	if err != nil {
		return err
	}
	return nil
}

// ------------------------------------------

// receive registered
func (h *Handler) handleCreated(registeredMsg *protogo.DMSMessage) error {
	if registeredMsg.Type != protogo.DMSMessageType_DMS_MESSAGE_TYPE_REGISTERED {
		return fmt.Errorf("sandbox - cannot handle message (%s) while in state: %s", registeredMsg.Type, h.state)
	}
	h.state = ready
	return h.afterCreated()
}

func (h *Handler) afterCreated() error {
	readyMsg := &protogo.DMSMessage{
		Type: protogo.DMSMessageType_DMS_MESSAGE_TYPE_READY,
	}
	return h.SendMessage(readyMsg)
}

// ------------------------------------------

func (h *Handler) handleReady(readyMsg *protogo.DMSMessage) error {

	// check new msg's TxId is equal to current TxId, if not equal, abandon this msg
	// abandoned msg is not error
	if len(h.currentTxId) > 0 && h.currentTxId != readyMsg.TxId {
		errMsg := fmt.Sprintf("abandon msg [%+v] because handler is handling existing tx [%s]\n",
			readyMsg, h.currentTxId)
		Logger.Error(errMsg)
		return nil
	}

	switch readyMsg.Type {
	case protogo.DMSMessageType_DMS_MESSAGE_TYPE_INIT:
		go func() {
			err := h.handleInit(readyMsg)
			if err != nil {
				Logger.Errorf("fail to handle init [%s]", err)
			}
		}()
	case protogo.DMSMessageType_DMS_MESSAGE_TYPE_INVOKE:
		go func() {
			err := h.handleInvoke(readyMsg)
			if err != nil {
				Logger.Errorf("fail to handle invoke [%s]", err)
			}
		}()
	case protogo.DMSMessageType_DMS_MESSAGE_TYPE_GET_STATE_RESPONSE:
		return h.handleResponse(readyMsg)
	case protogo.DMSMessageType_DMS_MESSAGE_TYPE_CALL_CONTRACT_RESPONSE:
		return h.handleResponse(readyMsg)
	//case protogo.DMSMessageType_DMS_MESSAGE_TYPE_COMPLETED:
	//	return h.handleCompleted(finishCh)
	case protogo.DMSMessageType_DMS_MESSAGE_TYPE_CREATE_KV_ITERATOR_RESPONSE:
		return h.handleResponse(readyMsg)
	case protogo.DMSMessageType_DMS_MESSAGE_TYPE_CONSUME_KV_ITERATOR_RESPONSE:
		return h.handleResponse(readyMsg)
	case protogo.DMSMessageType_DMS_MESSAGE_TYPE_CREATE_KEY_HISTORY_ITER_RESPONSE:
		return h.handleResponse(readyMsg)
	case protogo.DMSMessageType_DMS_MESSAGE_TYPE_CONSUME_KEY_HISTORY_ITER_RESPONSE:
		return h.handleResponse(readyMsg)
	case protogo.DMSMessageType_DMS_MESSAGE_TYPE_GET_SENDER_ADDRESS_RESPONSE:
		return h.handleResponse(readyMsg)
	}
	return nil
}

func (h *Handler) handleInit(readyMsg *protogo.DMSMessage) error {

	err := h.updateNewTx(readyMsg)
	if err != nil {
		return err
	}

	// deal with parameters
	var input protogo.Input
	err = proto.Unmarshal(readyMsg.Payload, &input)
	if err != nil {
		return err
	}
	args := input.Args

	stub := NewCMStub(h, args, h.contractName, h.contractVersion)

	// get result
	response := h.cmContract.InitContract(stub)

	// construct complete message
	writeMap := stub.GetWriteMap()
	events := stub.GetEvents()
	contractResponse := &protogo.ContractResponse{
		Response: &response,
		WriteMap: writeMap,
		Events:   events,
	}

	responseWithWriteMapPayload, err := proto.Marshal(contractResponse)
	if err != nil {
		return err
	}
	completedMsg := &protogo.DMSMessage{
		TxId:          h.currentTxId,
		Type:          protogo.DMSMessageType_DMS_MESSAGE_TYPE_COMPLETED,
		CurrentHeight: h.currentTxHeight,
		Payload:       responseWithWriteMapPayload,
	}

	h.resetTx()

	return h.SendMessage(completedMsg)
}

func (h *Handler) handleInvoke(readyMsg *protogo.DMSMessage) error {

	err := h.updateNewTx(readyMsg)
	if err != nil {
		return err
	}
	// deal with parameters
	var input protogo.Input
	err = proto.Unmarshal(readyMsg.Payload, &input)
	if err != nil {
		return err
	}
	args := input.Args

	stub := NewCMStub(h, args, h.contractName, h.contractVersion)

	response := h.cmContract.InvokeContract(stub)

	// construct complete message
	writeMap := stub.GetWriteMap()
	events := stub.GetEvents()
	contractResponse := &protogo.ContractResponse{
		Response: &response,
		WriteMap: writeMap,
		Events:   events,
	}

	// current height > 0, also send read map
	if h.currentTxHeight > 0 {
		contractResponse.ReadMap = stub.GetReadMap()
	}

	// construct complete message
	contractResponsePayload, err := proto.Marshal(contractResponse)
	if err != nil {
		return err
	}

	completedMsg := &protogo.DMSMessage{
		TxId:          h.currentTxId,
		Type:          protogo.DMSMessageType_DMS_MESSAGE_TYPE_COMPLETED,
		CurrentHeight: h.currentTxHeight,
		Payload:       contractResponsePayload,
	}

	h.resetTx()

	return h.SendMessage(completedMsg)
}

func (h *Handler) updateNewTx(readyMsg *protogo.DMSMessage) error {
	h.handlerLock.Lock()
	defer h.handlerLock.Unlock()

	Logger.Debugf("update new handler tx [%s]", readyMsg.TxId)

	if h.currentTxState == occupied && len(h.currentTxId) > 0 {
		errMsg := fmt.Sprintf("abandon new tx [%s] because handler is handling existing tx [%s]",
			readyMsg.TxId, h.currentTxId)
		return errors.New(errMsg)
	}

	h.currentTxId = readyMsg.TxId
	h.currentTxState = occupied
	h.currentTxHeight = readyMsg.CurrentHeight

	return nil
}

func (h *Handler) resetTx() {
	h.handlerLock.Lock()
	defer h.handlerLock.Unlock()

	Logger.Debugf("reset handler tx [%s]", h.currentTxId)

	h.currentTxId = ""
	h.currentTxState = empty
	h.currentTxHeight = 0
}

func (h *Handler) SendGetStateReq(key []byte, responseCh chan *protogo.DMSMessage) error {
	getStateMsg := &protogo.DMSMessage{
		TxId:          h.currentTxId,
		Type:          protogo.DMSMessageType_DMS_MESSAGE_TYPE_GET_STATE_REQUEST,
		CurrentHeight: h.currentTxHeight,
		Payload:       key,
	}

	h.responseCh = responseCh

	return h.SendMessage(getStateMsg)
}

func (h *Handler) SendCallContract(callContractPayload []byte, responseCh chan *protogo.DMSMessage) error {
	callContractMsg := &protogo.DMSMessage{
		TxId:          h.currentTxId,
		Type:          protogo.DMSMessageType_DMS_MESSAGE_TYPE_CALL_CONTRACT_REQUEST,
		CurrentHeight: h.currentTxHeight,
		Payload:       callContractPayload,
	}

	h.responseCh = responseCh

	return h.SendMessage(callContractMsg)
}

func (h *Handler) SendCreateKvIteratorReq(key []byte, responseCh chan *protogo.DMSMessage) error {
	createKvIteratorReq := &protogo.DMSMessage{
		TxId:          h.currentTxId,
		Type:          protogo.DMSMessageType_DMS_MESSAGE_TYPE_CREATE_KV_ITERATOR_REQUEST,
		CurrentHeight: h.currentTxHeight,
		Payload:       key,
	}

	h.responseCh = responseCh

	return h.SendMessage(createKvIteratorReq)
}

func (h *Handler) SendConsumeKvIteratorReq(key []byte, responseCh chan *protogo.DMSMessage) error {
	consumeKvIteratorReq := &protogo.DMSMessage{
		TxId:          h.currentTxId,
		Type:          protogo.DMSMessageType_DMS_MESSAGE_TYPE_CONSUME_KV_ITERATOR_REQUEST,
		CurrentHeight: h.currentTxHeight,
		Payload:       key,
	}

	h.responseCh = responseCh

	return h.SendMessage(consumeKvIteratorReq)
}

func (h *Handler) SendCreateKeyHistoryKvIterReq(key []byte, responseCh chan *protogo.DMSMessage) error {
	createKeyHistoryKvIterReq := &protogo.DMSMessage{
		TxId:          h.currentTxId,
		Type:          protogo.DMSMessageType_DMS_MESSAGE_TYPE_CREATE_KEY_HISTORY_ITER_REQUEST,
		CurrentHeight: h.currentTxHeight,
		Payload:       key,
	}

	h.responseCh = responseCh

	return h.SendMessage(createKeyHistoryKvIterReq)
}

func (h *Handler) SendConsumeKeyHistoryKvIterReq(key []byte, responseCh chan *protogo.DMSMessage) error {
	consumeKvIteratorReq := &protogo.DMSMessage{
		TxId:          h.currentTxId,
		Type:          protogo.DMSMessageType_DMS_MESSAGE_TYPE_CONSUME_KEY_HISTORY_ITER_REQUEST,
		CurrentHeight: h.currentTxHeight,
		Payload:       key,
	}

	h.responseCh = responseCh

	return h.SendMessage(consumeKvIteratorReq)
}

func (h *Handler) SendGetSenderAddrReq(key []byte, responseCh chan *protogo.DMSMessage) error {
	getSenderAddrReq := &protogo.DMSMessage{
		TxId:          h.currentTxId,
		Type:          protogo.DMSMessageType_DMS_MESSAGE_TYPE_GET_SENDER_ADDRESS_REQUEST,
		CurrentHeight: h.currentTxHeight,
		Payload:       key,
	}

	h.responseCh = responseCh

	return h.SendMessage(getSenderAddrReq)
}

func (h *Handler) handleResponse(readyMsg *protogo.DMSMessage) error {
	Logger.Debugf("handle response [%+v]", readyMsg)
	h.responseCh <- readyMsg
	Logger.Debugf("close response channel [%+v]", readyMsg)
	return nil
}

func (h *Handler) handleCompleted(finishCh chan bool) error {
	finishCh <- true
	return nil
}
