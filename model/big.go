package model

import (
	"database/sql/driver"
	"fmt"
	"math/big"
)

type BigInt big.Int

func (b *BigInt) Value() (driver.Value, error) {
	if b != nil {
		return (*big.Int)(b).String(), nil
	}
	return nil, nil
}

func (b *BigInt) Valid() bool {
	if b.String() == "" || b == nil {
		return false
	}
	return true
}

func (b *BigInt) Scan(value interface{}) error {
	if value == nil {
		b = nil
	}

	switch t := value.(type) {
	case []uint8:
		_, ok := (*big.Int)(b).SetString(string(value.([]uint8)), 10)
		if !ok {
			return fmt.Errorf("failed to load value to []uint8: %v", value)
		}
	default:
		return fmt.Errorf("could not scan type %T into BigInt", t)
	}
	return nil
}

func (z *BigInt) Set(x *BigInt) *BigInt {
	return (*BigInt)((*big.Int)(z).Set((*big.Int)(x)))
}

func (z *BigInt) SetUint64(x uint64) *BigInt {
	return (*BigInt)((*big.Int)(z).SetInt64(int64(x)))
}

func (z *BigInt) SetString(x string, base int) *BigInt {
	if res, err := (*big.Int)(z).SetString(x, base); err == true {
		return (*BigInt)(res)
	}
	return nil
}

func (x *BigInt) Cmp(y *BigInt) int {
	return (*big.Int)(x).Cmp((*big.Int)(y))
}

func (b *BigInt) String() string {
	return (*big.Int)(b).String()
}

func (b *BigInt) Text(base int) string {
	return (*big.Int)(b).Text(base)
}

func (x *BigInt) MarshalText() (text []byte, err error) {
	return (*big.Int)(x).MarshalText()
}

func (z *BigInt) UnmarshalText(text []byte) error {
	return (*big.Int)(z).UnmarshalText(text)
}

func (x *BigInt) MarshalJSON() ([]byte, error) {
	return x.MarshalText()
}

func (z *BigInt) UnmarshalJSON(text []byte) error {
	return z.UnmarshalText(text)
}
