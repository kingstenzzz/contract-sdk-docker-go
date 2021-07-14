package shim

import (
	"chainmaker.org/chainmaker-contract-sdk-docker-go/logger"
	"chainmaker.org/chainmaker-contract-sdk-docker-go/pb/protogo"
	"chainmaker.org/chainmaker-contract-sdk-docker-go/shim/internal"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"io"
	"os"
)

var Logger *zap.SugaredLogger

func GetClientStream(sockAddress string) (ClientStream, error) {

	// establish the connection
	conn, err := internal.NewClientConn(sockAddress)
	if err != nil {
		return nil, err
	}

	return internal.NewContractClient(conn)
}

func Start(cmContract CMContract) error {

	Logger = logger.NewDockerLogger("[Sandbox]")

	// passing sock address when initial the contract
	sockAddress := os.Args[0]
	handlerName := os.Args[1]
	contractName := os.Args[2]

	Logger.Debugf("sandbox - get address: %s", sockAddress)
	Logger.Debugf("sandbox - get handler name: %s", handlerName)
	Logger.Debugf("sandbox - get contract name: %s", contractName)

	// get sandbox stream
	stream, err := GetClientStream(sockAddress)
	if err != nil {
		return err
	}

	err = startClientChat(stream, cmContract, handlerName, contractName)
	if err != nil {
		return err
	}
	// wait to end
	Logger.Debugf("sandbox - end ...")
	return nil
}

func startClientChat(stream ClientStream, contract CMContract, handlerName, contractName string) error {
	defer stream.CloseSend()
	return chatWithManager(stream, contract, handlerName, contractName)
}

func chatWithManager(stream ClientStream, userContract CMContract, handlerName, contractName string) error {
	Logger.Debugf("sandbox - chat with manager")

	// Create the shim handler responsible for all control logic
	handler := newHandler(stream, userContract, handlerName, contractName)

	// Send the register
	payloadString := handlerName
	payload := []byte(payloadString)

	if err := handler.SendMessage(&protogo.DMSMessage{
		Type:         protogo.DMSMessageType_DMS_MESSAGE_TYPE_REGISTER,
		ContractName: contractName,
		Payload:      payload,
	}); err != nil {
		return fmt.Errorf("error sending chaincode REGISTER: %s", err)
	}

	// holds return values from gRPC Recv below
	type recvMsg struct {
		msg *protogo.DMSMessage
		err error
	}
	msgAvail := make(chan *recvMsg, 1)
	errc := make(chan error)
	fCh := make(chan bool, 1)

	receiveMessage := func() {
		in, err := stream.Recv()
		msgAvail <- &recvMsg{in, err}
	}

	// todo: change two goroutines, make recev and send different
	// finish condition: receive completed message
	go receiveMessage()
	for {
		select {
		case rmsg := <-msgAvail:
			switch {
			case rmsg.err == io.EOF:
				Logger.Debugf("server closed")
				return nil
			case rmsg.err != nil:
				Logger.Errorf("receive err: [%s]", rmsg.err)
				err := fmt.Errorf("receive failed: %s", rmsg.err)
				return err
			case rmsg.msg == nil:
				err := errors.New("received nil message, ending chaincode stream")
				return err
			default:
				err := handler.handleMessage(rmsg.msg, fCh)
				if err != nil {
					err = fmt.Errorf("error handling message: %s", err)
					return err
				}
			}

			go receiveMessage()

		case sendErr := <-errc:
			if sendErr != nil {
				err := fmt.Errorf("error sending: %s", sendErr)
				return err
			}

		case <-fCh:
			Logger.Debugf("sandbox - finished")
			close(msgAvail)
			close(fCh)
			close(errc)
			return nil
		}

	}
}
