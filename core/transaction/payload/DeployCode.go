package payload

import (
	. "DNA/core/code"
	"DNA/common/serialization"
	"io"
	"DNA/smartcontract/types"
	"DNA/common"
	"fmt"
)

type DeployCode struct {
	Code        *FunctionCode
	Params      []byte
	Name        string
	CodeVersion string
	Author      string
	Email       string
	Description string
	Language    types.LangType
	ProgramHash common.Uint160
}

func (dc *DeployCode) Data() []byte {
	// TODO: Data()

	return []byte{0}
}

func (dc *DeployCode) Serialize(w io.Writer) error {

	err := dc.Code.Serialize(w)
	if err != nil {
		return err
	}

	err = serialization.WriteVarBytes(w, dc.Params)

	if err != nil {
		return err
	}

	err = serialization.WriteVarString(w, dc.Name)
	if err != nil {
		return err
	}

	err = serialization.WriteVarString(w, dc.CodeVersion)
	if err != nil {
		return err
	}

	err = serialization.WriteVarString(w, dc.Author)
	if err != nil {
		return err
	}

	err = serialization.WriteVarString(w, dc.Email)
	if err != nil {
		return err
	}

	err = serialization.WriteVarString(w, dc.Description)
	if err != nil {
		return err
	}

	err = serialization.WriteByte(w, byte(dc.Language))
	if err != nil {
		return err
	}

	_, err = dc.ProgramHash.Serialize(w)
	if err != nil {
		return err
	}
	return nil
}

func (dc *DeployCode) Deserialize(r io.Reader) error {
	dc.Code = new(FunctionCode)
	err := dc.Code.Deserialize(r)
	if err != nil {
		return err
	}

	dc.Params, err = serialization.ReadVarBytes(r)
	if err != nil {
		return err
	}

	dc.Name, err = serialization.ReadVarString(r)
	if err != nil {
		return err
	}

	dc.CodeVersion, err = serialization.ReadVarString(r)
	if err != nil {
		return err
	}

	dc.Author, err = serialization.ReadVarString(r)
	if err != nil {
		return err
	}

	dc.Email, err = serialization.ReadVarString(r)
	if err != nil {
		return err
	}

	dc.Description, err = serialization.ReadVarString(r)
	if err != nil {
		return err
	}

	l, err := serialization.ReadByte(r)
	if err != nil {
		return err
	}
	dc.Language = types.LangType(l)

	u := new(common.Uint160)
	err = u.Deserialize(r)
	if err != nil {
		return err
	}
	dc.ProgramHash = *u
	return nil
}

func (dc *DeployCode) Print() {
	fmt.Println("### deploy print start ###")
	fmt.Println("FunctionCode Code:", dc.Code.Code)
	fmt.Println("FunctionCode ParameterTypes:", dc.Code.ParameterTypes)
	fmt.Println("FunctionCode ReturnType:", dc.Code.ReturnType)
	fmt.Println("FunctionCode CodeHash:", dc.Code.CodeHash())
	fmt.Println("Name:", dc.Name)
	fmt.Println("Name:", dc.Name)
	fmt.Println("CodeVersion:", dc.CodeVersion)
	fmt.Println("Author:", dc.Author)
	fmt.Println("Email:", dc.Email)
	fmt.Println("Description:", dc.Description)
	fmt.Println("Language:", dc.Language)
	fmt.Println("ProgramHash:", dc.ProgramHash)
	fmt.Println("### deploy print end ###")
}