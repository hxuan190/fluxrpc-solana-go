package benchcmp

import (
	"bytes"
	"strings"
	"testing"

	flux "github.com/fluxrpc/solana-go"
	fluxmemo "github.com/fluxrpc/solana-go/programs/memo"
	foundation "github.com/solana-foundation/solana-go/v2"
	foundationmemo "github.com/solana-foundation/solana-go/v2/programs/memo"
)

var (
	sinkFluxMemoInstruction       flux.Instruction
	sinkFoundationMemoInstruction foundation.Instruction
	sinkFluxMemoDecoded           *fluxmemo.Memo
	sinkFoundationMemoDecoded     *foundationmemo.MemoInstruction
	sinkMemoData                  []byte
	sinkMemoErr                   error
)

type memoInstructionBenchmark struct {
	name               string
	flux               flux.Instruction
	foundation         foundation.Instruction
	fluxData           []byte
	foundationData     []byte
	fluxAccounts       flux.AccountMetaSlice
	foundationAccounts []*foundation.AccountMeta
	newFlux            func() flux.Instruction
	newFoundation      func() foundation.Instruction
}

func newMemoInstructionBenchmarks(tb testing.TB) []memoInstructionBenchmark {
	tb.Helper()

	fluxSigners := []flux.PublicKey{
		flux.MustPublicKeyFromBase58("9nBUw2LGRdyEBqZCzKXfTvvHPWUCbVAG7RTeGDvbGGDS"),
		flux.MustPublicKeyFromBase58("HxUb9nJXjHrLDsHNwHqTBWkfxaLnbEwJXcYHgqmnbdrx"),
	}
	foundationSigners := make([]foundation.PublicKey, len(fluxSigners))
	for index, signer := range fluxSigners {
		foundationSigners[index] = foundation.PublicKey(signer)
	}

	ascii := []byte("rugcheck payment 4f2c9a")
	unicode := []byte("thanh toán ☀ 4f2c9a")
	long := []byte(strings.Repeat("memo ", 128))

	benchmarks := []memoInstructionBenchmark{
		{
			name:          "ASCII",
			newFlux:       func() flux.Instruction { return fluxmemo.NewMemoInstruction(ascii) },
			newFoundation: func() foundation.Instruction { return foundationmemo.NewMemoInstruction(ascii).Build() },
		},
		{
			name:          "UTF8",
			newFlux:       func() flux.Instruction { return fluxmemo.NewMemoInstruction(unicode) },
			newFoundation: func() foundation.Instruction { return foundationmemo.NewMemoInstruction(unicode).Build() },
		},
		{
			name:          "Long",
			newFlux:       func() flux.Instruction { return fluxmemo.NewMemoInstruction(long) },
			newFoundation: func() foundation.Instruction { return foundationmemo.NewMemoInstruction(long).Build() },
		},
		{
			name:    "TwoSigners",
			newFlux: func() flux.Instruction { return fluxmemo.NewMemoInstruction(ascii, fluxSigners...) },
			newFoundation: func() foundation.Instruction {
				return foundationmemo.NewMemoInstruction(ascii, foundationSigners...).Build()
			},
		},
	}

	for index := range benchmarks {
		entry := &benchmarks[index]
		entry.flux = entry.newFlux()
		entry.foundation = entry.newFoundation()
		var err error
		entry.fluxData, err = entry.flux.Data()
		if err != nil {
			tb.Fatalf("flux %s Data: %v", entry.name, err)
		}
		entry.foundationData, err = entry.foundation.Data()
		if err != nil {
			tb.Fatalf("foundation %s Data: %v", entry.name, err)
		}
		if !bytes.Equal(entry.fluxData, entry.foundationData) {
			tb.Fatalf("%s data differs: flux %x, foundation %x", entry.name, entry.fluxData, entry.foundationData)
		}
		entry.fluxAccounts = entry.flux.Accounts()
		entry.foundationAccounts = entry.foundation.Accounts()
	}
	return benchmarks
}

func BenchmarkMemoConstructorAndData(b *testing.B) {
	for _, test := range newMemoInstructionBenchmarks(b) {
		b.Run(test.name, func(b *testing.B) {
			b.Run("Flux", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkFluxMemoInstruction = test.newFlux()
					sinkMemoData, sinkMemoErr = sinkFluxMemoInstruction.Data()
				}
			})
			b.Run("Foundation", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkFoundationMemoInstruction = test.newFoundation()
					sinkMemoData, sinkMemoErr = sinkFoundationMemoInstruction.Data()
				}
			})
		})
	}
}

func BenchmarkMemoData(b *testing.B) {
	for _, test := range newMemoInstructionBenchmarks(b) {
		b.Run(test.name, func(b *testing.B) {
			b.Run("Flux", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkMemoData, sinkMemoErr = test.flux.Data()
				}
			})
			b.Run("Foundation", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkMemoData, sinkMemoErr = test.foundation.Data()
				}
			})
		})
	}
}

func BenchmarkMemoDecode(b *testing.B) {
	for _, test := range newMemoInstructionBenchmarks(b) {
		b.Run(test.name, func(b *testing.B) {
			b.Run("Flux", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkFluxMemoDecoded, sinkMemoErr = fluxmemo.DecodeInstruction(test.fluxAccounts, test.fluxData)
				}
			})
			b.Run("Foundation", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					sinkFoundationMemoDecoded, sinkMemoErr = foundationmemo.DecodeInstruction(test.foundationAccounts, test.foundationData)
				}
			})
		})
	}
}
