package shim

import (
	"fmt"
	"sync"

	"chainmaker.org/chainmaker-contract-sdk-docker-go/pb/protogo"
	"github.com/golang/protobuf/proto"
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
	serialLock    sync.Mutex
	contactStream ContactStream
	cmContract    CMContract
	state         state
	processName   string
	responseCh    chan *protogo.DMSMessage

	// related to each tx
	txId          string
	currentHeight uint32
}

// NewChaincodeHandler returns a new instance of the shim side handler.
func newHandler(chaincodeStream ContactStream, cmContract CMContract, processName string) *Handler {
	return &Handler{
		contactStream: chaincodeStream,
		cmContract:    cmContract,
		state:         created,
		processName:   processName,
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
		Type:    protogo.DMSMessageType_DMS_MESSAGE_TYPE_READY,
		Payload: nil,
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
	case protogo.DMSMessageType_DMS_MESSAGE_TYPE_GET_STATE_RESPONSE:
		return h.handleResponse(readyMsg)
	case protogo.DMSMessageType_DMS_MESSAGE_TYPE_CALL_CONTRACT_RESPONSE:
		return h.handleResponse(readyMsg)
	case protogo.DMSMessageType_DMS_MESSAGE_TYPE_COMPLETED:
		return h.handleCompleted(finishCh)
	}
	return nil
}

func (h *Handler) handleInit(readyMsg *protogo.DMSMessage) error {

	h.updateTx(readyMsg)

	// deal with parameters
	var input protogo.Input
	err := proto.UnmarshalMerge(readyMsg.Payload, &input)
	if err != nil {
		return err
	}
	args := input.Args

	stub := NewCMStub(h, args)

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
		TxId:          h.txId,
		Type:          protogo.DMSMessageType_DMS_MESSAGE_TYPE_COMPLETED,
		CurrentHeight: h.currentHeight,
		Payload:       responseWithWriteMapPayload,
	}

	return h.SendMessage(completedMsg)

}

func (h *Handler) handleInvoke(readyMsg *protogo.DMSMessage) error {
	h.updateTx(readyMsg)
	// deal with parameters
	var input protogo.Input
	err := proto.UnmarshalMerge(readyMsg.Payload, &input)
	args := input.Args

	stub := NewCMStub(h, args)

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
	if h.currentHeight > 0 {
		contractResponse.ReadMap = stub.GetReadMap()
	}

	// construct complete message
	contractResponsePayload, err := proto.Marshal(contractResponse)
	if err != nil {
		return err
	}

	completedMsg := &protogo.DMSMessage{
		TxId:          h.txId,
		Type:          protogo.DMSMessageType_DMS_MESSAGE_TYPE_COMPLETED,
		CurrentHeight: h.currentHeight,
		Payload:       contractResponsePayload,
	}

	return h.SendMessage(completedMsg)

}

func (h *Handler) updateTx(readyMsg *protogo.DMSMessage) {
	h.txId = readyMsg.TxId
	h.currentHeight = readyMsg.CurrentHeight
}

func (h *Handler) SendGetStateReq(key []byte, responseCh chan *protogo.DMSMessage) error {
	getStateMsg := &protogo.DMSMessage{
		TxId:          h.txId,
		Type:          protogo.DMSMessageType_DMS_MESSAGE_TYPE_GET_STATE_REQUEST,
		CurrentHeight: h.currentHeight,
		Payload:       key,
	}

	h.responseCh = responseCh

	return h.SendMessage(getStateMsg)
}

func (h *Handler) SendCallContract(callContractPayload []byte, responseCh chan *protogo.DMSMessage) error {
	callContractMsg := &protogo.DMSMessage{
		TxId:          h.txId,
		Type:          protogo.DMSMessageType_DMS_MESSAGE_TYPE_CALL_CONTRACT_REQUEST,
		CurrentHeight: h.currentHeight,
		Payload:       callContractPayload,
	}

	h.responseCh = responseCh

	return h.SendMessage(callContractMsg)
}

func (h *Handler) handleResponse(readyMsg *protogo.DMSMessage) error {
	h.responseCh <- readyMsg
	close(h.responseCh)
	h.responseCh = nil

	return nil
}

func (h *Handler) handleCompleted(finishCh chan bool) error {
	finishCh <- true
	return nil
}
