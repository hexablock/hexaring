package hexaring

import (
	"bytes"
	"crypto/sha256"
	"log"
	"sort"
	"testing"
)

type sortByteArrays [][]byte

func (b sortByteArrays) Len() int {
	return len(b)
}

func (b sortByteArrays) Less(i, j int) bool {
	// bytes package already implements Comparable for []byte.
	switch bytes.Compare(b[i], b[j]) {
	case -1:
		return true
	case 0, 1:
		return false
	default:
		log.Panic("not fail-able with `bytes.Comparable` bounded [-1, 1].")
		return false
	}
}

func (b sortByteArrays) Swap(i, j int) {
	b[j], b[i] = b[i], b[j]
}

func TestCalculateRingVertexes_4(t *testing.T) {
	sh := sha256.Sum256([]byte("my-test-string-to-hash"))
	h1 := CalculateRingVertexBytes(sh[:], 4)
	h2 := CalculateRingVertexBytes(h1[1], 4)

	s1 := sortByteArrays(h1)
	sort.Sort(s1)
	s2 := sortByteArrays(h2)
	sort.Sort(s2)

	for i := range h1 {
		if bytes.Compare(h1[i], h2[i]) != 0 {
			t.Fatal("not equal")
		}
	}

}

func TestCalculateRingVertexes_8(t *testing.T) {
	sh := sha256.Sum256([]byte("my-test-string-to-hash"))
	h1 := CalculateRingVertexBytes(sh[:], 8)
	h2 := CalculateRingVertexBytes(h1[1], 8)

	s1 := sortByteArrays(h1)
	sort.Sort(s1)
	s2 := sortByteArrays(h2)
	sort.Sort(s2)

	for i := range h1 {
		if bytes.Compare(h1[i], h2[i]) != 0 {
			t.Logf("%x %x", h1[i], h2[i])
			t.Fatal("not equal")
		}
	}

}
