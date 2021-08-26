package shim

import "chainmaker.org/chainmaker-contract-sdk-docker-go/pb/protogo"

type CMContract interface {
	InitContract(stub CMStubInterface) protogo.Response

	InvokeContract(stub CMStubInterface) protogo.Response
}

type CMStubInterface interface {
	GetArgs() map[string]string

	GetState(key []byte) ([]byte, error)

	PutState(key []byte, value []byte) error

	DelState(key []byte) error

	GetCreatorOrgId() (string, error)

	GetCreatorRole() (string, error)

	GetCreatorPk() (string, error)

	GetSenderOrgId() (string, error)

	GetSenderRole() (string, error)

	GetSenderPk() (string, error)

	GetBlockHeight() (int, error)

	GetTxId() (string, error)

	EmitEvent(topic string, data []string)

	Log(message string)
}
