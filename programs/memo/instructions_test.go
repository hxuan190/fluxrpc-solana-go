package memo

import (
	"bytes"
	"errors"
	"testing"

	solana "github.com/fluxrpc/solana-go"
)

func TestMemoInstruction(t *testing.T) {
	var inst solana.Instruction = NewMemoInstruction([]byte("hello ☀"))

	if inst.ProgramID() != ProgramID {
		t.Fatalf("ProgramID = %s, want %s", inst.ProgramID(), ProgramID)
	}
	if len(inst.Accounts()) != 0 {
		t.Fatalf("accounts = %v, want none", inst.Accounts())
	}

	data, err := inst.Data()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, []byte("hello ☀")) {
		t.Fatalf("data = %q, want %q", data, "hello ☀")
	}
}

func TestMemoSigners(t *testing.T) {
	first := solana.NewWallet().PublicKey()
	second := solana.NewWallet().PublicKey()

	inst := NewMemoInstruction([]byte("paid"), first, second)

	accounts := inst.Accounts()
	if len(accounts) != 2 {
		t.Fatalf("accounts = %d, want 2", len(accounts))
	}
	for i, want := range []solana.PublicKey{first, second} {
		if accounts[i].PublicKey != want {
			t.Fatalf("accounts[%d] = %s, want %s", i, accounts[i].PublicKey, want)
		}
		if !accounts[i].IsSigner {
			t.Fatalf("accounts[%d] (%s) should sign", i, want)
		}
		if accounts[i].IsWritable {
			t.Fatalf("accounts[%d] (%s) should not be writable", i, want)
		}
	}
}

func TestMemoRejectsInvalidUTF8(t *testing.T) {
	if _, err := NewMemoInstruction([]byte{0xff, 0xfe}).Data(); !errors.Is(err, ErrInvalidUTF8) {
		t.Fatalf("Data() error = %v, want ErrInvalidUTF8", err)
	}
	if _, err := DecodeInstruction(nil, []byte{0xff, 0xfe}); !errors.Is(err, ErrInvalidUTF8) {
		t.Fatalf("DecodeInstruction error = %v, want ErrInvalidUTF8", err)
	}
}

func TestMemoRoundTrip(t *testing.T) {
	signer := solana.NewWallet().PublicKey()
	inst := NewMemoInstruction([]byte("round trip"), signer)

	data, err := inst.Data()
	if err != nil {
		t.Fatal(err)
	}

	decoded, err := DecodeInstruction(inst.AccountMetaSlice, data)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.String() != "round trip" {
		t.Fatalf("decoded = %q, want %q", decoded.String(), "round trip")
	}
	if len(decoded.Accounts()) != 1 || decoded.Accounts()[0].PublicKey != signer {
		t.Fatalf("decoded accounts = %v, want signer %s", decoded.Accounts(), signer)
	}
}

func TestMemoSetProgramID(t *testing.T) {
	inst := NewMemoInstruction([]byte("legacy")).SetProgramID(ProgramIDV1)
	if inst.ProgramID() != ProgramIDV1 {
		t.Fatalf("ProgramID = %s, want %s", inst.ProgramID(), ProgramIDV1)
	}

	if (&Memo{}).ProgramID() != ProgramID {
		t.Fatalf("zero-value ProgramID = %s, want %s", (&Memo{}).ProgramID(), ProgramID)
	}
}

func TestMemoEmpty(t *testing.T) {
	data, err := NewMemoInstruction(nil).Data()
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 0 {
		t.Fatalf("data = %q, want empty", data)
	}
}

func FuzzDecodeInstruction(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte("hello"))
	f.Add([]byte("hello ☀"))
	f.Add([]byte{0xff, 0xfe})
	f.Add([]byte{0xe2, 0x98})
	f.Add(bytes.Repeat([]byte("a"), 1024))

	f.Fuzz(func(t *testing.T, data []byte) {
		decoded, err := DecodeInstruction(nil, data)
		if err != nil {
			if !errors.Is(err, ErrInvalidUTF8) {
				t.Fatalf("unexpected error: %v", err)
			}
			return
		}

		encoded, err := decoded.Data()
		if err != nil {
			t.Fatalf("re-encode: %v", err)
		}
		if !bytes.Equal(encoded, data) {
			t.Fatalf("re-encoded = %x, want %x", encoded, data)
		}
	})
}
