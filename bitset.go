package akara

import "math"

const (
	shift6   = 6
	overflow = 1 << shift6
	mask     = overflow - 1 // 0x3f // 0b111111 // 63
)

func NewBitSet(indicesToSetTrue ...int) *BitSet {
	bs := &BitSet{}

	for _, index := range indicesToSetTrue {
		bs.Set(index, true)
	}

	return bs
}

// BitSet is used for storing lots of bits.
type BitSet struct {
	groups []uint64
}

// Get will return the bit value at the given index in the entire bitset.
func (bs *BitSet) Get(idx int) bool {
	// suppose we were storing uint8 instead of uint64:
	// let groups = [0b00000000, 0b11111111]
	//                        ^      ^    ^
	//                        |      |    |
	// bitset.Get(0) ---------+      |    |
	// bitset.Get(13) ---------------+    |
	// bitset.Get(8) ---------------------+
	groupIdx, numGroups := idx>>shift6, len(bs.groups)

	if groupIdx >= numGroups {
		return false
	}

	return getBit(bs.groups[groupIdx], uint64(idx)&mask)
}

// GetAndClear will return the bit value at the given index in the entire bitset, and clear
// it in one step
func (bs *BitSet) GetAndClear(idx uint64) bool {
	groupIdx, numGroups := int(idx>>shift6), len(bs.groups)

	if groupIdx >= numGroups {
		return false
	}

	val := getBit(bs.groups[groupIdx], idx&mask)

	bs.groups[groupIdx] = clearBit(bs.groups[groupIdx], idx&mask)

	return val
}

// GetAndSet will return the bit value at the given index in the entire bitset, and set
// it in one step
func (bs *BitSet) GetAndSet(idx uint64) bool {
	groupIdx, numGroups := int(idx>>shift6), len(bs.groups)

	if groupIdx >= numGroups {
		return false
	}

	val := getBit(bs.groups[groupIdx], idx&mask)

	bs.groups[groupIdx] = setBit(bs.groups[groupIdx], idx&mask)

	return val
}

// Flip inverts the bit at the given index in the bitset
func (bs *BitSet) Flip(idx int) {
	bs.Set(idx, !bs.Get(idx))
}

// Set will set the bit value at the given index in the bitset
func (bs *BitSet) Set(idx int, b bool) {
	bs.assign(idx, b)
}

// Clear will clear all of the bits in the bitset
func (bs *BitSet) Clear() {
	bs.groups = make([]uint64, 0)
}

// Clone creates a copy of the BitSet
func (bs *BitSet) Clone() *BitSet {
	clone := &BitSet{groups: make([]uint64, len(bs.groups))}

	copy(clone.groups, bs.groups)

	return clone
}

// Invert all of the bits in the bitset (mutates the bitset!)
func (bs *BitSet) Invert() *BitSet {
	for groupIdx := range bs.groups {
		bs.groups[groupIdx] = ^bs.groups[groupIdx]
	}

	return bs
}

// And performs a bitwise AND against the argument bitset (mutates the bitset!)
func (bs *BitSet) And(other *BitSet) *BitSet {
	commonLength := int(math.Min(float64(len(bs.groups)), float64(len(other.groups))))

	for idx := commonLength - 1; idx >= 0; idx-- {
		bs.groups[idx] &= other.groups[idx]
	}

	return bs
}

// Or performs a bitwise OR against the argument bitset (mutates the bitset!)
func (bs *BitSet) Or(other *BitSet) *BitSet {
	commonLength := int(math.Min(float64(len(bs.groups)), float64(len(other.groups))))

	for idx := commonLength - 1; idx >= 0; idx-- {
		bs.groups[idx] |= other.groups[idx]
	}

	return bs
}

// Xor performs a bitwise XOR against the argument bitset (mutates the bitset!)
func (bs *BitSet) Xor(other *BitSet) *BitSet {
	commonLength := int(math.Min(float64(len(bs.groups)), float64(len(other.groups))))

	for idx := commonLength - 1; idx >= 0; idx-- {
		bs.groups[idx] ^= other.groups[idx]
	}

	return bs
}

// Empty returns false if any bit is set in the BitSet
func (bs *BitSet) Empty() bool {
	for idx := range bs.groups {
		if bs.groups[idx] > 0 {
			return false
		}
	}

	return true
}

// NotEmpty returns true if any bit is set in the BitSet
func (bs *BitSet) NotEmpty() bool {
	return !bs.Empty()
}

// ContainsAll returns true if this bitset contains all flags present in the other bitset
func (bs *BitSet) ContainsAll(other *BitSet) bool {
	if other.Empty() {
		return true
	}

	if bs.Empty() {
		return false
	}

	var numLowestGroups int

	numGroups, numOtherGroups := len(bs.groups), len(other.groups)

	if numGroups > numOtherGroups {
		numLowestGroups = numOtherGroups
	} else {
		numLowestGroups = numGroups
	}

	// if there are set bits in other which this bitset does not have, this
	// bitset cannot be a super set of other
	for idx := numGroups; idx < numOtherGroups; idx++ {
		if other.groups[idx] != 0 {
			return false
		}
	}

	// checking only the bit groups necessary, we AND a bit groups with the corresponding
	// bit groups from other. If the result no longer equals the other bit group then we know
	// to return false
	for idx := numLowestGroups - 1; idx >= 0; idx-- {
		if (bs.groups[idx] & other.groups[idx]) != other.groups[idx] {
			return false
		}
	}

	return true
}

// Intersects returns true if this bitset contains any of the bits from the other bitset
func (bs *BitSet) Intersects(other *BitSet) bool {
	if bs.Empty() {
		return false
	}

	var numLowestGroups int

	numGroups, numOtherGroups := len(bs.groups), len(other.groups)

	if numGroups > numOtherGroups {
		numLowestGroups = numOtherGroups
	} else {
		numLowestGroups = numGroups
	}

	// checking only the bit groups necessary, we AND a bit groups with the corresponding
	// bit groups from other. If the result is non-zero, this bitset shares bits with other
	for idx := numLowestGroups - 1; idx >= 0; idx-- {
		if (bs.groups[idx] & other.groups[idx]) != 0 {
			return true
		}
	}

	return false
}

// Equals tests if this bitset and the argument bitset are identical
func (bs *BitSet) Equals(other *BitSet) bool {
	if (bs == nil && other != nil) || (other == nil && bs != nil) {
		return false
	}

	if bs == nil && other == nil {
		return true
	}

	return bs.ContainsAll(other) && other.ContainsAll(bs)
}

// ToIntArray returns an array of all bit indices which are true
func (bs *BitSet) ToIntArray() []uint64 {
	result := make([]uint64, 0)

	numGroups := len(bs.groups)
	for groupIdx := 0; groupIdx < numGroups; groupIdx++ {
		for bitIdx := 0; bitIdx < overflow; bitIdx++ {
			if (bs.groups[groupIdx]>>bitIdx)&0b1 > 0 {
				result = append(result, uint64((groupIdx*overflow)+bitIdx))
			}
		}
	}

	return result
}

func (bs *BitSet) assign(idx int, b bool) {
	groupIdx, numGroups := idx>>shift6, len(bs.groups)

	if groupIdx >= numGroups {
		bs.extend(groupIdx + 1 - numGroups)
	}

	bits, pos := bs.groups[groupIdx], uint64(idx&mask)

	if b {
		bs.groups[groupIdx] = setBit(bits, pos)
	} else {
		bs.groups[groupIdx] = clearBit(bits, pos)
	}
}

func (bs *BitSet) extend(amount int) {
	bs.groups = append(bs.groups, make([]uint64, amount)...)
}

// Sets the bit at pos in the integer n.
func setBit(n, pos uint64) uint64 {
	return n | (1 << pos)
}

// Clears the bit at pos in n.
func clearBit(n, pos uint64) uint64 {
	return n & ^(uint64(1) << pos)
}

func getBit(n, pos uint64) bool {
	return (n & (1 << pos)) > 0
}
