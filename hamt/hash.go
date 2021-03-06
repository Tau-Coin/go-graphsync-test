package hamt

import (
	"fmt"

	"github.com/spaolacci/murmur3"
)

// hashBits is a helper that allows the reading of the 'next n bits' as an integer.
type HashBits struct {
	B        []byte
	consumed int
}

func mkmask(n int) byte {
	return (1 << uint(n)) - 1
}

// Next returns the next 'i' bits of the hashBits value as an integer, or an
// error if there aren't enough bits.
func (hb *HashBits) Next(i int) (int, error) {
	if hb.consumed+i > len(hb.B)*8 {
		return 0, fmt.Errorf("sharded directory too deep")
	}
	return hb.next(i), nil
}

func (hb *HashBits) next(i int) int {
	curbi := hb.consumed / 8
	leftb := 8 - (hb.consumed % 8)

	curb := hb.B[curbi]
	if i == leftb {
		out := int(mkmask(i) & curb)
		hb.consumed += i
		return out
	} else if i < leftb {
		a := curb & mkmask(leftb) // mask out the high bits we don't want
		b := a & ^mkmask(leftb-i) // mask out the low bits we don't want
		c := b >> uint(leftb-i)   // shift whats left down
		hb.consumed += i
		return int(c)
	} else {
		out := int(mkmask(leftb) & curb)
		out <<= uint(i - leftb)
		hb.consumed += leftb
		out += hb.next(i - leftb)
		return out
	}
}

var Hash = func(val string) []byte {
	h := murmur3.New64()
	h.Write([]byte(val))
	return h.Sum(nil)
}
