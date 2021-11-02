package shim

import "chainmaker.org/chainmaker-contract-sdk-docker-go/pb/protogo"

// CMContract user contract interface
type CMContract interface {
	// InitContract used to deploy and upgrade contract
	InitContract(stub CMStubInterface) protogo.Response
	// InvokeContract used to invoke contract
	InvokeContract(stub CMStubInterface) protogo.Response
}

type CMStubInterface interface {
	// GetArgs get arg from transaction parameters
	// @return: 参数map
	GetArgs() map[string][]byte
	// GetState get [key] from chain and db
	// @param key: 获取的参数名
	// @return1: 获取结果
	// @return2: 获取错误信息
	GetState(key []byte) ([]byte, error)
	// PutState put [key, value] to chain
	// @param1 key: 参数名
	// @param2 value: 参数值
	// @return1: 上传参数错误信息
	PutState(key []byte, value []byte) error
	// DelState delete [key] to chain
	// @param1 key: 删除的参数名
	// @return1：删除参数的错误信息
	DelState(key []byte) error
	// GetCreatorOrgId get tx creator org id
	// @return1: 合约创建者的组织ID
	// @return2: 获取错误信息
	GetCreatorOrgId() (string, error)
	// GetCreatorRole get tx creator role
	// @return1: 合约创建者的角色
	// @return2: 获取错误信息
	GetCreatorRole() (string, error)
	// GetCreatorPk get tx creator pk
	// @return1: 合约创建者的公钥
	// @return2: 获取错误信息
	GetCreatorPk() (string, error)
	// GetSenderOrgId get tx sender org id
	// @return1: 交易发起者的组织ID
	// @return2: 获取错误信息
	GetSenderOrgId() (string, error)
	// GetSenderRole get tx sender role
	// @return1: 交易发起者的角色
	// @return2: 获取错误信息
	GetSenderRole() (string, error)
	// GetSenderPk get tx sender pk
	// @return1: 交易发起者的公钥
	// @return2: 获取错误信息
	GetSenderPk() (string, error)
	// GetBlockHeight get tx block height
	// @return1: 当前块高度
	// @return2: 获取错误信息
	GetBlockHeight() (int, error)
	// GetTxId get current tx id
	// @return1: 交易ID
	// @return2: 获取错误信息
	GetTxId() (string, error)
	// GetTxTimeStamp get tx timestamp
	// @return1: 交易timestamp
	// @return2: 获取错误信息
	GetTxTimeStamp() (string, error)
	// EmitEvent emit event, you can subscribe to the event using the SDK
	// @param1 topic: 合约事件的主题
	// @param2 data: 合约事件的数据，参数数量不可大于16
	EmitEvent(topic string, data []string)
	// Log record log to chain server
	// @param message: 事情日志的信息
	Log(message string)
	// CallContract invoke another contract and get response
	// @param1: 合约名称
	// @param2: 合约版本
	// @param3: 合约参数
	// @return1: 合约结果
	CallContract(contractName, contractVersion string, args map[string][]byte) protogo.Response

	//NewIterator(key string, limit string) (ResultSet, error)

	//NewIteratorWithField(key string, startField string, limitField string) (ResultSet, error)
	//
	//NewIteratorPrefixWithKey(key string) (ResultSet, error)
	//
	//NewIteratorPrefixWithKeyField(startKey string, startField string) (ResultSet, error)
}

// ResultSet iterator query result
type ResultSet interface {
	// HasNext return does the next line exist
	HasNext() bool

	Close() (bool, error)
}

type ResultSetKV interface {
	ResultSet
	// Next return key,field,value,code
	Next() (string, string, []byte, error)
}
