package service

import (
	"DNA/core/ledger"
	"DNA/common"
	"math/big"
	"errors"
	"fmt"
	"DNA/vm/avm/types"
	"DNA/core/transaction"
	"DNA/smartcontract/states"
	"DNA/vm/avm"
	"DNA/common/log"
)

type StateReader struct {
	serviceMap map[string]func(*avm.ExecutionEngine) (bool, error)
}

func NewStateReader() *StateReader {
	var stateReader StateReader
	stateReader.serviceMap = make(map[string]func(*avm.ExecutionEngine) (bool, error), 0)
	stateReader.Register("Neo.Blockchain.GetHeight", stateReader.BlockChainGetHeight)
	stateReader.Register("Neo.Blockchain.GetHeader", stateReader.BlockChainGetHeader)
	stateReader.Register("Neo.Blockchain.GetBlock", stateReader.BlockChainGetBlock)
	stateReader.Register("Neo.Blockchain.GetTransaction", stateReader.BlockChainGetTransaction)
	stateReader.Register("Neo.Blockchain.GetAccount", stateReader.BlockChainGetAccount)
	stateReader.Register("Neo.Blockchain.GetValidators", stateReader.BlockChainGetValidators)
	stateReader.Register("Neo.Blockchain.GetAsset", stateReader.BlockChainGetAsset)

	stateReader.Register("Neo.Header.GetHash", stateReader.HeaderGetHash);
	stateReader.Register("Neo.Header.GetVersion", stateReader.HeaderGetVersion);
	stateReader.Register("Neo.Header.GetPrevHash", stateReader.HeaderGetPrevHash);
	stateReader.Register("Neo.Header.GetMerkleRoot", stateReader.HeaderGetMerkleRoot);
	stateReader.Register("Neo.Header.GetTimestamp", stateReader.HeaderGetTimestamp);
	stateReader.Register("Neo.Header.GetConsensusData", stateReader.HeaderGetConsensusData);
	stateReader.Register("Neo.Header.GetNextConsensus", stateReader.HeaderGetNextConsensus);

	stateReader.Register("Neo.Block.GetTransactionCount", stateReader.BlockGetTransactionCount);
	stateReader.Register("Neo.Block.GetTransactions", stateReader.BlockGetTransactions);
	stateReader.Register("Neo.Block.GetTransaction", stateReader.BlockGetTransaction);

	stateReader.Register("Neo.Transaction.GetHash", stateReader.TransactionGetHash);
	stateReader.Register("Neo.Transaction.GetType", stateReader.TransactionGetType);
	stateReader.Register("Neo.Transaction.GetAttributes", stateReader.TransactionGetAttributes);
	stateReader.Register("Neo.Transaction.GetInputs", stateReader.TransactionGetInputs);
	stateReader.Register("Neo.Transaction.GetOutputs", stateReader.TransactionGetOutputs);
	stateReader.Register("Neo.Transaction.GetReferences", stateReader.TransactionGetReferences);

	stateReader.Register("Neo.Attribute.GetUsage", stateReader.AttributeGetUsage);
	stateReader.Register("Neo.Attribute.GetData", stateReader.AttributeGetData);

	stateReader.Register("Neo.Input.GetHash", stateReader.InputGetHash);
	stateReader.Register("Neo.Input.GetIndex", stateReader.InputGetIndex);

	stateReader.Register("Neo.Output.GetAssetId", stateReader.OutputGetAssetId);
	stateReader.Register("Neo.Output.GetValue", stateReader.OutputGetValue);
	stateReader.Register("Neo.Output.GetScriptHash", stateReader.OutputGetCodeHash);

	stateReader.Register("Neo.Account.GetScriptHash", stateReader.AccountGetCodeHash);
	stateReader.Register("Neo.Account.GetBalance", stateReader.AccountGetBalance);

	stateReader.Register("Neo.Asset.GetAssetId", stateReader.AssetGetAssetId);
	stateReader.Register("Neo.Asset.GetAssetType", stateReader.AssetGetAssetType);
	stateReader.Register("Neo.Asset.GetAmount", stateReader.AssetGetAmount);
	stateReader.Register("Neo.Asset.GetAvailable", stateReader.AssetGetAvailable);
	stateReader.Register("Neo.Asset.GetPrecision", stateReader.AssetGetPrecision);
	stateReader.Register("Neo.Asset.GetOwner", stateReader.AssetGetOwner);
	stateReader.Register("Neo.Asset.GetAdmin", stateReader.AssetGetAdmin);
	stateReader.Register("Neo.Asset.GetIssuer", stateReader.AssetGetIssuer);

	stateReader.Register("Neo.Contract.GetScript", stateReader.ContractGetCode);

	stateReader.Register("Neo.Storage.GetContext", stateReader.StorageGetContext);

	return &stateReader
}


func (s *StateReader) Register(methodName string, handler func(*avm.ExecutionEngine) (bool, error)) bool {
	if _, ok := s.serviceMap[methodName]; ok {
		return false
	}
	s.serviceMap[methodName] = handler
	return true
}

func (s *StateReader) GetServiceMap() map[string]func(*avm.ExecutionEngine) (bool, error) {
	return s.serviceMap
}

func (s *StateReader) BlockChainGetHeight(e *avm.ExecutionEngine) (bool, error) {
	var i uint32
	if ledger.DefaultLedger == nil {
		i = 0
	}else {
		log.Error("[BlockChainGetHeight] DefaultLedger Store:", i)
		i = ledger.DefaultLedger.Store.GetHeight()
	}
	log.Error("[BlockChainGetHeight] height:", i)
	avm.PushData(e, i)
	return true, nil
}

func (s *StateReader) BlockChainGetHeader(e *avm.ExecutionEngine) (bool, error) {
	data := avm.PopByteArray(e)
	var (
		header *ledger.Header
		err error
	)
	l := len(data)
	if l <= 5 {
		b := new(big.Int)
		log.Error("[BlockChainGetHeight] data:", data)
		height := uint32(b.SetBytes(common.ToArrayReverse(data)).Int64())
		log.Error("[BlockChainGetHeight] height:", height)
		if ledger.DefaultLedger != nil {
			hash, err := ledger.DefaultLedger.Store.GetBlockHash(height)
			if err != nil { return false, err }
			header, err = ledger.DefaultLedger.Store.GetHeader(hash)
			if err != nil { return false, err }
		}else {
			header = nil
		}
	}else if l == 32 {
		hash, _ := common.Uint256ParseFromBytes(data)
		if ledger.DefaultLedger != nil {
			header, err = ledger.DefaultLedger.Store.GetHeader(hash)
			if err != nil { return false, err }
		}else {
			header = nil
		}
	}else {
		return false, errors.New("The data length is error in function blockchaningetheader!")
	}
	log.Error("[BlockChainGetHeader] header:", header)
	avm.PushData(e, header)
	return true, nil
}

func (s *StateReader) BlockChainGetBlock(e *avm.ExecutionEngine) (bool, error) {
	data := avm.PopByteArray(e)
	var (
		block *ledger.Block
	)
	l := len(data)
	if l <= 5 {
		b := new(big.Int)
		height := uint32(b.SetBytes(common.ToArrayReverse(data)).Int64())
		if ledger.DefaultLedger != nil {
			hash, err := ledger.DefaultLedger.Store.GetBlockHash(height)
			if err != nil { return false, err }
			block, err = ledger.DefaultLedger.Store.GetBlock(hash)
			if err != nil { return false, err }
		}else {
			block = nil
		}
	}else if l == 32 {
		hash, err := common.Uint256ParseFromBytes(data)
		if err != nil { return false, err }
		if ledger.DefaultLedger != nil {
			block, err = ledger.DefaultLedger.Store.GetBlock(hash)
			if err != nil { return false, err }
		}else {
			block = nil
		}
	}else {
		return false, errors.New("The data length is error in function blockchaningetblock!")
	}
	avm.PushData(e, block)
	return true, nil
}

func (s *StateReader) BlockChainGetTransaction(e *avm.ExecutionEngine) (bool, error) {
	d := avm.PopByteArray(e)
	hash, err := common.Uint256ParseFromBytes(d)
	if err != nil {
		return false, err
	}
	tx, err := ledger.DefaultLedger.Store.GetTransaction(hash)
	if err != nil {
		return false, err
	}

	avm.PushData(e, tx)
	return true, nil
}

func (s *StateReader) BlockChainGetAccount(e *avm.ExecutionEngine) (bool, error) {
	d := avm.PopByteArray(e)
	hash, err := common.Uint160ParseFromBytes(d)
	if err != nil { return false, err }
	account, err := ledger.DefaultLedger.Store.GetAccount(hash)
	avm.PushData(e, account)
	return true, nil
}

func (s *StateReader) BlockChainGetAsset(e *avm.ExecutionEngine) (bool, error) {
	d := avm.PopByteArray(e)
	hash, err := common.Uint256ParseFromBytes(d)
	if err != nil { return false, err }
	asset, err := ledger.DefaultLedger.Store.GetAsset(hash)
	if err != nil { return false, err }
	avm.PushData(e, asset)
	return true, nil
}

func (s *StateReader) BlockChainGetValidators(e *avm.ExecutionEngine) (bool, error) {
	return true, nil
}

func (s *StateReader) HeaderGetHash(e *avm.ExecutionEngine) (bool, error) {
	d := avm.PopInteropInterface(e)
	if d == nil {
		return false, fmt.Errorf("%v", "Get header error in function headergethash!", )
	}
   	h := d.(*ledger.Header).Blockdata.Hash()
	avm.PushData(e, h.ToArray())
	return true, nil
}

func (s *StateReader) HeaderGetVersion(e *avm.ExecutionEngine) (bool, error) {
	d := avm.PopInteropInterface(e)
	if d == nil {
		return false, fmt.Errorf("%v", "Get header error in function headergetversion")
	}
	version := d.(*ledger.Header).Blockdata.Version
	avm.PushData(e, version)
	return true, nil
}

func (s *StateReader) HeaderGetPrevHash(e *avm.ExecutionEngine) (bool, error) {
	d := avm.PopInteropInterface(e)
	if d == nil {
		return false, fmt.Errorf("%v", "Get header error in function HeaderGetPrevHash")
	}
	preHash := d.(*ledger.Header).Blockdata.PrevBlockHash
	avm.PushData(e, preHash.ToArray())
	return true, nil
}

func (s *StateReader) HeaderGetMerkleRoot(e *avm.ExecutionEngine) (bool, error) {
	d := avm.PopInteropInterface(e)
	if d == nil {
		return false, fmt.Errorf("%v", "Get header error in function HeaderGetMerkleRoot")
	}
	root := d.(*ledger.Header).Blockdata.TransactionsRoot
	avm.PushData(e, root.ToArray())
	return true, nil
}

func (s *StateReader) HeaderGetTimestamp(e *avm.ExecutionEngine) (bool, error) {
	d := avm.PopInteropInterface(e)
	if d == nil {
		return false, fmt.Errorf("%v", "Get header error in function HeaderGetTimestamp")
	}
	timeStamp := d.(*ledger.Header).Blockdata.Timestamp
	avm.PushData(e, timeStamp)
	return true, nil
}

func (s *StateReader) HeaderGetConsensusData(e *avm.ExecutionEngine) (bool, error) {
	d := avm.PopInteropInterface(e)
	if d == nil {
		return false, fmt.Errorf("%v", "Get header error in function HeaderGetConsensusData")
	}
	consensusData := d.(*ledger.Header).Blockdata.ConsensusData
	avm.PushData(e, consensusData)
	return true, nil
}

func (s *StateReader) HeaderGetNextConsensus(e *avm.ExecutionEngine) (bool, error) {
	d := avm.PopInteropInterface(e)
	if d == nil {
		return false, fmt.Errorf("%v", "Get header error in function HeaderGetNextConsensus")
	}
	nextBookKeeper := d.(*ledger.Header).Blockdata.NextBookKeeper
	avm.PushData(e, nextBookKeeper)
	return true, nil
}

func (s *StateReader) BlockGetTransactionCount(e *avm.ExecutionEngine) (bool, error) {
	d := avm.PopInteropInterface(e)
	if d == nil {
		return false, fmt.Errorf("%v", "Get block error in function BlockGetTransactionCount")
	}
	transactions := d.(*ledger.Block).Transactions
	avm.PushData(e, len(transactions))
	return true, nil
}

func (s *StateReader) BlockGetTransactions(e *avm.ExecutionEngine) (bool, error) {
	fmt.Println("[BlockGetTransactions]")
	d := avm.PopInteropInterface(e)
	fmt.Println("[BlockGetTransactions] data", d)
	if d == nil {
		return false, fmt.Errorf("%v", "Get block data error in function BlockGetTransactions")
	}
	transactions := d.(*ledger.Block).Transactions
	transactionList := make([]types.StackItemInterface, 0)
	fmt.Println("================len transactions==============", len(transactions))
	for _, v := range transactions {
		transactionList = append(transactionList, types.NewInteropInterface(v))
	}
	fmt.Println("================transactionList==============", transactionList)
	avm.PushData(e, transactionList)
	return true, nil
}

func (s *StateReader) BlockGetTransaction(e *avm.ExecutionEngine) (bool, error) {
	d := avm.PopInteropInterface(e)
	if d == nil {
		return false, fmt.Errorf("%v", "Get block data error in function BlockGetTransaction")
	}
	transactions := d.(*ledger.Block).Transactions
	index := avm.PopInt(e)
	avm.PushData(e, transactions[index])
	return true, nil
}

func (s *StateReader) TransactionGetHash(e *avm.ExecutionEngine) (bool, error) {
	d := avm.PopInteropInterface(e)
	if d == nil {
		return false, fmt.Errorf("%v", "Get transaction error in function TransactionGetHash")
	}
	txHash := d.(*transaction.Transaction).Hash()
	avm.PushData(e, txHash.ToArray())
	return true, nil
}

func (s *StateReader) TransactionGetType(e *avm.ExecutionEngine) (bool, error) {
	d := avm.PopInteropInterface(e)
	if d == nil {
		return false, fmt.Errorf("%v", "Get transaction error in function TransactionGetType")
	}
	txType := d.(*transaction.Transaction).TxType
	avm.PushData(e, int(txType))
	return true, nil
}

func (s *StateReader) TransactionGetAttributes(e *avm.ExecutionEngine) (bool, error) {
	return  true, nil
}

func (s *StateReader) TransactionGetInputs(e *avm.ExecutionEngine) (bool, error) {
	d := avm.PopInteropInterface(e)
	if d == nil {
		return false, fmt.Errorf("%v", "Get transaction error in function TransactionGetInputs")
	}
	inputs := d.(*transaction.Transaction).UTXOInputs
	inputList := make([]types.StackItemInterface, 0)
	for _, v := range inputs {
		inputList = append(inputList, types.NewInteropInterface(v))
	}
	avm.PushData(e, inputList)
	return true, nil
}

func (s *StateReader) TransactionGetOutputs(e *avm.ExecutionEngine) (bool, error) {
	d := avm.PopInteropInterface(e)
	if d == nil {
		return false, fmt.Errorf("%v", "Get transaction error in function TransactionGetOutputs")
	}
	outputs := d.(*transaction.Transaction).Outputs
	outputList := make([]types.StackItemInterface, 0)
	for _, v := range outputs {
		outputList = append(outputList, types.NewInteropInterface(v))
	}
	avm.PushData(e, outputList)
	return true, nil
}

func (s *StateReader) TransactionGetReferences(e *avm.ExecutionEngine) (bool, error) {
	d := avm.PopInteropInterface(e)
	if d == nil {
		return false, fmt.Errorf("%v", "Get transaction error in function TransactionGetReferences")
	}
	references, err := d.(*transaction.Transaction).GetReference()
	referenceList := make([]types.StackItemInterface, 0)
	for _, v := range references {
		referenceList = append(referenceList, types.NewInteropInterface(v))
	}
	avm.PushData(e, referenceList)
	if err != nil {return false, err}
	return true, nil
}

func (s *StateReader) AttributeGetUsage(e *avm.ExecutionEngine) (bool, error) {
	d := avm.PopInteropInterface(e)
	if d == nil {
		return false, fmt.Errorf("%v", "Get Attribute error in function AttributeGetUsage")
	}
	attribute := d.(*transaction.TxAttribute)
	avm.PushData(e, int(attribute.Usage))
	return true, nil
}

func (s *StateReader) AttributeGetData(e *avm.ExecutionEngine) (bool, error) {
	d := avm.PopInteropInterface(e)
	if d == nil {
		return false, fmt.Errorf("%v", "Get Attribute error in function AttributeGetUsage")
	}
	attribute := d.(*transaction.TxAttribute)
	avm.PushData(e, attribute.Data)
	return true, nil
}

func (s *StateReader) InputGetHash(e *avm.ExecutionEngine) (bool, error) {
	d := avm.PopInteropInterface(e)
	if d == nil {
		return false, fmt.Errorf("%v", "Get UTXOTxInput error in function InputGetHash")
	}
	input := d.(*transaction.UTXOTxInput)
	avm.PushData(e, input.ReferTxID.ToArray())
	return true, nil
}

func (s *StateReader) InputGetIndex(e *avm.ExecutionEngine) (bool, error) {
	d := avm.PopInteropInterface(e)
	if d == nil {
		return false, fmt.Errorf("%v", "Get transaction error in function TransactionGetReferences")
	}
	input := d.(*transaction.UTXOTxInput)
	avm.PushData(e, input.ReferTxOutputIndex)
	return true, nil
}

func (s *StateReader) OutputGetAssetId(e *avm.ExecutionEngine) (bool, error) {
	d := avm.PopInteropInterface(e)
	if d == nil {
		return false, fmt.Errorf("%v", "Get TxOutput error in function OutputGetAssetId")
	}
	output := d.(*transaction.TxOutput)
	avm.PushData(e, output.AssetID.ToArray())
	return true, nil
}

func (s *StateReader) OutputGetValue(e *avm.ExecutionEngine) (bool, error) {
	d := avm.PopInteropInterface(e)
	if d == nil {
		return false, fmt.Errorf("%v", "Get TxOutput error in function OutputGetValue")
	}
	output := d.(*transaction.TxOutput)
	avm.PushData(e, output.Value.GetData())
	return true, nil
}

func (s *StateReader) OutputGetCodeHash(e *avm.ExecutionEngine) (bool, error) {
	d := avm.PopInteropInterface(e)
	if d == nil {
		return false, fmt.Errorf("%v", "Get TxOutput error in function OutputGetCodeHash")
	}
	output := d.(*transaction.TxOutput)
	avm.PushData(e, output.ProgramHash.ToArray())
	return true, nil
}

func (s *StateReader) AccountGetCodeHash(e *avm.ExecutionEngine) (bool, error) {
	d := avm.PopInteropInterface(e)
	if d == nil {
		return false, fmt.Errorf("%v", "Get AccountState error in function AccountGetCodeHash")
	}
	accountState := d.(*states.AccountState).ProgramHash
	avm.PushData(e, accountState.ToArray())
	return true, nil
}

func (s *StateReader) AccountGetBalance(e *avm.ExecutionEngine) (bool, error) {
	d := avm.PopInteropInterface(e)
	if d == nil {
		return false, fmt.Errorf("%v", "Get AccountState error in function AccountGetCodeHash")
	}
	accountState := d.(*states.AccountState)
	assetIdByte := avm.PopByteArray(e)
	assetId, err := common.Uint256ParseFromBytes(assetIdByte)
	if err != nil {return false, err}
	balance := common.Fixed64(0)
	if v, ok := accountState.Balances[assetId]; ok {
		balance = v
	}
	avm.PushData(e, balance.GetData())
	return true, nil
}

func (s *StateReader) AssetGetAssetId(e *avm.ExecutionEngine) (bool, error) {
	d := avm.PopInteropInterface(e)
	if d == nil {
		return false, fmt.Errorf("%v", "Get AssetState error in function AssetGetAssetId")
	}
	assetState := d.(*states.AssetState)
	avm.PushData(e, assetState.AssetId.ToArray())
	return true, nil
}

func (s *StateReader) AssetGetAssetType(e *avm.ExecutionEngine) (bool, error) {
	d := avm.PopInteropInterface(e)
	if d == nil {
		return false, fmt.Errorf("%v", "Get AssetState error in function AssetGetAssetType")
	}
	assetState := d.(*states.AssetState)
	avm.PushData(e, int(assetState.AssetType))
	return true, nil
}

func (s *StateReader) AssetGetAmount(e *avm.ExecutionEngine) (bool, error) {
	d := avm.PopInteropInterface(e)
	if d == nil {
		return false, fmt.Errorf("%v", "Get AssetState error in function AssetGetAmount")
	}
	assetState := d.(*states.AssetState)
	avm.PushData(e, assetState.Amount.GetData())
	return true, nil
}

func (s *StateReader) AssetGetAvailable(e *avm.ExecutionEngine) (bool, error) {
	d := avm.PopInteropInterface(e)
	if d == nil {
		return false, fmt.Errorf("%v", "Get AssetState error in function AssetGetAvailable")
	}
	assetState := d.(*states.AssetState)
	avm.PushData(e, assetState.Available.GetData())
	return true, nil
}

func (s *StateReader) AssetGetPrecision(e *avm.ExecutionEngine) (bool, error) {
	d := avm.PopInteropInterface(e)
	if d == nil {
		return false, fmt.Errorf("%v", "Get AssetState error in function AssetGetPrecision")
	}
	assetState := d.(*states.AssetState)
	avm.PushData(e, int(assetState.Precision))
	return true, nil
}

func (s *StateReader) AssetGetOwner(e *avm.ExecutionEngine) (bool, error) {
	d := avm.PopInteropInterface(e)
	if d == nil {
		return false, fmt.Errorf("%v", "Get AssetState error in function AssetGetOwner")
	}
	assetState := d.(*states.AssetState)
	owner, err := assetState.Owner.EncodePoint(true)
	if err != nil {return false, err}
	avm.PushData(e, owner)
	return true, nil
}

func (s *StateReader) AssetGetAdmin(e *avm.ExecutionEngine) (bool, error) {
	d := avm.PopInteropInterface(e)
	if d == nil {
		return false, fmt.Errorf("%v", "Get AssetState error in function AssetGetAdmin")
	}
	assetState := d.(*states.AssetState)
	avm.PushData(e, assetState.Admin.ToArray())
	return true, nil
}

func (s *StateReader) AssetGetIssuer(e *avm.ExecutionEngine) (bool, error) {
	d := avm.PopInteropInterface(e)
	if d == nil {
		return false, fmt.Errorf("%v", "Get AssetState error in function AssetGetIssuer")
	}
	assetState := d.(*states.AssetState)
	avm.PushData(e, assetState.Issuer.ToArray())
	return true, nil
}

func (s *StateReader) ContractGetCode(e *avm.ExecutionEngine) (bool, error) {
	d := avm.PopInteropInterface(e)
	if d == nil {
		return false, fmt.Errorf("%v", "Get ContractState error in function ContractGetCode")
	}
	assetState := d.(*states.ContractState)
	avm.PushData(e, assetState.Code.Code)
	return true, nil
}

func (s *StateReader) StorageGetContext(e *avm.ExecutionEngine) (bool, error) {
	codeHash, err := common.Uint160ParseFromBytes(e.CurrentContext().GetCodeHash())
	if err != nil {return false, err}
	avm.PushData(e, NewStorageContext(&codeHash))
	return true, nil
}
