package akara

import "math/bits"

type componentMask []uint64

func (mask *componentMask) set(id uint32, value bool) {
	word := int(id >> 6)
	if !value && word >= len(*mask) {
		return
	}
	if word >= len(*mask) {
		*mask = append(*mask, make([]uint64, word+1-len(*mask))...)
	}
	bit := uint64(1) << (id & 63)
	if value {
		(*mask)[word] |= bit
	} else {
		(*mask)[word] &^= bit
	}
}

func (mask componentMask) containsAll(other componentMask) bool {
	for word, required := range other {
		if required == 0 {
			continue
		}
		if word >= len(mask) || mask[word]&required != required {
			return false
		}
	}
	return true
}

func (mask componentMask) intersects(other componentMask) bool {
	length := len(mask)
	if len(other) < length {
		length = len(other)
	}
	for word := 0; word < length; word++ {
		if mask[word]&other[word] != 0 {
			return true
		}
	}
	return false
}

func (mask componentMask) empty() bool {
	for _, word := range mask {
		if word != 0 {
			return false
		}
	}
	return true
}

func (mask componentMask) eachSet(visit func(uint32)) {
	for wordIndex, word := range mask {
		for word != 0 {
			bit := bits.TrailingZeros64(word)
			visit(uint32(wordIndex*64 + bit))
			word &= word - 1
		}
	}
}

func (mask componentMask) with(id uint32, value bool) componentMask {
	result := append(componentMask(nil), mask...)
	result.set(id, value)
	for len(result) > 0 && result[len(result)-1] == 0 {
		result = result[:len(result)-1]
	}
	return result
}

func (mask componentMask) hash() uint64 {
	const offset = uint64(14695981039346656037)
	const prime = uint64(1099511628211)
	hash := offset
	for _, word := range mask {
		hash ^= word
		hash *= prime
	}
	hash ^= uint64(len(mask))
	return hash * prime
}

func (mask componentMask) equals(other componentMask) bool {
	if len(mask) != len(other) {
		return false
	}
	for index := range mask {
		if mask[index] != other[index] {
			return false
		}
	}
	return true
}
