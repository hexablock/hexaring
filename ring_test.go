package hexaring

import (
	"bytes"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"

	chord "github.com/hexablock/go-chord"
)

var testkey = []byte("testkey")

func fastConf(host string) *Config {
	conf := DefaultConfig(host)
	conf.Meta = chord.Meta{"key": []byte("test")}
	conf.StabilizeMin = time.Duration(15 * time.Millisecond)
	conf.StabilizeMax = time.Duration(45 * time.Millisecond)
	return conf
}

func initTestRing(host string, peers ...string) (*Ring, error) {

	ln, _ := net.Listen("tcp", host)
	server := grpc.NewServer()
	go server.Serve(ln)

	conf := fastConf(host)

	if len(peers) == 0 {
		return Create(conf, server)
	}

	ps := NewInMemPeerStore()
	for _, p := range peers {
		ps.AddPeer(p)
	}

	return RetryJoin(conf, ps, server)
}

func TestRing(t *testing.T) {
	r1, err := initTestRing("127.0.0.1:33445")
	if err != nil {
		t.Fatal(err)
	}

	<-time.After(100 * time.Millisecond)
	// Test for first peer non-existent
	r2, err := initTestRing("127.0.0.1:44556", "127.0.0.1:65432", "127.0.0.1:33445")
	if err != nil {
		t.Fatal(err)
	}

	<-time.After(100 * time.Millisecond)

	if r1.Hostname() != "127.0.0.1:33445" {
		t.Fatal("wrong hostname")
	}
	if r2.Hostname() != "127.0.0.1:44556" {
		t.Fatal("wrong hostname")
	}

	locs1, err := r1.LookupReplicated(testkey, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(locs1) != 2 {
		t.Fatal("should have 2 locations")
	}

	locs2, err := r2.LookupReplicated(testkey, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(locs2) != 2 {
		t.Fatal("should have 2 locations")
	}

	if bytes.Compare(locs1[0].ID, locs2[0].ID) != 0 {
		t.Fatal("id mismatch")
	}

	if bytes.Compare(locs1[1].ID, locs2[1].ID) != 0 {
		t.Fatal("id mismatch")
	}

	if _, err = r1.LookupReplicated(testkey, 3); err == nil {
		t.Fatal("should fail")
	}

	vns1, _ := r1.Vnodes("")
	vns2, _ := r2.Vnodes("127.0.0.1:33445")

	if len(vns1) != len(vns2) {
		t.Fatal("different vnode count")
	}

}
