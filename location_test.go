package hexaring

import (
	"testing"
	"time"
)

func TestLocation(t *testing.T) {
	r1, err := initTestRing("127.0.0.1:33333")
	if err != nil {
		t.Fatal(err)
	}
	<-time.After(100 * time.Millisecond)

	r2, err := initTestRing("127.0.0.1:44444", "127.0.0.1:33333")
	if err != nil {
		t.Fatal(err)
	}
	<-time.After(100 * time.Millisecond)

	r3, err := initTestRing("127.0.0.1:55555", "127.0.0.1:44444")
	if err != nil {
		t.Fatal(err)
	}
	<-time.After(100 * time.Millisecond)

	locs1, _ := r1.LookupReplicated(testkey, 3)
	locs2, _ := r2.LookupReplicated(testkey, 3)
	locs3, _ := r3.LookupReplicated(testkey, 3)

	end1, _ := locs1.EndRange(locs1[2].ID)
	if !equalBytes(end1, locs1[0].ID) {
		t.Fatal("wrong range")
	}
	end2, _ := locs2.EndRange(locs2[1].ID)
	if !equalBytes(end2, locs2[2].ID) {
		t.Fatal("wrong range")
	}
	end3, _ := locs3.EndRange(locs3[0].ID)
	if !equalBytes(end3, locs3[1].ID) {
		t.Fatal("wrong range")
	}
}
