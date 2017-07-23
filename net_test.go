package hexaring

import (
	"crypto/sha1"
	"testing"
	"time"
)

func TestNetTransport(t *testing.T) {
	r1, err := initTestRing("127.0.0.1:12345")
	if err != nil {
		t.Fatal(err)
	}

	<-time.After(100 * time.Millisecond)
	// Test for first peer non-existent
	r2, err := initTestRing("127.0.0.1:23456", "127.0.0.1:12345")
	if err != nil {
		t.Fatal(err)
	}

	<-time.After(100 * time.Millisecond)

	if r1.NumSuccessors() != r2.NumSuccessors() {
		t.Fatal("successor mismatch")
	}

	client := NewNetClient(2*time.Second, 10*time.Second)
	locs, err := client.LookupReplicated("127.0.0.1:12345", testkey, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(locs) != 2 {
		t.Fatal("should have 2 locations")
	}

	vns, err := client.Lookup("127.0.0.1:23456", 3, testkey)
	if err != nil {
		t.Fatal(err)
	}
	if len(vns) != 3 {
		t.Fatal("should have 3 vnodes")
	}

	sh := sha1.Sum(testkey)

	locs1, err := client.LookupReplicatedHash("127.0.0.1:12345", sh[:], 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(locs1) != 2 {
		t.Fatal("should have 2 locations")
	}

	vns2, err := client.LookupHash("127.0.0.1:23456", 3, sh[:])
	if err != nil {
		t.Fatal(err)
	}
	if len(vns2) != 3 {
		t.Fatal("should have 3 vnodes")
	}

	// allow reap
	<-time.After(3 * time.Second)

	client.Shutdown()

	if _, err = client.Lookup("127.0.0.1:23456", 3, testkey); err == nil {
		t.Fatal("should fail with shutdown error")
	}
}
