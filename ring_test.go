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

func fastConf(host string) *chord.Config {
	conf := DefaultConfig(host)
	conf.Meta = chord.Meta{"key": []byte("test")}
	conf.StabilizeMin = time.Duration(15 * time.Millisecond)
	conf.StabilizeMax = time.Duration(45 * time.Millisecond)
	// enabled adaptive stabilization
	conf.StabilizeThresh = time.Duration(30 * time.Millisecond)
	return conf
}

func initTestRing(host string, peers ...string) (*Ring, error) {
	ln, err := net.Listen("tcp", host)
	if err != nil {
		return nil, err
	}
	server := grpc.NewServer()

	conf := fastConf(host)

	ps := NewInMemPeerStore()
	for _, p := range peers {
		ps.AddPeer(p)
	}

	trans := chord.NewGRPCTransport(2*time.Second, 1*time.Minute)

	r := New(conf, ps, trans)
	r.RegisterServer(server)

	go server.Serve(ln)

	if len(peers) == 0 {
		return r, r.Create()
	}

	return r, r.RetryJoin()
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

}

func TestRing_ScourReplicatedKey(t *testing.T) {
	r1, err := initTestRing("127.0.0.1:12445")
	if err != nil {
		t.Fatal(err)
	}
	//vns1, _ := r1.trans.ListVnodes("127.0.0.1:12445")
	<-time.After(100 * time.Millisecond)

	// Test for first peer non-existent
	r2, err := initTestRing("127.0.0.1:22556", "127.0.0.1:12445")
	if err != nil {
		t.Fatal(err)
	}
	//vns2, _ := r2.trans.ListVnodes("127.0.0.1:22556")
	<-time.After(100 * time.Millisecond)

	r3, err := initTestRing("127.0.0.1:22566", "127.0.0.1:12445")
	if err != nil {
		t.Fatal(err)
	}
	//vns3, _ := r3.trans.ListVnodes("127.0.0.1:22566")
	<-time.After(200 * time.Millisecond)

	key := []byte("some-data")

	var c1 int
	o1, err := r1.ScourReplicatedKey(key, 3, func(vn *chord.Vnode) error {
		c1++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	var c2 int
	o2, err := r2.ScourReplicatedKey(key, 3, func(vn *chord.Vnode) error {
		c2++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	var c3 int
	o3, err := r3.ScourReplicatedKey(key, 3, func(vn *chord.Vnode) error {
		c3++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if c1 != c2 || c2 != c3 || c3 != c1 {
		t.Fatal("scour count mismatch")
	}

	if o1 != c1 || o2 != c2 || o3 != c3 {
		t.Fatal("scour count mismatch")
	}

	slocs, _ := r1.LookupReplicated([]byte("mykey"), 3)
	for i, v := range slocs {
		t.Logf("loc.%d %x\n", i, v.ID)
	}

	if _, err = r1.ScourSector(slocs[1].ID, slocs[1].ID, func(*chord.Vnode) error { return nil }); err == nil {
		t.Fatal("should fail")
	}

	var c int
	r2.ScourSector(slocs[0].ID, slocs[1].ID, func(vn *chord.Vnode) error {
		c++
		return nil
	})

	r3.ScourSector(slocs[2].ID, slocs[0].ID, func(vn *chord.Vnode) error {
		c++
		return nil
	})

	if c <= 3 {
		t.Fatal("scour didnt visit required hosts")
	}

	// Wrap around test
	vst, _ := r1.ScourSector(slocs[1].ID, slocs[2].ID, func(vn *chord.Vnode) error {
		t.Logf("wrap %s\n", vn.Host)
		return nil
	})
	if vst != 3 {
		t.Fatal("should have visited 3 hosts")
	}
}
