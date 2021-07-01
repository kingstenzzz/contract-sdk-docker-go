package shim

import (
	"chainmaker.org/chainmaker-contract-sdk-docker-go/pb/protogo"
	"fmt"

	//"fmt"
	"github.com/golang/protobuf/proto"
	"sync"
)

type state string

const (
	created state = "created"

	prepare state = "prepare"

	ready state = "ready"
)

type ContactStream interface {
	Send(*protogo.ContractMessage) error
	Recv() (*protogo.ContractMessage, error)
}

type ClientStream interface {
	ContactStream
	CloseSend() error
}

type Handler struct {
	serialLock sync.Mutex

	contactStream ContactStream

	cmContract CMContract

	state state

	handlerName string
}

// NewChaincodeHandler returns a new instance of the shim side handler.
func newChaincodeHandler(chaincodeStream ContactStream, cmContract CMContract, handlerName string) *Handler {
	return &Handler{
		contactStream: chaincodeStream,
		cmContract:    cmContract,
		state:         created,
		handlerName:   handlerName,
	}
}

// SendMessage Send on the gRPC client.
func (h *Handler) SendMessage(msg *protogo.ContractMessage) error {
	h.serialLock.Lock()
	defer h.serialLock.Unlock()

	Logger.Debugf("sandbox - send message: [%v]", msg)

	return h.contactStream.Send(msg)
}

// handleMessage message handles loop for shim side of chaincode/peer stream.
func (h *Handler) handleMessage(msg *protogo.ContractMessage, errc chan error, finishCh chan bool) error {

	Logger.Debugf("sandbox - handle message: [%v]", msg)
	var err error

	switch h.state {
	case created:
		err = h.handleCreated(msg)
	case prepare:
		err = h.handlePrepare(msg)
	case ready:
		err = h.handleReady(msg, finishCh)
	default:
		panic(fmt.Sprintf("invalid handler state: %s", h.state))
	}
	if err != nil {
		payload := []byte(err.Error())
		errorMsg := &protogo.ContractMessage{Type: protogo.Type_ERROR, Payload: payload, HandlerName: h.handlerName}
		h.SendMessage(errorMsg)
		return err
	}

	return nil
}

func (h *Handler) handleCreated(registeredMsg *protogo.ContractMessage) error {
	if registeredMsg.Type != protogo.Type_REGISTERED {
		return fmt.Errorf("sandbox - handler [%s] cannot handle message (%s) while in state: %s", registeredMsg.HandlerName, registeredMsg.Type, h.state)
	}
	h.state = prepare
	return nil
}

func (h *Handler) handlePrepare(prepareMsg *protogo.ContractMessage) error {
	if prepareMsg.Type != protogo.Type_PREPARE {
		return fmt.Errorf("sandbox - handler [%s] cannot handle message (%s) while in state: %s", prepareMsg.HandlerName, prepareMsg.Type, h.state)
	}
	h.state = prepare

	return h.afterPrepare()
}

func (h *Handler) afterPrepare() error {
	readyMsg := &protogo.ContractMessage{
		Type:        protogo.Type_READY,
		HandlerName: h.handlerName,
		Payload:     nil,
	}
	h.state = ready
	return h.SendMessage(readyMsg)
}

func (h *Handler) handleReady(readyMsg *protogo.ContractMessage, finishCh chan bool) error {
	switch readyMsg.Type {
	case protogo.Type_INIT:
		return h.handleInit(readyMsg, finishCh)
	case protogo.Type_INVOKE:
		return h.handleInvoke(readyMsg, finishCh)
	case protogo.Type_RESPONSE:
		return h.handleResponse(readyMsg)
	case protogo.Type_COMPLETED:
		return h.handleCompleted(finishCh)
	}
	return nil
}

func (h *Handler) handleInit(readyMsg *protogo.ContractMessage, finishCh chan bool) error {

	// deal with parameters

	stub := NewCMStub(h, nil)
	result := h.cmContract.Init(stub)

	resultPayload, err := proto.Marshal(&result)
	if err != nil {
		return err
	}

	completedMsg := &protogo.ContractMessage{
		Type:        protogo.Type_COMPLETED,
		HandlerName: h.handlerName,
		Payload:     resultPayload,
	}

	return h.SendMessage(completedMsg)

}

func (h *Handler) handleInvoke(readyMsg *protogo.ContractMessage, finishCh chan bool) error {

	// deal with parameters
	//fmt.Println("in handle Invoke")

	// get input map

	var input protogo.Input
	err := proto.UnmarshalMerge(readyMsg.Payload, &input)

	args := input.Args

	stub := NewCMStub(h, args)
	result := h.cmContract.Invoke(stub)

	resultPayload, err := proto.Marshal(&result)
	if err != nil {
		return err
	}

	completedMsg := &protogo.ContractMessage{
		Type:        protogo.Type_COMPLETED,
		HandlerName: h.handlerName,
		Payload:     resultPayload,
	}

	return h.SendMessage(completedMsg)

}

func (h *Handler) handleResponse(readyMsg *protogo.ContractMessage) error {
	// handle get_state result
	return nil
}

func (h *Handler) handleCompleted(finishCh chan bool) error {
	finishCh <- true
	return nil
}
