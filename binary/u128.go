package binary

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
)

const Uint128Size = 16

var ErrInvalid128 = errors.New("binary: invalid 128-bit integer")

var (
	twoPow128  = new(big.Int).Lsh(big.NewInt(1), 128)
	maxUint128 = new(big.Int).Sub(twoPow128, big.NewInt(1))
	maxInt128  = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 127), big.NewInt(1))
	minInt128  = new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), 127))
)

type Uint128 struct {
	Lo uint64
	Hi uint64
}

type Int128 Uint128

func (v Uint128) BigInt() *big.Int {
	hi := new(big.Int).SetUint64(v.Hi)
	return hi.Lsh(hi, 64).Or(hi, new(big.Int).SetUint64(v.Lo))
}

func (v Uint128) String() string { return v.BigInt().String() }

func (v Uint128) MarshalJSON() ([]byte, error) {
	return []byte(`"` + v.String() + `"`), nil
}

func (v *Uint128) UnmarshalJSON(data []byte) error {
	parsed, null, err := parse128(data)
	if err != nil || null {
		return err
	}
	if parsed.Sign() < 0 || parsed.Cmp(maxUint128) > 0 {
		return fmt.Errorf("%w: uint128 out of range: %s", ErrInvalid128, parsed)
	}
	*v = from128(parsed)
	return nil
}

func (v Int128) BigInt() *big.Int {
	out := Uint128(v).BigInt()
	if v.Hi&(1<<63) != 0 {
		out.Sub(out, twoPow128)
	}
	return out
}

func (v Int128) String() string { return v.BigInt().String() }

func (v Int128) MarshalJSON() ([]byte, error) {
	return []byte(`"` + v.String() + `"`), nil
}

func (v *Int128) UnmarshalJSON(data []byte) error {
	parsed, null, err := parse128(data)
	if err != nil || null {
		return err
	}
	if parsed.Cmp(minInt128) < 0 || parsed.Cmp(maxInt128) > 0 {
		return fmt.Errorf("%w: int128 out of range: %s", ErrInvalid128, parsed)
	}
	if parsed.Sign() < 0 {
		parsed = new(big.Int).Add(parsed, twoPow128)
	}
	*v = Int128(from128(parsed))
	return nil
}

func parse128(data []byte) (*big.Int, bool, error) {
	s := string(data)
	if s == "null" {
		return nil, true, nil
	}
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}
	parsed, ok := new(big.Int).SetString(s, 0)
	if !ok {
		return nil, false, fmt.Errorf("%w: cannot parse %q", ErrInvalid128, s)
	}
	return parsed, false, nil
}

func from128(v *big.Int) Uint128 {
	var buf [Uint128Size]byte
	v.FillBytes(buf[:])
	return Uint128{
		Hi: binary.BigEndian.Uint64(buf[:8]),
		Lo: binary.BigEndian.Uint64(buf[8:]),
	}
}
