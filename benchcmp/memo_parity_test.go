package benchcmp

import (
	"bytes"
	"testing"

	fluxmemo "github.com/fluxrpc/solana-go/programs/memo"
	foundationmemo "github.com/solana-foundation/solana-go/v2/programs/memo"
)

func TestMemoInstructionParity(t *testing.T) {
	fixtures := newMemoInstructionBenchmarks(t)
	if len(fixtures) != 4 {
		t.Fatalf("parity fixture count = %d, want 4", len(fixtures))
	}
	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			stakeParityInstruction(t, fixture.flux, fixture.foundation)
		})
	}
}

func TestMemoDecodeParity(t *testing.T) {
	for _, fixture := range newMemoInstructionBenchmarks(t) {
		t.Run(fixture.name, func(t *testing.T) {
			fluxDecoded, err := fluxmemo.DecodeInstruction(fixture.fluxAccounts, fixture.fluxData)
			if err != nil {
				t.Fatalf("flux DecodeInstruction: %v", err)
			}
			foundationDecoded, err := foundationmemo.DecodeInstruction(fixture.foundationAccounts, fixture.foundationData)
			if err != nil {
				t.Fatalf("foundation DecodeInstruction: %v", err)
			}
			if !bytes.Equal(fluxDecoded.Message, foundationDecoded.Impl.(*foundationmemo.Create).Message) {
				t.Errorf("message = %x, foundation %x", fluxDecoded.Message,
					foundationDecoded.Impl.(*foundationmemo.Create).Message)
			}
			if len(fluxDecoded.Accounts()) != len(foundationDecoded.Accounts()) {
				t.Fatalf("account count = %d, foundation %d",
					len(fluxDecoded.Accounts()), len(foundationDecoded.Accounts()))
			}
		})
	}
}
