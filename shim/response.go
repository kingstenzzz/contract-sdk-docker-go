package shim

import "chainmaker.org/chainmaker-contract-sdk-docker-go/pb/protogo"

const (
	// OK constant - status code less than 400, endorser will endorse it.
	// OK means init or invoke successfully.
	OK = 200

	// ERROR constant - default error value
	ERROR = 500
)

// Success ...
func Success(payload []byte) protogo.Response {
	return protogo.Response{
		Status:  OK,
		Payload: payload,
	}
}

// Error ...
func Error(msg string) protogo.Response {
	return protogo.Response{
		Status:  ERROR,
		Message: msg,
	}
}