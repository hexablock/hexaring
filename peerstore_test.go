package hexaring

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestInMemPeerStore(t *testing.T) {
	ps := NewInMemPeerStore()
	ps.AddPeer("peer")
	ps.AddPeer("peer2")
	if len(ps.Peers()) != 2 {
		t.Fatal("should have 2 peers")
	}

	ps.RemovePeer("peer2")
	if len(ps.Peers()) != 1 {
		t.Fatal("should have 1 peers")
	}
	ps.RemovePeer("peer")
	if len(ps.Peers()) != 0 {
		t.Fatal("should have 0 peers")
	}
}

func TestPeerStore(t *testing.T) {
	tf, _ := ioutil.TempFile("/tmp", "peerstore")
	tf.Write([]byte("[]"))
	tf.Close()
	defer os.Remove(tf.Name())

	ps, err := NewPeerJSONStore(tf.Name())
	if err != nil {
		t.Fatal(err)
	}

	if !ps.AddPeer("peer") {
		t.Fatal("should not have peer")
	}

	if ps.AddPeer("peer") {
		t.Fatal("should have peer")
	}

	if !ps.AddPeer("peer2") {
		t.Fatal("peer2 should've been added")
	}

	ps.RemovePeer("peer")
	ps.RemovePeer("peer2")
	if len(ps.Peers()) != 0 {
		t.Fatal("should have 0 peers")
	}
}
