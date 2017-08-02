package contract

//parameter defined type.
type ContractParameterType byte

const (
	Signature ContractParameterType = iota
	Boolean
	Integer
	Hash160
	Hash256
	ByteArray
	PublicKey
	PublicKeyArray
	//Code
	//String
	Object
	Void = 0xff
)
