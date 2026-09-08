package memo

import (
	"unicode/utf8"

	solana "github.com/fluxrpc/solana-go"
)

type Memo struct {
	Message []byte

	solana.AccountMetaSlice
	programID solana.PublicKey
}

func NewMemoInstruction(message []byte, signers ...solana.PublicKey) *Memo {
	inst := &Memo{
		Message:          message,
		AccountMetaSlice: make(solana.AccountMetaSlice, 0, len(signers)),
		programID:        ProgramID,
	}
	for _, signer := range signers {
		inst.Append(signer.Meta().SIGNER())
	}
	return inst
}

func (inst *Memo) SetProgramID(programID solana.PublicKey) *Memo {
	inst.programID = programID
	return inst
}

func (inst *Memo) ProgramID() solana.PublicKey {
	if inst.programID.IsZero() {
		return ProgramID
	}
	return inst.programID
}

func (inst *Memo) Accounts() []*solana.AccountMeta { return inst.AccountMetaSlice }

func (inst *Memo) Data() ([]byte, error) {
	if !utf8.Valid(inst.Message) {
		return nil, ErrInvalidUTF8
	}
	return inst.Message, nil
}

func (inst *Memo) String() string { return string(inst.Message) }
