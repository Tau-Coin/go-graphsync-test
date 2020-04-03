package hamt

import (
        "math/big"
        "math/bits"
)

func IndexForBitPos(bp int, bitfield *big.Int) int {
        var x uint
        var count, i int
        w := bitfield.Bits()
        for x = uint(bp); x > bits.UintSize && i < len(w); x -= bits.UintSize {
                count += bits.OnesCount(uint(w[i]))
                i++
        }
        if i == len(w) {
                return count
        }
        return count + bits.OnesCount(uint(w[i])&((1<<x)-1))
}
