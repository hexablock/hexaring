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

	locs1, err := r1.LookupReplicated(testkey, 3)
	if err != nil {
		t.Fatal(err)
	}
	locs2, err := r2.LookupReplicated(testkey, 3)
	if err != nil {
		t.Fatal(err)
	}
	locs3, err := r3.LookupReplicated(testkey, 3)
	if err != nil {
		t.Fatal(err)
	}

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

	next1, err := locs3.GetNext(locs3[1].Host())
	if err != nil {
		t.Fatal(err)
	}
	if next1.Vnode.StringID() != locs3[2].Vnode.StringID() {
		t.Fatal("wrong location")
	}

	next2, err := locs3.GetNext(locs3[2].Host())
	if err != nil {
		t.Fatal(err)
	}
	if next2.Vnode.StringID() != locs3[0].Vnode.StringID() {
		t.Fatal("wrong location")
	}

	if _, err = locs3.GetNext("host"); err == nil {
		t.Fatal("shoud error")
	}
}
