package binary

import (
	"bytes"
	"encoding/json"
	"errors"
	"math"
	"math/big"
	"testing"
)

func TestUint128Wire(t *testing.T) {
	for _, test := range []struct {
		name string
		val  Uint128
		want string
	}{
		{"zero", Uint128{}, "0"},
		{"lo", Uint128{Lo: 1}, "1"},
		{"hi", Uint128{Hi: 1}, "18446744073709551616"},
		{"max", Uint128{Lo: math.MaxUint64, Hi: math.MaxUint64}, "340282366920938463463374607431768211455"},
	} {
		t.Run(test.name, func(t *testing.T) {
			enc := NewEncoder(nil)
			enc.WriteUint128(test.val)
			if err := enc.Err(); err != nil {
				t.Fatal(err)
			}
			data := enc.Bytes()
			if len(data) != Uint128Size {
				t.Fatalf("encoded %d bytes, want %d", len(data), Uint128Size)
			}

			dec := NewDecoder(data)
			got := dec.ReadUint128()
			if err := dec.Err(); err != nil {
				t.Fatal(err)
			}
			if got != test.val {
				t.Fatalf("round trip = %+v, want %+v", got, test.val)
			}
			if got.String() != test.want {
				t.Fatalf("String() = %s, want %s", got.String(), test.want)
			}
		})
	}
}

func TestUint128LittleEndianLayout(t *testing.T) {
	data := []byte{1, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0}
	got := NewDecoder(data).ReadUint128()
	if got.Lo != 1 || got.Hi != 2 {
		t.Fatalf("Lo=%d Hi=%d, want Lo=1 Hi=2", got.Lo, got.Hi)
	}
	want := new(big.Int).Add(new(big.Int).Lsh(big.NewInt(2), 64), big.NewInt(1))
	if got.BigInt().Cmp(want) != 0 {
		t.Fatalf("BigInt = %s, want %s", got.BigInt(), want)
	}
}

func TestInt128Signed(t *testing.T) {
	for _, test := range []struct {
		name string
		val  Int128
		want string
	}{
		{"zero", Int128{}, "0"},
		{"one", Int128{Lo: 1}, "1"},
		{"minus_one", Int128{Lo: math.MaxUint64, Hi: math.MaxUint64}, "-1"},
		{"min", Int128{Hi: 1 << 63}, "-170141183460469231731687303715884105728"},
		{"max", Int128{Lo: math.MaxUint64, Hi: math.MaxUint64 >> 1}, "170141183460469231731687303715884105727"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := test.val.String(); got != test.want {
				t.Fatalf("String() = %s, want %s", got, test.want)
			}
			enc := NewEncoder(nil)
			enc.WriteInt128(test.val)
			if got := NewDecoder(enc.Bytes()).ReadInt128(); got != test.val {
				t.Fatalf("round trip = %+v, want %+v", got, test.val)
			}
		})
	}
}

func TestUint128JSON(t *testing.T) {
	val := Uint128{Lo: 42, Hi: 7}
	data, err := json.Marshal(val)
	if err != nil {
		t.Fatal(err)
	}
	if want := `"` + val.String() + `"`; string(data) != want {
		t.Fatalf("MarshalJSON = %s, want %s", data, want)
	}

	var back Uint128
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatal(err)
	}
	if back != val {
		t.Fatalf("UnmarshalJSON = %+v, want %+v", back, val)
	}
}

func TestInt128JSONNegative(t *testing.T) {
	val := Int128{Lo: math.MaxUint64, Hi: math.MaxUint64}
	data, err := json.Marshal(val)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `"-1"` {
		t.Fatalf("MarshalJSON = %s, want \"-1\"", data)
	}

	var back Int128
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatal(err)
	}
	if back != val {
		t.Fatalf("UnmarshalJSON = %+v, want %+v", back, val)
	}
}

func TestUint128JSONAccepts(t *testing.T) {
	for _, in := range []string{`"255"`, `255`, `"0xff"`} {
		var v Uint128
		if err := json.Unmarshal([]byte(in), &v); err != nil {
			t.Fatalf("%s: %v", in, err)
		}
		if v.Lo != 255 || v.Hi != 0 {
			t.Fatalf("%s: got {Lo:%d Hi:%d}, want 255", in, v.Lo, v.Hi)
		}
	}
}

func TestUint128JSONNullLeavesValue(t *testing.T) {
	v := Uint128{Lo: 7, Hi: 9}
	if err := json.Unmarshal([]byte(`null`), &v); err != nil {
		t.Fatal(err)
	}
	if v.Lo != 7 || v.Hi != 9 {
		t.Fatalf("null changed value to {Lo:%d Hi:%d}", v.Lo, v.Hi)
	}
}

func TestUint128JSONRejects(t *testing.T) {
	for _, in := range []string{`"-1"`, `"340282366920938463463374607431768211456"`, `"nope"`, `""`} {
		var v Uint128
		if err := json.Unmarshal([]byte(in), &v); !errors.Is(err, ErrInvalid128) {
			t.Fatalf("%s: err = %v, want ErrInvalid128", in, err)
		}
	}
}

func TestInt128JSONRejectsOutOfRange(t *testing.T) {
	for _, in := range []string{`"170141183460469231731687303715884105728"`, `"-170141183460469231731687303715884105729"`} {
		var v Int128
		if err := json.Unmarshal([]byte(in), &v); !errors.Is(err, ErrInvalid128) {
			t.Fatalf("%s: err = %v, want ErrInvalid128", in, err)
		}
	}
}

func TestUint128ShortBuffer(t *testing.T) {
	dec := NewDecoder(make([]byte, Uint128Size-1))
	if got := dec.ReadUint128(); got != (Uint128{}) {
		t.Fatalf("value = %+v, want zero", got)
	}
	if !errors.Is(dec.Err(), ErrUnexpectedEOF) {
		t.Fatalf("Err() = %v, want ErrUnexpectedEOF", dec.Err())
	}
}

func FuzzUint128(f *testing.F) {
	f.Add(make([]byte, Uint128Size))
	f.Add(bytes.Repeat([]byte{0xff}, Uint128Size))
	f.Add([]byte{1, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0})

	f.Fuzz(func(t *testing.T, data []byte) {
		dec := NewDecoder(data)
		v := dec.ReadUint128()
		if dec.Err() != nil {
			return
		}

		enc := NewEncoder(nil)
		enc.WriteUint128(v)
		if !bytes.Equal(enc.Bytes(), data[:Uint128Size]) {
			t.Fatalf("re-encoded %x, want %x", enc.Bytes(), data[:Uint128Size])
		}
		if v.BigInt().Sign() < 0 {
			t.Fatalf("uint128 %s is negative", v)
		}

		raw, err := v.MarshalJSON()
		if err != nil {
			t.Fatal(err)
		}
		var back Uint128
		if err := back.UnmarshalJSON(raw); err != nil {
			t.Fatalf("%s: %v", raw, err)
		}
		if back != v {
			t.Fatalf("json round trip = %+v, want %+v", back, v)
		}

		signed := Int128(v)
		rawSigned, err := signed.MarshalJSON()
		if err != nil {
			t.Fatal(err)
		}
		var backSigned Int128
		if err := backSigned.UnmarshalJSON(rawSigned); err != nil {
			t.Fatalf("%s: %v", rawSigned, err)
		}
		if backSigned != signed {
			t.Fatalf("int128 json round trip = %+v, want %+v", backSigned, signed)
		}
	})
}
