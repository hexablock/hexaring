package hexaring

import (
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc"

	chord "github.com/hexablock/go-chord"
	"github.com/hexablock/log"
)

var errNoPeersFound = errors.New("no peers found")

// Config contains the configuration options for the chord ring.  It augments the
// default configuration for convenience
type Config struct {
	*chord.Config

	RPCTimeout  time.Duration
	MaxConnIdle time.Duration
}

// Ring is a node part of the chord ring allowing to perform ring operations.  This is
// used on peers participating in the ring.
type Ring struct {
	*chord.Ring                        // Underlying chord ring
	conf          *Config              // Hexaring config
	trans         *chord.GRPCTransport // Transport used by chord
	lookupService *NetTransport        // Serve up ring operations
}

// DefaultConfig returns a sane config
func DefaultConfig(hostname string) *Config {
	cfg := &Config{
		Config:      chord.DefaultConfig(hostname),
		RPCTimeout:  3 * time.Second,
		MaxConnIdle: 5 * time.Minute,
	}

	return cfg
}

func newRing(conf *Config, server *grpc.Server) *Ring {
	return &Ring{
		conf:  conf,
		trans: chord.NewGRPCTransport(server, conf.RPCTimeout, conf.MaxConnIdle),
	}
}

// LookupReplicated returns vnodes where a key and n replicas are located.
func (r *Ring) LookupReplicated(key []byte, n int) (LocationSet, error) {
	h := r.conf.HashFunc()
	h.Write(key)
	sh := h.Sum(nil)
	return r.LookupReplicatedHash(sh[:], n)
}

// LookupReplicatedHash returns vnodes where a key and n replicas are located.
// Each replica returned is a unique node.  It returns a n error if the lookup fails or
// enough unique nodes are not found.
func (r *Ring) LookupReplicatedHash(hash []byte, n int) (LocationSet, error) {

	hashes := CalculateRingVertexBytes(hash, int64(n))
	locations := map[string]*Location{}

	for i, h := range hashes {
		// Lookup successors for the replicated hash with the maximum allowable
		// successors.
		vs, err := r.LookupHash(r.conf.NumSuccessors, h)
		if err != nil {
			return nil, err
		}
		// Go through each successor selecting the first one by host that we do not
		// have..
		for _, vn := range vs {
			if _, ok := locations[vn.Host]; !ok {
				locations[vn.Host] = &Location{ID: h, Vnode: vn, Priority: int32(i)}
				break
			}
		}
	}
	// Check if we have the requested number of replicas
	if len(locations) != n {
		return nil, fmt.Errorf("not enough hosts found")
	}

	// Re-arrange by highest priority first
	locs := make(LocationSet, n)
	for _, v := range locations {
		locs[v.Priority] = v
	}

	return locs, nil
}

// Hostname returns the hostname of the node per the config.
func (r *Ring) Hostname() string {
	return r.conf.Hostname
}

// NumSuccessors returns the num of succesors per the config.
func (r *Ring) NumSuccessors() int {
	return r.conf.NumSuccessors
}

// Create creates a new ring based on the config
func Create(conf *Config, server *grpc.Server) (*Ring, error) {
	r := newRing(conf, server)

	ring, err := chord.Create(r.conf.Config, r.trans)
	if err == nil {
		r.Ring = ring
		r.lookupService = &NetTransport{r}
		RegisterLookupRPCServer(server, r.lookupService)
	}
	return r, err
}

// Join tries to join an existing ring using any of the peers from the PeerStore.
func Join(conf *Config, peerStore PeerStore, server *grpc.Server) (*Ring, error) {
	r := newRing(conf, server)
	err := joinRing(r, peerStore, server)
	return r, err
}

// RetryJoin implements exponential backoff rejoin.
func RetryJoin(conf *Config, peerStore PeerStore, server *grpc.Server) (*Ring, error) {

	retryInSec := 2
	tries := 0
	r := newRing(conf, server)

	for {
		tries++
		if tries == 3 {
			tries = 0
			retryInSec *= retryInSec
		}
		// Try each set of peers
		err := joinRing(r, peerStore, server)
		if err == nil {
			return r, nil
		}
		log.Printf("Failed to connect msg='%v'", err)

		// Wait before retying
		log.Printf("Trying in %d secs ...", retryInSec)
		<-time.After(time.Duration(retryInSec) * time.Second)
	}

}

func joinRing(r *Ring, peerStore PeerStore, server *grpc.Server) error {

	peers := peerStore.Peers()
	for _, peer := range peers {
		log.Printf("[INFO] Trying peer=%s", peer)

		ring, err := chord.Join(r.conf.Config, r.trans, peer)
		if err == nil {
			r.Ring = ring
			r.lookupService = &NetTransport{r}
			RegisterLookupRPCServer(server, r.lookupService)
			return nil
		}
		log.Printf("[ERROR] Failed to connect peer=%s msg='%v'", peer, err)

		// Wait before trying next peer
		<-time.After(500 * time.Millisecond)
	}

	return fmt.Errorf("all peers exhausted")
}
