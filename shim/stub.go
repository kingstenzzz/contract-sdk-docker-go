package shim

type CMStub struct {
	args    map[string]string
	Handler *Handler

	// snapshot
}

func NewCMStub(handler *Handler, args map[string]string) *CMStub {

	stub := &CMStub{
		args:    args,
		Handler: handler,
	}

	return stub
}

func (s *CMStub) GetArgs() map[string]string {
	return s.args
}


func (s *CMStub) GetState(key string) ([]byte, error) {
	return nil, nil
}

func (s *CMStub) PutState(key string, value []byte) error {
	return nil
}

func (s *CMStub) DelState(key string) error {
	return nil
}
