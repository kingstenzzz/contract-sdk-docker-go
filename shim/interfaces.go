package shim

import "chainmaker.org/chainmaker-contract-sdk-docker-go/pb/protogo"

type CMContract interface {
	Init(stub CMStubInterface) protogo.Response

	Invoke(stub CMStubInterface) protogo.Response
}

type CMStubInterface interface {
	GetArgs() map[string]string

	GetState(key string) ([]byte, error)

	PutState(key string, value []byte) error

	DelState(key string) error
}
