package shim

import (
	"chainmaker.org/chainmaker-contract-sdk-docker-go/pb/protogo"
	"chainmaker.org/chainmaker-contract-sdk-docker-go/shim/internal"
	"errors"
	"fmt"
	"io"
	"os"
)

func GetClientStream(sockAddress string) (ClientStream, error) {

	// establish the connection
	conn, err := internal.NewClientConn(sockAddress)
	if err != nil {
		fmt.Println(10)
		return nil, err
	}

	return internal.NewContractClient(conn)
}

func Start(cmContract CMContract) error {

	//flag.Parse()
	// passing sock address when initial the contract

	sockAddress := os.Args[0]
	handlerName := os.Args[1]
	fmt.Println("sandbox - get address: ", sockAddress)
	fmt.Println("sandbox - get handler name: ", handlerName)

	// get sandbox stream
	stream, err := GetClientStream(sockAddress)
	if err != nil {
		fmt.Println(3)
		return err
	}

	err = startClientChat(stream, cmContract, handlerName)
	if err != nil {
		fmt.Println(4)
		return err
	}
	// wait to end
	fmt.Println("sandbox - end ... ")
	err = stream.CloseSend()
	if err != nil {
		return err
	}
	return nil
}

func startClientChat(stream ClientStream, contract CMContract, handlerName string) error {

	return chatWithManager(stream, contract, handlerName)
}

func chatWithManager(stream ClientStream, contract CMContract, handlerName string) error {
	fmt.Println("sandbox - chat with manager")
	// Create the shim handler responsible for all control logic
	handler := newChaincodeHandler(stream, contract, handlerName)

	// Send the register
	payloadString := "first register"
	payload := []byte(payloadString)

	if err := handler.SendMessage(&protogo.ContractMessage{
		Type:        protogo.Type_REGISTER,
		Payload:     payload,
		HandlerName: handlerName,
	}); err != nil {
		return fmt.Errorf("error sending chaincode REGISTER: %s", err)
	}

	// holds return values from gRPC Recv below
	type recvMsg struct {
		msg *protogo.ContractMessage
		err error
	}
	msgAvail := make(chan *recvMsg, 1)
	errc := make(chan error)
	fCh := make(chan bool, 1)

	receiveMessage := func() {
		in, err := stream.Recv()
		msgAvail <- &recvMsg{in, err}
	}

	go receiveMessage()
	for {
		select {
		case rmsg := <-msgAvail:
			switch {
			case rmsg.err == io.EOF:
				fmt.Println("server closed")
				return errors.New("received EOF, ending chaincode stream")
			case rmsg.err != nil:
				err := fmt.Errorf("receive failed: %s", rmsg.err)
				return err
			case rmsg.msg == nil:
				err := errors.New("received nil message, ending chaincode stream")
				return err
			default:
				err := handler.handleMessage(rmsg.msg, errc, fCh)
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
			fmt.Println("sandbox - finished")
			close(msgAvail)
			close(fCh)
			close(errc)
			return nil
		}

	}
}
