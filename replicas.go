package hexaring

import "math/big"

// calculateReplicaHashBigInt computes r-1 additional hashes for a hash and returns respective int
// values. It divides the keyspace into r slices starting at provided hash.
// func calculateReplicaHashBigInt(hash []byte, bits, r int) []*big.Int {
// 	// Get the ring size
// 	var ringSize big.Int
// 	ringSize.Exp(big.NewInt(2), big.NewInt(int64(bits)), nil)
//
// 	// BigInt replica
// 	replicas := big.NewInt(int64(r))
// 	// Width of each sector
// 	width := new(big.Int).Div(&ringSize, replicas)
// 	// BigInt hash
// 	hashi := new(big.Int).SetBytes(hash)
//
// 	out := make([]*big.Int, r-1)
//
// 	for i := 1; i < r; i++ {
// 		bi := big.NewInt(int64(i))
// 		ti := new(big.Int).Mul(bi, width)
// 		ti.Add(hashi, ti)
//
// 		out[i-1] = new(big.Int).Mod(ti, &ringSize)
// 	}
//
// 	return out
// }

// CalculateRingVertexes returns the requested number of vertexes around the ring
// equi-distant from each other.
func CalculateRingVertexes(hash []byte, count int64) []*big.Int {
	var circum big.Int
	circum.Exp(big.NewInt(2), big.NewInt(int64(len(hash)*8)), nil)

	arcs := big.NewInt(count)
	arcWidth := new(big.Int).Div(&circum, arcs)

	offset := new(big.Int).SetBytes(hash)

	locs := make([]*big.Int, count)
	locs[0] = offset

	for i := int64(1); i < count; i++ {
		bigI := big.NewInt(i)
		piece := new(big.Int).Mul(bigI, arcWidth)
		po := new(big.Int).Add(piece, offset)
		locs[i] = new(big.Int).Mod(po, &circum)
	}

	return locs
}

// CalculateRingVertexBytes returns the a slice of bytes one for each vertex
func CalculateRingVertexBytes(hash []byte, count int64) [][]byte {
	vertexes := CalculateRingVertexes(hash, count)
	out := make([][]byte, len(vertexes))
	for i, v := range vertexes {
		out[i] = v.Bytes()
	}
	return out
}

// ReplicaHashes calculates replica hashes for a given hash.  It takes a hash,
// bit size and number of replicas to calculate and returns that number of
// replica hashes
// func ReplicaHashes(hash []byte, replicas int) [][]byte {
// 	out := make([][]byte, replicas)
//
// 	rsp := calculateReplicaHashBigInt(hash, len(hash)*8, replicas+1)
// 	for i, v := range rsp {
// 		out[i] = v.Bytes()
// 	}
//
// 	return out
// }

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
