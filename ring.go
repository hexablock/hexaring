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
	peers         PeerStore            // store containing known peers
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
	cfg.NumVnodes = 5                   // lowered from 8
	cfg.StabilizeMin = 10 * time.Second // lowered from 15
	cfg.StabilizeMax = 15 * time.Second // lowered from 45

	return cfg
}

func newRing(conf *Config, peers PeerStore, server *grpc.Server) *Ring {
	r := &Ring{
		conf:  conf,
		peers: peers,
		trans: chord.NewGRPCTransport(server, conf.RPCTimeout, conf.MaxConnIdle),
	}
	r.lookupService = NewNetTransport(server, r)

	return r
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

// Scour traverses each location up to the allowed number of succesors, issueing the
// callback for each node.  If the callback returns an error, it is immediately exits.  It
// returns the number of nodes visited and/or an error either from the lookup or callback.
func (r *Ring) Scour(locs LocationSet, cb func(*chord.Vnode) error) (int, error) {
	// Visited hosts
	visited := map[string]struct{}{}
	// Query primary replica locations first
	for _, loc := range locs {
		visited[loc.Vnode.Host] = struct{}{}
		// Return if callback returns an error
		if err := cb(loc.Vnode); err != nil {
			return len(visited), err
		}
	}

	var err error
	// Query succesors of each replica location
	for _, loc := range locs {
		// Get succesors of location.  Continue to the next if we fail
		vns, er := r.LookupHash(r.conf.NumSuccessors, loc.ID)
		if er != nil {
			err = er
			continue
		}

		// Query each successor
		for _, vn := range vns[1:] {
			// Skip if we've visted
			if _, ok := visited[vn.Host]; ok {
				continue
			}
			visited[vn.Host] = struct{}{}

			// Return if callback returns an error
			if err = cb(vn); err != nil {
				return len(visited), err
			}

		}
	}

	return len(visited), err
}

// ScourReplicatedKey finds the replca hash locations for the given key and calls Scour on
// each location.  This is a helper function to Scour.
func (r *Ring) ScourReplicatedKey(key []byte, replicas int, cb func(*chord.Vnode) error) (int, error) {
	locs, err := r.LookupReplicated(key, replicas)
	if err != nil {
		return 0, err
	}

	return r.Scour(locs, cb)
}

// Create creates a new ring based on the config
func Create(conf *Config, peers PeerStore, server *grpc.Server) (*Ring, error) {
	r := newRing(conf, peers, server)

	ring, err := chord.Create(r.conf.Config, r.trans)
	if err == nil {
		r.Ring = ring
	}

	return r, err
}

// Join tries to join an existing ring using any of the peers from the PeerStore.
func Join(conf *Config, peerStore PeerStore, server *grpc.Server) (*Ring, error) {
	r := newRing(conf, peerStore, server)
	err := joinRing(r, peerStore, server)
	return r, err
}

// RetryJoin implements exponential backoff rejoin.
func RetryJoin(conf *Config, peerStore PeerStore, server *grpc.Server) (*Ring, error) {

	retryInSec := 2
	tries := 0
	r := newRing(conf, peerStore, server)

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
			return nil
		}
		log.Printf("[ERROR] Failed to connect peer=%s msg='%v'", peer, err)

		// Wait before trying next peer
		<-time.After(500 * time.Millisecond)
	}

	return fmt.Errorf("all peers exhausted")
}
