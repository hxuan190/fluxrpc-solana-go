package benchcmp

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"math"
	"testing"

	fluxbin "github.com/fluxrpc/solana-go/binary"
	gaglbin "github.com/gagliardetto/binary"
)

const Uint128Bytes = 16

func u128Fixtures() []struct {
	name   string
	lo, hi uint64
} {
	return []struct {
		name   string
		lo, hi uint64
	}{
		{"zero", 0, 0},
		{"lo_one", 1, 0},
		{"hi_one", 0, 1},
		{"both", 42, 7},
		{"lo_max", math.MaxUint64, 0},
		{"hi_max", 0, math.MaxUint64},
		{"max", math.MaxUint64, math.MaxUint64},
		{"sign_bit", 0, 1 << 63},
		{"int_max", math.MaxUint64, math.MaxUint64 >> 1},
		{"liquidity", 0xDEADBEEFCAFEBABE, 0x0123456789ABCDEF},
	}
}

func u128Wire(lo, hi uint64) []byte {
	buf := make([]byte, 16)
	binary.LittleEndian.PutUint64(buf[:8], lo)
	binary.LittleEndian.PutUint64(buf[8:], hi)
	return buf
}

func TestUint128DecodeParity(t *testing.T) {
	for _, f := range u128Fixtures() {
		t.Run(f.name, func(t *testing.T) {
			data := u128Wire(f.lo, f.hi)

			flux := fluxbin.NewDecoder(data).ReadUint128()
			gagl, err := gaglbin.NewBinDecoder(data).ReadUint128(binary.LittleEndian)
			if err != nil {
				t.Fatal(err)
			}

			if flux.Lo != gagl.Lo || flux.Hi != gagl.Hi {
				t.Fatalf("flux{Lo:%d Hi:%d} != gagl{Lo:%d Hi:%d}", flux.Lo, flux.Hi, gagl.Lo, gagl.Hi)
			}
			if flux.BigInt().Cmp(gagl.BigInt()) != 0 {
				t.Fatalf("BigInt: flux %s, gagl %s", flux.BigInt(), gagl.BigInt())
			}
			if flux.String() != gagl.String() {
				t.Fatalf("String: flux %s, gagl %s", flux.String(), gagl.String())
			}
		})
	}
}

func TestUint128EncodeParity(t *testing.T) {
	for _, f := range u128Fixtures() {
		t.Run(f.name, func(t *testing.T) {
			enc := fluxbin.NewEncoder(nil)
			enc.WriteUint128(fluxbin.Uint128{Lo: f.lo, Hi: f.hi})
			if err := enc.Err(); err != nil {
				t.Fatal(err)
			}

			var buf bytes.Buffer
			gaglEnc := gaglbin.NewBinEncoder(&buf)
			if err := gaglEnc.Encode(gaglbin.Uint128{Lo: f.lo, Hi: f.hi}); err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(enc.Bytes(), buf.Bytes()) {
				t.Fatalf("flux %x != gagl %x", enc.Bytes(), buf.Bytes())
			}
			if !bytes.Equal(enc.Bytes(), u128Wire(f.lo, f.hi)) {
				t.Fatalf("flux %x is not little-endian lo||hi", enc.Bytes())
			}
		})
	}
}

func TestUint128JSONParity(t *testing.T) {
	for _, f := range u128Fixtures() {
		t.Run(f.name, func(t *testing.T) {
			fluxJSON, err := json.Marshal(fluxbin.Uint128{Lo: f.lo, Hi: f.hi})
			if err != nil {
				t.Fatal(err)
			}
			gaglJSON, err := json.Marshal(gaglbin.Uint128{Lo: f.lo, Hi: f.hi})
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(fluxJSON, gaglJSON) {
				t.Fatalf("flux %s != gagl %s", fluxJSON, gaglJSON)
			}

			var back fluxbin.Uint128
			if err := json.Unmarshal(gaglJSON, &back); err != nil {
				t.Fatal(err)
			}
			if back.Lo != f.lo || back.Hi != f.hi {
				t.Fatalf("flux read gagl JSON as {Lo:%d Hi:%d}", back.Lo, back.Hi)
			}
		})
	}
}

func TestInt128BigIntParity(t *testing.T) {
	for _, f := range u128Fixtures() {
		t.Run(f.name, func(t *testing.T) {
			data := u128Wire(f.lo, f.hi)

			flux := fluxbin.NewDecoder(data).ReadInt128()
			gagl, err := gaglbin.NewBinDecoder(data).ReadInt128(binary.LittleEndian)
			if err != nil {
				t.Fatal(err)
			}
			if flux.BigInt().Cmp(gagl.BigInt()) != 0 {
				t.Fatalf("BigInt: flux %s, gagl %s", flux.BigInt(), gagl.BigInt())
			}
		})
	}
}

func TestInt128TextIsSignedUnlikeGagliardetto(t *testing.T) {
	negative := u128Wire(math.MaxUint64, math.MaxUint64)

	flux := fluxbin.NewDecoder(negative).ReadInt128()
	gagl, err := gaglbin.NewBinDecoder(negative).ReadInt128(binary.LittleEndian)
	if err != nil {
		t.Fatal(err)
	}

	if flux.BigInt().String() != "-1" || gagl.BigInt().String() != "-1" {
		t.Fatalf("BigInt: flux %s, gagl %s; both must be -1", flux.BigInt(), gagl.BigInt())
	}

	if flux.String() != "-1" {
		t.Fatalf("flux String() = %s, want -1 to match its own BigInt", flux.String())
	}
	if gagl.String() != "340282366920938463463374607431768211455" {
		t.Fatalf("gagl String() = %s; expected the unsigned reading it has always emitted", gagl.String())
	}

	fluxJSON, err := json.Marshal(flux)
	if err != nil {
		t.Fatal(err)
	}
	if string(fluxJSON) != `"-1"` {
		t.Fatalf("flux JSON = %s, want \"-1\"", fluxJSON)
	}
}

func TestInt128JSONRoundTrip(t *testing.T) {
	for _, f := range u128Fixtures() {
		t.Run(f.name, func(t *testing.T) {
			val := fluxbin.Int128{Lo: f.lo, Hi: f.hi}
			data, err := json.Marshal(val)
			if err != nil {
				t.Fatal(err)
			}
			var back fluxbin.Int128
			if err := json.Unmarshal(data, &back); err != nil {
				t.Fatalf("%s: %v", data, err)
			}
			if back != val {
				t.Fatalf("round trip = %+v, want %+v", back, val)
			}
		})
	}
}

var (
	sinkFluxU128 fluxbin.Uint128
	sinkGaglU128 gaglbin.Uint128
	sinkU128Err  error
	sinkU128Byte []byte
)

func BenchmarkBinary_Uint128(b *testing.B) {
	data := u128Wire(0xDEADBEEFCAFEBABE, 0x0123456789ABCDEF)

	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkFluxU128 = fluxbin.NewDecoder(data).ReadUint128()
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkGaglU128, sinkU128Err = gaglbin.NewBinDecoder(data).ReadUint128(binary.LittleEndian)
		}
		if sinkU128Err != nil {
			b.Fatal(sinkU128Err)
		}
	})
}

func BenchmarkBinaryEncode_Uint128(b *testing.B) {
	value := fluxbin.Uint128{Lo: 0xDEADBEEFCAFEBABE, Hi: 0x0123456789ABCDEF}
	gaglValue := gaglbin.Uint128{Lo: value.Lo, Hi: value.Hi}

	b.Run("flux", func(b *testing.B) {
		dst := make([]byte, 0, Uint128Bytes)
		b.ReportAllocs()
		for b.Loop() {
			enc := fluxbin.NewEncoder(dst[:0])
			enc.WriteUint128(value)
			sinkU128Byte = enc.Bytes()
		}
	})
	b.Run("gagl", func(b *testing.B) {
		var buf bytes.Buffer
		enc := gaglbin.NewBinEncoder(&buf)
		b.ReportAllocs()
		for b.Loop() {
			buf.Reset()
			sinkU128Err = enc.WriteUint128(gaglValue, binary.LittleEndian)
			sinkU128Byte = buf.Bytes()
		}
		if sinkU128Err != nil {
			b.Fatal(sinkU128Err)
		}
	})
	b.Run("gagl_reflect", func(b *testing.B) {
		var buf bytes.Buffer
		enc := gaglbin.NewBinEncoder(&buf)
		b.ReportAllocs()
		for b.Loop() {
			buf.Reset()
			sinkU128Err = enc.Encode(gaglValue)
			sinkU128Byte = buf.Bytes()
		}
		if sinkU128Err != nil {
			b.Fatal(sinkU128Err)
		}
	})
}
