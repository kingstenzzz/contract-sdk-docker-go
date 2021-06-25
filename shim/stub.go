package shim

type CMStub struct {
	args    [][]byte
	Handler *Handler

	// todo: change to input protogo.Input
	// todo: method

	// snapshot
}

func NewCMStub(handler *Handler) *CMStub {

	stub := &CMStub{
		args:    nil,
		Handler: handler,
	}

	return stub
}

func (s *CMStub) GetArgs() [][]byte {
	return s.args
}

func (s *CMStub) GetStringArgs() []string {
	args := s.GetArgs()
	stringArgs := make([]string, 0, len(args))
	for _, barg := range args {
		stringArgs = append(stringArgs, string(barg))
	}

	return stringArgs
}

func (s *CMStub) GetFunctionAndParameters() (function string, params []string) {
	args := s.GetStringArgs()
	function = ""

	if len(args) >= 1 {
		function = args[0]
		params = args[1:]
	}
	return
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
