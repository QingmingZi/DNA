package service

import (
	"DNA/vm/avm"
	"DNA/smartcontract/states"
	"DNA/smartcontract/storage"
	"fmt"
	"DNA/common"
	"DNA/core/transaction"
	"encoding/hex"
	"DNA/crypto"
	"DNA/core/asset"
	"DNA/core/contract"
	"DNA/core/ledger"
	"DNA/core/code"
	"DNA/core/signature"
	"DNA/errors"
	"bytes"
	"DNA/core/store"
)

type StateMachine struct {
	*StateReader
	DBCache  storage.DBCache
	hashForVerifying []common.Uint160
}

func NewStateMachine(dbCache storage.DBCache) *StateMachine {
	var stateMachine StateMachine
	stateMachine.DBCache = dbCache
	stateMachine.StateReader = NewStateReader()
	stateMachine.StateReader.Register("AntShares.Blockchain.RegisterValidator", stateMachine.RegisterValidator)
	stateMachine.StateReader.Register("AntShares.Blockchain.CreateAsset", stateMachine.CreateAsset)
	stateMachine.StateReader.Register("AntShares.Blockchain.CreateContract", stateMachine.CreateContract)
	stateMachine.StateReader.Register("AntShares.Blockchain.GetContract", stateMachine.GetContract)
	stateMachine.StateReader.Register("AntShares.Asset.Renew", stateMachine.AssetRenew)
	stateMachine.StateReader.Register("AntShares.Storage.Get", stateMachine.StorageGet);
	stateMachine.StateReader.Register("AntShares.Contract.Destroy", stateMachine.ContractDestory)
	stateMachine.StateReader.Register("AntShares.Storage.Put", stateMachine.StoragePut)
	stateMachine.StateReader.Register("AntShares.Storage.Delete", stateMachine.StorageDelete)
	return &stateMachine
}

func (s *StateMachine) GetCodeHashsForVerifying(engine *avm.ExecutionEngine) ([]common.Uint160, error) {
	return engine.GetCodeContainer().(signature.SignableData).GetProgramHashes()
}

func (s *StateMachine) RegisterValidator(engine *avm.ExecutionEngine) (bool, error) {
	pubkeyByte := avm.PopByteArray(engine)
	pubkey, err := crypto.DecodePoint(pubkeyByte)
	if err != nil {
		return false, err
	}
	phs, err := s.GetCodeHashsForVerifying(engine)
	if err != nil {
		return false, err
	}
	c, err := contract.CreateSignatureRedeemScript(pubkey)
	if err != nil {
		return false, err
	}
	h, err := common.ToCodeHash(c)
	if err != nil {
		return false, err
	}
	if !contains(phs, h) {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[StateMachine], RegisterValidator failed.")
	}
	b := new(bytes.Buffer)
	pubkey.Serialize(b)
	validatorState, err := s.DBCache.GetOrAdd(store.ST_Validator, b.String(), &states.ValidatorState{PublicKey: pubkey})
	if err != nil {
		return false, err
	}
	avm.PushData(engine, validatorState)
	return true, nil
}

func (s *StateMachine) CreateAsset(engine *avm.ExecutionEngine) (bool, error) {
	tx := engine.GetCodeContainer().(*transaction.Transaction);
	assetId := tx.Hash()
	assertType := avm.PopBigInt(engine)
	name := avm.PopByteArray(engine)
	amount := avm.PopBigInt(engine)
	precision := avm.PopBigInt(engine)
	ownerByte := avm.PopByteArray(engine)
	owner, err := crypto.DecodePoint(ownerByte)
	if err != nil {
		return false, err
	}
	adminByte := avm.PopByteArray(engine)
	admin, err := common.Uint160ParseFromBytes(adminByte)
	if err != nil {
		return false, err
	}
	issueByte := avm.PopByteArray(engine)
	issue, err := common.Uint160ParseFromBytes(issueByte)
	if err != nil {
		return false, err
	}
	phs, err := s.GetCodeHashsForVerifying(engine)
	c, err := contract.CreateSignatureRedeemScript(owner)
	if err != nil {
		return false, err
	}
	h, err := common.ToCodeHash(c)
	if !contains(phs, h) {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[StateMachine], CreateAsset failed.")
	}
	b := new(bytes.Buffer)
	assetId.Serialize(b)
	assetState, err := s.DBCache.GetOrAdd(store.ST_Asset, b.String(), &states.AssetState{
		AssetId: assetId,
		AssetType: asset.AssetType(assertType.Int64()),
		Name: hex.EncodeToString(name),
		Amount: common.Fixed64(amount.Int64()),
		Precision: byte(precision.Int64()),
		Admin: admin,
		Issuer: issue,
		Owner: owner,
		Expiration: ledger.DefaultLedger.Store.GetHeight() + 1 + 2000000,
		IsFrozen: false,
	})
	if err != nil {
		return false, err
	}
	avm.PushData(engine, assetState)
	return true, nil
}

func (s *StateMachine) CreateContract(engine *avm.ExecutionEngine) (bool, error) {
	codeByte := avm.PopByteArray(engine)
	parameters := avm.PopByteArray(engine)
	parameterList := make([]contract.ContractParameterType, 0)
	for _, v := range parameters {
		parameterList = append(parameterList, contract.ContractParameterType(v))
	}
	returnType := avm.PopInt(engine)
	nameByte := avm.PopByteArray(engine)
	versionByte := avm.PopByteArray(engine)
	authorByte := avm.PopByteArray(engine)
	emailByte := avm.PopByteArray(engine)
	descByte := avm.PopByteArray(engine)
	funcCode := &code.FunctionCode{
		Code: codeByte,
		ParameterTypes: parameterList,
		ReturnType: contract.ContractParameterType(returnType),
	}
	contractState := &states.ContractState{
		Code: funcCode,
		Name: hex.EncodeToString(nameByte),
		Version: hex.EncodeToString(versionByte),
		Author: hex.EncodeToString(authorByte),
		Email: hex.EncodeToString(emailByte),
		Description: hex.EncodeToString(descByte),
	}
	avm.PushData(engine, contractState)
	return true, nil
}

func (s *StateMachine) GetContract(engine *avm.ExecutionEngine) (bool, error) {
	hashByte := avm.PopByteArray(engine)
	hash, err := common.Uint160ParseFromBytes(hashByte)
	if err != nil {
		return false, err
	}
	item, err := s.DBCache.TryGet(store.ST_Contract, storage.KeyToStr(&hash))
	if err != nil {
		return false, err
	}
	avm.PushData(engine, item.(*states.ContractState))
	return true, nil
}

func (s *StateMachine) AssetRenew(engine *avm.ExecutionEngine) (bool, error) {
	data := avm.PopInteropInterface(engine)
	years := avm.PopInt(engine)
	at := data.(*states.AssetState)
	height := ledger.DefaultLedger.Store.GetHeight() + 1
	b := new(bytes.Buffer)
	at.AssetId.Serialize(b)
	state, err := s.DBCache.TryGet(store.ST_Asset, b.String())
	if err != nil {
		return false, err
	}
	assetState := state.(*states.AssetState)
	if assetState.Expiration < height {
		assetState.Expiration = height
	}
	assetState.Expiration += uint32(years) * 2000000
	return true, nil
}

func (s *StateMachine) ContractDestory(engine *avm.ExecutionEngine) (bool, error) {
	data := engine.CurrentContext().CodeHash
	if data != nil {
		return false, nil
	}
	hash, err := common.Uint160ParseFromBytes(data)
	if err != nil {
		return false, err
	}
	keyStr := storage.KeyToStr(&hash)
	item, err := s.DBCache.TryGet(store.ST_Contract, keyStr)
	if err != nil || item == nil {
		return false, err
	}
	s.DBCache.GetWriteSet().Delete(keyStr)
	return true, nil
}

func (s *StateMachine) CheckStorageContext(context *StorageContext) (bool, error) {
	item, err := s.DBCache.TryGet(store.ST_Contract, storage.KeyToStr(context.codeHash))
	if err != nil {
		return false, err
	}
	if item == nil {
		return false, fmt.Errorf("check storage context fail, codehash=%v", context.codeHash)
	}
	return true, nil
}

func (s *StateMachine) StorageGet(engine *avm.ExecutionEngine) (bool, error) {
	opInterface := avm.PopInteropInterface(engine)
	context := opInterface.(*StorageContext)
	if exist, err := s.CheckStorageContext(context); !exist {
		return false, err
	}
	key := avm.PopByteArray(engine)
	storageKey := states.NewStorageKey(context.codeHash, key)
	item, err := s.DBCache.TryGet(store.ST_Storage, storage.KeyToStr(storageKey))
	if err != nil {
		return false, err
	}
	avm.PushData(engine, item.(*states.StorageItem).Value)
	return true, nil
}

func (s *StateMachine) StoragePut(engine *avm.ExecutionEngine) (bool, error) {
	opInterface := avm.PopInteropInterface(engine)
	context := opInterface.(*StorageContext)
	key := avm.PopByteArray(engine)
	value := avm.PopByteArray(engine)
	storageKey := states.NewStorageKey(context.codeHash, key)
	s.DBCache.GetOrAdd(store.ST_Storage, storage.KeyToStr(storageKey), states.NewStorageItem(value))
	return true, nil
}

func (s *StateMachine) StorageDelete(engine *avm.ExecutionEngine) (bool, error) {
	opInterface := avm.PopInteropInterface(engine)
	context := opInterface.(*StorageContext)
	key := avm.PopByteArray(engine)
	storageKey := states.NewStorageKey(context.codeHash, key)
	s.DBCache.GetWriteSet().Delete(storage.KeyToStr(storageKey))
	return true, nil
}

func contains(programHashes []common.Uint160, programHash common.Uint160) bool {
	for _, v := range programHashes {
		if v == programHash {
			return true
			break
		}
	}
	return false
}




