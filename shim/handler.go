package shim

import (
	"chainmaker.org/chainmaker-contract-sdk-docker-go/pb/protogo"
	"fmt"
	"github.com/golang/protobuf/proto"
	"sync"
)

type state string

const (
	created state = "created"

	ready state = "ready"
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
	serialLock sync.Mutex

	contactStream ContactStream
	cmContract    CMContract
	state         state

	handlerName  string
	contractName string
	responseCh   chan []byte
}

// NewChaincodeHandler returns a new instance of the shim side handler.
func newHandler(chaincodeStream ContactStream, cmContract CMContract, handlerName, contractName string) *Handler {
	return &Handler{
		contactStream: chaincodeStream,
		cmContract:    cmContract,
		state:         created,
		handlerName:   handlerName,
		contractName:  contractName,
		responseCh:    nil,
	}
}

// SendMessage Send on the gRPC client.
func (h *Handler) SendMessage(msg *protogo.DMSMessage) error {
	h.serialLock.Lock()
	defer h.serialLock.Unlock()

	Logger.Debugf("sandbox - send message: [%v]", msg)

	return h.contactStream.Send(msg)
}

// handleMessage message handles loop for shim side of chaincode/peer stream.
func (h *Handler) handleMessage(msg *protogo.DMSMessage, finishCh chan bool) error {

	Logger.Debugf("sandbox - handle message: [%v]", msg)
	var err error

	switch h.state {
	case created:
		err = h.handleCreated(msg)
	case ready:
		err = h.handleReady(msg, finishCh)
	default:
		panic(fmt.Sprintf("invalid handler state: %s", h.state))
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
		return fmt.Errorf("sandbox - handler [%s] cannot handle message (%s) while in state: %s", h.handlerName, registeredMsg.Type, h.state)
	}
	h.state = ready
	return h.afterCreated()
}

func (h *Handler) afterCreated() error {
	readyMsg := &protogo.DMSMessage{
		Type:         protogo.DMSMessageType_DMS_MESSAGE_TYPE_READY,
		ContractName: h.contractName,
		Payload:      nil,
	}
	return h.SendMessage(readyMsg)
}

// ------------------------------------------

func (h *Handler) handleReady(readyMsg *protogo.DMSMessage, finishCh chan bool) error {
	switch readyMsg.Type {
	case protogo.DMSMessageType_DMS_MESSAGE_TYPE_INIT:
		go func() {
			err := h.handleInit(readyMsg)
			if err != nil {
				Logger.Errorf("fail to handle init")
			}
		}()
	case protogo.DMSMessageType_DMS_MESSAGE_TYPE_INVOKE:
		go func() {
			err := h.handleInvoke(readyMsg)
			if err != nil {
				Logger.Errorf("fail to handle invoke")
			}
		}()
	case protogo.DMSMessageType_DMS_MESSAGE_TYPE_RESPONSE:
		return h.handleResponse(readyMsg)
	case protogo.DMSMessageType_DMS_MESSAGE_TYPE_COMPLETED:
		return h.handleCompleted(finishCh)
	}
	return nil
}

func (h *Handler) handleInit(readyMsg *protogo.DMSMessage) error {

	// deal with parameters
	var input protogo.Input
	err := proto.UnmarshalMerge(readyMsg.Payload, &input)
	if err != nil {
		return err
	}

	var args map[string]string

	stub := NewCMStub(h, args, h.contractName)

	// get result
	response := h.cmContract.InitContract(stub)

	// construct complete message
	writeMap := stub.GetWriteMap()
	responseWithWriteMap := &protogo.ResponseWithWriteMap{
		Response: &response,
		WriteMap: writeMap,
	}

	responseWithWriteMapPayload, err := proto.Marshal(responseWithWriteMap)
	if err != nil {
		return err
	}
	completedMsg := &protogo.DMSMessage{
		Type:    protogo.DMSMessageType_DMS_MESSAGE_TYPE_COMPLETED,
		Payload: responseWithWriteMapPayload,
	}

	return h.SendMessage(completedMsg)

}

func (h *Handler) handleInvoke(readyMsg *protogo.DMSMessage) error {
	// deal with parameters
	var input protogo.Input
	err := proto.UnmarshalMerge(readyMsg.Payload, &input)
	args := input.Args

	stub := NewCMStub(h, args, h.contractName)

	response := h.cmContract.InvokeContract(stub)

	// construct complete message
	writeMap := stub.GetWriteMap()
	responseWithWriteMap := &protogo.ResponseWithWriteMap{
		Response: &response,
		WriteMap: writeMap,
	}

	// construct complete message
	responseWithWriteMapPayload, err := proto.Marshal(responseWithWriteMap)
	if err != nil {
		return err
	}

	completedMsg := &protogo.DMSMessage{
		Type:    protogo.DMSMessageType_DMS_MESSAGE_TYPE_COMPLETED,
		Payload: responseWithWriteMapPayload,
	}

	return h.SendMessage(completedMsg)

}

func (h *Handler) SendGetStateReq(key []byte, responseCh chan []byte) error {
	getStateMsg := &protogo.DMSMessage{
		Type:         protogo.DMSMessageType_DMS_MESSAGE_TYPE_GET_STATE,
		ContractName: h.contractName,
		Payload:      key,
	}

	h.responseCh = responseCh

	return h.SendMessage(getStateMsg)
}

func (h *Handler) handleResponse(readyMsg *protogo.DMSMessage) error {
	h.responseCh <- readyMsg.Payload
	close(h.responseCh)
	h.responseCh = nil

	return nil
}

func (h *Handler) handleCompleted(finishCh chan bool) error {
	finishCh <- true
	return nil
}
