package hexaring

import (
	"hash"
	"math/big"
)

// CalculateRingVertexes returns the requested number of vertexes around the ring
// equi-distant from each other except for potentially the last one which may be larger
func CalculateRingVertexes(hash []byte, count int64) []*big.Int {
	var circum big.Int
	//circum.SetBytes(maxHash(len(hash)))
	circum.Exp(big.NewInt(2), big.NewInt(int64(len(hash))*8), nil)

	// Number of sections
	arcs := big.NewInt(count)
	// Size of each section
	arcWidth := new(big.Int).Div(&circum, arcs)
	// Starting offset
	offset := new(big.Int).SetBytes(hash)

	locs := make([]*big.Int, count)
	locs[0] = offset

	for i := int64(1); i < count; i++ {
		// Index
		bigI := big.NewInt(i)
		// Index times the width of the section
		piece := new(big.Int).Mul(bigI, arcWidth)
		// Add to offset
		po := new(big.Int).Add(piece, offset)
		locs[i] = new(big.Int).Mod(po, &circum)
	}

	return locs
}

// CalculateRingVertexBytes returns the a slice of bytes one for each vertex
func CalculateRingVertexBytes(hash []byte, count int64) [][]byte {
	// Hash size used when produces hashes are smaller.
	lh := len(hash)
	// Big int values of hashes
	vertexes := CalculateRingVertexes(hash, count)
	// Binary representation
	out := make([][]byte, len(vertexes))

	for i, v := range vertexes {
		b := v.Bytes()
		lb := len(b)
		// Account for modulo.  Padd required zeros
		if lb < lh {
			p := make([]byte, lh-lb)
			b = append(p, b...)
		}
		out[i] = b
	}
	return out
}

// BuildReplicaHashes hashes the given key and build the required additional hashes
// returning the requested count of hashes.
func BuildReplicaHashes(key []byte, count int64, h hash.Hash) [][]byte {
	h.Write(key)
	sh := h.Sum(nil)
	return CalculateRingVertexBytes(sh[:], count)
}

// replicasWithFault returns the number of replicas required in order to tolerate the
// given number of faulty nodes.
func replicasWithFault(faulty int) int {
	return (3 * faulty) + 1
}

// votesWithFault returns the number of votes required for propose in order to tolerate
// the given number of faulty nodes.
func votesWithFault(faulty int) int {
	return (2 * faulty) + 1
}

// commitsWithFault returns the number of commits required in order to tolerate the given
// number of faulty nodes.
func commitsWithFault(faulty int) int {
	return faulty + 1
}

func maxHash(s int) []byte {
	out := make([]byte, s)
	for i := range out {
		out[i] = 0xff
	}
	return out
}
