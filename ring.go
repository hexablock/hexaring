package hexaring

import (
	"bytes"
	"errors"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"

	chord "github.com/hexablock/go-chord"
	"github.com/hexablock/log"
)

var errNoPeersFound = errors.New("no peers found")

// Config contains the configuration options for the chord ring.  It augments the
// default configuration for convenience
// type Config struct {
// 	*chord.Config

// 	RPCTimeout  time.Duration
// 	MaxConnIdle time.Duration
// }
//

// Ring is a node part of the chord ring allowing to perform ring operations.  This is
// used on peers participating in the ring.
type Ring struct {
	*chord.Ring                        // Underlying chord ring
	conf          *chord.Config        // Hexaring config
	peers         PeerStore            // store containing known peers
	trans         *chord.GRPCTransport // Transport used by chord
	lookupService *NetTransport        // Serve up ring operations
}

// DefaultConfig returns a sane config
func DefaultConfig(hostname string) *chord.Config {
	// cfg := &Config{
	// 	Config:      chord.DefaultConfig(hostname),
	// 	RPCTimeout:  3 * time.Second,
	// 	MaxConnIdle: 5 * time.Minute,
	// }
	cfg := chord.DefaultConfig(hostname)
	cfg.NumVnodes = 5                  // lowered from 8
	cfg.StabilizeMin = 3 * time.Second // lowered from 15
	cfg.StabilizeMax = 7 * time.Second // lowered from 45
	//cfg.StabilizeThresh = 30 * time.Second // enables adaptive stabilization
	return cfg
}

// New instantiates a new ring
//func New(conf *Config, peers PeerStore, rpcTimeout, maxConnIdle time.Duration) *Ring {
func New(conf *chord.Config, peers PeerStore, trans *chord.GRPCTransport) *Ring {
	r := &Ring{
		conf:  conf,
		peers: peers,
		//trans: chord.NewGRPCTransport(rpcTimeout, maxConnIdle),
		trans: trans,
	}
	r.lookupService = NewNetTransport(r)

	return r
}

// RegisterServer registers the underlying transport to the grpc server
func (r *Ring) RegisterServer(server *grpc.Server) {
	// Register chord transport
	r.trans.RegisterServer(server)
	// Register hexing lookup service
	r.lookupService.RegisterServer(server)
}

// LookupReplicated returns vnodes where a key and n replicas are located.
func (r *Ring) LookupReplicated(key []byte, n int) (LocationSet, error) {
	h := r.conf.HashFunc()
	h.Write(key)
	sh := h.Sum(nil)
	return r.LookupReplicatedHash(sh[:], n)
}

// LookupReplicatedHashSerial returns vnodes where a key and n replicas are located.
// Each replica returned is a unique node.  It returns a n error if the lookup fails or
// enough unique nodes are not found.
func (r *Ring) LookupReplicatedHashSerial(hash []byte, n int) (LocationSet, error) {

	hashes := CalculateRingVertexBytes(hash, int64(n))
	locations := map[string]*Location{}

	for i, h := range hashes {
		// Lookup successors for the replicated hash with the maximum allowable
		// successors.
		vs, err := r.LookupHash(r.conf.NumSuccessors, h)
		if err != nil {
			return nil, err
		}
		// Go through each successor selecting the first one by host that we do not have.
		for j, vn := range vs {
			if _, ok := locations[vn.Host]; !ok {
				locations[vn.Host] = &Location{
					ID:       h,
					Priority: int32(i),
					Index:    int32(j),
					Vnode:    vn,
				}
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

// LookupReplicatedHash returns vnodes where a key and n replicas are located.  Each
// replica call is performed in its own go-routine. Each replica returned is a
// unique node.  It returns a n error if the lookup fails or enough unique nodes are
// not found.
func (r *Ring) LookupReplicatedHash(hash []byte, n int) (LocationSet, error) {
	hashes := CalculateRingVertexBytes(hash, int64(n))
	out := make(chan []*Location, n)

	var wg sync.WaitGroup
	wg.Add(n)
	for i, h := range hashes {
		// Lookup successors for the replicated hash with the maximum allowable
		// successors.
		go func(idx int, hsh []byte) {

			vs, err := r.LookupHash(r.conf.NumSuccessors, hsh)
			if err != nil {
				log.Println("[ERROR] Lookup failed:", err)
				//	return nil, err
				out <- nil
				return
			}

			locs := make([]*Location, len(vs))
			for j, v := range vs {
				locs[j] = &Location{ID: hsh, Vnode: v, Index: int32(j), Priority: int32(idx)}
			}
			out <- locs

			wg.Done()

		}(i, h)

	}

	go func() {
		wg.Wait()
		close(out)
	}()

	locations := make([][]*Location, n)
	// Sort by priority
	for la := range out {
		if la == nil {
			return nil, fmt.Errorf("not enough hosts found")
		}

		p := la[0].Priority
		locations[p] = la
	}

	locs := make(LocationSet, n)
	locs[0] = locations[0][0]

	// Update list with uniques by priority
	c := 1
	for _, l := range locations[1:] {
		for _, ll := range l {
			if containsHost(locs[:c], ll.Host()) {
				continue
			}

			locs[c] = ll
			c++
			break
		}
	}
	// Make sure we have the requested count
	for i := n - 1; i >= 0; i-- {
		if locs[i] == nil {
			return locs[:i], fmt.Errorf("not enough hosts found")
		}
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

// ScourSector scours all nodes between start and end hashes issueing the callback for
// the chosen vnodes.  It skips nodes that have already been visited.
func (r *Ring) ScourSector(start, end []byte, cb func(*chord.Vnode) error) (int, error) {
	var (
		cfunc func(a, b []byte) int
		max   = maxHash(len(start))
		zero  = make([]byte, len(start))
	)

	p := bytes.Compare(start, end)
	if p == 0 {
		return 0, fmt.Errorf("nothing to scour")
	} else if p > 0 {
		// Start > end
		cfunc = func(a, b []byte) int {
			if ((bytes.Compare(a, start) >= 0) && (bytes.Compare(a, max) <= 0)) ||
				((bytes.Compare(a, zero) >= 0) && (bytes.Compare(a, b) < 0)) {
				return -1
			}
			return 1
		}
	} else {
		// Start < end
		cfunc = func(a, b []byte) int { return bytes.Compare(a, b) }
	}

	// Proceed as normal where start < end
	visited := map[string]struct{}{}
	lkh := start
	for {

		vns, err := r.LookupHash(r.conf.NumSuccessors, lkh)
		if err != nil {
			return len(visited), err
		}

		for _, vn := range vns {
			if cfunc(vn.Id, end) > 0 {
				return len(visited), nil
			}

			// Skip if we've visted
			if _, ok := visited[vn.Host]; ok {
				continue
			}

			// Mark as visited
			visited[vn.Host] = struct{}{}

			// Return if callback returns an error
			if err = cb(vn); err != nil {
				return len(visited), err
			}
		}

		lkh = vns[len(vns)-1].Id

	}
}

// ScourReplica scours a replica location id upto the allowable number of vnodes.
func (r *Ring) ScourReplica(locID []byte, cb func(*chord.Vnode) error) (int, error) {
	visited := map[string]struct{}{}
	vns, err := r.LookupHash(r.conf.NumSuccessors, locID)
	if err != nil {
		return 0, err
	}

	for _, vn := range vns {
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

	return len(visited), nil
}

// Scour traverses each location up to the allowed number of succesors, issueing the
// callback for each node.  It skips nodes that have already been visited. If the callback
// returns an error, it is immediately exits.  It returns the number of nodes visited
// and/or an error either from the lookup or callback.
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

// Create creates a new ring.  This is only to be called once.
func (r *Ring) Create() error {
	ring, err := chord.Create(r.conf, r.trans)
	if err == nil {
		r.Ring = ring
	}
	return err
}

// Join tries to join an existing ring using any of the peers from the PeerStore.
func (r *Ring) Join() error {
	return joinRing(r, r.peers)
}

// RetryJoin keeps looping through the available peers to join.  It implements a backoff
// for each retry
func (r *Ring) RetryJoin() error {

	retryInSec := 2
	tries := 0

	for {
		tries++
		if tries == 3 {
			tries = 0
			retryInSec *= retryInSec
		}
		// Try each set of peers
		err := joinRing(r, r.peers)
		if err == nil {
			return nil
		}
		log.Printf("Failed to connect msg='%v'", err)

		// Wait before retying
		log.Printf("Trying in %d secs ...", retryInSec)
		<-time.After(time.Duration(retryInSec) * time.Second)
	}

}

// helper for join and retry-join
func joinRing(r *Ring, peerStore PeerStore) error {

	peers := peerStore.Peers()
	for _, peer := range peers {
		log.Printf("[INFO] Trying peer=%s", peer)

		ring, err := chord.Join(r.conf, r.trans, peer)
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

func containsHost(locs []*Location, host string) bool {
	for _, v := range locs {
		if v.Host() == host {
			return true
		}
	}
	return false
}
