/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package shim

import (
	"errors"
	"fmt"
	"io"
	"os"

	"chainmaker.org/chainmaker/chainmaker-contract-sdk-docker-go/logger"
	"chainmaker.org/chainmaker/chainmaker-contract-sdk-docker-go/pb/protogo"
	"chainmaker.org/chainmaker/chainmaker-contract-sdk-docker-go/shim/internal"
	"go.uber.org/zap"
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

	// passing sock address when initial the contract
	sockAddress := os.Args[0]
	processName := os.Args[1]
	contractName := os.Args[2]
	contractVersion := os.Args[3]

	Logger = logger.NewDockerLogger("[Sandbox]", os.Args[4])
	Logger.Debugf("loglevel: %s", os.Args[4])

	// get sandbox stream
	stream, err := GetClientStream(sockAddress)
	if err != nil {
		Logger.Errorf("sandbox process [%s] fail to establish stream", processName)
		return err
	}
	Logger.Debugf("sandbox process [%s] established the stream", processName)

	err = startClientChat(stream, cmContract, processName, contractName, contractVersion)
	if err != nil {
		Logger.Errorf("sandbox process [%s] fail to chat with manager", processName)
		return err
	}
	// wait to end
	Logger.Debugf("sandbox - end")
	return nil
}

func startClientChat(stream ClientStream, contract CMContract, processName, contractName, contractVersion string) error {
	defer func(stream ClientStream) {
		err := stream.CloseSend()
		if err != nil {
			Logger.Errorf("sandbox process [%s] close send err [%s]", processName)
			return
		}
	}(stream)
	return chatWithManager(stream, contract, processName, contractName, contractVersion)
}

func chatWithManager(stream ClientStream, userContract CMContract, processName, contractName, contractVersion string) error {
	Logger.Debugf("sandbox process [%s] - chat with manager", processName)

	// Create the shim handler responsible for all control logic
	handler := newHandler(stream, userContract, processName, contractName, contractVersion)

	// Send the register
	payloadString := processName
	payload := []byte(payloadString)

	if err := handler.SendMessage(&protogo.DMSMessage{
		Type:    protogo.DMSMessageType_DMS_MESSAGE_TYPE_REGISTER,
		Payload: payload,
	}); err != nil {
		return fmt.Errorf("error sending chaincode REGISTER: %s", err)
	}

	// holds return values from gRPC Recv below
	//type recvMsg struct {
	//	msg *protogo.DMSMessage
	//	err error
	//}

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			Logger.Errorf("sandbox process [%s] - recv eof", processName)
			return err
		}
		if err != nil {
			err := fmt.Errorf("receive failed: %s", err)
			Logger.Error(err)
			return err
		}
		if in == nil {
			err := errors.New("received nil message, ending chaincode stream")
			Logger.Error(err)
			return err
		}
		err = handler.handleMessage(in)
		if err != nil {
			err = fmt.Errorf("sandbox process [%s] error handling message: %s", processName, err)
			return err
		}
	}
	//msgAvail := make(chan *recvMsg, 1)
	//errc := make(chan error)
	//fCh := make(chan bool, 1)

	//receiveMessage := func() {
	//	in, err := stream.Recv()
	//	if err == io.EOF {
	//		Logger.Errorf("sandbox process [%s] - recv eof", processName)
	//		return
	//	}
	//	msgAvail <- &recvMsg{in, err}
	//}
	// finish condition: receive completed message
	//go receiveMessage()
	//for {
	//	select {
	//	case rmsg := <-msgAvail:
	//		switch {
	//		case rmsg.err == io.EOF:
	//			err := fmt.Errorf("receive end: %s", rmsg.err)
	//			return err
	//		case rmsg.err != nil:
	//			err := fmt.Errorf("receive failed: %s", rmsg.err)
	//			return err
	//		case rmsg.msg == nil:
	//			err := errors.New("received nil message, ending chaincode stream")
	//			return err
	//		default:
	//			err := handler.handleMessage(rmsg.msg, fCh)
	//			if err != nil {
	//				err = fmt.Errorf("sandbox process [%s] error handling message: %s", processName, err)
	//				return err
	//			}
	//		}
	//
	//		go receiveMessage()
	//
	//	case sendErr := <-errc:
	//		if sendErr != nil {
	//			err := fmt.Errorf("error sending: %s", sendErr)
	//			Logger.Errorf("\"sandbox process [%s] - err in send [%s]", processName, err)
	//			return err
	//		}
	//
	//	case <-fCh:
	//		close(msgAvail)
	//		close(fCh)
	//		close(errc)
	//		return nil
	//	}
	//}
}
