package memo

import (
	"unicode/utf8"

	solana "github.com/fluxrpc/solana-go"
)

func DecodeInstruction(accounts solana.AccountMetaSlice, data []byte) (*Memo, error) {
	if !utf8.Valid(data) {
		return nil, ErrInvalidUTF8
	}
	return &Memo{
		Message:          data,
		AccountMetaSlice: accounts,
		programID:        ProgramID,
	}, nil
}
